package usecase

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

	"github.com/jackc/pgx/v5"
	repoModels "github.com/twelvepills-936/tgapp-/internal/repository/models"
	ucModels "github.com/twelvepills-936/tgapp-/internal/usecase/models"
)

func (uc *useCase) RegisterByTelegram(ctx context.Context, input ucModels.RegisterByTelegramInput) (output ucModels.RegisterByTelegramOutput, err error) {
	// Validate input
	if err := input.Validate(); err != nil {
		return ucModels.RegisterByTelegramOutput{}, err
	}

	// decode init_data_raw (base64) and parse user json param similar to Node parseAuthToken path
	decoded, decodeErr := base64.StdEncoding.DecodeString(input.InitDataRaw)
	if decodeErr != nil {
		return output, fmt.Errorf("%w: failed to decode init data", ucModels.ErrInvalidInput)
	}
	params, parseErr := url.ParseQuery(string(decoded))
	if parseErr != nil {
		return output, fmt.Errorf("%w: failed to parse init data", ucModels.ErrInvalidInput)
	}
	userStr := params.Get("user")
	if userStr == "" {
		return output, fmt.Errorf("%w: user not found in init data", ucModels.ErrInvalidInput)
	}
	startParam := input.StartParam
	if startParam == "" {
		startParam = params.Get("start_param")
	}
	// we don't need full user struct now; minimally extract fields from json
	// to keep scope tight, parse a subset via a lightweight map
	tg := struct {
		ID           int64  `json:"id"`
		FirstName    string `json:"first_name"`
		Username     string `json:"username"`
		PhotoURL     string `json:"photo_url"`
		LanguageCode string `json:"language_code"`
	}{}
	if unmarshalErr := json.Unmarshal([]byte(userStr), &tg); unmarshalErr != nil {
		return output, fmt.Errorf("%w: failed to unmarshal user data", ucModels.ErrInvalidInput)
	}
	if tg.ID == 0 {
		return output, fmt.Errorf("%w: telegram user id is missing", ucModels.ErrInvalidInput)
	}

	tx, txErr := uc.repo.DBBeginTransaction(ctx)
	if txErr != nil {
		return output, txErr
	}
	defer func() {
		if err != nil && tx != nil {
			// Use background context for rollback to avoid context cancellation issues
			_ = tx.Rollback(context.Background())
		}
	}()

	refereeTelegramID := intToString(tg.ID)

	// if exists, still try to link referral from start_param, then return already registered
	if existing, checkErr := uc.repo.GetProfileByTelegramID(ctx, tx, refereeTelegramID); checkErr == nil {
		if startParam != "" {
			if _, linkErr := uc.linkReferral(ctx, tx, existing.ID, refereeTelegramID, startParam); linkErr != nil {
				err = linkErr
				return output, err
			}
		}
		if commitErr := tx.Commit(ctx); commitErr != nil {
			err = commitErr
			return output, err
		}
		err = ucModels.ErrProfileAlreadyRegistered
		return output, err
	} else if !errors.Is(checkErr, pgx.ErrNoRows) {
		err = checkErr
		return output, err
	}

	pid, createErr := uc.repo.CreateProfile(ctx, tx, repoModels.Profile{
		Name:             tg.FirstName,
		TelegramID:       refereeTelegramID,
		Avatar:           tg.PhotoURL,
		Location:         tg.LanguageCode,
		Role:             "",
		Description:      "",
		TelegramInitData: input.InitDataRaw,
		Username:         tg.Username,
		Verified:         false,
	})
	if createErr != nil {
		err = createErr
		return output, err
	}

	if _, walletErr := uc.repo.CreateWalletForUser(ctx, tx, pid); walletErr != nil {
		err = walletErr
		return output, err
	}

	if bonusErr := uc.grantRegistrationBonus(ctx, tx, pid); bonusErr != nil {
		err = bonusErr
		return output, err
	}

	wasReferred := false
	if startParam != "" {
		linked, linkErr := uc.linkReferral(ctx, tx, pid, refereeTelegramID, startParam)
		if linkErr != nil {
			err = linkErr
			return output, err
		}
		wasReferred = linked
	}

	if wasReferred {
		if bonusErr := uc.grantReferralSignupBonus(ctx, tx, pid); bonusErr != nil {
			err = bonusErr
			return output, err
		}
	}

	if commitErr := tx.Commit(ctx); commitErr != nil {
		err = commitErr
		return output, err
	}

	return ucModels.RegisterByTelegramOutput{ProfileID: pid}, nil
}

func (uc *useCase) grantRegistrationBonus(ctx context.Context, tx pgx.Tx, profileID int64) error {
	raw, err := uc.repo.GetAdminSettings(ctx, tx)
	if err != nil {
		return err
	}

	settings := mergeAdminSettings(raw)
	if settings.RegistrationBonus <= 0 {
		return nil
	}

	_, err = uc.repo.CreditProfileTokens(
		ctx,
		tx,
		profileID,
		0,
		settings.RegistrationBonus,
		"registration_bonus",
	)
	return err
}

// grantReferralSignupBonus rewards a newly registered user who came in via a
// referral link, on top of the standard registration bonus.
func (uc *useCase) grantReferralSignupBonus(ctx context.Context, tx pgx.Tx, profileID int64) error {
	raw, err := uc.repo.GetAdminSettings(ctx, tx)
	if err != nil {
		return err
	}

	settings := mergeAdminSettings(raw)
	if settings.ReferralRefereeBonus <= 0 {
		return nil
	}

	_, err = uc.repo.CreditProfileTokens(
		ctx,
		tx,
		profileID,
		0,
		settings.ReferralRefereeBonus,
		"referral_signup_bonus",
	)
	return err
}

func (uc *useCase) GetUserByTelegramID(ctx context.Context, telegramID string) (ucModels.GetProfileOutput, error) {
	// Validate telegram_id
	if telegramID == "" {
		return ucModels.GetProfileOutput{}, fmt.Errorf("telegram_id is required")
	}

	if len(telegramID) > 100 {
		return ucModels.GetProfileOutput{}, fmt.Errorf("telegram_id too long")
	}

	p, err := uc.repo.GetProfileByTelegramID(ctx, nil, telegramID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ucModels.GetProfileOutput{}, ucModels.ErrProfileNotFound
		}
		return ucModels.GetProfileOutput{}, err
	}
	theme := p.UITheme
	if theme == "" {
		theme = ucModels.ThemeLight
	}
	return ucModels.GetProfileOutput{Data: ucModels.ProfileUser{
		ID:         p.ID,
		Name:       p.Name,
		TelegramID: p.TelegramID,
		Avatar:     p.Avatar,
		Username:   p.Username,
		Verified:   p.Verified,
		Theme:      theme,
	}}, nil
}

func (uc *useCase) UpdateProfileTheme(ctx context.Context, input ucModels.UpdateProfileThemeInput) (ucModels.UpdateProfileThemeOutput, error) {
	if err := input.Validate(); err != nil {
		return ucModels.UpdateProfileThemeOutput{}, err
	}

	if err := uc.repo.UpdateProfileTheme(ctx, nil, input.TelegramID, input.Theme); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ucModels.UpdateProfileThemeOutput{}, ucModels.ErrProfileNotFound
		}
		return ucModels.UpdateProfileThemeOutput{}, err
	}
	return ucModels.UpdateProfileThemeOutput{Theme: input.Theme}, nil
}

// helpers
func intToString(v int64) string { return fmt.Sprintf("%d", v) }

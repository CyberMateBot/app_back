package usecase

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	repoModels "github.com/twelvepills-936/tgapp-/internal/repository/models"
	ucModels "github.com/twelvepills-936/tgapp-/internal/usecase/models"
	"github.com/twelvepills-936/tgapp-/pkg/jwtutil"
	"golang.org/x/crypto/bcrypt"
)

func (uc *useCase) BootstrapAdmin(ctx context.Context) error {
	email := strings.TrimSpace(strings.ToLower(os.Getenv("ADMIN_EMAIL")))
	password := os.Getenv("ADMIN_PASSWORD")
	if email == "" || password == "" {
		return nil
	}
	if uc.jwt.Secret == "" {
		return nil
	}

	if _, err := uc.repo.GetAdminByEmail(ctx, nil, email); err == nil {
		return nil
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return err
	}

	hash, hashErr := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if hashErr != nil {
		return hashErr
	}
	_, createErr := uc.repo.CreateAdmin(ctx, nil, email, string(hash))
	return createErr
}

func (uc *useCase) AdminLogin(ctx context.Context, input ucModels.AdminLoginInput) (ucModels.AdminLoginOutput, error) {
	var out ucModels.AdminLoginOutput
	if err := input.Validate(); err != nil {
		return out, err
	}
	if uc.jwt.Secret == "" {
		return out, fmt.Errorf("%w: jwt is not configured", ucModels.ErrInternalServerError)
	}

	email := strings.ToLower(strings.TrimSpace(input.Email))
	admin, err := uc.repo.GetAdminByEmail(ctx, nil, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return out, ucModels.ErrInvalidCredentials
		}
		return out, err
	}

	if cmpErr := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(input.Password)); cmpErr != nil {
		return out, ucModels.ErrInvalidCredentials
	}

	ttl := uc.jwt.AccessTokenTTL
	if ttl < 24*time.Hour {
		ttl = 24 * time.Hour
	}
	token, signErr := jwtutil.SignAdminToken(uc.jwt.Secret, ttl, admin.ID, admin.Email)
	if signErr != nil {
		return out, signErr
	}

	return ucModels.AdminLoginOutput{
		Token: token,
		Admin: ucModels.AdminUser{ID: admin.ID, Email: admin.Email},
	}, nil
}

func (uc *useCase) GetAdmin(ctx context.Context, adminID int64) (ucModels.AdminUser, error) {
	if adminID <= 0 {
		return ucModels.AdminUser{}, fmt.Errorf("%w: admin_id invalid", ucModels.ErrInvalidInput)
	}
	admin, err := uc.repo.GetAdminByID(ctx, nil, adminID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ucModels.AdminUser{}, ucModels.ErrAdminNotFound
		}
		return ucModels.AdminUser{}, err
	}
	return ucModels.AdminUser{ID: admin.ID, Email: admin.Email}, nil
}

func (uc *useCase) GetAdminStats(ctx context.Context) (ucModels.AdminStatsOutput, error) {
	stats, err := uc.repo.GetAdminStats(ctx, nil)
	if err != nil {
		return ucModels.AdminStatsOutput{}, err
	}
	return ucModels.AdminStatsOutput{
		TotalUsers:       stats.TotalUsers,
		ActiveUsersToday: stats.ActiveUsersToday,
		NewUsersToday:    stats.NewUsersToday,
		TotalMessages:    stats.TotalMessages,
	}, nil
}

func (uc *useCase) ListAdminUsers(ctx context.Context, input ucModels.AdminListUsersInput) (ucModels.AdminListUsersOutput, error) {
	input.Normalize()
	offset := (input.Page - 1) * input.PerPage
	items, total, err := uc.repo.ListAdminProfiles(ctx, nil, input.Search, input.PerPage, offset)
	if err != nil {
		return ucModels.AdminListUsersOutput{}, err
	}

	plans, planErr := uc.loadSubscriptionPlans(ctx)
	if planErr != nil {
		plans = nil
	}

	out := ucModels.AdminListUsersOutput{
		Data:  make([]ucModels.AdminUserItem, 0, len(items)),
		Total: total,
	}
	for _, p := range items {
		out.Data = append(out.Data, uc.mapAdminProfile(p, plans))
	}
	return out, nil
}

func (uc *useCase) GetAdminUser(ctx context.Context, userID int64) (ucModels.AdminUserItem, error) {
	if userID <= 0 {
		return ucModels.AdminUserItem{}, fmt.Errorf("%w: user_id invalid", ucModels.ErrInvalidInput)
	}
	p, err := uc.repo.GetAdminProfileByID(ctx, nil, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ucModels.AdminUserItem{}, ucModels.ErrAdminUserNotFound
		}
		return ucModels.AdminUserItem{}, err
	}
	plans, planErr := uc.loadSubscriptionPlans(ctx)
	if planErr != nil {
		plans = nil
	}
	return uc.mapAdminProfile(p, plans), nil
}

func (uc *useCase) UpdateAdminUserActive(ctx context.Context, input ucModels.AdminUpdateUserInput) (ucModels.AdminUserItem, error) {
	if input.UserID <= 0 {
		return ucModels.AdminUserItem{}, fmt.Errorf("%w: user_id invalid", ucModels.ErrInvalidInput)
	}
	if err := uc.repo.UpdateProfileActive(ctx, nil, input.UserID, input.IsActive); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ucModels.AdminUserItem{}, ucModels.ErrAdminUserNotFound
		}
		return ucModels.AdminUserItem{}, err
	}
	return uc.GetAdminUser(ctx, input.UserID)
}

func (uc *useCase) DeleteAdminUser(ctx context.Context, userID int64) error {
	if userID <= 0 {
		return fmt.Errorf("%w: user_id invalid", ucModels.ErrInvalidInput)
	}
	err := uc.repo.DeleteProfile(ctx, nil, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ucModels.ErrAdminUserNotFound
	}
	return err
}

func (uc *useCase) AdminCreditTokens(ctx context.Context, input ucModels.AdminTokenChangeInput) (ucModels.AdminTokenChangeOutput, error) {
	var out ucModels.AdminTokenChangeOutput
	if err := input.Validate(); err != nil {
		return out, err
	}

	tx, err := uc.repo.DBBeginTransaction(ctx)
	if err != nil {
		return out, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(context.Background())
		}
	}()

	result, err := uc.repo.CreditProfileTokens(ctx, tx, input.UserID, input.AdminID, input.Amount, input.Reason)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return out, ucModels.ErrAdminUserNotFound
		}
		return out, err
	}
	if commitErr := tx.Commit(ctx); commitErr != nil {
		return out, commitErr
	}

	return ucModels.AdminTokenChangeOutput{
		UserID:    result.ProfileID,
		Tokens:    result.BalanceAfter,
		Delta:     input.Amount,
		Operation: "credit",
	}, nil
}

func (uc *useCase) AdminDebitTokens(ctx context.Context, input ucModels.AdminTokenChangeInput) (ucModels.AdminTokenChangeOutput, error) {
	var out ucModels.AdminTokenChangeOutput
	if err := input.Validate(); err != nil {
		return out, err
	}

	tx, err := uc.repo.DBBeginTransaction(ctx)
	if err != nil {
		return out, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(context.Background())
		}
	}()

	result, err := uc.repo.DebitProfileTokens(ctx, tx, input.UserID, input.AdminID, input.Amount, input.Reason)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return out, ucModels.ErrAdminUserNotFound
		}
		if errors.Is(err, repoModels.ErrInsufficientTokens) {
			return out, ucModels.ErrInsufficientTokens
		}
		return out, err
	}
	if commitErr := tx.Commit(ctx); commitErr != nil {
		return out, commitErr
	}

	return ucModels.AdminTokenChangeOutput{
		UserID:    result.ProfileID,
		Tokens:    result.BalanceAfter,
		Delta:     -input.Amount,
		Operation: "debit",
	}, nil
}

func (uc *useCase) AdminBroadcast(ctx context.Context, input ucModels.AdminBroadcastInput, messenger interface {
	Active() bool
	SendText(chatID int64, text, parseMode string) error
}) (ucModels.AdminBroadcastOutput, error) {
	var out ucModels.AdminBroadcastOutput
	if err := input.Validate(); err != nil {
		return out, err
	}
	if messenger == nil || !messenger.Active() {
		return out, ucModels.ErrBroadcastNotReady
	}

	ids, err := uc.repo.ListBroadcastTelegramIDs(ctx, nil, input.Target == "active")
	if err != nil {
		return out, err
	}

	parseMode := strings.TrimSpace(input.ParseMode)
		for _, rawID := range ids {
		chatID, parseErr := strconv.ParseInt(strings.TrimSpace(rawID), 10, 64)
		if parseErr != nil || chatID <= 0 {
			out.Failed++
			continue
		}
		if sendErr := messenger.SendText(chatID, input.Message, parseMode); sendErr != nil {
			out.Failed++
			continue
		}
		out.Sent++
	}

	if input.AdminID > 0 {
		_, _ = uc.repo.CreateAdminBroadcast(ctx, nil, input.AdminID, input.Message, input.Target, parseMode, out.Sent, out.Failed)
	}

	return out, nil
}

func (uc *useCase) mapAdminProfile(p repoModels.AdminProfile, plans []ucModels.SubscriptionPlanItem) ucModels.AdminUserItem {
	tgID, _ := strconv.ParseInt(p.TelegramID, 10, 64)
	item := ucModels.AdminUserItem{
		ID:         p.ID,
		TelegramID: tgID,
		Username:   p.Username,
		FirstName:  p.Name,
		LastName:   "",
		IsActive:   p.IsActive,
		Tokens:     p.Tokens,
		CreatedAt:  p.CreatedAt.UTC().Format(time.RFC3339),
	}

	state := uc.buildSubscriptionState(p.SubscriptionPlanID, p.SubscriptionStarted, p.SubscriptionExpires, plans)
	item.SubscriptionPlanID = state.PlanID
	item.SubscriptionPlan = state.PlanName
	item.SubscriptionExpires = state.ExpiresAt
	item.SubscriptionDaysLeft = state.DaysLeft
	return item
}

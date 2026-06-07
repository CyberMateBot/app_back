package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	repoModels "github.com/twelvepills-936/tgapp-/internal/repository/models"
	ucModels "github.com/twelvepills-936/tgapp-/internal/usecase/models"
	"github.com/twelvepills-936/tgapp-/pkg/config"
	"github.com/twelvepills-936/tgapp-/pkg/jwtutil"
	"golang.org/x/crypto/bcrypt"
)

// NewWebAuthUseCase adds web-auth config to the base usecase.
// It is optional: if not called, web auth endpoints will not be registered.
func (uc *useCase) SetWebAuthConfig(cfg config.ConfigJWT) {
	uc.jwt = cfg
}

func (uc *useCase) RegisterWebAccount(ctx context.Context, input ucModels.RegisterWebAccountInput) (out ucModels.AuthTokensOutput, err error) {
	if err := input.Validate(); err != nil {
		return out, err
	}
	if uc.jwt.Secret == "" {
		return out, fmt.Errorf("%w: jwt is not configured", ucModels.ErrInternalServerError)
	}

	email := strings.TrimSpace(strings.ToLower(input.Email))

	// Check existence.
	if _, getErr := uc.repo.GetWebAccountByEmail(ctx, nil, email); getErr == nil {
		return out, ucModels.ErrWebAccountAlreadyExists
	} else if getErr != nil && !errors.Is(getErr, pgx.ErrNoRows) {
		return out, getErr
	}

	hash, hashErr := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if hashErr != nil {
		return out, hashErr
	}

	tx, txErr := uc.repo.DBBeginTransaction(ctx)
	if txErr != nil {
		return out, txErr
	}
	defer func() {
		if err != nil && tx != nil {
			_ = tx.Rollback(context.Background())
		}
	}()

	id, createErr := uc.repo.CreateWebAccount(ctx, tx, repoModels.WebAccount{
		Email:        email,
		PasswordHash: string(hash),
		DisplayName:  input.DisplayName,
	})
	if createErr != nil {
		err = createErr
		return out, err
	}

	if commitErr := tx.Commit(ctx); commitErr != nil {
		err = commitErr
		return out, err
	}

	token, signErr := jwtutil.SignAccessToken(uc.jwt.Secret, uc.jwt.AccessTokenTTL, id, email)
	if signErr != nil {
		return out, signErr
	}
	return ucModels.AuthTokensOutput{AccessToken: token}, nil
}

func (uc *useCase) LoginWebAccount(ctx context.Context, input ucModels.LoginWebAccountInput) (out ucModels.AuthTokensOutput, err error) {
	if err := input.Validate(); err != nil {
		return out, err
	}
	if uc.jwt.Secret == "" {
		return out, fmt.Errorf("%w: jwt is not configured", ucModels.ErrInternalServerError)
	}

	email := strings.TrimSpace(strings.ToLower(input.Email))
	acc, getErr := uc.repo.GetWebAccountByEmail(ctx, nil, email)
	if getErr != nil {
		if errors.Is(getErr, pgx.ErrNoRows) {
			return out, ucModels.ErrInvalidCredentials
		}
		return out, getErr
	}

	if cmpErr := bcrypt.CompareHashAndPassword([]byte(acc.PasswordHash), []byte(input.Password)); cmpErr != nil {
		return out, ucModels.ErrInvalidCredentials
	}

	token, signErr := jwtutil.SignAccessToken(uc.jwt.Secret, uc.jwt.AccessTokenTTL, acc.ID, acc.Email)
	if signErr != nil {
		return out, signErr
	}
	return ucModels.AuthTokensOutput{AccessToken: token}, nil
}

func (uc *useCase) GetWebAccount(ctx context.Context, webAccountID int64) (out ucModels.GetWebAccountOutput, err error) {
	if webAccountID <= 0 {
		return out, fmt.Errorf("%w: web_account_id invalid", ucModels.ErrInvalidInput)
	}
	acc, getErr := uc.repo.GetWebAccountByID(ctx, nil, webAccountID)
	if getErr != nil {
		if errors.Is(getErr, pgx.ErrNoRows) {
			return out, ucModels.ErrWebAccountNotFound
		}
		return out, getErr
	}
	return ucModels.GetWebAccountOutput{Data: ucModels.WebAccount{
		ID:          acc.ID,
		Email:       acc.Email,
		DisplayName: acc.DisplayName,
	}}, nil
}

func (uc *useCase) CreateWebPrompt(ctx context.Context, input ucModels.CreateWebPromptInput) (out ucModels.CreateWebPromptOutput, err error) {
	if err := input.Validate(); err != nil {
		return out, err
	}

	id, createErr := uc.repo.CreateWebPrompt(ctx, nil, repoModels.WebPrompt{
		WebAccountID: input.WebAccountID,
		Prompt:       input.Prompt,
		Category:     input.Category,
		Model:        input.Model,
	})
	if createErr != nil {
		return out, createErr
	}
	return ucModels.CreateWebPromptOutput{ID: id}, nil
}

func (uc *useCase) ListWebPrompts(ctx context.Context, input ucModels.ListWebPromptsInput) (out ucModels.ListWebPromptsOutput, err error) {
	if err := input.Validate(); err != nil {
		return out, err
	}
	items, listErr := uc.repo.ListWebPrompts(ctx, nil, input.WebAccountID, input.Limit, input.Offset)
	if listErr != nil {
		return out, listErr
	}

	outItems := make([]ucModels.WebPrompt, 0, len(items))
	for _, p := range items {
		outItems = append(outItems, ucModels.WebPrompt{
			ID:       p.ID,
			Prompt:   p.Prompt,
			Category: p.Category,
			Model:    p.Model,
		})
	}
	return ucModels.ListWebPromptsOutput{Items: outItems}, nil
}


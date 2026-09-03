package usecase

import (
	"github.com/twelvepills-936/tgapp-/internal"
	"github.com/twelvepills-936/tgapp-/pkg/config"
	"github.com/twelvepills-936/tgapp-/pkg/yookassa"
)

// useCase implements internal.UseCase.
type useCase struct {
	repo internal.Repository
	jwt  config.ConfigJWT

	yookassa          *yookassa.Client
	yookassaReturnURL string

	telegramBotToken string
}

// Option configures optional dependencies on the usecase layer.
type Option func(*useCase)

// WithYooKassa wires the YooKassa payments client used by checkout/webhook flows.
// returnURL is where YooKassa sends the buyer back to after paying.
func WithYooKassa(client *yookassa.Client, returnURL string) Option {
	return func(uc *useCase) {
		uc.yookassa = client
		uc.yookassaReturnURL = returnURL
	}
}

// WithTelegramBotToken wires the bot token used to verify the HMAC signature
// Telegram attaches to WebApp init data (see telegramauth.VerifyInitData).
// Without it, RegisterByTelegram cannot confirm a registration request
// actually came from Telegram rather than a forged payload.
func WithTelegramBotToken(token string) Option {
	return func(uc *useCase) {
		uc.telegramBotToken = token
	}
}

// NewUseCase wires repository layer into business logic.
func NewUseCase(
	repo internal.Repository,
	jwtCfg config.ConfigJWT,
	opts ...Option,
) internal.UseCase {
	uc := &useCase{
		repo: repo,
		jwt:  jwtCfg,
	}
	for _, opt := range opts {
		opt(uc)
	}
	return uc
}

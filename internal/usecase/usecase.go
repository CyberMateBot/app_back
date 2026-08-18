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

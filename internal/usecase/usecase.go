package usecase

import (
	"github.com/twelvepills-936/tgapp-/internal"
	"github.com/twelvepills-936/tgapp-/pkg/config"
)

// useCase implements internal.UseCase.
type useCase struct {
	repo internal.Repository
	jwt  config.ConfigJWT
}

// NewUseCase wires repository layer into business logic.
func NewUseCase(
	repo internal.Repository,
	jwtCfg config.ConfigJWT,
) internal.UseCase {
	return &useCase{
		repo: repo,
		jwt:  jwtCfg,
	}
}

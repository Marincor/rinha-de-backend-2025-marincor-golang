package healthcheck

import (
	"time"

	"github.com/marincor/rinha-de-backend-2025-marincor-golang/internal/application/dtos"
)

// var errOutOfSync = errors.New("out of sync")

type UseCase struct{}

func NewUseCase() *UseCase {
	return &UseCase{}
}

func (usecase *UseCase) Execute() (*dtos.Health, error) {
	now := time.Now()

	return &dtos.Health{
		Sync: &now,
	}, nil
}

package main

import (
	healthcheck "github.com/marincor/rinha-de-backend-2025-marincor-golang/internal/application/usecases/health_check"
	healthcontroller "github.com/marincor/rinha-de-backend-2025-marincor-golang/internal/presentation/controllers/health"
)

func makeHealthController() *healthcontroller.Controller {
	healthUseCase := healthcheck.NewUseCase()

	return healthcontroller.NewController(healthUseCase)
}

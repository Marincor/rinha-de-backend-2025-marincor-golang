package main

import (
	"github.com/marincor/rinha-de-backend-2025-marincor-golang/config"
	"github.com/marincor/rinha-de-backend-2025-marincor-golang/constants"
	healthcheck "github.com/marincor/rinha-de-backend-2025-marincor-golang/internal/application/usecases/health_check"
	processpayment "github.com/marincor/rinha-de-backend-2025-marincor-golang/internal/application/usecases/process_payment"
	retrievepaymentsummary "github.com/marincor/rinha-de-backend-2025-marincor-golang/internal/application/usecases/retrieve_payment_summary"
	"github.com/marincor/rinha-de-backend-2025-marincor-golang/internal/domain/contracts"
	"github.com/marincor/rinha-de-backend-2025-marincor-golang/internal/domain/entities"
	circuitbreaker "github.com/marincor/rinha-de-backend-2025-marincor-golang/internal/infra/circuit_breaker"
	paymentprocessor "github.com/marincor/rinha-de-backend-2025-marincor-golang/internal/infra/clients/payment_processor"
	"github.com/marincor/rinha-de-backend-2025-marincor-golang/internal/infra/clients/redis"
	healthcontroller "github.com/marincor/rinha-de-backend-2025-marincor-golang/internal/presentation/controllers/health"
	paymentcontroller "github.com/marincor/rinha-de-backend-2025-marincor-golang/internal/presentation/controllers/payment"
)

func makeHealthController() *healthcontroller.Controller {
	healthUseCase := healthcheck.NewUseCase()

	return healthcontroller.NewController(healthUseCase)
}

func makePaymentController(config *config.Config, workerPool contracts.WorkerPoolManager) *paymentcontroller.Controller {
	defaultPaymentProcessor := paymentprocessor.New(config.PaymentProcessorDefault, entities.Default)
	secondaryPaymentProcessor := paymentprocessor.New(config.PaymentProcessorFallback, entities.Fallback)

	paymentCircuitBreaker := circuitbreaker.New[*entities.PaymentResponse](
		constants.MaxAttemptsBeforeOpen, constants.RecoveryTimeout,
	)

	// paymentStorage := hazelcast.New("payments")

	paymentStorage := redis.New()

	paymentUseCase := processpayment.NewUseCase(
		defaultPaymentProcessor, secondaryPaymentProcessor,
		paymentCircuitBreaker,
		paymentStorage,
	)

	paymentSummaryUseCase := retrievepaymentsummary.NewUseCase(
		defaultPaymentProcessor, secondaryPaymentProcessor,
		paymentStorage,
	)

	return paymentcontroller.NewController(
		paymentUseCase,
		paymentSummaryUseCase,
		workerPool,
	)
}

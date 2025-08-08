package processpayment

import (
	"time"

	"github.com/marincor/rinha-de-backend-2025-marincor-golang/constants"
	"github.com/marincor/rinha-de-backend-2025-marincor-golang/helpers"
	"github.com/marincor/rinha-de-backend-2025-marincor-golang/internal/application/dtos"
	"github.com/marincor/rinha-de-backend-2025-marincor-golang/internal/domain/contracts"
	"github.com/marincor/rinha-de-backend-2025-marincor-golang/internal/domain/entities"
)

const (
	maxRetries   = 3
	initialDelay = time.Millisecond
	multiplier   = 2
	randomInt    = 10
)

type UseCase struct {
	defaultPaymentProcessor   contracts.PaymentProcessor
	secondaryPaymentProcessor contracts.PaymentProcessor
	paymentCircuitBreaker     contracts.CircuitBreaker[*entities.PaymentResponse]
	paymentStorage            contracts.Storage
}

func NewUseCase(
	defaultPaymentProcessor contracts.PaymentProcessor,
	secondaryPaymentProcessor contracts.PaymentProcessor,
	paymentCircuitBreaker contracts.CircuitBreaker[*entities.PaymentResponse],
	paymentStorage contracts.Storage,
) *UseCase {
	return &UseCase{
		defaultPaymentProcessor:   defaultPaymentProcessor,
		secondaryPaymentProcessor: secondaryPaymentProcessor,
		paymentCircuitBreaker:     paymentCircuitBreaker,
		paymentStorage:            paymentStorage,
	}
}

func (usecase *UseCase) Execute(paymentRequest *dtos.PaymentPayload) (*entities.PaymentResponse, error) {
	payload := &entities.PaymentRequest{
		CorrelationID: paymentRequest.CorrelationID.String(),
		Amount:        paymentRequest.Amount,
		RequestedAt:   time.Now().UTC().Format(constants.DefaultTimeFormat),
	}

	response, err := helpers.ExponentialBackoffRetry(
		func() (*entities.PaymentResponse, error) {
			return usecase.processPayment(payload)
		},
		maxRetries, initialDelay, multiplier, randomInt,
	)
	if err == nil {
		go func(currentResponse *entities.PaymentResponse, currentPayload *entities.PaymentRequest) {
			processorProvider := entities.Default
			if currentResponse.IsFallback {
				processorProvider = entities.Fallback
			}

			usecase.paymentStorage.Save(&entities.PaymentPayloadStorage{
				ID:                currentPayload.CorrelationID,
				Amount:            currentPayload.Amount,
				RequestedAt:       currentPayload.RequestedAt,
				ProcessorProvider: processorProvider,
			})
		}(response, payload)
	}

	return response, err
}

func (usecase *UseCase) processPayment(payload *entities.PaymentRequest) (*entities.PaymentResponse, error) {
	primaryPayment := func() (*entities.PaymentResponse, error) {
		return usecase.defaultPaymentProcessor.ProcessPayment(payload)
	}

	fallback := func() (*entities.PaymentResponse, error) {
		return usecase.secondaryPaymentProcessor.ProcessPayment(payload)
	}

	return usecase.paymentCircuitBreaker.Execute(primaryPayment, fallback)
}

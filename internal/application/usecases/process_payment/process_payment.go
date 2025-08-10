package processpayment

import (
	"errors"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/marincor/rinha-de-backend-2025-marincor-golang/constants"
	"github.com/marincor/rinha-de-backend-2025-marincor-golang/helpers"
	"github.com/marincor/rinha-de-backend-2025-marincor-golang/internal/application/dtos"
	"github.com/marincor/rinha-de-backend-2025-marincor-golang/internal/domain/contracts"
	"github.com/marincor/rinha-de-backend-2025-marincor-golang/internal/domain/entities"
)

const (
	maxRetries   = 2
	initialDelay = time.Millisecond
	multiplier   = 2
	randomInt    = 10
)

var (
	paymentRequestPool = sync.Pool{
		New: func() interface{} {
			return &entities.PaymentRequest{}
		},
	}

	timeStringCache = sync.Map{}
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

//nolint:funlen // long but necessary
func (usecase *UseCase) Execute(paymentRequest *dtos.PaymentPayload) (*entities.PaymentResponse, error) {
	payload, ok := paymentRequestPool.Get().(*entities.PaymentRequest)
	if !ok {
		return nil, constants.ErrGettingPaymentRequestFromPool
	}

	defer paymentRequestPool.Put(payload)

	*payload = entities.PaymentRequest{
		CorrelationID: paymentRequest.CorrelationID.String(),
		Amount:        paymentRequest.Amount,
		RequestedAt:   usecase.getTimeString(),
	}

	response, err := helpers.ExponentialBackoffRetry(
		func() (*entities.PaymentResponse, error) {
			internalResponse, err := usecase.processPayment(payload)
			if err != nil {
				alreadyProcessedCode := "422"
				if errors.Is(err, constants.ErrInvalidStatusCode) && strings.Contains(err.Error(), alreadyProcessedCode) {
					if internalResponse.ProcessorProvider == entities.Fallback {
						internalResponse.ProcessorProvider = entities.Default
					} else {
						internalResponse.ProcessorProvider = entities.Fallback
					}

					return internalResponse, nil
				}

				return internalResponse, err
			}

			return internalResponse, err
		},
		maxRetries, initialDelay, multiplier, randomInt,
	)
	if err == nil {
		go func(currentResponse *entities.PaymentResponse, currentPayload *entities.PaymentRequest) {
			log.Print(
				map[string]interface{}{
					"correlation_id": currentPayload.CorrelationID,
					"amount":         currentPayload.Amount,
					"requested_at":   currentPayload.RequestedAt,
					"processor":      currentResponse.ProcessorProvider,
					"action":         "saving",
				},
			)

			if err := usecase.paymentStorage.Save(&entities.PaymentPayloadStorage{
				ID:                currentPayload.CorrelationID,
				Amount:            currentPayload.Amount,
				RequestedAt:       currentPayload.RequestedAt,
				ProcessorProvider: currentResponse.ProcessorProvider,
			}); err != nil {
				log.Print(
					map[string]interface{}{
						"correlation_id": currentPayload.CorrelationID,
						"amount":         currentPayload.Amount,
						"requested_at":   currentPayload.RequestedAt,
						"processor":      currentResponse.ProcessorProvider,
						"action":         "error saving",
						"error":          err,
					},
				)
			}
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

func (usecase *UseCase) getTimeString() string {
	now := time.Now().UTC()

	// Using for second in repeated requests
	key := now.Unix()
	if cached, ok := timeStringCache.Load(key); ok {
		value, cachedOk := cached.(string)
		if cachedOk {
			return value
		}

		return now.Format(constants.DefaultTimeFormat)
	}

	formatted := now.Format(constants.DefaultTimeFormat)
	timeStringCache.Store(key, formatted)

	go usecase.cleanupTimeCache(key)

	return formatted
}

func (usecase *UseCase) cleanupTimeCache(currentKey int64) {
	timeToWait := 2

	time.Sleep(time.Duration(timeToWait) * time.Second)

	timeStringCache.Range(func(key, _ interface{}) bool {
		value, ok := key.(int64)
		if ok && value < currentKey-1 {
			timeStringCache.Delete(key)
		}

		return true
	})
}

package retrievepaymentsummary

import (
	"time"

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
	paymentStorage            contracts.Storage
}

func NewUseCase(
	defaultPaymentProcessor contracts.PaymentProcessor,
	secondaryPaymentProcessor contracts.PaymentProcessor,
	paymentStorage contracts.Storage,
) *UseCase {
	return &UseCase{
		defaultPaymentProcessor:   defaultPaymentProcessor,
		secondaryPaymentProcessor: secondaryPaymentProcessor,
		paymentStorage:            paymentStorage,
	}
}

func (usecase *UseCase) Execute(paymentSummaryFilter *dtos.PaymentSummaryFilters) (*entities.PaymentResultStorage, error) {
	return usecase.paymentStorage.Retrieve(&entities.PaymentSummaryFilters{
		From: paymentSummaryFilter.From,
		To:   paymentSummaryFilter.To,
	})
}

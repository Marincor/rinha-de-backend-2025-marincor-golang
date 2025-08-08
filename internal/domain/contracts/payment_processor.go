package contracts

import "github.com/marincor/rinha-de-backend-2025-marincor-golang/internal/domain/entities"

type PaymentProcessor interface {
	ProcessPayment(paymentRequest *entities.PaymentRequest) (*entities.PaymentResponse, error)
	PaymentsSummary(filters *entities.PaymentSummaryFilters) (*entities.PaymentSummaryResponse, error)
}

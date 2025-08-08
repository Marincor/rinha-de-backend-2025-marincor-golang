package contracts

import "github.com/marincor/rinha-de-backend-2025-marincor-golang/internal/domain/entities"

type Storage interface {
	Save(payload *entities.PaymentPayloadStorage) error
	Retrieve(payloadFilters *entities.PaymentSummaryFilters) (*entities.PaymentResultStorage, error)
}

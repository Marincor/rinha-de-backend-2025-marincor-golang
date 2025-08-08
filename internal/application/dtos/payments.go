package dtos

import (
	"time"

	"github.com/google/uuid"
)

type PaymentPayload struct {
	CorrelationID uuid.UUID `json:"correlationId"`
	Amount        float64   `json:"amount"`
}

type PaymentSummaryFilters struct {
	From *time.Time `query:"from"`
	To   *time.Time `query:"to"`
}

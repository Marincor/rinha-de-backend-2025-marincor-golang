package validators

import (
	"fmt"

	"github.com/marincor/rinha-de-backend-2025-marincor-golang/internal/application/dtos"
)

type Validator struct {
}

func New() *Validator {
	return &Validator{}
}

func (v *Validator) ValidatePaymentPayload(payload *dtos.PaymentPayload) error {
	correlationID := payload.CorrelationID.String()
	if correlationID == "" || correlationID == "00000000-0000-0000-0000-000000000000" {
		return fmt.Errorf("correlation_id is required")
	}

	if payload.Amount <= 0 {
		return fmt.Errorf("amount must be greater than 0")
	}

	return nil
}

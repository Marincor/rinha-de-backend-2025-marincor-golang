package entities

import "time"

type PaymentRequest struct {
	CorrelationID string
	RequestedAt   string
	Amount        float64
}

type PaymentResponse struct {
	Message           string `json:"message"`
	ProcessorProvider ProcessorProvider
}

type PaymentSummaryFilters struct {
	From *time.Time `query:"from"`
	To   *time.Time `query:"to"`
}

type Summary struct {
	TotalRequests int     `json:"totalRequests"`
	TotalAmount   float64 `json:"totalAmount"`
}

type PaymentSummaryResponse struct {
	Default  Summary `json:"default"`
	Fallback Summary `json:"fallback"`
}

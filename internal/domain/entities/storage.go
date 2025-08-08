package entities

type ProcessorProvider string

const (
	Default  ProcessorProvider = "default"
	Fallback ProcessorProvider = "fallback"
)

type PaymentPayloadStorage struct {
	ID                string
	ProcessorProvider ProcessorProvider
	Amount            float64
	RequestedAt       string
}

type PaymentResultStorage struct {
	PaymentSummaryResponse
}

type PaymentSummaryStorageFilters struct {
	PaymentSummaryFilters
}

package entities

type ProcessorProvider string

const (
	Default  ProcessorProvider = "default"
	Fallback ProcessorProvider = "fallback"
)

type PaymentPayloadStorage struct {
	ID                string
	ProcessorProvider ProcessorProvider
	RequestedAt       string
	Amount            float64
}

type PaymentResultStorage struct {
	PaymentSummaryResponse
}

type PaymentSummaryStorageFilters struct {
	PaymentSummaryFilters
}

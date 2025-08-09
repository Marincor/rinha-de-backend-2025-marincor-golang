package hazelcast

type PaymentEntry struct {
	ID          string  `json:"id"`
	RequestedAt string  `json:"requested_at"`
	Amount      float64 `json:"amount"`
}

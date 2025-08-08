package hazelcast

type PaymentEntry struct {
	ID          string  `json:"id"`
	Amount      float64 `json:"amount"`
	RequestedAt string  `json:"requested_at"`
}

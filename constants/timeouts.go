package constants

import "time"

const (
	DefaultRequestTimeout = 5 * time.Second
	MaxAttemptsBeforeOpen = 5
	RecoveryTimeout       = 10 * time.Second
)

package constants

import "time"

const (
	DefaultRequestTimeout = 100 * time.Millisecond
	MaxAttemptsBeforeOpen = 2
	RecoveryTimeout       = 10 * time.Second
)

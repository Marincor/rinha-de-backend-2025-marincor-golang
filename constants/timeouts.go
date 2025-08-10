package constants

import "time"

const (
	DefaultRequestTimeout = 30 * time.Second
	MaxAttemptsBeforeOpen = 2
	RecoveryTimeout       = 10 * time.Second
)

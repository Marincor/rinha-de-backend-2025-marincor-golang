package constants

import "time"

const (
	DefaultRequestTimeout = 10 * time.Second
	MaxAttemptsBeforeOpen = 3
	RecoveryTimeout       = 10 * time.Second
)

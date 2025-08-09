package constants

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidStatusCode           = errors.New("invalid status code")
	ErrAmountMustBeGreaterThanZero = errors.New("amount must be greater than 0")
	ErrCorrelationIDIsRequired     = errors.New("correlationId is required")
)

func NewErrorWrapper(err error, message any) error {
	return fmt.Errorf("%w: %v", err, message)
}

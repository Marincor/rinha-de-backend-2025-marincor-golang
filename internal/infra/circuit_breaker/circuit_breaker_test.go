package circuitbreaker_test

import (
	"errors"
	"math/rand"
	"sync"
	"testing"
	"time"

	circuitbreaker "github.com/marincor/rinha-de-backend-2025-marincor-golang/internal/infra/circuit_breaker"
	"github.com/stretchr/testify/assert"
)

var (
	errFallback             = errors.New("fallback")
	errOperation            = errors.New("operation failed")
	errOperationFailedAgain = errors.New("operation failed again")
	errSimulatedError       = errors.New("simulated failure")
	unexpected              = "unexpected error: %v"
)

func TestExecuteSuccess(t *testing.T) {
	t.Parallel()

	cb := circuitbreaker.New[int](3, 1*time.Second)

	result, err := cb.Execute(
		func() (int, error) {
			return 42, nil
		},
		func() (int, error) {
			return 0, errFallback
		},
	)
	if err != nil {
		t.Errorf(unexpected, err)
	}

	if result != 42 {
		t.Errorf("expected 42, got %d", result)
	}
}

func TestCircuitBreakerExecuteFailureFallBack(t *testing.T) {
	t.Parallel()

	cirbuitBreaker := circuitbreaker.New[int](3, 1*time.Second)

	// Simulate operation failure
	_, _ = cirbuitBreaker.Execute(
		func() (int, error) {
			return 0, errOperation
		},
		func() (int, error) {
			return 0, nil
		},
	)

	result, err := cirbuitBreaker.Execute(
		func() (int, error) {
			return 0, errOperationFailedAgain
		},
		func() (int, error) {
			return 99, nil
		},
	)
	if err != nil {
		t.Errorf(unexpected, err)
	}

	if result != 99 {
		t.Errorf("expected 99, got %d", result)
	}
}

func TestCircuitBreakerOpenState(t *testing.T) {
	t.Parallel()

	cirbuitBreaker := circuitbreaker.New[int](1, 1*time.Second)

	// This should open the circuit breaker
	_, _ = cirbuitBreaker.Execute(
		func() (int, error) {
			return 0, errOperation
		},
		func() (int, error) {
			return 0, nil
		},
	)

	_, err := cirbuitBreaker.Execute(
		func() (int, error) {
			t.Fatal("should not call operation when in an open state")

			return 0, nil
		},
		func() (int, error) {
			return 99, nil
		},
	)
	if err != nil {
		t.Errorf(unexpected, err)
	}
}

func TestCircuitBreakerHalfOpenState(t *testing.T) {
	t.Parallel()

	cirbuitBreaker := circuitbreaker.New[int](1, 500*time.Millisecond)

	// Open the circuit breaker
	_, _ = cirbuitBreaker.Execute(
		func() (int, error) {
			return 0, errOperation
		},
		func() (int, error) {
			return 0, nil
		},
	)

	time.Sleep(600 * time.Millisecond) // wait for the timeout to trigger half-open state

	result, err := cirbuitBreaker.Execute(
		func() (int, error) {
			return 42, nil
		},
		func() (int, error) {
			return 0, errFallback
		},
	)
	if err != nil {
		t.Errorf(unexpected, err)
	}

	if result != 42 {
		t.Errorf("expected 42, got %d", result)
	}
}

func TestCircuitBreakerRaceCondition(t *testing.T) {
	t.Parallel()

	failureThreshold := int32(5)
	recoveryTimeout := 100 * time.Millisecond

	circuitBreaker := circuitbreaker.New[int](failureThreshold, recoveryTimeout)

	numGoroutines := 100
	var waitGroup sync.WaitGroup

	operation := func() (int, error) {
		//nolint:gosec // this is just a test
		time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)

		return 0, errSimulatedError
	}

	fallback := func() (int, error) {
		//nolint:gosec // this is just a test
		time.Sleep(time.Duration(rand.Intn(40)) * time.Millisecond)

		return -1, nil
	}

	for range numGoroutines {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			if _, err := circuitBreaker.Execute(operation, fallback); err != nil {
				t.Errorf(unexpected, err)
			}
		}()
	}

	waitGroup.Wait()

	// Check the final state of the circuit breaker
	state := circuitBreaker.GetState()
	failureCount := circuitBreaker.GetCountFailure()

	// Asserts
	assert.Equal(t, circuitbreaker.Open, state, "The state should be Open")
	assert.GreaterOrEqual(t, failureCount, failureThreshold, "The failure count should be greater or equal to the failure threshold")
}

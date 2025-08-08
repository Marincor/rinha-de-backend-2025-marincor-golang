package circuitbreaker

import (
	"log"
	"reflect"
	"sync/atomic"
	"time"
)

const (
	Closed int32 = iota
	Open
	HalfOpen
)

type CircuitBreaker[T any] struct {
	lastFailureTime  atomic.Value
	typeName         string
	state            atomic.Int32
	failureCount     atomic.Int32
	failureThreshold int32
	recoveryTimeout  time.Duration
}

func getTypeName[T any](t T) string {
	return reflect.TypeOf(t).String()
}

func New[T any](failureThreshold int32, recoveryTimeout time.Duration) *CircuitBreaker[T] {
	var tName T

	circuitBreaker := &CircuitBreaker[T]{
		state:            atomic.Int32{},
		failureThreshold: failureThreshold,
		recoveryTimeout:  recoveryTimeout,
		typeName:         getTypeName[T](tName),
	}

	circuitBreaker.state.Store(Closed)

	return circuitBreaker
}

func (cb *CircuitBreaker[T]) Execute(
	operation func() (T, error),
	fallback func() (T, error),
) (T, error) {
	if cb.state.Load() == Open {
		lastFailureTime, ok := cb.lastFailureTime.Load().(time.Time)
		if ok && time.Since(lastFailureTime) > cb.recoveryTimeout {
			cb.state.Store(HalfOpen)
		} else {
			return fallback()
		}
	}

	result, err := operation()
	if err != nil {
		cb.handleFailure()

		return fallback()
	}

	cb.reset()

	return result, nil
}

func (cb *CircuitBreaker[T]) handleFailure() {
	currentFailures := cb.failureCount.Add(1)
	if currentFailures >= cb.failureThreshold || cb.state.Load() == HalfOpen {
		log.Printf("Circuit breaker %s failure count reached %d and circuit is now open", cb.typeName, cb.failureCount.Load())

		cb.state.Store(Open)
		cb.lastFailureTime.Store(time.Now())
	}
}

func (cb *CircuitBreaker[T]) GetState() int32 {
	return cb.state.Load()
}

func (cb *CircuitBreaker[T]) GetCountFailure() int32 {
	return cb.failureCount.Load()
}

func (cb *CircuitBreaker[T]) reset() {
	cb.failureCount.Store(0)
	cb.state.Store(Closed)
	cb.lastFailureTime.Store(time.Time{})
}

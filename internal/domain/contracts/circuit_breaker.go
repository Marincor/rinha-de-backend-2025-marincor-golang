package contracts

type CircuitBreaker[T any] interface {
	Execute(operation func() (T, error), fallback func() (T, error)) (T, error)
	GetState() int32
	GetCountFailure() int32
}

package helpers

import (
	"crypto/rand"
	"time"
)

type CallbackFunc[T any] func() (T, error)

func ExponentialBackoffRetry[T any](callback CallbackFunc[T], maxRetries int, initialDelay time.Duration, multiplier int, randomInt int) (T, error) {
	var attempt int

	for {
		result, err := callback()
		if err == nil {
			return result, nil
		}

		if attempt >= maxRetries-1 {
			var zero T

			return zero, err
		}

		delay := initialDelay * time.Duration(multiplier<<attempt) // 2^attempt

		jitter := generateJitter(randomInt)

		totalDelay := delay + jitter

		time.Sleep(totalDelay)

		attempt++
	}
}

// random variation expected between attempts.
func generateJitter(randomInt int) time.Duration {
	necessaryAmountOfBytes := 1
	randomValue := make([]byte, necessaryAmountOfBytes)
	randomByte := randomInt

	if _, err := rand.Read(randomValue); err == nil {
		randomByte = int(randomValue[0])
	}

	// limit bytes between 0 and randomInt -1 (because of % operator)
	jitter := time.Duration(randomByte%randomInt) * time.Second

	return jitter
}

//nolint:all // only test
package helpers

import (
	"errors"
	"testing"
	"time"
)

func TestExponentialBackoffRetrySuccessImmediately(t *testing.T) {
	callCount := 0
	callback := func() (string, error) {
		callCount++
		return "success", nil
	}

	result, err := ExponentialBackoffRetry(callback, 3, 10*time.Millisecond, 2, 1)
	if err != nil {
		t.Errorf("esperava sucesso, mas retornou erro: %v", err)
	}

	if result != "success" {
		t.Errorf("resultado inesperado: %s", result)
	}

	if callCount != 1 {
		t.Errorf("esperado 1 chamada, obtido %d", callCount)
	}
}

func TestExponentialBackoffRetrySuccessAfterRetries(t *testing.T) {
	callCount := 0
	callback := func() (string, error) {
		callCount++
		if callCount < 3 {
			return "", errors.ErrUnsupported
		}
		return "ok", nil
	}

	result, err := ExponentialBackoffRetry(callback, 5, 5*time.Millisecond, 2, 1)
	if err != nil {
		t.Errorf("esperado sucesso, obteve erro: %v", err)
	}

	if result != "ok" {
		t.Errorf("esperado 'ok', obteve '%s'", result)
	}

	if callCount != 3 {
		t.Errorf("esperado 3 tentativas, obteve %d", callCount)
	}
}

func TestExponentialBackoffRetryFailureAfterMaxRetries(t *testing.T) {
	callback := func() (int, error) {
		return 0, errors.ErrUnsupported
	}

	start := time.Now()
	_, err := ExponentialBackoffRetry(callback, 3, 5*time.Millisecond, 2, 1)
	elapsed := time.Since(start)

	if err == nil {
		t.Error("esperado erro, mas foi nil")
	}

	if elapsed < 10*time.Millisecond {
		t.Error("parece que não esperou entre tentativas")
	}
}

func TestGenerateJitterReturnsWithinExpectedRange(t *testing.T) {
	maxNumber := 10
	//nolint:intrange // false positive
	for i := 0; i < 100; i++ {
		j := generateJitter(maxNumber)
		if j < 0 || j > time.Duration(maxNumber-1)*time.Second {
			t.Errorf("jitter fora do intervalo esperado: %s", j)
		}
	}
}

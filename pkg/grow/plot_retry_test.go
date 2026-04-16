package grow_test

import (
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Mr-xiaotian/CelestialForge/pkg/grow"
)

// 重试后成功
func TestPlot_RetrySuccess(t *testing.T) {
	var attempts atomic.Int32
	cultivator := func(task int) (int, error) {
		n := attempts.Add(1)
		if n <= 2 {
			return 0, errors.New("transient error")
		}
		return task * 10, nil
	}

	plot := grow.NewPlot("test_retry_success", cultivator, nil,
		grow.WithTends(1),
		grow.WithMaxRetries(3),
	)
	results := plot.Start([]int{1})

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Fruit != 10 {
		t.Errorf("expected fruit 10, got %d", results[0].Fruit)
	}
	if attempts.Load() != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts.Load())
	}
}

// 重试耗尽仍失败
func TestPlot_RetryExhausted(t *testing.T) {
	var attempts atomic.Int32
	cultivator := func(task int) (int, error) {
		attempts.Add(1)
		return 0, errors.New("permanent error")
	}

	plot := grow.NewPlot("test_retry_exhausted", cultivator, nil,
		grow.WithTends(1),
		grow.WithMaxRetries(2),
	)
	results := plot.Start([]int{1})

	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
	if attempts.Load() != 3 {
		t.Errorf("expected 3 attempts (1 + 2 retries), got %d", attempts.Load())
	}
}

// retryIf 过滤不可重试错误
func TestPlot_RetryIf(t *testing.T) {
	var attempts atomic.Int32
	permanent := errors.New("permanent")
	cultivator := func(task int) (int, error) {
		attempts.Add(1)
		return 0, permanent
	}

	plot := grow.NewPlot("test_retry_if", cultivator, nil,
		grow.WithTends(1),
		grow.WithMaxRetries(3),
		grow.WithRetryIf(func(err error) bool {
			return !errors.Is(err, permanent)
		}),
	)
	results := plot.Start([]int{1})

	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
	if attempts.Load() != 1 {
		t.Errorf("expected 1 attempt (no retry for permanent error), got %d", attempts.Load())
	}
}

// retryDelay 验证间隔被调用
func TestPlot_RetryDelay(t *testing.T) {
	var attempts atomic.Int32
	cultivator := func(task int) (int, error) {
		n := attempts.Add(1)
		if n <= 1 {
			return 0, errors.New("transient")
		}
		return task, nil
	}

	start := time.Now()
	plot := grow.NewPlot("test_retry_delay", cultivator, nil,
		grow.WithTends(1),
		grow.WithMaxRetries(2),
		grow.WithRetryDelay(func(attempt int) time.Duration {
			return 100 * time.Millisecond
		}),
	)
	results := plot.Start([]int{1})
	elapsed := time.Since(start)

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if elapsed < 100*time.Millisecond {
		t.Errorf("expected at least 100ms delay, got %v", elapsed)
	}
}

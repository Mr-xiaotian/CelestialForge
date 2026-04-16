package tests

import (
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Mr-xiaotian/CelestialForge/pkg/grow"
)

// 全部失败
func TestPlot_AllError(t *testing.T) {
	cultivator := func(task int) (string, error) {
		return "", errors.New("always fail")
	}

	plot := grow.NewPlot("test_all_error", cultivator, nil, grow.WithTends(2))
	tasks := []int{1, 2, 3, 4, 5}

	results := plot.Start(tasks)
	for _, res := range results {
		t.Errorf("unexpected success: %v", res)
	}

	if plot.GetCompleted() != len(tasks) {
		t.Errorf("expected %d completed, got %d", len(tasks), plot.GetCompleted())
	}
	if int(plot.State()) != 2 {
		t.Errorf("expected state 2 (done), got %d", plot.State())
	}
}

// 部分失败
func TestPlot_PartialError(t *testing.T) {
	cultivator := func(task int) (int, error) {
		if task%2 == 0 {
			return 0, errors.New("even number error")
		}
		return task * 10, nil
	}

	plot := grow.NewPlot("test_partial_error", cultivator, nil, grow.WithTends(2))
	tasks := []int{1, 2, 3, 4, 5}

	results := plot.Start(tasks)
	successCount := 0
	for _, res := range results {
		if res.Fruit != res.Seed*10 {
			t.Fatalf("result %d does not match source %d", res.Fruit, res.Seed)
		}
		successCount++
	}

	if successCount != 3 {
		t.Errorf("expected 3 successes, got %d", successCount)
	}
	if plot.GetCompleted() != len(tasks) {
		t.Errorf("expected %d completed, got %d", len(tasks), plot.GetCompleted())
	}
}

// 全部成功
func TestPlot_AllSuccess(t *testing.T) {
	cultivator := func(task int) (int, error) {
		return task * 2, nil
	}

	plot := grow.NewPlot("test_all_success", cultivator, nil, grow.WithTends(3))
	tasks := []int{1, 2, 3, 4, 5}

	collects := plot.Start(tasks)
	results := map[int]int{}
	for _, res := range collects {
		results[res.Seed] = res.Fruit
		if res.Fruit != res.Seed*2 {
			t.Fatalf("result %d does not match source %d", res.Fruit, res.Seed)
		}
	}

	for _, task := range tasks {
		if results[task] != task*2 {
			t.Errorf("task %d: expected %d, got %d", task, task*2, results[task])
		}
	}
	if int(plot.State()) != 2 {
		t.Errorf("expected state 2 (done), got %d", plot.State())
	}
}

func TestPlot_Async(t *testing.T) {
	cultivator := func(task int) (int, error) {
		return task * 2, nil
	}

	plot := grow.NewPlot("test_async", cultivator, nil, grow.WithTends(3))
	results := map[int]int{}

	go plot.StartAsync()

	for task := range 5 {
		plot.Seed(task, task)
	}
	plot.Seal()

	plot.Harvest(func(res grow.Payload[int]) {
		results[res.Prev.(int)] = res.Value
	})

	plot.WaitAsync()

	if len(results) != 5 {
		t.Errorf("expected 5 results, got %d", len(results))
	}
	if int(plot.State()) != 2 {
		t.Errorf("expected state 2 (done), got %d", plot.State())
	}
}

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

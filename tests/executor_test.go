package tests

import (
	"errors"
	"testing"

	"github.com/Mr-xiaotian/CelestialForge/pkg/grow"
)

// 全部失败
func TestExecutor_AllError(t *testing.T) {
	processor := func(task int) (string, error) {
		return "", errors.New("always fail")
	}

	executor := grow.NewExecutor("test_all_error", processor, 2)
	tasks := []int{1, 2, 3, 4, 5}

	results := executor.Start(tasks)
	for _, res := range results {
		t.Errorf("unexpected success: %v", res)
	}

	if executor.GetCompleted() != len(tasks) {
		t.Errorf("expected %d completed, got %d", len(tasks), executor.GetCompleted())
	}
	if int(executor.State()) != 2 {
		t.Errorf("expected state 2 (done), got %d", executor.State())
	}
}

// 部分失败
func TestExecutor_PartialError(t *testing.T) {
	processor := func(task int) (int, error) {
		if task%2 == 0 {
			return 0, errors.New("even number error")
		}
		return task * 10, nil
	}

	executor := grow.NewExecutor("test_partial_error", processor, 2)
	tasks := []int{1, 2, 3, 4, 5}

	results := executor.Start(tasks)
	successCount := 0
	for _, res := range results {
		if res.Result != res.Task*10 {
			t.Fatalf("result %d does not match source %d", res.Result, res.Task)
		}
		successCount++
	}

	if successCount != 3 {
		t.Errorf("expected 3 successes, got %d", successCount)
	}
	if executor.GetCompleted() != len(tasks) {
		t.Errorf("expected %d completed, got %d", len(tasks), executor.GetCompleted())
	}
}

// 全部成功
func TestExecutor_AllSuccess(t *testing.T) {
	processor := func(task int) (int, error) {
		return task * 2, nil
	}

	executor := grow.NewExecutor("test_all_success", processor, 3)
	tasks := []int{1, 2, 3, 4, 5}

	collects := executor.Start(tasks)
	results := map[int]int{}
	for _, res := range collects {
		results[res.Task] = res.Result
		if res.Result != res.Task*2 {
			t.Fatalf("result %d does not match source %d", res.Result, res.Task)
		}
	}

	for _, task := range tasks {
		if results[task] != task*2 {
			t.Errorf("task %d: expected %d, got %d", task, task*2, results[task])
		}
	}
	if int(executor.State()) != 2 {
		t.Errorf("expected state 2 (done), got %d", executor.State())
	}
}

func TestExecutor_Async(t *testing.T) {
	processor := func(task int) (int, error) {
		return task * 2, nil
	}

	executor := grow.NewExecutor("test_async", processor, 3)
	executor.SetTotal(5)

	go executor.StartAsync()

	for task := range 5 {
		executor.Seed(task, task)
	}
	executor.ControlChan <- grow.ControlSignal{Source: "test"}

	results := map[int]int{}
	executor.Collect(func(res grow.Payload[int]) {
		results[res.Prev.(int)] = res.Value
	})

	executor.WaitAsync()

	if len(results) != 5 {
		t.Errorf("expected 5 results, got %d", len(results))
	}
	if int(executor.State()) != 2 {
		t.Errorf("expected state 2 (done), got %d", executor.State())
	}
}

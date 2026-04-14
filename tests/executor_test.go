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

	go executor.Start(tasks)
	executor.Drain(func(res grow.Payload[string]) {
		t.Errorf("unexpected success: %v", res)
	})

	if executor.GetComplated() != len(tasks) {
		t.Errorf("expected %d completed, got %d", len(tasks), executor.GetComplated())
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

	go executor.Start(tasks)

	successCount := 0
	executor.Drain(func(res grow.Payload[int]) {
		successCount++
	})

	if successCount != 3 {
		t.Errorf("expected 3 successes, got %d", successCount)
	}
	if executor.GetComplated() != len(tasks) {
		t.Errorf("expected %d completed, got %d", len(tasks), executor.GetComplated())
	}
}

// 全部成功
func TestExecutor_AllSuccess(t *testing.T) {
	processor := func(task int) (int, error) {
		return task * 2, nil
	}

	executor := grow.NewExecutor("test_all_success", processor, 3)
	tasks := []int{1, 2, 3, 4, 5}

	go executor.Start(tasks)

	results := map[int]int{}
	executor.Drain(func(res grow.Payload[int]) {
		results[res.ID] = res.Value
	})

	for i, task := range tasks {
		if results[i] != task*2 {
			t.Errorf("task %d: expected %d, got %d", i, task*2, results[i])
		}
	}
	if int(executor.State()) != 2 {
		t.Errorf("expected state 2 (done), got %d", executor.State())
	}
}

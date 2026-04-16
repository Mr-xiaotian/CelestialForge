package grow_test

import (
	"testing"

	"github.com/Mr-xiaotian/CelestialForge/pkg/grow"
)

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

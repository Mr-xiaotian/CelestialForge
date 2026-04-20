package grow_test

import (
	"testing"

	"github.com/Mr-xiaotian/CelestialForge/pkg/grow"
)

func TestPlot_Async(t *testing.T) {
	cultivator := func(seed int) (int, error) {
		return seed * 2, nil
	}

	plot := grow.NewPlot("test_async", cultivator, nil, grow.WithTends(3))
	plot.InitLocalEnv()
	plot.StartSpouts()
	fruits := map[int]int{}

	plot.StartAsync()
	for seed := range 5 {
		plot.Seed(seed)
	}
	plot.Seal()
	plot.Harvest(func(res grow.Payload[int]) {
		fruits[res.Prev.(int)] = res.Value
	}, 0)
	plot.WaitAsync()

	plot.StopSpouts()

	if len(fruits) != 5 {
		t.Errorf("expected 5 fruits, got %d", len(fruits))
	}
	if int(plot.GetState()) != 2 {
		t.Errorf("expected state 2 (done), got %d", plot.GetState())
	}
}

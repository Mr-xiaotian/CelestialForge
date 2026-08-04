package grow_test

import (
	"testing"

	"github.com/Mr-xiaotian/CelestialForge/pkg/grow"
)

func TestPlot_Async(t *testing.T) {
	cultivator := func(seed int) (int, error) {
		return seed * 2, nil
	}

	plot := grow.NewPlot("test_async", cultivator, grow.WithTends(3))
	plot.InitLocalEnv()
	plot.StartSpouts()
	fruits := map[int]int{}

	plot.StartAsync()
	// 优先启动Harvest, 避免seed数量过大导致seed与fruit chan全部阻塞
	plot.Harvest(func(res grow.Payload[int]) {
		fruits[res.Prev.(int)] = res.Value
	}, 0)
	for seed := range 50 {
		plot.Seed(seed)
	}
	plot.Seal()
	plot.WaitAsync()

	plot.StopSpouts()

	if len(fruits) != 50 {
		t.Errorf("expected 50 fruits, got %d", len(fruits))
	}
	if int(plot.GetState()) != 2 {
		t.Errorf("expected state 2 (done), got %d", plot.GetState())
	}
}

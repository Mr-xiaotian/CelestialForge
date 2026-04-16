package grow_test

import (
	"errors"
	"testing"

	"github.com/Mr-xiaotian/CelestialForge/pkg/grow"
)

// 全部失败
func TestPlot_AllError(t *testing.T) {
	cultivator := func(seed int) (string, error) {
		return "", errors.New("always fail")
	}

	plot := grow.NewPlot("test_all_error", cultivator, nil, grow.WithTends(2))
	seeds := []int{1, 2, 3, 4, 5}

	karmas := plot.Start(seeds)
	for _, res := range karmas {
		t.Errorf("unexpected success: %v", res)
	}

	if plot.GetCompleted() != len(seeds) {
		t.Errorf("expected %d completed, got %d", len(seeds), plot.GetCompleted())
	}
	if int(plot.State()) != 2 {
		t.Errorf("expected state 2 (done), got %d", plot.State())
	}
}

// 部分失败
func TestPlot_PartialError(t *testing.T) {
	cultivator := func(seed int) (int, error) {
		if seed%2 == 0 {
			return 0, errors.New("even number error")
		}
		return seed * 10, nil
	}

	plot := grow.NewPlot("test_partial_error", cultivator, nil, grow.WithTends(2))
	seeds := []int{1, 2, 3, 4, 5}

	karmas := plot.Start(seeds)
	successCount := 0
	for _, res := range karmas {
		if res.Fruit != res.Seed*10 {
			t.Fatalf("fruit %d does not match source %d", res.Fruit, res.Seed)
		}
		successCount++
	}

	if successCount != 3 {
		t.Errorf("expected 3 successes, got %d", successCount)
	}
	if plot.GetCompleted() != len(seeds) {
		t.Errorf("expected %d completed, got %d", len(seeds), plot.GetCompleted())
	}
}

// 全部成功
func TestPlot_AllSuccess(t *testing.T) {
	cultivator := func(seed int) (int, error) {
		return seed * 2, nil
	}

	plot := grow.NewPlot("test_all_success", cultivator, nil, grow.WithTends(3))
	seeds := []int{1, 2, 3, 4, 5}

	collects := plot.Start(seeds)
	karmas := map[int]int{}
	for _, res := range collects {
		karmas[res.Seed] = res.Fruit
		if res.Fruit != res.Seed*2 {
			t.Fatalf("fruit %d does not match source %d", res.Fruit, res.Seed)
		}
	}

	for _, seed := range seeds {
		if karmas[seed] != seed*2 {
			t.Errorf("seed %d: expected %d, got %d", seed, seed*2, karmas[seed])
		}
	}
	if int(plot.State()) != 2 {
		t.Errorf("expected state 2 (done), got %d", plot.State())
	}
}

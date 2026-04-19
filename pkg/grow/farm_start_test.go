package grow_test

import (
	"sort"
	"sync"
	"testing"

	"github.com/Mr-xiaotian/CelestialForge/pkg/grow"
)

func TestFarmStartLinear(t *testing.T) {
	root := grow.NewPlot("root", func(seed int) (int, error) { return seed * 2, nil }, nil, grow.WithTends(2))

	var (
		mu      sync.Mutex
		results []int
	)
	head := grow.NewPlot("head", func(seed int) (int, error) {
		mu.Lock()
		results = append(results, seed)
		mu.Unlock()
		return seed, nil
	}, nil, grow.WithTends(2))

	farm := grow.NewFarm("start_linear", "INFO")
	if err := farm.AddPlot(root, head); err != nil {
		t.Fatalf("AddPlot() error = %v", err)
	}
	if err := farm.Connect([]grow.PlotNode{root}, []grow.PlotNode{head}); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}

	if err := farm.Start(map[string][]any{
		"root": {1, 2, 3},
	}); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	sort.Ints(results)
	want := []int{2, 4, 6}
	if len(results) != len(want) {
		t.Fatalf("len(results) = %d, want %d", len(results), len(want))
	}
	for i := range want {
		if results[i] != want[i] {
			t.Fatalf("results[%d] = %d, want %d", i, results[i], want[i])
		}
	}
	if int(root.GetState()) != 2 {
		t.Fatalf("root state = %d, want 2", root.GetState())
	}
	if int(head.GetState()) != 2 {
		t.Fatalf("head state = %d, want 2", head.GetState())
	}
}

func TestFarmStartRejectNonRootInput(t *testing.T) {
	root := grow.NewPlot("root", func(seed int) (int, error) { return seed, nil }, nil)
	head := grow.NewPlot("head", func(seed int) (int, error) { return seed, nil }, nil)

	farm := grow.NewFarm("start_reject_non_root_input", "INFO")
	if err := farm.AddPlot(root, head); err != nil {
		t.Fatalf("AddPlot() error = %v", err)
	}
	if err := farm.Connect([]grow.PlotNode{root}, []grow.PlotNode{head}); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}

	if err := farm.Start(map[string][]any{
		"head": {1},
	}); err == nil {
		t.Fatal("Start() expected non-root input error, got nil")
	}
}

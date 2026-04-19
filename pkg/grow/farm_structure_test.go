package grow_test

import (
	"sync"
	"testing"

	"github.com/Mr-xiaotian/CelestialForge/pkg/grow"
)

func TestFarmStructure121(t *testing.T) {
	const seedCount = 50

	root := grow.NewPlot("root", func(seed int) (int, error) {
		return seed, nil
	}, nil, grow.WithTends(8))

	midA := grow.NewPlot("midA", func(seed int) (int, error) {
		return seed*10 + 1, nil
	}, nil, grow.WithTends(8))

	midB := grow.NewPlot("midB", func(seed int) (int, error) {
		return seed*10 + 2, nil
	}, nil, grow.WithTends(8))

	var (
		mu     sync.Mutex
		counts = make(map[int]int, seedCount*2)
	)

	head := grow.NewPlot("head", func(seed int) (int, error) {
		mu.Lock()
		counts[seed]++
		mu.Unlock()
		return seed, nil
	}, nil, grow.WithTends(8))

	farm := grow.NewFarm()
	if err := farm.AddPlot(root, midA, midB, head); err != nil {
		t.Fatalf("AddPlot() error = %v", err)
	}

	if err := farm.Connect([]grow.PlotNode{root}, []grow.PlotNode{midA, midB}); err != nil {
		t.Fatalf("Connect(root -> mids) error = %v", err)
	}
	if err := farm.Connect([]grow.PlotNode{midA, midB}, []grow.PlotNode{head}); err != nil {
		t.Fatalf("Connect(mids -> head) error = %v", err)
	}

	inputs := make([]any, 0, seedCount)
	for i := 0; i < seedCount; i++ {
		inputs = append(inputs, i)
	}

	if err := farm.Start(map[string][]any{
		"root": inputs,
	}); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	if !farm.IsRoot("root") {
		t.Fatal("root should remain root")
	}
	if !farm.IsHead("head") {
		t.Fatal("head should remain head")
	}
	if farm.IsHead("root") {
		t.Fatal("root should not be head after connecting downstream")
	}
	if farm.IsRoot("head") {
		t.Fatal("head should not be root after receiving upstream")
	}

	if got := len(counts); got != seedCount*2 {
		t.Fatalf("len(counts) = %d, want %d", got, seedCount*2)
	}

	for i := 0; i < seedCount; i++ {
		a := i*10 + 1
		b := i*10 + 2

		if counts[a] != 1 {
			t.Fatalf("head result %d count = %d, want 1", a, counts[a])
		}
		if counts[b] != 1 {
			t.Fatalf("head result %d count = %d, want 1", b, counts[b])
		}
	}

	if int(root.GetState()) != 2 {
		t.Fatalf("root state = %d, want 2", root.GetState())
	}
	if int(midA.GetState()) != 2 {
		t.Fatalf("midA state = %d, want 2", midA.GetState())
	}
	if int(midB.GetState()) != 2 {
		t.Fatalf("midB state = %d, want 2", midB.GetState())
	}
	if int(head.GetState()) != 2 {
		t.Fatalf("head state = %d, want 2", head.GetState())
	}
}

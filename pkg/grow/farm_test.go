package grow_test

import (
	"testing"

	"github.com/Mr-xiaotian/CelestialForge/pkg/grow"
)

func TestFarmAddPlot(t *testing.T) {
	farm := grow.NewFarm()
	plotA := grow.NewPlot("A", func(seed int) (int, error) { return seed, nil }, nil)
	plotB := grow.NewPlot("B", func(seed int) (int, error) { return seed, nil }, nil)

	if err := farm.AddPlot(plotA, plotB); err != nil {
		t.Fatalf("AddPlot() error = %v", err)
	}

	if farm.PlotCount() != 2 {
		t.Fatalf("farm.PlotCount() = %d, want 2", farm.PlotCount())
	}
	if !farm.HasPlot("A") {
		t.Fatalf("plot A should exist")
	}
	if !farm.HasPlot("B") {
		t.Fatalf("plot B should exist")
	}
	if got, ok := farm.GetPlot("A"); !ok || got != plotA {
		t.Fatalf("GetPlot(A) = %v, %v, want plotA, true", got, ok)
	}
	if got, ok := farm.GetPlot("B"); !ok || got != plotB {
		t.Fatalf("GetPlot(B) = %v, %v, want plotB, true", got, ok)
	}
}

func TestFarmAddPlotDuplicateName(t *testing.T) {
	farm := grow.NewFarm()
	plotA1 := grow.NewPlot("A", func(seed int) (int, error) { return seed, nil }, nil)
	plotA2 := grow.NewPlot("A", func(seed int) (int, error) { return seed, nil }, nil)

	if err := farm.AddPlot(plotA1); err != nil {
		t.Fatalf("AddPlot(first) error = %v", err)
	}
	if err := farm.AddPlot(plotA2); err == nil {
		t.Fatal("AddPlot(duplicate) expected error, got nil")
	}
}

func TestFarmConnectHyperEdge(t *testing.T) {
	farm := grow.NewFarm()
	source := grow.NewPlot("source", func(seed int) (int, error) { return seed * 2, nil }, nil)
	targetA := grow.NewPlot("targetA", func(seed int) (int, error) { return seed, nil }, nil)
	targetB := grow.NewPlot("targetB", func(seed int) (int, error) { return seed, nil }, nil)

	if err := farm.AddPlot(source, targetA, targetB); err != nil {
		t.Fatalf("AddPlot() error = %v", err)
	}

	if err := farm.Connect([]grow.PlotNode{source}, []grow.PlotNode{targetA, targetB, targetA}); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}

	if !farm.Connected("source", "targetA") {
		t.Fatal("source should connect to targetA")
	}
	if !farm.Connected("source", "targetB") {
		t.Fatal("source should connect to targetB")
	}
	if farm.Connected("targetA", "source") {
		t.Fatal("targetA should not connect back to source")
	}
}

func TestFarmConnectTypeMismatch(t *testing.T) {
	farm := grow.NewFarm()
	source := grow.NewPlot("source", func(seed int) (int, error) { return seed, nil }, nil)
	target := grow.NewPlot("target", func(seed string) (string, error) { return seed, nil }, nil)

	if err := farm.AddPlot(source, target); err != nil {
		t.Fatalf("AddPlot() error = %v", err)
	}

	if err := farm.Connect([]grow.PlotNode{source}, []grow.PlotNode{target}); err == nil {
		t.Fatal("Connect() expected type mismatch error, got nil")
	}
}

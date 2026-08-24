package grow_test

import (
	"reflect"
	"testing"

	"github.com/Mr-xiaotian/CelestialForge/pkg/grow"
)

func TestOrderGraph_BasicOperations(t *testing.T) {
	graph := grow.NewOrderGraph()
	graph.AddNode("isolated")
	graph.AddEdge("a", "b")
	graph.AddEdge("a", "c")
	graph.AddEdge("a", "b")

	if !graph.HasNode("isolated") {
		t.Fatalf("expected isolated node to exist")
	}

	wantNodes := []string{"isolated", "a", "b", "c"}
	if !reflect.DeepEqual(graph.Nodes(), wantNodes) {
		t.Fatalf("unexpected node order: got %v want %v", graph.Nodes(), wantNodes)
	}

	wantSuccessors := []string{"b", "c"}
	if !reflect.DeepEqual(graph.Successors("a"), wantSuccessors) {
		t.Fatalf("unexpected successors: got %v want %v", graph.Successors("a"), wantSuccessors)
	}

	wantPredecessors := []string{"a"}
	if !reflect.DeepEqual(graph.Predecessors("b"), wantPredecessors) {
		t.Fatalf("unexpected predecessors: got %v want %v", graph.Predecessors("b"), wantPredecessors)
	}
}

func TestGraphAlgorithms_TopoSortAndLevels(t *testing.T) {
	graph := grow.NewOrderGraphFromEdges(map[string][]string{
		"a": {"b", "c"},
		"b": {"d"},
		"c": {"d"},
	}, nil)

	if !grow.IsDAG(graph) {
		t.Fatalf("expected graph to be a DAG")
	}

	wantTopo := []string{"a", "b", "c", "d"}
	if !reflect.DeepEqual(grow.TopoSort(graph), wantTopo) {
		t.Fatalf("unexpected topo order: got %v want %v", grow.TopoSort(graph), wantTopo)
	}

	wantSources := []string{"a"}
	if !reflect.DeepEqual(grow.SourceNodes(graph), wantSources) {
		t.Fatalf("unexpected source nodes: got %v want %v", grow.SourceNodes(graph), wantSources)
	}

	levels, err := grow.ComputeNodeLevels(graph)
	if err != nil {
		t.Fatalf("compute node levels failed: %v", err)
	}

	wantLevels := map[string]int{
		"a": 0,
		"b": 1,
		"c": 1,
		"d": 2,
	}
	if !reflect.DeepEqual(levels, wantLevels) {
		t.Fatalf("unexpected node levels: got %v want %v", levels, wantLevels)
	}
}

func TestGraphAlgorithms_SCCAndCondensation(t *testing.T) {
	graph := grow.NewOrderGraphFromEdges(map[string][]string{
		"a": {"b"},
		"b": {"a", "c"},
		"c": {"d"},
		"d": {"c"},
	}, nil)

	if grow.IsDAG(graph) {
		t.Fatalf("expected graph to contain cycles")
	}

	sccs := grow.TarjanSCC(graph)
	if len(sccs) != 2 {
		t.Fatalf("unexpected SCC count: got %d want 2", len(sccs))
	}

	wantSCCs := [][]string{{"d", "c"}, {"b", "a"}}
	if !reflect.DeepEqual(sccs, wantSCCs) {
		t.Fatalf("unexpected SCCs: got %v want %v", sccs, wantSCCs)
	}

	sourceSCCs := grow.SourceSCCs(graph)
	wantSourceSCCs := [][]string{{"b", "a"}}
	if !reflect.DeepEqual(sourceSCCs, wantSourceSCCs) {
		t.Fatalf("unexpected source SCCs: got %v want %v", sourceSCCs, wantSourceSCCs)
	}

	condensation, _ := grow.GetCondensation(graph)
	wantCondensationNodes := []string{"scc_0", "scc_1"}
	if !reflect.DeepEqual(condensation.Nodes(), wantCondensationNodes) {
		t.Fatalf("unexpected condensation nodes: got %v want %v", condensation.Nodes(), wantCondensationNodes)
	}

	wantCondensationEdges := map[string][]string{
		"scc_0": {},
		"scc_1": {"scc_0"},
	}
	if !reflect.DeepEqual(condensation.OutEdges(), wantCondensationEdges) {
		t.Fatalf("unexpected condensation edges: got %v want %v", condensation.OutEdges(), wantCondensationEdges)
	}
}

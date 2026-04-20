package grow_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

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

	farm := grow.NewFarm("structure_121", "INFO")
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

func TestFarmStructure121PartialFailure(t *testing.T) {
	const seedCount = 20

	// root: 偶数失败，10 个成功
	root := grow.NewPlot("root", func(seed int) (int, error) {
		if seed%2 == 0 {
			return 0, fmt.Errorf("even seed %d", seed)
		}
		return seed, nil
	}, nil, grow.WithTends(4))

	// midA: 全部成功
	midA := grow.NewPlot("midA", func(seed int) (int, error) {
		return seed*10 + 1, nil
	}, nil, grow.WithTends(4))

	// midB: 能被 3 整除的失败，10 个输入中失败 seed=3,9 共 5 个（seed=1,3,5,7,9,11,13,15,17,19 中 3,9,15 能被3整除）
	// 修正：输入是 1,3,5,7,9,11,13,15,17,19（10个奇数），其中能被3整除的是 3,9,15 共 3 个，成功 7 个
	// 为了让 midB 恰好失败 5 个，改用 seed > 10 失败：失败 11,13,15,17,19 共 5 个，成功 1,3,5,7,9 共 5 个
	midB := grow.NewPlot("midB", func(seed int) (int, error) {
		if seed > 10 {
			return 0, fmt.Errorf("seed %d too large", seed)
		}
		return seed*10 + 2, nil
	}, nil, grow.WithTends(4))

	var (
		mu      sync.Mutex
		results = make(map[int]int)
	)
	head := grow.NewPlot("head", func(seed int) (int, error) {
		mu.Lock()
		results[seed]++
		mu.Unlock()
		return seed, nil
	}, nil, grow.WithTends(4))

	farm := grow.NewFarm("structure_121_partial_failure", "INFO")
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

	// midA 收到 10 个全部成功，midB 收到 10 个成功 5 个，head 共收到 15 个
	if got := len(results); got != 15 {
		t.Fatalf("len(results) = %d, want 15", got)
	}

	for _, p := range []grow.PlotNode{root, midA, midB, head} {
		if int(p.GetState()) != 2 {
			t.Fatalf("%s state = %d, want 2", p.GetName(), p.GetState())
		}
	}
}

func TestFarmStructureDisconnectedComponents(t *testing.T) {
	const seedCount = 50

	// 第一组: 1→2 (rootA → midA1, midA2)
	rootA := grow.NewPlot("rootA", func(seed int) (int, error) {
		return seed*10 + 1, nil
	}, nil, grow.WithTends(4))

	var (
		muA      sync.Mutex
		resultsA = make(map[int]int, seedCount*2)
	)
	midA1 := grow.NewPlot("midA1", func(seed int) (int, error) {
		muA.Lock()
		resultsA[seed]++
		muA.Unlock()
		return seed, nil
	}, nil, grow.WithTends(4))
	midA2 := grow.NewPlot("midA2", func(seed int) (int, error) {
		muA.Lock()
		resultsA[seed]++
		muA.Unlock()
		return seed, nil
	}, nil, grow.WithTends(4))

	// 第二组: 2→1 (rootB1, rootB2 → headB)
	rootB1 := grow.NewPlot("rootB1", func(seed int) (int, error) {
		return seed*10 + 3, nil
	}, nil, grow.WithTends(4))
	rootB2 := grow.NewPlot("rootB2", func(seed int) (int, error) {
		return seed*10 + 4, nil
	}, nil, grow.WithTends(4))

	var (
		muB      sync.Mutex
		resultsB = make(map[int]int, seedCount*2)
	)
	headB := grow.NewPlot("headB", func(seed int) (int, error) {
		muB.Lock()
		resultsB[seed]++
		muB.Unlock()
		return seed, nil
	}, nil, grow.WithTends(4))

	farm := grow.NewFarm("disconnected_components", "INFO")
	if err := farm.AddPlot(rootA, midA1, midA2, rootB1, rootB2, headB); err != nil {
		t.Fatalf("AddPlot() error = %v", err)
	}

	if err := farm.Connect([]grow.PlotNode{rootA}, []grow.PlotNode{midA1, midA2}); err != nil {
		t.Fatalf("Connect(rootA -> mids) error = %v", err)
	}
	if err := farm.Connect([]grow.PlotNode{rootB1, rootB2}, []grow.PlotNode{headB}); err != nil {
		t.Fatalf("Connect(roots -> headB) error = %v", err)
	}

	inputsA := make([]any, 0, seedCount)
	inputsB1 := make([]any, 0, seedCount)
	inputsB2 := make([]any, 0, seedCount)
	for i := 0; i < seedCount; i++ {
		inputsA = append(inputsA, i)
		inputsB1 = append(inputsB1, i)
		inputsB2 = append(inputsB2, i)
	}

	if err := farm.Start(map[string][]any{
		"rootA":  inputsA,
		"rootB1": inputsB1,
		"rootB2": inputsB2,
	}); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	// 验证第一组: rootA fan-out 到 midA1, midA2，各收到 seedCount 个
	if got := len(resultsA); got != seedCount {
		t.Fatalf("len(resultsA) = %d, want %d", got, seedCount)
	}
	for i := 0; i < seedCount; i++ {
		v := i*10 + 1
		if resultsA[v] != 2 {
			t.Fatalf("resultsA[%d] = %d, want 2 (one from each mid)", v, resultsA[v])
		}
	}

	// 验证第二组: rootB1, rootB2 fan-in 到 headB
	if got := len(resultsB); got != seedCount*2 {
		t.Fatalf("len(resultsB) = %d, want %d", got, seedCount*2)
	}
	for i := 0; i < seedCount; i++ {
		b1 := i*10 + 3
		b2 := i*10 + 4
		if resultsB[b1] != 1 {
			t.Fatalf("resultsB[%d] = %d, want 1", b1, resultsB[b1])
		}
		if resultsB[b2] != 1 {
			t.Fatalf("resultsB[%d] = %d, want 1", b2, resultsB[b2])
		}
	}

	// 验证所有 plot 状态
	for _, p := range []grow.PlotNode{rootA, midA1, midA2, rootB1, rootB2, headB} {
		if int(p.GetState()) != 2 {
			t.Fatalf("%s state = %d, want 2", p.GetName(), p.GetState())
		}
	}
}

func TestFarmStructure21FaninDifferentSpeed(t *testing.T) {
	const seedCount = 50

	rootFast := grow.NewPlot("rootFast", func(seed int) (int, error) {
		return seed*10 + 1, nil
	}, nil, grow.WithTends(4), grow.WithChanSize(50), grow.WithLogLevel("SUCCESS"))

	rootSlow := grow.NewPlot("rootSlow", func(seed int) (int, error) {
		time.Sleep(10 * time.Millisecond)
		return seed*10 + 2, nil
	}, nil, grow.WithTends(4), grow.WithChanSize(50), grow.WithLogLevel("SUCCESS"))

	var (
		mu      sync.Mutex
		counts  = make(map[int]int, seedCount*2)
		visited int
	)

	head := grow.NewPlot("head", func(seed int) (int, error) {
		mu.Lock()
		counts[seed]++
		visited++
		mu.Unlock()
		return seed, nil
	}, nil, grow.WithTends(8), grow.WithChanSize(100), grow.WithLogLevel("SUCCESS"))

	farm := grow.NewFarm("structure_21_fanin_different_speed", "INFO")
	if err := farm.AddPlot(rootFast, rootSlow, head); err != nil {
		t.Fatalf("AddPlot() error = %v", err)
	}
	if err := farm.Connect([]grow.PlotNode{rootFast, rootSlow}, []grow.PlotNode{head}); err != nil {
		t.Fatalf("Connect(roots -> head) error = %v", err)
	}

	fastInputs := make([]any, 0, seedCount)
	slowInputs := make([]any, 0, seedCount)
	for i := 0; i < seedCount; i++ {
		fastInputs = append(fastInputs, i)
		slowInputs = append(slowInputs, i)
	}

	if err := farm.Start(map[string][]any{
		"rootFast": fastInputs,
		"rootSlow": slowInputs,
	}); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	if !farm.IsRoot("rootFast") || !farm.IsRoot("rootSlow") {
		t.Fatal("both roots should remain root")
	}
	if farm.IsHead("rootFast") || farm.IsHead("rootSlow") {
		t.Fatal("roots should not remain head after connecting downstream")
	}
	if farm.IsRoot("head") {
		t.Fatal("head should not be root after receiving from two upstreams")
	}
	if !farm.IsHead("head") {
		t.Fatal("head should remain head")
	}

	if visited != seedCount*2 {
		t.Fatalf("visited = %d, want %d", visited, seedCount*2)
	}
	if got := len(counts); got != seedCount*2 {
		t.Fatalf("len(counts) = %d, want %d", got, seedCount*2)
	}

	for i := 0; i < seedCount; i++ {
		fast := i*10 + 1
		slow := i*10 + 2
		if counts[fast] != 1 {
			t.Fatalf("fanin fast result %d count = %d, want 1", fast, counts[fast])
		}
		if counts[slow] != 1 {
			t.Fatalf("fanin slow result %d count = %d, want 1", slow, counts[slow])
		}
	}

	if int(rootFast.GetState()) != 2 {
		t.Fatalf("rootFast state = %d, want 2", rootFast.GetState())
	}
	if int(rootSlow.GetState()) != 2 {
		t.Fatalf("rootSlow state = %d, want 2", rootSlow.GetState())
	}
	if int(head.GetState()) != 2 {
		t.Fatalf("head state = %d, want 2", head.GetState())
	}
}

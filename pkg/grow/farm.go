package grow

import "fmt"

// Farm 管理多节点静态图的注册与连接。
// 当前仅负责持有节点、校验名称唯一性，并建立超边式连接。
type Farm struct {
	plots       []PlotNode
	plotsByName map[string]PlotNode
	edges       map[string]map[string]struct{}
	roots       map[string]struct{}
	heads       map[string]struct{}
}

func NewFarm() *Farm {
	return &Farm{
		plotsByName: make(map[string]PlotNode),
		edges:       make(map[string]map[string]struct{}),
		roots:       make(map[string]struct{}),
		heads:       make(map[string]struct{}),
	}
}

// ==== 查询接口 ====

// PlotCount 返回当前已注册的 plot 数量。
func (f *Farm) PlotCount() int {
	return len(f.plots)
}

// HasPlot 返回指定名称的 plot 是否已注册。
func (f *Farm) HasPlot(name string) bool {
	if f.plotsByName == nil {
		return false
	}
	_, ok := f.plotsByName[name]
	return ok
}

// GetPlot 按名称返回已注册的 plot。
func (f *Farm) GetPlot(name string) (PlotNode, bool) {
	if f.plotsByName == nil {
		return nil, false
	}
	plot, ok := f.plotsByName[name]
	return plot, ok
}

// IsRoot 返回指定 plot 当前是否为 root。
// root plot 没有任何上游。
func (f *Farm) IsRoot(name string) bool {
	if f.roots == nil {
		return false
	}
	_, ok := f.roots[name]
	return ok
}

// IsHead 返回指定 plot 当前是否为 head。
// head plot 没有任何下游。
func (f *Farm) IsHead(name string) bool {
	if f.heads == nil {
		return false
	}
	_, ok := f.heads[name]
	return ok
}

// ==== 注册接口 ====

// AddPlot 将一个或多个 plot 注册到 farm 中。
// plot 名称必须唯一。
func (f *Farm) AddPlot(plots ...PlotNode) error {
	for _, plot := range plots {
		if plot == nil {
			return fmt.Errorf("plot is nil")
		}

		name := plot.GetName()
		if name == "" {
			return fmt.Errorf("plot name cannot be empty")
		}
		if _, exists := f.plotsByName[name]; exists {
			return fmt.Errorf("plot %q already exists", name)
		}

		f.plots = append(f.plots, plot)
		f.plotsByName[name] = plot
		f.roots[name] = struct{}{}
		f.heads[name] = struct{}{}
	}

	return nil
}

// ==== 连接接口 ====

// uniquePlots 确保 plot 名称唯一。
func uniquePlots(plots []PlotNode) []PlotNode {
	seen := make(map[string]struct{}, len(plots))
	unique := make([]PlotNode, 0, len(plots))
	for _, plot := range plots {
		if plot == nil {
			continue
		}
		name := plot.GetName()
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		unique = append(unique, plot)
	}
	return unique
}

// requireRegistered 确保 plot 已注册到 farm 中。
func (f *Farm) requireRegistered(plot PlotNode) error {
	if plot == nil {
		return fmt.Errorf("plot is nil")
	}
	if f.plotsByName == nil {
		return fmt.Errorf("plot %q is not registered in farm", plot.GetName())
	}
	if registered, ok := f.plotsByName[plot.GetName()]; !ok || registered != plot {
		return fmt.Errorf("plot %q is not registered in farm", plot.GetName())
	}
	return nil
}

func (f *Farm) addEdge(from, to string) {
	if f.edges[from] == nil {
		f.edges[from] = make(map[string]struct{})
	}
	f.edges[from][to] = struct{}{}
	delete(f.heads, from)
	delete(f.roots, to)
}

// Connected 返回 from -> to 是否已建立连接。
func (f *Farm) Connected(from, to string) bool {
	if f.edges == nil {
		return false
	}
	targets, ok := f.edges[from]
	if !ok {
		return false
	}
	_, ok = targets[to]
	return ok
}

// Connect 将源组中的每个 plot 与目标组中的每个 plot 相连。
// 这会形成一个“源组 x 目标组”的全连接，用于表达超边。
func (f *Farm) Connect(fromPlots []PlotNode, toPlots []PlotNode) error {
	fromUnique := uniquePlots(fromPlots)
	toUnique := uniquePlots(toPlots)

	if len(fromUnique) == 0 {
		return fmt.Errorf("from plots cannot be empty")
	}
	if len(toUnique) == 0 {
		return fmt.Errorf("to plots cannot be empty")
	}

	for _, from := range fromUnique {
		if err := f.requireRegistered(from); err != nil {
			return err
		}
	}
	for _, to := range toUnique {
		if err := f.requireRegistered(to); err != nil {
			return err
		}
	}

	for _, from := range fromUnique {
		for _, to := range toUnique {
			if err := from.ConnectTo(to); err != nil {
				return err
			}
			f.addEdge(from.GetName(), to.GetName())
		}
	}

	return nil
}

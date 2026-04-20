package grow

import (
	"fmt"
	"time"

	"github.com/Mr-xiaotian/CelestialForge/pkg/funnel"
)

// ==== Struct ====

// Farm 管理多个 Plot 组成的静态有向图。
// 负责节点注册、名称唯一性校验、超边式连接建立，
// 以及统一的 spout 管理和生命周期调度。
type Farm struct {
	name        string
	plots       []PlotNode
	plotsByName map[string]PlotNode
	edges       map[string]map[string]struct{}
	roots       map[string]struct{}
	heads       map[string]struct{}

	logSpout  *funnel.Spout[LogRecord]
	logInlet  *LogInlet
	failSpout *funnel.Spout[FailRecord]
	failInlet *FailInlet
}

// ==== Constructor ====

// NewFarm 创建一个 Farm 实例。
// name 为 farm 名称（用于日志标识），logLevel 为全局日志级别。
func NewFarm(name string, logLevel string) *Farm {
	logSpout := funnel.NewSpout(&LogRecordHandler{}, 100, time.Second)
	failSpout := funnel.NewSpout(&FailRecordHandler{}, 100, time.Second)
	logInlet := NewLogInlet(logSpout.GetQueue(), time.Second, logLevel)
	failInlet := NewFailInlet(failSpout.GetQueue(), time.Second)

	return &Farm{
		name:        name,
		plots:       make([]PlotNode, 0),
		plotsByName: make(map[string]PlotNode),
		edges:       make(map[string]map[string]struct{}),
		roots:       make(map[string]struct{}),
		heads:       make(map[string]struct{}),

		logSpout:  logSpout,
		logInlet:  logInlet,
		failSpout: failSpout,
		failInlet: failInlet,
	}
}

// ==== Getters ====

// PlotCount 返回已注册的 plot 数量。
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

// GetPlot 按名称返回已注册的 plot，未找到时 ok 为 false。
func (f *Farm) GetPlot(name string) (PlotNode, bool) {
	if f.plotsByName == nil {
		return nil, false
	}
	plot, ok := f.plotsByName[name]
	return plot, ok
}

// IsRoot 判断指定 plot 是否为 root（无上游）。
func (f *Farm) IsRoot(name string) bool {
	if f.roots == nil {
		return false
	}
	_, ok := f.roots[name]
	return ok
}

// IsHead 判断指定 plot 是否为 head（无下游）。
func (f *Farm) IsHead(name string) bool {
	if f.heads == nil {
		return false
	}
	_, ok := f.heads[name]
	return ok
}

// Connected 返回 from → to 是否已建立连接。
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

// rootPlots 返回所有 root plot（无上游的入口节点）。
func (f *Farm) rootPlots() []PlotNode {
	roots := make([]PlotNode, 0, len(f.roots))
	for name := range f.roots {
		if plot, ok := f.plotsByName[name]; ok {
			roots = append(roots, plot)
		}
	}
	return roots
}

// headPlots 返回所有 head plot（无下游的末端节点）。
func (f *Farm) headPlots() []PlotNode {
	heads := make([]PlotNode, 0, len(f.heads))
	for name := range f.heads {
		if plot, ok := f.plotsByName[name]; ok {
			heads = append(heads, plot)
		}
	}
	return heads
}

// ==== Registration ====

// AddPlot 将一个或多个 plot 注册到 farm。
// plot 名称不能为空且必须唯一，注册时默认标记为 root 和 head。
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

// requireRegistered 确保 plot 已注册到 farm 中，用于连接前校验。
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

// ==== Connection ====

// uniquePlots 对 plot 列表按名称去重，过滤 nil。
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

// addEdge 记录一条 from → to 的有向边，并更新 root/head 状态。
func (f *Farm) addEdge(from, to string) {
	if f.edges[from] == nil {
		f.edges[from] = make(map[string]struct{})
	}
	f.edges[from][to] = struct{}{}
	delete(f.heads, from)
	delete(f.roots, to)
}

// Connect 在源组和目标组之间建立全连接（笛卡尔积）。
// 每条连接调用 from.ConnectTo(to) 将上游 fruitChan 接入下游 seedChan，
// 并在下游登记上游名称用于 seal 聚合。
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
			to.AddUpstream(from.GetName())
			f.addEdge(from.GetName(), to.GetName())
		}
	}

	return nil
}

// ==== Execution ====

// validateStartInputs 校验输入参数：plot 必须已注册且为 root。
func (f *Farm) validateStartInputs(inputs map[string][]any) error {
	for name := range inputs {
		plot, ok := f.plotsByName[name]
		if !ok {
			return fmt.Errorf("plot %q is not registered in farm", name)
		}
		if !f.IsRoot(name) {
			return fmt.Errorf("plot %q is not a root plot", name)
		}
		if err := f.requireRegistered(plot); err != nil {
			return err
		}
	}
	return nil
}

// Start 同步启动整张 farm 图。
// inputs 按 plot 名称声明初始种子，仅允许注入 root plot。
// 流程：启动全局 spout → 绑定各 plot inlet → 启动所有 plot →
// 注入种子 → 封闭所有 root → 等待所有 plot 完成 → 停止 spout。
func (f *Farm) Start(inputs map[string][]any) error {
	if err := f.validateStartInputs(inputs); err != nil {
		return err
	}

	f.logSpout.Start()
	f.failSpout.Start()
	defer f.failSpout.Stop()
	defer f.logSpout.Stop()

	startTime := time.Now()
	f.logInlet.StartFarm(f.name)

	for _, plot := range f.plots {
		plot.BindInlet(f.logSpout.GetQueue(), f.failSpout.GetQueue())
	}

	for _, plot := range f.plots {
		plot.StartAsync()
	}

	for name, seeds := range inputs {
		plot := f.plotsByName[name]
		for idx, seed := range seeds {
			if err := plot.SeedAny(idx, seed); err != nil {
				return err
			}
		}
	}

	for _, plot := range f.rootPlots() {
		plot.Seal()
	}

	for _, plot := range f.plots {
		plot.WaitAsync()
	}

	f.logInlet.EndFarm(f.name, time.Since(startTime).Seconds())

	return nil
}

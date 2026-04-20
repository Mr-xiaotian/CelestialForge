# grow.Farm

> 源文件: `pkg/grow/farm.go`

## 概述

`farm.go` 实现了多个 Plot 组成的静态有向图管理器。Farm 负责节点注册、名称唯一性校验、超边式连接建立，以及统一的 spout 管理和生命周期调度。所有 Plot 共享 Farm 级别的日志和失败记录 spout。

## 类型

### `Farm`

多节点静态图管理器。

| 字段 | 类型 | 说明 |
|------|------|------|
| `name` | `string` | farm 名称（用于日志标识） |
| `plots` | `[]PlotNode` | 按注册顺序排列的所有 plot |
| `plotsByName` | `map[string]PlotNode` | 按名称索引的 plot 映射 |
| `edges` | `map[string]map[string]struct{}` | 有向边邻接表 |
| `roots` | `map[string]struct{}` | root plot 集合（无上游） |
| `heads` | `map[string]struct{}` | head plot 集合（无下游） |
| `logSpout` / `logInlet` | | 全局日志系统 |
| `failSpout` / `failInlet` | | 全局失败记录系统 |

#### 构造函数

##### `NewFarm(name, logLevel string) *Farm`

创建 Farm 实例，初始化全局日志和失败记录 spout。

#### Getter 方法

- **`PlotCount()`** — 返回已注册 plot 数量
- **`HasPlot(name)`** — 判断 plot 是否已注册
- **`GetPlot(name)`** — 按名称获取 plot
- **`IsRoot(name)`** — 判断是否为 root（无上游）
- **`IsHead(name)`** — 判断是否为 head（无下游）
- **`Connected(from, to)`** — 判断两个 plot 间是否已建立连接
- **`rootPlots()`** — 返回所有 root plot
- **`headPlots()`** — 返回所有 head plot

#### 注册方法

##### `AddPlot(plots ...PlotNode) error`

将一个或多个 plot 注册到 farm。名称不能为空且必须唯一，注册时默认标记为 root 和 head。

##### `requireRegistered(plot) error`

校验 plot 已注册到 farm 中，用于连接前校验。

#### 连接方法

##### `Connect(fromPlots, toPlots []PlotNode) error`

在源组和目标组之间建立全连接（笛卡尔积）。每条连接调用 `from.ConnectTo(to)` 将上游 fruitChan 接入下游 seedChan，并在下游登记上游名称用于 seal 聚合。

##### `addEdge(from, to string)`

记录一条有向边并更新 root/head 状态。

##### `uniquePlots(plots []PlotNode) []PlotNode`

包级函数，对 plot 列表按名称去重并过滤 nil。

#### 执行方法

##### `Start(inputs map[string][]any) error`

同步启动整张 farm 图。流程：

1. 校验输入（`validateStartInputs`）
2. 启动全局 spout
3. 绑定各 plot 的 inlet
4. 异步启动所有 plot
5. 向 root plot 注入种子
6. 封闭所有 root
7. 等待所有 plot 完成
8. 停止全局 spout

##### `validateStartInputs(inputs) error`

校验输入参数：plot 必须已注册且为 root。

## 使用示例

```go
// 创建 plot
plotA := grow.NewPlot("str2int", strconv.Atoi, nil, grow.WithTends(4))
plotB := grow.NewPlot("int2str", strconv.Itoa, nil, grow.WithTends(4))

// 创建 farm 并注册
farm := grow.NewFarm("pipeline", "INFO")
farm.AddPlot(plotA, plotB)

// 建立连接：A -> B
farm.Connect([]grow.PlotNode{plotA}, []grow.PlotNode{plotB})

// 启动
farm.Start(map[string][]any{
    "str2int": {"1", "2", "3"},
})
```

## 关联文件

- [plot.md](plot.md) — `PlotNode` 接口和 `Plot` 实现
- [log.md](log.md) — Farm 级别共享日志 spout
- [fail.md](fail.md) — Farm 级别共享失败记录 spout

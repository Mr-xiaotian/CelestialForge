# grow.Plot

> 源文件: `pkg/grow/plot.go`

## 概述

`plot.go` 实现了 `grow` 包的核心并发种子培育器。`Plot` 将一组种子分发给 tend 协程池并行处理，通过 `funnel` 包异步记录日志和失败信息。支持同步模式（`Start`）和异步模式（`StartAsync`），可单独运行（standalone）或由 `Farm` 统一调度。

## 接口

### `PlotNode`

Farm 管理 Plot 时使用的统一接口，擦除泛型参数使 Farm 可以用同一类型持有不同种子/果实类型的 Plot。

| 方法 | 说明 |
|------|------|
| `GetName() string` | 返回 plot 名称 |
| `GetState() int32` | 返回当前状态 |
| `GetSeedChanAny() any` | 返回 seedChan 的动态类型表示 |
| `ConnectTo(next PlotNode) error` | 连接到下游 plot |
| `AddUpstream(name string)` | 登记上游 plot 名称 |
| `BindInlet(logChan, failChan)` | 绑定日志和失败记录通道 |
| `StartAsync()` | 异步启动 |
| `WaitAsync()` | 等待异步完成 |
| `SeedAny(id int, seed any) error` | 弱类型播种 |
| `Seal()` | 发送终止信号 |

## 类型

### `Plot[S any, F any]`

并发种子培育器。S 为种子类型，F 为果实类型。

| 字段 | 类型 | 说明 |
|------|------|------|
| `name` | `string` | plot 名称，在 Farm 中需唯一 |
| `cultivator` | `func(S) (F, error)` | 培育函数 |
| `observers` | `[]Observer` | 进度观察者列表 |
| `numTends` | `int` | 最大并发 tend 数量 |
| `chanSize` | `int` | seedChan/fruitChan 缓冲区大小 |
| `maxRetries` | `int` | 最大重试次数（不含首次） |
| `retryDelay` | `func(attempt int) time.Duration` | 重试间隔策略 |
| `retryIf` | `func(error) bool` | 错误过滤器 |
| `logLevel` | `string` | 日志最低级别 |
| `seedChan` | `chan Payload[S]` | 种子输入通道 |
| `fruitChans` | `[]chan Payload[F]` | 果实输出通道列表（多下游 fan-out） |
| `upstreams` | `map[string]struct{}` | 已登记上游名称集合 |
| `sealedFrom` | `map[string]struct{}` | 已收到 seal 信号的上游集合 |
| `logSpout` / `logInlet` | | 日志系统（standalone 模式拥有 spout） |
| `failSpout` / `failInlet` | | 失败记录系统 |
| `ctx` / `cancel` | | 取消控制 |
| `wg` | `sync.WaitGroup` | 异步协程同步 |
| `state` | `atomic.Int32` | 0=idle, 1=running, 2=done |
| `Counter` | 嵌入 | 种子计数器 |

#### 构造函数

##### `NewPlot[S, F](name, cultivator, observers, ...Option) *Plot[S, F]`

创建 Plot 实例。初始时 `fruitChans` 为空，standalone 模式由 `InitLocalEnv` 添加本地 fruitChan，Farm 模式由 `ConnectTo` 添加下游 seedChan。

#### 初始化方法

- **`InitLocalEnv()`** — 初始化 standalone 模式所需的本地环境：创建 fruitChan、日志/失败 spout 并绑定 inlet
- **`BindInlet(logChan, failChan)`** — 绑定日志和失败记录的写入通道
- **`StartSpouts()`** / **`StopSpouts()`** — 启动/停止本地 spout（仅 standalone 模式）

#### 连接方法

- **`addFruitChan(fruitChan)`** — 添加一个下游果实通道
- **`AddUpstream(name)`** — 登记上游 plot 名称，供 seal 聚合使用
- **`ConnectTo(next PlotNode) error`** — 将果实输出连接到下游 plot 的种子输入，通过类型断言校验类型匹配
- **`resetSeals()`** — 重置所有上游 seal 状态
- **`markSealed(source) bool`** — 标记上游已 seal，全部完成时返回 true

#### Getter 方法

- **`GetName()`** — 返回名称
- **`GetState()`** — 返回状态
- **`GetSeedChanAny()`** — 以 any 类型返回 seedChan

#### 内部流水线

- **`seed(seeds []S)`** — 将种子切片逐个发送到 seedChan，完成后发送 `SignalSeal`
- **`tend(seedPayload, sem, done)`** — 照料单颗种子，执行 cultivator 并按策略重试
- **`sprout()`** — 调度器，从 seedChan 读取种子，通过信号量控制并发分发给 tend 协程
- **`harvest()`** — 从 `fruitChans[0]` 收集所有果实（仅 standalone 同步模式）
- **`bearFruit(seedPayload, fruit, startTime)`** — 处理成功：计数、日志、fan-out 到所有下游
- **`bearWeed(seedPayload, err, startTime)`** — 处理失败：计数、日志、失败记录

#### 同步 API

##### `Start(seeds []S) []Karma[S, F]`

同步启动 Plot，自动初始化本地环境、启停 spout，阻塞直到所有种子培育完成并返回 Karma 列表。

#### 异步 API

- **`SeedAny(id, seed any) error`** — 弱类型播种，供 Farm 统一注入使用
- **`Seed(id, seed S)`** — 播入单颗种子到 seedChan
- **`Seal()`** — 向 seedChan 发送 `SignalSeal`，通知 sprout 不再有新种子
- **`Harvest(sickle func(Payload[F]), chanIndex int)`** — 异步启动果实消费协程
- **`StartAsync()`** — 异步启动 sprout 调度器
- **`WaitAsync()`** — 等待所有异步协程退出

## 使用示例

```go
// Standalone 同步模式
plot := grow.NewPlot("hasher", hashFunc,
    grow.WithTends(8),
)
plot.AddObserver(grow.NewProgressBar("Hashing"))
karmas := plot.Start(files)

// Standalone 异步模式
plot := grow.NewPlot("worker", processFunc, grow.WithTends(4))
plot.InitLocalEnv()
plot.StartSpouts()
plot.StartAsync()
for i, item := range items {
    plot.Seed(i, item)
}
plot.Seal()
plot.Harvest(func(res grow.Payload[Result]) {
    fmt.Println(res.Value)
}, 0)
plot.WaitAsync()
plot.StopSpouts()
```

## 关联文件

- [farm.md](farm.md) — Farm 通过 PlotNode 接口管理多个 Plot 组成有向图
- [type.md](type.md) — `Payload` 和 `Karma` 类型定义
- [option.md](option.md) — `NewPlot` 的可选配置参数
- [counter.md](counter.md) — `Counter` 嵌入字段
- [observer.md](observer.md) — `Observer` 接口及 `ProgressBar` 实现
- [log.md](log.md) — 日志系统
- [fail.md](fail.md) — 失败记录系统
- [helper.md](helper.md) — `trunc` 截断函数

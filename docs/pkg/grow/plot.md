# grow.Plot

> 源文件: `pkg/grow/plot.go`

## 概述

`plot.go` 是 `grow` 包的核心文件，实现了一个泛型并发任务执行器。`Plot` 采用生产者-消费者模式，通过信号量控制并发度，将种子分发给多个 tend 协程并行培育。它集成了进度观察、结构化日志和失败记录三大子系统，提供同步（`Start`）和异步（`Seed`/`StartAsync`/`WaitAsync`/`Harvest`）两套 API，适用于批量任务处理场景（如并行文件哈希、批量网络请求等）。

## 类型

### `Plot[S any, F any]`

泛型并发任务执行器。`S` 为种子（输入）类型，`F` 为果实（输出）类型。

| 字段           | 类型                              | 说明                                       |
| -------------- | --------------------------------- | ------------------------------------------ |
| `Name`         | `string`                         | Plot 名称，用于日志和失败记录标识          |
| `cultivator`   | `func(S) (F, error)`            | 培育函数，由用户提供                       |
| `numTends`     | `int`                            | 最大并发 tend 数量                         |
| `maxRetries`   | `int`                            | 最大重试次数（默认 1，即不重试）           |
| `retryDelay`   | `func(attempt int) time.Duration`| 重试间隔策略                               |
| `retryIf`      | `func(error) bool`              | 判断错误是否值得重试                       |
| `SeedChan`     | channel                          | 种子载荷通道（`Payload[S]`）               |
| `FruitChan`    | channel                          | 果实载荷通道（`Payload[F]`）               |
| `ControlChan`  | channel                          | 控制信号通道（`ControlSignal`）            |
| `observers`    | `[]Observer`                     | 注册的进度观察者列表                       |
| `logSpout`     | funnel.Spout                     | 日志消费端                                 |
| `logInlet`     | `*LogInlet`                      | 日志生产端                                 |
| `failSpout`    | funnel.Spout                     | 失败记录消费端                             |
| `failInlet`    | `*FailInlet[S]`                  | 失败记录生产端                             |
| `ctx`          | `context.Context`                | 用于取消执行的上下文                       |
| `cancel`       | `context.CancelFunc`             | 取消函数                                   |
| `state`        | `atomic.Int32`                   | 执行器状态：0=idle, 1=running, 2=done      |
| `Counter`      | (embedded)                       | 嵌入的种子计数器                           |

### 构造函数

#### `NewPlot[S, F](name string, cultivator func(S)(F,error), observers []Observer, opts ...Option) *Plot[S,F]`

创建一个新的 Plot 实例。

**参数**:
- `name` — Plot 名称
- `cultivator` — 培育函数，接收类型 `S` 的种子，返回类型 `F` 的果实或错误
- `observers` — 进度观察者列表（如 `[]Observer{NewProgressBar("desc")}`），无需观察者时传 `nil`
- `opts` — 可选配置项（如 `WithTends(n)`, `WithMaxRetries(n)` 等）

### 方法

#### 状态查询

##### `State() int32`

返回当前 Plot 状态：`0`（空闲）、`1`（运行中）、`2`（已完成）。

#### 同步 API

##### `Start(seeds []S) []Karma[S,F]`

同步执行所有种子并返回结果列表。该方法会阻塞直到所有种子培育完成。内部依次调用 `seed` -> `sprout` -> `harvest`。

**参数**: `seeds` — 待培育的种子切片

**返回值**: `[]Karma[S,F]` — 所有成功培育的种子-果实配对

#### 异步 API

##### `Seed(id int, seed S)`

向种子通道播入单颗种子。用于异步模式下逐个添加种子。

**参数**:
- `id` — 种子 ID
- `seed` — 种子数据

##### `Seal()`

封闭种子入口，通知 sprout 调度器不再有新种子。

##### `StartAsync()`

启动异步执行模式。开始 `sprout` 调度循环，但不等待完成。外部通过 `Seed` 播种，通过 `Harvest` 收获。

##### `Harvest(onSuccess func(Payload[F]))`

逐个收获果实，阻塞直到 `FruitChan` 关闭。每当一颗种子成功培育时，调用 `onSuccess` 回调处理果实。

##### `WaitAsync()`

等待异步 Plot 结束并清理资源。

#### 取消 API

##### `Close()`

立即取消 Plot，强制停止所有操作。慎用，可能导致未完成的种子丢失。

### 内部方法

#### `seed(seeds []S)`

将种子切片包装为 `Payload[S]` 并逐个发送到 `SeedChan`，完成后发送 `ControlSignal`。

#### `tend(seedPayload Payload[S], sem, done chan struct{})`

照料单个种子的 tend 函数。包含 panic 恢复和重试机制。根据 `maxRetries`、`retryIf`、`retryDelay` 配置进行重试。

#### `sprout()`

基于信号量的种子调度器。从 `SeedChan` 读取种子，为每颗种子启动一个 tend 协程，通过信号量（`numTends`）控制最大并发数。同时监听 `ctx.Done()` 支持取消。

#### `harvest() []Karma[S,F]`

从 `FruitChan` 收集所有成功培育的果实，组装为 `Karma` 切片返回。

#### `processFruit(seedPayload, fruit, startTime)`

处理成功的种子：更新计数器、报告进度、记录日志、将果实发送到 `FruitChan`。

#### `handleWeed(seedPayload, err)`

处理失败的种子：更新计数器、报告进度、记录日志和失败记录。

## 使用示例

### 同步模式

```go
hashFile := func(path string) (string, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return "", err
    }
    hash := sha256.Sum256(data)
    return hex.EncodeToString(hash[:]), nil
}

bar := grow.NewProgressBar("Hashing files")
plot := grow.NewPlot("file-hasher", hashFile,
    []grow.Observer{bar},
    grow.WithTends(8),
)

files := []string{"a.txt", "b.txt", "c.txt"}
karmas := plot.Start(files)

for _, k := range karmas {
    fmt.Printf("%s -> %s\n", k.Seed, k.Fruit)
}
```

### 异步模式

```go
plot := grow.NewPlot("downloader", downloadFunc, nil,
    grow.WithTends(4),
)

go plot.StartAsync()

for i, url := range urls {
    plot.Seed(i, url)
}
plot.Seal()

plot.Harvest(func(p grow.Payload[[]byte]) {
    fmt.Printf("Downloaded item %d: %d bytes\n", p.ID, len(p.Value))
})

plot.WaitAsync()
```

### 带重试

```go
plot := grow.NewPlot("api-caller", callAPI, nil,
    grow.WithTends(4),
    grow.WithMaxRetries(3),
    grow.WithRetryDelay(func(attempt int) time.Duration {
        return time.Duration(attempt+1) * time.Second
    }),
    grow.WithRetryIf(func(err error) bool {
        return !errors.Is(err, ErrNotFound)
    }),
)
```

## 执行流程

```
Start(seeds)
  |
  v
seed(seeds) ──> SeedChan ──> sprout() ──> tend() ──> FruitChan ──> harvest()
  |                              |            |
  v                              |            ├── processFruit()
ControlChan ─────────────────────┘            |     ├── Counter.AddSuccess()
                                              |     ├── Observer.OnProgress()
                                              |     └── LogInlet.SeedSuccess()
                                              |
                                              └── handleWeed()
                                                    ├── Counter.AddFailed()
                                                    ├── Observer.OnProgress()
                                                    ├── LogInlet.SeedError()
                                                    └── FailInlet.SeedError()
```

## 关联文件

- [type.md](type.md) — `Payload` 和 `Karma` 数据类型定义
- [option.md](option.md) — `Option` 模式配置（`WithTends`、`WithMaxRetries` 等）
- [control.md](control.md) — `ControlSignal` 用于 `seed` 与 `sprout` 之间的协调
- [counter.md](counter.md) — `Counter` 嵌入到 `Plot` 中，跟踪种子完成状态
- [observer.md](observer.md) — `Observer` 接口及 `ProgressBar` 实现
- [log.md](log.md) — `LogInlet` 和 `LogRecordHandler` 提供结构化日志
- [fail.md](fail.md) — `FailInlet` 和 `FailRecordHandler` 提供失败记录
- [helper.md](helper.md) — `trunc` 函数用于截断日志和失败记录中的字符串

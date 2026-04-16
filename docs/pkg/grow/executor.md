# grow.Executor

> 源文件: `pkg/grow/executor.go`

## 概述

`executor.go` 是 `grow` 包的核心文件，实现了一个泛型并发任务执行器。`Executor` 采用生产者-消费者模式，通过信号量控制并发度，将任务分发给多个 tend 协程并行处理。它集成了进度观察、结构化日志和失败记录三大子系统，提供同步（`Start`）和异步（`Seed`/`StartAsync`/`WaitAsync`/`Collect`）两套 API，适用于批量任务处理场景（如并行文件哈希、批量网络请求等）。

## 类型

### `Executor[T any, R any]`

泛型并发任务执行器。`T` 为输入任务类型，`R` 为处理结果类型。

| 字段           | 类型                              | 说明                                       |
| -------------- | --------------------------------- | ------------------------------------------ |
| `Name`         | `string`                         | 执行器名称，用于日志和失败记录标识         |
| `processor`    | `func(T) (R, error)`            | 任务处理函数，由用户提供                   |
| `numTends`   | `int`                            | 最大并发 tend 数量                       |
| `wg`           | `sync.WaitGroup`                 | 等待所有 tend 完成                       |
| `TaskChan`     | channel                          | 任务载荷通道（`Payload[T]`）               |
| `ResultChan`   | channel                          | 结果载荷通道（`Payload[R]`）               |
| `ControlChan`  | channel                          | 控制信号通道（`ControlSignal`）            |
| `observers`    | `[]Observer`                     | 注册的进度观察者列表                       |
| `logSpout`     | funnel.Spout                     | 日志消费端                                 |
| `logInlet`     | `*LogInlet`                      | 日志生产端                                 |
| `failSpout`    | funnel.Spout                     | 失败记录消费端                             |
| `failInlet`    | `*FailInlet[T]`                  | 失败记录生产端                             |
| `state`        | `atomic.Int32`                   | 执行器状态：0=idle, 1=running, 2=done      |
| `Counter`      | (embedded)                       | 嵌入的任务计数器                           |

### 构造函数

#### `NewExecutor[T, R](name string, processor func(T)(R,error), numTends int, observers ...Observer) *Executor[T,R]`

创建一个新的执行器实例。

**参数**:
- `name` — 执行器名称
- `processor` — 任务处理函数，接收类型 `T` 的任务，返回类型 `R` 的结果或错误
- `numTends` — 最大并发 tend 数量
- `observers` — 可选的进度观察者（如 `ProgressBar`）

### 方法

#### 状态查询

##### `State() int32`

返回当前执行器状态：`0`（空闲）、`1`（运行中）、`2`（已完成）。

#### 同步 API

##### `Start(tasks []T) []TaskResult[T,R]`

同步执行所有任务并返回结果列表。该方法会阻塞直到所有任务处理完成。内部依次调用 `seed` -> `dispatch` -> `collect`。

**参数**: `tasks` — 待处理的任务切片

**返回值**: `[]TaskResult[T,R]` — 所有成功完成的任务及其结果

#### 异步 API

##### `Seed(id int, task T)`

向任务通道注入单个任务。用于异步模式下逐个添加任务。

**参数**:
- `id` — 任务 ID
- `task` — 任务数据

##### `StartAsync()`

启动异步执行模式。开始 `dispatch` 调度循环，但不等待完成。调用前需通过 `Seed` 注入任务。

##### `Collect(onSuccess func(Payload[R]))`

注册结果回调并启动结果收集。每当一个任务成功完成时，调用 `onSuccess` 回调处理结果。

##### `WaitAsync()`

等待异步执行完成。阻塞直到所有 tend 完成处理。

### 内部方法

#### `seed(tasks []T)`

将任务切片包装为 `Payload[T]` 并逐个发送到 `TaskChan`，完成后发送 `ControlSignal`。

#### `tend(taskPayload Payload[T])`

处理单个任务的 tend 函数。包含 panic 恢复机制，确保单个任务的 panic 不会导致整个执行器崩溃。

#### `dispatch()`

基于信号量的任务调度器。从 `TaskChan` 读取任务，为每个任务启动一个 tend 协程，通过信号量（`numTends`）控制最大并发数。

#### `collect() []TaskResult[T,R]`

从 `ResultChan` 收集所有成功处理的结果，组装为 `TaskResult` 切片返回。

#### `reportProgress()`

调用所有注册的 `Observer` 的 `OnProgress` 方法。

#### `notifyStart()`

将状态设为 running（1），调用所有 `Observer` 的 `OnStart` 方法。

#### `notifyFinish()`

将状态设为 done（2），调用所有 `Observer` 的 `OnFinish` 方法。

#### `processTaskSuccess(taskPayload, result, startTime)`

处理成功的任务：更新计数器、报告进度、记录日志、将结果发送到 `ResultChan`。

#### `handleTaskError(taskPayload, err)`

处理失败的任务：更新计数器、报告进度、记录日志和失败记录。

## 使用示例

### 同步模式

```go
// 定义处理函数
hashFile := func(path string) (string, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return "", err
    }
    hash := sha256.Sum256(data)
    return hex.EncodeToString(hash[:]), nil
}

// 创建带进度条的执行器
bar := grow.NewProgressBar("Hashing files")
executor := grow.NewExecutor("file-hasher", hashFile, 8, bar)

// 同步执行
files := []string{"a.txt", "b.txt", "c.txt"}
results := executor.Start(files)

for _, r := range results {
    fmt.Printf("%s -> %s\n", r.Task, r.Result)
}
```

### 异步模式

```go
executor := grow.NewExecutor("downloader", downloadFunc, 4)

// 注册结果回调
executor.Collect(func(p grow.Payload[[]byte]) {
    fmt.Printf("Downloaded item %d: %d bytes\n", p.ID, len(p.Value))
})

// 逐个注入任务
for i, url := range urls {
    executor.Seed(i, url)
}

// 启动异步执行
executor.StartAsync()

// 等待完成
executor.WaitAsync()
```

## 执行流程

```
Start(tasks)
  |
  v
seed(tasks) ──> TaskChan ──> dispatch() ──> tend() ──> ResultChan ──> collect()
  |                              |              |
  v                              |              ├── processTaskSuccess()
ControlChan ─────────────────────┘              |     ├── Counter.AddSuccess()
                                                |     ├── Observer.OnProgress()
                                                |     └── LogInlet.TaskSuccess()
                                                |
                                                └── handleTaskError()
                                                      ├── Counter.AddFailed()
                                                      ├── Observer.OnProgress()
                                                      ├── LogInlet.TaskError()
                                                      └── FailInlet.TaskError()
```

## 关联文件

- [type.md](type.md) — `Payload` 和 `TaskResult` 数据类型定义
- [controal.md](controal.md) — `ControlSignal` 用于 `seed` 与 `dispatch` 之间的协调
- [counter.md](counter.md) — `Counter` 嵌入到 `Executor` 中，跟踪任务完成状态
- [observer.md](observer.md) — `Observer` 接口及 `ProgressBar` 实现
- [log.md](log.md) — `LogInlet` 和 `LogRecordHandler` 提供结构化日志
- [fail.md](fail.md) — `FailInlet` 和 `FailRecordHandler` 提供失败记录
- [helper.md](helper.md) — `trunc` 函数用于截断日志和失败记录中的字符串
- [../../funnel/inlet.md](../../funnel/inlet.md) — 日志和失败记录使用的异步通道生产端
- [../../funnel/spout.md](../../funnel/spout.md) — 日志和失败记录使用的异步通道消费端

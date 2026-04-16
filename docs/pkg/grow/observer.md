# grow.Observer

> 源文件: `pkg/grow/observer.go`

## 概述

`observer.go` 定义了任务执行进度的观察者模式接口及其默认实现。`Observer` 接口提供了三个生命周期钩子，允许外部组件监听 `Executor` 的启动、进度更新和完成事件。包内提供了基于 `progressbar/v3` 的 `ProgressBar` 实现，在终端中以进度条形式实时展示任务处理进度。

## 类型

### `Observer` (interface)

任务执行进度的观察者接口。可通过 `NewExecutor` 的可变参数注入一个或多个观察者。

| 方法                                 | 说明                                             |
| ------------------------------------ | ------------------------------------------------ |
| `OnStart(total int)`                | 当 `Executor` 开始执行时调用，传入任务总数       |
| `OnProgress(completed, total int)`  | 每当一个任务完成（成功或失败）时调用             |
| `OnFinish(completed, total int)`    | 当所有任务执行完毕时调用                         |

### `ProgressBar`

基于 [`schollz/progressbar/v3`](https://github.com/schollz/progressbar) 的 `Observer` 实现，在终端中渲染实时进度条。

#### 构造函数

##### `NewProgressBar(description string) *ProgressBar`

创建一个带有指定描述文字的进度条。`description` 将显示在进度条的前缀位置。

#### 方法

##### `OnStart(total int)`

初始化进度条，设置最大值为 `total`。

##### `OnProgress(completed, total int)`

更新进度条的当前值为 `completed`。

##### `OnFinish(completed, total int)`

完成进度条渲染，标记为结束状态。

## 使用示例

```go
// 使用内置进度条
bar := grow.NewProgressBar("Processing files")
executor := grow.NewExecutor("hasher", hashFunc, 8, bar)
executor.Start(files)

// 自定义 Observer 实现
type MetricsObserver struct {
    startTime time.Time
}

func (m *MetricsObserver) OnStart(total int) {
    m.startTime = time.Now()
    log.Printf("Starting %d tasks", total)
}

func (m *MetricsObserver) OnProgress(completed, total int) {
    elapsed := time.Since(m.startTime)
    rate := float64(completed) / elapsed.Seconds()
    log.Printf("%.1f tasks/sec", rate)
}

func (m *MetricsObserver) OnFinish(completed, total int) {
    log.Printf("Finished %d tasks in %v", completed, time.Since(m.startTime))
}

executor := grow.NewExecutor("tend", processFunc, 4, bar, &MetricsObserver{})
```

## 关联文件

- [executor.md](executor.md) — `Executor` 在 `notifyStart`、`reportProgress`、`notifyFinish` 中依次调用所有注册的 `Observer`
- [counter.md](counter.md) — `Counter` 提供 `GetCompleted` 和 `GetTotal` 数据供 `Observer` 回调使用

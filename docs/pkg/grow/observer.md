# grow.Observer

> 源文件: `pkg/grow/observer.go`

## 概述

`observer.go` 定义了种子培育进度的观察者模式接口及其默认实现。`Observer` 接口提供了三个生命周期钩子，允许外部组件监听 `Plot` 的启动、进度更新和完成事件。包内提供了基于 `progressbar/v3` 的 `ProgressBar` 实现，在终端中以进度条形式实时展示种子培育进度。

## 接口

### `Observer`

种子培育进度的观察者接口。通过 `NewPlot` 的 `observers` 参数注入。

| 方法 | 说明 |
|------|------|
| `OnStart(total int)` | 当 `Plot` 开始执行时调用，传入种子总数 |
| `OnProgress(completed, total int)` | 每当一颗种子完成培育（成功或失败）时调用 |
| `OnFinish(completed, total int)` | 当所有种子培育完毕时调用 |

## 类型

### `ProgressBar`

基于 [`schollz/progressbar/v3`](https://github.com/schollz/progressbar) 的 `Observer` 实现，在终端中渲染实时进度条。

#### 构造函数

##### `NewProgressBar(description string) *ProgressBar`

创建一个带有指定描述文字的进度条。`description` 将显示在进度条的前缀位置。

#### 方法

##### `ensureBar(total int)`

延迟初始化进度条，在首次知道 total 时创建。

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
plot := grow.NewPlot("hasher", hashFunc,
    grow.WithTends(8),
)
plot.AddObserver(bar)
plot.Start(files)
```

## 关联文件

- [plot.md](plot.md) — `Plot` 在 `notifyStart`、`reportProgress`、`notifyFinish` 中依次调用所有注册的 `Observer`
- [counter.md](counter.md) — `Counter` 提供 `GetCompleted` 和 `GetSeedNum` 数据供 `Observer` 回调使用

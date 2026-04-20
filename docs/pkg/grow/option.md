# grow.Option

> 源文件: `pkg/grow/option.go`

## 概述

`option.go` 实现了 `Plot` 的函数式选项配置模式。调用方可以灵活配置并发数、通道大小、重试策略、日志级别等参数，同时保持 `NewPlot` 签名的稳定性和可扩展性。

## 类型

### `Option`

`func(*plotOptions)` 类型的函数，用于修改 `plotOptions` 的配置项。

### `plotOptions`

内部配置结构体，存储所有可选参数。

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `numTends` | `int` | `runtime.NumCPU()` | 并发 tend 协程数 |
| `chanSize` | `int` | `runtime.NumCPU()` | seedChan/fruitChan 缓冲区大小 |
| `maxRetries` | `int` | `1` | 最大重试次数（不含首次） |
| `retryDelay` | `func(int) time.Duration` | `func(int) { return 0 }` | 重试间隔策略 |
| `retryIf` | `func(error) bool` | `func(error) { return true }` | 错误过滤器 |
| `logLevel` | `string` | `"INFO"` | 日志最低级别 |

## 函数

### `WithTends(n int) Option`

设置并发 tend 协程数。

### `WithChanSize(n int) Option`

设置 seedChan/fruitChan 的缓冲区大小。

### `WithMaxRetries(n int) Option`

设置最大重试次数（不含首次执行）。例如 `WithMaxRetries(2)` 表示最多执行 3 次。

### `WithRetryDelay(fn func(attempt int) time.Duration) Option`

设置重试间隔策略。`attempt` 从 1 开始递增。

### `WithRetryIf(fn func(error) bool) Option`

设置错误过滤器，返回 true 的错误才会触发重试。

### `WithLogLevel(level string) Option`

设置日志最低级别。

## 使用示例

```go
plot := grow.NewPlot("api-caller", callAPI, nil,
    grow.WithTends(4),
    grow.WithChanSize(16),
    grow.WithMaxRetries(3),
    grow.WithRetryDelay(func(attempt int) time.Duration {
        return time.Duration(attempt) * time.Second
    }),
    grow.WithRetryIf(func(err error) bool {
        return !errors.Is(err, ErrPermanent)
    }),
    grow.WithLogLevel("DEBUG"),
)
```

## 关联文件

- [plot.md](plot.md) — `NewPlot` 接受 `...Option` 参数

# grow.Option

> 源文件: `pkg/grow/option.go`

## 概述

`option.go` 实现了 `Plot` 的 Option 模式配置。通过函数式选项，调用方可以灵活配置 `Plot` 的并发数、重试策略等参数，同时保持 `NewPlot` 签名的稳定性和可扩展性。

## 类型

### `Option`

`func(*plotOptions)` 类型的函数，用于修改 `plotOptions` 的配置项。

### `plotOptions`

内部配置结构体，存储所有可选参数。

| 字段         | 类型                              | 默认值                                      | 说明                       |
| ------------ | --------------------------------- | ------------------------------------------- | -------------------------- |
| `numTends`   | `int`                            | `runtime.NumCPU()`                          | 最大并发 tend 数量         |
| `maxRetries` | `int`                            | `1`（仅执行一次，不重试）                   | 最大重试次数               |
| `retryDelay` | `func(attempt int) time.Duration`| `func(int) time.Duration { return time.Second }` | 重试间隔策略         |
| `retryIf`    | `func(error) bool`              | `func(error) bool { return true }`          | 判断错误是否值得重试       |

## 函数

### `WithTends(n int) Option`

设置并发工作协程数。

### `WithMaxRetries(n int) Option`

设置最大重试次数。`maxRetries` 表示总执行次数，例如 `WithMaxRetries(3)` 表示最多执行 3 次。

### `WithRetryDelay(fn func(attempt int) time.Duration) Option`

设置重试间隔策略。`attempt` 从 0 开始。可实现固定间隔、指数退避等策略。

### `WithRetryIf(fn func(error) bool) Option`

设置哪些错误值得重试。返回 `true` 则重试，返回 `false` 则立即放弃。

## 使用示例

```go
// 基本用法
plot := grow.NewPlot("hasher", hashFunc, nil,
    grow.WithTends(8),
)

// 带重试的用法
plot := grow.NewPlot("api-caller", callAPI, nil,
    grow.WithTends(4),
    grow.WithMaxRetries(3),
    grow.WithRetryDelay(func(attempt int) time.Duration {
        return time.Duration(attempt+1) * time.Second  // 线性退避
    }),
    grow.WithRetryIf(func(err error) bool {
        return !errors.Is(err, ErrPermanent)  // 永久性错误不重试
    }),
)
```

## 关联文件

- [plot.md](plot.md) — `NewPlot` 接受 `...Option` 参数，内部通过 `defaultOptions()` 初始化默认值

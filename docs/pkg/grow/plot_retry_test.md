# plot_retry_test

> 源文件: `pkg/grow/plot_retry_test.go`

## 概述

`Plot` 重试机制的测试，采用黑盒测试（`package grow_test`）。

## 测试函数

| 测试函数 | 说明 |
|---------|------|
| `TestPlot_RetrySuccess` | 前 2 次失败第 3 次成功，`MaxRetries(3)` |
| `TestPlot_RetryExhausted` | 重试耗尽仍失败，`MaxRetries(2)` → 共 3 次尝试 |
| `TestPlot_RetryIf` | `RetryIf` 过滤永久性错误，仅执行 1 次 |
| `TestPlot_RetryDelay` | 验证重试间隔（100ms）被正确应用 |

## 关联文件

- [plot.md](plot.md) — `Plot.tend` 重试逻辑
- [option.md](option.md) — `WithMaxRetries`、`WithRetryDelay`、`WithRetryIf`

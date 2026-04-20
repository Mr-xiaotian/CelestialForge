# plot_async_test

> 源文件: `pkg/grow/plot_async_test.go`

## 概述

`Plot` 异步模式（`InitLocalEnv` + `StartSpouts` + `StartAsync` + `Seed` + `Seal` + `Harvest` + `WaitAsync` + `StopSpouts`）的测试，采用黑盒测试（`package grow_test`）。

## 测试函数

| 测试函数 | 说明 |
|---------|------|
| `TestPlot_Async` | 异步流程：逐个注入种子、密封、收获果实，验证 5 个结果 |

## 关联文件

- [plot.md](plot.md) — `Plot` 异步 API 实现

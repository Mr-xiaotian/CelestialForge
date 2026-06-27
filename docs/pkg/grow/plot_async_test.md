# plot_async_test

> 最后更新日期: 2026/06/28

> 源文件: `pkg/grow/plot_async_test.go`

## 概述

`Plot` 异步模式（`InitLocalEnv` + `StartSpouts` + `StartAsync` + `Seed` + `Seal` + `Harvest` + `WaitAsync` + `StopSpouts`）的测试，采用黑盒测试（`package grow_test`）。

## 测试函数

| 测试函数 | 说明 |
|---------|------|
| `TestPlot_Async` | 异步流程：先启动 `Harvest` 消费果实，再注入 50 颗种子、密封、等待完成，验证收到 50 个结果且 plot 状态为 2（done） |

## 关联文件

- [plot.md](plot.md) — `Plot` 异步 API 实现
# plot_sync_test

> 源文件: `pkg/grow/plot_sync_test.go`

## 概述

`Plot` 同步模式（`Start`）下的基本功能测试，采用黑盒测试（`package grow_test`）。

## 测试函数

| 测试函数 | 说明 |
|---------|------|
| `TestPlot_AllError` | 全部种子失败：验证无果实返回，状态码为 2 |
| `TestPlot_PartialError` | 部分种子失败：偶数种子失败，验证 3 个成功 |
| `TestPlot_AllSuccess` | 全部种子成功：验证每个果实 = 种子 × 2 |

## 关联文件

- [plot.md](plot.md) — `Plot.Start` 实现

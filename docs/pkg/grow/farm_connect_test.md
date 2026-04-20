# farm_connect_test

> 源文件: `pkg/grow/farm_connect_test.go`

## 概述

`Farm` 注册和连接功能的测试，采用黑盒测试（`package grow_test`）。

## 测试函数

| 测试函数 | 说明 |
|---------|------|
| `TestFarmAddPlot` | 注册 plot：验证 PlotCount、HasPlot、GetPlot、root/head 状态 |
| `TestFarmAddPlotDuplicateName` | 重名注册：验证返回错误 |
| `TestFarmConnectHyperEdge` | 超边连接（1→2）：验证边关系、root/head 状态更新、去重 |
| `TestFarmConnectTypeMismatch` | 类型不匹配连接：验证返回错误 |

## 关联文件

- [farm.md](farm.md) — `Farm.AddPlot`、`Farm.Connect` 实现

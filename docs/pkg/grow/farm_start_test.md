# farm_start_test

> 源文件: `pkg/grow/farm_start_test.go`

## 概述

`Farm` 启动执行功能的测试，采用黑盒测试（`package grow_test`）。

## 测试函数

| 测试函数 | 说明 |
|---------|------|
| `TestFarmStartLinear` | 线性管道（root→head）：验证数据正确流转和状态 |
| `TestFarmStartRejectNonRootInput` | 非 root 注入：验证返回错误 |

## 关联文件

- [farm.md](farm.md) — `Farm.Start` 实现

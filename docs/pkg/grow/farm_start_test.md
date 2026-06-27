# farm_start_test

> 最后更新日期: 2026/06/28

> 源文件: `pkg/grow/farm_start_test.go`

## 概述

`Farm` 启动执行功能的测试，采用黑盒测试（`package grow_test`）。

## 测试函数

| 测试函数 | 说明 |
|---------|------|
| `TestFarmStartLinear` | 线性管道（root→head）：向 root 注入 3 个种子，验证 head 收到翻倍结果 `{2,4,6}`，且两个 plot 最终状态均为 2（done） |

## 关联文件

- [farm.md](farm.md) — `Farm.Start` 实现
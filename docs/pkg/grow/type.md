# grow.Type

> 源文件: `pkg/grow/type.go`

## 概述

`type.go` 定义了 `grow` 包的基础数据类型和信号常量。`Payload` 作为管道阶段的统一数据载体，同时承载正常数据和控制信号；`Karma` 记录种子与果实的配对结果。

## 常量

| 常量 | 值 | 说明 |
|------|------|------|
| `SignalNone` | `0` | 正常数据 |
| `SignalSeal` | `1` | 终止信号，通知下游不再有新数据 |

## 类型

### `Payload[V any]`

管道阶段的统一数据载体。数据流和控制流共用同一通道，通过 `Signal` 字段区分。

| 字段 | 类型 | 说明 |
|------|------|------|
| `ID` | `int` | 种子序号，用于追踪到原始输入 |
| `Value` | `V` | 数据值 |
| `Prev` | `any` | 上一阶段的种子值 |
| `Signal` | `int` | 控制信号（`SignalNone` 或 `SignalSeal`） |
| `Source` | `string` | 来源 Plot 名称 |

### `Karma[S any, F any]`

种子与果实的配对，记录一颗种子培育后的完整结果。

| 字段 | 类型 | 说明 |
|------|------|------|
| `Seed` | `S` | 原始种子 |
| `Fruit` | `F` | 培育结果 |

## 关联文件

- [plot.md](plot.md) — `Payload` 在 seedChan/fruitChan 中流转，`Karma` 由 `harvest` 返回

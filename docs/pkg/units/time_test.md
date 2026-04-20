# time_test

> 源文件: `pkg/units/time_test.go`

## 概述

`units.HumanTime` 时间格式化功能的测试，采用黑盒测试（`package units_test`）。使用表驱动测试模式，验证秒数到人类可读时间字符串的转换是否正确。

## 测试函数

| 测试函数 | 说明 |
|---------|------|
| `TestHumanTime` | 表驱动测试，验证 `units.NewHumanTime().String()` 的输出格式 |

**测试用例：**

| 输入（秒） | 期望输出 |
|------------|----------|
| 97 | `1m 37.00s` |
| 0 | `0s` |
| 1008 | `16m 48.00s` |
| 81 | `1m 21.00s` |

## 关联文件

- [time.md](time.md) — `HumanTime` 实现

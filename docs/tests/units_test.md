# tests.TestHumanTime

> 源文件: `tests/units_test.go`

## 概述

`units.HumanTime` 时间格式化功能的测试套件。采用表驱动测试模式，验证秒数到人类可读时间字符串的转换是否正确。测试覆盖了零值、常规值和较大值等多种输入场景，确保格式化输出符合预期的 "Xm Ys" 格式规范。

## 类型/函数

### `TestHumanTime(t *testing.T)`

表驱动测试，验证 `units.NewHumanTime().String()` 的输出格式。

**测试用例：**

| 输入（秒） | 期望输出 |
|------------|----------|
| 97 | `1m 37.00s` |
| 0 | `0s` |
| 1008 | `16m 48.00s` |
| 81 | `1m 21.00s` |

## 使用示例

```bash
# 运行时间格式化测试
go test -v -run TestHumanTime ./tests/

# 运行所有 units 相关测试
go test -v -run "HumanTime|HumanBytes" ./tests/
```

## 关联文件

- [executor_test.md](executor_test.md) — 同一测试包中的执行器测试
- [file_test.md](file_test.md) — 同一测试包中的文件操作测试
- [structs_test.md](structs_test.md) — 同一测试包中的数据结构测试
- [../cmd/debug/main.md](../cmd/debug/main.md) — debug_bytes 函数提供 HumanBytes 的手动调试

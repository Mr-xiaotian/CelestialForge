# units 包测试

> 源文件: `pkg/units/time_test.go`

## 概述

`units.HumanTime` 时间格式化功能的测试套件，采用黑盒测试（`package units_test`）。使用表驱动测试模式，验证秒数到人类可读时间字符串的转换是否正确。

## 测试文件

### `time_test.go`

#### `TestHumanTime(t *testing.T)`

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
# 运行 units 包测试
go test -v ./pkg/units/
```

## 关联文件

- [file_test.md](file_test.md) — file 包测试
- [structs_test.md](structs_test.md) — structs 包测试

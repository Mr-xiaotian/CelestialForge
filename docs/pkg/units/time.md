# units.time

> 源文件: `pkg/units/time.go`

## 概述

本文件定义了 `HumanTime` 类型，用于以人类可读的方式表示和操作时间持续时长。`HumanTime` 的底层类型为 `float64`，存储以秒为单位的时长值，支持负数。格式化输出使用 `d`（天）、`h`（小时）、`m`（分钟）、`s`（秒）等单位，并提供精度和零值显示的控制参数。该类型在 `pkg/units/time_test.go` 中有完整的测试覆盖。

## 类型/函数

### `HumanTime`

底层类型为 `float64`，以秒为单位存储时间持续时长。

**格式化方法：**

| 方法 | 签名 | 说明 |
|------|------|------|
| `String` | `(t HumanTime) String() string` | 等价于 `t.Format(2, false)` |
| `Format` | `(t HumanTime) Format(precision int, showZero bool) string` | 格式化为 `"1d 2h 3m 4.56s"`，`precision` 控制秒的小数位数，`showZero` 控制是否显示值为零的单位 |

**算术方法：**

| 方法 | 签名 | 说明 |
|------|------|------|
| `Add` | `(t HumanTime) Add(other HumanTime) HumanTime` | 加法运算 |
| `Sub` | `(t HumanTime) Sub(other HumanTime) HumanTime` | 减法运算 |
| `Mul` | `(t HumanTime) Mul(n float64) HumanTime` | 乘法运算（标量） |
| `Div` | `(t HumanTime) Div(n float64) HumanTime` | 除法运算（标量） |
| `Neg` | `(t HumanTime) Neg() HumanTime` | 取反（返回负值） |

**转换方法：**

| 方法 | 签名 | 说明 |
|------|------|------|
| `Float64` | `(t HumanTime) Float64() float64` | 返回原始秒数 |

### `NewHumanTime`

```go
func NewHumanTime(seconds float64) HumanTime
```

根据秒数创建 `HumanTime` 实例。

### `ParseHumanTime`

```go
func ParseHumanTime(text string) (HumanTime, error)
```

解析人类可读的时间字符串，支持 `"1d 2h 3m 4.56s"` 格式。

## 使用示例

```go
package main

import "celestialforge/pkg/units"

func main() {
	// 创建 HumanTime
	duration := units.NewHumanTime(90061.5) // 1天 1小时 1分钟 1.5秒
	fmt.Println(duration)                   // "1d 1h 1m 1.5s"

	// 自定义格式化
	fmt.Println(duration.Format(0, false))  // "1d 1h 1m 2s"（精度为0，四舍五入）
	fmt.Println(duration.Format(3, true))   // "1d 1h 1m 1.500s"（显示完整精度）

	// 解析字符串
	parsed, err := units.ParseHumanTime("2h 30m")
	if err != nil {
		panic(err)
	}
	fmt.Println(parsed.Float64()) // 9000

	// 算术运算
	a := units.NewHumanTime(3600) // 1h
	b := units.NewHumanTime(1800) // 30m
	fmt.Println(a.Add(b))         // "1h 30m"
	fmt.Println(a.Neg())          // "-1h"
}
```

## 关联文件

- [bytes.md](bytes.md) — 同属 units 包的字节单位类型

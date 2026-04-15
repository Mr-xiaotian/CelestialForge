# units.bytes

> 源文件: `pkg/units/bytes.go`

## 概述

本文件定义了 `HumanBytes` 类型，用于以人类可读的方式表示和操作字节大小。该类型在整个项目中被广泛使用，特别是在 file 包中（如 `FileInfo.Size`、`GetFileSize`、`GetDirSize`、`DuplicateReport` 等）。`HumanBytes` 采用二进制单位制（1 KB = 1024 B），支持算术运算以及人类可读格式的字符串化和解析。

## 类型/函数

### `HumanBytes`

底层类型为 `int64`，表示字节数。

**算术方法：**

| 方法 | 签名 | 说明 |
|------|------|------|
| `Add` | `(h HumanBytes) Add(other HumanBytes) HumanBytes` | 加法运算 |
| `Sub` | `(h HumanBytes) Sub(other HumanBytes) HumanBytes` | 减法运算 |
| `Mul` | `(h HumanBytes) Mul(n int64) HumanBytes` | 乘法运算（标量） |
| `Div` | `(h HumanBytes) Div(n int64) HumanBytes` | 除法运算（标量） |
| `Mod` | `(h HumanBytes) Mod(n int64) HumanBytes` | 取模运算（标量） |

**转换方法：**

| 方法 | 签名 | 说明 |
|------|------|------|
| `String` | `(h HumanBytes) String() string` | 返回人类可读格式，如 `"1GB 512MB"` |
| `Int64` | `(h HumanBytes) Int64() int64` | 返回原始字节数 |

### `NewHumanBytes`

```go
func NewHumanBytes(b int64) HumanBytes
```

根据字节数创建 `HumanBytes` 实例。

### `ParseHumanBytes`

```go
func ParseHumanBytes(text string) (HumanBytes, error)
```

解析人类可读的字节字符串。支持 `"1GB 512MB"` 和 `"1.5GB"` 等格式。

### 包级变量

- `unitNames` — 单位名称列表：`["B", "KB", "MB", "GB", "TB"]`
- `unitMap` — 单位到字节数的映射，采用二进制单位制（1 KB = 1024 B）
- `parseRe` — 用于解析字节字符串的正则表达式：`(\d+(?:\.\d+)?)([A-Za-z]+)`

## 使用示例

```go
package main

import "celestialforge/pkg/units"

func main() {
	// 创建 HumanBytes
	size := units.NewHumanBytes(1024 * 1024 * 1024 + 512 * 1024 * 1024) // 1.5 GB
	fmt.Println(size) // "1GB 512MB"

	// 解析字符串
	parsed, err := units.ParseHumanBytes("1GB 512MB")
	if err != nil {
		panic(err)
	}
	fmt.Println(parsed.Int64()) // 1610612736

	// 算术运算
	a := units.NewHumanBytes(1024) // 1KB
	b := units.NewHumanBytes(2048) // 2KB
	fmt.Println(a.Add(b))         // "3KB"
	fmt.Println(b.Mul(3))         // "6KB"
}
```

## 关联文件

- [time.md](time.md) — 同属 units 包的时间单位类型

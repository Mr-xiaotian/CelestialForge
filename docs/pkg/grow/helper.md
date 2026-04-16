# grow.Helper

> 源文件: `pkg/grow/helper.go`

## 概述

`helper.go` 提供了 `grow` 包内部使用的工具函数。目前包含字符串截断函数 `trunc`，用于在日志和失败记录中将过长的种子/果实文本表示截短为可读长度，避免日志文件过于膨胀。

## 函数

### `trunc(s string, maxLen int) string`

将字符串 `s` 截断到最大长度 `maxLen`。如果字符串长度未超过 `maxLen`，原样返回。内部使用 `[]rune` 转换以正确处理中文等多字节字符。

**截断策略**: 保留前 1/3 和后 1/3 的内容，中间用 `"..."` 连接。每段至少 1 个字符（通过 `max(1, maxLen/3)` 保证）。这种方式能同时保留字符串的开头（通常包含类型/标识信息）和结尾（通常包含关键数据），比单纯截断尾部更有利于调试。

**参数**:
- `s` — 待截断的字符串
- `maxLen` — 最大允许长度

**返回值**: 截断后的字符串，长度不超过 `maxLen`

## 使用示例

```go
// 短字符串，原样返回
trunc("hello", 100) // => "hello"

// 长字符串，前1/3 + "..." + 后1/3
trunc("abcdefghijklmnopqrstuvwxyz", 15)
// => "abcde...vwxyz"
```

## 关联文件

- [plot.md](plot.md) — `processFruit` 和 `handleWeed` 方法在生成日志和失败记录时使用 `trunc` 截断种子和果实的字符串表示
- [log.md](log.md) — 日志中的 `seedRepr` 和 `fruitRepr` 经过 `trunc` 处理
- [fail.md](fail.md) — 失败记录中的种子描述经过 `trunc` 处理

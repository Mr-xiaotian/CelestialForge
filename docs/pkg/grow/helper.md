# grow.Helper

> 源文件: `pkg/grow/helper.go`

## 概述

`helper.go` 提供了 `grow` 包内部使用的工具函数。目前包含字符串截断函数 `trunc`，用于在日志和失败记录中将过长的种子/果实文本表示截短为可读长度，避免日志文件过于膨胀。

## 函数

### `trunc(s string, maxLen int) string`

将字符串 `s` 截断到最大长度 `maxLen`。如果字符串长度未超过 `maxLen`，原样返回。内部使用 `[]rune` 转换以正确处理中文等多字节字符。

**截断策略**: 保留前 1/3 和后 1/3 的内容，中间用 `"..."` 连接。每段至少 1 个字符（通过 `max(1, maxLen/3)` 保证）。

## 使用示例

```go
trunc("hello", 100)                        // => "hello"
trunc("abcdefghijklmnopqrstuvwxyz", 15)    // => "abcde...vwxyz"
```

## 关联文件

- [plot.md](plot.md) — `bearFruit` 和 `bearWeed` 方法在生成日志和失败记录时使用 `trunc` 截断种子和果实的字符串表示
- [log.md](log.md) — 日志中的 `seedRepr` 和 `fruitRepr` 经过 `trunc` 处理

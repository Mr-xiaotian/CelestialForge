# structs.symmetric_map

> 源文件: `pkg/structs/symmetric_map.go`

## 概述

本文件定义了 `SymmetricMap` 泛型类型，实现了一个双向一对一映射（对称映射）。每个元素最多出现一次，通过 `forward` 和 `reverse` 两个内部字典实现双向查找。`Set` 操作是幂等的，当设置新的配对关系时会自动解除旧的冲突配对。可以通过 `allowSelf` 参数控制是否允许元素与自身配对。该类型在 `pkg/structs/symmetric_map_test.go` 中有完整的测试覆盖。

## 类型/函数

### `Pair[T comparable]`

表示一个配对关系的结构体。

| 字段 | 类型 | 说明 |
|------|------|------|
| `A` | `T` | 配对的第一个元素 |
| `B` | `T` | 配对的第二个元素 |

### `SymmetricMap[T comparable]`

泛型双向一对一映射结构体。

| 内部字段 | 类型 | 说明 |
|----------|------|------|
| `forward` | `map[T]T` | 正向映射（a -> b），存储代表键 |
| `reverse` | `map[T]T` | 反向映射（b -> a） |
| `allowSelf` | `bool` | 是否允许自配对（a <-> a） |

#### 构造函数

| 函数 | 签名 | 说明 |
|------|------|------|
| `NewSymmetricMap` | `func NewSymmetricMap[T comparable](allowSelf bool) *SymmetricMap[T]` | 创建空的对称映射 |
| `FromMap` | `func FromMap[T comparable](m map[T]T, allowSelf bool) (*SymmetricMap[T], error)` | 从普通 map 构建对称映射 |
| `FromPairs` | `func FromPairs[T comparable](pairs []Pair[T], allowSelf bool) (*SymmetricMap[T], error)` | 从 Pair 切片构建对称映射 |

#### 基础操作

| 方法 | 签名 | 说明 |
|------|------|------|
| `Set` | `(m *SymmetricMap[T]) Set(a, b T) error` | 设置配对关系，幂等操作，会自动解除旧的冲突配对 |
| `Get` | `(m *SymmetricMap[T]) Get(x T) (T, bool)` | 双向查找，返回配对方和是否存在 |
| `MustGet` | `(m *SymmetricMap[T]) MustGet(x T) T` | 双向查找，未找到时 panic |
| `Delete` | `(m *SymmetricMap[T]) Delete(x T) bool` | 删除包含 x 的配对，返回是否存在 |
| `Pop` | `(m *SymmetricMap[T]) Pop(x T) (T, bool)` | 删除配对并返回配对方 |

#### 视图/便捷方法

| 方法 | 签名 | 说明 |
|------|------|------|
| `Contains` | `(m *SymmetricMap[T]) Contains(x T) bool` | 检查元素是否存在于映射中 |
| `Len` | `(m *SymmetricMap[T]) Len() int` | 返回配对数量 |
| `Keys` | `(m *SymmetricMap[T]) Keys() []T` | 返回所有代表键（仅正向侧） |
| `Values` | `(m *SymmetricMap[T]) Values() []T` | 返回所有值（仅正向侧的值） |
| `Items` | `(m *SymmetricMap[T]) Items() []Pair[T]` | 返回所有配对 |
| `Clear` | `(m *SymmetricMap[T]) Clear()` | 清空所有配对 |

#### 显示

| 方法 | 签名 | 说明 |
|------|------|------|
| `String` | `(m *SymmetricMap[T]) String() string` | 返回格式如 `"SymmetricMap(a <-> b, c <-> d)"` |

## 使用示例

```go
package main

import "celestialforge/pkg/structs"

func main() {
	// 创建对称映射（不允许自配对）
	sm := structs.NewSymmetricMap[string](false)

	// 设置配对
	sm.Set("alice", "bob")
	sm.Set("charlie", "diana")

	// 双向查找
	partner, ok := sm.Get("alice")   // "bob", true
	partner, ok = sm.Get("bob")      // "alice", true

	// Set 的幂等性和自动解除
	sm.Set("alice", "charlie") // 自动解除 alice<->bob 和 charlie<->diana

	// 从 Pair 切片构建
	pairs := []structs.Pair[int]{
		{A: 1, B: 2},
		{A: 3, B: 4},
	}
	intMap, err := structs.FromPairs(pairs, false)
	if err != nil {
		panic(err)
	}

	// 遍历
	for _, pair := range intMap.Items() {
		fmt.Printf("%d <-> %d\n", pair.A, pair.B)
	}

	// 删除并返回配对方
	val, ok := intMap.Pop(1) // val=2, ok=true

	fmt.Println(sm) // "SymmetricMap(alice <-> charlie)"
}
```

## 关联文件

- [../../tests/structs_test.md](../../tests/structs_test.md) — SymmetricMap 的测试文件

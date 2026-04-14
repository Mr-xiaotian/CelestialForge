package structs

import (
	"fmt"
	"strings"
)

// SymmetricMap 一对一双向映射：每个元素至多出现一次，与唯一的伙伴绑定。
//   - 迭代/Keys() 仅遍历"代表键"（forward 的键），每对只出现一次。
//   - Contains(x) 表示 x 在任意一侧出现。
type SymmetricMap[T comparable] struct {
	forward   map[T]T // a -> b（代表键）
	reverse   map[T]T // b -> a
	allowSelf bool
}

// Pair 表示一对映射关系
type Pair[T comparable] struct {
	A, B T
}

// NewSymmetricMap 创建空的 SymmetricMap
func NewSymmetricMap[T comparable](allowSelf bool) *SymmetricMap[T] {
	return &SymmetricMap[T]{
		forward:   make(map[T]T),
		reverse:   make(map[T]T),
		allowSelf: allowSelf,
	}
}

// FromMap 从普通 map 构建 SymmetricMap
func FromMap[T comparable](m map[T]T, allowSelf bool) (*SymmetricMap[T], error) {
	sm := NewSymmetricMap[T](allowSelf)
	for a, b := range m {
		if err := sm.Set(a, b); err != nil {
			return nil, err
		}
	}
	return sm, nil
}

// FromPairs 从配对切片构建 SymmetricMap
func FromPairs[T comparable](pairs []Pair[T], allowSelf bool) (*SymmetricMap[T], error) {
	sm := NewSymmetricMap[T](allowSelf)
	for _, p := range pairs {
		if err := sm.Set(p.A, p.B); err != nil {
			return nil, err
		}
	}
	return sm, nil
}

// ============ 基础操作 ============

// Set 设置一对映射 a <-> b。若 a 或 b 已有配对，先解绑旧关系。
func (m *SymmetricMap[T]) Set(a, b T) error {
	if !m.allowSelf && a == b {
		return fmt.Errorf("self-pair is not allowed (a == b)")
	}

	// 已有相同配对则幂等返回
	if existing, ok := m.Get(a); ok && existing == b {
		return nil
	}

	// a 或 b 若已配对，先解绑
	if m.Contains(a) {
		m.Delete(a)
	}
	if m.Contains(b) {
		m.Delete(b)
	}

	m.forward[a] = b
	m.reverse[b] = a
	return nil
}

// Get 获取与 x 绑定的伙伴。第二个返回值表示是否找到。
func (m *SymmetricMap[T]) Get(x T) (T, bool) {
	if v, ok := m.forward[x]; ok {
		return v, true
	}
	if v, ok := m.reverse[x]; ok {
		return v, true
	}
	var zero T
	return zero, false
}

// MustGet 获取与 x 绑定的伙伴，未找到则 panic。
func (m *SymmetricMap[T]) MustGet(x T) T {
	v, ok := m.Get(x)
	if !ok {
		panic(fmt.Sprintf("SymmetricMap: key %v not found", x))
	}
	return v
}

// Delete 删除 x 所在的配对。
func (m *SymmetricMap[T]) Delete(x T) bool {
	if y, ok := m.forward[x]; ok {
		delete(m.forward, x)
		delete(m.reverse, y)
		return true
	}
	if y, ok := m.reverse[x]; ok {
		delete(m.reverse, x)
		delete(m.forward, y)
		return true
	}
	return false
}

// Pop 移除并返回与 x 绑定的伙伴。
func (m *SymmetricMap[T]) Pop(x T) (T, bool) {
	if y, ok := m.forward[x]; ok {
		delete(m.forward, x)
		delete(m.reverse, y)
		return y, true
	}
	if y, ok := m.reverse[x]; ok {
		delete(m.reverse, x)
		delete(m.forward, y)
		return y, true
	}
	var zero T
	return zero, false
}

// ============ 视图/便捷方法 ============

// Contains 判断 x 是否在映射的任意一侧出现。
func (m *SymmetricMap[T]) Contains(x T) bool {
	_, inF := m.forward[x]
	_, inR := m.reverse[x]
	return inF || inR
}

// Len 返回配对数量。
func (m *SymmetricMap[T]) Len() int {
	return len(m.forward)
}

// Keys 返回所有代表键（每对只出现一次）。
func (m *SymmetricMap[T]) Keys() []T {
	keys := make([]T, 0, len(m.forward))
	for k := range m.forward {
		keys = append(keys, k)
	}
	return keys
}

// Values 返回所有代表值（代表键对应的伙伴）。
func (m *SymmetricMap[T]) Values() []T {
	vals := make([]T, 0, len(m.forward))
	for _, v := range m.forward {
		vals = append(vals, v)
	}
	return vals
}

// Items 返回所有配对。
func (m *SymmetricMap[T]) Items() []Pair[T] {
	pairs := make([]Pair[T], 0, len(m.forward))
	for a, b := range m.forward {
		pairs = append(pairs, Pair[T]{A: a, B: b})
	}
	return pairs
}

// Clear 清空所有配对关系。
func (m *SymmetricMap[T]) Clear() {
	clear(m.forward)
	clear(m.reverse)
}

// ============ 显示 ============

// String 返回人类可读表示，如 "SymmetricMap(a <-> b, c <-> d)"
func (m *SymmetricMap[T]) String() string {
	if len(m.forward) == 0 {
		return "SymmetricMap()"
	}
	parts := make([]string, 0, len(m.forward))
	for a, b := range m.forward {
		parts = append(parts, fmt.Sprintf("%v <-> %v", a, b))
	}
	return "SymmetricMap(" + strings.Join(parts, ", ") + ")"
}

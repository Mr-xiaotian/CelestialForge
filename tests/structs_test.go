package tests

import (
	"testing"

	"github.com/Mr-xiaotian/CelestialForge/pkg/structs"
)

func TestSymmetricMap_SetGet(t *testing.T) {
	m := structs.NewSymmetricMap[string](false)

	if err := m.Set("a", "b"); err != nil {
		t.Fatalf("Set(a, b) error: %v", err)
	}

	// 正向查找
	v, ok := m.Get("a")
	if !ok || v != "b" {
		t.Errorf("Get(a) = %q, %v, want b, true", v, ok)
	}

	// 反向查找
	v, ok = m.Get("b")
	if !ok || v != "a" {
		t.Errorf("Get(b) = %q, %v, want a, true", v, ok)
	}

	// 不存在的键
	_, ok = m.Get("c")
	if ok {
		t.Error("Get(c) should return false")
	}
}

func TestSymmetricMap_SelfPair(t *testing.T) {
	// 默认不允许 self-pair
	m := structs.NewSymmetricMap[int](false)
	if err := m.Set(1, 1); err == nil {
		t.Error("Set(1, 1) should return error when allowSelf=false")
	}

	// 允许 self-pair
	m2 := structs.NewSymmetricMap[int](true)
	if err := m2.Set(1, 1); err != nil {
		t.Fatalf("Set(1, 1) with allowSelf=true error: %v", err)
	}
	v, ok := m2.Get(1)
	if !ok || v != 1 {
		t.Errorf("Get(1) = %d, %v, want 1, true", v, ok)
	}
}

func TestSymmetricMap_Idempotent(t *testing.T) {
	m := structs.NewSymmetricMap[string](false)
	_ = m.Set("a", "b")
	_ = m.Set("a", "b") // 重复设置应幂等

	if m.Len() != 1 {
		t.Errorf("Len() = %d, want 1 after idempotent Set", m.Len())
	}
}

func TestSymmetricMap_Rebind(t *testing.T) {
	m := structs.NewSymmetricMap[string](false)
	_ = m.Set("a", "b")
	_ = m.Set("a", "c") // a 改绑到 c，b 应被释放

	v, ok := m.Get("a")
	if !ok || v != "c" {
		t.Errorf("Get(a) = %q, %v, want c, true", v, ok)
	}

	// b 已不在映射中
	if m.Contains("b") {
		t.Error("b should not be in map after rebind")
	}

	if m.Len() != 1 {
		t.Errorf("Len() = %d, want 1", m.Len())
	}
}

func TestSymmetricMap_Delete(t *testing.T) {
	m := structs.NewSymmetricMap[string](false)
	_ = m.Set("a", "b")
	_ = m.Set("c", "d")

	// 从正向侧删除
	if !m.Delete("a") {
		t.Error("Delete(a) should return true")
	}
	if m.Contains("a") || m.Contains("b") {
		t.Error("a and b should be removed")
	}

	// 从反向侧删除
	if !m.Delete("d") {
		t.Error("Delete(d) should return true")
	}
	if m.Contains("c") || m.Contains("d") {
		t.Error("c and d should be removed")
	}

	// 删除不存在的键
	if m.Delete("x") {
		t.Error("Delete(x) should return false")
	}

	if m.Len() != 0 {
		t.Errorf("Len() = %d, want 0", m.Len())
	}
}

func TestSymmetricMap_Pop(t *testing.T) {
	m := structs.NewSymmetricMap[string](false)
	_ = m.Set("a", "b")

	// 正向 pop
	v, ok := m.Pop("a")
	if !ok || v != "b" {
		t.Errorf("Pop(a) = %q, %v, want b, true", v, ok)
	}

	if m.Len() != 0 {
		t.Errorf("Len() = %d after Pop, want 0", m.Len())
	}

	// 反向 pop
	_ = m.Set("c", "d")
	v, ok = m.Pop("d")
	if !ok || v != "c" {
		t.Errorf("Pop(d) = %q, %v, want c, true", v, ok)
	}

	// pop 不存在的键
	_, ok = m.Pop("x")
	if ok {
		t.Error("Pop(x) should return false")
	}
}

func TestSymmetricMap_Contains(t *testing.T) {
	m := structs.NewSymmetricMap[int](false)
	_ = m.Set(1, 2)

	if !m.Contains(1) {
		t.Error("Contains(1) should be true")
	}
	if !m.Contains(2) {
		t.Error("Contains(2) should be true")
	}
	if m.Contains(3) {
		t.Error("Contains(3) should be false")
	}
}

func TestSymmetricMap_KeysValuesItems(t *testing.T) {
	m := structs.NewSymmetricMap[string](false)
	_ = m.Set("a", "b")
	_ = m.Set("c", "d")

	keys := m.Keys()
	if len(keys) != 2 {
		t.Errorf("len(Keys()) = %d, want 2", len(keys))
	}

	vals := m.Values()
	if len(vals) != 2 {
		t.Errorf("len(Values()) = %d, want 2", len(vals))
	}

	items := m.Items()
	if len(items) != 2 {
		t.Errorf("len(Items()) = %d, want 2", len(items))
	}

	// 验证 Items 的内容能互相查找
	for _, p := range items {
		v, ok := m.Get(p.A)
		if !ok || v != p.B {
			t.Errorf("Get(%v) = %v, %v, want %v, true", p.A, v, ok, p.B)
		}
	}
}

func TestSymmetricMap_Clear(t *testing.T) {
	m := structs.NewSymmetricMap[int](false)
	_ = m.Set(1, 2)
	_ = m.Set(3, 4)

	m.Clear()

	if m.Len() != 0 {
		t.Errorf("Len() = %d after Clear, want 0", m.Len())
	}
	if m.Contains(1) || m.Contains(2) {
		t.Error("map should be empty after Clear")
	}
}

func TestSymmetricMap_MustGet(t *testing.T) {
	m := structs.NewSymmetricMap[string](false)
	_ = m.Set("a", "b")

	if v := m.MustGet("a"); v != "b" {
		t.Errorf("MustGet(a) = %q, want b", v)
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("MustGet on missing key should panic")
		}
	}()
	m.MustGet("z")
}

func TestSymmetricMap_String(t *testing.T) {
	m := structs.NewSymmetricMap[int](false)
	s := m.String()
	if s != "SymmetricMap()" {
		t.Errorf("empty String() = %q, want SymmetricMap()", s)
	}

	_ = m.Set(1, 2)
	s = m.String()
	if s == "SymmetricMap()" {
		t.Error("non-empty String() should not be SymmetricMap()")
	}
	t.Logf("String() = %s", s)
}

func TestSymmetricMap_FromMap(t *testing.T) {
	m, err := structs.FromMap(map[string]string{
		"a": "b",
		"c": "d",
	}, false)
	if err != nil {
		t.Fatalf("FromMap error: %v", err)
	}

	if m.Len() != 2 {
		t.Errorf("Len() = %d, want 2", m.Len())
	}

	v, ok := m.Get("a")
	if !ok || v != "b" {
		t.Errorf("Get(a) = %q, %v, want b, true", v, ok)
	}
}

func TestSymmetricMap_FromPairs(t *testing.T) {
	m, err := structs.FromPairs([]structs.Pair[int]{
		{A: 1, B: 2},
		{A: 3, B: 4},
	}, false)
	if err != nil {
		t.Fatalf("FromPairs error: %v", err)
	}

	if m.Len() != 2 {
		t.Errorf("Len() = %d, want 2", m.Len())
	}

	v, ok := m.Get(2)
	if !ok || v != 1 {
		t.Errorf("Get(2) = %d, %v, want 1, true", v, ok)
	}
}

func TestSymmetricMap_ConflictResolution(t *testing.T) {
	m := structs.NewSymmetricMap[string](false)
	_ = m.Set("a", "b")
	_ = m.Set("c", "b") // b 要从 a<->b 解绑，重新绑定到 c<->b

	// a 应被释放
	if m.Contains("a") {
		t.Error("a should be removed after b was rebound to c")
	}

	v, ok := m.Get("c")
	if !ok || v != "b" {
		t.Errorf("Get(c) = %q, %v, want b, true", v, ok)
	}

	v, ok = m.Get("b")
	if !ok || v != "c" {
		t.Errorf("Get(b) = %q, %v, want c, true", v, ok)
	}

	if m.Len() != 1 {
		t.Errorf("Len() = %d, want 1", m.Len())
	}
}

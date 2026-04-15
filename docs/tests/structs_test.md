# tests.TestSymmetricMap

> 源文件: `tests/structs_test.go`

## 概述

`structs.SymmetricMap` 对称映射数据结构的综合测试套件。覆盖所有公开 API 的功能测试，包括基本的存取操作、自对（self-pair）处理、幂等性、重绑定（rebind）时的自动清理、删除、弹出、包含检查、视图方法、清空、panic 行为、字符串格式化以及多种构造方式。特别关注了边界情况和冲突解决策略的正确性。

## 类型/函数

### `TestSymmetricMap_SetGet(t *testing.T)`

测试基本的 Set/Get 操作以及双向查找功能。设置 `a <-> b` 后，从 `a` 或 `b` 任一方向均可查找到对方。

### `TestSymmetricMap_SelfPair(t *testing.T)`

测试自对行为：当 `allowSelf=false` 时，尝试将元素映射到自身应被拒绝；当 `allowSelf=true` 时，应正常接受。

### `TestSymmetricMap_Idempotent(t *testing.T)`

测试幂等性：对同一对重复调用 `Set` 不会产生副作用，映射状态保持不变。

### `TestSymmetricMap_Rebind(t *testing.T)`

测试重绑定行为：当已存在 `a <-> b` 时，设置 `a <-> c` 应自动移除旧的 `a <-> b` 关系。

### `TestSymmetricMap_Delete(t *testing.T)`

测试从正向和反向两侧进行删除操作，确保删除后双向关系均被清除。

### `TestSymmetricMap_Pop(t *testing.T)`

测试弹出操作：返回指定键的配对值并同时移除该对。

### `TestSymmetricMap_Contains(t *testing.T)`

测试包含检查：验证正向和反向两侧的存在性查询。

### `TestSymmetricMap_KeysValuesItems(t *testing.T)`

验证 `Keys()`、`Values()`、`Items()` 等视图方法返回正确的数据。

### `TestSymmetricMap_Clear(t *testing.T)`

测试清空操作：调用后映射应完全为空。

### `TestSymmetricMap_MustGet(t *testing.T)`

测试强制获取：当键不存在时应触发 panic。

### `TestSymmetricMap_String(t *testing.T)`

验证字符串格式化输出的格式是否符合预期。

### `TestSymmetricMap_FromMap(t *testing.T)`

测试从 `map[T]T` 构造 `SymmetricMap` 的工厂方法。

### `TestSymmetricMap_FromPairs(t *testing.T)`

测试从 `[]Pair[T]` 切片构造 `SymmetricMap` 的工厂方法。

### `TestSymmetricMap_ConflictResolution(t *testing.T)`

测试冲突解决：当已存在 `a <-> b` 时，设置 `c <-> b` 应自动移除 `a` 的绑定关系，最终只保留 `c <-> b`。

## 使用示例

```bash
# 运行所有 SymmetricMap 测试
go test -v -run TestSymmetricMap ./tests/

# 运行特定测试
go test -v -run TestSymmetricMap_ConflictResolution ./tests/

# 运行并查看覆盖率
go test -v -cover -run TestSymmetricMap ./tests/
```

## 关联文件

- [executor_test.md](executor_test.md) — 同一测试包中的执行器测试
- [file_test.md](file_test.md) — 同一测试包中的文件操作测试
- [units_test.md](units_test.md) — 同一测试包中的单位格式化测试

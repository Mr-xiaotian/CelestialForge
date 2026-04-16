# structs 包测试

> 源文件: `pkg/structs/symmetric_map_test.go`

## 概述

`structs.SymmetricMap` 对称映射数据结构的综合测试套件，采用黑盒测试（`package structs_test`）。覆盖所有公开 API 的功能测试，包括基本的存取操作、自对（self-pair）处理、幂等性、重绑定（rebind）时的自动清理、删除、弹出、包含检查、视图方法、清空、panic 行为、字符串格式化以及多种构造方式。

## 测试文件

### `symmetric_map_test.go`

| 测试函数 | 说明 |
|---------|------|
| `TestSymmetricMap_SetGet` | 基本 Set/Get 及双向查找 |
| `TestSymmetricMap_SelfPair` | 自对行为（allowSelf 开关） |
| `TestSymmetricMap_Idempotent` | 重复 Set 幂等性 |
| `TestSymmetricMap_Rebind` | 重绑定自动清理旧关系 |
| `TestSymmetricMap_Delete` | 正向/反向删除 |
| `TestSymmetricMap_Pop` | 弹出并移除 |
| `TestSymmetricMap_Contains` | 双向存在性检查 |
| `TestSymmetricMap_KeysValuesItems` | 视图方法 |
| `TestSymmetricMap_Clear` | 清空操作 |
| `TestSymmetricMap_MustGet` | 键不存在时 panic |
| `TestSymmetricMap_String` | 字符串格式化 |
| `TestSymmetricMap_FromMap` | 从 map 构造 |
| `TestSymmetricMap_FromPairs` | 从 Pair 切片构造 |
| `TestSymmetricMap_ConflictResolution` | 冲突解决策略 |

## 使用示例

```bash
# 运行 structs 包测试
go test -v ./pkg/structs/

# 运行特定测试
go test -v -run TestSymmetricMap_ConflictResolution ./pkg/structs/
```

## 关联文件

- [file_test.md](file_test.md) — file 包测试
- [units_test.md](units_test.md) — units 包测试

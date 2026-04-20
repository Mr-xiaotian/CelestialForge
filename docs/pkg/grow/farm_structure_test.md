# farm_structure_test

> 源文件: `pkg/grow/farm_structure_test.go`

## 概述

`Farm` 拓扑结构的集成测试，采用黑盒测试（`package grow_test`）。验证多节点有向图在不同拓扑下的数据流转正确性。

## 测试函数

| 测试函数 | 说明 |
|---------|------|
| `TestFarmStructure121` | 菱形结构（1→2→1）：50 个种子经两条中间路径汇聚到 head，验证 100 个不重复结果 |
| `TestFarmStructure21FaninDifferentSpeed` | 扇入（2→1）不同速度：一个快 root 一个慢 root 汇入 head，验证全部到达 |

## 关联文件

- [farm.md](farm.md) — `Farm` 有向图管理器

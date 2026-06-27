# farm_structure_test

> 最后更新日期: 2026/06/28

> 源文件: `pkg/grow/farm_structure_test.go`

## 概述

`Farm` 拓扑结构的集成测试，采用黑盒测试（`package grow_test`）。验证多节点有向图在不同拓扑、不同速度以及部分失败场景下的数据流转正确性。

## 测试函数

| 测试函数 | 说明 |
|---------|------|
| `TestFarmStructure121` | 菱形结构（1→2→1）：50 个种子经两条中间路径汇聚到 head，验证 100 个不重复结果，root/head 状态正确 |
| `TestFarmStructure121PartialFailure` | 菱形结构下的部分失败：root 偶数失败、midB 大半失败，验证各节点 fruit/weed 计数及 head 最终收到 15 个果实 |
| `TestFarmStructureDisconnectedComponents` | 不连通组件：一组 1→2 扇出和一组 2→1 扇入同时运行，验证各自结果集合完整 |
| `TestFarmStructure21FaninDifferentSpeed` | 扇入（2→1）不同速度：一个快 root 一个慢 root 汇入 head，验证全部到达且状态为 done |

## 关联文件

- [farm.md](farm.md) — `Farm` 有向图管理器
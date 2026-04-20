# grow.Counter

> 源文件: `pkg/grow/counter.go`

## 概述

`counter.go` 实现了一个线程安全的种子计数器，用于跟踪 `Plot` 中种子的总数、成功数和失败数。内部使用 `sync/atomic` 包的原子操作保证在高并发场景下计数的正确性，支持多个 tend 协程同时更新计数而无需额外加锁。

## 类型

### `Counter`

并发安全的种子进度计数器。嵌入到 `Plot` 中，供 tend 协程在种子培育完成时更新计数。

| 字段 | 类型 | 说明 |
|------|------|------|
| `seedNum` | `atomic.Int64` | 种子总数 |
| `fruitNum` | `atomic.Int64` | 成功培育的种子数（果实） |
| `weedNum` | `atomic.Int64` | 失败的种子数（杂草） |

## 构造函数

### `NewCounter() *Counter`

创建并返回一个新的 `Counter` 实例，所有计数初始为零。

## Setters

### `SetSeedNum(seedNum int)`

原子地设置种子总数。

## Adders

### `AddSeedNum(addNNum int)`

原子地增加种子总数。

### `AddFruitNum(addNNum int)`

原子地增加成功数（果实）。

### `AddWeedNum(addNNum int)`

原子地增加失败数（杂草）。

## Getters

### `GetSeedNum() int`

返回种子总数。

### `GetFruitNum() int`

返回成功数（果实）。

### `GetWeedNum() int`

返回失败数（杂草）。

### `GetCompleted() int`

返回已完成总数（`fruitNum + weedNum`）。

## Predicates

### `IsFinish() bool`

判断所有种子是否已全部完成，即 `GetCompleted() == GetSeedNum()`。`Plot` 的 `sprout` 方法用此判断是否可以结束调度。

## 使用示例

```go
counter := grow.NewCounter()
counter.SetSeedNum(100)

// 在 tend 协程中
counter.AddFruitNum(1)
// 或
counter.AddWeedNum(1)

// 检查进度
fmt.Printf("Progress: %d/%d\n", counter.GetCompleted(), counter.GetSeedNum())

if counter.IsFinish() {
    fmt.Println("All seeds completed")
}
```

## 关联文件

- [plot.md](plot.md) — `Counter` 作为嵌入字段存在于 `Plot` 中，由 `bearFruit` 和 `bearWeed` 方法更新
- [observer.md](observer.md) — `Observer` 接口的 `OnProgress` 回调依赖 `Counter` 提供的完成数和总数

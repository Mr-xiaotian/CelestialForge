# grow.Counter

> 源文件: `pkg/grow/counter.go`

## 概述

`counter.go` 实现了一个线程安全的种子计数器，用于跟踪 `Plot` 中种子的总数、成功数和失败数。内部使用 `sync/atomic` 包的原子操作保证在高并发场景下计数的正确性，支持多个 tend 协程同时更新计数而无需额外加锁。

## 类型

### `Counter`

并发安全的种子进度计数器。嵌入到 `Plot` 中，供 tend 协程在种子培育完成时更新计数。

| 字段      | 类型           | 说明                                 |
| --------- | -------------- | ------------------------------------ |
| `total`   | `atomic.Int64` | 种子总数，由 `AddTotal` 累加         |
| `success` | `atomic.Int64` | 成功培育的种子数（原子计数）         |
| `failed`  | `atomic.Int64` | 失败的种子数（原子计数）             |

### 构造函数

#### `NewCounter() *Counter`

创建并返回一个新的 `Counter` 实例，所有计数初始为零。

### 方法

#### `AddTotal(addNNum int)`

原子地将种子总数增加 `addNNum`。在 `seed()` 或 `Seed()` 方法中调用。

#### `AddSuccess(addNNum int)`

原子地将成功计数增加 `addNNum`。每当一个 tend 成功培育种子后调用。

#### `AddFailed(addNNum int)`

原子地将失败计数增加 `addNNum`。每当一个 tend 培育种子出错后调用。

#### `GetTotal() int`

返回种子总数。

#### `GetSuccess() int`

返回当前成功培育的种子数。

#### `GetFailed() int`

返回当前失败的种子数。

#### `GetCompleted() int`

返回已完成的种子总数（`success + failed`），不区分成功与失败。

#### `IsFinish() bool`

判断所有种子是否已全部完成，即 `total > 0 && GetCompleted() == GetTotal()`。`Plot` 的 `sprout` 方法用此判断是否可以结束调度。

## 使用示例

```go
counter := grow.NewCounter()
counter.AddTotal(100)

// 在 tend 协程中
counter.AddSuccess(1)
// 或
counter.AddFailed(1)

// 检查进度
fmt.Printf("Progress: %d/%d\n", counter.GetCompleted(), counter.GetTotal())

if counter.IsFinish() {
    fmt.Println("All seeds completed")
}
```

## 关联文件

- [plot.md](plot.md) — `Counter` 作为嵌入字段存在于 `Plot` 中，由 `processFruit` 和 `handleWeed` 方法更新
- [observer.md](observer.md) — `Observer` 接口的 `OnProgress` 回调依赖 `Counter` 提供的完成数和总数

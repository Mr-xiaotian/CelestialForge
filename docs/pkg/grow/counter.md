# grow.Counter

> 源文件: `pkg/grow/counter.go`

## 概述

`counter.go` 实现了一个线程安全的任务计数器，用于跟踪 `Executor` 中任务的总数、成功数和失败数。内部使用 `sync/atomic` 包的原子操作保证在高并发场景下计数的正确性，支持多个 worker 协程同时更新计数而无需额外加锁。

## 类型

### `Counter`

并发安全的任务进度计数器。嵌入到 `Executor` 中，供 worker 协程在任务完成时更新计数。

| 字段      | 类型           | 说明                               |
| --------- | -------------- | ---------------------------------- |
| `total`   | `int`          | 任务总数，由 `SetTotal` 设定       |
| `success` | `atomic.Int64` | 成功完成的任务数（原子计数）       |
| `failed`  | `atomic.Int64` | 失败的任务数（原子计数）           |

### 构造函数

#### `NewCounter() *Counter`

创建并返回一个新的 `Counter` 实例，所有计数初始为零。

### 方法

#### `SetTotal(total int)`

设置任务总数。通常在 `Executor` 启动时根据输入任务列表的长度调用一次。

#### `AddSuccess(addNNum int)`

原子地将成功计数增加 `addNNum`。每当一个 worker 成功处理任务后调用。

#### `AddFailed(addNNum int)`

原子地将失败计数增加 `addNNum`。每当一个 worker 处理任务出错后调用。

#### `GetTotal() int`

返回任务总数。

#### `GetSuccess() int`

返回当前成功完成的任务数。

#### `GetFailed() int`

返回当前失败的任务数。

#### `GetCompleted() int`

返回已完成的任务总数（`success + failed`），不区分成功与失败。

#### `IsFinish() bool`

判断所有任务是否已全部完成，即 `GetCompleted() == GetTotal()`。`Executor` 用此方法判断是否可以结束执行流程。

## 使用示例

```go
counter := grow.NewCounter()
counter.SetTotal(100)

// 在 worker 协程中
counter.AddSuccess(1)
// 或
counter.AddFailed(1)

// 检查进度
fmt.Printf("Progress: %d/%d\n", counter.GetCompleted(), counter.GetTotal())

if counter.IsFinish() {
    fmt.Println("All tasks completed")
}
```

## 关联文件

- [executor.md](executor.md) — `Counter` 作为嵌入字段存在于 `Executor` 中，由 `processTaskSuccess` 和 `handleTaskError` 方法更新
- [observer.md](observer.md) — `Observer` 接口的 `OnProgress` 回调依赖 `Counter` 提供的完成数和总数

# grow.Type

> 源文件: `pkg/grow/type.go`

## 概述

`type.go` 定义了 `grow` 包中核心的泛型数据类型，用于在并发执行管道中传递任务载荷和封装任务结果。这两个类型是 `Executor` 内部任务调度与结果收集机制的基础数据结构。

## 类型

### `Payload[V any]`

任务载荷结构体，用于在 `Executor` 的内部通道中包装和传递任务数据。

| 字段    | 类型  | 说明                                                         |
| ------- | ----- | ------------------------------------------------------------ |
| `ID`    | `int` | 任务的唯一标识符，用于追踪任务在管道中的位置                 |
| `Value` | `V`   | 泛型任务值，承载实际需要处理的数据                           |
| `Prev`  | `any` | 上一阶段的处理结果，用于支持多阶段管道中的数据传递（链式执行场景） |

### `TaskResult[T any, R any]`

任务结果结构体，将原始任务与其处理结果绑定在一起，便于调用方在获取结果时追溯对应的输入任务。

| 字段     | 类型 | 说明                         |
| -------- | ---- | ---------------------------- |
| `Task`   | `T`  | 原始任务值                   |
| `Result` | `R`  | 任务处理后产出的结果值       |

## 使用示例

```go
// 构造一个任务载荷
payload := grow.Payload[string]{
    ID:    1,
    Value: "https://example.com",
    Prev:  nil,
}

// 从 Executor.Start 获取结果
results := executor.Start(tasks)
for _, r := range results {
    fmt.Printf("Task: %v -> Result: %v\n", r.Task, r.Result)
}
```

## 关联文件

- [executor.md](executor.md) — `Executor` 在内部通道中使用 `Payload` 包装任务，`Start` 和 `Collect` 方法返回 `TaskResult`
- [counter.md](counter.md) — `Counter` 通过 `Payload.ID` 追踪任务完成状态

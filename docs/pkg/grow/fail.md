# grow.Fail

> 源文件: `pkg/grow/fail.go`

## 概述

`fail.go` 实现了 `Executor` 的失败任务记录系统，基于 `funnel` 包的生产者-消费者模式将失败任务异步写入 JSONL 文件。每条失败记录包含时间戳、Executor 名称、任务 ID、原始任务数据和错误信息，便于事后排查和任务重试。失败记录按日期分目录存储在 `fallback/` 下。

## 类型

### `FailRecord[T any]`

单条失败任务记录，使用泛型 `T` 保留原始任务的完整类型信息。

| 字段           | 类型     | 说明                                 |
| -------------- | -------- | ------------------------------------ |
| `FormatTime`   | `string` | 格式化的时间戳                       |
| `ExecutorName` | `string` | 产生失败的 Executor 名称             |
| `TaskID`       | `int`    | 任务 ID（对应 `Payload.ID`）         |
| `TaskValue`    | `T`      | 原始任务数据，保留完整类型信息       |
| `ErrorMessage` | `string` | 错误信息                             |

### `FailRecordHandler[T any]`

失败记录的消费端处理器，实现了 `funnel.RecordHandler[FailRecord[T]]` 接口，负责将失败记录序列化为 JSON 后写入文件。

| 字段       | 类型       | 说明                   |
| ---------- | ---------- | ---------------------- |
| `FailPath` | `string`   | 失败记录文件路径       |
| `FailFile` | `*os.File` | 失败记录文件句柄       |

#### 方法

- **`BeforeStart()`** — 创建 `fallback/{date}/` 目录（按日期分目录），打开 JSONL 文件以追加模式写入
- **`HandleRecord(record FailRecord[T])`** — 将 `FailRecord` 序列化为 JSON 并写入一行（JSONL 格式）
- **`AfterStop()`** — 关闭失败记录文件句柄

### `FailInlet[T any]`

失败记录的生产端入口，封装了 `funnel.Inlet[FailRecord[T]]`。

| 字段    | 类型                              | 说明                                       |
| ------- | --------------------------------- | ------------------------------------------ |
| `Inlet` | `funnel.Inlet[FailRecord[T]]`    | 嵌入的 funnel 入口，负责向通道发送记录     |

#### 构造函数

##### `NewFailInlet[T any](ch chan<- FailRecord[T], timeout time.Duration) *FailInlet[T]`

创建失败记录入口。`timeout` 为向通道发送记录时的超时时间。

#### 方法

##### `TaskError(executorName string, taskID int, task T, err error)`

记录一个失败任务。构造 `FailRecord` 并通过 `Inlet` 异步发送到消费端。参数包括 Executor 名称、任务 ID、原始任务数据和错误对象。

## 使用示例

```go
// 失败记录系统由 Executor 内部自动初始化和管理
// 以下展示其内部工作原理

ch := make(chan grow.FailRecord[string], 100)

// 消费端
handler := &grow.FailRecordHandler[string]{
    FailPath: "fallback/2024-01-15/file-hasher.jsonl",
}
spout := funnel.NewSpout(ch, handler)

// 生产端
inlet := grow.NewFailInlet(ch, 5*time.Second)

// 记录失败任务
inlet.TaskError("file-hasher", 42, "/path/to/file.dat", errors.New("read timeout"))
```

**JSONL 输出示例**:

```json
{"FormatTime":"2024-01-15 10:30:45","ExecutorName":"file-hasher","TaskID":42,"TaskValue":"/path/to/file.dat","ErrorMessage":"read timeout"}
```

## 关联文件

- [executor.md](executor.md) — `Executor` 在 `handleTaskError` 中调用 `FailInlet.TaskError` 记录失败任务
- [type.md](type.md) — `FailRecord.TaskID` 对应 `Payload.ID`
- [../../funnel/inlet.md](../../funnel/inlet.md) — `FailInlet` 嵌入了 `funnel.Inlet`，用于异步发送失败记录
- [../../funnel/spout.md](../../funnel/spout.md) — `FailRecordHandler` 实现了 `funnel.RecordHandler` 接口，由 `funnel.Spout` 驱动消费

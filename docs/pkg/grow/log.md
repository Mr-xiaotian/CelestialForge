# grow.Log

> 源文件: `pkg/grow/log.go`

## 概述

`log.go` 实现了 `Executor` 的结构化日志系统，基于 `funnel` 包的生产者-消费者模式异步写入日志文件。日志系统支持多级别过滤（TRACE/DEBUG/INFO/SUCCESS/WARNING/ERROR），并提供针对 `Executor` 生命周期事件的便捷日志方法。日志通过 `LogInlet`（生产端）写入通道，由 `LogRecordHandler`（消费端）格式化后持久化到磁盘。

## 常量

### `levelOrder`

日志级别优先级映射表，数值越小优先级越低：

| 级别      | 值  |
| --------- | --- |
| `TRACE`   | 0   |
| `DEBUG`   | 1   |
| `INFO`    | 2   |
| `SUCCESS` | 3   |
| `WARNING` | 4   |
| `ERROR`   | 5   |

## 类型

### `LogRecord`

单条日志记录。

| 字段         | 类型     | 说明                     |
| ------------ | -------- | ------------------------ |
| `FormatTime` | `string` | 格式化的时间戳           |
| `Level`      | `string` | 日志级别（如 `"INFO"`）  |
| `Message`    | `string` | 日志消息内容             |

### `LogRecordHandler`

日志记录的消费端处理器，实现了 `funnel.RecordHandler[LogRecord]` 接口，负责将日志写入文件。

| 字段      | 类型      | 说明               |
| --------- | --------- | ------------------ |
| `LogPath` | `string`  | 日志文件路径       |
| `logFile` | `*os.File`| 日志文件句柄       |

#### 方法

- **`BeforeStart()`** — 创建 `logs/` 目录（如不存在），打开日志文件以追加模式写入
- **`HandleRecord(record LogRecord)`** — 将 `LogRecord` 格式化为一行文本写入日志文件
- **`AfterStop()`** — 关闭日志文件句柄，释放资源

### `LogInlet`

日志的生产端入口，封装了 `funnel.Inlet[LogRecord]`，提供面向 `Executor` 的高层日志方法。

| 字段       | 类型                      | 说明                                       |
| ---------- | ------------------------- | ------------------------------------------ |
| `Inlet`    | `funnel.Inlet[LogRecord]` | 嵌入的 funnel 入口，负责向通道发送记录     |
| `minLevel` | `int`                     | 最低日志级别，低于此级别的日志将被过滤     |

#### 构造函数

##### `NewLogInlet(ch chan<- LogRecord, timeout time.Duration, level string) *LogInlet`

创建日志入口。`level` 参数指定最低日志级别（如 `"INFO"` 将过滤掉 TRACE 和 DEBUG）。`timeout` 为向通道发送记录时的超时时间。

#### 方法

##### `log(level, message string)`

内部日志方法。检查 `level` 是否达到 `minLevel`，如果达到则构造 `LogRecord` 并通过 `Inlet` 发送到通道。

##### `StartExecutor(name string, n int)`

记录 Executor 启动事件。级别为 INFO，包含 Executor 名称和 tend 数量。

##### `EndExecutor(name string, useTime float64, success, failed int)`

记录 Executor 结束事件。级别为 INFO，包含耗时、成功数和失败数统计。

##### `TaskSuccess(name, taskRepr, resultRepr string, useTime float64)`

记录单个任务成功完成。级别为 SUCCESS，包含任务和结果的字符串表示（经 `trunc` 截断）。

##### `TaskError(name, taskRepr string, err error)`

记录单个任务执行失败。级别为 ERROR，包含任务的字符串表示和错误信息。

## 使用示例

```go
// 日志系统由 Executor 内部自动初始化和管理
// 以下展示其内部工作原理

ch := make(chan grow.LogRecord, 100)

// 消费端
handler := &grow.LogRecordHandler{LogPath: "logs/executor.log"}
spout := funnel.NewSpout(ch, handler)

// 生产端
inlet := grow.NewLogInlet(ch, 5*time.Second, "INFO")

// Executor 生命周期日志
inlet.StartExecutor("file-hasher", 8)
inlet.TaskSuccess("file-hasher", "file_a.txt", "abc123...", 0.35)
inlet.TaskError("file-hasher", "file_b.txt", errors.New("permission denied"))
inlet.EndExecutor("file-hasher", 12.5, 99, 1)
```

## 关联文件

- [executor.md](executor.md) — `Executor` 在内部创建日志通道、`LogInlet` 和 `LogRecordHandler`，并在任务处理过程中调用日志方法
- [helper.md](helper.md) — `trunc` 函数用于截断 `taskRepr` 和 `resultRepr`
- [../../funnel/inlet.md](../../funnel/inlet.md) — `LogInlet` 嵌入了 `funnel.Inlet`，用于异步发送日志记录
- [../../funnel/spout.md](../../funnel/spout.md) — `LogRecordHandler` 实现了 `funnel.RecordHandler` 接口，由 `funnel.Spout` 驱动消费

# grow.Log

> 源文件: `pkg/grow/log.go`

## 概述

`log.go` 实现了 `Plot` 和 `Farm` 的结构化日志系统，基于 `funnel` 包的生产者-消费者模式异步写入日志文件。日志系统支持多级别过滤（TRACE/DEBUG/SUCCESS/INFO/WARNING/ERROR/CRITICAL），并提供针对 Plot 和 Farm 生命周期事件的便捷日志方法。

## 常量

### `levelOrder`

日志级别优先级映射表，数值越小优先级越低：

| 级别 | 值 |
|------|------|
| `TRACE` | 0 |
| `DEBUG` | 1 |
| `SUCCESS` | 2 |
| `INFO` | 3 |
| `WARNING` | 4 |
| `ERROR` | 5 |
| `CRITICAL` | 6 |

## 类型

### `LogRecord`

单条日志记录。

| 字段 | 类型 | 说明 |
|------|------|------|
| `FormatTime` | `string` | 格式化的时间戳 |
| `Level` | `string` | 日志级别 |
| `Message` | `string` | 日志消息内容 |

### `LogRecordHandler`

日志记录的消费端处理器，实现了 `funnel.RecordHandler[LogRecord]` 接口，负责将日志写入文件。

| 字段 | 类型 | 说明 |
|------|------|------|
| `LogPath` | `string` | 日志文件路径 |
| `logFile` | `*os.File` | 日志文件句柄 |

#### 方法

- **`BeforeStart()`** — 创建 `logs/` 目录，打开日志文件（按日期命名，追加模式）
- **`HandleRecord(record LogRecord)`** — 将 `LogRecord` 格式化为一行文本写入日志文件
- **`AfterStop()`** — 关闭日志文件句柄

### `LogInlet`

日志的生产端入口，封装了 `funnel.Inlet[LogRecord]`，提供面向 Plot/Farm 的高层日志方法。

| 字段 | 类型 | 说明 |
|------|------|------|
| `Inlet` | `funnel.Inlet[LogRecord]` | 嵌入的 funnel 入口 |
| `minLevel` | `int` | 最低日志级别，低于此级别的日志将被过滤 |

#### 构造函数

##### `NewLogInlet(ch chan<- LogRecord, timeout time.Duration, level string) *LogInlet`

创建日志入口。`level` 参数指定最低日志级别，不存在则默认 INFO。

## 日志方法

### `log(level, message string)`

内部日志方法。低于 `minLevel` 的日志会被静默丢弃。

### `StartFarm(farmName string)`

记录 Farm 启动。级别为 INFO。

### `EndFarm(farmName string, useTime float64)`

记录 Farm 结束，包含总耗时。级别为 INFO。

### `StartPlot(plotName string, numTends int)`

记录 Plot 启动，包含 tend 数量。级别为 INFO。

### `EndPlot(plotName string, useTime float64, fruitNum, weedNum int)`

记录 Plot 结束，包含耗时、成功数和失败数。级别为 INFO。

### `SeedRipen(plotName, seedRepr, fruitRepr string, useTime float64)`

记录种子成熟（培育成功），包含种子和果实的字符串表示及耗时。级别为 SUCCESS。

### `SeedWither(plotName, seedRepr string, err error, startTime time.Time)`

记录种子枯萎（培育失败），包含错误信息和耗时。级别为 ERROR。

### `SeedReplant(plotName, seedRepr string, attempt int, err error)`

记录种子重新种植（重试），包含当前尝试次数和错误信息。级别为 WARNING。

## 使用示例

```go
ch := make(chan grow.LogRecord, 100)

handler := &grow.LogRecordHandler{}
spout := funnel.NewSpout(ch, handler)

inlet := grow.NewLogInlet(ch, 5*time.Second, "INFO")

inlet.StartPlot("file-hasher", 8)
inlet.SeedRipen("file-hasher", "file_a.txt", "abc123...", 0.35)
inlet.SeedWither("file-hasher", "file_b.txt", errors.New("permission denied"), time.Now())
inlet.EndPlot("file-hasher", 12.5, 99, 1)
```

## 关联文件

- [plot.md](plot.md) — `Plot` 在内部创建日志通道、`LogInlet` 和 `LogRecordHandler`
- [farm.md](farm.md) — `Farm` 共享一套日志 spout 给所有 Plot
- [helper.md](helper.md) — `trunc` 函数用于截断 `seedRepr` 和 `fruitRepr`

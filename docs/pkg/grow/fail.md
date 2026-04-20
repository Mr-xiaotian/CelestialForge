# grow.Fail

> 源文件: `pkg/grow/fail.go`

## 概述

`fail.go` 实现了 `Plot` 的失败种子记录系统，基于 `funnel` 包的生产者-消费者模式将失败种子异步写入 JSONL 文件。每条失败记录包含时间戳、Plot 名称、种子字符串表示和错误信息，便于事后排查和种子重试。失败记录按日期分目录存储在 `fallback/` 下。

## 类型

### `FailRecord`

单条失败种子记录。

| 字段 | 类型 | 说明 |
|------|------|------|
| `FormatTime` | `string` | 格式化的时间戳 |
| `PlotName` | `string` | 产生失败的 Plot 名称 |
| `SeedID` | `int` | 种子 ID |
| `SeedString` | `string` | 种子的字符串表示 |
| `ErrorMessage` | `string` | 错误信息 |

### `FailRecordHandler`

失败记录的消费端处理器，实现了 `funnel.RecordHandler[FailRecord]` 接口，负责将失败记录序列化为 JSON 后写入文件。

| 字段 | 类型 | 说明 |
|------|------|------|
| `FailPath` | `string` | 失败记录文件路径 |
| `FailFile` | `*os.File` | 失败记录文件句柄 |

#### 方法

- **`BeforeStart()`** — 创建 `fallback/{date}/` 目录，打开 JSONL 文件（追加模式）
- **`HandleRecord(record FailRecord)`** — 将 `FailRecord` 序列化为 JSON 并写入一行
- **`AfterStop()`** — 关闭 JSONL 文件句柄

### `FailInlet`

失败记录的生产端入口，封装了 `funnel.Inlet[FailRecord]`。

| 字段 | 类型 | 说明 |
|------|------|------|
| `Inlet` | `funnel.Inlet[FailRecord]` | 嵌入的 funnel 入口 |

#### 构造函数

##### `NewFailInlet(ch chan<- FailRecord, timeout time.Duration) *FailInlet`

创建失败记录入口。`timeout` 为向通道发送记录时的超时时间。

## 失败记录方法

### `SeedWither(plotName string, seed any, err error)`

记录一颗失败的种子。将 `seed` 格式化为字符串后构造 `FailRecord` 并通过 `Inlet` 异步发送。

## 使用示例

```go
ch := make(chan grow.FailRecord, 100)

handler := &grow.FailRecordHandler{}
spout := funnel.NewSpout(ch, handler)

inlet := grow.NewFailInlet(ch, 5*time.Second)

inlet.SeedWither("file-hasher", "/path/to/file.dat", errors.New("read timeout"))
```

**JSONL 输出示例**:

```json
{"FormatTime":"2024-01-15 10:30:45","PlotName":"file-hasher","SeedID":0,"SeedString":"/path/to/file.dat","ErrorMessage":"read timeout"}
```

## 关联文件

- [plot.md](plot.md) — `Plot` 在 `bearWeed` 中调用 `FailInlet.SeedWither` 记录失败种子
- [farm.md](farm.md) — `Farm` 共享一套失败记录 spout 给所有 Plot

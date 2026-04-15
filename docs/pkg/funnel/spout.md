# funnel.Spout

> 源文件: `pkg/funnel/spout.go`

## 概述

`Spout` 是 funnel 包中的消费端（Consumer），负责从通道中读取记录并交给 `RecordHandler` 处理。它管理着一个内部缓冲通道和消费循环，支持优雅启动与停止，并通过 `sync.WaitGroup` 确保所有记录在关闭前被完整处理。

该设计源自 Python 项目中的 BaseListener 模式，在 Go 中使用泛型实现，支持任意记录类型。`Spout` 与 `Inlet` 构成生产者-消费者对，通过共享通道进行通信。

## 类型/函数

### `RecordHandler[T any]`

定义记录处理的生命周期接口。任何实现此接口的类型都可以作为 `Spout` 的处理器。

**方法：**

| 方法                           | 说明                                       |
|--------------------------------|--------------------------------------------|
| `BeforeStart() error`         | 在消费循环启动前调用，用于初始化资源       |
| `HandleRecord(record T) error`| 处理单条记录，由消费循环逐条调用           |
| `AfterStop() error`           | 在消费循环结束后调用，用于清理和释放资源   |

### `Spout[T any]`

泛型消费端结构体，从通道读取记录并委托给 `RecordHandler` 处理。

**字段（未导出）：**

| 字段      | 类型                    | 说明                                   |
|-----------|-------------------------|----------------------------------------|
| `ch`      | `chan T`                | 内部缓冲通道，生产端通过此通道写入记录 |
| `ctx`     | `context.Context`       | 用于取消消费循环的上下文               |
| `cancel`  | `context.CancelFunc`    | 取消函数，由 `Stop` 调用               |
| `wg`      | `sync.WaitGroup`        | 用于等待消费循环退出                   |
| `timeout` | `time.Duration`         | 超时控制                               |
| `handler` | `RecordHandler[T]`      | 记录处理器实例                         |

### `NewSpout[T any](handler RecordHandler[T], bufferSize int, timeout time.Duration) *Spout[T]`

创建一个新的 `Spout` 实例。内部会分配一个指定大小的缓冲通道，并创建可取消的上下文。

**参数：**

- `handler` — 实现 `RecordHandler[T]` 接口的处理器，负责记录的实际处理逻辑。
- `bufferSize` — 内部通道的缓冲大小，控制生产端在消费端处理速度不足时可以积压的记录数。
- `timeout` — 超时控制时间。

### `(*Spout[T]) GetQueue() chan<- T`

返回内部通道的单向写引用。该返回值通常传递给 `NewInlet` 用于创建对应的生产端。

### `(*Spout[T]) Start() error`

启动消费循环。执行流程：

1. 调用 `handler.BeforeStart()` 进行初始化。
2. 启动一个 goroutine 运行内部消费循环 `spout()`。
3. 如果 `BeforeStart` 返回错误，`Start` 将直接返回该错误而不启动循环。

### `(*Spout[T]) spout()`

内部消费循环（未导出）。持续从通道读取记录并调用 `handler.HandleRecord` 处理，直到上下文被取消或通道被关闭。

### `(*Spout[T]) Stop() error`

优雅停止消费端。执行流程：

1. 取消上下文，通知消费循环退出。
2. 通过 `wg.Wait()` 等待消费循环完成所有已读取记录的处理。
3. 调用 `handler.AfterStop()` 进行资源清理。

## 使用示例

```go
package main

import (
    "fmt"
    "time"

    "your_module/pkg/funnel"
)

// 实现 RecordHandler 接口
type MyHandler struct{}

func (h *MyHandler) BeforeStart() error {
    fmt.Println("handler initialized")
    return nil
}

func (h *MyHandler) HandleRecord(record string) error {
    fmt.Println("processing:", record)
    return nil
}

func (h *MyHandler) AfterStop() error {
    fmt.Println("handler cleaned up")
    return nil
}

func main() {
    handler := &MyHandler{}
    spout := funnel.NewSpout[string](handler, 100, 5*time.Second)

    // 启动消费端
    if err := spout.Start(); err != nil {
        panic(err)
    }

    // 创建生产端并绑定到消费端的通道
    inlet := funnel.NewInlet(spout.GetQueue(), 5*time.Second)

    // 发送记录
    inlet.Send("record-1")
    inlet.Send("record-2")

    // 关闭生产端，停止消费端
    inlet.Close()
    spout.Stop()
}
```

## 关联文件

- [inlet.md](inlet.md) — `Inlet` 是对应的生产端，通过 `Spout.GetQueue()` 返回的通道向 `Spout` 发送记录
- `pkg/grow/log.go` — `LogRecordHandler` 实现了 `RecordHandler` 接口，用于处理日志记录
- `pkg/grow/fail.go` — `FailRecordHandler` 实现了 `RecordHandler` 接口，用于处理失败记录

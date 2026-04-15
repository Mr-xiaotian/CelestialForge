# funnel.Inlet

> 源文件: `pkg/funnel/inlet.go`

## 概述

`Inlet` 是 funnel 包中的生产端（Producer），负责向通道写入记录。它封装了一个单向写通道，并提供上下文取消和超时控制能力，确保在通道阻塞或系统关闭时 `Send` 操作能够安全退出而不会永久挂起。

该设计源自 Python 项目中的 BaseSinker 模式，在 Go 中使用泛型实现，支持任意记录类型。

## 类型/函数

### `Inlet[T any]`

泛型生产端结构体，将记录发送到绑定的通道中。

**字段（未导出）：**

| 字段      | 类型                    | 说明                         |
|-----------|-------------------------|------------------------------|
| `ch`      | `chan<- T`              | 绑定的单向写通道             |
| `ctx`     | `context.Context`       | 用于取消控制的上下文         |
| `cancel`  | `context.CancelFunc`    | 取消函数，由 `Close` 调用    |
| `timeout` | `time.Duration`         | 单次 `Send` 的超时时间       |

### `NewInlet[T any](ch chan<- T, timeout time.Duration) *Inlet[T]`

创建一个新的 `Inlet` 实例，绑定到指定的写通道。内部会创建一个可取消的上下文，用于在调用 `Close` 后使后续的 `Send` 立即返回错误。

**参数：**

- `ch` — 目标写通道，通常由 `Spout.GetQueue()` 获取。
- `timeout` — 每次 `Send` 的最大等待时间，超时后返回错误。

### `(*Inlet[T]) Send(record T) error`

向通道发送一条记录。该方法通过 `select` 同时监听三个条件：

1. 成功写入通道 — 返回 `nil`。
2. 上下文被取消（`Close` 已调用） — 返回 `context.Canceled`。
3. 超时 — 返回 `"inlet send timeout after <duration>"` 错误。

### `(*Inlet[T]) Close()`

关闭 `Inlet`，取消内部上下文。调用后，所有正在阻塞或后续的 `Send` 调用将立即返回上下文取消错误。注意：`Close` 不会关闭底层通道，通道的关闭由 `Spout` 端管理。

## 使用示例

```go
package main

import (
    "fmt"
    "time"

    "your_module/pkg/funnel"
)

func main() {
    // 通常通过 Spout.GetQueue() 获取通道
    ch := make(chan string, 10)

    inlet := funnel.NewInlet(ch, 5*time.Second)
    defer inlet.Close()

    // 发送记录
    if err := inlet.Send("hello"); err != nil {
        fmt.Println("send failed:", err)
    }
}
```

在实际项目中，`Inlet` 通常与 `Spout` 配合使用：

```go
spout := funnel.NewSpout(handler, 100, 5*time.Second)
spout.Start()

inlet := funnel.NewInlet(spout.GetQueue(), 5*time.Second)
inlet.Send(record)

inlet.Close()
spout.Stop()
```

## 关联文件

- [spout.md](spout.md) — `Spout` 是对应的消费端，`Inlet` 通过 `Spout.GetQueue()` 返回的通道与其连接
- `pkg/grow/log.go` — `LogInlet` 使用 `Inlet` 发送日志记录到 `LogRecordHandler`
- `pkg/grow/fail.go` — `FailInlet` 使用 `Inlet` 发送失败记录到 `FailRecordHandler`

# grow.Control

> 源文件: `pkg/grow/control.go`

## 概述

`control.go` 定义了用于 `Plot` 内部控制流的信号类型。`ControlSignal` 通过专用的控制通道在 `Plot` 的各个协程之间传递控制指令，用于协调种子调度的启停等操作。

## 类型

### `ControlSignal`

控制信号结构体，用于在 `Plot` 的 `ControlChan` 通道中传递控制指令。

| 字段     | 类型     | 说明                                                         |
| -------- | -------- | ------------------------------------------------------------ |
| `Source` | `string` | 信号来源标识，说明是哪个组件发出的控制信号（如 `"seed"` 表示种子注入完成） |

## 使用示例

```go
// 在 Plot 内部，seed 完成后发送控制信号
e.ControlChan <- grow.ControlSignal{
    Source: "seed",
}

// sprout 协程接收控制信号后关闭种子分发
signal := <-e.ControlChan
if signal.Source == "seed" {
    // 所有种子已注入，准备关闭调度
}
```

## 关联文件

- [plot.md](plot.md) — `Plot` 的 `seed` 方法发送 `ControlSignal`，`sprout` 方法接收并据此协调工作流程

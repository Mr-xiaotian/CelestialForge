# grow.Type

> 源文件: `pkg/grow/type.go`

## 概述

`type.go` 定义了 `grow` 包中核心的泛型数据类型，用于在并发执行管道中传递种子载荷和封装培育结果。这两个类型是 `Plot` 内部种子调度与果实收集机制的基础数据结构。

## 类型

### `Payload[V any]`

种子载荷结构体，用于在 `Plot` 的内部通道中包装和传递种子数据。

| 字段    | 类型  | 说明                                                         |
| ------- | ----- | ------------------------------------------------------------ |
| `ID`    | `int` | 种子的唯一标识符，用于追踪种子在管道中的位置                 |
| `Value` | `V`   | 泛型值，承载实际需要处理的数据                               |
| `Prev`  | `any` | 上一阶段的处理结果，用于支持多阶段管道中的数据传递（链式执行场景） |

### `Karma[S any, F any]`

种子与果实的配对结构体，将原始种子与其培育结果绑定在一起，便于调用方在获取果实时追溯对应的种子。

| 字段    | 类型 | 说明                   |
| ------- | ---- | ---------------------- |
| `Seed`  | `S`  | 原始种子值             |
| `Fruit` | `F`  | 培育后产出的果实值     |

## 使用示例

```go
// 构造一个种子载荷
payload := grow.Payload[string]{
    ID:    1,
    Value: "https://example.com",
    Prev:  nil,
}

// 从 Plot.Start 获取结果
karmas := plot.Start(seeds)
for _, k := range karmas {
    fmt.Printf("Seed: %v -> Fruit: %v\n", k.Seed, k.Fruit)
}
```

## 关联文件

- [plot.md](plot.md) — `Plot` 在内部通道中使用 `Payload` 包装种子，`Start` 方法返回 `Karma`
- [counter.md](counter.md) — `Counter` 通过 `Payload.ID` 追踪种子完成状态

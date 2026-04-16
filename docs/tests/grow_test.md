# grow 包测试

> 源文件: `pkg/grow/plot_sync_test.go`, `pkg/grow/plot_async_test.go`, `pkg/grow/plot_retry_test.go`

## 概述

`grow.Plot` 的功能测试套件，采用黑盒测试（`package grow_test`）。测试按功能拆分为三个文件，覆盖同步执行、异步执行和重试机制。

## 测试文件

### `plot_sync_test.go`

同步模式（`Start`）下的基本功能测试。

| 测试函数 | 说明 |
|---------|------|
| `TestPlot_AllError` | 全部种子失败：验证无果实返回，状态码为 2 |
| `TestPlot_PartialError` | 部分种子失败：偶数种子失败，验证 3 个成功 |
| `TestPlot_AllSuccess` | 全部种子成功：验证每个果实 = 种子 × 2 |

### `plot_async_test.go`

异步模式（`StartAsync` + `Seed` + `Seal` + `Harvest` + `WaitAsync`）的测试。

| 测试函数 | 说明 |
|---------|------|
| `TestPlot_Async` | 异步流程：逐个注入种子、密封、收获果实，验证 5 个结果 |

### `plot_retry_test.go`

重试机制的测试。

| 测试函数 | 说明 |
|---------|------|
| `TestPlot_RetrySuccess` | 前 2 次失败第 3 次成功，`MaxRetries(3)` |
| `TestPlot_RetryExhausted` | 重试耗尽仍失败，`MaxRetries(2)` → 共 3 次尝试 |
| `TestPlot_RetryIf` | `RetryIf` 过滤永久性错误，仅执行 1 次 |
| `TestPlot_RetryDelay` | 验证重试间隔（100ms）被正确应用 |

## 使用示例

```bash
# 运行所有 grow 包测试
go test -v ./pkg/grow/

# 运行重试相关测试
go test -v -run TestPlot_Retry ./pkg/grow/
```

## 关联文件

- [../pkg/grow/plot.md](../pkg/grow/plot.md) — Plot 核心实现
- [../pkg/grow/option.md](../pkg/grow/option.md) — Option 配置（WithMaxRetries 等）

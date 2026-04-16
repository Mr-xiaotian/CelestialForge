# cmd/duplicate.main

> 源文件: `cmd/duplicate/main.go`

## 概述

这是一个正式的命令行工具，用于扫描指定目录下的重复文件并生成报告。通过命令行参数控制扫描路径、输出文件路径和并发工作线程数。底层调用 `pkg/file` 包的 `ScanDuplicateFile` 函数执行实际的重复文件检测，并将结果格式化输出到指定的报告文件中。

## 类型/函数

### `main()`

程序入口，解析命令行参数并执行重复文件扫描流程：

1. 通过 `flag` 包解析三个命令行参数
2. 调用 `file.ScanDuplicateFile` 执行扫描
3. 将扫描结果写入指定的输出文件

**命令行参数：**

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `--scan-path` | string | `""` (必填) | 要扫描的目录路径 |
| `--output-path` | string | `duplicate_report.txt` | 报告输出文件路径 |
| `--num-tends` | int | `4` | 并发工作线程数 |

## 使用示例

```bash
# 扫描指定目录，使用默认输出路径和工作线程数
go run cmd/duplicate/main.go --scan-path /path/to/scan

# 指定所有参数
go run cmd/duplicate/main.go \
  --scan-path /path/to/scan \
  --output-path result.txt \
  --num-tends 8

# 编译后运行
go build -o duplicate cmd/duplicate/main.go
./duplicate --scan-path /data/photos --num-tends 16
```

## 关联文件

- [../../tests/file_test.md](../../tests/file_test.md) — 包含 TestDuplicateReport 和 TestDuplicateReportEmpty 测试
- [../debug/main.md](../debug/main.md) — 调试入口中的 debug_duplicate 提供类似功能

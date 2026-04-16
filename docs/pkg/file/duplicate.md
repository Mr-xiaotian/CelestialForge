# file.Duplicate

> 源文件: `pkg/file/duplicate.go`

## 概述

本文件实现了高效的重复文件扫描功能，采用三级过滤流水线逐步缩小候选集，避免对所有文件进行完整哈希计算的高昂开销。流水线的三个阶段分别为：按文件大小分组（零IO开销）、按前 4KB 快照哈希过滤（低IO开销）、按完整文件哈希确认（仅对少量候选文件执行）。第二和第三阶段通过 `grow.NewPlot` 进行并行计算，并配合 `grow.NewProgressBar` 显示进度。最终结果以 `map[FileInfo][]string` 形式返回，键为包含哈希和大小的文件信息，值为具有相同内容的文件路径列表。还提供 `DuplicateReport` 函数生成格式化的重复文件报告。

## 类型/函数

### `getSizeDuplicate`（内部）

```go
func getSizeDuplicate(fileInfoMap FileInfoMap) []string
```

流水线第一阶段：按文件大小分组。将所有文件按 `Size` 分组，提取出大小相同且大小大于 0 的文件路径列表。此步骤无需任何额外 IO 操作。

### `getSnapshotDuplicate`（内部）

```go
func getSnapshotDuplicate(fileSizeDuplicates []string, numTends int) ([]string, error)
```

流水线第二阶段：按快照哈希过滤。使用 `grow.NewPlot` 并行调用 `GetFileSnapshotSHA1` 计算每个候选文件的前 4KB 哈希值，筛选出快照哈希相同的文件。

### `getHashDuplicate`（内部）

```go
func getHashDuplicate(fileSnapshotDuplicates []string, fileInfoMap FileInfoMap, numTends int) (map[FileInfo][]string, error)
```

流水线第三阶段：按完整哈希确认。使用 `grow.NewPlot` 并行调用 `GetFileSHA1` 计算候选文件的完整哈希值，最终确认真正的重复文件。返回结果中 `FileInfo` 包含 `Hash` 和 `Size` 字段。

### `ScanDuplicateFile`

```go
func ScanDuplicateFile(path string, numTends int) (map[FileInfo][]string, error)
```

公开 API，扫描指定目录下的重复文件。

**参数：**

| 参数 | 类型 | 说明 |
|------|------|------|
| `path` | `string` | 要扫描的目录路径 |
| `numTends` | `int` | 并行工作协程数量 |

**返回值：**

| 返回值 | 类型 | 说明 |
|--------|------|------|
| 重复文件映射 | `map[FileInfo][]string` | 键为文件信息（含哈希和大小），值为重复文件路径列表 |
| `error` | `error` | 处理过程中的错误 |

**执行流程：**
1. 调用 `GetFilesInfoRecursive` 获取所有文件信息
2. 调用 `getSizeDuplicate` 按大小初筛
3. 调用 `getSnapshotDuplicate` 按快照哈希二次筛选
4. 调用 `getHashDuplicate` 按完整哈希最终确认

### `DuplicateReport`

```go
func DuplicateReport(identicalMap map[FileInfo][]string) string
```

生成重复文件的详细报告文本。

**功能：**
- 按总占用大小（单个文件大小 x 重复数量）降序排列
- 使用 `str.FormatTable` 格式化每组重复文件的路径和大小
- 统计并输出总重复大小、总重复文件数和最多重复的文件组信息

## 使用示例

```go
package main

import (
    "fmt"
    "log"
    "github.com/Mr-xiaotian/CelestialForge/pkg/file"
)

func main() {
    duplicates, err := file.ScanDuplicateFile("/data/photos", 8)
    if err != nil {
        log.Fatalf("扫描失败: %v", err)
    }

    report := file.DuplicateReport(duplicates)
    fmt.Println(report)
}
```

## 关联文件

- [api.md](api.md) -- 使用 `FileInfo` 和 `FileInfoMap` 数据类型
- [info.md](info.md) -- 调用 `GetFilesInfoRecursive` 获取文件元信息
- [hash.md](hash.md) -- 使用 `GetFileSnapshotSHA1` 和 `GetFileSHA1` 计算哈希
- [../../str/table.md](../../str/table.md) -- 使用 `str.FormatTable` 格式化报告表格

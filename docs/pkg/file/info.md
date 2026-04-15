# file.Info

> 源文件: `pkg/file/info.go`

## 概述

本文件提供递归遍历目录并收集文件元信息的功能。核心函数 `GetFilesInfoRecursive` 使用 `filepath.WalkDir` 递归遍历指定目录下的所有文件（跳过目录本身），为每个文件创建 `FileInfo` 记录（包含大小和修改时间），并以完整路径为键存入 `FileInfoMap`。该函数在调用前会验证目录是否存在，若目录不存在则返回带有中文提示的错误信息。此函数是重复文件扫描（`ScanDuplicateFile`）的第一步，为后续的大小分组、哈希比对提供基础数据。

## 类型/函数

### `GetFilesInfoRecursive`

```go
func GetFilesInfoRecursive(root string) (FileInfoMap, error)
```

递归获取指定目录下所有文件的元信息。

**参数：**

| 参数 | 类型 | 说明 |
|------|------|------|
| `root` | `string` | 要遍历的根目录路径 |

**返回值：**

| 返回值 | 类型 | 说明 |
|--------|------|------|
| `FileInfoMap` | `map[string]FileInfo` | 以文件完整路径为键的文件信息映射 |
| `error` | `error` | 目录不存在或遍历过程中的错误 |

**行为细节：**
- 使用 `os.Stat` 预检查目录是否存在
- 使用 `filepath.WalkDir` 进行递归遍历，性能优于 `filepath.Walk`
- 仅收集文件信息，跳过目录条目
- 每个文件记录其 `Size`（通过 `units.NewHumanBytes` 转换）和 `Mtime`（修改时间）
- 不计算文件哈希值（`Hash` 字段留空），哈希计算在后续流水线阶段执行

## 使用示例

```go
package main

import (
    "fmt"
    "log"
    "github.com/Mr-xiaotian/CelestialForge/pkg/file"
)

func main() {
    files, err := file.GetFilesInfoRecursive("/data/documents")
    if err != nil {
        log.Fatalf("遍历失败: %v", err)
    }

    fmt.Printf("共发现 %d 个文件\n", len(files))
    for path, info := range files {
        fmt.Printf("  %s  大小: %s  修改时间: %s\n", path, info.Size, info.Mtime)
    }
}
```

## 关联文件

- [api.md](api.md) -- 定义了返回类型 `FileInfoMap` 和 `FileInfo`
- [duplicate.md](duplicate.md) -- `ScanDuplicateFile` 调用 `GetFilesInfoRecursive` 作为流水线第一步

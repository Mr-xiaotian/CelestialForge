# file.Api

> 源文件: `pkg/file/api.go`

## 概述

本文件定义了 `file` 包中最核心的数据类型：`FileInfo` 和 `FileInfoMap`。`FileInfo` 是文件元信息的统一表示，封装了文件路径、大小、修改时间和哈希值四个维度的信息。`FileInfoMap` 则是以文件路径为键的映射类型，用于批量管理文件元信息。这两个类型是整个 `file` 包的数据基础，被 `info.go`、`duplicate.go`、`size.go`、`mtime.go` 等模块广泛使用。

## 类型/函数

### `FileInfo`

文件元信息结构体，包含路径、大小、修改时间和哈希值。

| 字段 | 类型 | 说明 |
|------|------|------|
| `Path` | `string` | 文件的完整路径 |
| `Size` | `units.HumanBytes` | 文件大小，使用 `units.HumanBytes` 类型支持人类可读的格式化输出 |
| `Mtime` | `time.Time` | 文件的最后修改时间 |
| `Hash` | `string` | 文件的哈希值（十六进制字符串） |

### `FileInfoMap`

```go
type FileInfoMap map[string]FileInfo
```

文件路径到 `FileInfo` 的映射类型。键为文件的完整路径字符串，值为对应的 `FileInfo` 结构体。该类型由 `GetFilesInfoRecursive` 函数生成，并在重复文件扫描流水线中作为核心数据结构传递。

## 使用示例

```go
package main

import (
    "fmt"
    "time"
    "github.com/Mr-xiaotian/CelestialForge/pkg/file"
    "github.com/Mr-xiaotian/CelestialForge/pkg/units"
)

func main() {
    info := file.FileInfo{
        Path:  "/data/example.txt",
        Size:  units.NewHumanBytes(1024),
        Mtime: time.Now(),
        Hash:  "abc123",
    }

    files := make(file.FileInfoMap)
    files[info.Path] = info

    for path, fi := range files {
        fmt.Printf("路径: %s, 大小: %s\n", path, fi.Size)
    }
}
```

## 关联文件

- [info.md](info.md) -- `GetFilesInfoRecursive` 返回 `FileInfoMap` 类型
- [duplicate.md](duplicate.md) -- 重复文件扫描流水线中大量使用 `FileInfo` 和 `FileInfoMap`
- [size.md](size.md) -- `GetFileSize`/`GetDirSize` 返回 `units.HumanBytes`，与 `FileInfo.Size` 类型一致
- [mtime.md](mtime.md) -- `GetFileMtime`/`GetDirMtime` 返回 `time.Time`，与 `FileInfo.Mtime` 类型一致

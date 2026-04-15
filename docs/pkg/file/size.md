# file.Size

> 源文件: `pkg/file/size.go`

## 概述

本文件提供文件和目录大小的查询功能。`GetFileSize` 用于获取单个文件的大小，会拒绝对目录路径的调用并返回错误。`GetDirSize` 用于递归计算目录下所有文件的总大小，通过 `filepath.WalkDir` 遍历并累加每个文件的字节数。两个函数均返回 `units.HumanBytes` 类型，支持人类可读的格式化输出（如 "1.5 MB"）。

## 类型/函数

### `GetFileSize`

```go
func GetFileSize(path string) (units.HumanBytes, error)
```

返回指定文件的大小。

**参数：**

| 参数 | 类型 | 说明 |
|------|------|------|
| `path` | `string` | 文件路径 |

**返回值：**

| 返回值 | 类型 | 说明 |
|--------|------|------|
| 大小 | `units.HumanBytes` | 文件大小 |
| `error` | `error` | 文件不存在、路径为目录等错误 |

**行为细节：**
- 通过 `os.Stat` 获取文件信息
- 如果路径指向目录，返回错误 `"路径是目录而非文件"`
- 返回值通过 `units.NewHumanBytes` 封装

### `GetDirSize`

```go
func GetDirSize(path string) (units.HumanBytes, error)
```

递归计算目录下所有文件的总大小。

**参数：**

| 参数 | 类型 | 说明 |
|------|------|------|
| `path` | `string` | 目录路径 |

**返回值：**

| 返回值 | 类型 | 说明 |
|--------|------|------|
| 大小 | `units.HumanBytes` | 目录下所有文件的总大小 |
| `error` | `error` | 遍历过程中的错误 |

**行为细节：**
- 使用 `filepath.WalkDir` 递归遍历
- 仅累加文件大小，跳过目录条目
- 遍历过程中遇到无法访问的文件会返回错误

## 使用示例

```go
package main

import (
    "fmt"
    "log"
    "github.com/Mr-xiaotian/CelestialForge/pkg/file"
)

func main() {
    // 获取单个文件大小
    size, err := file.GetFileSize("/data/archive.zip")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("文件大小: %s\n", size)

    // 获取目录总大小
    dirSize, err := file.GetDirSize("/data/project")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("目录总大小: %s\n", dirSize)
}
```

## 关联文件

- [api.md](api.md) -- `FileInfo.Size` 字段使用相同的 `units.HumanBytes` 类型

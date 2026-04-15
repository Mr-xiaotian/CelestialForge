# file.Mtime

> 源文件: `pkg/file/mtime.go`

## 概述

本文件提供文件和目录修改时间的查询功能。`GetFileMtime` 用于获取单个文件的最后修改时间，会拒绝对目录路径的调用并返回错误。`GetDirMtime` 用于递归查找目录下所有文件中最晚的修改时间，通过 `filepath.WalkDir` 遍历所有文件并比较修改时间，返回最大值。两个函数均返回 `time.Time` 类型。

## 类型/函数

### `GetFileMtime`

```go
func GetFileMtime(path string) (time.Time, error)
```

返回指定文件的最后修改时间。

**参数：**

| 参数 | 类型 | 说明 |
|------|------|------|
| `path` | `string` | 文件路径 |

**返回值：**

| 返回值 | 类型 | 说明 |
|--------|------|------|
| 修改时间 | `time.Time` | 文件的最后修改时间 |
| `error` | `error` | 文件不存在、路径为目录等错误 |

**行为细节：**
- 通过 `os.Stat` 获取文件信息
- 如果路径指向目录，返回错误 `"路径是目录而非文件"`
- 返回 `info.ModTime()`

### `GetDirMtime`

```go
func GetDirMtime(path string) (time.Time, error)
```

返回目录下所有文件中最晚的修改时间。

**参数：**

| 参数 | 类型 | 说明 |
|------|------|------|
| `path` | `string` | 目录路径 |

**返回值：**

| 返回值 | 类型 | 说明 |
|--------|------|------|
| 修改时间 | `time.Time` | 目录下所有文件中最晚的修改时间 |
| `error` | `error` | 遍历过程中的错误 |

**行为细节：**
- 使用 `filepath.WalkDir` 递归遍历
- 仅检查文件的修改时间，跳过目录条目
- 使用 `time.Time.After` 比较并保留最晚时间
- 如果目录为空（无文件），返回 `time.Time` 的零值

## 使用示例

```go
package main

import (
    "fmt"
    "log"
    "github.com/Mr-xiaotian/CelestialForge/pkg/file"
)

func main() {
    // 获取单个文件的修改时间
    mtime, err := file.GetFileMtime("/data/config.yaml")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("文件修改时间: %s\n", mtime.Format("2006-01-02 15:04:05"))

    // 获取目录下最晚的修改时间
    dirMtime, err := file.GetDirMtime("/data/project")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("目录最晚修改时间: %s\n", dirMtime.Format("2006-01-02 15:04:05"))
}
```

## 关联文件

- [api.md](api.md) -- `FileInfo.Mtime` 字段使用相同的 `time.Time` 类型

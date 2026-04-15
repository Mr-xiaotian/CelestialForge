# file.Hash

> 源文件: `pkg/file/hash.go`

## 概述

本文件实现了完整的文件和目录哈希计算功能，支持 MD5、SHA1、SHA256 三种哈希算法。提供两种文件哈希模式：完整哈希（读取整个文件，使用 1MB 缓冲区）和快照哈希（仅读取文件前 4KB，适用于快速预筛选）。目录哈希通过递归计算所有子项的哈希值，将每个子项编码为 `"D:名称:哈希"` 或 `"F:名称:哈希"` 格式，排序后拼接再计算哈希，支持按目录名和文件扩展名进行排除过滤。该模块是重复文件检测流水线中的核心计算引擎。

## 类型/函数

### `HashType`

```go
type HashType string
```

哈希算法类型的枚举，可选值：

| 常量 | 值 | 说明 |
|------|----|------|
| `MD5` | `"md5"` | MD5 哈希算法 |
| `SHA1` | `"sha1"` | SHA1 哈希算法 |
| `SHA256` | `"sha256"` | SHA256 哈希算法 |

### `newHash`

```go
func newHash(hashType HashType) (hash.Hash, error)
```

内部函数，根据 `HashType` 创建对应的 `hash.Hash` 实例。不支持的哈希类型会返回错误。

### `hashBytes`

```go
func hashBytes(data []byte, hashType HashType) (string, error)
```

内部函数，对字节切片计算哈希并返回十六进制编码字符串。

### `GetFileHash`

```go
func GetFileHash(path string, hashType HashType) (string, error)
```

计算文件的完整哈希值。使用 1MB 缓冲区通过 `io.CopyBuffer` 流式读取文件内容，避免将整个文件加载到内存中，适合处理大文件。

### `GetFileSnapshotHash`

```go
func GetFileSnapshotHash(path string, hashType HashType) (string, error)
```

计算文件的快照哈希值。仅读取文件的前 4KB（通过 `io.LimitReader` 限制），用于快速预筛选可能相同的文件。在重复文件检测中作为第二阶段过滤使用。

### 便捷函数（文件哈希）

| 函数 | 说明 |
|------|------|
| `GetFileMD5(path string) (string, error)` | 获取文件完整 MD5 |
| `GetFileSHA1(path string) (string, error)` | 获取文件完整 SHA1 |
| `GetFileSHA256(path string) (string, error)` | 获取文件完整 SHA256 |
| `GetFileSnapshotMD5(path string) (string, error)` | 获取文件快照 MD5 |
| `GetFileSnapshotSHA1(path string) (string, error)` | 获取文件快照 SHA1 |
| `GetFileSnapshotSHA256(path string) (string, error)` | 获取文件快照 SHA256 |

### `GetDirHash`

```go
func GetDirHash(dirPath string, hashType HashType, excludeDirs, excludeExts []string) (string, error)
```

计算整个目录的递归哈希值。

**参数：**

| 参数 | 类型 | 说明 |
|------|------|------|
| `dirPath` | `string` | 目录路径 |
| `hashType` | `HashType` | 哈希算法类型 |
| `excludeDirs` | `[]string` | 要排除的目录名（不含路径，如 `.git`） |
| `excludeExts` | `[]string` | 要排除的文件扩展名（含点，如 `.tmp`） |

**算法细节：**
- 递归遍历目录结构，对每个子项计算哈希
- 排序规则：目录排在文件前面，同类按名称升序排列
- 每个子项编码为 `"D:名称:哈希"`（目录）或 `"F:名称:哈希"`（文件）
- 空目录哈希基于 `"[EMPTY]"` 计算
- 不存在的路径哈希基于 `"[MISSING]"` 计算

### 便捷函数（目录哈希）

| 函数 | 说明 |
|------|------|
| `GetDirMD5(dirPath string, excludeDirs, excludeExts []string) (string, error)` | 获取目录 MD5 |
| `GetDirSHA1(dirPath string, excludeDirs, excludeExts []string) (string, error)` | 获取目录 SHA1 |
| `GetDirSHA256(dirPath string, excludeDirs, excludeExts []string) (string, error)` | 获取目录 SHA256 |

## 使用示例

```go
package main

import (
    "fmt"
    "log"
    "github.com/Mr-xiaotian/CelestialForge/pkg/file"
)

func main() {
    // 计算单个文件的 SHA256
    hash, err := file.GetFileSHA256("/data/important.dat")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("文件哈希: %s\n", hash)

    // 快照哈希（仅前 4KB）用于快速比对
    snapshot, err := file.GetFileSnapshotSHA1("/data/large_file.bin")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("快照哈希: %s\n", snapshot)

    // 计算目录哈希，排除 .git 目录和 .tmp 文件
    dirHash, err := file.GetDirSHA256(
        "/data/project",
        []string{".git", "node_modules"},
        []string{".tmp", ".log"},
    )
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("目录哈希: %s\n", dirHash)
}
```

## 关联文件

- [duplicate.md](duplicate.md) -- `getSnapshotDuplicate` 使用 `GetFileSnapshotSHA1`，`getHashDuplicate` 使用 `GetFileSHA1`
- [api.md](api.md) -- `FileInfo.Hash` 字段存储由本模块计算的哈希值

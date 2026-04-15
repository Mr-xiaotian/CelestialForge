# str.Table

> 源文件: `pkg/str/table.go`

## 概述

本文件提供纯文本表格格式化功能。`FormatTable` 函数接受二维字符串数据和列名定义，自动计算各列的最大宽度，输出对齐的纯文本表格。表格由表头行、分隔线和数据行三部分组成，列之间以两个空格分隔，每列使用左对齐的固定宽度格式化。该函数被 `file/duplicate.go` 中的 `DuplicateReport` 用于格式化重复文件的路径和大小信息。

## 类型/函数

### `FormatTable`

```go
func FormatTable(data [][]string, columns []string) string
```

将二维字符串数据格式化为对齐的纯文本表格。

**参数：**

| 参数 | 类型 | 说明 |
|------|------|------|
| `data` | `[][]string` | 表格数据，每行是一个字符串切片 |
| `columns` | `[]string` | 表头列名 |

**返回值：**

| 返回值 | 类型 | 说明 |
|--------|------|------|
| 表格文本 | `string` | 格式化后的纯文本表格 |

**算法细节：**
1. 以列名长度为初始宽度，遍历所有数据行更新各列的最大宽度
2. 使用 `fmt.Sprintf("%-*s", width, value)` 进行左对齐格式化
3. 列之间以两个空格 `"  "` 分隔
4. 表头和数据之间使用等宽的 `"-"` 分隔线
5. 通过 `strings.Builder` 高效拼接字符串

**输出格式示例：**

```
Item                      Size
------------------------  --------
/data/photos/img001.jpg   1.5 MB
/data/photos/img002.jpg   1.5 MB
```

## 使用示例

```go
package main

import (
    "fmt"
    "github.com/Mr-xiaotian/CelestialForge/pkg/str"
)

func main() {
    data := [][]string{
        {"/data/file1.txt", "1.2 MB"},
        {"/data/file2.txt", "3.4 MB"},
        {"/data/subdir/file3.txt", "567 KB"},
    }
    columns := []string{"Item", "Size"}

    table := str.FormatTable(data, columns)
    fmt.Println(table)
}
```

## 关联文件

- [../../file/duplicate.md](../../file/duplicate.md) -- `DuplicateReport` 使用 `FormatTable` 格式化重复文件报告

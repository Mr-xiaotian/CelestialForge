# size_test

> 源文件: `pkg/file/size_test.go`

## 概述

`file.GetFileSize` 功能的测试，采用黑盒测试（`package file_test`）。使用 `testdata/size_mtime/` 目录下预先准备的测试文件进行验证。

## 测试函数

| 测试函数 | 说明 |
|---------|------|
| `TestGetFileSize` | 表驱动测试，检查返回的字节数是否与预期的精确值匹配 |

## 关联文件

- [size.md](size.md) — `GetFileSize` 实现

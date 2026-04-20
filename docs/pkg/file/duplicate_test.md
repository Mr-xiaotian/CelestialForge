# duplicate_test

> 源文件: `pkg/file/duplicate_test.go`

## 概述

`file.DuplicateReport` 功能的测试，采用黑盒测试（`package file_test`）。使用 `testdata/duplicate/` 目录下预先准备的测试文件进行验证。

## 测试函数

| 测试函数 | 说明 |
|---------|------|
| `TestDuplicateReport` | 扫描测试目录，验证生成的重复文件报告包含预期的关键字符串 |
| `TestDuplicateReportEmpty` | 边界条件：输入为 `nil` 时返回空字符串 |

## 关联文件

- [duplicate.md](duplicate.md) — `DuplicateReport` 实现

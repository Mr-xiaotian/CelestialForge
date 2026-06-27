# cmd/debug.main

> 最后更新日期: 2026/06/28
> 源文件: `cmd/debug/main.go`

## 概述

这是项目的开发调试入口，用于手动测试 `pkg/file` 和 `pkg/units` 包中的各种功能。开发者可以通过在 `main()` 函数中注释/取消注释不同的调试函数来快速验证特定功能的行为。该文件不属于正式的命令行工具，仅用于开发阶段的功能验证和调试。

## 类型/函数

### `debug_info()`

调用 `file.GetFilesInfoRecursive` 递归获取 `tests/testdata` 目录下所有文件的信息并打印输出。

### `debug_duplicate()`

调用 `file.ScanDuplicateFile` 扫描当前目录下的重复文件，再通过 `file.DuplicateReport` 生成并输出格式化报告。

### `debug_bytes()`

演示 `units.HumanBytes` 的算术运算功能，展示人类可读的字节单位转换。

### `debug_size()`

演示 `file.GetDirSize` 的用法，获取并展示目录大小。

### `debug_mtime()`

演示 `file.GetDirMtime` 的用法，获取并展示目录的修改时间。

### `debug_hash()`

演示 `file.GetFileMD5` 和 `file.GetFileSnapshotMD5` 的用法，计算文件的 MD5 哈希值。

### `main()`

程序入口。默认仅执行 `debug_duplicate()`，其他调试函数已注释。开发者可根据需要取消注释来运行特定的调试功能。

## 使用示例

```bash
# 直接运行调试程序
go run cmd/debug/main.go

# 如需测试其他功能，编辑 main() 函数取消相应注释
# 例如取消 debug_info() 的注释以测试递归文件信息获取
```

## 关联文件

- [../../pkg/file/info.md](../../pkg/file/info.md) — `GetFilesInfoRecursive` 递归获取文件信息
- [../../pkg/file/duplicate.md](../../pkg/file/duplicate.md) — `ScanDuplicateFile` 与 `DuplicateReport` 重复文件扫描与报告
- [../../pkg/file/size.md](../../pkg/file/size.md) — `GetDirSize` 目录大小计算
- [../../pkg/file/mtime.md](../../pkg/file/mtime.md) — `GetDirMtime` 目录修改时间
- [../../pkg/file/hash.md](../../pkg/file/hash.md) — `GetFileMD5` 与 `GetFileSnapshotMD5` 文件哈希
- [../../pkg/units/bytes.md](../../pkg/units/bytes.md) — `HumanBytes` 字节单位
- [../duplicate/main.md](../duplicate/main.md) — 重复文件扫描的正式 CLI 工具

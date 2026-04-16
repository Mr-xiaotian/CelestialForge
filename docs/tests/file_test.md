# file 包测试

> 源文件: `pkg/file/size_test.go`, `pkg/file/mtime_test.go`, `pkg/file/hash_test.go`, `pkg/file/duplicate_test.go`

## 概述

`pkg/file` 包的功能测试套件，采用黑盒测试（`package file_test`），使用 `testdata/` 目录下预先准备的测试文件进行验证。测试按源文件拆分为独立的测试文件。

## 测试文件

### `size_test.go`

#### `TestGetFileSize(t *testing.T)`

表驱动测试，使用 `testdata/size_mtime/` 目录下的文件验证 `file.GetFileSize` 函数。检查返回的字节数是否与预期的精确值匹配。

### `mtime_test.go`

#### `TestGetFileMtime(t *testing.T)`

验证 `file.GetFileMtime` 函数返回的修改时间不为零值且不在未来时间。

### `hash_test.go`

#### `TestGetFileSHA1(t *testing.T)`

使用已知的 SHA1 哈希值验证 `file.GetFileSHA1` 函数的正确性。

### `duplicate_test.go`

#### `TestDuplicateReport(t *testing.T)`

扫描 `testdata/duplicate/` 目录，验证生成的重复文件报告包含预期的关键字符串。

#### `TestDuplicateReportEmpty(t *testing.T)`

测试边界条件：当输入为 `nil` 时，`DuplicateReport` 应返回空字符串。

## 使用示例

```bash
# 运行所有 file 包测试
go test -v ./pkg/file/

# 运行特定测试
go test -v -run TestGetFileSHA1 ./pkg/file/
```

## 关联文件

- [units_test.md](units_test.md) — units 包测试
- [structs_test.md](structs_test.md) — structs 包测试
- [../bench/bench_hash_test.md](../bench/bench_hash_test.md) — 哈希算法的性能基准测试

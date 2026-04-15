# tests.TestFile

> 源文件: `tests/file_test.go`

## 概述

`pkg/file` 包的功能测试套件，使用 `testdata/` 目录下预先准备的测试文件进行验证。涵盖文件大小获取、修改时间获取、SHA1 哈希计算以及重复文件扫描报告等核心功能的测试。采用表驱动测试（table-driven tests）模式组织用例，确保各函数在不同输入下的行为正确。

## 类型/函数

### `TestGetFileSize(t *testing.T)`

表驱动测试，使用 `testdata/size_mtime/` 目录下的文件验证 `file.GetFileSize` 函数。检查返回的字节数是否与预期的精确值匹配。

### `TestGetFileMtime(t *testing.T)`

验证 `file.GetFileMtime` 函数返回的修改时间：
- 不为零值（即文件存在且可读取修改时间）
- 不在未来时间（确保时间值合理）

### `TestGetFileSHA1(t *testing.T)`

使用已知的 SHA1 哈希值验证 `file.GetFileSHA1` 函数的正确性。将计算结果与预存的期望哈希值进行精确比较。

### `TestDuplicateReport(t *testing.T)`

扫描 `testdata/duplicate/` 目录，验证生成的重复文件报告：
- 报告包含预期的关键字符串
- 重复文件被正确识别和分组

### `TestDuplicateReportEmpty(t *testing.T)`

测试边界条件：当输入为 `nil` 时，`DuplicateReport` 应返回空字符串，不应 panic 或返回错误。

## 使用示例

```bash
# 运行所有文件相关测试
go test -v -run TestGetFile ./tests/
go test -v -run TestDuplicate ./tests/

# 运行单个测试
go test -v -run TestGetFileSHA1 ./tests/

# 查看测试覆盖率
go test -v -cover -run "TestGetFile|TestDuplicate" ./tests/
```

## 关联文件

- [executor_test.md](executor_test.md) — 同一测试包中的执行器测试
- [units_test.md](units_test.md) — 同一测试包中的单位格式化测试
- [structs_test.md](structs_test.md) — 同一测试包中的数据结构测试
- [../bench/bench_hash_test.md](../bench/bench_hash_test.md) — 哈希算法的性能基准测试
- [../cmd/duplicate/main.md](../cmd/duplicate/main.md) — 重复文件扫描的 CLI 工具

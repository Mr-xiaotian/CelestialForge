# tests.TestExecutor

> 源文件: `tests/executor_test.go`

## 概述

`grow.Executor` 执行器的测试套件，覆盖全部失败、部分失败和全部成功三种场景。每个测试用例都验证了执行器的计数器（`Counter.GetCompleted()`）、状态码（`State()`）以及结果的正确性。通过这些测试确保执行器在各种错误场景下都能正确处理任务并返回预期的结果。

## 类型/函数

### `TestExecutor_AllError(t *testing.T)`

测试所有任务均返回错误的场景。期望：
- 结果数量为 0
- 执行器状态码为 2（表示执行完成但全部失败）

### `TestExecutor_PartialError(t *testing.T)`

测试部分任务失败的场景。偶数编号的任务返回错误，奇数编号的任务成功。在 5 个任务中期望：
- 3 个成功结果
- 执行器正确记录完成数量

### `TestExecutor_AllSuccess(t *testing.T)`

测试所有任务均成功的场景。每个任务返回输入值的 2 倍。验证：
- 所有结果的数量正确
- 每个结果值等于对应任务输入的 2 倍

## 使用示例

```bash
# 运行所有 Executor 测试
go test -v -run TestExecutor ./tests/

# 运行特定测试
go test -v -run TestExecutor_AllSuccess ./tests/

# 运行并查看覆盖率
go test -v -cover -run TestExecutor ./tests/
```

## 关联文件

- [file_test.md](file_test.md) — 同一测试包中的文件操作测试
- [units_test.md](units_test.md) — 同一测试包中的单位格式化测试
- [structs_test.md](structs_test.md) — 同一测试包中的数据结构测试

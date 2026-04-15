# bench.BenchmarkFileHash

> 源文件: `bench/bench_hash_test.go`

## 概述

文件哈希算法的性能基准测试套件。通过对不同大小的临时文件（1KB、1MB、10MB、100MB）执行多种哈希算法（MD5、SHA1、SHA256）来衡量各算法在不同文件规模下的吞吐量表现。使用 `b.SetBytes` 设置字节数以便在基准测试结果中直接显示吞吐量（MB/s）。临时文件使用 32KB 写入缓冲区生成随机数据。

## 类型/函数

### `generateTempFile(size int64) (string, error)`

创建指定大小的临时文件，内容为随机数据。使用 32KB 的写入缓冲区逐块写入，确保高效生成大文件。返回临时文件路径供基准测试使用。

### `BenchmarkFileHash(b *testing.B)`

综合基准测试函数，遍历所有文件大小和哈希算法的组合进行测试。

**测试文件大小梯度：**

| 名称 | 大小 |
|------|------|
| 1KB | 1,024 字节 |
| 1MB | 1,048,576 字节 |
| 10MB | 10,485,760 字节 |
| 100MB | 104,857,600 字节 |

**哈希算法：** MD5、SHA1、SHA256

### `BenchmarkHashComparison(b *testing.B)`

在固定的 10MB 文件上比较所有哈希算法的性能差异，便于直接横向对比不同算法的吞吐量。

## 使用示例

```bash
# 运行所有哈希基准测试
go test -bench=. -benchmem ./bench/

# 仅运行哈希比较测试
go test -bench=BenchmarkHashComparison -benchmem ./bench/

# 运行特定大小的测试（例如 10MB）
go test -bench=BenchmarkFileHash/10MB -benchmem ./bench/
```

## 关联文件

- [../tests/file_test.md](../tests/file_test.md) — file 包的功能测试，包含 SHA1 哈希正确性验证
- [../cmd/debug/main.md](../cmd/debug/main.md) — debug_hash 函数提供 MD5 哈希的手动调试

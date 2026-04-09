package bench

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/Mr-xiaotian/CelestialForge/pkg/file"
)

// 测试文件大小梯度
var fileSizes = []struct {
	name string
	size int64
}{
	{"1KB", 1 << 10},
	{"1MB", 1 << 20},
	{"10MB", 10 << 20},
	{"100MB", 100 << 20},
}

// 哈希算法列表
var hashTypes = []struct {
	name     string
	hashType file.HashType
}{
	{"MD5", file.MD5},
	{"SHA1", file.SHA1},
	{"SHA256", file.SHA256},
}

// generateTempFile 生成指定大小的临时文件，填充随机数据
func generateTempFile(dir string, size int64) (string, error) {
	f, err := os.CreateTemp(dir, "benchfile-*")
	if err != nil {
		return "", err
	}
	defer f.Close()

	buf := make([]byte, 32*1024) // 32KB 写入缓冲区
	var written int64
	for written < size {
		n := int64(len(buf))
		if remaining := size - written; remaining < n {
			n = remaining
		}
		rand.Read(buf[:n])
		w, err := f.Write(buf[:n])
		if err != nil {
			os.Remove(f.Name())
			return "", err
		}
		written += int64(w)
	}

	return f.Name(), nil
}

func BenchmarkFileHash(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "bench-hash-*")
	if err != nil {
		b.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// 预先为每种大小生成一个测试文件
	testFiles := make(map[string]string)
	for _, fs := range fileSizes {
		path, err := generateTempFile(tmpDir, fs.size)
		if err != nil {
			b.Fatalf("生成 %s 测试文件失败: %v", fs.name, err)
		}
		testFiles[fs.name] = path
	}

	for _, fs := range fileSizes {
		for _, ht := range hashTypes {
			benchName := fmt.Sprintf("%s/%s", fs.name, ht.name)
			filePath := testFiles[fs.name]

			b.Run(benchName, func(b *testing.B) {
				info, _ := os.Stat(filePath)
				b.SetBytes(info.Size())
				b.ResetTimer()

				for i := 0; i < b.N; i++ {
					_, err := file.GetFileHash(filePath, ht.hashType)
					if err != nil {
						b.Fatalf("哈希计算失败: %v", err)
					}
				}
			})
		}
	}
}

// BenchmarkHashComparison 以表格形式对比同一文件大小下不同算法的性能
func BenchmarkHashComparison(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "bench-hash-cmp-*")
	if err != nil {
		b.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// 使用 10MB 文件作为对比基准
	filePath := filepath.Join(tmpDir, "compare-10mb.bin")
	f, err := os.Create(filePath)
	if err != nil {
		b.Fatal(err)
	}
	buf := make([]byte, 32*1024)
	var written int64
	for written < 10<<20 {
		n := int64(len(buf))
		if remaining := 10<<20 - written; remaining < n {
			n = remaining
		}
		rand.Read(buf[:n])
		w, _ := f.Write(buf[:n])
		written += int64(w)
	}
	f.Close()

	for _, ht := range hashTypes {
		b.Run(ht.name, func(b *testing.B) {
			b.SetBytes(10 << 20)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				file.GetFileHash(filePath, ht.hashType)
			}
		})
	}
}

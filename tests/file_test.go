package tests

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Mr-xiaotian/CelestialForge/pkg/file"
)

func TestGetFileSize(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		wantSize int64
		wantErr  bool
	}{
		{"普通文本文件", "testdata/size_mtime/testfile.txt", 97, false},
		{"空文件", "testdata/size_mtime/empty.txt", 0, false},
		{"较大文件", "testdata/size_mtime/large.bin", 1008, false},
		{"JSON文件", "testdata/size_mtime/data.json", 81, false},
		{"不存在的文件", "testdata/size_mtime/nonexistent.txt", 0, true},
		{"目录而非文件", "testdata/size_mtime", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			size, err := file.GetFileSize(tt.path)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetFileSize(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
			}
			if !tt.wantErr && size != tt.wantSize {
				t.Errorf("GetFileSize(%q) = %d, want %d", tt.path, size, tt.wantSize)
			}
		})
	}
}

func TestGetFileMtime(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"普通文本文件", "testdata/size_mtime/testfile.txt", false},
		{"空文件", "testdata/size_mtime/empty.txt", false},
		{"较大文件", "testdata/size_mtime/large.bin", false},
		{"JSON文件", "testdata/size_mtime/data.json", false},
		{"不存在的文件", "testdata/size_mtime/nonexistent.txt", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mtime, err := file.GetFileMtime(tt.path)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetFileMtime(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
			}
			if !tt.wantErr {
				if mtime.IsZero() {
					t.Error("GetFileMtime 返回了零值时间")
				}
				if mtime.After(time.Now()) {
					t.Error("GetFileMtime 返回了未来的时间")
				}
				t.Logf("文件 %s 最后修改时间: %v", tt.path, mtime)
			}
		})
	}
}

func TestGetFileSHA1(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		wantHash string
		wantErr  bool
	}{
		{"普通文本文件", "testdata/hash/testfile.txt", "98ce93c7ae27eb5e9e7c7933163ad5545a2fa2bf", false},
		{"空文件", "testdata/hash/empty.txt", "da39a3ee5e6b4b0d3255bfef95601890afd80709", false},
		{"较大文件", "testdata/hash/large.bin", "bbd47d1d53d8150aba6f81feecba834292350e41", false},
		{"JSON文件", "testdata/hash/data.json", "421337b7ed8ad2680a9c6044490cddc31c1f92f9", false},
		{"不存在的文件", "testdata/hash/nonexistent.txt", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := file.GetFileSHA1(tt.path)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetFileSHA1(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
			}
			if !tt.wantErr && hash != tt.wantHash {
				t.Errorf("GetFileSHA1(%q) = %q, want %q", tt.path, hash, tt.wantHash)
			}
		})
	}
}

func TestDuplicateReport(t *testing.T) {
	duplicates, err := file.GetDuplicateFile(filepath.Join("testdata", "duplicate"))
	if err != nil {
		t.Fatalf("GetDuplicateFile 失败: %v", err)
	}

	report := file.DuplicateReport(duplicates)
	if report == "" {
		t.Fatal("DuplicateReport 返回了空报告")
	}

	// 验证报告包含关键内容
	checks := []string{
		"Identical items found:",
		"Hash:",
		"Total size of duplicate items:",
		"Total number of duplicate items:",
		"Item with the most duplicates:",
		"dup_a1.txt",
		"dup_b1.txt",
	}
	for _, s := range checks {
		if !strings.Contains(report, s) {
			t.Errorf("报告中缺少 %q", s)
		}
	}

	t.Logf("生成的报告:\n%s", report)
}

func TestDuplicateReportEmpty(t *testing.T) {
	report := file.DuplicateReport(nil)
	if report != "" {
		t.Errorf("空输入应返回空字符串, 实际得到: %q", report)
	}
}

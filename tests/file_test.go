package tests

import (
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
		{"普通文本文件", "testdata/testfile.txt", 97, false},
		{"空文件", "testdata/empty.txt", 0, false},
		{"较大文件", "testdata/large.bin", 1008, false},
		{"JSON文件", "testdata/data.json", 81, false},
		{"不存在的文件", "testdata/nonexistent.txt", 0, true},
		{"目录而非文件", "testdata", 0, true},
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
		{"普通文本文件", "testdata/testfile.txt", false},
		{"空文件", "testdata/empty.txt", false},
		{"较大文件", "testdata/large.bin", false},
		{"JSON文件", "testdata/data.json", false},
		{"不存在的文件", "testdata/nonexistent.txt", true},
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

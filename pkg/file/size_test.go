package file_test

import (
	"testing"

	"github.com/Mr-xiaotian/CelestialForge/pkg/file"
)

func TestGetFileSize(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		wantSize int64
		wantErr  bool
	}{
		{"普通文本文件", "../../testdata/size_mtime/testfile.txt", 97, false},
		{"空文件", "../../testdata/size_mtime/empty.txt", 0, false},
		{"较大文件", "../../testdata/size_mtime/large.bin", 1008, false},
		{"JSON文件", "../../testdata/size_mtime/data.json", 81, false},
		{"不存在的文件", "../../testdata/size_mtime/nonexistent.txt", 0, true},
		{"目录而非文件", "../../testdata/size_mtime", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			size, err := file.GetFileSize(tt.path)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetFileSize(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
			}
			if !tt.wantErr && int64(size) != tt.wantSize {
				t.Errorf("GetFileSize(%q) = %d, want %d", tt.path, size, tt.wantSize)
			}
		})
	}
}

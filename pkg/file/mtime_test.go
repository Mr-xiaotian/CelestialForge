package file_test

import (
	"testing"
	"time"

	"github.com/Mr-xiaotian/CelestialForge/pkg/file"
)

func TestGetFileMtime(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"普通文本文件", "../../testdata/size_mtime/testfile.txt", false},
		{"空文件", "../../testdata/size_mtime/empty.txt", false},
		{"较大文件", "../../testdata/size_mtime/large.bin", false},
		{"JSON文件", "../../testdata/size_mtime/data.json", false},
		{"不存在的文件", "../../testdata/size_mtime/nonexistent.txt", true},
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

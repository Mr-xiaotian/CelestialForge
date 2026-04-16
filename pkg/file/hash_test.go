package file_test

import (
	"testing"

	"github.com/Mr-xiaotian/CelestialForge/pkg/file"
)

func TestGetFileSHA1(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		wantHash string
		wantErr  bool
	}{
		{"普通文本文件", "../../testdata/hash/testfile.txt", "98ce93c7ae27eb5e9e7c7933163ad5545a2fa2bf", false},
		{"空文件", "../../testdata/hash/empty.txt", "da39a3ee5e6b4b0d3255bfef95601890afd80709", false},
		{"较大文件", "../../testdata/hash/large.bin", "bbd47d1d53d8150aba6f81feecba834292350e41", false},
		{"JSON文件", "../../testdata/hash/data.json", "421337b7ed8ad2680a9c6044490cddc31c1f92f9", false},
		{"不存在的文件", "../../testdata/hash/nonexistent.txt", "", true},
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

package tests

import (
	"path/filepath"
	"sort"
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

func TestGetFilesRecursive(t *testing.T) {
	tests := []struct {
		name      string
		root      string
		wantFiles []string
		wantErr   bool
	}{
		{
			"遍历整个testdata目录",
			"testdata",
			[]string{
				filepath.Join("testdata", ".hidden"),
				filepath.Join("testdata", "a", "b", "c", "deep.txt"),
				filepath.Join("testdata", "data.json"),
				filepath.Join("testdata", "docs", "README.md"),
				filepath.Join("testdata", "docs", "config.toml"),
				filepath.Join("testdata", "empty.txt"),
				filepath.Join("testdata", "emptydir", ".gitkeep"),
				filepath.Join("testdata", "large.bin"),
				filepath.Join("testdata", "src", "main.go"),
				filepath.Join("testdata", "src", "util", "math.go"),
				filepath.Join("testdata", "testfile.txt"),
			},
			false,
		},
		{
			"遍历子目录src",
			filepath.Join("testdata", "src"),
			[]string{
				filepath.Join("testdata", "src", "main.go"),
				filepath.Join("testdata", "src", "util", "math.go"),
			},
			false,
		},
		{
			"遍历深层嵌套目录",
			filepath.Join("testdata", "a"),
			[]string{
				filepath.Join("testdata", "a", "b", "c", "deep.txt"),
			},
			false,
		},
		{
			"遍历仅含隐藏文件的目录",
			filepath.Join("testdata", "emptydir"),
			[]string{
				filepath.Join("testdata", "emptydir", ".gitkeep"),
			},
			false,
		},
		{
			"遍历docs子目录",
			filepath.Join("testdata", "docs"),
			[]string{
				filepath.Join("testdata", "docs", "README.md"),
				filepath.Join("testdata", "docs", "config.toml"),
			},
			false,
		},
		{
			"不存在的目录",
			"testdata/nonexistent_dir",
			nil,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files, err := file.GetFilesRecursive(tt.root)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetFilesRecursive(%q) error = %v, wantErr %v", tt.root, err, tt.wantErr)
			}
			if !tt.wantErr {
				sort.Strings(files)
				sort.Strings(tt.wantFiles)
				if len(files) != len(tt.wantFiles) {
					t.Fatalf("文件数量不匹配: got %d, want %d\ngot:  %v\nwant: %v", len(files), len(tt.wantFiles), files, tt.wantFiles)
				}
				for i := range files {
					if files[i] != tt.wantFiles[i] {
						t.Errorf("文件[%d]不匹配: got %q, want %q", i, files[i], tt.wantFiles[i])
					}
				}
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
		{"普通文本文件", "testdata/testfile.txt", "98ce93c7ae27eb5e9e7c7933163ad5545a2fa2bf", false},
		{"空文件", "testdata/empty.txt", "da39a3ee5e6b4b0d3255bfef95601890afd80709", false},
		{"较大文件", "testdata/large.bin", "bbd47d1d53d8150aba6f81feecba834292350e41", false},
		{"JSON文件", "testdata/data.json", "421337b7ed8ad2680a9c6044490cddc31c1f92f9", false},
		{"不存在的文件", "testdata/nonexistent.txt", "", true},
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

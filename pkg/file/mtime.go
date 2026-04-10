package file

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

// GetFileMtime 返回文件最后修改时间
func GetFileMtime(path string) (time.Time, error) {
	info, err := os.Stat(path)
	if err != nil {
		return time.Time{}, fmt.Errorf("无法获取文件信息: %w", err)
	}

	if info.IsDir() {
		return time.Time{}, fmt.Errorf("路径是目录而非文件: %s", path)
	}

	return info.ModTime(), nil
}

// GetDirMtime 返回目录下所有文件的最后修改时间
func GetDirMtime(path string) (time.Time, error) {
	dirMtime := time.Time{}

	err := filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err // 跳过无法访问的文件
		}
		if !d.IsDir() {
			info, _ := d.Info()
			if info.ModTime().After(dirMtime) {
				dirMtime = info.ModTime()
			}
		}
		return nil
	})

	return dirMtime, err
}

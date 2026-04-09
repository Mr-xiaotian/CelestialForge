package file

import (
	"fmt"
	"os"
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

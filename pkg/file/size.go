package file

import (
	"fmt"
	"os"
)

// GetFileSize 返回文件大小（字节）
func GetFileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, fmt.Errorf("无法获取文件信息: %w", err)
	}

	if info.IsDir() {
		return 0, fmt.Errorf("路径是目录而非文件: %s", path)
	}

	return info.Size(), nil
}

package file

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/Mr-xiaotian/CelestialForge/pkg/units"
)

// GetFileSize 返回文件大小（字节）
func GetFileSize(path string) (units.HumanBytes, error) {
	info, err := os.Stat(path)
	if err != nil {
		return units.NewHumanBytes(0), fmt.Errorf("无法获取文件信息: %w", err)
	}

	if info.IsDir() {
		return units.NewHumanBytes(0), fmt.Errorf("路径是目录而非文件: %s", path)
	}

	return units.NewHumanBytes(info.Size()), nil
}

// GetDirSize 返回目录下所有文件的总大小（字节）
func GetDirSize(path string) (units.HumanBytes, error) {
	dirSize := int64(0)

	err := filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err // 跳过无法访问的文件
		}
		if !d.IsDir() {
			info, _ := d.Info()
			dirSize += info.Size()
		}
		return nil
	})

	return units.NewHumanBytes(dirSize), err
}

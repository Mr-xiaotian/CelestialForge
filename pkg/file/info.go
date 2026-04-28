package file

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/Mr-xiaotian/CelestialForge/pkg/units"
)

// GetFilesInfoRecursive 递归获取所有文件
func GetFilesInfoRecursive(root string) (FileInfoMap, error) {
	if _, err := os.Stat(root); err != nil && !os.IsExist(err) {
		// 处理不存在的情况
		return nil, fmt.Errorf("目录不存在: %w", err)
	}

	files := make(FileInfoMap)
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err // 跳过无法访问的文件
		}
		if d.IsDir() {
			return nil // 跳过目录
		}
		info, _ := d.Info()
		files[path] = FileInfo{
			Size:  units.NewHumanBytes(info.Size()),
			Mtime: info.ModTime(),
		} // 完整路径作为键，包含大小和修改时间
		return nil
	})

	return files, err
}

package file

import (
	"io/fs"
	"path/filepath"
)

// GetFilesRecursive 递归获取所有文件
func GetFilesRecursive(root string) ([]string, error) {
	var files []string

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err // 跳过无法访问的文件
		}
		if !d.IsDir() {
			files = append(files, path) // 完整路径
		}
		return nil
	})

	return files, err
}

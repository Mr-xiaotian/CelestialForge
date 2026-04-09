package file

import (
	"io/fs"
	"path/filepath"
)

// GetFilesInfoRecursive 递归获取所有文件
func GetFilesInfoRecursive(root string) (FileInfoMap, error) {
	files := make(FileInfoMap)

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err // 跳过无法访问的文件
		}
		if !d.IsDir() {
			info, _ := d.Info()
			files[path] = FileInfo{
				Size:  info.Size(),
				Mtime: info.ModTime(),
			} // 完整路径
		}
		return nil
	})

	return files, err
}

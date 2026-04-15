package file

import (
	"time"

	"github.com/Mr-xiaotian/CelestialForge/pkg/units"
)

// FileInfo 文件元信息，包含路径、大小、修改时间和哈希值。
type FileInfo struct {
	Path  string
	Size  units.HumanBytes
	Mtime time.Time
	Hash  string
}

// FileInfoMap 文件路径到 FileInfo 的映射。
type FileInfoMap map[string]FileInfo

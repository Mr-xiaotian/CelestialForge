package file

import (
	"time"

	"github.com/Mr-xiaotian/CelestialForge/pkg/units"
)

type FileInfo struct {
	Path  string
	Size  units.HumanBytes
	Mtime time.Time
	Hash  string
}

type FileInfoMap map[string]FileInfo

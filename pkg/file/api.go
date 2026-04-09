package file

import "time"

type FileInfo struct {
	Path  string
	Size  int64
	Mtime time.Time
	Hash  string
}

type FileInfoMap map[string]FileInfo

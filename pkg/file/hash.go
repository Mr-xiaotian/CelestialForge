package file

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// HashType 哈希算法类型
type HashType string

const (
	MD5    HashType = "md5"
	SHA1   HashType = "sha1"
	SHA256 HashType = "sha256"
)

// ============ file hash ============

// newHash 根据 HashType 创建对应的 hash.Hash
func newHash(hashType HashType) (hash.Hash, error) {
	switch hashType {
	case MD5:
		return md5.New(), nil
	case SHA1:
		return sha1.New(), nil
	case SHA256:
		return sha256.New(), nil
	default:
		return nil, fmt.Errorf("不支持的哈希类型: %s", hashType)
	}
}

// hashBytes 对字节切片计算哈希并返回十六进制字符串
func hashBytes(data []byte, hashType HashType) (string, error) {
	h, err := newHash(hashType)
	if err != nil {
		return "", err
	}
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil)), nil
}

// GetFileHash 计算文件哈希值
func GetFileHash(path string, hashType HashType) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	h, err := newHash(hashType)
	if err != nil {
		return "", err
	}

	buf := make([]byte, 1024*1024) // 1MB 缓冲区
	if _, err := io.CopyBuffer(h, file, buf); err != nil {
		return "", fmt.Errorf("计算哈希失败: %w", err)
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// GetFileSnapshotHash 获取文件快照哈希值
func GetFileSnapshotHash(path string, hashType HashType) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("open file failed: %w", err)
	}
	defer file.Close()

	h, err := newHash(hashType)
	if err != nil {
		return "", err
	}

	// 限制读取量：最大只读取前 4KB
	const limit = 4096
	if _, err := io.Copy(h, io.LimitReader(file, limit)); err != nil {
		return "", fmt.Errorf("calculate snapshot hash failed: %w", err)
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// 便捷函数

// GetFileMD5 获取文件 MD5
func GetFileMD5(path string) (string, error) {
	return GetFileHash(path, MD5)
}

// GetFileSHA1 获取文件 SHA1
func GetFileSHA1(path string) (string, error) {
	return GetFileHash(path, SHA1)
}

// GetFileSHA256 获取文件 SHA256
func GetFileSHA256(path string) (string, error) {
	return GetFileHash(path, SHA256)
}

// GetFileSnapshotMD5 获取文件快照 MD5
func GetFileSnapshotMD5(path string) (string, error) {
	return GetFileSnapshotHash(path, MD5)
}

// GetFileSnapshotSHA1 获取文件快照 SHA1
func GetFileSnapshotSHA1(path string) (string, error) {
	return GetFileSnapshotHash(path, SHA1)
}

// GetFileSnapshotSHA256 获取文件快照 SHA256
func GetFileSnapshotSHA256(path string) (string, error) {
	return GetFileSnapshotHash(path, SHA256)
}

// ============ dir hash ============

// GetDirHash 计算整个文件夹的哈希值（递归包含子文件）。
// 目录哈希由子节点哈希组合而来：每个子项编码为 "D:name:hash" 或 "F:name:hash"，
// 按目录优先、名称排序后拼接计算哈希。
//
//   - excludeDirs: 要排除的目录名（不含路径）。
//   - excludeExts: 要排除的文件扩展名（含点，例如 ".tmp"）。
func GetDirHash(dirPath string, hashType HashType, excludeDirs, excludeExts []string) (string, error) {
	excludeDirSet := make(map[string]struct{}, len(excludeDirs))
	for _, d := range excludeDirs {
		excludeDirSet[d] = struct{}{}
	}
	excludeExtSet := make(map[string]struct{}, len(excludeExts))
	for _, ext := range excludeExts {
		excludeExtSet[strings.ToLower(ext)] = struct{}{}
	}

	return computeDirHash(dirPath, hashType, excludeDirSet, excludeExtSet)
}

// computeDirHash 递归计算目录或文件的哈希
func computeDirHash(path string, hashType HashType, excludeDirs, excludeExts map[string]struct{}) (string, error) {
	// --- 排除目录 ---
	if _, ok := excludeDirs[filepath.Base(path)]; ok {
		return "", nil
	}

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return hashBytes([]byte("[MISSING]"), hashType)
		}
		return "", fmt.Errorf("获取路径信息失败: %w", err)
	}

	// --- 文件 ---
	if !info.IsDir() {
		ext := strings.ToLower(filepath.Ext(path))
		if _, ok := excludeExts[ext]; ok {
			return "", nil
		}
		return GetFileHash(path, hashType)
	}

	// --- 目录：递归计算子项哈希 ---
	entries, err := os.ReadDir(path)
	if err != nil {
		return "", fmt.Errorf("读取目录失败: %w", err)
	}

	// 排序：目录优先，名称升序
	sort.Slice(entries, func(i, j int) bool {
		di, dj := entries[i].IsDir(), entries[j].IsDir()
		if di != dj {
			return di // 目录排在前面
		}
		return entries[i].Name() < entries[j].Name()
	})

	var combined []byte
	for _, entry := range entries {
		childPath := filepath.Join(path, entry.Name())
		h, err := computeDirHash(childPath, hashType, excludeDirs, excludeExts)
		if err != nil {
			return "", err
		}
		if h == "" {
			continue
		}
		tag := "F"
		if entry.IsDir() {
			tag = "D"
		}
		combined = append(combined, fmt.Appendf(nil, "%s:%s:%s", tag, entry.Name(), h)...)
	}

	if len(combined) == 0 {
		combined = []byte("[EMPTY]")
	}

	return hashBytes(combined, hashType)
}

// ============ 目录哈希便捷函数 ============

// GetDirMD5 获取目录 MD5
func GetDirMD5(dirPath string, excludeDirs, excludeExts []string) (string, error) {
	return GetDirHash(dirPath, MD5, excludeDirs, excludeExts)
}

// GetDirSHA1 获取目录 SHA1
func GetDirSHA1(dirPath string, excludeDirs, excludeExts []string) (string, error) {
	return GetDirHash(dirPath, SHA1, excludeDirs, excludeExts)
}

// GetDirSHA256 获取目录 SHA256
func GetDirSHA256(dirPath string, excludeDirs, excludeExts []string) (string, error) {
	return GetDirHash(dirPath, SHA256, excludeDirs, excludeExts)
}

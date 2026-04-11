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
)

// HashType 哈希算法类型
type HashType string

const (
	MD5    HashType = "md5"
	SHA1   HashType = "sha1"
	SHA256 HashType = "sha256"
)

// GetFileHash 计算文件哈希值
func GetFileHash(path string, hashType HashType) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	var h hash.Hash
	switch hashType {
	case MD5:
		h = md5.New()
	case SHA1:
		h = sha1.New()
	case SHA256:
		h = sha256.New()
	default:
		return "", fmt.Errorf("不支持的哈希类型: %s", hashType)
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

	// 1. 确定哈希算法
	var h hash.Hash
	switch hashType {
	case MD5:
		h = md5.New()
	case SHA1:
		h = sha1.New()
	case SHA256:
		h = sha256.New()
	default:
		return "", fmt.Errorf("unsupported hash type: %s", hashType)
	}

	// 2. 限制读取量：最大只读取前 4KB
	// io.LimitReader 会在达到 4096 字节或文件末尾时停止
	const limit = 4096
	if _, err := io.Copy(h, io.LimitReader(file, limit)); err != nil {
		return "", fmt.Errorf("calculate snapshot hash failed: %w", err)
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// ============ 便捷函数 ============

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

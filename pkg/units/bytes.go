package units

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	unitNames = []string{"B", "KB", "MB", "GB", "TB"}
	unitMap   = map[string]int64{
		"B":  1,
		"KB": 1024,
		"MB": 1024 * 1024,
		"GB": 1024 * 1024 * 1024,
		"TB": 1024 * 1024 * 1024 * 1024,
	}
	parseRe = regexp.MustCompile(`(\d+(?:\.\d+)?)([A-Za-z]+)`)
)

// HumanBytes 是一个基于 int64 的类型，自动以人类可读格式显示字节大小
type HumanBytes int64

// String 转换为人类可读字符串（如 "1GB 512MB"）
func (h HumanBytes) String() string {
	if h <= 0 {
		return "0B"
	}

	remaining := int64(h)
	var parts []string

	for i := len(unitNames) - 1; i >= 0; i-- {
		unitSize := unitMap[unitNames[i]]
		if remaining >= unitSize {
			value := remaining / unitSize
			remaining %= unitSize
			parts = append(parts, fmt.Sprintf("%d%s", value, unitNames[i]))
		}
	}

	return strings.Join(parts, " ")
}

// Int64 返回原始字节数
func (h HumanBytes) Int64() int64 {
	return int64(h)
}

// Add 加法
func (h HumanBytes) Add(other HumanBytes) HumanBytes {
	return HumanBytes(int64(h) + int64(other))
}

// Sub 减法
func (h HumanBytes) Sub(other HumanBytes) HumanBytes {
	return HumanBytes(int64(h) - int64(other))
}

// Mul 乘法
func (h HumanBytes) Mul(n int64) HumanBytes {
	return HumanBytes(int64(h) * n)
}

// Div 整除
func (h HumanBytes) Div(n int64) HumanBytes {
	return HumanBytes(int64(h) / n)
}

// Mod 取模
func (h HumanBytes) Mod(n int64) HumanBytes {
	return HumanBytes(int64(h) % n)
}

// NewHumanBytes 用 int64 创建 HumanBytes
func NewHumanBytes(b int64) HumanBytes {
	return HumanBytes(b)
}

// ParseHumanBytes 解析人类可读字符串，如 "1GB 512MB" 或 "1.5GB"
func ParseHumanBytes(text string) (HumanBytes, error) {
	matches := parseRe.FindAllStringSubmatch(text, -1)
	if len(matches) == 0 {
		return 0, fmt.Errorf("无法解析输入: %s", text)
	}

	var total float64
	for _, m := range matches {
		value, err := strconv.ParseFloat(m[1], 64)
		if err != nil {
			return 0, fmt.Errorf("无法解析数值: %s", m[1])
		}
		unit := strings.ToUpper(m[2])
		multiplier, ok := unitMap[unit]
		if !ok {
			return 0, fmt.Errorf("未知单位: %s", m[2])
		}
		total += value * float64(multiplier)
	}

	return HumanBytes(int64(total)), nil
}

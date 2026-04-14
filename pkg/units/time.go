package units

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

var (
	timeUnitNames   = []string{"d", "h", "m", "s"}
	timeUnitFactors = map[string]float64{
		"d": 86400,
		"h": 3600,
		"m": 60,
		"s": 1,
	}
	timeParseRe = regexp.MustCompile(`(\d+(?:\.\d+)?)\s*([dhms])`)
)

// HumanTime 表示可读的时间长度（duration），内部以秒（float64）存储。
// 打印时自动格式化为 "1d 2h 3m 4.56s" 形式。
type HumanTime float64

// String 转换为人类可读字符串（如 "1d 2h 3m 4.56s"）
func (t HumanTime) String() string {
	return t.Format(2, false)
}

// Format 格式化为人类可读字符串
//   - precision: 小数秒的精度位数
//   - showZero: 是否显示值为 0 的中间单位
func (t HumanTime) Format(precision int, showZero bool) string {
	remaining := math.Abs(float64(t))
	var parts []string

	for _, unit := range timeUnitNames {
		factor := timeUnitFactors[unit]
		if unit == "s" {
			// 秒作为最后一个单位，处理小数部分
			break
		}
		if remaining >= factor || (showZero && len(parts) > 0) {
			val := int(remaining / factor)
			remaining -= float64(val) * factor
			if val > 0 || showZero {
				parts = append(parts, fmt.Sprintf("%d%s", val, unit))
			}
		}
	}

	// 处理剩余秒数（含小数）
	if remaining > 0 {
		parts = append(parts, strconv.FormatFloat(remaining, 'f', precision, 64)+"s")
	}

	if len(parts) == 0 {
		parts = append(parts, "0s")
	}

	sign := ""
	if t < 0 {
		sign = "-"
	}
	return sign + strings.Join(parts, " ")
}

// Float64 返回原始秒数
func (t HumanTime) Float64() float64 {
	return float64(t)
}

// Add 加法
func (t HumanTime) Add(other HumanTime) HumanTime {
	return HumanTime(float64(t) + float64(other))
}

// Sub 减法
func (t HumanTime) Sub(other HumanTime) HumanTime {
	return HumanTime(float64(t) - float64(other))
}

// Mul 乘法
func (t HumanTime) Mul(n float64) HumanTime {
	return HumanTime(float64(t) * n)
}

// Div 除法
func (t HumanTime) Div(n float64) HumanTime {
	return HumanTime(float64(t) / n)
}

// Neg 取反
func (t HumanTime) Neg() HumanTime {
	return HumanTime(-float64(t))
}

// NewHumanTime 用秒数创建 HumanTime
func NewHumanTime(seconds float64) HumanTime {
	return HumanTime(seconds)
}

// ParseHumanTime 解析人类可读字符串，如 "1d 2h 3m 4.56s"
func ParseHumanTime(text string) (HumanTime, error) {
	matches := timeParseRe.FindAllStringSubmatch(strings.ToLower(text), -1)
	if len(matches) == 0 {
		return 0, fmt.Errorf("无法解析时间长度: %s", text)
	}

	var total float64
	for _, m := range matches {
		val, err := strconv.ParseFloat(m[1], 64)
		if err != nil {
			return 0, fmt.Errorf("无法解析数值: %s", m[1])
		}
		unit := m[2]
		factor, ok := timeUnitFactors[unit]
		if !ok {
			return 0, fmt.Errorf("未知单位: %s", unit)
		}
		total += val * factor
	}

	return HumanTime(total), nil
}

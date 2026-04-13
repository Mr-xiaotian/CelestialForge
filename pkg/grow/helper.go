package grow

func trunc(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	segmentLen := max(1, maxLen/3) // 确保每段至少1字符
	return s[:segmentLen*2] + "..." + s[len(s)-segmentLen:]
}

package str

import (
	"fmt"
	"strings"
)

func FormatTable(data [][]string, columns []string) string {
	widths := make([]int, len(columns))
	for i, col := range columns {
		widths[i] = len(col)
	}
	for _, row := range data {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	var sb strings.Builder
	// header
	for i, col := range columns {
		if i > 0 {
			sb.WriteString("  ")
		}
		sb.WriteString(fmt.Sprintf("%-*s", widths[i], col))
	}
	sb.WriteString("\n")
	// separator
	for i, w := range widths {
		if i > 0 {
			sb.WriteString("  ")
		}
		sb.WriteString(strings.Repeat("-", w))
	}
	sb.WriteString("\n")
	// rows
	for _, row := range data {
		for i, cell := range row {
			if i > 0 {
				sb.WriteString("  ")
			}
			if i < len(widths) {
				sb.WriteString(fmt.Sprintf("%-*s", widths[i], cell))
			}
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

package file_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/Mr-xiaotian/CelestialForge/pkg/file"
)

func TestDuplicateReport(t *testing.T) {
	duplicates, err := file.ScanDuplicateFile(filepath.Join("../../testdata", "duplicate"), 4)
	if err != nil {
		t.Fatalf("ScanDuplicateFile 失败: %v", err)
	}

	report := file.DuplicateReport(duplicates)
	if report == "" {
		t.Fatal("DuplicateReport 返回了空报告")
	}

	// 验证报告包含关键内容
	checks := []string{
		"Identical items found:",
		"Hash:",
		"Total size of duplicate items:",
		"Total number of duplicate items:",
		"Item with the most duplicates:",
		"dup_a1.txt",
		"dup_b1.txt",
	}
	for _, s := range checks {
		if !strings.Contains(report, s) {
			t.Errorf("报告中缺少 %q", s)
		}
	}

	t.Logf("生成的报告:\n%s", report)
}

func TestDuplicateReportEmpty(t *testing.T) {
	report := file.DuplicateReport(nil)
	if report != "" {
		t.Errorf("空输入应返回空字符串, 实际得到: %q", report)
	}
}

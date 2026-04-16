package units_test

import (
	"testing"

	"github.com/Mr-xiaotian/CelestialForge/pkg/units"
)

func TestHumanTime(t *testing.T) {
	// 测试代码
	tests := []struct {
		name    string
		time    float64
		wantStr string
		wantErr bool
	}{
		{"", 97, "1m 37.00s", false},
		{"", 0, "0s", false},
		{"", 1008, "16m 48.00s", false},
		{"", 81, "1m 21.00s", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			humanTime := units.NewHumanTime(tt.time)
			if humanTime.String() != tt.wantStr {
				t.Errorf("NewHumanTime(%v) = %s, want %s", tt.time, humanTime, tt.wantStr)
			}
		})
	}

}

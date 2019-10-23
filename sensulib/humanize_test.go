package sensulib_test

import (
	"math"
	"testing"

	"github.com/julian7/sensu-base-checks/sensulib"
)

func TestSizeToHuman(t *testing.T) {
	tests := []struct {
		name string
		size uint64
		want string
	}{
		{"low bytes", 5, "5 B"},
		{"bytes", 55, "55 B"},
		{"low kilobytes", 1800, "1.8 KiB"},
		{"kilobytes", 20000, "20 KiB"},
		{"kilobytes 2", 20481, "20 KiB"},
		{"low megabytes", uint64(0x5.32p20), "5.2 MiB"},
		{"megabytes", uint64(0x12.6p20), "18 MiB"},
		{"low gigabytes", uint64(0x5.32p30), "5.2 GiB"},
		{"gigabytes", uint64(0x12.6p30), "18 GiB"},
		{"low terabytes", uint64(0x5.32p40), "5.2 TiB"},
		{"terabytes", uint64(0x12.6p40), "18 TiB"},
		{"low petabytes", uint64(0x5.32p50), "5.2 PiB"},
		{"petabytes", uint64(0x12.6p50), "18 PiB"},
		{"low exabytes", uint64(0x5.32p60), "5.2 EiB"},
		{"max exabytes", uint64(0x8.6p60), "8.4 EiB"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sensulib.SizeToHuman(tt.size); got != tt.want {
				t.Errorf("SizeToHuman() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPercentToHuman(t *testing.T) {
	tt := []struct {
		name      string
		value     float64
		precision int
		want      string
	}{
		{"round to 2", math.Pi, 2, "3.14%"},
		{"round to 1", math.Pi, 1, "3.1%"},
		{"round to 0", math.Pi, 0, "3%"},
		{"trim trailing zeroes", 80.0, 2, "80%"},
		{"trim trailing zeroes 2", 80.1, 2, "80.1%"},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			if got := sensulib.PercentToHuman(tc.value, tc.precision); got != tc.want {
				t.Errorf("PercentToHuman() = %v, want: %v", got, tc.want)
			}
		})
	}
}

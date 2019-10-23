package sensulib

import (
	"fmt"
	"math"
	"strconv"
)

func logn(n, b float64) float64 {
	return math.Log(n) / math.Log(b)
}

// SizeToHuman converts byte size to IEC size string
func SizeToHuman(size uint64) string {
	var precision int

	prefixes := []string{"B", "KiB", "MiB", "GiB", "TiB", "PiB", "EiB"}
	s := float64(size)
	e := math.Floor(logn(s, 1024))
	suffix := prefixes[int(e)]
	val := math.Floor(s/math.Pow(1024, e)*10+0.5) / 10

	if val < 10 && e > 0 {
		precision = 1
	}

	return fmt.Sprintf("%.[1]*[2]f %[3]s", precision, val, suffix)
}

// PercentToHuman returns a string value of percentage, with at most a given precision
func PercentToHuman(percent float64, precision int) string {
	exponent := math.Pow(10, float64(precision))
	roundedPercent := math.Round(percent*exponent) / exponent

	return fmt.Sprintf("%s%%", strconv.FormatFloat(roundedPercent, 'f', -1, 64))
}

package security

import (
	"math"
)

// EstimateLength estimates the length needed to generate a random string without collisions.
func EstimateLength(n int64, alphabetSize int64) int64 {
	length := math.Log10(float64(n)) / math.Log10(float64(alphabetSize))
	return int64(math.Ceil(length))
}

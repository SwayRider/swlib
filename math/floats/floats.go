// Package floats provides utilities for floating-point comparisons
// with epsilon tolerance to handle precision issues.
//
// # Usage
//
//	if floats.Equal64(0.1+0.2, 0.3) {
//	    // Handles floating-point precision issues
//	}
package floats

// Default epsilon values for floating-point comparisons.
const Epsilon64 float64 = 1e-12
const Epsilon32 float32 = 1e-6

// Compare64 compares two float64 values with epsilon tolerance.
// Returns -1 if a < b, 0 if a == b (within epsilon), 1 if a > b.
// An optional custom epsilon can be provided.
func Compare64(a, b float64, epsilon ...float64) int {
	if len(epsilon) == 0 {
		epsilon = []float64{Epsilon64}
	}

	if a < b-epsilon[0] {
		return -1
	} else if a > b+epsilon[0] {
		return 1
	} else {
		return 0
	}
}

// Equal64 checks if two float64 values are equal within epsilon tolerance.
func Equal64(a, b float64, epsilon ...float64) bool {
	return Compare64(a, b, epsilon...) == 0
}

// IsZero64 checks if a float64 value is zero within epsilon tolerance.
func IsZero64(a float64, epsilon ...float64) bool {
	return Equal64(a, 0, epsilon...)
}

// Compare32 compares two float32 values with epsilon tolerance.
// Returns -1 if a < b, 0 if a == b (within epsilon), 1 if a > b.
func Compare32(a, b float32, epsilon ...float32) int {
	if len(epsilon) == 0 {
		epsilon = []float32{Epsilon32}
	}

	if a < b-epsilon[0] {
		return -1
	} else if a > b+epsilon[0] {
		return 1
	} else {
		return 0
	}
}

// Equal32 checks if two float32 values are equal within epsilon tolerance.
func Equal32(a, b float32, epsilon ...float32) bool {
	return Compare32(a, b, epsilon...) == 0
}

// IsZero32 checks if a float32 value is zero within epsilon tolerance.
func IsZero32(a float32, epsilon ...float32) bool {
	return Equal32(a, 0, epsilon...)
}

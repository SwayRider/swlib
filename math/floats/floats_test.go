package floats_test

import (
	"testing"

	"github.com/swayrider/swlib/math/floats"
)

func TestCompare64_LessThan(t *testing.T) {
	result := floats.Compare64(1.0, 2.0)
	if result != -1 {
		t.Errorf("Compare64(1.0, 2.0) = %d, want -1", result)
	}
}

func TestCompare64_GreaterThan(t *testing.T) {
	result := floats.Compare64(2.0, 1.0)
	if result != 1 {
		t.Errorf("Compare64(2.0, 1.0) = %d, want 1", result)
	}
}

func TestCompare64_Equal(t *testing.T) {
	result := floats.Compare64(1.0, 1.0)
	if result != 0 {
		t.Errorf("Compare64(1.0, 1.0) = %d, want 0", result)
	}
}

func TestCompare64_EqualWithinEpsilon(t *testing.T) {
	a := 1.0
	b := 1.0 + floats.Epsilon64/2
	result := floats.Compare64(a, b)
	if result != 0 {
		t.Errorf("Compare64(%v, %v) = %d, want 0 (values within default epsilon)", a, b, result)
	}
}

func TestCompare64_CustomEpsilon(t *testing.T) {
	a := 1.0
	b := 1.1
	customEpsilon := 0.2

	result := floats.Compare64(a, b, customEpsilon)
	if result != 0 {
		t.Errorf("Compare64(%v, %v, %v) = %d, want 0 (values within custom epsilon)", a, b, customEpsilon, result)
	}

	smallEpsilon := 0.01
	result = floats.Compare64(a, b, smallEpsilon)
	if result != -1 {
		t.Errorf("Compare64(%v, %v, %v) = %d, want -1 (values outside custom epsilon)", a, b, smallEpsilon, result)
	}
}

func TestEqual64(t *testing.T) {
	tests := []struct {
		name     string
		a        float64
		b        float64
		epsilon  []float64
		expected bool
	}{
		{"exact equal", 1.0, 1.0, nil, true},
		{"within default epsilon", 1.0, 1.0 + floats.Epsilon64/2, nil, true},
		{"outside default epsilon", 1.0, 2.0, nil, false},
		{"within custom epsilon", 1.0, 1.5, []float64{1.0}, true},
		{"outside custom epsilon", 1.0, 1.5, []float64{0.1}, false},
		{"negative numbers equal", -5.0, -5.0, nil, true},
		{"negative numbers not equal", -5.0, -6.0, nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result bool
			if tt.epsilon != nil {
				result = floats.Equal64(tt.a, tt.b, tt.epsilon...)
			} else {
				result = floats.Equal64(tt.a, tt.b)
			}
			if result != tt.expected {
				t.Errorf("Equal64(%v, %v) = %v, want %v", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestIsZero64(t *testing.T) {
	tests := []struct {
		name     string
		a        float64
		epsilon  []float64
		expected bool
	}{
		{"exact zero", 0.0, nil, true},
		{"within default epsilon", floats.Epsilon64 / 2, nil, true},
		{"outside default epsilon", 1.0, nil, false},
		{"negative within epsilon", -floats.Epsilon64 / 2, nil, true},
		{"within custom epsilon", 0.5, []float64{1.0}, true},
		{"outside custom epsilon", 0.5, []float64{0.1}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result bool
			if tt.epsilon != nil {
				result = floats.IsZero64(tt.a, tt.epsilon...)
			} else {
				result = floats.IsZero64(tt.a)
			}
			if result != tt.expected {
				t.Errorf("IsZero64(%v) = %v, want %v", tt.a, result, tt.expected)
			}
		})
	}
}

func TestCompare32_LessThan(t *testing.T) {
	result := floats.Compare32(1.0, 2.0)
	if result != -1 {
		t.Errorf("Compare32(1.0, 2.0) = %d, want -1", result)
	}
}

func TestCompare32_GreaterThan(t *testing.T) {
	result := floats.Compare32(2.0, 1.0)
	if result != 1 {
		t.Errorf("Compare32(2.0, 1.0) = %d, want 1", result)
	}
}

func TestCompare32_Equal(t *testing.T) {
	result := floats.Compare32(1.0, 1.0)
	if result != 0 {
		t.Errorf("Compare32(1.0, 1.0) = %d, want 0", result)
	}
}

func TestCompare32_EqualWithinEpsilon(t *testing.T) {
	a := float32(1.0)
	b := float32(1.0) + floats.Epsilon32/2
	result := floats.Compare32(a, b)
	if result != 0 {
		t.Errorf("Compare32(%v, %v) = %d, want 0 (values within default epsilon)", a, b, result)
	}
}

func TestCompare32_CustomEpsilon(t *testing.T) {
	a := float32(1.0)
	b := float32(1.1)
	customEpsilon := float32(0.2)

	result := floats.Compare32(a, b, customEpsilon)
	if result != 0 {
		t.Errorf("Compare32(%v, %v, %v) = %d, want 0 (values within custom epsilon)", a, b, customEpsilon, result)
	}

	smallEpsilon := float32(0.01)
	result = floats.Compare32(a, b, smallEpsilon)
	if result != -1 {
		t.Errorf("Compare32(%v, %v, %v) = %d, want -1 (values outside custom epsilon)", a, b, smallEpsilon, result)
	}
}

func TestEqual32(t *testing.T) {
	tests := []struct {
		name     string
		a        float32
		b        float32
		epsilon  []float32
		expected bool
	}{
		{"exact equal", 1.0, 1.0, nil, true},
		{"within default epsilon", 1.0, 1.0 + floats.Epsilon32/2, nil, true},
		{"outside default epsilon", 1.0, 2.0, nil, false},
		{"within custom epsilon", 1.0, 1.5, []float32{1.0}, true},
		{"outside custom epsilon", 1.0, 1.5, []float32{0.1}, false},
		{"negative numbers equal", -5.0, -5.0, nil, true},
		{"negative numbers not equal", -5.0, -6.0, nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result bool
			if tt.epsilon != nil {
				result = floats.Equal32(tt.a, tt.b, tt.epsilon...)
			} else {
				result = floats.Equal32(tt.a, tt.b)
			}
			if result != tt.expected {
				t.Errorf("Equal32(%v, %v) = %v, want %v", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestIsZero32(t *testing.T) {
	tests := []struct {
		name     string
		a        float32
		epsilon  []float32
		expected bool
	}{
		{"exact zero", 0.0, nil, true},
		{"within default epsilon", floats.Epsilon32 / 2, nil, true},
		{"outside default epsilon", 1.0, nil, false},
		{"negative within epsilon", -floats.Epsilon32 / 2, nil, true},
		{"within custom epsilon", 0.5, []float32{1.0}, true},
		{"outside custom epsilon", 0.5, []float32{0.1}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result bool
			if tt.epsilon != nil {
				result = floats.IsZero32(tt.a, tt.epsilon...)
			} else {
				result = floats.IsZero32(tt.a)
			}
			if result != tt.expected {
				t.Errorf("IsZero32(%v) = %v, want %v", tt.a, result, tt.expected)
			}
		})
	}
}

package env

import (
	"os"
	"testing"
)

// Test environment variable keys - use unique names to avoid conflicts
const (
	testEnvKey     = "TEST_ENV_VAR"
	testEnvKeyArr  = "TEST_ENV_VAR_ARR"
	testEnvKeyInt  = "TEST_ENV_VAR_INT"
	testEnvKeyBool = "TEST_ENV_VAR_BOOL"
)

// Helper to set and cleanup environment variables
func setEnv(t *testing.T, key, value string) {
	t.Helper()
	if err := os.Setenv(key, value); err != nil {
		t.Fatalf("failed to set env var: %v", err)
	}
	t.Cleanup(func() {
		os.Unsetenv(key)
	})
}

// Helper to unset environment variable
func unsetEnv(t *testing.T, key string) {
	t.Helper()
	os.Unsetenv(key)
	t.Cleanup(func() {
		os.Unsetenv(key)
	})
}

// =============================================================================
// Get Tests
// =============================================================================

func TestGet_EnvSet(t *testing.T) {
	setEnv(t, testEnvKey, "test-value")

	result := Get(testEnvKey, "fallback")
	if result != "test-value" {
		t.Errorf("expected 'test-value', got '%s'", result)
	}
}

func TestGet_EnvNotSet(t *testing.T) {
	unsetEnv(t, testEnvKey)

	result := Get(testEnvKey, "fallback")
	if result != "fallback" {
		t.Errorf("expected 'fallback', got '%s'", result)
	}
}

func TestGet_EnvEmpty(t *testing.T) {
	setEnv(t, testEnvKey, "")

	result := Get(testEnvKey, "fallback")
	if result != "fallback" {
		t.Errorf("expected 'fallback' for empty env var, got '%s'", result)
	}
}

func TestGet_EmptyFallback(t *testing.T) {
	unsetEnv(t, testEnvKey)

	result := Get(testEnvKey, "")
	if result != "" {
		t.Errorf("expected empty string, got '%s'", result)
	}
}

// =============================================================================
// GetAsStringArr Tests
// =============================================================================

func TestGetAsStringArr_EnvSet(t *testing.T) {
	setEnv(t, testEnvKeyArr, "a,b,c")

	result := GetAsStringArr(testEnvKeyArr, "x")
	if len(result) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(result))
	}
	if result[0] != "a" || result[1] != "b" || result[2] != "c" {
		t.Errorf("unexpected values: %v", result)
	}
}

func TestGetAsStringArr_EnvNotSet(t *testing.T) {
	unsetEnv(t, testEnvKeyArr)

	result := GetAsStringArr(testEnvKeyArr, "x,y,z")
	if len(result) != 3 {
		t.Fatalf("expected 3 elements from fallback, got %d", len(result))
	}
	if result[0] != "x" || result[1] != "y" || result[2] != "z" {
		t.Errorf("unexpected fallback values: %v", result)
	}
}

func TestGetAsStringArr_SingleValue(t *testing.T) {
	setEnv(t, testEnvKeyArr, "single")

	result := GetAsStringArr(testEnvKeyArr, "fallback")
	if len(result) != 1 {
		t.Fatalf("expected 1 element, got %d", len(result))
	}
	if result[0] != "single" {
		t.Errorf("expected 'single', got '%s'", result[0])
	}
}

func TestGetAsStringArr_EmptyEnv(t *testing.T) {
	setEnv(t, testEnvKeyArr, "")

	result := GetAsStringArr(testEnvKeyArr, "fallback")
	if len(result) != 1 || result[0] != "fallback" {
		t.Errorf("expected fallback for empty env, got %v", result)
	}
}

func TestGetAsStringArr_WithSpaces(t *testing.T) {
	setEnv(t, testEnvKeyArr, "a, b , c")

	result := GetAsStringArr(testEnvKeyArr, "x")
	// Note: spaces are preserved, not trimmed
	if len(result) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(result))
	}
	if result[0] != "a" || result[1] != " b " || result[2] != " c" {
		t.Errorf("expected spaces to be preserved: %v", result)
	}
}

// =============================================================================
// GetAsInt Tests
// =============================================================================

func TestGetAsInt_ValidInt(t *testing.T) {
	setEnv(t, testEnvKeyInt, "42")

	result := GetAsInt(testEnvKeyInt, 0)
	if result != 42 {
		t.Errorf("expected 42, got %d", result)
	}
}

func TestGetAsInt_EnvNotSet(t *testing.T) {
	unsetEnv(t, testEnvKeyInt)

	result := GetAsInt(testEnvKeyInt, 100)
	if result != 100 {
		t.Errorf("expected fallback 100, got %d", result)
	}
}

func TestGetAsInt_InvalidInt(t *testing.T) {
	setEnv(t, testEnvKeyInt, "not-a-number")

	result := GetAsInt(testEnvKeyInt, 50)
	if result != 50 {
		t.Errorf("expected fallback 50 for invalid int, got %d", result)
	}
}

func TestGetAsInt_NegativeInt(t *testing.T) {
	setEnv(t, testEnvKeyInt, "-123")

	result := GetAsInt(testEnvKeyInt, 0)
	if result != -123 {
		t.Errorf("expected -123, got %d", result)
	}
}

func TestGetAsInt_Zero(t *testing.T) {
	setEnv(t, testEnvKeyInt, "0")

	result := GetAsInt(testEnvKeyInt, 99)
	if result != 0 {
		t.Errorf("expected 0, got %d", result)
	}
}

func TestGetAsInt_Float(t *testing.T) {
	setEnv(t, testEnvKeyInt, "3.14")

	result := GetAsInt(testEnvKeyInt, 0)
	// Float is not a valid int, should return fallback
	if result != 0 {
		t.Errorf("expected fallback 0 for float, got %d", result)
	}
}

func TestGetAsInt_EmptyString(t *testing.T) {
	setEnv(t, testEnvKeyInt, "")

	result := GetAsInt(testEnvKeyInt, 25)
	if result != 25 {
		t.Errorf("expected fallback 25 for empty string, got %d", result)
	}
}

// =============================================================================
// GetAsIntArr Tests
// =============================================================================

func TestGetAsIntArr_ValidInts(t *testing.T) {
	setEnv(t, testEnvKeyArr, "1,2,3,4,5")

	result := GetAsIntArr(testEnvKeyArr, "0")
	if len(result) != 5 {
		t.Fatalf("expected 5 elements, got %d", len(result))
	}
	expected := []int{1, 2, 3, 4, 5}
	for i, v := range result {
		if v != expected[i] {
			t.Errorf("at index %d: expected %d, got %d", i, expected[i], v)
		}
	}
}

func TestGetAsIntArr_EnvNotSet(t *testing.T) {
	unsetEnv(t, testEnvKeyArr)

	result := GetAsIntArr(testEnvKeyArr, "10,20,30")
	if len(result) != 3 {
		t.Fatalf("expected 3 elements from fallback, got %d", len(result))
	}
	if result[0] != 10 || result[1] != 20 || result[2] != 30 {
		t.Errorf("unexpected fallback values: %v", result)
	}
}

func TestGetAsIntArr_MixedValidInvalid(t *testing.T) {
	setEnv(t, testEnvKeyArr, "1,invalid,3,abc,5")

	result := GetAsIntArr(testEnvKeyArr, "0")
	// Invalid values should be skipped
	if len(result) != 3 {
		t.Fatalf("expected 3 valid elements, got %d: %v", len(result), result)
	}
	if result[0] != 1 || result[1] != 3 || result[2] != 5 {
		t.Errorf("unexpected values: %v", result)
	}
}

func TestGetAsIntArr_AllInvalid(t *testing.T) {
	setEnv(t, testEnvKeyArr, "a,b,c")

	result := GetAsIntArr(testEnvKeyArr, "0")
	// All invalid, should return empty slice
	if len(result) != 0 {
		t.Errorf("expected empty slice for all invalid values, got %v", result)
	}
}

func TestGetAsIntArr_NegativeNumbers(t *testing.T) {
	setEnv(t, testEnvKeyArr, "-1,-2,-3")

	result := GetAsIntArr(testEnvKeyArr, "0")
	if len(result) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(result))
	}
	if result[0] != -1 || result[1] != -2 || result[2] != -3 {
		t.Errorf("unexpected values: %v", result)
	}
}

// =============================================================================
// GetAsFloat64 Tests
// =============================================================================

func TestGetAsFloat64_ValidFloat(t *testing.T) {
	setEnv(t, testEnvKey, "3.14159")

	result := GetAsFloat64(testEnvKey, 0.0)
	if result != 3.14159 {
		t.Errorf("expected 3.14159, got %f", result)
	}
}

func TestGetAsFloat64_EnvNotSet(t *testing.T) {
	unsetEnv(t, testEnvKey)

	result := GetAsFloat64(testEnvKey, 2.5)
	if result != 2.5 {
		t.Errorf("expected fallback 2.5, got %f", result)
	}
}

func TestGetAsFloat64_InvalidFloat(t *testing.T) {
	setEnv(t, testEnvKey, "not-a-float")

	result := GetAsFloat64(testEnvKey, 1.0)
	if result != 1.0 {
		t.Errorf("expected fallback 1.0 for invalid float, got %f", result)
	}
}

func TestGetAsFloat64_Integer(t *testing.T) {
	setEnv(t, testEnvKey, "42")

	result := GetAsFloat64(testEnvKey, 0.0)
	if result != 42.0 {
		t.Errorf("expected 42.0, got %f", result)
	}
}

func TestGetAsFloat64_Negative(t *testing.T) {
	setEnv(t, testEnvKey, "-3.5")

	result := GetAsFloat64(testEnvKey, 0.0)
	if result != -3.5 {
		t.Errorf("expected -3.5, got %f", result)
	}
}

func TestGetAsFloat64_Scientific(t *testing.T) {
	setEnv(t, testEnvKey, "1.5e10")

	result := GetAsFloat64(testEnvKey, 0.0)
	if result != 1.5e10 {
		t.Errorf("expected 1.5e10, got %f", result)
	}
}

func TestGetAsFloat64_Zero(t *testing.T) {
	setEnv(t, testEnvKey, "0.0")

	result := GetAsFloat64(testEnvKey, 99.9)
	if result != 0.0 {
		t.Errorf("expected 0.0, got %f", result)
	}
}

// =============================================================================
// GetAsBool Tests
// =============================================================================

func TestGetAsBool_True(t *testing.T) {
	testCases := []string{"true", "TRUE", "True", "1", "t", "T"}
	for _, tc := range testCases {
		setEnv(t, testEnvKeyBool, tc)

		result := GetAsBool(testEnvKeyBool, false)
		if !result {
			t.Errorf("expected true for '%s', got false", tc)
		}
	}
}

func TestGetAsBool_False(t *testing.T) {
	testCases := []string{"false", "FALSE", "False", "0", "f", "F"}
	for _, tc := range testCases {
		setEnv(t, testEnvKeyBool, tc)

		result := GetAsBool(testEnvKeyBool, true)
		if result {
			t.Errorf("expected false for '%s', got true", tc)
		}
	}
}

func TestGetAsBool_EnvNotSet(t *testing.T) {
	unsetEnv(t, testEnvKeyBool)

	resultTrue := GetAsBool(testEnvKeyBool, true)
	if !resultTrue {
		t.Error("expected fallback true")
	}

	resultFalse := GetAsBool(testEnvKeyBool, false)
	if resultFalse {
		t.Error("expected fallback false")
	}
}

func TestGetAsBool_Invalid(t *testing.T) {
	setEnv(t, testEnvKeyBool, "yes")

	result := GetAsBool(testEnvKeyBool, false)
	// "yes" is not a valid bool string for strconv.ParseBool
	if result != false {
		t.Errorf("expected fallback false for invalid bool string, got true")
	}
}

func TestGetAsBool_EmptyString(t *testing.T) {
	setEnv(t, testEnvKeyBool, "")

	result := GetAsBool(testEnvKeyBool, true)
	if !result {
		t.Error("expected fallback true for empty string")
	}
}

// =============================================================================
// GetAsFloat64Arr Tests
// =============================================================================

func TestGetAsFloat64Arr_ValidFloats(t *testing.T) {
	setEnv(t, testEnvKeyArr, "1.1,2.2,3.3")

	result := GetAsFloat64Arr(testEnvKeyArr, "0.0")
	if len(result) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(result))
	}
	if result[0] != 1.1 || result[1] != 2.2 || result[2] != 3.3 {
		t.Errorf("unexpected values: %v", result)
	}
}

func TestGetAsFloat64Arr_EnvNotSet(t *testing.T) {
	unsetEnv(t, testEnvKeyArr)

	result := GetAsFloat64Arr(testEnvKeyArr, "5.5,6.6")
	if len(result) != 2 {
		t.Fatalf("expected 2 elements from fallback, got %d", len(result))
	}
	if result[0] != 5.5 || result[1] != 6.6 {
		t.Errorf("unexpected fallback values: %v", result)
	}
}

func TestGetAsFloat64Arr_MixedValidInvalid(t *testing.T) {
	setEnv(t, testEnvKeyArr, "1.0,invalid,3.0,abc,5.0")

	result := GetAsFloat64Arr(testEnvKeyArr, "0.0")
	if len(result) != 3 {
		t.Fatalf("expected 3 valid elements, got %d: %v", len(result), result)
	}
	if result[0] != 1.0 || result[1] != 3.0 || result[2] != 5.0 {
		t.Errorf("unexpected values: %v", result)
	}
}

func TestGetAsFloat64Arr_Integers(t *testing.T) {
	setEnv(t, testEnvKeyArr, "1,2,3")

	result := GetAsFloat64Arr(testEnvKeyArr, "0")
	if len(result) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(result))
	}
	if result[0] != 1.0 || result[1] != 2.0 || result[2] != 3.0 {
		t.Errorf("unexpected values: %v", result)
	}
}

func TestGetAsFloat64Arr_Scientific(t *testing.T) {
	setEnv(t, testEnvKeyArr, "1e3,2e-2")

	result := GetAsFloat64Arr(testEnvKeyArr, "0")
	if len(result) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(result))
	}
	if result[0] != 1000.0 || result[1] != 0.02 {
		t.Errorf("unexpected values: %v", result)
	}
}

// =============================================================================
// GetAsBoolArr Tests
// =============================================================================

func TestGetAsBoolArr_ValidBools(t *testing.T) {
	setEnv(t, testEnvKeyArr, "true,false,true")

	result := GetAsBoolArr(testEnvKeyArr, "false")
	if len(result) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(result))
	}
	if result[0] != true || result[1] != false || result[2] != true {
		t.Errorf("unexpected values: %v", result)
	}
}

func TestGetAsBoolArr_EnvNotSet(t *testing.T) {
	unsetEnv(t, testEnvKeyArr)

	result := GetAsBoolArr(testEnvKeyArr, "true,true")
	if len(result) != 2 {
		t.Fatalf("expected 2 elements from fallback, got %d", len(result))
	}
	if result[0] != true || result[1] != true {
		t.Errorf("unexpected fallback values: %v", result)
	}
}

func TestGetAsBoolArr_MixedValidInvalid(t *testing.T) {
	setEnv(t, testEnvKeyArr, "true,invalid,false,yes,1")

	result := GetAsBoolArr(testEnvKeyArr, "false")
	// "yes" is invalid for ParseBool, but "1" is valid (true)
	if len(result) != 3 {
		t.Fatalf("expected 3 valid elements, got %d: %v", len(result), result)
	}
	if result[0] != true || result[1] != false || result[2] != true {
		t.Errorf("unexpected values: %v", result)
	}
}

func TestGetAsBoolArr_NumericBools(t *testing.T) {
	setEnv(t, testEnvKeyArr, "1,0,1,0")

	result := GetAsBoolArr(testEnvKeyArr, "false")
	if len(result) != 4 {
		t.Fatalf("expected 4 elements, got %d", len(result))
	}
	if result[0] != true || result[1] != false || result[2] != true || result[3] != false {
		t.Errorf("unexpected values: %v", result)
	}
}

func TestGetAsBoolArr_AllInvalid(t *testing.T) {
	setEnv(t, testEnvKeyArr, "yes,no,on,off")

	result := GetAsBoolArr(testEnvKeyArr, "false")
	// "yes", "no", "on", "off" are all invalid for ParseBool
	if len(result) != 0 {
		t.Errorf("expected empty slice for all invalid values, got %v", result)
	}
}

// =============================================================================
// Edge Cases
// =============================================================================

func TestEdgeCase_VeryLongValue(t *testing.T) {
	// Create a long string
	longValue := ""
	for i := 0; i < 1000; i++ {
		longValue += "a"
	}
	setEnv(t, testEnvKey, longValue)

	result := Get(testEnvKey, "short")
	if len(result) != 1000 {
		t.Errorf("expected 1000 char string, got %d chars", len(result))
	}
}

func TestEdgeCase_SpecialCharacters(t *testing.T) {
	setEnv(t, testEnvKey, "!@#$%^&*()_+-=[]{}|;':\",./<>?")

	result := Get(testEnvKey, "fallback")
	if result != "!@#$%^&*()_+-=[]{}|;':\",./<>?" {
		t.Errorf("unexpected result for special chars: %s", result)
	}
}

func TestEdgeCase_Unicode(t *testing.T) {
	setEnv(t, testEnvKey, "日本語テスト")

	result := Get(testEnvKey, "fallback")
	if result != "日本語テスト" {
		t.Errorf("unexpected result for unicode: %s", result)
	}
}

func TestEdgeCase_WhitespaceValue(t *testing.T) {
	setEnv(t, testEnvKey, "   ")

	result := Get(testEnvKey, "fallback")
	// Non-empty string (whitespace) should be returned
	if result != "   " {
		t.Errorf("expected whitespace string, got '%s'", result)
	}
}

func TestEdgeCase_NewlineInValue(t *testing.T) {
	setEnv(t, testEnvKey, "line1\nline2")

	result := Get(testEnvKey, "fallback")
	if result != "line1\nline2" {
		t.Errorf("expected value with newline, got '%s'", result)
	}
}

func TestEdgeCase_CommaOnlyArray(t *testing.T) {
	setEnv(t, testEnvKeyArr, ",,,")

	result := GetAsStringArr(testEnvKeyArr, "fallback")
	// Should result in 4 empty strings
	if len(result) != 4 {
		t.Errorf("expected 4 elements, got %d", len(result))
	}
	for i, v := range result {
		if v != "" {
			t.Errorf("expected empty string at index %d, got '%s'", i, v)
		}
	}
}

func TestEdgeCase_LargeIntArray(t *testing.T) {
	// Create a large array
	values := ""
	for i := 0; i < 100; i++ {
		if i > 0 {
			values += ","
		}
		values += "42"
	}
	setEnv(t, testEnvKeyArr, values)

	result := GetAsIntArr(testEnvKeyArr, "0")
	if len(result) != 100 {
		t.Errorf("expected 100 elements, got %d", len(result))
	}
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkGet(b *testing.B) {
	os.Setenv(testEnvKey, "benchmark-value")
	defer os.Unsetenv(testEnvKey)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Get(testEnvKey, "fallback")
	}
}

func BenchmarkGetAsInt(b *testing.B) {
	os.Setenv(testEnvKeyInt, "12345")
	defer os.Unsetenv(testEnvKeyInt)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetAsInt(testEnvKeyInt, 0)
	}
}

func BenchmarkGetAsFloat64(b *testing.B) {
	os.Setenv(testEnvKey, "123.456")
	defer os.Unsetenv(testEnvKey)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetAsFloat64(testEnvKey, 0.0)
	}
}

func BenchmarkGetAsBool(b *testing.B) {
	os.Setenv(testEnvKeyBool, "true")
	defer os.Unsetenv(testEnvKeyBool)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetAsBool(testEnvKeyBool, false)
	}
}

func BenchmarkGetAsStringArr(b *testing.B) {
	os.Setenv(testEnvKeyArr, "a,b,c,d,e")
	defer os.Unsetenv(testEnvKeyArr)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetAsStringArr(testEnvKeyArr, "x")
	}
}

func BenchmarkGetAsIntArr(b *testing.B) {
	os.Setenv(testEnvKeyArr, "1,2,3,4,5")
	defer os.Unsetenv(testEnvKeyArr)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetAsIntArr(testEnvKeyArr, "0")
	}
}

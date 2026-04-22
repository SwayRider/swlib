package str

import (
	"testing"
)

// =============================================================================
// NullTerm Tests
// =============================================================================

func TestNullTerm_WithNulls(t *testing.T) {
	input := []byte("hello\x00\x00\x00")

	result := NullTerm(input)
	if result != "hello" {
		t.Errorf("expected 'hello', got '%s'", result)
	}
}

func TestNullTerm_NoNulls(t *testing.T) {
	input := []byte("hello")

	result := NullTerm(input)
	if result != "hello" {
		t.Errorf("expected 'hello', got '%s'", result)
	}
}

func TestNullTerm_OnlyNulls(t *testing.T) {
	input := []byte("\x00\x00\x00")

	result := NullTerm(input)
	if result != "" {
		t.Errorf("expected empty string, got '%s'", result)
	}
}

func TestNullTerm_EmptySlice(t *testing.T) {
	input := []byte{}

	result := NullTerm(input)
	if result != "" {
		t.Errorf("expected empty string, got '%s'", result)
	}
}

func TestNullTerm_SingleNull(t *testing.T) {
	input := []byte("\x00")

	result := NullTerm(input)
	if result != "" {
		t.Errorf("expected empty string, got '%s'", result)
	}
}

func TestNullTerm_NullInMiddle(t *testing.T) {
	// Null in middle should remain, only trailing nulls removed
	input := []byte("hel\x00lo\x00\x00")

	result := NullTerm(input)
	if result != "hel\x00lo" {
		t.Errorf("expected 'hel\\x00lo', got '%s'", result)
	}
}

func TestNullTerm_NullAtStart(t *testing.T) {
	input := []byte("\x00hello\x00")

	result := NullTerm(input)
	if result != "\x00hello" {
		t.Errorf("expected '\\x00hello', got '%s'", result)
	}
}

func TestNullTerm_Unicode(t *testing.T) {
	input := []byte("日本語\x00\x00")

	result := NullTerm(input)
	if result != "日本語" {
		t.Errorf("expected '日本語', got '%s'", result)
	}
}

func TestNullTerm_SpecialCharacters(t *testing.T) {
	input := []byte("!@#$%^&*()\x00")

	result := NullTerm(input)
	if result != "!@#$%^&*()" {
		t.Errorf("expected '!@#$%%^&*()', got '%s'", result)
	}
}

func TestNullTerm_Newlines(t *testing.T) {
	input := []byte("line1\nline2\x00")

	result := NullTerm(input)
	if result != "line1\nline2" {
		t.Errorf("expected 'line1\\nline2', got '%s'", result)
	}
}

func TestNullTerm_LargeInput(t *testing.T) {
	// Create a large byte slice
	input := make([]byte, 10000)
	for i := 0; i < 9990; i++ {
		input[i] = 'a'
	}
	// Last 10 bytes are nulls

	result := NullTerm(input)
	if len(result) != 9990 {
		t.Errorf("expected 9990 chars, got %d", len(result))
	}
}

// =============================================================================
// ToPtr Tests
// =============================================================================

func TestToPtr_NonEmpty(t *testing.T) {
	input := "hello"

	result := ToPtr(input)
	if result == nil {
		t.Fatal("expected non-nil pointer")
	}
	if *result != "hello" {
		t.Errorf("expected 'hello', got '%s'", *result)
	}
}

func TestToPtr_Empty(t *testing.T) {
	input := ""

	result := ToPtr(input)
	if result != nil {
		t.Errorf("expected nil for empty string, got '%s'", *result)
	}
}

func TestToPtr_SingleChar(t *testing.T) {
	input := "a"

	result := ToPtr(input)
	if result == nil {
		t.Fatal("expected non-nil pointer")
	}
	if *result != "a" {
		t.Errorf("expected 'a', got '%s'", *result)
	}
}

func TestToPtr_Whitespace(t *testing.T) {
	input := "   "

	result := ToPtr(input)
	if result == nil {
		t.Fatal("expected non-nil pointer for whitespace")
	}
	if *result != "   " {
		t.Errorf("expected '   ', got '%s'", *result)
	}
}

func TestToPtr_Unicode(t *testing.T) {
	input := "日本語"

	result := ToPtr(input)
	if result == nil {
		t.Fatal("expected non-nil pointer")
	}
	if *result != "日本語" {
		t.Errorf("expected '日本語', got '%s'", *result)
	}
}

func TestToPtr_SpecialCharacters(t *testing.T) {
	input := "!@#$%^&*()"

	result := ToPtr(input)
	if result == nil {
		t.Fatal("expected non-nil pointer")
	}
	if *result != "!@#$%^&*()" {
		t.Errorf("expected '!@#$%%^&*()', got '%s'", *result)
	}
}

func TestToPtr_Newlines(t *testing.T) {
	input := "line1\nline2"

	result := ToPtr(input)
	if result == nil {
		t.Fatal("expected non-nil pointer")
	}
	if *result != "line1\nline2" {
		t.Errorf("expected 'line1\\nline2', got '%s'", *result)
	}
}

func TestToPtr_NullChar(t *testing.T) {
	// A string with null chars is non-empty
	input := "\x00"

	result := ToPtr(input)
	if result == nil {
		t.Fatal("expected non-nil pointer for null char string")
	}
	if *result != "\x00" {
		t.Error("expected null char string")
	}
}

func TestToPtr_LongString(t *testing.T) {
	// Create a long string
	input := ""
	for i := 0; i < 10000; i++ {
		input += "a"
	}

	result := ToPtr(input)
	if result == nil {
		t.Fatal("expected non-nil pointer")
	}
	if len(*result) != 10000 {
		t.Errorf("expected 10000 chars, got %d", len(*result))
	}
}

func TestToPtr_PointerIsUnique(t *testing.T) {
	input1 := "same"
	input2 := "same"

	result1 := ToPtr(input1)
	result2 := ToPtr(input2)

	// Pointers should be different even for same value
	if result1 == result2 {
		t.Error("expected different pointers for different calls")
	}
	// But values should be equal
	if *result1 != *result2 {
		t.Error("expected equal values")
	}
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkNullTerm_Short(b *testing.B) {
	input := []byte("hello\x00\x00\x00")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NullTerm(input)
	}
}

func BenchmarkNullTerm_Long(b *testing.B) {
	input := make([]byte, 10000)
	for i := 0; i < 9990; i++ {
		input[i] = 'a'
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NullTerm(input)
	}
}

func BenchmarkNullTerm_NoNulls(b *testing.B) {
	input := []byte("hello world without nulls")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NullTerm(input)
	}
}

func BenchmarkToPtr_NonEmpty(b *testing.B) {
	input := "hello"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ToPtr(input)
	}
}

func BenchmarkToPtr_Empty(b *testing.B) {
	input := ""

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ToPtr(input)
	}
}

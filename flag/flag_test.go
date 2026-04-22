package flag

import (
	goflag "flag"
	"testing"
)

// =============================================================================
// StringArr Tests
// =============================================================================

func TestStringArr_Set_CommaSeparated(t *testing.T) {
	var arr StringArr

	err := arr.Set("a,b,c")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(arr) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(arr))
	}
	if arr[0] != "a" || arr[1] != "b" || arr[2] != "c" {
		t.Errorf("unexpected values: %v", arr)
	}
}

func TestStringArr_Set_SingleValue(t *testing.T) {
	var arr StringArr

	err := arr.Set("single")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(arr) != 1 {
		t.Fatalf("expected 1 element, got %d", len(arr))
	}
	if arr[0] != "single" {
		t.Errorf("expected 'single', got '%s'", arr[0])
	}
}

func TestStringArr_Set_Multiple(t *testing.T) {
	var arr StringArr

	arr.Set("first")
	arr.Set("second")
	arr.Set("third")

	if len(arr) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(arr))
	}
	if arr[0] != "first" || arr[1] != "second" || arr[2] != "third" {
		t.Errorf("unexpected values: %v", arr)
	}
}

func TestStringArr_Set_WithSpaces(t *testing.T) {
	var arr StringArr

	err := arr.Set(" a , b , c ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Spaces should be trimmed
	if len(arr) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(arr))
	}
	if arr[0] != "a" || arr[1] != "b" || arr[2] != "c" {
		t.Errorf("expected trimmed values, got: %v", arr)
	}
}

func TestStringArr_Set_EmptyParts(t *testing.T) {
	var arr StringArr

	err := arr.Set("a,,b,  ,c")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Empty parts should be skipped
	if len(arr) != 3 {
		t.Fatalf("expected 3 non-empty elements, got %d: %v", len(arr), arr)
	}
	if arr[0] != "a" || arr[1] != "b" || arr[2] != "c" {
		t.Errorf("unexpected values: %v", arr)
	}
}

func TestStringArr_String(t *testing.T) {
	arr := StringArr{"a", "b", "c"}

	result := arr.String()
	if result != "a,b,c" {
		t.Errorf("expected 'a,b,c', got '%s'", result)
	}
}

func TestStringArr_String_Empty(t *testing.T) {
	arr := StringArr{}

	result := arr.String()
	if result != "" {
		t.Errorf("expected empty string, got '%s'", result)
	}
}

func TestStringArr_String_Single(t *testing.T) {
	arr := StringArr{"single"}

	result := arr.String()
	if result != "single" {
		t.Errorf("expected 'single', got '%s'", result)
	}
}

func TestStringArrayParser_WithFlagSet(t *testing.T) {
	fs := goflag.NewFlagSet("test", goflag.ContinueOnError)
	parser := StringArrayParser(fs)

	// Pass nil default since we're testing flag parsing
	result := parser("hosts", nil, "host list")

	fs.Parse([]string{"-hosts", "host1", "-hosts", "host2"})

	if len(*result) != 2 {
		t.Fatalf("expected 2 elements, got %d: %v", len(*result), *result)
	}
	if (*result)[0] != "host1" || (*result)[1] != "host2" {
		t.Errorf("unexpected values: %v", *result)
	}
}

func TestStringArrayParser_DefaultValue(t *testing.T) {
	fs := goflag.NewFlagSet("test", goflag.ContinueOnError)
	parser := StringArrayParser(fs)

	result := parser("hosts", StringArr{"default1", "default2"}, "host list")

	fs.Parse([]string{})

	if len(*result) != 2 {
		t.Fatalf("expected 2 default elements, got %d: %v", len(*result), *result)
	}
	if (*result)[0] != "default1" || (*result)[1] != "default2" {
		t.Errorf("unexpected default values: %v", *result)
	}
}

func TestStringArrayParser_NilFlagSet(t *testing.T) {
	// This will use global flagset - just make sure it doesn't panic
	parser := StringArrayParser()
	result := parser("test-global-str", nil, "test")
	// Can't easily test global flagset parsing, just verify it returns something
	if result == nil {
		t.Error("expected non-nil result")
	}
}

// =============================================================================
// IntArr Tests
// =============================================================================

func TestIntArr_Set_CommaSeparated(t *testing.T) {
	var arr IntArr

	err := arr.Set("1,2,3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(arr) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(arr))
	}
	if arr[0] != 1 || arr[1] != 2 || arr[2] != 3 {
		t.Errorf("unexpected values: %v", arr)
	}
}

func TestIntArr_Set_SingleValue(t *testing.T) {
	var arr IntArr

	err := arr.Set("42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(arr) != 1 {
		t.Fatalf("expected 1 element, got %d", len(arr))
	}
	if arr[0] != 42 {
		t.Errorf("expected 42, got %d", arr[0])
	}
}

func TestIntArr_Set_Multiple(t *testing.T) {
	var arr IntArr

	arr.Set("10")
	arr.Set("20")
	arr.Set("30")

	if len(arr) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(arr))
	}
	if arr[0] != 10 || arr[1] != 20 || arr[2] != 30 {
		t.Errorf("unexpected values: %v", arr)
	}
}

func TestIntArr_Set_WithSpaces(t *testing.T) {
	var arr IntArr

	err := arr.Set(" 1 , 2 , 3 ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(arr) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(arr))
	}
	if arr[0] != 1 || arr[1] != 2 || arr[2] != 3 {
		t.Errorf("unexpected values: %v", arr)
	}
}

func TestIntArr_Set_Negative(t *testing.T) {
	var arr IntArr

	err := arr.Set("-1,-2,-3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(arr) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(arr))
	}
	if arr[0] != -1 || arr[1] != -2 || arr[2] != -3 {
		t.Errorf("unexpected values: %v", arr)
	}
}

func TestIntArr_Set_Invalid(t *testing.T) {
	var arr IntArr

	err := arr.Set("not-a-number")
	if err == nil {
		t.Error("expected error for invalid integer")
	}
}

func TestIntArr_Set_InvalidInComma(t *testing.T) {
	var arr IntArr

	err := arr.Set("1,invalid,3")
	if err == nil {
		t.Error("expected error for invalid integer in comma-separated")
	}
}

func TestIntArr_String(t *testing.T) {
	arr := IntArr{1, 2, 3}

	result := arr.String()
	if result != "1,2,3" {
		t.Errorf("expected '1,2,3', got '%s'", result)
	}
}

func TestIntArr_String_Empty(t *testing.T) {
	arr := IntArr{}

	result := arr.String()
	if result != "" {
		t.Errorf("expected empty string, got '%s'", result)
	}
}

func TestIntArr_String_Negative(t *testing.T) {
	arr := IntArr{-1, -2, -3}

	result := arr.String()
	if result != "-1,-2,-3" {
		t.Errorf("expected '-1,-2,-3', got '%s'", result)
	}
}

func TestIntArrayParser_WithFlagSet(t *testing.T) {
	fs := goflag.NewFlagSet("test", goflag.ContinueOnError)
	parser := IntArrayParser(fs)

	// Pass nil default since we're testing flag parsing
	result := parser("ports", nil, "port list")

	fs.Parse([]string{"-ports", "80", "-ports", "443"})

	if len(*result) != 2 {
		t.Fatalf("expected 2 elements, got %d: %v", len(*result), *result)
	}
	if (*result)[0] != 80 || (*result)[1] != 443 {
		t.Errorf("unexpected values: %v", *result)
	}
}

func TestIntArrayParser_DefaultValue(t *testing.T) {
	fs := goflag.NewFlagSet("test", goflag.ContinueOnError)
	parser := IntArrayParser(fs)

	result := parser("ports", IntArr{8080, 8081}, "port list")

	fs.Parse([]string{})

	if len(*result) != 2 {
		t.Fatalf("expected 2 default elements, got %d: %v", len(*result), *result)
	}
}

// =============================================================================
// FloatArr Tests
// =============================================================================

func TestFloatArr_Set_CommaSeparated(t *testing.T) {
	var arr FloatArr

	err := arr.Set("1.1,2.2,3.3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(arr) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(arr))
	}
	if arr[0] != 1.1 || arr[1] != 2.2 || arr[2] != 3.3 {
		t.Errorf("unexpected values: %v", arr)
	}
}

func TestFloatArr_Set_SingleValue(t *testing.T) {
	var arr FloatArr

	err := arr.Set("3.14159")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(arr) != 1 {
		t.Fatalf("expected 1 element, got %d", len(arr))
	}
	if arr[0] != 3.14159 {
		t.Errorf("expected 3.14159, got %f", arr[0])
	}
}

func TestFloatArr_Set_Multiple(t *testing.T) {
	var arr FloatArr

	arr.Set("1.0")
	arr.Set("2.0")
	arr.Set("3.0")

	if len(arr) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(arr))
	}
	if arr[0] != 1.0 || arr[1] != 2.0 || arr[2] != 3.0 {
		t.Errorf("unexpected values: %v", arr)
	}
}

func TestFloatArr_Set_WithSpaces(t *testing.T) {
	var arr FloatArr

	err := arr.Set(" 1.0 , 2.0 , 3.0 ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(arr) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(arr))
	}
}

func TestFloatArr_Set_Integers(t *testing.T) {
	var arr FloatArr

	err := arr.Set("1,2,3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(arr) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(arr))
	}
	if arr[0] != 1.0 || arr[1] != 2.0 || arr[2] != 3.0 {
		t.Errorf("unexpected values: %v", arr)
	}
}

func TestFloatArr_Set_Scientific(t *testing.T) {
	var arr FloatArr

	err := arr.Set("1e10,2.5e-3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(arr) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(arr))
	}
	if arr[0] != 1e10 || arr[1] != 2.5e-3 {
		t.Errorf("unexpected values: %v", arr)
	}
}

func TestFloatArr_Set_Invalid(t *testing.T) {
	var arr FloatArr

	err := arr.Set("not-a-number")
	if err == nil {
		t.Error("expected error for invalid float")
	}
}

func TestFloatArr_Set_InvalidInComma(t *testing.T) {
	var arr FloatArr

	err := arr.Set("1.0,invalid,3.0")
	if err == nil {
		t.Error("expected error for invalid float in comma-separated")
	}
}

func TestFloatArr_String(t *testing.T) {
	arr := FloatArr{1.5, 2.5, 3.5}

	result := arr.String()
	if result != "1.5,2.5,3.5" {
		t.Errorf("expected '1.5,2.5,3.5', got '%s'", result)
	}
}

func TestFloatArr_String_Empty(t *testing.T) {
	arr := FloatArr{}

	result := arr.String()
	if result != "" {
		t.Errorf("expected empty string, got '%s'", result)
	}
}

func TestFloatArr_String_Integer(t *testing.T) {
	arr := FloatArr{1.0, 2.0, 3.0}

	result := arr.String()
	if result != "1,2,3" {
		t.Errorf("expected '1,2,3', got '%s'", result)
	}
}

func TestFloatArrayParser_WithFlagSet(t *testing.T) {
	fs := goflag.NewFlagSet("test", goflag.ContinueOnError)
	parser := FloatArrayParser(fs)

	// Pass nil default since we're testing flag parsing
	result := parser("weights", nil, "weights")

	fs.Parse([]string{"-weights", "0.5", "-weights", "0.3"})

	if len(*result) != 2 {
		t.Fatalf("expected 2 elements, got %d: %v", len(*result), *result)
	}
	if (*result)[0] != 0.5 || (*result)[1] != 0.3 {
		t.Errorf("unexpected values: %v", *result)
	}
}

func TestFloatArrayParser_DefaultValue(t *testing.T) {
	fs := goflag.NewFlagSet("test", goflag.ContinueOnError)
	parser := FloatArrayParser(fs)

	result := parser("weights", FloatArr{0.5, 0.3, 0.2}, "weights")

	fs.Parse([]string{})

	if len(*result) != 3 {
		t.Fatalf("expected 3 default elements, got %d: %v", len(*result), *result)
	}
}

// =============================================================================
// BoolArr Tests
// =============================================================================

func TestBoolArr_Set_CommaSeparated(t *testing.T) {
	var arr BoolArr

	err := arr.Set("true,false,true")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(arr) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(arr))
	}
	if arr[0] != true || arr[1] != false || arr[2] != true {
		t.Errorf("unexpected values: %v", arr)
	}
}

func TestBoolArr_Set_SingleValue(t *testing.T) {
	var arr BoolArr

	err := arr.Set("true")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(arr) != 1 {
		t.Fatalf("expected 1 element, got %d", len(arr))
	}
	if arr[0] != true {
		t.Error("expected true")
	}
}

func TestBoolArr_Set_Multiple(t *testing.T) {
	var arr BoolArr

	arr.Set("true")
	arr.Set("false")
	arr.Set("true")

	if len(arr) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(arr))
	}
	if arr[0] != true || arr[1] != false || arr[2] != true {
		t.Errorf("unexpected values: %v", arr)
	}
}

func TestBoolArr_Set_Numeric(t *testing.T) {
	var arr BoolArr

	err := arr.Set("1,0,1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(arr) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(arr))
	}
	if arr[0] != true || arr[1] != false || arr[2] != true {
		t.Errorf("unexpected values: %v", arr)
	}
}

func TestBoolArr_Set_CaseInsensitive(t *testing.T) {
	var arr BoolArr

	err := arr.Set("TRUE,FALSE,True,False")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(arr) != 4 {
		t.Fatalf("expected 4 elements, got %d", len(arr))
	}
	if arr[0] != true || arr[1] != false || arr[2] != true || arr[3] != false {
		t.Errorf("unexpected values: %v", arr)
	}
}

func TestBoolArr_Set_WithSpaces(t *testing.T) {
	var arr BoolArr

	err := arr.Set(" true , false , true ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(arr) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(arr))
	}
}

func TestBoolArr_Set_Invalid(t *testing.T) {
	var arr BoolArr

	err := arr.Set("yes")
	if err == nil {
		t.Error("expected error for invalid bool")
	}
}

func TestBoolArr_Set_InvalidInComma(t *testing.T) {
	var arr BoolArr

	err := arr.Set("true,invalid,false")
	if err == nil {
		t.Error("expected error for invalid bool in comma-separated")
	}
}

func TestBoolArr_String(t *testing.T) {
	arr := BoolArr{true, false, true}

	result := arr.String()
	if result != "true,false,true" {
		t.Errorf("expected 'true,false,true', got '%s'", result)
	}
}

func TestBoolArr_String_Empty(t *testing.T) {
	arr := BoolArr{}

	result := arr.String()
	if result != "" {
		t.Errorf("expected empty string, got '%s'", result)
	}
}

func TestBoolArr_String_Single(t *testing.T) {
	arr := BoolArr{true}

	result := arr.String()
	if result != "true" {
		t.Errorf("expected 'true', got '%s'", result)
	}
}

func TestBoolArrayParser_WithFlagSet(t *testing.T) {
	fs := goflag.NewFlagSet("test", goflag.ContinueOnError)
	parser := BoolArrayParser(fs)

	// Pass nil default since we're testing flag parsing
	result := parser("flags", nil, "flags")

	fs.Parse([]string{"-flags", "true", "-flags", "false"})

	if len(*result) != 2 {
		t.Fatalf("expected 2 elements, got %d: %v", len(*result), *result)
	}
	if (*result)[0] != true || (*result)[1] != false {
		t.Errorf("unexpected values: %v", *result)
	}
}

func TestBoolArrayParser_DefaultValue(t *testing.T) {
	fs := goflag.NewFlagSet("test", goflag.ContinueOnError)
	parser := BoolArrayParser(fs)

	result := parser("flags", BoolArr{true, false}, "flags")

	fs.Parse([]string{})

	if len(*result) != 2 {
		t.Fatalf("expected 2 default elements, got %d: %v", len(*result), *result)
	}
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkStringArr_Set_Single(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var arr StringArr
		arr.Set("value")
	}
}

func BenchmarkStringArr_Set_CommaSeparated(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var arr StringArr
		arr.Set("a,b,c,d,e")
	}
}

func BenchmarkIntArr_Set_Single(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var arr IntArr
		arr.Set("42")
	}
}

func BenchmarkIntArr_Set_CommaSeparated(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var arr IntArr
		arr.Set("1,2,3,4,5")
	}
}

func BenchmarkFloatArr_Set_Single(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var arr FloatArr
		arr.Set("3.14")
	}
}

func BenchmarkFloatArr_Set_CommaSeparated(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var arr FloatArr
		arr.Set("1.0,2.0,3.0,4.0,5.0")
	}
}

func BenchmarkBoolArr_Set_Single(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var arr BoolArr
		arr.Set("true")
	}
}

func BenchmarkBoolArr_Set_CommaSeparated(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var arr BoolArr
		arr.Set("true,false,true,false,true")
	}
}

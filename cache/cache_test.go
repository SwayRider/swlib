package cache

import (
	"sync"
	"testing"
)

// Test cache keys
const (
	testKey1 LocalCacheKey = "test-key-1"
	testKey2 LocalCacheKey = "test-key-2"
	testKey3 LocalCacheKey = "test-key-3"
)

// Helper to clear cache between tests
func clearCache() {
	llock.Lock()
	defer llock.Unlock()
	lcache = make(map[LocalCacheKey]any)
}

// =============================================================================
// Basic Operations Tests
// =============================================================================

func TestLCSet_AndGet(t *testing.T) {
	clearCache()

	LCSet(testKey1, "test-value")

	value, ok := LCGet(testKey1)
	if !ok {
		t.Fatal("expected key to exist")
	}
	if value != "test-value" {
		t.Errorf("expected 'test-value', got '%v'", value)
	}
}

func TestLCGet_NonExistent(t *testing.T) {
	clearCache()

	value, ok := LCGet("non-existent-key")
	if ok {
		t.Error("expected ok to be false for non-existent key")
	}
	if value != nil {
		t.Errorf("expected nil value, got '%v'", value)
	}
}

func TestLCSet_Overwrite(t *testing.T) {
	clearCache()

	LCSet(testKey1, "first-value")
	LCSet(testKey1, "second-value")

	value, ok := LCGet(testKey1)
	if !ok {
		t.Fatal("expected key to exist")
	}
	if value != "second-value" {
		t.Errorf("expected 'second-value', got '%v'", value)
	}
}

func TestLCHas_Exists(t *testing.T) {
	clearCache()

	LCSet(testKey1, "value")

	if !LCHas(testKey1) {
		t.Error("expected LCHas to return true for existing key")
	}
}

func TestLCHas_NotExists(t *testing.T) {
	clearCache()

	if LCHas("non-existent") {
		t.Error("expected LCHas to return false for non-existent key")
	}
}

func TestLCDel(t *testing.T) {
	clearCache()

	LCSet(testKey1, "value")
	if !LCHas(testKey1) {
		t.Fatal("expected key to exist before delete")
	}

	LCDel(testKey1)

	if LCHas(testKey1) {
		t.Error("expected key to not exist after delete")
	}

	value, ok := LCGet(testKey1)
	if ok {
		t.Error("expected ok to be false after delete")
	}
	if value != nil {
		t.Error("expected nil value after delete")
	}
}

func TestLCDel_NonExistent(t *testing.T) {
	clearCache()

	// Should not panic
	LCDel("non-existent-key")
}

// =============================================================================
// Different Value Types Tests
// =============================================================================

func TestLCSet_StringValue(t *testing.T) {
	clearCache()

	LCSet(testKey1, "string-value")

	value, ok := LCGet(testKey1)
	if !ok {
		t.Fatal("expected key to exist")
	}
	str, isString := value.(string)
	if !isString {
		t.Fatal("expected string type")
	}
	if str != "string-value" {
		t.Errorf("expected 'string-value', got '%s'", str)
	}
}

func TestLCSet_IntValue(t *testing.T) {
	clearCache()

	LCSet(testKey1, 42)

	value, ok := LCGet(testKey1)
	if !ok {
		t.Fatal("expected key to exist")
	}
	num, isInt := value.(int)
	if !isInt {
		t.Fatal("expected int type")
	}
	if num != 42 {
		t.Errorf("expected 42, got %d", num)
	}
}

func TestLCSet_BoolValue(t *testing.T) {
	clearCache()

	LCSet(testKey1, true)

	value, ok := LCGet(testKey1)
	if !ok {
		t.Fatal("expected key to exist")
	}
	b, isBool := value.(bool)
	if !isBool {
		t.Fatal("expected bool type")
	}
	if !b {
		t.Error("expected true")
	}
}

func TestLCSet_SliceValue(t *testing.T) {
	clearCache()

	slice := []string{"a", "b", "c"}
	LCSet(testKey1, slice)

	value, ok := LCGet(testKey1)
	if !ok {
		t.Fatal("expected key to exist")
	}
	result, isSlice := value.([]string)
	if !isSlice {
		t.Fatal("expected []string type")
	}
	if len(result) != 3 {
		t.Errorf("expected 3 elements, got %d", len(result))
	}
	if result[0] != "a" || result[1] != "b" || result[2] != "c" {
		t.Errorf("unexpected slice values: %v", result)
	}
}

func TestLCSet_MapValue(t *testing.T) {
	clearCache()

	m := map[string]int{"one": 1, "two": 2}
	LCSet(testKey1, m)

	value, ok := LCGet(testKey1)
	if !ok {
		t.Fatal("expected key to exist")
	}
	result, isMap := value.(map[string]int)
	if !isMap {
		t.Fatal("expected map[string]int type")
	}
	if result["one"] != 1 || result["two"] != 2 {
		t.Errorf("unexpected map values: %v", result)
	}
}

func TestLCSet_StructValue(t *testing.T) {
	clearCache()

	type TestStruct struct {
		ID   string
		Name string
		Age  int
	}

	s := TestStruct{ID: "123", Name: "John", Age: 30}
	LCSet(testKey1, s)

	value, ok := LCGet(testKey1)
	if !ok {
		t.Fatal("expected key to exist")
	}
	result, isStruct := value.(TestStruct)
	if !isStruct {
		t.Fatal("expected TestStruct type")
	}
	if result.ID != "123" || result.Name != "John" || result.Age != 30 {
		t.Errorf("unexpected struct values: %+v", result)
	}
}

func TestLCSet_PointerValue(t *testing.T) {
	clearCache()

	type TestStruct struct {
		ID string
	}

	s := &TestStruct{ID: "ptr-123"}
	LCSet(testKey1, s)

	value, ok := LCGet(testKey1)
	if !ok {
		t.Fatal("expected key to exist")
	}
	result, isPtr := value.(*TestStruct)
	if !isPtr {
		t.Fatal("expected *TestStruct type")
	}
	if result.ID != "ptr-123" {
		t.Errorf("expected ID 'ptr-123', got '%s'", result.ID)
	}

	// Verify it's the same pointer
	if result != s {
		t.Error("expected same pointer reference")
	}
}

func TestLCSet_NilValue(t *testing.T) {
	clearCache()

	LCSet(testKey1, nil)

	value, ok := LCGet(testKey1)
	if !ok {
		t.Fatal("expected key to exist")
	}
	if value != nil {
		t.Errorf("expected nil value, got '%v'", value)
	}
}

// =============================================================================
// Multiple Keys Tests
// =============================================================================

func TestMultipleKeys(t *testing.T) {
	clearCache()

	LCSet(testKey1, "value1")
	LCSet(testKey2, "value2")
	LCSet(testKey3, "value3")

	v1, _ := LCGet(testKey1)
	v2, _ := LCGet(testKey2)
	v3, _ := LCGet(testKey3)

	if v1 != "value1" {
		t.Errorf("expected 'value1', got '%v'", v1)
	}
	if v2 != "value2" {
		t.Errorf("expected 'value2', got '%v'", v2)
	}
	if v3 != "value3" {
		t.Errorf("expected 'value3', got '%v'", v3)
	}
}

func TestDeleteOneKey_OthersRemain(t *testing.T) {
	clearCache()

	LCSet(testKey1, "value1")
	LCSet(testKey2, "value2")

	LCDel(testKey1)

	if LCHas(testKey1) {
		t.Error("expected testKey1 to be deleted")
	}
	if !LCHas(testKey2) {
		t.Error("expected testKey2 to still exist")
	}
}

// =============================================================================
// Thread Safety Tests
// =============================================================================

func TestConcurrentSet(t *testing.T) {
	clearCache()

	var wg sync.WaitGroup
	iterations := 1000

	// Concurrent writes to different keys
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			key := LocalCacheKey("concurrent-key-" + string(rune('0'+n%10)))
			LCSet(key, n)
		}(i)
	}

	wg.Wait()

	// Verify no panics occurred and cache is in valid state
	for i := 0; i < 10; i++ {
		key := LocalCacheKey("concurrent-key-" + string(rune('0'+i)))
		if !LCHas(key) {
			t.Errorf("expected key '%s' to exist", key)
		}
	}
}

func TestConcurrentSetSameKey(t *testing.T) {
	clearCache()

	var wg sync.WaitGroup
	iterations := 1000
	key := LocalCacheKey("same-key")

	// Concurrent writes to the same key
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			LCSet(key, n)
		}(i)
	}

	wg.Wait()

	// Verify key exists with some value
	value, ok := LCGet(key)
	if !ok {
		t.Fatal("expected key to exist")
	}
	_, isInt := value.(int)
	if !isInt {
		t.Error("expected int value")
	}
}

func TestConcurrentGet(t *testing.T) {
	clearCache()

	LCSet(testKey1, "test-value")

	var wg sync.WaitGroup
	iterations := 1000

	// Concurrent reads
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			value, ok := LCGet(testKey1)
			if !ok {
				t.Error("expected key to exist")
			}
			if value != "test-value" {
				t.Errorf("expected 'test-value', got '%v'", value)
			}
		}()
	}

	wg.Wait()
}

func TestConcurrentHas(t *testing.T) {
	clearCache()

	LCSet(testKey1, "value")

	var wg sync.WaitGroup
	iterations := 1000

	// Concurrent Has checks
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if !LCHas(testKey1) {
				t.Error("expected key to exist")
			}
		}()
	}

	wg.Wait()
}

func TestConcurrentSetGetDel(t *testing.T) {
	clearCache()

	var wg sync.WaitGroup
	iterations := 100

	// Mix of concurrent operations
	for i := 0; i < iterations; i++ {
		wg.Add(3)

		// Writer
		go func(n int) {
			defer wg.Done()
			LCSet(testKey1, n)
		}(i)

		// Reader
		go func() {
			defer wg.Done()
			LCGet(testKey1)
		}()

		// Deleter (occasionally)
		go func(n int) {
			defer wg.Done()
			if n%10 == 0 {
				LCDel(testKey1)
			}
		}(i)
	}

	wg.Wait()
	// Test passes if no race conditions or panics
}

// =============================================================================
// LocalCacheKey Type Tests
// =============================================================================

func TestLocalCacheKey_DifferentKeys(t *testing.T) {
	clearCache()

	key1 := LocalCacheKey("user:123")
	key2 := LocalCacheKey("user:456")

	LCSet(key1, "user-123-data")
	LCSet(key2, "user-456-data")

	v1, _ := LCGet(key1)
	v2, _ := LCGet(key2)

	if v1 != "user-123-data" {
		t.Errorf("expected 'user-123-data', got '%v'", v1)
	}
	if v2 != "user-456-data" {
		t.Errorf("expected 'user-456-data', got '%v'", v2)
	}
}

func TestLocalCacheKey_EmptyString(t *testing.T) {
	clearCache()

	emptyKey := LocalCacheKey("")
	LCSet(emptyKey, "empty-key-value")

	value, ok := LCGet(emptyKey)
	if !ok {
		t.Fatal("expected empty key to work")
	}
	if value != "empty-key-value" {
		t.Errorf("expected 'empty-key-value', got '%v'", value)
	}
}

func TestLocalCacheKey_SpecialCharacters(t *testing.T) {
	clearCache()

	specialKey := LocalCacheKey("key:with/special!@#$%^&*()characters")
	LCSet(specialKey, "special-value")

	value, ok := LCGet(specialKey)
	if !ok {
		t.Fatal("expected special key to work")
	}
	if value != "special-value" {
		t.Errorf("expected 'special-value', got '%v'", value)
	}
}

func TestLocalCacheKey_Unicode(t *testing.T) {
	clearCache()

	unicodeKey := LocalCacheKey("キー日本語")
	LCSet(unicodeKey, "unicode-value")

	value, ok := LCGet(unicodeKey)
	if !ok {
		t.Fatal("expected unicode key to work")
	}
	if value != "unicode-value" {
		t.Errorf("expected 'unicode-value', got '%v'", value)
	}
}

// =============================================================================
// Edge Cases
// =============================================================================

func TestLCGet_AfterClear(t *testing.T) {
	clearCache()

	LCSet(testKey1, "value")
	clearCache()

	_, ok := LCGet(testKey1)
	if ok {
		t.Error("expected key to not exist after clear")
	}
}

func TestLCSet_LargeValue(t *testing.T) {
	clearCache()

	// Create a large slice
	largeSlice := make([]byte, 1024*1024) // 1MB
	for i := range largeSlice {
		largeSlice[i] = byte(i % 256)
	}

	LCSet(testKey1, largeSlice)

	value, ok := LCGet(testKey1)
	if !ok {
		t.Fatal("expected key to exist")
	}
	result, isSlice := value.([]byte)
	if !isSlice {
		t.Fatal("expected []byte type")
	}
	if len(result) != 1024*1024 {
		t.Errorf("expected 1MB slice, got %d bytes", len(result))
	}
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkLCSet(b *testing.B) {
	clearCache()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		LCSet(testKey1, "benchmark-value")
	}
}

func BenchmarkLCGet(b *testing.B) {
	clearCache()
	LCSet(testKey1, "benchmark-value")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		LCGet(testKey1)
	}
}

func BenchmarkLCHas(b *testing.B) {
	clearCache()
	LCSet(testKey1, "benchmark-value")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		LCHas(testKey1)
	}
}

func BenchmarkLCDel(b *testing.B) {
	clearCache()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		LCSet(testKey1, "value")
		LCDel(testKey1)
	}
}

func BenchmarkConcurrentReads(b *testing.B) {
	clearCache()
	LCSet(testKey1, "benchmark-value")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			LCGet(testKey1)
		}
	})
}

func BenchmarkConcurrentWrites(b *testing.B) {
	clearCache()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			LCSet(testKey1, "benchmark-value")
		}
	})
}

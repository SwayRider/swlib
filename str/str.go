// Package str provides simple string manipulation utilities.
package str

import "strings"

// NullTerm removes null terminators (\x00) from the end of a byte slice
// and returns the result as a string. This is useful when working with
// C-style null-terminated strings or binary data.
//
// Example:
//
//	data := []byte("hello\x00\x00\x00")
//	cleaned := str.NullTerm(data) // "hello"
func NullTerm(bytes []byte) string {
	return strings.TrimRight(string(bytes), "\x00")
}

// ToPtr converts a string to a pointer. Returns nil for empty strings.
// This is useful for optional string fields in structs or API requests.
//
// Example:
//
//	name := "John"
//	namePtr := str.ToPtr(name) // *string pointing to "John"
//
//	empty := ""
//	emptyPtr := str.ToPtr(empty) // nil
func ToPtr(in string) *string {
	if len(in) == 0 {
		return nil
	}
	return &in
}

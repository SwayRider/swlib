// Package env provides utilities for reading environment variables with
// type conversion and fallback values.
//
// All functions return a fallback value if the environment variable is not set
// or cannot be parsed to the requested type.
//
// # Usage
//
//	port := env.GetAsInt("PORT", 8080)
//	debug := env.GetAsBool("DEBUG", false)
//	hosts := env.GetAsStringArr("HOSTS", "localhost")
package env

import (
	"os"
	"strconv"
	"strings"
)

// Get returns the value of the environment variable with the given key, or
// the fallback value if the key is not set
//
// Parameters:
//   - key: The key of the environment variable
//   - fallback: The fallback value if the key is not set
//
// Returns:
//   - string: The value of the environment variable
func Get(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

// GetAsStringArr returns the value of the environment variable as a string slice.
// Values are expected to be comma-separated. Returns the fallback split by commas
// if the environment variable is not set.
//
// Example:
//
//	// HOSTS=host1,host2,host3
//	hosts := env.GetAsStringArr("HOSTS", "localhost")
func GetAsStringArr(key, fallback string) []string {
	valStr := os.Getenv(key)
	if valStr != "" {
		return strings.Split(valStr, ",")
	}
	return strings.Split(fallback, ",")
}

// GetAsInt returns the value of the environment variable with the given key,
// or the fallback value if the key is not set
//
// Parameters:
//   - key: The key of the environment variable
//   - fallback: The fallback value if the key is not set
//
// Returns:
//   - int: The value of the environment variable
func GetAsInt(key string, fallback int) int {
	valStr := os.Getenv(key)
	if val, err := strconv.Atoi(valStr); err == nil {
		return val
	}
	return fallback
}

// GetAsIntArr returns the value of the environment variable as an integer slice.
// Values are expected to be comma-separated integers. Invalid integers are skipped.
//
// Example:
//
//	// PORTS=8080,8081,8082
//	ports := env.GetAsIntArr("PORTS", "8080")
func GetAsIntArr(key string, fallback string) []int {
	strArr := GetAsStringArr(key, fallback)
	intArr := make([]int, 0, len(strArr))
	for _, str := range strArr {
		if val, err := strconv.Atoi(str); err == nil {
			intArr = append(intArr, val)
		}
	}
	return intArr
}

// GetAsFloat64 returns the value of the environment variable as a float64.
// Returns the fallback value if the variable is not set or cannot be parsed.
//
// Example:
//
//	// TIMEOUT=30.5
//	timeout := env.GetAsFloat64("TIMEOUT", 10.0)
func GetAsFloat64(key string, fallback float64) float64 {
	valStr := os.Getenv(key)
	if val, err := strconv.ParseFloat(valStr, 64); err == nil {
		return val
	}
	return fallback
}

// GetAsBool returns the value of the environment variable with the given key,
// or the fallback value if the key is not set
//
// Parameters:
//   - key: The key of the environment variable
//   - fallback: The fallback value if the key is not set
//
// Returns:
//   - bool: The value of the environment variable
func GetAsBool(key string, fallback bool) bool {
	valStr := os.Getenv(key)
	if val, err := strconv.ParseBool(valStr); err == nil {
		return val
	}
	return fallback
}

// GetAsFloat64Arr returns the value of the environment variable as a float64 slice.
// Values are expected to be comma-separated floats. Invalid floats are skipped.
//
// Example:
//
//	// WEIGHTS=0.5,0.3,0.2
//	weights := env.GetAsFloat64Arr("WEIGHTS", "1.0")
func GetAsFloat64Arr(key string, fallback string) []float64 {
	strArr := GetAsStringArr(key, fallback)
	floatArr := make([]float64, 0, len(strArr))
	for _, str := range strArr {
		if val, err := strconv.ParseFloat(str, 64); err == nil {
			floatArr = append(floatArr, val)
		}
	}
	return floatArr
}

// GetAsBoolArr returns the value of the environment variable as a boolean slice.
// Values are expected to be comma-separated booleans. Invalid booleans are skipped.
//
// Example:
//
//	// FEATURES=true,false,true
//	features := env.GetAsBoolArr("FEATURES", "false")
func GetAsBoolArr(key string, fallback string) []bool {
	strArr := GetAsStringArr(key, fallback)
	boolArr := make([]bool, 0, len(strArr))
	for _, str := range strArr {
		if val, err := strconv.ParseBool(str); err == nil {
			boolArr = append(boolArr, val)
		}
	}
	return boolArr
}

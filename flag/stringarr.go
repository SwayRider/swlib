// Package flag provides custom flag types for parsing array values from
// command-line arguments. These types implement the flag.Value interface
// and support both comma-separated values and multiple flag occurrences.
//
// # Usage
//
// Single flag with comma-separated values:
//
//	./myapp -hosts "host1,host2,host3"
//
// Multiple flag occurrences:
//
//	./myapp -hosts host1 -hosts host2 -hosts host3
//
// # Available Types
//
//   - StringArr: String array flag
//   - IntArr: Integer array flag
//   - FloatArr: Float64 array flag
//   - BoolArr: Boolean array flag
//
// # Example
//
//	var hosts flag.StringArr
//	goflag.Var(&hosts, "hosts", "Host addresses")
//	goflag.Parse()
//	for _, host := range hosts {
//	    fmt.Println(host)
//	}
package flag

import (
	goflag "flag"
	"strings"
)

// StringArr is a string slice that implements the flag.Value interface.
// It supports both comma-separated values and multiple flag occurrences.
type StringArr []string

// Set implements flag.Value interface. Parses comma-separated values
// or appends a single value to the array.
func (s *StringArr) Set(value string) error {
	if strings.Contains(value, ",") {
		for part := range strings.SplitSeq(value, ",") {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				*s = append(*s, trimmed)
			}
		}
		return nil
	}

	*s = append(*s, value)
	return nil
}

// String implements flag.Value interface. Returns comma-separated representation.
func (s StringArr) String() string {
	return strings.Join(s, ",")
}

// StringArrayParser returns a function that creates and registers a StringArr flag.
// This is useful when working with custom FlagSets or the app configuration system.
//
// Example:
//
//	parseStringArr := flag.StringArrayParser()
//	hosts := parseStringArr(fs, "hosts", nil, "Host addresses")
func StringArrayParser(flagset ...*goflag.FlagSet) func(string, StringArr, string) *StringArr {
	return func(name string, defaultValue StringArr, usage string) *StringArr {
		var res StringArr
		if len(flagset) > 0 && flagset[0] != nil {
			flagset[0].Var(&res, name, usage)
		} else {
			goflag.Var(&res, name, usage)
		}
		if len(res) == 0 {
			res = defaultValue
		}
		return &res
	}
}

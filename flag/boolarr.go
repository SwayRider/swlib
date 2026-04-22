package flag

import (
	goflag "flag"
	"strconv"
	"strings"
)

// BoolArr is a boolean slice that implements the flag.Value interface.
// It supports both comma-separated values and multiple flag occurrences.
type BoolArr []bool

// Set implements flag.Value interface. Parses comma-separated booleans
// or appends a single boolean to the array.
func (b *BoolArr) Set(value string) error {
	if strings.Contains(value, ",") {
		for part := range strings.SplitSeq(value, ",") {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				v, err := strconv.ParseBool(trimmed)
				if err != nil {
					return err
				}
				*b = append(*b, v)
			}
		}
		return nil
	}

	v, err := strconv.ParseBool(value)
	if err != nil {
		return err
	}
	*b = append(*b, v)
	return nil
}

// String implements flag.Value interface. Returns comma-separated representation.
func (b BoolArr) String() string {
	strArr := make([]string, 0, len(b))
	for _, v := range b {
		strArr = append(strArr, strconv.FormatBool(v))
	}
	return strings.Join(strArr, ",")
}

// BoolArrayParser returns a function that creates and registers a BoolArr flag.
func BoolArrayParser(flagset ...*goflag.FlagSet) func(string, BoolArr, string) *BoolArr {
	return func(name string, defaultValue BoolArr, usage string) *BoolArr {
		var res BoolArr
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

package flag

import (
	goflag "flag"
	"strconv"
	"strings"
)

// IntArr is an integer slice that implements the flag.Value interface.
// It supports both comma-separated values and multiple flag occurrences.
type IntArr []int

// Set implements flag.Value interface. Parses comma-separated integers
// or appends a single integer to the array.
func (i *IntArr) Set(value string) error {
	if strings.Contains(value, ",") {
		for part := range strings.SplitSeq(value, ",") {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				v, err := strconv.Atoi(trimmed)
				if err != nil {
					return err
				}
				*i = append(*i, v)
			}
		}
		return nil
	}

	v, err := strconv.Atoi(value)
	if err != nil {
		return err
	}
	*i = append(*i, v)
	return nil
}

// String implements flag.Value interface. Returns comma-separated representation.
func (i IntArr) String() string {
	strArr := make([]string, 0, len(i))
	for _, v := range i {
		strArr = append(strArr, strconv.Itoa(v))
	}
	return strings.Join(strArr, ",")
}

// IntArrayParser returns a function that creates and registers an IntArr flag.
func IntArrayParser(flagset ...*goflag.FlagSet) func(string, IntArr, string) *IntArr {
	return func(name string, defaultValue IntArr, usage string) *IntArr {
		var res IntArr
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

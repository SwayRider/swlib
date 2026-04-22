package flag

import (
	goflag "flag"
	"strconv"
	"strings"
)

// FloatArr is a float64 slice that implements the flag.Value interface.
// It supports both comma-separated values and multiple flag occurrences.
type FloatArr []float64

// Set implements flag.Value interface. Parses comma-separated floats
// or appends a single float to the array.
func (f *FloatArr) Set(value string) error {
	if strings.Contains(value, ",") {
		for part := range strings.SplitSeq(value, ",") {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				v, err := strconv.ParseFloat(trimmed, 64)
				if err != nil {
					return err
				}
				*f = append(*f, v)
			}
		}
		return nil
	}

	v, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return err
	}
	*f = append(*f, v)
	return nil
}

// String implements flag.Value interface. Returns comma-separated representation.
func (f FloatArr) String() string {
	strArr := make([]string, 0, len(f))
	for _, v := range f {
		strArr = append(strArr, strconv.FormatFloat(v, 'f', -1, 64))
	}
	return strings.Join(strArr, ",")
}

// FloatArrayParser returns a function that creates and registers a FloatArr flag.
func FloatArrayParser(flagset ...*goflag.FlagSet) func(string, FloatArr, string) *FloatArr {
	return func(name string, defaultValue FloatArr, usage string) *FloatArr {
		var res FloatArr
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

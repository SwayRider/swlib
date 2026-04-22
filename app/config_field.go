package app

import (
	"flag"
	"fmt"

	"github.com/swayrider/swlib/env"
	flg "github.com/swayrider/swlib/flag"
)

// ConfigFieldValue is a type constraint for valid configuration field value types.
// Supported types include scalar types (string, int, float64, bool) and array
// types (StringArr, IntArr, FloatArr, BoolArr).
type ConfigFieldValue interface {
	string | int | float64 | bool | flg.StringArr | flg.IntArr | flg.FloatArr | flg.BoolArr
}

// ConfigField represents a single configuration field that can be set via
// environment variables or CLI flags. It provides access to the field's
// metadata and current value.
type ConfigField interface {
	// Name returns the CLI flag name (e.g., "http-port")
	Name() string
	// Env returns the environment variable name (e.g., "HTTP_PORT")
	Env() string
	// Description returns the human-readable description of the field
	Description() string
	// Value returns the current value of the field after parsing
	Value() any

	configureFlag()
	setFlagValue()
}

// NewIntConfigField creates a new integer configuration field.
//
// Parameters:
//   - name: CLI flag name (e.g., "http-port")
//   - envVar: Environment variable name (e.g., "HTTP_PORT")
//   - description: Human-readable description for help text
//   - defaultValue: Default value if not set via env or flag
//   - flagset: Optional custom FlagSet (uses default if not provided)
//
// Example:
//
//	field := NewIntConfigField("http-port", "HTTP_PORT", "HTTP server port", 8080)
func NewIntConfigField(
	name string,
	envVar string,
	description string,
	defaultValue int,
	flagset ...*flag.FlagSet,
) ConfigField {
	var ff func(string, int, string) *int

	if len(flagset) > 0 && flagset[0] != nil {
		ff = flagset[0].Int
	} else {
		ff = flag.Int
	}

	return &configField[int]{
		name:         name,
		envVar:       envVar,
		descr:        description,
		defaultValue: defaultValue,

		flagFunc: ff,
		envFunc:  env.GetAsInt,
	}
}

// NewIntArrConfigField creates a new integer array configuration field.
// Values can be provided as comma-separated lists via environment variables
// or as multiple flag occurrences.
//
// Example:
//
//	field := NewIntArrConfigField("ports", "PORTS", "Server ports", []int{8080, 8081})
func NewIntArrConfigField(
	name string,
	envVar string,
	description string,
	devaultValue []int,
	flagset ...*flag.FlagSet,
) ConfigField {
	return &configField[flg.IntArr]{
		name:         name,
		envVar:       envVar,
		descr:        description,
		defaultValue: devaultValue,

		flagFunc: flg.IntArrayParser(flagset...),
		envFunc: func(key string, fallback flg.IntArr) flg.IntArr {
			tmp := env.GetAsIntArr(key, fallback.String())
			return flg.IntArr(tmp)
		},

		isArrField: true,
	}
}

// NewStringConfigField creates a new string configuration field.
//
// Example:
//
//	field := NewStringConfigField("db-host", "DB_HOST", "Database hostname", "localhost")
func NewStringConfigField(
	name string,
	envVar string,
	description string,
	defaultValue string,
	flagset ...*flag.FlagSet,
) ConfigField {
	var ff func(string, string, string) *string

	if len(flagset) > 0 && flagset[0] != nil {
		ff = flagset[0].String
	} else {
		ff = flag.String
	}

	return &configField[string]{
		name:         name,
		envVar:       envVar,
		descr:        description,
		defaultValue: defaultValue,

		flagFunc: ff,
		envFunc:  env.Get,
	}
}

// NewStringArrConfigField creates a new string array configuration field.
// Values can be provided as comma-separated lists via environment variables
// or as multiple flag occurrences.
//
// Example:
//
//	field := NewStringArrConfigField("hosts", "HOSTS", "Allowed hosts", []string{"localhost"})
func NewStringArrConfigField(
	name string,
	envVar string,
	description string,
	defaultValue []string,
	flagset ...*flag.FlagSet,
) ConfigField {
	return &configField[flg.StringArr]{
		name:         name,
		envVar:       envVar,
		descr:        description,
		defaultValue: defaultValue,

		flagFunc: flg.StringArrayParser(flagset...),
		envFunc: func(key string, fallback flg.StringArr) flg.StringArr {
			tmp := env.GetAsStringArr(key, fallback.String())
			return flg.StringArr(tmp)
		},

		isArrField: true,
	}
}

// NewBoolConfigField creates a new boolean configuration field.
// Boolean values can be set via environment variables using "true", "false", "1", "0".
//
// Example:
//
//	field := NewBoolConfigField("debug", "DEBUG", "Enable debug mode", false)
func NewBoolConfigField(
	name string,
	envVar string,
	description string,
	defaultValue bool,
	flagset ...*flag.FlagSet,
) ConfigField {
	var ff func(string, bool, string) *bool

	if len(flagset) > 0 && flagset[0] != nil {
		ff = flagset[0].Bool
	} else {
		ff = flag.Bool
	}

	return &configField[bool]{
		name:         name,
		envVar:       envVar,
		descr:        description,
		defaultValue: defaultValue,

		flagFunc: ff,
		envFunc:  env.GetAsBool,
	}
}

// NewBoolArrConfigField creates a new boolean array configuration field.
// Values can be provided as comma-separated lists via environment variables
// or as multiple flag occurrences.
//
// Example:
//
//	field := NewBoolArrConfigField("features", "FEATURES", "Feature flags", []bool{true, false})
func NewBoolArrConfigField(
	name string,
	envVar string,
	description string,
	devaultValue []bool,
	flagset ...*flag.FlagSet,
) ConfigField {
	return &configField[flg.BoolArr]{
		name:         name,
		envVar:       envVar,
		descr:        description,
		defaultValue: devaultValue,

		flagFunc: flg.BoolArrayParser(flagset...),
		envFunc: func(key string, fallback flg.BoolArr) flg.BoolArr {
			tmp := env.GetAsBoolArr(key, fallback.String())
			return flg.BoolArr(tmp)
		},

		isArrField: true,
	}
}

// NewFloatConfigField creates a new float64 configuration field.
//
// Example:
//
//	field := NewFloatConfigField("timeout", "TIMEOUT", "Request timeout in seconds", 30.0)
func NewFloatConfigField(
	name string,
	envVar string,
	description string,
	defaultValue float64,
	flagset ...*flag.FlagSet,
) ConfigField {
	var ff func(string, float64, string) *float64

	if len(flagset) > 0 && flagset[0] != nil {
		ff = flagset[0].Float64
	} else {
		ff = flag.Float64
	}

	return &configField[float64]{
		name:         name,
		envVar:       envVar,
		descr:        description,
		defaultValue: defaultValue,

		flagFunc: ff,
		envFunc:  env.GetAsFloat64,
	}
}

// NewFloatArrConfigField creates a new float64 array configuration field.
// Values can be provided as comma-separated lists via environment variables
// or as multiple flag occurrences.
//
// Example:
//
//	field := NewFloatArrConfigField("weights", "WEIGHTS", "Feature weights", []float64{0.5, 0.3, 0.2})
func NewFloatArrConfigField(
	name string,
	envVar string,
	description string,
	devaultValue []float64,
	flagset ...*flag.FlagSet,
) ConfigField {
	return &configField[flg.FloatArr]{
		name:         name,
		envVar:       envVar,
		descr:        description,
		defaultValue: devaultValue,

		flagFunc: flg.FloatArrayParser(flagset...),
		envFunc: func(key string, fallback flg.FloatArr) flg.FloatArr {
			tmp := env.GetAsFloat64Arr(key, fallback.String())
			return flg.FloatArr(tmp)
		},

		isArrField: true,
	}
}

type configField[T ConfigFieldValue] struct {
	name         string
	envVar       string
	descr        string
	defaultValue T
	value        T
	flagValue    *T

	flagFunc func(string, T, string) *T
	envFunc  func(string, T) T

	isArrField bool
}

func (f configField[T]) Name() string {
	return f.name
}

func (f configField[T]) Env() string {
	return f.envVar
}

func (f configField[T]) Description() string {
	return f.descr
}

func (f configField[T]) Value() any {
	return f.value
}

func (f *configField[T]) configureFlag() {
	f.defaultValue = f.envFunc(f.envVar, f.defaultValue)
	if f.isArrField {
		var def T
		f.flagValue = f.flagFunc(f.name, def, f.descr)
		return
	}
	f.flagValue = f.flagFunc(f.name, f.defaultValue, f.descr)
}

func (f *configField[T]) setFlagValue() {
	if f.isArrField {
		iface := any(f.flagValue)
		switch vv := iface.(type) {
		case *flg.StringArr:
			if len(*vv) > 0 {
				f.value = *f.flagValue
				return
			}
		case *flg.IntArr:
			if len(*vv) > 0 {
				f.value = *f.flagValue
				return
			}
		case *flg.BoolArr:
			if len(*vv) > 0 {
				f.value = *f.flagValue
				return
			}
		case *flg.FloatArr:
			if len(*vv) > 0 {
				f.value = *f.flagValue
				return
			}
		default:
			panic(fmt.Sprintf("Unsupported array type: %T", f.flagValue))
		}

		f.value = f.defaultValue
		return
	}

	if f.flagValue != nil {
		f.value = *f.flagValue
		return
	}
}

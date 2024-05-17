// Copy from https://github.com/opentofu/opentofu/blob/v1.7.1/internal/lang/funcs/datetime.go
package hclfuncs

import (
	"fmt"
	"github.com/jehiah/go-strftime"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"time"
)

// TimeCmpFunc is a function that compares two timestamps.
var TimeCmpFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name: "timestamp_a",
			Type: cty.String,
		},
		{
			Name: "timestamp_b",
			Type: cty.String,
		},
	},
	Type:         function.StaticReturnType(cty.Number),
	RefineResult: refineNotNull,
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		tsA, err := parseTimestamp(args[0].AsString())
		if err != nil {
			return cty.UnknownVal(cty.String), function.NewArgError(0, err)
		}
		tsB, err := parseTimestamp(args[1].AsString())
		if err != nil {
			return cty.UnknownVal(cty.String), function.NewArgError(1, err)
		}

		switch {
		case tsA.Equal(tsB):
			return cty.NumberIntVal(0), nil
		case tsA.Before(tsB):
			return cty.NumberIntVal(-1), nil
		default:
			// By elimintation, tsA must be after tsB.
			return cty.NumberIntVal(1), nil
		}
	},
})

func parseTimestamp(ts string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		switch err := err.(type) {
		case *time.ParseError:
			// If err is a time.ParseError then its string representation is not
			// appropriate since it relies on details of Go's strange date format
			// representation, which a caller of our functions is not expected
			// to be familiar with.
			//
			// Therefore we do some light transformation to get a more suitable
			// error that should make more sense to our callers. These are
			// still not awesome error messages, but at least they refer to
			// the timestamp portions by name rather than by Go's example
			// values.
			if err.LayoutElem == "" && err.ValueElem == "" && err.Message != "" {
				return time.Time{}, fmt.Errorf("not a valid RFC3339 timestamp: %w", err)
			}
			var what string
			switch err.LayoutElem {
			case "2006":
				what = "year"
			case "01":
				what = "month"
			case "02":
				what = "day of month"
			case "15":
				what = "hour"
			case "04":
				what = "minute"
			case "05":
				what = "second"
			case "Z07:00":
				what = "UTC offset"
			case "T":
				return time.Time{}, fmt.Errorf("not a valid RFC3339 timestamp: missing required time introducer 'T'")
			case ":", "-":
				if err.ValueElem == "" {
					return time.Time{}, fmt.Errorf("not a valid RFC3339 timestamp: end of string where %q is expected", err.LayoutElem)
				} else {
					return time.Time{}, fmt.Errorf("not a valid RFC3339 timestamp: found %q where %q is expected", err.ValueElem, err.LayoutElem)
				}
			default:
				// Should never get here, because time.RFC3339 includes only the
				// above portions, but since that might change in future we'll
				// be robust here.
				what = "timestamp segment"
			}
			if err.ValueElem == "" {
				return time.Time{}, fmt.Errorf("not a valid RFC3339 timestamp: end of string before %s", what)
			} else {
				return time.Time{}, fmt.Errorf("not a valid RFC3339 timestamp: cannot use %q as %s", err.ValueElem, what)
			}
		}
		return time.Time{}, err
	}
	return t, nil
}

// LegacyIsotimeFunc constructs a function that returns a string representation
// of the current date and time using golang's datetime formatting.
var LegacyIsotimeFunc = function.New(&function.Spec{
	Params: []function.Parameter{},
	VarParam: &function.Parameter{
		Name: "format",
		Type: cty.String,
	},
	Type: function.StaticReturnType(cty.String),
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		if len(args) > 1 {
			return cty.StringVal(""), fmt.Errorf("too many values, 1 needed: %v", args)
		} else if len(args) == 0 {
			return cty.StringVal(InitTime.Format(time.RFC3339)), nil
		}
		format := args[0].AsString()
		return cty.StringVal(InitTime.Format(format)), nil
	},
})

// LegacyStrftimeFunc constructs a function that returns a string representation
// of the current date and time using golang's strftime datetime formatting.
var LegacyStrftimeFunc = function.New(&function.Spec{
	Params: []function.Parameter{},
	VarParam: &function.Parameter{
		Name: "format",
		Type: cty.String,
	},
	Type: function.StaticReturnType(cty.String),
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		if len(args) > 1 {
			return cty.StringVal(""), fmt.Errorf("too many values, 1 needed: %v", args)
		} else if len(args) == 0 {
			return cty.StringVal(InitTime.Format(time.RFC3339)), nil
		}
		format := args[0].AsString()
		return cty.StringVal(strftime.Format(format, InitTime)), nil
	},
})

// TimestampFunc constructs a function that returns a string representation of the current date and time.
var TimestampFunc = function.New(&function.Spec{
	Params: []function.Parameter{},
	Type:   function.StaticReturnType(cty.String),
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		return cty.StringVal(time.Now().UTC().Format(time.RFC3339)), nil
	},
})

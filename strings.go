package hclfuncs

import (
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"strings"
)

// StartsWithFunc constructs a function that checks if a string starts with
// a specific prefix using strings.HasPrefix
// Copy from https://github.com/opentofu/opentofu/blob/v1.7.1/internal/lang/funcs/string.go
var StartsWithFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name:         "str",
			Type:         cty.String,
			AllowUnknown: true,
		},
		{
			Name: "prefix",
			Type: cty.String,
		},
	},
	Type:         function.StaticReturnType(cty.Bool),
	RefineResult: refineNotNull,
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		prefix := args[1].AsString()

		if !args[0].IsKnown() {
			// If the unknown value has a known prefix then we might be
			// able to still produce a known result.
			if prefix == "" {
				// The empty string is a prefix of any string.
				return cty.True, nil
			}
			if knownPrefix := args[0].Range().StringPrefix(); knownPrefix != "" {
				if strings.HasPrefix(knownPrefix, prefix) {
					return cty.True, nil
				}
				if len(knownPrefix) >= len(prefix) {
					// If the prefix we're testing is no longer than the known
					// prefix and it didn't match then the full string with
					// that same prefix can't match either.
					return cty.False, nil
				}
			}
			return cty.UnknownVal(cty.Bool), nil
		}

		str := args[0].AsString()

		if strings.HasPrefix(str, prefix) {
			return cty.True, nil
		}

		return cty.False, nil
	},
})

// EndsWithFunc constructs a function that checks if a string ends with
// a specific suffix using strings.HasSuffix
// Copy from https://github.com/opentofu/opentofu/blob/v1.7.1/internal/lang/funcs/string.go
var EndsWithFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name: "str",
			Type: cty.String,
		},
		{
			Name: "suffix",
			Type: cty.String,
		},
	},
	Type:         function.StaticReturnType(cty.Bool),
	RefineResult: refineNotNull,
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		str := args[0].AsString()
		suffix := args[1].AsString()

		if strings.HasSuffix(str, suffix) {
			return cty.True, nil
		}

		return cty.False, nil
	},
})

// StrContainsFunc searches a given string for another given substring,
// if found the function returns true, otherwise returns false.
// Copy from https://github.com/opentofu/opentofu/blob/v1.7.1/internal/lang/funcs/string.go
var StrContainsFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name: "str",
			Type: cty.String,
		},
		{
			Name: "substr",
			Type: cty.String,
		},
	},
	Type: function.StaticReturnType(cty.Bool),
	Impl: func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {
		str := args[0].AsString()
		substr := args[1].AsString()

		if strings.Contains(str, substr) {
			return cty.True, nil
		}

		return cty.False, nil
	},
})

// Copy from https://github.com/opentofu/opentofu/blob/v1.7.1/internal/lang/funcs/refinements.go
func refineNotNull(b *cty.RefinementBuilder) *cty.RefinementBuilder {
	return b.NotNull()
}

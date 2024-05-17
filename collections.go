package hclfuncs

import (
	"codeberg.org/6543/go-yaml2json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/function/stdlib"
	"math/big"
	"sort"
)

// IndexFunc constructs a function that finds the element index for a given value in a list.
var IndexFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name: "list",
			Type: cty.DynamicPseudoType,
		},
		{
			Name: "value",
			Type: cty.DynamicPseudoType,
		},
	},
	Type: function.StaticReturnType(cty.Number),
	Impl: func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {
		if !(args[0].Type().IsListType() || args[0].Type().IsTupleType()) {
			return cty.NilVal, errors.New("argument must be a list or tuple")
		}

		if !args[0].IsKnown() {
			return cty.UnknownVal(cty.Number), nil
		}

		if args[0].LengthInt() == 0 { // Easy path
			return cty.NilVal, errors.New("cannot search an empty list")
		}

		for it := args[0].ElementIterator(); it.Next(); {
			i, v := it.Element()
			eq, err := stdlib.Equal(v, args[1])
			if err != nil {
				return cty.NilVal, err
			}
			if !eq.IsKnown() {
				return cty.UnknownVal(cty.Number), nil
			}
			if eq.True() {
				return i, nil
			}
		}
		return cty.NilVal, errors.New("item not found")

	},
})

var ComplimentFunction = function.New(&function.Spec{
	Description: "Return the compliment of list1 and all otherLists.",
	Params: []function.Parameter{
		{
			Name:             "list1",
			Description:      "the first list, will return all elements that in this list but not in any of other lists.",
			Type:             cty.Set(cty.DynamicPseudoType),
			AllowDynamicType: true,
		},
	},
	VarParam: &function.Parameter{
		Name:             "otherList",
		Description:      "other_list",
		Type:             cty.Set(cty.DynamicPseudoType),
		AllowDynamicType: true,
	},
	Type:         setOperationReturnType,
	RefineResult: refineNonNull,
	Impl: setOperationImpl(func(s1, s2 cty.ValueSet) cty.ValueSet {
		return s1.Subtract(s2)
	}, false),
})

// MatchkeysFunc constructs a function that constructs a new list by taking a
// subset of elements from one list whose indexes match the corresponding
// indexes of values in another list.
// Copy from https://github.com/opentofu/opentofu/blob/v1.7.1/internal/lang/funcs/collection.go
var MatchkeysFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name: "values",
			Type: cty.List(cty.DynamicPseudoType),
		},
		{
			Name: "keys",
			Type: cty.List(cty.DynamicPseudoType),
		},
		{
			Name: "searchset",
			Type: cty.List(cty.DynamicPseudoType),
		},
	},
	Type: func(args []cty.Value) (cty.Type, error) {
		ty, _ := convert.UnifyUnsafe([]cty.Type{args[1].Type(), args[2].Type()})
		if ty == cty.NilType {
			return cty.NilType, errors.New("keys and searchset must be of the same type")
		}

		// the return type is based on args[0] (values)
		return args[0].Type(), nil
	},
	RefineResult: refineNotNull,
	Impl: func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {
		if !args[0].IsKnown() {
			return cty.UnknownVal(cty.List(retType.ElementType())), nil
		}

		if args[0].LengthInt() != args[1].LengthInt() {
			return cty.ListValEmpty(retType.ElementType()), errors.New("length of keys and values should be equal")
		}

		output := make([]cty.Value, 0)
		values := args[0]

		// Keys and searchset must be the same type.
		// We can skip error checking here because we've already verified that
		// they can be unified in the Type function
		ty, _ := convert.UnifyUnsafe([]cty.Type{args[1].Type(), args[2].Type()})
		keys, _ := convert.Convert(args[1], ty)
		searchset, _ := convert.Convert(args[2], ty)

		// if searchset is empty, return an empty list.
		if searchset.LengthInt() == 0 {
			return cty.ListValEmpty(retType.ElementType()), nil
		}

		if !values.IsWhollyKnown() || !keys.IsWhollyKnown() {
			return cty.UnknownVal(retType), nil
		}

		i := 0
		for it := keys.ElementIterator(); it.Next(); {
			_, key := it.Element()
			for iter := searchset.ElementIterator(); iter.Next(); {
				_, search := iter.Element()
				eq, err := stdlib.Equal(key, search)
				if err != nil {
					return cty.NilVal, err
				}
				if !eq.IsKnown() {
					return cty.ListValEmpty(retType.ElementType()), nil
				}
				if eq.True() {
					v := values.Index(cty.NumberIntVal(int64(i)))
					output = append(output, v)
					break
				}
			}
			i++
		}

		// if we haven't matched any key, then output is an empty list.
		if len(output) == 0 {
			return cty.ListValEmpty(retType.ElementType()), nil
		}
		return cty.ListVal(output), nil
	},
})

// SumFunc constructs a function that returns the sum of all
// numbers provided in a list
// Copy from https://github.com/opentofu/opentofu/blob/v1.7.1/internal/lang/funcs/collection.go
var SumFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name: "list",
			Type: cty.DynamicPseudoType,
		},
	},
	Type:         function.StaticReturnType(cty.Number),
	RefineResult: refineNotNull,
	Impl: func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {

		if !args[0].CanIterateElements() {
			return cty.NilVal, function.NewArgErrorf(0, "cannot sum noniterable")
		}

		if args[0].LengthInt() == 0 { // Easy path
			return cty.NilVal, function.NewArgErrorf(0, "cannot sum an empty list")
		}

		arg := args[0].AsValueSlice()
		ty := args[0].Type()

		if !ty.IsListType() && !ty.IsSetType() && !ty.IsTupleType() {
			return cty.NilVal, function.NewArgErrorf(0, fmt.Sprintf("argument must be list, set, or tuple. Received %s", ty.FriendlyName()))
		}

		if !args[0].IsWhollyKnown() {
			return cty.UnknownVal(cty.Number), nil
		}

		// big.Float.Add can panic if the input values are opposing infinities,
		// so we must catch that here in order to remain within
		// the cty Function abstraction.
		defer func() {
			if r := recover(); r != nil {
				if _, ok := r.(big.ErrNaN); ok {
					ret = cty.NilVal
					err = fmt.Errorf("can't compute sum of opposing infinities")
				} else {
					// not a panic we recognize
					panic(r)
				}
			}
		}()

		s := arg[0]
		if s.IsNull() {
			return cty.NilVal, function.NewArgErrorf(0, "argument must be list, set, or tuple of number values")
		}
		s, err = convert.Convert(s, cty.Number)
		if err != nil {
			return cty.NilVal, function.NewArgErrorf(0, "argument must be list, set, or tuple of number values")
		}
		for _, v := range arg[1:] {
			if v.IsNull() {
				return cty.NilVal, function.NewArgErrorf(0, "argument must be list, set, or tuple of number values")
			}
			v, err = convert.Convert(v, cty.Number)
			if err != nil {
				return cty.NilVal, function.NewArgErrorf(0, "argument must be list, set, or tuple of number values")
			}
			s = s.Add(v)
		}

		return s, nil
	},
})

// TransposeFunc constructs a function that takes a map of lists of strings and
// swaps the keys and values to produce a new map of lists of strings.
// Copy from https://github.com/opentofu/opentofu/blob/v1.7.1/internal/lang/funcs/collection.go
var TransposeFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name: "values",
			Type: cty.Map(cty.List(cty.String)),
		},
	},
	Type:         function.StaticReturnType(cty.Map(cty.List(cty.String))),
	RefineResult: refineNotNull,
	Impl: func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {
		inputMap := args[0]
		if !inputMap.IsWhollyKnown() {
			return cty.UnknownVal(retType), nil
		}

		outputMap := make(map[string]cty.Value)
		tmpMap := make(map[string][]string)

		for it := inputMap.ElementIterator(); it.Next(); {
			inKey, inVal := it.Element()
			for iter := inVal.ElementIterator(); iter.Next(); {
				_, val := iter.Element()
				if !val.Type().Equals(cty.String) {
					return cty.MapValEmpty(cty.List(cty.String)), errors.New("input must be a map of lists of strings")
				}

				outKey := val.AsString()
				if _, ok := tmpMap[outKey]; !ok {
					tmpMap[outKey] = make([]string, 0)
				}
				outVal := tmpMap[outKey]
				outVal = append(outVal, inKey.AsString())
				sort.Strings(outVal)
				tmpMap[outKey] = outVal
			}
		}

		for outKey, outVal := range tmpMap {
			values := make([]cty.Value, 0)
			for _, v := range outVal {
				values = append(values, cty.StringVal(v))
			}
			outputMap[outKey] = cty.ListVal(values)
		}

		if len(outputMap) == 0 {
			return cty.MapValEmpty(cty.List(cty.String)), nil
		}

		return cty.MapVal(outputMap), nil
	},
})

var LengthFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name:             "value",
			Type:             cty.DynamicPseudoType,
			AllowDynamicType: true,
			AllowUnknown:     true,
		},
	},
	Type: func(args []cty.Value) (cty.Type, error) {
		collTy := args[0].Type()
		switch {
		case collTy == cty.String || collTy.IsTupleType() || collTy.IsObjectType() || collTy.IsListType() || collTy.IsMapType() || collTy.IsSetType() || collTy == cty.DynamicPseudoType:
			return cty.Number, nil
		default:
			return cty.Number, errors.New("argument must be a string, a collection type, or a structural type")
		}
	},
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		coll := args[0]
		collTy := args[0].Type()
		switch {
		case collTy == cty.DynamicPseudoType:
			return cty.UnknownVal(cty.Number), nil
		case collTy.IsTupleType():
			l := len(collTy.TupleElementTypes())
			return cty.NumberIntVal(int64(l)), nil
		case collTy.IsObjectType():
			l := len(collTy.AttributeTypes())
			return cty.NumberIntVal(int64(l)), nil
		case collTy == cty.String:
			// We'll delegate to the cty stdlib strlen function here, because
			// it deals with all of the complexities of tokenizing unicode
			// grapheme clusters.
			return stdlib.Strlen(coll)
		case collTy.IsListType() || collTy.IsSetType() || collTy.IsMapType():
			return coll.Length(), nil
		default:
			// Should never happen, because of the checks in our Type func above
			return cty.UnknownVal(cty.Number), errors.New("impossible value type for length(...)")
		}
	},
})

var YAML2JsonFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name: "src",
			Type: cty.String,
		},
	},
	Type: func(args []cty.Value) (cty.Type, error) {
		if !args[0].IsKnown() {
			return cty.DynamicPseudoType, nil
		}
		if args[0].IsNull() {
			return cty.NilType, function.NewArgErrorf(0, "YAML source code cannot be null")
		}
		return cty.String, nil
	},
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		json, err := yaml2json.Convert([]byte(args[0].AsString()))
		if err != nil {
			return cty.NilVal, err
		}
		return cty.StringVal(string(json)), nil
	},
})

// AllTrueFunc constructs a function that returns true if all elements of the
// list are true. If the list is empty, return true.
var AllTrueFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name: "list",
			Type: cty.List(cty.Bool),
		},
	},
	Type: function.StaticReturnType(cty.Bool),
	Impl: func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {
		result := cty.True
		for it := args[0].ElementIterator(); it.Next(); {
			_, v := it.Element()
			if !v.IsKnown() {
				return cty.UnknownVal(cty.Bool), nil
			}
			if v.IsNull() {
				return cty.False, nil
			}
			result = result.And(v)
			if result.False() {
				return cty.False, nil
			}
		}
		return result, nil
	},
})

// AnyTrueFunc constructs a function that returns true if any element of the
// list is true. If the list is empty, return false.
var AnyTrueFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name: "list",
			Type: cty.List(cty.Bool),
		},
	},
	Type: function.StaticReturnType(cty.Bool),
	Impl: func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {
		result := cty.False
		var hasUnknown bool
		for it := args[0].ElementIterator(); it.Next(); {
			_, v := it.Element()
			if !v.IsKnown() {
				hasUnknown = true
				continue
			}
			if v.IsNull() {
				continue
			}
			result = result.Or(v)
			if result.True() {
				return cty.True, nil
			}
		}
		if hasUnknown {
			return cty.UnknownVal(cty.Bool), nil
		}
		return result, nil
	},
})

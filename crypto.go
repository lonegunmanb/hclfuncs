package hclfuncs

import (
	"github.com/hashicorp/go-uuid"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

// Copy from https://github.com/opentofu/opentofu/blob/v1.7.1/internal/lang/funcs/crypto.go
var UUIDFunc = function.New(&function.Spec{
	Params:       []function.Parameter{},
	Type:         function.StaticReturnType(cty.String),
	RefineResult: refineNotNull,
	Impl: func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {
		result, err := uuid.GenerateUUID()
		if err != nil {
			return cty.UnknownVal(cty.String), err
		}
		return cty.StringVal(result), nil
	},
})

package hclfuncs

import (
	"fmt"
	"github.com/hashicorp/go-version"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

var SemverCheck = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name: "constraint",
			Type: cty.String,
		},
		{
			Name: "version",
			Type: cty.String,
		},
	},
	Type:         function.StaticReturnType(cty.Bool),
	RefineResult: refineNotNull,
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		constraintStr := args[0].AsString()
		versionStr := args[1].AsString()
		constraint, err := version.NewConstraint(constraintStr)
		if err != nil {
			return cty.NilVal, fmt.Errorf("invalid constraint %s, %v+", constraintStr, err)
		}
		v, err := version.NewVersion(versionStr)
		if err != nil {
			return cty.NilVal, fmt.Errorf("invalid version %s, %v+", versionStr, err)
		}
		return cty.BoolVal(constraint.Check(v)), nil
	},
})

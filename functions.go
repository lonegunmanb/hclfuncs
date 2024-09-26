package hclfuncs

import (
	"fmt"
	"github.com/hashicorp/go-cty-funcs/uuid"
	ctyyaml "github.com/zclconf/go-cty-yaml"
	"os"
	"time"

	"github.com/hashicorp/go-cty-funcs/cidr"
	"github.com/hashicorp/go-cty-funcs/collection"
	"github.com/hashicorp/go-cty-funcs/crypto"
	"github.com/hashicorp/go-cty-funcs/encoding"
	"github.com/hashicorp/go-cty-funcs/filesystem"
	"github.com/hashicorp/hcl/v2/ext/tryfunc"
	"github.com/hashicorp/hcl/v2/ext/typeexpr"
	commontpl "github.com/hashicorp/packer-plugin-sdk/template"
	"github.com/timandy/routine"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/function/stdlib"
)

// InitTime is the UTC time when this package was initialized. It is
// used as the timestamp for all configuration templates so that they
// match for a single build.
var InitTime time.Time

var GoroutineLocalEnv = routine.NewThreadLocal[map[string]string]()

func init() {
	InitTime = time.Now().UTC()
}

func Functions(baseDir string) map[string]function.Function {
	r := map[string]function.Function{
		"alltrue":          AllTrueFunc,
		"anytrue":          AnyTrueFunc,
		"abs":              stdlib.AbsoluteFunc,
		"abspath":          filesystem.AbsPathFunc,
		"basename":         filesystem.BasenameFunc,
		"base64decode":     encoding.Base64DecodeFunc,
		"base64encode":     encoding.Base64EncodeFunc,
		"bcrypt":           crypto.BcryptFunc,
		"can":              tryfunc.CanFunc,
		"ceil":             stdlib.CeilFunc,
		"chomp":            stdlib.ChompFunc,
		"chunklist":        stdlib.ChunklistFunc,
		"cidrcontains":     CidrContainsFunc,
		"cidrhost":         cidr.HostFunc,
		"cidrnetmask":      cidr.NetmaskFunc,
		"cidrsubnet":       cidr.SubnetFunc,
		"cidrsubnets":      cidr.SubnetsFunc,
		"coalesce":         collection.CoalesceFunc,
		"coalescelist":     stdlib.CoalesceListFunc,
		"compact":          stdlib.CompactFunc,
		"concat":           stdlib.ConcatFunc,
		"consul_key":       ConsulFunc,
		"contains":         stdlib.ContainsFunc,
		"convert":          typeexpr.ConvertFunc,
		"csvdecode":        stdlib.CSVDecodeFunc,
		"dirname":          filesystem.DirnameFunc,
		"distinct":         stdlib.DistinctFunc,
		"endswith":         EndsWithFunc,
		"element":          stdlib.ElementFunc,
		"file":             filesystem.MakeFileFunc(baseDir, false),
		"fileexists":       filesystem.MakeFileExistsFunc(baseDir),
		"fileset":          filesystem.MakeFileSetFunc(baseDir),
		"flatten":          stdlib.FlattenFunc,
		"floor":            stdlib.FloorFunc,
		"format":           stdlib.FormatFunc,
		"formatdate":       stdlib.FormatDateFunc,
		"formatlist":       stdlib.FormatListFunc,
		"indent":           stdlib.IndentFunc,
		"index":            IndexFunc, // stdlib.IndexFunc is not compatible
		"issensitive":      IsSensitiveFunc,
		"join":             stdlib.JoinFunc,
		"jsondecode":       stdlib.JSONDecodeFunc,
		"jsonencode":       stdlib.JSONEncodeFunc,
		"keys":             stdlib.KeysFunc,
		"legacy_isotime":   LegacyIsotimeFunc,
		"legacy_strftime":  LegacyStrftimeFunc,
		"length":           LengthFunc,
		"log":              stdlib.LogFunc,
		"lookup":           stdlib.LookupFunc,
		"lower":            stdlib.LowerFunc,
		"matchkeys":        MatchkeysFunc,
		"max":              stdlib.MaxFunc,
		"md5":              crypto.Md5Func,
		"merge":            stdlib.MergeFunc,
		"min":              stdlib.MinFunc,
		"nonsensitive":     NonsensitiveFunc,
		"parseint":         stdlib.ParseIntFunc,
		"pathexpand":       filesystem.PathExpandFunc,
		"pow":              stdlib.PowFunc,
		"range":            stdlib.RangeFunc,
		"regex":            stdlib.RegexFunc,
		"regexall":         stdlib.RegexAllFunc,
		"regex_replace":    stdlib.RegexReplaceFunc,
		"replace":          ReplaceFunc,
		"reverse":          stdlib.ReverseListFunc,
		"rsadecrypt":       crypto.RsaDecryptFunc,
		"sensitive":        SensitiveFunc,
		"setintersection":  stdlib.SetIntersectionFunc,
		"setproduct":       stdlib.SetProductFunc,
		"setsubtract":      stdlib.SetSubtractFunc,
		"setunion":         stdlib.SetUnionFunc,
		"sha1":             crypto.Sha1Func,
		"sha256":           crypto.Sha256Func,
		"sha512":           crypto.Sha512Func,
		"signum":           stdlib.SignumFunc,
		"slice":            stdlib.SliceFunc,
		"sort":             stdlib.SortFunc,
		"split":            stdlib.SplitFunc,
		"startswith":       StartsWithFunc,
		"strcontains":      StrContainsFunc,
		"strrev":           stdlib.ReverseFunc,
		"substr":           stdlib.SubstrFunc,
		"sum":              SumFunc,
		"textdecodebase64": TextDecodeBase64Func,
		"textencodebase64": TextEncodeBase64Func,
		"timestamp":        TimestampFunc,
		"timeadd":          stdlib.TimeAddFunc,
		"timecmp":          TimeCmpFunc,
		"title":            stdlib.TitleFunc,
		"transpose":        TransposeFunc,
		"trim":             stdlib.TrimFunc,
		"trimprefix":       stdlib.TrimPrefixFunc,
		"trimspace":        stdlib.TrimSpaceFunc,
		"trimsuffix":       stdlib.TrimSuffixFunc,
		"try":              tryfunc.TryFunc,
		"upper":            stdlib.UpperFunc,
		"urlencode":        URLEncodeFunc,
		"urldecode":        URLDecodeFunc,
		"uuid":             UUIDFunc,
		"uuidv4":           uuid.V4Func,
		"uuidv5":           uuid.V5Func,
		"values":           stdlib.ValuesFunc,
		"vault":            VaultFunc,
		"yamldecode":       ctyyaml.YAMLDecodeFunc,
		"yamlencode":       ctyyaml.YAMLEncodeFunc,
		"yaml2json":        YAML2JsonFunc,
		"zipmap":           stdlib.ZipmapFunc,
		"compliment":       ComplimentFunction,
		"env":              EnvFunction,
		"tostring":         MakeToFunc(cty.String),
		"tonumber":         MakeToFunc(cty.Number),
		"tobool":           MakeToFunc(cty.Bool),
		"toset":            MakeToFunc(cty.Set(cty.DynamicPseudoType)),
		"tolist":           MakeToFunc(cty.List(cty.DynamicPseudoType)),
		"tomap":            MakeToFunc(cty.Map(cty.DynamicPseudoType)),
	}
	return r
}

var EnvFunction = function.New(&function.Spec{
	Description: "Read environment variable, return empty string if the variable is not set.",
	Params: []function.Parameter{
		{
			Name:         "key",
			Description:  "Environment variable name",
			Type:         cty.String,
			AllowUnknown: true,
			AllowMarked:  true,
		},
	},
	Type: function.StaticReturnType(cty.String),
	RefineResult: func(builder *cty.RefinementBuilder) *cty.RefinementBuilder {
		return builder.NotNull()
	},
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		key := args[0]
		if !key.IsKnown() {
			return cty.UnknownVal(cty.String), nil
		}
		envKey := key.AsString()
		localEnv := GoroutineLocalEnv.Get()
		if localEnv != nil {
			if env, ok := localEnv[envKey]; ok {
				return cty.StringVal(env), nil
			}
		}
		env := os.Getenv(envKey)
		return cty.StringVal(env), nil
	},
})

func setOperationReturnType(args []cty.Value) (ret cty.Type, err error) {
	var etys []cty.Type
	for _, arg := range args {
		ty := arg.Type().ElementType()

		if arg.IsKnown() && arg.LengthInt() == 0 && ty.Equals(cty.DynamicPseudoType) {
			continue
		}

		etys = append(etys, ty)
	}

	if len(etys) == 0 {
		return cty.Set(cty.DynamicPseudoType), nil
	}

	newEty, _ := convert.UnifyUnsafe(etys)
	if newEty == cty.NilType {
		return cty.NilType, fmt.Errorf("given sets must all have compatible element types")
	}
	return cty.Set(newEty), nil
}

func refineNonNull(b *cty.RefinementBuilder) *cty.RefinementBuilder {
	return b.NotNull()
}

func setOperationImpl(f func(s1, s2 cty.ValueSet) cty.ValueSet, allowUnknowns bool) function.ImplFunc {
	return func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {
		first := args[0]
		first, err = convert.Convert(first, retType)
		if err != nil {
			return cty.NilVal, function.NewArgError(0, err)
		}
		if !allowUnknowns && !first.IsWhollyKnown() {
			// This set function can produce a correct result only when all
			// elements are known, because eventually knowing the unknown
			// values may cause the result to have fewer known elements, or
			// might cause a result with no unknown elements at all to become
			// one with a different length.
			return cty.UnknownVal(retType), nil
		}

		set := first.AsValueSet()
		for i, arg := range args[1:] {
			arg, err := convert.Convert(arg, retType)
			if err != nil {
				return cty.NilVal, function.NewArgError(i+1, err)
			}
			if !allowUnknowns && !arg.IsWhollyKnown() {
				// (For the same reason as we did this check for "first" above.)
				return cty.UnknownVal(retType), nil
			}

			argSet := arg.AsValueSet()
			set = f(set, argSet)
		}
		return cty.SetValFromValueSet(set), nil
	}
}

// ConsulFunc constructs a function that retrieves KV secrets from HC vault
var ConsulFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name: "key",
			Type: cty.String,
		},
	},
	Type: function.StaticReturnType(cty.String),
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		key := args[0].AsString()
		val, err := commontpl.Consul(key)

		return cty.StringVal(val), err
	},
})

// VaultFunc constructs a function that retrieves KV secrets from HC vault
var VaultFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name: "path",
			Type: cty.String,
		},
		{
			Name: "key",
			Type: cty.String,
		},
	},
	Type: function.StaticReturnType(cty.String),
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		path := args[0].AsString()
		key := args[1].AsString()

		val, err := commontpl.Vault(path, key)

		return cty.StringVal(val), err
	},
})

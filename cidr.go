package hclfuncs

import (
	"fmt"
	"github.com/apparentlymart/go-cidr/cidr"
	"github.com/lonegunmanb/hclfuncs/ipaddr"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

// CidrContainsFunc constructs a function that checks whether a given IP address
// is within a given IP network address prefix.
var CidrContainsFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name: "containing_prefix",
			Type: cty.String,
		},
		{
			Name: "contained_ip_or_prefix",
			Type: cty.String,
		},
	},
	Type: function.StaticReturnType(cty.Bool),
	Impl: func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {
		prefix := args[0].AsString()
		addr := args[1].AsString()

		// The first argument must be a CIDR prefix.
		_, containing, err := ipaddr.ParseCIDR(prefix)
		if err != nil {
			return cty.UnknownVal(cty.Bool), err
		}

		// The second argument can be either an IP address or a CIDR prefix.
		// We will try parsing it as an IP address first.
		startIP := ipaddr.ParseIP(addr)
		var endIP ipaddr.IP

		// If the second argument did not parse as an IP, we will try parsing it
		// as a CIDR prefix.
		if startIP == nil {
			_, contained, err := ipaddr.ParseCIDR(addr)

			// If that also fails, we'll return an error.
			if err != nil {
				return cty.UnknownVal(cty.Bool), fmt.Errorf("invalid IP address or prefix: %s", addr)
			}

			// Otherwise, we will want to know the start and the end IP of the
			// prefix, so that we can check whether both are contained in the
			// containing prefix.
			startIP, endIP = cidr.AddressRange(contained)
		}

		// We require that both addresses are of the same type, so that
		// we can't accidentally compare an IPv4 address to an IPv6 prefix.
		// The underlying Go function will always return false if this happens,
		// but we want to return an error instead so that the caller can
		// distinguish between a "legitimate" false result and an erroneous
		// check.
		if (startIP.To4() == nil) != (containing.IP.To4() == nil) {
			return cty.UnknownVal(cty.Bool), fmt.Errorf("address family mismatch: %s vs. %s", prefix, addr)
		}

		// If the second argument was an IP address, we will check whether it
		// is contained in the containing prefix, and that's our result.
		result := containing.Contains(startIP)

		// If the second argument was a CIDR prefix, we will also check whether
		// the end IP of the prefix is contained in the containing prefix.
		// Once CIDR is contained in another CIDR iff both the start and the
		// end IP of the contained CIDR are contained in the containing CIDR.
		if endIP != nil {
			result = result && containing.Contains(endIP)
		}

		return cty.BoolVal(result), nil
	},
})

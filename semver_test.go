package hclfuncs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zclconf/go-cty/cty"
)

func TestSemverCheck(t *testing.T) {
	tests := []struct {
		constraint string
		version    string
		expected   bool
	}{
		{">= 1.0, < 1.2", "1.1.5", true},
		{"< 1.0, < 1.2", "1.1.5", false},
		{"= 1.0", "1.1.5", false},
		{"= 1.0", "1.0.0", true},
		{"1.0", "1.0.0", true},
		{"~> 1.0", "2.0", false},
		{"~> 1.0", "1.1", true},
		{"~> 1.0", "1.2.3", true},
		{"~> 1.0.0", "1.2.3", false},
		{"~> 1.0.0", "1.0.7", true},
		{"~> 1.0.0", "1.1.0", false},
		{"~> 1.0.7", "1.0.4", false},
		{"~> 1.0.7", "1.0.7", true},
		{"~> 1.0.7", "1.0.8", true},
		{"~> 1.0.7", "1.0.7.5", true},
		{"~> 1.0.7", "1.0.6.99", false},
		{"~> 1.0.7", "1.0.8.0", true},
		{"~> 1.0.9.5", "1.0.9.5", true},
		{"~> 1.0.9.5", "1.0.9.4", false},
		{"~> 1.0.9.5", "1.0.9.6", true},
		{"~> 1.0.9.5", "1.0.9.5.0", true},
		{"~> 1.0.9.5", "1.0.9.5.1", true},
		{"~> 2.0", "2.1.0-beta", false},
		{"~> 2.1.0-a", "2.2.0", false},
		{"~> 2.1.0-a", "2.1.0", false},
		{"~> 2.1.0-a", "2.1.0-beta", true},
		{"~> 2.1.0-a", "2.2.0-alpha", false},
		{"> 2.0", "2.1.0-beta", false},
		{">= 2.1.0-a", "2.1.0-beta", true},
		{">= 2.1.0-a", "2.1.1-beta", false},
		{">= 2.0.0", "2.1.0-beta", false},
		{">= 2.1.0-a", "2.1.1", true},
		{">= 2.1.0-a", "2.1.1-beta", false},
		{">= 2.1.0-a", "2.1.1", true},
	}

	for _, tc := range tests {
		t.Run(tc.constraint+"_"+tc.version, func(t *testing.T) {
			result, err := SemverCheck.Call([]cty.Value{
				cty.StringVal(tc.constraint),
				cty.StringVal(tc.version),
			})
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, result.True())
		})
	}
}

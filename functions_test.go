package hclfuncs

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/packer-plugin-sdk/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFunction_Env(t *testing.T) {
	v := uuid.NewString()
	t.Setenv("TEST_ENV", v)
	code := `env("TEST_ENV")`
	exp, diag := hclsyntax.ParseExpression([]byte(code), "test.hcl", hcl.InitialPos)
	require.False(t, diag.HasErrors())
	value, diag := exp.Value(&hcl.EvalContext{
		Functions: Functions("."),
	})
	require.False(t, diag.HasErrors())
	assert.Equal(t, v, value.AsString())
}

func TestFunction_EnvShouldHonorGoroutineLocalEnv(t *testing.T) {
	v0 := uuid.NewString()
	v1 := uuid.NewString()
	t.Setenv("TEST_ENV", v0)
	code := `env("TEST_ENV")`
	exp, diag := hclsyntax.ParseExpression([]byte(code), "test.hcl", hcl.InitialPos)
	require.False(t, diag.HasErrors())
	done := make(chan struct{})
	defer close(done)
	go func() {
		GoroutineLocalEnv.Set(map[string]string{
			"TEST_ENV": v1,
		})
		value, diag := exp.Value(&hcl.EvalContext{
			Functions: Functions("."),
		})
		require.False(t, diag.HasErrors())
		assert.Equal(t, v1, value.AsString())
		done <- struct{}{}
	}()
	value, diag := exp.Value(&hcl.EvalContext{
		Functions: Functions("."),
	})
	require.False(t, diag.HasErrors())
	assert.Equal(t, v0, value.AsString())
	select {
	case <-done:
		{
		}
	case <-time.After(10 * time.Millisecond):
		{
			t.Fatal("timeout")
		}
	}
}

func TestFunction_Compliment(t *testing.T) {
	code := `compliment([2, 4, 6, 8, 10, 12], [4, 6, 8], [12])`
	exp, diag := hclsyntax.ParseExpression([]byte(code), "test.hcl", hcl.InitialPos)
	require.False(t, diag.HasErrors())
	value, diag := exp.Value(&hcl.EvalContext{Functions: Functions(".")})
	require.False(t, diag.HasErrors())
	slice := value.AsValueSlice()
	assert.Len(t, slice, 2)
	i0, _ := slice[0].AsBigFloat().Int64()
	i1, _ := slice[1].AsBigFloat().Int64()
	assert.Equal(t, int64(2), i0)
	assert.Equal(t, int64(10), i1)
}

func TestFunction_Yaml2Json(t *testing.T) {
	code := `yaml2json("name: build\n\non:\n  push:\n    branches:\n        - main\n  pull_request:\n")`
	exp, diag := hclsyntax.ParseExpression([]byte(code), "test.hcl", hcl.InitialPos)
	require.False(t, diag.HasErrors())
	value, diag := exp.Value(&hcl.EvalContext{Functions: Functions(".")})
	require.False(t, diag.HasErrors())
	j := value.AsString()
	m := make(map[string]any)
	err := json.Unmarshal([]byte(j), &m)
	require.NoError(t, err)
	expected := map[string]any{
		"name": "build",
		"on": map[string]any{
			"push": map[string]any{
				"branches": []any{"main"},
			},
			"pull_request": nil,
		},
	}
	assert.Equal(t, expected, m)
}

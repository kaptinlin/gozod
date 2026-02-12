package types

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/kaptinlin/gozod/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	errTransformFailed = errors.New("transform error")
	errOnlyTrueAllowed = errors.New("only true values allowed")
)

func TestTransform_DefaultAndPrefault(t *testing.T) {
	t.Run("Default skips transform", func(t *testing.T) {
		transformCalled := false
		schema := String().Default("default_value").Transform(func(input string, _ *core.RefinementContext) (any, error) {
			transformCalled = true
			return strings.ToUpper(input), nil
		})

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "default_value", result)
		assert.False(t, transformCalled)
	})

	t.Run("Prefault goes through transform", func(t *testing.T) {
		transformCalled := false
		schema := String().Prefault("prefault_value").Transform(func(input string, _ *core.RefinementContext) (any, error) {
			transformCalled = true
			return strings.ToUpper(input), nil
		})

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "PREFAULT_VALUE", result)
		assert.True(t, transformCalled)
	})

	t.Run("Default takes precedence over Prefault", func(t *testing.T) {
		transformCalled := false
		schema := String().Default("default_value").Prefault("prefault_value").Transform(func(input string, _ *core.RefinementContext) (any, error) {
			transformCalled = true
			return strings.ToUpper(input), nil
		})

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "default_value", result)
		assert.False(t, transformCalled)
	})

	t.Run("valid input goes through transform", func(t *testing.T) {
		transformCalled := false
		schema := String().Default("default_value").Prefault("prefault_value").Transform(func(input string, _ *core.RefinementContext) (any, error) {
			transformCalled = true
			return strings.ToUpper(input), nil
		})

		result, err := schema.Parse("input_value")
		require.NoError(t, err)
		assert.Equal(t, "INPUT_VALUE", result)
		assert.True(t, transformCalled)
	})

	t.Run("transform error does not fall back to Prefault", func(t *testing.T) {
		schema := String().Prefault("prefault_value").Transform(func(_ string, _ *core.RefinementContext) (any, error) {
			return nil, errTransformFailed
		})

		_, err := schema.Parse("input_value")
		require.Error(t, err)
		assert.ErrorIs(t, err, errTransformFailed)
	})

	t.Run("DefaultFunc skips transform", func(t *testing.T) {
		defaultCalled := false
		transformCalled := false
		schema := String().DefaultFunc(func() string {
			defaultCalled = true
			return "func_default"
		}).Transform(func(input string, _ *core.RefinementContext) (any, error) {
			transformCalled = true
			return strings.ToUpper(input), nil
		})

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "func_default", result)
		assert.True(t, defaultCalled)
		assert.False(t, transformCalled)
	})

	t.Run("PrefaultFunc goes through transform", func(t *testing.T) {
		prefaultCalled := false
		transformCalled := false
		schema := String().PrefaultFunc(func() string {
			prefaultCalled = true
			return "func_prefault"
		}).Transform(func(input string, _ *core.RefinementContext) (any, error) {
			transformCalled = true
			return strings.ToUpper(input), nil
		})

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "FUNC_PREFAULT", result)
		assert.True(t, prefaultCalled)
		assert.True(t, transformCalled)
	})
}

func TestBool_Transform(t *testing.T) {
	boolToString := Bool().Transform(func(b bool, _ *core.RefinementContext) (any, error) {
		if b {
			return "YES", nil
		}
		return "NO", nil
	})

	tests := []struct {
		name     string
		input    any
		expected any
		wantErr  bool
	}{
		{name: "true to YES", input: true, expected: "YES"},
		{name: "false to NO", input: false, expected: "NO"},
		{name: "invalid input type", input: "not a bool", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := boolToString.Parse(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBool_ChainedTransform(t *testing.T) {
	// bool -> string -> int
	transform := Bool().
		Transform(func(b bool, _ *core.RefinementContext) (any, error) {
			if b {
				return "YES", nil
			}
			return "NO", nil
		}).
		Transform(func(s any, _ *core.RefinementContext) (any, error) {
			if s.(string) == "YES" {
				return 1, nil
			}
			return 0, nil
		})

	tests := []struct {
		name     string
		input    any
		expected any
	}{
		{name: "true -> YES -> 1", input: true, expected: 1},
		{name: "false -> NO -> 0", input: false, expected: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := transform.Parse(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBool_TransformError(t *testing.T) {
	transform := Bool().Transform(func(b bool, _ *core.RefinementContext) (any, error) {
		if !b {
			return nil, errOnlyTrueAllowed
		}
		return "accepted", nil
	})

	t.Run("success", func(t *testing.T) {
		result, err := transform.Parse(true)
		require.NoError(t, err)
		assert.Equal(t, "accepted", result)
	})

	t.Run("error", func(t *testing.T) {
		_, err := transform.Parse(false)
		require.Error(t, err)
		assert.ErrorIs(t, err, errOnlyTrueAllowed)
	})
}

func TestBool_MustParse_Transform(t *testing.T) {
	transform := Bool().Transform(func(b bool, _ *core.RefinementContext) (any, error) {
		return fmt.Sprintf("value: %t", b), nil
	})

	t.Run("success", func(t *testing.T) {
		result := transform.MustParse(true)
		assert.Equal(t, "value: true", result)
	})

	t.Run("panics on invalid input", func(t *testing.T) {
		assert.Panics(t, func() {
			transform.MustParse("invalid")
		})
	})
}

func TestBool_ComplexTransform(t *testing.T) {
	configTransform := Bool().Transform(func(enabled bool, _ *core.RefinementContext) (any, error) {
		config := map[string]any{
			"feature_enabled": enabled,
			"max_connections": 100,
			"timeout":         30,
		}
		if enabled {
			config["max_connections"] = 1000
			config["timeout"] = 60
		}
		return config, nil
	})

	t.Run("enabled", func(t *testing.T) {
		result, err := configTransform.Parse(true)
		require.NoError(t, err)

		config, ok := result.(map[string]any)
		require.True(t, ok)
		assert.Equal(t, true, config["feature_enabled"])
		assert.Equal(t, 1000, config["max_connections"])
		assert.Equal(t, 60, config["timeout"])
	})

	t.Run("disabled", func(t *testing.T) {
		result, err := configTransform.Parse(false)
		require.NoError(t, err)

		config, ok := result.(map[string]any)
		require.True(t, ok)
		assert.Equal(t, false, config["feature_enabled"])
		assert.Equal(t, 100, config["max_connections"])
		assert.Equal(t, 30, config["timeout"])
	})
}

func TestZodTransform_InterfaceMethods(t *testing.T) {
	toUpper := String().Transform(func(s string, _ *core.RefinementContext) (any, error) {
		return strings.ToUpper(s), nil
	})

	t.Run("IsOptional returns false by default", func(t *testing.T) {
		assert.False(t, toUpper.IsOptional())
	})

	t.Run("IsOptional returns true for optional source", func(t *testing.T) {
		schema := String().Optional()
		assert.True(t, schema.IsOptional())
	})

	t.Run("IsNilable returns false by default", func(t *testing.T) {
		assert.False(t, toUpper.IsNilable())
	})

	t.Run("IsNilable returns true for nilable source", func(t *testing.T) {
		schema := String().Nilable()
		assert.True(t, schema.IsNilable())
	})

	t.Run("MustParse succeeds", func(t *testing.T) {
		assert.Equal(t, "HELLO", toUpper.MustParse("hello"))
	})

	t.Run("MustParse panics on invalid input", func(t *testing.T) {
		assert.Panics(t, func() { toUpper.MustParse(123) })
	})

	t.Run("MustParse panics on transform error", func(t *testing.T) {
		schema := String().Transform(func(s string, _ *core.RefinementContext) (any, error) {
			if s == "" {
				return nil, fmt.Errorf("empty string: %w", errTransformFailed)
			}
			return s, nil
		})
		assert.Panics(t, func() { schema.MustParse("") })
	})

	t.Run("ParseAny returns any type", func(t *testing.T) {
		toLen := String().Transform(func(s string, _ *core.RefinementContext) (any, error) {
			return len(s), nil
		})
		result, err := toLen.ParseAny("hello")
		require.NoError(t, err)
		assert.Equal(t, 5, result)
	})

	t.Run("ParseAny with empty string", func(t *testing.T) {
		schema := String().Transform(func(s string, _ *core.RefinementContext) (any, error) {
			if s == "" {
				return "was empty", nil
			}
			return s, nil
		})
		result, err := schema.ParseAny("")
		require.NoError(t, err)
		assert.Equal(t, "was empty", result)
	})

	t.Run("ParseAny returns error for invalid input", func(t *testing.T) {
		_, err := toUpper.ParseAny(123)
		assert.Error(t, err)
	})

	t.Run("Parse with context", func(t *testing.T) {
		schema := String().Transform(func(s string, _ *core.RefinementContext) (any, error) {
			return "transformed:" + s, nil
		})
		result, err := schema.Parse("test", core.NewParseContext())
		require.NoError(t, err)
		assert.Equal(t, "transformed:test", result)
	})

	t.Run("MustParse with context", func(t *testing.T) {
		schema := String().Transform(func(s string, _ *core.RefinementContext) (any, error) {
			return strings.ToLower(s), nil
		})
		result := schema.MustParse("HELLO", core.NewParseContext())
		assert.Equal(t, "hello", result)
	})

	t.Run("GetInternals returns valid internals", func(t *testing.T) {
		assert.NotNil(t, toUpper.GetInternals())
	})

	t.Run("internals Type is ZodTypeTransform", func(t *testing.T) {
		assert.Equal(t, core.ZodTypeTransform, toUpper.GetInternals().Type)
	})

	t.Run("GetInner returns the source schema", func(t *testing.T) {
		inner := String().Min(5)
		schema := inner.Transform(func(s string, _ *core.RefinementContext) (any, error) {
			return s, nil
		})
		assert.NotNil(t, schema.GetInner())
	})
}

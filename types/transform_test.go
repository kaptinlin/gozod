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

// Test error variables
var (
	errTransformError  = errors.New("transform error")
	errOnlyTrueAllowed = errors.New("only true values allowed")
)

// =============================================================================
// Default and prefault tests with Transform interaction
// =============================================================================

func TestTransform_DefaultAndPrefault(t *testing.T) {
	t.Run("Default skips transform", func(t *testing.T) {
		// Default should bypass transform and return immediately
		transformCalled := false
		schema := String().Default("default_value").Transform(func(input string, ctx *core.RefinementContext) (any, error) {
			transformCalled = true
			return strings.ToUpper(input), nil
		})

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "default_value", result)
		assert.False(t, transformCalled, "Transform should not be called when Default is used")
	})

	t.Run("Prefault goes through transform", func(t *testing.T) {
		// Prefault should go through the full validation and transform pipeline
		transformCalled := false
		schema := String().Prefault("prefault_value").Transform(func(input string, ctx *core.RefinementContext) (any, error) {
			transformCalled = true
			return strings.ToUpper(input), nil
		})

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "PREFAULT_VALUE", result)
		assert.True(t, transformCalled, "Transform should be called when Prefault is used")
	})

	t.Run("Default has higher priority than Prefault with transform", func(t *testing.T) {
		// When both Default and Prefault are set, Default should take precedence and skip transform
		transformCalled := false
		schema := String().Default("default_value").Prefault("prefault_value").Transform(func(input string, ctx *core.RefinementContext) (any, error) {
			transformCalled = true
			return strings.ToUpper(input), nil
		})

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "default_value", result)
		assert.False(t, transformCalled, "Transform should not be called when Default is used")
	})

	t.Run("Valid input goes through transform", func(t *testing.T) {
		// Valid input should go through transform, ignoring Default and Prefault
		transformCalled := false
		schema := String().Default("default_value").Prefault("prefault_value").Transform(func(input string, ctx *core.RefinementContext) (any, error) {
			transformCalled = true
			return strings.ToUpper(input), nil
		})

		result, err := schema.Parse("input_value")
		require.NoError(t, err)
		assert.Equal(t, "INPUT_VALUE", result)
		assert.True(t, transformCalled, "Transform should be called for valid input")
	})

	t.Run("Transform error with Prefault fallback", func(t *testing.T) {
		// When transform fails, it should not fall back to Prefault
		schema := String().Prefault("prefault_value").Transform(func(input string, ctx *core.RefinementContext) (any, error) {
			return nil, fmt.Errorf("%w", errTransformError)
		})

		// Transform error should be returned, not fall back to Prefault
		_, err := schema.Parse("input_value")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "transform error")
	})

	t.Run("DefaultFunc skips transform", func(t *testing.T) {
		// DefaultFunc should also bypass transform
		defaultCalled := false
		transformCalled := false
		schema := String().DefaultFunc(func() string {
			defaultCalled = true
			return "func_default"
		}).Transform(func(input string, ctx *core.RefinementContext) (any, error) {
			transformCalled = true
			return strings.ToUpper(input), nil
		})

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "func_default", result)
		assert.True(t, defaultCalled, "DefaultFunc should be called")
		assert.False(t, transformCalled, "Transform should not be called when DefaultFunc is used")
	})

	t.Run("PrefaultFunc goes through transform", func(t *testing.T) {
		// PrefaultFunc should go through transform
		prefaultCalled := false
		transformCalled := false
		schema := String().PrefaultFunc(func() string {
			prefaultCalled = true
			return "func_prefault"
		}).Transform(func(input string, ctx *core.RefinementContext) (any, error) {
			transformCalled = true
			return strings.ToUpper(input), nil
		})

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "FUNC_PREFAULT", result)
		assert.True(t, prefaultCalled, "PrefaultFunc should be called")
		assert.True(t, transformCalled, "Transform should be called when PrefaultFunc is used")
	})
}

func TestBool_Transform(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected any
		wantErr  bool
	}{
		{
			name:     "true to YES",
			input:    true,
			expected: "YES",
			wantErr:  false,
		},
		{
			name:     "false to NO",
			input:    false,
			expected: "NO",
			wantErr:  false,
		},
		{
			name:     "invalid input type",
			input:    "not a bool",
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			boolSchema := Bool()

			// Create transform: bool -> string
			transform := boolSchema.Transform(func(b bool, ctx *core.RefinementContext) (any, error) {
				if b {
					return "YES", nil
				}
				return "NO", nil
			})

			result, err := transform.Parse(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestBool_ChainedTransform(t *testing.T) {
	boolSchema := Bool()

	// Create chained transform: bool -> string -> int
	transform := boolSchema.
		Transform(func(b bool, ctx *core.RefinementContext) (any, error) {
			if b {
				return "YES", nil
			}
			return "NO", nil
		}).
		Transform(func(s any, ctx *core.RefinementContext) (any, error) {
			str := s.(string)
			if str == "YES" {
				return 1, nil
			}
			return 0, nil
		})

	tests := []struct {
		name     string
		input    any
		expected any
	}{
		{
			name:     "true -> YES -> 1",
			input:    true,
			expected: 1,
		},
		{
			name:     "false -> NO -> 0",
			input:    false,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := transform.Parse(tt.input)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestBool_TransformError(t *testing.T) {
	boolSchema := Bool()

	// Create transform that only accepts true
	transform := boolSchema.Transform(func(b bool, ctx *core.RefinementContext) (any, error) {
		if !b {
			return nil, fmt.Errorf("%w", errOnlyTrueAllowed)
		}
		return "accepted", nil
	})

	// Test successful case
	result, err := transform.Parse(true)
	if err != nil {
		t.Errorf("unexpected error for true input: %v", err)
	}
	if result != "accepted" {
		t.Errorf("expected 'accepted', got %v", result)
	}

	// Test error case
	_, err = transform.Parse(false)
	if err == nil {
		t.Errorf("expected error for false input")
	}
	if !strings.Contains(err.Error(), "only true values allowed") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestBool_MustParse_Transform(t *testing.T) {
	boolSchema := Bool()

	transform := boolSchema.Transform(func(b bool, ctx *core.RefinementContext) (any, error) {
		return fmt.Sprintf("value: %t", b), nil
	})

	// Test successful MustParse
	result := transform.MustParse(true)
	expected := "value: true"
	if result != expected {
		t.Errorf("expected %v, got %v", expected, result)
	}

	// Test panic on error
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic for invalid input")
		}
	}()

	transform.MustParse("invalid")
}

func TestBool_ComplexTransform(t *testing.T) {
	boolSchema := Bool()

	// Transform to a configuration object
	configTransform := boolSchema.Transform(func(enabled bool, ctx *core.RefinementContext) (any, error) {
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

	// Test enabled configuration
	result, err := configTransform.Parse(true)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	config, ok := result.(map[string]any)
	if !ok {
		t.Errorf("expected map[string]any, got %T", result)
		return
	}

	if config["feature_enabled"] != true {
		t.Errorf("expected feature_enabled to be true")
	}
	if config["max_connections"] != 1000 {
		t.Errorf("expected max_connections to be 1000, got %v", config["max_connections"])
	}
	if config["timeout"] != 60 {
		t.Errorf("expected timeout to be 60, got %v", config["timeout"])
	}

	// Test disabled configuration
	result, err = configTransform.Parse(false)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	config, ok = result.(map[string]any)
	if !ok {
		t.Errorf("expected map[string]any, got %T", result)
		return
	}

	if config["feature_enabled"] != false {
		t.Errorf("expected feature_enabled to be false")
	}
	if config["max_connections"] != 100 {
		t.Errorf("expected max_connections to be 100, got %v", config["max_connections"])
	}
	if config["timeout"] != 30 {
		t.Errorf("expected timeout to be 30, got %v", config["timeout"])
	}
}

// =============================================================================
// ZodTransform interface methods tests
// =============================================================================

func TestZodTransform_InterfaceMethods(t *testing.T) {
	t.Run("IsOptional returns false for non-optional transform", func(t *testing.T) {
		transform := String().Transform(func(s string, ctx *core.RefinementContext) (any, error) {
			return strings.ToUpper(s), nil
		})

		assert.False(t, transform.IsOptional(), "Non-optional transform should return false")
	})

	t.Run("IsOptional returns true for optional schema", func(t *testing.T) {
		// For optional, we check the inner schema's optional status
		optionalSchema := String().Optional()
		assert.True(t, optionalSchema.IsOptional(), "Optional schema should return true")
	})

	t.Run("IsNilable returns false for non-nilable transform", func(t *testing.T) {
		transform := String().Transform(func(s string, ctx *core.RefinementContext) (any, error) {
			return len(s), nil
		})

		assert.False(t, transform.IsNilable(), "Non-nilable transform should return false")
	})

	t.Run("IsNilable returns true for nilable schema", func(t *testing.T) {
		// For nilable, we check the inner schema's nilable status
		nilableSchema := String().Nilable()
		assert.True(t, nilableSchema.IsNilable(), "Nilable schema should return true")
	})

	t.Run("MustParse succeeds for valid input", func(t *testing.T) {
		transform := String().Transform(func(s string, ctx *core.RefinementContext) (any, error) {
			return strings.ToUpper(s), nil
		})

		result := transform.MustParse("hello")
		assert.Equal(t, "HELLO", result)
	})

	t.Run("MustParse panics for invalid input", func(t *testing.T) {
		transform := String().Transform(func(s string, ctx *core.RefinementContext) (any, error) {
			return strings.ToUpper(s), nil
		})

		assert.Panics(t, func() {
			transform.MustParse(123) // Invalid input type
		})
	})

	t.Run("MustParse panics when transform returns error", func(t *testing.T) {
		transform := String().Transform(func(s string, ctx *core.RefinementContext) (any, error) {
			if s == "" {
				return nil, fmt.Errorf("%w: empty string", errTransformError)
			}
			return s, nil
		})

		assert.Panics(t, func() {
			transform.MustParse("") // Will trigger transform error
		})
	})

	t.Run("ParseAny returns any type result", func(t *testing.T) {
		transform := String().Transform(func(s string, ctx *core.RefinementContext) (any, error) {
			return len(s), nil
		})

		result, err := transform.ParseAny("hello")
		require.NoError(t, err)
		assert.Equal(t, 5, result)
		assert.IsType(t, 0, result) // Result should be int
	})

	t.Run("ParseAny with empty string", func(t *testing.T) {
		transform := String().Transform(func(s string, ctx *core.RefinementContext) (any, error) {
			if s == "" {
				return "was empty", nil
			}
			return s, nil
		})

		result, err := transform.ParseAny("")
		require.NoError(t, err)
		assert.Equal(t, "was empty", result)
	})

	t.Run("ParseAny returns error for invalid input", func(t *testing.T) {
		transform := String().Transform(func(s string, ctx *core.RefinementContext) (any, error) {
			return s, nil
		})

		_, err := transform.ParseAny(123) // Invalid type
		assert.Error(t, err)
	})

	t.Run("Parse with context parameter", func(t *testing.T) {
		transform := String().Transform(func(s string, ctx *core.RefinementContext) (any, error) {
			return "transformed:" + s, nil
		})

		ctx := core.NewParseContext()
		result, err := transform.Parse("test", ctx)
		require.NoError(t, err)
		assert.Equal(t, "transformed:test", result)
	})

	t.Run("MustParse with context parameter", func(t *testing.T) {
		transform := String().Transform(func(s string, ctx *core.RefinementContext) (any, error) {
			return strings.ToLower(s), nil
		})

		ctx := core.NewParseContext()
		result := transform.MustParse("HELLO", ctx)
		assert.Equal(t, "hello", result)
	})

	t.Run("GetInternals returns valid internals", func(t *testing.T) {
		transform := String().Transform(func(s string, ctx *core.RefinementContext) (any, error) {
			return s, nil
		})

		internals := transform.GetInternals()
		assert.NotNil(t, internals)
	})

	t.Run("internals Type is ZodTypeTransform", func(t *testing.T) {
		transform := String().Transform(func(s string, ctx *core.RefinementContext) (any, error) {
			return s, nil
		})

		internals := transform.GetInternals()
		assert.Equal(t, core.ZodTypeTransform, internals.Type)
	})

	t.Run("GetInner returns the inner schema", func(t *testing.T) {
		innerSchema := String().Min(5)
		transform := innerSchema.Transform(func(s string, ctx *core.RefinementContext) (any, error) {
			return s, nil
		})

		inner := transform.GetInner()
		assert.NotNil(t, inner)
	})
}

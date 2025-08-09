package types

import (
	"fmt"
	"strings"
	"testing"

	"github.com/kaptinlin/gozod/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			return nil, fmt.Errorf("transform error")
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
			return nil, fmt.Errorf("only true values allowed")
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

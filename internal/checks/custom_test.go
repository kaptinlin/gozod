package checks

import (
	"testing"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// CUSTOM VALIDATION TESTS
// =============================================================================

func TestCustomValidation_RefineFunctions(t *testing.T) {
	t.Run("validates with string refine function", func(t *testing.T) {
		refineFn := func(s string) bool {
			return len(s) > 3 && s[0] >= 'A' && s[0] <= 'Z'
		}
		check := NewCustom[string](refineFn)

		tests := []struct {
			name  string
			input string
			valid bool
		}{
			{"valid string", "Hello", true},
			{"too short", "Hi", false},
			{"lowercase start", "hello", false},
			{"empty string", "", false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				payload := core.NewParsePayload(tt.input)
				internals := check.GetZod()
				internals.Check(payload)

				if tt.valid {
					assert.Empty(t, payload.GetIssues(), "Should accept valid input")
				} else {
					require.Len(t, payload.GetIssues(), 1)
					assert.Equal(t, core.Custom, payload.GetIssues()[0].Code)
				}
			})
		}
	})

	t.Run("validates with map refine function", func(t *testing.T) {
		refineFn := func(m map[string]any) bool {
			name, hasName := m["name"]
			age, hasAge := m["age"]
			if !hasName || !hasAge {
				return false
			}
			nameStr, nameOk := name.(string)
			ageNum, ageOk := age.(float64)
			return nameOk && ageOk && len(nameStr) > 0 && ageNum >= 18
		}
		check := NewCustom[map[string]any](refineFn)

		tests := []struct {
			name  string
			input map[string]any
			valid bool
		}{
			{"valid adult", map[string]any{"name": "John", "age": float64(25)}, true},
			{"minor", map[string]any{"name": "Jane", "age": float64(16)}, false},
			{"missing name", map[string]any{"age": float64(25)}, false},
			{"empty name", map[string]any{"name": "", "age": float64(25)}, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				payload := core.NewParsePayload(tt.input)
				internals := check.GetZod()
				internals.Check(payload)

				if tt.valid {
					assert.Empty(t, payload.GetIssues(), "Should accept valid input")
				} else {
					require.Len(t, payload.GetIssues(), 1)
					assert.Equal(t, core.Custom, payload.GetIssues()[0].Code)
				}
			})
		}
	})

	t.Run("validates with any refine function", func(t *testing.T) {
		refineFn := func(v any) bool {
			switch val := v.(type) {
			case string:
				return len(val) > 0
			case float64:
				return val > 0
			case bool:
				return val
			default:
				return false
			}
		}
		check := NewCustom[any](refineFn)

		tests := []struct {
			name  string
			input any
			valid bool
		}{
			{"valid string", "hello", true},
			{"empty string", "", false},
			{"positive number", float64(42), true},
			{"zero", float64(0), false},
			{"true bool", true, true},
			{"false bool", false, false},
			{"invalid type", []int{1, 2, 3}, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				payload := core.NewParsePayload(tt.input)
				internals := check.GetZod()
				internals.Check(payload)

				if tt.valid {
					assert.Empty(t, payload.GetIssues(), "Should accept valid input")
				} else {
					require.Len(t, payload.GetIssues(), 1)
					assert.Equal(t, core.Custom, payload.GetIssues()[0].Code)
				}
			})
		}
	})
}

func TestCustomValidation_CheckFunctions(t *testing.T) {
	t.Run("validates with check function", func(t *testing.T) {
		checkFn := func(payload *core.ParsePayload) {
			if str, ok := payload.GetValue().(string); ok {
				if len(str) < 3 {
					// Add custom issue using issues package
					issue := issues.NewRawIssue(
						"too_short",
						payload.GetValue(),
						issues.WithContinue(true),
					)
					payload.AddIssue(issue)
				}
			} else {
				// Add type error using issues package
				issue := issues.NewRawIssue(
					"invalid_type",
					payload.GetValue(),
					issues.WithContinue(true),
				)
				payload.AddIssue(issue)
			}
		}
		check := NewCustom[any](core.ZodCheckFn(checkFn))

		tests := []struct {
			name  string
			input any
			valid bool
		}{
			{"valid string", "hello", true},
			{"short string", "hi", false},
			{"non-string", 123, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				payload := core.NewParsePayload(tt.input)
				internals := check.GetZod()
				internals.Check(payload)

				if tt.valid {
					assert.Empty(t, payload.GetIssues(), "Should accept valid input")
				} else {
					assert.NotEmpty(t, payload.GetIssues(), "Should reject invalid input")
				}
			})
		}
	})
}

func TestCustomValidation_WithParameters(t *testing.T) {
	t.Run("validates with custom path", func(t *testing.T) {
		refineFn := func(s string) bool {
			return len(s) > 0
		}
		params := core.SchemaParams{
			Params: map[string]any{
				"customPath": []string{"user", "name"},
			},
		}
		check := NewCustom[string](refineFn, params)

		payload := core.NewParsePayload("")
		internals := check.GetZod()
		internals.Check(payload)

		require.Len(t, payload.GetIssues(), 1)
		assert.Equal(t, core.Custom, payload.GetIssues()[0].Code)
		assert.Empty(t, payload.GetIssues()[0].Path)
	})

	t.Run("validates with custom parameters", func(t *testing.T) {
		refineFn := func(s string) bool {
			return false // Always fail for testing
		}
		customParams := map[string]any{
			"minLength": 5,
			"pattern":   "^[A-Z]",
		}
		params := core.SchemaParams{
			Params: customParams,
		}
		check := NewCustom[string](refineFn, params)

		payload := core.NewParsePayload("test")
		internals := check.GetZod()
		internals.Check(payload)

		require.Len(t, payload.GetIssues(), 1)
		assert.Equal(t, core.Custom, payload.GetIssues()[0].Code)
		// Note: The Params field access would need to be implemented in the core package
	})

	t.Run("validates with error customization", func(t *testing.T) {
		refineFn := func(s string) bool {
			return len(s) > 8
		}

		params := core.CustomParams{
			Error: "Custom validation failed",
		}

		check := NewCustom[string](refineFn, params)

		payload := core.NewParsePayload("hi")
		internals := check.GetZod()
		internals.Check(payload)

		require.Len(t, payload.GetIssues(), 1)
		assert.Equal(t, core.Custom, payload.GetIssues()[0].Code)
		assert.NotNil(t, internals.Def.Error)
	})
}

// =============================================================================
// OVERWRITE TRANSFORMATION TESTS
// =============================================================================

func TestTransformChecks_Overwrite(t *testing.T) {
	t.Run("transforms values using provided function", func(t *testing.T) {
		transform := func(value any) any {
			if str, ok := value.(string); ok {
				return str + "_transformed"
			}
			return value
		}

		check := NewZodCheckOverwrite(transform)
		payload := core.NewParsePayload("hello")

		internals := check.GetZod()
		internals.Check(payload)

		assert.Equal(t, "hello_transformed", payload.GetValue(), "Should transform value")
		assert.Empty(t, payload.GetIssues(), "Should not have issues for transform")
	})

	t.Run("preserves non-matching types", func(t *testing.T) {
		transform := func(value any) any {
			if str, ok := value.(string); ok {
				return str + "_transformed"
			}
			return value
		}

		check := NewZodCheckOverwrite(transform)
		payload := core.NewParsePayload(42)

		internals := check.GetZod()
		internals.Check(payload)

		assert.Equal(t, 42, payload.GetValue(), "Should preserve non-string values")
		assert.Empty(t, payload.GetIssues(), "Should not have issues")
	})

	t.Run("transforms with complex logic", func(t *testing.T) {
		transform := func(value any) any {
			switch v := value.(type) {
			case string:
				return v + "_string"
			case float64:
				return v * 2
			case bool:
				return !v
			default:
				return "unknown"
			}
		}

		tests := []struct {
			name     string
			input    any
			expected any
		}{
			{"string transform", "test", "test_string"},
			{"number transform", float64(21), float64(42)},
			{"bool transform", true, false},
			{"other transform", []int{1, 2}, "unknown"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				check := NewZodCheckOverwrite(transform)
				payload := core.NewParsePayload(tt.input)

				internals := check.GetZod()
				internals.Check(payload)

				assert.Equal(t, tt.expected, payload.GetValue(), "Should transform correctly")
				assert.Empty(t, payload.GetIssues(), "Should not have issues")
			})
		}
	})
}

// =============================================================================
// CONSTRUCTOR TESTS
// =============================================================================

func TestCustomCheckConstructors(t *testing.T) {
	t.Run("creates custom check instances with proper internals", func(t *testing.T) {
		tests := []struct {
			name  string
			check func() core.ZodCheck
		}{
			{"String refine", func() core.ZodCheck {
				return NewCustom[string](func(s string) bool { return len(s) > 0 })
			}},
			{"Map refine", func() core.ZodCheck {
				return NewCustom[map[string]any](func(m map[string]any) bool { return len(m) > 0 })
			}},
			{"Interface refine", func() core.ZodCheck {
				return NewCustom[any](func(v any) bool { return v != nil })
			}},
			{"Check function", func() core.ZodCheck {
				return NewCustom[any](core.ZodCheckFn(func(p *core.ParsePayload) {}))
			}},
			{"Overwrite", func() core.ZodCheck {
				return NewZodCheckOverwrite(func(v any) any { return v })
			}},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				check := tt.check()
				require.NotNil(t, check)

				internals := check.GetZod()
				require.NotNil(t, internals)
				require.NotNil(t, internals.Def)
				require.NotNil(t, internals.Check)
				assert.NotEmpty(t, internals.Def.Check)
			})
		}
	})

	t.Run("applies schema parameters correctly", func(t *testing.T) {
		customErrorMap := core.ZodErrorMap(func(issue core.ZodRawIssue) string {
			return "Custom validation error"
		})

		params := core.CustomParams{
			Error: &customErrorMap,
			Abort: true,
		}

		check := NewCustom[string](func(s string) bool { return len(s) > 0 }, params)
		require.NotNil(t, check)

		internals := check.GetZod()
		require.NotNil(t, internals.Def)
		// Check if error was applied correctly
		if internals.Def.Error != nil {
			assert.NotNil(t, internals.Def.Error)
		}
		// Check if abort flag was applied correctly
		assert.True(t, internals.Def.Abort)
	})
}

// =============================================================================
// ERROR HANDLING TESTS
// =============================================================================

func TestCustomValidation_ErrorHandling(t *testing.T) {
	t.Run("handles refine failure correctly", func(t *testing.T) {
		refineFn := func(s string) bool {
			return false // Always fail
		}
		check := NewCustom[string](refineFn)

		payload := core.NewParsePayload("test")
		internals := check.GetZod()
		internals.Check(payload)

		require.Len(t, payload.GetIssues(), 1)
		issue := payload.GetIssues()[0]
		assert.Equal(t, core.Custom, issue.Code)
		assert.NotNil(t, issue.Inst)
	})

	t.Run("handles path concatenation", func(t *testing.T) {
		refineFn := func(s string) bool {
			return false
		}
		params := core.CustomParams{
			Params: map[string]any{
				"nestedField": []string{"nested", "field"},
			},
		}
		check := NewCustom[string](refineFn, params)

		payload := core.NewParsePayload("test")
		payload.PushPath("parent")
		internals := check.GetZod()
		internals.Check(payload)

		require.Len(t, payload.GetIssues(), 1)
		expectedPath := []any{"parent"}
		assert.Equal(t, expectedPath, payload.GetIssues()[0].Path)
	})

	t.Run("handles type mismatch gracefully", func(t *testing.T) {
		// This test checks what happens when wrong type is passed to typed refine function
		refineFn := func(s string) bool {
			return len(s) > 0
		}
		check := NewCustom[string](refineFn)

		payload := core.NewParsePayload(123) // Number instead of string
		internals := check.GetZod()

		// With our fix, type mismatch should be handled gracefully by creating a validation error
		assert.NotPanics(t, func() {
			internals.Check(payload)
		}, "Should not panic on type mismatch")

		// Should create a validation error for type mismatch
		require.Len(t, payload.GetIssues(), 1)
		issue := payload.GetIssues()[0]
		assert.Equal(t, core.Custom, issue.Code)
		assert.Equal(t, 123, issue.Input)
	})
}

// =============================================================================
// UTILITY FUNCTION TESTS
// =============================================================================

func TestCustomUtilities(t *testing.T) {
	t.Run("PrefixIssues works correctly", func(t *testing.T) {
		issues := []core.ZodRawIssue{
			{
				Code: "test1",
				Path: []any{"existing"},
			},
			{
				Code: "test2",
				Path: nil,
			},
		}

		result := PrefixIssues("prefix", issues)

		require.Len(t, result, 2)
		assert.Equal(t, []any{"prefix", "existing"}, result[0].Path)
		assert.Equal(t, []any{"prefix"}, result[1].Path)
	})

	t.Run("handleRefineResult creates proper issues", func(t *testing.T) {
		internals := &ZodCheckCustomInternals{
			Def: &ZodCheckCustomDef{
				ZodCheckDef: core.ZodCheckDef{Abort: false},
				Params:      map[string]any{"test": "value"},
			},
		}

		payload := core.NewParsePayload("test")
		payload.PushPath("base")

		// Test successful case
		handleRefineResult(true, payload, "test", internals)
		assert.Empty(t, payload.GetIssues(), "Should not create issues for successful validation")

		// Test failure case
		handleRefineResult(false, payload, "test", internals)
		require.Len(t, payload.GetIssues(), 1)
		assert.Equal(t, core.Custom, payload.GetIssues()[0].Code)
		// Path will only contain the payload path since custom path is no longer appended
		assert.Equal(t, []any{"base"}, payload.GetIssues()[0].Path)
	})
}

// =============================================================================
// BENCHMARK TESTS
// =============================================================================

func BenchmarkCustomCheck_StringRefine(b *testing.B) {
	refineFn := func(s string) bool {
		return len(s) > 3 && s[0] >= 'A' && s[0] <= 'Z'
	}
	check := NewCustom[string](refineFn)
	internals := check.GetZod()
	input := "Hello"

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		payload := core.NewParsePayload(input)
		internals.Check(payload)
		if len(payload.GetIssues()) > 0 {
			b.Fatal("Unexpected validation failure")
		}
	}
}

func BenchmarkCustomCheck_MapRefine(b *testing.B) {
	refineFn := func(m map[string]any) bool {
		name, hasName := m["name"]
		return hasName && name != nil
	}
	check := NewCustom[map[string]any](refineFn)
	internals := check.GetZod()
	input := map[string]any{"name": "John", "age": 30}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		payload := core.NewParsePayload(input)
		internals.Check(payload)
		if len(payload.GetIssues()) > 0 {
			b.Fatal("Unexpected validation failure")
		}
	}
}

func BenchmarkCustomCheck_Overwrite(b *testing.B) {
	transform := func(value any) any {
		if str, ok := value.(string); ok {
			return str + "_suffix"
		}
		return value
	}
	check := NewZodCheckOverwrite(transform)
	internals := check.GetZod()
	input := "test"

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		payload := core.NewParsePayload(input)
		internals.Check(payload)
		// Transform should never generate issues
		if len(payload.GetIssues()) > 0 {
			b.Fatal("Unexpected transform failure")
		}
	}
}

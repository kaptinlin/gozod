package engine

import (
	"errors"
	"testing"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// TEST HELPERS
// =============================================================================

// mockValidator creates a simple validator function for testing
func mockValidator[T any](shouldFail bool) func(T, []core.ZodCheck, *core.ParseContext) (T, error) {
	return func(value T, checks []core.ZodCheck, ctx *core.ParseContext) (T, error) {
		if shouldFail {
			return value, errors.New("mock validation failed")
		}
		return value, nil
	}
}

// mockConverter creates a simple converter function for testing
func mockConverter[T any](result any, ctx *core.ParseContext, expectedType core.ZodTypeCode) (T, error) {
	if result == nil {
		var zero T
		return zero, nil
	}
	if val, ok := result.(T); ok {
		return val, nil
	}
	var zero T
	return zero, CreateInvalidTypeError(expectedType, result, ctx)
}

// mockTypeExtractor creates a type extractor for testing
func mockTypeExtractor[T any](shouldExtract bool) func(any) (T, bool) {
	return func(input any) (T, bool) {
		if !shouldExtract {
			var zero T
			return zero, false
		}
		if val, ok := input.(T); ok {
			return val, true
		}
		var zero T
		return zero, false
	}
}

// mockPtrExtractor creates a pointer extractor for testing
func mockPtrExtractor[T any](shouldExtract bool) func(any) (*T, bool) {
	return func(input any) (*T, bool) {
		if !shouldExtract {
			return nil, false
		}
		if ptr, ok := input.(*T); ok {
			return ptr, true
		}
		return nil, false
	}
}

// mockCheck creates a simple check for testing
type mockCheck struct{}

func (m *mockCheck) GetZod() *core.ZodCheckInternals {
	return &core.ZodCheckInternals{
		Def: &core.ZodCheckDef{
			Check: "mock",
		},
		OnAttach: []func(schema any){},
	}
}

// createMockInternals creates mock internals for testing
func createMockInternals() *core.ZodTypeInternals {
	return &core.ZodTypeInternals{
		Type:   "mock",
		Checks: []core.ZodCheck{},
		Values: make(map[any]struct{}),
		Bag:    make(map[string]any),
	}
}

// =============================================================================
// PARSEPRIMITIVE TESTS
// =============================================================================

func TestParsePrimitive(t *testing.T) {
	t.Run("successful parsing with valid input", func(t *testing.T) {
		internals := createMockInternals()
		validator := mockValidator[string](false)

		result, err := ParsePrimitive[string, string](
			"test",
			internals,
			"string",
			validator,
			mockConverter[string],
		)

		require.NoError(t, err)
		assert.Equal(t, "test", result)
	})

	t.Run("validation failure", func(t *testing.T) {
		internals := createMockInternals()
		// Add a dummy check to trigger validation
		internals.AddCheck(&mockCheck{})
		validator := mockValidator[string](true)

		_, err := ParsePrimitive[string, string](
			"test",
			internals,
			"string",
			validator,
			mockConverter[string],
		)

		require.Error(t, err)
		// The error should contain the validation failure message
		assert.Contains(t, err.Error(), "mock validation failed")
	})

	t.Run("nil input with optional flag", func(t *testing.T) {
		internals := createMockInternals()
		internals.SetOptional(true)
		validator := mockValidator[string](false)

		result, err := ParsePrimitive[string, string](
			nil,
			internals,
			"string",
			validator,
			mockConverter[string],
		)

		require.NoError(t, err)
		assert.Equal(t, "", result) // zero value for string
	})

	t.Run("nil input with nilable flag", func(t *testing.T) {
		internals := createMockInternals()
		internals.SetNilable(true)
		validator := mockValidator[string](false)

		result, err := ParsePrimitive[string, *string](
			nil,
			internals,
			"string",
			validator,
			func(result any, ctx *core.ParseContext, expectedType core.ZodTypeCode) (*string, error) {
				if result == nil {
					return nil, nil
				}
				if val, ok := result.(string); ok {
					return &val, nil
				}
				return nil, CreateInvalidTypeError(expectedType, result, ctx)
			},
		)

		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("nil input with default value", func(t *testing.T) {
		internals := createMockInternals()
		internals.SetDefaultValue("default")
		validator := mockValidator[string](false)

		result, err := ParsePrimitive[string, string](
			nil,
			internals,
			"string",
			validator,
			mockConverter[string],
		)

		require.NoError(t, err)
		assert.Equal(t, "default", result)
	})

	t.Run("nil input with default function", func(t *testing.T) {
		internals := createMockInternals()
		internals.SetDefaultFunc(func() any { return "from_func" })
		validator := mockValidator[string](false)

		result, err := ParsePrimitive[string, string](
			nil,
			internals,
			"string",
			validator,
			mockConverter[string],
		)

		require.NoError(t, err)
		assert.Equal(t, "from_func", result)
	})

	t.Run("prefault value after parsing failure", func(t *testing.T) {
		internals := createMockInternals()
		internals.SetPrefaultValue("prefault")
		validator := mockValidator[string](false)

		result, err := ParsePrimitive[string, string](
			123, // Wrong type, will cause parsing failure
			internals,
			"string",
			validator,
			mockConverter[string],
		)

		require.NoError(t, err)
		assert.Equal(t, "prefault", result)
	})

	t.Run("prefault function after parsing failure", func(t *testing.T) {
		internals := createMockInternals()
		internals.SetPrefaultFunc(func() any { return "prefault_func" })
		validator := mockValidator[string](false)

		result, err := ParsePrimitive[string, string](
			123, // Wrong type, will cause parsing failure
			internals,
			"string",
			validator,
			mockConverter[string],
		)

		require.NoError(t, err)
		assert.Equal(t, "prefault_func", result)
	})

	t.Run("transform function applied", func(t *testing.T) {
		internals := createMockInternals()
		internals.SetTransform(func(value any, ctx *core.RefinementContext) (any, error) {
			return "transformed_" + value.(string), nil
		})
		validator := mockValidator[string](false)

		result, err := ParsePrimitive[string, string](
			"test",
			internals,
			"string",
			validator,
			mockConverter[string],
		)

		require.NoError(t, err)
		assert.Equal(t, "transformed_test", result)
	})

	t.Run("transform function error", func(t *testing.T) {
		internals := createMockInternals()
		internals.SetTransform(func(value any, ctx *core.RefinementContext) (any, error) {
			return nil, errors.New("transform failed")
		})
		validator := mockValidator[string](false)

		_, err := ParsePrimitive[string, string](
			"test",
			internals,
			"string",
			validator,
			mockConverter[string],
		)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "transform failed")
	})

	t.Run("with parse context", func(t *testing.T) {
		internals := createMockInternals()
		validator := mockValidator[string](false)
		ctx := &core.ParseContext{ReportInput: true}

		result, err := ParsePrimitive[string, string](
			"test",
			internals,
			"string",
			validator,
			mockConverter[string],
			ctx,
		)

		require.NoError(t, err)
		assert.Equal(t, "test", result)
	})
}

// =============================================================================
// PARSETYPE TESTS
// =============================================================================

func TestParseComplex(t *testing.T) {
	t.Run("successful parsing with type extractor", func(t *testing.T) {
		internals := createMockInternals()
		typeExtractor := mockTypeExtractor[string](true)
		ptrExtractor := mockPtrExtractor[string](false)
		validator := mockValidator[string](false)

		result, err := ParseComplex[string](
			"test",
			internals,
			"string",
			typeExtractor,
			ptrExtractor,
			validator,
		)

		require.NoError(t, err)
		assert.Equal(t, "test", result)
	})

	t.Run("successful parsing with pointer extractor", func(t *testing.T) {
		internals := createMockInternals()
		typeExtractor := mockTypeExtractor[string](false)
		ptrExtractor := mockPtrExtractor[string](true)
		validator := mockValidator[string](false)

		testStr := "test"
		result, err := ParseComplex[string](
			&testStr,
			internals,
			"string",
			typeExtractor,
			ptrExtractor,
			validator,
		)

		require.NoError(t, err)
		// ParseComplex returns the pointer when using pointer extractor
		assert.Equal(t, &testStr, result)
	})

	t.Run("extraction failure", func(t *testing.T) {
		internals := createMockInternals()
		typeExtractor := mockTypeExtractor[string](false)
		ptrExtractor := mockPtrExtractor[string](false)
		validator := mockValidator[string](false)

		_, err := ParseComplex[string](
			123, // Wrong type
			internals,
			"string",
			typeExtractor,
			ptrExtractor,
			validator,
		)

		require.Error(t, err)
		var zodErr *issues.ZodError
		require.True(t, errors.As(err, &zodErr))
		assert.Equal(t, core.InvalidType, zodErr.Issues[0].Code)
	})

	t.Run("validation failure", func(t *testing.T) {
		internals := createMockInternals()
		// Add a dummy check to trigger validation
		internals.AddCheck(&mockCheck{})
		typeExtractor := mockTypeExtractor[string](true)
		ptrExtractor := mockPtrExtractor[string](false)
		validator := mockValidator[string](true)

		_, err := ParseComplex[string](
			"test",
			internals,
			"string",
			typeExtractor,
			ptrExtractor,
			validator,
		)

		require.Error(t, err)
		// The error should contain the validation failure message
		assert.Contains(t, err.Error(), "mock validation failed")
	})

	t.Run("nil input with optional flag", func(t *testing.T) {
		internals := createMockInternals()
		internals.SetOptional(true)
		typeExtractor := mockTypeExtractor[string](false)
		ptrExtractor := mockPtrExtractor[string](false)
		validator := mockValidator[string](false)

		result, err := ParseComplex[string](
			nil,
			internals,
			"string",
			typeExtractor,
			ptrExtractor,
			validator,
		)

		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("with transform function", func(t *testing.T) {
		internals := createMockInternals()
		internals.SetTransform(func(value any, ctx *core.RefinementContext) (any, error) {
			return "transformed_" + value.(string), nil
		})
		typeExtractor := mockTypeExtractor[string](true)
		ptrExtractor := mockPtrExtractor[string](false)
		validator := mockValidator[string](false)

		result, err := ParseComplex[string](
			"test",
			internals,
			"string",
			typeExtractor,
			ptrExtractor,
			validator,
		)

		require.NoError(t, err)
		assert.Equal(t, "transformed_test", result)
	})
}

// =============================================================================
// INTERNAL HELPER TESTS
// =============================================================================

func TestProcessModifiers(t *testing.T) {
	t.Run("non-nil input bypasses modifiers", func(t *testing.T) {
		internals := createMockInternals()
		internals.SetOptional(true)
		parseCore := func(any) (any, error) { return "parsed", nil }

		result, handled, err := processModifiers[string](
			"input",
			internals,
			"string",
			parseCore,
			&core.ParseContext{},
		)

		assert.False(t, handled)
		assert.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("nil input with optional returns nil", func(t *testing.T) {
		internals := createMockInternals()
		internals.SetOptional(true)
		parseCore := func(any) (any, error) { return "parsed", nil }

		result, handled, err := processModifiers[string](
			nil,
			internals,
			"string",
			parseCore,
			&core.ParseContext{},
		)

		assert.True(t, handled)
		assert.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("nil input with nilable returns nil", func(t *testing.T) {
		internals := createMockInternals()
		internals.SetNilable(true)
		parseCore := func(any) (any, error) { return "parsed", nil }

		result, handled, err := processModifiers[string](
			nil,
			internals,
			"string",
			parseCore,
			&core.ParseContext{},
		)

		assert.True(t, handled)
		assert.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("nil input with default value", func(t *testing.T) {
		internals := createMockInternals()
		internals.SetDefaultValue("default")
		parseCore := func(input any) (any, error) {
			return input, nil
		}

		result, handled, err := processModifiers[string](
			nil,
			internals,
			"string",
			parseCore,
			&core.ParseContext{},
		)

		assert.True(t, handled)
		assert.NoError(t, err)
		assert.Equal(t, "default", result)
	})

	t.Run("nil input without modifiers returns error", func(t *testing.T) {
		internals := createMockInternals()
		parseCore := func(any) (any, error) { return "parsed", nil }

		result, handled, err := processModifiers[string](
			nil,
			internals,
			"string",
			parseCore,
			&core.ParseContext{},
		)

		assert.True(t, handled)
		assert.Error(t, err)
		assert.Nil(t, result)

		var zodErr *issues.ZodError
		require.True(t, errors.As(err, &zodErr))
		assert.Equal(t, core.InvalidType, zodErr.Issues[0].Code)
	})
}

func TestApplyTransformIfPresent(t *testing.T) {
	t.Run("no transform function", func(t *testing.T) {
		internals := createMockInternals()
		ctx := &core.ParseContext{}

		result, err := applyTransformIfPresent("input", internals, ctx)

		assert.NoError(t, err)
		assert.Equal(t, "input", result)
	})

	t.Run("with transform function", func(t *testing.T) {
		internals := createMockInternals()
		internals.SetTransform(func(value any, ctx *core.RefinementContext) (any, error) {
			return "transformed", nil
		})
		ctx := &core.ParseContext{}

		result, err := applyTransformIfPresent("input", internals, ctx)

		assert.NoError(t, err)
		assert.Equal(t, "transformed", result)
	})

	t.Run("transform function error", func(t *testing.T) {
		internals := createMockInternals()
		internals.SetTransform(func(value any, ctx *core.RefinementContext) (any, error) {
			return nil, errors.New("transform error")
		})
		ctx := &core.ParseContext{}

		_, err := applyTransformIfPresent("input", internals, ctx)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "transform error")
	})
}

func TestTryPrefaultFallback(t *testing.T) {
	t.Run("no prefault available", func(t *testing.T) {
		internals := createMockInternals()
		parseCore := func(any) (any, error) { return "parsed", nil }
		originalErr := errors.New("original error")

		result, err := tryPrefaultFallback[string](internals, parseCore, originalErr)

		assert.Error(t, err)
		assert.Equal(t, originalErr, err)
		assert.Nil(t, result)
	})

	t.Run("prefault value success", func(t *testing.T) {
		internals := createMockInternals()
		internals.SetPrefaultValue("prefault")
		parseCore := func(input any) (any, error) {
			return input, nil
		}
		originalErr := errors.New("original error")

		result, err := tryPrefaultFallback[string](internals, parseCore, originalErr)

		assert.NoError(t, err)
		assert.Equal(t, "prefault", result)
	})

	t.Run("prefault function success", func(t *testing.T) {
		internals := createMockInternals()
		internals.SetPrefaultFunc(func() any { return "prefault_func" })
		parseCore := func(input any) (any, error) {
			return input, nil
		}
		originalErr := errors.New("original error")

		result, err := tryPrefaultFallback[string](internals, parseCore, originalErr)

		assert.NoError(t, err)
		assert.Equal(t, "prefault_func", result)
	})
}

func TestGetOrCreateContext(t *testing.T) {
	t.Run("no context provided", func(t *testing.T) {
		ctx := getOrCreateContext()

		assert.NotNil(t, ctx)
		assert.False(t, ctx.ReportInput)
	})

	t.Run("context provided", func(t *testing.T) {
		providedCtx := &core.ParseContext{ReportInput: true}
		ctx := getOrCreateContext(providedCtx)

		assert.Equal(t, providedCtx, ctx)
		assert.True(t, ctx.ReportInput)
	})

	t.Run("multiple contexts provided", func(t *testing.T) {
		ctx1 := &core.ParseContext{ReportInput: true}
		ctx2 := &core.ParseContext{ReportInput: false}
		ctx := getOrCreateContext(ctx1, ctx2)

		assert.Equal(t, ctx1, ctx) // Should use first one
		assert.True(t, ctx.ReportInput)
	})
}

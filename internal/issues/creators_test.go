package issues

import (
	"errors"
	"testing"
	"time"

	"github.com/kaptinlin/gozod/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRawIssueFromMessage(t *testing.T) {
	message := "Custom error message"
	input := "test_input"
	instance := "test_instance"

	issue := NewRawIssueFromMessage(message, input, instance)

	assert.Equal(t, core.Custom, issue.Code)
	assert.Equal(t, message, issue.Message)
	assert.Equal(t, input, issue.Input)
	assert.Equal(t, instance, issue.Inst)
	assert.Empty(t, issue.Path)
	// Properties are initialized lazily, so initially nil
	assert.Nil(t, issue.Properties)
}

func TestErrorCreationHelpers(t *testing.T) {
	t.Run("CreateInvalidTypeError", func(t *testing.T) {
		err := CreateInvalidTypeError(core.ZodTypeString, "test_input", nil)

		require.NotNil(t, err)
		assert.IsType(t, &ZodError{}, err)
		var zodErr *ZodError
		errors.As(err, &zodErr)
		require.Len(t, zodErr.Issues, 1)
		require.Equal(t, core.InvalidType, zodErr.Issues[0].Code)
		assert.Equal(t, core.ZodTypeString, zodErr.Issues[0].Expected)
	})

	t.Run("CreateTooBigError", func(t *testing.T) {
		err := CreateTooBigError(100, true, "number", 150, nil)

		require.NotNil(t, err)
		assert.IsType(t, &ZodError{}, err)
		var zodErr *ZodError
		errors.As(err, &zodErr)
		require.Len(t, zodErr.Issues, 1)
		require.Equal(t, core.TooBig, zodErr.Issues[0].Code)
		assert.Equal(t, 100, zodErr.Issues[0].Maximum)
		assert.True(t, zodErr.Issues[0].Inclusive)
		assert.Equal(t, "number", zodErr.Issues[0].Origin)
	})

	t.Run("CreateTooSmallError", func(t *testing.T) {
		err := CreateTooSmallError(5, false, "string", 3, nil)

		require.NotNil(t, err)
		assert.IsType(t, &ZodError{}, err)
		var zodErr *ZodError
		errors.As(err, &zodErr)
		require.Len(t, zodErr.Issues, 1)
		require.Equal(t, core.TooSmall, zodErr.Issues[0].Code)
		assert.Equal(t, 5, zodErr.Issues[0].Minimum)
		assert.False(t, zodErr.Issues[0].Inclusive)
		assert.Equal(t, "string", zodErr.Issues[0].Origin)
	})

	t.Run("CreateInvalidFormatError", func(t *testing.T) {
		err := CreateInvalidFormatError("email", "invalid@", nil)

		require.NotNil(t, err)
		assert.IsType(t, &ZodError{}, err)
		var zodErr *ZodError
		errors.As(err, &zodErr)
		require.Len(t, zodErr.Issues, 1)
		require.Equal(t, core.InvalidFormat, zodErr.Issues[0].Code)
		assert.Equal(t, "email", zodErr.Issues[0].Format)
	})

	t.Run("CreateNotMultipleOfError", func(t *testing.T) {
		err := CreateNotMultipleOfError(2, "number", 7, nil)

		require.NotNil(t, err)
		assert.IsType(t, &ZodError{}, err)
		var zodErr *ZodError
		errors.As(err, &zodErr)
		require.Len(t, zodErr.Issues, 1)
		require.Equal(t, core.NotMultipleOf, zodErr.Issues[0].Code)
		assert.Equal(t, 2, zodErr.Issues[0].Divisor)
	})

	t.Run("CreateCustomError", func(t *testing.T) {
		err := CreateCustomError("Custom validation failed", nil, "test_input", nil)

		require.NotNil(t, err)
		assert.IsType(t, &ZodError{}, err)
		var zodErr *ZodError
		errors.As(err, &zodErr)
		require.Len(t, zodErr.Issues, 1)
		require.Equal(t, core.Custom, zodErr.Issues[0].Code)
		require.Equal(t, "Custom validation failed", zodErr.Issues[0].Message)
	})

	t.Run("CreateInvalidValueError", func(t *testing.T) {
		validValues := []any{"val1", "val2", "val3"}
		err := CreateInvalidValueError(validValues, "invalid_value", nil)

		require.NotNil(t, err)
		assert.IsType(t, &ZodError{}, err)
		var zodErr *ZodError
		errors.As(err, &zodErr)
		require.Len(t, zodErr.Issues, 1)
		require.Equal(t, core.InvalidValue, zodErr.Issues[0].Code)
		assert.Equal(t, validValues, zodErr.Issues[0].Values)
	})

	t.Run("CreateUnrecognizedKeysError", func(t *testing.T) {
		keys := []string{"extraKey1", "extraKey2"}
		err := CreateUnrecognizedKeysError(keys, nil, nil)

		require.NotNil(t, err)
		assert.IsType(t, &ZodError{}, err)
		var zodErr *ZodError
		errors.As(err, &zodErr)
		require.Len(t, zodErr.Issues, 1)
		require.Equal(t, core.UnrecognizedKeys, zodErr.Issues[0].Code)
		assert.Equal(t, keys, zodErr.Issues[0].Keys)
	})
}

func TestCreationHelpersWithOptions(t *testing.T) {
	t.Run("helper functions accept additional options", func(t *testing.T) {
		path := []any{"user", "email"}
		issue := NewRawIssue(core.InvalidFormat, "invalid@",
			WithPath(path),
			WithMessage("Custom email error"),
			WithFormat("email"),
		)

		assert.Equal(t, path, issue.Path)
		assert.Equal(t, "Custom email error", issue.Message)
		format, _ := issue.Properties["format"].(string)
		assert.Equal(t, "email", format)
	})

	// Note: Low-level issue creation functions have been removed
	// in favor of high-level error creation functions
	t.Skip("Low-level issue creation functions have been removed")
}

func TestComplexCreationScenarios(t *testing.T) {
	t.Run("create issue with all possible properties", func(t *testing.T) {
		keys := []string{"key1", "key2"}
		values := []any{"val1", "val2"}
		params := map[string]any{"custom": "data", "level": 2}
		path := []any{"nested", "deep", "field"}
		instance := "complex_schema"

		issue := NewRawIssue(core.Custom, "complex_input",
			WithExpected("complex_type"),
			WithReceived("simple_type"),
			WithMinimum(10),
			WithMaximum(100),
			WithInclusive(true),
			WithOrigin("validation"),
			WithFormat("complex_format"),
			WithPattern("^complex.*$"),
			WithPrefix("complex_"),
			WithSuffix("_complex"),
			WithIncludes("complex"),
			WithDivisor(5),
			WithKeys(keys),
			WithValues(values),
			WithAlgorithm("COMPLEX256"),
			WithParams(params),
			WithPath(path),
			WithInst(instance),
			WithContinue(false),
			WithMessage("Complex validation error"),
		)

		// Verify all properties are set correctly
		assert.Equal(t, core.Custom, issue.Code)
		assert.Equal(t, "complex_input", issue.Input)
		assert.Equal(t, "Complex validation error", issue.Message)
		assert.Equal(t, "complex_type", issue.Properties["expected"])
		assert.Equal(t, "simple_type", issue.Properties["received"])
		assert.Equal(t, 10, issue.Properties["minimum"])
		assert.Equal(t, 100, issue.Properties["maximum"])
		assert.True(t, issue.Properties["inclusive"].(bool))
		assert.Equal(t, "validation", issue.Properties["origin"])
		assert.Equal(t, "complex_format", issue.Properties["format"])
		assert.Equal(t, "^complex.*$", issue.Properties["pattern"])
		assert.Equal(t, "complex_", issue.Properties["prefix"])
		assert.Equal(t, "_complex", issue.Properties["suffix"])
		assert.Equal(t, "complex", issue.Properties["includes"])
		assert.Equal(t, 5, issue.Properties["divisor"])
		assert.Equal(t, keys, issue.Properties["keys"])
		assert.Equal(t, values, issue.Properties["values"])
		assert.Equal(t, "COMPLEX256", issue.Properties["algorithm"])
		assert.Equal(t, params, issue.Properties["params"])
		assert.Equal(t, path, issue.Path)
		assert.Equal(t, instance, issue.Inst)
		assert.False(t, issue.Continue)
	})

	// Note: Low-level issue creation functions have been removed
	// in favor of high-level error creation functions
	t.Skip("Low-level issue creation functions have been removed")
}

func TestCreationEdgeCases(t *testing.T) {
	// Note: Low-level issue creation functions have been removed
	// in favor of high-level error creation functions
	t.Skip("Low-level issue creation functions have been removed")
}

func TestCreationPerformance(t *testing.T) {
	// Note: Low-level issue creation functions have been removed
	// in favor of high-level error creation functions
	t.Skip("Low-level issue creation functions have been removed")
}

func TestHighLevelErrorAPI(t *testing.T) {
	ctx := &core.ParseContext{}

	t.Run("CreateFinalError", func(t *testing.T) {
		properties := map[string]any{
			"expected": "string",
			"received": "number",
		}
		err := CreateFinalError(core.InvalidType, "Type mismatch", properties, 123, ctx, nil)

		assert.Error(t, err)
		assert.IsType(t, &ZodError{}, err)
		var zodErr *ZodError
		errors.As(err, &zodErr)
		assert.Len(t, zodErr.Issues, 1)
		assert.Equal(t, core.InvalidType, zodErr.Issues[0].Code)
	})

	t.Run("CreateInvalidTypeError", func(t *testing.T) {
		err := CreateInvalidTypeError(core.ZodTypeString, 123, ctx)

		assert.Error(t, err)
		assert.IsType(t, &ZodError{}, err)
		var zodErr *ZodError
		errors.As(err, &zodErr)
		assert.Len(t, zodErr.Issues, 1)
		assert.Equal(t, core.InvalidType, zodErr.Issues[0].Code)
	})

	t.Run("CreateNonOptionalError", func(t *testing.T) {
		err := CreateNonOptionalError(ctx)

		assert.Error(t, err)
		assert.IsType(t, &ZodError{}, err)
		var zodErr *ZodError
		errors.As(err, &zodErr)
		assert.Len(t, zodErr.Issues, 1)
		assert.Equal(t, core.InvalidType, zodErr.Issues[0].Code)
	})

	t.Run("CreateInvalidValueError", func(t *testing.T) {
		validValues := []any{"val1", "val2", "val3"}
		err := CreateInvalidValueError(validValues, "invalid", ctx)

		assert.Error(t, err)
		assert.IsType(t, &ZodError{}, err)
		var zodErr *ZodError
		errors.As(err, &zodErr)
		assert.Len(t, zodErr.Issues, 1)
		assert.Equal(t, core.InvalidValue, zodErr.Issues[0].Code)
	})

	t.Run("CreateTooBigError", func(t *testing.T) {
		err := CreateTooBigError(100, true, "number", 150, ctx)

		assert.Error(t, err)
		assert.IsType(t, &ZodError{}, err)
		var zodErr *ZodError
		errors.As(err, &zodErr)
		assert.Len(t, zodErr.Issues, 1)
		assert.Equal(t, core.TooBig, zodErr.Issues[0].Code)
	})

	t.Run("CreateTooSmallError", func(t *testing.T) {
		err := CreateTooSmallError(5, false, "string", 3, ctx)

		assert.Error(t, err)
		assert.IsType(t, &ZodError{}, err)
		var zodErr *ZodError
		errors.As(err, &zodErr)
		assert.Len(t, zodErr.Issues, 1)
		assert.Equal(t, core.TooSmall, zodErr.Issues[0].Code)
	})

	t.Run("CreateInvalidFormatError", func(t *testing.T) {
		err := CreateInvalidFormatError("email", "invalid@", ctx)

		assert.Error(t, err)
		assert.IsType(t, &ZodError{}, err)
		var zodErr *ZodError
		errors.As(err, &zodErr)
		assert.Len(t, zodErr.Issues, 1)
		assert.Equal(t, core.InvalidFormat, zodErr.Issues[0].Code)
	})

	t.Run("CreateCustomError", func(t *testing.T) {
		properties := map[string]any{"custom": "value"}
		err := CreateCustomError("Custom validation failed", properties, "input", ctx)

		assert.Error(t, err)
		assert.IsType(t, &ZodError{}, err)
		var zodErr *ZodError
		errors.As(err, &zodErr)
		assert.Len(t, zodErr.Issues, 1)
		assert.Equal(t, core.Custom, zodErr.Issues[0].Code)
		assert.Equal(t, "Custom validation failed", zodErr.Issues[0].Message)
	})

	t.Run("CreateUnrecognizedKeysError", func(t *testing.T) {
		keys := []string{"extraKey1", "extraKey2"}
		err := CreateUnrecognizedKeysError(keys, nil, ctx)

		assert.Error(t, err)
		assert.IsType(t, &ZodError{}, err)
		var zodErr *ZodError
		errors.As(err, &zodErr)
		assert.Len(t, zodErr.Issues, 1)
		assert.Equal(t, core.UnrecognizedKeys, zodErr.Issues[0].Code)
	})
}

func TestAPIComparison(t *testing.T) {
	ctx := &core.ParseContext{}

	t.Run("high-level error API functionality", func(t *testing.T) {
		// Test the new high-level API
		err := CreateInvalidTypeError(core.ZodTypeString, 123, ctx)

		// Should produce proper error
		assert.IsType(t, &ZodError{}, err)

		// Should be *ZodError type
		var zodErr *ZodError
		ok := errors.As(err, &zodErr)
		require.True(t, ok)

		assert.Len(t, zodErr.Issues, 1)
		assert.Equal(t, core.InvalidType, zodErr.Issues[0].Code)
	})

	t.Run("performance test for high-level API", func(t *testing.T) {
		// Test performance of the new high-level API
		n := 1000

		// Measure new API performance
		start := time.Now()
		for i := 0; i < n; i++ {
			_ = CreateInvalidTypeError(core.ZodTypeString, i, ctx)
		}
		duration := time.Since(start)

		// Should complete in reasonable time (less than 1 second for 1000 operations)
		assert.True(t, duration < time.Second, "High-level API should be performant")
	})
}

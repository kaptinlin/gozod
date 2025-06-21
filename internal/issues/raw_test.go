package issues

import (
	"testing"

	"github.com/kaptinlin/gozod/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//////////////////////////////////////////
//////////   Raw Issue Creation Tests  ///
//////////////////////////////////////////

func TestNewRawIssue(t *testing.T) {
	t.Run("creates raw issue with basic properties", func(t *testing.T) {
		issue := NewRawIssue("invalid_type", "test_input",
			WithExpected("string"),
			WithReceived("number"),
			WithMessage("Custom message"),
		)

		require.Equal(t, core.InvalidType, issue.Code)
		require.Equal(t, "test_input", issue.Input)
		require.Equal(t, "Custom message", issue.Message)
		require.NotNil(t, issue.Properties)
		assert.Equal(t, "string", issue.Properties["expected"])
		assert.Equal(t, "number", issue.Properties["received"])
		assert.Empty(t, issue.Path) // Should be initialized as empty slice
	})

	t.Run("creates raw issue with path information", func(t *testing.T) {
		path := []any{"user", "name"}
		issue := NewRawIssue("too_small", "hi",
			WithMinimum(5),
			WithOrigin("string"),
			WithPath(path),
		)

		require.Equal(t, core.TooSmall, issue.Code)
		assert.Equal(t, path, issue.Path)
		assert.Equal(t, 5, issue.Properties["minimum"])
		assert.Equal(t, "string", issue.Properties["origin"])
	})

	t.Run("creates raw issue with all option types", func(t *testing.T) {
		keys := []string{"invalid_key1", "invalid_key2"}
		values := []any{"val1", "val2"}
		params := map[string]any{"custom": "data"}

		issue := NewRawIssue("custom", nil,
			WithKeys(keys),
			WithValues(values),
			WithFormat("email"),
			WithPattern(".*@.*"),
			WithPrefix("test_"),
			WithSuffix("_end"),
			WithIncludes("@"),
			WithDivisor(2),
			WithAlgorithm("HS256"),
			WithParams(params),
			WithInclusive(true),
			WithContinue(false),
		)

		require.Equal(t, core.Custom, issue.Code)
		assert.Equal(t, keys, issue.Properties["keys"])
		assert.Equal(t, values, issue.Properties["values"])
		assert.Equal(t, "email", issue.Properties["format"])
		assert.Equal(t, ".*@.*", issue.Properties["pattern"])
		assert.Equal(t, "test_", issue.Properties["prefix"])
		assert.Equal(t, "_end", issue.Properties["suffix"])
		assert.Equal(t, "@", issue.Properties["includes"])
		assert.Equal(t, 2, issue.Properties["divisor"])
		assert.Equal(t, "HS256", issue.Properties["algorithm"])
		assert.Equal(t, params, issue.Properties["params"])
		assert.Equal(t, true, issue.Properties["inclusive"])
		assert.Equal(t, false, issue.Continue)
	})

	t.Run("handles nil properties map gracefully", func(t *testing.T) {
		issue := NewRawIssue("custom", "test")
		// Properties are initialized lazily, so initially nil
		assert.Nil(t, issue.Properties)

		// But when we use an option that needs properties, it gets initialized
		WithExpected("string")(&issue)
		require.NotNil(t, issue.Properties)
		assert.Equal(t, "string", issue.Properties["expected"])
	})
}

//////////////////////////////////////////
//////////   Option Functions Tests    ///
//////////////////////////////////////////

func TestOptionFunctions(t *testing.T) {
	t.Run("WithOrigin sets origin property", func(t *testing.T) {
		issue := NewRawIssue("test", nil, WithOrigin("string"))
		assert.Equal(t, "string", issue.Properties["origin"])
	})

	t.Run("WithMessage sets message field", func(t *testing.T) {
		issue := NewRawIssue("test", nil, WithMessage("Test message"))
		assert.Equal(t, "Test message", issue.Message)
	})

	t.Run("WithMinimum sets minimum property", func(t *testing.T) {
		issue := NewRawIssue("test", nil, WithMinimum(10))
		assert.Equal(t, 10, issue.Properties["minimum"])
	})

	t.Run("WithMaximum sets maximum property", func(t *testing.T) {
		issue := NewRawIssue("test", nil, WithMaximum(100))
		assert.Equal(t, 100, issue.Properties["maximum"])
	})

	t.Run("WithExpected sets expected property", func(t *testing.T) {
		issue := NewRawIssue("test", nil, WithExpected("string"))
		assert.Equal(t, "string", issue.Properties["expected"])
	})

	t.Run("WithReceived sets received property", func(t *testing.T) {
		issue := NewRawIssue("test", nil, WithReceived("number"))
		assert.Equal(t, "number", issue.Properties["received"])
	})

	t.Run("WithPath sets path field", func(t *testing.T) {
		path := []any{"user", "name"}
		issue := NewRawIssue("test", nil, WithPath(path))
		assert.Equal(t, path, issue.Path)
	})

	t.Run("WithInclusive sets inclusive property", func(t *testing.T) {
		issue := NewRawIssue("test", nil, WithInclusive(true))
		assert.Equal(t, true, issue.Properties["inclusive"])
	})

	t.Run("WithFormat sets format property", func(t *testing.T) {
		issue := NewRawIssue("test", nil, WithFormat("email"))
		assert.Equal(t, "email", issue.Properties["format"])
	})

	t.Run("WithContinue sets continue field", func(t *testing.T) {
		issue := NewRawIssue("test", nil, WithContinue(false))
		assert.False(t, issue.Continue)
	})

	t.Run("WithPattern sets pattern property", func(t *testing.T) {
		issue := NewRawIssue("test", nil, WithPattern(".*@.*"))
		assert.Equal(t, ".*@.*", issue.Properties["pattern"])
	})

	t.Run("WithPrefix sets prefix property", func(t *testing.T) {
		issue := NewRawIssue("test", nil, WithPrefix("test_"))
		assert.Equal(t, "test_", issue.Properties["prefix"])
	})

	t.Run("WithSuffix sets suffix property", func(t *testing.T) {
		issue := NewRawIssue("test", nil, WithSuffix("_end"))
		assert.Equal(t, "_end", issue.Properties["suffix"])
	})

	t.Run("WithIncludes sets includes property", func(t *testing.T) {
		issue := NewRawIssue("test", nil, WithIncludes("@"))
		assert.Equal(t, "@", issue.Properties["includes"])
	})

	t.Run("WithDivisor sets divisor property", func(t *testing.T) {
		issue := NewRawIssue("test", nil, WithDivisor(2.5))
		assert.Equal(t, 2.5, issue.Properties["divisor"])
	})

	t.Run("WithKeys sets keys property", func(t *testing.T) {
		keys := []string{"key1", "key2"}
		issue := NewRawIssue("test", nil, WithKeys(keys))
		assert.Equal(t, keys, issue.Properties["keys"])
	})

	t.Run("WithValues sets values property", func(t *testing.T) {
		values := []any{"val1", "val2"}
		issue := NewRawIssue("test", nil, WithValues(values))
		assert.Equal(t, values, issue.Properties["values"])
	})

	t.Run("WithAlgorithm sets algorithm property", func(t *testing.T) {
		issue := NewRawIssue("test", nil, WithAlgorithm("HS256"))
		assert.Equal(t, "HS256", issue.Properties["algorithm"])
	})

	t.Run("WithParams sets params property", func(t *testing.T) {
		params := map[string]any{"custom": "data"}
		issue := NewRawIssue("test", nil, WithParams(params))
		assert.Equal(t, params, issue.Properties["params"])
	})

	t.Run("WithInst sets inst field", func(t *testing.T) {
		inst := "test_instance"
		issue := NewRawIssue("test", nil, WithInst(inst))
		assert.Equal(t, inst, issue.Inst)
	})
}

//////////////////////////////////////////
//////////   Property Initialization Tests ///
//////////////////////////////////////////

func TestPropertyInitialization(t *testing.T) {
	t.Run("properties map is initialized when needed", func(t *testing.T) {
		// Start with empty issue
		issue := ZodRawIssue{
			Code:  "test",
			Input: nil,
		}

		// Apply options that need properties
		WithExpected("string")(&issue)
		WithReceived("number")(&issue)

		require.NotNil(t, issue.Properties)
		assert.Equal(t, "string", issue.Properties["expected"])
		assert.Equal(t, "number", issue.Properties["received"])
	})

	t.Run("properties map is reused when it exists", func(t *testing.T) {
		issue := ZodRawIssue{
			Code:       "test",
			Properties: map[string]any{"existing": "value"},
		}

		WithExpected("string")(&issue)

		assert.Equal(t, "value", issue.Properties["existing"])
		assert.Equal(t, "string", issue.Properties["expected"])
	})

	t.Run("path is initialized as empty slice", func(t *testing.T) {
		issue := NewRawIssue("test", nil)
		require.NotNil(t, issue.Path)
		assert.Empty(t, issue.Path)
	})
}

//////////////////////////////////////////
//////////   Edge Cases Tests          ///
//////////////////////////////////////////

func TestRawIssueEdgeCases(t *testing.T) {
	t.Run("handles nil input gracefully", func(t *testing.T) {
		issue := NewRawIssue(core.IssueCode("test"), nil)
		assert.Equal(t, core.IssueCode("test"), issue.Code)
		assert.Nil(t, issue.Input)
	})

	t.Run("handles empty code gracefully", func(t *testing.T) {
		issue := NewRawIssue("", "test")
		assert.Empty(t, issue.Code)
		assert.Equal(t, "test", issue.Input)
	})

	t.Run("handles nil path in WithPath", func(t *testing.T) {
		issue := NewRawIssue("test", nil, WithPath(nil))
		assert.Nil(t, issue.Path)
	})

	t.Run("handles empty path in WithPath", func(t *testing.T) {
		issue := NewRawIssue("test", nil, WithPath([]any{}))
		require.NotNil(t, issue.Path)
		assert.Empty(t, issue.Path)
	})

	t.Run("handles nil properties maps in option functions", func(t *testing.T) {
		issue := ZodRawIssue{Code: "test", Properties: nil}

		assert.NotPanics(t, func() {
			WithExpected("string")(&issue)
		})

		require.NotNil(t, issue.Properties)
		assert.Equal(t, "string", issue.Properties["expected"])
	})
}

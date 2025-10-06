package issues

import (
	"testing"

	"github.com/kaptinlin/gozod/core"
	"github.com/stretchr/testify/assert"
)

//////////////////////////////////////////
//////////   Issue Constants Tests     ///
//////////////////////////////////////////

func TestIssueConstants(t *testing.T) {
	t.Run("all issue codes are defined correctly", func(t *testing.T) {
		expectedCodes := map[IssueCode]string{
			InvalidType:      "invalid_type",
			TooBig:           "too_big",
			TooSmall:         "too_small",
			InvalidFormat:    "invalid_format",
			NotMultipleOf:    "not_multiple_of",
			UnrecognizedKeys: "unrecognized_keys",
			InvalidUnion:     "invalid_union",
			InvalidKey:       "invalid_key",
			InvalidElement:   "invalid_element",
			InvalidValue:     "invalid_value",
			Custom:           "custom",
		}

		for code, expected := range expectedCodes {
			assert.Equal(t, expected, string(code))
		}
	})

	t.Run("issue codes can be used for comparison", func(t *testing.T) {
		// Test issue code usage in ZodRawIssue
		rawIssue := ZodRawIssue{
			Code: InvalidType,
		}
		assert.Equal(t, InvalidType, rawIssue.Code)

		// Test issue code usage in ZodIssue
		finalizedIssue := ZodIssue{
			ZodIssueBase: ZodIssueBase{Code: TooBig},
		}
		assert.Equal(t, TooBig, finalizedIssue.Code)
	})
}

//////////////////////////////////////////
//////////   ZodIssue Interface Tests  ///
//////////////////////////////////////////

func TestZodIssueInterface(t *testing.T) {
	t.Run("ZodIssue implements error interface", func(t *testing.T) {
		issue := ZodIssue{
			ZodIssueBase: ZodIssueBase{
				Code:    "custom",
				Message: "Test error message",
			},
		}

		var err error = issue
		assert.Equal(t, "Test error message", err.Error())
	})

	t.Run("ZodIssue string representation", func(t *testing.T) {
		issue := ZodIssue{
			ZodIssueBase: ZodIssueBase{
				Code:    "invalid_type",
				Message: "Expected string, received number",
				Path:    []any{"user", "name"},
			},
		}

		str := issue.String()
		assert.Contains(t, str, "ZodIssue")
		assert.Contains(t, str, "invalid_type")
		assert.Contains(t, str, "Expected string, received number")
		assert.Contains(t, str, "[user name]")
	})

	t.Run("ZodIssue handles empty path", func(t *testing.T) {
		issue := ZodIssue{
			ZodIssueBase: ZodIssueBase{
				Code:    "custom",
				Message: "Error",
				Path:    []any{},
			},
		}

		str := issue.String()
		assert.Contains(t, str, "Path: []")
	})

	t.Run("ZodIssue handles nil path", func(t *testing.T) {
		issue := ZodIssue{
			ZodIssueBase: ZodIssueBase{
				Code:    "custom",
				Message: "Error",
				Path:    nil,
			},
		}

		// Should not panic
		assert.NotPanics(t, func() {
			_ = issue.String()
		})
	})
}

//////////////////////////////////////////
//////////   Type Definition Tests     ///
//////////////////////////////////////////

func TestTypeDefinitions(t *testing.T) {
	t.Run("ParseParams type definition", func(t *testing.T) {
		params := ParseParams{
			Error:       nil,
			ReportInput: true,
		}

		assert.Nil(t, params.Error)
		assert.True(t, params.ReportInput)
	})

	t.Run("ZodIssueBase type definition", func(t *testing.T) {
		base := ZodIssueBase{
			Code:    IssueCode("test_code"),
			Input:   "test_input",
			Path:    []any{"field"},
			Message: "test_message",
		}

		assert.Equal(t, IssueCode("test_code"), base.Code)
		assert.Equal(t, "test_input", base.Input)
		assert.Equal(t, []any{"field"}, base.Path)
		assert.Equal(t, "test_message", base.Message)
	})

	t.Run("ZodRawIssue type definition", func(t *testing.T) {
		rawIssue := ZodRawIssue{
			Code:       IssueCode("test_code"),
			Input:      "test_input",
			Path:       []any{"field"},
			Message:    "test_message",
			Properties: map[string]any{"key": "value"},
			Continue:   false,
			Inst:       nil,
		}

		assert.Equal(t, IssueCode("test_code"), rawIssue.Code)
		assert.Equal(t, "test_input", rawIssue.Input)
		assert.Equal(t, []any{"field"}, rawIssue.Path)
		assert.Equal(t, "test_message", rawIssue.Message)
		assert.Equal(t, "value", rawIssue.Properties["key"])
		assert.False(t, rawIssue.Continue)
		assert.Nil(t, rawIssue.Inst)
	})

	t.Run("ZodIssue type definition with all fields", func(t *testing.T) {
		issue := ZodIssue{
			ZodIssueBase: ZodIssueBase{
				Code:    InvalidType,
				Message: "Test message",
				Path:    []any{"field"},
			},
			Expected:  core.ZodTypeString,
			Received:  core.ZodTypeNumber,
			Minimum:   5,
			Maximum:   100,
			Inclusive: true,
			Keys:      []string{"key1", "key2"},
			Format:    "email",
			Divisor:   2,
			Pattern:   ".*@.*",
			Includes:  "@",
			Prefix:    "test_",
			Suffix:    "_end",
			Values:    []any{"val1", "val2"},
			Algorithm: "HS256",
			Origin:    "string",
			Key:       "field_name",
			Params:    map[string]any{"custom": "data"},
		}

		assert.Equal(t, InvalidType, issue.Code)
		assert.Equal(t, core.ZodTypeString, issue.Expected)
		assert.Equal(t, core.ZodTypeNumber, issue.Received)
		assert.Equal(t, 5, issue.Minimum)
		assert.Equal(t, 100, issue.Maximum)
		assert.True(t, issue.Inclusive)
		assert.Equal(t, []string{"key1", "key2"}, issue.Keys)
		assert.Equal(t, "email", issue.Format)
		assert.Equal(t, 2, issue.Divisor)
		assert.Equal(t, ".*@.*", issue.Pattern)
		assert.Equal(t, "@", issue.Includes)
		assert.Equal(t, "test_", issue.Prefix)
		assert.Equal(t, "_end", issue.Suffix)
		assert.Equal(t, []any{"val1", "val2"}, issue.Values)
		assert.Equal(t, "HS256", issue.Algorithm)
		assert.Equal(t, "string", issue.Origin)
		assert.Equal(t, "field_name", issue.Key)
		assert.Equal(t, map[string]any{"custom": "data"}, issue.Params)
	})
}

//////////////////////////////////////////
//////////   Issue Subtype Tests       ///
//////////////////////////////////////////

func TestIssueSubtypes(t *testing.T) {
	t.Run("ZodIssueInvalidType structure", func(t *testing.T) {
		issue := ZodIssueInvalidType{
			ZodIssueBase: ZodIssueBase{
				Code:    InvalidType,
				Message: "Type error",
			},
			Expected: core.ZodTypeString,
			Received: core.ParsedTypeNumber,
		}

		assert.Equal(t, InvalidType, issue.Code)
		assert.Equal(t, core.ZodTypeString, issue.Expected)
		assert.Equal(t, core.ParsedTypeNumber, issue.Received)
	})

	t.Run("ZodIssueTooBig structure", func(t *testing.T) {
		issue := ZodIssueTooBig{
			ZodIssueBase: ZodIssueBase{
				Code:    TooBig,
				Message: "Value too big",
			},
			Origin:    "number",
			Maximum:   100,
			Inclusive: true,
		}

		assert.Equal(t, TooBig, issue.Code)
		assert.Equal(t, "number", issue.Origin)
		assert.Equal(t, 100, issue.Maximum)
		assert.True(t, issue.Inclusive)
	})

	t.Run("ZodIssueTooSmall structure", func(t *testing.T) {
		issue := ZodIssueTooSmall{
			ZodIssueBase: ZodIssueBase{
				Code:    TooSmall,
				Message: "Value too small",
			},
			Origin:    "string",
			Minimum:   5,
			Inclusive: false,
		}

		assert.Equal(t, TooSmall, issue.Code)
		assert.Equal(t, "string", issue.Origin)
		assert.Equal(t, 5, issue.Minimum)
		assert.False(t, issue.Inclusive)
	})

	t.Run("ZodIssueInvalidStringFormat structure", func(t *testing.T) {
		issue := ZodIssueInvalidStringFormat{
			ZodIssueBase: ZodIssueBase{
				Code:    InvalidFormat,
				Message: "Invalid format",
			},
			Format:  "email",
			Pattern: ".*@.*",
		}

		assert.Equal(t, InvalidFormat, issue.Code)
		assert.Equal(t, "email", issue.Format)
		assert.Equal(t, ".*@.*", issue.Pattern)
	})

	t.Run("string format issue subtypes", func(t *testing.T) {
		regexIssue := ZodIssueStringInvalidRegex{
			ZodIssueInvalidStringFormat: ZodIssueInvalidStringFormat{
				ZodIssueBase: ZodIssueBase{Code: InvalidFormat},
				Format:       "regex",
			},
			Pattern: "^[a-z]+$",
		}

		jwtIssue := ZodIssueStringInvalidJWT{
			ZodIssueInvalidStringFormat: ZodIssueInvalidStringFormat{
				ZodIssueBase: ZodIssueBase{Code: InvalidFormat},
				Format:       "jwt",
			},
			Algorithm: "HS256",
		}

		startsWithIssue := ZodIssueStringStartsWith{
			ZodIssueInvalidStringFormat: ZodIssueInvalidStringFormat{
				ZodIssueBase: ZodIssueBase{Code: InvalidFormat},
				Format:       "starts_with",
			},
			Prefix: "prefix_",
		}

		assert.Equal(t, "^[a-z]+$", regexIssue.Pattern)
		assert.Equal(t, "HS256", jwtIssue.Algorithm)
		assert.Equal(t, "prefix_", startsWithIssue.Prefix)
	})
}

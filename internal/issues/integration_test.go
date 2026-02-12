package issues

import (
	"fmt"
	"testing"

	"github.com/kaptinlin/gozod/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIssueLifecycleIntegration(t *testing.T) {
	t.Run("full issue lifecycle from creation to finalization", func(t *testing.T) {
		// Create raw issue
		rawIssue := NewRawIssue(core.InvalidType, "123",
			WithExpected("string"),
			WithReceived("number"),
			WithPath([]any{"user", "name"}),
		)

		// Verify raw issue properties
		require.Equal(t, core.InvalidType, rawIssue.Code)
		require.Equal(t, core.ZodTypeString, rawIssue.Expected())
		require.Equal(t, core.ZodTypeNumber, rawIssue.Received())

		// Finalize issue
		ctx := &core.ParseContext{ReportInput: true}
		config := &core.ZodConfig{}
		finalIssue := FinalizeIssue(rawIssue, ctx, config)

		// Verify finalized issue
		require.Equal(t, core.InvalidType, finalIssue.Code)
		require.Equal(t, "123", finalIssue.Input)
		require.Equal(t, []any{"user", "name"}, finalIssue.Path)
		assert.Equal(t, core.ZodTypeString, finalIssue.Expected)
		assert.Equal(t, core.ZodTypeNumber, finalIssue.Received)

		// Verify type-specific accessors work
		expected, ok := finalIssue.ExpectedType()
		require.True(t, ok)
		assert.Equal(t, core.ZodTypeString, expected)
	})

	t.Run("issue creation with all helper functions", func(t *testing.T) {
		testCases := []struct {
			name   string
			create func() core.ZodRawIssue
		}{
			{
				name: "invalid type helper",
				create: func() core.ZodRawIssue {
					return NewRawIssue(core.InvalidType, "test", WithExpected("string"), WithReceived("number"))
				},
			},
			{
				name: "too big helper",
				create: func() core.ZodRawIssue {
					return NewRawIssue(core.TooBig, 150, WithMaximum(100), WithInclusive(true), WithOrigin("number"))
				},
			},
			{
				name: "too small helper",
				create: func() core.ZodRawIssue {
					return NewRawIssue(core.TooSmall, 3, WithMinimum(5), WithInclusive(false), WithOrigin("string"))
				},
			},
			{
				name:   "invalid format helper",
				create: func() core.ZodRawIssue { return NewRawIssue(core.InvalidFormat, "invalid@", WithFormat("email")) },
			},
			{
				name: "not multiple of helper",
				create: func() core.ZodRawIssue {
					return NewRawIssue(core.NotMultipleOf, 7, WithDivisor(2), WithOrigin("number"))
				},
			},
			{
				name:   "custom helper",
				create: func() core.ZodRawIssue { return NewRawIssueFromMessage("Custom error", "test", nil) },
			},
		}

		config := &core.ZodConfig{}
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				rawIssue := tc.create()

				// Should be able to finalize without errors
				finalIssue := FinalizeIssue(rawIssue, nil, config)
				assert.NotEmpty(t, finalIssue.Code)
				assert.NotEmpty(t, finalIssue.Message)
			})
		}
	})
}

func TestErrorProcessingIntegration(t *testing.T) {
	t.Run("complete error processing workflow", func(t *testing.T) {
		// Create a complex error scenario
		issues := []ZodIssue{
			{
				ZodIssueBase: ZodIssueBase{
					Code:    core.UnrecognizedKeys,
					Message: "Unknown field 'extraField'",
					Path:    []any{},
				},
				Keys: []string{"extraField"},
			},
			{
				ZodIssueBase: ZodIssueBase{
					Code:    core.InvalidType,
					Message: "Expected string, received number",
					Path:    []any{"user", "name"},
				},
				Expected: "string",
				Received: "number",
			},
			{
				ZodIssueBase: ZodIssueBase{
					Code:    core.TooSmall,
					Message: "Must be at least 18",
					Path:    []any{"user", "age"},
				},
				Minimum: 18,
			},
			{
				ZodIssueBase: ZodIssueBase{
					Code:    core.InvalidFormat,
					Message: "Invalid email format",
					Path:    []any{"user", "contacts", 0, "email"},
				},
				Format: "email",
			},
		}

		zodErr := NewZodError(issues)

		// Test all formatting methods work together
		formatted := FormatError(zodErr)
		tree := TreeifyError(zodErr)
		flattened := FlattenError(zodErr)
		prettified := PrettifyError(zodErr)

		// Verify formatted structure
		require.Contains(t, formatted["_errors"].([]string), "Unknown field 'extraField'")

		userFormatted := formatted["user"].(ZodFormattedError)
		nameFormatted := userFormatted["name"].(ZodFormattedError)
		assert.Contains(t, nameFormatted["_errors"].([]string), "Expected string, received number")

		// Verify tree structure
		assert.Contains(t, tree.Errors, "Unknown field 'extraField'")
		userProp := tree.Properties["user"]
		assert.Contains(t, userProp.Properties["name"].Errors, "Expected string, received number")
		assert.Contains(t, userProp.Properties["age"].Errors, "Must be at least 18")

		// Verify flattened structure
		assert.Contains(t, flattened.FormErrors, "Unknown field 'extraField'")
		assert.Contains(t, flattened.FieldErrors["user"], "Expected string, received number")
		assert.Contains(t, flattened.FieldErrors["user"], "Must be at least 18")
		assert.Contains(t, flattened.FieldErrors["user"], "Invalid email format")

		// Verify prettified output
		assert.Contains(t, prettified, "Unknown field 'extraField'")
		assert.Contains(t, prettified, "Expected string, received number")
		assert.Contains(t, prettified, "user.name: Expected string, received number")
		assert.Contains(t, prettified, "user.contacts[0].email: Invalid email format")
	})

	t.Run("error identification and conversion workflow", func(t *testing.T) {
		// Create ZodError
		zodErr := NewZodError([]ZodIssue{
			{
				ZodIssueBase: ZodIssueBase{
					Code:    core.InvalidType,
					Message: "Test error",
					Path:    []any{"field"},
				},
			},
		})

		// Wrap it in a regular error
		wrappedErr := fmt.Errorf("validation failed: %w", zodErr)

		// Test identification workflow
		var extractedErr *ZodError
		isZod := IsZodError(wrappedErr, &extractedErr)

		require.True(t, isZod)
		require.NotNil(t, extractedErr)
		assert.Equal(t, zodErr, extractedErr)

		// Process the extracted error
		prettified := PrettifyError(extractedErr)
		assert.Contains(t, prettified, "Test error")
		assert.Contains(t, prettified, "field: Test error")
	})

	t.Run("custom mapper integration", func(t *testing.T) {
		issues := []ZodIssue{
			{
				ZodIssueBase: ZodIssueBase{
					Code:    core.InvalidType,
					Message: "Type error",
					Path:    []any{"field1"},
				},
				Expected: "string",
				Received: "number",
			},
			{
				ZodIssueBase: ZodIssueBase{
					Code:    core.TooSmall,
					Message: "Size error",
					Path:    []any{"field2"},
				},
				Minimum: 5,
			},
		}

		err := NewZodError(issues)

		// Create a comprehensive mapper
		detailedMapper := func(issue ZodIssue) string {
			switch issue.Code {
			case core.InvalidType:
				return fmt.Sprintf("TYPE_ERROR: Expected '%s' but received '%s' at %s",
					issue.Expected, issue.Received, ToDotPath(issue.Path))
			case core.TooSmall:
				return fmt.Sprintf("SIZE_ERROR: Value must be at least %v at %s",
					issue.Minimum, ToDotPath(issue.Path))
			case core.InvalidValue, core.InvalidFormat, core.InvalidUnion, core.InvalidKey,
				core.InvalidElement, core.TooBig, core.NotMultipleOf, core.UnrecognizedKeys, core.Custom,
				core.InvalidSchema, core.InvalidDiscriminator, core.IncompatibleTypes, core.MissingRequired,
				core.TypeConversion, core.NilPointer:
				return fmt.Sprintf("ERROR: %s at %s",
					issue.Message, ToDotPath(issue.Path))
			default:
				return fmt.Sprintf("UNKNOWN_ERROR: %s at %s",
					issue.Message, ToDotPath(issue.Path))
			}
		}

		// Test all mapper variants work consistently
		formatted := FormatErrorWithMapper(err, detailedMapper)
		tree := TreeifyErrorWithMapper(err, detailedMapper)
		flattened := FlattenErrorWithMapper(err, detailedMapper)

		// Verify consistent custom formatting
		field1Formatted := formatted["field1"].(ZodFormattedError)
		field1Errors := field1Formatted["_errors"].([]string)
		assert.Contains(t, field1Errors[0], "TYPE_ERROR: Expected 'string' but received 'number'")

		field1Tree := tree.Properties["field1"]
		assert.Contains(t, field1Tree.Errors[0], "TYPE_ERROR: Expected 'string' but received 'number'")

		field1Flattened := flattened.FieldErrors["field1"]
		assert.Contains(t, field1Flattened[0], "TYPE_ERROR: Expected 'string' but received 'number'")
	})
}

// TestComplexIssueTypes moved to errors_test.go to avoid duplication

func TestPerformanceIntegration(t *testing.T) {
	t.Run("raw issue creation is efficient", func(t *testing.T) {
		// This test ensures issue creation doesn't have performance regressions
		for i := 0; i < 1000; i++ {
			issue := NewRawIssue(core.InvalidType, i,
				WithExpected("string"),
				WithReceived("number"),
			)
			assert.Equal(t, core.InvalidType, issue.Code)
		}
	})

	t.Run("issue finalization is efficient", func(t *testing.T) {
		config := &core.ZodConfig{}

		// Test finalization performance
		for i := 0; i < 1000; i++ {
			rawIssue := NewRawIssue(core.TooSmall, i,
				WithMinimum(100),
				WithOrigin("number"),
			)

			issue := FinalizeIssue(rawIssue, nil, config)
			assert.Equal(t, core.TooSmall, issue.Code)
		}
	})

	t.Run("error formatting is efficient", func(t *testing.T) {
		// Create a complex error with multiple issues
		issues := []ZodIssue{}
		for i := 0; i < 50; i++ {
			issues = append(issues, ZodIssue{
				ZodIssueBase: ZodIssueBase{
					Code:    core.InvalidType,
					Message: fmt.Sprintf("Error %d", i),
					Path:    []any{fmt.Sprintf("field%d", i)},
				},
			})
		}

		zodErr := NewZodError(issues)

		// Test that formatting operations are efficient
		for i := 0; i < 10; i++ {
			formatted := FormatError(zodErr)
			tree := TreeifyError(zodErr)
			flattened := FlattenError(zodErr)
			prettified := PrettifyError(zodErr)

			assert.NotNil(t, formatted)
			assert.NotNil(t, tree)
			assert.NotNil(t, flattened)
			assert.NotEmpty(t, prettified)
		}
	})

	t.Run("API validation error response simulation", func(t *testing.T) {
		// Simulate typical API validation error response
		rawIssue1 := NewRawIssue(core.InvalidType, "john123", WithExpected("string"), WithReceived("number"), WithPath([]any{"username"}))

		rawIssue2 := NewRawIssue(core.TooSmall, "ab", WithMinimum(3), WithInclusive(true), WithOrigin("string"), WithPath([]any{"password"}))

		rawIssue3 := NewRawIssue(core.InvalidFormat, "invalid.email", WithFormat("email"), WithPath([]any{"email"}))

		// Convert raw issues to finalized issues
		issues := []core.ZodIssue{
			FinalizeIssue(rawIssue1, nil, nil),
			FinalizeIssue(rawIssue2, nil, nil),
			FinalizeIssue(rawIssue3, nil, nil),
		}

		zodErr := NewZodError(issues)

		// Create API-friendly error format
		apiMapper := func(issue core.ZodIssue) string {
			switch issue.Code {
			case core.InvalidType:
				return fmt.Sprintf("Field must be of type %s", issue.Expected)
			case core.TooSmall:
				if issue.Inclusive {
					return fmt.Sprintf("Field must be at least %v characters", issue.Minimum)
				}
				return fmt.Sprintf("Field must be more than %v characters", issue.Minimum)
			case core.InvalidFormat:
				return fmt.Sprintf("Field must be a valid %s", issue.Format)
			case core.InvalidValue, core.InvalidUnion, core.InvalidKey,
				core.InvalidElement, core.TooBig, core.NotMultipleOf, core.UnrecognizedKeys, core.Custom,
				core.InvalidSchema, core.InvalidDiscriminator, core.IncompatibleTypes, core.MissingRequired,
				core.TypeConversion, core.NilPointer:
				return issue.Message
			default:
				return issue.Message
			}
		}

		apiErrors := FlattenErrorWithMapper(zodErr, apiMapper)

		// Verify API response format
		require.Contains(t, apiErrors.FieldErrors, "username")
		assert.Equal(t, "Field must be of type string", apiErrors.FieldErrors["username"][0])

		require.Contains(t, apiErrors.FieldErrors, "password")
		assert.Equal(t, "Field must be at least 3 characters", apiErrors.FieldErrors["password"][0])

		require.Contains(t, apiErrors.FieldErrors, "email")
		assert.Equal(t, "Field must be a valid email", apiErrors.FieldErrors["email"][0])
	})

	t.Run("development vs production error handling", func(t *testing.T) {
		rawIssue := NewRawIssue(core.InvalidType, "test", WithExpected("string"), WithReceived("number"), WithPath([]any{"data", "field"}))

		// Development mode - include input
		devCtx := &core.ParseContext{ReportInput: true}
		devIssue := FinalizeIssue(rawIssue, devCtx, nil)
		assert.Equal(t, "test", devIssue.Input)

		// Production mode - exclude sensitive input
		prodCtx := &core.ParseContext{ReportInput: false}
		prodIssue := FinalizeIssue(rawIssue, prodCtx, nil)
		assert.Nil(t, prodIssue.Input)

		// Both should have same validation logic
		assert.Equal(t, devIssue.Code, prodIssue.Code)
		assert.Equal(t, devIssue.Path, prodIssue.Path)
		assert.Equal(t, devIssue.Expected, prodIssue.Expected)
		assert.Equal(t, devIssue.Received, prodIssue.Received)
	})
}

func TestEdgeCasesIntegration(t *testing.T) {
	t.Run("handles missing properties gracefully", func(t *testing.T) {
		issue := ZodRawIssue{
			Code:       "test",
			Properties: nil,
		}

		// Should not panic
		assert.NotPanics(t, func() {
			_ = issue.Expected()
			_ = issue.Minimum()
			_ = issue.Inclusive()
		})
	})

	t.Run("handles wrong property types gracefully", func(t *testing.T) {
		issue := ZodRawIssue{
			Code: "test",
			Properties: map[string]any{
				"expected":  123,     // Wrong type (should be string)
				"minimum":   "hello", // Wrong type (should be number)
				"inclusive": "yes",   // Wrong type (should be bool)
			},
		}

		// Should return empty/default values for wrong types
		assert.Empty(t, issue.Expected())
		assert.Equal(t, "hello", issue.Minimum()) // Returns any, so preserves wrong type
		assert.False(t, issue.Inclusive())        // Wrong type returns false
	})

	t.Run("finalization with nil config", func(t *testing.T) {
		rawIssue := NewRawIssue(core.Custom, nil)

		// Should not panic with nil config
		assert.NotPanics(t, func() {
			issue := FinalizeIssue(rawIssue, nil, nil)
			assert.Equal(t, core.Custom, issue.Code)
		})
	})

	t.Run("finalization with nil context", func(t *testing.T) {
		rawIssue := NewRawIssue(core.Custom, "test_input")
		config := &core.ZodConfig{}

		issue := FinalizeIssue(rawIssue, nil, config)

		assert.Equal(t, core.Custom, issue.Code)
		// When ctx is nil, input is included by default (ReportInput defaults to true)
		assert.Equal(t, "test_input", issue.Input)
	})

	t.Run("property mapping handles all field types", func(t *testing.T) {
		props := map[string]any{
			"expected":  "string",
			"received":  "number",
			"minimum":   5,
			"maximum":   100,
			"inclusive": true,
			"format":    "email",
			"pattern":   ".*@.*",
			"prefix":    "test_",
			"suffix":    "_end",
			"includes":  "@",
			"algorithm": "HS256",
			"divisor":   2.5,
			"keys":      []string{"key1", "key2"},
			"values":    []any{"val1", "val2"},
			"origin":    "string",
			"key":       "field_name",
			"params":    map[string]any{"custom": "data"},
		}

		issue := ZodIssue{}

		// Should not panic
		assert.NotPanics(t, func() {
			MapPropertiesToIssue(&issue, props)
		})

		// Verify all properties were mapped correctly
		assert.Equal(t, core.ZodTypeString, issue.Expected)
		assert.Equal(t, core.ZodTypeNumber, issue.Received)
		assert.Equal(t, 5, issue.Minimum)
		assert.Equal(t, 100, issue.Maximum)
		assert.True(t, issue.Inclusive)
		assert.Equal(t, "email", issue.Format)
		assert.Equal(t, ".*@.*", issue.Pattern)
		assert.Equal(t, "test_", issue.Prefix)
		assert.Equal(t, "_end", issue.Suffix)
		assert.Equal(t, "@", issue.Includes)
		assert.Equal(t, "HS256", issue.Algorithm)
		assert.Equal(t, 2.5, issue.Divisor)
		assert.Equal(t, []string{"key1", "key2"}, issue.Keys)
		assert.Equal(t, []any{"val1", "val2"}, issue.Values)
		assert.Equal(t, "string", issue.Origin)
		assert.Equal(t, "field_name", issue.Key)
		assert.Equal(t, map[string]any{"custom": "data"}, issue.Params)
	})

	t.Run("property mapping with nil properties", func(t *testing.T) {
		issue := ZodIssue{}

		// Should not panic
		assert.NotPanics(t, func() {
			MapPropertiesToIssue(&issue, nil)
		})

		// Should remain unchanged
		assert.Empty(t, issue.Expected)
		assert.Nil(t, issue.Minimum)
	})

	t.Run("property mapping with wrong types", func(t *testing.T) {
		props := map[string]any{
			"expected":  123,                      // Wrong type
			"keys":      "not_a_slice",            // Wrong type
			"inclusive": "not_a_bool",             // Wrong type
			"params":    []string{"not", "a_map"}, // Wrong type
		}

		issue := ZodIssue{}

		// Should not panic with wrong types
		assert.NotPanics(t, func() {
			MapPropertiesToIssue(&issue, props)
		})

		// Wrong types should not be assigned
		assert.Empty(t, issue.Expected)  // Wrong type, should remain empty
		assert.Nil(t, issue.Keys)        // Wrong type, should remain nil
		assert.False(t, issue.Inclusive) // Wrong type, should remain false
		assert.Nil(t, issue.Params)      // Wrong type, should remain nil
	})
}

func TestRealWorldScenarios(t *testing.T) {
	t.Run("complex nested validation scenario", func(t *testing.T) {
		// Simulate a complex form validation with multiple error types
		issues := []ZodIssue{
			{
				ZodIssueBase: ZodIssueBase{
					Code:    core.UnrecognizedKeys,
					Message: "Unknown field in root object",
					Path:    []any{},
				},
				Keys: []string{"unknownField"},
			},
			{
				ZodIssueBase: ZodIssueBase{
					Code:    core.InvalidType,
					Message: "Expected string",
					Path:    []any{"user", "firstName"},
				},
				Expected: "string",
				Received: "number",
			},
			{
				ZodIssueBase: ZodIssueBase{
					Code:    core.TooSmall,
					Message: "Age must be at least 18",
					Path:    []any{"user", "age"},
				},
				Minimum:   18,
				Inclusive: true,
			},
			{
				ZodIssueBase: ZodIssueBase{
					Code:    core.InvalidFormat,
					Message: "Invalid email format",
					Path:    []any{"user", "contacts", 0, "email"},
				},
				Format: "email",
			},
			{
				ZodIssueBase: ZodIssueBase{
					Code:    core.TooSmall,
					Message: "Phone number too short",
					Path:    []any{"user", "contacts", 0, "phone"},
				},
				Minimum:   10,
				Inclusive: true,
			},
			{
				ZodIssueBase: ZodIssueBase{
					Code:    core.InvalidType,
					Message: "Expected array",
					Path:    []any{"user", "tags"},
				},
				Expected: "array",
				Received: "string",
			},
		}

		zodErr := NewZodError(issues)

		// Test comprehensive error formatting
		formatted := FormatError(zodErr)
		tree := TreeifyError(zodErr)
		flattened := FlattenError(zodErr)
		prettified := PrettifyError(zodErr)

		// Verify structured access to errors
		assert.NotNil(t, formatted)
		assert.NotNil(t, tree)
		assert.NotNil(t, flattened)
		assert.NotEmpty(t, prettified)

		// Test specific error location access
		userTree := tree.Properties["user"]
		require.NotNil(t, userTree)

		firstNameErrors := userTree.Properties["firstName"].Errors
		assert.Contains(t, firstNameErrors, "Expected string")

		contactsArray := userTree.Properties["contacts"]
		require.NotNil(t, contactsArray)
		require.NotNil(t, contactsArray.Items)
		require.Len(t, contactsArray.Items, 1)

		contact0 := contactsArray.Items[0]
		assert.Contains(t, contact0.Properties["email"].Errors, "Invalid email format")
		assert.Contains(t, contact0.Properties["phone"].Errors, "Phone number too short")
	})

	t.Run("API validation error response simulation", func(t *testing.T) {
		// Simulate typical API validation error response
		rawIssue1 := NewRawIssue(core.InvalidType, "john123", WithExpected("string"), WithReceived("number"), WithPath([]any{"username"}))

		rawIssue2 := NewRawIssue(core.TooSmall, "ab", WithMinimum(3), WithInclusive(true), WithOrigin("string"), WithPath([]any{"password"}))

		rawIssue3 := NewRawIssue(core.InvalidFormat, "invalid.email", WithFormat("email"), WithPath([]any{"email"}))

		// Convert raw issues to finalized issues
		issues := []core.ZodIssue{
			FinalizeIssue(rawIssue1, nil, nil),
			FinalizeIssue(rawIssue2, nil, nil),
			FinalizeIssue(rawIssue3, nil, nil),
		}

		zodErr := NewZodError(issues)

		// Create API-friendly error format
		apiMapper := func(issue core.ZodIssue) string {
			switch issue.Code {
			case core.InvalidType:
				return fmt.Sprintf("Field must be of type %s", issue.Expected)
			case core.TooSmall:
				if issue.Inclusive {
					return fmt.Sprintf("Field must be at least %v characters", issue.Minimum)
				}
				return fmt.Sprintf("Field must be more than %v characters", issue.Minimum)
			case core.InvalidFormat:
				return fmt.Sprintf("Field must be a valid %s", issue.Format)
			case core.InvalidValue, core.InvalidUnion, core.InvalidKey,
				core.InvalidElement, core.TooBig, core.NotMultipleOf, core.UnrecognizedKeys, core.Custom,
				core.InvalidSchema, core.InvalidDiscriminator, core.IncompatibleTypes, core.MissingRequired,
				core.TypeConversion, core.NilPointer:
				return issue.Message
			default:
				return issue.Message
			}
		}

		apiErrors := FlattenErrorWithMapper(zodErr, apiMapper)

		// Verify API response format
		require.Contains(t, apiErrors.FieldErrors, "username")
		assert.Equal(t, "Field must be of type string", apiErrors.FieldErrors["username"][0])

		require.Contains(t, apiErrors.FieldErrors, "password")
		assert.Equal(t, "Field must be at least 3 characters", apiErrors.FieldErrors["password"][0])

		require.Contains(t, apiErrors.FieldErrors, "email")
		assert.Equal(t, "Field must be a valid email", apiErrors.FieldErrors["email"][0])
	})

	t.Run("development vs production error handling", func(t *testing.T) {
		rawIssue := NewRawIssue(core.InvalidType, "test", WithExpected("string"), WithReceived("number"), WithPath([]any{"data", "field"}))

		// Development mode - include input
		devCtx := &core.ParseContext{ReportInput: true}
		devIssue := FinalizeIssue(rawIssue, devCtx, nil)
		assert.Equal(t, "test", devIssue.Input)

		// Production mode - exclude sensitive input
		prodCtx := &core.ParseContext{ReportInput: false}
		prodIssue := FinalizeIssue(rawIssue, prodCtx, nil)
		assert.Nil(t, prodIssue.Input)

		// Both should have same validation logic
		assert.Equal(t, devIssue.Code, prodIssue.Code)
		assert.Equal(t, devIssue.Path, prodIssue.Path)
		assert.Equal(t, devIssue.Expected, prodIssue.Expected)
		assert.Equal(t, devIssue.Received, prodIssue.Received)
	})
}

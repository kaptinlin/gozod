package issues

import (
	"fmt"
	"strings"
	"testing"

	"github.com/kaptinlin/gozod/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//////////////////////////////////////////
//////////   Complex Issue Type Tests  ///
//////////////////////////////////////////

func TestComplexIssueTypes(t *testing.T) {
	t.Run("handles invalid_union issues", func(t *testing.T) {
		// Create mock union errors - simulate failed union validation
		unionErrors := [][]ZodIssue{
			{
				{
					ZodIssueBase: ZodIssueBase{
						Code:    "invalid_type",
						Message: "Expected string, received number",
						Path:    []any{"field1"},
					},
					Expected: "string",
					Received: "number",
				},
			},
			{
				{
					ZodIssueBase: ZodIssueBase{
						Code:    "too_small",
						Message: "Must be at least 5",
						Path:    []any{"field2"},
					},
					Minimum: 5,
				},
			},
		}

		// Create invalid_union issue
		unionIssue := ZodIssue{
			ZodIssueBase: ZodIssueBase{
				Code:    "invalid_union",
				Message: "Invalid union input",
				Path:    []any{},
			},
			Errors: unionErrors,
		}

		issues := []ZodIssue{unionIssue}
		err := NewZodError(issues)

		// Test that all formatting methods handle union errors correctly
		formatted := FormatError(err)
		tree := TreeifyError(err)
		flattened := FlattenError(err)
		prettified := PrettifyError(err)

		// All should process without panicking
		assert.NotNil(t, formatted)
		assert.NotNil(t, tree)
		assert.NotNil(t, flattened)
		assert.NotEmpty(t, prettified)

		// Verify that union errors are processed correctly
		assert.Contains(t, flattened.FormErrors, "Invalid union input")
		assert.Contains(t, prettified, "Invalid union input")
	})

	t.Run("handles invalid_key issues", func(t *testing.T) {
		// Create nested issues for invalid key validation
		nestedIssues := []ZodIssue{
			{
				ZodIssueBase: ZodIssueBase{
					Code:    "invalid_format",
					Message: "Invalid key format",
					Path:    []any{"keyField"},
				},
				Format: "string",
			},
		}

		// Create invalid_key issue
		keyIssue := ZodIssue{
			ZodIssueBase: ZodIssueBase{
				Code:    "invalid_key",
				Message: "Invalid key in record",
				Path:    []any{"record"},
			},
			Issues: nestedIssues,
			Key:    "invalidKey",
			Origin: "record",
		}

		issues := []ZodIssue{keyIssue}
		err := NewZodError(issues)

		// Test that all formatting methods handle key errors correctly
		formatted := FormatError(err)
		tree := TreeifyError(err)
		flattened := FlattenError(err)
		prettified := PrettifyError(err)

		// All should process without panicking
		assert.NotNil(t, formatted)
		assert.NotNil(t, tree)
		assert.NotNil(t, flattened)
		assert.NotEmpty(t, prettified)

		// Verify that key errors are processed correctly
		assert.Contains(t, flattened.FieldErrors["record"], "Invalid key in record")
		assert.Contains(t, prettified, "Invalid key in record")
	})

	t.Run("handles invalid_element issues", func(t *testing.T) {
		// Create nested issues for invalid element validation
		nestedIssues := []ZodIssue{
			{
				ZodIssueBase: ZodIssueBase{
					Code:    "invalid_type",
					Message: "Expected number, received string",
					Path:    []any{"element"},
				},
				Expected: "number",
				Received: "string",
			},
		}

		// Create invalid_element issue
		elementIssue := ZodIssue{
			ZodIssueBase: ZodIssueBase{
				Code:    "invalid_element",
				Message: "Invalid array element",
				Path:    []any{"array", 0},
			},
			Issues: nestedIssues,
			Key:    0,
			Origin: "array",
		}

		issues := []ZodIssue{elementIssue}
		err := NewZodError(issues)

		// Test that all formatting methods handle element errors correctly
		formatted := FormatError(err)
		tree := TreeifyError(err)
		flattened := FlattenError(err)
		prettified := PrettifyError(err)

		// All should process without panicking
		assert.NotNil(t, formatted)
		assert.NotNil(t, tree)
		assert.NotNil(t, flattened)
		assert.NotEmpty(t, prettified)

		// Verify that element errors are processed correctly
		assert.Contains(t, flattened.FieldErrors["array"], "Invalid array element")
		assert.Contains(t, prettified, "Invalid array element")
		assert.Contains(t, prettified, "array[0]: Invalid array element")
	})

	t.Run("processes standard issue types correctly", func(t *testing.T) {
		issues := []ZodIssue{
			{
				ZodIssueBase: ZodIssueBase{
					Code:    "invalid_type",
					Message: "Type mismatch",
					Path:    []any{"field1"},
				},
			},
			{
				ZodIssueBase: ZodIssueBase{
					Code:    "too_small",
					Message: "Value too small",
					Path:    []any{"field2"},
				},
			},
			{
				ZodIssueBase: ZodIssueBase{
					Code:    "custom",
					Message: "Custom validation failed",
					Path:    []any{"field3"},
				},
			},
		}

		err := NewZodError(issues)

		// Test that all standard formats handle these correctly
		formatted := FormatError(err)
		tree := TreeifyError(err)
		flattened := FlattenError(err)
		prettified := PrettifyError(err)

		// All should process without panicking
		assert.NotNil(t, formatted)
		assert.NotNil(t, tree)
		assert.NotNil(t, flattened)
		assert.NotEmpty(t, prettified)

		// Verify content is preserved
		assert.Contains(t, flattened.FieldErrors["field1"], "Type mismatch")
		assert.Contains(t, flattened.FieldErrors["field2"], "Value too small")
		assert.Contains(t, flattened.FieldErrors["field3"], "Custom validation failed")
	})
}

//////////////////////////////////////////
//////////   Error Integration Tests   ///
//////////////////////////////////////////

func TestErrorIntegration(t *testing.T) {
	t.Run("complete error processing workflow", func(t *testing.T) {
		// Create a complex error scenario
		issues := []ZodIssue{
			{
				ZodIssueBase: ZodIssueBase{
					Code:    "unrecognized_keys",
					Message: "Unknown field 'extraField'",
					Path:    []any{},
				},
				Keys: []string{"extraField"},
			},
			{
				ZodIssueBase: ZodIssueBase{
					Code:    "invalid_type",
					Message: "Expected string, received number",
					Path:    []any{"user", "name"},
				},
				Expected: "string",
				Received: "number",
			},
			{
				ZodIssueBase: ZodIssueBase{
					Code:    "too_small",
					Message: "Must be at least 18",
					Path:    []any{"user", "age"},
				},
				Minimum: 18,
			},
			{
				ZodIssueBase: ZodIssueBase{
					Code:    "invalid_format",
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
					Code:    "invalid_type",
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
					Code:    "invalid_type",
					Message: "Type error",
					Path:    []any{"field1"},
				},
				Expected: "string",
				Received: "number",
			},
			{
				ZodIssueBase: ZodIssueBase{
					Code:    "too_small",
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

	t.Run("comprehensive error handling workflow", func(t *testing.T) {
		// Test the complete end-to-end error processing workflow

		// Step 1: Create various types of errors
		issues := []ZodIssue{
			// Root level error
			{
				ZodIssueBase: ZodIssueBase{
					Code:    "custom",
					Message: "General validation failed",
					Path:    []any{},
				},
			},
			// Simple field error
			{
				ZodIssueBase: ZodIssueBase{
					Code:    "invalid_type",
					Message: "Expected string",
					Path:    []any{"name"},
				},
				Expected: "string",
				Received: "number",
			},
			// Nested object error
			{
				ZodIssueBase: ZodIssueBase{
					Code:    "too_small",
					Message: "Age must be at least 18",
					Path:    []any{"profile", "age"},
				},
				Minimum: 18,
			},
			// Array element error
			{
				ZodIssueBase: ZodIssueBase{
					Code:    "invalid_format",
					Message: "Invalid email format",
					Path:    []any{"contacts", 0, "email"},
				},
				Format: "email",
			},
			// Deep nested error
			{
				ZodIssueBase: ZodIssueBase{
					Code:    "not_multiple_of",
					Message: "Must be multiple of 5",
					Path:    []any{"settings", "preferences", "theme", "spacing"},
				},
				Divisor: 5,
			},
		}

		// Step 2: Create ZodError
		zodErr := NewZodError(issues)

		// Step 3: Test error identification
		var extractedErr *ZodError
		isZod := IsZodError(zodErr, &extractedErr)
		require.True(t, isZod)
		assert.Equal(t, zodErr, extractedErr)

		// Step 4: Test all formatting methods
		formatted := FormatError(zodErr)
		tree := TreeifyError(zodErr)
		flattened := FlattenError(zodErr)
		prettified := PrettifyError(zodErr)

		// Step 5: Verify all outputs are consistent and complete
		require.NotNil(t, formatted)
		require.NotNil(t, tree)
		require.NotNil(t, flattened)
		require.NotEmpty(t, prettified)

		// Step 6: Verify content integrity across all formats
		// Root level errors
		assert.Contains(t, formatted["_errors"].([]string), "General validation failed")
		assert.Contains(t, tree.Errors, "General validation failed")
		assert.Contains(t, flattened.FormErrors, "General validation failed")
		assert.Contains(t, prettified, "General validation failed")

		// Field level errors
		assert.Contains(t, flattened.FieldErrors["name"], "Expected string")
		assert.Contains(t, flattened.FieldErrors["profile"], "Age must be at least 18")
		assert.Contains(t, flattened.FieldErrors["contacts"], "Invalid email format")
		assert.Contains(t, flattened.FieldErrors["settings"], "Must be multiple of 5")

		// Path formatting in prettified output
		assert.Contains(t, prettified, "name: Expected string")
		assert.Contains(t, prettified, "profile.age: Age must be at least 18")
		assert.Contains(t, prettified, "contacts[0].email: Invalid email format")
		assert.Contains(t, prettified, "settings.preferences.theme.spacing: Must be multiple of 5")

		// Step 7: Test custom mapping integration
		customMapper := func(issue ZodIssue) string {
			return fmt.Sprintf("[%s] %s", strings.ToUpper(string(issue.Code)), issue.Message)
		}

		customFormatted := FormatErrorWithMapper(zodErr, customMapper)
		customTree := TreeifyErrorWithMapper(zodErr, customMapper)
		customFlattened := FlattenErrorWithMapper(zodErr, customMapper)

		// Verify custom formatting is applied
		nameField := customFormatted["name"].(ZodFormattedError)
		nameErrors := nameField["_errors"].([]string)
		assert.Contains(t, nameErrors[0], "[INVALID_TYPE] Expected string")

		// Verify tree formatting
		nameTreeProp := customTree.Properties["name"]
		assert.Contains(t, nameTreeProp.Errors[0], "[INVALID_TYPE] Expected string")

		// Verify flattened formatting
		nameFlattened := customFlattened.FieldErrors["name"]
		assert.Contains(t, nameFlattened[0], "[INVALID_TYPE] Expected string")

		// Step 8: Test wrapped error handling
		wrappedErr := fmt.Errorf("validation process failed: %w", zodErr)
		var wrappedExtracted *ZodError
		isWrappedZod := IsZodError(wrappedErr, &wrappedExtracted)
		require.True(t, isWrappedZod)
		assert.Equal(t, zodErr, wrappedExtracted)
	})
}

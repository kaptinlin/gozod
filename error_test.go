package gozod

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//////////////////////////////////////////
//////////   ZodError Creation Tests   ///
//////////////////////////////////////////

func TestZodErrorCreation(t *testing.T) {
	t.Run("creates error with basic properties", func(t *testing.T) {
		issues := []ZodIssue{
			{
				ZodIssueBase: ZodIssueBase{
					Code:    "invalid_type",
					Message: "Invalid input: expected string, received number",
					Path:    []interface{}{"name"},
				},
				Expected: "string",
				Received: "number",
			},
		}

		err := NewZodError(issues)

		require.NotNil(t, err)
		require.Equal(t, "ZodError", err.Name)
		require.Len(t, err.Issues, 1)
		require.Len(t, err.Zod.Def, 1)
		assert.Equal(t, issues, err.Issues)
		assert.Equal(t, issues, err.Zod.Def)
		assert.Nil(t, err.Type)
		assert.Nil(t, err.Zod.Output)
	})

	t.Run("creates error with multiple issues", func(t *testing.T) {
		issues := []ZodIssue{
			{
				ZodIssueBase: ZodIssueBase{
					Code:    "invalid_type",
					Message: "Invalid type",
					Path:    []interface{}{"name"},
				},
			},
			{
				ZodIssueBase: ZodIssueBase{
					Code:    "too_small",
					Message: "Too small",
					Path:    []interface{}{"age"},
				},
			},
			{
				ZodIssueBase: ZodIssueBase{
					Code:    "invalid_format",
					Message: "Invalid email",
					Path:    []interface{}{"email"},
				},
			},
		}

		err := NewZodError(issues)

		require.Len(t, err.Issues, 3)
		assert.Equal(t, "invalid_type", err.Issues[0].Code)
		assert.Equal(t, "too_small", err.Issues[1].Code)
		assert.Equal(t, "invalid_format", err.Issues[2].Code)
	})

	t.Run("creates error with empty issues slice", func(t *testing.T) {
		issues := []ZodIssue{}
		err := NewZodError(issues)

		require.NotNil(t, err)
		assert.Empty(t, err.Issues)
		assert.Empty(t, err.Zod.Def)
	})

	t.Run("creates error with nil issues slice", func(t *testing.T) {
		err := NewZodError(nil)

		require.NotNil(t, err)
		assert.Nil(t, err.Issues)
		assert.Nil(t, err.Zod.Def)
	})
}

//////////////////////////////////////////
//////////   ZodError Interface Tests  ///
//////////////////////////////////////////

func TestZodErrorInterface(t *testing.T) {
	t.Run("implements error interface correctly", func(t *testing.T) {
		issues := []ZodIssue{
			{
				ZodIssueBase: ZodIssueBase{
					Code:    "invalid_type",
					Message: "Invalid input: expected string, received number",
					Path:    []interface{}{"user", "name"},
				},
			},
		}

		zodErr := NewZodError(issues)
		var err error = zodErr

		require.NotNil(t, err)
		errorStr := err.Error()
		assert.Contains(t, errorStr, "✖")
		assert.Contains(t, errorStr, "Invalid input: expected string, received number")
		assert.Contains(t, errorStr, "→ at user.name")
	})

	t.Run("error string contains all issues", func(t *testing.T) {
		issues := []ZodIssue{
			{
				ZodIssueBase: ZodIssueBase{
					Code:    "invalid_type",
					Message: "First error",
					Path:    []interface{}{"field1"},
				},
			},
			{
				ZodIssueBase: ZodIssueBase{
					Code:    "too_small",
					Message: "Second error",
					Path:    []interface{}{"field2"},
				},
			},
		}

		err := NewZodError(issues)
		errorStr := err.Error()

		assert.Contains(t, errorStr, "First error")
		assert.Contains(t, errorStr, "Second error")
		assert.Contains(t, errorStr, "field1")
		assert.Contains(t, errorStr, "field2")
	})
}

//////////////////////////////////////////
//////////   IsZodError Function Tests ///
//////////////////////////////////////////

func TestIsZodError(t *testing.T) {
	t.Run("identifies ZodError correctly", func(t *testing.T) {
		zodErr := NewZodError([]ZodIssue{
			{ZodIssueBase: ZodIssueBase{Code: "test", Message: "test"}},
		})

		var target *ZodError
		result := IsZodError(zodErr, &target)

		require.True(t, result)
		require.NotNil(t, target)
		assert.Equal(t, zodErr, target)
	})

	t.Run("returns false for non-ZodError", func(t *testing.T) {
		//nolint:err113 // Intentional regular error for testing IsZodError function
		regularErr := fmt.Errorf("regular error")

		var target *ZodError
		result := IsZodError(regularErr, &target)

		assert.False(t, result)
		assert.Nil(t, target)
	})

	t.Run("returns false for nil error", func(t *testing.T) {
		var target *ZodError
		result := IsZodError(nil, &target)

		assert.False(t, result)
		assert.Nil(t, target)
	})

	t.Run("works without target parameter", func(t *testing.T) {
		zodErr := NewZodError([]ZodIssue{
			{ZodIssueBase: ZodIssueBase{Code: "test", Message: "test"}},
		})

		result := IsZodError(zodErr, nil)
		assert.True(t, result)

		//nolint:err113 // Intentional regular error for testing IsZodError function
		regularErr := fmt.Errorf("regular error")
		result = IsZodError(regularErr, nil)
		assert.False(t, result)
	})

	t.Run("handles wrapped errors", func(t *testing.T) {
		zodErr := NewZodError([]ZodIssue{
			{ZodIssueBase: ZodIssueBase{Code: "test", Message: "test"}},
		})
		wrappedErr := fmt.Errorf("wrapped: %w", zodErr)

		var target *ZodError
		result := IsZodError(wrappedErr, &target)

		require.True(t, result)
		require.NotNil(t, target)
		assert.Equal(t, zodErr, target)
	})
}

//////////////////////////////////////////
//////////   FormatError Tests         ///
//////////////////////////////////////////

func TestFormatError(t *testing.T) {
	t.Run("formats simple field errors", func(t *testing.T) {
		issues := []ZodIssue{
			{
				ZodIssueBase: ZodIssueBase{
					Code:    "invalid_type",
					Message: "Expected string",
					Path:    []interface{}{"username"},
				},
			},
			{
				ZodIssueBase: ZodIssueBase{
					Code:    "too_small",
					Message: "Too short",
					Path:    []interface{}{"password"},
				},
			},
		}

		err := NewZodError(issues)
		formatted := FormatError(err)

		require.NotNil(t, formatted)

		// Check root _errors exists
		rootErrors, exists := formatted["_errors"]
		require.True(t, exists)
		assert.Empty(t, rootErrors.([]string))

		// Check field errors
		usernameField, exists := formatted["username"]
		require.True(t, exists)
		usernameFormatted := usernameField.(ZodFormattedError)
		usernameErrors := usernameFormatted["_errors"].([]string)
		assert.Contains(t, usernameErrors, "Expected string")

		passwordField, exists := formatted["password"]
		require.True(t, exists)
		passwordFormatted := passwordField.(ZodFormattedError)
		passwordErrors := passwordFormatted["_errors"].([]string)
		assert.Contains(t, passwordErrors, "Too short")
	})

	t.Run("formats nested path errors", func(t *testing.T) {
		issues := []ZodIssue{
			{
				ZodIssueBase: ZodIssueBase{
					Code:    "invalid_type",
					Message: "Invalid nested field",
					Path:    []interface{}{"user", "profile", "email"},
				},
			},
		}

		err := NewZodError(issues)
		formatted := FormatError(err)

		// Navigate to nested structure
		userField, exists := formatted["user"]
		require.True(t, exists)
		userFormatted := userField.(ZodFormattedError)

		profileField, exists := userFormatted["profile"]
		require.True(t, exists)
		profileFormatted := profileField.(ZodFormattedError)

		emailField, exists := profileFormatted["email"]
		require.True(t, exists)
		emailFormatted := emailField.(ZodFormattedError)
		emailErrors := emailFormatted["_errors"].([]string)
		assert.Contains(t, emailErrors, "Invalid nested field")
	})

	t.Run("formats root level errors", func(t *testing.T) {
		issues := []ZodIssue{
			{
				ZodIssueBase: ZodIssueBase{
					Code:    "custom",
					Message: "Root level error",
					Path:    []interface{}{},
				},
			},
		}

		err := NewZodError(issues)
		formatted := FormatError(err)

		rootErrors := formatted["_errors"].([]string)
		assert.Contains(t, rootErrors, "Root level error")
	})

	t.Run("handles empty error", func(t *testing.T) {
		err := NewZodError([]ZodIssue{})
		formatted := FormatError(err)

		require.NotNil(t, formatted)
		rootErrors := formatted["_errors"].([]string)
		assert.Empty(t, rootErrors)
	})
}

//////////////////////////////////////////
//////////   TreeifyError Tests        ///
//////////////////////////////////////////

func TestTreeifyError(t *testing.T) {
	t.Run("creates tree with field properties", func(t *testing.T) {
		issues := []ZodIssue{
			{
				ZodIssueBase: ZodIssueBase{
					Code:    "invalid_type",
					Message: "Invalid name",
					Path:    []interface{}{"user", "name"},
				},
			},
			{
				ZodIssueBase: ZodIssueBase{
					Code:    "too_small",
					Message: "Age too small",
					Path:    []interface{}{"user", "age"},
				},
			},
		}

		err := NewZodError(issues)
		tree := TreeifyError(err)

		require.NotNil(t, tree)
		assert.Empty(t, tree.Errors)
		require.NotNil(t, tree.Properties)

		userProp, exists := tree.Properties["user"]
		require.True(t, exists)
		require.NotNil(t, userProp.Properties)

		nameProp, exists := userProp.Properties["name"]
		require.True(t, exists)
		assert.Contains(t, nameProp.Errors, "Invalid name")

		ageProp, exists := userProp.Properties["age"]
		require.True(t, exists)
		assert.Contains(t, ageProp.Errors, "Age too small")
	})

	t.Run("creates tree with array items", func(t *testing.T) {
		issues := []ZodIssue{
			{
				ZodIssueBase: ZodIssueBase{
					Code:    "invalid_type",
					Message: "Invalid item 0",
					Path:    []interface{}{"items", 0},
				},
			},
			{
				ZodIssueBase: ZodIssueBase{
					Code:    "invalid_type",
					Message: "Invalid item 2",
					Path:    []interface{}{"items", 2},
				},
			},
		}

		err := NewZodError(issues)
		tree := TreeifyError(err)

		itemsProp, exists := tree.Properties["items"]
		require.True(t, exists)
		require.NotNil(t, itemsProp.Items)
		require.Len(t, itemsProp.Items, 3) // Should create items 0, 1, 2

		assert.Contains(t, itemsProp.Items[0].Errors, "Invalid item 0")
		assert.Empty(t, itemsProp.Items[1].Errors) // Item 1 should be empty
		assert.Contains(t, itemsProp.Items[2].Errors, "Invalid item 2")
	})

	t.Run("handles root level errors", func(t *testing.T) {
		issues := []ZodIssue{
			{
				ZodIssueBase: ZodIssueBase{
					Code:    "custom",
					Message: "Root error",
					Path:    []interface{}{},
				},
			},
		}

		err := NewZodError(issues)
		tree := TreeifyError(err)

		assert.Contains(t, tree.Errors, "Root error")
	})

	t.Run("handles mixed path types", func(t *testing.T) {
		issues := []ZodIssue{
			{
				ZodIssueBase: ZodIssueBase{
					Code:    "invalid_type",
					Message: "Mixed path error",
					Path:    []interface{}{"users", 0, "profile", "settings", 1, "value"},
				},
			},
		}

		err := NewZodError(issues)
		tree := TreeifyError(err)

		// Navigate through mixed string/int path
		usersProp := tree.Properties["users"]
		require.NotNil(t, usersProp)

		user0 := usersProp.Items[0]
		require.NotNil(t, user0)

		profileProp := user0.Properties["profile"]
		require.NotNil(t, profileProp)

		settingsProp := profileProp.Properties["settings"]
		require.NotNil(t, settingsProp)

		setting1 := settingsProp.Items[1]
		require.NotNil(t, setting1)

		valueProp := setting1.Properties["value"]
		require.NotNil(t, valueProp)

		assert.Contains(t, valueProp.Errors, "Mixed path error")
	})
}

//////////////////////////////////////////
//////////   FlattenError Tests        ///
//////////////////////////////////////////

func TestFlattenError(t *testing.T) {
	t.Run("flattens field errors correctly", func(t *testing.T) {
		issues := []ZodIssue{
			{
				ZodIssueBase: ZodIssueBase{
					Code:    "invalid_type",
					Message: "Username invalid",
					Path:    []interface{}{"username"},
				},
			},
			{
				ZodIssueBase: ZodIssueBase{
					Code:    "too_small",
					Message: "Password too short",
					Path:    []interface{}{"password"},
				},
			},
		}

		err := NewZodError(issues)
		flattened := FlattenError(err)

		require.NotNil(t, flattened)
		assert.Empty(t, flattened.FormErrors)

		require.Contains(t, flattened.FieldErrors, "username")
		assert.Contains(t, flattened.FieldErrors["username"], "Username invalid")

		require.Contains(t, flattened.FieldErrors, "password")
		assert.Contains(t, flattened.FieldErrors["password"], "Password too short")
	})

	t.Run("handles form level errors", func(t *testing.T) {
		issues := []ZodIssue{
			{
				ZodIssueBase: ZodIssueBase{
					Code:    "unrecognized_keys",
					Message: "Unknown fields",
					Path:    []interface{}{},
				},
			},
			{
				ZodIssueBase: ZodIssueBase{
					Code:    "invalid_type",
					Message: "Field error",
					Path:    []interface{}{"field"},
				},
			},
		}

		err := NewZodError(issues)
		flattened := FlattenError(err)

		assert.Contains(t, flattened.FormErrors, "Unknown fields")
		assert.Contains(t, flattened.FieldErrors["field"], "Field error")
	})

	t.Run("uses only first level for nested paths", func(t *testing.T) {
		issues := []ZodIssue{
			{
				ZodIssueBase: ZodIssueBase{
					Code:    "invalid_type",
					Message: "Nested error",
					Path:    []interface{}{"user", "profile", "email"},
				},
			},
			{
				ZodIssueBase: ZodIssueBase{
					Code:    "invalid_type",
					Message: "Array error",
					Path:    []interface{}{"items", 0, "name"},
				},
			},
		}

		err := NewZodError(issues)
		flattened := FlattenError(err)

		// Both errors should be grouped under their first-level keys
		assert.Contains(t, flattened.FieldErrors["user"], "Nested error")
		assert.Contains(t, flattened.FieldErrors["items"], "Array error")
	})

	t.Run("handles empty error", func(t *testing.T) {
		err := NewZodError([]ZodIssue{})
		flattened := FlattenError(err)

		require.NotNil(t, flattened)
		assert.Empty(t, flattened.FormErrors)
		assert.Empty(t, flattened.FieldErrors)
	})
}

//////////////////////////////////////////
//////////   ToDotPath Tests           ///
//////////////////////////////////////////

func TestToDotPath(t *testing.T) {
	t.Run("converts simple paths", func(t *testing.T) {
		testCases := []struct {
			name     string
			path     []interface{}
			expected string
		}{
			{
				name:     "simple field path",
				path:     []interface{}{"user", "name"},
				expected: "user.name",
			},
			{
				name:     "single field",
				path:     []interface{}{"username"},
				expected: "username",
			},
			{
				name:     "empty path",
				path:     []interface{}{},
				expected: "",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := ToDotPath(tc.path)
				assert.Equal(t, tc.expected, result)
			})
		}
	})

	t.Run("converts array indices", func(t *testing.T) {
		testCases := []struct {
			name     string
			path     []interface{}
			expected string
		}{
			{
				name:     "array access",
				path:     []interface{}{"items", 0},
				expected: "items[0]",
			},
			{
				name:     "nested array access",
				path:     []interface{}{"users", 1, "posts", 2, "title"},
				expected: "users[1].posts[2].title",
			},
			{
				name:     "array only",
				path:     []interface{}{0, 1, 2},
				expected: "[0][1][2]",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := ToDotPath(tc.path)
				assert.Equal(t, tc.expected, result)
			})
		}
	})

	t.Run("handles special characters in field names", func(t *testing.T) {
		testCases := []struct {
			name     string
			path     []interface{}
			expected string
		}{
			{
				name:     "field with spaces",
				path:     []interface{}{"user name"},
				expected: `["user name"]`,
			},
			{
				name:     "field with special chars",
				path:     []interface{}{"user-email"},
				expected: `["user-email"]`,
			},
			{
				name:     "mixed valid and invalid names",
				path:     []interface{}{"user", "full name", "first"},
				expected: `user["full name"].first`,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := ToDotPath(tc.path)
				assert.Equal(t, tc.expected, result)
			})
		}
	})

	t.Run("handles non-string non-int types", func(t *testing.T) {
		path := []interface{}{"field", 3.14, true}
		result := ToDotPath(path)

		assert.Contains(t, result, "field")
		assert.Contains(t, result, "3.14")
		assert.Contains(t, result, "true")
	})
}

//////////////////////////////////////////
//////////   PrettifyError Tests       ///
//////////////////////////////////////////

func TestPrettifyError(t *testing.T) {
	t.Run("formats single error correctly", func(t *testing.T) {
		issues := []ZodIssue{
			{
				ZodIssueBase: ZodIssueBase{
					Code:    "invalid_type",
					Message: "Expected string, received number",
					Path:    []interface{}{"username"},
				},
			},
		}

		err := NewZodError(issues)
		prettified := PrettifyError(err)

		assert.Contains(t, prettified, "✖ Expected string, received number")
		assert.Contains(t, prettified, "→ at username")
	})

	t.Run("formats multiple errors with sorting", func(t *testing.T) {
		issues := []ZodIssue{
			{
				ZodIssueBase: ZodIssueBase{
					Code:    "invalid_type",
					Message: "Deep nested error",
					Path:    []interface{}{"user", "profile", "settings", "theme"},
				},
			},
			{
				ZodIssueBase: ZodIssueBase{
					Code:    "too_small",
					Message: "Root error",
					Path:    []interface{}{},
				},
			},
			{
				ZodIssueBase: ZodIssueBase{
					Code:    "invalid_format",
					Message: "Simple field error",
					Path:    []interface{}{"email"},
				},
			},
		}

		err := NewZodError(issues)
		prettified := PrettifyError(err)

		// Check that all errors are present
		assert.Contains(t, prettified, "Root error")
		assert.Contains(t, prettified, "Simple field error")
		assert.Contains(t, prettified, "Deep nested error")

		// Check that path information is included
		assert.Contains(t, prettified, "→ at email")
		assert.Contains(t, prettified, "→ at user.profile.settings.theme")

		// Root error should not have path information
		lines := strings.Split(prettified, "\n")

		// Find the line with root error and verify no path follows
		for i, line := range lines {
			if strings.Contains(line, "Root error") {
				// Next line should not contain path (since root error has empty path)
				if i+1 < len(lines) {
					nextLine := lines[i+1]
					if strings.Contains(nextLine, "→ at") {
						t.Error("Root error should not have path information")
					}
				}
				break
			}
		}
	})

	t.Run("handles errors without paths", func(t *testing.T) {
		issues := []ZodIssue{
			{
				ZodIssueBase: ZodIssueBase{
					Code:    "custom",
					Message: "Custom validation failed",
					Path:    []interface{}{},
				},
			},
		}

		err := NewZodError(issues)
		prettified := PrettifyError(err)

		assert.Contains(t, prettified, "✖ Custom validation failed")
		assert.NotContains(t, prettified, "→ at")
	})

	t.Run("handles empty error", func(t *testing.T) {
		err := NewZodError([]ZodIssue{})
		prettified := PrettifyError(err)

		assert.Empty(t, prettified)
	})
}

//////////////////////////////////////////
//////////   Mapper Function Tests     ///
//////////////////////////////////////////

func TestMapperFunctions(t *testing.T) {
	t.Run("FormatErrorWithMapper applies custom formatting", func(t *testing.T) {
		issues := []ZodIssue{
			{
				ZodIssueBase: ZodIssueBase{
					Code:    "invalid_type",
					Message: "Original message",
					Path:    []interface{}{"field"},
				},
				Expected: "string",
				Received: "number",
			},
		}

		err := NewZodError(issues)

		customMapper := func(issue ZodIssue) string {
			return fmt.Sprintf("[%s] Expected: %s, Got: %s",
				strings.ToUpper(issue.Code), issue.Expected, issue.Received)
		}

		formatted := FormatErrorWithMapper(err, customMapper)

		fieldFormatted := formatted["field"].(ZodFormattedError)
		fieldErrors := fieldFormatted["_errors"].([]string)

		require.Len(t, fieldErrors, 1)
		assert.Equal(t, "[INVALID_TYPE] Expected: string, Got: number", fieldErrors[0])
	})

	t.Run("TreeifyErrorWithMapper applies custom formatting", func(t *testing.T) {
		issues := []ZodIssue{
			{
				ZodIssueBase: ZodIssueBase{
					Code:    "too_small",
					Message: "Too small",
					Path:    []interface{}{"age"},
				},
				Minimum: 18,
			},
		}

		err := NewZodError(issues)

		contextMapper := func(issue ZodIssue) string {
			return fmt.Sprintf("Error: %s (Min: %v, Code: %s)",
				issue.Message, issue.Minimum, issue.Code)
		}

		tree := TreeifyErrorWithMapper(err, contextMapper)

		ageProp := tree.Properties["age"]
		require.NotNil(t, ageProp)
		require.Len(t, ageProp.Errors, 1)
		assert.Equal(t, "Error: Too small (Min: 18, Code: too_small)", ageProp.Errors[0])
	})

	t.Run("FlattenErrorWithMapper applies custom formatting", func(t *testing.T) {
		issues := []ZodIssue{
			{
				ZodIssueBase: ZodIssueBase{
					Code:    "invalid_format",
					Message: "Invalid email",
					Path:    []interface{}{"email"},
				},
				Format: "email",
			},
		}

		err := NewZodError(issues)

		detailMapper := func(issue ZodIssue) string {
			return fmt.Sprintf("VALIDATION_ERROR: %s (Expected format: %s)",
				issue.Message, issue.Format)
		}

		flattened := FlattenErrorWithMapper(err, detailMapper)

		require.Contains(t, flattened.FieldErrors, "email")
		emailErrors := flattened.FieldErrors["email"]
		require.Len(t, emailErrors, 1)
		assert.Equal(t, "VALIDATION_ERROR: Invalid email (Expected format: email)", emailErrors[0])
	})

	t.Run("mapper functions handle edge cases", func(t *testing.T) {
		issues := []ZodIssue{
			{
				ZodIssueBase: ZodIssueBase{
					Code:    "custom",
					Message: "",
					Path:    []interface{}{},
				},
			},
		}

		err := NewZodError(issues)

		safeMapper := func(issue ZodIssue) string {
			if issue.Message == "" {
				return fmt.Sprintf("No message provided for code: %s", issue.Code)
			}
			return issue.Message
		}

		formatted := FormatErrorWithMapper(err, safeMapper)
		rootErrors := formatted["_errors"].([]string)

		require.Len(t, rootErrors, 1)
		assert.Equal(t, "No message provided for code: custom", rootErrors[0])
	})
}

//////////////////////////////////////////
//////////   Complex Issue Type Tests  ///
//////////////////////////////////////////

func TestComplexIssueTypes(t *testing.T) {
	t.Run("handles invalid_union issues", func(t *testing.T) {
		// TODO: Test invalid_union issues when union error structures are fully implemented
		t.Skip("TODO: Test invalid_union processing when union validation error structures are available")

		// This would test:
		// - issue.Errors field processing
		// - Nested union error handling
		// - Error propagation through union alternatives
	})

	t.Run("handles invalid_key issues", func(t *testing.T) {
		// TODO: Test invalid_key issues when record/object key validation is fully implemented
		t.Skip("TODO: Test invalid_key processing when record key validation is available")

		// This would test:
		// - issue.Issues field processing
		// - Key validation error handling
		// - Nested key validation errors
	})

	t.Run("handles invalid_element issues", func(t *testing.T) {
		// TODO: Test invalid_element issues when array element validation is fully implemented
		t.Skip("TODO: Test invalid_element processing when array element validation is available")

		// This would test:
		// - issue.Issues field processing with issue.Key
		// - Array element validation errors
		// - Nested element validation handling
	})

	t.Run("processes standard issue types correctly", func(t *testing.T) {
		issues := []ZodIssue{
			{
				ZodIssueBase: ZodIssueBase{
					Code:    "invalid_type",
					Message: "Type mismatch",
					Path:    []interface{}{"field1"},
				},
			},
			{
				ZodIssueBase: ZodIssueBase{
					Code:    "too_small",
					Message: "Value too small",
					Path:    []interface{}{"field2"},
				},
			},
			{
				ZodIssueBase: ZodIssueBase{
					Code:    "custom",
					Message: "Custom validation failed",
					Path:    []interface{}{"field3"},
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
//////////   Error Edge Cases and Error Handling ///
//////////////////////////////////////////

func TestErrorEdgeCasesAndErrorHandling(t *testing.T) {
	t.Run("handles nil issues gracefully", func(t *testing.T) {
		err := NewZodError(nil)

		// All functions should handle nil issues without panicking
		assert.NotPanics(t, func() {
			_ = FormatError(err)
		})
		assert.NotPanics(t, func() {
			_ = TreeifyError(err)
		})
		assert.NotPanics(t, func() {
			_ = FlattenError(err)
		})
		assert.NotPanics(t, func() {
			_ = PrettifyError(err)
		})
	})

	t.Run("handles issues with nil paths", func(t *testing.T) {
		issues := []ZodIssue{
			{
				ZodIssueBase: ZodIssueBase{
					Code:    "custom",
					Message: "Error with nil path",
					Path:    nil,
				},
			},
		}

		err := NewZodError(issues)

		assert.NotPanics(t, func() {
			formatted := FormatError(err)
			assert.NotNil(t, formatted)
		})

		assert.NotPanics(t, func() {
			tree := TreeifyError(err)
			assert.NotNil(t, tree)
		})

		assert.NotPanics(t, func() {
			flattened := FlattenError(err)
			assert.NotNil(t, flattened)
		})
	})

	t.Run("handles issues with empty messages", func(t *testing.T) {
		issues := []ZodIssue{
			{
				ZodIssueBase: ZodIssueBase{
					Code:    "custom",
					Message: "",
					Path:    []interface{}{"field"},
				},
			},
		}

		err := NewZodError(issues)
		prettified := PrettifyError(err)

		// Should still format properly with empty message
		assert.Contains(t, prettified, "✖")
		assert.Contains(t, prettified, "→ at field")
	})

	t.Run("handles very deep nesting", func(t *testing.T) {
		deepPath := []interface{}{}
		for i := 0; i < 20; i++ {
			deepPath = append(deepPath, fmt.Sprintf("level%d", i))
		}

		issues := []ZodIssue{
			{
				ZodIssueBase: ZodIssueBase{
					Code:    "invalid_type",
					Message: "Deep nested error",
					Path:    deepPath,
				},
			},
		}

		err := NewZodError(issues)

		// Should handle deep nesting without issues
		assert.NotPanics(t, func() {
			_ = FormatError(err)
			_ = TreeifyError(err)
			_ = FlattenError(err)
			_ = PrettifyError(err)
		})
	})

	t.Run("handles large numbers of issues", func(t *testing.T) {
		issues := []ZodIssue{}
		for i := 0; i < 100; i++ {
			issues = append(issues, ZodIssue{
				ZodIssueBase: ZodIssueBase{
					Code:    "invalid_type",
					Message: fmt.Sprintf("Error %d", i),
					Path:    []interface{}{fmt.Sprintf("field%d", i)},
				},
			})
		}

		err := NewZodError(issues)

		// Should handle large numbers of issues efficiently
		assert.NotPanics(t, func() {
			formatted := FormatError(err)
			assert.NotNil(t, formatted)
		})

		prettified := PrettifyError(err)
		assert.Contains(t, prettified, "Error 0")
		assert.Contains(t, prettified, "Error 99")
	})

	t.Run("handles special characters in paths and messages", func(t *testing.T) {
		issues := []ZodIssue{
			{
				ZodIssueBase: ZodIssueBase{
					Code:    "custom",
					Message: "Error with unicode: 中文测试 🚀",
					Path:    []interface{}{"field@special", "nested-field", "数据"},
				},
			},
		}

		err := NewZodError(issues)

		assert.NotPanics(t, func() {
			formatted := FormatError(err)
			tree := TreeifyError(err)
			flattened := FlattenError(err)
			prettified := PrettifyError(err)

			assert.NotNil(t, formatted)
			assert.NotNil(t, tree)
			assert.NotNil(t, flattened)
			assert.NotEmpty(t, prettified)
		})
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
					Path:    []interface{}{},
				},
				Keys: []string{"extraField"},
			},
			{
				ZodIssueBase: ZodIssueBase{
					Code:    "invalid_type",
					Message: "Expected string, received number",
					Path:    []interface{}{"user", "name"},
				},
				Expected: "string",
				Received: "number",
			},
			{
				ZodIssueBase: ZodIssueBase{
					Code:    "too_small",
					Message: "Must be at least 18",
					Path:    []interface{}{"user", "age"},
				},
				Minimum: 18,
			},
			{
				ZodIssueBase: ZodIssueBase{
					Code:    "invalid_format",
					Message: "Invalid email format",
					Path:    []interface{}{"user", "contacts", 0, "email"},
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
		assert.Contains(t, prettified, "✖ Unknown field 'extraField'")
		assert.Contains(t, prettified, "✖ Expected string, received number")
		assert.Contains(t, prettified, "→ at user.name")
		assert.Contains(t, prettified, "→ at user.contacts[0].email")
	})

	t.Run("error identification and conversion workflow", func(t *testing.T) {
		// Create ZodError
		zodErr := NewZodError([]ZodIssue{
			{
				ZodIssueBase: ZodIssueBase{
					Code:    "invalid_type",
					Message: "Test error",
					Path:    []interface{}{"field"},
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
		assert.Contains(t, prettified, "→ at field")
	})

	t.Run("custom mapper integration", func(t *testing.T) {
		issues := []ZodIssue{
			{
				ZodIssueBase: ZodIssueBase{
					Code:    "invalid_type",
					Message: "Type error",
					Path:    []interface{}{"field1"},
				},
				Expected: "string",
				Received: "number",
			},
			{
				ZodIssueBase: ZodIssueBase{
					Code:    "too_small",
					Message: "Size error",
					Path:    []interface{}{"field2"},
				},
				Minimum: 5,
			},
		}

		err := NewZodError(issues)

		// Create a comprehensive mapper
		detailedMapper := func(issue ZodIssue) string {
			switch issue.Code {
			case "invalid_type":
				return fmt.Sprintf("TYPE_ERROR: Expected '%s' but received '%s' at %s",
					issue.Expected, issue.Received, ToDotPath(issue.Path))
			case "too_small":
				return fmt.Sprintf("SIZE_ERROR: Value must be at least %v at %s",
					issue.Minimum, ToDotPath(issue.Path))
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

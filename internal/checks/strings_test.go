package checks

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kaptinlin/gozod/core"
)

func TestStringContentChecks(t *testing.T) {
	t.Run("Includes validates substring presence", func(t *testing.T) {
		check := Includes("test")

		// Test valid case
		payload := core.NewParsePayload("this is a test string")
		executeCheck(check, payload)
		assert.Len(t, payload.Issues(), 0, "Expected no issues for string containing substring")

		// Test invalid case
		payload = core.NewParsePayload("this string doesn't contain the word")
		executeCheck(check, payload)
		assert.Len(t, payload.Issues(), 1, "Expected 1 issue for string not containing substring")
	})

	t.Run("StartsWith validates prefix", func(t *testing.T) {
		check := StartsWith("hello")

		// Test valid case
		payload := core.NewParsePayload("hello world")
		executeCheck(check, payload)
		assert.Len(t, payload.Issues(), 0, "Expected no issues for string with correct prefix")

		// Test invalid case
		payload = core.NewParsePayload("hi world")
		executeCheck(check, payload)
		assert.Len(t, payload.Issues(), 1, "Expected 1 issue for string with wrong prefix")
	})

	t.Run("EndsWith validates suffix", func(t *testing.T) {
		check := EndsWith("world")

		// Test valid case
		payload := core.NewParsePayload("hello world")
		executeCheck(check, payload)
		assert.Len(t, payload.Issues(), 0, "Expected no issues for string with correct suffix")

		// Test invalid case
		payload = core.NewParsePayload("hello universe")
		executeCheck(check, payload)
		assert.Len(t, payload.Issues(), 1, "Expected 1 issue for string with wrong suffix")
	})
}

func TestCaseValidationChecks(t *testing.T) {
	t.Run("Lowercase validates lowercase strings", func(t *testing.T) {
		check := Lowercase()

		// Test valid case
		payload := core.NewParsePayload("hello world")
		executeCheck(check, payload)
		assert.Len(t, payload.Issues(), 0, "Expected no issues for lowercase string")

		// Test invalid case
		payload = core.NewParsePayload("Hello World")
		executeCheck(check, payload)
		assert.Len(t, payload.Issues(), 1, "Expected 1 issue for non-lowercase string")
	})

	t.Run("Uppercase validates uppercase strings", func(t *testing.T) {
		check := Uppercase()

		// Test valid case
		payload := core.NewParsePayload("HELLO WORLD")
		executeCheck(check, payload)
		assert.Len(t, payload.Issues(), 0, "Expected no issues for uppercase string")

		// Test invalid case
		payload = core.NewParsePayload("Hello World")
		executeCheck(check, payload)
		assert.Len(t, payload.Issues(), 1, "Expected 1 issue for non-uppercase string")
	})
}

func TestRegexCheck(t *testing.T) {
	t.Run("Regex validates against pattern", func(t *testing.T) {
		pattern := regexp.MustCompile(`^\d{3}-\d{3}-\d{4}$`)
		check := Regex(pattern)

		// Test valid case
		payload := core.NewParsePayload("123-456-7890")
		executeCheck(check, payload)
		assert.Len(t, payload.Issues(), 0, "Expected no issues for matching regex")

		// Test invalid case
		payload = core.NewParsePayload("123-45-6789")
		executeCheck(check, payload)
		assert.Len(t, payload.Issues(), 1, "Expected 1 issue for non-matching regex")
	})
}

func TestStringCustomMessages(t *testing.T) {
	t.Run("Custom error messages work for string checks", func(t *testing.T) {
		check := Includes("test", "Must include the word 'test'")

		payload := core.NewParsePayload("hello world")
		executeCheck(check, payload)

		require.Len(t, payload.Issues(), 1, "Expected 1 issue")

		internals := check.Zod()
		assert.NotNil(t, internals.Def.Error, "Expected custom error mapping to be set")
	})
}

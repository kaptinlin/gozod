package checks

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kaptinlin/gozod/core"
)

func TestFormatChecks(t *testing.T) {
	t.Run("Email validates email format", func(t *testing.T) {
		check := Email()

		// Test valid case
		payload := core.NewParsePayload("test@example.com")
		executeCheck(check, payload)
		assert.Len(t, payload.Issues(), 0, "Expected no issues for valid email")

		// Test invalid case
		payload = core.NewParsePayload("not-an-email")
		executeCheck(check, payload)
		assert.Len(t, payload.Issues(), 1, "Expected 1 issue for invalid email")
	})

	t.Run("URL validates URL format", func(t *testing.T) {
		check := URL()

		// Test valid case
		payload := core.NewParsePayload("https://example.com")
		executeCheck(check, payload)
		assert.Len(t, payload.Issues(), 0, "Expected no issues for valid URL")

		// Test invalid case
		payload = core.NewParsePayload("not-a-url")
		executeCheck(check, payload)
		assert.Len(t, payload.Issues(), 1, "Expected 1 issue for invalid URL")
	})
}

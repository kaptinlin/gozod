package checks

import (
	"testing"

	"github.com/kaptinlin/gozod/core"
)

func TestFormatChecks(t *testing.T) {
	t.Run("Email validates email format", func(t *testing.T) {
		check := Email()

		// Test valid case
		payload := core.NewParsePayload("test@example.com")
		executeCheck(check, payload)
		if len(payload.Issues()) != 0 {
			t.Errorf("Expected no issues for valid email, got %d", len(payload.Issues()))
		}

		// Test invalid case
		payload = core.NewParsePayload("not-an-email")
		executeCheck(check, payload)
		if len(payload.Issues()) != 1 {
			t.Errorf("Expected 1 issue for invalid email, got %d", len(payload.Issues()))
		}
	})

	t.Run("URL validates URL format", func(t *testing.T) {
		check := URL()

		// Test valid case
		payload := core.NewParsePayload("https://example.com")
		executeCheck(check, payload)
		if len(payload.Issues()) != 0 {
			t.Errorf("Expected no issues for valid URL, got %d", len(payload.Issues()))
		}

		// Test invalid case
		payload = core.NewParsePayload("not-a-url")
		executeCheck(check, payload)
		if len(payload.Issues()) != 1 {
			t.Errorf("Expected 1 issue for invalid URL, got %d", len(payload.Issues()))
		}
	})
}

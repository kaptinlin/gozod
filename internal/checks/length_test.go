package checks

import (
	"testing"

	"github.com/kaptinlin/gozod/core"
)

func TestLengthChecks(t *testing.T) {
	t.Run("MinLength validates minimum length", func(t *testing.T) {
		check := MinLength(5)

		// Test valid case
		payload := core.NewParsePayload("hello")
		executeCheck(check, payload)
		if len(payload.GetIssues()) != 0 {
			t.Errorf("Expected no issues for valid length, got %d", len(payload.GetIssues()))
		}

		// Test invalid case
		payload = core.NewParsePayload("hi")
		executeCheck(check, payload)
		if len(payload.GetIssues()) != 1 {
			t.Errorf("Expected 1 issue for invalid length, got %d", len(payload.GetIssues()))
		}
	})
}

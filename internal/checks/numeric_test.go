package checks

import (
	"testing"

	"github.com/kaptinlin/gozod/core"
)

func TestNumericChecks(t *testing.T) {
	t.Run("Gte validates minimum value", func(t *testing.T) {
		check := Gte(5)

		// Test valid case
		payload := core.NewParsePayload(10)
		executeCheck(check, payload)
		if len(payload.GetIssues()) != 0 {
			t.Errorf("Expected no issues for valid value, got %d", len(payload.GetIssues()))
		}

		// Test invalid case
		payload = core.NewParsePayload(3)
		executeCheck(check, payload)
		if len(payload.GetIssues()) != 1 {
			t.Errorf("Expected 1 issue for invalid value, got %d", len(payload.GetIssues()))
		}
	})
}

package checks

import (
	"testing"

	"github.com/kaptinlin/gozod/core"
)

func TestNumericChecks(t *testing.T) {
	t.Run("Lt validates less than", func(t *testing.T) {
		check := Lt(10)

		// Test valid case
		payload := &core.ParsePayload{
			Value:  5,
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)
		if len(payload.Issues) != 0 {
			t.Errorf("Expected no issues for 5 < 10, got %d", len(payload.Issues))
		}

		// Test invalid case
		payload = &core.ParsePayload{
			Value:  15,
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)
		if len(payload.Issues) != 1 {
			t.Errorf("Expected 1 issue for 15 >= 10, got %d", len(payload.Issues))
		}
	})

	t.Run("Gt validates greater than", func(t *testing.T) {
		check := Gt(0)

		// Test valid case
		payload := &core.ParsePayload{
			Value:  5,
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)
		if len(payload.Issues) != 0 {
			t.Errorf("Expected no issues for 5 > 0, got %d", len(payload.Issues))
		}

		// Test invalid case
		payload = &core.ParsePayload{
			Value:  -1,
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)
		if len(payload.Issues) != 1 {
			t.Errorf("Expected 1 issue for -1 <= 0, got %d", len(payload.Issues))
		}
	})

	t.Run("MultipleOf validates multiples", func(t *testing.T) {
		check := MultipleOf(3)

		// Test valid case
		payload := &core.ParsePayload{
			Value:  9,
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)
		if len(payload.Issues) != 0 {
			t.Errorf("Expected no issues for 9 %% 3 == 0, got %d", len(payload.Issues))
		}

		// Test invalid case
		payload = &core.ParsePayload{
			Value:  10,
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)
		if len(payload.Issues) != 1 {
			t.Errorf("Expected 1 issue for 10 %% 3 != 0, got %d", len(payload.Issues))
		}
	})

	t.Run("Positive convenience function", func(t *testing.T) {
		check := Positive()

		// Test valid case
		payload := &core.ParsePayload{
			Value:  5,
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)
		if len(payload.Issues) != 0 {
			t.Errorf("Expected no issues for positive number, got %d", len(payload.Issues))
		}

		// Test invalid case
		payload = &core.ParsePayload{
			Value:  0,
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)
		if len(payload.Issues) != 1 {
			t.Errorf("Expected 1 issue for zero, got %d", len(payload.Issues))
		}
	})

	t.Run("Custom error messages work", func(t *testing.T) {
		check := Lt(10, "Value must be less than 10")

		payload := &core.ParsePayload{
			Value:  15,
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)

		if len(payload.Issues) != 1 {
			t.Fatalf("Expected 1 issue, got %d", len(payload.Issues))
		}

		internals := check.GetZod()
		if internals.Def.Error == nil {
			t.Error("Expected custom error mapping to be set")
		}
	})
}

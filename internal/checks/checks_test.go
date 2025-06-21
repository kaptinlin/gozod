package checks

import (
	"testing"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/internal/utils"
	"github.com/kaptinlin/gozod/pkg/validate"
)

// =============================================================================
// TESTING UTILITIES
// =============================================================================

// executeCheck executes the check logic - utility for testing
func executeCheck(check core.ZodCheck, payload *core.ParsePayload) {
	if internals := check.GetZod(); internals != nil {
		// check conditional execution
		if internals.When != nil && !internals.When(payload) {
			return // skip check
		}

		// execute check function
		if internals.Check != nil {
			internals.Check(payload)
		}
	}
}

func TestDirectCheckCreation(t *testing.T) {
	def := &core.ZodCheckDef{
		Check: "test_check",
	}

	check := &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if payload.Value == nil {
				return
			}
			if str, ok := payload.Value.(string); ok && str == "invalid" {
				payload.Issues = append(payload.Issues, issues.CreateCustomIssue("test error", nil, payload.Value))
			}
		},
	}

	if check == nil {
		t.Fatal("Check creation returned nil")
	}

	internals := check.GetZod()
	if internals == nil {
		t.Fatal("GetZod returned nil")
	}

	if internals.Def.Check != "test_check" {
		t.Errorf("Expected check type 'test_check', got '%s'", internals.Def.Check)
	}
}

func TestNormalizeParams(t *testing.T) {
	def := &core.ZodCheckDef{
		Check: "test",
	}

	normalizeParams(def, []string{})
	if def.Error != nil {
		t.Error("Expected no error mapping for empty params")
	}

	normalizeParams(def, []string{"custom error message"})
	if def.Error == nil {
		t.Error("Expected error mapping for string param")
	}
}

func TestCheckExecution(t *testing.T) {
	def := &core.ZodCheckDef{
		Check: "length_check",
	}

	check := &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.MaxLength(payload.Value, 5) {
				origin := utils.GetOriginFromValue(payload.Value)
				payload.Issues = append(payload.Issues, issues.CreateTooBigIssue(5, true, origin, payload.Value))
			}
		},
	}

	payload := &core.ParsePayload{
		Value:  "test",
		Issues: make([]core.ZodRawIssue, 0),
	}

	executeCheck(check, payload)

	if len(payload.Issues) != 0 {
		t.Errorf("Expected no issues for valid input, got %d", len(payload.Issues))
	}

	payload = &core.ParsePayload{
		Value:  "this is too long",
		Issues: make([]core.ZodRawIssue, 0),
	}

	executeCheck(check, payload)

	if len(payload.Issues) != 1 {
		t.Errorf("Expected 1 issue for invalid input, got %d", len(payload.Issues))
	}
}

func TestConditionalExecution(t *testing.T) {
	def := &core.ZodCheckDef{
		Check: "conditional_check",
	}

	check := &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			payload.Issues = append(payload.Issues, issues.CreateCustomIssue("should not execute", nil, payload.Value))
		},
		When: func(payload *core.ParsePayload) bool {
			_, ok := payload.Value.(string)
			return ok
		},
	}

	payload := &core.ParsePayload{
		Value:  42,
		Issues: make([]core.ZodRawIssue, 0),
	}

	executeCheck(check, payload)

	if len(payload.Issues) != 0 {
		t.Errorf("Expected no issues for skipped check, got %d", len(payload.Issues))
	}

	payload = &core.ParsePayload{
		Value:  "test",
		Issues: make([]core.ZodRawIssue, 0),
	}

	executeCheck(check, payload)

	if len(payload.Issues) != 1 {
		t.Errorf("Expected 1 issue for executed check, got %d", len(payload.Issues))
	}
}

func BenchmarkDirectCheckCreation(b *testing.B) {
	def := &core.ZodCheckDef{
		Check: "benchmark_check",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		check := &core.ZodCheckInternals{
			Def: def,
			Check: func(payload *core.ParsePayload) {
				// Simple validation
			},
		}
		_ = check
	}
}

func BenchmarkCheckExecution(b *testing.B) {
	def := &core.ZodCheckDef{
		Check: "benchmark_check",
	}

	check := &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			validate.MaxLength(payload.Value, 10)
		},
	}

	payload := &core.ParsePayload{
		Value:  "test",
		Issues: make([]core.ZodRawIssue, 0, 1),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		payload.Issues = payload.Issues[:0]
		executeCheck(check, payload)
	}
}

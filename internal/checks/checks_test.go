package checks

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
			if payload.GetValue() == nil {
				return
			}
			if str, ok := payload.GetValue().(string); ok && str == "invalid" {
				payload.AddIssue(issues.CreateCustomIssue("test error", nil, payload.GetValue()))
			}
		},
	}

	internals := check.GetZod()
	require.NotNil(t, internals, "GetZod returned nil")
	assert.Equal(t, "test_check", internals.Def.Check)
}

func TestCheckExecution(t *testing.T) {
	def := &core.ZodCheckDef{
		Check: "length_check",
	}

	check := &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.MaxLength(payload.GetValue(), 5) {
				origin := utils.GetOriginFromValue(payload.GetValue())
				payload.AddIssue(issues.CreateTooBigIssue(5, true, origin, payload.GetValue()))
			}
		},
	}

	payload := core.NewParsePayload("test")

	executeCheck(check, payload)

	if len(payload.GetIssues()) != 0 {
		t.Errorf("Expected no issues for valid input, got %d", len(payload.GetIssues()))
	}

	payload = core.NewParsePayload("this is too long")

	executeCheck(check, payload)

	if len(payload.GetIssues()) != 1 {
		t.Errorf("Expected 1 issue for invalid input, got %d", len(payload.GetIssues()))
	}
}

func TestConditionalExecution(t *testing.T) {
	def := &core.ZodCheckDef{
		Check: "conditional_check",
	}

	check := &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			payload.AddIssue(issues.CreateCustomIssue("should not execute", nil, payload.GetValue()))
		},
		When: func(payload *core.ParsePayload) bool {
			_, ok := payload.GetValue().(string)
			return ok
		},
	}

	payload := core.NewParsePayload(42)

	executeCheck(check, payload)

	if len(payload.GetIssues()) != 0 {
		t.Errorf("Expected no issues for skipped check, got %d", len(payload.GetIssues()))
	}

	payload = core.NewParsePayload("test")

	executeCheck(check, payload)

	if len(payload.GetIssues()) != 1 {
		t.Errorf("Expected 1 issue for executed check, got %d", len(payload.GetIssues()))
	}
}

func BenchmarkDirectCheckCreation(b *testing.B) {
	def := &core.ZodCheckDef{
		Check: "benchmark_check",
	}

	b.ResetTimer()
	for b.Loop() {
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
			validate.MaxLength(payload.GetValue(), 10)
		},
	}

	b.ResetTimer()
	for b.Loop() {
		newPayload := core.NewParsePayload("test")
		executeCheck(check, newPayload)
	}
}

package engine

import (
	"testing"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/issues"
)

func TestRunChecks(t *testing.T) {
	t.Run("successful validation", func(t *testing.T) {
		payload := core.NewParsePayload("test")

		check := checks.NewCustom[string](func(v any) bool {
			if s, ok := v.(string); ok {
				return len(s) > 2
			}
			return false
		}, core.SchemaParams{})

		checkList := []core.ZodCheck{check}

		result := RunChecks(checkList, payload)

		if result == nil {
			t.Fatal("Expected result, got nil")
		}

		if len(result.Issues()) != 0 {
			t.Errorf("Expected no issues, got %d", len(result.Issues()))
		}
	})

	t.Run("with context parameter", func(t *testing.T) {
		payload := core.NewParsePayload("test")

		check := checks.NewCustom[string](func(v any) bool {
			return true
		}, core.SchemaParams{})

		checkList := []core.ZodCheck{check}
		ctx := &core.ParseContext{}

		result := RunChecks(checkList, payload, ctx)

		if result == nil {
			t.Fatal("Expected result, got nil")
		}
	})

	t.Run("failed validation", func(t *testing.T) {
		payload := core.NewParsePayload("x") // Short string

		check := checks.NewCustom[string](func(v any) bool {
			if s, ok := v.(string); ok {
				return len(s) > 2
			}
			return false
		}, core.SchemaParams{})

		checkList := []core.ZodCheck{check}

		result := RunChecks(checkList, payload)

		if result == nil {
			t.Fatal("Expected result, got nil")
		}

		if len(result.Issues()) == 0 {
			t.Error("Expected validation issues, got none")
		}
	})

	t.Run("nil payload", func(t *testing.T) {
		check := checks.NewCustom[string](func(v any) bool { return true }, core.SchemaParams{})
		result := RunChecks([]core.ZodCheck{check}, nil)

		if result != nil {
			t.Error("Expected nil result for nil payload")
		}
	})

	t.Run("empty checks", func(t *testing.T) {
		payload := core.NewParsePayload("test")
		result := RunChecks([]core.ZodCheck{}, payload)

		if result != payload {
			t.Error("Expected same payload for empty checks")
		}
	})
}

func TestRunChecksOnValue(t *testing.T) {
	t.Run("successful validation", func(t *testing.T) {
		payload := core.NewParsePayload(nil)

		check := checks.NewCustom[string](func(v any) bool {
			if s, ok := v.(string); ok {
				return len(s) > 2
			}
			return false
		}, core.SchemaParams{})

		checkList := []core.ZodCheck{check}

		result := RunChecksOnValue("test", checkList, payload)

		if result == nil {
			t.Fatal("Expected result, got nil")
		}

		if len(result.Issues()) != 0 {
			t.Errorf("Expected no issues, got %d", len(result.Issues()))
		}
	})

	t.Run("with context parameter", func(t *testing.T) {
		payload := core.NewParsePayload(nil)

		check := checks.NewCustom[string](func(v any) bool {
			return true
		}, core.SchemaParams{})

		checkList := []core.ZodCheck{check}
		ctx := &core.ParseContext{}

		result := RunChecksOnValue("test", checkList, payload, ctx)

		if result == nil {
			t.Fatal("Expected result, got nil")
		}
	})

	t.Run("failed validation", func(t *testing.T) {
		payload := core.NewParsePayload(nil)

		check := checks.NewCustom[string](func(v any) bool {
			if s, ok := v.(string); ok {
				return len(s) > 5 // This should fail for "test"
			}
			return false
		}, core.CustomParams{})

		checkList := []core.ZodCheck{check}

		result := RunChecksOnValue("test", checkList, payload)

		if result == nil {
			t.Fatal("Expected result, got nil")
		}

		if len(result.Issues()) == 0 {
			t.Error("Expected validation issues, got none")
		}
	})

	t.Run("nil payload", func(t *testing.T) {
		check := checks.NewCustom[string](func(v any) bool { return true }, core.CustomParams{})

		result := RunChecksOnValue("test", []core.ZodCheck{check}, nil)

		if result != nil {
			t.Error("Expected nil result for nil payload")
		}
	})

	t.Run("empty checks", func(t *testing.T) {
		payload := core.NewParsePayload("initial")
		initialIssues := len(payload.Issues())

		result := RunChecksOnValue("test", []core.ZodCheck{}, payload)

		if result == nil {
			t.Fatal("Expected result, got nil")
		}

		if len(result.Issues()) != initialIssues {
			t.Error("Expected no new issues for empty checks")
		}
	})
}

func TestExecuteChecks(t *testing.T) {
	t.Run("custom error mapping", func(t *testing.T) {
		payload := core.NewParsePayload("test")

		check := checks.NewCustom[string](func(v any) bool {
			return false // Always fail
		}, core.CustomParams{
			Error: "Custom error message",
		})

		checkList := []core.ZodCheck{check}

		result := executeChecks("test", checkList, payload, nil)

		if result == nil {
			t.Fatal("Expected result, got nil")
		}

		if len(result.Issues()) == 0 {
			t.Fatal("Expected validation issues")
		}

		issue := result.Issues()[0]
		if issue.Message != "Custom error message" {
			t.Errorf("Expected custom error message, got %s", issue.Message)
		}
	})

	t.Run("abort flag early termination", func(t *testing.T) {
		payload := core.NewParsePayload("test")

		check1 := checks.NewCustom[string](func(v any) bool {
			return false // Always fail
		}, core.CustomParams{
			Abort: true,
		})

		check2 := checks.NewCustom[string](func(v any) bool {
			return false // Always fail
		}, core.CustomParams{})

		checkList := []core.ZodCheck{check1, check2}

		result := executeChecks("test", checkList, payload, nil)

		if result == nil {
			t.Fatal("Expected result, got nil")
		}

		issues := result.Issues()
		if len(issues) == 0 {
			t.Fatal("Expected at least one issue")
		}
	})

	t.Run("abort skips check with When predicate", func(t *testing.T) {
		payload := core.NewParsePayload("hi")

		// Check 1: always fails, with Abort: true
		check1 := checks.NewCustom[string](func(v any) bool {
			return false
		}, core.CustomParams{
			Abort: true,
		})

		// Check 2: has a When predicate (like built-in Min/Max) — should be skipped
		check2 := checks.NewCustom[string](func(v any) bool {
			return false // Would fail if reached
		}, core.CustomParams{
			When: func(_ *core.ParsePayload) bool {
				return true // When passes, but abort should prevent execution
			},
		})

		checkList := []core.ZodCheck{check1, check2}
		result := executeChecks("hi", checkList, payload, nil)

		if result == nil {
			t.Fatal("Expected result, got nil")
		}

		// Should have exactly 1 issue from check1, not 2
		if len(result.Issues()) != 1 {
			t.Errorf("Expected exactly 1 issue (abort should skip When-gated check), got %d", len(result.Issues()))
		}
	})

	t.Run("memory optimization with sufficient capacity", func(t *testing.T) {
		payload := core.NewParsePayload("test")

		// Pre-allocate some capacity
		payload.AddIssue(issues.NewRawIssue(core.InvalidType, "test", issues.WithExpected("string")))
		payload.SetIssues(make([]core.ZodRawIssue, 1, 10)) // Capacity 10

		// Create multiple checks to test memory allocation
		var checkList []core.ZodCheck
		for i := 0; i < 5; i++ {
			check := checks.NewCustom[string](func(v any) bool {
				return false // Always fail to generate issues
			}, core.CustomParams{})
			checkList = append(checkList, check)
		}

		result := executeChecks("test", checkList, payload, nil)

		if result == nil {
			t.Fatal("Expected result, got nil")
		}

		// Should have issues from all checks
		if len(result.Issues()) <= 1 {
			t.Error("Expected multiple issues from checks")
		}
	})

	t.Run("performance with no issues", func(t *testing.T) {
		payload := core.NewParsePayload("test")

		// Create multiple checks that all pass
		var checkList []core.ZodCheck
		for i := 0; i < 10; i++ {
			check := checks.NewCustom[string](func(v any) bool {
				return true // Always pass
			}, core.CustomParams{})
			checkList = append(checkList, check)
		}

		result := executeChecks("test", checkList, payload, nil)

		if result == nil {
			t.Fatal("Expected result, got nil")
		}

		if len(result.Issues()) != 0 {
			t.Errorf("Expected no issues, got %d", len(result.Issues()))
		}
	})
}

func TestCheckAborted(t *testing.T) {
	t.Run("no issues", func(t *testing.T) {
		payload := core.NewParsePayload("test")

		aborted := CheckAborted(*payload, 0)
		if aborted {
			t.Error("Expected not aborted for empty issues")
		}
	})

	t.Run("start index out of bounds", func(t *testing.T) {
		payload := core.NewParsePayload("test")
		issue := issues.NewRawIssue(core.InvalidType, "test", issues.WithExpected("string"))
		payload.AddIssue(issue)

		aborted := CheckAborted(*payload, 10)
		if aborted {
			t.Error("Expected not aborted for out of bounds index")
		}
	})

	t.Run("issues with continue flag", func(t *testing.T) {
		payload := core.NewParsePayload("test")
		issue := issues.NewRawIssue(core.InvalidType, "test", issues.WithExpected("string"))
		issue.Continue = true
		payload.AddIssue(issue)

		aborted := CheckAborted(*payload, 0)
		if aborted {
			t.Error("Expected not aborted when Continue is true")
		}
	})

	t.Run("issues with abort flag", func(t *testing.T) {
		payload := core.NewParsePayload("test")
		issue := issues.NewRawIssue(core.InvalidType, "test", issues.WithExpected("string"))
		issue.Continue = false
		payload.AddIssue(issue)

		aborted := CheckAborted(*payload, 0)
		if !aborted {
			t.Error("Expected aborted when Continue is false")
		}
	})
}

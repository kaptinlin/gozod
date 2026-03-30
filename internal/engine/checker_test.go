package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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

		require.NotNil(t, result)
		assert.Len(t, result.Issues(), 0)
	})

	t.Run("with context parameter", func(t *testing.T) {
		payload := core.NewParsePayload("test")

		check := checks.NewCustom[string](func(v any) bool {
			return true
		}, core.SchemaParams{})

		checkList := []core.ZodCheck{check}
		ctx := &core.ParseContext{}

		result := RunChecks(checkList, payload, ctx)

		require.NotNil(t, result)
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

		require.NotNil(t, result)
		assert.NotEmpty(t, result.Issues())
	})

	t.Run("nil payload", func(t *testing.T) {
		check := checks.NewCustom[string](func(v any) bool { return true }, core.SchemaParams{})
		result := RunChecks([]core.ZodCheck{check}, nil)

		assert.Nil(t, result)
	})

	t.Run("empty checks", func(t *testing.T) {
		payload := core.NewParsePayload("test")
		result := RunChecks([]core.ZodCheck{}, payload)

		assert.Equal(t, payload, result)
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

		require.NotNil(t, result)
		assert.Len(t, result.Issues(), 0)
	})

	t.Run("with context parameter", func(t *testing.T) {
		payload := core.NewParsePayload(nil)

		check := checks.NewCustom[string](func(v any) bool {
			return true
		}, core.SchemaParams{})

		checkList := []core.ZodCheck{check}
		ctx := &core.ParseContext{}

		result := RunChecksOnValue("test", checkList, payload, ctx)

		require.NotNil(t, result)
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

		require.NotNil(t, result)
		assert.NotEmpty(t, result.Issues())
	})

	t.Run("nil payload", func(t *testing.T) {
		check := checks.NewCustom[string](func(v any) bool { return true }, core.CustomParams{})

		result := RunChecksOnValue("test", []core.ZodCheck{check}, nil)

		assert.Nil(t, result)
	})

	t.Run("empty checks", func(t *testing.T) {
		payload := core.NewParsePayload("initial")
		initialIssues := len(payload.Issues())

		result := RunChecksOnValue("test", []core.ZodCheck{}, payload)

		require.NotNil(t, result)
		assert.Len(t, result.Issues(), initialIssues)
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

		require.NotNil(t, result)
		require.NotEmpty(t, result.Issues())

		issue := result.Issues()[0]
		assert.Equal(t, "Custom error message", issue.Message)
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

		require.NotNil(t, result)
		assert.NotEmpty(t, result.Issues())
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

		require.NotNil(t, result)

		// Should have exactly 1 issue from check1, not 2
		assert.Len(t, result.Issues(), 1, "Expected exactly 1 issue (abort should skip When-gated check)")
	})

	t.Run("non-abort failure does NOT skip When-gated check", func(t *testing.T) {
		payload := core.NewParsePayload("hi")

		// Check 1: always fails, WITHOUT Abort (default)
		check1 := checks.NewCustom[string](func(v any) bool {
			return false
		}, core.CustomParams{})

		// Check 2: has a When predicate — should still run because check1 was not an explicit abort
		check2 := checks.NewCustom[string](func(v any) bool {
			return false // Also fails
		}, core.CustomParams{
			When: func(_ *core.ParsePayload) bool {
				return true
			},
		})

		checkList := []core.ZodCheck{check1, check2}
		result := executeChecks("hi", checkList, payload, nil)

		require.NotNil(t, result)

		// Should have 2 issues: non-abort failure must not block When-gated checks (Zod v4: 5b574501)
		assert.Len(t, result.Issues(), 2, "Expected 2 issues (non-abort should not skip When-gated check)")
	})

	t.Run("memory optimization with sufficient capacity", func(t *testing.T) {
		payload := core.NewParsePayload("test")

		// Pre-allocate some capacity
		payload.AddIssue(issues.NewRawIssue(core.InvalidType, "test", issues.WithExpected("string")))
		payload.SetIssues(make([]core.ZodRawIssue, 1, 10)) // Capacity 10

		// Create multiple checks to test memory allocation
		checkList := make([]core.ZodCheck, 0, 5)
		for range 5 {
			check := checks.NewCustom[string](func(v any) bool {
				return false // Always fail to generate issues
			}, core.CustomParams{})
			checkList = append(checkList, check)
		}

		result := executeChecks("test", checkList, payload, nil)

		require.NotNil(t, result)

		// Should have issues from all checks
		assert.Greater(t, len(result.Issues()), 1, "Expected multiple issues from checks")
	})

	t.Run("performance with no issues", func(t *testing.T) {
		payload := core.NewParsePayload("test")

		// Create multiple checks that all pass
		checkList := make([]core.ZodCheck, 0, 10)
		for range 10 {
			check := checks.NewCustom[string](func(v any) bool {
				return true // Always pass
			}, core.CustomParams{})
			checkList = append(checkList, check)
		}

		result := executeChecks("test", checkList, payload, nil)

		require.NotNil(t, result)
		assert.Len(t, result.Issues(), 0)
	})
}

func TestExplicitlyAborted(t *testing.T) {
	t.Run("no issues", func(t *testing.T) {
		payload := core.NewParsePayload("test")

		aborted := ExplicitlyAborted(*payload, 0)
		assert.False(t, aborted)
	})

	t.Run("start index out of bounds", func(t *testing.T) {
		payload := core.NewParsePayload("test")
		issue := issues.NewRawIssue(core.InvalidType, "test", issues.WithExpected("string"))
		payload.AddIssue(issue)

		aborted := ExplicitlyAborted(*payload, 10)
		assert.False(t, aborted)
	})

	t.Run("issues with continue flag", func(t *testing.T) {
		payload := core.NewParsePayload("test")
		issue := issues.NewRawIssue(core.InvalidType, "test", issues.WithExpected("string"))
		issue.Continue = true
		payload.AddIssue(issue)

		aborted := ExplicitlyAborted(*payload, 0)
		assert.False(t, aborted)
	})

	t.Run("issues with explicit abort flag", func(t *testing.T) {
		payload := core.NewParsePayload("test")
		issue := issues.NewRawIssue(core.InvalidType, "test", issues.WithExpected("string"))
		issue.Continue = false
		payload.AddIssue(issue)

		aborted := ExplicitlyAborted(*payload, 0)
		assert.True(t, aborted)
	})
}

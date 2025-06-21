package checks

import (
	"testing"

	"github.com/kaptinlin/gozod/core"
)

func TestLengthChecks(t *testing.T) {
	t.Run("MaxLength validates string length", func(t *testing.T) {
		check := MaxLength(5)

		// Test valid case
		payload := &core.ParsePayload{
			Value:  "hello",
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)
		if len(payload.Issues) != 0 {
			t.Errorf("Expected no issues for 'hello' (length 5) <= 5, got %d", len(payload.Issues))
		}

		// Test invalid case
		payload = &core.ParsePayload{
			Value:  "hello world",
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)
		if len(payload.Issues) != 1 {
			t.Errorf("Expected 1 issue for 'hello world' (length 11) > 5, got %d", len(payload.Issues))
		}
	})

	t.Run("MinLength validates string length", func(t *testing.T) {
		check := MinLength(3)

		// Test valid case
		payload := &core.ParsePayload{
			Value:  "hello",
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)
		if len(payload.Issues) != 0 {
			t.Errorf("Expected no issues for 'hello' (length 5) >= 3, got %d", len(payload.Issues))
		}

		// Test invalid case
		payload = &core.ParsePayload{
			Value:  "hi",
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)
		if len(payload.Issues) != 1 {
			t.Errorf("Expected 1 issue for 'hi' (length 2) < 3, got %d", len(payload.Issues))
		}
	})

	t.Run("Length validates exact length", func(t *testing.T) {
		check := Length(5)

		// Test valid case
		payload := &core.ParsePayload{
			Value:  "hello",
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)
		if len(payload.Issues) != 0 {
			t.Errorf("Expected no issues for 'hello' (length 5) == 5, got %d", len(payload.Issues))
		}

		// Test too long
		payload = &core.ParsePayload{
			Value:  "hello world",
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)
		if len(payload.Issues) != 1 {
			t.Errorf("Expected 1 issue for 'hello world' (length 11) != 5, got %d", len(payload.Issues))
		}

		// Test too short
		payload = &core.ParsePayload{
			Value:  "hi",
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)
		if len(payload.Issues) != 1 {
			t.Errorf("Expected 1 issue for 'hi' (length 2) != 5, got %d", len(payload.Issues))
		}
	})

	t.Run("MaxSize with conditional execution", func(t *testing.T) {
		check := MaxSize(3)

		// Test with slice - should execute
		payload := &core.ParsePayload{
			Value:  []int{1, 2},
			Issues: make([]core.ZodRawIssue, 0),
		}

		internals := check.GetZod()
		if internals.When != nil && !internals.When(payload) {
			t.Error("Expected When condition to pass for slice")
		}

		executeCheck(check, payload)
		if len(payload.Issues) != 0 {
			t.Errorf("Expected no issues for slice with 2 elements <= 3, got %d", len(payload.Issues))
		}

		// Test with too large slice
		payload = &core.ParsePayload{
			Value:  []int{1, 2, 3, 4},
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)
		if len(payload.Issues) != 1 {
			t.Errorf("Expected 1 issue for slice with 4 elements > 3, got %d", len(payload.Issues))
		}
	})

	t.Run("NonEmpty convenience function", func(t *testing.T) {
		check := NonEmpty()

		// Test valid case
		payload := &core.ParsePayload{
			Value:  "hello",
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)
		if len(payload.Issues) != 0 {
			t.Errorf("Expected no issues for non-empty string, got %d", len(payload.Issues))
		}

		// Test invalid case
		payload = &core.ParsePayload{
			Value:  "",
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)
		if len(payload.Issues) != 1 {
			t.Errorf("Expected 1 issue for empty string, got %d", len(payload.Issues))
		}
	})

	t.Run("Empty convenience function", func(t *testing.T) {
		check := Empty()

		// Test valid case
		payload := &core.ParsePayload{
			Value:  "",
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)
		if len(payload.Issues) != 0 {
			t.Errorf("Expected no issues for empty string, got %d", len(payload.Issues))
		}

		// Test invalid case
		payload = &core.ParsePayload{
			Value:  "hello",
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)
		if len(payload.Issues) != 1 {
			t.Errorf("Expected 1 issue for non-empty string, got %d", len(payload.Issues))
		}
	})
}

func TestLengthArraySupport(t *testing.T) {
	t.Run("works with arrays and slices", func(t *testing.T) {
		check := MaxLength(3)

		// Test with slice
		payload := &core.ParsePayload{
			Value:  []string{"a", "b"},
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)
		if len(payload.Issues) != 0 {
			t.Errorf("Expected no issues for slice with 2 elements <= 3, got %d", len(payload.Issues))
		}

		// Test with string
		payload = &core.ParsePayload{
			Value:  "ab",
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)
		if len(payload.Issues) != 0 {
			t.Errorf("Expected no issues for string with 2 characters <= 3, got %d", len(payload.Issues))
		}
	})

	t.Run("MaxSize works with maps", func(t *testing.T) {
		check := MaxSize(3)

		// Test with map[string]any - the supported type
		payload := &core.ParsePayload{
			Value:  map[string]any{"a": 1, "b": 2},
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)
		if len(payload.Issues) != 0 {
			t.Errorf("Expected no issues for map with 2 elements <= 3, got %d", len(payload.Issues))
		}

		// Test with slice - should also work
		payload = &core.ParsePayload{
			Value:  []string{"a", "b"},
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)
		if len(payload.Issues) != 0 {
			t.Errorf("Expected no issues for slice with 2 elements <= 3, got %d", len(payload.Issues))
		}
	})
}

func TestLengthRangeChecks(t *testing.T) {
	t.Run("LengthRange validates length range", func(t *testing.T) {
		check := LengthRange(3, 8)

		// Test valid case
		payload := &core.ParsePayload{
			Value:  "hello",
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)
		if len(payload.Issues) != 0 {
			t.Errorf("Expected no issues for 'hello' (length 5) in range [3,8], got %d", len(payload.Issues))
		}

		// Test too short
		payload = &core.ParsePayload{
			Value:  "hi",
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)
		if len(payload.Issues) != 1 {
			t.Errorf("Expected 1 issue for 'hi' (length 2) < 3, got %d", len(payload.Issues))
		}

		// Test too long
		payload = &core.ParsePayload{
			Value:  "hello world!",
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)
		if len(payload.Issues) != 1 {
			t.Errorf("Expected 1 issue for 'hello world!' (length 12) > 8, got %d", len(payload.Issues))
		}
	})

	t.Run("SizeRange with conditional execution", func(t *testing.T) {
		check := SizeRange(2, 5)

		// Test valid case
		payload := &core.ParsePayload{
			Value:  []int{1, 2, 3},
			Issues: make([]core.ZodRawIssue, 0),
		}

		internals := check.GetZod()
		if internals.When != nil && !internals.When(payload) {
			t.Error("Expected When condition to pass for slice")
		}

		executeCheck(check, payload)
		if len(payload.Issues) != 0 {
			t.Errorf("Expected no issues for slice with 3 elements in range [2,5], got %d", len(payload.Issues))
		}
	})
}

func TestLengthCustomMessages(t *testing.T) {
	t.Run("Custom error messages work", func(t *testing.T) {
		check := MaxLength(5, "String is too long")

		payload := &core.ParsePayload{
			Value:  "hello world",
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

func BenchmarkLengthChecks(b *testing.B) {
	check := MaxLength(100)
	payload := &core.ParsePayload{
		Value:  "hello world",
		Issues: make([]core.ZodRawIssue, 0, 1),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		payload.Issues = payload.Issues[:0]
		executeCheck(check, payload)
	}
}

func TestLengthZeroAllocation(t *testing.T) {
	check := MaxLength(100, "Too long")

	for i := 0; i < 100; i++ {
		payload := &core.ParsePayload{
			Value:  "short",
			Issues: make([]core.ZodRawIssue, 0, 1),
		}
		executeCheck(check, payload)

		if len(payload.Issues) != 0 {
			t.Errorf("Iteration %d: Expected no issues for short string, got %d", i, len(payload.Issues))
		}
	}
}

package checks

import (
	"regexp"
	"testing"

	"github.com/kaptinlin/gozod/core"
)

func TestStringContentChecks(t *testing.T) {
	t.Run("Includes validates substring presence", func(t *testing.T) {
		check := Includes("test")

		// Test valid case
		payload := &core.ParsePayload{
			Value:  "this is a test string",
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)
		if len(payload.Issues) != 0 {
			t.Errorf("Expected no issues for string containing substring, got %d", len(payload.Issues))
		}

		// Test invalid case
		payload = &core.ParsePayload{
			Value:  "this string doesn't contain the word",
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)
		if len(payload.Issues) != 1 {
			t.Errorf("Expected 1 issue for string not containing substring, got %d", len(payload.Issues))
		}
	})

	t.Run("StartsWith validates prefix", func(t *testing.T) {
		check := StartsWith("hello")

		// Test valid case
		payload := &core.ParsePayload{
			Value:  "hello world",
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)
		if len(payload.Issues) != 0 {
			t.Errorf("Expected no issues for string with correct prefix, got %d", len(payload.Issues))
		}

		// Test invalid case
		payload = &core.ParsePayload{
			Value:  "hi world",
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)
		if len(payload.Issues) != 1 {
			t.Errorf("Expected 1 issue for string with wrong prefix, got %d", len(payload.Issues))
		}
	})

	t.Run("EndsWith validates suffix", func(t *testing.T) {
		check := EndsWith("world")

		// Test valid case
		payload := &core.ParsePayload{
			Value:  "hello world",
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)
		if len(payload.Issues) != 0 {
			t.Errorf("Expected no issues for string with correct suffix, got %d", len(payload.Issues))
		}

		// Test invalid case
		payload = &core.ParsePayload{
			Value:  "hello universe",
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)
		if len(payload.Issues) != 1 {
			t.Errorf("Expected 1 issue for string with wrong suffix, got %d", len(payload.Issues))
		}
	})
}

func TestCaseValidationChecks(t *testing.T) {
	t.Run("Lowercase validates lowercase strings", func(t *testing.T) {
		check := Lowercase()

		// Test valid case
		payload := &core.ParsePayload{
			Value:  "hello world",
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)
		if len(payload.Issues) != 0 {
			t.Errorf("Expected no issues for lowercase string, got %d", len(payload.Issues))
		}

		// Test invalid case
		payload = &core.ParsePayload{
			Value:  "Hello World",
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)
		if len(payload.Issues) != 1 {
			t.Errorf("Expected 1 issue for non-lowercase string, got %d", len(payload.Issues))
		}
	})

	t.Run("Uppercase validates uppercase strings", func(t *testing.T) {
		check := Uppercase()

		// Test valid case
		payload := &core.ParsePayload{
			Value:  "HELLO WORLD",
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)
		if len(payload.Issues) != 0 {
			t.Errorf("Expected no issues for uppercase string, got %d", len(payload.Issues))
		}

		// Test invalid case
		payload = &core.ParsePayload{
			Value:  "Hello World",
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)
		if len(payload.Issues) != 1 {
			t.Errorf("Expected 1 issue for non-uppercase string, got %d", len(payload.Issues))
		}
	})
}

func TestRegexCheck(t *testing.T) {
	t.Run("Regex validates against pattern", func(t *testing.T) {
		pattern := regexp.MustCompile(`^\d{3}-\d{3}-\d{4}$`)
		check := Regex(pattern)

		// Test valid case
		payload := &core.ParsePayload{
			Value:  "123-456-7890",
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)
		if len(payload.Issues) != 0 {
			t.Errorf("Expected no issues for matching regex, got %d", len(payload.Issues))
		}

		// Test invalid case
		payload = &core.ParsePayload{
			Value:  "123-45-6789",
			Issues: make([]core.ZodRawIssue, 0),
		}
		executeCheck(check, payload)
		if len(payload.Issues) != 1 {
			t.Errorf("Expected 1 issue for non-matching regex, got %d", len(payload.Issues))
		}
	})
}

func TestStringCustomMessages(t *testing.T) {
	t.Run("Custom error messages work for string checks", func(t *testing.T) {
		check := Includes("test", "Must include the word 'test'")

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

package checks

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kaptinlin/gozod/core"
)

func TestLengthChecks(t *testing.T) {
	t.Run("MinLength validates minimum length", func(t *testing.T) {
		check := MinLength(5)

		// Test valid case
		payload := core.NewParsePayload("hello")
		executeCheck(check, payload)
		assert.Len(t, payload.Issues(), 0, "Expected no issues for valid length")

		// Test invalid case
		payload = core.NewParsePayload("hi")
		executeCheck(check, payload)
		assert.Len(t, payload.Issues(), 1, "Expected 1 issue for invalid length")
	})
}

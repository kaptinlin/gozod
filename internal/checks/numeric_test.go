package checks

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kaptinlin/gozod/core"
)

func TestNumericChecks(t *testing.T) {
	t.Run("Gte validates minimum value", func(t *testing.T) {
		check := Gte(5)

		// Test valid case
		payload := core.NewParsePayload(10)
		executeCheck(check, payload)
		assert.Len(t, payload.Issues(), 0, "Expected no issues for valid value")

		// Test invalid case
		payload = core.NewParsePayload(3)
		executeCheck(check, payload)
		assert.Len(t, payload.Issues(), 1, "Expected 1 issue for invalid value")
	})
}

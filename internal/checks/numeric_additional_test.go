package checks

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kaptinlin/gozod/core"
)

func TestNumericCheckConstructors_ValidateAndAnnotate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		check      core.ZodCheck
		valid      any
		invalid    any
		wantCode   core.IssueCode
		wantBound  string
		wantValue  any
		inclusive  bool
		wantBagKey string
	}{
		{name: "less than", check: Lt(5), valid: 4, invalid: 5, wantCode: core.TooBig, wantBound: "maximum", wantValue: 5, inclusive: false, wantBagKey: "exclusiveMaximum"},
		{name: "less than or equal", check: Lte(5), valid: 5, invalid: 6, wantCode: core.TooBig, wantBound: "maximum", wantValue: 5, inclusive: true, wantBagKey: "maximum"},
		{name: "greater than", check: Gt(5), valid: 6, invalid: 5, wantCode: core.TooSmall, wantBound: "minimum", wantValue: 5, inclusive: false, wantBagKey: "exclusiveMinimum"},
		{name: "greater than or equal", check: Gte(5), valid: 5, invalid: 4, wantCode: core.TooSmall, wantBound: "minimum", wantValue: 5, inclusive: true, wantBagKey: "minimum"},
		{name: "positive", check: Positive(), valid: 1, invalid: 0, wantCode: core.TooSmall, wantBound: "minimum", wantValue: 0, inclusive: false, wantBagKey: "exclusiveMinimum"},
		{name: "negative", check: Negative(), valid: -1, invalid: 0, wantCode: core.TooBig, wantBound: "maximum", wantValue: 0, inclusive: false, wantBagKey: "exclusiveMaximum"},
		{name: "non positive", check: NonPositive(), valid: 0, invalid: 1, wantCode: core.TooBig, wantBound: "maximum", wantValue: 0, inclusive: true, wantBagKey: "maximum"},
		{name: "non negative", check: NonNegative(), valid: 0, invalid: -1, wantCode: core.TooSmall, wantBound: "minimum", wantValue: 0, inclusive: true, wantBagKey: "minimum"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			requireCheckAccepts(t, tt.check, tt.valid)
			issue := requireCheckRejects(t, tt.check, tt.invalid, tt.wantCode)
			assert.Equal(t, tt.inclusive, issue.Inclusive())
			if tt.wantBound == "minimum" {
				assert.Equal(t, tt.wantValue, issue.Minimum())
			} else {
				assert.Equal(t, tt.wantValue, issue.Maximum())
			}

			schema := newCheckAttachSchema(core.ZodTypeNumber)
			bag := applyCheckAttach(t, tt.check, schema)
			assert.Equal(t, tt.wantValue, bag[tt.wantBagKey])
		})
	}
}

func TestMultipleOf_ValidatesAndAnnotates(t *testing.T) {
	t.Parallel()

	check := MultipleOf(3)
	requireCheckAccepts(t, check, 9)

	issue := requireCheckRejects(t, check, 10, core.NotMultipleOf)
	assert.Equal(t, 3, issue.Divisor())
	assert.Equal(t, "integer", issue.Origin())

	schema := newCheckAttachSchema(core.ZodTypeNumber)
	bag := applyCheckAttach(t, check, schema)
	assert.Equal(t, 3, bag["multipleOf"])
}

func TestNumericOnAttach_MergesStricterBounds(t *testing.T) {
	t.Parallel()

	schema := newCheckAttachSchema(core.ZodTypeNumber)
	applyCheckAttach(t, Gt(0), schema)
	applyCheckAttach(t, Gt(5), schema)
	applyCheckAttach(t, Lt(100), schema)
	bag := applyCheckAttach(t, Lt(10), schema)

	assert.Equal(t, 5, bag["exclusiveMinimum"])
	assert.Equal(t, 10, bag["exclusiveMaximum"])
}

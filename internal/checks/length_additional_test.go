package checks

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kaptinlin/gozod/core"
)

func TestLengthCheckConstructors_ValidateAndAnnotate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		check      core.ZodCheck
		valid      any
		invalid    any
		wantCode   core.IssueCode
		wantKey    string
		wantValue  any
		wantOrigin string
	}{
		{name: "max length", check: MaxLength(3), valid: "go", invalid: "gozod", wantCode: core.TooBig, wantKey: "maxLength", wantValue: 3, wantOrigin: "string"},
		{name: "min length", check: MinLength(3), valid: "gozod", invalid: "go", wantCode: core.TooSmall, wantKey: "minLength", wantValue: 3, wantOrigin: "string"},
		{name: "non empty", check: NonEmpty(), valid: "go", invalid: "", wantCode: core.TooSmall, wantKey: "minLength", wantValue: 1, wantOrigin: "string"},
		{name: "empty", check: Empty(), valid: "", invalid: "go", wantCode: core.TooBig, wantKey: "maxLength", wantValue: 0, wantOrigin: "string"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			requireCheckAccepts(t, tt.check, tt.valid)
			issue := requireCheckRejects(t, tt.check, tt.invalid, tt.wantCode)
			assert.Equal(t, tt.wantOrigin, issue.Origin())
			assert.True(t, issue.Inclusive())

			schema := newCheckAttachSchema(core.ZodTypeString)
			bag := applyCheckAttach(t, tt.check, schema)
			assert.Equal(t, tt.wantValue, bag[tt.wantKey])
		})
	}
}

func TestExactLength_ReportsTooSmallAndTooBig(t *testing.T) {
	t.Parallel()

	check := Length(3)
	requireCheckAccepts(t, check, "abc")

	tooSmall := requireCheckRejects(t, check, "go", core.TooSmall)
	assert.Equal(t, 3, tooSmall.Minimum())
	assert.Equal(t, "string", tooSmall.Origin())

	tooBig := requireCheckRejects(t, check, "gozod", core.TooBig)
	assert.Equal(t, 3, tooBig.Maximum())
	assert.Equal(t, "string", tooBig.Origin())

	schema := newCheckAttachSchema(core.ZodTypeString)
	bag := applyCheckAttach(t, check, schema)
	assert.Equal(t, 3, bag["minLength"])
	assert.Equal(t, 3, bag["maxLength"])
}

func TestLengthRange_ReportsOutsideBounds(t *testing.T) {
	t.Parallel()

	check := LengthRange(2, 4)
	requireCheckAccepts(t, check, "go")
	requireCheckAccepts(t, check, "gozo")

	tooSmall := requireCheckRejects(t, check, "g", core.TooSmall)
	assert.Equal(t, 2, tooSmall.Minimum())
	tooBig := requireCheckRejects(t, check, "gozod", core.TooBig)
	assert.Equal(t, 4, tooBig.Maximum())

	schema := newCheckAttachSchema(core.ZodTypeString)
	bag := applyCheckAttach(t, check, schema)
	assert.Equal(t, 2, bag["minLength"])
	assert.Equal(t, 4, bag["maxLength"])
}

func TestSizeChecks_ValidateCollectionsAndAnnotateArrayBounds(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		check      core.ZodCheck
		valid      any
		invalid    any
		wantCode   core.IssueCode
		wantKey    string
		wantValue  any
		wantOrigin string
	}{
		{name: "max size", check: MaxSize(2), valid: []int{1, 2}, invalid: []int{1, 2, 3}, wantCode: core.TooBig, wantKey: "maxItems", wantValue: 2, wantOrigin: "slice"},
		{name: "min size", check: MinSize(2), valid: []int{1, 2}, invalid: []int{1}, wantCode: core.TooSmall, wantKey: "minItems", wantValue: 2, wantOrigin: "slice"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			requireCheckAccepts(t, tt.check, tt.valid)
			issue := requireCheckRejects(t, tt.check, tt.invalid, tt.wantCode)
			assert.Equal(t, tt.wantOrigin, issue.Origin())

			schema := newCheckAttachSchema(core.ZodTypeSlice)
			schema.internals.Bag = map[string]any{"type": "array"}
			bag := applyCheckAttach(t, tt.check, schema)
			assert.Equal(t, tt.wantValue, bag[tt.wantKey])
		})
	}
}

func TestExactSize_ReportsTooSmallAndTooBig(t *testing.T) {
	t.Parallel()

	check := Size(2)
	requireCheckAccepts(t, check, []int{1, 2})

	tooSmall := requireCheckRejects(t, check, []int{1}, core.TooSmall)
	assert.Equal(t, 2, tooSmall.Minimum())
	assert.Equal(t, "slice", tooSmall.Origin())

	tooBig := requireCheckRejects(t, check, []int{1, 2, 3}, core.TooBig)
	assert.Equal(t, 2, tooBig.Maximum())
	assert.Equal(t, "slice", tooBig.Origin())

	schema := newCheckAttachSchema(core.ZodTypeSlice)
	schema.internals.Bag = map[string]any{"type": "array"}
	bag := applyCheckAttach(t, check, schema)
	assert.Equal(t, 2, bag["minItems"])
	assert.Equal(t, 2, bag["maxItems"])
}

func TestSizeRange_ReportsOutsideBounds(t *testing.T) {
	t.Parallel()

	check := SizeRange(2, 3)
	requireCheckAccepts(t, check, []int{1, 2})
	requireCheckAccepts(t, check, []int{1, 2, 3})

	tooSmall := requireCheckRejects(t, check, []int{1}, core.TooSmall)
	assert.Equal(t, 2, tooSmall.Minimum())
	tooBig := requireCheckRejects(t, check, []int{1, 2, 3, 4}, core.TooBig)
	assert.Equal(t, 3, tooBig.Maximum())

	schema := newCheckAttachSchema(core.ZodTypeSlice)
	schema.internals.Bag = map[string]any{"type": "array"}
	bag := applyCheckAttach(t, check, schema)
	assert.Equal(t, 2, bag["minItems"])
	assert.Equal(t, 3, bag["maxItems"])
}

func TestSizeChecks_SkipValuesWithoutSize(t *testing.T) {
	t.Parallel()

	payload := core.NewParsePayload(42)
	executeCheck(MinSize(1), payload)
	assert.Len(t, payload.Issues(), 0)
}

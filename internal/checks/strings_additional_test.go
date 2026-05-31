package checks

import (
	"regexp"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kaptinlin/gozod/core"
)

func TestStringRangeChecks_ValidateLexicographicBounds(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		check     core.ZodCheck
		valid     any
		invalid   any
		wantCode  core.IssueCode
		wantBound string
		wantValue any
	}{
		{name: "string greater than or equal", check: StringGte("2024-01-01"), valid: "2024-06-01", invalid: "2023-12-31", wantCode: core.TooSmall, wantBound: "minimum", wantValue: "2024-01-01"},
		{name: "string less than or equal", check: StringLte("2024-12-31"), valid: "2024-06-01", invalid: "2025-01-01", wantCode: core.TooBig, wantBound: "maximum", wantValue: "2024-12-31"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			requireCheckAccepts(t, tt.check, tt.valid)
			issue := requireCheckRejects(t, tt.check, tt.invalid, tt.wantCode)
			assert.Equal(t, "string", issue.Origin())
			if tt.wantBound == "minimum" {
				assert.Equal(t, tt.wantValue, issue.Minimum())
			} else {
				assert.Equal(t, tt.wantValue, issue.Maximum())
			}
		})
	}
}

func TestStringRangeChecks_RejectNonStringInput(t *testing.T) {
	t.Parallel()

	issue := requireCheckRejects(t, StringGte("a"), 123, core.InvalidType)
	assert.Equal(t, core.ZodTypeString, issue.Expected())

	issue = requireCheckRejects(t, StringLte("z"), 123, core.InvalidType)
	assert.Equal(t, core.ZodTypeString, issue.Expected())
}

func TestAddPatternToSchema_DeduplicatesPatterns(t *testing.T) {
	t.Parallel()

	schema := newCheckAttachSchema(core.ZodTypeString)
	for _, check := range []core.ZodCheck{
		Regex(regexp.MustCompile(`^[a-z]+$`)),
		Regex(regexp.MustCompile(`^[a-z]+$`)),
		Includes("go"),
	} {
		applyCheckAttach(t, check, schema)
	}

	patterns, ok := schema.internals.Bag["patterns"].([]string)
	require.True(t, ok)
	want := []string{`^[a-z]+$`, `go`}
	if diff := cmp.Diff(want, patterns); diff != "" {
		t.Errorf("patterns mismatch (-want +got):\n%s", diff)
	}
}

package tagparser_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kaptinlin/gozod/pkg/tagparser"
)

func TestCompileRule(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		rule       tagparser.TagRule
		wantOp     tagparser.RuleOp
		wantMethod string
	}{
		{
			name:   "structural",
			rule:   tagparser.TagRule{Name: "required"},
			wantOp: tagparser.RuleStructural,
		},
		{
			name:       "numeric method",
			rule:       tagparser.TagRule{Name: "gte", Params: []string{"1"}},
			wantOp:     tagparser.RuleMethod,
			wantMethod: "Gte",
		},
		{
			name:       "string check",
			rule:       tagparser.TagRule{Name: "email"},
			wantOp:     tagparser.RuleStringCheck,
			wantMethod: "Email",
		},
		{
			name:       "uuid constructor special",
			rule:       tagparser.TagRule{Name: "uuid"},
			wantOp:     tagparser.RuleStringCheck,
			wantMethod: "",
		},
		{
			name:       "default",
			rule:       tagparser.TagRule{Name: "default", Params: []string{"hello"}},
			wantOp:     tagparser.RuleDefault,
			wantMethod: "Default",
		},
		{
			name:   "unknown",
			rule:   tagparser.TagRule{Name: "mystery"},
			wantOp: tagparser.RuleUnsupported,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tagparser.CompileRule(tt.rule)
			assert.Equal(t, tt.wantOp, got.Op)
			assert.Equal(t, tt.wantMethod, got.Method)
		})
	}
}

func TestRulePlan_JoinedValue(t *testing.T) {
	t.Parallel()

	plan := tagparser.CompileRule(tagparser.TagRule{
		Name:   "prefault",
		Params: []string{"hello", "world"},
	})
	got, ok := plan.JoinedValue()

	assert.True(t, ok)
	assert.Equal(t, "hello world", got)
}

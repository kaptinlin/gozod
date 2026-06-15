package tagparser_test

import (
	"reflect"
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

func TestCompileFieldPlan(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                  string
		field                 tagparser.FieldInfo
		wantRuntimeOptional   tagparser.OptionalPlacement
		wantGeneratedOptional tagparser.OptionalPlacement
		wantOperationNames    []string
		wantOperationMethods  []string
	}{
		{
			name: "default makes generated optional inner",
			field: tagparser.FieldInfo{
				Type: reflect.TypeFor[string](),
				Rules: []tagparser.TagRule{
					{Name: "default", Params: []string{"guest"}},
				},
			},
			wantRuntimeOptional:   tagparser.OptionalPlacementNone,
			wantGeneratedOptional: tagparser.OptionalPlacementBeforeOperations,
			wantOperationNames:    []string{"default"},
			wantOperationMethods:  []string{"Default"},
		},
		{
			name: "pointer default makes runtime and generated optional inner",
			field: tagparser.FieldInfo{
				Type: reflect.TypeFor[*string](),
				Rules: []tagparser.TagRule{
					{Name: "default", Params: []string{"guest"}},
				},
			},
			wantRuntimeOptional:   tagparser.OptionalPlacementBeforeOperations,
			wantGeneratedOptional: tagparser.OptionalPlacementBeforeOperations,
			wantOperationNames:    []string{"default"},
			wantOperationMethods:  []string{"Default"},
		},
		{
			name: "plain optional stays outer",
			field: tagparser.FieldInfo{
				Type: reflect.TypeFor[string](),
				Rules: []tagparser.TagRule{
					{Name: "min", Params: []string{"2"}},
				},
			},
			wantRuntimeOptional:   tagparser.OptionalPlacementNone,
			wantGeneratedOptional: tagparser.OptionalPlacementAfterOperations,
			wantOperationNames:    []string{"min"},
			wantOperationMethods:  []string{"Min"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tagparser.CompileFieldPlan(&tt.field)

			assert.Equal(t, tt.wantRuntimeOptional, got.RuntimePointerOptional)
			assert.Equal(t, tt.wantGeneratedOptional, got.GeneratedOptional)
			assert.Len(t, got.Operations, len(tt.wantOperationNames))
			for i, operation := range got.Operations {
				assert.Equal(t, tt.wantOperationNames[i], operation.Rule.Name)
				assert.Equal(t, tt.wantOperationMethods[i], operation.Method)
			}
		})
	}
}

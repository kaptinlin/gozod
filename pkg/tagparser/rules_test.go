package tagparser_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kaptinlin/gozod/pkg/tagparser"
)

func TestCompileFieldPlan(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                  string
		field                 tagparser.FieldInfo
		wantRuntimeOptional   tagparser.OptionalPlacement
		wantGeneratedOptional tagparser.OptionalPlacement
		wantOperationNames    []string
		wantOperations        []tagparser.RuleOp
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
			wantOperations:        []tagparser.RuleOp{tagparser.RuleDefault},
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
			wantOperations:        []tagparser.RuleOp{tagparser.RuleDefault},
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
			wantOperations:        []tagparser.RuleOp{tagparser.RuleMin},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tagparser.CompileFieldPlan(&tt.field)
			assert.NoError(t, err)

			assert.Equal(t, tt.wantRuntimeOptional, got.RuntimePointerOptional)
			assert.Equal(t, tt.wantGeneratedOptional, got.GeneratedOptional)
			assert.Len(t, got.Operations, len(tt.wantOperationNames))
			for i, operation := range got.Operations {
				assert.Equal(t, tt.wantOperationNames[i], operation.Name)
				assert.Equal(t, tt.wantOperations[i], operation.Op)
			}
		})
	}
}

func TestCompileFieldPlanRejectsUnknownRuleWithFieldContext(t *testing.T) {
	t.Parallel()

	field := tagparser.FieldInfo{
		Name:     "Age",
		Type:     reflect.TypeFor[int](),
		GoZodTag: "mystery=1",
		Rules: []tagparser.TagRule{
			{Name: "mystery", Params: []string{"1"}},
		},
	}

	_, err := tagparser.CompileFieldPlan(&field)

	assert.Error(t, err)
	assert.ErrorContains(t, err, "Age")
	assert.ErrorContains(t, err, "mystery=1")
}

func TestCompileFieldPlanRejectsUndefinedUUIDVariant(t *testing.T) {
	t.Parallel()

	field := tagparser.FieldInfo{
		Name:     "ID",
		Type:     reflect.TypeFor[string](),
		GoZodTag: "uuid:v4",
		Rules:    []tagparser.TagRule{{Name: "uuid:v4"}},
	}

	_, err := tagparser.CompileFieldPlan(&field)

	assert.ErrorIs(t, err, tagparser.ErrUnknownRule)
	assert.ErrorContains(t, err, "ID")
	assert.ErrorContains(t, err, "uuid:v4")
}

func TestCompileFieldPlanRejectsMissingOperand(t *testing.T) {
	t.Parallel()

	field := tagparser.FieldInfo{
		Name:     "Age",
		Type:     reflect.TypeFor[int](),
		GoZodTag: "gte",
		Rules:    []tagparser.TagRule{{Name: "gte"}},
	}

	_, err := tagparser.CompileFieldPlan(&field)

	assert.Error(t, err)
	assert.ErrorContains(t, err, "Age")
	assert.ErrorContains(t, err, "gte")
}

func TestCompileFieldPlanRejectsExtraOperand(t *testing.T) {
	t.Parallel()

	field := tagparser.FieldInfo{
		Name:     "Age",
		Type:     reflect.TypeFor[int](),
		GoZodTag: "gte=1 2",
		Rules: []tagparser.TagRule{
			{Name: "gte", Params: []string{"1", "2"}},
		},
	}

	_, err := tagparser.CompileFieldPlan(&field)

	assert.Error(t, err)
	assert.ErrorContains(t, err, "Age")
	assert.ErrorContains(t, err, "gte=1 2")
}

func TestCompileFieldPlanRejectsInvalidNumericOperand(t *testing.T) {
	t.Parallel()

	field := tagparser.FieldInfo{
		Name:     "Age",
		Type:     reflect.TypeFor[int](),
		GoZodTag: "gte=adult",
		Rules: []tagparser.TagRule{
			{Name: "gte", Params: []string{"adult"}},
		},
	}

	_, err := tagparser.CompileFieldPlan(&field)

	assert.Error(t, err)
	assert.ErrorContains(t, err, "Age")
	assert.ErrorContains(t, err, "gte=adult")
}

func TestCompileFieldPlanRejectsNumericRuleOnBool(t *testing.T) {
	t.Parallel()

	field := tagparser.FieldInfo{
		Name:     "Enabled",
		Type:     reflect.TypeFor[bool](),
		GoZodTag: "gte=1",
		Rules: []tagparser.TagRule{
			{Name: "gte", Params: []string{"1"}},
		},
	}

	_, err := tagparser.CompileFieldPlan(&field)

	assert.Error(t, err)
	assert.ErrorContains(t, err, "Enabled")
	assert.ErrorContains(t, err, "gte=1")
}

func TestCompileFieldPlanRejectsInvalidRegex(t *testing.T) {
	t.Parallel()

	field := tagparser.FieldInfo{
		Name:     "Code",
		Type:     reflect.TypeFor[string](),
		GoZodTag: "regex=[",
		Rules: []tagparser.TagRule{
			{Name: "regex", Params: []string{"["}},
		},
	}

	_, err := tagparser.CompileFieldPlan(&field)

	assert.Error(t, err)
	assert.ErrorContains(t, err, "Code")
	assert.ErrorContains(t, err, "regex=[")
}

func TestCompileFieldPlanRejectsInvalidDefault(t *testing.T) {
	t.Parallel()

	field := tagparser.FieldInfo{
		Name:     "Enabled",
		Type:     reflect.TypeFor[bool](),
		GoZodTag: "default=maybe",
		Rules: []tagparser.TagRule{
			{Name: "default", Params: []string{"maybe"}},
		},
	}

	_, err := tagparser.CompileFieldPlan(&field)

	assert.Error(t, err)
	assert.ErrorContains(t, err, "Enabled")
	assert.ErrorContains(t, err, "default=maybe")
}

func TestCompileFieldPlanParsesNumericOperandForFieldFamily(t *testing.T) {
	t.Parallel()

	field := tagparser.FieldInfo{
		Name:     "Age",
		Type:     reflect.TypeFor[int](),
		GoZodTag: "gte=18",
		Rules: []tagparser.TagRule{
			{Name: "gte", Params: []string{"18"}},
		},
	}

	plan, err := tagparser.CompileFieldPlan(&field)

	assert.NoError(t, err)
	if assert.Len(t, plan.Operations, 1) {
		assert.Equal(t, tagparser.FieldFamilySignedInteger, plan.Operations[0].Family)
		assert.Equal(t, int64(18), plan.Operations[0].Operand)
	}
}

func TestCompileFieldPlanRejectsCoercionForCompositeFamilies(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		fieldType reflect.Type
	}{
		{name: "struct", fieldType: reflect.TypeFor[struct{ Value string }]()},
		{name: "slice", fieldType: reflect.TypeFor[[]string]()},
		{name: "map", fieldType: reflect.TypeFor[map[string]string]()},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			plan, err := tagparser.CompileFieldPlan(&tagparser.FieldInfo{
				Name:  "Value",
				Type:  test.fieldType,
				Rules: []tagparser.TagRule{{Name: "coerce"}},
			})

			assert.Empty(t, plan.Operations)
			assert.ErrorIs(t, err, tagparser.ErrInapplicableRule)
		})
	}
}

func TestCompileFieldPlanParsesStructuredDefault(t *testing.T) {
	t.Parallel()

	field := tagparser.FieldInfo{
		Name:     "Scores",
		Type:     reflect.TypeFor[[]int](),
		GoZodTag: "default=[1,2]",
		Rules: []tagparser.TagRule{
			{Name: "default", Params: []string{"[1,2]"}},
		},
	}

	plan, err := tagparser.CompileFieldPlan(&field)

	assert.NoError(t, err)
	if assert.Len(t, plan.Operations, 1) {
		assert.Equal(t, []int{1, 2}, plan.Operations[0].Operand)
	}
}

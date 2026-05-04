package core

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type internalsCheck struct {
	internals *ZodCheckInternals
}

func (c internalsCheck) Zod() *ZodCheckInternals {
	return c.internals
}

func TestZodTypeInternals_CloneCopiesMutableState(t *testing.T) {
	t.Parallel()

	check := internalsCheck{internals: &ZodCheckInternals{}}
	original := &ZodTypeInternals{
		Type:          ZodTypeString,
		Checks:        []ZodCheck{check},
		DefaultValue:  map[string][]int{"scores": {1, 2}},
		PrefaultValue: []string{"prefault"},
		Values:        map[any]struct{}{"a": {}},
		Bag:           map[string]any{"description": "original"},
	}

	cloned := original.Clone()
	require.NotSame(t, original, cloned)

	original.Checks = append(original.Checks, check)
	original.DefaultValue.(map[string][]int)["scores"][0] = 99
	original.PrefaultValue.([]string)[0] = "changed"
	original.Values["b"] = struct{}{}
	original.Bag["description"] = "changed"

	assert.Len(t, cloned.Checks, 1)
	assert.Contains(t, cloned.Values, "a")
	assert.NotContains(t, cloned.Values, "b")
	assert.Equal(t, "original", cloned.Bag["description"])

	wantDefault := map[string][]int{"scores": {1, 2}}
	if diff := cmp.Diff(wantDefault, cloned.DefaultValue); diff != "" {
		t.Errorf("Clone() default mismatch (-want +got):\n%s", diff)
	}
	wantPrefault := []string{"prefault"}
	if diff := cmp.Diff(wantPrefault, cloned.PrefaultValue); diff != "" {
		t.Errorf("Clone() prefault mismatch (-want +got):\n%s", diff)
	}
}

func TestZodTypeInternals_ModifierFlags(t *testing.T) {
	t.Parallel()

	internals := &ZodTypeInternals{}
	internals.SetOptional(true)
	internals.SetNilable(true)
	internals.SetNonOptional(true)
	internals.SetExactOptional(true)
	internals.SetCoerce(true)
	internals.SetDefaultValue("default")
	internals.SetDefaultFunc(func() any { return "computed" })
	internals.SetPrefaultValue("prefault")
	internals.SetPrefaultFunc(func() any { return "computed prefault" })
	internals.SetTransform(func(value any, ctx *RefinementContext) (any, error) { return value, nil })
	internals.AddCheck(internalsCheck{internals: &ZodCheckInternals{}})

	assert.True(t, internals.IsOptional())
	assert.True(t, internals.IsNilable())
	assert.True(t, internals.IsNonOptional())
	assert.True(t, internals.IsExactOptional())
	assert.True(t, internals.IsCoerce())
	assert.Equal(t, "default", internals.DefaultValue)
	assert.Equal(t, "computed", internals.DefaultFunc())
	assert.Equal(t, "prefault", internals.PrefaultValue)
	assert.Equal(t, "computed prefault", internals.PrefaultFunc())
	require.NotNil(t, internals.Transform)
	got, err := internals.Transform("value", NewRefinementContext(nil, "value"))
	require.NoError(t, err)
	assert.Equal(t, "value", got)
	assert.Len(t, internals.Checks, 1)
}

func TestConvertToZodSchema(t *testing.T) {
	t.Parallel()

	schema := newRegistrySchema()
	got, err := ConvertToZodSchema(schema)
	require.NoError(t, err)
	assert.Same(t, schema, got)

	_, err = ConvertToZodSchema("not a schema")
	require.ErrorIs(t, err, ErrSchemaNotZodSchema)
}

func TestZodCheckInternals_Zod(t *testing.T) {
	t.Parallel()

	check := &ZodCheckInternals{}
	assert.Same(t, check, check.Zod())
}

func TestRefinementContext_IssuesAndJoinedError(t *testing.T) {
	t.Parallel()

	ctx := &ParseContext{ReportInput: true}
	refinement := NewRefinementContext(ctx, "value")
	assert.Same(t, ctx, refinement.ParseContext)
	assert.Equal(t, "value", refinement.Value)
	assert.NoError(t, refinement.Err())

	refinement.AddIssue(ZodIssue{ZodIssueBase: ZodIssueBase{Message: "first"}})
	refinement.AddIssue(ZodIssue{ZodIssueBase: ZodIssueBase{Message: "second"}})

	issues := refinement.Issues()
	require.Len(t, issues, 2)
	issues[0].Message = "mutated"
	assert.Equal(t, "first", refinement.Issues()[0].Message)

	err := refinement.Err()
	require.Error(t, err)
	var joined interface{ Unwrap() []error }
	require.ErrorAs(t, err, &joined)
	assert.Len(t, joined.Unwrap(), 2)
}

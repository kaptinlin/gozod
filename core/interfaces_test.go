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
		Type:   ZodTypeString,
		Checks: []ZodCheck{check},
		Values: map[any]struct{}{"a": {}},
		Bag:    map[string]any{"description": "original", "patterns": []string{"^[a-z]+$"}},
	}
	original.SetDefaultValue(map[string][]int{"scores": {1, 2}})
	original.SetPrefaultValue([]string{"prefault"})

	cloned := original.Clone()
	require.NotSame(t, original, cloned)

	original.Checks = append(original.Checks, check)
	original.Modifiers[0].Value.(map[string][]int)["scores"][0] = 99
	original.Modifiers[1].Value.([]string)[0] = "changed"
	original.Values["b"] = struct{}{}
	original.Bag["description"] = "changed"
	original.Bag["patterns"].([]string)[0] = "changed"

	assert.Len(t, cloned.Checks, 1)
	assert.Contains(t, cloned.Values, "a")
	assert.NotContains(t, cloned.Values, "b")
	assert.Equal(t, "original", cloned.Bag["description"])
	assert.Equal(t, []string{"^[a-z]+$"}, cloned.Bag["patterns"])

	wantDefault := map[string][]int{"scores": {1, 2}}
	if diff := cmp.Diff(wantDefault, cloned.Modifiers[0].Value); diff != "" {
		t.Errorf("Clone() default mismatch (-want +got):\n%s", diff)
	}
	wantPrefault := []string{"prefault"}
	if diff := cmp.Diff(wantPrefault, cloned.Modifiers[1].Value); diff != "" {
		t.Errorf("Clone() prefault mismatch (-want +got):\n%s", diff)
	}
}

func TestZodTypeInternals_ModifierFlags(t *testing.T) {
	t.Parallel()

	optional := &ZodTypeInternals{}
	optional.SetOptional(true)
	assert.True(t, optional.IsOptional())

	nilable := &ZodTypeInternals{}
	nilable.SetNilable(true)
	assert.True(t, nilable.IsNilable())

	nonoptional := &ZodTypeInternals{}
	nonoptional.SetNonOptional(true)
	assert.True(t, nonoptional.IsNonOptional())

	exactOptional := &ZodTypeInternals{}
	exactOptional.SetExactOptional(true)
	assert.True(t, exactOptional.IsExactOptional())
	assert.True(t, exactOptional.IsOptional())

	internals := &ZodTypeInternals{}
	internals.SetCoerce(true)
	internals.SetDefaultValue("default")
	internals.SetDefaultFunc(func() any { return "computed" })
	internals.SetPrefaultValue("prefault")
	internals.SetPrefaultFunc(func() any { return "computed prefault" })
	internals.SetTransform(func(value any, ctx *RefinementContext) (any, error) { return value, nil })
	internals.AddCheck(internalsCheck{internals: &ZodCheckInternals{}})

	assert.True(t, internals.IsCoerce())
	require.Len(t, internals.Modifiers, 4)
	assert.Equal(t, "default", internals.Modifiers[0].Value)
	assert.Equal(t, "computed", internals.Modifiers[1].Func())
	assert.Equal(t, "prefault", internals.Modifiers[2].Value)
	assert.Equal(t, "computed prefault", internals.Modifiers[3].Func())
	require.NotNil(t, internals.Transform)
	got, err := internals.Transform("value", NewRefinementContext(nil, "value"))
	require.NoError(t, err)
	assert.Equal(t, "value", got)
	assert.Len(t, internals.Checks, 1)
}

func TestZodTypeInternals_IsOptionalUsesOrderedModifiers(t *testing.T) {
	t.Parallel()

	internals := &ZodTypeInternals{
		Modifiers: []ZodModifier{{Kind: ZodModifierOptional}},
	}

	assert.True(t, internals.IsOptional())
}

func TestZodTypeInternals_IsNilableUsesOrderedModifiers(t *testing.T) {
	t.Parallel()

	internals := &ZodTypeInternals{
		Modifiers: []ZodModifier{{Kind: ZodModifierNilable}},
	}

	assert.True(t, internals.IsNilable())
}

func TestZodTypeInternals_IsNonOptionalUsesOrderedModifiers(t *testing.T) {
	t.Parallel()

	internals := &ZodTypeInternals{
		Modifiers: []ZodModifier{{Kind: ZodModifierNonOptional}},
	}

	assert.True(t, internals.IsNonOptional())
}

func TestZodTypeInternals_IsExactOptionalUsesOrderedModifiers(t *testing.T) {
	t.Parallel()

	internals := &ZodTypeInternals{
		Modifiers: []ZodModifier{{Kind: ZodModifierExactOptional}},
	}

	assert.True(t, internals.IsExactOptional())
}

func TestZodTypeInternals_PresenceQueriesFollowModifierOrder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		modifiers         []ZodModifier
		wantOptional      bool
		wantNilable       bool
		wantNonOptional   bool
		wantExactOptional bool
	}{
		{
			name:            "outer nonoptional overrides optional",
			modifiers:       []ZodModifier{{Kind: ZodModifierOptional}, {Kind: ZodModifierNonOptional}},
			wantNonOptional: true,
		},
		{
			name:         "outer optional overrides nonoptional",
			modifiers:    []ZodModifier{{Kind: ZodModifierNonOptional}, {Kind: ZodModifierOptional}},
			wantOptional: true,
		},
		{
			name:        "outer nilable overrides nonoptional",
			modifiers:   []ZodModifier{{Kind: ZodModifierNonOptional}, {Kind: ZodModifierNilable}},
			wantNilable: true,
		},
		{
			name:              "outer exact optional overrides optional",
			modifiers:         []ZodModifier{{Kind: ZodModifierOptional}, {Kind: ZodModifierExactOptional}},
			wantOptional:      true,
			wantExactOptional: true,
		},
		{
			name:         "outer optional overrides exact optional",
			modifiers:    []ZodModifier{{Kind: ZodModifierExactOptional}, {Kind: ZodModifierOptional}},
			wantOptional: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			internals := &ZodTypeInternals{Modifiers: tt.modifiers}
			assert.Equal(t, tt.wantOptional, internals.IsOptional())
			assert.Equal(t, tt.wantNilable, internals.IsNilable())
			assert.Equal(t, tt.wantNonOptional, internals.IsNonOptional())
			assert.Equal(t, tt.wantExactOptional, internals.IsExactOptional())
		})
	}
}

func TestZodTypeInternals_FallbackQueriesFollowModifierOrder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		modifiers    []ZodModifier
		wantDefault  bool
		wantFallback bool
	}{
		{
			name: "outer default claims nil",
			modifiers: []ZodModifier{
				{Kind: ZodModifierOptional},
				{Kind: ZodModifierDefault, Value: "default", HasValue: true},
			},
			wantDefault:  true,
			wantFallback: true,
		},
		{
			name: "outer optional shadows default",
			modifiers: []ZodModifier{
				{Kind: ZodModifierDefault, Value: "default", HasValue: true},
				{Kind: ZodModifierOptional},
			},
		},
		{
			name: "outer prefault claims nil",
			modifiers: []ZodModifier{
				{Kind: ZodModifierOptional},
				{Kind: ZodModifierPrefault, Value: "prefault", HasValue: true},
			},
			wantFallback: true,
		},
		{
			name: "outer optional shadows prefault",
			modifiers: []ZodModifier{
				{Kind: ZodModifierPrefault, Value: "prefault", HasValue: true},
				{Kind: ZodModifierOptional},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			internals := &ZodTypeInternals{Modifiers: tt.modifiers}
			assert.Equal(t, tt.wantDefault, internals.NilInputUsesDefault())
			assert.Equal(t, tt.wantFallback, internals.NilInputUsesFallback())
		})
	}
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

package presencefixture

import (
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kaptinlin/gozod"
)

func TestGeneratedPresenceShapeMatchesRuntime(t *testing.T) {
	runtime := gozod.MustFromStruct[Presence]()
	generated := Presence{}.Schema()

	for name, runtimeField := range runtime.Shape() {
		generatedField := generated.Shape()[name]
		require.NotNil(t, generatedField, "field %s", name)
		assert.Equal(t, runtimeField.IsOptional(), generatedField.IsOptional(), "field %s optional", name)
		assert.Equal(t, runtimeField.IsNilable(), generatedField.IsNilable(), "field %s nilable", name)
	}
}

func TestGeneratedPresenceMapOutputAndIssuesMatchRuntime(t *testing.T) {
	runtime := gozod.Object(gozod.MustFromStruct[Presence]().Shape())
	generated := gozod.Object(Presence{}.Schema().Shape())
	value := "valid"
	input := map[string]any{
		"required_value":            "valid",
		"required_pointer":          &value,
		"required_nilable_value":    "valid",
		"required_nilable_pointer":  nil,
		"required_default_value":    "valid",
		"required_default_pointer":  &value,
		"required_prefault_value":   "valid",
		"required_prefault_pointer": &value,
	}

	runtimeResult, err := runtime.Parse(input)
	require.NoError(t, err)
	generatedResult, err := generated.Parse(input)
	require.NoError(t, err)
	if diff := cmp.Diff(runtimeResult, generatedResult); diff != "" {
		t.Fatalf("map output mismatch (-runtime +generated):\n%s", diff)
	}

	invalid := map[string]any{"required_value": "x"}
	_, runtimeErr := runtime.Parse(invalid)
	_, generatedErr := generated.Parse(invalid)
	runtimeZod, ok := errors.AsType[*gozod.ZodError](runtimeErr)
	require.True(t, ok)
	generatedZod, ok := errors.AsType[*gozod.ZodError](generatedErr)
	require.True(t, ok)
	sortIssues := cmpopts.SortSlices(func(a, b gozod.ZodIssue) bool {
		aKey := fmt.Sprintf("%v/%s/%s", a.Path, a.Code, a.Message)
		bKey := fmt.Sprintf("%v/%s/%s", b.Path, b.Code, b.Message)
		return aKey < bKey
	})
	if diff := cmp.Diff(runtimeZod.Issues, generatedZod.Issues, sortIssues); diff != "" {
		t.Fatalf("issues mismatch (-runtime +generated):\n%s", diff)
	}
}

func TestGeneratedPresenceStructOutputMatchesRuntime(t *testing.T) {
	value := "valid"
	input := Presence{
		RequiredValue:           value,
		OptionalValue:           value,
		RequiredPointer:         &value,
		OptionalPointer:         &value,
		RequiredNilableValue:    value,
		OptionalNilableValue:    value,
		RequiredNilablePointer:  &value,
		OptionalNilablePointer:  &value,
		RequiredDefaultValue:    value,
		OptionalDefaultValue:    value,
		RequiredDefaultPointer:  &value,
		OptionalDefaultPointer:  &value,
		RequiredPrefaultValue:   value,
		OptionalPrefaultValue:   value,
		RequiredPrefaultPointer: &value,
		OptionalPrefaultPointer: &value,
	}

	runtimeResult, err := gozod.MustFromStruct[Presence]().Parse(input)
	require.NoError(t, err)
	generatedResult, err := Presence{}.Schema().Parse(input)
	require.NoError(t, err)
	if diff := cmp.Diff(runtimeResult, generatedResult); diff != "" {
		t.Fatalf("struct output mismatch (-runtime +generated):\n%s", diff)
	}
}

func TestGeneratedPresenceJSONSchemaRequiredMatchesRuntime(t *testing.T) {
	runtime, err := gozod.ToJSONSchema(gozod.MustFromStruct[Presence]())
	require.NoError(t, err)
	generated, err := gozod.ToJSONSchema(Presence{}.Schema())
	require.NoError(t, err)
	assert.ElementsMatch(t, runtime.Required, generated.Required)
}

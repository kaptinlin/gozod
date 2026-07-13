package coercionfixture

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kaptinlin/gozod"
)

func TestGeneratedCoercionMatchesRuntime(t *testing.T) {
	runtimeSchema := gozod.MustFromStruct[CoerceStruct]()
	generatedSchema := CoerceStruct{}.Schema()
	runtimeObject := gozod.Object(runtimeSchema.Shape())
	generatedObject := gozod.Object(generatedSchema.Shape())

	input := map[string]any{
		"name":   42,
		"email":  []byte("user@example.com"),
		"age":    "7",
		"active": "true",
		"score":  "1.5",
	}
	runtimeResult, err := runtimeObject.Parse(input)
	require.NoError(t, err)
	generatedResult, err := generatedObject.Parse(input)
	require.NoError(t, err)
	if diff := cmp.Diff(runtimeResult, generatedResult); diff != "" {
		t.Errorf("generated coercion result mismatch (-runtime +generated):\n%s", diff)
	}

	invalid := map[string]any{
		"name":   "name",
		"email":  "user@example.com",
		"age":    "invalid",
		"active": true,
		"score":  1.5,
	}
	_, runtimeErr := runtimeObject.Parse(invalid)
	_, generatedErr := generatedObject.Parse(invalid)
	runtimeZod, ok := errors.AsType[*gozod.ZodError](runtimeErr)
	require.True(t, ok)
	generatedZod, ok := errors.AsType[*gozod.ZodError](generatedErr)
	require.True(t, ok)
	require.NotEmpty(t, runtimeZod.Issues)
	require.NotEmpty(t, generatedZod.Issues)
	assert.Equal(t, runtimeZod.Issues[0].Code, generatedZod.Issues[0].Code)
	if diff := cmp.Diff(runtimeZod.Issues[0].Path, generatedZod.Issues[0].Path); diff != "" {
		t.Errorf("generated coercion issue path mismatch (-runtime +generated):\n%s", diff)
	}
}

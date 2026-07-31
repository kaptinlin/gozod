package namedscalarfixture

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kaptinlin/gozod"
	"github.com/kaptinlin/gozod/core"
)

func TestGeneratedNamedScalarsPreserveDeclaredOutputTypes(t *testing.T) {
	pointer := UserID("pointer")
	input := Scalars{
		ID:           UserID("user-1"),
		Count:        Counter(7),
		Active:       Enabled(true),
		Alias:        "label",
		Pointer:      &pointer,
		Slice:        []UserID{"one", "two"},
		Map:          map[string]Counter{"visits": 3},
		Array:        [2]UserID{"left", "right"},
		DefaultSlice: []UserID{"provided"},
		PrefaultMap:  map[string]UserID{"provided": "user-2"},
	}

	runtimeResult, err := gozod.MustFromStruct[Scalars]().Parse(input)
	require.NoError(t, err)
	generatedResult, err := input.Schema().Parse(input)
	require.NoError(t, err)
	assert.Equal(t, runtimeResult, generatedResult)
	assert.Equal(t, input, generatedResult)
}

func TestGeneratedNamedScalarsUseUnderlyingValidationFamilies(t *testing.T) {
	pointer := UserID("pointer")
	input := Scalars{
		ID:           UserID("x"),
		Count:        Counter(0),
		Active:       Enabled(true),
		Alias:        "label",
		Pointer:      &pointer,
		Slice:        []UserID{"one"},
		Map:          map[string]Counter{"visits": 1},
		Array:        [2]UserID{"left", "right"},
		DefaultSlice: []UserID{"provided"},
		PrefaultMap:  map[string]UserID{"provided": "user-2"},
	}

	_, runtimeErr := gozod.MustFromStruct[Scalars]().Parse(input)
	_, generatedErr := input.Schema().Parse(input)

	require.Error(t, runtimeErr)
	require.Error(t, generatedErr)
	runtimeZod, ok := errors.AsType[*gozod.ZodError](runtimeErr)
	require.True(t, ok)
	generatedZod, ok := errors.AsType[*gozod.ZodError](generatedErr)
	require.True(t, ok)
	assert.Equal(t, gozod.FlattenError(runtimeZod), gozod.FlattenError(generatedZod))
}

func TestGeneratedNamedScalarSliceDefaultMatchesRuntime(t *testing.T) {
	pointer := UserID("pointer")
	input := map[string]any{
		"id":      UserID("user-1"),
		"count":   Counter(7),
		"active":  Enabled(true),
		"alias":   "label",
		"pointer": &pointer,
		"slice":   []UserID{"one"},
		"map":     map[string]Counter{"visits": 3},
		"array":   [2]UserID{"left", "right"},
	}

	runtime := gozod.Object(gozod.MustFromStruct[Scalars]().Shape())
	generated := gozod.Object((Scalars{}).Schema().Shape())
	runtimeResult, err := runtime.Parse(input)
	require.NoError(t, err)
	generatedResult, err := generated.Parse(input)
	require.NoError(t, err)
	defaultSlice := []string{"primary", "backup"}
	prefaultMap := map[string]string{"primary": "user-1"}
	assert.Equal(t, &defaultSlice, runtimeResult["default_slice"])
	assert.Equal(t, &prefaultMap, runtimeResult["prefault_map"])
	assert.Equal(t, runtimeResult, generatedResult)
}

func TestStructReturnsErrorForUnassignableParsedField(t *testing.T) {
	type target struct {
		ID UserID
	}
	schema := gozod.Struct[target](gozod.StructSchema{
		"ID": gozod.String().Transform(func(string, *core.RefinementContext) (any, error) {
			return struct{}{}, nil
		}),
	})

	assert.NotPanics(t, func() {
		_, err := schema.Parse(target{ID: "user-1"})
		require.Error(t, err)
		assert.ErrorContains(t, err, "Failed to set field ID")
	})
}

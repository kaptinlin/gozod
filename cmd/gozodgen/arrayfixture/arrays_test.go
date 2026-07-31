package arrayfixture

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kaptinlin/gozod"
)

func TestGeneratedFixedArrayLengthMatchesRuntime(t *testing.T) {
	runtime := gozod.Object(gozod.MustFromStruct[Arrays]().Shape())
	generated := gozod.Object(Arrays{}.Schema().Shape())

	for _, length := range []int{2, 3, 4} {
		t.Run(string(rune('0'+length)), func(t *testing.T) {
			values := make([]string, length)
			input := map[string]any{"fixed": values, "dynamic": values}

			_, runtimeErr := runtime.Parse(input)
			_, generatedErr := generated.Parse(input)
			assert.Equal(t, length != 3, runtimeErr != nil)
			assert.Equal(t, runtimeErr != nil, generatedErr != nil)
		})
	}
}

func TestGeneratedFixedArrayTypedOutputMatchesRuntime(t *testing.T) {
	defaultPointer := [2]string{"provided", "default"}
	prefaultPointer := [2]string{"provided", "prefault"}
	input := Arrays{
		Fixed:           [3]string{"a", "b", "c"},
		Dynamic:         []string{"a", "b", "c", "d"},
		DefaultValue:    [2]string{"value", "default"},
		DefaultPointer:  &defaultPointer,
		PrefaultValue:   [2]string{"value", "prefault"},
		PrefaultPointer: &prefaultPointer,
	}

	runtimeResult, err := gozod.MustFromStruct[Arrays]().Parse(input)
	require.NoError(t, err)
	generatedResult, err := input.Schema().Parse(input)
	require.NoError(t, err)
	assert.Equal(t, runtimeResult, generatedResult)
	assert.Equal(t, input, generatedResult)
}

func TestGeneratedFixedArrayFallbacksMatchRuntime(t *testing.T) {
	runtime := gozod.Object(gozod.MustFromStruct[Arrays]().Shape())
	generated := gozod.Object(Arrays{}.Schema().Shape())
	input := map[string]any{
		"fixed":   []string{"a", "b", "c"},
		"dynamic": []string{"value"},
	}

	runtimeResult, err := runtime.Parse(input)
	require.NoError(t, err)
	generatedResult, err := generated.Parse(input)
	require.NoError(t, err)

	defaultValue := []string{"left", "right"}
	defaultPointer := []string{"north", "south"}
	prefaultValue := []string{"up", "down"}
	prefaultPointer := []string{"east", "west"}
	assert.Equal(t, &defaultValue, runtimeResult["default_value"])
	assert.Equal(t, &defaultPointer, runtimeResult["default_pointer"])
	assert.Equal(t, &prefaultValue, runtimeResult["prefault_value"])
	assert.Equal(t, &prefaultPointer, runtimeResult["prefault_pointer"])
	assert.Equal(t, runtimeResult, generatedResult)
}

func TestGeneratedCompositeFixedArrayOutputMatchesRuntime(t *testing.T) {
	pointer := [2]string{"left", "right"}
	input := CompositeArrays{
		Nested:         [2][3]string{{"a", "b", "c"}, {"d", "e", "f"}},
		Pointer:        &pointer,
		DefaultNested:  [2][3]string{{"provided", "default", "value"}},
		PrefaultNested: [2][3]string{{"provided", "prefault", "value"}},
	}

	runtimeResult, err := gozod.MustFromStruct[CompositeArrays]().Parse(input)
	require.NoError(t, err)
	generatedResult, err := input.Schema().Parse(input)
	require.NoError(t, err)
	assert.Equal(t, runtimeResult, generatedResult)
	assert.Equal(t, input, generatedResult)
}

func TestGeneratedCompositeFixedArrayFallbacksMatchRuntime(t *testing.T) {
	runtime := gozod.Object(gozod.MustFromStruct[CompositeArrays]().Shape())
	generated := gozod.Object((CompositeArrays{}).Schema().Shape())
	input := map[string]any{
		"nested":  [][]string{{"provided", "", ""}, {"", "", ""}},
		"pointer": []string{"left", "right"},
	}

	runtimeResult, err := runtime.Parse(input)
	require.NoError(t, err)
	generatedResult, err := generated.Parse(input)
	require.NoError(t, err)
	defaultValue := []any{
		[3]string{"a", "b", "c"},
		[3]string{"d", "e", "f"},
	}
	prefaultValue := []any{
		[3]string{"g", "h", "i"},
		[3]string{"j", "k", "l"},
	}
	assert.Equal(t, &defaultValue, runtimeResult["default_nested"])
	assert.Equal(t, &prefaultValue, runtimeResult["prefault_nested"])
	assert.Equal(t, runtimeResult, generatedResult)
}

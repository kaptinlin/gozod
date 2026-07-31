package anyfixture

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kaptinlin/gozod"
)

func TestGeneratedExplicitAnyFieldsPreserveDynamicValues(t *testing.T) {
	input := Values{
		Any:   map[string]any{"enabled": true},
		Empty: []any{"one", 2},
	}

	runtimeResult, err := gozod.MustFromStruct[Values]().Parse(input)
	require.NoError(t, err)
	generatedResult, err := input.Schema().Parse(input)
	require.NoError(t, err)
	assert.Equal(t, runtimeResult, generatedResult)
	assert.Equal(t, input, generatedResult)
}

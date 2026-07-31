package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/types"
)

type presenceValueDefault struct {
	Value string `json:"value" gozod:"default=guest"`
}

func TestPresenceParity_ValueDefaultIsOptional(t *testing.T) {
	runtime, err := types.FromStruct[presenceValueDefault]()
	require.NoError(t, err)
	generated := types.Struct[presenceValueDefault](core.StructSchema{
		"value": types.String().Optional().Default("guest"),
	})

	runtimeField := runtime.Shape()["value"]
	generatedField := generated.Shape()["value"]
	assert.Equal(t, generatedField.IsOptional(), runtimeField.IsOptional())
}

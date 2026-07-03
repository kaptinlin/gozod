package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kaptinlin/gozod/core"
)

func TestCompositeConstructorsDoNotExposeStructuralChecks(t *testing.T) {
	cases := []struct {
		name   string
		schema core.ZodSchema
	}{
		{name: "map", schema: Map(String(), Int())},
		{name: "object", schema: Object(core.ObjectSchema{"name": String()})},
		{name: "record", schema: Record(String(), Int())},
		{name: "set", schema: Set[string](String())},
		{name: "slice", schema: Slice[string](String())},
		{name: "tuple", schema: Tuple(String(), Int())},
		{name: "union", schema: UnionOf(String(), Int())},
		{name: "xor", schema: XorOf(String(), Int())},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			assert.Empty(t, tt.schema.Internals().Checks)
		})
	}
}

func TestCompositeStructuralValidationRunsWithoutExposedChecks(t *testing.T) {
	schema := Slice[string](String().Min(2))

	_, err := schema.Parse([]string{"x"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "Too small")
}

package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFunction_TypeSpecificInternals(t *testing.T) {
	t.Run("Input method sets input schema", func(t *testing.T) {
		schema := Function().Input(nil)
		require.NotNil(t, schema)
		assert.Nil(t, schema.internals.Input)
	})

	t.Run("Output method sets output schema", func(t *testing.T) {
		schema := Function().Output(nil)
		require.NotNil(t, schema)
		assert.Nil(t, schema.internals.Output)
	})

	t.Run("Input and Output chaining", func(t *testing.T) {
		schema := Function().Input(nil).Output(nil)
		require.NotNil(t, schema)
		assert.Nil(t, schema.internals.Input)
		assert.Nil(t, schema.internals.Output)
	})

	t.Run("function parameters validation", func(t *testing.T) {
		schema := Function(FunctionParams{Input: nil, Output: nil})
		require.NotNil(t, schema)
		assert.Nil(t, schema.internals.Input)
		assert.Nil(t, schema.internals.Output)
	})
}

package types

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kaptinlin/gozod/core"
)

func TestModifierOrder_NilInputMatrix(t *testing.T) {
	t.Run("primitive outer default shadows inner optional", func(t *testing.T) {
		schema := String().Optional().Default("fallback")

		got, err := schema.Parse(nil)

		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "fallback", *got)
	})

	t.Run("primitive outer optional shadows inner default", func(t *testing.T) {
		schema := String().Default("fallback").Optional()

		got, err := schema.Parse(nil)

		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("primitive outer nonoptional rejects inner optional", func(t *testing.T) {
		schema := String().Optional().NonOptional()

		_, err := schema.Parse(nil)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "nonoptional")
	})

	t.Run("primitive outer optional accepts inner nonoptional", func(t *testing.T) {
		schema := String().NonOptional().Optional()

		got, err := schema.Parse(nil)

		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("primitive nilable accepts nil", func(t *testing.T) {
		schema := String().Nilable()

		got, err := schema.Parse(nil)

		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("primitive nullish accepts nil", func(t *testing.T) {
		schema := String().Nullish()

		got, err := schema.Parse(nil)

		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("primitive prefault validates fallback", func(t *testing.T) {
		schema := String().Min(5).Prefault("bad")

		_, err := schema.Parse(nil)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "Too small")
	})

	t.Run("composite outer optional shadows inner default", func(t *testing.T) {
		schema := Slice[string](String()).Default([]string{"fallback"}).Optional()

		got, err := schema.Parse(nil)

		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("composite outer default shadows inner optional", func(t *testing.T) {
		schema := Slice[string](String()).Optional().Default([]string{"fallback"})

		got, err := schema.Parse(nil)

		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, []string{"fallback"}, *got)
	})

	t.Run("composite nullish accepts nil", func(t *testing.T) {
		schema := Slice[string](String()).Nullish()

		got, err := schema.Parse(nil)

		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("bigint outer optional shadows inner default", func(t *testing.T) {
		schema := BigInt().Default(big.NewInt(42)).Optional()

		got, err := schema.Parse(nil)

		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("exact optional only accepts absent object fields", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"name": String().ExactOptional(),
		})

		got, err := schema.Parse(map[string]any{})
		require.NoError(t, err)
		assert.NotContains(t, got, "name")

		_, err = schema.Parse(map[string]any{"name": nil})
		require.Error(t, err)
	})

	t.Run("outer optional accepts explicit nil over inner exact optional", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"name": String().ExactOptional().Optional(),
		})

		got, err := schema.Parse(map[string]any{"name": nil})

		require.NoError(t, err)
		assert.Contains(t, got, "name")
		assert.Nil(t, got["name"])
	})
}

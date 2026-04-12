package types

import (
	"reflect"
	"testing"

	"github.com/kaptinlin/gozod/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestArray_CloneFromDoesNotShareState(t *testing.T) {
	source := Array([]any{String(), Int()}).Describe("source array")
	target := Array([]any{Bool()}).Meta(core.GlobalMeta{Description: "target array"})

	target.CloneFrom(source)

	assert.NotSame(t, source.internals, target.internals)
	assert.NotEqual(t, reflect.ValueOf(source.internals.Items).Pointer(), reflect.ValueOf(target.internals.Items).Pointer())

	target.internals.SetOptional(true)
	target.internals.Items[0] = Bool()

	assert.False(t, source.IsOptional())

	_, err := source.Parse([]any{"name", 1})
	require.NoError(t, err)

	_, err = target.Parse([]any{"name", 1})
	require.Error(t, err)

	meta, ok := core.GlobalRegistry.Get(target)
	require.True(t, ok)
	assert.Equal(t, "source array", meta.Description)
}

func TestSlice_CloneFromDoesNotShareState(t *testing.T) {
	source := Slice[string](String()).Describe("source slice")
	target := Slice[string](Bool()).Meta(core.GlobalMeta{Description: "target slice"})

	target.CloneFrom(source)

	assert.NotSame(t, source.internals, target.internals)

	target.internals.SetOptional(true)
	target.internals.Element = Int()

	assert.False(t, source.IsOptional())

	_, err := source.Parse([]string{"name"})
	require.NoError(t, err)

	meta, ok := core.GlobalRegistry.Get(target)
	require.True(t, ok)
	assert.Equal(t, "source slice", meta.Description)
}

func TestCollectionCloneFromDoesNotShareInternals(t *testing.T) {
	t.Run("map", func(t *testing.T) {
		source := Map(String(), Int()).Describe("source map")
		target := Map(Bool(), Float64()).Meta(core.GlobalMeta{Description: "target map"})

		target.CloneFrom(source)

		assert.NotSame(t, source.internals, target.internals)
		target.internals.SetOptional(true)
		assert.False(t, source.IsOptional())

		meta, ok := core.GlobalRegistry.Get(target)
		require.True(t, ok)
		assert.Equal(t, "source map", meta.Description)
	})

	t.Run("set", func(t *testing.T) {
		source := Set[string](String()).Describe("source set")
		target := Set[string](Bool()).Meta(core.GlobalMeta{Description: "target set"})

		target.CloneFrom(source)

		assert.NotSame(t, source.internals, target.internals)
		target.internals.SetOptional(true)
		assert.False(t, source.IsOptional())

		meta, ok := core.GlobalRegistry.Get(target)
		require.True(t, ok)
		assert.Equal(t, "source set", meta.Description)
	})

	t.Run("record", func(t *testing.T) {
		source := Record(String(), Int()).Describe("source record")
		target := Record(Bool(), Float64()).Meta(core.GlobalMeta{Description: "target record"})

		target.CloneFrom(source)

		assert.NotSame(t, source.internals, target.internals)
		target.internals.SetOptional(true)
		assert.False(t, source.IsOptional())

		meta, ok := core.GlobalRegistry.Get(target)
		require.True(t, ok)
		assert.Equal(t, "source record", meta.Description)
	})

	t.Run("intersection", func(t *testing.T) {
		source := Intersection(String(), String()).Describe("source intersection")
		target := Intersection(Bool(), Bool()).Meta(core.GlobalMeta{Description: "target intersection"})

		target.CloneFrom(source)

		assert.NotSame(t, source.internals, target.internals)
		target.internals.SetOptional(true)
		assert.False(t, source.IsOptional())

		meta, ok := core.GlobalRegistry.Get(target)
		require.True(t, ok)
		assert.Equal(t, "source intersection", meta.Description)
	})
}

func TestPrimitiveCloneFromSyncsMetadataAndState(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		source := String().Describe("source string")
		target := String().Meta(core.GlobalMeta{Description: "target string"})

		target.CloneFrom(source)

		assert.NotSame(t, source.internals, target.internals)
		target.internals.SetOptional(true)
		assert.False(t, source.IsOptional())

		meta, ok := core.GlobalRegistry.Get(target)
		require.True(t, ok)
		assert.Equal(t, "source string", meta.Description)
	})

	t.Run("file", func(t *testing.T) {
		source := File().Describe("source file")
		target := File().Meta(core.GlobalMeta{Description: "target file"})

		target.CloneFrom(source)

		assert.NotSame(t, source.internals, target.internals)
		target.internals.SetOptional(true)
		assert.False(t, source.IsOptional())

		meta, ok := core.GlobalRegistry.Get(target)
		require.True(t, ok)
		assert.Equal(t, "source file", meta.Description)
	})

	t.Run("never", func(t *testing.T) {
		source := Never().Describe("source never")
		target := Never().Meta(core.GlobalMeta{Description: "target never"})

		target.CloneFrom(source)

		assert.NotSame(t, source.internals, target.internals)
		target.internals.SetOptional(true)
		assert.False(t, source.IsOptional())

		meta, ok := core.GlobalRegistry.Get(target)
		require.True(t, ok)
		assert.Equal(t, "source never", meta.Description)
	})

	t.Run("function", func(t *testing.T) {
		source := Function().Describe("source function")
		target := Function().Meta(core.GlobalMeta{Description: "target function"})

		target.CloneFrom(source)

		assert.NotSame(t, source.internals, target.internals)
		target.internals.SetOptional(true)
		assert.False(t, source.IsOptional())

		meta, ok := core.GlobalRegistry.Get(target)
		require.True(t, ok)
		assert.Equal(t, "source function", meta.Description)
	})

	t.Run("union", func(t *testing.T) {
		source := Union([]any{String(), Int()}).Describe("source union")
		target := Union([]any{Bool()}).Meta(core.GlobalMeta{Description: "target union"})

		target.CloneFrom(source)

		assert.NotSame(t, source.internals, target.internals)
		target.internals.SetOptional(true)
		assert.False(t, source.IsOptional())

		meta, ok := core.GlobalRegistry.Get(target)
		require.True(t, ok)
		assert.Equal(t, "source union", meta.Description)
	})

	t.Run("xor", func(t *testing.T) {
		source := Xor([]any{String(), Int()}).Describe("source xor")
		target := Xor([]any{Bool()}).Meta(core.GlobalMeta{Description: "target xor"})

		target.CloneFrom(source)

		assert.NotSame(t, source.internals, target.internals)
		target.internals.SetOptional(true)
		assert.False(t, source.IsOptional())

		meta, ok := core.GlobalRegistry.Get(target)
		require.True(t, ok)
		assert.Equal(t, "source xor", meta.Description)
	})

	t.Run("discriminated union", func(t *testing.T) {
		source := DiscriminatedUnion(
			"type",
			[]any{
				Object(core.ObjectSchema{"type": Literal("a"), "name": String()}),
				Object(core.ObjectSchema{"type": Literal("b"), "age": Int()}),
			},
		).Describe("source discriminated union")
		target := DiscriminatedUnion(
			"type",
			[]any{
				Object(core.ObjectSchema{"type": Literal("x")}),
			},
		).Meta(core.GlobalMeta{Description: "target discriminated union"})

		target.CloneFrom(source)

		assert.NotSame(t, source.internals, target.internals)
		target.internals.SetOptional(true)
		assert.False(t, source.IsOptional())

		meta, ok := core.GlobalRegistry.Get(target)
		require.True(t, ok)
		assert.Equal(t, "source discriminated union", meta.Description)
	})
}

func TestModifierChainsPreserveMetadataOnScalarSchemas(t *testing.T) {
	t.Run("bool", func(t *testing.T) {
		schema := Bool().Describe("bool schema").Optional()
		meta, ok := core.GlobalRegistry.Get(schema)
		require.True(t, ok)
		assert.Equal(t, "bool schema", meta.Description)
	})

	t.Run("integer", func(t *testing.T) {
		schema := Int().Describe("int schema").Optional()
		meta, ok := core.GlobalRegistry.Get(schema)
		require.True(t, ok)
		assert.Equal(t, "int schema", meta.Description)
	})

	t.Run("time", func(t *testing.T) {
		schema := Time().Describe("time schema").Optional()
		meta, ok := core.GlobalRegistry.Get(schema)
		require.True(t, ok)
		assert.Equal(t, "time schema", meta.Description)
	})

	t.Run("file", func(t *testing.T) {
		schema := File().Describe("file schema").Optional()
		meta, ok := core.GlobalRegistry.Get(schema)
		require.True(t, ok)
		assert.Equal(t, "file schema", meta.Description)
	})

	t.Run("bigint", func(t *testing.T) {
		schema := BigInt().Describe("bigint schema").Optional()
		meta, ok := core.GlobalRegistry.Get(schema)
		require.True(t, ok)
		assert.Equal(t, "bigint schema", meta.Description)
	})

	t.Run("array", func(t *testing.T) {
		schema := Array([]any{String()}).Describe("array schema").Optional()
		meta, ok := core.GlobalRegistry.Get(schema)
		require.True(t, ok)
		assert.Equal(t, "array schema", meta.Description)
	})

	t.Run("slice", func(t *testing.T) {
		schema := Slice[string](String()).Describe("slice schema").Optional()
		meta, ok := core.GlobalRegistry.Get(schema)
		require.True(t, ok)
		assert.Equal(t, "slice schema", meta.Description)
	})

	t.Run("map", func(t *testing.T) {
		schema := Map(String(), Int()).Describe("map schema").Optional()
		meta, ok := core.GlobalRegistry.Get(schema)
		require.True(t, ok)
		assert.Equal(t, "map schema", meta.Description)
	})

	t.Run("set", func(t *testing.T) {
		schema := Set[string](String()).Describe("set schema").Optional()
		meta, ok := core.GlobalRegistry.Get(schema)
		require.True(t, ok)
		assert.Equal(t, "set schema", meta.Description)
	})

	t.Run("record", func(t *testing.T) {
		schema := Record(String(), Int()).Describe("record schema").Optional()
		meta, ok := core.GlobalRegistry.Get(schema)
		require.True(t, ok)
		assert.Equal(t, "record schema", meta.Description)
	})

	t.Run("intersection", func(t *testing.T) {
		schema := Intersection(String(), String()).Describe("intersection schema").Optional()
		meta, ok := core.GlobalRegistry.Get(schema)
		require.True(t, ok)
		assert.Equal(t, "intersection schema", meta.Description)
	})

	t.Run("literal", func(t *testing.T) {
		schema := Literal("x").Describe("literal schema").Optional()
		meta, ok := core.GlobalRegistry.Get(schema)
		require.True(t, ok)
		assert.Equal(t, "literal schema", meta.Description)
	})

	t.Run("tuple", func(t *testing.T) {
		schema := Tuple(String(), Int()).Describe("tuple schema").Optional()
		meta, ok := core.GlobalRegistry.Get(schema)
		require.True(t, ok)
		assert.Equal(t, "tuple schema", meta.Description)
	})
}

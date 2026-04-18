package types

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kaptinlin/gozod/core"
)

func TestObject_CloneFromDoesNotShareState(t *testing.T) {
	source := Object(core.ObjectSchema{
		"name": String(),
	}).Describe("source object")
	target := Object(core.ObjectSchema{
		"age": Int(),
	}).Meta(core.GlobalMeta{Description: "target object"})

	target.CloneFrom(source)

	assert.NotSame(t, source.internals, target.internals)
	assert.NotEqual(t, reflect.ValueOf(source.internals.Shape).Pointer(), reflect.ValueOf(target.internals.Shape).Pointer())

	target.internals.SetOptional(true)
	target.internals.Shape["age"] = Int()

	assert.False(t, source.IsOptional())
	_, exists := source.internals.Shape["age"]
	assert.False(t, exists)

	meta, ok := core.GlobalRegistry.Get(target)
	require.True(t, ok)
	assert.Equal(t, "source object", meta.Description)
}

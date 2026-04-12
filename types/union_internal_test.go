package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnion_CloneFromDoesNotShareInternals(t *testing.T) {
	source := Union([]any{String(), Int()})
	target := Union([]any{Bool(), Float64()})

	target.CloneFrom(source)

	assert.NotSame(t, source.internals, target.internals)

	target.internals.SetOptional(true)
	assert.False(t, source.IsOptional())
	assert.True(t, target.IsOptional())
}

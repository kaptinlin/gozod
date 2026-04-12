package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestZodXor_CloneFromDoesNotShareInternals(t *testing.T) {
	source := XorOf(String(), Int())
	target := XorOf(Bool(), Float64())

	target.CloneFrom(source)

	assert.NotSame(t, source.internals, target.internals)

	target.internals.SetOptional(true)
	assert.False(t, source.IsOptional())
	assert.True(t, target.IsOptional())
}

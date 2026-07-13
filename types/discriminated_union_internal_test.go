package types

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kaptinlin/gozod/core"
)

func TestDiscriminatedUnion_CloneFromDoesNotShareInternals(t *testing.T) {
	source := MustDiscriminatedUnion("type", []core.ZodSchema{
		Object(core.ObjectSchema{
			"type": LiteralOf([]string{"user"}),
			"name": String(),
		}),
	})
	target := MustDiscriminatedUnion("kind", []core.ZodSchema{
		Object(core.ObjectSchema{
			"kind": LiteralOf([]string{"admin"}),
			"role": String(),
		}),
	})

	target.CloneFrom(source)

	assert.NotSame(t, source.internals, target.internals)

	target.internals.SetOptional(true)
	assert.False(t, source.IsOptional())
	assert.True(t, target.IsOptional())
}

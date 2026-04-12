package types

import (
	"testing"

	"github.com/kaptinlin/gozod/core"
	"github.com/stretchr/testify/assert"
)

func TestDiscriminatedUnion_CloneFromDoesNotShareInternals(t *testing.T) {
	source := DiscriminatedUnion("type", []any{
		Object(core.ObjectSchema{
			"type": LiteralOf([]string{"user"}),
			"name": String(),
		}),
	})
	target := DiscriminatedUnion("kind", []any{
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

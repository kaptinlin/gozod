package gozod_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kaptinlin/gozod"
)

func TestDiscriminatedUnion_ConstructsValidSchema(t *testing.T) {
	user := gozod.Object(gozod.ObjectSchema{
		"type": gozod.Literal("user"),
		"name": gozod.String(),
	})
	admin := gozod.Object(gozod.ObjectSchema{
		"type": gozod.Literal("admin"),
		"name": gozod.String(),
	})

	schema, err := gozod.DiscriminatedUnion("type", []gozod.ZodSchema{user, admin})
	require.NoError(t, err)

	got, err := schema.Parse(map[string]any{"type": "user", "name": "Ada"})
	require.NoError(t, err)
	assert.Equal(t, "Ada", got.(map[string]any)["name"])
}

func TestDiscriminatedUnion_ExposesConstructionErrorsAtRoot(t *testing.T) {
	first := gozod.Object(gozod.ObjectSchema{"type": gozod.Literal("duplicate")})
	second := gozod.Object(gozod.ObjectSchema{"type": gozod.Literal("duplicate")})

	schema, err := gozod.DiscriminatedUnion("type", []gozod.ZodSchema{first, second})

	assert.Nil(t, schema)
	require.ErrorIs(t, err, gozod.ErrDuplicateDiscriminator)
	detail, ok := errors.AsType[*gozod.DiscriminatorError](err)
	require.True(t, ok)
	assert.Equal(t, 1, detail.Option)
	assert.Equal(t, "type", detail.Field)
	assert.Equal(t, "duplicate", detail.Value)
}

package gozod_test

import (
	"errors"
	"testing"

	lib "github.com/kaptinlin/jsonschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kaptinlin/gozod"
)

func TestJSONSchemaImportErrorIsAvailableFromRoot(t *testing.T) {
	schema := &lib.Schema{Not: &lib.Schema{Boolean: new(true)}}

	_, err := gozod.FromJSONSchema(schema)
	var importErr *gozod.JSONSchemaImportError
	require.True(t, errors.As(err, &importErr))
	assert.Equal(t, "not", importErr.Keyword)
	assert.Equal(t, "/not", importErr.Pointer)
	assert.ErrorIs(t, err, gozod.ErrUnsupportedJSONSchemaKeyword)
}

func TestJSONSchemaLossyImportIsAvailableFromRoot(t *testing.T) {
	document := &lib.Schema{Not: &lib.Schema{Boolean: new(true)}}

	schema, losses, err := gozod.FromJSONSchemaLossy(document)
	require.NoError(t, err)
	require.NotNil(t, schema)
	require.Len(t, losses, 1)

	loss := losses[0]
	assert.Equal(t, "not", loss.Keyword)
	assert.Equal(t, "/not", loss.Pointer)
	assert.ErrorIs(t, loss, gozod.ErrUnsupportedJSONSchemaKeyword)
	var typedLoss gozod.JSONSchemaImportLossError
	assert.ErrorAs(t, loss, &typedLoss)
}

func TestJSONSchemaRegistryIDErrorIsAvailableFromRoot(t *testing.T) {
	registry := gozod.NewRegistry[gozod.GlobalMeta]().Add(gozod.String(), gozod.GlobalMeta{})

	result, err := gozod.ToJSONSchemaRegistry(registry)

	assert.Nil(t, result)
	assert.ErrorIs(t, err, gozod.ErrInvalidRegistrySchemaID)
}

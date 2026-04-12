package types_test

import (
	"testing"

	"github.com/kaptinlin/gozod/core"
	. "github.com/kaptinlin/gozod/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDescribe tests the Describe method on ZodString
func TestDescribe(t *testing.T) {
	desc := "A valid user ID"
	schema := String().Describe(desc)

	// Validation should passes
	_, err := schema.Parse("user123")
	require.NoError(t, err)

	// Verify metadata registration
	// We check if the schema instance is registered in the GlobalRegistry
	// Note: schema returned by Describe is the one registered

	meta, ok := core.GlobalRegistry.Get(schema)
	require.True(t, ok, "Metadata not found in GlobalRegistry")

	assert.Equal(t, desc, meta.Description)
}

// TestMeta tests the Meta method on ZodString
func TestMeta(t *testing.T) {
	metaData := core.GlobalMeta{
		ID:          "meta-test-id",
		Title:       "Meta Test Title",
		Description: "Meta description",
		Examples:    []any{"example1", "example2"},
	}

	schema := String().Meta(metaData)

	_, err := schema.Parse("valid")
	require.NoError(t, err)

	registered, ok := core.GlobalRegistry.Get(schema)
	require.True(t, ok, "Metadata not found")

	assert.Equal(t, metaData.ID, registered.ID)
	assert.Equal(t, metaData.Title, registered.Title)
	assert.Equal(t, metaData.Description, registered.Description)
	assert.Len(t, registered.Examples, 2)
}

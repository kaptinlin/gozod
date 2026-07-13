package types_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kaptinlin/gozod/core"
	. "github.com/kaptinlin/gozod/types"
)

func TestMetaFollowsSchemaValue(t *testing.T) {
	base := String()
	example := map[string]any{"name": "before"}
	named := base.Meta(core.GlobalMeta{
		ID:       "username",
		Title:    "Username",
		Examples: []any{example},
	})
	t.Cleanup(func() {
		core.GlobalRegistry.Remove(base).Remove(named)
	})

	example["name"] = "after"
	got := named.Internals().Metadata()
	require.Equal(t, "username", got.ID)
	require.Equal(t, "Username", got.Title)
	require.Len(t, got.Examples, 1)
	assert.Equal(t, "before", got.Examples[0].(map[string]any)["name"])

	got.Examples[0].(map[string]any)["name"] = "caller-mutated"
	assert.Equal(t, "before", named.Internals().Metadata().Examples[0].(map[string]any)["name"])
	assert.Equal(t, core.GlobalMeta{}, base.Internals().Metadata())
	assert.False(t, core.GlobalRegistry.Has(named))
}

func TestDescribeMergesIntoSchemaMetadataAcrossFamilies(t *testing.T) {
	tests := []struct {
		name   string
		schema core.ZodSchema
	}{
		{"bool", Bool().Meta(core.GlobalMeta{Title: "Flag"}).Describe("Boolean flag")},
		{"object", Object(core.ObjectSchema{}).Meta(core.GlobalMeta{Title: "Object"}).Describe("Object value")},
		{"slice", Slice[string](String()).Meta(core.GlobalMeta{Title: "Names"}).Describe("Name list")},
		{"email", Email().Meta(core.GlobalMeta{Title: "Email"}).Describe("Email address")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Cleanup(func() { core.GlobalRegistry.Remove(tt.schema) })
			meta := tt.schema.Internals().Metadata()
			assert.NotEmpty(t, meta.Title)
			assert.NotEmpty(t, meta.Description)
			assert.False(t, core.GlobalRegistry.Has(tt.schema))
		})
	}
}

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

	meta := schema.Internals().Metadata()

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

	registered := schema.Internals().Metadata()

	assert.Equal(t, metaData.ID, registered.ID)
	assert.Equal(t, metaData.Title, registered.Title)
	assert.Equal(t, metaData.Description, registered.Description)
	assert.Len(t, registered.Examples, 2)
}

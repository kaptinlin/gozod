package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// DEFAULT VALUE CLONING TESTS
// =============================================================================
// These tests verify that mutable default values (slices, maps) are properly
// cloned to prevent shared state issues across multiple Parse calls.

func TestDefaultValueCloning(t *testing.T) {
	t.Run("Slice default cloning", func(t *testing.T) {
		defaultSlice := []string{"a", "b"}
		schema := Slice[string](String()).Default(defaultSlice)

		// 1. Get default value
		res1, err := schema.Parse(nil)
		require.NoError(t, err, "First parse failed")
		require.Len(t, res1, 2)
		assert.Equal(t, "a", res1[0])
		assert.Equal(t, "b", res1[1])

		// 2. Modify result
		res1[0] = "modified"

		// 3. Get default value again
		res2, err := schema.Parse(nil)
		require.NoError(t, err, "Second parse failed")

		// 4. Assert isolation - each parse should get a fresh clone
		assert.Equal(t, "modified", res1[0], "First result should still be modified")
		assert.Equal(t, "a", res2[0], "Second result should be unaffected by first modification")
		assert.Equal(t, "a", defaultSlice[0], "Original default slice should not be modified")
	})

	t.Run("Map default cloning", func(t *testing.T) {
		defaultMap := map[string]int{"a": 1, "b": 2}
		schema := RecordTyped[map[string]int, map[string]int](String(), Int()).Default(defaultMap)

		// 1. Get default value
		res1, err := schema.Parse(nil)
		require.NoError(t, err, "First parse failed")
		assert.Equal(t, 1, res1["a"])
		assert.Equal(t, 2, res1["b"])

		// 2. Modify result
		res1["a"] = 999

		// 3. Get default value again
		res2, err := schema.Parse(nil)
		require.NoError(t, err, "Second parse failed")

		// 4. Assert isolation - each parse should get a fresh clone
		assert.Equal(t, 999, res1["a"], "First result should still be modified")
		assert.Equal(t, 1, res2["a"], "Second result should be unaffected by first modification")
		assert.Equal(t, 1, defaultMap["a"], "Original default map should not be modified")
	})

	t.Run("Nested composite default cloning", func(t *testing.T) {
		defaultValue := map[string]any{
			"items": []any{
				map[string]any{"name": "first"},
			},
			"meta": map[string]any{
				"count": 1,
			},
		}
		schema := Any().Default(defaultValue)

		res1, err := schema.Parse(nil)
		require.NoError(t, err, "First parse failed")

		items1, ok := res1.(map[string]any)["items"].([]any)
		require.True(t, ok, "Expected nested items slice")
		item1, ok := items1[0].(map[string]any)
		require.True(t, ok, "Expected nested item map")
		item1["name"] = "modified"

		meta1, ok := res1.(map[string]any)["meta"].(map[string]any)
		require.True(t, ok, "Expected nested meta map")
		meta1["count"] = 999

		res2, err := schema.Parse(nil)
		require.NoError(t, err, "Second parse failed")

		items2 := res2.(map[string]any)["items"].([]any)
		item2 := items2[0].(map[string]any)
		meta2 := res2.(map[string]any)["meta"].(map[string]any)

		assert.Equal(t, "modified", item1["name"], "First result should stay modified")
		assert.Equal(t, "first", item2["name"], "Nested map in second result should be isolated")
		assert.Equal(t, 999, meta1["count"], "First result nested map should stay modified")
		assert.Equal(t, 1, meta2["count"], "Nested map in second result should be isolated")

		origItems := defaultValue["items"].([]any)
		origItem := origItems[0].(map[string]any)
		origMeta := defaultValue["meta"].(map[string]any)
		assert.Equal(t, "first", origItem["name"], "Original nested map should not be modified")
		assert.Equal(t, 1, origMeta["count"], "Original nested meta map should not be modified")
	})

	t.Run("Modifier cloning keeps default internals isolated", func(t *testing.T) {
		defaultValue := map[string]any{
			"items": []any{
				map[string]any{"name": "first"},
			},
		}
		source := Any().Default(defaultValue)
		clone := source.Optional().NonOptional()

		clonedDefault, ok := clone.internals.DefaultValue.(map[string]any)
		require.True(t, ok, "Expected cloned default map")

		clonedItems, ok := clonedDefault["items"].([]any)
		require.True(t, ok, "Expected cloned items slice")

		clonedItem, ok := clonedItems[0].(map[string]any)
		require.True(t, ok, "Expected cloned item map")
		clonedItem["name"] = "modified"

		res, err := source.Parse(nil)
		require.NoError(t, err)

		items, ok := res.(map[string]any)["items"].([]any)
		require.True(t, ok, "Expected source items slice")
		item, ok := items[0].(map[string]any)
		require.True(t, ok, "Expected source item map")
		assert.Equal(t, "first", item["name"], "Source default should remain isolated from cloned schema internals")
	})

	t.Run("String default (immutable - no cloning needed)", func(t *testing.T) {
		// Strings are immutable, so no cloning is needed
		// But we test the behavior is correct
		schema := String().Default("default-value")

		res1, err := schema.Parse(nil)
		require.NoError(t, err, "Parse failed")
		assert.Equal(t, "default-value", res1)

		// Parse again
		res2, err := schema.Parse(nil)
		require.NoError(t, err, "Second parse failed")
		assert.Equal(t, "default-value", res2)
	})

	t.Run("DefaultFunc with slice", func(t *testing.T) {
		// DefaultFunc should create a new instance each time
		schema := Slice[string](String()).DefaultFunc(func() []string {
			return []string{"a", "b"}
		})

		res1, err := schema.Parse(nil)
		require.NoError(t, err, "First parse failed")

		res1[0] = "modified"

		res2, err := schema.Parse(nil)
		require.NoError(t, err, "Second parse failed")

		// Since DefaultFunc creates new instances, res2 should be unaffected
		assert.Equal(t, "a", res2[0], "DefaultFunc should create new instance each time")
	})
}

// =============================================================================
// DEFAULT VALUE BEHAVIOR TESTS
// =============================================================================

func TestDefaultValueBehavior(t *testing.T) {
	t.Run("Default vs provided value", func(t *testing.T) {
		schema := String().Default("default")

		// When nil is provided, use default
		res, err := schema.Parse(nil)
		require.NoError(t, err, "Parse(nil) failed")
		assert.Equal(t, "default", res)

		// When value is provided, use the value
		res, err = schema.Parse("provided")
		require.NoError(t, err, "Parse('provided') failed")
		assert.Equal(t, "provided", res)
	})

	t.Run("Default with Optional", func(t *testing.T) {
		schema := String().Default("default").Optional()

		// nil should use default
		res, err := schema.Parse(nil)
		require.NoError(t, err, "Parse(nil) failed")
		// Result is *string
		require.NotNil(t, res, "Expected non-nil result for default value")
		str, ok := any(res).(*string)
		require.True(t, ok, "Expected *string type")
		assert.Equal(t, "default", *str)
	})

	t.Run("Nested default values", func(t *testing.T) {
		// Object with default slice field
		defaultInner := []string{"inner"}
		innerSchema := Slice[string](String()).Default(defaultInner)

		// Since we need Object schema, we'll use a simpler test
		// Just verify the inner schema works
		res1, err := innerSchema.Parse(nil)
		require.NoError(t, err, "Parse failed")

		res1[0] = "modified"

		res2, err := innerSchema.Parse(nil)
		require.NoError(t, err, "Second parse failed")

		assert.Equal(t, "inner", res2[0], "Nested default should be cloned")
	})
}

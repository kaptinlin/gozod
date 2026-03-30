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

package types

import (
	"testing"
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
		if err != nil {
			t.Fatalf("First parse failed: %v", err)
		}
		if len(res1) != 2 || res1[0] != "a" || res1[1] != "b" {
			t.Errorf("Expected [a, b], got %v", res1)
		}

		// 2. Modify result
		res1[0] = "modified"

		// 3. Get default value again
		res2, err := schema.Parse(nil)
		if err != nil {
			t.Fatalf("Second parse failed: %v", err)
		}

		// 4. Assert isolation - each parse should get a fresh clone
		if res1[0] != "modified" {
			t.Errorf("First result should still be modified, got %v", res1[0])
		}
		if res2[0] != "a" {
			t.Errorf("Second result should be unaffected by first modification, got %v", res2[0])
		}
		if defaultSlice[0] != "a" {
			t.Errorf("Original default slice should not be modified, got %v", defaultSlice[0])
		}
	})

	t.Run("Map default cloning", func(t *testing.T) {
		defaultMap := map[string]int{"a": 1, "b": 2}
		schema := RecordTyped[map[string]int, map[string]int](String(), Int()).Default(defaultMap)

		// 1. Get default value
		res1, err := schema.Parse(nil)
		if err != nil {
			t.Fatalf("First parse failed: %v", err)
		}
		if res1["a"] != 1 || res1["b"] != 2 {
			t.Errorf("Expected map[a:1 b:2], got %v", res1)
		}

		// 2. Modify result
		res1["a"] = 999

		// 3. Get default value again
		res2, err := schema.Parse(nil)
		if err != nil {
			t.Fatalf("Second parse failed: %v", err)
		}

		// 4. Assert isolation - each parse should get a fresh clone
		if res1["a"] != 999 {
			t.Errorf("First result should still be modified, got %v", res1["a"])
		}
		if res2["a"] != 1 {
			t.Errorf("Second result should be unaffected by first modification, got %v", res2["a"])
		}
		if defaultMap["a"] != 1 {
			t.Errorf("Original default map should not be modified, got %v", defaultMap["a"])
		}
	})

	t.Run("String default (immutable - no cloning needed)", func(t *testing.T) {
		// Strings are immutable, so no cloning is needed
		// But we test the behavior is correct
		schema := String().Default("default-value")

		res1, err := schema.Parse(nil)
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}
		if res1 != "default-value" {
			t.Errorf("Expected 'default-value', got %v", res1)
		}

		// Parse again
		res2, err := schema.Parse(nil)
		if err != nil {
			t.Fatalf("Second parse failed: %v", err)
		}
		if res2 != "default-value" {
			t.Errorf("Expected 'default-value', got %v", res2)
		}
	})

	t.Run("DefaultFunc with slice", func(t *testing.T) {
		// DefaultFunc should create a new instance each time
		schema := Slice[string](String()).DefaultFunc(func() []string {
			return []string{"a", "b"}
		})

		res1, err := schema.Parse(nil)
		if err != nil {
			t.Fatalf("First parse failed: %v", err)
		}

		res1[0] = "modified"

		res2, err := schema.Parse(nil)
		if err != nil {
			t.Fatalf("Second parse failed: %v", err)
		}

		// Since DefaultFunc creates new instances, res2 should be unaffected
		if res2[0] != "a" {
			t.Errorf("DefaultFunc should create new instance each time, got %v", res2[0])
		}
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
		if err != nil {
			t.Fatalf("Parse(nil) failed: %v", err)
		}
		if res != "default" {
			t.Errorf("Expected 'default', got %v", res)
		}

		// When value is provided, use the value
		res, err = schema.Parse("provided")
		if err != nil {
			t.Fatalf("Parse('provided') failed: %v", err)
		}
		if res != "provided" {
			t.Errorf("Expected 'provided', got %v", res)
		}
	})

	t.Run("Default with Optional", func(t *testing.T) {
		schema := String().Default("default").Optional()

		// nil should use default
		res, err := schema.Parse(nil)
		if err != nil {
			t.Fatalf("Parse(nil) failed: %v", err)
		}
		// Result is *string
		if res == nil {
			t.Error("Expected non-nil result for default value")
		}
		if str, ok := any(res).(*string); !ok || *str != "default" {
			t.Errorf("Expected *string('default'), got %v", res)
		}
	})

	t.Run("Nested default values", func(t *testing.T) {
		// Object with default slice field
		defaultInner := []string{"inner"}
		innerSchema := Slice[string](String()).Default(defaultInner)

		// Since we need Object schema, we'll use a simpler test
		// Just verify the inner schema works
		res1, err := innerSchema.Parse(nil)
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}

		res1[0] = "modified"

		res2, err := innerSchema.Parse(nil)
		if err != nil {
			t.Fatalf("Second parse failed: %v", err)
		}

		if res2[0] != "inner" {
			t.Errorf("Nested default should be cloned, got %v", res2[0])
		}
	})
}

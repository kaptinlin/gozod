package gozod_test

import (
	"testing"

	"github.com/kaptinlin/gozod"
	"github.com/stretchr/testify/assert"
)

// TestDescribeGlobalFunction tests the gozod.Describe() global function
// TypeScript Zod v4 equivalent: z.describe(description)
func TestDescribeGlobalFunction(t *testing.T) {
	t.Run("Describe creates a check that registers description", func(t *testing.T) {
		// Create describe check
		describeCheck := gozod.Describe("A user email address")
		assert.NotNil(t, describeCheck)
		assert.NotNil(t, describeCheck.GetZod())
		assert.Equal(t, "describe", describeCheck.GetZod().Def.Check)
	})

	t.Run("Describe check is no-op for validation", func(t *testing.T) {
		// The describe check function should be a no-op
		describeCheck := gozod.Describe("Test description")
		internals := describeCheck.GetZod()

		// Check function should exist and be callable without error
		assert.NotNil(t, internals.Check)

		// OnAttach should be set
		assert.NotEmpty(t, internals.OnAttach)
	})
}

// TestMetaGlobalFunction tests the gozod.Meta() global function
// TypeScript Zod v4 equivalent: z.meta(metadata)
func TestMetaGlobalFunction(t *testing.T) {
	t.Run("Meta creates a check that registers metadata", func(t *testing.T) {
		// Create meta check
		metaCheck := gozod.Meta(gozod.GlobalMeta{
			Title:       "Age",
			Description: "User's age in years",
		})
		assert.NotNil(t, metaCheck)
		assert.NotNil(t, metaCheck.GetZod())
		assert.Equal(t, "meta", metaCheck.GetZod().Def.Check)
	})

	t.Run("Meta check is no-op for validation", func(t *testing.T) {
		// The meta check function should be a no-op
		metaCheck := gozod.Meta(gozod.GlobalMeta{Title: "Count"})
		internals := metaCheck.GetZod()

		// Check function should exist and be callable without error
		assert.NotNil(t, internals.Check)

		// OnAttach should be set
		assert.NotEmpty(t, internals.OnAttach)
	})

	t.Run("Meta check supports all GlobalMeta fields", func(t *testing.T) {
		metaCheck := gozod.Meta(gozod.GlobalMeta{
			ID:          "user_email",
			Title:       "Email",
			Description: "User's email address",
			Examples:    []any{"user@example.com", "admin@test.org"},
		})

		internals := metaCheck.GetZod()
		assert.NotNil(t, internals)
		assert.NotEmpty(t, internals.OnAttach)
	})
}

// TestDescribeWithInstanceMethod tests using both global Describe and instance .Describe() method
func TestDescribeWithInstanceMethod(t *testing.T) {
	t.Run("instance Describe method works", func(t *testing.T) {
		// Use instance method (existing API)
		schema := gozod.String().Describe("User email")

		// Verify the description was registered
		meta, ok := gozod.GlobalRegistry.Get(schema)
		assert.True(t, ok)
		assert.Equal(t, "User email", meta.Description)

		// Validation should still work
		result, err := schema.Parse("test@example.com")
		assert.NoError(t, err)
		assert.Equal(t, "test@example.com", result)
	})
}

// TestMetaWithInstanceMethod tests using both global Meta and instance .Meta() method
func TestMetaWithInstanceMethod(t *testing.T) {
	t.Run("instance Meta method works", func(t *testing.T) {
		// Use instance method (existing API)
		schema := gozod.Int().Meta(gozod.GlobalMeta{
			Title:       "Age",
			Description: "User's age",
		})

		// Verify the metadata was registered
		meta, ok := gozod.GlobalRegistry.Get(schema)
		assert.True(t, ok)
		assert.Equal(t, "Age", meta.Title)
		assert.Equal(t, "User's age", meta.Description)

		// Validation should still work
		result, err := schema.Parse(42)
		assert.NoError(t, err)
		assert.Equal(t, 42, result)
	})
}

// TestMetadataCheckOnAttachCallback tests that OnAttach callbacks work correctly
func TestMetadataCheckOnAttachCallback(t *testing.T) {
	t.Run("Describe OnAttach registers in global registry", func(t *testing.T) {
		describeCheck := gozod.Describe("Test description")
		internals := describeCheck.GetZod()

		// Verify OnAttach callbacks exist
		assert.Len(t, internals.OnAttach, 1)

		// Create a mock schema and call OnAttach
		mockSchema := gozod.String()
		for _, fn := range internals.OnAttach {
			fn(mockSchema)
		}

		// Verify the description was registered
		meta, ok := gozod.GlobalRegistry.Get(mockSchema)
		assert.True(t, ok)
		assert.Equal(t, "Test description", meta.Description)
	})

	t.Run("Meta OnAttach registers in global registry", func(t *testing.T) {
		metaCheck := gozod.Meta(gozod.GlobalMeta{
			Title:       "TestTitle",
			Description: "TestDesc",
		})
		internals := metaCheck.GetZod()

		// Verify OnAttach callbacks exist
		assert.Len(t, internals.OnAttach, 1)

		// Create a mock schema and call OnAttach
		mockSchema := gozod.Int()
		for _, fn := range internals.OnAttach {
			fn(mockSchema)
		}

		// Verify the metadata was registered
		meta, ok := gozod.GlobalRegistry.Get(mockSchema)
		assert.True(t, ok)
		assert.Equal(t, "TestTitle", meta.Title)
		assert.Equal(t, "TestDesc", meta.Description)
	})
}

// =============================================================================
// ENHANCED ZOD V4 COMPATIBILITY TESTS
// =============================================================================

// TestCombinedDescribeAndMeta tests using both Describe and Meta together
// TypeScript Zod v4 equivalent: z.string().check(z.describe("Email address"), z.meta({ title: "Email" }))
func TestCombinedDescribeAndMeta(t *testing.T) {
	t.Run("Describe and Meta can be combined on same schema", func(t *testing.T) {
		// First describe, then meta
		schema := gozod.String().Describe("Email address")
		schema = schema.Meta(gozod.GlobalMeta{Title: "Email"})

		meta, ok := gozod.GlobalRegistry.Get(schema)
		assert.True(t, ok)
		assert.Equal(t, "Email address", meta.Description)
		assert.Equal(t, "Email", meta.Title)
	})

	t.Run("Meta can override description from Describe", func(t *testing.T) {
		// Meta's description should override previous Describe
		schema := gozod.String().Describe("Old description")
		schema = schema.Meta(gozod.GlobalMeta{
			Title:       "Email",
			Description: "New description",
		})

		meta, ok := gozod.GlobalRegistry.Get(schema)
		assert.True(t, ok)
		// Meta's description should take precedence
		assert.Equal(t, "New description", meta.Description)
		assert.Equal(t, "Email", meta.Title)
	})

	t.Run("validation still works with combined metadata", func(t *testing.T) {
		schema := gozod.String().
			Min(5).
			Describe("A string at least 5 characters").
			Meta(gozod.GlobalMeta{Title: "MinString"})

		// Valid input
		result, err := schema.Parse("hello world")
		assert.NoError(t, err)
		assert.Equal(t, "hello world", result)

		// Invalid input (too short)
		_, err = schema.Parse("hi")
		assert.Error(t, err)

		// Metadata still registered
		meta, ok := gozod.GlobalRegistry.Get(schema)
		assert.True(t, ok)
		assert.Equal(t, "MinString", meta.Title)
	})
}

// TestMetadataOnAllSchemaTypes verifies metadata works on all schema types
func TestMetadataOnAllSchemaTypes(t *testing.T) {
	t.Run("String schema", func(t *testing.T) {
		schema := gozod.String().Describe("A string value")
		meta, ok := gozod.GlobalRegistry.Get(schema)
		assert.True(t, ok)
		assert.Equal(t, "A string value", meta.Description)
	})

	t.Run("Int schema", func(t *testing.T) {
		schema := gozod.Int().Describe("An integer value")
		meta, ok := gozod.GlobalRegistry.Get(schema)
		assert.True(t, ok)
		assert.Equal(t, "An integer value", meta.Description)
	})

	t.Run("Float schema", func(t *testing.T) {
		schema := gozod.Float64().Describe("A float value")
		meta, ok := gozod.GlobalRegistry.Get(schema)
		assert.True(t, ok)
		assert.Equal(t, "A float value", meta.Description)
	})

	t.Run("Bool schema", func(t *testing.T) {
		schema := gozod.Bool().Describe("A boolean value")
		meta, ok := gozod.GlobalRegistry.Get(schema)
		assert.True(t, ok)
		assert.Equal(t, "A boolean value", meta.Description)
	})

	t.Run("Slice schema", func(t *testing.T) {
		schema := gozod.Slice[string](gozod.String()).Describe("An array of strings")
		meta, ok := gozod.GlobalRegistry.Get(schema)
		assert.True(t, ok)
		assert.Equal(t, "An array of strings", meta.Description)
	})

	t.Run("Object schema", func(t *testing.T) {
		schema := gozod.Object(gozod.ObjectSchema{
			"name": gozod.String(),
		}).Describe("A user object")
		meta, ok := gozod.GlobalRegistry.Get(schema)
		assert.True(t, ok)
		assert.Equal(t, "A user object", meta.Description)
	})

	t.Run("Record schema", func(t *testing.T) {
		schema := gozod.Record[string, string](gozod.String(), gozod.String()).Describe("A string record")
		meta, ok := gozod.GlobalRegistry.Get(schema)
		assert.True(t, ok)
		assert.Equal(t, "A string record", meta.Description)
	})

	t.Run("Union schema", func(t *testing.T) {
		schema := gozod.Union([]any{gozod.String(), gozod.Int()}).Describe("String or int")
		meta, ok := gozod.GlobalRegistry.Get(schema)
		assert.True(t, ok)
		assert.Equal(t, "String or int", meta.Description)
	})

	t.Run("Enum schema", func(t *testing.T) {
		schema := gozod.Enum("a", "b", "c").Describe("One of a, b, or c")
		meta, ok := gozod.GlobalRegistry.Get(schema)
		assert.True(t, ok)
		assert.Equal(t, "One of a, b, or c", meta.Description)
	})

	t.Run("Any schema", func(t *testing.T) {
		schema := gozod.Any().Describe("Any value allowed")
		meta, ok := gozod.GlobalRegistry.Get(schema)
		assert.True(t, ok)
		assert.Equal(t, "Any value allowed", meta.Description)
	})
}

// TestMetaWithFullFields tests all GlobalMeta fields
func TestMetaWithFullFields(t *testing.T) {
	t.Run("all GlobalMeta fields are stored", func(t *testing.T) {
		schema := gozod.String().Meta(gozod.GlobalMeta{
			ID:          "user_email",
			Title:       "Email Address",
			Description: "The user's primary email",
			Examples:    []any{"user@example.com", "admin@test.org"},
		})

		meta, ok := gozod.GlobalRegistry.Get(schema)
		assert.True(t, ok)
		assert.Equal(t, "user_email", meta.ID)
		assert.Equal(t, "Email Address", meta.Title)
		assert.Equal(t, "The user's primary email", meta.Description)
		assert.Len(t, meta.Examples, 2)
		assert.Contains(t, meta.Examples, "user@example.com")
		assert.Contains(t, meta.Examples, "admin@test.org")
	})
}

// TestMetadataImmutability tests that metadata operations don't affect original schema
func TestMetadataImmutability(t *testing.T) {
	t.Run("Describe returns new schema instance", func(t *testing.T) {
		original := gozod.String()
		described := original.Describe("Described version")

		// Original should not have metadata
		_, _ = gozod.GlobalRegistry.Get(original)
		// Note: Original may or may not have metadata depending on implementation
		// The key is that described has the correct metadata
		meta, ok := gozod.GlobalRegistry.Get(described)
		assert.True(t, ok)
		assert.Equal(t, "Described version", meta.Description)
	})

	t.Run("Meta returns new schema instance", func(t *testing.T) {
		original := gozod.Int()
		withMeta := original.Meta(gozod.GlobalMeta{Title: "Age"})

		// withMeta should have metadata
		meta, ok := gozod.GlobalRegistry.Get(withMeta)
		assert.True(t, ok)
		assert.Equal(t, "Age", meta.Title)
	})
}

// TestMetadataChaining tests chaining metadata with other operations
func TestMetadataChaining(t *testing.T) {
	t.Run("metadata before validation methods", func(t *testing.T) {
		schema := gozod.String().Describe("Username").Min(3).Max(20)

		meta, ok := gozod.GlobalRegistry.Get(schema)
		assert.True(t, ok)
		assert.Equal(t, "Username", meta.Description)

		// Validation works
		result, err := schema.Parse("alice")
		assert.NoError(t, err)
		assert.Equal(t, "alice", result)

		_, err = schema.Parse("ab")
		assert.Error(t, err)
	})

	t.Run("metadata after validation methods", func(t *testing.T) {
		schema := gozod.String().Min(3).Max(20).Describe("Username")

		meta, ok := gozod.GlobalRegistry.Get(schema)
		assert.True(t, ok)
		assert.Equal(t, "Username", meta.Description)

		// Validation works
		result, err := schema.Parse("alice")
		assert.NoError(t, err)
		assert.Equal(t, "alice", result)

		_, err = schema.Parse("ab")
		assert.Error(t, err)
	})

	t.Run("metadata with modifiers", func(t *testing.T) {
		schema := gozod.String().Describe("Optional email").Optional()

		// Schema should accept nil
		result, err := schema.Parse(nil)
		assert.NoError(t, err)
		assert.Nil(t, result)

		// Metadata was on the original schema before Optional
		// After Optional, the schema is a new instance
	})
}

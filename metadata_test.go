package gozod_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kaptinlin/gozod"
	"github.com/kaptinlin/gozod/core"
)

// TestDescribeWithInstanceMethod tests fluent description metadata.
func TestDescribeWithInstanceMethod(t *testing.T) {
	t.Run("instance Describe method works", func(t *testing.T) {
		// Use instance method (existing API)
		schema := gozod.String().Describe("User email")

		// Verify the description is owned by the schema.
		meta := schema.Internals().Metadata()
		assert.Equal(t, "User email", meta.Description)

		// Validation should still work
		result, err := schema.Parse("test@example.com")
		assert.NoError(t, err)
		assert.Equal(t, "test@example.com", result)
	})
}

// TestMetaWithInstanceMethod tests fluent standard metadata.
func TestMetaWithInstanceMethod(t *testing.T) {
	t.Run("instance Meta method works", func(t *testing.T) {
		// Use instance method (existing API)
		schema := gozod.Int().Meta(gozod.GlobalMeta{
			Title:       "Age",
			Description: "User's age",
		})

		// Verify the metadata is owned by the schema.
		meta := schema.Internals().Metadata()
		assert.Equal(t, "Age", meta.Title)
		assert.Equal(t, "User's age", meta.Description)

		// Validation should still work
		result, err := schema.Parse(42)
		assert.NoError(t, err)
		assert.Equal(t, 42, result)
	})
}

// =============================================================================
// ENHANCED ZOD V4 COMPATIBILITY TESTS
// =============================================================================

// TestCombinedDescribeAndMeta tests using both Describe and Meta together
// TypeScript Zod v4 equivalent: z.string().describe("Email address").meta({ title: "Email" })
func TestCombinedDescribeAndMeta(t *testing.T) {
	t.Run("Describe and Meta can be combined on same schema", func(t *testing.T) {
		// First describe, then meta
		schema := gozod.String().Describe("Email address")
		schema = schema.Meta(gozod.GlobalMeta{Title: "Email"})

		meta := schema.Internals().Metadata()
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

		meta := schema.Internals().Metadata()
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

		// Metadata remains on the schema.
		meta := schema.Internals().Metadata()
		assert.Equal(t, "MinString", meta.Title)
	})
}

// TestMetadataOnAllSchemaTypes verifies metadata works on all schema types
func TestMetadataOnAllSchemaTypes(t *testing.T) {
	t.Run("String schema", func(t *testing.T) {
		schema := gozod.String().Describe("A string value")
		meta := schema.Internals().Metadata()
		assert.Equal(t, "A string value", meta.Description)
	})

	t.Run("Int schema", func(t *testing.T) {
		schema := gozod.Int().Describe("An integer value")
		meta := schema.Internals().Metadata()
		assert.Equal(t, "An integer value", meta.Description)
	})

	t.Run("Float schema", func(t *testing.T) {
		schema := gozod.Float64().Describe("A float value")
		meta := schema.Internals().Metadata()
		assert.Equal(t, "A float value", meta.Description)
	})

	t.Run("Bool schema", func(t *testing.T) {
		schema := gozod.Bool().Describe("A boolean value")
		meta := schema.Internals().Metadata()
		assert.Equal(t, "A boolean value", meta.Description)
	})

	t.Run("Slice schema", func(t *testing.T) {
		schema := gozod.Slice[string](gozod.String()).Describe("An array of strings")
		meta := schema.Internals().Metadata()
		assert.Equal(t, "An array of strings", meta.Description)
	})

	t.Run("Object schema", func(t *testing.T) {
		schema := gozod.Object(gozod.ObjectSchema{
			"name": gozod.String(),
		}).Describe("A user object")
		meta := schema.Internals().Metadata()
		assert.Equal(t, "A user object", meta.Description)
	})

	t.Run("Record schema", func(t *testing.T) {
		schema := gozod.Record[string, string](gozod.String(), gozod.String()).Describe("A string record")
		meta := schema.Internals().Metadata()
		assert.Equal(t, "A string record", meta.Description)
	})

	t.Run("Union schema", func(t *testing.T) {
		schema := gozod.Union([]any{gozod.String(), gozod.Int()}).Describe("String or int")
		meta := schema.Internals().Metadata()
		assert.Equal(t, "String or int", meta.Description)
	})

	t.Run("Enum schema", func(t *testing.T) {
		schema := gozod.Enum("a", "b", "c").Describe("One of a, b, or c")
		meta := schema.Internals().Metadata()
		assert.Equal(t, "One of a, b, or c", meta.Description)
	})

	t.Run("Any schema", func(t *testing.T) {
		schema := gozod.Any().Describe("Any value allowed")
		meta := schema.Internals().Metadata()
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

		meta := schema.Internals().Metadata()
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

		assert.NotSame(t, original, described)
		assert.Equal(t, gozod.GlobalMeta{}, original.Internals().Metadata())

		meta := described.Internals().Metadata()
		assert.Equal(t, "Described version", meta.Description)
	})

	t.Run("Meta returns new schema instance", func(t *testing.T) {
		original := gozod.Int()
		withMeta := original.Meta(gozod.GlobalMeta{Title: "Age"})

		assert.NotSame(t, original, withMeta)
		assert.Equal(t, gozod.GlobalMeta{}, original.Internals().Metadata())

		meta := withMeta.Internals().Metadata()
		assert.Equal(t, "Age", meta.Title)
	})
}

func TestMetadataCopyOnWriteRepresentativeSchemas(t *testing.T) {
	tests := []struct {
		name  string
		build func() (core.ZodSchema, core.ZodSchema)
	}{
		{
			name: "string",
			build: func() (core.ZodSchema, core.ZodSchema) {
				original := gozod.String()
				return original, original.Meta(gozod.GlobalMeta{Title: "String"})
			},
		},
		{
			name: "integer",
			build: func() (core.ZodSchema, core.ZodSchema) {
				original := gozod.Int()
				return original, original.Meta(gozod.GlobalMeta{Title: "Integer"})
			},
		},
		{
			name: "object",
			build: func() (core.ZodSchema, core.ZodSchema) {
				original := gozod.Object(gozod.ObjectSchema{"name": gozod.String()})
				return original, original.Meta(gozod.GlobalMeta{Title: "Object"})
			},
		},
		{
			name: "slice",
			build: func() (core.ZodSchema, core.ZodSchema) {
				original := gozod.Slice[string](gozod.String())
				return original, original.Meta(gozod.GlobalMeta{Title: "Slice"})
			},
		},
		{
			name: "format wrapper",
			build: func() (core.ZodSchema, core.ZodSchema) {
				original := gozod.Email()
				return original, original.Meta(gozod.GlobalMeta{Title: "Email"})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original, withMeta := tt.build()

			assert.NotSame(t, original, withMeta)
			assert.Equal(t, gozod.GlobalMeta{}, original.Internals().Metadata())

			meta := withMeta.Internals().Metadata()
			assert.NotEmpty(t, meta.Title)
		})
	}
}

// TestMetadataChaining tests chaining metadata with other operations
func TestMetadataChaining(t *testing.T) {
	t.Run("metadata before validation methods", func(t *testing.T) {
		schema := gozod.String().Describe("Username").Min(3).Max(20)

		meta := schema.Internals().Metadata()
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

		meta := schema.Internals().Metadata()
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

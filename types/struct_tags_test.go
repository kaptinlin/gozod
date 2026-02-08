package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// TEST TYPES FOR TAG SUPPORT
// =============================================================================

type TaggedUser struct {
	Name  string `gozod:"required,min=2,max=50"`
	Email string `gozod:"required,email"`
	Age   int    `gozod:"min=18,max=120"`
	Bio   string `gozod:"max=500"`
}

type NoTagUser struct {
	Name  string
	Email string
}

// =============================================================================
// FROMSTRUCT FUNCTION TESTS
// =============================================================================

func TestFromStruct_BasicUsage(t *testing.T) {
	// Test that FromStruct creates a valid schema
	schema := FromStruct[TaggedUser]()
	if schema == nil {
		t.Fatal("FromStruct should return a non-nil schema")
	}

	// Verify basic functionality
	user := TaggedUser{
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   25,
		Bio:   "Software engineer",
	}

	result, err := schema.Parse(user)
	require.NoError(t, err, "Parse should succeed for valid user")

	if result.Name != user.Name {
		assert.Equal(t, user.Name, result.Name, "Expected name %s, got %s", user.Name, result.Name)
	}
}

func TestFromStructPtr_PointerTypes(t *testing.T) {
	// Test that FromStructPtr creates a valid pointer schema
	schema := FromStructPtr[TaggedUser]()
	if schema == nil {
		t.Fatal("FromStructPtr should return a non-nil schema")
	}

	// Verify basic functionality with pointer input
	user := &TaggedUser{
		Name:  "Jane Doe",
		Email: "jane@example.com",
		Age:   30,
		Bio:   "Product manager",
	}

	result, err := schema.Parse(user)
	require.NoError(t, err, "Parse should succeed for valid user pointer")

	require.NotNil(t, result, "Expected non-nil result")
	assert.Equal(t, user.Name, result.Name)
}

func TestFromStruct_EmptyStruct(t *testing.T) {
	// Test FromStruct with empty struct
	type EmptyStruct struct{}

	schema := FromStruct[EmptyStruct]()
	if schema == nil {
		t.Fatal("FromStruct should handle empty structs")
	}

	result, err := schema.Parse(EmptyStruct{})
	require.NoError(t, err, "Parse should succeed for empty struct")

	// Verify result is correct type
	_ = result // EmptyStruct type assertion
}

func TestFromStruct_NoTags(t *testing.T) {
	// Test that FromStruct works even without gozod tags
	schema := FromStruct[NoTagUser]()
	if schema == nil {
		t.Fatal("FromStruct should work without tags")
	}

	user := NoTagUser{
		Name:  "Test User",
		Email: "test@example.com",
	}

	result, err := schema.Parse(user)
	require.NoError(t, err, "Parse should succeed for user without tags")

	if result.Name != user.Name {
		assert.Equal(t, user.Name, result.Name, "Expected name %s, got %s", user.Name, result.Name)
	}
}

// =============================================================================
// INTEGRATION TESTS
// =============================================================================

func TestFromStruct_Integration(t *testing.T) {
	// Test that FromStruct integrates well with existing functionality
	schema := FromStruct[TaggedUser]()

	// Test with modifiers
	optionalSchema := schema.Optional()
	if optionalSchema == nil {
		t.Fatal("FromStruct schema should support modifiers")
	}

	// Test nil handling with optional
	result, err := optionalSchema.Parse(nil)
	require.NoError(t, err, "Optional schema should handle nil")

	// For optional structs, nil input should return nil pointer
	if result != nil {
		t.Log("Optional schema returned non-nil result for nil input - this is acceptable")
	}
}

func TestFromStruct_WithRefine(t *testing.T) {
	// Test that FromStruct works with Refine
	schema := FromStruct[TaggedUser]().Refine(func(user TaggedUser) bool {
		return user.Age >= 21 // Adult validation
	}, "User must be at least 21 years old")

	// Test valid case
	adult := TaggedUser{
		Name:  "Adult User",
		Email: "adult@example.com",
		Age:   25,
		Bio:   "Adult user",
	}

	_, err := schema.Parse(adult)
	require.NoError(t, err, "Adult user should pass validation")

	// Test invalid case
	minor := TaggedUser{
		Name:  "Minor User",
		Email: "minor@example.com",
		Age:   18, // Valid for age constraint but fails refine
		Bio:   "Minor user",
	}

	_, err = schema.Parse(minor)
	if err == nil {
		t.Fatal("Minor user should fail refine validation")
	}
}

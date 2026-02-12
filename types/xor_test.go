package types

import (
	"testing"

	"github.com/kaptinlin/gozod/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Optional / Nilable / Nullish / NonOptional modifiers
// =============================================================================

func TestZodXor_Optional(t *testing.T) {
	schema := XorOf(String(), Int()).Optional()

	result, err := schema.Parse("hello")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "hello", *result)

	result, err = schema.Parse(nil)
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestZodXor_Nilable(t *testing.T) {
	schema := XorOf(String(), Int()).Nilable()

	result, err := schema.Parse(nil)
	require.NoError(t, err)
	assert.Nil(t, result)

	result, err = schema.Parse(42)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 42, *result)
}

func TestZodXor_Nullish(t *testing.T) {
	schema := XorOf(String(), Int()).Nullish()

	result, err := schema.Parse(nil)
	require.NoError(t, err)
	assert.Nil(t, result)

	result, err = schema.Parse("test")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "test", *result)

	assert.True(t, schema.IsOptional())
	assert.True(t, schema.IsNilable())
}

func TestZodXor_NonOptional(t *testing.T) {
	schema := XorOf(String(), Int()).Optional().NonOptional()

	result, err := schema.Parse("hello")
	require.NoError(t, err)
	assert.Equal(t, "hello", result)

	_, err = schema.Parse(nil)
	assert.Error(t, err)

	assert.False(t, schema.IsOptional())
}

// =============================================================================
// Default / DefaultFunc
// =============================================================================

func TestZodXor_Default(t *testing.T) {
	schema := XorOf(String(), Int()).Default("fallback")

	result, err := schema.Parse(nil)
	require.NoError(t, err)
	assert.Equal(t, "fallback", result)

	result, err = schema.Parse(42)
	require.NoError(t, err)
	assert.Equal(t, 42, result)
}

func TestZodXor_DefaultFunc(t *testing.T) {
	called := false
	schema := XorOf(String(), Int()).DefaultFunc(func() any {
		called = true
		return "generated"
	})

	result, err := schema.Parse(nil)
	require.NoError(t, err)
	assert.Equal(t, "generated", result)
	assert.True(t, called)
}

// =============================================================================
// Prefault / PrefaultFunc
// =============================================================================

func TestZodXor_Prefault(t *testing.T) {
	schema := XorOf(String(), Int()).Prefault("prefault_val")

	result, err := schema.Parse(nil)
	require.NoError(t, err)
	assert.Equal(t, "prefault_val", result)

	result, err = schema.Parse(99)
	require.NoError(t, err)
	assert.Equal(t, 99, result)
}

func TestZodXor_PrefaultFunc(t *testing.T) {
	called := false
	schema := XorOf(String(), Int()).PrefaultFunc(func() any {
		called = true
		return 100
	})

	result, err := schema.Parse(nil)
	require.NoError(t, err)
	assert.Equal(t, 100, result)
	assert.True(t, called)
}

func TestZodXor_DefaultOverridesPrefault(t *testing.T) {
	schema := XorOf(String(), Int()).Default("default").Prefault("prefault")

	result, err := schema.Parse(nil)
	require.NoError(t, err)
	assert.Equal(t, "default", result)
}

// =============================================================================
// Describe / Meta
// =============================================================================

func TestZodXor_Describe(t *testing.T) {
	schema := XorOf(String(), Int()).Describe("a string or int")

	meta, ok := core.GlobalRegistry.Get(schema)
	require.True(t, ok)
	assert.Equal(t, "a string or int", meta.Description)
}

func TestZodXor_Meta(t *testing.T) {
	schema := XorOf(String(), Int()).Meta(core.GlobalMeta{
		ID:          "xor-test",
		Title:       "XOR Schema",
		Description: "exclusive union",
	})

	meta, ok := core.GlobalRegistry.Get(schema)
	require.True(t, ok)
	assert.Equal(t, "xor-test", meta.ID)
	assert.Equal(t, "XOR Schema", meta.Title)
	assert.Equal(t, "exclusive union", meta.Description)
}

// =============================================================================
// Refine / RefineAny
// =============================================================================

func TestZodXor_Refine(t *testing.T) {
	schema := XorOf(String(), Int()).Refine(func(v any) bool {
		switch val := v.(type) {
		case string:
			return len(val) > 2
		case int:
			return val > 0
		default:
			return false
		}
	}, core.SchemaParams{Error: "custom refine error"})

	result, err := schema.Parse("hello")
	require.NoError(t, err)
	assert.Equal(t, "hello", result)

	result, err = schema.Parse(10)
	require.NoError(t, err)
	assert.Equal(t, 10, result)

	_, err = schema.Parse("ab")
	assert.Error(t, err)

	_, err = schema.Parse(-1)
	assert.Error(t, err)
}

func TestZodXor_RefineAny(t *testing.T) {
	schema := XorOf(String(), Bool()).RefineAny(func(v any) bool {
		return v == "yes" || v == true
	})

	result, err := schema.Parse("yes")
	require.NoError(t, err)
	assert.Equal(t, "yes", result)

	result, err = schema.Parse(true)
	require.NoError(t, err)
	assert.Equal(t, true, result)

	_, err = schema.Parse("no")
	assert.Error(t, err)

	_, err = schema.Parse(false)
	assert.Error(t, err)
}

// =============================================================================
// Transform / Pipe
// =============================================================================

func TestZodXor_Transform(t *testing.T) {
	schema := XorOf(String(), Int()).Transform(
		func(v any, _ *core.RefinementContext) (any, error) {
			if val, ok := v.(string); ok {
				return len(val), nil
			}
			if val, ok := v.(int); ok {
				return val * 2, nil
			}
			return v, nil
		},
	)

	result, err := schema.Parse("hello")
	require.NoError(t, err)
	assert.Equal(t, 5, result)

	result, err = schema.Parse(3)
	require.NoError(t, err)
	assert.Equal(t, 6, result)
}

func TestZodXor_Pipe(t *testing.T) {
	// Pipe xor output into an any schema for further validation
	target := Any()
	schema := XorOf(String(), Int()).Pipe(target)

	result, err := schema.Parse("hello")
	require.NoError(t, err)
	assert.Equal(t, "hello", result)

	result, err = schema.Parse(42)
	require.NoError(t, err)
	assert.Equal(t, 42, result)
}

// =============================================================================
// And / Or composition
// =============================================================================

func TestZodXor_And(t *testing.T) {
	schema := XorOf(String(), Int()).And(String())

	// "hello" matches xor (string) AND string
	result, err := schema.Parse("hello")
	require.NoError(t, err)
	assert.Equal(t, "hello", result)

	// 42 matches xor (int) but NOT string intersection
	_, err = schema.Parse(42)
	assert.Error(t, err)
}

func TestZodXor_Or(t *testing.T) {
	schema := XorOf(String(), Int()).Or(Bool())

	result, err := schema.Parse("hello")
	require.NoError(t, err)
	assert.Equal(t, "hello", result)

	result, err = schema.Parse(42)
	require.NoError(t, err)
	assert.Equal(t, 42, result)

	result, err = schema.Parse(true)
	require.NoError(t, err)
	assert.Equal(t, true, result)

	_, err = schema.Parse(3.14)
	assert.Error(t, err)
}

// =============================================================================
// CloneFrom
// =============================================================================

func TestZodXor_CloneFrom(t *testing.T) {
	source := XorOf(String(), Int()).Describe("source schema")
	target := XorOf(Bool(), Float64())

	target.CloneFrom(source)

	// After clone, target should behave like source
	result, err := target.Parse("hello")
	require.NoError(t, err)
	assert.Equal(t, "hello", result)

	result, err = target.Parse(42)
	require.NoError(t, err)
	assert.Equal(t, 42, result)
}

// =============================================================================
// Options returns defensive copy
// =============================================================================

func TestZodXor_OptionsDefensiveCopy(t *testing.T) {
	schema := XorOf(String(), Int(), Bool())

	options := schema.Options()
	assert.Len(t, options, 3)

	// Modify the returned slice - should not affect original
	options[0] = nil

	// Original should be unaffected
	original := schema.Options()
	assert.Len(t, original, 3)
	assert.NotNil(t, original[0])
}

// =============================================================================
// GetInternals / IsOptional / IsNilable
// =============================================================================

func TestZodXor_GetInternals(t *testing.T) {
	schema := XorOf(String(), Int())

	internals := schema.GetInternals()
	require.NotNil(t, internals)
}

func TestZodXor_IsOptional_Default(t *testing.T) {
	schema := XorOf(String(), Int())
	assert.False(t, schema.IsOptional())
	assert.False(t, schema.IsNilable())
}

func TestZodXor_IsOptional_AfterOptional(t *testing.T) {
	schema := XorOf(String(), Int()).Optional()
	assert.True(t, schema.IsOptional())
	assert.False(t, schema.IsNilable())
}

func TestZodXor_IsNilable_AfterNilable(t *testing.T) {
	schema := XorOf(String(), Int()).Nilable()
	assert.False(t, schema.IsOptional())
	assert.True(t, schema.IsNilable())
}

// =============================================================================
// Modifier immutability
// =============================================================================

func TestZodXor_ModifierImmutability(t *testing.T) {
	original := XorOf(String(), Int())
	_ = original.Optional()

	// Original should not be affected
	_, err := original.Parse(nil)
	assert.Error(t, err)
	assert.False(t, original.IsOptional())
}

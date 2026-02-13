package jsonschema

import (
	"testing"

	"github.com/kaptinlin/gozod/core"
	lib "github.com/kaptinlin/jsonschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFromJSONSchema_String(t *testing.T) {
	t.Run("basic string", func(t *testing.T) {
		schema := &lib.Schema{}
		schema.Type = []string{"string"}

		zodSchema, err := FromJSONSchema(schema)
		require.NoError(t, err)

		result, err := zodSchema.ParseAny("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		_, err = zodSchema.ParseAny(123)
		assert.Error(t, err)
	})

	t.Run("string with constraints", func(t *testing.T) {
		minLen := float64(2)
		maxLen := float64(10)
		schema := &lib.Schema{}
		schema.Type = []string{"string"}
		schema.MinLength = &minLen
		schema.MaxLength = &maxLen

		zodSchema, err := FromJSONSchema(schema)
		require.NoError(t, err)

		result, err := zodSchema.ParseAny("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		_, err = zodSchema.ParseAny("a")
		assert.Error(t, err, "should fail for too short string")

		_, err = zodSchema.ParseAny("this is too long")
		assert.Error(t, err, "should fail for too long string")
	})

	t.Run("string with email format", func(t *testing.T) {
		format := "email"
		schema := &lib.Schema{}
		schema.Type = []string{"string"}
		schema.Format = &format

		zodSchema, err := FromJSONSchema(schema)
		require.NoError(t, err)

		result, err := zodSchema.ParseAny("test@example.com")
		require.NoError(t, err)
		assert.Equal(t, "test@example.com", result)

		_, err = zodSchema.ParseAny("not-an-email")
		assert.Error(t, err)
	})
}

func TestFromJSONSchema_Number(t *testing.T) {
	t.Run("basic number", func(t *testing.T) {
		schema := &lib.Schema{}
		schema.Type = []string{"number"}

		zodSchema, err := FromJSONSchema(schema)
		require.NoError(t, err)

		result, err := zodSchema.ParseAny(3.14)
		require.NoError(t, err)
		assert.Equal(t, 3.14, result)
	})

	t.Run("integer", func(t *testing.T) {
		schema := &lib.Schema{}
		schema.Type = []string{"integer"}

		zodSchema, err := FromJSONSchema(schema)
		require.NoError(t, err)

		result, err := zodSchema.ParseAny(42)
		require.NoError(t, err)
		assert.Equal(t, 42, result)
	})
}

func TestFromJSONSchema_Boolean(t *testing.T) {
	schema := &lib.Schema{}
	schema.Type = []string{"boolean"}

	zodSchema, err := FromJSONSchema(schema)
	require.NoError(t, err)

	result, err := zodSchema.ParseAny(true)
	require.NoError(t, err)
	assert.Equal(t, true, result)

	result, err = zodSchema.ParseAny(false)
	require.NoError(t, err)
	assert.Equal(t, false, result)
}

func TestFromJSONSchema_Null(t *testing.T) {
	schema := &lib.Schema{}
	schema.Type = []string{"null"}

	zodSchema, err := FromJSONSchema(schema)
	require.NoError(t, err)

	result, err := zodSchema.ParseAny(nil)
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestFromJSONSchema_Array(t *testing.T) {
	t.Run("basic array", func(t *testing.T) {
		itemSchema := &lib.Schema{}
		itemSchema.Type = []string{"string"}

		schema := &lib.Schema{}
		schema.Type = []string{"array"}
		schema.Items = itemSchema

		zodSchema, err := FromJSONSchema(schema)
		require.NoError(t, err)

		result, err := zodSchema.ParseAny([]any{"a", "b", "c"})
		require.NoError(t, err)
		assert.Equal(t, []any{"a", "b", "c"}, result)
	})

	t.Run("array with min/max items", func(t *testing.T) {
		minItems := float64(1)
		maxItems := float64(3)
		itemSchema := &lib.Schema{}
		itemSchema.Type = []string{"string"}

		schema := &lib.Schema{}
		schema.Type = []string{"array"}
		schema.Items = itemSchema
		schema.MinItems = &minItems
		schema.MaxItems = &maxItems

		zodSchema, err := FromJSONSchema(schema)
		require.NoError(t, err)

		_, err = zodSchema.ParseAny([]any{})
		assert.Error(t, err, "should fail for empty array")

		_, err = zodSchema.ParseAny([]any{"a", "b", "c", "d"})
		assert.Error(t, err, "should fail for too many items")
	})
}

func TestFromJSONSchema_Object(t *testing.T) {
	t.Run("basic object", func(t *testing.T) {
		nameSchema := &lib.Schema{}
		nameSchema.Type = []string{"string"}

		ageSchema := &lib.Schema{}
		ageSchema.Type = []string{"integer"}

		props := lib.SchemaMap{
			"name": nameSchema,
			"age":  ageSchema,
		}

		schema := &lib.Schema{}
		schema.Type = []string{"object"}
		schema.Properties = &props
		schema.Required = []string{"name"}

		zodSchema, err := FromJSONSchema(schema)
		require.NoError(t, err)

		result, err := zodSchema.ParseAny(map[string]any{
			"name": "John",
			"age":  30,
		})
		require.NoError(t, err)
		assert.Equal(t, "John", result.(map[string]any)["name"])
	})

	t.Run("empty object", func(t *testing.T) {
		schema := &lib.Schema{}
		schema.Type = []string{"object"}

		zodSchema, err := FromJSONSchema(schema)
		require.NoError(t, err)

		result, err := zodSchema.ParseAny(map[string]any{})
		require.NoError(t, err)
		assert.NotNil(t, result)
	})
}

func TestFromJSONSchema_Enum(t *testing.T) {
	t.Run("string enum", func(t *testing.T) {
		schema := &lib.Schema{}
		schema.Enum = []any{"red", "green", "blue"}

		zodSchema, err := FromJSONSchema(schema)
		require.NoError(t, err)

		result, err := zodSchema.ParseAny("red")
		require.NoError(t, err)
		assert.Equal(t, "red", result)

		_, err = zodSchema.ParseAny("yellow")
		assert.Error(t, err)
	})
}

func TestFromJSONSchema_Const(t *testing.T) {
	schema := lib.Const("constant")

	zodSchema, err := FromJSONSchema(schema)
	require.NoError(t, err)

	result, err := zodSchema.ParseAny("constant")
	require.NoError(t, err)
	assert.Equal(t, "constant", result)

	_, err = zodSchema.ParseAny("other")
	assert.Error(t, err)
}

func TestFromJSONSchema_AnyOf(t *testing.T) {
	stringSchema := &lib.Schema{}
	stringSchema.Type = []string{"string"}

	intSchema := &lib.Schema{}
	intSchema.Type = []string{"integer"}

	schema := &lib.Schema{}
	schema.AnyOf = []*lib.Schema{stringSchema, intSchema}

	zodSchema, err := FromJSONSchema(schema)
	require.NoError(t, err)

	result, err := zodSchema.ParseAny("hello")
	require.NoError(t, err)
	assert.Equal(t, "hello", result)

	result, err = zodSchema.ParseAny(42)
	require.NoError(t, err)
	assert.Equal(t, 42, result)
}

func TestFromJSONSchema_AllOf(t *testing.T) {
	// AllOf with two object schemas
	nameSchema := &lib.Schema{}
	nameSchema.Type = []string{"string"}

	ageSchema := &lib.Schema{}
	ageSchema.Type = []string{"integer"}

	props1 := lib.SchemaMap{"name": nameSchema}
	props2 := lib.SchemaMap{"age": ageSchema}

	obj1 := &lib.Schema{}
	obj1.Type = []string{"object"}
	obj1.Properties = &props1

	obj2 := &lib.Schema{}
	obj2.Type = []string{"object"}
	obj2.Properties = &props2

	schema := &lib.Schema{}
	schema.AllOf = []*lib.Schema{obj1, obj2}

	zodSchema, err := FromJSONSchema(schema)
	require.NoError(t, err)
	require.NotNil(t, zodSchema)
}

func TestFromJSONSchema_BooleanSchema(t *testing.T) {
	t.Run("true schema accepts anything", func(t *testing.T) {
		trueVal := true
		schema := &lib.Schema{Boolean: &trueVal}

		zodSchema, err := FromJSONSchema(schema)
		require.NoError(t, err)
		require.NotNil(t, zodSchema)
	})

	t.Run("false schema rejects everything", func(t *testing.T) {
		falseVal := false
		schema := &lib.Schema{Boolean: &falseVal}

		zodSchema, err := FromJSONSchema(schema)
		require.NoError(t, err)

		_, err = zodSchema.ParseAny("anything")
		assert.Error(t, err)
	})
}

func TestFromJSONSchema_MultiType(t *testing.T) {
	t.Run("string or integer", func(t *testing.T) {
		schema := &lib.Schema{}
		schema.Type = []string{"string", "integer"}

		zodSchema, err := FromJSONSchema(schema)
		require.NoError(t, err)

		result, err := zodSchema.ParseAny("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		result, err = zodSchema.ParseAny(42)
		require.NoError(t, err)
		assert.Equal(t, 42, result)
	})
}

func TestFromJSONSchema_StrictMode(t *testing.T) {
	t.Run("strict mode fails on if/then/else", func(t *testing.T) {
		ifSchema := &lib.Schema{}
		ifSchema.Type = []string{"string"}

		schema := &lib.Schema{}
		schema.If = ifSchema

		_, err := FromJSONSchema(schema, FromJSONSchemaOptions{StrictMode: true})
		assert.ErrorIs(t, err, ErrJSONSchemaIfThenElse)
	})
}

func TestFromJSONSchema_NilSchema(t *testing.T) {
	zodSchema, err := FromJSONSchema(nil)
	require.NoError(t, err)
	require.NotNil(t, zodSchema)
}

func TestFromJSONSchema_PrefixItems(t *testing.T) {
	t.Run("basic tuple from prefixItems", func(t *testing.T) {
		stringSchema := &lib.Schema{}
		stringSchema.Type = []string{"string"}

		intSchema := &lib.Schema{}
		intSchema.Type = []string{"integer"}

		schema := &lib.Schema{}
		schema.Type = []string{"array"}
		schema.PrefixItems = []*lib.Schema{stringSchema, intSchema}

		zodSchema, err := FromJSONSchema(schema)
		require.NoError(t, err)

		result, err := zodSchema.ParseAny([]any{"hello", 42})
		require.NoError(t, err)
		assert.Equal(t, []any{"hello", 42}, result)

		// Wrong type at position 0
		_, err = zodSchema.ParseAny([]any{123, 42})
		assert.Error(t, err)

		// Wrong type at position 1
		_, err = zodSchema.ParseAny([]any{"hello", "world"})
		assert.Error(t, err)
	})

	t.Run("tuple with rest element", func(t *testing.T) {
		stringSchema := &lib.Schema{}
		stringSchema.Type = []string{"string"}

		boolSchema := &lib.Schema{}
		boolSchema.Type = []string{"boolean"}

		schema := &lib.Schema{}
		schema.Type = []string{"array"}
		schema.PrefixItems = []*lib.Schema{stringSchema}
		schema.Items = boolSchema // rest elements must be boolean

		zodSchema, err := FromJSONSchema(schema)
		require.NoError(t, err)

		result, err := zodSchema.ParseAny([]any{"hello", true, false})
		require.NoError(t, err)
		assert.Equal(t, []any{"hello", true, false}, result)

		// Rest elements must be boolean
		_, err = zodSchema.ParseAny([]any{"hello", "not-bool"})
		assert.Error(t, err)
	})
}

func TestFromJSONSchema_Metadata(t *testing.T) {
	t.Run("extracts title and description", func(t *testing.T) {
		title := "User Name"
		desc := "The user's full name"
		schema := &lib.Schema{
			Title:       &title,
			Description: &desc,
		}
		schema.Type = []string{"string"}

		zodSchema, err := FromJSONSchema(schema)
		require.NoError(t, err)

		meta, ok := core.GlobalRegistry.Get(zodSchema)
		require.True(t, ok, "Expected metadata to be registered")
		assert.Equal(t, "User Name", meta.Title)
		assert.Equal(t, "The user's full name", meta.Description)
	})

	t.Run("extracts $id and examples", func(t *testing.T) {
		schema := &lib.Schema{
			ID:       "https://example.com/schemas/name",
			Examples: []any{"John", "Jane"},
		}
		schema.Type = []string{"string"}

		zodSchema, err := FromJSONSchema(schema)
		require.NoError(t, err)

		meta, ok := core.GlobalRegistry.Get(zodSchema)
		require.True(t, ok, "Expected metadata to be registered")
		assert.Equal(t, "https://example.com/schemas/name", meta.ID)
		assert.Equal(t, []any{"John", "Jane"}, meta.Examples)
	})

	t.Run("no metadata when fields are empty", func(t *testing.T) {
		schema := &lib.Schema{}
		schema.Type = []string{"integer"}

		zodSchema, err := FromJSONSchema(schema)
		require.NoError(t, err)

		_, ok := core.GlobalRegistry.Get(zodSchema)
		assert.False(t, ok, "Expected no metadata when all fields are empty")
	})
}

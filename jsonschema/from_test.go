package jsonschema

import (
	"testing"

	lib "github.com/kaptinlin/jsonschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kaptinlin/gozod/core"
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
		schema := &lib.Schema{}
		schema.Type = []string{"string"}
		schema.MinLength = new(float64(2))
		schema.MaxLength = new(float64(10))

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
		schema := &lib.Schema{}
		schema.Type = []string{"string"}
		schema.Format = new("email")

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
		itemSchema := &lib.Schema{}
		itemSchema.Type = []string{"string"}

		schema := &lib.Schema{}
		schema.Type = []string{"array"}
		schema.Items = itemSchema
		schema.MinItems = new(float64(1))
		schema.MaxItems = new(float64(3))

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

		schema := &lib.Schema{}
		schema.Type = []string{"object"}
		schema.Properties = new(lib.SchemaMap{
			"name": nameSchema,
			"age":  ageSchema,
		})
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

	obj1 := &lib.Schema{}
	obj1.Type = []string{"object"}
	obj1.Properties = new(lib.SchemaMap{"name": nameSchema})

	obj2 := &lib.Schema{}
	obj2.Type = []string{"object"}
	obj2.Properties = new(lib.SchemaMap{"age": ageSchema})

	schema := &lib.Schema{}
	schema.AllOf = []*lib.Schema{obj1, obj2}

	zodSchema, err := FromJSONSchema(schema)
	require.NoError(t, err)
	require.NotNil(t, zodSchema)
}

func TestFromJSONSchema_BooleanSchema(t *testing.T) {
	t.Run("true schema accepts anything", func(t *testing.T) {
		schema := &lib.Schema{Boolean: new(true)}

		zodSchema, err := FromJSONSchema(schema)
		require.NoError(t, err)
		require.NotNil(t, zodSchema)
	})

	t.Run("false schema rejects everything", func(t *testing.T) {
		schema := &lib.Schema{Boolean: new(false)}

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

func TestFromJSONSchema_StrictModeUnsupportedKeywords(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		schema func() *lib.Schema
		want   error
	}{
		{
			name: "patternProperties",
			schema: func() *lib.Schema {
				patternSchema := &lib.Schema{}
				patternSchema.Type = []string{"string"}
				return &lib.Schema{PatternProperties: &lib.SchemaMap{"^x-": patternSchema}}
			},
			want: ErrJSONSchemaPatternProperties,
		},
		{
			name: "dynamicRef",
			schema: func() *lib.Schema {
				return &lib.Schema{DynamicRef: "#node"}
			},
			want: ErrJSONSchemaDynamicRef,
		},
		{
			name: "unevaluatedProperties",
			schema: func() *lib.Schema {
				return &lib.Schema{UnevaluatedProperties: &lib.Schema{Boolean: new(false)}}
			},
			want: ErrJSONSchemaUnevaluatedProps,
		},
		{
			name: "unevaluatedItems",
			schema: func() *lib.Schema {
				return &lib.Schema{UnevaluatedItems: &lib.Schema{Boolean: new(false)}}
			},
			want: ErrJSONSchemaUnevaluatedItems,
		},
		{
			name: "dependentSchemas",
			schema: func() *lib.Schema {
				return &lib.Schema{DependentSchemas: map[string]*lib.Schema{"card": {Boolean: new(true)}}}
			},
			want: ErrJSONSchemaDependentSchemas,
		},
		{
			name: "propertyNames",
			schema: func() *lib.Schema {
				return &lib.Schema{PropertyNames: &lib.Schema{Boolean: new(true)}}
			},
			want: ErrJSONSchemaPropertyNames,
		},
		{
			name: "contains",
			schema: func() *lib.Schema {
				return &lib.Schema{Contains: &lib.Schema{Boolean: new(true)}}
			},
			want: ErrJSONSchemaContains,
		},
		{
			name: "minContains",
			schema: func() *lib.Schema {
				minContains := float64(1)
				return &lib.Schema{MinContains: &minContains}
			},
			want: ErrJSONSchemaContains,
		},
		{
			name: "maxContains",
			schema: func() *lib.Schema {
				maxContains := float64(2)
				return &lib.Schema{MaxContains: &maxContains}
			},
			want: ErrJSONSchemaContains,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := FromJSONSchema(tt.schema(), FromJSONSchemaOptions{StrictMode: true})
			require.ErrorIs(t, err, tt.want)
		})
	}
}

func TestFromJSONSchema_ObjectAdditionalProperties(t *testing.T) {
	t.Parallel()

	t.Run("false rejects unknown keys", func(t *testing.T) {
		t.Parallel()

		nameSchema := &lib.Schema{}
		nameSchema.Type = []string{"string"}

		schema := &lib.Schema{}
		schema.Type = []string{"object"}
		schema.Properties = &lib.SchemaMap{"name": nameSchema}
		schema.Required = []string{"name"}
		schema.AdditionalProperties = &lib.Schema{Boolean: new(false)}

		zodSchema, err := FromJSONSchema(schema)
		require.NoError(t, err)

		_, err = zodSchema.ParseAny(map[string]any{"name": "Ada", "extra": true})
		require.Error(t, err)
	})

	t.Run("schema validates unknown keys", func(t *testing.T) {
		t.Parallel()

		nameSchema := &lib.Schema{}
		nameSchema.Type = []string{"string"}
		additionalSchema := &lib.Schema{}
		additionalSchema.Type = []string{"integer"}

		schema := &lib.Schema{}
		schema.Type = []string{"object"}
		schema.Properties = &lib.SchemaMap{"name": nameSchema}
		schema.Required = []string{"name"}
		schema.AdditionalProperties = additionalSchema

		zodSchema, err := FromJSONSchema(schema)
		require.NoError(t, err)

		result, err := zodSchema.ParseAny(map[string]any{"name": "Ada", "extra": 1})
		require.NoError(t, err)
		assert.Equal(t, "Ada", result.(map[string]any)["name"])
		assert.Equal(t, 1, result.(map[string]any)["extra"])

		_, err = zodSchema.ParseAny(map[string]any{"name": "Ada", "extra": "wrong"})
		require.Error(t, err)
	})

	t.Run("record validates values", func(t *testing.T) {
		t.Parallel()

		valueSchema := &lib.Schema{}
		valueSchema.Type = []string{"boolean"}

		schema := &lib.Schema{}
		schema.Type = []string{"object"}
		schema.AdditionalProperties = valueSchema

		zodSchema, err := FromJSONSchema(schema)
		require.NoError(t, err)

		result, err := zodSchema.ParseAny(map[string]any{"enabled": true})
		require.NoError(t, err)
		assert.Equal(t, true, result.(map[string]any)["enabled"])

		_, err = zodSchema.ParseAny(map[string]any{"enabled": "yes"})
		require.Error(t, err)
	})
}

func TestFromJSONSchema_NumberAndIntegerConstraints(t *testing.T) {
	t.Parallel()

	t.Run("number constraints", func(t *testing.T) {
		t.Parallel()

		minimum := lib.NewRat(1)
		maximum := lib.NewRat(10)
		exclusiveMinimum := lib.NewRat(0)
		exclusiveMaximum := lib.NewRat(11)
		multipleOf := lib.NewRat(0.5)

		schema := &lib.Schema{}
		schema.Type = []string{"number"}
		schema.Minimum = minimum
		schema.Maximum = maximum
		schema.ExclusiveMinimum = exclusiveMinimum
		schema.ExclusiveMaximum = exclusiveMaximum
		schema.MultipleOf = multipleOf

		zodSchema, err := FromJSONSchema(schema)
		require.NoError(t, err)

		_, err = zodSchema.ParseAny(5.5)
		require.NoError(t, err)
		_, err = zodSchema.ParseAny(0)
		require.Error(t, err)
		_, err = zodSchema.ParseAny(11)
		require.Error(t, err)
		_, err = zodSchema.ParseAny(5.25)
		require.Error(t, err)
	})

	t.Run("integer constraints", func(t *testing.T) {
		t.Parallel()

		minimum := lib.NewRat(1)
		maximum := lib.NewRat(10)
		exclusiveMinimum := lib.NewRat(0)
		exclusiveMaximum := lib.NewRat(11)
		multipleOf := lib.NewRat(2)

		schema := &lib.Schema{}
		schema.Type = []string{"integer"}
		schema.Minimum = minimum
		schema.Maximum = maximum
		schema.ExclusiveMinimum = exclusiveMinimum
		schema.ExclusiveMaximum = exclusiveMaximum
		schema.MultipleOf = multipleOf

		zodSchema, err := FromJSONSchema(schema)
		require.NoError(t, err)

		_, err = zodSchema.ParseAny(4)
		require.NoError(t, err)
		_, err = zodSchema.ParseAny(0)
		require.Error(t, err)
		_, err = zodSchema.ParseAny(11)
		require.Error(t, err)
		_, err = zodSchema.ParseAny(5)
		require.Error(t, err)
	})
}

func TestFromJSONSchema_OneOfUsesExclusiveUnion(t *testing.T) {
	t.Parallel()

	first := &lib.Schema{}
	first.Type = []string{"string"}
	second := &lib.Schema{}
	second.Type = []string{"integer"}

	schema := &lib.Schema{}
	schema.OneOf = []*lib.Schema{first, second}

	zodSchema, err := FromJSONSchema(schema)
	require.NoError(t, err)

	_, err = zodSchema.ParseAny("value")
	require.NoError(t, err)
	_, err = zodSchema.ParseAny(1)
	require.NoError(t, err)

	second.Type = []string{"string"}
	zodSchema, err = FromJSONSchema(schema)
	require.NoError(t, err)

	_, err = zodSchema.ParseAny("value")
	require.Error(t, err)
}

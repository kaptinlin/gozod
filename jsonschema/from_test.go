package jsonschema

import (
	"regexp/syntax"
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

	t.Run("string with pattern", func(t *testing.T) {
		schema := &lib.Schema{}
		schema.Type = []string{"string"}
		schema.Pattern = new("^[A-Z]{2}\\d+$")

		zodSchema, err := FromJSONSchema(schema)
		require.NoError(t, err)

		result, err := zodSchema.ParseAny("AB12")
		require.NoError(t, err)
		assert.Equal(t, "AB12", result)

		_, err = zodSchema.ParseAny("ab12")
		require.Error(t, err)
	})

	t.Run("invalid pattern returns sentinel", func(t *testing.T) {
		schema := &lib.Schema{}
		schema.Type = []string{"string"}
		schema.Pattern = new("[")

		_, err := FromJSONSchema(schema)
		require.ErrorIs(t, err, ErrJSONSchemaPatternCompile)

		var syntaxErr *syntax.Error
		require.ErrorAs(t, err, &syntaxErr)
		assert.Equal(t, syntax.ErrMissingBracket, syntaxErr.Code)
		assert.Equal(t, "[", syntaxErr.Expr)
	})
}

func TestFromJSONSchema_StringFormats(t *testing.T) {
	tests := []struct {
		name    string
		format  string
		valid   string
		invalid string
	}{
		{name: "uuid", format: "uuid", valid: "550e8400-e29b-41d4-a716-446655440000", invalid: "not-a-uuid"},
		{name: "uri", format: "uri", valid: "https://example.com", invalid: "not a url"},
		{name: "url", format: "url", valid: "https://example.com", invalid: "not a url"},
		{name: "date-time", format: "date-time", valid: "2023-12-25T15:30:45Z", invalid: "2023-12-25 15:30:45"},
		{name: "date", format: "date", valid: "2024-02-29", invalid: "2024-02-30"},
		{name: "time", format: "time", valid: "15:30:45", invalid: "25:00:00"},
		{name: "ipv4", format: "ipv4", valid: "192.168.0.1", invalid: "999.999.999.999"},
		{name: "ipv6", format: "ipv6", valid: "2001:db8::1", invalid: "192.168.0.1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := &lib.Schema{}
			schema.Type = []string{"string"}
			schema.Format = &tt.format

			zodSchema, err := FromJSONSchema(schema)
			require.NoError(t, err)

			result, err := zodSchema.ParseAny(tt.valid)
			require.NoError(t, err)
			assert.Equal(t, tt.valid, result)

			_, err = zodSchema.ParseAny(tt.invalid)
			require.Error(t, err)
		})
	}
}

func TestFromJSONSchema_UnknownStringFormatFallsBack(t *testing.T) {
	schema := &lib.Schema{}
	schema.Type = []string{"string"}
	schema.Format = new("slug")

	zodSchema, err := FromJSONSchema(schema)
	require.NoError(t, err)

	result, err := zodSchema.ParseAny("not a slug, but still a string")
	require.NoError(t, err)
	assert.Equal(t, "not a slug, but still a string", result)

	_, err = zodSchema.ParseAny(123)
	require.Error(t, err)
}

func TestFromJSONSchema_StringFormatPreservesLengthAndPatternConstraints(t *testing.T) {
	format := "email"
	minLength := float64(20)
	pattern := "^[^@]+@example\\.com$"
	schema := &lib.Schema{
		Type:      []string{"string"},
		Format:    &format,
		MinLength: &minLength,
		Pattern:   &pattern,
	}

	zodSchema, err := FromJSONSchema(schema)
	require.NoError(t, err)

	_, err = zodSchema.ParseAny("short@example.com")
	require.Error(t, err)

	_, err = zodSchema.ParseAny("long-enough@example.org")
	require.Error(t, err)

	result, err := zodSchema.ParseAny("long-enough@example.com")
	require.NoError(t, err)
	assert.Equal(t, "long-enough@example.com", result)
}

func TestFromJSONSchema_InvalidReferencesAndTypes(t *testing.T) {
	t.Run("unresolved ref errors", func(t *testing.T) {
		_, err := FromJSONSchema(&lib.Schema{Ref: "#/$defs/Missing"})
		require.ErrorIs(t, err, ErrInvalidJSONSchema)
	})

	t.Run("unknown type errors", func(t *testing.T) {
		_, err := FromJSONSchema(&lib.Schema{Type: []string{"strng"}})
		require.ErrorIs(t, err, ErrUnsupportedJSONSchemaType)
	})

	t.Run("unknown multi-type member errors", func(t *testing.T) {
		_, err := FromJSONSchema(&lib.Schema{Type: []string{"string", "strng"}})
		require.ErrorIs(t, err, ErrUnsupportedJSONSchemaType)
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

	t.Run("optional properties validate only when present", func(t *testing.T) {
		nameSchema := &lib.Schema{}
		nameSchema.Type = []string{"string"}

		activeSchema := &lib.Schema{}
		activeSchema.Type = []string{"boolean"}

		tagSchema := &lib.Schema{}
		tagSchema.Type = []string{"string"}
		tagsSchema := &lib.Schema{}
		tagsSchema.Type = []string{"array"}
		tagsSchema.Items = tagSchema

		citySchema := &lib.Schema{}
		citySchema.Type = []string{"string"}
		profileSchema := &lib.Schema{}
		profileSchema.Type = []string{"object"}
		profileSchema.Properties = &lib.SchemaMap{"city": citySchema}

		schema := &lib.Schema{}
		schema.Type = []string{"object"}
		schema.Properties = &lib.SchemaMap{
			"name":    nameSchema,
			"active":  activeSchema,
			"tags":    tagsSchema,
			"profile": profileSchema,
		}
		schema.Required = []string{"name"}

		zodSchema, err := FromJSONSchema(schema)
		require.NoError(t, err)

		result, err := zodSchema.ParseAny(map[string]any{"name": "Ada"})
		require.NoError(t, err)
		assert.Equal(t, "Ada", result.(map[string]any)["name"])

		result, err = zodSchema.ParseAny(map[string]any{
			"name":    "Ada",
			"active":  true,
			"tags":    []any{"go", "zod"},
			"profile": map[string]any{"city": "London"},
		})
		require.NoError(t, err)
		got := result.(map[string]any)
		assert.Equal(t, true, got["active"])
		assert.Equal(t, []any{"go", "zod"}, got["tags"])

		_, err = zodSchema.ParseAny(map[string]any{"name": "Ada", "active": "yes"})
		require.Error(t, err)

		_, err = zodSchema.ParseAny(map[string]any{"name": "Ada", "tags": []any{"go", 1}})
		require.Error(t, err)
	})
}

func TestFromJSONSchema_ObjectOptionalEnumProperty(t *testing.T) {
	t.Parallel()

	statusSchema := &lib.Schema{Enum: []any{"draft", "published"}}
	schema := &lib.Schema{}
	schema.Type = []string{"object"}
	schema.Properties = &lib.SchemaMap{"status": statusSchema}

	zodSchema, err := FromJSONSchema(schema)
	require.NoError(t, err)

	result, err := zodSchema.ParseAny(map[string]any{})
	require.NoError(t, err)
	assert.NotContains(t, result.(map[string]any), "status")

	result, err = zodSchema.ParseAny(map[string]any{"status": "draft"})
	require.NoError(t, err)
	assert.Equal(t, "draft", result.(map[string]any)["status"])

	_, err = zodSchema.ParseAny(map[string]any{"status": "archived"})
	require.Error(t, err)
}

func TestFromJSONSchema_ObjectPropertyConversionError(t *testing.T) {
	t.Parallel()

	codeSchema := &lib.Schema{}
	codeSchema.Type = []string{"string"}
	codeSchema.Pattern = new("[")

	schema := &lib.Schema{}
	schema.Type = []string{"object"}
	schema.Properties = &lib.SchemaMap{"code": codeSchema}
	schema.Required = []string{"code"}

	_, err := FromJSONSchema(schema)
	require.ErrorIs(t, err, ErrJSONSchemaPatternCompile)
}

func TestFromJSONSchema_ResolvedRef(t *testing.T) {
	target := &lib.Schema{}
	target.Type = []string{"string"}
	target.MinLength = new(float64(3))

	schema := &lib.Schema{ResolvedRef: target}

	zodSchema, err := FromJSONSchema(schema)
	require.NoError(t, err)

	result, err := zodSchema.ParseAny("Ada")
	require.NoError(t, err)
	assert.Equal(t, "Ada", result)

	_, err = zodSchema.ParseAny("Al")
	require.Error(t, err)
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

func TestJSONSchemaSupportedSubsetRoundTrip(t *testing.T) {
	minLength := float64(2)
	maxLength := float64(20)
	minItems := float64(1)
	maxItems := float64(3)
	email := "email"

	input := &lib.Schema{
		Type: []string{"object"},
		Properties: &lib.SchemaMap{
			"name": {
				Type:      []string{"string"},
				MinLength: &minLength,
				MaxLength: &maxLength,
			},
			"email": {
				Type:   []string{"string"},
				Format: &email,
			},
			"tags": {
				Type:     []string{"array"},
				Items:    &lib.Schema{Type: []string{"string"}},
				MinItems: &minItems,
				MaxItems: &maxItems,
			},
		},
		Required:             []string{"name", "email"},
		AdditionalProperties: &lib.Schema{Boolean: new(false)},
	}

	zodSchema, err := FromJSONSchema(input)
	require.NoError(t, err)

	exported, err := ToJSONSchema(zodSchema)
	require.NoError(t, err)
	require.Equal(t, lib.SchemaType{"object"}, exported.Type)
	assert.ElementsMatch(t, []string{"name", "email"}, exported.Required)
	require.NotNil(t, exported.AdditionalProperties)
	require.NotNil(t, exported.AdditionalProperties.Boolean)
	assert.False(t, *exported.AdditionalProperties.Boolean)

	require.NotNil(t, exported.Properties)
	nameSchema := (*exported.Properties)["name"]
	require.NotNil(t, nameSchema)
	assert.Equal(t, lib.SchemaType{"string"}, nameSchema.Type)
	require.NotNil(t, nameSchema.MinLength)
	require.NotNil(t, nameSchema.MaxLength)
	assert.Equal(t, minLength, *nameSchema.MinLength)
	assert.Equal(t, maxLength, *nameSchema.MaxLength)

	emailSchema := (*exported.Properties)["email"]
	require.NotNil(t, emailSchema)
	require.NotNil(t, emailSchema.Format)
	assert.Equal(t, email, *emailSchema.Format)

	tagsSchema := (*exported.Properties)["tags"]
	require.NotNil(t, tagsSchema)
	assert.Equal(t, lib.SchemaType{"array"}, tagsSchema.Type)
	require.NotNil(t, tagsSchema.Items)
	assert.Equal(t, lib.SchemaType{"string"}, tagsSchema.Items.Type)
	require.NotNil(t, tagsSchema.MinItems)
	require.NotNil(t, tagsSchema.MaxItems)
	assert.Equal(t, minItems, *tagsSchema.MinItems)
	assert.Equal(t, maxItems, *tagsSchema.MaxItems)
}

func TestFromJSONSchema_UnsupportedDefaultsToError(t *testing.T) {
	t.Run("default conversion fails on if/then/else", func(t *testing.T) {
		ifSchema := &lib.Schema{}
		ifSchema.Type = []string{"string"}

		schema := &lib.Schema{}
		schema.If = ifSchema

		_, err := FromJSONSchema(schema)
		assert.ErrorIs(t, err, ErrJSONSchemaIfThenElse)
	})

	t.Run("explicit lossy mode reports ignored keywords", func(t *testing.T) {
		ifSchema := &lib.Schema{}
		ifSchema.Type = []string{"string"}

		schema := &lib.Schema{}
		schema.If = ifSchema

		var keywords []string
		zodSchema, err := FromJSONSchema(schema, FromJSONSchemaOptions{
			AllowLossy:    true,
			LossyKeywords: &keywords,
		})
		require.NoError(t, err)
		require.NotNil(t, zodSchema)
		assert.Equal(t, []string{"if/then/else"}, keywords)
	})
}

func TestFromJSONSchema_UnsupportedImportKeywordContract(t *testing.T) {
	samples := map[string]func() *lib.Schema{
		"if/then/else": func() *lib.Schema {
			return &lib.Schema{If: &lib.Schema{Type: []string{"string"}}}
		},
		"patternProperties": func() *lib.Schema {
			return &lib.Schema{PatternProperties: &lib.SchemaMap{"^x-": {Type: []string{"string"}}}}
		},
		"$dynamicRef": func() *lib.Schema {
			return &lib.Schema{DynamicRef: "#node"}
		},
		"unevaluatedProperties": func() *lib.Schema {
			return &lib.Schema{UnevaluatedProperties: &lib.Schema{Boolean: new(false)}}
		},
		"unevaluatedItems": func() *lib.Schema {
			return &lib.Schema{UnevaluatedItems: &lib.Schema{Boolean: new(false)}}
		},
		"not": func() *lib.Schema {
			return &lib.Schema{Not: &lib.Schema{Type: []string{"string"}}}
		},
		"uniqueItems": func() *lib.Schema {
			return &lib.Schema{UniqueItems: new(true)}
		},
		"dependentSchemas": func() *lib.Schema {
			return &lib.Schema{DependentSchemas: map[string]*lib.Schema{"card": {Boolean: new(true)}}}
		},
		"propertyNames": func() *lib.Schema {
			return &lib.Schema{PropertyNames: &lib.Schema{Boolean: new(true)}}
		},
		"contains": func() *lib.Schema {
			return &lib.Schema{Contains: &lib.Schema{Boolean: new(true)}}
		},
	}

	require.Len(t, samples, len(unsupportedImportKeywords))
	seen := make(map[string]struct{}, len(unsupportedImportKeywords))
	for _, keyword := range unsupportedImportKeywords {
		if _, ok := seen[keyword.keyword]; ok {
			t.Fatalf("duplicate unsupported keyword contract: %s", keyword.keyword)
		}
		seen[keyword.keyword] = struct{}{}
		if _, ok := samples[keyword.keyword]; !ok {
			t.Fatalf("missing unsupported keyword sample: %s", keyword.keyword)
		}
	}

	for _, keyword := range unsupportedImportKeywords {
		t.Run(keyword.keyword, func(t *testing.T) {
			schema := samples[keyword.keyword]()
			require.True(t, keyword.present(schema), "sample must trigger contract predicate")

			_, err := FromJSONSchema(schema)
			require.ErrorIs(t, err, keyword.err)

			var lossyKeywords []string
			zodSchema, err := FromJSONSchema(schema, FromJSONSchemaOptions{
				AllowLossy:    true,
				LossyKeywords: &lossyKeywords,
			})
			require.NoError(t, err)
			require.NotNil(t, zodSchema)
			assert.Equal(t, []string{keyword.keyword}, lossyKeywords)
		})
	}
}

func TestFromJSONSchema_LossyRecordsEveryUnsupportedKeyword(t *testing.T) {
	schema := &lib.Schema{
		If:          &lib.Schema{Type: []string{"string"}},
		DynamicRef:  "#node",
		Not:         &lib.Schema{Type: []string{"integer"}},
		UniqueItems: new(true),
	}

	var keywords []string
	zodSchema, err := FromJSONSchema(schema, FromJSONSchemaOptions{
		AllowLossy:    true,
		LossyKeywords: &keywords,
	})
	require.NoError(t, err)
	require.NotNil(t, zodSchema)
	assert.Equal(t, []string{"if/then/else", "$dynamicRef", "not", "uniqueItems"}, keywords)
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

func TestFromJSONSchema_DefaultUnsupportedKeywords(t *testing.T) {
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
			name: "not",
			schema: func() *lib.Schema {
				return &lib.Schema{Not: &lib.Schema{Type: []string{"string"}}}
			},
			want: ErrUnsupportedJSONSchemaKeyword,
		},
		{
			name: "uniqueItems",
			schema: func() *lib.Schema {
				return &lib.Schema{UniqueItems: new(true)}
			},
			want: ErrUnsupportedJSONSchemaKeyword,
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

			_, err := FromJSONSchema(tt.schema())
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

	t.Run("true preserves unknown keys", func(t *testing.T) {
		t.Parallel()

		nameSchema := &lib.Schema{}
		nameSchema.Type = []string{"string"}

		schema := &lib.Schema{}
		schema.Type = []string{"object"}
		schema.Properties = &lib.SchemaMap{"name": nameSchema}
		schema.Required = []string{"name"}
		schema.AdditionalProperties = &lib.Schema{Boolean: new(true)}

		zodSchema, err := FromJSONSchema(schema)
		require.NoError(t, err)

		result, err := zodSchema.ParseAny(map[string]any{"name": "Ada", "extra": true})
		require.NoError(t, err)
		assert.Equal(t, "Ada", result.(map[string]any)["name"])
		assert.Equal(t, true, result.(map[string]any)["extra"])
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

	t.Run("schema conversion errors are returned", func(t *testing.T) {
		t.Parallel()

		nameSchema := &lib.Schema{}
		nameSchema.Type = []string{"string"}

		additionalSchema := &lib.Schema{}
		additionalSchema.Type = []string{"string"}
		additionalSchema.Pattern = new("[")

		schema := &lib.Schema{}
		schema.Type = []string{"object"}
		schema.Properties = &lib.SchemaMap{"name": nameSchema}
		schema.Required = []string{"name"}
		schema.AdditionalProperties = additionalSchema

		_, err := FromJSONSchema(schema)
		require.ErrorIs(t, err, ErrJSONSchemaPatternCompile)
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

	t.Run("fractional integer bounds use integer-domain semantics", func(t *testing.T) {
		t.Parallel()

		schema := &lib.Schema{}
		schema.Type = []string{"integer"}
		schema.Minimum = lib.NewRat("1.5")
		schema.ExclusiveMaximum = lib.NewRat("4.5")

		zodSchema, err := FromJSONSchema(schema)
		require.NoError(t, err)

		_, err = zodSchema.ParseAny(1)
		require.Error(t, err)
		_, err = zodSchema.ParseAny(2)
		require.NoError(t, err)
		_, err = zodSchema.ParseAny(4)
		require.NoError(t, err)
		_, err = zodSchema.ParseAny(5)
		require.Error(t, err)
	})

	t.Run("fractional integer multipleOf maps to exact integer divisor", func(t *testing.T) {
		t.Parallel()

		schema := &lib.Schema{}
		schema.Type = []string{"integer"}
		schema.MultipleOf = lib.NewRat("1.5")

		zodSchema, err := FromJSONSchema(schema)
		require.NoError(t, err)

		_, err = zodSchema.ParseAny(3)
		require.NoError(t, err)
		_, err = zodSchema.ParseAny(2)
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

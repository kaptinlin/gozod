package jsonschema

import (
	"errors"
	"os"
	"regexp"
	"regexp/syntax"
	"strconv"
	"strings"
	"sync"
	"testing"

	lib "github.com/kaptinlin/jsonschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/types"
)

func compileImportSchema(t *testing.T, source string) *lib.Schema {
	t.Helper()
	compiler := lib.NewCompiler()
	schema, err := compiler.Compile([]byte(source))
	require.NoError(t, err)
	return schema
}

func importLossKeywords(losses []ImportLossError) []string {
	keywords := make([]string, len(losses))
	for i, loss := range losses {
		keywords[i] = loss.Keyword
	}
	return keywords
}

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

func TestFromJSONSchema_ContentMediaTypeJSON(t *testing.T) {
	schema := compileImportSchema(t, `{
		"type": "string",
		"contentMediaType": "application/json"
	}`)
	imported, err := FromJSONSchema(schema)
	require.NoError(t, err)

	for _, input := range []any{
		`{"name":"Ada"}`,
		`[1,true,null]`,
		`{"name":`,
		`plain text`,
		42,
	} {
		dependencyValid := schema.Validate(input).IsValid()
		result, parseErr := imported.ParseAny(input)
		assert.Equal(t, dependencyValid, parseErr == nil, "input %#v", input)
		if parseErr == nil {
			assert.Equal(t, input, result)
		}
	}
}

func TestFromJSONSchema_ContentMediaTypeComposesWithFormat(t *testing.T) {
	compiler := lib.NewCompiler().SetAssertFormat(true)
	schema, err := compiler.Compile([]byte(`{
		"type": "string",
		"format": "email",
		"contentMediaType": "application/json"
	}`))
	require.NoError(t, err)
	imported, err := FromJSONSchema(schema)
	require.NoError(t, err)

	for _, input := range []any{
		`test@example.com`,
		`"test@example.com"`,
	} {
		dependencyValid := schema.Validate(input).IsValid()
		_, parseErr := imported.ParseAny(input)
		assert.Equal(t, dependencyValid, parseErr == nil, "input %#v", input)
	}
}

func TestFromJSONSchema_RejectsContentEncodingWithoutExactCheck(t *testing.T) {
	schema := compileImportSchema(t, `{
		"type": "string",
		"contentEncoding": "base64"
	}`)
	candidate := types.Base64()
	tests := []struct {
		name            string
		input           any
		dependencyValid bool
		candidateValid  bool
	}{
		{name: "canonical", input: "aGVsbG8=", dependencyValid: true, candidateValid: true},
		{name: "folded", input: "aG\nVsbG8=", dependencyValid: true, candidateValid: false},
		{name: "invalid", input: "%%%", dependencyValid: false, candidateValid: false},
		{name: "non-string", input: 42, dependencyValid: false, candidateValid: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.dependencyValid, schema.Validate(test.input).IsValid())
			_, parseErr := candidate.ParseAny(test.input)
			assert.Equal(t, test.candidateValid, parseErr == nil)
		})
	}

	imported, err := FromJSONSchema(schema)
	assert.Nil(t, imported)
	assert.ErrorIs(t, err, ErrUnsupportedJSONSchemaKeyword)
	var importErr *ImportError
	require.ErrorAs(t, err, &importErr)
	assert.Equal(t, "contentEncoding", importErr.Keyword)
	assert.Equal(t, "/contentEncoding", importErr.Pointer)
}

func TestFromJSONSchema_RejectsCombinedContentPipeline(t *testing.T) {
	schema := compileImportSchema(t, `{
		"type": "string",
		"contentEncoding": "base64",
		"contentMediaType": "application/json"
	}`)

	imported, err := FromJSONSchema(schema)
	assert.Nil(t, imported)
	var importErr *ImportError
	require.ErrorAs(t, err, &importErr)
	assert.Equal(t, "contentEncoding", importErr.Keyword)
	assert.Equal(t, "/contentEncoding", importErr.Pointer)
}

func TestFromJSONSchema_RejectsUnsupportedContentMediaType(t *testing.T) {
	schema := compileImportSchema(t, `{
		"type": "string",
		"contentMediaType": "image/png"
	}`)

	imported, err := FromJSONSchema(schema)
	assert.Nil(t, imported)
	assert.ErrorIs(t, err, ErrUnsupportedJSONSchemaKeyword)
	var importErr *ImportError
	require.ErrorAs(t, err, &importErr)
	assert.Equal(t, "contentMediaType", importErr.Keyword)
	assert.Equal(t, "/contentMediaType", importErr.Pointer)
}

func TestFromJSONSchema_RejectsContentSchema(t *testing.T) {
	schema := compileImportSchema(t, `{
		"type": "string",
		"contentMediaType": "application/json",
		"contentSchema": {
			"type": "object",
			"properties": {"name": {"type": "string"}},
			"required": ["name"]
		}
	}`)

	imported, err := FromJSONSchema(schema)
	assert.Nil(t, imported)
	assert.ErrorIs(t, err, ErrUnsupportedJSONSchemaKeyword)
	var importErr *ImportError
	require.ErrorAs(t, err, &importErr)
	assert.Equal(t, "contentSchema", importErr.Keyword)
	assert.Equal(t, "/contentSchema", importErr.Pointer)
}

func TestFromJSONSchema_RejectsUnanchoredContentMediaType(t *testing.T) {
	schema := compileImportSchema(t, `{
		"contentMediaType": "application/json"
	}`)

	imported, err := FromJSONSchema(schema)
	assert.Nil(t, imported)
	assert.ErrorIs(t, err, ErrUnsupportedJSONSchemaKeyword)
	var importErr *ImportError
	require.ErrorAs(t, err, &importErr)
	assert.Equal(t, "contentMediaType", importErr.Keyword)
	assert.Equal(t, "/contentMediaType", importErr.Pointer)
}

func TestFromJSONSchema_ContentAssertionsDoNotApplyToExplicitNonStringType(t *testing.T) {
	schema := compileImportSchema(t, `{
		"type": "number",
		"contentEncoding": "custom",
		"contentMediaType": "image/png",
		"contentSchema": false
	}`)
	imported, err := FromJSONSchema(schema)
	require.NoError(t, err)

	for _, input := range []any{42.5, "plain text", true} {
		dependencyValid := schema.Validate(input).IsValid()
		_, parseErr := imported.ParseAny(input)
		assert.Equal(t, dependencyValid, parseErr == nil, "input %#v", input)
	}
}

func TestFromJSONSchema_ContentMediaTypeJSONWithMultipleTypes(t *testing.T) {
	schema := compileImportSchema(t, `{
		"type": ["string", "number"],
		"contentMediaType": "application/json"
	}`)
	imported, err := FromJSONSchema(schema)
	require.NoError(t, err)

	for _, input := range []any{`{"name":"Ada"}`, `plain text`, 42.5, true} {
		dependencyValid := schema.Validate(input).IsValid()
		_, parseErr := imported.ParseAny(input)
		assert.Equal(t, dependencyValid, parseErr == nil, "input %#v", input)
	}
}

func TestFromJSONSchema_RejectsUnanchoredStringAssertion(t *testing.T) {
	schema := compileImportSchema(t, `{"minLength": 2}`)
	require.False(t, schema.Validate("a").IsValid())
	require.True(t, schema.Validate("ab").IsValid())
	require.True(t, schema.Validate(42).IsValid())

	imported, err := FromJSONSchema(schema)
	assert.Nil(t, imported)
	assert.ErrorIs(t, err, ErrUnsupportedJSONSchemaKeyword)
	var importErr *ImportError
	require.ErrorAs(t, err, &importErr)
	assert.Equal(t, "minLength", importErr.Keyword)
	assert.Equal(t, "/minLength", importErr.Pointer)
}

func TestFromJSONSchema_RejectsEveryUnanchoredStringAssertion(t *testing.T) {
	tests := []struct {
		keyword string
		schema  *lib.Schema
	}{
		{keyword: "maxLength", schema: &lib.Schema{MaxLength: new(float64(2))}},
		{keyword: "pattern", schema: &lib.Schema{Pattern: new("^x")}},
	}

	for _, test := range tests {
		t.Run(test.keyword, func(t *testing.T) {
			imported, err := FromJSONSchema(test.schema)
			assert.Nil(t, imported)
			var importErr *ImportError
			require.ErrorAs(t, err, &importErr)
			assert.Equal(t, test.keyword, importErr.Keyword)
			assert.Equal(t, "/"+test.keyword, importErr.Pointer)
		})
	}
}

func TestFromJSONSchema_RejectsUnanchoredNumericAssertion(t *testing.T) {
	schema := compileImportSchema(t, `{"minimum": 1}`)
	require.False(t, schema.Validate(0).IsValid())
	require.True(t, schema.Validate(1).IsValid())
	require.True(t, schema.Validate("zero").IsValid())

	imported, err := FromJSONSchema(schema)
	assert.Nil(t, imported)
	var importErr *ImportError
	require.ErrorAs(t, err, &importErr)
	assert.Equal(t, "minimum", importErr.Keyword)
	assert.Equal(t, "/minimum", importErr.Pointer)
}

func TestFromJSONSchema_RejectsEveryUnanchoredNumericAssertion(t *testing.T) {
	tests := []struct {
		keyword string
		schema  *lib.Schema
	}{
		{keyword: "maximum", schema: &lib.Schema{Maximum: lib.NewRat(1)}},
		{keyword: "exclusiveMinimum", schema: &lib.Schema{ExclusiveMinimum: lib.NewRat(1)}},
		{keyword: "exclusiveMaximum", schema: &lib.Schema{ExclusiveMaximum: lib.NewRat(1)}},
		{keyword: "multipleOf", schema: &lib.Schema{MultipleOf: lib.NewRat(2)}},
	}

	for _, test := range tests {
		t.Run(test.keyword, func(t *testing.T) {
			imported, err := FromJSONSchema(test.schema)
			assert.Nil(t, imported)
			var importErr *ImportError
			require.ErrorAs(t, err, &importErr)
			assert.Equal(t, test.keyword, importErr.Keyword)
			assert.Equal(t, "/"+test.keyword, importErr.Pointer)
		})
	}
}

func TestFromJSONSchema_RejectsUnanchoredArrayAssertion(t *testing.T) {
	schema := compileImportSchema(t, `{"minItems": 2}`)
	require.False(t, schema.Validate([]any{1}).IsValid())
	require.True(t, schema.Validate([]any{1, 2}).IsValid())
	require.True(t, schema.Validate("item").IsValid())

	imported, err := FromJSONSchema(schema)
	assert.Nil(t, imported)
	var importErr *ImportError
	require.ErrorAs(t, err, &importErr)
	assert.Equal(t, "minItems", importErr.Keyword)
	assert.Equal(t, "/minItems", importErr.Pointer)
}

func TestFromJSONSchema_RejectsEveryUnanchoredArrayAssertion(t *testing.T) {
	tests := []struct {
		keyword string
		schema  *lib.Schema
	}{
		{keyword: "maxItems", schema: &lib.Schema{MaxItems: new(float64(2))}},
		{keyword: "prefixItems", schema: &lib.Schema{PrefixItems: []*lib.Schema{{Boolean: new(true)}}}},
		{keyword: "items", schema: &lib.Schema{Items: &lib.Schema{Boolean: new(false)}}},
	}

	for _, test := range tests {
		t.Run(test.keyword, func(t *testing.T) {
			imported, err := FromJSONSchema(test.schema)
			assert.Nil(t, imported)
			var importErr *ImportError
			require.ErrorAs(t, err, &importErr)
			assert.Equal(t, test.keyword, importErr.Keyword)
			assert.Equal(t, "/"+test.keyword, importErr.Pointer)
		})
	}
}

func TestFromJSONSchema_RejectsUnanchoredObjectAssertion(t *testing.T) {
	schema := compileImportSchema(t, `{"required": ["name"]}`)
	require.False(t, schema.Validate(map[string]any{}).IsValid())
	require.True(t, schema.Validate(map[string]any{"name": "Ada"}).IsValid())
	require.True(t, schema.Validate("Ada").IsValid())

	imported, err := FromJSONSchema(schema)
	assert.Nil(t, imported)
	var importErr *ImportError
	require.ErrorAs(t, err, &importErr)
	assert.Equal(t, "required", importErr.Keyword)
	assert.Equal(t, "/required", importErr.Pointer)
}

func TestFromJSONSchema_RejectsEveryUnanchoredObjectAssertion(t *testing.T) {
	tests := []struct {
		keyword string
		schema  *lib.Schema
	}{
		{
			keyword: "properties",
			schema:  &lib.Schema{Properties: &lib.SchemaMap{"name": {Type: []string{"string"}}}},
		},
		{
			keyword: "additionalProperties",
			schema:  &lib.Schema{AdditionalProperties: &lib.Schema{Boolean: new(false)}},
		},
		{keyword: "minProperties", schema: &lib.Schema{MinProperties: new(float64(1))}},
		{keyword: "maxProperties", schema: &lib.Schema{MaxProperties: new(float64(1))}},
	}

	for _, test := range tests {
		t.Run(test.keyword, func(t *testing.T) {
			imported, err := FromJSONSchema(test.schema)
			assert.Nil(t, imported)
			var importErr *ImportError
			require.ErrorAs(t, err, &importErr)
			assert.Equal(t, test.keyword, importErr.Keyword)
			assert.Equal(t, "/"+test.keyword, importErr.Pointer)
		})
	}
}

func TestFromJSONSchema_UnanchoredAssertionErrorCarriesNestedPointer(t *testing.T) {
	schema := &lib.Schema{
		Type: []string{"object"},
		Properties: &lib.SchemaMap{
			"value": {MinLength: new(float64(2))},
		},
	}

	imported, err := FromJSONSchema(schema)
	assert.Nil(t, imported)
	var importErr *ImportError
	require.ErrorAs(t, err, &importErr)
	assert.Equal(t, "minLength", importErr.Keyword)
	assert.Equal(t, "/properties/value/minLength", importErr.Pointer)
}

func TestFromJSONSchema_LossyRecordsEveryUnanchoredAssertion(t *testing.T) {
	schema := &lib.Schema{
		MinLength:     new(float64(1)),
		Minimum:       lib.NewRat(1),
		MinItems:      new(float64(1)),
		MinProperties: new(float64(1)),
	}
	imported, losses, err := FromJSONSchemaLossy(schema)
	require.NoError(t, err)
	require.NotNil(t, imported)
	assert.Equal(t, []string{"minItems", "minLength", "minProperties", "minimum"}, importLossKeywords(losses))
}

func TestFromJSONSchema_AllowsAnnotationOnlyAndExistingAssertionBases(t *testing.T) {
	metadata := core.NewRegistry[core.GlobalMeta]()
	tests := []struct {
		name   string
		schema *lib.Schema
	}{
		{
			name: "annotations only",
			schema: &lib.Schema{
				Title:       new("Value"),
				Description: new("Any annotated value"),
				Examples:    []any{"example"},
			},
		},
		{name: "const", schema: lib.Const("fixed")},
		{name: "enum", schema: &lib.Schema{Enum: []any{"a", "b"}}},
		{name: "composition", schema: &lib.Schema{AllOf: []*lib.Schema{{Type: []string{"string"}}}}},
		{
			name: "resolved ref",
			schema: &lib.Schema{
				Ref:         "#/$defs/value",
				ResolvedRef: &lib.Schema{Type: []string{"string"}},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			imported, err := FromJSONSchema(test.schema, FromJSONSchemaOptions{Metadata: metadata})
			require.NoError(t, err)
			require.NotNil(t, imported)
		})
	}
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

func TestFromJSONSchema_IntegerMinimumRejectsOutOfRangeOperand(t *testing.T) {
	minimum := "2147483648"
	if strconv.IntSize == 64 {
		minimum = "9223372036854775808"
	}
	schema := &lib.Schema{
		Type:    []string{"integer"},
		Minimum: lib.NewRat(minimum),
	}

	imported, err := FromJSONSchema(schema)
	assert.Nil(t, imported)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrInvalidJSONSchema)

	importErr, ok := errors.AsType[*ImportError](err)
	require.True(t, ok)
	assert.Equal(t, "minimum", importErr.Keyword)
	assert.Equal(t, "/minimum", importErr.Pointer)
}

func TestFromJSONSchema_IntegerMaximumRejectsOutOfRangeOperand(t *testing.T) {
	maximum := "-2147483649"
	if strconv.IntSize == 64 {
		maximum = "-9223372036854775809"
	}
	schema := &lib.Schema{
		Type:    []string{"integer"},
		Maximum: lib.NewRat(maximum),
	}

	imported, err := FromJSONSchema(schema)
	assert.Nil(t, imported)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrInvalidJSONSchema)

	var importErr *ImportError
	require.True(t, errors.As(err, &importErr))
	assert.Equal(t, "maximum", importErr.Keyword)
	assert.Equal(t, "/maximum", importErr.Pointer)
}

func TestFromJSONSchema_IntegerExclusiveMinimumRejectsOverflowingAdjustment(t *testing.T) {
	exclusiveMinimum := "2147483647"
	if strconv.IntSize == 64 {
		exclusiveMinimum = "9223372036854775807"
	}
	schema := &lib.Schema{
		Type:             []string{"integer"},
		ExclusiveMinimum: lib.NewRat(exclusiveMinimum),
	}

	imported, err := FromJSONSchema(schema)
	assert.Nil(t, imported)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrInvalidJSONSchema)

	var importErr *ImportError
	require.True(t, errors.As(err, &importErr))
	assert.Equal(t, "exclusiveMinimum", importErr.Keyword)
	assert.Equal(t, "/exclusiveMinimum", importErr.Pointer)
}

func TestFromJSONSchema_IntegerExclusiveMaximumRejectsOverflowingAdjustment(t *testing.T) {
	exclusiveMaximum := "-2147483648"
	if strconv.IntSize == 64 {
		exclusiveMaximum = "-9223372036854775808"
	}
	schema := &lib.Schema{
		Type:             []string{"integer"},
		ExclusiveMaximum: lib.NewRat(exclusiveMaximum),
	}

	imported, err := FromJSONSchema(schema)
	assert.Nil(t, imported)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrInvalidJSONSchema)

	var importErr *ImportError
	require.True(t, errors.As(err, &importErr))
	assert.Equal(t, "exclusiveMaximum", importErr.Keyword)
	assert.Equal(t, "/exclusiveMaximum", importErr.Pointer)
}

func TestFromJSONSchema_IntegerMultipleOfRejectsOutOfRangeOperand(t *testing.T) {
	multipleOf := "2147483648"
	if strconv.IntSize == 64 {
		multipleOf = "9223372036854775808"
	}
	schema := &lib.Schema{
		Type:       []string{"integer"},
		MultipleOf: lib.NewRat(multipleOf),
	}

	imported, err := FromJSONSchema(schema)
	assert.Nil(t, imported)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrInvalidJSONSchema)

	var importErr *ImportError
	require.True(t, errors.As(err, &importErr))
	assert.Equal(t, "multipleOf", importErr.Keyword)
	assert.Equal(t, "/multipleOf", importErr.Pointer)
}

func TestFromJSONSchema_RejectsNonPositiveMultipleOf(t *testing.T) {
	tests := []struct {
		name    string
		schema  *lib.Schema
		lossy   bool
		pointer string
	}{
		{
			name:    "strict number zero at root",
			schema:  &lib.Schema{Type: []string{"number"}, MultipleOf: lib.NewRat(0)},
			pointer: "/multipleOf",
		},
		{
			name: "strict integer negative in property",
			schema: &lib.Schema{
				Type: []string{"object"},
				Properties: &lib.SchemaMap{
					"count": {Type: []string{"integer"}, MultipleOf: lib.NewRat(-1)},
				},
			},
			pointer: "/properties/count/multipleOf",
		},
		{
			name: "lossy number negative in items",
			schema: &lib.Schema{
				Type:  []string{"array"},
				Items: &lib.Schema{Type: []string{"number"}, MultipleOf: lib.NewRat(-1)},
			},
			lossy:   true,
			pointer: "/items/multipleOf",
		},
		{
			name:    "lossy integer zero at root",
			schema:  &lib.Schema{Type: []string{"integer"}, MultipleOf: lib.NewRat(0)},
			lossy:   true,
			pointer: "/multipleOf",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var (
				imported core.ZodSchema
				losses   []ImportLossError
				err      error
			)
			if test.lossy {
				imported, losses, err = FromJSONSchemaLossy(test.schema)
			} else {
				imported, err = FromJSONSchema(test.schema)
			}

			assert.Nil(t, imported)
			assert.Empty(t, losses)
			require.ErrorIs(t, err, ErrInvalidJSONSchema)
			var importErr *ImportError
			require.ErrorAs(t, err, &importErr)
			assert.Equal(t, "multipleOf", importErr.Keyword)
			assert.Equal(t, test.pointer, importErr.Pointer)
		})
	}
}

func TestFromJSONSchema_IntegerBoundsPreserveRepresentablePlatformEdges(t *testing.T) {
	maxInt := int(^uint(0) >> 1)
	minInt := -maxInt - 1
	tests := []struct {
		name    string
		schema  *lib.Schema
		valid   int
		invalid int
	}{
		{
			name:    "minimum",
			schema:  &lib.Schema{Type: []string{"integer"}, Minimum: lib.NewRat(strconv.FormatInt(int64(maxInt), 10))},
			valid:   maxInt,
			invalid: maxInt - 1,
		},
		{
			name:    "maximum",
			schema:  &lib.Schema{Type: []string{"integer"}, Maximum: lib.NewRat(strconv.FormatInt(int64(minInt), 10))},
			valid:   minInt,
			invalid: minInt + 1,
		},
		{
			name: "exclusiveMinimum",
			schema: &lib.Schema{
				Type:             []string{"integer"},
				ExclusiveMinimum: lib.NewRat(strconv.FormatInt(int64(maxInt-1), 10)),
			},
			valid:   maxInt,
			invalid: maxInt - 1,
		},
		{
			name: "exclusiveMaximum",
			schema: &lib.Schema{
				Type:             []string{"integer"},
				ExclusiveMaximum: lib.NewRat(strconv.FormatInt(int64(minInt+1), 10)),
			},
			valid:   minInt,
			invalid: minInt + 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			imported, err := FromJSONSchema(test.schema)
			require.NoError(t, err)

			result, err := imported.ParseAny(test.valid)
			require.NoError(t, err)
			assert.Equal(t, test.valid, result)

			_, err = imported.ParseAny(test.invalid)
			require.Error(t, err)
		})
	}
}

func TestFromJSONSchema_IntegerConstraintsPreserveRationalSemantics(t *testing.T) {
	tests := []struct {
		name    string
		schema  *lib.Schema
		valid   int
		invalid int
	}{
		{
			name:    "minimum rounds up",
			schema:  &lib.Schema{Type: []string{"integer"}, Minimum: lib.NewRat("3/2")},
			valid:   2,
			invalid: 1,
		},
		{
			name:    "maximum rounds down",
			schema:  &lib.Schema{Type: []string{"integer"}, Maximum: lib.NewRat("3/2")},
			valid:   1,
			invalid: 2,
		},
		{
			name: "exclusive minimum rounds up",
			schema: &lib.Schema{
				Type:             []string{"integer"},
				ExclusiveMinimum: lib.NewRat("3/2"),
			},
			valid:   2,
			invalid: 1,
		},
		{
			name: "exclusive maximum rounds down",
			schema: &lib.Schema{
				Type:             []string{"integer"},
				ExclusiveMaximum: lib.NewRat("3/2"),
			},
			valid:   1,
			invalid: 2,
		},
		{
			name:    "multipleOf uses reduced numerator",
			schema:  &lib.Schema{Type: []string{"integer"}, MultipleOf: lib.NewRat("3/2")},
			valid:   6,
			invalid: 4,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			imported, err := FromJSONSchema(test.schema)
			require.NoError(t, err)

			_, err = imported.ParseAny(test.valid)
			require.NoError(t, err)
			_, err = imported.ParseAny(test.invalid)
			require.Error(t, err)
		})
	}
}

func TestFromJSONSchema_IntegerOverflowCarriesNestedPointer(t *testing.T) {
	maximum := "-2147483649"
	if strconv.IntSize == 64 {
		maximum = "-9223372036854775809"
	}
	schema := &lib.Schema{
		Type: []string{"object"},
		Properties: &lib.SchemaMap{
			"limit/value": {
				Type:    []string{"integer"},
				Maximum: lib.NewRat(maximum),
			},
		},
	}

	imported, err := FromJSONSchema(schema)
	assert.Nil(t, imported)
	require.ErrorIs(t, err, ErrInvalidJSONSchema)
	var importErr *ImportError
	require.ErrorAs(t, err, &importErr)
	assert.Equal(t, "maximum", importErr.Keyword)
	assert.Equal(t, "/properties/limit~1value/maximum", importErr.Pointer)
}

func TestFromJSONSchemaLossy_IntegerOverflowRemainsAnError(t *testing.T) {
	multipleOf := "2147483648"
	if strconv.IntSize == 64 {
		multipleOf = "9223372036854775808"
	}
	schema := &lib.Schema{Type: []string{"integer"}, MultipleOf: lib.NewRat(multipleOf)}

	imported, losses, err := FromJSONSchemaLossy(schema)
	assert.Nil(t, imported)
	assert.Empty(t, losses)
	require.ErrorIs(t, err, ErrInvalidJSONSchema)
	var importErr *ImportError
	require.ErrorAs(t, err, &importErr)
	assert.Equal(t, "multipleOf", importErr.Keyword)
	assert.Equal(t, "/multipleOf", importErr.Pointer)
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

func TestFromJSONSchema_ObjectPropertyCountAssertions(t *testing.T) {
	schema := &lib.Schema{
		Type: []string{"object"},
		Properties: &lib.SchemaMap{
			"a": {Type: []string{"string"}},
			"b": {Type: []string{"string"}},
			"c": {Type: []string{"string"}},
		},
		MinProperties: new(float64(1)),
		MaxProperties: new(float64(2)),
	}

	imported, err := FromJSONSchema(schema)
	require.NoError(t, err)
	for _, input := range []any{
		map[string]any{},
		map[string]any{"a": "a"},
		map[string]any{"a": "a", "b": "b"},
		map[string]any{"a": "a", "b": "b", "c": "c"},
	} {
		dependencyValid := schema.Validate(input).IsValid()
		_, parseErr := imported.ParseAny(input)
		assert.Equal(t, dependencyValid, parseErr == nil, "input %#v", input)
	}
}

func TestFromJSONSchema_RejectsDependentRequired(t *testing.T) {
	schema := &lib.Schema{
		Type: []string{"object"},
		Properties: &lib.SchemaMap{
			"account": {
				Type: []string{"object"},
				Properties: &lib.SchemaMap{
					"credit_card":     {Type: []string{"string"}},
					"billing_address": {Type: []string{"string"}},
				},
				DependentRequired: map[string][]string{"credit_card": {"billing_address"}},
			},
		},
	}

	imported, err := FromJSONSchema(schema)
	assert.Nil(t, imported)
	assert.ErrorIs(t, err, ErrUnsupportedJSONSchemaKeyword)
	var importErr *ImportError
	require.ErrorAs(t, err, &importErr)
	assert.Equal(t, "dependentRequired", importErr.Keyword)
	assert.Equal(t, "/properties/account/dependentRequired", importErr.Pointer)
}

func TestFromJSONSchema_NestedObjectPropertyCountIssuePath(t *testing.T) {
	schema := &lib.Schema{
		Type: []string{"object"},
		Properties: &lib.SchemaMap{
			"profile": {
				Type: []string{"object"},
				Properties: &lib.SchemaMap{
					"name": {Type: []string{"string"}},
				},
				MinProperties: new(float64(1)),
			},
		},
		Required: []string{"profile"},
	}
	input := map[string]any{"profile": map[string]any{}}

	require.False(t, schema.Validate(input).IsValid())
	imported, err := FromJSONSchema(schema)
	require.NoError(t, err)
	_, err = imported.ParseAny(input)
	assert.ErrorContains(t, err, "profile")
}

func TestFromJSONSchema_RecordLikePropertyCountAssertions(t *testing.T) {
	schema := &lib.Schema{
		Type:                 []string{"object"},
		AdditionalProperties: &lib.Schema{Type: []string{"string"}},
		MinProperties:        new(float64(1)),
		MaxProperties:        new(float64(2)),
	}

	imported, err := FromJSONSchema(schema)
	require.NoError(t, err)
	for _, input := range []any{
		map[string]any{},
		map[string]any{"a": "a"},
		map[string]any{"a": "a", "b": "b", "c": "c"},
	} {
		dependencyValid := schema.Validate(input).IsValid()
		_, parseErr := imported.ParseAny(input)
		assert.Equal(t, dependencyValid, parseErr == nil, "input %#v", input)
	}
}

func TestFromJSONSchema_RoundTripsExportedRecord(t *testing.T) {
	original := types.Record(types.String(), types.Int())
	exported, err := ToJSONSchema(original)
	require.NoError(t, err)
	require.NotNil(t, exported.PropertyNames)
	require.NotNil(t, exported.AdditionalProperties)

	inputs := []any{
		map[string]any{"a": 1, "b": 2},
		map[string]any{},
		map[string]any{"a": "one"},
		"not an object",
	}
	for _, input := range inputs {
		dependencyValid := exported.Validate(input).IsValid()
		_, originalErr := original.ParseAny(input)
		assert.Equal(t, dependencyValid, originalErr == nil, "original input %#v", input)
	}

	imported, err := FromJSONSchema(exported)
	require.NoError(t, err)
	for _, input := range inputs {
		dependencyValid := exported.Validate(input).IsValid()
		_, parseErr := imported.ParseAny(input)
		assert.Equal(t, dependencyValid, parseErr == nil, "imported input %#v", input)
	}
}

func TestFromJSONSchema_RoundTripsExportedRegexRecord(t *testing.T) {
	original := types.Record(
		types.String().Regex(regexp.MustCompile("^[a-z]+$")),
		types.Int(),
	)
	exported, err := ToJSONSchema(original)
	require.NoError(t, err)
	imported, err := FromJSONSchema(exported)
	require.NoError(t, err)

	for _, input := range []any{
		map[string]any{"valid": 1},
		map[string]any{"INVALID": 1},
		map[string]any{"valid": "one"},
	} {
		dependencyValid := exported.Validate(input).IsValid()
		_, originalErr := original.ParseAny(input)
		_, importedErr := imported.ParseAny(input)
		assert.Equal(t, dependencyValid, originalErr == nil, "original input %#v", input)
		assert.Equal(t, dependencyValid, importedErr == nil, "imported input %#v", input)
	}
}

func TestFromJSONSchema_RoundTripsExportedLooseRecord(t *testing.T) {
	original := types.LooseRecord(
		types.String().Regex(regexp.MustCompile("^[a-z]+$")),
		types.Int(),
	)
	exported, err := ToJSONSchema(original)
	require.NoError(t, err)
	require.NotNil(t, exported.PatternProperties)
	require.Len(t, *exported.PatternProperties, 1)

	inputs := []any{
		map[string]any{"valid": 1},
		map[string]any{"valid": "one"},
		map[string]any{"UNMATCHED": "one"},
		map[string]any{"valid": 1, "UNMATCHED": "one"},
		"not an object",
	}
	for _, input := range inputs {
		dependencyValid := exported.Validate(input).IsValid()
		_, originalErr := original.ParseAny(input)
		assert.Equal(t, dependencyValid, originalErr == nil, "original input %#v", input)
	}

	imported, err := FromJSONSchema(exported)
	require.NoError(t, err)
	for _, input := range inputs {
		dependencyValid := exported.Validate(input).IsValid()
		result, parseErr := imported.ParseAny(input)
		assert.Equal(t, dependencyValid, parseErr == nil, "imported input %#v", input)
		if parseErr == nil {
			assert.Equal(t, input, result)
		}
	}
}

func TestFromJSONSchema_InvalidPatternPropertyReportsKeywordAndPointer(t *testing.T) {
	patterns := lib.SchemaMap{"[": {Type: []string{"string"}}}
	schema := &lib.Schema{
		Type:              []string{"object"},
		PatternProperties: &patterns,
	}

	imported, err := FromJSONSchema(schema)
	assert.Nil(t, imported)
	assert.ErrorIs(t, err, ErrJSONSchemaPatternCompile)
	var importErr *ImportError
	require.ErrorAs(t, err, &importErr)
	assert.Equal(t, "patternProperties", importErr.Keyword)
	assert.Equal(t, "/patternProperties/[", importErr.Pointer)
}

func TestFromJSONSchema_RejectsExhaustivePropertyNamesRecord(t *testing.T) {
	schema := &lib.Schema{
		Type: []string{"object"},
		PropertyNames: &lib.Schema{
			Type: []string{"string"},
			Enum: []any{"a", "b"},
		},
		AdditionalProperties: &lib.Schema{Type: []string{"integer"}},
	}

	imported, err := FromJSONSchema(schema)
	assert.Nil(t, imported)
	assert.ErrorIs(t, err, ErrJSONSchemaPropertyNames)
	var importErr *ImportError
	require.ErrorAs(t, err, &importErr)
	assert.Equal(t, "propertyNames", importErr.Keyword)
	assert.Equal(t, "/propertyNames", importErr.Pointer)
}

func TestFromJSONSchema_RejectsConflictingRecordShapes(t *testing.T) {
	stringNames := func() *lib.Schema { return &lib.Schema{Type: []string{"string"}} }
	valueSchema := func() *lib.Schema { return &lib.Schema{Type: []string{"integer"}} }
	tests := []struct {
		name    string
		schema  *lib.Schema
		keyword string
	}{
		{
			name: "propertyNames with named properties",
			schema: &lib.Schema{
				Type:                 []string{"object"},
				Properties:           &lib.SchemaMap{"name": {Type: []string{"string"}}},
				PropertyNames:        stringNames(),
				AdditionalProperties: valueSchema(),
			},
			keyword: "propertyNames",
		},
		{
			name: "non-string propertyNames",
			schema: &lib.Schema{
				Type:                 []string{"object"},
				PropertyNames:        &lib.Schema{Type: []string{"number"}},
				AdditionalProperties: valueSchema(),
			},
			keyword: "propertyNames",
		},
		{
			name: "multiple patterns",
			schema: &lib.Schema{
				Type: []string{"object"},
				PatternProperties: &lib.SchemaMap{
					"^a": valueSchema(),
					"^b": valueSchema(),
				},
			},
			keyword: "patternProperties",
		},
		{
			name: "pattern with additionalProperties",
			schema: &lib.Schema{
				Type:                 []string{"object"},
				PatternProperties:    &lib.SchemaMap{"^a": valueSchema()},
				AdditionalProperties: &lib.Schema{Boolean: new(false)},
			},
			keyword: "patternProperties",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			imported, err := FromJSONSchema(test.schema)
			assert.Nil(t, imported)
			var importErr *ImportError
			require.ErrorAs(t, err, &importErr)
			assert.Equal(t, test.keyword, importErr.Keyword)
		})
	}
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
	var importErr *ImportError
	require.ErrorAs(t, err, &importErr)
	assert.Equal(t, "pattern", importErr.Keyword)
	assert.Equal(t, "/properties/code/pattern", importErr.Pointer)
	var syntaxErr *syntax.Error
	assert.ErrorAs(t, err, &syntaxErr)
}

func TestFromJSONSchema_NestedImportErrorCarriesJSONPointer(t *testing.T) {
	schema := &lib.Schema{
		Type: []string{"object"},
		Properties: &lib.SchemaMap{
			"a/b~c": {Not: &lib.Schema{Boolean: new(true)}},
		},
	}

	imported, err := FromJSONSchema(schema)
	assert.Nil(t, imported)
	assert.ErrorIs(t, err, ErrUnsupportedJSONSchemaKeyword)
	var importErr *ImportError
	require.True(t, errors.As(err, &importErr))
	assert.Equal(t, "not", importErr.Keyword)
	assert.Equal(t, "/properties/a~1b~0c/not", importErr.Pointer)
}

func TestFromJSONSchema_TypeErrorCarriesJSONPointer(t *testing.T) {
	schema := &lib.Schema{Type: []string{"imaginary"}}

	imported, err := FromJSONSchema(schema)
	assert.Nil(t, imported)
	assert.ErrorIs(t, err, ErrUnsupportedJSONSchemaType)
	var importErr *ImportError
	require.ErrorAs(t, err, &importErr)
	assert.Equal(t, "type", importErr.Keyword)
	assert.Equal(t, "/type", importErr.Pointer)
}

func TestFromJSONSchema_RecursiveImportErrorPointers(t *testing.T) {
	unsupported := func() *lib.Schema {
		return &lib.Schema{Not: &lib.Schema{Boolean: new(true)}}
	}
	tests := []struct {
		name    string
		schema  *lib.Schema
		pointer string
	}{
		{
			name:    "items",
			schema:  &lib.Schema{Type: []string{"array"}, Items: unsupported()},
			pointer: "/items/not",
		},
		{
			name: "prefixItems",
			schema: &lib.Schema{
				Type:        []string{"array"},
				PrefixItems: []*lib.Schema{{Type: []string{"string"}}, unsupported()},
			},
			pointer: "/prefixItems/1/not",
		},
		{
			name:    "allOf",
			schema:  &lib.Schema{AllOf: []*lib.Schema{unsupported()}},
			pointer: "/allOf/0/not",
		},
		{
			name:    "anyOf",
			schema:  &lib.Schema{AnyOf: []*lib.Schema{unsupported()}},
			pointer: "/anyOf/0/not",
		},
		{
			name:    "oneOf",
			schema:  &lib.Schema{OneOf: []*lib.Schema{unsupported()}},
			pointer: "/oneOf/0/not",
		},
		{
			name: "$ref",
			schema: &lib.Schema{
				Ref:         "#/$defs/value",
				ResolvedRef: unsupported(),
			},
			pointer: "/$ref/not",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			imported, err := FromJSONSchema(test.schema)
			assert.Nil(t, imported)
			var importErr *ImportError
			require.ErrorAs(t, err, &importErr)
			assert.Equal(t, "not", importErr.Keyword)
			assert.Equal(t, test.pointer, importErr.Pointer)
		})
	}
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

func TestFromJSONSchema_ConjoinsResolvedRefAndStringSibling(t *testing.T) {
	target := &lib.Schema{Type: []string{"string"}}
	schema := &lib.Schema{
		Ref:         "#/$defs/name",
		ResolvedRef: target,
		MinLength:   new(float64(3)),
	}

	imported, err := FromJSONSchema(schema)
	require.NoError(t, err)
	for _, input := range []any{"Ada", "Al", 1} {
		dependencyValid := schema.Validate(input).IsValid()
		_, parseErr := imported.ParseAny(input)
		assert.Equal(t, dependencyValid, parseErr == nil, "input %#v", input)
	}
}

func TestFromJSONSchema_RecursiveResolvedRefUsesCompletedSchema(t *testing.T) {
	node := &lib.Schema{
		Type: []string{"object"},
		Properties: &lib.SchemaMap{
			"value": {Type: []string{"string"}, MinLength: new(float64(2))},
		},
		Required: []string{"value"},
	}
	(*node.Properties)["next"] = &lib.Schema{Ref: "#", ResolvedRef: node}

	imported, err := FromJSONSchema(node)
	require.NoError(t, err)
	input := map[string]any{
		"value": "root",
		"next":  map[string]any{"value": "x"},
	}
	require.False(t, node.Validate(input).IsValid())
	_, err = imported.ParseAny(input)
	assert.Error(t, err)
}

func TestFromJSONSchema_RejectsDirectSelfReference(t *testing.T) {
	schema := &lib.Schema{Ref: "#"}
	schema.ResolvedRef = schema

	imported, err := FromJSONSchema(schema)
	assert.Nil(t, imported)
	assert.ErrorIs(t, err, ErrJSONSchemaCircularRef)
}

func TestFromJSONSchema_ConjoinsResolvedRefObjectSibling(t *testing.T) {
	target := &lib.Schema{
		Type: []string{"object"},
		Properties: &lib.SchemaMap{
			"id": {Type: []string{"string"}},
		},
		Required: []string{"id"},
	}
	schema := &lib.Schema{
		Ref:         "#/$defs/base",
		ResolvedRef: target,
		Properties: &lib.SchemaMap{
			"name": {Type: []string{"string"}, MinLength: new(float64(2))},
		},
		Required: []string{"name"},
	}

	imported, err := FromJSONSchema(schema)
	require.NoError(t, err)
	for _, input := range []any{
		map[string]any{"id": "1", "name": "Ada"},
		map[string]any{"id": "1"},
		map[string]any{"name": "Ada"},
		map[string]any{"id": "1", "name": "A"},
	} {
		dependencyValid := schema.Validate(input).IsValid()
		_, parseErr := imported.ParseAny(input)
		assert.Equal(t, dependencyValid, parseErr == nil, "input %#v", input)
	}
}

func TestFromJSONSchema_ConjoinsResolvedRefAndAllOfSibling(t *testing.T) {
	target := &lib.Schema{Type: []string{"string"}}
	schema := &lib.Schema{
		Ref:         "#/$defs/base",
		ResolvedRef: target,
		AllOf: []*lib.Schema{{
			Type:      []string{"string"},
			MinLength: new(float64(3)),
		}},
	}

	imported, err := FromJSONSchema(schema)
	require.NoError(t, err)
	for _, input := range []any{"Ada", "Al", 1} {
		dependencyValid := schema.Validate(input).IsValid()
		_, parseErr := imported.ParseAny(input)
		assert.Equal(t, dependencyValid, parseErr == nil, "input %#v", input)
	}
}

func TestFromJSONSchema_SharedResolvedRef(t *testing.T) {
	target := &lib.Schema{Type: []string{"string"}, MinLength: new(float64(2))}
	ref := func() *lib.Schema {
		return &lib.Schema{Ref: "#/$defs/name", ResolvedRef: target}
	}
	schema := &lib.Schema{
		Type: []string{"object"},
		Properties: &lib.SchemaMap{
			"first": ref(),
			"last":  ref(),
		},
		Required: []string{"first", "last"},
	}

	imported, err := FromJSONSchema(schema)
	require.NoError(t, err)
	for _, input := range []any{
		map[string]any{"first": "Ada", "last": "Li"},
		map[string]any{"first": "A", "last": "Li"},
		map[string]any{"first": "Ada", "last": "L"},
	} {
		dependencyValid := schema.Validate(input).IsValid()
		_, parseErr := imported.ParseAny(input)
		assert.Equal(t, dependencyValid, parseErr == nil, "input %#v", input)
	}
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

func TestFromJSONSchema_ConjoinsIncompatibleTypeAndConst(t *testing.T) {
	schema := lib.Const(1)
	schema.Type = []string{"string"}

	require.False(t, schema.Validate(1).IsValid())

	imported, err := FromJSONSchema(schema)
	require.NoError(t, err)
	_, err = imported.ParseAny(1)
	assert.Error(t, err)
}

func TestFromJSONSchema_ConjoinsBasicAssertionSiblings(t *testing.T) {
	tests := []struct {
		name   string
		schema func() *lib.Schema
		inputs []any
		valid  []bool
	}{
		{
			name: "compatible type and const",
			schema: func() *lib.Schema {
				schema := lib.Const("fixed")
				schema.Type = []string{"string"}
				return schema
			},
			inputs: []any{"fixed", "other", 1},
			valid:  []bool{true, false, false},
		},
		{
			name: "type filters mixed enum",
			schema: func() *lib.Schema {
				return &lib.Schema{Type: []string{"string"}, Enum: []any{"fixed", 1}}
			},
			inputs: []any{"fixed", 1, "other"},
			valid:  []bool{true, false, false},
		},
		{
			name: "compatible const and enum",
			schema: func() *lib.Schema {
				schema := lib.Const("fixed")
				schema.Enum = []any{"fixed", "other"}
				return schema
			},
			inputs: []any{"fixed", "other"},
			valid:  []bool{true, false},
		},
		{
			name: "incompatible const and enum",
			schema: func() *lib.Schema {
				schema := lib.Const("fixed")
				schema.Enum = []any{"other"}
				return schema
			},
			inputs: []any{"fixed", "other"},
			valid:  []bool{false, false},
		},
		{
			name: "all three compatible",
			schema: func() *lib.Schema {
				schema := lib.Const("fixed")
				schema.Type = []string{"string"}
				schema.Enum = []any{"fixed", "other"}
				return schema
			},
			inputs: []any{"fixed", "other", 1},
			valid:  []bool{true, false, false},
		},
		{
			name: "all three incompatible",
			schema: func() *lib.Schema {
				schema := lib.Const(1)
				schema.Type = []string{"string"}
				schema.Enum = []any{1, "other"}
				return schema
			},
			inputs: []any{1, "other"},
			valid:  []bool{false, false},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			schema := test.schema()
			imported, err := FromJSONSchema(schema)
			require.NoError(t, err)

			for i, input := range test.inputs {
				dependencyValid := schema.Validate(input).IsValid()
				assert.Equal(t, test.valid[i], dependencyValid, "dependency result for %#v", input)

				_, parseErr := imported.ParseAny(input)
				assert.Equal(t, dependencyValid, parseErr == nil, "imported result for %#v", input)
			}
		})
	}
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

func TestFromJSONSchema_ConjoinsBaseAndAnyOf(t *testing.T) {
	schema := &lib.Schema{
		Type:      []string{"string"},
		MinLength: new(float64(3)),
		AnyOf: []*lib.Schema{
			lib.Const("ab"),
			lib.Const("valid"),
			lib.Const(1),
		},
	}

	imported, err := FromJSONSchema(schema)
	require.NoError(t, err)
	for _, input := range []any{"valid", "ab", 1, "other"} {
		dependencyValid := schema.Validate(input).IsValid()
		_, parseErr := imported.ParseAny(input)
		assert.Equal(t, dependencyValid, parseErr == nil, "input %#v", input)
	}
}

func TestFromJSONSchema_ConjoinsCompositionSiblings(t *testing.T) {
	allOf := []*lib.Schema{{Enum: []any{"a", "b"}}}
	anyOf := []*lib.Schema{lib.Const("a"), lib.Const("c")}
	oneOf := []*lib.Schema{{Enum: []any{"a", "b"}}, lib.Const("b")}
	tests := []struct {
		name   string
		allOf  []*lib.Schema
		anyOf  []*lib.Schema
		oneOf  []*lib.Schema
		inputs []any
	}{
		{name: "allOf and anyOf", allOf: allOf, anyOf: anyOf, inputs: []any{"a", "b", "c", 1}},
		{name: "allOf and oneOf", allOf: allOf, oneOf: oneOf, inputs: []any{"a", "b", "c", 1}},
		{name: "anyOf and oneOf", anyOf: anyOf, oneOf: oneOf, inputs: []any{"a", "b", "c", 1}},
		{name: "all three", allOf: allOf, anyOf: anyOf, oneOf: oneOf, inputs: []any{"a", "b", "c", 1}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			schema := &lib.Schema{
				Type:  []string{"string"},
				AllOf: test.allOf,
				AnyOf: test.anyOf,
				OneOf: test.oneOf,
			}
			imported, err := FromJSONSchema(schema)
			require.NoError(t, err)

			for _, input := range test.inputs {
				dependencyValid := schema.Validate(input).IsValid()
				_, parseErr := imported.ParseAny(input)
				assert.Equal(t, dependencyValid, parseErr == nil, "input %#v", input)
			}
		})
	}
}

func TestFromJSONSchema_ConjoinsObjectPropertiesAndAllOf(t *testing.T) {
	schema := &lib.Schema{
		Type: []string{"object"},
		Properties: &lib.SchemaMap{
			"name": {Type: []string{"string"}},
		},
		Required: []string{"name"},
		AllOf: []*lib.Schema{{
			Type: []string{"object"},
			Properties: &lib.SchemaMap{
				"active": {Type: []string{"boolean"}},
			},
			Required: []string{"active"},
		}},
	}

	imported, err := FromJSONSchema(schema)
	require.NoError(t, err)
	for _, input := range []any{
		map[string]any{"name": "Ada", "active": true},
		map[string]any{"name": "Ada"},
		map[string]any{"active": true},
		map[string]any{"name": 1, "active": true},
	} {
		dependencyValid := schema.Validate(input).IsValid()
		_, parseErr := imported.ParseAny(input)
		assert.Equal(t, dependencyValid, parseErr == nil, "input %#v", input)
	}
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
	assert.Nil(t, emailSchema.Format)
	require.NotNil(t, emailSchema.Pattern)

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

		zodSchema, losses, err := FromJSONSchemaLossy(schema)
		require.NoError(t, err)
		require.NotNil(t, zodSchema)
		require.Len(t, losses, 1)
		assert.Equal(t, "if/then/else", losses[0].Keyword)
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
		"dependentRequired": func() *lib.Schema {
			return &lib.Schema{DependentRequired: map[string][]string{"card": {"billing"}}}
		},
		"contentEncoding": func() *lib.Schema {
			return &lib.Schema{Type: []string{"string"}, ContentEncoding: new("base64")}
		},
		"contentMediaType": func() *lib.Schema {
			return &lib.Schema{Type: []string{"string"}, ContentMediaType: new("image/png")}
		},
		"contentSchema": func() *lib.Schema {
			return &lib.Schema{Type: []string{"string"}, ContentSchema: &lib.Schema{Boolean: new(true)}}
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

			zodSchema, losses, err := FromJSONSchemaLossy(schema)
			require.NoError(t, err)
			require.NotNil(t, zodSchema)
			require.Len(t, losses, 1)
			assert.Equal(t, keyword.keyword, losses[0].Keyword)
			assert.ErrorIs(t, losses[0], keyword.err)
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

	zodSchema, losses, err := FromJSONSchemaLossy(schema)
	require.NoError(t, err)
	require.NotNil(t, zodSchema)
	assert.Equal(t, []string{"$dynamicRef", "if/then/else", "not", "uniqueItems"}, importLossKeywords(losses))
}

func TestFromJSONSchemaLossy_ReturnsTypedLoss(t *testing.T) {
	schema := &lib.Schema{
		Type: []string{"object"},
		Properties: &lib.SchemaMap{
			"value": {Not: &lib.Schema{Boolean: new(true)}},
		},
	}

	imported, losses, err := FromJSONSchemaLossy(schema)
	require.NoError(t, err)
	require.NotNil(t, imported)
	require.Len(t, losses, 1)
	assert.Equal(t, "not", losses[0].Keyword)
	assert.Equal(t, "/properties/value/not", losses[0].Pointer)
	assert.ErrorIs(t, losses[0], ErrUnsupportedJSONSchemaKeyword)
}

func TestFromJSONSchemaLossy_PreservesTheSameKeywordAtDifferentLocations(t *testing.T) {
	document := &lib.Schema{
		Type: []string{"object"},
		Properties: &lib.SchemaMap{
			"first":  {Not: &lib.Schema{Boolean: new(true)}},
			"second": {Not: &lib.Schema{Boolean: new(true)}},
		},
	}

	imported, losses, err := FromJSONSchemaLossy(document)
	require.NoError(t, err)
	require.NotNil(t, imported)
	require.Len(t, losses, 2)
	assert.Equal(t, []string{
		"/properties/first/not",
		"/properties/second/not",
	}, []string{losses[0].Pointer, losses[1].Pointer})
}

func TestFromJSONSchemaLossy_SortsLossesByLocation(t *testing.T) {
	schema := &lib.Schema{
		Not:       &lib.Schema{Boolean: new(true)},
		MinLength: new(float64(1)),
		Maximum:   lib.NewRat(1),
		Items:     &lib.Schema{Boolean: new(false)},
	}

	_, losses, err := FromJSONSchemaLossy(schema)
	require.NoError(t, err)
	require.Len(t, losses, 4)
	assert.Equal(t, []string{
		"/items",
		"/maximum",
		"/minLength",
		"/not",
	}, []string{losses[0].Pointer, losses[1].Pointer, losses[2].Pointer, losses[3].Pointer})
}

func TestFromJSONSchemaLossy_NoLossReturnsEmptySnapshot(t *testing.T) {
	document := &lib.Schema{Type: []string{"string"}}

	imported, losses, err := FromJSONSchemaLossy(document)
	require.NoError(t, err)
	require.NotNil(t, imported)
	assert.Empty(t, losses)
}

func TestFromJSONSchemaLossy_ReturnedSnapshotsAreIndependent(t *testing.T) {
	document := &lib.Schema{Not: &lib.Schema{Boolean: new(true)}}

	_, first, err := FromJSONSchemaLossy(document)
	require.NoError(t, err)
	_, second, err := FromJSONSchemaLossy(document)
	require.NoError(t, err)
	require.Len(t, first, 1)
	require.Len(t, second, 1)

	first[0].Keyword = "caller-mutated"
	assert.Equal(t, "not", second[0].Keyword)
}

func TestFromJSONSchemaLossy_ConcurrentCallsReturnIdenticalSnapshots(t *testing.T) {
	document := &lib.Schema{
		Not:       &lib.Schema{Boolean: new(true)},
		MinLength: new(float64(1)),
		Maximum:   lib.NewRat(1),
	}
	_, want, err := FromJSONSchemaLossy(document)
	require.NoError(t, err)

	type result struct {
		losses []ImportLossError
		err    error
	}
	const callers = 16
	results := make(chan result, callers)
	var wg sync.WaitGroup
	for range callers {
		wg.Go(func() {
			_, losses, err := FromJSONSchemaLossy(document)
			results <- result{losses: losses, err: err}
		})
	}
	wg.Wait()
	close(results)

	for got := range results {
		require.NoError(t, got.err)
		assert.Equal(t, want, got.losses)
	}
}

func TestFromJSONSchemaLossy_FatalErrorPreservesEarlierLosses(t *testing.T) {
	document := &lib.Schema{
		Type: []string{"object"},
		Not:  &lib.Schema{Boolean: new(true)},
		Properties: &lib.SchemaMap{
			"value": {
				Type:    []string{"string"},
				Pattern: new("["),
			},
		},
	}

	imported, losses, err := FromJSONSchemaLossy(document)
	require.Nil(t, imported)
	require.ErrorIs(t, err, ErrJSONSchemaPatternCompile)
	require.Len(t, losses, 1)
	assert.Equal(t, "not", losses[0].Keyword)
	assert.Equal(t, "/not", losses[0].Pointer)

	strict, err := FromJSONSchema(document)
	require.Nil(t, strict)
	assert.ErrorIs(t, err, ErrUnsupportedJSONSchemaKeyword)
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
		t.Cleanup(func() { core.GlobalRegistry.Remove(zodSchema) })

		meta := zodSchema.Internals().Metadata()
		assert.Equal(t, "User Name", meta.Title)
		assert.Equal(t, "The user's full name", meta.Description)
		assert.False(t, core.GlobalRegistry.Has(zodSchema))
	})

	t.Run("explicit registry isolates imported metadata", func(t *testing.T) {
		title := "User Name"
		desc := "The user's full name"
		schema := &lib.Schema{
			Title:       &title,
			Description: &desc,
		}
		schema.Type = []string{"string"}
		registry := core.NewRegistry[core.GlobalMeta]()

		zodSchema, err := FromJSONSchema(schema, FromJSONSchemaOptions{Metadata: registry})
		require.NoError(t, err)

		meta, ok := registry.Get(zodSchema)
		require.True(t, ok, "Expected metadata to be registered in caller registry")
		assert.Equal(t, "User Name", meta.Title)
		assert.Equal(t, "The user's full name", meta.Description)

		assert.Equal(t, core.GlobalMeta{}, zodSchema.Internals().Metadata())
		assert.False(t, core.GlobalRegistry.Has(zodSchema))
	})

	t.Run("extracts $id and examples", func(t *testing.T) {
		schema := &lib.Schema{
			ID:       "https://example.com/schemas/name",
			Examples: []any{"John", "Jane"},
		}
		schema.Type = []string{"string"}

		zodSchema, err := FromJSONSchema(schema)
		require.NoError(t, err)
		t.Cleanup(func() { core.GlobalRegistry.Remove(zodSchema) })

		meta := zodSchema.Internals().Metadata()
		assert.Equal(t, "https://example.com/schemas/name", meta.ID)
		assert.Equal(t, []any{"John", "Jane"}, meta.Examples)
		assert.False(t, core.GlobalRegistry.Has(zodSchema))
	})

	t.Run("no metadata when fields are empty", func(t *testing.T) {
		schema := &lib.Schema{}
		schema.Type = []string{"integer"}

		zodSchema, err := FromJSONSchema(schema)
		require.NoError(t, err)

		assert.Equal(t, core.GlobalMeta{}, zodSchema.Internals().Metadata())
		assert.False(t, core.GlobalRegistry.Has(zodSchema))
	})
}

func TestFromJSONSchema_ExplicitRegistrySnapshotsImportedExamples(t *testing.T) {
	nested := []any{"before"}
	example := map[string]any{"names": nested}
	document := &lib.Schema{
		Type:     []string{"string"},
		Examples: []any{example},
	}
	registry := core.NewRegistry[core.GlobalMeta]()

	imported, err := FromJSONSchema(document, FromJSONSchemaOptions{Metadata: registry})
	require.NoError(t, err)
	nested[0] = "after"
	document.Examples[0] = "replaced"

	meta, ok := registry.Get(imported)
	require.True(t, ok)
	require.Len(t, meta.Examples, 1)
	assert.Equal(t, []any{"before"}, meta.Examples[0].(map[string]any)["names"])
}

func TestFromJSONSchema_DefaultMetadataSnapshotsImportedExamples(t *testing.T) {
	nested := []any{"before"}
	document := &lib.Schema{
		Type:     []string{"string"},
		Examples: []any{map[string]any{"names": nested}},
	}

	imported, err := FromJSONSchema(document)
	require.NoError(t, err)
	nested[0] = "after"
	document.Examples[0] = "replaced"

	meta := imported.Internals().Metadata()
	require.Len(t, meta.Examples, 1)
	assert.Equal(t, []any{"before"}, meta.Examples[0].(map[string]any)["names"])
}

func TestFromJSONSchema_MetadataMutationDoesNotChangeSourceDocument(t *testing.T) {
	tests := []struct {
		name string
		load func(*lib.Schema) (core.ZodSchema, core.GlobalMeta, error)
	}{
		{
			name: "schema-owned metadata",
			load: func(document *lib.Schema) (core.ZodSchema, core.GlobalMeta, error) {
				imported, err := FromJSONSchema(document)
				if err != nil {
					return nil, core.GlobalMeta{}, err
				}
				return imported, imported.Internals().Metadata(), nil
			},
		},
		{
			name: "explicit registry",
			load: func(document *lib.Schema) (core.ZodSchema, core.GlobalMeta, error) {
				registry := core.NewRegistry[core.GlobalMeta]()
				imported, err := FromJSONSchema(document, FromJSONSchemaOptions{Metadata: registry})
				if err != nil {
					return nil, core.GlobalMeta{}, err
				}
				meta, ok := registry.Get(imported)
				if !ok {
					return nil, core.GlobalMeta{}, errors.New("imported metadata missing from explicit registry")
				}
				return imported, meta, nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sourceNested := []any{"source"}
			document := &lib.Schema{
				Type:     []string{"string"},
				Examples: []any{map[string]any{"names": sourceNested}},
			}

			_, meta, err := tt.load(document)
			require.NoError(t, err)
			meta.Examples[0].(map[string]any)["names"].([]any)[0] = "destination"

			assert.Equal(t, "source", sourceNested[0])
			assert.Equal(t, "source", document.Examples[0].(map[string]any)["names"].([]any)[0])
		})
	}
}

func TestFromJSONSchema_MetadataRoundTripsThroughOwningDestination(t *testing.T) {
	tests := []struct {
		name      string
		roundTrip func(*lib.Schema) (*lib.Schema, error)
	}{
		{
			name: "schema-owned metadata",
			roundTrip: func(document *lib.Schema) (*lib.Schema, error) {
				imported, err := FromJSONSchema(document)
				if err != nil {
					return nil, err
				}
				return ToJSONSchema(imported)
			},
		},
		{
			name: "explicit registry",
			roundTrip: func(document *lib.Schema) (*lib.Schema, error) {
				registry := core.NewRegistry[core.GlobalMeta]()
				imported, err := FromJSONSchema(document, FromJSONSchemaOptions{Metadata: registry})
				if err != nil {
					return nil, err
				}
				return ToJSONSchema(imported, Options{Metadata: registry})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			document := &lib.Schema{
				ID:          "https://example.com/schemas/name",
				Title:       new("Name"),
				Description: new("A display name"),
				Type:        []string{"string"},
				Examples:    []any{map[string]any{"value": "Ada"}},
			}

			exported, err := tt.roundTrip(document)
			require.NoError(t, err)
			assert.Equal(t, "#/$defs/"+document.ID, exported.Ref)
			definition, ok := exported.Defs[document.ID]
			require.True(t, ok)
			require.NotNil(t, definition.Title)
			assert.Equal(t, *document.Title, *definition.Title)
			require.NotNil(t, exported.Description)
			assert.Equal(t, *document.Description, *exported.Description)
			assert.Equal(t, document.Examples, definition.Examples)
		})
	}
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

func TestFromJSONSchema_UnsupportedKeywordDocsMatchCode(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile("../docs/json-schema.md")
	require.NoError(t, err)
	doc := string(content)

	for _, keyword := range unsupportedImportKeywords {
		t.Run(keyword.keyword, func(t *testing.T) {
			t.Parallel()

			for _, token := range strings.Split(keyword.keyword, "/") {
				token = strings.TrimSpace(token)
				require.NotEmpty(t, token)
				assert.Contains(t, doc, token)
			}
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

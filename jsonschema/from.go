// Package jsonschema provides JSON Schema conversion for GoZod schemas.
package jsonschema

import (
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"regexp"
	"slices"

	lib "github.com/kaptinlin/jsonschema"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/types"
)

// Conversion errors for FromJSONSchema operations.
var (
	ErrUnsupportedJSONSchemaType    = errors.New("unsupported JSON Schema type")
	ErrUnsupportedJSONSchemaKeyword = errors.New("unsupported JSON Schema keyword")
	ErrInvalidJSONSchema            = errors.New("invalid JSON Schema")
	ErrJSONSchemaCircularRef        = errors.New("circular reference detected in JSON Schema")
	ErrJSONSchemaPatternCompile     = errors.New("failed to compile JSON Schema pattern")
	ErrJSONSchemaIfThenElse         = errors.New("if/then/else is not supported")
	ErrJSONSchemaPatternProperties  = errors.New("patternProperties is not supported")
	ErrJSONSchemaDynamicRef         = errors.New("$dynamicRef is not supported")
	ErrJSONSchemaUnevaluatedProps   = errors.New("unevaluatedProperties is not supported")
	ErrJSONSchemaUnevaluatedItems   = errors.New("unevaluatedItems is not supported")
	ErrJSONSchemaDependentSchemas   = errors.New("dependentSchemas is not supported")
	ErrJSONSchemaPropertyNames      = errors.New("propertyNames is not supported")
	ErrJSONSchemaContains           = errors.New("contains/minContains/maxContains is not supported")
)

// FromJSONSchemaOptions configures the JSON Schema to GoZod conversion.
type FromJSONSchemaOptions struct {
	// AllowLossy permits conversion to ignore unsupported JSON Schema keywords.
	// The default is fail-closed: unsupported keywords return an error.
	AllowLossy bool
	// LossyKeywords receives unsupported keywords ignored when AllowLossy is true.
	LossyKeywords *[]string
	// Metadata receives imported JSON Schema metadata. Nil keeps the historical
	// behavior of writing metadata to core.GlobalRegistry.
	Metadata *core.Registry[core.GlobalMeta]
}

// FromJSONSchema converts a kaptinlin/jsonschema Schema to a GoZod schema.
// Returns core.ZodSchema for maximum flexibility.
func FromJSONSchema(schema *lib.Schema, opts ...FromJSONSchemaOptions) (core.ZodSchema, error) {
	var options FromJSONSchemaOptions
	if len(opts) > 0 {
		options = opts[0]
	}

	ctx := &fromJSONSchemaContext{
		seen:    make(map[*lib.Schema]core.ZodSchema),
		options: options,
	}

	return ctx.convert(schema)
}

// fromJSONSchemaContext holds conversion state.
type fromJSONSchemaContext struct {
	seen    map[*lib.Schema]core.ZodSchema
	options FromJSONSchemaOptions
}

// convert dispatches to the appropriate converter based on schema type.
func (ctx *fromJSONSchemaContext) convert(s *lib.Schema) (core.ZodSchema, error) {
	if s == nil {
		return types.Unknown(), nil
	}

	// Handle circular references
	if existing, ok := ctx.seen[s]; ok {
		return existing, nil
	}

	// Handle boolean schema
	if s.Boolean != nil {
		if *s.Boolean {
			return types.Unknown(), nil // true schema accepts anything
		}
		return types.Never(), nil // false schema rejects everything
	}

	// Handle $ref (already pre-resolved by kaptinlin/jsonschema).
	if s.Ref != "" {
		if s.ResolvedRef == nil {
			return nil, fmt.Errorf("%w: unresolved $ref %q", ErrInvalidJSONSchema, s.Ref)
		}
		return ctx.convert(s.ResolvedRef)
	}
	if s.ResolvedRef != nil {
		return ctx.convert(s.ResolvedRef)
	}

	unsupported := ctx.unsupportedFeatures(s)
	if len(unsupported) > 0 {
		if !ctx.options.AllowLossy {
			return nil, unsupported[0].err
		}
		ctx.recordLossyKeywords(unsupported)
	}

	// Handle composition keywords first
	var result core.ZodSchema
	var convErr error

	switch {
	case len(s.AllOf) > 0:
		result, convErr = ctx.convertAllOf(s)
	case len(s.AnyOf) > 0:
		result, convErr = ctx.convertAnyOf(s)
	case len(s.OneOf) > 0:
		result, convErr = ctx.convertOneOf(s)
	case s.Const != nil:
		result, convErr = ctx.convertConst(s)
	case len(s.Enum) > 0:
		result, convErr = ctx.convertEnum(s)
	default:
		result, convErr = ctx.convertByType(s)
	}

	if convErr != nil {
		return nil, convErr
	}

	// Extract metadata from JSON Schema and attach to the GoZod schema.
	ctx.attachMeta(s, result)
	return result, nil
}

// attachMeta extracts metadata fields from a JSON Schema node and attaches them
// to the GoZod schema via the selected registry (Zod v4: 456af1ea).
// Captures: $id, title, description, examples.
func (ctx *fromJSONSchemaContext) attachMeta(s *lib.Schema, schema core.ZodSchema) {
	if schema == nil {
		return
	}

	var meta core.GlobalMeta
	if s.ID != "" {
		meta.ID = s.ID
	}
	if s.Title != nil && *s.Title != "" {
		meta.Title = *s.Title
	}
	if s.Description != nil && *s.Description != "" {
		meta.Description = *s.Description
	}
	if len(s.Examples) > 0 {
		meta.Examples = s.Examples
	}

	if meta.ID != "" || meta.Title != "" || meta.Description != "" || len(meta.Examples) > 0 {
		registry := ctx.options.Metadata
		if registry == nil {
			registry = core.GlobalRegistry
		}
		registry.Add(schema, meta)
	}
}

type unsupportedFeature struct {
	keyword string
	err     error
}

type unsupportedImportKeyword struct {
	keyword string
	err     error
	present func(*lib.Schema) bool
}

var unsupportedImportKeywords = []unsupportedImportKeyword{
	{
		keyword: "if/then/else",
		err:     ErrJSONSchemaIfThenElse,
		present: func(s *lib.Schema) bool {
			return s.If != nil || s.Then != nil || s.Else != nil
		},
	},
	{
		keyword: "patternProperties",
		err:     ErrJSONSchemaPatternProperties,
		present: func(s *lib.Schema) bool {
			return s.PatternProperties != nil && len(*s.PatternProperties) > 0
		},
	},
	{
		keyword: "$dynamicRef",
		err:     ErrJSONSchemaDynamicRef,
		present: func(s *lib.Schema) bool {
			return s.DynamicRef != ""
		},
	},
	{
		keyword: "unevaluatedProperties",
		err:     ErrJSONSchemaUnevaluatedProps,
		present: func(s *lib.Schema) bool {
			return s.UnevaluatedProperties != nil
		},
	},
	{
		keyword: "unevaluatedItems",
		err:     ErrJSONSchemaUnevaluatedItems,
		present: func(s *lib.Schema) bool {
			return s.UnevaluatedItems != nil
		},
	},
	{
		keyword: "not",
		err:     ErrUnsupportedJSONSchemaKeyword,
		present: func(s *lib.Schema) bool {
			return s.Not != nil
		},
	},
	{
		keyword: "uniqueItems",
		err:     ErrUnsupportedJSONSchemaKeyword,
		present: func(s *lib.Schema) bool {
			return s.UniqueItems != nil && *s.UniqueItems
		},
	},
	{
		keyword: "dependentSchemas",
		err:     ErrJSONSchemaDependentSchemas,
		present: func(s *lib.Schema) bool {
			return len(s.DependentSchemas) > 0
		},
	},
	{
		keyword: "propertyNames",
		err:     ErrJSONSchemaPropertyNames,
		present: func(s *lib.Schema) bool {
			return s.PropertyNames != nil
		},
	},
	{
		keyword: "contains",
		err:     ErrJSONSchemaContains,
		present: func(s *lib.Schema) bool {
			return s.Contains != nil || s.MinContains != nil || s.MaxContains != nil
		},
	},
}

// unsupportedFeatures returns unsupported keywords present on the schema.
func (ctx *fromJSONSchemaContext) unsupportedFeatures(s *lib.Schema) []unsupportedFeature {
	var unsupported []unsupportedFeature
	for _, keyword := range unsupportedImportKeywords {
		if keyword.present(s) {
			unsupported = append(unsupported, unsupportedFeature{
				keyword: keyword.keyword,
				err:     keyword.err,
			})
		}
	}
	return unsupported
}

func (ctx *fromJSONSchemaContext) recordLossyKeywords(features []unsupportedFeature) {
	if ctx.options.LossyKeywords == nil {
		return
	}
	for _, feature := range features {
		*ctx.options.LossyKeywords = append(*ctx.options.LossyKeywords, feature.keyword)
	}
}

// convertByType converts based on the type keyword.
func (ctx *fromJSONSchemaContext) convertByType(s *lib.Schema) (core.ZodSchema, error) {
	// Handle multi-type (e.g., ["string", "null"])
	if len(s.Type) > 1 {
		return ctx.convertMultiType(s)
	}

	// Handle single type - direct comparison is more efficient than slices.Contains for single element
	if len(s.Type) == 1 {
		switch s.Type[0] {
		case "string":
			return ctx.convertString(s)
		case "number":
			return ctx.convertNumber(s)
		case "integer":
			return ctx.convertInteger(s)
		case "boolean":
			return types.Bool(), nil
		case "null":
			return types.Nil(), nil
		case "array":
			return ctx.convertArray(s)
		case "object":
			return ctx.convertObject(s)
		default:
			return nil, fmt.Errorf("%w: %s", ErrUnsupportedJSONSchemaType, s.Type[0])
		}
	}

	// No type specified - return Unknown (accepts anything)
	return types.Unknown(), nil
}

func isSupportedJSONSchemaType(typeName string) bool {
	switch typeName {
	case "string", "number", "integer", "boolean", "null", "array", "object":
		return true
	default:
		return false
	}
}

// convertMultiType handles schemas with multiple types like ["string", "null"].
func (ctx *fromJSONSchemaContext) convertMultiType(s *lib.Schema) (core.ZodSchema, error) {
	for _, typeName := range s.Type {
		if !isSupportedJSONSchemaType(typeName) {
			return nil, fmt.Errorf("%w: %s", ErrUnsupportedJSONSchemaType, typeName)
		}
	}

	schemas := make([]core.ZodSchema, 0, len(s.Type))

	typeChecks := []struct {
		typeName string
		convert  func(*lib.Schema) (core.ZodSchema, error)
	}{
		{"string", ctx.convertString},
		{"number", ctx.convertNumber},
		{"integer", ctx.convertInteger},
		{"boolean", func(_ *lib.Schema) (core.ZodSchema, error) { return types.Bool(), nil }},
		{"null", func(_ *lib.Schema) (core.ZodSchema, error) { return types.Nil(), nil }},
		{"array", ctx.convertArray},
		{"object", ctx.convertObject},
	}

	for _, tc := range typeChecks {
		if slices.Contains(s.Type, tc.typeName) {
			schema, err := tc.convert(s)
			if err != nil {
				return nil, err
			}
			schemas = append(schemas, schema)
		}
	}

	if len(schemas) == 0 {
		return types.Unknown(), nil
	}
	if len(schemas) == 1 {
		return schemas[0], nil
	}

	return types.Union(zodSchemasToAny(schemas)), nil
}

// convertString converts a string type schema.
func (ctx *fromJSONSchemaContext) convertString(s *lib.Schema) (core.ZodSchema, error) {
	schema := types.String()
	if s.Format != nil {
		if formatSchema := ctx.getFormatSchema(*s.Format); formatSchema != nil {
			schema = formatSchema
		}
	}

	// Apply constraints
	if s.MinLength != nil {
		schema = schema.Min(int(*s.MinLength))
	}
	if s.MaxLength != nil {
		schema = schema.Max(int(*s.MaxLength))
	}
	if s.Pattern != nil {
		re, err := regexp.Compile(*s.Pattern)
		if err != nil {
			return nil, fmt.Errorf("%w: %q: %w", ErrJSONSchemaPatternCompile, *s.Pattern, err)
		}
		schema = schema.Regex(re)
	}

	return schema, nil
}

// getFormatSchema returns a dedicated string schema for known formats, or nil for unknown formats.
func (ctx *fromJSONSchemaContext) getFormatSchema(format string) *types.ZodString[string] {
	switch format {
	case "email":
		return types.Email().ZodString
	case "uuid":
		return types.UUID().ZodString
	case "uri", "url":
		return types.URL().ZodString
	case "date-time":
		return types.IsoDateTime().ZodString
	case "date":
		return types.IsoDate().ZodString
	case "time":
		return types.IsoTime().ZodString
	case "ipv4":
		return types.IPv4().ZodString
	case "ipv6":
		return types.IPv6().ZodString
	default:
		return nil
	}
}

// convertNumber converts a number type schema.
func (ctx *fromJSONSchemaContext) convertNumber(s *lib.Schema) (core.ZodSchema, error) {
	schema := types.Number()

	// Apply constraints
	if s.Minimum != nil {
		val, _ := s.Minimum.Float64()
		schema = schema.Min(val)
	}
	if s.Maximum != nil {
		val, _ := s.Maximum.Float64()
		schema = schema.Max(val)
	}
	if s.ExclusiveMinimum != nil {
		val, _ := s.ExclusiveMinimum.Float64()
		schema = schema.Gt(val)
	}
	if s.ExclusiveMaximum != nil {
		val, _ := s.ExclusiveMaximum.Float64()
		schema = schema.Lt(val)
	}
	if s.MultipleOf != nil {
		val, _ := s.MultipleOf.Float64()
		schema = schema.MultipleOf(val)
	}

	return schema, nil
}

// convertInteger converts an integer type schema.
func (ctx *fromJSONSchemaContext) convertInteger(s *lib.Schema) (core.ZodSchema, error) {
	schema := types.Int()

	if s.Minimum != nil {
		schema = schema.Min(ratCeilInt64(s.Minimum))
	}
	if s.Maximum != nil {
		schema = schema.Max(ratFloorInt64(s.Maximum))
	}
	if s.ExclusiveMinimum != nil {
		schema = schema.Min(ratFloorInt64(s.ExclusiveMinimum) + 1)
	}
	if s.ExclusiveMaximum != nil {
		schema = schema.Max(ratCeilInt64(s.ExclusiveMaximum) - 1)
	}
	if s.MultipleOf != nil {
		if divisor := ratIntegerMultipleDivisor(s.MultipleOf); divisor > 1 {
			schema = schema.MultipleOf(divisor)
		}
	}

	return schema, nil
}

func ratFloorInt64(r *lib.Rat) int64 {
	var floor big.Int
	floor.Div(r.Num(), r.Denom())
	return floor.Int64()
}

func ratCeilInt64(r *lib.Rat) int64 {
	ceil := ratFloorInt64(r)
	if !r.IsInt() {
		ceil++
	}
	return ceil
}

func ratIntegerMultipleDivisor(r *lib.Rat) int64 {
	var numerator big.Int
	numerator.Abs(r.Num())
	return numerator.Int64()
}

// convertArray converts an array type schema.
func (ctx *fromJSONSchemaContext) convertArray(s *lib.Schema) (core.ZodSchema, error) {
	// Handle prefixItems (tuple-like arrays, JSON Schema Draft 2020-12)
	if len(s.PrefixItems) > 0 {
		return ctx.convertTuple(s)
	}

	var itemSchema core.ZodSchema = types.Unknown()

	// Handle items schema
	if s.Items != nil {
		var err error
		itemSchema, err = ctx.convert(s.Items)
		if err != nil {
			return nil, err
		}
	}

	schema := types.Slice[any](itemSchema)

	// Apply constraints
	if s.MinItems != nil {
		schema = schema.Min(int(*s.MinItems))
	}
	if s.MaxItems != nil {
		schema = schema.Max(int(*s.MaxItems))
	}
	// Note: uniqueItems is not directly supported in GoZod

	return schema, nil
}

// convertTuple converts prefixItems to a Tuple schema.
func (ctx *fromJSONSchemaContext) convertTuple(s *lib.Schema) (core.ZodSchema, error) {
	items := make([]core.ZodSchema, len(s.PrefixItems))
	for i, itemSchema := range s.PrefixItems {
		converted, err := ctx.convert(itemSchema)
		if err != nil {
			return nil, err
		}
		items[i] = converted
	}

	// Handle rest element from s.Items (elements beyond prefixItems)
	var rest core.ZodSchema
	if s.Items != nil {
		var err error
		rest, err = ctx.convert(s.Items)
		if err != nil {
			return nil, err
		}
	}

	if rest != nil {
		return types.TupleWithRest(items, rest), nil
	}
	return types.Tuple(items...), nil
}

// convertObject converts an object type schema.
func (ctx *fromJSONSchemaContext) convertObject(s *lib.Schema) (core.ZodSchema, error) {
	// Handle record-like objects (additionalProperties without properties)
	if s.Properties == nil || len(*s.Properties) == 0 {
		if s.AdditionalProperties != nil {
			valueSchema, err := ctx.convert(s.AdditionalProperties)
			if err != nil {
				return nil, err
			}
			return types.Record(types.String(), valueSchema), nil
		}
		// Empty object with no constraints
		return types.Object(core.ObjectSchema{}), nil
	}

	// Build object shape
	shape := make(core.ObjectSchema)
	requiredSet := make(map[string]bool)
	for _, req := range s.Required {
		requiredSet[req] = true
	}

	// Mark schema for circular reference detection
	placeholder := types.Object(core.ObjectSchema{})
	ctx.seen[s] = placeholder

	// Convert each property
	for key, propSchema := range *s.Properties {
		propZodSchema, err := ctx.convert(propSchema)
		if err != nil {
			return nil, err
		}

		// Make optional if not in required list
		if !requiredSet[key] {
			propZodSchema = makeOptional(propZodSchema)
		}

		shape[key] = propZodSchema
	}

	result := types.Object(shape)

	// Handle additionalProperties
	if s.AdditionalProperties != nil {
		if s.AdditionalProperties.Boolean != nil {
			if *s.AdditionalProperties.Boolean {
				result = result.Passthrough()
			} else {
				result = result.Strict()
			}
		} else {
			// It's a schema - use catchall
			catchallSchema, err := ctx.convert(s.AdditionalProperties)
			if err != nil {
				return nil, err
			}
			result = result.Passthrough().WithCatchall(catchallSchema)
		}
	}

	// Update the placeholder reference
	ctx.seen[s] = result

	return result, nil
}

// makeOptional wraps a schema in Optional if it supports it.
func makeOptional(schema core.ZodSchema) core.ZodSchema {
	if schema == nil {
		return nil
	}

	method := reflect.ValueOf(schema).MethodByName("Optional")
	if !method.IsValid() || method.Type().NumIn() != 0 || method.Type().NumOut() != 1 {
		return schema
	}

	optional, ok := method.Call(nil)[0].Interface().(core.ZodSchema)
	if !ok {
		return schema
	}
	return optional
}

// convertConst converts a const value.
func (ctx *fromJSONSchemaContext) convertConst(s *lib.Schema) (core.ZodSchema, error) {
	if s.Const == nil {
		return types.Unknown(), nil
	}
	return types.Literal(s.Const.Value), nil
}

// convertEnum converts an enum schema.
func (ctx *fromJSONSchemaContext) convertEnum(s *lib.Schema) (core.ZodSchema, error) {
	if len(s.Enum) == 0 {
		return types.Unknown(), nil
	}

	// Check if all values are strings
	allStrings := true
	stringVals := make([]string, 0, len(s.Enum))
	for _, v := range s.Enum {
		if str, ok := v.(string); ok {
			stringVals = append(stringVals, str)
		} else {
			allStrings = false
			break
		}
	}

	if allStrings {
		return types.Enum(stringVals...), nil
	}

	// Mixed types - use union of literals
	literals := make([]any, len(s.Enum))
	for i, v := range s.Enum {
		literals[i] = types.Literal(v)
	}
	return types.Union(literals), nil
}

// convertAllOf converts allOf (intersection).
func (ctx *fromJSONSchemaContext) convertAllOf(s *lib.Schema) (core.ZodSchema, error) {
	schemas, err := ctx.convertSchemaList(s.AllOf)
	if err != nil {
		return nil, err
	}

	if len(schemas) == 0 {
		return types.Unknown(), nil
	}
	if len(schemas) == 1 {
		return schemas[0], nil
	}

	// Chain intersections: A & B & C => Intersection(Intersection(A, B), C)
	result := types.Intersection(schemas[0], schemas[1])
	for i := 2; i < len(schemas); i++ {
		result = types.Intersection(result, schemas[i])
	}
	return result, nil
}

// convertAnyOf converts anyOf (union).
func (ctx *fromJSONSchemaContext) convertAnyOf(s *lib.Schema) (core.ZodSchema, error) {
	schemas, err := ctx.convertSchemaList(s.AnyOf)
	if err != nil {
		return nil, err
	}

	if len(schemas) == 0 {
		return types.Unknown(), nil
	}
	if len(schemas) == 1 {
		return schemas[0], nil
	}

	return types.Union(zodSchemasToAny(schemas)), nil
}

// convertOneOf converts oneOf (exclusive union).
// Uses Xor for proper exclusive union semantics (exactly one must match).
func (ctx *fromJSONSchemaContext) convertOneOf(s *lib.Schema) (core.ZodSchema, error) {
	schemas, err := ctx.convertSchemaList(s.OneOf)
	if err != nil {
		return nil, err
	}

	if len(schemas) == 0 {
		return types.Unknown(), nil
	}
	if len(schemas) == 1 {
		return schemas[0], nil
	}

	return types.Xor(zodSchemasToAny(schemas)), nil
}

// convertSchemaList converts a slice of JSON Schemas to GoZod schemas.
func (ctx *fromJSONSchemaContext) convertSchemaList(schemas []*lib.Schema) ([]core.ZodSchema, error) {
	result := make([]core.ZodSchema, 0, len(schemas))
	for _, subSchema := range schemas {
		converted, err := ctx.convert(subSchema)
		if err != nil {
			return nil, err
		}
		result = append(result, converted)
	}
	return result, nil
}

// zodSchemasToAny converts a slice of ZodSchema to []any for Union/Xor constructors.
func zodSchemasToAny(schemas []core.ZodSchema) []any {
	options := make([]any, len(schemas))
	for i, s := range schemas {
		options[i] = s
	}
	return options
}

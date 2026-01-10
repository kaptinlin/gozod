package jsonschema

import (
	"errors"
	"regexp"
	"slices"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/types"
	lib "github.com/kaptinlin/jsonschema"
)

// FromJSONSchema conversion errors.
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
	// StrictMode causes conversion to fail on unsupported features.
	// If false, unsupported features are silently ignored.
	StrictMode bool
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

	// Handle $ref (already pre-resolved by kaptinlin/jsonschema)
	if s.ResolvedRef != nil {
		return ctx.convert(s.ResolvedRef)
	}

	// Check for unsupported features in strict mode
	if ctx.options.StrictMode {
		if err := ctx.checkUnsupportedFeatures(s); err != nil {
			return nil, err
		}
	}

	// Handle composition keywords first
	if len(s.AllOf) > 0 {
		return ctx.convertAllOf(s)
	}
	if len(s.AnyOf) > 0 {
		return ctx.convertAnyOf(s)
	}
	if len(s.OneOf) > 0 {
		return ctx.convertOneOf(s)
	}

	// Handle const
	if s.Const != nil {
		return ctx.convertConst(s)
	}

	// Handle enum
	if len(s.Enum) > 0 {
		return ctx.convertEnum(s)
	}

	// Handle type-based conversion
	return ctx.convertByType(s)
}

// checkUnsupportedFeatures returns an error if unsupported keywords are present.
func (ctx *fromJSONSchemaContext) checkUnsupportedFeatures(s *lib.Schema) error {
	if s.If != nil || s.Then != nil || s.Else != nil {
		return ErrJSONSchemaIfThenElse
	}
	if s.PatternProperties != nil && len(*s.PatternProperties) > 0 {
		return ErrJSONSchemaPatternProperties
	}
	if s.DynamicRef != "" {
		return ErrJSONSchemaDynamicRef
	}
	if s.UnevaluatedProperties != nil {
		return ErrJSONSchemaUnevaluatedProps
	}
	if s.UnevaluatedItems != nil {
		return ErrJSONSchemaUnevaluatedItems
	}
	if len(s.DependentSchemas) > 0 {
		return ErrJSONSchemaDependentSchemas
	}
	if s.PropertyNames != nil {
		return ErrJSONSchemaPropertyNames
	}
	if s.Contains != nil || s.MinContains != nil || s.MaxContains != nil {
		return ErrJSONSchemaContains
	}
	return nil
}

// convertByType converts based on the type keyword.
func (ctx *fromJSONSchemaContext) convertByType(s *lib.Schema) (core.ZodSchema, error) {
	// Handle multi-type (e.g., ["string", "null"])
	if len(s.Type) > 1 {
		return ctx.convertMultiType(s)
	}

	// Handle single type
	if len(s.Type) == 1 {
		switch {
		case slices.Contains(s.Type, "string"):
			return ctx.convertString(s)
		case slices.Contains(s.Type, "number"):
			return ctx.convertNumber(s)
		case slices.Contains(s.Type, "integer"):
			return ctx.convertInteger(s)
		case slices.Contains(s.Type, "boolean"):
			return types.Bool(), nil
		case slices.Contains(s.Type, "null"):
			return types.Nil(), nil
		case slices.Contains(s.Type, "array"):
			return ctx.convertArray(s)
		case slices.Contains(s.Type, "object"):
			return ctx.convertObject(s)
		}
	}

	// No type specified - return Unknown (accepts anything)
	return types.Unknown(), nil
}

// convertMultiType handles schemas with multiple types like ["string", "null"].
func (ctx *fromJSONSchemaContext) convertMultiType(s *lib.Schema) (core.ZodSchema, error) {
	schemas := make([]core.ZodSchema, 0)

	// Check each possible type
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

	// Convert to []any for Union
	options := make([]any, len(schemas))
	for i, s := range schemas {
		options[i] = s
	}
	return types.Union(options), nil
}

// convertString converts a string type schema.
func (ctx *fromJSONSchemaContext) convertString(s *lib.Schema) (core.ZodSchema, error) {
	// Check for format first - some formats use dedicated types
	if s.Format != nil {
		formatSchema := ctx.getFormatSchema(*s.Format)
		if formatSchema != nil {
			// Note: Format schemas don't support minLength/maxLength/pattern
			// In strict mode, we could error here if those are present
			return formatSchema, nil
		}
	}

	// Standard string schema
	schema := types.String()

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
			return nil, ErrJSONSchemaPatternCompile
		}
		schema = schema.Regex(re)
	}

	return schema, nil
}

// getFormatSchema returns a dedicated schema for known formats, or nil for unknown formats.
func (ctx *fromJSONSchemaContext) getFormatSchema(format string) core.ZodSchema {
	switch format {
	case "email":
		return types.Email()
	case "uuid":
		return types.Uuid()
	case "uri", "url":
		return types.URL()
	case "date-time":
		return types.IsoDateTime()
	case "date":
		return types.IsoDate()
	case "time":
		return types.IsoTime()
	case "ipv4":
		return types.IPv4()
	case "ipv6":
		return types.IPv6()
	default:
		// Unknown format - return nil to fall back to basic string
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

	// Apply constraints
	if s.Minimum != nil {
		val, _ := s.Minimum.Float64()
		schema = schema.Min(int64(val))
	}
	if s.Maximum != nil {
		val, _ := s.Maximum.Float64()
		schema = schema.Max(int64(val))
	}
	if s.ExclusiveMinimum != nil {
		val, _ := s.ExclusiveMinimum.Float64()
		schema = schema.Gt(int64(val))
	}
	if s.ExclusiveMaximum != nil {
		val, _ := s.ExclusiveMaximum.Float64()
		schema = schema.Lt(int64(val))
	}
	if s.MultipleOf != nil {
		val, _ := s.MultipleOf.Float64()
		schema = schema.MultipleOf(int64(val))
	}

	return schema, nil
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
			continue // Skip on error in this context
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
		// Check if it's a boolean false (strict mode)
		if s.AdditionalProperties.Boolean != nil && !*s.AdditionalProperties.Boolean {
			result = result.Strict()
		} else if s.AdditionalProperties.Boolean == nil {
			// It's a schema - use catchall
			catchallSchema, err := ctx.convert(s.AdditionalProperties)
			if err == nil {
				result = result.Catchall(catchallSchema)
			}
		}
		// If true, default passthrough behavior
	}

	// Update the placeholder reference
	ctx.seen[s] = result

	return result, nil
}

// makeOptional wraps a schema in Optional if it supports it.
func makeOptional(schema core.ZodSchema) core.ZodSchema {
	// Try different schema types
	switch s := schema.(type) {
	case *types.ZodString[string]:
		return s.Optional()
	case *types.ZodIntegerTyped[int, int]:
		return s.Optional()
	case *types.ZodFloatTyped[float64, float64]:
		return s.Optional()
	case *types.ZodBool[bool]:
		return s.Optional()
	case *types.ZodSlice[any, []any]:
		return s.Optional()
	case *types.ZodObject[map[string]any, map[string]any]:
		return s.Optional()
	default:
		// For unknown types, return as-is
		return schema
	}
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
	schemas := make([]core.ZodSchema, 0, len(s.AllOf))
	for _, subSchema := range s.AllOf {
		converted, err := ctx.convert(subSchema)
		if err != nil {
			return nil, err
		}
		schemas = append(schemas, converted)
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
	schemas := make([]core.ZodSchema, 0, len(s.AnyOf))
	for _, subSchema := range s.AnyOf {
		converted, err := ctx.convert(subSchema)
		if err != nil {
			return nil, err
		}
		schemas = append(schemas, converted)
	}

	if len(schemas) == 0 {
		return types.Unknown(), nil
	}
	if len(schemas) == 1 {
		return schemas[0], nil
	}

	// Convert to []any for Union
	options := make([]any, len(schemas))
	for i, s := range schemas {
		options[i] = s
	}
	return types.Union(options), nil
}

// convertOneOf converts oneOf (exclusive union).
// Uses Xor for proper exclusive union semantics (exactly one must match).
func (ctx *fromJSONSchemaContext) convertOneOf(s *lib.Schema) (core.ZodSchema, error) {
	schemas := make([]core.ZodSchema, 0, len(s.OneOf))
	for _, subSchema := range s.OneOf {
		converted, err := ctx.convert(subSchema)
		if err != nil {
			return nil, err
		}
		schemas = append(schemas, converted)
	}

	if len(schemas) == 0 {
		return types.Unknown(), nil
	}
	if len(schemas) == 1 {
		return schemas[0], nil
	}

	// Use Xor for exclusive union (oneOf semantics)
	options := make([]any, len(schemas))
	for i, s := range schemas {
		options[i] = s
	}
	return types.Xor(options), nil
}

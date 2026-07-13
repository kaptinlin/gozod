// Package jsonschema provides JSON Schema conversion for GoZod schemas.
package jsonschema

import (
	"cmp"
	"errors"
	"fmt"
	"maps"
	"math"
	"math/big"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"

	lib "github.com/kaptinlin/jsonschema"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/pkg/cloneutil"
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

// ImportError identifies the JSON Schema keyword and RFC 6901 location that failed to import.
type ImportError struct {
	Keyword string
	Pointer string
	Err     error
}

func (e *ImportError) Error() string {
	pointer := e.Pointer
	if pointer == "" {
		pointer = "<root>"
	}
	return fmt.Sprintf("import JSON Schema keyword %q at %s: %v", e.Keyword, pointer, e.Err)
}

// Unwrap preserves sentinel and dependency error inspection.
func (e *ImportError) Unwrap() error { return e.Err }

// ImportLossError describes validation semantics intentionally omitted by a lossy import.
type ImportLossError struct {
	Keyword string
	Pointer string
	Err     error
}

func (l ImportLossError) Error() string {
	pointer := l.Pointer
	if pointer == "" {
		pointer = "<root>"
	}
	return fmt.Sprintf("lossy JSON Schema import omitted keyword %q at %s: %v", l.Keyword, pointer, l.Err)
}

// Unwrap preserves sentinel and typed cause inspection.
func (l ImportLossError) Unwrap() error { return l.Err }

// FromJSONSchemaOptions configures the JSON Schema to GoZod conversion.
type FromJSONSchemaOptions struct {
	// Metadata receives imported JSON Schema metadata. Nil stores metadata on
	// the returned schema.
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
		seen:    make(map[*lib.Schema]*fromSchemaCell),
		options: options,
	}

	return ctx.convert(schema)
}

// FromJSONSchemaLossy converts a JSON Schema while reporting omitted validation semantics.
func FromJSONSchemaLossy(
	schema *lib.Schema,
	opts ...FromJSONSchemaOptions,
) (core.ZodSchema, []ImportLossError, error) {
	var options FromJSONSchemaOptions
	if len(opts) > 0 {
		options = opts[0]
	}
	losses := make([]ImportLossError, 0)
	ctx := &fromJSONSchemaContext{
		seen:    make(map[*lib.Schema]*fromSchemaCell),
		options: options,
		lossy:   true,
		losses:  &losses,
	}
	imported, err := ctx.convert(schema)
	return imported, normalizeImportLosses(losses), err
}

func normalizeImportLosses(losses []ImportLossError) []ImportLossError {
	slices.SortStableFunc(losses, func(a, b ImportLossError) int {
		return cmp.Or(cmp.Compare(a.Pointer, b.Pointer), cmp.Compare(a.Keyword, b.Keyword))
	})
	unique := losses[:0]
	for _, loss := range losses {
		if len(unique) > 0 {
			previous := unique[len(unique)-1]
			if previous.Pointer == loss.Pointer && previous.Keyword == loss.Keyword {
				continue
			}
		}
		unique = append(unique, loss)
	}
	return unique
}

// fromJSONSchemaContext holds conversion state.
type fromJSONSchemaContext struct {
	seen    map[*lib.Schema]*fromSchemaCell
	options FromJSONSchemaOptions
	path    []string
	lossy   bool
	losses  *[]ImportLossError
}

type fromSchemaCell struct {
	schema core.ZodSchema
	lazy   core.ZodSchema
}

func (c *fromSchemaCell) reference() core.ZodSchema {
	if c.schema != nil {
		return c.schema
	}
	if c.lazy == nil {
		c.lazy = types.LazyAny(func() any { return c.schema })
	}
	return c.lazy
}

func (ctx *fromJSONSchemaContext) at(tokens ...string) *fromJSONSchemaContext {
	derived := *ctx
	derived.path = append(slices.Clone(ctx.path), tokens...)
	return &derived
}

func (ctx *fromJSONSchemaContext) importError(keyword string, err error) error {
	return ctx.importErrorAt(keyword, err, keyword)
}

func (ctx *fromJSONSchemaContext) importErrorAt(keyword string, err error, path ...string) error {
	tokens := append(slices.Clone(ctx.path), path...)
	for i := range tokens {
		tokens[i] = strings.ReplaceAll(strings.ReplaceAll(tokens[i], "~", "~0"), "/", "~1")
	}
	pointer := ""
	if len(tokens) > 0 {
		pointer = "/" + strings.Join(tokens, "/")
	}
	return &ImportError{Keyword: keyword, Pointer: pointer, Err: err}
}

// convert dispatches to the appropriate converter based on schema type.
func (ctx *fromJSONSchemaContext) convert(s *lib.Schema) (core.ZodSchema, error) {
	if s == nil {
		return types.Unknown(), nil
	}

	if cell, ok := ctx.seen[s]; ok {
		return cell.reference(), nil
	}
	cell := &fromSchemaCell{}
	ctx.seen[s] = cell

	result, err := ctx.convertUnseen(s)
	if err != nil {
		delete(ctx.seen, s)
		return nil, err
	}
	cell.schema = result
	return result, nil
}

func (ctx *fromJSONSchemaContext) convertUnseen(s *lib.Schema) (core.ZodSchema, error) {
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
			return nil, ctx.importError("$ref", fmt.Errorf("%w: unresolved $ref %q", ErrInvalidJSONSchema, s.Ref))
		}
		return ctx.convertResolvedRef(s)
	}
	if s.ResolvedRef != nil {
		return ctx.convertResolvedRef(s)
	}

	unsupported := ctx.unsupportedFeatures(s)
	unsupported = append(unsupported, unanchoredValidationFeatures(s)...)
	if len(unsupported) > 0 {
		if !ctx.lossy {
			return nil, ctx.importError(unsupported[0].keyword, unsupported[0].err)
		}
		ctx.recordLosses(unsupported)
	}

	result, convErr := ctx.convertAssertions(s)

	if convErr != nil {
		return nil, convErr
	}

	// Extract metadata from JSON Schema and attach to the GoZod schema.
	ctx.attachMeta(s, result)
	return result, nil
}

func (ctx *fromJSONSchemaContext) convertResolvedRef(s *lib.Schema) (core.ZodSchema, error) {
	if s.ResolvedRef == s {
		return nil, ctx.importError("$ref", fmt.Errorf("%w: %s", ErrJSONSchemaCircularRef, s.Ref))
	}
	target, err := ctx.at("$ref").convert(s.ResolvedRef)
	if err != nil {
		return nil, err
	}

	siblings := *s
	siblings.Ref = ""
	siblings.ResolvedRef = nil
	if len(siblings.Type) == 0 && len(s.ResolvedRef.Type) > 0 {
		siblings.Type = slices.Clone(s.ResolvedRef.Type)
	}
	local, err := ctx.convert(&siblings)
	if err != nil {
		return nil, err
	}
	return types.Intersection(target, local), nil
}

func (ctx *fromJSONSchemaContext) convertAssertions(s *lib.Schema) (core.ZodSchema, error) {
	assertions := make([]core.ZodSchema, 0, 4)
	if len(s.Type) > 0 || s.Const != nil || len(s.Enum) > 0 {
		basic, err := ctx.convertBasicAssertions(s)
		if err != nil {
			return nil, err
		}
		assertions = append(assertions, basic)
	}
	if len(s.AllOf) > 0 {
		allOf, err := ctx.convertAllOf(s)
		if err != nil {
			return nil, err
		}
		assertions = append(assertions, allOf)
	}
	if len(s.AnyOf) > 0 {
		anyOf, err := ctx.convertAnyOf(s)
		if err != nil {
			return nil, err
		}
		assertions = append(assertions, anyOf)
	}
	if len(s.OneOf) > 0 {
		oneOf, err := ctx.convertOneOf(s)
		if err != nil {
			return nil, err
		}
		assertions = append(assertions, oneOf)
	}

	if len(assertions) == 0 {
		return ctx.convertBasicAssertions(s)
	}
	return conjoinSchemas(assertions), nil
}

func conjoinSchemas(schemas []core.ZodSchema) core.ZodSchema {
	result := schemas[0]
	for _, schema := range schemas[1:] {
		result = types.Intersection(result, schema)
	}
	return result
}

func (ctx *fromJSONSchemaContext) convertBasicAssertions(s *lib.Schema) (core.ZodSchema, error) {
	assertions := make([]core.ZodSchema, 0, 3)
	if len(s.Type) > 0 {
		base, err := ctx.convertByType(s)
		if err != nil {
			return nil, err
		}
		assertions = append(assertions, base)
	}
	if s.Const != nil {
		constant, err := ctx.convertConst(s)
		if err != nil {
			return nil, err
		}
		assertions = append(assertions, constant)
	}
	if len(s.Enum) > 0 {
		enum, err := ctx.convertEnum(s)
		if err != nil {
			return nil, err
		}
		assertions = append(assertions, enum)
	}

	if len(assertions) == 0 {
		return ctx.convertByType(s)
	}
	return conjoinSchemas(assertions), nil
}

// attachMeta extracts metadata fields from a JSON Schema node and attaches them
// to the returned schema or the caller-selected registry.
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
		meta = cloneutil.Clone(meta).(core.GlobalMeta)
		if ctx.options.Metadata != nil {
			ctx.options.Metadata.Add(schema, meta)
			return
		}
		schema.Internals().SetMetadata(meta)
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
			return s.PatternProperties != nil && len(*s.PatternProperties) > 0 &&
				!isImportablePatternRecord(s)
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
		keyword: "dependentRequired",
		err:     ErrUnsupportedJSONSchemaKeyword,
		present: func(s *lib.Schema) bool {
			return len(s.DependentRequired) > 0
		},
	},
	{
		keyword: "contentEncoding",
		err:     ErrUnsupportedJSONSchemaKeyword,
		present: func(s *lib.Schema) bool {
			return s.ContentEncoding != nil && contentMayApply(s)
		},
	},
	{
		keyword: "contentMediaType",
		err:     ErrUnsupportedJSONSchemaKeyword,
		present: func(s *lib.Schema) bool {
			if s.ContentMediaType == nil || !contentMayApply(s) {
				return false
			}
			return *s.ContentMediaType != "application/json" || len(s.Type) == 0
		},
	},
	{
		keyword: "contentSchema",
		err:     ErrUnsupportedJSONSchemaKeyword,
		present: func(s *lib.Schema) bool {
			return s.ContentSchema != nil && contentMayApply(s)
		},
	},
	{
		keyword: "propertyNames",
		err:     ErrJSONSchemaPropertyNames,
		present: func(s *lib.Schema) bool {
			return s.PropertyNames != nil && !isImportablePropertyNamesRecord(s)
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

func contentMayApply(s *lib.Schema) bool {
	return len(s.Type) == 0 || slices.Contains(s.Type, "string")
}

func isImportablePropertyNamesRecord(s *lib.Schema) bool {
	return len(s.Type) == 1 && s.Type[0] == "object" &&
		(s.Properties == nil || len(*s.Properties) == 0) &&
		(s.PatternProperties == nil || len(*s.PatternProperties) == 0) &&
		isImportableRecordKeySchema(s.PropertyNames) &&
		s.AdditionalProperties != nil && len(s.Required) == 0
}

func isImportableRecordKeySchema(s *lib.Schema) bool {
	return s != nil && s.Boolean == nil &&
		len(s.Type) == 1 && s.Type[0] == "string" &&
		s.Const == nil && len(s.Enum) == 0 &&
		len(s.AllOf) == 0 && len(s.AnyOf) == 0 && len(s.OneOf) == 0 &&
		s.Ref == "" && s.ResolvedRef == nil
}

func isImportablePatternRecord(s *lib.Schema) bool {
	return len(s.Type) == 1 && s.Type[0] == "object" &&
		(s.Properties == nil || len(*s.Properties) == 0) &&
		s.PropertyNames == nil &&
		s.PatternProperties != nil && len(*s.PatternProperties) == 1 &&
		s.AdditionalProperties == nil && len(s.Required) == 0
}

func unanchoredValidationFeatures(s *lib.Schema) []unsupportedFeature {
	if len(s.Type) > 0 {
		return nil
	}
	var features []unsupportedFeature
	if s.MinLength != nil {
		features = append(features, unsupportedFeature{keyword: "minLength", err: ErrUnsupportedJSONSchemaKeyword})
	}
	if s.MaxLength != nil {
		features = append(features, unsupportedFeature{keyword: "maxLength", err: ErrUnsupportedJSONSchemaKeyword})
	}
	if s.Pattern != nil {
		features = append(features, unsupportedFeature{keyword: "pattern", err: ErrUnsupportedJSONSchemaKeyword})
	}
	if s.Minimum != nil {
		features = append(features, unsupportedFeature{keyword: "minimum", err: ErrUnsupportedJSONSchemaKeyword})
	}
	if s.Maximum != nil {
		features = append(features, unsupportedFeature{keyword: "maximum", err: ErrUnsupportedJSONSchemaKeyword})
	}
	if s.ExclusiveMinimum != nil {
		features = append(features, unsupportedFeature{keyword: "exclusiveMinimum", err: ErrUnsupportedJSONSchemaKeyword})
	}
	if s.ExclusiveMaximum != nil {
		features = append(features, unsupportedFeature{keyword: "exclusiveMaximum", err: ErrUnsupportedJSONSchemaKeyword})
	}
	if s.MultipleOf != nil {
		features = append(features, unsupportedFeature{keyword: "multipleOf", err: ErrUnsupportedJSONSchemaKeyword})
	}
	if s.MinItems != nil {
		features = append(features, unsupportedFeature{keyword: "minItems", err: ErrUnsupportedJSONSchemaKeyword})
	}
	if s.MaxItems != nil {
		features = append(features, unsupportedFeature{keyword: "maxItems", err: ErrUnsupportedJSONSchemaKeyword})
	}
	if len(s.PrefixItems) > 0 {
		features = append(features, unsupportedFeature{keyword: "prefixItems", err: ErrUnsupportedJSONSchemaKeyword})
	}
	if s.Items != nil {
		features = append(features, unsupportedFeature{keyword: "items", err: ErrUnsupportedJSONSchemaKeyword})
	}
	if s.Properties != nil && len(*s.Properties) > 0 {
		features = append(features, unsupportedFeature{keyword: "properties", err: ErrUnsupportedJSONSchemaKeyword})
	}
	if s.AdditionalProperties != nil {
		features = append(features, unsupportedFeature{keyword: "additionalProperties", err: ErrUnsupportedJSONSchemaKeyword})
	}
	if len(s.Required) > 0 {
		features = append(features, unsupportedFeature{keyword: "required", err: ErrUnsupportedJSONSchemaKeyword})
	}
	if s.MinProperties != nil {
		features = append(features, unsupportedFeature{keyword: "minProperties", err: ErrUnsupportedJSONSchemaKeyword})
	}
	if s.MaxProperties != nil {
		features = append(features, unsupportedFeature{keyword: "maxProperties", err: ErrUnsupportedJSONSchemaKeyword})
	}
	return features
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

func (ctx *fromJSONSchemaContext) recordLosses(features []unsupportedFeature) {
	for _, feature := range features {
		if ctx.losses != nil {
			var importErr *ImportError
			if !errors.As(ctx.importError(feature.keyword, feature.err), &importErr) {
				continue
			}
			*ctx.losses = append(*ctx.losses, ImportLossError{
				Keyword: importErr.Keyword,
				Pointer: importErr.Pointer,
				Err:     importErr.Err,
			})
		}
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
			return nil, ctx.importError(
				"type",
				fmt.Errorf("%w: %s", ErrUnsupportedJSONSchemaType, s.Type[0]),
			)
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
			return nil, ctx.importError(
				"type",
				fmt.Errorf("%w: %s", ErrUnsupportedJSONSchemaType, typeName),
			)
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
	if s.ContentMediaType != nil && *s.ContentMediaType == "application/json" {
		schema = schema.JSON()
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
			return nil, ctx.importError(
				"pattern",
				fmt.Errorf("%w: %q: %w", ErrJSONSchemaPatternCompile, *s.Pattern, err),
			)
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
		itemSchema, err = ctx.at("items").convert(s.Items)
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
		converted, err := ctx.at("prefixItems", strconv.Itoa(i)).convert(itemSchema)
		if err != nil {
			return nil, err
		}
		items[i] = converted
	}

	// Handle rest element from s.Items (elements beyond prefixItems)
	var rest core.ZodSchema
	if s.Items != nil {
		var err error
		rest, err = ctx.at("items").convert(s.Items)
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
	if isImportablePatternRecord(s) {
		pattern := slices.Sorted(maps.Keys(*s.PatternProperties))[0]
		compiled, err := regexp.Compile(pattern)
		if err != nil {
			return nil, ctx.importErrorAt(
				"patternProperties",
				fmt.Errorf("%w: %q: %w", ErrJSONSchemaPatternCompile, pattern, err),
				"patternProperties",
				pattern,
			)
		}
		valueSchema, err := ctx.at("patternProperties", pattern).convert((*s.PatternProperties)[pattern])
		if err != nil {
			return nil, err
		}
		keySchema := types.String().Regex(compiled)
		return ctx.applyRecordPropertyBounds(s, types.LooseRecord(keySchema, valueSchema))
	}

	if isImportablePropertyNamesRecord(s) {
		keySchema, err := ctx.at("propertyNames").convert(s.PropertyNames)
		if err != nil {
			return nil, err
		}
		valueSchema, err := ctx.at("additionalProperties").convert(s.AdditionalProperties)
		if err != nil {
			return nil, err
		}
		return ctx.applyRecordPropertyBounds(s, types.Record(keySchema, valueSchema))
	}

	// Handle record-like objects (additionalProperties without properties)
	if s.Properties == nil || len(*s.Properties) == 0 {
		if s.AdditionalProperties != nil {
			valueSchema, err := ctx.at("additionalProperties").convert(s.AdditionalProperties)
			if err != nil {
				return nil, err
			}
			return ctx.applyRecordPropertyBounds(s, types.Record(types.String(), valueSchema))
		}
		// Empty object with no constraints
		return ctx.applyObjectPropertyBounds(s, types.Object(core.ObjectSchema{}))
	}

	// Build object shape
	shape := make(core.ObjectSchema)
	requiredSet := make(map[string]bool)
	for _, req := range s.Required {
		requiredSet[req] = true
	}

	// Convert each property
	for _, key := range slices.Sorted(maps.Keys(*s.Properties)) {
		propSchema := (*s.Properties)[key]
		propZodSchema, err := ctx.at("properties", key).convert(propSchema)
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
			catchallSchema, err := ctx.at("additionalProperties").convert(s.AdditionalProperties)
			if err != nil {
				return nil, err
			}
			result = result.Passthrough().WithCatchall(catchallSchema)
		}
	}

	return ctx.applyObjectPropertyBounds(s, result)
}

func (ctx *fromJSONSchemaContext) applyObjectPropertyBounds(
	s *lib.Schema,
	schema *types.ZodObject[map[string]any, map[string]any],
) (*types.ZodObject[map[string]any, map[string]any], error) {
	if s.MinProperties != nil {
		minimum, err := ctx.propertyBound("minProperties", *s.MinProperties)
		if err != nil {
			return nil, err
		}
		schema = schema.Min(minimum)
	}
	if s.MaxProperties != nil {
		maximum, err := ctx.propertyBound("maxProperties", *s.MaxProperties)
		if err != nil {
			return nil, err
		}
		schema = schema.Max(maximum)
	}
	return schema, nil
}

func (ctx *fromJSONSchemaContext) applyRecordPropertyBounds(
	s *lib.Schema,
	schema *types.ZodRecord[map[string]any, map[string]any],
) (*types.ZodRecord[map[string]any, map[string]any], error) {
	if s.MinProperties != nil {
		minimum, err := ctx.propertyBound("minProperties", *s.MinProperties)
		if err != nil {
			return nil, err
		}
		schema = schema.Min(minimum)
	}
	if s.MaxProperties != nil {
		maximum, err := ctx.propertyBound("maxProperties", *s.MaxProperties)
		if err != nil {
			return nil, err
		}
		schema = schema.Max(maximum)
	}
	return schema, nil
}

func (ctx *fromJSONSchemaContext) propertyBound(keyword string, value float64) (int, error) {
	if value < 0 || value != math.Trunc(value) || value > float64(math.MaxInt) {
		return 0, ctx.importError(
			keyword,
			fmt.Errorf("%w: %s must be a non-negative integer", ErrInvalidJSONSchema, keyword),
		)
	}
	return int(value), nil
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
	schemas, err := ctx.convertSchemaList("allOf", s.AllOf)
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
	schemas, err := ctx.convertSchemaList("anyOf", s.AnyOf)
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
	schemas, err := ctx.convertSchemaList("oneOf", s.OneOf)
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
func (ctx *fromJSONSchemaContext) convertSchemaList(keyword string, schemas []*lib.Schema) ([]core.ZodSchema, error) {
	result := make([]core.ZodSchema, 0, len(schemas))
	for i, subSchema := range schemas {
		converted, err := ctx.at(keyword, strconv.Itoa(i)).convert(subSchema)
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

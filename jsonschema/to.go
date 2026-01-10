package jsonschema

import (
	"cmp"
	"errors"
	"fmt"
	"math"
	"reflect"
	"slices"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/types"
	lib "github.com/kaptinlin/jsonschema"
)

// Sentinel errors reused across this package for consistent error handling.
var (
	ErrUnsupportedInputType          = errors.New("unsupported input type")
	ErrCircularReference             = errors.New("circular reference detected")
	ErrUnrepresentableType           = errors.New("unrepresentable type")
	ErrSchemaNotObjectOrStruct       = errors.New("schema is not a ZodObject or ZodStruct")
	ErrSliceElementNotSchema         = errors.New("slice element is not a ZodSchema")
	ErrArrayItemNotSchema            = errors.New("array item is not a ZodSchema")
	ErrUnhandledArrayLike            = errors.New("unhandled array-like type")
	ErrUnionInvalid                  = errors.New("schema is not a union type with Options method")
	ErrUnionNoMembers                = errors.New("union has no member schemas")
	ErrIntersectionInvalid           = errors.New("schema is not an intersection type")
	ErrInvalidEnumSchema             = errors.New("invalid enum schema")
	ErrEnumExtractValues             = errors.New("unable to extract enum values")
	ErrLiteralNoValuesMethod         = errors.New("schema does not have a Values method")
	ErrLiteralUnexpectedReturnValues = errors.New("unexpected number of return values from Values method")
	ErrExpectedDiscriminatedUnion    = errors.New("expected a discriminated union schema")
	ErrExpectedRecord                = errors.New("expected a record schema with ValueType()")
	ErrRecordValueNotSchema          = errors.New("record value type is not a valid schema")
	ErrMapNoMethods                  = errors.New("schema does not implement KeyType() and ValueType() methods for map conversion")
	ErrMapKeyNotSchema               = errors.New("map key type is not a valid schema")
	ErrMapValueNotSchema             = errors.New("map value type is not a valid schema")
)

// OverrideContext provides context for the Override function.
type OverrideContext struct {
	ZodSchema  core.ZodSchema
	JSONSchema *lib.Schema
}

// Options defines the configuration options for JSON schema conversion.
type Options struct {
	// A registry used to look up metadata for each schema.
	// Any schema with an ID property will be extracted as a $def.
	Metadata *core.Registry[core.GlobalMeta]

	// How to handle unrepresentable types:
	// "throw" (default) - Unrepresentable types throw an error.
	// "any" - Unrepresentable types become {}.
	Unrepresentable string

	// How to handle cycles:
	// "ref" (default) - Cycles will be broken using $defs.
	// "throw" - Cycles will throw an error if encountered.
	Cycles string

	// How to handle reused schemas:
	// "inline" (default) - Reused schemas will be inlined.
	// "ref" - Reused schemas will be extracted as $defs.
	Reused string

	// A function used to convert ID values to URIs for external $refs.
	URI func(id string) string

	// Target specifies the JSON Schema version.
	// "draft-2020-12" (default) or "draft-07".
	Target string

	// Override is a custom logic to modify the schema after generation.
	Override func(ctx OverrideContext)

	// IO specifies whether to convert the "input" or "output" schema.
	// "output" (default) or "input".
	IO string
}

// ToJSONSchema converts a GoZod schema or registry into a JSON Schema instance.
func ToJSONSchema(input any, opts ...Options) (*lib.Schema, error) {
	var options Options
	if len(opts) > 0 {
		options = opts[0]
	}

	switch v := input.(type) {
	case core.ZodSchema:
		return toJSONSchemaSingle(v, options)
	case *core.Registry[core.GlobalMeta]:
		return toJSONSchemaRegistry(v, options)
	default:
		return nil, fmt.Errorf("%w: %T", ErrUnsupportedInputType, v)
	}
}

// toJSONSchemaSingle handles the conversion of a single ZodSchema.
func toJSONSchemaSingle(schema core.ZodSchema, opts Options) (*lib.Schema, error) {
	c := newConverter(opts)
	s, err := c.convert(schema)
	if err != nil {
		return nil, err
	}

	if len(c.defs) > 0 {
		if s.Defs == nil {
			s.Defs = make(map[string]*lib.Schema)
		}
		// Sort keys to ensure deterministic output order
		defKeys := make([]string, 0, len(c.defs))
		for defKey := range c.defs {
			defKeys = append(defKeys, defKey)
		}
		slices.Sort(defKeys)
		for _, defKey := range defKeys {
			s.Defs[defKey] = c.defs[defKey]
		}
	}

	// NOTE: The `discriminator` property is no longer automatically added.
	// This functionality was removed to allow returning a direct *jsonschema.Schema instance.
	// The `oneOf` keyword will be used for discriminated unions, which is compliant
	// with JSON Schema standards.

	return s, nil
}

// toJSONSchemaRegistry handles the conversion of a schema Registry.
func toJSONSchemaRegistry(reg *core.Registry[core.GlobalMeta], opts Options) (*lib.Schema, error) {
	// Ensure converter has access to registry metadata for ID extraction.
	opts.Metadata = reg
	c := newConverter(opts)

	// First pass: process all schemas to populate seen map and defs.
	schemasInRegistry := make([]core.ZodSchema, 0)
	reg.Range(func(schema core.ZodSchema, meta core.GlobalMeta) bool {
		schemasInRegistry = append(schemasInRegistry, schema)
		return true
	})

	for _, s := range schemasInRegistry {
		if _, err := c.convert(s); err != nil {
			return nil, err // Early exit on conversion error.
		}
	}

	// Create a root schema to hold all definitions.
	rootSchema := &lib.Schema{}
	if len(c.defs) > 0 {
		rootSchema.Defs = make(map[string]*lib.Schema, len(c.defs))
		// Sort keys to ensure deterministic output order
		defKeys := make([]string, 0, len(c.defs))
		for key := range c.defs {
			defKeys = append(defKeys, key)
		}
		slices.Sort(defKeys)
		for _, key := range defKeys {
			rootSchema.Defs[key] = c.defs[key]
		}
	}

	return rootSchema, nil
}

// converter holds the state for a single conversion run.
type converter struct {
	opts        Options
	seen        map[core.ZodSchema]*lib.Schema
	counts      map[core.ZodSchema]int
	refs        map[core.ZodSchema]string
	auto        int
	path        []string
	defs        map[string]*lib.Schema
	depth       int
	idCache     map[core.ZodSchema]string         // cache for getID results
	unwrapCache map[core.ZodSchema]core.ZodSchema // cache for unwrapSchema results
}

func newConverter(opts Options) *converter {
	return &converter{
		opts:        opts,
		seen:        make(map[core.ZodSchema]*lib.Schema),
		counts:      make(map[core.ZodSchema]int),
		refs:        make(map[core.ZodSchema]string),
		defs:        make(map[string]*lib.Schema),
		idCache:     make(map[core.ZodSchema]string),
		unwrapCache: make(map[core.ZodSchema]core.ZodSchema),
	}
}

// unwrapSchema recursively unwraps well-known wrapper types (Optional, Nilable, etc.)
// by following a GetInner() method if implemented. This allows features like
// ID hoisting and reused-schema detection to operate on the underlying core
// schema rather than wrapper instances.
func (c *converter) unwrapSchema(s core.ZodSchema) core.ZodSchema {
	if cached, ok := c.unwrapCache[s]; ok {
		return cached
	}
	visited := map[core.ZodSchema]struct{}{}
	for {
		if s == nil {
			c.unwrapCache[s] = nil
			return nil
		}
		if _, ok := visited[s]; ok {
			// cycle guard
			c.unwrapCache[s] = s
			return s
		}
		visited[s] = struct{}{}
		if getter, ok := s.(interface{ GetInner() core.ZodSchema }); ok {
			inner := getter.GetInner()
			if inner != nil && inner != s {
				s = inner
				continue
			}
		}
		break
	}
	c.unwrapCache[s] = s
	return s
}

func (c *converter) convert(schema core.ZodSchema) (*lib.Schema, error) {
	// Track recursion depth
	c.depth++
	defer func() { c.depth-- }()

	if schema == nil {
		return nil, nil
	}

	// Unwrap schema for reuse detection purposes.
	baseKey := c.unwrapSchema(schema)
	// Track visit count on baseKey for reuse detection
	c.counts[baseKey]++

	// Cycle detection – return already converted schema or placeholder
	if s, ok := c.seen[schema]; ok {
		if c.opts.Cycles == "throw" {
			return nil, ErrCircularReference
		}
		// If we are in reuse-by-ref mode and have a registered definition, return a $ref instead
		if c.opts.Reused == "ref" {
			if name, ok := c.refs[baseKey]; ok {
				return &lib.Schema{Ref: "#/$defs/" + name}, nil
			}
		}
		return s, nil
	}

	// Insert placeholder to break potential cycles early
	placeholder := &lib.Schema{}
	c.seen[schema] = placeholder

	internals := schema.GetInternals()

	var finalSchema *lib.Schema
	var err error

	// Handle Optional / Nilable fields
	switch {
	case internals.IsOptional() && !internals.IsNilable():
		// Optional fields: render as the underlying schema (no null union).
		converted, errInner := c.doConvert(schema)
		if errInner != nil {
			return nil, errInner
		}
		finalSchema = converted
	case internals.IsNilable():
		// Nilable: underlying schema OR null
		baseSchema, derr := c.doConvert(schema)
		if derr != nil {
			return nil, derr
		}

		// Special case: if the base schema is already a pure null type,
		// don't wrap it in anyOf to avoid duplication
		if internals.Type == core.ZodTypeNil {
			finalSchema = baseSchema
		} else {
			finalSchema = &lib.Schema{
				AnyOf: []*lib.Schema{
					baseSchema,
					{Type: []string{"null"}},
				},
			}
		}
	default:
		finalSchema, err = c.doConvert(schema)
		if err != nil {
			return nil, err
		}
	}

	// Attach metadata (title, description, examples) if available
	c.applyMeta(schema, finalSchema)

	// Ensure a definition is registered **before** any placeholder replacement so that future
	// conversions (especially wrappers / lazy) can immediately resolve a $ref.
	if c.opts.Reused == "ref" {
		// Skip primitives – only composite schemas are worth extracting.
		composite := internals.Type == core.ZodTypeObject || internals.Type == core.ZodTypeStruct || internals.Type == core.ZodTypeSlice || internals.Type == core.ZodTypeArray || internals.Type == core.ZodTypeRecord || internals.Type == core.ZodTypeUnion || internals.Type == core.ZodTypeIntersection
		if composite && !internals.IsOptional() && !internals.IsNilable() {
			if _, exists := c.refs[baseKey]; !exists {
				c.auto++
				name := fmt.Sprintf("def%d", c.auto)
				c.refs[baseKey] = name
				c.defs[name] = finalSchema
			}
		}
	}

	// Replace placeholder with actual schema
	*placeholder = *finalSchema

	// ID-based hoisting to $defs
	if id := c.getID(schema); id != "" {
		// Ensure definition stored under $defs
		if _, exists := c.defs[id]; !exists {
			c.defs[id] = finalSchema
		}

		ref := "#/$defs/" + id
		if c.opts.URI != nil {
			ref = c.opts.URI(id)
		}

		// Special handling for Nilable wrapper – keep union structure but replace inner schema with $ref.
		if internals.IsNilable() {
			if len(placeholder.AnyOf) >= 1 {
				// Preserve description on property if present on definition
				underlying := placeholder.AnyOf[0]
				var desc *string
				if finalSchema.Description != nil && *finalSchema.Description != "" {
					desc = finalSchema.Description
					finalSchema.Description = nil
				}
				placeholder.AnyOf[0] = &lib.Schema{Ref: ref, Description: desc}
				// Update defs with the original underlying schema (not the $ref wrapper)
				c.defs[id] = underlying
			}
		} else {
			// Default behaviour: replace entire placeholder with a $ref schema.
			refSchema := &lib.Schema{Ref: ref}
			if finalSchema.Description != nil && *finalSchema.Description != "" {
				refSchema.Description = finalSchema.Description
				finalSchema.Description = nil
			}
			*placeholder = *refSchema
		}
	}

	// Automatic hoisting for reused schemas
	if c.counts[baseKey] > 1 && c.opts.Reused == "ref" {
		name, ok := c.refs[baseKey]
		if !ok {
			c.auto++
			name = fmt.Sprintf("def%d", c.auto)
			c.refs[baseKey] = name
			c.defs[name] = finalSchema
		}
		ref := "#/$defs/" + name
		placeholder.Ref = ref
		placeholder.Type = nil
		placeholder.OneOf = nil
		placeholder.Properties = nil
		placeholder.Items = nil
	}

	// Apply override at the very end.
	if c.opts.Override != nil {
		c.opts.Override(OverrideContext{
			ZodSchema:  schema,
			JSONSchema: placeholder,
		})
	}

	return placeholder, nil
}

func (c *converter) doConvert(schema core.ZodSchema) (*lib.Schema, error) {
	internals := schema.GetInternals()
	var jsonSchema *lib.Schema
	var err error

	// Execute OnAttach callbacks so that checks can annotate Bag/metadata
	for _, chk := range internals.Checks {
		if zc := chk.GetZod(); zc != nil {
			for _, fn := range zc.OnAttach {
				if fn != nil {
					fn(schema)
				}
			}
		}
	}

	switch internals.Type {
	case core.ZodTypeString:
		jsonSchema = &lib.Schema{Type: []string{"string"}}
		c.applyStringBag(jsonSchema, internals)
	case core.ZodTypeIPv4, core.ZodTypeIPv6, core.ZodTypeHostname, core.ZodTypeMAC, core.ZodTypeE164:
		jsonSchema = &lib.Schema{Type: []string{"string"}}
		c.applyStringBag(jsonSchema, internals)
	case core.ZodTypeCIDRv4, core.ZodTypeCIDRv6, core.ZodTypeURL:
		jsonSchema = &lib.Schema{Type: []string{"string"}}
		c.applyStringBag(jsonSchema, internals)
	case core.ZodTypeInt, core.ZodTypeInteger, core.ZodTypeInt8, core.ZodTypeInt16, core.ZodTypeInt32, core.ZodTypeInt64,
		core.ZodTypeUint, core.ZodTypeUint8, core.ZodTypeUint16, core.ZodTypeUint32, core.ZodTypeUint64, core.ZodTypeUintptr:
		jsonSchema = &lib.Schema{Type: []string{"integer"}}
		c.applyNumericRangeDefaults(internals.Type, jsonSchema, internals)
	case core.ZodTypeFloat:
		jsonSchema = &lib.Schema{Type: []string{"number"}}
	case core.ZodTypeFloat32, core.ZodTypeFloat64:
		jsonSchema = &lib.Schema{Type: []string{"number"}}
		c.applyNumericRangeDefaults(internals.Type, jsonSchema, internals)
	case core.ZodTypeBool:
		jsonSchema = &lib.Schema{Type: []string{"boolean"}}
	case core.ZodTypeNil:
		jsonSchema = &lib.Schema{Type: []string{"null"}}
	case core.ZodTypeAny, core.ZodTypeUnknown:
		jsonSchema = &lib.Schema{}
	case core.ZodTypeNever:
		// Never type should not match anything, which is represented as {"not": {}}
		// However, empty schemas are omitted due to omitempty tags in the jsonschema library
		// As a workaround, we create a special structure that will force {"not": {}} output
		// by using Boolean schema which serializes differently
		emptyNotSchema := &lib.Schema{}
		emptyNotSchema.Boolean = new(bool) // Setting boolean forces non-omitted serialization
		*emptyNotSchema.Boolean = true     // true means "match everything", so "not true" = "match nothing" (Never behavior)
		jsonSchema = &lib.Schema{Not: emptyNotSchema}
	case core.ZodTypeUnion:
		jsonSchema, err = c.convertUnion(schema)
	case core.ZodTypeXor:
		jsonSchema, err = c.convertXor(schema)
	case core.ZodTypePipe, core.ZodTypePipeline:
		// Handle pipeline schemas differently depending on IO mode.
		if c.opts.IO == "input" {
			if inp, ok := schema.(interface{ GetInner() core.ZodSchema }); ok {
				return c.convert(inp.GetInner())
			}
		} else { // output mode (default)
			if outpHolder, ok := schema.(interface{ GetOutput() core.ZodSchema }); ok {
				tgt := outpHolder.GetOutput()
				if tgt != nil {
					return c.convert(tgt)
				}
			}
		}
		// If target not available, fallback to unrepresentable handling.
		if c.opts.Unrepresentable == "any" {
			return &lib.Schema{}, nil
		}
		return nil, fmt.Errorf("%w: %s for '%s' IO", ErrUnrepresentableType, internals.Type, c.opts.IO)
	case core.ZodTypeTransform:
		// For transforms, expose input schema in IO:"input"; otherwise treat as unrepresentable.
		if c.opts.IO == "input" {
			if inp, ok := schema.(interface{ GetInner() core.ZodSchema }); ok {
				return c.convert(inp.GetInner())
			}
		}
		if c.opts.Unrepresentable == "any" {
			return &lib.Schema{}, nil
		}
		return nil, fmt.Errorf("%w: transform", ErrUnrepresentableType)
	case core.ZodTypeDiscriminated:
		jsonSchema, err = c.convertDiscriminatedUnion(schema)
	case core.ZodTypeIntersection:
		jsonSchema, err = c.convertIntersection(schema)
	case core.ZodTypeRecord:
		jsonSchema, err = c.convertRecord(schema)
	case core.ZodTypeObject, core.ZodTypeStruct:
		jsonSchema, err = c.convertObject(schema)
	case core.ZodTypeSlice, core.ZodTypeArray:
		jsonSchema, err = c.convertArray(schema)
	case core.ZodTypeTuple:
		jsonSchema, err = c.convertTuple(schema)
	case core.ZodTypeEnum:
		jsonSchema, err = c.convertEnum(schema)
	case core.ZodTypeLiteral:
		jsonSchema, err = c.convertLiteral(schema)
	case core.ZodTypeFile:
		jsonSchema, err = c.convertFile(schema)
	case core.ZodTypeLazy:
		jsonSchema, err = c.convertLazy(schema)
	case core.ZodTypeMap:
		jsonSchema, err = c.convertMap(schema)
	case core.ZodTypeSet:
		// Set cannot be represented in JSON Schema (similar to Map)
		return nil, fmt.Errorf("%w: %s", ErrUnrepresentableType, core.ZodTypeSet)
	case core.ZodTypeNumber:
		jsonSchema = &lib.Schema{Type: []string{"number"}}
	case core.ZodTypeBigInt:
		// Per user request, BigInt is not supported for JSON Schema generation.
		return nil, fmt.Errorf("%w: %s", ErrUnrepresentableType, core.ZodTypeBigInt)
	case core.ZodTypeDate:
		jsonSchema = &lib.Schema{Type: []string{"string"}, Format: ptrToString("date-time")}
	case core.ZodTypeEmail:
		jsonSchema = &lib.Schema{Type: []string{"string"}, Format: ptrToString("email")}
	case core.ZodTypeTime:
		jsonSchema = &lib.Schema{Type: []string{"string"}, Format: ptrToString("time")}
	case core.ZodTypeISODateTime, core.ZodTypeIso:
		jsonSchema = &lib.Schema{Type: []string{"string"}, Format: ptrToString("date-time")}
	case core.ZodTypeISODate:
		jsonSchema = &lib.Schema{Type: []string{"string"}, Format: ptrToString("date")}
	case core.ZodTypeISOTime:
		jsonSchema = &lib.Schema{Type: []string{"string"}, Format: ptrToString("time")}
	case core.ZodTypeISODuration:
		jsonSchema = &lib.Schema{Type: []string{"string"}, Format: ptrToString("duration")}
	case core.ZodTypeOptional, core.ZodTypeNilable, core.ZodTypeDefault, core.ZodTypePrefault, core.ZodTypeRefine, core.ZodTypeCheck:
		if s, ok := schema.(interface{ GetInner() core.ZodSchema }); ok {
			return c.convert(s.GetInner())
		}
		if c.opts.Unrepresentable == "any" {
			return &lib.Schema{}, nil
		}
		return nil, fmt.Errorf("%w: %s", ErrUnrepresentableType, internals.Type)
	case core.ZodTypeNaN, core.ZodTypeStringBool, core.ZodTypeFunction, core.ZodTypeCustom, core.ZodTypeComplex64, core.ZodTypeComplex128, core.ZodTypeNonOptional:
		// These types are unrepresentable in JSON Schema.
		// They will fall through to the default case.
		fallthrough
	default:
		if c.opts.Unrepresentable == "any" {
			jsonSchema = &lib.Schema{}
		} else {
			err = fmt.Errorf("%w: %s", ErrUnrepresentableType, internals.Type)
		}
	}

	if err != nil {
		return nil, err
	}

	// Map generic bag properties produced by checks → JSON Schema keywords
	if internals != nil && internals.Bag != nil {
		bag := internals.Bag
		c.applyBag(jsonSchema, bag)
	}

	return jsonSchema, nil
}

func (c *converter) applyStringBag(jsonSchema *lib.Schema, internals *core.ZodTypeInternals) {
	if internals.Bag == nil {
		return
	}

	// Process aggregated patterns from the bag
	if patternsRaw, ok := internals.Bag["patterns"]; ok {
		if patterns, ok := patternsRaw.([]string); ok && len(patterns) > 0 {
			// Deduplicate patterns to handle schemas that add the same check multiple ways
			uniquePatterns := make(map[string]struct{})
			var result []string
			for _, p := range patterns {
				if _, ok := uniquePatterns[p]; !ok {
					uniquePatterns[p] = struct{}{}
					result = append(result, p)
				}
			}
			patterns = result

			if len(patterns) == 1 {
				p := patterns[0]
				jsonSchema.Pattern = &p
			} else {
				jsonSchema.AllOf = make([]*lib.Schema, len(patterns))
				for i, p := range patterns {
					pattern := p
					jsonSchema.AllOf[i] = &lib.Schema{
						Pattern: &pattern,
					}
				}
			}

			// Remove patterns from bag so applyBag doesn't re-add them
			delete(internals.Bag, "patterns")
		}
	}

	// Apply other string-related properties from the bag
	if val, ok := internals.Bag["format"].(string); ok {
		jsonSchema.Format = &val
	}
	if v, ok := internals.Bag["minLength"]; ok {
		if n, ok := toFloat(v); ok {
			jsonSchema.MinLength = &n
		}
	}
	if v, ok := internals.Bag["maxLength"]; ok {
		if n, ok := toFloat(v); ok {
			jsonSchema.MaxLength = &n
		}
	}
	if v, ok := internals.Bag["contentEncoding"]; ok {
		if ce, ok := v.(string); ok {
			jsonSchema.ContentEncoding = &ce
		}
	}
	if v, ok := internals.Bag["contentMediaType"]; ok {
		if cmt, ok := v.(string); ok {
			jsonSchema.ContentMediaType = &cmt
		}
	}
}

// catchaller is an interface for schemas that have a catch-all schema.
type catchaller interface {
	GetCatchall() core.ZodSchema
}

// unknownKeysHandler is an interface for schemas that handle unknown keys.
// We use a generic method signature to avoid circular imports.
type unknownKeysHandler interface {
	GetUnknownKeys() any
}

// unwrapper is an interface for schemas that can unwrap to reveal inner schemas.
type unwrapper interface {
	Unwrap() core.ZodType[any]
}

func (c *converter) convertObject(schema core.ZodSchema) (*lib.Schema, error) {
	// Try ObjectSchema first (for ZodObject)
	if s, ok := schema.(interface{ Shape() core.ObjectSchema }); ok {
		shape := s.Shape()
		return c.convertObjectFromShape(schema, shape)
	}

	// Try StructSchema (for ZodStruct)
	if s, ok := schema.(interface{ Shape() core.StructSchema }); ok {
		shape := s.Shape()
		// StructSchema and ObjectSchema are the same type alias
		return c.convertObjectFromShape(schema, shape)
	}

	return nil, ErrSchemaNotObjectOrStruct
}

func (c *converter) convertObjectFromShape(schema core.ZodSchema, shape core.ObjectSchema) (*lib.Schema, error) {
	properties := make(map[string]*lib.Schema, len(shape))
	required := make([]string, 0)

	for key, propSchema := range shape {
		c.path = append(c.path, "properties", key)
		propJsonSchema, err := c.convert(propSchema)
		if err != nil {
			return nil, err
		}
		properties[key] = propJsonSchema
		c.path = c.path[:len(c.path)-2]

		isRequired := !propSchema.GetInternals().IsOptional()
		// In "input" mode, fields with defaults are not required.
		if c.opts.IO == "input" {
			pInternals := propSchema.GetInternals()
			if pInternals.DefaultValue != nil || pInternals.DefaultFunc != nil {
				isRequired = false
			}
		}
		if isRequired {
			required = append(required, key)
		}
	}

	jsonSchema := &lib.Schema{
		Type: []string{"object"},
	}

	// Only add properties if there are any
	if len(properties) > 0 {
		schemaMap := lib.SchemaMap(properties)
		jsonSchema.Properties = &schemaMap
	}
	if len(required) > 0 {
		slices.SortFunc(required, func(a, b string) int { return cmp.Compare(b, a) })
		jsonSchema.Required = required
	}

	// Handle additionalProperties based on catchall and unknown keys mode
	// Try to call GetUnknownKeys and GetCatchall methods
	if s, ok := schema.(catchaller); ok {
		if catchallSchema := s.GetCatchall(); catchallSchema != nil {
			// If there's a catchall schema, convert it to additionalProperties
			c.path = append(c.path, "additionalProperties")
			catchallJsonSchema, err := c.convert(catchallSchema)
			if err != nil {
				return nil, err
			}
			jsonSchema.AdditionalProperties = catchallJsonSchema
			c.path = c.path[:len(c.path)-1]
		}
	}

	if jsonSchema.AdditionalProperties == nil {
		// Prefer the fast interface assertion path.
		if uk, ok := schema.(unknownKeysHandler); ok {
			modeStr := fmt.Sprint(uk.GetUnknownKeys())
			if modeStr == string(types.ObjectModePassthrough) || modeStr == "passthrough" {
				trueValue := true
				jsonSchema.AdditionalProperties = &lib.Schema{Boolean: &trueValue}
			} else {
				// "strict" and "strip" (or unknown) => additionalProperties false
				falseValue := false
				jsonSchema.AdditionalProperties = &lib.Schema{Boolean: &falseValue}
			}
		} else if unknownKeysMethod := reflect.ValueOf(schema).MethodByName("GetUnknownKeys"); unknownKeysMethod.IsValid() {
			if results := unknownKeysMethod.Call(nil); len(results) == 1 {
				modeStr := fmt.Sprint(results[0].Interface())
				if modeStr == "passthrough" {
					trueValue := true
					jsonSchema.AdditionalProperties = &lib.Schema{Boolean: &trueValue}
				} else {
					falseValue := false
					jsonSchema.AdditionalProperties = &lib.Schema{Boolean: &falseValue}
				}
			}
		} else {
			// Default for objects without unknown keys info
			falseValue := false
			jsonSchema.AdditionalProperties = &lib.Schema{Boolean: &falseValue}
		}
	}

	return jsonSchema, nil
}

func (c *converter) convertArray(schema core.ZodSchema) (*lib.Schema, error) {
	jsonSchema := &lib.Schema{Type: []string{"array"}}

	// Handle ZodSlice (variable-length arrays)
	if s, ok := schema.(interface{ Element() any }); ok {
		if elemSchema, ok := s.Element().(core.ZodSchema); ok {
			items, err := c.convert(elemSchema)
			if err != nil {
				return nil, err
			}
			jsonSchema.Items = items
		} else {
			return nil, ErrSliceElementNotSchema
		}
	} else if s, ok := schema.(interface {
		Items() []any
		Rest() any
	}); ok { // Handle ZodArray (tuples)
		elements := s.Items()
		jsonSchema.PrefixItems = make([]*lib.Schema, len(elements))
		for i, el := range elements {
			if itemSchema, ok := el.(core.ZodSchema); ok {
				converted, err := c.convert(itemSchema)
				if err != nil {
					return nil, err
				}
				jsonSchema.PrefixItems[i] = converted
			} else {
				return nil, fmt.Errorf("%w: index %d", ErrArrayItemNotSchema, i)
			}
		}

		if rest := s.Rest(); rest != nil {
			if restSchema, ok := rest.(core.ZodSchema); ok {
				items, err := c.convert(restSchema)
				if err != nil {
					return nil, err
				}
				jsonSchema.Items = items
			}
		} else {
			// If only one tuple element and no rest, treat as standard variable-length array for compatibility
			if len(elements) == 1 {
				jsonSchema.Items = jsonSchema.PrefixItems[0]
				jsonSchema.PrefixItems = nil
			} else {
				// Fixed-length tuple
				itemCount := float64(len(elements))
				jsonSchema.MinItems = &itemCount
				jsonSchema.MaxItems = &itemCount
			}
		}
	} else {
		return nil, ErrUnhandledArrayLike
	}

	return jsonSchema, nil
}

// convertTuple handles ZodTuple -> JSON Schema array with prefixItems
func (c *converter) convertTuple(schema core.ZodSchema) (*lib.Schema, error) {
	tupleSchema, ok := schema.(interface {
		GetItems() []core.ZodSchema
		GetRest() core.ZodSchema
	})
	if !ok {
		return nil, fmt.Errorf("%w: expected tuple schema with GetItems/GetRest methods", ErrUnhandledArrayLike)
	}

	items := tupleSchema.GetItems()
	rest := tupleSchema.GetRest()

	jsonSchema := &lib.Schema{Type: []string{"array"}}

	// Convert each item schema to prefixItems
	if len(items) > 0 {
		jsonSchema.PrefixItems = make([]*lib.Schema, len(items))
		for i, itemSchema := range items {
			c.path = append(c.path, fmt.Sprintf("prefixItems[%d]", i))
			converted, err := c.convert(itemSchema)
			if err != nil {
				return nil, err
			}
			jsonSchema.PrefixItems[i] = converted
			c.path = c.path[:len(c.path)-1]
		}
	}

	// Handle rest element schema
	if rest != nil {
		c.path = append(c.path, "items")
		restConverted, err := c.convert(rest)
		if err != nil {
			return nil, err
		}
		jsonSchema.Items = restConverted
		c.path = c.path[:len(c.path)-1]
	} else {
		// No rest element - fixed length tuple
		// Set min/max items to enforce exact length
		itemCount := float64(len(items))
		// Calculate required count (items that are not optional)
		requiredCount := 0
		for _, item := range items {
			if !item.GetInternals().IsOptional() {
				requiredCount++
			} else {
				break // Stop at first optional item
			}
		}
		// Actually, we need to find the last required item
		lastRequired := -1
		for i := len(items) - 1; i >= 0; i-- {
			if !items[i].GetInternals().IsOptional() {
				lastRequired = i
				break
			}
		}
		minItems := float64(lastRequired + 1)
		jsonSchema.MinItems = &minItems
		jsonSchema.MaxItems = &itemCount
	}

	return jsonSchema, nil
}

// applyMeta copies GlobalMeta fields from registry onto the generated JSON Schema node.
func (c *converter) applyMeta(schema core.ZodSchema, jsonSchema *lib.Schema) {
	if jsonSchema == nil || schema == nil {
		return
	}

	// Determine which registry to use. Explicit opts takes precedence, otherwise fallback to global.
	reg := c.opts.Metadata
	if reg == nil {
		reg = core.GlobalRegistry
	}
	if reg == nil {
		return
	}

	meta, ok := reg.Get(schema)
	if !ok {
		// If no direct meta, check if the schema is a wrapper (e.g., Optional) and has meta on its inner type.
		if s, ok := schema.(interface{ GetInner() core.ZodSchema }); ok {
			meta, ok = reg.Get(s.GetInner())
			if !ok {
				return
			}
		} else {
			return
		}
	}

	if meta.Title != "" && jsonSchema.Title == nil {
		t := meta.Title
		jsonSchema.Title = &t
	}
	if meta.Description != "" && (jsonSchema.Description == nil || *jsonSchema.Description == "") {
		d := meta.Description
		jsonSchema.Description = &d
	}
	if len(meta.Examples) > 0 && len(jsonSchema.Examples) == 0 {
		jsonSchema.Examples = meta.Examples
	}
}

// getID retrieves meta.ID for schema via opts.Metadata or global registry.
func (c *converter) getID(schema core.ZodSchema) string {
	if id, ok := c.idCache[schema]; ok {
		return id
	}
	reg := c.opts.Metadata
	if reg == nil {
		reg = core.GlobalRegistry
	}
	if reg == nil {
		return ""
	}
	meta, ok := reg.Get(schema)
	if !ok {
		// If no direct meta, check if the schema is a wrapper (e.g., Optional) and has meta on its inner type.
		if s, ok := schema.(interface{ GetInner() core.ZodSchema }); ok {
			meta, ok = reg.Get(s.GetInner())
			if !ok {
				return ""
			}
		} else {
			return ""
		}
	}
	c.idCache[schema] = meta.ID
	return meta.ID
}

// applyBag copies well-known constraint keys from Zod internals.Bag to JSON Schema.
func (c *converter) applyBag(js *lib.Schema, bag map[string]any) {
	// First handle pattern/patterns specially as they may need to merge
	if v, ok := bag["pattern"]; ok {
		if p, ok := v.(string); ok {
			if js.Pattern == nil {
				js.Pattern = &p
			} else {
				if js.AllOf == nil {
					js.AllOf = []*lib.Schema{{Pattern: js.Pattern}}
				}
				js.AllOf = append(js.AllOf, &lib.Schema{Pattern: &p})
				js.Pattern = nil
			}
		}
	}

	if v, ok := bag["patterns"]; ok {
		if patterns, ok := v.([]string); ok && len(patterns) > 0 {
			// Move existing single pattern to allOf
			if js.Pattern != nil {
				if js.AllOf == nil {
					js.AllOf = []*lib.Schema{}
				}
				js.AllOf = append(js.AllOf, &lib.Schema{Pattern: js.Pattern})
				js.Pattern = nil
			}

			if len(patterns) == 1 && js.AllOf == nil {
				js.Pattern = &patterns[0]
			} else {
				if js.AllOf == nil {
					js.AllOf = []*lib.Schema{}
				}
				for _, p := range patterns {
					pCopy := p
					js.AllOf = append(js.AllOf, &lib.Schema{Pattern: &pCopy})
				}
			}
		}
	}

	// Table-driven simple setters to minimize reflection and branching.
	for k, v := range bag {
		switch k {
		case "minLength":
			if f, ok := toFloat(v); ok {
				js.MinLength = &f
			}
		case "maxLength":
			if f, ok := toFloat(v); ok {
				js.MaxLength = &f
			}
		case "format":
			if s, ok := v.(string); ok {
				js.Format = &s
			}
		case "contentEncoding":
			if s, ok := v.(string); ok {
				js.ContentEncoding = &s
			}
		case "contentMediaType":
			if s, ok := v.(string); ok {
				js.ContentMediaType = &s
			}
		case "minimum":
			if r, ok := toRat(v); ok {
				js.Minimum = &r
			}
		case "maximum":
			if r, ok := toRat(v); ok {
				js.Maximum = &r
			}
		case "multipleOf":
			if r, ok := toRat(v); ok {
				js.MultipleOf = &r
			}
		case "exclusiveMinimum":
			if r, ok := toRat(v); ok {
				js.ExclusiveMinimum = &r
			}
		case "exclusiveMaximum":
			if r, ok := toRat(v); ok {
				js.ExclusiveMaximum = &r
			}
		case "minItems":
			if f, ok := toFloat(v); ok {
				js.MinItems = &f
			}
		case "maxItems":
			if f, ok := toFloat(v); ok {
				js.MaxItems = &f
			}
		case "minSize":
			if f, ok := toFloat(v); ok {
				js.MinLength = &f
			}
		case "maxSize":
			if f, ok := toFloat(v); ok {
				js.MaxLength = &f
			}
		case "size":
			if f, ok := toFloat(v); ok {
				js.MinLength = &f
				js.MaxLength = &f
			}
		case "mime":
			if mimes, ok := v.([]string); ok && len(mimes) == 1 {
				cm := mimes[0]
				js.ContentMediaType = &cm
			}
		}
	}
}

// toFloat converts numeric types to float64.
func toFloat(v any) (float64, bool) {
	switch x := v.(type) {
	case int:
		return float64(x), true
	case int32:
		return float64(x), true
	case int64:
		return float64(x), true
	case uint:
		return float64(x), true
	case uint32:
		return float64(x), true
	case uint64:
		return float64(x), true
	case float32:
		return float64(x), true
	case float64:
		return x, true
	default:
		return 0, false
	}
}

// toRat converts numeric values to jsonschema.Rat (alias for big.Rat wrapper).
func toRat(v any) (lib.Rat, bool) {
	if f, ok := toFloat(v); ok {
		return *lib.NewRat(f), true
	}
	return lib.Rat{}, false
}

// convertUnion handles ZodUnion -> JSON Schema anyOf
func (c *converter) convertUnion(schema core.ZodSchema) (*lib.Schema, error) {
	u, ok := schema.(interface{ Options() []core.ZodSchema })
	if !ok {
		return nil, ErrUnionInvalid
	}
	opts := u.Options()
	if len(opts) == 0 {
		return nil, ErrUnionNoMembers
	}

	// Check if this is an "optional union" (type + null)
	if len(opts) == 2 {
		var nonNullSchema, nullSchema core.ZodSchema
		for _, opt := range opts {
			if isNullSchema(opt) {
				nullSchema = opt
			} else {
				nonNullSchema = opt
			}
		}

		// If we have exactly one non-null schema and one null schema,
		// this is an optional/nilable field
		if nonNullSchema != nil && nullSchema != nil {
			// Check if the original schema is Optional (not Nilable)
			if schema.GetInternals().IsOptional() && !schema.GetInternals().IsNilable() {
				// For Optional fields, just return the non-null schema
				// The optionality is handled at the object level (not in required array)
				return c.convert(nonNullSchema)
			}

			// For Nilable fields, use anyOf with null
			anyOf := make([]*lib.Schema, 0, 2)

			c.path = append(c.path, "anyOf[0]")
			nonNullJS, err := c.convert(nonNullSchema)
			if err != nil {
				return nil, err
			}
			anyOf = append(anyOf, nonNullJS)
			c.path = c.path[:len(c.path)-1]

			c.path = append(c.path, "anyOf[1]")
			nullJS, err := c.convert(nullSchema)
			if err != nil {
				return nil, err
			}
			anyOf = append(anyOf, nullJS)
			c.path = c.path[:len(c.path)-1]

			return &lib.Schema{AnyOf: anyOf}, nil
		}
	}

	// Regular union - use anyOf
	anyOf := make([]*lib.Schema, 0, len(opts))
	for i, mem := range opts {
		c.path = append(c.path, fmt.Sprintf("anyOf[%d]", i))
		s, err := c.convert(mem)
		if err != nil {
			return nil, err
		}
		anyOf = append(anyOf, s)
		c.path = c.path[:len(c.path)-1]
	}

	return &lib.Schema{AnyOf: anyOf}, nil
}

// convertXor handles ZodXor -> JSON Schema oneOf (exclusive union - exactly one must match)
func (c *converter) convertXor(schema core.ZodSchema) (*lib.Schema, error) {
	x, ok := schema.(interface{ Options() []core.ZodSchema })
	if !ok {
		return nil, ErrUnionInvalid
	}
	opts := x.Options()
	if len(opts) == 0 {
		return nil, ErrUnionNoMembers
	}

	// Xor uses oneOf (exactly one must match)
	oneOf := make([]*lib.Schema, 0, len(opts))
	for i, opt := range opts {
		c.path = append(c.path, fmt.Sprintf("oneOf[%d]", i))
		converted, err := c.convert(opt)
		if err != nil {
			return nil, err
		}
		oneOf = append(oneOf, converted)
		c.path = c.path[:len(c.path)-1]
	}

	return &lib.Schema{OneOf: oneOf}, nil
}

// isNullSchema checks if a schema represents null/nil type
func isNullSchema(schema core.ZodSchema) bool {
	// Check the schema type
	internals := schema.GetInternals()
	return internals.Type == core.ZodTypeNil
}

// convertIntersection handles ZodIntersection -> JSON Schema allOf (two schemas)
func (c *converter) convertIntersection(schema core.ZodSchema) (*lib.Schema, error) {
	inter, ok := schema.(interface {
		Left() core.ZodSchema
		Right() core.ZodSchema
	})
	if !ok {
		return nil, ErrIntersectionInvalid
	}
	leftSchema, err := c.convert(inter.Left())
	if err != nil {
		return nil, err
	}
	rightSchema, err := c.convert(inter.Right())
	if err != nil {
		return nil, err
	}

	return &lib.Schema{AllOf: []*lib.Schema{leftSchema, rightSchema}}, nil
}

// convertRecord handles ZodRecord -> JSON Schema object with additionalProperties
func (c *converter) convertRecord(schema core.ZodSchema) (*lib.Schema, error) {
	recordSchema, ok := schema.(interface {
		KeyType() any
		ValueType() any
	})
	if !ok {
		return nil, ErrExpectedRecord
	}

	// Convert key schema for propertyNames
	var propertyNames *lib.Schema
	if keyType := recordSchema.KeyType(); keyType != nil {
		if keySchema, ok := keyType.(core.ZodSchema); ok {
			var err error
			propertyNames, err = c.convert(keySchema)
			if err != nil {
				return nil, err
			}
		}
	}

	// Convert value schema for additionalProperties
	var additionalProperties *lib.Schema
	if valueType := recordSchema.ValueType(); valueType != nil {
		if valueSchema, ok := valueType.(core.ZodSchema); ok {
			var err error
			additionalProperties, err = c.convert(valueSchema)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, ErrRecordValueNotSchema
		}
	}

	return &lib.Schema{
		Type:                 []string{"object"},
		PropertyNames:        propertyNames,
		AdditionalProperties: additionalProperties,
	}, nil
}

// convertEnum handles ZodEnum -> JSON Schema enum
func (c *converter) convertEnum(schema core.ZodSchema) (*lib.Schema, error) {
	// Use reflection to call Options() method to get enum values
	var enumValues []any
	if optionsMethod := reflect.ValueOf(schema).MethodByName("Options"); optionsMethod.IsValid() {
		if results := optionsMethod.Call(nil); len(results) == 1 {
			slice := results[0]
			if slice.IsValid() && slice.Kind() == reflect.Slice {
				for i := 0; i < slice.Len(); i++ {
					enumValues = append(enumValues, slice.Index(i).Interface())
				}
			}
		}
	}
	if len(enumValues) == 0 {
		return nil, ErrEnumExtractValues
	}

	// Ensure deterministic order for enum values to avoid map iteration randomness
	if len(enumValues) > 0 {
		switch enumValues[0].(type) {
		case string:
			slices.SortStableFunc(enumValues, func(a, b interface{}) int {
				return cmp.Compare(a.(string), b.(string))
			})
		case int, int32, int64, uint, uint32, uint64, float64, float32:
			slices.SortStableFunc(enumValues, func(a, b interface{}) int {
				return cmp.Compare(fmt.Sprintf("%v", a), fmt.Sprintf("%v", b))
			})
		}
	}

	js := &lib.Schema{Enum: enumValues}
	// Determine type from first value
	switch enumValues[0].(type) {
	case string:
		js.Type = []string{"string"}
	case int, int32, int64, uint, uint32, uint64, float64, float32:
		js.Type = []string{"number"}
	}
	return js, nil
}

// convertLiteral handles ZodLiteral -> JSON Schema const/enum
func (c *converter) convertLiteral(schema core.ZodSchema) (*lib.Schema, error) {
	// Use reflection to call Values() method to get literal values
	var values []any
	if valuesMethod := reflect.ValueOf(schema).MethodByName("Values"); valuesMethod.IsValid() {
		if results := valuesMethod.Call(nil); len(results) == 1 {
			sliceValue := results[0]
			if sliceValue.IsValid() && sliceValue.Kind() == reflect.Slice {
				values = make([]any, sliceValue.Len())
				for i := 0; i < sliceValue.Len(); i++ {
					values[i] = sliceValue.Index(i).Interface()
				}
			}
		}
	}
	if len(values) == 0 {
		return nil, ErrLiteralNoValuesMethod
	}

	jsonSchema := &lib.Schema{}

	// Flatten if a single slice/array literal is provided (represents multiple literal values).
	if len(values) == 1 {
		rv := reflect.ValueOf(values[0])
		if rv.IsValid() && (rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array) {
			flat := make([]any, rv.Len())
			for i := 0; i < rv.Len(); i++ {
				flat[i] = rv.Index(i).Interface()
			}
			values = flat
		}
	}

	if len(values) > 0 {
		switch values[0].(type) {
		case string:
			jsonSchema.Type = []string{"string"}
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
			jsonSchema.Type = []string{"number"}
		case bool:
			jsonSchema.Type = []string{"boolean"}
		}
	}

	if len(values) == 1 {
		// Single literal value → const
		jsonSchema.Const = &lib.ConstValue{Value: values[0], IsSet: true}
		return jsonSchema, nil
	}

	// Multiple literal values → enum
	jsonSchema.Enum = values
	return jsonSchema, nil
}

// convertFile handles ZodFile -> JSON Schema file representation
func (c *converter) convertFile(schema core.ZodSchema) (*lib.Schema, error) {
	s := &lib.Schema{
		Type:            []string{"string"},
		Format:          ptrToString("binary"),
		ContentEncoding: ptrToString("binary"),
	}

	internals := schema.GetInternals()
	c.applyBag(s, internals.Bag) // Bag is applied here.

	// Handle multiple MIME types
	if mimes, ok := internals.Bag["mime"].([]string); ok && len(mimes) > 1 {
		anyOf := make([]*lib.Schema, len(mimes))
		for i, mime := range mimes {
			mimeCopy := mime
			itemSchema := &lib.Schema{
				Type:             []string{"string"},
				Format:           ptrToString("binary"),
				ContentEncoding:  ptrToString("binary"),
				ContentMediaType: &mimeCopy,
			}
			// Apply min/max/size constraints using helper conversion
			if v, ok := internals.Bag["size"]; ok {
				if f, ok := toFloat(v); ok {
					itemSchema.MinLength = &f
					itemSchema.MaxLength = &f
				}
			} else {
				if v, ok := internals.Bag["minSize"]; ok {
					if f, ok := toFloat(v); ok {
						itemSchema.MinLength = &f
					}
				}
				if v, ok := internals.Bag["maxSize"]; ok {
					if f, ok := toFloat(v); ok {
						itemSchema.MaxLength = &f
					}
				}
			}
			anyOf[i] = itemSchema
		}
		s.AnyOf = anyOf
		// Clear top-level properties that are now in anyOf
		s.Type = nil
		s.Format = nil
		s.ContentEncoding = nil
		s.ContentMediaType = nil
		s.MinLength = nil
		s.MaxLength = nil

		// Remove size-related bag entries so upstream code won't reapply them
		delete(internals.Bag, "minSize")
		delete(internals.Bag, "maxSize")
		delete(internals.Bag, "size")
	}

	return s, nil
}

func ptrToString(s string) *string { return &s }

// convertLazy resolves inner schema and delegates conversion.
func (c *converter) convertLazy(schema core.ZodSchema) (*lib.Schema, error) {
	// Use interface assertion to get inner schema
	if s, ok := schema.(unwrapper); ok {
		inner := s.Unwrap()

		// Try core.ZodSchema first
		if zodSchema, ok := inner.(core.ZodSchema); ok {
			// Check if this creates a cycle by looking for the inner schema in our path
			if _, found := c.seen[zodSchema]; found {
				// This is a cycle, return a reference to the root or $defs if available
				if id := c.getID(zodSchema); id != "" {
					return &lib.Schema{Ref: "#/$defs/" + id}, nil
				}
				if name, ok := c.refs[c.unwrapSchema(zodSchema)]; ok {
					return &lib.Schema{Ref: "#/$defs/" + name}, nil
				}
				return &lib.Schema{Ref: "#"}, nil
			}
			return c.convert(zodSchema)
		}

		// `inner` is already a core.ZodType[any]; attempt to call GetInner via reflection to unwrap further
		if method := reflect.ValueOf(inner).MethodByName("GetInner"); method.IsValid() {
			if results := method.Call(nil); len(results) == 1 {
				actualInner := results[0].Interface()
				if zodSchema, ok := actualInner.(core.ZodSchema); ok {
					// Detect potential cycles
					if _, found := c.seen[zodSchema]; found {
						if id := c.getID(zodSchema); id != "" {
							return &lib.Schema{Ref: "#/$defs/" + id}, nil
						}
						if name, ok := c.refs[c.unwrapSchema(zodSchema)]; ok {
							return &lib.Schema{Ref: "#/$defs/" + name}, nil
						}
						return &lib.Schema{Ref: "#"}, nil
					}
					return c.convert(zodSchema)
				}
			}
		}
	}
	// Fallback: treat as any
	return &lib.Schema{}, nil
}

// convertMap converts ZodMap -> JSON Schema object with additionalProperties
func (c *converter) convertMap(schema core.ZodSchema) (*lib.Schema, error) {
	mapSchema, ok := schema.(interface {
		KeyType() any
		ValueType() any
	})
	if !ok {
		return nil, ErrMapNoMethods
	}

	keySchema, ok := mapSchema.KeyType().(core.ZodSchema)
	if !ok {
		return nil, ErrMapKeyNotSchema
	}

	if keySchema.GetInternals().Type != core.ZodTypeString {
		return nil, fmt.Errorf("%w: map with non-string keys", ErrUnrepresentableType)
	}

	valueSchema, ok := mapSchema.ValueType().(core.ZodSchema)
	if !ok {
		return nil, ErrMapValueNotSchema
	}

	additionalProps, err := c.convert(valueSchema)
	if err != nil {
		return nil, err
	}

	return &lib.Schema{
		Type:                 []string{"object"},
		AdditionalProperties: additionalProps,
	}, nil
}

func (c *converter) convertDiscriminatedUnion(schema core.ZodSchema) (*lib.Schema, error) {
	du, ok := schema.(interface {
		Discriminator() string
		Options() []core.ZodSchema
	})
	if !ok {
		return nil, ErrExpectedDiscriminatedUnion
	}

	options := du.Options()
	oneOf := make([]*lib.Schema, len(options))

	for i, option := range options {
		convertedOption, err := c.convert(option)
		if err != nil {
			return nil, err
		}
		oneOf[i] = convertedOption
	}

	return &lib.Schema{
		OneOf: oneOf,
	}, nil
}

// numericRangeDefaults maps ZodTypeCode to its inclusive minimum and maximum.
var numericRangeDefaults = map[core.ZodTypeCode][2]float64{
	// Go platform dependent int range (assuming 64-bit build).
	core.ZodTypeInt:     {float64(math.MinInt), float64(math.MaxInt)},
	core.ZodTypeInteger: {float64(math.MinInt), float64(math.MaxInt)},

	// Signed integers
	core.ZodTypeInt8:  {-128, 127},
	core.ZodTypeInt16: {-32768, 32767},
	core.ZodTypeInt32: {-2147483648, 2147483647},
	// Using float64 to hold these large ints – precision loss is acceptable for JSON-Schema ranges
	core.ZodTypeInt64: {float64(math.MinInt64), float64(math.MaxInt64)}, // Int64 full range

	// Unsigned integers
	core.ZodTypeUint:   {0, float64(math.MaxUint)},
	core.ZodTypeUint8:  {0, 255},
	core.ZodTypeUint16: {0, 65535},
	core.ZodTypeUint32: {0, 4294967295},
	core.ZodTypeUint64: {0, 1.844674407371e19}, // approximate math.MaxUint64 as float64

	// Floats
	core.ZodTypeFloat32: {-math.MaxFloat32, math.MaxFloat32},
	core.ZodTypeFloat64: {-math.MaxFloat64, math.MaxFloat64},
}

// applyNumericRangeDefaults populates default numeric range constraints for a given schema based on its type.
func (c *converter) applyNumericRangeDefaults(zodType core.ZodTypeCode, js *lib.Schema, internals *core.ZodTypeInternals) {
	// Apply only at top-level (depth==1)
	if c.depth != 1 {
		return
	}

	// Do not override explicit constraints from bag.
	if internals != nil && internals.Bag != nil {
		bag := internals.Bag
		if _, hasMin := bag["minimum"]; hasMin {
			return
		}
		if _, hasMinEx := bag["exclusiveMinimum"]; hasMinEx {
			return
		}
		if _, hasMax := bag["maximum"]; hasMax {
			return
		}
		if _, hasMaxEx := bag["exclusiveMaximum"]; hasMaxEx {
			return
		}
	}

	rng, ok := numericRangeDefaults[zodType]
	if !ok {
		return
	}

	// Only set if not already set.
	if js.Minimum == nil {
		js.Minimum = lib.NewRat(rng[0])
	}
	if js.Maximum == nil {
		js.Maximum = lib.NewRat(rng[1])
	}
}

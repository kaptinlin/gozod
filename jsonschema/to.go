package jsonschema

import (
	"cmp"
	"errors"
	"fmt"
	"maps"
	"reflect"
	"slices"

	lib "github.com/kaptinlin/jsonschema"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/pkg/cloneutil"
	"github.com/kaptinlin/gozod/types"
)

// Conversion errors for ToJSONSchema operations.
var (
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
	ErrInvalidRegistrySchemaID       = errors.New("invalid registry schema ID")
	ErrMapNoMethods                  = errors.New("schema does not implement KeyType() and ValueType() methods for map conversion")
	ErrMapKeyNotSchema               = errors.New("map key type is not a valid schema")
	ErrMapValueNotSchema             = errors.New("map value type is not a valid schema")
	ErrInvalidJSONSchemaOption       = errors.New("invalid JSON Schema option")
)

// OverrideContext provides context for the Override function.
type OverrideContext struct {
	ZodSchema  core.ZodSchema
	JSONSchema *lib.Schema
}

// UnrepresentableMode controls how export handles schemas with no faithful JSON Schema representation.
type UnrepresentableMode string

// CyclesMode controls how export handles cyclic schema graphs.
type CyclesMode string

// ReusedMode controls how export handles reused schema instances.
type ReusedMode string

// IOMode selects whether export represents schema input or output shape.
type IOMode string

// Supported JSON Schema export option values.
const (
	UnrepresentableThrow UnrepresentableMode = "throw"
	UnrepresentableAny   UnrepresentableMode = "any"

	CyclesRef   CyclesMode = "ref"
	CyclesThrow CyclesMode = "throw"

	ReusedInline ReusedMode = "inline"
	ReusedRef    ReusedMode = "ref"

	IOOutput IOMode = "output"
	IOInput  IOMode = "input"
)

// Options defines the configuration options for JSON schema conversion.
type Options struct {
	// Metadata provides whole-record overrides for schemas present in the registry.
	// Schemas absent from the registry use their schema-owned metadata.
	Metadata *core.Registry[core.GlobalMeta]

	// How to handle unrepresentable types:
	// "throw" (default) - Unrepresentable types throw an error.
	// "any" - Unrepresentable types become {}.
	Unrepresentable UnrepresentableMode

	// How to handle cycles:
	// "ref" (default) - Cycles will be broken using $defs.
	// "throw" - Cycles will throw an error if encountered.
	Cycles CyclesMode

	// How to handle reused schemas:
	// "inline" (default) - Reused schemas will be inlined.
	// "ref" - Reused schemas will be extracted as $defs.
	Reused ReusedMode

	// A function used to convert ID values to URIs for external $refs.
	URI func(id string) string

	// Override is a custom logic to modify the schema after generation.
	Override func(ctx OverrideContext)

	// IO specifies whether to convert the "input" or "output" schema.
	// "output" (default) or "input".
	IO IOMode
}

// ToJSONSchema converts a GoZod schema into a JSON Schema instance.
func ToJSONSchema(schema core.ZodSchema, opts ...Options) (*lib.Schema, error) {
	options, err := optionsFrom(opts)
	if err != nil {
		return nil, err
	}
	return toJSONSchemaSingle(schema, options)
}

// ToJSONSchemaRegistry converts a schema registry into a JSON Schema document.
func ToJSONSchemaRegistry(
	registry *core.Registry[core.GlobalMeta],
	opts ...Options,
) (*lib.Schema, error) {
	if registry == nil {
		return nil, fmt.Errorf("registry is nil: %w", ErrInvalidRegistrySchemaID)
	}
	options, err := optionsFrom(opts)
	if err != nil {
		return nil, err
	}
	return toJSONSchemaRegistry(registry, options)
}

func optionsFrom(opts []Options) (Options, error) {
	var options Options
	if len(opts) > 0 {
		options = opts[0]
	}
	return normalizeOptions(options)
}

func normalizeOptions(opts Options) (Options, error) {
	if opts.Unrepresentable == "" {
		opts.Unrepresentable = UnrepresentableThrow
	}
	if opts.Unrepresentable != UnrepresentableThrow && opts.Unrepresentable != UnrepresentableAny {
		return opts, fmt.Errorf("%w: Unrepresentable=%q", ErrInvalidJSONSchemaOption, opts.Unrepresentable)
	}

	if opts.Cycles == "" {
		opts.Cycles = CyclesRef
	}
	if opts.Cycles != CyclesRef && opts.Cycles != CyclesThrow {
		return opts, fmt.Errorf("%w: Cycles=%q", ErrInvalidJSONSchemaOption, opts.Cycles)
	}

	if opts.Reused == "" {
		opts.Reused = ReusedInline
	}
	if opts.Reused != ReusedInline && opts.Reused != ReusedRef {
		return opts, fmt.Errorf("%w: Reused=%q", ErrInvalidJSONSchemaOption, opts.Reused)
	}

	if opts.IO == "" {
		opts.IO = IOOutput
	}
	if opts.IO != IOOutput && opts.IO != IOInput {
		return opts, fmt.Errorf("%w: IO=%q", ErrInvalidJSONSchemaOption, opts.IO)
	}

	return opts, nil
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
			s.Defs = make(map[string]*lib.Schema, len(c.defs))
		}
		for _, key := range slices.Sorted(maps.Keys(c.defs)) {
			s.Defs[key] = c.defs[key]
		}
	}

	return s, nil
}

type registryEntry struct {
	schema core.ZodSchema
	meta   core.GlobalMeta
}

// toJSONSchemaRegistry handles the conversion of a schema Registry.
func toJSONSchemaRegistry(reg *core.Registry[core.GlobalMeta], opts Options) (*lib.Schema, error) {
	var entries []registryEntry
	reg.Range(func(schema core.ZodSchema, meta core.GlobalMeta) bool {
		meta = cloneutil.Clone(meta).(core.GlobalMeta)
		entries = append(entries, registryEntry{schema: schema, meta: meta})
		return true
	})
	seenIDs := make(map[string]struct{}, len(entries))
	for _, entry := range entries {
		if entry.meta.ID == "" {
			return nil, fmt.Errorf("registry schema ID is missing: %w", ErrInvalidRegistrySchemaID)
		}
		if _, exists := seenIDs[entry.meta.ID]; exists {
			return nil, fmt.Errorf("registry schema ID is duplicate %q: %w", entry.meta.ID, ErrInvalidRegistrySchemaID)
		}
		seenIDs[entry.meta.ID] = struct{}{}
	}
	slices.SortFunc(entries, func(a, b registryEntry) int {
		return cmp.Compare(a.meta.ID, b.meta.ID)
	})

	c := newConverter(opts)
	c.batchMetadata = make(map[core.ZodSchema]core.GlobalMeta, len(entries))
	for _, entry := range entries {
		c.batchMetadata[entry.schema] = entry.meta
	}

	// First pass: process all schemas to populate seen map and defs.
	for _, entry := range entries {
		if _, err := c.convert(entry.schema); err != nil {
			return nil, err
		}
	}

	// Create a root schema to hold all definitions.
	rootSchema := &lib.Schema{}
	if len(c.defs) > 0 {
		rootSchema.Defs = make(map[string]*lib.Schema, len(c.defs))
		for _, key := range slices.Sorted(maps.Keys(c.defs)) {
			rootSchema.Defs[key] = c.defs[key]
		}
	}

	return rootSchema, nil
}

// converter holds the state for a single conversion run.
type converter struct {
	opts          Options
	seen          map[core.ZodSchema]*lib.Schema
	counts        map[core.ZodSchema]int
	refs          map[core.ZodSchema]string
	batchMetadata map[core.ZodSchema]core.GlobalMeta
	auto          int
	path          []string
	defs          map[string]*lib.Schema
	idCache       map[core.ZodSchema]string         // cache for getID results
	unwrapCache   map[core.ZodSchema]core.ZodSchema // cache for unwrapSchema results
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
// by following a Inner() method if implemented. This allows features like
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
		if getter, ok := s.(interface{ Inner() core.ZodSchema }); ok {
			inner := getter.Inner()
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

	internals := schema.Internals()

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
		// Skip primitives -- only composite schemas are worth extracting.
		if isCompositeType(internals.Type) && !internals.IsOptional() && !internals.IsNilable() {
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
	internals := schema.Internals()
	bag := cloneBag(internals)
	var jsonSchema *lib.Schema
	var err error

	switch internals.Type {
	case core.ZodTypeString,
		core.ZodTypeIPv4, core.ZodTypeIPv6, core.ZodTypeHostname, core.ZodTypeMAC, core.ZodTypeE164,
		core.ZodTypeCIDRv4, core.ZodTypeCIDRv6, core.ZodTypeURL:
		jsonSchema = &lib.Schema{Type: []string{"string"}}
		c.applyStringBag(jsonSchema, bag)
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
			if inp, ok := schema.(interface{ Inner() core.ZodSchema }); ok {
				return c.convert(inp.Inner())
			}
		} else { // output mode (default)
			if outpHolder, ok := schema.(interface{ Output() core.ZodSchema }); ok {
				tgt := outpHolder.Output()
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
			if inp, ok := schema.(interface{ Inner() core.ZodSchema }); ok {
				return c.convert(inp.Inner())
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
		jsonSchema, err = c.convertFile(bag)
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
		jsonSchema = &lib.Schema{Type: []string{"string"}, Format: new("date-time")}
	case core.ZodTypeEmail:
		jsonSchema = &lib.Schema{Type: []string{"string"}, Format: new("email")}
	case core.ZodTypeTime:
		jsonSchema = &lib.Schema{Type: []string{"string"}, Format: new("time")}
	case core.ZodTypeISODateTime, core.ZodTypeIso:
		jsonSchema = &lib.Schema{Type: []string{"string"}, Format: new("date-time")}
	case core.ZodTypeISODate:
		jsonSchema = &lib.Schema{Type: []string{"string"}, Format: new("date")}
	case core.ZodTypeISOTime:
		jsonSchema = &lib.Schema{Type: []string{"string"}, Format: new("time")}
	case core.ZodTypeISODuration:
		jsonSchema = &lib.Schema{Type: []string{"string"}, Format: new("duration")}
	case core.ZodTypeOptional, core.ZodTypeNilable, core.ZodTypeDefault, core.ZodTypePrefault, core.ZodTypeRefine, core.ZodTypeCheck:
		if s, ok := schema.(interface{ Inner() core.ZodSchema }); ok {
			return c.convert(s.Inner())
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

	if len(bag) > 0 {
		c.applyBag(jsonSchema, bag)
	}

	return jsonSchema, nil
}

// catchaller is an interface for schemas that have a catch-all schema.
type catchaller interface {
	Catchall() core.ZodSchema
}

// unknownKeysHandler is an interface for schemas that handle unknown keys.
// We use a generic method signature to avoid circular imports.
type unknownKeysHandler interface {
	UnknownKeys() any
}

// unwrapper is an interface for schemas that can unwrap to reveal inner schemas.
type unwrapper interface {
	Unwrap() core.ZodType[any]
}

// booleanSchema returns a JSON Schema boolean schema (true or false).
func booleanSchema(value bool) *lib.Schema {
	return &lib.Schema{Boolean: &value}
}

// resolveUnknownKeysMode determines if a schema uses passthrough mode for unknown keys.
// Returns true for passthrough, false for strict/strip/default.
func resolveUnknownKeysMode(schema core.ZodSchema) bool {
	// Prefer the fast interface assertion path.
	if uk, ok := schema.(unknownKeysHandler); ok {
		modeStr := fmt.Sprint(uk.UnknownKeys())
		return modeStr == string(types.ObjectModePassthrough) || modeStr == "passthrough"
	}
	// Fallback to reflection.
	if method := reflect.ValueOf(schema).MethodByName("UnknownKeys"); method.IsValid() {
		if results := method.Call(nil); len(results) == 1 {
			return fmt.Sprint(results[0].Interface()) == "passthrough"
		}
	}
	return false
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
		propJSONSchema, err := c.convert(propSchema)
		if err != nil {
			return nil, err
		}
		properties[key] = propJSONSchema
		c.path = c.path[:len(c.path)-2]

		isRequired := !propSchema.Internals().IsOptional()
		// In "input" mode, fields with defaults are not required.
		if c.opts.IO == "input" {
			pInternals := propSchema.Internals()
			if pInternals.NilInputUsesDefault() {
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
		jsonSchema.Properties = new(lib.SchemaMap(properties))
	}
	if len(required) > 0 {
		slices.SortFunc(required, func(a, b string) int { return cmp.Compare(b, a) })
		jsonSchema.Required = required
	}

	// Handle additionalProperties based on catchall and unknown keys mode
	// Try to call UnknownKeys and Catchall methods
	if s, ok := schema.(catchaller); ok {
		if catchallSchema := s.Catchall(); catchallSchema != nil {
			// If there's a catchall schema, convert it to additionalProperties
			c.path = append(c.path, "additionalProperties")
			catchallJSONSchema, err := c.convert(catchallSchema)
			if err != nil {
				return nil, err
			}
			jsonSchema.AdditionalProperties = catchallJSONSchema
			c.path = c.path[:len(c.path)-1]
		}
	}

	if jsonSchema.AdditionalProperties == nil {
		jsonSchema.AdditionalProperties = booleanSchema(resolveUnknownKeysMode(schema))
	}

	return jsonSchema, nil
}

func (c *converter) convertArray(schema core.ZodSchema) (*lib.Schema, error) {
	jsonSchema := &lib.Schema{Type: []string{"array"}}

	// Handle ZodSlice (variable-length arrays)
	if s, ok := schema.(interface{ Element() core.ZodSchema }); ok {
		elemSchema := s.Element()
		if elemSchema == nil {
			return nil, ErrSliceElementNotSchema
		}
		items, err := c.convert(elemSchema)
		if err != nil {
			return nil, err
		}
		jsonSchema.Items = items
	} else if s, ok := schema.(interface {
		Items() []core.ZodSchema
		Rest() core.ZodSchema
	}); ok { // Handle ZodArray (tuples)
		elements := s.Items()
		jsonSchema.PrefixItems = make([]*lib.Schema, len(elements))
		for i, itemSchema := range elements {
			if itemSchema == nil {
				return nil, fmt.Errorf("%w: index %d", ErrArrayItemNotSchema, i)
			}
			converted, err := c.convert(itemSchema)
			if err != nil {
				return nil, err
			}
			jsonSchema.PrefixItems[i] = converted
		}

		if rest := s.Rest(); rest != nil {
			items, err := c.convert(rest)
			if err != nil {
				return nil, err
			}
			jsonSchema.Items = items
		} else {
			// If only one tuple element and no rest, treat as standard variable-length array for compatibility
			if len(elements) == 1 {
				jsonSchema.Items = jsonSchema.PrefixItems[0]
				jsonSchema.PrefixItems = nil
			} else {
				// Fixed-length tuple
				jsonSchema.MinItems = new(float64(len(elements)))
				jsonSchema.MaxItems = new(float64(len(elements)))
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
		Items() []core.ZodSchema
		Rest() core.ZodSchema
	})
	if !ok {
		return nil, fmt.Errorf("%w: expected tuple schema with Items/Rest methods", ErrUnhandledArrayLike)
	}

	items := tupleSchema.Items()
	rest := tupleSchema.Rest()

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
		// No rest element - fixed length tuple.
		minItems := 0
		for i := len(items) - 1; i >= 0; i-- {
			if !items[i].Internals().IsOptional() {
				minItems = i + 1
				break
			}
		}
		minItemsFloat := float64(minItems)
		jsonSchema.MinItems = &minItemsFloat
		jsonSchema.MaxItems = new(float64(len(items)))
	}

	return jsonSchema, nil
}

// applyMeta copies the selected metadata onto the generated JSON Schema node.
func (c *converter) applyMeta(schema core.ZodSchema, jsonSchema *lib.Schema) {
	if jsonSchema == nil || schema == nil {
		return
	}

	meta, ok := c.lookupMeta(schema)
	if !ok {
		return
	}

	if meta.Title != "" && jsonSchema.Title == nil {
		jsonSchema.Title = new(meta.Title)
	}
	if meta.Description != "" && (jsonSchema.Description == nil || *jsonSchema.Description == "") {
		jsonSchema.Description = new(meta.Description)
	}
	if len(meta.Examples) > 0 && len(jsonSchema.Examples) == 0 {
		jsonSchema.Examples = cloneutil.Clone(meta.Examples).([]any)
	}
}

// getID retrieves the selected metadata ID for schema.
func (c *converter) getID(schema core.ZodSchema) string {
	if id, ok := c.idCache[schema]; ok {
		return id
	}
	meta, ok := c.lookupMeta(schema)
	if !ok {
		return ""
	}
	c.idCache[schema] = meta.ID
	return meta.ID
}

// lookupMeta selects an explicit registry entry or falls back to schema-owned metadata.
func (c *converter) lookupMeta(schema core.ZodSchema) (core.GlobalMeta, bool) {
	if c.batchMetadata != nil {
		if meta, ok := lookupSnapshotMeta(c.batchMetadata, schema); ok {
			return meta, true
		}
	} else if c.opts.Metadata != nil {
		if meta, ok := lookupRegistryMeta(c.opts.Metadata, schema); ok {
			return meta, true
		}
	}
	for schema != nil {
		meta := schema.Internals().Metadata()
		if hasMetadata(meta) {
			return meta, true
		}
		inner, ok := schema.(interface{ Inner() core.ZodSchema })
		if !ok {
			break
		}
		next := inner.Inner()
		if next == schema {
			break
		}
		schema = next
	}
	return core.GlobalMeta{}, false
}

func lookupSnapshotMeta(
	metadata map[core.ZodSchema]core.GlobalMeta,
	schema core.ZodSchema,
) (core.GlobalMeta, bool) {
	if meta, ok := metadata[schema]; ok {
		return meta, true
	}
	if inner, ok := schema.(interface{ Inner() core.ZodSchema }); ok {
		meta, found := metadata[inner.Inner()]
		return meta, found
	}
	return core.GlobalMeta{}, false
}

func lookupRegistryMeta(
	registry *core.Registry[core.GlobalMeta],
	schema core.ZodSchema,
) (core.GlobalMeta, bool) {
	if meta, ok := registry.Get(schema); ok {
		return meta, true
	}
	if inner, ok := schema.(interface{ Inner() core.ZodSchema }); ok {
		return registry.Get(inner.Inner())
	}
	return core.GlobalMeta{}, false
}

func hasMetadata(meta core.GlobalMeta) bool {
	return meta.ID != "" || meta.Title != "" || meta.Description != "" || len(meta.Examples) > 0
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
			if schema.Internals().IsOptional() && !schema.Internals().IsNilable() {
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

// isNullSchema checks if a schema represents null/nil type.
func isNullSchema(schema core.ZodSchema) bool {
	return schema.Internals().Type == core.ZodTypeNil
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

	// For loose records with regex-based key patterns, emit patternProperties
	// instead of propertyNames for more semantically correct JSON Schema that
	// composes better with allOf/intersections (Zod v4: e01cd02b).
	if looseSchema, ok := schema.(interface{ IsLoose() bool }); ok && looseSchema.IsLoose() {
		if propertyNames != nil && propertyNames.Pattern != nil {
			patternProps := lib.SchemaMap{
				*propertyNames.Pattern: additionalProperties,
			}
			return &lib.Schema{
				Type:              []string{"object"},
				PatternProperties: &patternProps,
			}, nil
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
				for i := range slice.Len() {
					enumValues = append(enumValues, slice.Index(i).Interface())
				}
			}
		}
	}
	if len(enumValues) == 0 {
		return nil, ErrEnumExtractValues
	}

	// Ensure deterministic order for enum values to avoid map iteration randomness
	switch enumValues[0].(type) {
	case string:
		slices.SortStableFunc(enumValues, func(a, b any) int {
			return cmp.Compare(a.(string), b.(string))
		})
	case int, int32, int64, uint, uint32, uint64, float64, float32:
		slices.SortStableFunc(enumValues, func(a, b any) int {
			return cmp.Compare(fmt.Sprintf("%v", a), fmt.Sprintf("%v", b))
		})
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
				for i := range sliceValue.Len() {
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
			for i := range rv.Len() {
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

// compositeTypes lists schema types that are worth extracting into $defs for reuse.
var compositeTypes = map[core.ZodTypeCode]struct{}{
	core.ZodTypeObject:       {},
	core.ZodTypeStruct:       {},
	core.ZodTypeSlice:        {},
	core.ZodTypeArray:        {},
	core.ZodTypeRecord:       {},
	core.ZodTypeUnion:        {},
	core.ZodTypeIntersection: {},
}

// isCompositeType returns true for schema types worth extracting into $defs.
func isCompositeType(t core.ZodTypeCode) bool {
	_, ok := compositeTypes[t]
	return ok
}

// convertLazy resolves inner schema and delegates conversion.
func (c *converter) convertLazy(schema core.ZodSchema) (*lib.Schema, error) {
	s, ok := schema.(unwrapper)
	if !ok {
		return c.unrepresentableLazy()
	}
	innerValue := any(s.Unwrap())
	if wrapper, ok := innerValue.(interface{ Inner() any }); ok {
		innerValue = wrapper.Inner()
	}
	inner, ok := innerValue.(core.ZodSchema)
	if !ok || inner == nil {
		return c.unrepresentableLazy()
	}

	if _, found := c.seen[inner]; found {
		if c.opts.Cycles == CyclesThrow {
			return nil, ErrCircularReference
		}
		if id := c.getID(inner); id != "" {
			return &lib.Schema{Ref: "#/$defs/" + id}, nil
		}
		if name, ok := c.refs[c.unwrapSchema(inner)]; ok {
			return &lib.Schema{Ref: "#/$defs/" + name}, nil
		}
		return &lib.Schema{Ref: "#"}, nil
	}
	return c.convert(inner)
}

func (c *converter) unrepresentableLazy() (*lib.Schema, error) {
	if c.opts.Unrepresentable == UnrepresentableAny {
		return &lib.Schema{}, nil
	}
	return nil, fmt.Errorf("%w: unresolved lazy target", ErrUnrepresentableType)
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

	if keySchema.Internals().Type != core.ZodTypeString {
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

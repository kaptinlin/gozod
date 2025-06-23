package types

import (
	"errors"
	"reflect"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/pkg/structx"
)

// Error definitions for struct transformations
var (
	ErrTransformNilStruct = errors.New("cannot transform nil struct value")
	ErrExpectedStruct     = errors.New("expected struct type")
)

//////////////////////////////////////////
//////////////////////////////////////////

// =============================================================================
// CORE TYPE DEFINITIONS
// =============================================================================

// ZodStructDef defines the configuration for struct/object validation
type ZodStructDef struct {
	core.ZodTypeDef
	Type     core.ZodTypeCode       // "struct"
	Shape    core.StructSchema      // Field schemas
	Catchall core.ZodType[any, any] // Schema for unrecognized keys
	Mode     string                 // "strict", "strip", "loose"
}

// ZodStructInternals contains struct validator internal state
type ZodStructInternals struct {
	core.ZodTypeInternals
	Def      *ZodStructDef          // Schema definition
	Shape    core.StructSchema      // Field schemas
	Mode     string                 // Validation mode
	Catchall core.ZodType[any, any] // Catchall schema
	Bag      map[string]any         // Runtime configuration and custom checks
}

// ZodStruct represents a struct/object validation schema with type safety
type ZodStruct struct {
	internals *ZodStructInternals
}

// Mode constants for struct validation
const (
	STRICT_MODE = "strict" // Error on unknown keys
	STRIP_MODE  = "strip"  // Strip unknown keys (default)
	LOOSE_MODE  = "loose"  // Allow unknown keys
)

// =============================================================================
// VALIDATION METHODS
// =============================================================================

// GetInternals returns the internal state of the schema
func (z *ZodStruct) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// GetZod returns the struct-specific internals for framework usage
func (z *ZodStruct) GetZod() *ZodStructInternals {
	return z.internals
}

// Shape provides access to the internal field schemas
func (z *ZodStruct) Shape() core.StructSchema {
	return z.internals.Shape
}

// Parse validates and parses with smart type inference
func (z *ZodStruct) Parse(input any, ctx ...*core.ParseContext) (any, error) {
	// ---------------------------------------------------------------------
	// 1. nil check
	// ---------------------------------------------------------------------
	if input == nil {
		if z.internals.Nilable {
			return nil, nil
		}
		rawIssue := issues.CreateInvalidTypeIssue("struct", input)
		finalIssue := issues.FinalizeIssue(rawIssue, nil, core.GetConfig())
		return nil, issues.NewZodError([]core.ZodIssue{finalIssue})
	}

	// ---------------------------------------------------------------------
	// 2. convert input to map[string]any for internal validation
	// ---------------------------------------------------------------------
	var (
		objectMap         map[string]any
		fromMapPointer    bool // input is *map[string]any
		returnOriginalObj bool // input is struct or *struct, need to keep original return
	)

	// helper: convert nested struct to map[string]any
	var convertNested func(any) any
	convertNested = func(val any) any {
		if val == nil {
			return nil
		}

		// if map, recursively convert its value
		if m, ok := val.(map[string]any); ok {
			for k, v := range m {
				m[k] = convertNested(v)
			}
			return m
		}

		rv := reflect.ValueOf(val)
		if rv.Kind() == reflect.Ptr {
			if rv.IsNil() {
				return val
			}
			rv = rv.Elem()
		}
		if rv.Kind() == reflect.Struct {
			nestedMap := structx.Marshal(val)
			if nestedMap != nil {
				return convertNested(nestedMap)
			}
		}

		return val
	}

	switch v := input.(type) {
	case map[string]any:
		objectMap = v
	case *map[string]any:
		// *map[string]any needs special handling for nil
		if v == nil {
			if z.internals.Nilable {
				return (*map[string]any)(nil), nil
			}
			rawIssue := issues.CreateInvalidTypeIssue("struct", input)
			finalIssue := issues.FinalizeIssue(rawIssue, nil, core.GetConfig())
			return nil, issues.NewZodError([]core.ZodIssue{finalIssue})
		}
		objectMap = *v
		fromMapPointer = true
	default:
		// Check if it is a struct pointer and is nil
		rv := reflect.ValueOf(v)
		if rv.Kind() == reflect.Ptr && rv.IsNil() {
			if z.internals.Nilable {
				return v, nil
			}
			rawIssue := issues.CreateInvalidTypeIssue("struct", input)
			finalIssue := issues.FinalizeIssue(rawIssue, nil, core.GetConfig())
			return nil, issues.NewZodError([]core.ZodIssue{finalIssue})
		}

		// try to convert struct (value or pointer) to map[string]any
		if m := structx.Marshal(v); m != nil {
			objectMap = convertNested(m).(map[string]any)
			returnOriginalObj = true
		} else {
			// non map/struct type, error
			rawIssue := issues.CreateInvalidTypeIssue("struct", input)
			finalIssue := issues.FinalizeIssue(rawIssue, nil, core.GetConfig())
			return nil, issues.NewZodError([]core.ZodIssue{finalIssue})
		}
	}

	// Validate the object using unified parsing infrastructure
	// Use constructor instead of direct struct literal to respect private fields
	payload := core.NewParsePayloadWithPath(objectMap, make([]any, 0))

	var parseCtx *core.ParseContext
	if len(ctx) > 0 {
		parseCtx = ctx[0]
	}

	result := z.internals.Parse(payload, parseCtx)
	// Use getter methods instead of direct field access
	if len(result.GetIssues()) > 0 {
		finalizedIssues := make([]core.ZodIssue, len(result.GetIssues()))
		for i, rawIssue := range result.GetIssues() {
			finalizedIssues[i] = issues.FinalizeIssue(rawIssue, parseCtx, core.GetConfig())
		}
		return nil, issues.NewZodError(finalizedIssues)
	}

	// Run checks on the validated object
	if len(z.internals.Checks) > 0 {
		// Use constructor and getter methods instead of direct field access
		checksPayload := core.NewParsePayload(result.GetValue())
		engine.RunChecksOnValue(result.GetValue(), z.internals.Checks, checksPayload, parseCtx)
		if len(checksPayload.GetIssues()) > 0 {
			finalizedIssues := make([]core.ZodIssue, len(checksPayload.GetIssues()))
			for i, rawIssue := range checksPayload.GetIssues() {
				finalizedIssues[i] = issues.FinalizeIssue(rawIssue, parseCtx, core.GetConfig())
			}
			return nil, issues.NewZodError(finalizedIssues)
		}
	}

	// Use getter method instead of direct field access
	validatedMap := result.GetValue().(map[string]any)

	// keep pointer return for *map[string]any
	if fromMapPointer {
		return &validatedMap, nil
	}

	// if input is struct or *struct, return original input (keep type consistency)
	if returnOriginalObj {
		return input, nil
	}

	// default return map
	return validatedMap, nil
}

// MustParse validates and parses, panicking on failure
func (z *ZodStruct) MustParse(input any, ctx ...*core.ParseContext) any {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// CloneFrom implements Cloneable interface
func (z *ZodStruct) CloneFrom(source any) {
	if src, ok := source.(interface{ GetZod() *ZodStructInternals }); ok {
		srcState := src.GetZod()
		tgtState := z.GetZod()

		// Copy Shape
		if srcState.Shape != nil {
			tgtState.Shape = make(core.StructSchema)
			for key, value := range srcState.Shape {
				tgtState.Shape[key] = value
			}
		}

		// Copy Mode
		tgtState.Mode = srcState.Mode

		// Copy Catchall
		tgtState.Catchall = srcState.Catchall
	}
}

// Refine adds type-safe custom validation logic
func (z *ZodStruct) Refine(fn func(map[string]any) bool, params ...any) *ZodStruct {
	result := z.RefineAny(func(v any) bool {
		if objMap, ok := v.(map[string]any); ok {
			return fn(objMap)
		}
		return false
	}, params...)
	return result.(*ZodStruct)
}

// RefineAny adds flexible custom validation logic
func (z *ZodStruct) RefineAny(fn func(any) bool, params ...any) core.ZodType[any, any] {
	check := checks.NewCustom[any](fn, params...)
	return engine.AddCheck(z, check)
}

// TransformAny adds flexible transformation logic
func (z *ZodStruct) TransformAny(fn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	transform := Transform[any, any](fn)
	return &ZodPipe[any, any]{
		in:  any(z).(core.ZodType[any, any]),
		out: any(transform).(core.ZodType[any, any]),
	}
}

// Pipe connects this schema to another schema
func (z *ZodStruct) Pipe(schema core.ZodType[any, any]) core.ZodType[any, any] {
	return &ZodPipe[any, any]{
		in:  any(z).(core.ZodType[any, any]),
		out: schema,
	}
}

//////////////////////////
// MODIFIER METHODS
//////////////////////////

// Optional makes the struct optional
func (z *ZodStruct) Optional() core.ZodType[any, any] {
	return Optional(any(z).(core.ZodType[any, any]))
}

// Nilable makes the struct nilable while preserving type inference
func (z *ZodStruct) Nilable() core.ZodType[any, any] {
	z.internals.SetNilable()
	return z
}

// Nullish makes the struct both optional and nilable
func (z *ZodStruct) Nullish() core.ZodType[any, any] {
	return Nullish(any(z).(core.ZodType[any, any]))
}

// Unwrap returns the inner type (for basic types, returns self)
func (z *ZodStruct) Unwrap() core.ZodType[any, any] {
	return any(z).(core.ZodType[any, any])
}

//////////////////////////
// WRAPPER TYPES
//////////////////////////

// DEFAULT METHODS

// Default creates a default wrapper with type safety
func (z *ZodStruct) Default(value map[string]any) core.ZodType[any, any] {
	return Default(z, value)
}

// DefaultFunc creates a default wrapper with function
func (z *ZodStruct) DefaultFunc(fn func() map[string]any) core.ZodType[any, any] {
	return DefaultFunc(z, func() any { return fn() })
}

// PREFAULT METHODS

// Prefault creates a prefault wrapper with type safety
func (z *ZodStruct) Prefault(value map[string]any) core.ZodType[any, any] {
	return Prefault(any(z).(core.ZodType[any, any]), value)
}

// PrefaultFunc creates a prefault wrapper with function
func (z *ZodStruct) PrefaultFunc(fn func() map[string]any) core.ZodType[any, any] {
	return PrefaultFunc(any(z).(core.ZodType[any, any]), func() any { return fn() })
}

//////////////////////////
// STRUCT MANIPULATION METHODS
//////////////////////////

// Pick creates a new struct with only the specified keys
func (z *ZodStruct) Pick(keys []string) *ZodStruct {
	newShape := make(core.StructSchema)
	for _, key := range keys {
		if schema, exists := z.internals.Shape[key]; exists {
			newShape[key] = schema
		}
	}

	def := &ZodStructDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeStruct,
			Checks: make([]core.ZodCheck, 0),
		},
		Type:     core.ZodTypeStruct,
		Shape:    newShape,
		Catchall: z.internals.Catchall,
	}

	return createZodStructFromDef(def)
}

// Omit creates a new struct without the specified keys
func (z *ZodStruct) Omit(keys []string) *ZodStruct {
	keySet := make(map[string]struct{})
	for _, key := range keys {
		keySet[key] = struct{}{}
	}

	newShape := make(core.StructSchema)
	for key, schema := range z.internals.Shape {
		if _, shouldOmit := keySet[key]; !shouldOmit {
			newShape[key] = schema
		}
	}

	def := &ZodStructDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeStruct,
			Checks: make([]core.ZodCheck, 0),
		},
		Type:     core.ZodTypeStruct,
		Shape:    newShape,
		Catchall: z.internals.Catchall,
	}

	return createZodStructFromDef(def)
}

// Extend adds new fields to the struct
func (z *ZodStruct) Extend(augmentation core.StructSchema) *ZodStruct {
	newShape := make(core.StructSchema)

	// Copy existing shape
	for key, schema := range z.internals.Shape {
		newShape[key] = schema
	}

	// Add new fields
	for key, schema := range augmentation {
		newShape[key] = schema
	}

	def := &ZodStructDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeStruct,
			Checks: make([]core.ZodCheck, 0),
		},
		Type:     core.ZodTypeStruct,
		Shape:    newShape,
		Catchall: z.internals.Catchall,
	}

	return createZodStructFromDef(def)
}

// Partial makes specified fields optional
func (z *ZodStruct) Partial(keys ...[]string) *ZodStruct {
	var targetKeys map[string]struct{}

	if len(keys) > 0 && len(keys[0]) > 0 {
		// Partial with specific keys
		targetKeys = make(map[string]struct{})
		for _, key := range keys[0] {
			targetKeys[key] = struct{}{}
		}
	}

	newShape := make(core.StructSchema)
	for key, schema := range z.internals.Shape {
		if targetKeys == nil || len(targetKeys) == 0 {
			// Make all fields optional
			newShape[key] = Optional(schema)
		} else if _, shouldMakeOptional := targetKeys[key]; shouldMakeOptional {
			// Make specific field optional
			newShape[key] = Optional(schema)
		} else {
			// Keep field as is
			newShape[key] = schema
		}
	}

	def := &ZodStructDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeStruct,
			Checks: make([]core.ZodCheck, 0),
		},
		Type:     core.ZodTypeStruct,
		Shape:    newShape,
		Catchall: z.internals.Catchall,
	}

	return createZodStructFromDef(def)
}

// Merge combines two struct schemas
func (z *ZodStruct) Merge(other *ZodStruct) *ZodStruct {
	newShape := make(core.StructSchema)

	// Copy current shape
	for key, schema := range z.internals.Shape {
		newShape[key] = schema
	}

	// Merge other shape (overwrites conflicting keys)
	for key, schema := range other.internals.Shape {
		newShape[key] = schema
	}

	def := &ZodStructDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeStruct,
			Checks: make([]core.ZodCheck, 0),
		},
		Type:     core.ZodTypeStruct,
		Shape:    newShape,
		Catchall: z.internals.Catchall,
	}

	return createZodStructFromDef(def)
}

// Catchall sets a schema for unknown keys and switches to loose mode
func (z *ZodStruct) Catchall(catchallSchema core.ZodType[any, any]) *ZodStruct {
	def := &ZodStructDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeStruct,
			Checks: make([]core.ZodCheck, 0),
		},
		Type:     core.ZodTypeStruct,
		Shape:    z.internals.Shape,
		Catchall: catchallSchema,
		Mode:     LOOSE_MODE,
	}

	return createZodStructFromDef(def)
}

// Passthrough allows unknown keys to pass through
func (z *ZodStruct) Passthrough() *ZodStruct {
	def := &ZodStructDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeStruct,
			Checks: make([]core.ZodCheck, 0),
		},
		Type:     core.ZodTypeStruct,
		Shape:    z.internals.Shape,
		Catchall: z.internals.Catchall,
		Mode:     LOOSE_MODE,
	}

	return createZodStructFromDef(def)
}

// Strict sets strict mode (error on unknown keys)
func (z *ZodStruct) Strict() *ZodStruct {
	def := &ZodStructDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeStruct,
			Checks: make([]core.ZodCheck, 0),
		},
		Type:     core.ZodTypeStruct,
		Shape:    z.internals.Shape,
		Catchall: z.internals.Catchall,
		Mode:     STRICT_MODE,
	}

	return createZodStructFromDef(def)
}

// Strip sets strip mode (remove unknown keys)
func (z *ZodStruct) Strip() *ZodStruct {
	def := &ZodStructDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeStruct,
			Checks: make([]core.ZodCheck, 0),
		},
		Type:     core.ZodTypeStruct,
		Shape:    z.internals.Shape,
		Catchall: z.internals.Catchall,
		Mode:     STRIP_MODE,
	}

	return createZodStructFromDef(def)
}

// Keyof creates a union of all struct keys
func (z *ZodStruct) Keyof() core.ZodType[any, any] {
	keys := make([]core.ZodType[any, any], 0, len(z.internals.Shape))
	for key := range z.internals.Shape {
		keys = append(keys, Literal(key))
	}

	if len(keys) == 0 {
		return any(Never()).(core.ZodType[any, any])
	}

	return any(Union(keys)).(core.ZodType[any, any])
}

//////////////////////////
// CONSTRUCTOR FUNCTIONS
//////////////////////////

// createZodStructFromDef creates a ZodStruct from definition
func createZodStructFromDef(def *ZodStructDef) *ZodStruct {
	internals := &ZodStructInternals{
		ZodTypeInternals: core.ZodTypeInternals{
			Version: core.Version,
			Type:    def.Type,
			Checks:  def.Checks,
			Coerce:  def.Coerce,
			Bag:     make(map[string]any),
		},
		Def:      def,
		Shape:    def.Shape,
		Mode:     def.Mode,
		Catchall: def.Catchall,
		Bag:      make(map[string]any),
	}

	// Initialize default mode if not set
	if internals.Mode == "" {
		internals.Mode = STRIP_MODE
	}

	internals.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any, any] {
		structDef := &ZodStructDef{
			ZodTypeDef: *newDef,
			Type:       core.ZodTypeStruct,
			Shape:      internals.Shape,
			Catchall:   internals.Catchall,
			Mode:       internals.Mode,
		}
		return any(createZodStructFromDef(structDef)).(core.ZodType[any, any])
	}

	// Set up the parse function
	internals.Parse = func(payload *core.ParsePayload, parseCtx *core.ParseContext) *core.ParsePayload {
		// Validate input is an object
		if payload.GetValue() == nil {
			if !internals.Nilable {
				issue := issues.CreateInvalidTypeIssue("struct", payload.GetValue())
				issue.Inst = internals
				payload.AddIssue(issue)
			}
			return payload
		}

		// Validate and process each field
		objMap, _ := payload.GetValue().(map[string]any)
		processedKeys := make(map[string]struct{})
		result := make(map[string]any)

		// Process defined fields
		for fieldKey, fieldSchema := range internals.Shape {
			processedKeys[fieldKey] = struct{}{}

			if fieldValue, exists := objMap[fieldKey]; exists {
				// Field exists, validate it - use constructor instead of direct struct literal
				fieldPayload := core.NewParsePayloadWithPath(fieldValue, append(payload.GetPath(), fieldKey))

				fieldResult := fieldSchema.GetInternals().Parse(fieldPayload, parseCtx)
				// Use getter methods instead of direct field access
				if len(fieldResult.GetIssues()) > 0 {
					// Add field issues with proper path
					for _, issue := range fieldResult.GetIssues() {
						payload.AddIssueWithPath(issue, append(payload.GetPath(), fieldKey))
					}
				} else {
					result[fieldKey] = fieldResult.GetValue()
				}
			} else {
				// Field missing, check if it's optional
				if optionalSchema, isOptional := fieldSchema.(interface{ IsOptional() bool }); isOptional && optionalSchema.IsOptional() {
					// Optional field, skip
					continue
				} else {
					// Required field missing
					issue := issues.CreateMissingKeyIssue(fieldKey, func(iss *core.ZodRawIssue) {
						iss.Inst = internals
					})
					payload.AddIssueWithPath(issue, append(payload.GetPath(), fieldKey))
				}
			}
		}

		// Handle unrecognized keys based on mode
		var unrecognizedKeys []string
		for key := range objMap {
			if _, processed := processedKeys[key]; !processed {
				unrecognizedKeys = append(unrecognizedKeys, key)
			}
		}

		if len(unrecognizedKeys) > 0 {
			switch internals.Mode {
			case STRICT_MODE:
				// Error on unrecognized keys
				issue := issues.CreateUnrecognizedKeysIssue(unrecognizedKeys, payload.GetValue())
				issue.Inst = internals
				payload.AddIssue(issue)
			case STRIP_MODE:
				// Strip unrecognized keys (do nothing)
			case LOOSE_MODE:
				// Allow unrecognized keys
				for _, key := range unrecognizedKeys {
					if internals.Catchall != nil {
						// Validate unknown key value using catchall schema's full Parse to preserve wrapper behaviour
						parsedVal, err := internals.Catchall.Parse(objMap[key], parseCtx)
						if err != nil {
							var zodErr *issues.ZodError
							if errors.As(err, &zodErr) {
								for _, issue := range zodErr.Issues {
									// Convert ZodError to RawIssue using standardized converter
									raw := issues.ConvertZodIssueToRaw(issue)

									// Prepend struct field key to issue path for catchall field errors - use getter method
									raw.Path = append(payload.GetPath(), append([]any{key}, issue.Path...)...)
									raw.Inst = internals

									payload.AddIssue(raw)
								}
							} else {
								raw := issues.CreateInvalidTypeIssue("unknown", objMap[key])
								raw.Path = append(payload.GetPath(), key)
								raw.Inst = internals
								payload.AddIssue(raw)
							}
						} else {
							result[key] = parsedVal
						}
					} else {
						result[key] = objMap[key]
					}
				}
			}
		}

		payload.SetValue(result)
		return payload
	}

	zodSchema := &ZodStruct{internals: internals}
	engine.InitZodType(any(zodSchema).(core.ZodType[any, any]), &def.ZodTypeDef)
	return zodSchema
}

// Struct creates a new struct schema
func Struct(shape core.StructSchema, params ...any) *ZodStruct {
	def := &ZodStructDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeStruct,
			Checks: make([]core.ZodCheck, 0),
		},
		Type:     core.ZodTypeStruct,
		Shape:    shape,
		Catchall: nil,
		Mode:     STRIP_MODE,
	}

	schema := createZodStructFromDef(def)

	if len(params) > 0 {
		param := params[0]

		// Handle different parameter types
		switch p := param.(type) {
		case string:
			// String parameter becomes Error field
			errorMap := issues.CreateErrorMap(p)
			if errorMap != nil {
				def.Error = errorMap
				schema.internals.Error = errorMap
			}
		case core.SchemaParams:
			// Handle core.SchemaParams
			if p.Error != nil {
				errorMap := issues.CreateErrorMap(p.Error)
				if errorMap != nil {
					def.Error = errorMap
					schema.internals.Error = errorMap
				}
			}
			if p.Description != "" {
				schema.internals.Bag["description"] = p.Description
			}
			if p.Abort {
				schema.internals.Bag["abort"] = true
			}
			if len(p.Path) > 0 {
				schema.internals.Bag["path"] = p.Path
			}
			if len(p.Params) > 0 {
				schema.internals.Bag["params"] = p.Params
			}
		}
	}

	return schema
}

// StrictStruct creates a struct schema in strict mode
func StrictStruct(shape core.StructSchema, params ...any) *ZodStruct {
	schema := Struct(shape, params...)
	schema.internals.Mode = STRICT_MODE
	schema.internals.Def.Mode = STRICT_MODE
	return schema
}

// LooseStruct creates a struct schema in loose mode
func LooseStruct(shape core.StructSchema, params ...any) *ZodStruct {
	schema := Struct(shape, params...)
	schema.internals.Mode = LOOSE_MODE
	schema.internals.Def.Mode = LOOSE_MODE
	return schema
}

// =============================================================================
// UTILITY FUNCTIONS
// =============================================================================

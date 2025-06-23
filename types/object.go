package types

import (
	"errors"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/pkg/mapx"
	"github.com/kaptinlin/gozod/pkg/reflectx"
)

// Error definitions for object transformations
var (
	ErrTransformNilObject  = errors.New("cannot transform nil object")
	ErrCannotConvertObject = errors.New("cannot convert to object")
)

// =============================================================================
// OBJECT MODE CONSTANTS
// =============================================================================

// ObjectMode defines how to handle unknown keys in object validation
type ObjectMode string

const (
	OBJECT_STRICT_MODE ObjectMode = "strict" // Error on unknown keys
	OBJECT_STRIP_MODE  ObjectMode = "strip"  // Strip unknown keys (default)
	OBJECT_LOOSE_MODE  ObjectMode = "loose"  // Allow unknown keys
)

// =============================================================================
// OBJECT TYPE DEFINITIONS (Three-Layer Architecture)
// =============================================================================

// ZodObjectDef defines object validation configuration
type ZodObjectDef struct {
	core.ZodTypeDef
	Type        core.ZodTypeCode       // "object"
	Shape       core.ObjectSchema      // Field definitions
	Catchall    core.ZodType[any, any] // Catchall schema for unknown keys
	UnknownKeys ObjectMode             // How to handle unknown keys
}

// ZodObjectInternals contains object validator internal state
type ZodObjectInternals struct {
	core.ZodTypeInternals
	Def   *ZodObjectDef     // Schema definition
	Shape core.ObjectSchema // Field definitions map
	Bag   map[string]any    // Runtime configuration
}

// ZodObject represents fixed-field object validation
type ZodObject struct {
	internals *ZodObjectInternals
}

// Shape provides access to the internal field schemas
func (z *ZodObject) Shape() core.ObjectSchema {
	return z.internals.Shape
}

// =============================================================================
// CONSTRUCTOR FUNCTIONS
// =============================================================================

// Object creates object schema for fixed-field validation
func Object(shape core.ObjectSchema, params ...any) *ZodObject {
	def := &ZodObjectDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   "object",
			Checks: make([]core.ZodCheck, 0),
		},
		Type:        "object",
		Shape:       shape,
		UnknownKeys: OBJECT_STRIP_MODE, // Default mode
	}

	schema := createZodObjectFromDef(def)

	// Apply schema parameters
	if len(params) > 0 {
		if param, ok := params[0].(core.SchemaParams); ok {
			// Store coerce flag in bag

			// Handle schema-level error configuration
			if param.Error != nil {
				errorMap := issues.CreateErrorMap(param.Error)
				if errorMap != nil {
					def.Error = errorMap
					schema.internals.Error = errorMap
				}
			}

			// Store additional parameters
			if param.Params != nil {
				for key, value := range param.Params {
					schema.internals.Bag[key] = value
				}
			}
		}
	}

	return schema
}

// StrictObject creates strict object (disallows unknown keys)
func StrictObject(shape core.ObjectSchema, params ...any) *ZodObject {
	schema := Object(shape, params...)
	schema.internals.Def.UnknownKeys = OBJECT_STRICT_MODE
	return schema
}

// LooseObject creates loose object (allows unknown keys)
func LooseObject(shape core.ObjectSchema, params ...any) *ZodObject {
	schema := Object(shape, params...)
	schema.internals.Def.UnknownKeys = OBJECT_LOOSE_MODE
	return schema
}

// =============================================================================
// OBJECT UTILITY FUNCTIONS
// =============================================================================

// No private utility functions - use pkg packages directly

// =============================================================================
// OBJECT CREATION HELPER
// =============================================================================

// createZodObjectFromDef creates a ZodObject from definition following the unified pattern
func createZodObjectFromDef(def *ZodObjectDef, params ...any) *ZodObject {
	// Create internals with modern pattern
	internals := &ZodObjectInternals{
		ZodTypeInternals: engine.NewBaseZodTypeInternals(def.Type),
		Def:              def,
		Shape:            def.Shape,
		Bag:              make(map[string]any),
	}

	// Apply schema parameters following unified pattern
	for _, p := range params {
		if param, ok := p.(core.SchemaParams); ok {
			if param.Error != nil {
				errorMap := issues.CreateErrorMap(param.Error)
				if errorMap != nil {
					def.Error = errorMap
					internals.Error = errorMap
				}
			}
			if param.Description != "" {
				internals.Bag["description"] = param.Description
			}
			if param.Abort {
				internals.Bag["abort"] = true
			}
			if len(param.Path) > 0 {
				internals.Bag["path"] = param.Path
			}
			if len(param.Params) > 0 {
				internals.Bag["params"] = param.Params
			}
		}
	}

	// Set up simplified constructor for cloning
	internals.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any, any] {
		objectDef := &ZodObjectDef{
			ZodTypeDef:  *newDef,
			Type:        "object",
			Shape:       def.Shape,
			UnknownKeys: def.UnknownKeys,
		}
		newSchema := createZodObjectFromDef(objectDef)
		return any(newSchema).(core.ZodType[any, any])
	}

	// Set up parse function
	internals.Parse = func(payload *core.ParsePayload, ctx *core.ParseContext) *core.ParsePayload {
		schema := &ZodObject{internals: internals}
		result, err := schema.Parse(payload.GetValue(), ctx)
		if err != nil {
			var zodErr *issues.ZodError
			if errors.As(err, &zodErr) {
				for _, issue := range zodErr.Issues {
					// Convert ZodError to RawIssue using standardized converter
					rawIssue := issues.ConvertZodIssueToRaw(issue)
					rawIssue.Path = issue.Path
					payload.AddIssue(rawIssue)
				}
			}
			return payload
		}
		payload.SetValue(result)
		return payload
	}

	zodSchema := &ZodObject{internals: internals}

	// Use unified infrastructure for initialization
	engine.InitZodType(zodSchema, &def.ZodTypeDef)

	return zodSchema
}

// =============================================================================
// CORE INTERFACE METHODS
// =============================================================================

// GetInternals returns the internal state of the schema
func (z *ZodObject) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// Parse validates the input value using smart type inference
func (z *ZodObject) Parse(input any, ctx ...*core.ParseContext) (any, error) {
	parseCtx := (*core.ParseContext)(nil)
	if len(ctx) > 0 {
		parseCtx = ctx[0]
	}

	// 1. nil handling
	if input == nil {
		if !z.internals.Nilable {
			rawIssue := issues.CreateInvalidTypeIssue("object", input)
			finalIssue := issues.FinalizeIssue(rawIssue, nil, core.GetConfig())
			return nil, issues.NewZodError([]core.ZodIssue{finalIssue})
		}
		return (*map[string]any)(nil), nil
	}

	// 3. smart type inference: check if input is nil pointer using reflectx
	// --- Pointer smart inference -------------------------------------------------
	// If caller provides *map[string]any (or any pointer whose underlying value is
	// a map that can be converted to map[string]any), we validate the pointed
	// value but return the original pointer so that tests can assert pointer
	// identity preservation. This mirrors zod v4 behaviour.
	if reflectx.IsPointer(input) && !reflectx.IsNilPointer(input) {
		if deref, ok := reflectx.Deref(input); ok {
			// Try to convert to map[string]any
			if obj, ok := mapx.Extract(deref); ok {
				if stringMap, ok := obj.(map[string]any); ok {
					// Validate the dereferenced value via parseObjectCore
					// Use constructor instead of direct struct literal to respect private fields
					payload := core.NewParsePayloadWithPath(stringMap, make([]any, 0))
					resPayload := parseObjectCore(payload, z.internals, parseCtx)
					// Use getter method instead of direct field access
					if len(resPayload.GetIssues()) > 0 {
						return nil, issues.NewZodError(issues.ConvertRawIssuesToIssues(resPayload.GetIssues(), parseCtx))
					}

					// Also run checks if any
					if len(z.internals.Checks) > 0 {
						// Use getter method instead of direct field access
						checkPayload := core.NewParsePayload(resPayload.GetValue())
						engine.RunChecksOnValue(resPayload.GetValue(), z.internals.Checks, checkPayload, parseCtx)
						// Use getter method instead of direct field access
						if len(checkPayload.GetIssues()) > 0 {
							return nil, issues.NewZodError(issues.ConvertRawIssuesToIssues(checkPayload.GetIssues(), parseCtx))
						}
					}

					// Validation succeeded: return original pointer to preserve identity
					return input, nil
				}
			}
		}
	}

	// 3. smart type inference: check if input is nil pointer using reflectx
	if reflectx.IsNil(input) {
		if !z.internals.Nilable {
			rawIssue := issues.CreateInvalidTypeIssue("object", input)
			finalIssue := issues.FinalizeIssue(rawIssue, nil, core.GetConfig())
			return nil, issues.NewZodError([]core.ZodIssue{finalIssue})
		}
		return (*map[string]any)(nil), nil
	}

	// 4. strict type requirement: only map[string]any or *map[string]any accepted
	var objectMap map[string]any
	switch v := input.(type) {
	case map[string]any:
		objectMap = v
	case *map[string]any:
		if v != nil {
			objectMap = *v
		}
	default:
		rawIssue := issues.CreateInvalidTypeIssue("object", input)
		finalIssue := issues.FinalizeIssue(rawIssue, nil, core.GetConfig())
		return nil, issues.NewZodError([]core.ZodIssue{finalIssue})
	}

	// 6. use parseObjectCore for full object validation and field filtering
	// Use constructor instead of direct struct literal to respect private fields
	payload := core.NewParsePayloadWithPath(objectMap, make([]any, 0))

	result := parseObjectCore(payload, z.internals, parseCtx)
	// Use getter method instead of direct field access
	if len(result.GetIssues()) > 0 {
		return nil, issues.NewZodError(issues.ConvertRawIssuesToIssues(result.GetIssues(), parseCtx))
	}

	// 7. run additional checks
	if len(z.internals.Checks) > 0 {
		// Use constructor and getter methods instead of direct field access
		checksPayload := core.NewParsePayload(result.GetValue())
		engine.RunChecksOnValue(result.GetValue(), z.internals.Checks, checksPayload, parseCtx)
		if len(checksPayload.GetIssues()) > 0 {
			return nil, issues.NewZodError(issues.ConvertRawIssuesToIssues(checksPayload.GetIssues(), parseCtx))
		}
	}

	// 8. return validated map - smart type inference preserves original type structure
	return result.GetValue(), nil
}

// MustParse validates the input value and panics on failure
func (z *ZodObject) MustParse(input any, ctx ...*core.ParseContext) any {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// Transform provides type-safe object transformation, supporting smart dereferencing
// Automatically handles input of map[string]any, struct, *struct, and nil pointer
func (z *ZodObject) Transform(fn func(map[string]any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	return z.TransformAny(func(input any, ctx *core.RefinementContext) (any, error) {
		if input == nil || reflectx.IsNil(input) {
			return nil, ErrTransformNilObject
		}

		// Convert to map[string]any for consistent processing
		objMap := mapx.FromAny(input)
		if objMap == nil {
			return nil, ErrCannotConvertObject
		}

		return fn(objMap, ctx)
	})
}

// TransformAny flexible version of Transform - same implementation as Transform, providing backward compatibility
// Implements ZodType[any, any] interface: TransformAny(fn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any]
func (z *ZodObject) TransformAny(fn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	transform := Transform[any, any](fn)

	return &ZodPipe[any, any]{
		in:  any(z).(core.ZodType[any, any]),
		out: any(transform).(core.ZodType[any, any]),
		def: core.ZodTypeDef{Type: "pipe"},
	}
}

// Pipe creates a new schema by piping the output of this schema to another
func (z *ZodObject) Pipe(target core.ZodType[any, any]) core.ZodType[any, any] {
	return &ZodPipe[any, any]{
		in:  any(z).(core.ZodType[any, any]),
		out: target,
		def: core.ZodTypeDef{Type: "pipe"},
	}
}

// =============================================================================
// OBJECT OPERATIONS
// =============================================================================

// Pick creates a new object schema with only the specified fields
func (z *ZodObject) Pick(keys []string) *ZodObject {
	newShape := make(core.ObjectSchema)
	for _, key := range keys {
		if schema, exists := z.internals.Shape[key]; exists {
			newShape[key] = schema
		}
	}

	newDef := &ZodObjectDef{
		ZodTypeDef:  z.internals.Def.ZodTypeDef,
		Type:        z.internals.Def.Type,
		Shape:       newShape,
		UnknownKeys: z.internals.Def.UnknownKeys,
	}

	newSchema := createZodObjectFromDef(newDef)

	// Copy runtime configuration from original schema
	for key, value := range z.internals.Bag {
		newSchema.internals.Bag[key] = value
	}

	return newSchema
}

// Omit creates a new object schema without the specified fields
func (z *ZodObject) Omit(keys []string) *ZodObject {
	omitSet := make(map[string]struct{})
	for _, key := range keys {
		omitSet[key] = struct{}{}
	}

	newShape := make(core.ObjectSchema)
	for key, schema := range z.internals.Shape {
		if _, shouldOmit := omitSet[key]; !shouldOmit {
			newShape[key] = schema
		}
	}

	newDef := &ZodObjectDef{
		ZodTypeDef:  z.internals.Def.ZodTypeDef,
		Type:        z.internals.Def.Type,
		Shape:       newShape,
		UnknownKeys: z.internals.Def.UnknownKeys,
	}

	newSchema := createZodObjectFromDef(newDef)

	// Copy runtime configuration from original schema
	for key, value := range z.internals.Bag {
		newSchema.internals.Bag[key] = value
	}

	return newSchema
}

// Extend creates a new object schema with additional fields
func (z *ZodObject) Extend(extension core.ObjectSchema) *ZodObject {
	newShape := make(core.ObjectSchema)

	// Copy existing fields
	for key, schema := range z.internals.Shape {
		newShape[key] = schema
	}

	// Add extension fields (overrides existing ones)
	for key, schema := range extension {
		newShape[key] = schema
	}

	newDef := &ZodObjectDef{
		ZodTypeDef:  z.internals.Def.ZodTypeDef,
		Type:        z.internals.Def.Type,
		Shape:       newShape,
		UnknownKeys: z.internals.Def.UnknownKeys,
	}

	newSchema := createZodObjectFromDef(newDef)

	// Copy runtime configuration from original schema
	for key, value := range z.internals.Bag {
		newSchema.internals.Bag[key] = value
	}

	return newSchema
}

// Partial makes all fields optional
func (z *ZodObject) Partial() *ZodObject {
	newShape := make(core.ObjectSchema)
	for key, schema := range z.internals.Shape {
		newShape[key] = Optional(schema)
	}

	newDef := &ZodObjectDef{
		ZodTypeDef:  z.internals.Def.ZodTypeDef,
		Type:        z.internals.Def.Type,
		Shape:       newShape,
		UnknownKeys: z.internals.Def.UnknownKeys,
	}

	newSchema := createZodObjectFromDef(newDef)

	// Copy runtime configuration from original schema
	for key, value := range z.internals.Bag {
		newSchema.internals.Bag[key] = value
	}

	return newSchema
}

// Required makes all fields required
func (z *ZodObject) Required(fields ...[]string) *ZodObject {
	newShape := make(core.ObjectSchema)

	// If specific fields provided, only make those required
	var targetFields map[string]struct{}
	if len(fields) > 0 && len(fields[0]) > 0 {
		targetFields = make(map[string]struct{})
		for _, field := range fields[0] {
			targetFields[field] = struct{}{}
		}
	}

	for key, schema := range z.internals.Shape {
		if targetFields != nil {
			// Only make specified fields required
			if _, shouldRequire := targetFields[key]; shouldRequire {
				// Remove optional wrapper if present
				if optionalType, ok := schema.(*ZodOptional[core.ZodType[any, any]]); ok {
					newShape[key] = optionalType.Unwrap()
				} else {
					newShape[key] = schema
				}
			} else {
				newShape[key] = schema
			}
		} else {
			// Make all fields required
			if optionalType, ok := schema.(*ZodOptional[core.ZodType[any, any]]); ok {
				newShape[key] = optionalType.Unwrap()
			} else {
				newShape[key] = schema
			}
		}
	}

	newDef := &ZodObjectDef{
		ZodTypeDef:  z.internals.Def.ZodTypeDef,
		Type:        z.internals.Def.Type,
		Shape:       newShape,
		UnknownKeys: z.internals.Def.UnknownKeys,
	}

	newSchema := createZodObjectFromDef(newDef)

	// Copy runtime configuration from original schema
	for key, value := range z.internals.Bag {
		newSchema.internals.Bag[key] = value
	}

	return newSchema
}

// Merge combines this object schema with another
func (z *ZodObject) Merge(other *ZodObject) *ZodObject {
	newShape := make(core.ObjectSchema)

	// Copy fields from this schema
	for key, schema := range z.internals.Shape {
		newShape[key] = schema
	}

	// Override with fields from other schema
	for key, schema := range other.internals.Shape {
		newShape[key] = schema
	}

	newDef := &ZodObjectDef{
		ZodTypeDef:  z.internals.Def.ZodTypeDef,
		Type:        z.internals.Def.Type,
		Shape:       newShape,
		UnknownKeys: other.internals.Def.UnknownKeys,
	}

	newSchema := createZodObjectFromDef(newDef)

	// Copy runtime configuration from both schemas (other takes precedence)
	for key, value := range z.internals.Bag {
		newSchema.internals.Bag[key] = value
	}
	for key, value := range other.internals.Bag {
		newSchema.internals.Bag[key] = value
	}

	return newSchema
}

// Catchall sets a schema to validate unknown keys
func (z *ZodObject) Catchall(catchallSchema core.ZodType[any, any]) *ZodObject {
	newDef := &ZodObjectDef{
		ZodTypeDef:  z.internals.Def.ZodTypeDef,
		Type:        z.internals.Def.Type,
		Shape:       z.internals.Shape,
		Catchall:    catchallSchema,
		UnknownKeys: OBJECT_LOOSE_MODE, // Catchall implies loose mode
	}

	newSchema := createZodObjectFromDef(newDef)

	// Copy runtime configuration from original schema
	for key, value := range z.internals.Bag {
		newSchema.internals.Bag[key] = value
	}

	return newSchema
}

// Passthrough allows unknown keys to pass through (alias for loose mode)
func (z *ZodObject) Passthrough() *ZodObject {
	newDef := &ZodObjectDef{
		ZodTypeDef:  z.internals.Def.ZodTypeDef,
		Type:        z.internals.Def.Type,
		Shape:       z.internals.Shape,
		Catchall:    Unknown(),
		UnknownKeys: OBJECT_LOOSE_MODE,
	}

	newSchema := createZodObjectFromDef(newDef)

	// Copy runtime configuration from original schema
	for key, value := range z.internals.Bag {
		newSchema.internals.Bag[key] = value
	}

	return newSchema
}

// Strict sets strict mode (rejects unknown keys)
func (z *ZodObject) Strict() *ZodObject {
	newDef := &ZodObjectDef{
		ZodTypeDef:  z.internals.Def.ZodTypeDef,
		Type:        z.internals.Def.Type,
		Shape:       z.internals.Shape,
		UnknownKeys: OBJECT_STRICT_MODE,
	}

	newSchema := createZodObjectFromDef(newDef)

	// Copy runtime configuration from original schema
	for key, value := range z.internals.Bag {
		newSchema.internals.Bag[key] = value
	}

	return newSchema
}

// Strip sets strip mode (removes unknown keys) - default behavior
func (z *ZodObject) Strip() *ZodObject {
	newDef := &ZodObjectDef{
		ZodTypeDef:  z.internals.Def.ZodTypeDef,
		Type:        z.internals.Def.Type,
		Shape:       z.internals.Shape,
		UnknownKeys: OBJECT_STRIP_MODE,
	}

	newSchema := createZodObjectFromDef(newDef)

	// Copy runtime configuration from original schema
	for key, value := range z.internals.Bag {
		newSchema.internals.Bag[key] = value
	}

	return newSchema
}

// Keyof creates an enum schema from the object keys
func (z *ZodObject) Keyof() core.ZodType[any, any] {
	keys := make([]string, 0, len(z.internals.Shape))
	for key := range z.internals.Shape {
		keys = append(keys, key)
	}

	if len(keys) == 0 {
		return Never() // Empty object has no keys
	}

	return EnumSlice(keys)
}

// =============================================================================
// WRAPPER METHODS
// =============================================================================

//////////////////////////////////////////
//////////   Utility Methods     //////////
//////////////////////////////////////////

// Refine adds a type-safe validation function to the object
func (z *ZodObject) Refine(fn func(map[string]any) bool, params ...any) *ZodObject {
	result := z.RefineAny(func(v any) bool {
		if v == nil || reflectx.IsNil(v) {
			return true // let the upper logic decide
		}
		objectMap := mapx.FromAny(v)
		if objectMap == nil {
			return false
		}
		return fn(objectMap)
	}, params...)
	return result.(*ZodObject)
}

// RefineAny adds flexible custom validation logic to the object schema
func (z *ZodObject) RefineAny(fn func(any) bool, params ...any) core.ZodType[any, any] {
	check := checks.NewCustom[any](fn, params...)
	return engine.AddCheck(any(z).(core.ZodType[any, any]), check)
}

// Optional makes the object optional
func (z *ZodObject) Optional() core.ZodType[any, any] {
	return any(Optional(any(z).(core.ZodType[any, any]))).(core.ZodType[any, any])
}

// Nilable makes the object nilable
func (z *ZodObject) Nilable() core.ZodType[any, any] {
	cloned := engine.Clone(any(z).(core.ZodType[any, any]), func(def *core.ZodTypeDef) {
	}).(*ZodObject)
	cloned.internals.SetNilable()
	return any(cloned).(core.ZodType[any, any])
}

// Check adds modern validation using direct payload access
func (z *ZodObject) Check(fn core.CheckFn) *ZodObject {
	check := checks.NewCustom[map[string]any](fn, core.SchemaParams{})
	result := engine.AddCheck(z, check)
	return result.(*ZodObject)
}

////////////////////////////
////   OBJECT DEFAULT WRAPPER ////
////////////////////////////

type ZodObjectDefault struct {
	*ZodDefault[*ZodObject] // embed pointer to allow method promotion
}

func (s ZodObjectDefault) Parse(input any, ctx ...*core.ParseContext) (any, error) {
	return s.ZodDefault.Parse(input, ctx...)
}

////////////////////////////
////   DEFAULT method   ////
////////////////////////////

// Default sets the default value for the object
func (z *ZodObject) Default(value map[string]any) ZodObjectDefault {
	return ZodObjectDefault{
		&ZodDefault[*ZodObject]{
			innerType:    z,
			defaultValue: value,
			isFunction:   false,
		},
	}
}

// DefaultFunc sets the default function for the object
func (z *ZodObject) DefaultFunc(fn func() map[string]any) ZodObjectDefault {
	genericFn := func() any { return fn() }
	return ZodObjectDefault{
		&ZodDefault[*ZodObject]{
			innerType:   z,
			defaultFunc: genericFn,
			isFunction:  true,
		},
	}
}

////////////////////////////
////   OBJECTDEFAULT CHAIN METHODS ////
////////////////////////////

// Pick creates a new object schema with only the specified fields, return ZodObjectDefault
func (s ZodObjectDefault) Pick(keys []string) ZodObjectDefault {
	newInner := s.innerType.Pick(keys)

	// filter the default value, only keep the fields of Pick
	var newDefaultValue any
	var newDefaultFunc func() any

	if s.isFunction && s.defaultFunc != nil {
		// if it is a function default value, create a new function to filter the result
		newDefaultFunc = func() any {
			originalValue := s.defaultFunc()
			if originalMap, ok := originalValue.(map[string]any); ok {
				filteredMap := make(map[string]any)
				for _, key := range keys {
					if value, exists := originalMap[key]; exists {
						filteredMap[key] = value
					}
				}
				return filteredMap
			}
			return originalValue
		}
	} else {
		// if it is a static default value, filter directly
		if originalMap, ok := s.defaultValue.(map[string]any); ok {
			filteredMap := make(map[string]any)
			for _, key := range keys {
				if value, exists := originalMap[key]; exists {
					filteredMap[key] = value
				}
			}
			newDefaultValue = filteredMap
		} else {
			newDefaultValue = s.defaultValue
		}
	}

	return ZodObjectDefault{
		&ZodDefault[*ZodObject]{
			innerType:    newInner,
			defaultValue: newDefaultValue,
			defaultFunc:  newDefaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

// Omit creates a new object schema with the specified fields omitted, return ZodObjectDefault
func (s ZodObjectDefault) Omit(keys []string) ZodObjectDefault {
	newInner := s.innerType.Omit(keys)
	return ZodObjectDefault{
		&ZodDefault[*ZodObject]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

// Extend extends the object schema, return ZodObjectDefault
func (s ZodObjectDefault) Extend(extension core.ObjectSchema) ZodObjectDefault {
	newInner := s.innerType.Extend(extension)
	return ZodObjectDefault{
		&ZodDefault[*ZodObject]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

// Partial makes all fields optional, return ZodObjectDefault
func (s ZodObjectDefault) Partial() ZodObjectDefault {
	newInner := s.innerType.Partial()
	return ZodObjectDefault{
		&ZodDefault[*ZodObject]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

// Required makes the specified fields required, return ZodObjectDefault
func (s ZodObjectDefault) Required(fields ...[]string) ZodObjectDefault {
	newInner := s.innerType.Required(fields...)
	return ZodObjectDefault{
		&ZodDefault[*ZodObject]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

// Merge merges another object schema, return ZodObjectDefault
func (s ZodObjectDefault) Merge(other *ZodObject) ZodObjectDefault {
	newInner := s.innerType.Merge(other)
	return ZodObjectDefault{
		&ZodDefault[*ZodObject]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

// Refine adds a flexible validation function to the object schema, return ZodObjectDefault
func (s ZodObjectDefault) Refine(fn func(map[string]any) bool, params ...any) ZodObjectDefault {
	newInner := s.innerType.Refine(fn, params...)
	return ZodObjectDefault{
		&ZodDefault[*ZodObject]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

// Transform adds data transformation, return a generic ZodType
func (s ZodObjectDefault) Transform(fn func(map[string]any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	// use the TransformAny method of the embedded ZodDefault
	return s.TransformAny(func(input any, ctx *core.RefinementContext) (any, error) {
		// smartly handle object value
		if input == nil || reflectx.IsNil(input) {
			return nil, ErrTransformNilObject
		}

		// convert to map[string]any
		objMap := mapx.FromAny(input)
		if objMap == nil {
			return nil, ErrCannotConvertObject
		}

		return fn(objMap, ctx)
	})
}

// Optional modifier - correctly wrap Default wrapper
func (s ZodObjectDefault) Optional() core.ZodType[any, any] {
	// wrap the current ZodObjectDefault instance, keep Default logic
	return Optional(any(s).(core.ZodType[any, any]))
}

// Nilable modifier - correctly wrap Default wrapper
func (s ZodObjectDefault) Nilable() core.ZodType[any, any] {
	// wrap the current ZodObjectDefault instance, keep Default logic
	return Nilable(any(s).(core.ZodType[any, any]))
}

////////////////////////////
////   OBJECT PREFAULT WRAPPER ////
////////////////////////////

// ZodObjectPrefault is the Prefault wrapper for object type
// provide perfect type safety and chain call support
type ZodObjectPrefault struct {
	*ZodPrefault[*ZodObject] // embed pointer to allow method promotion
}

////////////////////////////
////   PREFAULT method   ////
////////////////////////////

// Prefault adds a prefault value to the object schema, return ZodObjectPrefault
func (z *ZodObject) Prefault(value map[string]any) ZodObjectPrefault {
	return ZodObjectPrefault{
		&ZodPrefault[*ZodObject]{
			innerType:     z,
			prefaultValue: value,
			isFunction:    false,
		},
	}
}

// PrefaultFunc adds a prefault function to the object schema, return ZodObjectPrefault
func (z *ZodObject) PrefaultFunc(fn func() map[string]any) ZodObjectPrefault {
	genericFn := func() any { return fn() }
	return ZodObjectPrefault{
		&ZodPrefault[*ZodObject]{
			innerType:    z,
			prefaultFunc: genericFn,
			isFunction:   true,
		},
	}
}

////////////////////////////
////   OBJECTPREFAULT chain methods ////
////////////////////////////

// Pick selects the specified fields, return ZodObjectPrefault
func (o ZodObjectPrefault) Pick(keys []string) ZodObjectPrefault {
	newInner := o.innerType.Pick(keys)

	// filter prefaultValue, only keep the fields of Pick
	var newPrefaultValue any
	var newPrefaultFunc func() any

	if o.isFunction && o.prefaultFunc != nil {
		// if it is a function default value, create a new function to filter the result
		newPrefaultFunc = func() any {
			originalValue := o.prefaultFunc()
			if originalMap, ok := originalValue.(map[string]any); ok {
				filteredMap := make(map[string]any)
				for _, key := range keys {
					if value, exists := originalMap[key]; exists {
						filteredMap[key] = value
					}
				}
				return filteredMap
			}
			return originalValue
		}
	} else {
		// if it is a static default value, filter directly
		if originalMap, ok := o.prefaultValue.(map[string]any); ok {
			filteredMap := make(map[string]any)
			for _, key := range keys {
				if value, exists := originalMap[key]; exists {
					filteredMap[key] = value
				}
			}
			newPrefaultValue = filteredMap
		} else {
			newPrefaultValue = o.prefaultValue
		}
	}

	return ZodObjectPrefault{
		&ZodPrefault[*ZodObject]{
			innerType:     newInner,
			prefaultValue: newPrefaultValue,
			prefaultFunc:  newPrefaultFunc,
			isFunction:    o.isFunction,
		},
	}
}

// Omit excludes the specified fields, return ZodObjectPrefault
func (o ZodObjectPrefault) Omit(keys []string) ZodObjectPrefault {
	newInner := o.innerType.Omit(keys)

	// create the omit set
	omitSet := make(map[string]struct{})
	for _, key := range keys {
		omitSet[key] = struct{}{}
	}

	// filter prefaultValue, exclude the fields of Omit
	var newPrefaultValue any
	var newPrefaultFunc func() any

	if o.isFunction && o.prefaultFunc != nil {
		// if it is a function default value, create a new function to filter the result
		newPrefaultFunc = func() any {
			originalValue := o.prefaultFunc()
			if originalMap, ok := originalValue.(map[string]any); ok {
				filteredMap := make(map[string]any)
				for key, value := range originalMap {
					if _, shouldOmit := omitSet[key]; !shouldOmit {
						filteredMap[key] = value
					}
				}
				return filteredMap
			}
			return originalValue
		}
	} else {
		// if it is a static default value, filter directly
		if originalMap, ok := o.prefaultValue.(map[string]any); ok {
			filteredMap := make(map[string]any)
			for key, value := range originalMap {
				if _, shouldOmit := omitSet[key]; !shouldOmit {
					filteredMap[key] = value
				}
			}
			newPrefaultValue = filteredMap
		} else {
			newPrefaultValue = o.prefaultValue
		}
	}

	return ZodObjectPrefault{
		&ZodPrefault[*ZodObject]{
			innerType:     newInner,
			prefaultValue: newPrefaultValue,
			prefaultFunc:  newPrefaultFunc,
			isFunction:    o.isFunction,
		},
	}
}

// Partial makes all fields optional, return ZodObjectPrefault
func (o ZodObjectPrefault) Partial() ZodObjectPrefault {
	newInner := o.innerType.Partial()
	return ZodObjectPrefault{
		&ZodPrefault[*ZodObject]{
			innerType:     newInner,
			prefaultValue: o.prefaultValue,
			prefaultFunc:  o.prefaultFunc,
			isFunction:    o.isFunction,
		},
	}
}

// Refine adds a flexible validation function to the object schema, return ZodObjectPrefault
func (o ZodObjectPrefault) Refine(fn func(map[string]any) bool, params ...any) ZodObjectPrefault {
	newInner := o.innerType.Refine(fn, params...)
	return ZodObjectPrefault{
		&ZodPrefault[*ZodObject]{
			innerType:     newInner,
			prefaultValue: o.prefaultValue,
			prefaultFunc:  o.prefaultFunc,
			isFunction:    o.isFunction,
		},
	}
}

// Transform adds data transformation, return a generic ZodType
func (o ZodObjectPrefault) Transform(fn func(map[string]any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	// use the TransformAny method of the embedded ZodPrefault
	return o.TransformAny(func(input any, ctx *core.RefinementContext) (any, error) {
		// smartly handle object value
		if input == nil || reflectx.IsNil(input) {
			return nil, ErrTransformNilObject
		}

		// convert to map[string]any
		objMap := mapx.FromAny(input)
		if objMap == nil {
			return nil, ErrCannotConvertObject
		}

		return fn(objMap, ctx)
	})
}

// Optional modifier - correctly wrap Prefault wrapper
func (o ZodObjectPrefault) Optional() core.ZodType[any, any] {
	// wrap the current ZodObjectPrefault instance, keep Prefault logic
	return Optional(any(o).(core.ZodType[any, any]))
}

// Nilable modifier - correctly wrap Prefault wrapper
func (o ZodObjectPrefault) Nilable() core.ZodType[any, any] {
	// wrap the current ZodObjectPrefault instance, keep Prefault logic
	return Nilable(any(o).(core.ZodType[any, any]))
}

// =============================================================================
// INTERNAL CHECK MANAGEMENT
// =============================================================================

// CloneFrom implements Cloneable interface for type-specific state copying
func (z *ZodObject) CloneFrom(source any) {
	if src, ok := source.(*ZodObject); ok {
		// Copy type-specific fields from Bag
		if src.internals.Bag != nil {
			if z.internals.Bag == nil {
				z.internals.Bag = make(map[string]any)
			}
			for k, v := range src.internals.Bag {
				z.internals.Bag[k] = v
			}
		}

		// Copy shape and other object-specific fields
		if src.internals.Shape != nil {
			z.internals.Shape = make(core.ObjectSchema)
			for k, v := range src.internals.Shape {
				z.internals.Shape[k] = v
			}
		}

		// Copy object definition fields
		if src.internals.Def != nil {
			z.internals.Def.UnknownKeys = src.internals.Def.UnknownKeys
		}
	}
}

// =============================================================================
// CORE PARSING LOGIC
// =============================================================================

// parseObjectCore handles object-specific parsing logic
func parseObjectCore(payload *core.ParsePayload, internals *ZodObjectInternals, ctx *core.ParseContext) *core.ParsePayload {
	// 1. type check - only accept map[string]any (no coercion for object type)
	objectData, ok := payload.GetValue().(map[string]any)
	if !ok {
		issue := issues.CreateInvalidTypeIssue("object", payload.GetValue())
		issue.Inst = internals
		payload.AddIssue(issue)
		return payload
	}

	// 2. field validation - validate each field according to Shape definition
	result := make(map[string]any)
	processedKeys := make(map[string]struct{})

	// validate defined fields
	for fieldName, fieldSchema := range internals.Shape {
		// Use getter method instead of direct field access
		fieldPath := make([]any, 0, len(payload.GetPath())+1)
		fieldPath = append(fieldPath, payload.GetPath()...)
		fieldPath = append(fieldPath, fieldName)
		fieldValue, exists := objectData[fieldName]

		if !exists {
			// check if field is optional
			if engine.IsOptionalField(fieldSchema) {
				// for optional fields, try parsing nil to get default values
				// this handles Default types that should provide default values

				// use the schema's Parse method directly, not GetInternals().Parse
				// this ensures wrapper types like Default work correctly
				fieldResultValue, fieldErr := fieldSchema.Parse(nil, ctx)
				if fieldErr == nil && fieldResultValue != nil {
					// default value provided, include it in result
					result[fieldName] = fieldResultValue
				}
				// if no default value or parsing failed, skip the field (Optional behavior)
				continue
			} else {
				// missing required field
				issue := issues.CreateMissingKeyIssue(fieldName, func(iss *core.ZodRawIssue) {
					iss.Inst = internals
				})
				payload.AddIssueWithPath(issue, fieldPath)
				continue
			}
		}

		// no coercion for object type (non-primitive)
		actualFieldSchema := fieldSchema

		// fix nil pointer error: for wrapper types, use engine.Parse method directly
		if actualFieldSchema == nil {
			// if schema is nil, create an error
			issue := issues.CreateInvalidTypeIssue("unknown", fieldValue)
			issue.Inst = internals
			payload.AddIssueWithPath(issue, fieldPath)
			continue
		}

		// use the schema's Parse method directly, this ensures wrapper types (like Prefault, Default) work correctly
		fieldResultValue, fieldErr := actualFieldSchema.Parse(fieldValue, ctx)
		if fieldErr != nil {
			// when parsing a field, convert the error to payload format
			var zodErr *issues.ZodError
			if errors.As(fieldErr, &zodErr) {
				// Convert ZodIssue to ZodRawIssue and update the error path
				for _, issue := range zodErr.Issues {
					// Convert ZodError to RawIssue using standardized converter
					rawIssue := issues.ConvertZodIssueToRaw(issue)
					rawIssue.Inst = internals

					// Copy additional properties from the original issue
					if issue.Expected != "" {
						rawIssue.Properties["expected"] = issue.Expected
					}
					if issue.Received != "" {
						rawIssue.Properties["received"] = issue.Received
					}
					if issue.Format != "" {
						rawIssue.Properties["format"] = issue.Format
					}

					payload.AddIssueWithPath(rawIssue, append(fieldPath, issue.Path...))
				}
			} else {
				// if not ZodError, create a generic error
				issue := issues.CreateInvalidTypeIssue("unknown", fieldValue)
				issue.Inst = internals
				payload.AddIssueWithPath(issue, fieldPath)
			}
		} else {
			result[fieldName] = fieldResultValue
		}

		processedKeys[fieldName] = struct{}{}
	}

	// 3. handle unknown keys according to UnknownKeys mode
	for key, value := range objectData {
		if _, processed := processedKeys[key]; !processed {
			// Unknown key found
			switch internals.Def.UnknownKeys {
			case OBJECT_STRICT_MODE:
				// strict mode: error on unknown keys
				issue := issues.CreateUnrecognizedKeysIssue([]string{key}, payload.GetValue())
				issue.Inst = internals
				payload.AddIssueWithPath(issue, append(payload.GetPath(), key))

			case OBJECT_LOOSE_MODE:
				// loose mode: allow unknown keys to pass through
				if internals.Def.Catchall != nil {
					// Validate unknown key value with catchall schema using full Parse to preserve wrapper behaviour
					parsedVal, err := internals.Def.Catchall.Parse(value, ctx)
					if err != nil {
						// Convert ZodError to raw issues with correct path
						var zodErr *issues.ZodError
						if errors.As(err, &zodErr) {
							for _, issue := range zodErr.Issues {
								// Convert ZodError to RawIssue using standardized converter
								raw := issues.ConvertZodIssueToRaw(issue)
								raw.Inst = internals

								// Prepend object key to issue path for catchall field errors
								path := append(payload.GetPath(), append([]any{key}, issue.Path...)...)
								payload.AddIssueWithPath(raw, path)
							}
						} else {
							// Fallback generic invalid type issue
							raw := issues.CreateInvalidTypeIssue("unknown", value)
							raw.Inst = internals
							payload.AddIssueWithPath(raw, append(payload.GetPath(), key))
						}
					} else {
						// Successful validation – adopt transformed value
						result[key] = parsedVal
					}
				} else {
					// no catchall, pass through as-is
					result[key] = value
				}

			case OBJECT_STRIP_MODE:
				// strip mode: ignore unknown keys (default behavior)
				// do nothing - key is stripped
			}
		}
	}

	// 4. run custom checks if object validation succeeded
	if len(payload.GetIssues()) == 0 {
		if customChecks, exists := internals.Bag["customChecks"].([]core.ZodCheck); exists {
			engine.RunChecksOnValue(result, customChecks, payload, ctx)
		}
	}

	// update payload with validated object
	payload.SetValue(result)
	return payload
}

// Unwrap returns the inner type (for basic types, return self)
func (z *ZodObject) Unwrap() core.ZodType[any, any] {
	return any(z).(core.ZodType[any, any])
}

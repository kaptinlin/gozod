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

// Error definitions for record transformations
var (
	ErrTransformNilRecord = errors.New("cannot transform nil record value")
	ErrExpectedRecord     = errors.New("expected record type")
)

//////////////////////////
////   RECORD TYPES  ////
//////////////////////////

// ZodRecordDef defines the configuration for record validation
type ZodRecordDef struct {
	core.ZodTypeDef
	Type      core.ZodTypeCode       // "record"
	KeyType   core.ZodType[any, any] // Schema for validating keys
	ValueType core.ZodType[any, any] // Schema for validating values
	Checks    []core.ZodCheck        // Record-specific validation checks
}

// ZodRecordInternals contains record validator internal state
type ZodRecordInternals struct {
	core.ZodTypeInternals
	Def       *ZodRecordDef              // Schema definition
	KeyType   core.ZodType[any, any]     // Key validation schema
	ValueType core.ZodType[any, any]     // Value validation schema
	Isst      issues.ZodIssueInvalidType // Invalid type issue template
	Bag       map[string]any             // Additional metadata
}

// ZodRecord represents a record validation schema for key-value pairs
type ZodRecord struct {
	internals *ZodRecordInternals
}

// GetInternals returns the internal state of the schema
func (z *ZodRecord) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// GetZod returns the record-specific internals for framework usage
func (z *ZodRecord) GetZod() *ZodRecordInternals {
	return z.internals
}

// CloneFrom implements Cloneable interface for type-specific state copying
func (z *ZodRecord) CloneFrom(source any) {
	if src, ok := source.(*ZodRecord); ok {
		// Copy type-specific fields from Bag
		if src.internals.Bag != nil {
			if z.internals.Bag == nil {
				z.internals.Bag = make(map[string]any)
			}
			for k, v := range src.internals.Bag {
				z.internals.Bag[k] = v
			}
		}

		// Copy key and value schemas
		z.internals.KeyType = src.internals.KeyType
		z.internals.ValueType = src.internals.ValueType
	}
}

// Parse validates input using unified parsing template
func (z *ZodRecord) Parse(input any, ctx ...*core.ParseContext) (any, error) {
	parseCtx := (*core.ParseContext)(nil)
	if len(ctx) > 0 {
		parseCtx = ctx[0]
	}

	// Use engine.ParseType unified template
	return engine.ParseType[map[any]any](
		input,
		&z.internals.ZodTypeInternals,
		core.ZodTypeRecord,
		mapx.ExtractRecord,
		func(v any) (*map[any]any, bool) {
			// Handle pointer types using reflectx
			if reflectx.IsPointer(v) {
				if deref, ok := reflectx.Deref(v); ok {
					if record, ok := mapx.ExtractRecord(deref); ok {
						return &record, true
					}
				}
			}
			return nil, false
		},
		func(value map[any]any, checks []core.ZodCheck, ctx *core.ParseContext) error {
			// Validate record and run checks
			_, err := z.validateRecordAndRunChecks(value, ctx)
			return err
		},
		parseCtx,
	)
}

// MustParse validates the input value and panics on failure
func (z *ZodRecord) MustParse(input any, ctx ...*core.ParseContext) any {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// validateRecordAndRunChecks validates record keys/values and runs checks
func (z *ZodRecord) validateRecordAndRunChecks(recordMap map[any]any, ctx *core.ParseContext) (map[any]any, error) {
	// Run basic checks first
	if len(z.internals.Checks) > 0 {
		payload := &core.ParsePayload{
			Value:  recordMap,
			Issues: make([]core.ZodRawIssue, 0),
			Path:   make([]any, 0),
		}
		engine.RunChecksOnValue(recordMap, z.internals.Checks, payload, ctx)
		if len(payload.Issues) > 0 {
			return nil, issues.NewZodError(issues.ConvertRawIssuesToIssues(payload.Issues, ctx))
		}
	}

	// Validate keys and values
	result := make(map[any]any)
	for key, value := range recordMap {
		// Validate key - use engine.TryApplyCoercion
		var validatedKey any
		var err error

		validatedKey, err = z.internals.KeyType.Parse(key, ctx)

		if err != nil {
			// Use issues.CreateInvalidKeyIssue with correct parameters
			keyStr, _ := reflectx.ToString(key)
			rawIssue := issues.CreateInvalidKeyIssue(
				keyStr,
				"record",
				recordMap,
			)
			finalIssue := issues.FinalizeIssue(rawIssue, ctx, core.GetConfig())
			return nil, issues.NewZodError([]core.ZodIssue{finalIssue})
		}

		// Validate value - use engine.TryApplyCoercion
		var validatedValue any

		validatedValue, err = z.internals.ValueType.Parse(value, ctx)

		if err != nil {
			rawIssue := issues.NewRawIssue("invalid_value", recordMap)
			rawIssue.Message = err.Error()
			rawIssue.Path = []any{key}
			finalIssue := issues.FinalizeIssue(rawIssue, ctx, core.GetConfig())
			return nil, issues.NewZodError([]core.ZodIssue{finalIssue})
		}

		result[validatedKey] = validatedValue
	}
	return result, nil
}

//////////////////////////
// TRANSFORM METHODS
//////////////////////////

// Transform provides type-safe record transformation
func (z *ZodRecord) Transform(fn func(map[any]any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	wrappedFn := func(input any, ctx *core.RefinementContext) (any, error) {
		recordVal, isNil, err := extractRecordPointerValue(input)
		if err != nil {
			return nil, err
		}
		if isNil {
			return nil, ErrTransformNilRecord
		}
		return fn(recordVal, ctx)
	}
	return z.TransformAny(wrappedFn)
}

// TransformAny provides flexible transformation that accepts any input type
func (z *ZodRecord) TransformAny(fn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	// Create a generic Transform schema wrapping provided function
	transform := Transform[any, any](fn)

	// Chain current record schema with transform via ZodPipe so that:
	// 1) input is validated by the record schema (z)
	// 2) output of validation is passed into transform
	return &ZodPipe[any, any]{
		in:  any(z).(core.ZodType[any, any]),
		out: any(transform).(core.ZodType[any, any]),
		def: core.ZodTypeDef{Type: "pipe"},
	}
}

// Pipe chains the current record schema with another schema, mirroring zod.pipe behaviour.
// The resulting schema first validates using the record schema, and then feeds the result
// into the provided `out` schema.
func (z *ZodRecord) Pipe(out core.ZodType[any, any]) core.ZodType[any, any] {
	return &ZodPipe[any, any]{
		in:  any(z).(core.ZodType[any, any]),
		out: out,
		def: core.ZodTypeDef{Type: "pipe"},
	}
}

///////////////////////////
////   RECORD WRAPPERS ////
///////////////////////////

// Optional makes the record schema optional (equivalent to TypeScript z.record().optional())
func (z *ZodRecord) Optional() core.ZodType[any, any] {
	return Optional(any(z).(core.ZodType[any, any]))
}

// Nilable makes the record nilable
func (z *ZodRecord) Nilable() core.ZodType[any, any] {
	cloned := engine.Clone(any(z).(core.ZodType[any, any]), func(def *core.ZodTypeDef) {
		// setNilable only changes nil handling, not other logic
	})
	internals := cloned.GetInternals()
	internals.SetNilable()
	return cloned
}

// Nullish makes the record both optional and nilable (equivalent to TypeScript z.record().nullish())
func (z *ZodRecord) Nullish() core.ZodType[any, any] {
	return Nullish(any(z).(core.ZodType[any, any]))
}

////////////////////////////
////   INTERNAL METHODS ////
////////////////////////////

// Refine provides type-safe record validation
func (z *ZodRecord) Refine(fn func(map[any]any) bool, params ...any) *ZodRecord {
	result := z.RefineAny(func(v any) bool {
		// Use mapx.ExtractRecord for robust extraction
		recordMap, ok := mapx.ExtractRecord(v)
		if !ok {
			return false
		}
		return fn(recordMap)
	}, params...)
	return result.(*ZodRecord)
}

// RefineAny flexible version of validation that accepts any type
func (z *ZodRecord) RefineAny(fn func(any) bool, params ...any) core.ZodType[any, any] {
	check := checks.NewCustom[any](fn, params...)
	return engine.AddCheck(any(z).(core.ZodType[any, any]), check)
}

// Unwrap returns the inner type (for basic types, returns self)
func (z *ZodRecord) Unwrap() core.ZodType[any, any] {
	return any(z).(core.ZodType[any, any])
}

// --- Default & Prefault wrappers --------------------------------------------------

// ZodRecordDefault embeds ZodDefault for method promotion
type ZodRecordDefault struct {
	*ZodDefault[*ZodRecord]
}

// ZodRecordPrefault embeds ZodPrefault for method promotion
type ZodRecordPrefault struct {
	*ZodPrefault[*ZodRecord]
}

// Default adds a default value to the record schema
func (z *ZodRecord) Default(value map[any]any) ZodRecordDefault {
	return ZodRecordDefault{
		&ZodDefault[*ZodRecord]{
			innerType:    z,
			defaultValue: value,
			isFunction:   false,
		},
	}
}

// DefaultFunc adds a default function to the record schema
func (z *ZodRecord) DefaultFunc(fn func() map[any]any) ZodRecordDefault {
	genericFn := func() any { return fn() }
	return ZodRecordDefault{
		&ZodDefault[*ZodRecord]{
			innerType:   z,
			defaultFunc: genericFn,
			isFunction:  true,
		},
	}
}

// Prefault adds a prefault value to the record schema
func (z *ZodRecord) Prefault(value map[any]any) ZodRecordPrefault {
	baseInternals := z.GetInternals()
	internals := &core.ZodTypeInternals{
		Version:     core.Version,
		Type:        core.ZodTypePrefault,
		Checks:      baseInternals.Checks,
		Coerce:      baseInternals.Coerce,
		Optional:    baseInternals.Optional,
		Nilable:     baseInternals.Nilable,
		Constructor: baseInternals.Constructor,
		Values:      baseInternals.Values,
		Pattern:     baseInternals.Pattern,
		Error:       baseInternals.Error,
		Bag:         baseInternals.Bag,
	}

	return ZodRecordPrefault{
		&ZodPrefault[*ZodRecord]{
			internals:     internals,
			innerType:     z,
			prefaultValue: value,
			prefaultFunc:  nil,
			isFunction:    false,
		},
	}
}

// PrefaultFunc adds a prefault function to the record schema
func (z *ZodRecord) PrefaultFunc(fn func() map[any]any) ZodRecordPrefault {
	baseInternals := z.GetInternals()
	internals := &core.ZodTypeInternals{
		Version:     core.Version,
		Type:        core.ZodTypePrefault,
		Checks:      baseInternals.Checks,
		Coerce:      baseInternals.Coerce,
		Optional:    baseInternals.Optional,
		Nilable:     baseInternals.Nilable,
		Constructor: baseInternals.Constructor,
		Values:      baseInternals.Values,
		Pattern:     baseInternals.Pattern,
		Error:       baseInternals.Error,
		Bag:         baseInternals.Bag,
	}

	genericFn := func() any { return fn() }
	return ZodRecordPrefault{
		&ZodPrefault[*ZodRecord]{
			internals:     internals,
			innerType:     z,
			prefaultValue: nil,
			prefaultFunc:  genericFn,
			isFunction:    true,
		},
	}
}

// createZodRecordFromDef creates a ZodRecord from definition
func createZodRecordFromDef(def *ZodRecordDef) *ZodRecord {
	internals := &ZodRecordInternals{
		ZodTypeInternals: engine.NewBaseZodTypeInternals(def.Type),
		Def:              def,
		KeyType:          def.KeyType,
		ValueType:        def.ValueType,
		Isst:             issues.ZodIssueInvalidType{Expected: core.ZodTypeRecord},
		Bag:              make(map[string]any),
	}

	// Set up simplified constructor for cloning
	internals.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any, any] {
		recordDef := &ZodRecordDef{
			ZodTypeDef: *newDef,
			Type:       core.ZodTypeRecord,
			KeyType:    def.KeyType,   // Preserve original key type
			ValueType:  def.ValueType, // Preserve original value type
		}
		return any(createZodRecordFromDef(recordDef)).(core.ZodType[any, any])
	}

	// Set up parse function
	internals.Parse = func(payload *core.ParsePayload, ctx *core.ParseContext) *core.ParsePayload {
		schema := &ZodRecord{internals: internals}
		result, err := schema.Parse(payload.Value, ctx)
		if err != nil {
			var zodErr *issues.ZodError
			if errors.As(err, &zodErr) {
				for _, issue := range zodErr.Issues {
					rawIssue := core.ZodRawIssue{
						Code:    issue.Code,
						Input:   issue.Input,
						Path:    issue.Path,
						Message: issue.Message,
					}
					payload.Issues = append(payload.Issues, rawIssue)
				}
			}
			return payload
		}
		payload.Value = result
		return payload
	}

	zodSchema := &ZodRecord{internals: internals}

	// Initialize the schema with proper error handling support
	engine.InitZodType(zodSchema, &def.ZodTypeDef)

	return zodSchema
}

// Record creates a new record schema
func Record(keyType, valueType core.ZodType[any, any], params ...any) *ZodRecord {
	def := &ZodRecordDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeRecord,
			Checks: make([]core.ZodCheck, 0),
		},
		Type:      core.ZodTypeRecord,
		KeyType:   keyType,
		ValueType: valueType,
	}

	schema := createZodRecordFromDef(def)

	// Apply schema parameters
	if len(params) > 0 {
		if param, ok := params[0].(core.SchemaParams); ok {
			// Apply parameters using engine.ApplySchemaParams
			engine.ApplySchemaParams(&def.ZodTypeDef, param)

			// Handle additional parameters
			if param.Description != "" {
				schema.internals.Bag["description"] = param.Description
			}
			if param.Abort {
				schema.internals.Bag["abort"] = true
			}
			if len(param.Path) > 0 {
				schema.internals.Bag["path"] = param.Path
			}
			if len(param.Params) > 0 {
				schema.internals.Bag["params"] = param.Params
			}
		}
	}

	return schema
}

////////////////////////////
////   HELPER FUNCTIONS ////
////////////////////////////

// extractRecordPointerValue extract record pointer value
func extractRecordPointerValue(input any) (map[any]any, bool, error) {
	// Use mapx.ExtractRecord for robust extraction
	if record, ok := mapx.ExtractRecord(input); ok {
		return record, false, nil
	}

	// Handle pointer to record types
	if reflectx.IsPointer(input) {
		if reflectx.IsNilPointer(input) {
			return nil, true, nil
		}
		if deref, ok := reflectx.Deref(input); ok {
			if record, ok := mapx.ExtractRecord(deref); ok {
				return record, false, nil
			}
		}
	}

	return nil, false, ErrExpectedRecord
}

////////////////////////////
////   PARTIAL RECORD   ////
////////////////////////////

// PartialRecord creates a partial record schema where keys are optional
func PartialRecord(keyType, valueType core.ZodType[any, any], params ...any) *ZodRecord {
	// For simplicity, just use the original keyType for now
	// In a full implementation, this would create a union with optional keys
	return Record(keyType, valueType, params...)
}

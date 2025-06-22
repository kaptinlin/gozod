package types

import (
	"errors"
	"reflect"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/pkg/mapx"
	"github.com/kaptinlin/gozod/pkg/reflectx"
)

// Error definitions for map transformations
var (
	ErrTransformNilMap = errors.New("cannot transform nil map")
)

// =============================================================================
// MAP TYPE DEFINITION (Go map[K]V validation)
// =============================================================================

// ZodMapDef defines the configuration for map validation
type ZodMapDef struct {
	core.ZodTypeDef
	Type      string // "map"
	KeyType   any    // The key schema (type-erased)
	ValueType any    // The value schema (type-erased)
}

// ZodMapInternals contains map validator internal state
type ZodMapInternals struct {
	core.ZodTypeInternals
	Def       *ZodMapDef
	KeyType   any            // The key schema
	ValueType any            // The value schema
	Bag       map[string]any // Runtime configuration
}

// ZodMap represents a map validation schema for Go map[K]V types
type ZodMap struct {
	internals *ZodMapInternals
}

// =============================================================================
// MAP TYPE METHODS
// =============================================================================

// GetInternals returns the internal state of the map schema
func (z *ZodMap) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// GetZod returns the map-specific internals for framework usage
func (z *ZodMap) GetZod() *ZodMapInternals {
	return z.internals
}

// Parse parse map with smart type inference
func (z *ZodMap) Parse(input any, ctx ...*core.ParseContext) (any, error) {
	parseCtx := (*core.ParseContext)(nil)
	if len(ctx) > 0 {
		parseCtx = ctx[0]
	}

	// ---------------------------------------------------------------------
	// Fast-path: if the input is a *map[…], we want to preserve the original
	// pointer identity and type. The original implementation relied on
	// engine.ParseType with *map[any]any which caused the pointer case
	// (e.g. *map[string]int) to be treated as an invalid type. We therefore
	// perform validation manually for pointer-to-map inputs and, on success,
	// return the original pointer unchanged so that downstream code and
	// tests that assert pointer equality continue to work.
	// ---------------------------------------------------------------------

	if rv := reflect.ValueOf(input); rv.Kind() == reflect.Ptr && !rv.IsNil() && rv.Elem().Kind() == reflect.Map {
		// Handle nilable flag when the pointer itself is nil (already checked via IsNil above).

		// Extract record in generic form for validation purposes.
		record, ok := mapx.ExtractRecord(rv.Elem().Interface())
		if !ok {
			rawIssue := issues.CreateInvalidTypeIssue("map", input)
			finalIssue := issues.FinalizeIssue(rawIssue, parseCtx, core.GetConfig())
			return nil, issues.NewZodError([]core.ZodIssue{finalIssue})
		}

		// Validator function reused from the original inline implementation.
		validate := func(value map[any]any) error {
			// Run schema-level checks first (Length/Min/Max etc.).
			if len(z.internals.Checks) > 0 {
				payload := &core.ParsePayload{Value: value, Issues: make([]core.ZodRawIssue, 0)}
				engine.RunChecksOnValue(value, z.internals.Checks, payload, parseCtx)
				if len(payload.Issues) > 0 {
					finalized := make([]core.ZodIssue, len(payload.Issues))
					for i, raw := range payload.Issues {
						finalized[i] = issues.FinalizeIssue(raw, parseCtx, core.GetConfig())
					}
					return issues.NewZodError(finalized)
				}
			}

			// Validate individual key / value entries.
			var allErrors []core.ZodIssue
			for key, val := range value {
				// Key validation.
				validatedKey, keyErr := z.internals.KeyType.(core.ZodType[any, any]).Parse(key, parseCtx)
				if keyErr != nil {
					var zErr *issues.ZodError
					if errors.As(keyErr, &zErr) {
						for _, iss := range zErr.Issues {
							converted := core.ZodIssue{ZodIssueBase: core.ZodIssueBase{Code: iss.Code, Input: iss.Input, Path: append([]any{key}, iss.Path...), Message: iss.Message}}
							allErrors = append(allErrors, converted)
						}
					} else {
						raw := issues.CreateInvalidTypeIssue("key", key)
						raw.Path = []any{key}
						allErrors = append(allErrors, issues.FinalizeIssue(raw, parseCtx, core.GetConfig()))
					}
					continue
				}

				// Value validation.
				validatedValue, valErr := z.internals.ValueType.(core.ZodType[any, any]).Parse(val, parseCtx)
				if valErr != nil {
					var zErr *issues.ZodError
					if errors.As(valErr, &zErr) {
						for _, iss := range zErr.Issues {
							converted := core.ZodIssue{ZodIssueBase: core.ZodIssueBase{Code: iss.Code, Input: iss.Input, Path: append([]any{key}, iss.Path...), Message: iss.Message}}
							allErrors = append(allErrors, converted)
						}
					} else {
						raw := issues.CreateInvalidTypeIssue("value", val)
						raw.Path = []any{key}
						allErrors = append(allErrors, issues.FinalizeIssue(raw, parseCtx, core.GetConfig()))
					}
					continue
				}

				// Mutate the underlying map entry with validated results (maintains pointer identity).
				// Use reflection to set values while respecting original key types.
				keyVal := reflect.ValueOf(key)
				// Ensure key convertibility before setting to avoid panic.
				if keyVal.Type().ConvertibleTo(rv.Elem().Type().Key()) {
					rv.Elem().SetMapIndex(keyVal.Convert(rv.Elem().Type().Key()), reflect.ValueOf(validatedValue))
				}

				// Also update the record to continue downstream validations if any.
				record[validatedKey] = validatedValue
			}

			if len(allErrors) > 0 {
				return issues.NewZodError(allErrors)
			}
			return nil
		}

		if err := validate(record); err != nil {
			return nil, err
		}

		return input, nil // pointer identity preserved
	}

	originalInput := input // Keep original for type preservation

	result, err := engine.ParseType[map[any]any](
		input,
		&z.internals.ZodTypeInternals,
		"map",
		func(v any) (map[any]any, bool) {
			// Use reflectx for nil checking
			if reflectx.IsNil(v) {
				return nil, false
			}

			// Use mapx for map extraction and conversion
			if mapData, ok := mapx.ExtractRecord(v); ok {
				return mapData, true
			}
			return nil, false
		},
		func(v any) (*map[any]any, bool) {
			if ptr, ok := v.(*map[any]any); ok {
				return ptr, true
			}
			return nil, false
		},
		func(value map[any]any, checks []core.ZodCheck, ctx *core.ParseContext) error {
			if len(checks) > 0 {
				payload := &core.ParsePayload{
					Value:  value,
					Issues: make([]core.ZodRawIssue, 0),
				}
				engine.RunChecksOnValue(value, checks, payload, ctx)
				if len(payload.Issues) > 0 {
					finalizedIssues := make([]core.ZodIssue, len(payload.Issues))
					for i, rawIssue := range payload.Issues {
						finalizedIssues[i] = issues.FinalizeIssue(rawIssue, ctx, core.GetConfig())
					}
					return issues.NewZodError(finalizedIssues)
				}
			}

			// Validate map keys and values using schemas
			result := make(map[any]any)
			var allErrors []core.ZodIssue

			for key, val := range value {
				// Validate key
				validatedKey, keyErr := z.internals.KeyType.(core.ZodType[any, any]).Parse(key, parseCtx)
				if keyErr != nil {
					var zodErr *issues.ZodError
					if errors.As(keyErr, &zodErr) {
						for _, issue := range zodErr.Issues {
							convertedIssue := core.ZodIssue{
								ZodIssueBase: core.ZodIssueBase{
									Code:    issue.Code,
									Input:   issue.Input,
									Path:    append([]any{key}, issue.Path...),
									Message: issue.Message,
								},
							}
							allErrors = append(allErrors, convertedIssue)
						}
					} else {
						rawIssue := issues.CreateInvalidTypeIssue("key", key)
						rawIssue.Path = []any{key}
						finalIssue := issues.FinalizeIssue(rawIssue, parseCtx, core.GetConfig())
						allErrors = append(allErrors, finalIssue)
					}
					continue
				}

				// Validate value
				validatedValue, valueErr := z.internals.ValueType.(core.ZodType[any, any]).Parse(val, parseCtx)
				if valueErr != nil {
					var zodErr *issues.ZodError
					if errors.As(valueErr, &zodErr) {
						for _, issue := range zodErr.Issues {
							convertedIssue := core.ZodIssue{
								ZodIssueBase: core.ZodIssueBase{
									Code:    issue.Code,
									Input:   issue.Input,
									Path:    append([]any{key}, issue.Path...),
									Message: issue.Message,
								},
							}
							allErrors = append(allErrors, convertedIssue)
						}
					} else {
						rawIssue := issues.CreateInvalidTypeIssue("value", val)
						rawIssue.Path = []any{key}
						finalIssue := issues.FinalizeIssue(rawIssue, parseCtx, core.GetConfig())
						allErrors = append(allErrors, finalIssue)
					}
					continue
				}

				result[validatedKey] = validatedValue
			}

			if len(allErrors) > 0 {
				return issues.NewZodError(allErrors)
			}

			// Update the original map
			for k := range value {
				delete(value, k)
			}
			for k, v := range result {
				value[k] = v
			}

			return nil
		},
		parseCtx,
	)

	if err != nil {
		return nil, err
	}

	// If the original input is a map (or pointer to map) and validation succeeded,
	// return it as-is to keep the exact type for callers (tests rely on this).
	if reflectx.IsMap(originalInput) || reflectx.IsPointer(originalInput) {
		return originalInput, nil
	}

	return result, nil
}

// MustParse parses the input and panics on validation failure
func (z *ZodMap) MustParse(input any, ctx ...*core.ParseContext) any {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// Transform transform map with given function
func (z *ZodMap) Transform(fn func(map[any]any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	return z.TransformAny(func(input any, ctx *core.RefinementContext) (any, error) {
		// Use reflectx for nil checking
		if reflectx.IsNil(input) {
			return nil, ErrTransformNilMap
		}

		// Use mapx for map extraction
		if mapData, ok := mapx.ExtractRecord(input); ok {
			return fn(mapData, ctx)
		}

		return nil, errors.New("expected map type for transformation")
	})
}

// TransformAny creates a transform with given function
func (z *ZodMap) TransformAny(fn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	transform := Transform[any, any](fn)

	return &ZodPipe[any, any]{
		in:  any(z).(core.ZodType[any, any]),
		out: any(transform).(core.ZodType[any, any]),
		def: core.ZodTypeDef{Type: "pipe"},
	}
}

// Pipe creates a pipe with another schema
func (z *ZodMap) Pipe(next core.ZodType[any, any]) core.ZodType[any, any] {
	return &ZodPipe[any, any]{
		in:  any(z).(core.ZodType[any, any]),
		out: next,
		def: core.ZodTypeDef{Type: "pipe"},
	}
}

// Optional makes the map optional
func (z *ZodMap) Optional() core.ZodType[any, any] {
	return any(Optional(any(z).(core.ZodType[any, any]))).(core.ZodType[any, any])
}

// Nilable make the map nilable
func (z *ZodMap) Nilable() core.ZodType[any, any] {
	return engine.Clone(z, func(def *core.ZodTypeDef) {
	}).(*ZodMap).setNilable()
}

// setNilable set nilable flag
func (z *ZodMap) setNilable() core.ZodType[any, any] {
	z.internals.Nilable = true
	return z
}

// Nullish makes the map both optional and nilable
func (z *ZodMap) Nullish() core.ZodType[any, any] {
	return any(Nullish(any(z).(core.ZodType[any, any]))).(core.ZodType[any, any])
}

// Refine adds a flexible validation function to the map
func (z *ZodMap) Refine(fn func(map[any]any) bool, params ...any) *ZodMap {
	result := z.RefineAny(func(v any) bool {
		// Use reflectx for nil checking
		if reflectx.IsNil(v) {
			return true // let the upper logic decide
		}

		// Use mapx for map extraction
		if mapData, ok := mapx.ExtractRecord(v); ok {
			return fn(mapData)
		}
		return false
	}, params...)
	return result.(*ZodMap)
}

// RefineAny adds a custom validation function to the map
func (z *ZodMap) RefineAny(fn func(any) bool, params ...any) core.ZodType[any, any] {
	check := checks.NewCustom[any](fn, params...)
	return engine.AddCheck(z, check)
}

// Check adds modern validation using direct payload access
func (z *ZodMap) Check(fn core.CheckFn) *ZodMap {
	check := checks.NewCustom[map[any]any](fn, core.SchemaParams{})
	result := engine.AddCheck(z, check)
	return result.(*ZodMap)
}

// Length adds a size validation check to the map
func (z *ZodMap) Length(size int, params ...any) *ZodMap {
	check := checks.NewCustom[map[any]any](func(v map[any]any) bool {
		return len(v) == size
	}, params...)
	result := engine.AddCheck(z, check)
	return result.(*ZodMap)
}

// Min adds a minimum size validation check to the map
func (z *ZodMap) Min(minimum int, params ...any) *ZodMap {
	check := checks.NewCustom[map[any]any](func(v map[any]any) bool {
		return len(v) >= minimum
	}, params...)
	result := engine.AddCheck(z, check)
	return result.(*ZodMap)
}

// Max adds a maximum size validation check to the map
func (z *ZodMap) Max(maximum int, params ...any) *ZodMap {
	check := checks.NewCustom[map[any]any](func(v map[any]any) bool {
		return len(v) <= maximum
	}, params...)
	result := engine.AddCheck(z, check)
	return result.(*ZodMap)
}

// ZodMapDefault is a Default wrapper for map type
type ZodMapDefault struct {
	*ZodDefault[*ZodMap]
}

type ZodMapPrefault struct {
	*ZodPrefault[*ZodMap] // embed specific pointer, allow method promotion
}

// Default adds a default value to the map
func (z *ZodMap) Default(value any) ZodMapDefault {
	return ZodMapDefault{
		&ZodDefault[*ZodMap]{
			innerType:    z,
			defaultValue: value,
			isFunction:   false,
		},
	}
}

// DefaultFunc adds a default function to the map
func (z *ZodMap) DefaultFunc(fn func() any) ZodMapDefault {
	genericFn := func() any { return fn() }
	return ZodMapDefault{
		&ZodDefault[*ZodMap]{
			innerType:   z,
			defaultFunc: genericFn,
			isFunction:  true,
		},
	}
}

// Prefault adds a prefault value to the map
func (z *ZodMap) Prefault(value any) ZodMapPrefault {
	baseInternals := z.GetInternals()
	internals := &core.ZodTypeInternals{
		Version:     core.Version,
		Type:        core.ZodTypePrefault,
		Checks:      baseInternals.Checks,
		Optional:    baseInternals.Optional,
		Nilable:     baseInternals.Nilable,
		Constructor: baseInternals.Constructor,
		Values:      baseInternals.Values,
		Pattern:     baseInternals.Pattern,
		Error:       baseInternals.Error,
		Bag:         baseInternals.Bag,
	}

	return ZodMapPrefault{
		&ZodPrefault[*ZodMap]{
			internals:     internals,
			innerType:     z,
			prefaultValue: value,
			prefaultFunc:  nil,
			isFunction:    false,
		},
	}
}

// PrefaultFunc adds a prefault function to the map
func (z *ZodMap) PrefaultFunc(fn func() any) ZodMapPrefault {
	baseInternals := z.GetInternals()
	internals := &core.ZodTypeInternals{
		Version:     core.Version,
		Type:        core.ZodTypePrefault,
		Checks:      baseInternals.Checks,
		Optional:    baseInternals.Optional,
		Nilable:     baseInternals.Nilable,
		Constructor: baseInternals.Constructor,
		Values:      baseInternals.Values,
		Pattern:     baseInternals.Pattern,
		Error:       baseInternals.Error,
		Bag:         baseInternals.Bag,
	}

	genericFn := func() any { return fn() }
	return ZodMapPrefault{
		&ZodPrefault[*ZodMap]{
			internals:     internals,
			innerType:     z,
			prefaultValue: nil,
			prefaultFunc:  genericFn,
			isFunction:    true,
		},
	}
}

// CloneFrom implements Cloneable interface for type-specific state copying
func (z *ZodMap) CloneFrom(source any) {
	if src, ok := source.(*ZodMap); ok {
		// Copy type-specific fields from Bag
		if src.internals.Bag != nil {
			if z.internals.Bag == nil {
				z.internals.Bag = make(map[string]any)
			}
			for k, v := range src.internals.Bag {
				z.internals.Bag[k] = v
			}
		}

		// Copy map-specific fields
		z.internals.KeyType = src.internals.KeyType
		z.internals.ValueType = src.internals.ValueType
	}
}

// Unwrap returns the inner type (for basic types, returns self)
func (z *ZodMap) Unwrap() core.ZodType[any, any] {
	return any(z).(core.ZodType[any, any])
}

// =============================================================================
// MAP TYPE CONSTRUCTORS
// =============================================================================

// createZodMapFromDef creates a ZodMap instance from a definition following the unified pattern
func createZodMapFromDef(def *ZodMapDef, params ...any) *ZodMap {
	// Create internals with modern pattern
	internals := &ZodMapInternals{
		ZodTypeInternals: engine.NewBaseZodTypeInternals("map"),
		Def:              def,
		KeyType:          def.KeyType,
		ValueType:        def.ValueType,
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
		mapDef := &ZodMapDef{
			ZodTypeDef: *newDef,
			Type:       "map",
			KeyType:    def.KeyType,
			ValueType:  def.ValueType,
		}
		return any(createZodMapFromDef(mapDef)).(core.ZodType[any, any])
	}

	// Set up parse function
	internals.Parse = func(payload *core.ParsePayload, ctx *core.ParseContext) *core.ParsePayload {
		schema := &ZodMap{internals: internals}
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

	zodSchema := &ZodMap{internals: internals}

	// Use unified infrastructure for initialization
	engine.InitZodType(any(zodSchema).(core.ZodType[any, any]), &def.ZodTypeDef)

	return zodSchema
}

// Map creates a new map schema with the given key and value schemas
func Map(keySchema, valueSchema core.ZodType[any, any], params ...any) *ZodMap {
	// Create map definition
	def := &ZodMapDef{
		ZodTypeDef: core.ZodTypeDef{Type: "map"},
		Type:       "map",
		KeyType:    keySchema,
		ValueType:  valueSchema,
	}

	return createZodMapFromDef(def, params...)
}

// ==================== ZodMapDefault chain call methods ====================

// Length adds a size validation check to the map, returns ZodMapDefault support chain call
func (s ZodMapDefault) Length(size int, params ...any) ZodMapDefault {
	newInner := s.innerType.Length(size, params...)
	return ZodMapDefault{
		&ZodDefault[*ZodMap]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

// Min adds a minimum size validation check to the map, returns ZodMapDefault support chain call
func (s ZodMapDefault) Min(minimum int, params ...any) ZodMapDefault {
	newInner := s.innerType.Min(minimum, params...)
	return ZodMapDefault{
		&ZodDefault[*ZodMap]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

// Max adds a maximum size validation check to the map, returns ZodMapDefault support chain call
func (s ZodMapDefault) Max(maximum int, params ...any) ZodMapDefault {
	newInner := s.innerType.Max(maximum, params...)
	return ZodMapDefault{
		&ZodDefault[*ZodMap]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

// Refine adds a flexible validation function to the map, returns ZodMapDefault support chain call
func (s ZodMapDefault) Refine(fn func(map[any]any) bool, params ...any) ZodMapDefault {
	newInner := s.innerType.Refine(fn, params...)
	return ZodMapDefault{
		&ZodDefault[*ZodMap]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

// Transform adds a data transformation function to the map, returns ZodType support transform pipeline
func (s ZodMapDefault) Transform(fn func(map[any]any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	return s.ZodDefault.TransformAny(func(input any, ctx *core.RefinementContext) (any, error) {
		// Use reflectx for nil checking
		if reflectx.IsNil(input) {
			return nil, ErrTransformNilMap
		}

		// Use mapx for map extraction
		if mapData, ok := mapx.ExtractRecord(input); ok {
			return fn(mapData, ctx)
		}

		return nil, errors.New("expected map type for transformation")
	})
}

// Check adds a modern validation function to the map, returns ZodMapDefault support chain call
func (s ZodMapDefault) Check(fn core.CheckFn) ZodMapDefault {
	newInner := s.innerType.Check(fn)
	return ZodMapDefault{
		&ZodDefault[*ZodMap]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

// Optional adds an optional check to the map, returns ZodType support chain call
func (s ZodMapDefault) Optional() core.ZodType[any, any] {
	return Optional(any(s).(core.ZodType[any, any]))
}

// Nilable adds a nilable check to the map, returns ZodType support chain call
func (s ZodMapDefault) Nilable() core.ZodType[any, any] {
	return Nilable(any(s).(core.ZodType[any, any]))
}

// ==================== ZodMapPrefault chain call methods ====================

// Length adds a size validation check to the map, returns ZodMapPrefault support chain call
func (s ZodMapPrefault) Length(size int, params ...any) ZodMapPrefault {
	newInner := s.innerType.Length(size, params...)
	return ZodMapPrefault{
		&ZodPrefault[*ZodMap]{
			internals:     s.internals,
			innerType:     newInner,
			prefaultValue: s.prefaultValue,
			prefaultFunc:  s.prefaultFunc,
			isFunction:    s.isFunction,
		},
	}
}

// Min adds a minimum size validation check to the map, returns ZodMapPrefault support chain call
func (s ZodMapPrefault) Min(minimum int, params ...any) ZodMapPrefault {
	newInner := s.innerType.Min(minimum, params...)
	return ZodMapPrefault{
		&ZodPrefault[*ZodMap]{
			internals:     s.internals,
			innerType:     newInner,
			prefaultValue: s.prefaultValue,
			prefaultFunc:  s.prefaultFunc,
			isFunction:    s.isFunction,
		},
	}
}

// Max adds a maximum size validation check to the map, returns ZodMapPrefault support chain call
func (s ZodMapPrefault) Max(maximum int, params ...any) ZodMapPrefault {
	newInner := s.innerType.Max(maximum, params...)
	return ZodMapPrefault{
		&ZodPrefault[*ZodMap]{
			internals:     s.internals,
			innerType:     newInner,
			prefaultValue: s.prefaultValue,
			prefaultFunc:  s.prefaultFunc,
			isFunction:    s.isFunction,
		},
	}
}

// Refine adds a flexible validation function to the map, returns ZodMapPrefault support chain call
func (s ZodMapPrefault) Refine(fn func(map[any]any) bool, params ...any) ZodMapPrefault {
	newInner := s.innerType.Refine(fn, params...)
	return ZodMapPrefault{
		&ZodPrefault[*ZodMap]{
			internals:     s.internals,
			innerType:     newInner,
			prefaultValue: s.prefaultValue,
			prefaultFunc:  s.prefaultFunc,
			isFunction:    s.isFunction,
		},
	}
}

// Transform adds a data transformation function to the map, returns ZodType support transform pipeline
func (s ZodMapPrefault) Transform(fn func(map[any]any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	return s.ZodPrefault.TransformAny(func(input any, ctx *core.RefinementContext) (any, error) {
		// Use reflectx for nil checking
		if reflectx.IsNil(input) {
			return nil, ErrTransformNilMap
		}

		// Use mapx for map extraction
		if mapData, ok := mapx.ExtractRecord(input); ok {
			return fn(mapData, ctx)
		}

		return nil, errors.New("expected map type for transformation")
	})
}

// TransformAny adds a generic data transformation function to the map, returns ZodType support transform pipeline
func (s ZodMapPrefault) TransformAny(fn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	return s.ZodPrefault.TransformAny(fn)
}

// Check adds a modern validation function to the map, returns ZodMapPrefault support chain call
func (s ZodMapPrefault) Check(fn core.CheckFn) ZodMapPrefault {
	newInner := s.innerType.Check(fn)
	return ZodMapPrefault{
		&ZodPrefault[*ZodMap]{
			internals:     s.internals,
			innerType:     newInner,
			prefaultValue: s.prefaultValue,
			prefaultFunc:  s.prefaultFunc,
			isFunction:    s.isFunction,
		},
	}
}

// Optional adds an optional check to the map, returns ZodType support chain call
func (s ZodMapPrefault) Optional() core.ZodType[any, any] {
	return Optional(any(s).(core.ZodType[any, any]))
}

// Nilable adds a nilable check to the map, returns ZodType support chain call
func (s ZodMapPrefault) Nilable() core.ZodType[any, any] {
	return Nilable(any(s).(core.ZodType[any, any]))
}

// ==================== ZodMapPrefault core interface methods ====================

// Parse implements ZodType interface, delegates to embedded ZodPrefault
func (s ZodMapPrefault) Parse(input any, ctx ...*core.ParseContext) (any, error) {
	return s.ZodPrefault.Parse(input, ctx...)
}

// MustParse implements ZodType interface, delegates to embedded ZodPrefault
func (s ZodMapPrefault) MustParse(input any, ctx ...*core.ParseContext) any {
	return s.ZodPrefault.MustParse(input, ctx...)
}

// GetInternals implements ZodType interface, delegates to embedded ZodPrefault
func (s ZodMapPrefault) GetInternals() *core.ZodTypeInternals {
	return s.ZodPrefault.GetInternals()
}

// Pipe implements ZodType interface, delegates to embedded ZodPrefault
func (s ZodMapPrefault) Pipe(out core.ZodType[any, any]) core.ZodType[any, any] {
	return s.ZodPrefault.Pipe(out)
}

// RefineAny implements ZodType interface, delegates to embedded ZodPrefault
func (s ZodMapPrefault) RefineAny(fn func(any) bool, params ...any) core.ZodType[any, any] {
	return s.ZodPrefault.RefineAny(fn, params...)
}

// Unwrap implements ZodType interface, delegates to embedded ZodPrefault
func (s ZodMapPrefault) Unwrap() core.ZodType[any, any] {
	return s.ZodPrefault.Unwrap()
}

// Prefault implements ZodType interface, delegates to embedded ZodPrefault
func (s ZodMapPrefault) Prefault(value any) core.ZodType[any, any] {
	return s.ZodPrefault.Prefault(value)
}

// PrefaultFunc implements ZodType interface, delegates to embedded ZodPrefault
func (s ZodMapPrefault) PrefaultFunc(fn func() any) core.ZodType[any, any] {
	return s.ZodPrefault.PrefaultFunc(fn)
}

package types

import (
	"fmt"
	"mime/multipart"
	"os"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/pkg/reflectx"
	"github.com/kaptinlin/gozod/pkg/slicex"
)

//////////////////////////////
////   FILE DEFINITION    ////
//////////////////////////////

// ZodFileDef defines the configuration for file validation
type ZodFileDef struct {
	core.ZodTypeDef
	Type string // "file"
}

// ZodFileInternals contains file validator internal state
type ZodFileInternals struct {
	core.ZodTypeInternals
	Def *ZodFileDef    // File definition
	Bag map[string]any // Additional metadata (minimum, maximum, mime)
}

// ZodFile represents a file validation schema
type ZodFile struct {
	internals *ZodFileInternals
}

//////////////////////////////
////   INTERNAL METHODS   ////
//////////////////////////////

// createZodFileFromDef creates a ZodFile from definition using pkg utilities
func createZodFileFromDef(def *ZodFileDef) *ZodFile {
	internals := &ZodFileInternals{
		ZodTypeInternals: engine.NewBaseZodTypeInternals(def.Type),
		Def:              def,
		Bag:              make(map[string]any),
	}

	// Copy checks from definition
	if len(def.Checks) > 0 {
		internals.Checks = append(internals.Checks, def.Checks...)
	}

	// Set up constructor for cloning
	internals.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any, any] {
		fileDef := &ZodFileDef{
			ZodTypeDef: *newDef,
			Type:       "file",
		}
		return any(createZodFileFromDef(fileDef)).(core.ZodType[any, any])
	}

	internals.Parse = func(payload *core.ParsePayload, ctx *core.ParseContext) *core.ParsePayload {
		return parseZodFile(payload, def, &internals.ZodTypeInternals, ctx)
	}

	zodSchema := &ZodFile{internals: internals}
	engine.InitZodType(zodSchema, &def.ZodTypeDef)
	return zodSchema
}

// parseZodFile implements the core parsing logic for file type
func parseZodFile(payload *core.ParsePayload, def *ZodFileDef, internals *core.ZodTypeInternals, ctx *core.ParseContext) *core.ParsePayload {
	input := payload.Value

	// 1. Unified nil handling
	if input == nil {
		if !internals.Nilable {
			rawIssue := issues.CreateInvalidTypeIssue("file", input)
			payload.Issues = append(payload.Issues, rawIssue)
			return payload
		}
		payload.Value = (*multipart.FileHeader)(nil) // Return typed nil pointer
		return payload
	}

	// 2. Smart type inference: preserve exact file type
	switch v := input.(type) {
	case *multipart.FileHeader:
		if v == nil {
			if !internals.Nilable {
				rawIssue := issues.CreateInvalidTypeIssue("file", input)
				payload.Issues = append(payload.Issues, rawIssue)
				return payload
			}
			payload.Value = (*multipart.FileHeader)(nil)
			return payload
		}
		// *multipart.FileHeader → *multipart.FileHeader (preserve pointer)
		if len(internals.Checks) > 0 {
			engine.RunChecksOnValue(v, internals.Checks, payload, ctx)
			if len(payload.Issues) > 0 {
				return payload
			}
		}
		payload.Value = v // Keep original pointer
		return payload

	case multipart.FileHeader:
		// multipart.FileHeader → multipart.FileHeader (value type)
		if len(internals.Checks) > 0 {
			engine.RunChecksOnValue(v, internals.Checks, payload, ctx)
			if len(payload.Issues) > 0 {
				return payload
			}
		}
		payload.Value = v
		return payload

	case *os.File:
		if v == nil {
			if !internals.Nilable {
				rawIssue := issues.CreateInvalidTypeIssue("file", input)
				payload.Issues = append(payload.Issues, rawIssue)
				return payload
			}
			payload.Value = (*os.File)(nil)
			return payload
		}
		// *os.File → *os.File (preserve pointer)
		if len(internals.Checks) > 0 {
			engine.RunChecksOnValue(v, internals.Checks, payload, ctx)
			if len(payload.Issues) > 0 {
				return payload
			}
		}
		payload.Value = v // Keep original pointer
		return payload

	case os.File:
		// os.File → os.File (value type)
		if len(internals.Checks) > 0 {
			engine.RunChecksOnValue(v, internals.Checks, payload, ctx)
			if len(payload.Issues) > 0 {
				return payload
			}
		}
		payload.Value = v
		return payload

	default:
		// 3. Type coercion (if enabled) - files typically don't support coercion
		// Files generally don't support coercion, skip this section

		// 4. Unified error creation
		rawIssue := issues.CreateInvalidTypeIssue("file", input)
		payload.Issues = append(payload.Issues, rawIssue)
		return payload
	}
}

//////////////////////////////
////   CONSTRUCTORS       ////
//////////////////////////////

// File creates a new file schema using pkg utilities
func File(params ...any) *ZodFile {
	def := &ZodFileDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   "file",
			Checks: make([]core.ZodCheck, 0),
		},
		Type: "file",
	}

	schema := createZodFileFromDef(def)

	// Apply schema parameters following unified pattern
	if len(params) > 0 {
		// Convert any params to core.SchemaParams format
		if param, ok := params[0].(core.SchemaParams); ok {
			// Handle schema-level error configuration using issues.CreateErrorMap
			if param.Error != nil {
				errorMap := issues.CreateErrorMap(param.Error)
				if errorMap != nil {
					def.Error = errorMap
					schema.internals.Error = errorMap
				}
			}

			// Handle coercion configuration
			if param.Coerce {
				schema.internals.Coerce = true
				if schema.internals.Bag == nil {
					schema.internals.Bag = make(map[string]any)
				}
				schema.internals.Bag["coerce"] = true
			}

			// Handle description
			if param.Description != "" {
				if schema.internals.Bag == nil {
					schema.internals.Bag = make(map[string]any)
				}
				schema.internals.Bag["description"] = param.Description
			}
		}
	}

	return schema
}

// =============================================================================
// ZODTYPE INTERFACE IMPLEMENTATION WITH SMART TYPE INFERENCE
// =============================================================================

// Parse implements intelligent type inference and validation
func (z *ZodFile) Parse(input any, ctx ...*core.ParseContext) (any, error) {
	parseCtx := (*core.ParseContext)(nil)
	if len(ctx) > 0 {
		parseCtx = ctx[0]
	}

	// 1. Unified nil handling
	if input == nil {
		if !z.internals.Nilable {
			rawIssue := issues.CreateInvalidTypeIssue("file", input)
			finalIssue := issues.FinalizeIssue(rawIssue, parseCtx, core.GetConfig())
			return nil, issues.NewZodError([]core.ZodIssue{finalIssue})
		}
		return (*multipart.FileHeader)(nil), nil
	}

	// 2. Smart type inference: file → file, *file → *file
	switch v := input.(type) {
	case *multipart.FileHeader:
		if v == nil {
			if !z.internals.Nilable {
				rawIssue := issues.CreateInvalidTypeIssue("file", input)
				finalIssue := issues.FinalizeIssue(rawIssue, parseCtx, core.GetConfig())
				return nil, issues.NewZodError([]core.ZodIssue{finalIssue})
			}
			return (*multipart.FileHeader)(nil), nil
		}
		// *multipart.FileHeader → *multipart.FileHeader (preserve pointer)
		if len(z.internals.Checks) > 0 {
			payload := &core.ParsePayload{Value: v, Issues: make([]core.ZodRawIssue, 0)}
			engine.RunChecksOnValue(v, z.internals.Checks, payload, parseCtx)
			if len(payload.Issues) > 0 {
				finalizedIssues := make([]core.ZodIssue, len(payload.Issues))
				for i, rawIssue := range payload.Issues {
					finalizedIssues[i] = issues.FinalizeIssue(rawIssue, parseCtx, core.GetConfig())
				}
				return nil, issues.NewZodError(finalizedIssues)
			}
		}
		return v, nil

	case multipart.FileHeader:
		// multipart.FileHeader → multipart.FileHeader (value type)
		if len(z.internals.Checks) > 0 {
			payload := &core.ParsePayload{Value: v, Issues: make([]core.ZodRawIssue, 0)}
			engine.RunChecksOnValue(v, z.internals.Checks, payload, parseCtx)
			if len(payload.Issues) > 0 {
				finalizedIssues := make([]core.ZodIssue, len(payload.Issues))
				for i, rawIssue := range payload.Issues {
					finalizedIssues[i] = issues.FinalizeIssue(rawIssue, parseCtx, core.GetConfig())
				}
				return nil, issues.NewZodError(finalizedIssues)
			}
		}
		return v, nil

	case *os.File:
		if v == nil {
			if !z.internals.Nilable {
				rawIssue := issues.CreateInvalidTypeIssue("file", input)
				finalIssue := issues.FinalizeIssue(rawIssue, parseCtx, core.GetConfig())
				return nil, issues.NewZodError([]core.ZodIssue{finalIssue})
			}
			return (*os.File)(nil), nil
		}
		// *os.File → *os.File (preserve pointer)
		if len(z.internals.Checks) > 0 {
			payload := &core.ParsePayload{Value: v, Issues: make([]core.ZodRawIssue, 0)}
			engine.RunChecksOnValue(v, z.internals.Checks, payload, parseCtx)
			if len(payload.Issues) > 0 {
				finalizedIssues := make([]core.ZodIssue, len(payload.Issues))
				for i, rawIssue := range payload.Issues {
					finalizedIssues[i] = issues.FinalizeIssue(rawIssue, parseCtx, core.GetConfig())
				}
				return nil, issues.NewZodError(finalizedIssues)
			}
		}
		return v, nil

	case os.File:
		// os.File → os.File (value type)
		if len(z.internals.Checks) > 0 {
			payload := &core.ParsePayload{Value: v, Issues: make([]core.ZodRawIssue, 0)}
			engine.RunChecksOnValue(v, z.internals.Checks, payload, parseCtx)
			if len(payload.Issues) > 0 {
				finalizedIssues := make([]core.ZodIssue, len(payload.Issues))
				for i, rawIssue := range payload.Issues {
					finalizedIssues[i] = issues.FinalizeIssue(rawIssue, parseCtx, core.GetConfig())
				}
				return nil, issues.NewZodError(finalizedIssues)
			}
		}
		return v, nil

	default:
		// 3. Type coercion (if enabled) - files typically don't support coercion
		// Files generally don't support coercion, skip this section

		// 4. Unified error creation
		rawIssue := issues.CreateInvalidTypeIssue("file", input)
		finalIssue := issues.FinalizeIssue(rawIssue, parseCtx, core.GetConfig())
		return nil, issues.NewZodError([]core.ZodIssue{finalIssue})
	}
}

// MustParse parses the input and panics on failure
func (z *ZodFile) MustParse(input any, ctx ...*core.ParseContext) any {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// =============================================================================
// TYPE-SAFE REFINE METHODS
// =============================================================================

// Refine adds a type-safe refinement check for file types
func (z *ZodFile) Refine(fn func(any) bool, params ...any) *ZodFile {
	result := z.RefineAny(func(v any) bool {
		val, isNil, err := extractFileValue(v)
		if err != nil {
			return false
		}
		if isNil {
			return true // Let upper logic decide
		}
		return fn(val)
	}, params...)
	return result.(*ZodFile)
}

// RefineAny adds a refinement check that accepts any input type
func (z *ZodFile) RefineAny(fn func(any) bool, params ...any) core.ZodType[any, any] {
	check := checks.NewCustom[any](fn, params...)
	return engine.AddCheck(any(z).(core.ZodType[any, any]), check)
}

// =============================================================================
// FILE-SPECIFIC VALIDATION METHODS
// =============================================================================

// Min adds minimum file size validation using pkg utilities
func (z *ZodFile) Min(minimum int64, params ...any) *ZodFile {
	check := checks.NewCustom[any](func(v any) bool {
		if fileHeader, ok := v.(*multipart.FileHeader); ok {
			return fileHeader.Size >= minimum
		}
		if file, ok := v.(*os.File); ok {
			if stat, err := file.Stat(); err == nil {
				return stat.Size() >= minimum
			}
		}
		if fileValue, ok := v.(multipart.FileHeader); ok {
			return fileValue.Size >= minimum
		}
		if fileValue, ok := v.(os.File); ok {
			if stat, err := fileValue.Stat(); err == nil {
				return stat.Size() >= minimum
			}
		}
		return false
	}, params...)
	result := engine.AddCheck(any(z).(core.ZodType[any, any]), check)
	return result.(*ZodFile)
}

// Max adds maximum file size validation using pkg utilities
func (z *ZodFile) Max(maximum int64, params ...any) *ZodFile {
	check := checks.NewCustom[any](func(v any) bool {
		if fileHeader, ok := v.(*multipart.FileHeader); ok {
			return fileHeader.Size <= maximum
		}
		if file, ok := v.(*os.File); ok {
			if stat, err := file.Stat(); err == nil {
				return stat.Size() <= maximum
			}
		}
		if fileValue, ok := v.(multipart.FileHeader); ok {
			return fileValue.Size <= maximum
		}
		if fileValue, ok := v.(os.File); ok {
			if stat, err := fileValue.Stat(); err == nil {
				return stat.Size() <= maximum
			}
		}
		return false
	}, params...)
	result := engine.AddCheck(any(z).(core.ZodType[any, any]), check)
	return result.(*ZodFile)
}

// Size adds exact file size validation using pkg utilities
func (z *ZodFile) Size(size int64, params ...any) *ZodFile {
	check := checks.NewCustom[any](func(v any) bool {
		if fileHeader, ok := v.(*multipart.FileHeader); ok {
			return fileHeader.Size == size
		}
		if file, ok := v.(*os.File); ok {
			if stat, err := file.Stat(); err == nil {
				return stat.Size() == size
			}
		}
		if fileValue, ok := v.(multipart.FileHeader); ok {
			return fileValue.Size == size
		}
		if fileValue, ok := v.(os.File); ok {
			if stat, err := fileValue.Stat(); err == nil {
				return stat.Size() == size
			}
		}
		return false
	}, params...)
	result := engine.AddCheck(any(z).(core.ZodType[any, any]), check)
	return result.(*ZodFile)
}

// Mime adds MIME type validation using pkg utilities
func (z *ZodFile) Mime(mimeTypes []string, params ...any) *ZodFile {
	check := checks.NewCustom[any](func(v any) bool {
		var contentType string
		if fileHeader, ok := v.(*multipart.FileHeader); ok {
			contentType = fileHeader.Header.Get("Content-Type")
		} else if fileValue, ok := v.(multipart.FileHeader); ok {
			contentType = fileValue.Header.Get("Content-Type")
		} else {
			return false
		}

		// Use slicex.Contains for efficient slice searching
		return slicex.Contains(mimeTypes, contentType)
	}, params...)
	result := engine.AddCheck(any(z).(core.ZodType[any, any]), check)
	return result.(*ZodFile)
}

// =============================================================================
// TRANSFORM METHODS
// =============================================================================

// Transform creates a transformation pipeline for file types
func (z *ZodFile) Transform(fn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	return z.TransformAny(fn)
}

// TransformAny creates a transformation pipeline that accepts any input type
func (z *ZodFile) TransformAny(fn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	transform := Transform[any, any](fn)
	return &ZodPipe[any, any]{
		in:  any(z).(core.ZodType[any, any]),
		out: any(transform).(core.ZodType[any, any]),
		def: core.ZodTypeDef{Type: "pipe"},
	}
}

// =============================================================================
// NILABLE MODIFIER WITH CLONE PATTERN
// =============================================================================

// Nilable creates a new file schema that accepts nil values
func (z *ZodFile) Nilable() core.ZodType[any, any] {
	return z.setNilable()
}

func (z *ZodFile) setNilable() core.ZodType[any, any] {
	cloned := engine.Clone(z, func(def *core.ZodTypeDef) {
		// Clone operates on ZodTypeDef level
	})
	cloned.(*ZodFile).internals.Nilable = true
	return cloned
}

// =============================================================================
// WRAPPER METHODS FOR MODIFIERS
// =============================================================================

// Optional makes the file schema optional
func (z *ZodFile) Optional() core.ZodType[any, any] {
	return Optional(any(z).(core.ZodType[any, any]))
}

// Nullish makes the file schema both optional and nullable
func (z *ZodFile) Nullish() core.ZodType[any, any] {
	return Nullish(any(z).(core.ZodType[any, any]))
}

// Pipe creates a validation pipeline
func (z *ZodFile) Pipe(out core.ZodType[any, any]) core.ZodType[any, any] {
	return &ZodPipe[any, any]{
		in:  any(z).(core.ZodType[any, any]),
		out: out,
		def: core.ZodTypeDef{Type: "pipe"},
	}
}

// GetInternals returns the internal type information
func (z *ZodFile) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// =============================================================================
// HELPER CONSTRUCTORS FOR COMPATIBILITY
// =============================================================================

// GetZod returns the file-specific internals for framework usage
func (z *ZodFile) GetZod() *ZodFileInternals {
	return z.internals
}

// CloneFrom implements the Cloneable interface
// Copies type-specific state from another ZodFile instance
func (z *ZodFile) CloneFrom(source any) {
	if src, ok := source.(*ZodFile); ok {
		if src.internals.Bag != nil {
			z.internals.Bag = make(map[string]any)
			for k, v := range src.internals.Bag {
				z.internals.Bag[k] = v
			}
		}
	}
}

// =============================================================================
// HELPER FUNCTIONS FOR TYPE EXTRACTION
// =============================================================================

// extractFileValue extracts file value using pkg utilities, handling various input types
func extractFileValue(input any) (any, bool, error) {
	// Use reflectx.IsNil for unified nil checking
	if reflectx.IsNil(input) {
		return nil, true, nil
	}

	switch v := input.(type) {
	case *multipart.FileHeader:
		if reflectx.IsNil(v) {
			return nil, true, nil
		}
		return v, false, nil
	case multipart.FileHeader:
		return v, false, nil
	case *os.File:
		if reflectx.IsNil(v) {
			return nil, true, nil
		}
		return v, false, nil
	case os.File:
		return v, false, nil
	default:
		return nil, false, fmt.Errorf("expected file, got %T", input)
	}
}

// extractFilePointerValue extracts file value from pointer types

// Unwrap returns the inner type (for basic types, returns self)
func (z *ZodFile) Unwrap() core.ZodType[any, any] {
	return any(z).(core.ZodType[any, any])
}

////////////////////////////
////   FILE DEFAULT WRAPPER ////
////////////////////////////

// ZodFileDefault is a wrapper for the ZodFile type that allows for default values
type ZodFileDefault struct {
	*ZodDefault[*ZodFile]
}

////////////////////////////
////   DEFAULT method   ////
////////////////////////////

// Default adds a default value to the file
func (z *ZodFile) Default(value any) ZodFileDefault {
	return ZodFileDefault{
		&ZodDefault[*ZodFile]{
			innerType:    z,
			defaultValue: value,
			isFunction:   false,
		},
	}
}

// DefaultFunc adds a default function to the file
func (z *ZodFile) DefaultFunc(fn func() any) ZodFileDefault {
	return ZodFileDefault{
		&ZodDefault[*ZodFile]{
			innerType:   z,
			defaultFunc: fn,
			isFunction:  true,
		},
	}
}

////////////////////////////
////   FILEDEFAULT chain methods ////
////////////////////////////

// Min adds minimum file size validation, returns ZodFileDefault support chain call
func (s ZodFileDefault) Min(minimum int64, params ...any) ZodFileDefault {
	newInner := s.innerType.Min(minimum, params...)
	return ZodFileDefault{
		&ZodDefault[*ZodFile]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

// Max adds maximum file size validation, returns ZodFileDefault support chain call
func (s ZodFileDefault) Max(maximum int64, params ...any) ZodFileDefault {
	newInner := s.innerType.Max(maximum, params...)
	return ZodFileDefault{
		&ZodDefault[*ZodFile]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

// Mime adds MIME type validation, returns ZodFileDefault support chain call
func (s ZodFileDefault) Mime(mimeTypes []string, params ...any) ZodFileDefault {
	newInner := s.innerType.Mime(mimeTypes, params...)
	return ZodFileDefault{
		&ZodDefault[*ZodFile]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

// Refine adds a flexible validation function to the file, returns ZodFileDefault support chain call
func (s ZodFileDefault) Refine(fn func(any) bool, params ...any) ZodFileDefault {
	newInner := s.innerType.Refine(fn, params...)
	return ZodFileDefault{
		&ZodDefault[*ZodFile]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

// Transform adds data transformation, returns a generic ZodType support transform pipeline
func (s ZodFileDefault) Transform(fn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	// Use the TransformAny method of the embedded ZodDefault
	return s.TransformAny(fn)
}

// Optional adds an optional check to the file, returns ZodType support chain call
func (s ZodFileDefault) Optional() core.ZodType[any, any] {
	// wrap the current ZodFileDefault instance, keeping Default logic
	return Optional(any(s).(core.ZodType[any, any]))
}

// Nilable adds a nilable check to the file, returns ZodType support chain call
func (s ZodFileDefault) Nilable() core.ZodType[any, any] {
	// wrap the current ZodFileDefault instance, keeping Default logic
	return Nilable(any(s).(core.ZodType[any, any]))
}

////////////////////////////
////   FILE PREFAULT WRAPPER ////
////////////////////////////

// ZodFilePrefault is a wrapper for the ZodFile type that allows for prefault values
type ZodFilePrefault struct {
	*ZodPrefault[*ZodFile] // embed pointer, allowing method promotion
}

////////////////////////////
////   PREFAULT method   ////
////////////////////////////

// Prefault adds a prefault value to the file
func (z *ZodFile) Prefault(value any) ZodFilePrefault {
	// construct Prefault's internals, Type = "prefault", copy inner type's checks/coerce/optional/nilable
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

	return ZodFilePrefault{
		&ZodPrefault[*ZodFile]{
			internals:     internals,
			innerType:     z,
			prefaultValue: value,
			prefaultFunc:  nil,
			isFunction:    false,
		},
	}
}

// PrefaultFunc adds a prefault function to the file
func (z *ZodFile) PrefaultFunc(fn func() any) ZodFilePrefault {
	// construct Prefault's internals, Type = "prefault", copy inner type's checks/coerce/optional/nilable
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

	return ZodFilePrefault{
		&ZodPrefault[*ZodFile]{
			internals:     internals,
			innerType:     z,
			prefaultValue: nil,
			prefaultFunc:  fn,
			isFunction:    true,
		},
	}
}

////////////////////////////
////   FILEPREFAULT chain methods ////
////////////////////////////

// Min adds minimum file size validation, returns ZodFilePrefault support chain call
func (f ZodFilePrefault) Min(minimum int64, params ...any) ZodFilePrefault {
	newInner := f.innerType.Min(minimum, params...)

	// construct new internals
	baseInternals := newInner.GetInternals()
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

	return ZodFilePrefault{
		&ZodPrefault[*ZodFile]{
			internals:     internals,
			innerType:     newInner,
			prefaultValue: f.prefaultValue,
			prefaultFunc:  f.prefaultFunc,
			isFunction:    f.isFunction,
		},
	}
}

// Max adds maximum file size validation, returns ZodFilePrefault support chain call
func (f ZodFilePrefault) Max(maximum int64, params ...any) ZodFilePrefault {
	newInner := f.innerType.Max(maximum, params...)

	// construct new internals
	baseInternals := newInner.GetInternals()
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

	return ZodFilePrefault{
		&ZodPrefault[*ZodFile]{
			internals:     internals,
			innerType:     newInner,
			prefaultValue: f.prefaultValue,
			prefaultFunc:  f.prefaultFunc,
			isFunction:    f.isFunction,
		},
	}
}

// Mime adds MIME type validation, returns ZodFilePrefault support chain call
func (f ZodFilePrefault) Mime(mimeTypes []string, params ...any) ZodFilePrefault {
	newInner := f.innerType.Mime(mimeTypes, params...)

	// construct new internals
	baseInternals := newInner.GetInternals()
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

	return ZodFilePrefault{
		&ZodPrefault[*ZodFile]{
			internals:     internals,
			innerType:     newInner,
			prefaultValue: f.prefaultValue,
			prefaultFunc:  f.prefaultFunc,
			isFunction:    f.isFunction,
		},
	}
}

// Refine adds a flexible validation function to the file, returns ZodFilePrefault support chain call
func (f ZodFilePrefault) Refine(fn func(any) bool, params ...any) ZodFilePrefault {
	newInner := f.innerType.Refine(fn, params...)

	// construct new internals
	baseInternals := newInner.GetInternals()
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

	return ZodFilePrefault{
		&ZodPrefault[*ZodFile]{
			internals:     internals,
			innerType:     newInner,
			prefaultValue: f.prefaultValue,
			prefaultFunc:  f.prefaultFunc,
			isFunction:    f.isFunction,
		},
	}
}

// Transform adds data transformation, returns a generic ZodType support transform pipeline
func (f ZodFilePrefault) Transform(fn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	// use the TransformAny method of the embedded ZodPrefault
	return f.TransformAny(fn)
}

// Optional adds an optional check to the file, returns ZodType support chain call
func (f ZodFilePrefault) Optional() core.ZodType[any, any] {
	// wrap the current ZodFilePrefault instance, keeping Prefault logic
	return Optional(any(f).(core.ZodType[any, any]))
}

// Nilable adds a nilable check to the file, returns ZodType support chain call
func (f ZodFilePrefault) Nilable() core.ZodType[any, any] {
	// wrap the current ZodFilePrefault instance, keeping Prefault logic
	return Nilable(any(f).(core.ZodType[any, any]))
}

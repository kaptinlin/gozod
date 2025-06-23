package types

import (
	"errors"
	"mime/multipart"
	"os"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/pkg/reflectx"
)

// Error definitions for file transformations
var (
	ErrExpectedFile     = errors.New("expected file type")
	ErrTransformNilFile = errors.New("cannot transform nil file")
)

//////////////////////////////
////   FILE DEFINITION    ////
//////////////////////////////

// ZodFileDef defines the configuration for file validation
type ZodFileDef struct {
	core.ZodTypeDef
	Type   core.ZodTypeCode // Type identifier using type-safe constants
	Checks []core.ZodCheck  // File-specific validation checks
}

// ZodFileInternals contains file validator internal state
type ZodFileInternals struct {
	core.ZodTypeInternals
	Def  *ZodFileDef                // Schema definition
	Isst issues.ZodIssueInvalidType // Invalid type issue template
	Bag  map[string]any             // Additional metadata (minimum, maximum, mime)
}

// ZodFile represents a file validation schema with type safety
type ZodFile struct {
	internals *ZodFileInternals
}

// GetInternals returns the internal state of the schema
func (z *ZodFile) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// Parse implements intelligent type inference and validation using engine.ParseType
func (z *ZodFile) Parse(input any, ctx ...*core.ParseContext) (any, error) {
	parseCtx := (*core.ParseContext)(nil)
	if len(ctx) > 0 {
		parseCtx = ctx[0]
	}

	// Use engine.ParseType unified template with improved file handling
	return engine.ParseType[any](
		input,
		&z.internals.ZodTypeInternals,
		core.ZodTypeFile,
		// Type checker function - handles all file value types
		func(v any) (any, bool) {
			switch file := v.(type) {
			case *multipart.FileHeader:
				return file, true
			case multipart.FileHeader:
				return file, true
			case *os.File:
				return file, true
			case os.File:
				return file, true
			default:
				return nil, false
			}
		},
		// Pointer checker function - handles pointer dereferencing for file types
		func(v any) (*any, bool) {
			switch ptr := v.(type) {
			case **multipart.FileHeader:
				if ptr != nil {
					anyPtr := any(*ptr)
					return &anyPtr, true
				}
			case **os.File:
				if ptr != nil {
					anyPtr := any(*ptr)
					return &anyPtr, true
				}
			}
			return nil, false
		},
		// Validator function
		func(value any, checks []core.ZodCheck, ctx *core.ParseContext) error {
			return validateFile(value, checks, ctx)
		},
		parseCtx,
	)
}

// MustParse validates the input value and panics on failure
func (z *ZodFile) MustParse(input any, ctx ...*core.ParseContext) any {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

//////////////////////////////
////   INTERNAL METHODS   ////
//////////////////////////////

// createZodFileFromDef creates a ZodFile from definition using pkg utilities
func createZodFileFromDef(def *ZodFileDef) *ZodFile {
	internals := &ZodFileInternals{
		ZodTypeInternals: engine.NewBaseZodTypeInternals(def.Type),
		Def:              def,
		Isst:             issues.ZodIssueInvalidType{Expected: core.ZodTypeFile},
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
			Type:       core.ZodTypeFile,
			Checks:     newDef.Checks,
		}
		return any(createZodFileFromDef(fileDef)).(core.ZodType[any, any])
	}

	internals.Parse = func(payload *core.ParsePayload, ctx *core.ParseContext) *core.ParsePayload {
		result, err := engine.ParseType[any](
			payload.GetValue(),
			&internals.ZodTypeInternals,
			core.ZodTypeFile,
			// Type checker function - handles all file value types
			func(v any) (any, bool) {
				switch file := v.(type) {
				case *multipart.FileHeader:
					return file, true
				case multipart.FileHeader:
					return file, true
				case *os.File:
					return file, true
				case os.File:
					return file, true
				default:
					return nil, false
				}
			},
			// Pointer checker function - handles pointer dereferencing for file types
			func(v any) (*any, bool) {
				switch ptr := v.(type) {
				case **multipart.FileHeader:
					if ptr != nil {
						anyPtr := any(*ptr)
						return &anyPtr, true
					}
				case **os.File:
					if ptr != nil {
						anyPtr := any(*ptr)
						return &anyPtr, true
					}
				}
				return nil, false
			},
			// Validator function
			func(value any, checks []core.ZodCheck, ctx *core.ParseContext) error {
				return validateFile(value, checks, ctx)
			},
			ctx,
		)

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

	zodSchema := &ZodFile{internals: internals}
	engine.InitZodType(zodSchema, &def.ZodTypeDef)
	return zodSchema
}

//////////////////////////////
////   CONSTRUCTORS       ////
//////////////////////////////

// File creates a new file schema with unified parameter handling
func File(params ...any) *ZodFile {
	def := &ZodFileDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeFile,
			Checks: make([]core.ZodCheck, 0),
		},
		Type:   core.ZodTypeFile,
		Checks: make([]core.ZodCheck, 0),
	}

	schema := createZodFileFromDef(def)

	if len(params) > 0 {
		param := params[0]

		// Handle different parameter types
		switch p := param.(type) {
		case string:
			// String parameter becomes Error field
			errorMap := core.ZodErrorMap(func(issue core.ZodRawIssue) string {
				return p
			})
			def.Error = &errorMap
			schema.internals.Error = &errorMap
		case core.SchemaParams:
			// Handle core.SchemaParams
			if p.Error != nil {
				// Handle string error messages by converting to ZodErrorMap
				if errStr, ok := p.Error.(string); ok {
					errorMap := core.ZodErrorMap(func(issue core.ZodRawIssue) string {
						return errStr
					})
					def.Error = &errorMap
					schema.internals.Error = &errorMap
				} else if errorMap, ok := p.Error.(core.ZodErrorMap); ok {
					def.Error = &errorMap
					schema.internals.Error = &errorMap
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
		}
	}

	return schema
}

//////////////////////////////
////   VALIDATION METHODS  ////
//////////////////////////////

// Min adds minimum file size validation
func (z *ZodFile) Min(minimum int64, params ...any) *ZodFile {
	checkFn := func(payload *core.ParsePayload) {
		size := getFileSize(payload.GetValue())
		if size < minimum {
			raw := issues.CreateTooSmallIssue(minimum, true, "file", payload.GetValue())
			payload.AddIssue(raw)
		}
	}

	check := checks.NewCustom[any](checkFn, params...)
	result := engine.AddCheck(z, check)
	return result.(*ZodFile)
}

// Max adds maximum file size validation
func (z *ZodFile) Max(maximum int64, params ...any) *ZodFile {
	checkFn := func(payload *core.ParsePayload) {
		size := getFileSize(payload.GetValue())
		if size > maximum {
			raw := issues.CreateTooBigIssue(maximum, true, "file", payload.GetValue())
			payload.AddIssue(raw)
		}
	}

	check := checks.NewCustom[any](checkFn, params...)
	result := engine.AddCheck(z, check)
	return result.(*ZodFile)
}

// Size enforces exact file size validation
func (z *ZodFile) Size(expect int64, params ...any) *ZodFile {
	checkFn := func(payload *core.ParsePayload) {
		size := getFileSize(payload.GetValue())
		if size != expect {
			var raw core.ZodRawIssue
			if size > expect {
				raw = issues.CreateTooBigIssue(expect, true, "file", payload.GetValue())
			} else {
				raw = issues.CreateTooSmallIssue(expect, true, "file", payload.GetValue())
			}
			payload.AddIssue(raw)
		}
	}

	check := checks.NewCustom[any](checkFn, params...)
	result := engine.AddCheck(z, check)
	return result.(*ZodFile)
}

// Mime validates allowed MIME types
func (z *ZodFile) Mime(mimeTypes []string, params ...any) *ZodFile {
	check := checks.Mime(mimeTypes, params...)
	result := engine.AddCheck(z, check)
	return result.(*ZodFile)
}

//////////////////////////////
////   TRANSFORM METHODS   ////
//////////////////////////////

// Transform provides type-safe file transformation with smart dereferencing
func (z *ZodFile) Transform(fn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	return z.TransformAny(func(input any, ctx *core.RefinementContext) (any, error) {
		file, ok := extractFileValue(input)
		if !ok {
			// Try pointer dereferencing
			if ptr, ptrOk := reflectx.Deref(input); ptrOk {
				if file, ok = extractFileValue(ptr); !ok {
					return nil, ErrExpectedFile
				}
			} else {
				return nil, ErrExpectedFile
			}
		}

		return fn(file, ctx)
	})
}

// TransformAny flexible version of transformation
func (z *ZodFile) TransformAny(fn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	transform := Transform[any, any](fn)

	return &ZodPipe[any, any]{
		in:  any(z).(core.ZodType[any, any]),
		out: any(transform).(core.ZodType[any, any]),
		def: core.ZodTypeDef{Type: "pipe"},
	}
}

// Pipe operation for pipeline chaining
func (z *ZodFile) Pipe(out core.ZodType[any, any]) core.ZodType[any, any] {
	return &ZodPipe[any, any]{
		in:  z,
		out: out,
		def: core.ZodTypeDef{Type: "pipe"},
	}
}

//////////////////////////////
////   MODIFIER METHODS    ////
//////////////////////////////

// Optional makes the file optional
func (z *ZodFile) Optional() core.ZodType[any, any] {
	return any(Optional(any(z).(core.ZodType[any, any]))).(core.ZodType[any, any])
}

// Nilable makes the file nilable while preserving type inference
func (z *ZodFile) Nilable() core.ZodType[any, any] {
	return Nilable(any(z).(core.ZodType[any, any]))
}

// Nullish makes the file both optional and nilable
func (z *ZodFile) Nullish() core.ZodType[any, any] {
	return any(Nullish(any(z).(core.ZodType[any, any]))).(core.ZodType[any, any])
}

//////////////////////////////
////   REFINE METHODS      ////
//////////////////////////////

// Refine adds type-safe custom validation logic
func (z *ZodFile) Refine(fn func(any) bool, params ...any) *ZodFile {
	result := z.RefineAny(func(v any) bool {
		file, ok := extractFileValue(v)
		if !ok {
			// Try pointer dereferencing
			if ptr, ptrOk := reflectx.Deref(v); ptrOk {
				file, ok = extractFileValue(ptr)
			}
		}
		if !ok {
			return false // Let type validation handle wrong types
		}
		return fn(file)
	}, params...)
	return result.(*ZodFile)
}

// RefineAny adds flexible custom validation logic
func (z *ZodFile) RefineAny(fn func(any) bool, params ...any) core.ZodType[any, any] {
	check := checks.NewCustom[any](fn, params...)
	return engine.AddCheck(z, check)
}

//////////////////////////////
////   UNWRAP METHODS      ////
//////////////////////////////

// Unwrap returns the inner type (for basic types, returns self)
func (z *ZodFile) Unwrap() core.ZodType[any, any] {
	return any(z).(core.ZodType[any, any])
}

//////////////////////////////
////   WRAPPER TYPES       ////
//////////////////////////////

// ZodFileDefault is a default value wrapper for file type
type ZodFileDefault struct {
	*ZodDefault[*ZodFile]
}

//////////////////////////////
////   DEFAULT METHODS     ////
//////////////////////////////

// Default creates a default wrapper with type safety
func (z *ZodFile) Default(value any) ZodFileDefault {
	return ZodFileDefault{
		&ZodDefault[*ZodFile]{
			innerType:    z,
			defaultValue: value,
			isFunction:   false,
		},
	}
}

// DefaultFunc creates a default wrapper with function
func (z *ZodFile) DefaultFunc(fn func() any) ZodFileDefault {
	return ZodFileDefault{
		&ZodDefault[*ZodFile]{
			innerType:   z,
			defaultFunc: fn,
			isFunction:  true,
		},
	}
}

//////////////////////////////
////   FILEDEFAULT chain methods ////
//////////////////////////////

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

//////////////////////////////
////   FILEPREFAULT WRAPPER ////
//////////////////////////////

// ZodFilePrefault is a prefault value wrapper for file type
type ZodFilePrefault struct {
	*ZodPrefault[*ZodFile]
}

//////////////////////////////
////   PREFAULT method     ////
//////////////////////////////

// Prefault creates a prefault wrapper with type safety
func (z *ZodFile) Prefault(value any) ZodFilePrefault {
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

// PrefaultFunc creates a prefault wrapper with function
func (z *ZodFile) PrefaultFunc(fn func() any) ZodFilePrefault {
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

//////////////////////////////
////   FILEPREFAULT chain methods ////
//////////////////////////////

// Min adds minimum file size validation, returns ZodFilePrefault support chain call
func (f ZodFilePrefault) Min(minimum int64, params ...any) ZodFilePrefault {
	newInner := f.innerType.Min(minimum, params...)

	baseInternals := newInner.GetInternals()
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

	baseInternals := newInner.GetInternals()
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

	baseInternals := newInner.GetInternals()
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

	baseInternals := newInner.GetInternals()
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

//////////////////////////////
////   UTILITY METHODS     ////
//////////////////////////////

// GetZod returns the file-specific internals
func (z *ZodFile) GetZod() *ZodFileInternals {
	return z.internals
}

// CloneFrom implements Cloneable interface
func (z *ZodFile) CloneFrom(source any) {
	if src, ok := source.(interface{ GetZod() *ZodFileInternals }); ok {
		srcState := src.GetZod()
		tgtState := z.GetZod()

		if len(srcState.Bag) > 0 {
			if tgtState.Bag == nil {
				tgtState.Bag = make(map[string]any)
			}
			for key, value := range srcState.Bag {
				tgtState.Bag[key] = value
			}
		}

		if len(srcState.ZodTypeInternals.Bag) > 0 {
			if tgtState.ZodTypeInternals.Bag == nil {
				tgtState.ZodTypeInternals.Bag = make(map[string]any)
			}
			for key, value := range srcState.ZodTypeInternals.Bag {
				tgtState.ZodTypeInternals.Bag[key] = value
			}
		}
	}
}

// extractFileValue extracts file value using smart type checking
func extractFileValue(input any) (any, bool) {
	switch v := input.(type) {
	case *multipart.FileHeader:
		return v, true
	case multipart.FileHeader:
		return v, true
	case *os.File:
		return v, true
	case os.File:
		return v, true
	default:
		return nil, false
	}
}

// getFileSize returns file size in bytes (0 if unknown)
func getFileSize(v any) int64 {
	switch f := v.(type) {
	case *multipart.FileHeader:
		return f.Size
	case multipart.FileHeader:
		return f.Size
	case *os.File:
		if stat, err := f.Stat(); err == nil {
			return stat.Size()
		}
	case os.File:
		if stat, err := f.Stat(); err == nil {
			return stat.Size()
		}
	}
	return 0
}

//////////////////////////////
////   VALIDATION FUNCTIONS ////
//////////////////////////////

// validateFile validates file values with checks
func validateFile(value any, checks []core.ZodCheck, ctx *core.ParseContext) error {
	if len(checks) > 0 {
		// Use constructor instead of direct struct literal to respect private fields
		payload := core.NewParsePayload(value)
		engine.RunChecksOnValue(value, checks, payload, ctx)
		if len(payload.GetIssues()) > 0 {
			return issues.NewZodError(issues.ConvertRawIssuesToIssues(payload.GetIssues(), ctx))
		}
	}
	return nil
}

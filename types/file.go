package types

import (
	"mime/multipart"
	"os"
	"reflect"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/internal/utils"
)

// =============================================================================
// TYPE DEFINITIONS
// =============================================================================

// ZodFileDef defines the configuration for file validation.
type ZodFileDef struct {
	core.ZodTypeDef
}

// ZodFileInternals contains file validator internal state.
type ZodFileInternals struct {
	core.ZodTypeInternals
	Def *ZodFileDef
}

// ZodFile validates file inputs (*os.File, *multipart.FileHeader, multipart.File)
// with constraint types T (base) and R (output).
type ZodFile[T any, R any] struct {
	internals *ZodFileInternals
}

// =============================================================================
// CORE METHODS
// =============================================================================

// Internals returns the internal state of the schema.
func (z *ZodFile[T, R]) Internals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// IsOptional reports whether this schema accepts undefined/missing values.
func (z *ZodFile[T, R]) IsOptional() bool {
	return z.internals.IsOptional()
}

// IsNilable reports whether this schema accepts nil values.
func (z *ZodFile[T, R]) IsNilable() bool {
	return z.internals.IsNilable()
}

// Parse validates input and returns a value matching the generic type R.
func (z *ZodFile[T, R]) Parse(input any, ctx ...*core.ParseContext) (R, error) {
	result, err := engine.ParseComplex(
		input,
		&z.internals.ZodTypeInternals,
		z.expectedType(),
		z.extractFileForEngine,
		z.extractFilePtrForEngine,
		z.validateFileForEngine,
		ctx...,
	)
	if err != nil {
		var zero R
		return zero, err
	}

	return convertFileResult[T, R](result)
}

// MustParse validates input and panics on failure.
func (z *ZodFile[T, R]) MustParse(input any, ctx ...*core.ParseContext) R {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// StrictParse provides compile-time type safety by requiring exact type T.
func (z *ZodFile[T, R]) StrictParse(input T, ctx ...*core.ParseContext) (R, error) {
	result, err := engine.ParseComplexStrict(
		input,
		&z.internals.ZodTypeInternals,
		z.expectedType(),
		z.extractFileForEngine,
		z.extractFilePtrForEngine,
		z.validateFileForEngine,
		ctx...,
	)
	if err != nil {
		var zero R
		return zero, err
	}

	return convertToFileConstraintType[T, R](result), nil
}

// MustStrictParse provides compile-time type safety and panics on failure.
func (z *ZodFile[T, R]) MustStrictParse(input T, ctx ...*core.ParseContext) R {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// ParseAny validates input and returns any type for runtime interface usage.
func (z *ZodFile[T, R]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// =============================================================================
// MODIFIER METHODS
// =============================================================================

// Optional returns a new schema that accepts nil, with pointer constraint.
func (z *ZodFile[T, R]) Optional() *ZodFile[T, *T] {
	in := z.internals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// ExactOptional accepts absent keys but rejects explicit nil values.
func (z *ZodFile[T, R]) ExactOptional() *ZodFile[T, R] {
	in := z.internals.Clone()
	in.SetExactOptional(true)
	return z.withInternals(in)
}

// Nilable returns a new schema that accepts nil values, with pointer constraint.
func (z *ZodFile[T, R]) Nilable() *ZodFile[T, *T] {
	in := z.internals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish returns a new schema combining optional and nilable modifiers.
func (z *ZodFile[T, R]) Nullish() *ZodFile[T, *T] {
	in := z.internals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Default sets a fallback value returned when input is nil (short-circuits validation).
func (z *ZodFile[T, R]) Default(v T) *ZodFile[T, R] {
	in := z.internals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

// DefaultFunc sets a fallback function called when input is nil (short-circuits validation).
func (z *ZodFile[T, R]) DefaultFunc(fn func() T) *ZodFile[T, R] {
	in := z.internals.Clone()
	in.SetDefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Prefault sets a fallback value that goes through the full validation pipeline.
func (z *ZodFile[T, R]) Prefault(v T) *ZodFile[T, R] {
	in := z.internals.Clone()
	in.SetPrefaultValue(v)
	return z.withInternals(in)
}

// PrefaultFunc sets a fallback function that goes through the full validation pipeline.
func (z *ZodFile[T, R]) PrefaultFunc(fn func() T) *ZodFile[T, R] {
	in := z.internals.Clone()
	in.SetPrefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Meta stores metadata for this file schema in the global registry.
func (z *ZodFile[T, R]) Meta(meta core.GlobalMeta) *ZodFile[T, R] {
	core.GlobalRegistry.Add(z, meta)
	return z
}

// Describe registers a description in the global registry.
func (z *ZodFile[T, R]) Describe(desc string) *ZodFile[T, R] {
	in := z.internals.Clone()
	existing, ok := core.GlobalRegistry.Get(z)
	if !ok {
		existing = core.GlobalMeta{}
	}
	existing.Description = desc
	clone := z.withInternals(in)
	core.GlobalRegistry.Add(clone, existing)
	return clone
}

// =============================================================================
// VALIDATION METHODS
// =============================================================================

// Min sets minimum file size validation.
func (z *ZodFile[T, R]) Min(minimum int64, params ...any) *ZodFile[T, R] {
	return z.withCheck(checks.MinFileSize(minimum, params...))
}

// Max sets maximum file size validation.
func (z *ZodFile[T, R]) Max(maximum int64, params ...any) *ZodFile[T, R] {
	return z.withCheck(checks.MaxFileSize(maximum, params...))
}

// Size sets exact file size validation.
func (z *ZodFile[T, R]) Size(expected int64, params ...any) *ZodFile[T, R] {
	return z.withCheck(checks.FileSize(expected, params...))
}

// Mime sets MIME type validation.
func (z *ZodFile[T, R]) Mime(mimeTypes []string, params ...any) *ZodFile[T, R] {
	return z.withCheck(checks.Mime(mimeTypes, params...))
}

// =============================================================================
// REFINEMENT METHODS
// =============================================================================

// Refine applies a custom validation function matching the schema's output type R.
func (z *ZodFile[T, R]) Refine(fn func(R) bool, params ...any) *ZodFile[T, R] {
	wrapper := func(v any) bool {
		if v == nil {
			return true
		}
		if val, ok := convertToFileType[T, R](v); ok {
			return fn(val)
		}
		return false
	}

	sp := utils.NormalizeParams(params...)
	var msg any
	if sp.Error != nil {
		msg = sp.Error
	}

	check := checks.NewCustom[any](wrapper, msg)
	return z.withCheck(check)
}

// RefineAny applies a custom validation function that receives the raw value.
func (z *ZodFile[T, R]) RefineAny(fn func(any) bool, params ...any) *ZodFile[T, R] {
	sp := utils.NormalizeParams(params...)
	var msg any
	if sp.Error != nil {
		msg = sp.Error
	}
	check := checks.NewCustom[any](fn, msg)
	return z.withCheck(check)
}

// Check adds a custom validation function that can push multiple issues.
func (z *ZodFile[T, R]) Check(fn func(value R, payload *core.ParsePayload), params ...any) *ZodFile[T, R] {
	wrapper := func(payload *core.ParsePayload) {
		if val, ok := payload.Value().(R); ok {
			fn(val, payload)
		}
	}
	check := checks.NewCustom[any](wrapper, utils.NormalizeCustomParams(params...))
	return z.withCheck(check)
}

// With is an alias for Check (Zod v4 API compatibility).
func (z *ZodFile[T, R]) With(fn func(value R, payload *core.ParsePayload), params ...any) *ZodFile[T, R] {
	return z.Check(fn, params...)
}

// =============================================================================
// TRANSFORMATION AND PIPELINE METHODS
// =============================================================================

// Transform applies a transformation function to the parsed value.
func (z *ZodFile[T, R]) Transform(fn func(T, *core.RefinementContext) (any, error)) *core.ZodTransform[R, any] {
	wrapperFn := func(input R, ctx *core.RefinementContext) (any, error) {
		baseValue := extractFileValue[T, R](input)
		return fn(baseValue, ctx)
	}
	return core.NewZodTransform[R, any](z, wrapperFn)
}

// Overwrite transforms the input value while preserving the original type.
func (z *ZodFile[T, R]) Overwrite(transform func(R) R, params ...any) *ZodFile[T, R] {
	transformAny := func(input any) any {
		converted, ok := convertToFileType[T, R](input)
		if !ok {
			return input
		}
		return transform(converted)
	}
	check := checks.NewZodCheckOverwrite(transformAny, params...)
	return z.withCheck(check)
}

// Pipe creates a validation pipeline with another schema.
func (z *ZodFile[T, R]) Pipe(target core.ZodType[any]) *core.ZodPipe[R, any] {
	wrapperFn := func(input R, ctx *core.ParseContext) (any, error) {
		baseValue := extractFileValue[T, R](input)
		return target.Parse(baseValue, ctx)
	}
	return core.NewZodPipe[R, any](z, target, wrapperFn)
}

// =============================================================================
// COMPOSITION METHODS
// =============================================================================

// And creates an intersection with another schema.
func (z *ZodFile[T, R]) And(other any) *ZodIntersection[any, any] {
	return Intersection(z, other)
}

// Or creates a union with another schema.
func (z *ZodFile[T, R]) Or(other any) *ZodUnion[any, any] {
	return Union([]any{z, other})
}

// =============================================================================
// HELPER AND PRIVATE METHODS
// =============================================================================

// expectedType returns the schema's type code, defaulting to ZodTypeFile.
func (z *ZodFile[T, R]) expectedType() core.ZodTypeCode {
	if z.internals.Type != "" {
		return z.internals.Type
	}
	return core.ZodTypeFile
}

// withCheck clones internals, adds a check, and returns a new schema (Copy-on-Write).
func (z *ZodFile[T, R]) withCheck(check core.ZodCheck) *ZodFile[T, R] {
	in := z.internals.Clone()
	in.AddCheck(check)
	return z.withInternals(in)
}

// withPtrInternals creates a new pointer-constraint schema from cloned internals.
func (z *ZodFile[T, R]) withPtrInternals(in *core.ZodTypeInternals) *ZodFile[T, *T] {
	return &ZodFile[T, *T]{internals: &ZodFileInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
	}}
}

// withInternals creates a new schema preserving generic type parameters.
func (z *ZodFile[T, R]) withInternals(in *core.ZodTypeInternals) *ZodFile[T, R] {
	return &ZodFile[T, R]{internals: &ZodFileInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
	}}
}

// CloneFrom copies configuration from another schema of the same type.
func (z *ZodFile[T, R]) CloneFrom(source any) {
	if src, ok := source.(*ZodFile[T, R]); ok {
		z.internals = src.internals
	}
}

// =============================================================================
// TYPE CONVERSION HELPERS
// =============================================================================

// convertFileResult converts engine output to the constraint type R,
// handling pointer wrapping for optional/nilable schemas.
func convertFileResult[T any, R any](result any) (R, error) {
	switch v := result.(type) {
	case *any:
		if v != nil {
			zeroType := reflect.TypeFor[R]()
			if zeroType.Kind() == reflect.Pointer {
				if converted, ok := any(v).(R); ok {
					return converted, nil
				}
				if zeroType == reflect.TypeFor[*any]() {
					return any(v).(R), nil
				}
			}
			return convertToFileConstraintType[T, R](*v), nil
		}
		var zero R
		return zero, nil
	case nil:
		var zero R
		return zero, nil
	default:
		return convertToFileConstraintType[T, R](result), nil
	}
}

// convertToFileConstraintType converts any value to constraint type R.
func convertToFileConstraintType[T any, R any](value any) R {
	var zero R

	// For any type (R = any), return the value directly.
	if reflect.TypeFor[R]() == reflect.TypeFor[any]() {
		return any(value).(R) //nolint:unconvert // Required for generic type constraint conversion
	}

	if typedValue, ok := value.(R); ok {
		return typedValue
	}

	if ptr, ok := value.(*R); ok && ptr != nil {
		return *ptr
	}

	// Handle conversion from T to *T (for FilePtr cases).
	if reflect.TypeFor[R]().Kind() == reflect.Pointer {
		if value != nil {
			valuePtr := &value
			if converted, ok := any(valuePtr).(R); ok {
				return converted
			}
		}
	}

	return zero
}

// extractFileValue extracts the base type T from constraint type R.
func extractFileValue[T any, R any](value R) T {
	if typedValue, ok := any(value).(T); ok {
		return typedValue
	}

	if ptr, ok := any(value).(*T); ok && ptr != nil {
		return *ptr
	}

	var zero T
	return zero
}

// convertToFileType converts any value to constraint type R with success flag.
func convertToFileType[T any, R any](v any) (R, bool) {
	if typedValue, ok := v.(R); ok {
		return typedValue, true
	}

	if ptr, ok := v.(*R); ok && ptr != nil {
		return *ptr, true
	}

	if baseValue, ok := v.(T); ok {
		if converted, ok := any(baseValue).(R); ok {
			return converted, true
		}
	}

	if ptr, ok := v.(*T); ok && ptr != nil {
		if converted, ok := any(*ptr).(R); ok {
			return converted, true
		}
	}

	var zero R
	return zero, false
}

// =============================================================================
// FILE EXTRACTION HELPERS
// =============================================================================

// extractFile recognizes supported file types from input.
func extractFile(v any) (any, bool) {
	switch file := v.(type) {
	case *os.File:
		return file, true
	case *multipart.FileHeader:
		return file, true
	case multipart.File:
		return file, true
	default:
		return nil, false
	}
}

// extractFileForEngine extracts a file value from input for engine.ParseComplex.
func (z *ZodFile[T, R]) extractFileForEngine(input any) (any, bool) {
	return extractFile(input)
}

// extractFilePtrForEngine extracts a pointer to file from input for engine.ParseComplex.
func (z *ZodFile[T, R]) extractFilePtrForEngine(input any) (*any, bool) {
	if ptr, ok := input.(*any); ok {
		return ptr, true
	}

	result, ok := extractFile(input)
	if !ok {
		return nil, false
	}
	return &result, true
}

// validateFileForEngine validates a file value against checks.
func (z *ZodFile[T, R]) validateFileForEngine(value any, cs []core.ZodCheck, ctx *core.ParseContext) (any, error) {
	return engine.ApplyChecks[any](value, cs, ctx)
}

// =============================================================================
// PUBLIC CONSTRUCTORS
// =============================================================================

// newZodFileFromDef constructs a new ZodFile from a definition.
func newZodFileFromDef[T any, R any](def *ZodFileDef) *ZodFile[T, R] {
	internals := &ZodFileInternals{
		ZodTypeInternals: engine.NewBaseZodTypeInternals(def.Type),
		Def:              def,
	}

	internals.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any] {
		fileDef := &ZodFileDef{ZodTypeDef: *newDef}
		return any(newZodFileFromDef[T, R](fileDef)).(core.ZodType[any])
	}

	schema := &ZodFile[T, R]{internals: internals}

	if def.Error != nil {
		internals.Error = def.Error
	}

	for _, check := range def.Checks {
		internals.AddCheck(check)
	}

	return schema
}

// newFileDef creates a ZodFileDef from optional params.
func newFileDef(params ...any) *ZodFileDef {
	sp := utils.NormalizeParams(utils.FirstParam(params...))

	def := &ZodFileDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeFile,
			Checks: []core.ZodCheck{},
		},
	}

	if sp != nil {
		utils.ApplySchemaParams(&def.ZodTypeDef, sp)
	}

	return def
}

// File creates a new file validation schema.
func File(params ...any) *ZodFile[any, any] {
	return newZodFileFromDef[any, any](newFileDef(params...))
}

// FilePtr creates a new file validation schema with pointer constraint.
func FilePtr(params ...any) *ZodFile[any, *any] {
	return newZodFileFromDef[any, *any](newFileDef(params...))
}

// FileTyped creates a new file schema with specific constraint types.
func FileTyped[T any, R any](params ...any) *ZodFile[T, R] {
	return newZodFileFromDef[T, R](newFileDef(params...))
}

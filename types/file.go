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

// ZodFileDef represents the definition for a file type
type ZodFileDef struct {
	core.ZodTypeDef
}

// ZodFileInternals holds the internal state for file validation
type ZodFileInternals struct {
	core.ZodTypeInternals
	Def *ZodFileDef // Schema definition
}

// ZodFile represents a file validation schema with constraint types T and R
type ZodFile[T any, R any] struct {
	internals *ZodFileInternals
}

// =============================================================================
// CORE INTERFACE METHODS
// =============================================================================

// GetInternals returns the internal state for interface compatibility
func (z *ZodFile[T, R]) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// IsOptional returns whether the file type is optional
func (z *ZodFile[T, R]) IsOptional() bool {
	return z.internals.Optional
}

// IsNilable returns whether the file type is nilable
func (z *ZodFile[T, R]) IsNilable() bool {
	return z.internals.Nilable
}

// =============================================================================
// PARSING METHODS
// =============================================================================

// Parse returns a validated file using direct validation for better performance
func (z *ZodFile[T, R]) Parse(input any, ctx ...*core.ParseContext) (R, error) {
	result, err := engine.ParseComplex(
		input,
		&z.internals.ZodTypeInternals,
		core.ZodTypeFile,
		z.extractFileForEngine,
		z.extractFilePtrForEngine,
		z.validateFileForEngine,
		ctx...,
	)
	if err != nil {
		var zero R
		return zero, err
	}

	// Handle different return types
	switch v := result.(type) {
	case *any:
		if v != nil {
			// Check if R is itself a pointer type (*any, etc.)
			var zero R
			zeroType := reflect.TypeOf(zero)
			if zeroType != nil && zeroType.Kind() == reflect.Ptr {
				// R is a pointer type, try to convert the entire *any to R first
				if converted, ok := any(v).(R); ok {
					return converted, nil
				}
				// If direct conversion fails, check if R is specifically *any
				if zeroType == reflect.TypeOf((*any)(nil)) {
					// R is *any, return v directly to preserve Overwrite transformations
					return any(v).(R), nil
				}
			}
			// R is not a pointer type, dereference and convert
			return convertToFileConstraintType[T, R](*v), nil
		}
		// Return zero value for nil pointer
		var zero R
		return zero, nil
	case nil:
		// Return zero value for nil result
		var zero R
		return zero, nil
	default:
		// Handle all other types including concrete types like *multipart.FileHeader
		return convertToFileConstraintType[T, R](result), nil
	}
}

// MustParse is the variant that panics on error
func (z *ZodFile[T, R]) MustParse(input any, ctx ...*core.ParseContext) R {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// StrictParse validates input with compile-time type safety and enhanced performance.
// This method provides zero-overhead abstraction with strict type constraints.
func (z *ZodFile[T, R]) StrictParse(input T, ctx ...*core.ParseContext) (R, error) {
	result, err := engine.ParseComplexStrict(
		input,
		&z.internals.ZodTypeInternals,
		core.ZodTypeFile,
		z.extractFileForEngine,
		z.extractFilePtrForEngine,
		z.validateFileForEngine,
		ctx...,
	)
	if err != nil {
		var zero R
		return zero, err
	}

	// Convert result to constraint type R
	return convertToFileConstraintType[T, R](result), nil
}

// MustStrictParse validates input with compile-time type safety and panics on failure.
// This method provides zero-overhead abstraction with strict type constraints.
func (z *ZodFile[T, R]) MustStrictParse(input T, ctx ...*core.ParseContext) R {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// ParseAny validates the input value and returns any type (for runtime interface)
func (z *ZodFile[T, R]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// =============================================================================
// MODIFIER METHODS
// =============================================================================

// Optional makes the file type optional and returns pointer constraint
func (z *ZodFile[T, R]) Optional() *ZodFile[T, *T] {
	in := z.internals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// Nilable makes the file type nilable and returns pointer constraint
func (z *ZodFile[T, R]) Nilable() *ZodFile[T, *T] {
	in := z.internals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish makes the file type both optional and nilable
func (z *ZodFile[T, R]) Nullish() *ZodFile[T, *T] {
	in := z.internals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Default sets a default value for the file type
func (z *ZodFile[T, R]) Default(v T) *ZodFile[T, R] {
	in := z.internals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

// DefaultFunc sets a default function for the file type
func (z *ZodFile[T, R]) DefaultFunc(fn func() T) *ZodFile[T, R] {
	in := z.internals.Clone()
	in.SetDefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Prefault sets a prefault value for the file type
func (z *ZodFile[T, R]) Prefault(v T) *ZodFile[T, R] {
	in := z.internals.Clone()
	in.SetPrefaultValue(v)
	return z.withInternals(in)
}

// PrefaultFunc sets a prefault function for the file type
func (z *ZodFile[T, R]) PrefaultFunc(fn func() T) *ZodFile[T, R] {
	in := z.internals.Clone()
	in.SetPrefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Meta adds metadata to the file type
func (z *ZodFile[T, R]) Meta(meta core.GlobalMeta) *ZodFile[T, R] {
	core.GlobalRegistry.Add(z, meta)
	return z
}

// Describe registers a description in the global registry.
// TypeScript Zod v4 equivalent: schema.describe(description)
func (z *ZodFile[T, R]) Describe(description string) *ZodFile[T, R] {
	newInternals := z.internals.Clone()
	existing, ok := core.GlobalRegistry.Get(z)
	if !ok {
		existing = core.GlobalMeta{}
	}
	existing.Description = description
	clone := z.withInternals(newInternals)
	core.GlobalRegistry.Add(clone, existing)
	return clone
}

// =============================================================================
// VALIDATION METHODS
// =============================================================================

// Min sets minimum file size validation
func (z *ZodFile[T, R]) Min(minimum int64, params ...any) *ZodFile[T, R] {
	check := checks.MinFileSize(minimum, params...)
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Max sets maximum file size validation
func (z *ZodFile[T, R]) Max(maximum int64, params ...any) *ZodFile[T, R] {
	check := checks.MaxFileSize(maximum, params...)
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Size sets exact file size validation
func (z *ZodFile[T, R]) Size(expected int64, params ...any) *ZodFile[T, R] {
	check := checks.FileSize(expected, params...)
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Mime sets MIME type validation
func (z *ZodFile[T, R]) Mime(mimeTypes []string, params ...any) *ZodFile[T, R] {
	check := checks.Mime(mimeTypes, params...)
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Refine adds custom validation logic
func (z *ZodFile[T, R]) Refine(fn func(R) bool, params ...any) *ZodFile[T, R] {
	wrapper := func(v any) bool {
		if v == nil {
			return true // Skip refinement for nil values
		}
		// Convert any to constraint type R for type-safe refinement
		if constraintVal, ok := convertToFileType[T, R](v); ok {
			return fn(constraintVal)
		}
		return false
	}

	// Use unified parameter handling
	param := utils.FirstParam(params...)
	customParams := utils.NormalizeCustomParams(param)
	check := checks.NewCustom[any](wrapper, customParams)
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// RefineAny adds custom validation logic with any type
func (z *ZodFile[T, R]) RefineAny(fn func(any) bool, params ...any) *ZodFile[T, R] {
	// Use unified parameter handling
	param := utils.FirstParam(params...)
	customParams := utils.NormalizeCustomParams(param)
	check := checks.NewCustom[any](fn, customParams)
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// =============================================================================
// TRANSFORMATION METHODS
// =============================================================================

// Transform creates a transformation pipeline
func (z *ZodFile[T, R]) Transform(fn func(T, *core.RefinementContext) (any, error)) *core.ZodTransform[R, any] {
	wrapperFn := func(input R, ctx *core.RefinementContext) (any, error) {
		baseValue := extractFileValue[T, R](input)
		return fn(baseValue, ctx)
	}
	return core.NewZodTransform[R, any](z, wrapperFn)
}

// Overwrite applies a transformation function to the file value
func (z *ZodFile[T, R]) Overwrite(transform func(R) R, params ...any) *ZodFile[T, R] {
	// Create a transformation function that works with the constraint type R
	transformAny := func(input any) any {
		// Try to convert input to constraint type R
		converted, ok := convertToFileType[T, R](input)
		if !ok {
			// If conversion fails, return original value
			return input
		}

		// Apply transformation directly on constraint type R
		return transform(converted)
	}

	check := checks.NewZodCheckOverwrite(transformAny, params...)
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Pipe creates a pipeline to another schema
func (z *ZodFile[T, R]) Pipe(target core.ZodType[any]) *core.ZodPipe[R, any] {
	wrapperFn := func(input R, ctx *core.ParseContext) (any, error) {
		baseValue := extractFileValue[T, R](input)
		return target.Parse(baseValue, ctx)
	}
	return core.NewZodPipe[R, any](z, target, wrapperFn)
}

// =============================================================================
// INTERNAL HELPER METHODS
// =============================================================================

// withPtrInternals creates a new instance with pointer constraint type
func (z *ZodFile[T, R]) withPtrInternals(in *core.ZodTypeInternals) *ZodFile[T, *T] {
	return &ZodFile[T, *T]{
		internals: &ZodFileInternals{
			ZodTypeInternals: *in,
			Def:              z.internals.Def,
		},
	}
}

func (z *ZodFile[T, R]) withInternals(in *core.ZodTypeInternals) *ZodFile[T, R] {
	return &ZodFile[T, R]{
		internals: &ZodFileInternals{
			ZodTypeInternals: *in,
			Def:              z.internals.Def,
		},
	}
}

// CloneFrom copies configuration from another schema
func (z *ZodFile[T, R]) CloneFrom(source any) {
	if src, ok := source.(*ZodFile[T, R]); ok {
		z.internals = src.internals
	}
}

// =============================================================================
// TYPE CONVERSION UTILITIES
// =============================================================================

// convertToFileConstraintType converts any value to constraint type R
func convertToFileConstraintType[T any, R any](value any) R {
	var zero R

	// For any type (R = any), return the value directly
	// This handles File().Prefault() case where R is any
	if reflect.TypeOf(zero) == reflect.TypeOf((*any)(nil)).Elem() {
		return any(value).(R) //nolint:unconvert // Required for generic type constraint conversion
	}

	// Direct type conversion - try this first
	if typedValue, ok := value.(R); ok {
		return typedValue
	}

	// Try pointer conversion (dereference)
	if ptr, ok := value.(*R); ok && ptr != nil {
		return *ptr
	}

	// Handle conversion from T to *T (for FilePtr cases)
	if reflect.TypeOf(zero).Kind() == reflect.Ptr {
		// R is a pointer type, try to create pointer to value
		if value != nil {
			// Create a pointer to the value
			valuePtr := &value
			if converted, ok := any(valuePtr).(R); ok {
				return converted
			}
		}
	}

	// Return zero value if conversion fails
	return zero
}

// extractFileValue extracts the base type T from constraint type R
func extractFileValue[T any, R any](value R) T {
	if typedValue, ok := any(value).(T); ok {
		return typedValue
	}

	// Try pointer dereferencing
	if ptr, ok := any(value).(*T); ok && ptr != nil {
		return *ptr
	}

	// Return zero value if extraction fails
	var zero T
	return zero
}

// convertToFileType converts any value to constraint type R with success flag
func convertToFileType[T any, R any](v any) (R, bool) {
	if typedValue, ok := v.(R); ok {
		return typedValue, true
	}

	// Try pointer conversion
	if ptr, ok := v.(*R); ok && ptr != nil {
		return *ptr, true
	}

	// Try converting through T
	if baseValue, ok := v.(T); ok {
		if converted, ok := any(baseValue).(R); ok {
			return converted, true
		}
	}

	// Try pointer to T conversion
	if ptr, ok := v.(*T); ok && ptr != nil {
		if converted, ok := any(*ptr).(R); ok {
			return converted, true
		}
	}

	var zero R
	return zero, false
}

// =============================================================================
// FILE EXTRACTION UTILITIES
// =============================================================================

// extractFile extracts file-like values from various input types
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

// =============================================================================
// CONSTRUCTOR FUNCTIONS
// =============================================================================

// newZodFileFromDef creates a new ZodFile instance from a definition
func newZodFileFromDef[T any, R any](def *ZodFileDef) *ZodFile[T, R] {
	internals := &ZodFileInternals{
		ZodTypeInternals: engine.NewBaseZodTypeInternals(def.Type),
		Def:              def,
	}

	// Provide constructor for AddCheck functionality
	internals.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any] {
		fileDef := &ZodFileDef{
			ZodTypeDef: *newDef,
		}
		return any(newZodFileFromDef[T, R](fileDef)).(core.ZodType[any])
	}

	schema := &ZodFile[T, R]{internals: internals}

	// Set error if provided
	if def.Error != nil {
		internals.Error = def.Error
	}

	// Set checks if provided
	if len(def.Checks) > 0 {
		for _, check := range def.Checks {
			internals.AddCheck(check)
		}
	}

	return schema
}

// =============================================================================
// PUBLIC CONSTRUCTOR FUNCTIONS
// =============================================================================

// File creates a new file schema
func File(params ...any) *ZodFile[any, any] {
	param := utils.FirstParam(params...)
	normalizedParams := utils.NormalizeParams(param)

	def := &ZodFileDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeFile,
			Checks: []core.ZodCheck{},
		},
	}

	// Use utils.ApplySchemaParams for consistent parameter handling
	if normalizedParams != nil {
		utils.ApplySchemaParams(&def.ZodTypeDef, normalizedParams)
	}

	return newZodFileFromDef[any, any](def)
}

// FilePtr creates a new file schema with pointer constraint
func FilePtr(params ...any) *ZodFile[any, *any] {
	param := utils.FirstParam(params...)
	normalizedParams := utils.NormalizeParams(param)

	def := &ZodFileDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeFile,
			Checks: []core.ZodCheck{},
		},
	}

	// Use utils.ApplySchemaParams for consistent parameter handling
	if normalizedParams != nil {
		utils.ApplySchemaParams(&def.ZodTypeDef, normalizedParams)
	}

	return newZodFileFromDef[any, *any](def)
}

// FileTyped creates a new file schema with specific constraint types
func FileTyped[T any, R any](params ...any) *ZodFile[T, R] {
	param := utils.FirstParam(params...)
	normalizedParams := utils.NormalizeParams(param)

	def := &ZodFileDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeFile,
			Checks: []core.ZodCheck{},
		},
	}

	// Use utils.ApplySchemaParams for consistent parameter handling
	if normalizedParams != nil {
		utils.ApplySchemaParams(&def.ZodTypeDef, normalizedParams)
	}

	return newZodFileFromDef[T, R](def)
}

// extractFileForEngine extracts any from input for engine.ParseComplex
func (z *ZodFile[T, R]) extractFileForEngine(input any) (any, bool) {
	return extractFile(input)
}

// extractFilePtrForEngine extracts pointer to any from input for engine.ParseComplex
func (z *ZodFile[T, R]) extractFilePtrForEngine(input any) (*any, bool) {
	// Try direct pointer extraction
	if ptr, ok := input.(*any); ok {
		return ptr, true
	}

	// Try extracting file and return pointer to it
	result, ok := extractFile(input)
	if !ok {
		return nil, false
	}
	return &result, true
}

// validateFileForEngine validates any for engine.ParseComplex
func (z *ZodFile[T, R]) validateFileForEngine(value any, checks []core.ZodCheck, ctx *core.ParseContext) (any, error) {
	return engine.ApplyChecks[any](value, checks, ctx)
}

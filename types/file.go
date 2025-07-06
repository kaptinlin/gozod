package types

import (
	"mime/multipart"
	"os"
	"reflect"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/internal/utils"
)

// =============================================================================
// TYPE DEFINITIONS
// =============================================================================

// ZodFileDef defines the configuration for file validation
type ZodFileDef struct {
	core.ZodTypeDef
}

// ZodFileInternals contains file validator internal state
type ZodFileInternals struct {
	core.ZodTypeInternals
	Def *ZodFileDef // Schema definition
}

// ZodFile represents a file validation schema with dual generic parameters
// T = base type (any), R = constraint type (any or *any)
type ZodFile[T any, R any] struct {
	internals *ZodFileInternals
}

// =============================================================================
// CORE METHODS
// =============================================================================

// GetInternals returns the internal state of the schema
func (z *ZodFile[T, R]) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// IsOptional returns true if this schema accepts undefined/missing values
func (z *ZodFile[T, R]) IsOptional() bool {
	return z.internals.ZodTypeInternals.IsOptional()
}

// IsNilable returns true if this schema accepts nil values
func (z *ZodFile[T, R]) IsNilable() bool {
	return z.internals.ZodTypeInternals.IsNilable()
}

// Parse returns a validated file using direct validation for better performance
func (z *ZodFile[T, R]) Parse(input any, ctx ...*core.ParseContext) (R, error) {
	var parseCtx *core.ParseContext
	if len(ctx) > 0 && ctx[0] != nil {
		parseCtx = ctx[0]
	} else {
		parseCtx = &core.ParseContext{}
	}

	// Handle nil input
	if input == nil {
		// Check if nil is allowed (optional/nilable)
		if z.internals.ZodTypeInternals.Optional || z.internals.ZodTypeInternals.Nilable {
			var zero R
			return zero, nil
		}

		// Try default values
		if z.internals.ZodTypeInternals.DefaultValue != nil {
			return z.Parse(z.internals.ZodTypeInternals.DefaultValue, parseCtx)
		}
		if z.internals.ZodTypeInternals.DefaultFunc != nil {
			defaultValue := z.internals.ZodTypeInternals.DefaultFunc()
			return z.Parse(defaultValue, parseCtx)
		}

		// File type expects a file, nil is invalid
		rawIssue := issues.CreateInvalidTypeIssue(core.ZodTypeFile, input)
		finalIssue := issues.FinalizeIssue(rawIssue, parseCtx, core.GetConfig())
		return *new(R), issues.NewZodError([]core.ZodIssue{finalIssue})
	}

	// Validate that the value is a valid file type
	if _, ok := extractFile(input); !ok {
		// Try prefault on type error
		if z.internals.ZodTypeInternals.PrefaultValue != nil {
			return z.Parse(z.internals.ZodTypeInternals.PrefaultValue, parseCtx)
		}
		if z.internals.ZodTypeInternals.PrefaultFunc != nil {
			prefaultValue := z.internals.ZodTypeInternals.PrefaultFunc()
			return z.Parse(prefaultValue, parseCtx)
		}

		// Invalid file type - create appropriate error
		rawIssue := issues.CreateInvalidTypeIssue(core.ZodTypeFile, input)
		finalIssue := issues.FinalizeIssue(rawIssue, parseCtx, core.GetConfig())
		return *new(R), issues.NewZodError([]core.ZodIssue{finalIssue})
	}

	// Run validation checks if any exist and capture transformed result
	if len(z.internals.ZodTypeInternals.Checks) > 0 {
		transformedInput, err := engine.ApplyChecks[any](input, z.internals.ZodTypeInternals.Checks, parseCtx)
		if err != nil {
			// Try prefault on validation failure
			if z.internals.ZodTypeInternals.PrefaultValue != nil {
				return z.Parse(z.internals.ZodTypeInternals.PrefaultValue, parseCtx)
			}
			if z.internals.ZodTypeInternals.PrefaultFunc != nil {
				prefaultValue := z.internals.ZodTypeInternals.PrefaultFunc()
				return z.Parse(prefaultValue, parseCtx)
			}

			return *new(R), err
		}
		input = transformedInput
	}

	// Apply transform if present
	if z.internals.ZodTypeInternals.Transform != nil {
		refCtx := &core.RefinementContext{ParseContext: parseCtx}
		result, err := z.internals.ZodTypeInternals.Transform(input, refCtx)
		if err != nil {
			return *new(R), err
		}
		return convertToFileConstraintType[T, R](result), nil
	}

	// Convert result to constraint type R and return
	return convertToFileConstraintType[T, R](input), nil
}

// MustParse is the variant that panics on error
func (z *ZodFile[T, R]) MustParse(input any, ctx ...*core.ParseContext) R {
	result, err := z.Parse(input, ctx...)
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
	in := z.internals.ZodTypeInternals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// Nilable makes the file type nilable and returns pointer constraint
func (z *ZodFile[T, R]) Nilable() *ZodFile[T, *T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish combines optional and nilable modifiers
func (z *ZodFile[T, R]) Nullish() *ZodFile[T, *T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Default preserves current constraint type R
func (z *ZodFile[T, R]) Default(v T) *ZodFile[T, R] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

// DefaultFunc preserves current constraint type R
func (z *ZodFile[T, R]) DefaultFunc(fn func() T) *ZodFile[T, R] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetDefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Prefault provides fallback values on validation failure
func (z *ZodFile[T, R]) Prefault(v T) *ZodFile[T, R] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetPrefaultValue(v)
	return z.withInternals(in)
}

// PrefaultFunc provides dynamic fallback values
func (z *ZodFile[T, R]) PrefaultFunc(fn func() T) *ZodFile[T, R] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetPrefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// =============================================================================
// VALIDATION METHODS
// =============================================================================

// Min validates minimum file size in bytes
func (z *ZodFile[T, R]) Min(minimum int64, params ...any) *ZodFile[T, R] {
	check := checks.MinFileSize(minimum, params...)
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Max validates maximum file size in bytes
func (z *ZodFile[T, R]) Max(maximum int64, params ...any) *ZodFile[T, R] {
	check := checks.MaxFileSize(maximum, params...)
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Size validates exact file size in bytes
func (z *ZodFile[T, R]) Size(expected int64, params ...any) *ZodFile[T, R] {
	check := checks.FileSize(expected, params...)
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Mime validates allowed MIME types
func (z *ZodFile[T, R]) Mime(mimeTypes []string, params ...any) *ZodFile[T, R] {
	check := checks.Mime(mimeTypes, params...)
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Refine applies type-safe validation with constraint type R
func (z *ZodFile[T, R]) Refine(fn func(R) bool, args ...any) *ZodFile[T, R] {
	// Wrapper converts the raw value to R before calling fn
	wrapper := func(v any) bool {
		// Convert value to constraint type R and call the refinement function
		if constraintValue, ok := convertToFileConstraintValue[T, R](v); ok {
			return fn(constraintValue)
		}
		return false
	}

	// Use unified parameter handling
	param := utils.GetFirstParam(args...)
	normalizedParams := utils.NormalizeParams(param)

	var checkParams any
	if normalizedParams.Error != nil {
		checkParams = normalizedParams
	}

	check := checks.NewCustom[any](wrapper, checkParams)
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// RefineAny provides flexible validation without type conversion
func (z *ZodFile[T, R]) RefineAny(fn func(any) bool, args ...any) *ZodFile[T, R] {
	// Use unified parameter handling
	param := utils.GetFirstParam(args...)
	normalizedParams := utils.NormalizeParams(param)

	var checkParams any
	if normalizedParams.Error != nil {
		checkParams = normalizedParams
	}

	check := checks.NewCustom[any](fn, checkParams)
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// =============================================================================
// TRANSFORMATION AND PIPELINE METHODS
// =============================================================================

// Transform creates a type-safe transformation using WrapFn pattern
func (z *ZodFile[T, R]) Transform(fn func(T, *core.RefinementContext) (any, error)) *core.ZodTransform[R, any] {
	wrapperFn := func(input R, ctx *core.RefinementContext) (any, error) {
		baseValue := extractFileValue[T, R](input)
		return fn(baseValue, ctx)
	}
	return core.NewZodTransform[R, any](z, wrapperFn)
}

// Overwrite transforms the input value while preserving the original type.
// Unlike Transform, this method doesn't change the inferred type and returns an instance of the original class.
// The transformation function is stored as a check, so it doesn't modify the inferred type.
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
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Pipe creates a pipeline using WrapFn pattern
func (z *ZodFile[T, R]) Pipe(target core.ZodType[any]) *core.ZodPipe[R, any] {
	wrapperFn := func(input R, ctx *core.ParseContext) (any, error) {
		baseValue := extractFileValue[T, R](input)
		return target.Parse(baseValue, ctx)
	}
	return core.NewZodPipe[R, any](z, wrapperFn)
}

// =============================================================================
// HELPER METHODS
// =============================================================================

func (z *ZodFile[T, R]) withPtrInternals(in *core.ZodTypeInternals) *ZodFile[T, *T] {
	return &ZodFile[T, *T]{internals: &ZodFileInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
	}}
}

func (z *ZodFile[T, R]) withInternals(in *core.ZodTypeInternals) *ZodFile[T, R] {
	return &ZodFile[T, R]{internals: &ZodFileInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
	}}
}

// CloneFrom copies configuration from another schema
func (z *ZodFile[T, R]) CloneFrom(source any) {
	if src, ok := source.(*ZodFile[T, R]); ok {
		z.internals = src.internals
	}
}

// =============================================================================
// TYPE CONVERSION HELPERS
// =============================================================================

// convertToFileConstraintType converts a base type T to constraint type R
func convertToFileConstraintType[T any, R any](value any) R {
	var zero R
	switch any(zero).(type) {
	case *any:
		// Need to return *any from any
		if value != nil {
			valueCopy := value
			return any(&valueCopy).(R)
		}
		return any((*any)(nil)).(R)
	default:
		// Return value directly as R
		return any(value).(R)
	}
}

// extractFileValue extracts the base type T from constraint type R
func extractFileValue[T any, R any](value R) T {
	// Handle direct assignment (when T == R)
	if directValue, ok := any(value).(T); ok {
		return directValue
	}

	// Handle pointer dereferencing
	if ptrValue := reflect.ValueOf(value); ptrValue.Kind() == reflect.Ptr && !ptrValue.IsNil() {
		if derefValue, ok := ptrValue.Elem().Interface().(T); ok {
			return derefValue
		}
	}

	// Fallback to zero value
	var zero T
	return zero
}

// convertToFileType converts any value to the file constraint type R with strict type checking
func convertToFileType[T any, R any](v any) (R, bool) {
	var zero R

	if v == nil {
		// Handle nil values for pointer types
		zeroType := reflect.TypeOf((*R)(nil)).Elem()
		if zeroType.Kind() == reflect.Ptr {
			return zero, true // zero value for pointer types is nil
		}
		return zero, false // nil not allowed for value types
	}

	// Check if input is a valid file type
	if _, isValidFile := extractFile(v); !isValidFile {
		return zero, false // Reject all non-file types
	}

	// Convert to target constraint type R
	zeroType := reflect.TypeOf((*R)(nil)).Elem()
	if zeroType.Kind() == reflect.Ptr {
		// R is *any - return pointer to the file
		if converted, ok := any(&v).(R); ok { //nolint:unconvert
			return converted, true
		}
	} else {
		// R is any - return the file directly
		if converted, ok := any(v).(R); ok { //nolint:unconvert
			return converted, true
		}
	}

	return zero, false
}

// convertToFileConstraintValue converts any value to constraint type R if possible
func convertToFileConstraintValue[T any, R any](value any) (R, bool) {
	var zero R

	// Direct type match
	if r, ok := any(value).(R); ok {
		return r, true
	}

	// Handle pointer conversion for file types
	if _, ok := any(zero).(*any); ok {
		// Need to convert any to *any
		if value != nil {
			valueCopy := value
			return any(&valueCopy).(R), true
		}
		return any((*any)(nil)).(R), true
	}

	return zero, false
}

// =============================================================================
// FILE TYPE DETECTION AND VALIDATION
// =============================================================================

// extractFile attempts to extract a file from various types
func extractFile(v any) (any, bool) {
	switch f := v.(type) {
	case *os.File:
		return f, true
	case os.File:
		return f, true
	case *multipart.FileHeader:
		return f, true
	case multipart.FileHeader:
		return f, true
	default:
		return nil, false
	}
}

// =============================================================================
// CONSTRUCTOR FUNCTIONS
// =============================================================================

// newZodFileFromDef constructs new ZodFile from definition
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
// FACTORY FUNCTIONS
// =============================================================================

// File creates file schema that accepts file values - returns value constraint
func File(params ...any) *ZodFile[any, any] {
	return FileTyped[any, any](params...)
}

// FilePtr creates file schema that accepts file values - returns pointer constraint
func FilePtr(params ...any) *ZodFile[any, *any] {
	return FileTyped[any, *any](params...)
}

// FileTyped creates typed file schema with generic constraints
func FileTyped[T any, R any](params ...any) *ZodFile[T, R] {
	param := utils.GetFirstParam(params...)
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

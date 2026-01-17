package types

import (
	"regexp"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/internal/utils"
	"github.com/kaptinlin/gozod/pkg/coerce"
	"github.com/kaptinlin/gozod/pkg/regex"
	"github.com/kaptinlin/gozod/pkg/validate"
)

// =============================================================================
// TYPE CONSTRAINTS
// =============================================================================

// NetworkConstraint defines the constraint for network address types (string-based)
type NetworkConstraint interface {
	string | *string
}

// =============================================================================
// IPv4 TYPE DEFINITIONS
// =============================================================================

// ZodIPv4Def defines the configuration for IPv4 address validation
type ZodIPv4Def struct {
	core.ZodTypeDef
}

// ZodIPv4Internals contains IPv4 validator internal state
type ZodIPv4Internals struct {
	core.ZodTypeInternals
	Def *ZodIPv4Def // Schema definition
}

// ZodIPv4 represents an IPv4 address validation schema
type ZodIPv4[T NetworkConstraint] struct {
	internals *ZodIPv4Internals
}

// =============================================================================
// IPv4 CORE METHODS
// =============================================================================

// GetInternals returns the internal state of the schema
func (z *ZodIPv4[T]) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// IsOptional returns true if this schema accepts undefined/missing values
func (z *ZodIPv4[T]) IsOptional() bool {
	return z.internals.IsOptional()
}

// IsNilable returns true if this schema accepts nil values
func (z *ZodIPv4[T]) IsNilable() bool {
	return z.internals.IsNilable()
}

// Coerce implements type conversion interface using coerce package
func (z *ZodIPv4[T]) Coerce(input any) (any, bool) {
	result, err := coerce.ToString(input)
	return result, err == nil
}

// Parse returns type-safe IPv4 address using unified engine API
func (z *ZodIPv4[T]) Parse(input any, ctx ...*core.ParseContext) (T, error) {
	return engine.ParsePrimitive[string, T](
		input,
		&z.internals.ZodTypeInternals,
		core.ZodTypeIPv4,
		engine.ApplyChecks[string], // Universal validator
		engine.ConvertToConstraintType[string, T], // Universal converter
		ctx...,
	)
}

// MustParse is the variant that panics on error
func (z *ZodIPv4[T]) MustParse(input any, ctx ...*core.ParseContext) T {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// StrictParse provides compile-time type safety by requiring exact type matching.
// This eliminates runtime type checking overhead for maximum performance.
// The input must exactly match the schema's constraint type T.
func (z *ZodIPv4[T]) StrictParse(input T, ctx ...*core.ParseContext) (T, error) {
	// Create validator that applies checks to the constraint type T
	validator := func(value string, checks []core.ZodCheck, c *core.ParseContext) (string, error) {
		return engine.ApplyChecks[string](value, checks, c)
	}

	return engine.ParsePrimitiveStrict[string, T](
		input,
		&z.internals.ZodTypeInternals,
		core.ZodTypeIPv4,
		validator,
		ctx...,
	)
}

// MustStrictParse is the strict mode variant that panics on error.
// Provides compile-time type safety with maximum performance.
func (z *ZodIPv4[T]) MustStrictParse(input T, ctx ...*core.ParseContext) T {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// ParseAny validates the input value and returns any type (for runtime interface)
func (z *ZodIPv4[T]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// =============================================================================
// IPv4 MODIFIER METHODS
// =============================================================================

// Optional allows nil values, returns pointer type for nullable semantics
func (z *ZodIPv4[T]) Optional() *ZodIPv4[*string] {
	in := z.internals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// Nilable allows nil values, returns pointer type
func (z *ZodIPv4[T]) Nilable() *ZodIPv4[*string] {
	in := z.internals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish combines optional and nilable modifiers
func (z *ZodIPv4[T]) Nullish() *ZodIPv4[*string] {
	in := z.internals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Default preserves current generic type T
func (z *ZodIPv4[T]) Default(v string) *ZodIPv4[T] {
	in := z.internals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

// DefaultFunc preserves current generic type T
func (z *ZodIPv4[T]) DefaultFunc(fn func() string) *ZodIPv4[T] {
	in := z.internals.Clone()
	in.SetDefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Prefault provides fallback values on validation failure
func (z *ZodIPv4[T]) Prefault(v string) *ZodIPv4[T] {
	in := z.internals.Clone()
	in.SetPrefaultValue(v)
	return z.withInternals(in)
}

// PrefaultFunc provides dynamic fallback values
func (z *ZodIPv4[T]) PrefaultFunc(fn func() string) *ZodIPv4[T] {
	in := z.internals.Clone()
	in.SetPrefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// =============================================================================
// IPv4 TRANSFORMATION AND PIPELINE METHODS
// =============================================================================

// Transform creates type-safe transformation using WrapFn pattern
func (z *ZodIPv4[T]) Transform(fn func(string, *core.RefinementContext) (any, error)) *core.ZodTransform[T, any] {
	wrapperFn := func(input T, ctx *core.RefinementContext) (any, error) {
		str := extractNetworkString(input)
		return fn(str, ctx)
	}
	return core.NewZodTransform[T, any](z, wrapperFn)
}

// Pipe creates a pipeline using WrapFn pattern
func (z *ZodIPv4[T]) Pipe(target core.ZodType[any]) *core.ZodPipe[T, any] {
	wrapperFn := func(input T, ctx *core.ParseContext) (any, error) {
		str := extractNetworkString(input)
		return target.Parse(str, ctx)
	}
	return core.NewZodPipe[T, any](z, target, wrapperFn)
}

// =============================================================================
// IPv4 REFINEMENT METHODS
// =============================================================================

// Refine adds type-safe custom validation logic
func (z *ZodIPv4[T]) Refine(fn func(T) bool, params ...any) *ZodIPv4[T] {
	wrapper := func(v any) bool {
		var zero T

		switch any(zero).(type) {
		case string:
			if v == nil {
				return false
			}
			if strVal, ok := v.(string); ok {
				return fn(any(strVal).(T))
			}
			return false
		case *string:
			if v == nil {
				// Call the refinement function with typed nil
				return fn(any((*string)(nil)).(T))
			}
			if strVal, ok := v.(string); ok {
				sCopy := strVal
				ptr := &sCopy
				return fn(any(ptr).(T))
			}
			return false
		default:
			return false
		}
	}

	schemaParams := utils.NormalizeParams(params...)
	var errorMessage any
	if schemaParams.Error != nil {
		errorMessage = schemaParams.Error // Pass the actual error message, not the SchemaParams
	}

	check := checks.NewCustom[any](wrapper, errorMessage)
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// RefineAny adds flexible custom validation logic
func (z *ZodIPv4[T]) RefineAny(fn func(any) bool, params ...any) *ZodIPv4[T] {
	schemaParams := utils.NormalizeParams(params...)
	var errorMessage any
	if schemaParams.Error != nil {
		errorMessage = schemaParams.Error // Pass the actual error message, not the SchemaParams
	}

	check := checks.NewCustom[any](fn, errorMessage)
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// =============================================================================
// IPv4 HELPER AND PRIVATE METHODS
// =============================================================================

// withPtrInternals creates new instance with pointer type
func (z *ZodIPv4[T]) withPtrInternals(in *core.ZodTypeInternals) *ZodIPv4[*string] {
	return &ZodIPv4[*string]{internals: &ZodIPv4Internals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
	}}
}

// withInternals creates new instance preserving generic type T
func (z *ZodIPv4[T]) withInternals(in *core.ZodTypeInternals) *ZodIPv4[T] {
	return &ZodIPv4[T]{internals: &ZodIPv4Internals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
	}}
}

// CloneFrom copies configuration from another schema
func (z *ZodIPv4[T]) CloneFrom(source any) {
	if src, ok := source.(*ZodIPv4[T]); ok {
		originalChecks := z.internals.Checks
		*z.internals = *src.internals
		z.internals.Checks = originalChecks
	}
}

// extractNetworkString extracts string value from generic network type T
func extractNetworkString[T NetworkConstraint](value T) string {
	if ptr, ok := any(value).(*string); ok {
		if ptr != nil {
			return *ptr
		}
		return ""
	}
	return any(value).(string)
}

// newZodIPv4FromDef constructs new ZodIPv4 from definition
func newZodIPv4FromDef[T NetworkConstraint](def *ZodIPv4Def) *ZodIPv4[T] {
	internals := &ZodIPv4Internals{
		ZodTypeInternals: core.ZodTypeInternals{
			Type:   def.Type,
			Checks: def.Checks,
			Coerce: def.Coerce,
			Bag:    make(map[string]any),
		},
		Def: def,
	}

	// Add IPv4 validation check
	ipv4Check := checks.IPv4()
	internals.Checks = append(internals.Checks, ipv4Check)

	// Provide constructor for AddCheck functionality
	internals.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any] {
		ipv4Def := &ZodIPv4Def{
			ZodTypeDef: *newDef,
		}
		return any(newZodIPv4FromDef[T](ipv4Def)).(core.ZodType[any])
	}

	if def.Error != nil {
		internals.Error = def.Error
	}

	return &ZodIPv4[T]{internals: internals}
}

// =============================================================================
// CONSTRUCTORS AND FACTORY FUNCTIONS
// =============================================================================

// IPv4 creates IPv4 address schema with type-inference support
func IPv4(params ...any) *ZodIPv4[string] {
	return IPv4Typed[string](params...)
}

// IPv4Ptr creates schema for *string IPv4
func IPv4Ptr(params ...any) *ZodIPv4[*string] {
	return IPv4Typed[*string](params...)
}

// IPv4Typed is the generic constructor for IPv4 address schemas
func IPv4Typed[T NetworkConstraint](params ...any) *ZodIPv4[T] {
	schemaParams := utils.NormalizeParams(params...)

	def := &ZodIPv4Def{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeIPv4,
			Checks: []core.ZodCheck{},
		},
	}

	// Apply normalized parameters to schema definition
	utils.ApplySchemaParams(&def.ZodTypeDef, schemaParams)

	return newZodIPv4FromDef[T](def)
}

// CoercedIPv4 creates coerced IPv4 schema that attempts string conversion
func CoercedIPv4(params ...any) *ZodIPv4[string] {
	schema := IPv4Typed[string](params...)
	schema.internals.Coerce = true
	return schema
}

// =============================================================================
// IPv6 TYPE DEFINITIONS
// =============================================================================

// ZodIPv6Def defines the configuration for IPv6 address validation
type ZodIPv6Def struct {
	core.ZodTypeDef
}

// ZodIPv6Internals contains IPv6 validator internal state
type ZodIPv6Internals struct {
	core.ZodTypeInternals
	Def *ZodIPv6Def // Schema definition
}

// ZodIPv6 represents an IPv6 address validation schema
type ZodIPv6[T NetworkConstraint] struct {
	internals *ZodIPv6Internals
}

// =============================================================================
// IPv6 CORE METHODS
// =============================================================================

// GetInternals returns the internal state of the schema
func (z *ZodIPv6[T]) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// IsOptional returns true if this schema accepts undefined/missing values
func (z *ZodIPv6[T]) IsOptional() bool {
	return z.internals.IsOptional()
}

// IsNilable returns true if this schema accepts nil values
func (z *ZodIPv6[T]) IsNilable() bool {
	return z.internals.IsNilable()
}

// Coerce implements type conversion interface using coerce package
func (z *ZodIPv6[T]) Coerce(input any) (any, bool) {
	result, err := coerce.ToString(input)
	return result, err == nil
}

// Parse returns type-safe IPv6 address using unified engine API
func (z *ZodIPv6[T]) Parse(input any, ctx ...*core.ParseContext) (T, error) {
	return engine.ParsePrimitive[string, T](
		input,
		&z.internals.ZodTypeInternals,
		core.ZodTypeIPv6,
		engine.ApplyChecks[string], // Universal validator
		engine.ConvertToConstraintType[string, T], // Universal converter
		ctx...,
	)
}

// MustParse is the variant that panics on error
func (z *ZodIPv6[T]) MustParse(input any, ctx ...*core.ParseContext) T {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// StrictParse provides compile-time type safety by requiring exact type matching.
// This eliminates runtime type checking overhead for maximum performance.
// The input must exactly match the schema's constraint type T.
func (z *ZodIPv6[T]) StrictParse(input T, ctx ...*core.ParseContext) (T, error) {
	// Create validator that applies checks to the constraint type T
	validator := func(value string, checks []core.ZodCheck, c *core.ParseContext) (string, error) {
		return engine.ApplyChecks[string](value, checks, c)
	}

	return engine.ParsePrimitiveStrict[string, T](
		input,
		&z.internals.ZodTypeInternals,
		core.ZodTypeIPv6,
		validator,
		ctx...,
	)
}

// MustStrictParse is the strict mode variant that panics on error.
// Provides compile-time type safety with maximum performance.
func (z *ZodIPv6[T]) MustStrictParse(input T, ctx ...*core.ParseContext) T {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// ParseAny validates the input value and returns any type (for runtime interface)
func (z *ZodIPv6[T]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// =============================================================================
// IPv6 MODIFIER METHODS
// =============================================================================

// Optional allows nil values, returns pointer type for nullable semantics
func (z *ZodIPv6[T]) Optional() *ZodIPv6[*string] {
	in := z.internals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// Nilable allows nil values, returns pointer type
func (z *ZodIPv6[T]) Nilable() *ZodIPv6[*string] {
	in := z.internals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish combines optional and nilable modifiers
func (z *ZodIPv6[T]) Nullish() *ZodIPv6[*string] {
	in := z.internals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Default preserves current generic type T
func (z *ZodIPv6[T]) Default(v string) *ZodIPv6[T] {
	in := z.internals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

// DefaultFunc preserves current generic type T
func (z *ZodIPv6[T]) DefaultFunc(fn func() string) *ZodIPv6[T] {
	in := z.internals.Clone()
	in.SetDefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Prefault provides fallback values on validation failure
func (z *ZodIPv6[T]) Prefault(v string) *ZodIPv6[T] {
	in := z.internals.Clone()
	in.SetPrefaultValue(v)
	return z.withInternals(in)
}

// PrefaultFunc provides dynamic fallback values
func (z *ZodIPv6[T]) PrefaultFunc(fn func() string) *ZodIPv6[T] {
	in := z.internals.Clone()
	in.SetPrefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// =============================================================================
// IPv6 TRANSFORMATION AND PIPELINE METHODS
// =============================================================================

// Transform creates type-safe transformation using WrapFn pattern
func (z *ZodIPv6[T]) Transform(fn func(string, *core.RefinementContext) (any, error)) *core.ZodTransform[T, any] {
	wrapperFn := func(input T, ctx *core.RefinementContext) (any, error) {
		str := extractNetworkString(input)
		return fn(str, ctx)
	}
	return core.NewZodTransform[T, any](z, wrapperFn)
}

// Pipe creates a pipeline using WrapFn pattern
func (z *ZodIPv6[T]) Pipe(target core.ZodType[any]) *core.ZodPipe[T, any] {
	wrapperFn := func(input T, ctx *core.ParseContext) (any, error) {
		str := extractNetworkString(input)
		return target.Parse(str, ctx)
	}
	return core.NewZodPipe[T, any](z, target, wrapperFn)
}

// =============================================================================
// IPv6 REFINEMENT METHODS
// =============================================================================

// Refine adds type-safe custom validation logic
func (z *ZodIPv6[T]) Refine(fn func(T) bool, params ...any) *ZodIPv6[T] {
	wrapper := func(v any) bool {
		var zero T

		switch any(zero).(type) {
		case string:
			if v == nil {
				return false
			}
			if strVal, ok := v.(string); ok {
				return fn(any(strVal).(T))
			}
			return false
		case *string:
			if v == nil {
				// Call the refinement function with typed nil
				return fn(any((*string)(nil)).(T))
			}
			if strVal, ok := v.(string); ok {
				sCopy := strVal
				ptr := &sCopy
				return fn(any(ptr).(T))
			}
			return false
		default:
			return false
		}
	}

	schemaParams := utils.NormalizeParams(params...)
	var errorMessage any
	if schemaParams.Error != nil {
		errorMessage = schemaParams.Error // Pass the actual error message, not the SchemaParams
	}

	check := checks.NewCustom[any](wrapper, errorMessage)
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// RefineAny adds flexible custom validation logic
func (z *ZodIPv6[T]) RefineAny(fn func(any) bool, params ...any) *ZodIPv6[T] {
	schemaParams := utils.NormalizeParams(params...)
	var errorMessage any
	if schemaParams.Error != nil {
		errorMessage = schemaParams.Error // Pass the actual error message, not the SchemaParams
	}

	check := checks.NewCustom[any](fn, errorMessage)
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// =============================================================================
// IPv6 HELPER AND PRIVATE METHODS
// =============================================================================

// withPtrInternals creates new instance with pointer type
func (z *ZodIPv6[T]) withPtrInternals(in *core.ZodTypeInternals) *ZodIPv6[*string] {
	return &ZodIPv6[*string]{internals: &ZodIPv6Internals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
	}}
}

// withInternals creates new instance preserving generic type T
func (z *ZodIPv6[T]) withInternals(in *core.ZodTypeInternals) *ZodIPv6[T] {
	return &ZodIPv6[T]{internals: &ZodIPv6Internals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
	}}
}

// CloneFrom copies configuration from another schema
func (z *ZodIPv6[T]) CloneFrom(source any) {
	if src, ok := source.(*ZodIPv6[T]); ok {
		originalChecks := z.internals.Checks
		*z.internals = *src.internals
		z.internals.Checks = originalChecks
	}
}

// newZodIPv6FromDef constructs new ZodIPv6 from definition
func newZodIPv6FromDef[T NetworkConstraint](def *ZodIPv6Def) *ZodIPv6[T] {
	internals := &ZodIPv6Internals{
		ZodTypeInternals: core.ZodTypeInternals{
			Type:   def.Type,
			Checks: def.Checks,
			Coerce: def.Coerce,
			Bag:    make(map[string]any),
		},
		Def: def,
	}

	// Add IPv6 validation check
	ipv6Check := checks.IPv6()
	internals.Checks = append(internals.Checks, ipv6Check)

	// Provide constructor for AddCheck functionality
	internals.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any] {
		ipv6Def := &ZodIPv6Def{
			ZodTypeDef: *newDef,
		}
		return any(newZodIPv6FromDef[T](ipv6Def)).(core.ZodType[any])
	}

	if def.Error != nil {
		internals.Error = def.Error
	}

	return &ZodIPv6[T]{internals: internals}
}

// IPv6 creates IPv6 address schema with type-inference support
func IPv6(params ...any) *ZodIPv6[string] {
	return IPv6Typed[string](params...)
}

// IPv6Ptr creates schema for *string IPv6
func IPv6Ptr(params ...any) *ZodIPv6[*string] {
	return IPv6Typed[*string](params...)
}

// IPv6Typed is the generic constructor for IPv6 address schemas
func IPv6Typed[T NetworkConstraint](params ...any) *ZodIPv6[T] {
	schemaParams := utils.NormalizeParams(params...)

	def := &ZodIPv6Def{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeIPv6,
			Checks: []core.ZodCheck{},
		},
	}

	// Apply normalized parameters to schema definition
	utils.ApplySchemaParams(&def.ZodTypeDef, schemaParams)

	return newZodIPv6FromDef[T](def)
}

// CoercedIPv6 creates coerced IPv6 schema that attempts string conversion
func CoercedIPv6(params ...any) *ZodIPv6[string] {
	schema := IPv6Typed[string](params...)
	schema.internals.Coerce = true
	return schema
}

// =============================================================================
// CIDRv4 TYPE DEFINITIONS
// =============================================================================

// ZodCIDRv4Def defines the configuration for CIDRv4 notation validation
type ZodCIDRv4Def struct {
	core.ZodTypeDef
}

// ZodCIDRv4Internals contains CIDRv4 validator internal state
type ZodCIDRv4Internals struct {
	core.ZodTypeInternals
	Def *ZodCIDRv4Def // Schema definition
}

// ZodCIDRv4 represents a CIDRv4 notation validation schema
type ZodCIDRv4[T NetworkConstraint] struct {
	internals *ZodCIDRv4Internals
}

// =============================================================================
// CIDRv4 CORE METHODS
// =============================================================================

// GetInternals returns the internal state of the schema
func (z *ZodCIDRv4[T]) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// IsOptional returns true if this schema accepts undefined/missing values
func (z *ZodCIDRv4[T]) IsOptional() bool {
	return z.internals.IsOptional()
}

// IsNilable returns true if this schema accepts nil values
func (z *ZodCIDRv4[T]) IsNilable() bool {
	return z.internals.IsNilable()
}

// Coerce implements type conversion interface using coerce package
func (z *ZodCIDRv4[T]) Coerce(input any) (any, bool) {
	result, err := coerce.ToString(input)
	return result, err == nil
}

// Parse returns type-safe CIDRv4 notation using unified engine API
func (z *ZodCIDRv4[T]) Parse(input any, ctx ...*core.ParseContext) (T, error) {
	return engine.ParsePrimitive[string, T](
		input,
		&z.internals.ZodTypeInternals,
		core.ZodTypeCIDRv4,
		engine.ApplyChecks[string], // Universal validator
		engine.ConvertToConstraintType[string, T], // Universal converter
		ctx...,
	)
}

// MustParse is the variant that panics on error
func (z *ZodCIDRv4[T]) MustParse(input any, ctx ...*core.ParseContext) T {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// StrictParse provides compile-time type safety by requiring exact type matching.
// This eliminates runtime type checking overhead for maximum performance.
// The input must exactly match the schema's constraint type T.
func (z *ZodCIDRv4[T]) StrictParse(input T, ctx ...*core.ParseContext) (T, error) {
	// Create validator that applies checks to the constraint type T
	validator := func(value string, checks []core.ZodCheck, c *core.ParseContext) (string, error) {
		return engine.ApplyChecks[string](value, checks, c)
	}

	return engine.ParsePrimitiveStrict[string, T](
		input,
		&z.internals.ZodTypeInternals,
		core.ZodTypeCIDRv4,
		validator,
		ctx...,
	)
}

// MustStrictParse is the strict mode variant that panics on error.
// Provides compile-time type safety with maximum performance.
func (z *ZodCIDRv4[T]) MustStrictParse(input T, ctx ...*core.ParseContext) T {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// ParseAny validates the input and returns any type (for runtime interface)
func (z *ZodCIDRv4[T]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// =============================================================================
// CIDRv4 MODIFIER METHODS
// =============================================================================

// Optional allows nil values, returns pointer type for nullable semantics
func (z *ZodCIDRv4[T]) Optional() *ZodCIDRv4[*string] {
	in := z.internals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// Nilable allows nil values, returns pointer type
func (z *ZodCIDRv4[T]) Nilable() *ZodCIDRv4[*string] {
	in := z.internals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish combines optional and nilable modifiers
func (z *ZodCIDRv4[T]) Nullish() *ZodCIDRv4[*string] {
	in := z.internals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Default preserves current generic type T
func (z *ZodCIDRv4[T]) Default(v string) *ZodCIDRv4[T] {
	in := z.internals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

// DefaultFunc preserves current generic type T
func (z *ZodCIDRv4[T]) DefaultFunc(fn func() string) *ZodCIDRv4[T] {
	in := z.internals.Clone()
	in.SetDefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Prefault provides fallback values on validation failure
func (z *ZodCIDRv4[T]) Prefault(v string) *ZodCIDRv4[T] {
	in := z.internals.Clone()
	in.SetPrefaultValue(v)
	return z.withInternals(in)
}

// PrefaultFunc provides dynamic fallback values
func (z *ZodCIDRv4[T]) PrefaultFunc(fn func() string) *ZodCIDRv4[T] {
	in := z.internals.Clone()
	in.SetPrefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// =============================================================================
// CIDRv4 TRANSFORMATION AND PIPELINE METHODS
// =============================================================================

// Transform creates type-safe transformation using WrapFn pattern
func (z *ZodCIDRv4[T]) Transform(fn func(string, *core.RefinementContext) (any, error)) *core.ZodTransform[T, any] {
	wrapperFn := func(input T, ctx *core.RefinementContext) (any, error) {
		str := extractNetworkString(input)
		return fn(str, ctx)
	}
	return core.NewZodTransform[T, any](z, wrapperFn)
}

// Pipe creates a pipeline using WrapFn pattern
func (z *ZodCIDRv4[T]) Pipe(target core.ZodType[any]) *core.ZodPipe[T, any] {
	wrapperFn := func(input T, ctx *core.ParseContext) (any, error) {
		str := extractNetworkString(input)
		return target.Parse(str, ctx)
	}
	return core.NewZodPipe[T, any](z, target, wrapperFn)
}

// =============================================================================
// CIDRv4 REFINEMENT METHODS
// =============================================================================

// Refine adds type-safe custom validation logic
func (z *ZodCIDRv4[T]) Refine(fn func(T) bool, params ...any) *ZodCIDRv4[T] {
	wrapper := func(v any) bool {
		var zero T

		switch any(zero).(type) {
		case string:
			if v == nil {
				return false
			}
			if strVal, ok := v.(string); ok {
				return fn(any(strVal).(T))
			}
			return false
		case *string:
			if v == nil {
				// Call the refinement function with typed nil
				return fn(any((*string)(nil)).(T))
			}
			if strVal, ok := v.(string); ok {
				sCopy := strVal
				ptr := &sCopy
				return fn(any(ptr).(T))
			}
			return false
		default:
			return false
		}
	}

	schemaParams := utils.NormalizeParams(params...)
	var errorMessage any
	if schemaParams.Error != nil {
		errorMessage = schemaParams.Error // Pass the actual error message, not the SchemaParams
	}

	check := checks.NewCustom[any](wrapper, errorMessage)
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// RefineAny adds flexible custom validation logic
func (z *ZodCIDRv4[T]) RefineAny(fn func(any) bool, params ...any) *ZodCIDRv4[T] {
	schemaParams := utils.NormalizeParams(params...)
	var errorMessage any
	if schemaParams.Error != nil {
		errorMessage = schemaParams.Error // Pass the actual error message, not the SchemaParams
	}

	check := checks.NewCustom[any](fn, errorMessage)
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// =============================================================================
// CIDRv4 HELPER AND PRIVATE METHODS
// =============================================================================

// withPtrInternals creates new instance with pointer type
func (z *ZodCIDRv4[T]) withPtrInternals(in *core.ZodTypeInternals) *ZodCIDRv4[*string] {
	return &ZodCIDRv4[*string]{internals: &ZodCIDRv4Internals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
	}}
}

// withInternals creates new instance preserving generic type T
func (z *ZodCIDRv4[T]) withInternals(in *core.ZodTypeInternals) *ZodCIDRv4[T] {
	return &ZodCIDRv4[T]{internals: &ZodCIDRv4Internals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
	}}
}

// CloneFrom copies configuration from another schema
func (z *ZodCIDRv4[T]) CloneFrom(source any) {
	if src, ok := source.(*ZodCIDRv4[T]); ok {
		originalChecks := z.internals.Checks
		*z.internals = *src.internals
		z.internals.Checks = originalChecks
	}
}

// newZodCIDRv4FromDef constructs new ZodCIDRv4 from definition
func newZodCIDRv4FromDef[T NetworkConstraint](def *ZodCIDRv4Def) *ZodCIDRv4[T] {
	internals := &ZodCIDRv4Internals{
		ZodTypeInternals: core.ZodTypeInternals{
			Type:   def.Type,
			Checks: def.Checks,
			Coerce: def.Coerce,
			Bag:    make(map[string]any),
		},
		Def: def,
	}

	// Add CIDRv4 validation check
	cidrv4Check := checks.CIDRv4()
	internals.Checks = append(internals.Checks, cidrv4Check)

	// Provide constructor for AddCheck functionality
	internals.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any] {
		cidrv4Def := &ZodCIDRv4Def{
			ZodTypeDef: *newDef,
		}
		return any(newZodCIDRv4FromDef[T](cidrv4Def)).(core.ZodType[any])
	}

	if def.Error != nil {
		internals.Error = def.Error
	}

	return &ZodCIDRv4[T]{internals: internals}
}

// CIDRv4 creates CIDRv4 notation schema with type-inference support
func CIDRv4(params ...any) *ZodCIDRv4[string] {
	return CIDRv4Typed[string](params...)
}

// CIDRv4Ptr creates schema for *string CIDRv4
func CIDRv4Ptr(params ...any) *ZodCIDRv4[*string] {
	return CIDRv4Typed[*string](params...)
}

// CIDRv4Typed is the generic constructor for CIDRv4 notation schemas
func CIDRv4Typed[T NetworkConstraint](params ...any) *ZodCIDRv4[T] {
	schemaParams := utils.NormalizeParams(params...)

	def := &ZodCIDRv4Def{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeCIDRv4,
			Checks: []core.ZodCheck{},
		},
	}

	// Apply normalized parameters to schema definition
	utils.ApplySchemaParams(&def.ZodTypeDef, schemaParams)

	return newZodCIDRv4FromDef[T](def)
}

// CoercedCIDRv4 creates coerced CIDRv4 schema that attempts string conversion
func CoercedCIDRv4(params ...any) *ZodCIDRv4[string] {
	schema := CIDRv4Typed[string](params...)
	schema.internals.Coerce = true
	return schema
}

// =============================================================================
// CIDRv6 TYPE DEFINITIONS
// =============================================================================

// ZodCIDRv6Def defines the configuration for CIDRv6 notation validation
type ZodCIDRv6Def struct {
	core.ZodTypeDef
}

// ZodCIDRv6Internals contains CIDRv6 validator internal state
type ZodCIDRv6Internals struct {
	core.ZodTypeInternals
	Def *ZodCIDRv6Def // Schema definition
}

// ZodCIDRv6 represents a CIDRv6 notation validation schema
type ZodCIDRv6[T NetworkConstraint] struct {
	internals *ZodCIDRv6Internals
}

// =============================================================================
// CIDRv6 CORE METHODS
// =============================================================================

// GetInternals returns the internal state of the schema
func (z *ZodCIDRv6[T]) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// IsOptional returns true if this schema accepts undefined/missing values
func (z *ZodCIDRv6[T]) IsOptional() bool {
	return z.internals.IsOptional()
}

// IsNilable returns true if this schema accepts nil values
func (z *ZodCIDRv6[T]) IsNilable() bool {
	return z.internals.IsNilable()
}

// Coerce implements type conversion interface using coerce package
func (z *ZodCIDRv6[T]) Coerce(input any) (any, bool) {
	result, err := coerce.ToString(input)
	return result, err == nil
}

// Parse returns type-safe CIDRv6 notation using unified engine API
func (z *ZodCIDRv6[T]) Parse(input any, ctx ...*core.ParseContext) (T, error) {
	return engine.ParsePrimitive[string, T](
		input,
		&z.internals.ZodTypeInternals,
		core.ZodTypeCIDRv6,
		engine.ApplyChecks[string], // Universal validator
		engine.ConvertToConstraintType[string, T], // Universal converter
		ctx...,
	)
}

// MustParse is the variant that panics on error
func (z *ZodCIDRv6[T]) MustParse(input any, ctx ...*core.ParseContext) T {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// StrictParse provides compile-time type safety by requiring exact type matching.
// This eliminates runtime type checking overhead for maximum performance.
// The input must exactly match the schema's constraint type T.
func (z *ZodCIDRv6[T]) StrictParse(input T, ctx ...*core.ParseContext) (T, error) {
	// Create validator that applies checks to the constraint type T
	validator := func(value string, checks []core.ZodCheck, c *core.ParseContext) (string, error) {
		return engine.ApplyChecks[string](value, checks, c)
	}

	return engine.ParsePrimitiveStrict[string, T](
		input,
		&z.internals.ZodTypeInternals,
		core.ZodTypeCIDRv6,
		validator,
		ctx...,
	)
}

// MustStrictParse is the strict mode variant that panics on error.
// Provides compile-time type safety with maximum performance.
func (z *ZodCIDRv6[T]) MustStrictParse(input T, ctx ...*core.ParseContext) T {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// ParseAny validates the input and returns any type (for runtime interface)
func (z *ZodCIDRv6[T]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// =============================================================================
// CIDRv6 MODIFIER METHODS
// =============================================================================

// Optional allows nil values, returns pointer type for nullable semantics
func (z *ZodCIDRv6[T]) Optional() *ZodCIDRv6[*string] {
	in := z.internals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// Nilable allows nil values, returns pointer type
func (z *ZodCIDRv6[T]) Nilable() *ZodCIDRv6[*string] {
	in := z.internals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish combines optional and nilable modifiers
func (z *ZodCIDRv6[T]) Nullish() *ZodCIDRv6[*string] {
	in := z.internals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Default preserves current generic type T
func (z *ZodCIDRv6[T]) Default(v string) *ZodCIDRv6[T] {
	in := z.internals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

// DefaultFunc preserves current generic type T
func (z *ZodCIDRv6[T]) DefaultFunc(fn func() string) *ZodCIDRv6[T] {
	in := z.internals.Clone()
	in.SetDefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Prefault provides fallback values on validation failure
func (z *ZodCIDRv6[T]) Prefault(v string) *ZodCIDRv6[T] {
	in := z.internals.Clone()
	in.SetPrefaultValue(v)
	return z.withInternals(in)
}

// PrefaultFunc provides dynamic fallback values
func (z *ZodCIDRv6[T]) PrefaultFunc(fn func() string) *ZodCIDRv6[T] {
	in := z.internals.Clone()
	in.SetPrefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// =============================================================================
// CIDRv6 TRANSFORMATION AND PIPELINE METHODS
// =============================================================================

// Transform creates type-safe transformation using WrapFn pattern
func (z *ZodCIDRv6[T]) Transform(fn func(string, *core.RefinementContext) (any, error)) *core.ZodTransform[T, any] {
	wrapperFn := func(input T, ctx *core.RefinementContext) (any, error) {
		str := extractNetworkString(input)
		return fn(str, ctx)
	}
	return core.NewZodTransform[T, any](z, wrapperFn)
}

// Pipe creates a pipeline using WrapFn pattern
func (z *ZodCIDRv6[T]) Pipe(target core.ZodType[any]) *core.ZodPipe[T, any] {
	wrapperFn := func(input T, ctx *core.ParseContext) (any, error) {
		str := extractNetworkString(input)
		return target.Parse(str, ctx)
	}
	return core.NewZodPipe[T, any](z, target, wrapperFn)
}

// =============================================================================
// CIDRv6 REFINEMENT METHODS
// =============================================================================

// Refine adds type-safe custom validation logic
func (z *ZodCIDRv6[T]) Refine(fn func(T) bool, params ...any) *ZodCIDRv6[T] {
	wrapper := func(v any) bool {
		var zero T

		switch any(zero).(type) {
		case string:
			if v == nil {
				return false
			}
			if strVal, ok := v.(string); ok {
				return fn(any(strVal).(T))
			}
			return false
		case *string:
			if v == nil {
				// Call the refinement function with typed nil
				return fn(any((*string)(nil)).(T))
			}
			if strVal, ok := v.(string); ok {
				sCopy := strVal
				ptr := &sCopy
				return fn(any(ptr).(T))
			}
			return false
		default:
			return false
		}
	}

	schemaParams := utils.NormalizeParams(params...)
	var errorMessage any
	if schemaParams.Error != nil {
		errorMessage = schemaParams.Error // Pass the actual error message, not the SchemaParams
	}

	check := checks.NewCustom[any](wrapper, errorMessage)
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// RefineAny adds flexible custom validation logic
func (z *ZodCIDRv6[T]) RefineAny(fn func(any) bool, params ...any) *ZodCIDRv6[T] {
	schemaParams := utils.NormalizeParams(params...)
	var errorMessage any
	if schemaParams.Error != nil {
		errorMessage = schemaParams.Error // Pass the actual error message, not the SchemaParams
	}

	check := checks.NewCustom[any](fn, errorMessage)
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// =============================================================================
// CIDRv6 HELPER AND PRIVATE METHODS
// =============================================================================

// withPtrInternals creates new instance with pointer type
func (z *ZodCIDRv6[T]) withPtrInternals(in *core.ZodTypeInternals) *ZodCIDRv6[*string] {
	return &ZodCIDRv6[*string]{internals: &ZodCIDRv6Internals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
	}}
}

// withInternals creates new instance preserving generic type T
func (z *ZodCIDRv6[T]) withInternals(in *core.ZodTypeInternals) *ZodCIDRv6[T] {
	return &ZodCIDRv6[T]{internals: &ZodCIDRv6Internals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
	}}
}

// CloneFrom copies configuration from another schema
func (z *ZodCIDRv6[T]) CloneFrom(source any) {
	if src, ok := source.(*ZodCIDRv6[T]); ok {
		originalChecks := z.internals.Checks
		*z.internals = *src.internals
		z.internals.Checks = originalChecks
	}
}

// newZodCIDRv6FromDef constructs new ZodCIDRv6 from definition
func newZodCIDRv6FromDef[T NetworkConstraint](def *ZodCIDRv6Def) *ZodCIDRv6[T] {
	internals := &ZodCIDRv6Internals{
		ZodTypeInternals: core.ZodTypeInternals{
			Type:   def.Type,
			Checks: def.Checks,
			Coerce: def.Coerce,
			Bag:    make(map[string]any),
		},
		Def: def,
	}

	// Add CIDRv6 validation check
	cidrv6Check := checks.CIDRv6()
	internals.Checks = append(internals.Checks, cidrv6Check)

	// Provide constructor for AddCheck functionality
	internals.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any] {
		cidrv6Def := &ZodCIDRv6Def{
			ZodTypeDef: *newDef,
		}
		return any(newZodCIDRv6FromDef[T](cidrv6Def)).(core.ZodType[any])
	}

	if def.Error != nil {
		internals.Error = def.Error
	}

	return &ZodCIDRv6[T]{internals: internals}
}

// CIDRv6 creates CIDRv6 notation schema with type-inference support
func CIDRv6(params ...any) *ZodCIDRv6[string] {
	return CIDRv6Typed[string](params...)
}

// CIDRv6Ptr creates schema for *string CIDRv6
func CIDRv6Ptr(params ...any) *ZodCIDRv6[*string] {
	return CIDRv6Typed[*string](params...)
}

// CIDRv6Typed is the generic constructor for CIDRv6 notation schemas
func CIDRv6Typed[T NetworkConstraint](params ...any) *ZodCIDRv6[T] {
	schemaParams := utils.NormalizeParams(params...)

	def := &ZodCIDRv6Def{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeCIDRv6,
			Checks: []core.ZodCheck{},
		},
	}

	// Apply normalized parameters to schema definition
	utils.ApplySchemaParams(&def.ZodTypeDef, schemaParams)

	return newZodCIDRv6FromDef[T](def)
}

// CoercedCIDRv6 creates coerced CIDRv6 schema that attempts string conversion
func CoercedCIDRv6(params ...any) *ZodCIDRv6[string] {
	schema := CIDRv6Typed[string](params...)
	schema.internals.Coerce = true
	return schema
}

// =============================================================================
// URL OPTIONS TYPES
// =============================================================================

// URLOptions defines options for URL validation (similar to IsoDatetimeOptions)
type URLOptions struct {
	// Hostname validation pattern
	Hostname *regexp.Regexp
	// Protocol validation pattern
	Protocol *regexp.Regexp
}

// =============================================================================
// URL TYPE DEFINITIONS
// =============================================================================

// ZodURLDef defines the configuration for URL validation
type ZodURLDef struct {
	core.ZodTypeDef
}

// ZodURLInternals contains URL validator internal state
type ZodURLInternals struct {
	core.ZodTypeInternals
	Def *ZodURLDef // Schema definition
}

// ZodURL represents a URL validation schema
type ZodURL[T NetworkConstraint] struct {
	internals *ZodURLInternals
}

// =============================================================================
// URL CORE METHODS
// =============================================================================

// GetInternals returns the internal state of the schema
func (z *ZodURL[T]) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// IsOptional returns true if this schema accepts undefined/missing values
func (z *ZodURL[T]) IsOptional() bool {
	return z.internals.IsOptional()
}

// IsNilable returns true if this schema accepts nil values
func (z *ZodURL[T]) IsNilable() bool {
	return z.internals.IsNilable()
}

// Coerce implements type conversion interface using coerce package
func (z *ZodURL[T]) Coerce(input any) (any, bool) {
	result, err := coerce.ToString(input)
	return result, err == nil
}

// Parse returns type-safe URL using unified engine API
func (z *ZodURL[T]) Parse(input any, ctx ...*core.ParseContext) (T, error) {
	return engine.ParsePrimitive[string, T](
		input,
		&z.internals.ZodTypeInternals,
		core.ZodTypeURL,
		engine.ApplyChecks[string], // Universal validator
		engine.ConvertToConstraintType[string, T], // Universal converter
		ctx...,
	)
}

// MustParse is the variant that panics on error
func (z *ZodURL[T]) MustParse(input any, ctx ...*core.ParseContext) T {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// StrictParse provides compile-time type safety by requiring exact type matching.
// This eliminates runtime type checking overhead for maximum performance.
// The input must exactly match the schema's constraint type T.
func (z *ZodURL[T]) StrictParse(input T, ctx ...*core.ParseContext) (T, error) {
	// Create validator that applies checks to the constraint type T
	validator := func(value string, checks []core.ZodCheck, c *core.ParseContext) (string, error) {
		return engine.ApplyChecks[string](value, checks, c)
	}

	return engine.ParsePrimitiveStrict[string, T](
		input,
		&z.internals.ZodTypeInternals,
		core.ZodTypeURL,
		validator,
		ctx...,
	)
}

// MustStrictParse is the strict mode variant that panics on error.
// Provides compile-time type safety with maximum performance.
func (z *ZodURL[T]) MustStrictParse(input T, ctx ...*core.ParseContext) T {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// ParseAny validates the input value and returns any type (for runtime interface)
func (z *ZodURL[T]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// =============================================================================
// URL MODIFIER METHODS
// =============================================================================

// Optional allows nil values, returns pointer type for nullable semantics
func (z *ZodURL[T]) Optional() *ZodURL[*string] {
	in := z.internals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// Nilable allows nil values, returns pointer type
func (z *ZodURL[T]) Nilable() *ZodURL[*string] {
	in := z.internals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish combines optional and nilable modifiers
func (z *ZodURL[T]) Nullish() *ZodURL[*string] {
	in := z.internals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Default preserves current generic type T
func (z *ZodURL[T]) Default(v string) *ZodURL[T] {
	in := z.internals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

// DefaultFunc preserves current generic type T
func (z *ZodURL[T]) DefaultFunc(fn func() string) *ZodURL[T] {
	in := z.internals.Clone()
	in.SetDefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Prefault provides fallback values on validation failure
func (z *ZodURL[T]) Prefault(v string) *ZodURL[T] {
	in := z.internals.Clone()
	in.SetPrefaultValue(v)
	return z.withInternals(in)
}

// PrefaultFunc provides dynamic fallback values
func (z *ZodURL[T]) PrefaultFunc(fn func() string) *ZodURL[T] {
	in := z.internals.Clone()
	in.SetPrefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// =============================================================================
// URL REFINEMENT METHODS
// =============================================================================

// Refine adds type-safe custom validation logic
func (z *ZodURL[T]) Refine(fn func(T) bool, params ...any) *ZodURL[T] {
	wrapper := func(v any) bool {
		var zero T

		switch any(zero).(type) {
		case string:
			if v == nil {
				return false
			}
			if strVal, ok := v.(string); ok {
				return fn(any(strVal).(T))
			}
			return false
		case *string:
			if v == nil {
				// Call the refinement function with typed nil
				return fn(any((*string)(nil)).(T))
			}
			if strVal, ok := v.(string); ok {
				sCopy := strVal
				ptr := &sCopy
				return fn(any(ptr).(T))
			}
			return false
		default:
			return false
		}
	}

	schemaParams := utils.NormalizeParams(params...)
	var errorMessage any
	if schemaParams.Error != nil {
		errorMessage = schemaParams.Error // Pass the actual error message, not the SchemaParams
	}

	check := checks.NewCustom[any](wrapper, errorMessage)
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// RefineAny adds a custom refinement function for any type
func (z *ZodURL[T]) RefineAny(fn func(any) bool, params ...any) *ZodURL[T] {
	schemaParams := utils.NormalizeParams(params...)
	var errorMessage any
	if schemaParams.Error != nil {
		errorMessage = schemaParams.Error // Pass the actual error message, not the SchemaParams
	}

	check := checks.NewCustom[any](fn, errorMessage)
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// =============================================================================
// URL INTERNAL HELPER METHODS
// =============================================================================

// withPtrInternals creates a new ZodURL instance with pointer internals
func (z *ZodURL[T]) withPtrInternals(in *core.ZodTypeInternals) *ZodURL[*string] {
	return &ZodURL[*string]{
		internals: &ZodURLInternals{
			ZodTypeInternals: *in,
			Def:              z.internals.Def,
		},
	}
}

// withInternals creates a new ZodURL instance with the given internals
func (z *ZodURL[T]) withInternals(in *core.ZodTypeInternals) *ZodURL[T] {
	return &ZodURL[T]{
		internals: &ZodURLInternals{
			ZodTypeInternals: *in,
			Def:              z.internals.Def,
		},
	}
}

// CloneFrom clones internals from another ZodURL instance
func (z *ZodURL[T]) CloneFrom(source any) {
	if srcURL, ok := source.(*ZodURL[T]); ok {
		z.internals = srcURL.internals
	}
}

// newZodURLFromDef creates a new ZodURL instance from definition
func newZodURLFromDef[T NetworkConstraint](def *ZodURLDef) *ZodURL[T] {
	internals := &ZodURLInternals{
		ZodTypeInternals: core.ZodTypeInternals{
			Type:   core.ZodTypeURL,
			Checks: def.Checks,
			Coerce: def.Coerce,
			Bag:    make(map[string]any),
		},
		Def: def,
	}

	// Add basic URL validation check
	internals.AddCheck(checks.URL())

	// Provide constructor for AddCheck functionality
	internals.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any] {
		urlDef := &ZodURLDef{
			ZodTypeDef: *newDef,
		}
		return any(newZodURLFromDef[T](urlDef)).(core.ZodType[any])
	}

	if def.Error != nil {
		internals.Error = def.Error
	}

	return &ZodURL[T]{
		internals: internals,
	}
}

// =============================================================================
// URL CONSTRUCTOR FUNCTIONS
// =============================================================================

// URL creates a new URL schema with flexible parameter support
// Supports various parameter combinations:
//
//	URL() - basic URL schema
//	URL("error message") - with custom error message
//	URL(URLOptions{Hostname: pattern}) - with URL options
//	URL(URLOptions{...}, "error message") - options with error message
//	URL(URLOptions{...}, core.SchemaParams{Description: "URL"}) - options with schema params
//	URL(core.SchemaParams{Description: "URL"}) - with schema parameters only
func URL(params ...any) *ZodURL[string] {
	return URLTyped[string](params...)
}

// URLPtr creates a new URL schema with pointer type and flexible parameter support
// Supports the same parameter combinations as URL() but returns *ZodURL[*string]
func URLPtr(params ...any) *ZodURL[*string] {
	return URLTyped[*string](params...)
}

// URLTyped creates a new URL schema with specific type and flexible parameter support
func URLTyped[T NetworkConstraint](params ...any) *ZodURL[T] {
	var options *URLOptions
	var schemaParams *core.SchemaParams
	var errorMessage string

	// Parse parameters
	for _, param := range params {
		switch p := param.(type) {
		case URLOptions:
			options = &p
		case core.SchemaParams:
			schemaParams = &p
		case string:
			errorMessage = p
		}
	}

	def := &ZodURLDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:     core.ZodTypeURL,
			Required: true,
			Checks:   []core.ZodCheck{},
		},
	}

	// Apply schema parameters if provided
	if schemaParams != nil {
		utils.ApplySchemaParams(&def.ZodTypeDef, schemaParams)
	}

	// Apply error message if provided
	if errorMessage != "" {
		// Create a simple error function for the error message
		errorFunc := core.ZodErrorMap(func(issue core.ZodRawIssue) string {
			return errorMessage
		})
		def.Error = &errorFunc
	}

	schema := newZodURLFromDef[T](def)

	// Apply URL options if provided
	if options != nil {
		if options.Hostname != nil || options.Protocol != nil {
			validateOptions := validate.URLOptions{
				Hostname: options.Hostname,
				Protocol: options.Protocol,
			}
			check := checks.URLWithOptions(validateOptions)
			schema.internals.AddCheck(check)
		}
	}

	return schema
}

// CoercedURL creates coerced URL schema that attempts string conversion
func CoercedURL(params ...any) *ZodURL[string] {
	schema := URLTyped[string](params...)
	schema.internals.Coerce = true
	return schema
}

// HttpURL creates an HTTP/HTTPS URL schema
// This is a convenience function that creates a URL schema restricted to
// http:// and https:// protocols with domain hostname validation.
//
// TypeScript Zod v4 equivalent: z.httpUrl()
//
// Examples:
//
//	schema := HttpURL()
//	schema.Parse("https://example.com")      // valid
//	schema.Parse("http://sub.example.com")   // valid
//	schema.Parse("ftp://example.com")        // invalid - not http/https
//	schema.Parse("http://localhost")         // invalid - not a domain
func HttpURL(params ...any) *ZodURL[string] {
	return HttpURLTyped[string](params...)
}

// HttpURLPtr creates an HTTP/HTTPS URL schema for pointer types
func HttpURLPtr(params ...any) *ZodURL[*string] {
	return HttpURLTyped[*string](params...)
}

// HttpURLTyped creates an HTTP/HTTPS URL schema with specific type
func HttpURLTyped[T NetworkConstraint](params ...any) *ZodURL[T] {
	// Create URL with http/https protocol and domain hostname restriction
	httpOptions := URLOptions{
		Protocol: regex.HTTPProtocol,
		Hostname: regex.Domain,
	}

	// Prepend the HTTP options to user-provided params
	allParams := append([]any{httpOptions}, params...)
	return URLTyped[T](allParams...)
}

// =============================================================================
// HOSTNAME TYPE DEFINITIONS
// =============================================================================

// ZodHostnameDef defines the configuration for hostname validation
type ZodHostnameDef struct {
	core.ZodTypeDef
}

// ZodHostnameInternals contains hostname validator internal state
type ZodHostnameInternals struct {
	core.ZodTypeInternals
	Def *ZodHostnameDef
}

// ZodHostname represents a DNS hostname validation schema
type ZodHostname[T NetworkConstraint] struct {
	internals *ZodHostnameInternals
}

// =============================================================================
// HOSTNAME CORE METHODS
// =============================================================================

// GetInternals returns the internal state of the schema
func (z *ZodHostname[T]) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// IsOptional returns true if this schema accepts undefined/missing values
func (z *ZodHostname[T]) IsOptional() bool {
	return z.internals.IsOptional()
}

// IsNilable returns true if this schema accepts nil values
func (z *ZodHostname[T]) IsNilable() bool {
	return z.internals.IsNilable()
}

// Coerce implements type conversion interface
func (z *ZodHostname[T]) Coerce(input any) (any, bool) {
	result, err := coerce.ToString(input)
	return result, err == nil
}

// Parse validates and returns a hostname string
func (z *ZodHostname[T]) Parse(input any, ctx ...*core.ParseContext) (T, error) {
	return engine.ParsePrimitive[string, T](
		input,
		&z.internals.ZodTypeInternals,
		core.ZodTypeHostname,
		engine.ApplyChecks[string],
		engine.ConvertToConstraintType[string, T],
		ctx...,
	)
}

// MustParse is the variant that panics on error
func (z *ZodHostname[T]) MustParse(input any, ctx ...*core.ParseContext) T {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// StrictParse provides compile-time type safety
func (z *ZodHostname[T]) StrictParse(input T, ctx ...*core.ParseContext) (T, error) {
	return engine.ParsePrimitiveStrict[string, T](
		input,
		&z.internals.ZodTypeInternals,
		core.ZodTypeHostname,
		engine.ApplyChecks[string],
		ctx...,
	)
}

// MustStrictParse is the strict mode variant that panics on error
func (z *ZodHostname[T]) MustStrictParse(input T, ctx ...*core.ParseContext) T {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// ParseAny validates input and returns untyped result
func (z *ZodHostname[T]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// =============================================================================
// HOSTNAME MODIFIER METHODS
// =============================================================================

// Optional returns a schema that accepts nil values
func (z *ZodHostname[T]) Optional() *ZodHostname[*string] {
	in := z.internals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// Nilable returns a schema that accepts nil values
func (z *ZodHostname[T]) Nilable() *ZodHostname[*string] {
	in := z.internals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Default sets a default value
func (z *ZodHostname[T]) Default(v string) *ZodHostname[T] {
	in := z.internals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

// Describe adds a description to the schema
func (z *ZodHostname[T]) Describe(description string) *ZodHostname[T] {
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
// HOSTNAME REFINEMENT METHODS
// =============================================================================

// Refine adds type-safe custom validation logic
func (z *ZodHostname[T]) Refine(fn func(T) bool, params ...any) *ZodHostname[T] {
	wrapper := func(v any) bool {
		var zero T

		switch any(zero).(type) {
		case string:
			if v == nil {
				return false
			}
			if strVal, ok := v.(string); ok {
				return fn(any(strVal).(T))
			}
			return false
		case *string:
			if v == nil {
				// Call the refinement function with typed nil
				return fn(any((*string)(nil)).(T))
			}
			if strVal, ok := v.(string); ok {
				sCopy := strVal
				ptr := &sCopy
				return fn(any(ptr).(T))
			}
			return false
		default:
			return false
		}
	}

	schemaParams := utils.NormalizeParams(params...)
	var errorMessage any
	if schemaParams.Error != nil {
		errorMessage = schemaParams.Error
	}

	check := checks.NewCustom[any](wrapper, errorMessage)
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// RefineAny adds flexible custom validation logic
func (z *ZodHostname[T]) RefineAny(fn func(any) bool, params ...any) *ZodHostname[T] {
	schemaParams := utils.NormalizeParams(params...)
	var errorMessage any
	if schemaParams.Error != nil {
		errorMessage = schemaParams.Error
	}

	check := checks.NewCustom[any](fn, errorMessage)
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// =============================================================================
// HOSTNAME COMPOSITION METHODS
// =============================================================================

// And creates an intersection with another schema
func (z *ZodHostname[T]) And(other any) *ZodIntersection[any, any] {
	return Intersection(z, other)
}

// Or creates a union with another schema
func (z *ZodHostname[T]) Or(other any) *ZodUnion[any, any] {
	return Union([]any{z, other})
}

// =============================================================================
// HOSTNAME HELPER METHODS
// =============================================================================

func (z *ZodHostname[T]) withPtrInternals(in *core.ZodTypeInternals) *ZodHostname[*string] {
	return &ZodHostname[*string]{internals: &ZodHostnameInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
	}}
}

func (z *ZodHostname[T]) withInternals(in *core.ZodTypeInternals) *ZodHostname[T] {
	return &ZodHostname[T]{internals: &ZodHostnameInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
	}}
}

// =============================================================================
// HOSTNAME CONSTRUCTOR FUNCTIONS
// =============================================================================

// Hostname creates a new hostname schema
func Hostname(params ...any) *ZodHostname[string] {
	return HostnameTyped[string](params...)
}

// HostnamePtr creates a new hostname schema with pointer type
func HostnamePtr(params ...any) *ZodHostname[*string] {
	return HostnameTyped[*string](params...)
}

// HostnameTyped creates a new hostname schema with specific type
func HostnameTyped[T NetworkConstraint](params ...any) *ZodHostname[T] {
	schemaParams := utils.NormalizeParams(params...)
	def := &ZodHostnameDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:     core.ZodTypeHostname,
			Required: true,
			Checks:   []core.ZodCheck{checks.Hostname()},
		},
	}

	if schemaParams != nil {
		utils.ApplySchemaParams(&def.ZodTypeDef, schemaParams)
	}

	internals := &ZodHostnameInternals{
		ZodTypeInternals: core.ZodTypeInternals{
			Type:   core.ZodTypeHostname,
			Checks: def.Checks,
			Bag:    make(map[string]any),
		},
		Def: def,
	}

	return &ZodHostname[T]{internals: internals}
}

// =============================================================================
// MAC ADDRESS VALIDATION
// =============================================================================

// ZodMACDef defines the configuration for MAC address validation
type ZodMACDef struct {
	core.ZodTypeDef
	Delimiter string // MAC address delimiter (default ":")
}

// ZodMACInternals contains MAC address validator internal state
type ZodMACInternals struct {
	core.ZodTypeInternals
	Def *ZodMACDef
}

// Clone creates a deep copy of internals
func (i *ZodMACInternals) Clone() *core.ZodTypeInternals {
	return i.ZodTypeInternals.Clone()
}

// ZodMAC represents a MAC address validation schema
type ZodMAC[T NetworkConstraint] struct {
	internals *ZodMACInternals
}

// =============================================================================
// CORE METHODS
// =============================================================================

func (z *ZodMAC[T]) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

func (z *ZodMAC[T]) IsOptional() bool {
	return z.internals.Optional
}

func (z *ZodMAC[T]) IsNilable() bool {
	return z.internals.Nilable
}

func (z *ZodMAC[T]) Coerce(input any) (any, bool) {
	return input, false
}

func (z *ZodMAC[T]) Parse(input any, ctx ...*core.ParseContext) (T, error) {
	return engine.ParsePrimitive[string, T](
		input,
		&z.internals.ZodTypeInternals,
		core.ZodTypeMAC,
		engine.ApplyChecks[string],
		engine.ConvertToConstraintType[string, T],
		ctx...,
	)
}

func (z *ZodMAC[T]) MustParse(input any, ctx ...*core.ParseContext) T {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

func (z *ZodMAC[T]) StrictParse(input T, ctx ...*core.ParseContext) (T, error) {
	return engine.ParsePrimitiveStrict[string, T](
		input,
		&z.internals.ZodTypeInternals,
		core.ZodTypeMAC,
		engine.ApplyChecks[string],
		ctx...,
	)
}

func (z *ZodMAC[T]) MustStrictParse(input T, ctx ...*core.ParseContext) T {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

func (z *ZodMAC[T]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// =============================================================================
// MODIFIER METHODS
// =============================================================================

func (z *ZodMAC[T]) Optional() *ZodMAC[*string] {
	in := z.internals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

func (z *ZodMAC[T]) Nilable() *ZodMAC[*string] {
	in := z.internals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

func (z *ZodMAC[T]) Default(v string) *ZodMAC[T] {
	in := z.internals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

func (z *ZodMAC[T]) Describe(description string) *ZodMAC[T] {
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
// MAC REFINEMENT METHODS
// =============================================================================

// Refine adds type-safe custom validation logic
func (z *ZodMAC[T]) Refine(fn func(T) bool, params ...any) *ZodMAC[T] {
	wrapper := func(v any) bool {
		var zero T

		switch any(zero).(type) {
		case string:
			if v == nil {
				return false
			}
			if strVal, ok := v.(string); ok {
				return fn(any(strVal).(T))
			}
			return false
		case *string:
			if v == nil {
				return fn(any((*string)(nil)).(T))
			}
			if strVal, ok := v.(string); ok {
				sCopy := strVal
				ptr := &sCopy
				return fn(any(ptr).(T))
			}
			return false
		default:
			return false
		}
	}

	schemaParams := utils.NormalizeParams(params...)
	var errorMessage any
	if schemaParams.Error != nil {
		errorMessage = schemaParams.Error
	}

	check := checks.NewCustom[any](wrapper, errorMessage)
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// RefineAny adds flexible custom validation logic
func (z *ZodMAC[T]) RefineAny(fn func(any) bool, params ...any) *ZodMAC[T] {
	schemaParams := utils.NormalizeParams(params...)
	var errorMessage any
	if schemaParams.Error != nil {
		errorMessage = schemaParams.Error
	}

	check := checks.NewCustom[any](fn, errorMessage)
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// =============================================================================
// COMPOSITION METHODS (Zod v4 Compatibility)
// =============================================================================

func (z *ZodMAC[T]) And(other any) *ZodIntersection[any, any] {
	return Intersection(z, other)
}

func (z *ZodMAC[T]) Or(other any) *ZodUnion[any, any] {
	return Union([]any{z, other})
}

// =============================================================================
// HELPER METHODS
// =============================================================================

func (z *ZodMAC[T]) withPtrInternals(in *core.ZodTypeInternals) *ZodMAC[*string] {
	return &ZodMAC[*string]{internals: &ZodMACInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
	}}
}

func (z *ZodMAC[T]) withInternals(in *core.ZodTypeInternals) *ZodMAC[T] {
	return &ZodMAC[T]{internals: &ZodMACInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
	}}
}

// =============================================================================
// CONSTRUCTORS
// =============================================================================

// MAC creates a MAC address schema with default colon delimiter
func MAC(params ...any) *ZodMAC[string] {
	return MACWithDelimiter(":", params...)
}

// MACPtr creates a MAC address schema for pointer types
func MACPtr(params ...any) *ZodMAC[*string] {
	return MACTyped[*string](":", params...)
}

// MACWithDelimiter creates a MAC address schema with custom delimiter
func MACWithDelimiter(delimiter string, params ...any) *ZodMAC[string] {
	return MACTyped[string](delimiter, params...)
}

// MACTyped creates a MAC address schema with custom delimiter and type
func MACTyped[T NetworkConstraint](delimiter string, params ...any) *ZodMAC[T] {
	if delimiter == "" {
		delimiter = ":"
	}
	schemaParams := utils.NormalizeParams(params...)
	def := &ZodMACDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:     core.ZodTypeMAC,
			Required: true,
			Checks:   []core.ZodCheck{checks.MACWithDelimiter(delimiter)},
		},
		Delimiter: delimiter,
	}

	if schemaParams != nil {
		utils.ApplySchemaParams(&def.ZodTypeDef, schemaParams)
	}

	internals := &ZodMACInternals{
		ZodTypeInternals: core.ZodTypeInternals{
			Type:   core.ZodTypeMAC,
			Checks: def.Checks,
			Bag:    make(map[string]any),
		},
		Def: def,
	}

	return &ZodMAC[T]{internals: internals}
}

// =============================================================================
// E.164 PHONE NUMBER VALIDATION
// =============================================================================

// ZodE164Def defines the configuration for E.164 phone number validation
type ZodE164Def struct {
	core.ZodTypeDef
}

// ZodE164Internals contains E.164 validator internal state
type ZodE164Internals struct {
	core.ZodTypeInternals
	Def *ZodE164Def
}

// Clone creates a deep copy of internals
func (i *ZodE164Internals) Clone() *core.ZodTypeInternals {
	return i.ZodTypeInternals.Clone()
}

// ZodE164 represents an E.164 phone number validation schema
type ZodE164[T NetworkConstraint] struct {
	internals *ZodE164Internals
}

// =============================================================================
// E164 CORE METHODS
// =============================================================================

func (z *ZodE164[T]) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

func (z *ZodE164[T]) IsOptional() bool {
	return z.internals.Optional
}

func (z *ZodE164[T]) IsNilable() bool {
	return z.internals.Nilable
}

func (z *ZodE164[T]) Coerce(input any) (any, bool) {
	return input, false
}

func (z *ZodE164[T]) Parse(input any, ctx ...*core.ParseContext) (T, error) {
	return engine.ParsePrimitive[string, T](
		input,
		&z.internals.ZodTypeInternals,
		core.ZodTypeE164,
		engine.ApplyChecks[string],
		engine.ConvertToConstraintType[string, T],
		ctx...,
	)
}

func (z *ZodE164[T]) MustParse(input any, ctx ...*core.ParseContext) T {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

func (z *ZodE164[T]) StrictParse(input T, ctx ...*core.ParseContext) (T, error) {
	return engine.ParsePrimitiveStrict[string, T](
		input,
		&z.internals.ZodTypeInternals,
		core.ZodTypeE164,
		engine.ApplyChecks[string],
		ctx...,
	)
}

func (z *ZodE164[T]) MustStrictParse(input T, ctx ...*core.ParseContext) T {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

func (z *ZodE164[T]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// =============================================================================
// E164 MODIFIER METHODS
// =============================================================================

func (z *ZodE164[T]) Optional() *ZodE164[*string] {
	in := z.internals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

func (z *ZodE164[T]) Nilable() *ZodE164[*string] {
	in := z.internals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

func (z *ZodE164[T]) Default(v string) *ZodE164[T] {
	in := z.internals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

func (z *ZodE164[T]) Describe(description string) *ZodE164[T] {
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
// E164 REFINEMENT METHODS
// =============================================================================

// Refine adds type-safe custom validation logic
func (z *ZodE164[T]) Refine(fn func(T) bool, params ...any) *ZodE164[T] {
	wrapper := func(v any) bool {
		var zero T

		switch any(zero).(type) {
		case string:
			if v == nil {
				return false
			}
			if strVal, ok := v.(string); ok {
				return fn(any(strVal).(T))
			}
			return false
		case *string:
			if v == nil {
				return fn(any((*string)(nil)).(T))
			}
			if strVal, ok := v.(string); ok {
				sCopy := strVal
				ptr := &sCopy
				return fn(any(ptr).(T))
			}
			return false
		default:
			return false
		}
	}

	schemaParams := utils.NormalizeParams(params...)
	var errorMessage any
	if schemaParams.Error != nil {
		errorMessage = schemaParams.Error
	}

	check := checks.NewCustom[any](wrapper, errorMessage)
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// RefineAny adds flexible custom validation logic
func (z *ZodE164[T]) RefineAny(fn func(any) bool, params ...any) *ZodE164[T] {
	schemaParams := utils.NormalizeParams(params...)
	var errorMessage any
	if schemaParams.Error != nil {
		errorMessage = schemaParams.Error
	}

	check := checks.NewCustom[any](fn, errorMessage)
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// =============================================================================
// E164 COMPOSITION METHODS
// =============================================================================

func (z *ZodE164[T]) And(other any) *ZodIntersection[any, any] {
	return Intersection(z, other)
}

func (z *ZodE164[T]) Or(other any) *ZodUnion[any, any] {
	return Union([]any{z, other})
}

// =============================================================================
// E164 HELPER METHODS
// =============================================================================

func (z *ZodE164[T]) withPtrInternals(in *core.ZodTypeInternals) *ZodE164[*string] {
	return &ZodE164[*string]{internals: &ZodE164Internals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
	}}
}

func (z *ZodE164[T]) withInternals(in *core.ZodTypeInternals) *ZodE164[T] {
	return &ZodE164[T]{internals: &ZodE164Internals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
	}}
}

// =============================================================================
// E164 CONSTRUCTOR FUNCTIONS
// =============================================================================

// E164 creates an E.164 phone number schema
func E164(params ...any) *ZodE164[string] {
	return E164Typed[string](params...)
}

// E164Ptr creates an E.164 phone number schema for pointer types
func E164Ptr(params ...any) *ZodE164[*string] {
	return E164Typed[*string](params...)
}

// E164Typed creates an E.164 phone number schema with specific type
func E164Typed[T NetworkConstraint](params ...any) *ZodE164[T] {
	schemaParams := utils.NormalizeParams(params...)
	def := &ZodE164Def{
		ZodTypeDef: core.ZodTypeDef{
			Type:     core.ZodTypeE164,
			Required: true,
			Checks:   []core.ZodCheck{checks.E164()},
		},
	}

	if schemaParams != nil {
		utils.ApplySchemaParams(&def.ZodTypeDef, schemaParams)
	}

	internals := &ZodE164Internals{
		ZodTypeInternals: core.ZodTypeInternals{
			Type:   core.ZodTypeE164,
			Checks: def.Checks,
			Bag:    make(map[string]any),
		},
		Def: def,
	}

	return &ZodE164[T]{internals: internals}
}

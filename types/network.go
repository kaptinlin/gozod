package types

import (
	"regexp"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/internal/utils"
	"github.com/kaptinlin/gozod/pkg/coerce"
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
	return z.internals.ZodTypeInternals.IsOptional()
}

// IsNilable returns true if this schema accepts nil values
func (z *ZodIPv4[T]) IsNilable() bool {
	return z.internals.ZodTypeInternals.IsNilable()
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
	in := z.internals.ZodTypeInternals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// Nilable allows nil values, returns pointer type
func (z *ZodIPv4[T]) Nilable() *ZodIPv4[*string] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish combines optional and nilable modifiers
func (z *ZodIPv4[T]) Nullish() *ZodIPv4[*string] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Default preserves current generic type T
func (z *ZodIPv4[T]) Default(v string) *ZodIPv4[T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

// DefaultFunc preserves current generic type T
func (z *ZodIPv4[T]) DefaultFunc(fn func() string) *ZodIPv4[T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetDefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Prefault provides fallback values on validation failure
func (z *ZodIPv4[T]) Prefault(v string) *ZodIPv4[T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetPrefaultValue(v)
	return z.withInternals(in)
}

// PrefaultFunc provides dynamic fallback values
func (z *ZodIPv4[T]) PrefaultFunc(fn func() string) *ZodIPv4[T] {
	in := z.internals.ZodTypeInternals.Clone()
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
	newInternals := z.internals.ZodTypeInternals.Clone()
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
	newInternals := z.internals.ZodTypeInternals.Clone()
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
		originalChecks := z.internals.ZodTypeInternals.Checks
		*z.internals = *src.internals
		z.internals.ZodTypeInternals.Checks = originalChecks
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
	internals.ZodTypeInternals.Checks = append(internals.ZodTypeInternals.Checks, ipv4Check)

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
	schema.internals.ZodTypeInternals.Coerce = true
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
	return z.internals.ZodTypeInternals.IsOptional()
}

// IsNilable returns true if this schema accepts nil values
func (z *ZodIPv6[T]) IsNilable() bool {
	return z.internals.ZodTypeInternals.IsNilable()
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
	in := z.internals.ZodTypeInternals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// Nilable allows nil values, returns pointer type
func (z *ZodIPv6[T]) Nilable() *ZodIPv6[*string] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish combines optional and nilable modifiers
func (z *ZodIPv6[T]) Nullish() *ZodIPv6[*string] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Default preserves current generic type T
func (z *ZodIPv6[T]) Default(v string) *ZodIPv6[T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

// DefaultFunc preserves current generic type T
func (z *ZodIPv6[T]) DefaultFunc(fn func() string) *ZodIPv6[T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetDefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Prefault provides fallback values on validation failure
func (z *ZodIPv6[T]) Prefault(v string) *ZodIPv6[T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetPrefaultValue(v)
	return z.withInternals(in)
}

// PrefaultFunc provides dynamic fallback values
func (z *ZodIPv6[T]) PrefaultFunc(fn func() string) *ZodIPv6[T] {
	in := z.internals.ZodTypeInternals.Clone()
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
// IPv6 TYPE CONVERSION
// =============================================================================

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
	newInternals := z.internals.ZodTypeInternals.Clone()
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
	newInternals := z.internals.ZodTypeInternals.Clone()
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
		originalChecks := z.internals.ZodTypeInternals.Checks
		*z.internals = *src.internals
		z.internals.ZodTypeInternals.Checks = originalChecks
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
	internals.ZodTypeInternals.Checks = append(internals.ZodTypeInternals.Checks, ipv6Check)

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
	schema.internals.ZodTypeInternals.Coerce = true
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
	return z.internals.ZodTypeInternals.IsOptional()
}

// IsNilable returns true if this schema accepts nil values
func (z *ZodCIDRv4[T]) IsNilable() bool {
	return z.internals.ZodTypeInternals.IsNilable()
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
	in := z.internals.ZodTypeInternals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// Nilable allows nil values, returns pointer type
func (z *ZodCIDRv4[T]) Nilable() *ZodCIDRv4[*string] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish combines optional and nilable modifiers
func (z *ZodCIDRv4[T]) Nullish() *ZodCIDRv4[*string] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Default preserves current generic type T
func (z *ZodCIDRv4[T]) Default(v string) *ZodCIDRv4[T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

// DefaultFunc preserves current generic type T
func (z *ZodCIDRv4[T]) DefaultFunc(fn func() string) *ZodCIDRv4[T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetDefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Prefault provides fallback values on validation failure
func (z *ZodCIDRv4[T]) Prefault(v string) *ZodCIDRv4[T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetPrefaultValue(v)
	return z.withInternals(in)
}

// PrefaultFunc provides dynamic fallback values
func (z *ZodCIDRv4[T]) PrefaultFunc(fn func() string) *ZodCIDRv4[T] {
	in := z.internals.ZodTypeInternals.Clone()
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
// CIDRv4 TYPE CONVERSION
// =============================================================================

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
	newInternals := z.internals.ZodTypeInternals.Clone()
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
	newInternals := z.internals.ZodTypeInternals.Clone()
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
		originalChecks := z.internals.ZodTypeInternals.Checks
		*z.internals = *src.internals
		z.internals.ZodTypeInternals.Checks = originalChecks
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
	internals.ZodTypeInternals.Checks = append(internals.ZodTypeInternals.Checks, cidrv4Check)

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
	schema.internals.ZodTypeInternals.Coerce = true
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
	return z.internals.ZodTypeInternals.IsOptional()
}

// IsNilable returns true if this schema accepts nil values
func (z *ZodCIDRv6[T]) IsNilable() bool {
	return z.internals.ZodTypeInternals.IsNilable()
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
	in := z.internals.ZodTypeInternals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// Nilable allows nil values, returns pointer type
func (z *ZodCIDRv6[T]) Nilable() *ZodCIDRv6[*string] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish combines optional and nilable modifiers
func (z *ZodCIDRv6[T]) Nullish() *ZodCIDRv6[*string] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Default preserves current generic type T
func (z *ZodCIDRv6[T]) Default(v string) *ZodCIDRv6[T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

// DefaultFunc preserves current generic type T
func (z *ZodCIDRv6[T]) DefaultFunc(fn func() string) *ZodCIDRv6[T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetDefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Prefault provides fallback values on validation failure
func (z *ZodCIDRv6[T]) Prefault(v string) *ZodCIDRv6[T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetPrefaultValue(v)
	return z.withInternals(in)
}

// PrefaultFunc provides dynamic fallback values
func (z *ZodCIDRv6[T]) PrefaultFunc(fn func() string) *ZodCIDRv6[T] {
	in := z.internals.ZodTypeInternals.Clone()
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
	newInternals := z.internals.ZodTypeInternals.Clone()
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
	newInternals := z.internals.ZodTypeInternals.Clone()
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
		originalChecks := z.internals.ZodTypeInternals.Checks
		*z.internals = *src.internals
		z.internals.ZodTypeInternals.Checks = originalChecks
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
	internals.ZodTypeInternals.Checks = append(internals.ZodTypeInternals.Checks, cidrv6Check)

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
	schema.internals.ZodTypeInternals.Coerce = true
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
	return z.internals.ZodTypeInternals.IsOptional()
}

// IsNilable returns true if this schema accepts nil values
func (z *ZodURL[T]) IsNilable() bool {
	return z.internals.ZodTypeInternals.IsNilable()
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
	in := z.internals.ZodTypeInternals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// Nilable allows nil values, returns pointer type
func (z *ZodURL[T]) Nilable() *ZodURL[*string] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish combines optional and nilable modifiers
func (z *ZodURL[T]) Nullish() *ZodURL[*string] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Default preserves current generic type T
func (z *ZodURL[T]) Default(v string) *ZodURL[T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

// DefaultFunc preserves current generic type T
func (z *ZodURL[T]) DefaultFunc(fn func() string) *ZodURL[T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetDefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Prefault provides fallback values on validation failure
func (z *ZodURL[T]) Prefault(v string) *ZodURL[T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetPrefaultValue(v)
	return z.withInternals(in)
}

// PrefaultFunc provides dynamic fallback values
func (z *ZodURL[T]) PrefaultFunc(fn func() string) *ZodURL[T] {
	in := z.internals.ZodTypeInternals.Clone()
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
	newInternals := z.internals.ZodTypeInternals.Clone()
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
	newInternals := z.internals.ZodTypeInternals.Clone()
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
	schema.internals.ZodTypeInternals.Coerce = true
	return schema
}

package core

import (
	"fmt"
	"regexp"
	"sync/atomic"
)

// Static error variables to comply with err113
var (
	ErrSchemaNotZodSchema = fmt.Errorf("schema does not implement ZodSchema interface")
)

// =============================================================================
// CORE SCHEMA INTERFACE & INTERNALS
// =============================================================================

// Global priority counter for modifier application order tracking
var modifierPriorityCounter int64

//
// This file defines the **public contract** for every schema in the `gozod`
// ecosystem. It establishes a minimal, universal interface that ensures all
// schema types, whether built-in or custom, behave consistently.
//
// Key Abstractions:
//
//	1. ZodType[T] - The primary, generic interface for compile-time type safety.
//	   All schemas implement this.
//
//	2. ZodSchema - A non-generic counterpart for runtime-only scenarios where
//	   the static type is unknown (e.g., validating fields of a map[string]any).
//
//	3. ZodTypeInternals - The internal state backing every schema. While it is
//	   "internal," it is exposed via the `GetInternals()` method to enable
//	   advanced composition and introspection.

// -----------------------------------------------------------------------------
// PUBLIC SCHEMA INTERFACES
// -----------------------------------------------------------------------------

// ZodType is the universal interface for all validation schemas. It is the
// cornerstone of `gozod`, providing a consistent API for any data type.
//
// Generic Parameter:
//   - T: The output type that this schema validates and returns.
type ZodType[T any] interface {
	// Parse validates the input against this schema and returns the typed result.
	// This is the core validation method.
	Parse(input any, ctx ...*ParseContext) (T, error)

	// MustParse is a convenience method that validates input and panics on error.
	// It simplifies code where validation is expected to succeed.
	MustParse(input any, ctx ...*ParseContext) T

	// GetInternals provides access to the internal state of this schema,
	// allowing for advanced composition and framework integration.
	GetInternals() *ZodTypeInternals

	// IsOptional returns true if this schema accepts undefined/missing values.
	// This is a convenience method equivalent to GetInternals().IsOptional().
	IsOptional() bool

	// IsNilable returns true if this schema accepts nil values.
	// This is a convenience method equivalent to GetInternals().IsNilable().
	IsNilable() bool
}

// ZodSchema is a non-generic version of the schema interface, used for
// dynamic validation at runtime when the specific type `T` is not known.
type ZodSchema interface {
	// ParseAny validates input and returns an untyped `any` result. This is
	// crucial for dynamic structures like maps or objects.
	ParseAny(input any, ctx ...*ParseContext) (any, error)

	// GetInternals provides access to the internal state of this schema.
	GetInternals() *ZodTypeInternals
}

// -----------------------------------------------------------------------------
// CLONEABLE INTERFACE
// -----------------------------------------------------------------------------

// Cloneable is an optional interface for schemas that need to support
// deep-copying of their internal state. This is essential for modifiers
// like `.Optional()` or `.Transform()` that create new schema instances
// without modifying the original.
type Cloneable interface {
	// CloneFrom copies configuration and state from a source schema instance.
	// The source should be of the same concrete type.
	CloneFrom(source any)
}

// -----------------------------------------------------------------------------
// INTERNAL SCHEMA STATE
// -----------------------------------------------------------------------------

// ZodTypeInternals holds the complete configuration and state for any schema.
// It acts as the backbone, storing everything from validation checks to default
// values and custom error maps.
type ZodTypeInternals struct {
	Type   ZodTypeCode                                                  // Type identifier using type-safe constants
	Checks []ZodCheck                                                   // List of validation checks to apply
	Parse  func(payload *ParsePayload, ctx *ParseContext) *ParsePayload // The core parsing function for the type

	// Core validation flags
	Coerce      bool // Whether to enable type coercion
	Optional    bool // Whether the field is optional
	Nilable     bool // Whether nil values are allowed
	NonOptional bool // True if .NonOptional() was applied, for error reporting

	// Default/Prefault values
	DefaultValue  any
	DefaultFunc   func() any
	PrefaultValue any
	PrefaultFunc  func() any

	// Modifier priority tracking (higher number = applied later = higher priority)
	OptionalPriority int // Priority when Optional/Nilable was applied
	PrefaultPriority int // Priority when Prefault was applied
	DefaultPriority  int // Priority when Default was applied

	// Transform function
	Transform func(any, *RefinementContext) (any, error)

	// Optionality configuration
	OptIn  string // Optionality mode input
	OptOut string // Optionality mode output

	// Constructor and configuration
	Constructor func(def *ZodTypeDef) ZodType[any] // Factory function
	Values      map[any]struct{}                   // Valid values for literal types
	Pattern     *regexp.Regexp                     // Regex pattern for string validation
	Error       *ZodErrorMap                       // Custom error mapping
	Bag         map[string]any                     // Additional configuration storage
}

// =============================================================================
// ZODTYPEINTERNALS CONVENIENCE METHODS
// =============================================================================

// Clone creates a deep copy of the internals for immutable modifications.
func (z *ZodTypeInternals) Clone() *ZodTypeInternals {
	if z == nil {
		return nil
	}

	cp := *z // Shallow copy the struct

	// Deep copy Checks slice
	if len(z.Checks) > 0 {
		cp.Checks = make([]ZodCheck, len(z.Checks))
		copy(cp.Checks, z.Checks)
	}

	// Deep copy Values map
	if len(z.Values) > 0 {
		cp.Values = make(map[any]struct{}, len(z.Values))
		for k, v := range z.Values {
			cp.Values[k] = v
		}
	}

	// Deep copy Bag map
	if len(z.Bag) > 0 {
		cp.Bag = make(map[string]any, len(z.Bag))
		for k, v := range z.Bag {
			cp.Bag[k] = v
		}
	}

	return &cp
}

// IsOptional returns true if the field is optional.
func (z *ZodTypeInternals) IsOptional() bool {
	return z.Optional
}

// IsNilable returns true if nil values are allowed.
func (z *ZodTypeInternals) IsNilable() bool {
	return z.Nilable
}

// IsCoerce returns true if type coercion is enabled.
func (z *ZodTypeInternals) IsCoerce() bool {
	return z.Coerce
}

// IsNonOptional returns true if the field is non-optional.
func (z *ZodTypeInternals) IsNonOptional() bool {
	return z.NonOptional
}

// SetOptional marks the field as optional.
func (z *ZodTypeInternals) SetOptional(value bool) {
	z.Optional = value
	if value {
		z.OptionalPriority = int(atomic.AddInt64(&modifierPriorityCounter, 1))
	}
}

// SetNilable allows nil values for this field.
func (z *ZodTypeInternals) SetNilable(value bool) {
	z.Nilable = value
	if value {
		z.OptionalPriority = int(atomic.AddInt64(&modifierPriorityCounter, 1))
	}
}

// SetNonOptional marks the field as nonoptional (disallow nil with custom expected tag).
func (z *ZodTypeInternals) SetNonOptional(value bool) {
	z.NonOptional = value
}

// SetCoerce enables type coercion for this field.
func (z *ZodTypeInternals) SetCoerce(value bool) {
	z.Coerce = value
}

// SetDefaultValue sets a default value.
func (z *ZodTypeInternals) SetDefaultValue(value any) {
	z.DefaultValue = value
	z.DefaultPriority = int(atomic.AddInt64(&modifierPriorityCounter, 1))
}

// SetDefaultFunc sets a default value function.
func (z *ZodTypeInternals) SetDefaultFunc(fn func() any) {
	z.DefaultFunc = fn
	z.DefaultPriority = int(atomic.AddInt64(&modifierPriorityCounter, 1))
}

// SetPrefaultValue sets a prefault value.
func (z *ZodTypeInternals) SetPrefaultValue(value any) {
	z.PrefaultValue = value
	z.PrefaultPriority = int(atomic.AddInt64(&modifierPriorityCounter, 1))
}

// SetPrefaultFunc sets a prefault value function.
func (z *ZodTypeInternals) SetPrefaultFunc(fn func() any) {
	z.PrefaultFunc = fn
	z.PrefaultPriority = int(atomic.AddInt64(&modifierPriorityCounter, 1))
}

// SetTransform sets a transform function.
func (z *ZodTypeInternals) SetTransform(fn func(any, *RefinementContext) (any, error)) {
	z.Transform = fn
}

// AddCheck adds a validation check to the internals.
func (z *ZodTypeInternals) AddCheck(check ZodCheck) {
	z.Checks = append(z.Checks, check)
}

// =============================================================================
// SCHEMA CONVERSION UTILITIES
// =============================================================================

// ConvertToZodSchema converts a value to the ZodSchema interface, returning
// an error if the type does not implement the interface.
func ConvertToZodSchema(schema any) (ZodSchema, error) {
	if zodSchema, ok := schema.(ZodSchema); ok {
		return zodSchema, nil
	}
	return nil, fmt.Errorf("%w: %T", ErrSchemaNotZodSchema, schema)
}

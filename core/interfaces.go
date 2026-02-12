// Package core provides the foundational types, interfaces, and constants
// shared by all schema implementations in the gozod validation library.
package core

import (
	"errors"
	"fmt"
	"maps"
	"regexp"
	"sync/atomic"
)

// ErrSchemaNotZodSchema is a sentinel error returned when a value does not
// implement the ZodSchema interface.
var ErrSchemaNotZodSchema = errors.New("schema does not implement ZodSchema interface")

// Global priority counter for modifier application order tracking.
// Only relative priority matters, not absolute values.
var modifierPriorityCounter int64

// ZodType is the universal interface for all validation schemas.
//
// Generic Parameter:
//   - T: The output type that this schema validates and returns.
type ZodType[T any] interface {
	// Parse validates the input and returns the typed result.
	Parse(input any, ctx ...*ParseContext) (T, error)

	// MustParse validates input and panics on error.
	MustParse(input any, ctx ...*ParseContext) T

	// Internals provides access to the internal state of this schema.
	Internals() *ZodTypeInternals

	// IsOptional reports whether this schema accepts missing values.
	IsOptional() bool

	// IsNilable reports whether this schema accepts nil values.
	IsNilable() bool
}

// ZodSchema is a non-generic version of the schema interface, used for
// dynamic validation when the specific type T is not known.
type ZodSchema interface {
	// ParseAny validates input and returns an untyped result.
	ParseAny(input any, ctx ...*ParseContext) (any, error)

	// Internals provides access to the internal state of this schema.
	Internals() *ZodTypeInternals
}

// Cloneable is an optional interface for schemas that support deep-copying.
// This is essential for copy-on-write modifiers like .Optional() or .Transform().
type Cloneable interface {
	// CloneFrom copies configuration from a source schema instance.
	CloneFrom(source any)
}

// ZodTypeInternals holds the complete configuration and state for any schema.
type ZodTypeInternals struct {
	Type   ZodTypeCode                                                  // Type identifier
	Checks []ZodCheck                                                   // Validation checks to apply
	Parse  func(payload *ParsePayload, ctx *ParseContext) *ParsePayload // Core parsing function

	// Core validation flags
	Coerce        bool // Whether to enable type coercion
	Optional      bool // Whether the field is optional
	Nilable       bool // Whether nil values are allowed
	NonOptional   bool // True if .NonOptional() was applied
	ExactOptional bool // True if .ExactOptional() was applied

	// Default/Prefault values
	DefaultValue  any
	DefaultFunc   func() any
	PrefaultValue any
	PrefaultFunc  func() any

	// Modifier priority tracking (higher = applied later)
	OptionalPriority int
	PrefaultPriority int
	DefaultPriority  int

	// Transform function
	Transform func(any, *RefinementContext) (any, error)

	// Optionality configuration
	OptIn  string
	OptOut string

	// Constructor and configuration
	Constructor func(def *ZodTypeDef) ZodType[any]
	Values      map[any]struct{}
	Pattern     *regexp.Regexp
	Error       *ZodErrorMap
	Bag         map[string]any
}

// Clone creates a deep copy of the internals for copy-on-write modifications.
//
// Deep copied: Checks slice, Values map, Bag map.
// Shared (immutable): function pointers, Pattern, Error.
func (z *ZodTypeInternals) Clone() *ZodTypeInternals {
	if z == nil {
		return nil
	}
	cp := *z
	if len(z.Checks) > 0 {
		cp.Checks = make([]ZodCheck, len(z.Checks))
		copy(cp.Checks, z.Checks)
	}
	if len(z.Values) > 0 {
		cp.Values = maps.Clone(z.Values)
	}
	if len(z.Bag) > 0 {
		cp.Bag = maps.Clone(z.Bag)
	}
	return &cp
}

// IsOptional reports whether the field is optional.
func (z *ZodTypeInternals) IsOptional() bool {
	return z.Optional
}

// IsNilable reports whether nil values are allowed.
func (z *ZodTypeInternals) IsNilable() bool {
	return z.Nilable
}

// IsCoerce reports whether type coercion is enabled.
func (z *ZodTypeInternals) IsCoerce() bool {
	return z.Coerce
}

// IsNonOptional reports whether the field is non-optional.
func (z *ZodTypeInternals) IsNonOptional() bool {
	return z.NonOptional
}

// IsExactOptional reports whether exact optional mode is enabled.
func (z *ZodTypeInternals) IsExactOptional() bool {
	return z.ExactOptional
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

// SetNonOptional marks the field as nonoptional.
func (z *ZodTypeInternals) SetNonOptional(value bool) {
	z.NonOptional = value
}

// SetExactOptional enables exact optional mode.
// Also sets Optional=true because ExactOptional implies Optional
// for absent key handling.
func (z *ZodTypeInternals) SetExactOptional(value bool) {
	z.ExactOptional = value
	if value {
		z.Optional = true
		z.OptionalPriority = int(atomic.AddInt64(&modifierPriorityCounter, 1))
	}
}

// SetCoerce enables type coercion.
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

// AddCheck adds a validation check.
func (z *ZodTypeInternals) AddCheck(check ZodCheck) {
	z.Checks = append(z.Checks, check)
}

// ConvertToZodSchema converts a value to ZodSchema, returning
// an error if the type does not implement the interface.
func ConvertToZodSchema(schema any) (ZodSchema, error) {
	if zodSchema, ok := schema.(ZodSchema); ok {
		return zodSchema, nil
	}
	return nil, fmt.Errorf("%w: %T", ErrSchemaNotZodSchema, schema)
}

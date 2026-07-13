// Package core provides the foundational types, interfaces, and constants
// shared by all schema implementations in the gozod validation library.
package core

import (
	"errors"
	"fmt"
	"maps"
	"regexp"
	"slices"

	"github.com/kaptinlin/gozod/pkg/cloneutil"
)

// ErrSchemaNotZodSchema is a sentinel error returned when a value does not
// implement the ZodSchema interface.
var ErrSchemaNotZodSchema = errors.New("schema does not implement ZodSchema interface")

// ZodModifierKind identifies an ordered fluent modifier.
type ZodModifierKind string

const (
	// ZodModifierOptional records an Optional modifier.
	ZodModifierOptional ZodModifierKind = "optional"
	// ZodModifierNilable records a Nilable modifier.
	ZodModifierNilable ZodModifierKind = "nilable"
	// ZodModifierNonOptional records a NonOptional modifier.
	ZodModifierNonOptional ZodModifierKind = "nonoptional"
	// ZodModifierExactOptional records an ExactOptional modifier.
	ZodModifierExactOptional ZodModifierKind = "exact_optional"
	// ZodModifierDefault records a Default modifier.
	ZodModifierDefault ZodModifierKind = "default"
	// ZodModifierPrefault records a Prefault modifier.
	ZodModifierPrefault ZodModifierKind = "prefault"
)

// ZodModifier records one fluent modifier application in call order.
type ZodModifier struct {
	Kind     ZodModifierKind
	Value    any
	HasValue bool
	Func     func() any
}

// ZodSchema is the minimal non-generic runtime contract shared by all schemas.
//
// It is intentionally limited to the capabilities required by dynamic runtime
// code paths:
//   - parse an untyped input
//   - inspect shared runtime flags
//   - access core internals used by the engine and registries
//
// Fluent methods such as Describe/Meta/RefineAny and compile-time constrained
// methods such as StrictParse are intentionally not part of this boundary.
// Those capabilities live on typed schemas or on separate generic constraints.
type ZodSchema interface {
	// ParseAny validates input and returns an untyped result.
	ParseAny(input any, ctx ...*ParseContext) (any, error)

	// Internals provides access to the internal state of this schema.
	Internals() *ZodTypeInternals

	// IsOptional reports whether this schema accepts missing values.
	IsOptional() bool

	// IsNilable reports whether this schema accepts nil values.
	IsNilable() bool
}

// ZodType is the typed parsing contract layered on top of ZodSchema.
//
// This is the boundary most concrete schema implementations satisfy. It keeps
// runtime parsing typed without forcing every dynamic code path to depend on
// compile-time constrained parsing or fluent chaining contracts.
//
// Generic Parameter:
//   - T: The output type that this schema validates and returns.
type ZodType[T any] interface {
	ZodSchema

	// Parse validates the input and returns the typed result.
	Parse(input any, ctx ...*ParseContext) (T, error)

	// MustParse validates input and panics on error.
	MustParse(input any, ctx ...*ParseContext) T
}

// StrictZodType extends ZodType with compile-time constrained parsing.
//
// Input and output are separate because many schemas accept a base type T
// while returning a constraint type R such as *T. This interface models the
// "strict side" of the API surface only; fluent metadata and refinement
// contracts remain separate.
type StrictZodType[In, Out any] interface {
	ZodType[Out]

	// StrictParse validates a compile-time constrained input.
	StrictParse(input In, ctx ...*ParseContext) (Out, error)

	// MustStrictParse validates constrained input and panics on error.
	MustStrictParse(input In, ctx ...*ParseContext) Out
}

// Cloneable is an optional interface for schemas that support deep-copying.
// This is essential for copy-on-write modifiers like .Optional() or .Transform().
type Cloneable interface {
	// CloneFrom copies configuration from a source schema instance.
	CloneFrom(source any)
}

// ZodTypeInternals holds the complete configuration and state for any schema.
type ZodTypeInternals struct {
	Type   ZodTypeCode // Type identifier
	Checks []ZodCheck  // Validation checks to apply
	Parse  func(payload *ParsePayload, ctx *ParseContext) *ParsePayload

	// Core validation flags
	Coerce bool

	// Ordered fluent modifier applications.
	Modifiers []ZodModifier

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
	metadata    GlobalMeta
}

// Clone creates a deep copy of the internals for copy-on-write modifications.
// Deep copied: Checks slice, Values map, Bag map and values.
// Shared (immutable): function pointers, Pattern, Error.
func (z *ZodTypeInternals) Clone() *ZodTypeInternals {
	if z == nil {
		return nil
	}
	cp := *z
	if len(z.Checks) > 0 {
		cp.Checks = slices.Clone(z.Checks)
	}
	if len(z.Modifiers) > 0 {
		cp.Modifiers = slices.Clone(z.Modifiers)
		for i := range cp.Modifiers {
			if cp.Modifiers[i].HasValue {
				cp.Modifiers[i].Value = cloneutil.Clone(cp.Modifiers[i].Value)
			}
		}
	}
	if len(z.Values) > 0 {
		cp.Values = maps.Clone(z.Values)
	}
	if len(z.Bag) > 0 {
		cp.Bag = make(map[string]any, len(z.Bag))
		for key, value := range z.Bag {
			cp.Bag[key] = cloneutil.Clone(value)
		}
	}
	cp.metadata = cloneGlobalMeta(z.metadata)
	return &cp
}

// Metadata returns a defensive copy of the schema's standard metadata.
func (z *ZodTypeInternals) Metadata() GlobalMeta {
	if z == nil {
		return GlobalMeta{}
	}
	return cloneGlobalMeta(z.metadata)
}

// SetMetadata replaces the schema's standard metadata with a defensive copy.
func (z *ZodTypeInternals) SetMetadata(meta GlobalMeta) {
	if z == nil {
		return
	}
	z.metadata = cloneGlobalMeta(meta)
}

// MergeMetadata overlays non-zero fields onto the schema's standard metadata.
func (z *ZodTypeInternals) MergeMetadata(update GlobalMeta) {
	if z == nil {
		return
	}
	z.SetMetadata(MergeGlobalMeta(z.metadata, update))
}

func cloneGlobalMeta(meta GlobalMeta) GlobalMeta {
	if meta.Examples == nil {
		return meta
	}
	examples := make([]any, len(meta.Examples))
	for i, example := range meta.Examples {
		examples[i] = cloneutil.Clone(example)
	}
	meta.Examples = examples
	return meta
}

// IsOptional reports whether the field is optional.
func (z *ZodTypeInternals) IsOptional() bool {
	for i := len(z.Modifiers) - 1; i >= 0; i-- {
		switch z.Modifiers[i].Kind {
		case ZodModifierOptional, ZodModifierExactOptional:
			return true
		case ZodModifierNonOptional:
			return false
		case ZodModifierNilable, ZodModifierDefault, ZodModifierPrefault:
		}
	}
	return false
}

// IsNilable reports whether nil values are allowed.
func (z *ZodTypeInternals) IsNilable() bool {
	for i := len(z.Modifiers) - 1; i >= 0; i-- {
		switch z.Modifiers[i].Kind {
		case ZodModifierNilable:
			return true
		case ZodModifierNonOptional, ZodModifierExactOptional:
			return false
		case ZodModifierOptional, ZodModifierDefault, ZodModifierPrefault:
		}
	}
	return false
}

// IsCoerce reports whether type coercion is enabled.
func (z *ZodTypeInternals) IsCoerce() bool {
	return z.Coerce
}

// IsNonOptional reports whether the field is non-optional.
func (z *ZodTypeInternals) IsNonOptional() bool {
	for i := len(z.Modifiers) - 1; i >= 0; i-- {
		switch z.Modifiers[i].Kind {
		case ZodModifierNonOptional:
			return true
		case ZodModifierOptional, ZodModifierNilable, ZodModifierExactOptional:
			return false
		case ZodModifierDefault, ZodModifierPrefault:
		}
	}
	return false
}

// IsExactOptional reports whether exact optional mode is enabled.
func (z *ZodTypeInternals) IsExactOptional() bool {
	for i := len(z.Modifiers) - 1; i >= 0; i-- {
		switch z.Modifiers[i].Kind {
		case ZodModifierExactOptional:
			return true
		case ZodModifierOptional, ZodModifierNilable, ZodModifierNonOptional,
			ZodModifierDefault, ZodModifierPrefault:
			return false
		}
	}
	return false
}

// NilInputUsesDefault reports whether nil input is claimed by an outer default modifier.
func (z *ZodTypeInternals) NilInputUsesDefault() bool {
	for i := len(z.Modifiers) - 1; i >= 0; i-- {
		switch modifier := z.Modifiers[i]; modifier.Kind {
		case ZodModifierDefault:
			return modifier.HasValue || modifier.Func != nil
		case ZodModifierPrefault, ZodModifierOptional, ZodModifierNilable,
			ZodModifierNonOptional, ZodModifierExactOptional:
			return false
		}
	}
	return false
}

// NilInputUsesFallback reports whether the outer nil-input modifier supplies a value.
func (z *ZodTypeInternals) NilInputUsesFallback() bool {
	for i := len(z.Modifiers) - 1; i >= 0; i-- {
		switch modifier := z.Modifiers[i]; modifier.Kind {
		case ZodModifierDefault, ZodModifierPrefault:
			return modifier.HasValue || modifier.Func != nil
		case ZodModifierOptional, ZodModifierNilable, ZodModifierNonOptional,
			ZodModifierExactOptional:
			return false
		}
	}
	return false
}

// SetOptional marks the field as optional.
func (z *ZodTypeInternals) SetOptional(value bool) {
	if value {
		z.Modifiers = append(z.Modifiers, ZodModifier{Kind: ZodModifierOptional})
	}
}

// SetNilable allows nil values for this field.
func (z *ZodTypeInternals) SetNilable(value bool) {
	if value {
		z.Modifiers = append(z.Modifiers, ZodModifier{Kind: ZodModifierNilable})
	}
}

// SetNonOptional marks the field as nonoptional.
func (z *ZodTypeInternals) SetNonOptional(value bool) {
	if value {
		z.Modifiers = append(z.Modifiers, ZodModifier{Kind: ZodModifierNonOptional})
	}
}

// SetExactOptional enables exact optional mode.
func (z *ZodTypeInternals) SetExactOptional(value bool) {
	if value {
		z.Modifiers = append(z.Modifiers, ZodModifier{Kind: ZodModifierExactOptional})
	}
}

// SetCoerce enables type coercion.
func (z *ZodTypeInternals) SetCoerce(value bool) {
	z.Coerce = value
}

// SetDefaultValue sets a default value.
func (z *ZodTypeInternals) SetDefaultValue(value any) {
	z.Modifiers = append(z.Modifiers, ZodModifier{
		Kind:     ZodModifierDefault,
		Value:    cloneutil.Clone(value),
		HasValue: true,
	})
}

// SetDefaultFunc sets a default value function.
func (z *ZodTypeInternals) SetDefaultFunc(fn func() any) {
	z.Modifiers = append(z.Modifiers, ZodModifier{
		Kind: ZodModifierDefault,
		Func: fn,
	})
}

// SetPrefaultValue sets a prefault value.
func (z *ZodTypeInternals) SetPrefaultValue(value any) {
	z.Modifiers = append(z.Modifiers, ZodModifier{
		Kind:     ZodModifierPrefault,
		Value:    cloneutil.Clone(value),
		HasValue: true,
	})
}

// SetPrefaultFunc sets a prefault value function.
func (z *ZodTypeInternals) SetPrefaultFunc(fn func() any) {
	z.Modifiers = append(z.Modifiers, ZodModifier{
		Kind: ZodModifierPrefault,
		Func: fn,
	})
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

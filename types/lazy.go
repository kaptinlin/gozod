package types

import (
	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/internal/issues"
)

//////////////////////////
// CORE TYPE DEFINITIONS
//////////////////////////

// ZodLazyDef defines a lazy schema that gets its inner type from a function
type ZodLazyDef struct {
	core.ZodTypeDef
	Type   core.ZodTypeCode              // "lazy"
	Getter func() core.ZodType[any, any] // Schema getter function
}

// ZodLazyInternals contains lazy validator internal state with cached schema resolution
type ZodLazyInternals struct {
	core.ZodTypeInternals
	Def       *ZodLazyDef            // Definition
	InnerType core.ZodType[any, any] // Cached inner schema
	Cached    bool                   // Whether inner schema is cached
	Bag       map[string]any         // Additional metadata
}

// ZodLazy represents a lazy validation schema for recursive type definitions
type ZodLazy struct {
	internals *ZodLazyInternals
}

//////////////////////////
// CONSTRUCTOR FUNCTIONS
//////////////////////////

// Lazy creates a lazy schema that defers evaluation until needed, enabling recursive type definitions
func Lazy(getter func() core.ZodType[any, any]) core.ZodType[any, any] {
	def := &ZodLazyDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   "lazy",
			Checks: make([]core.ZodCheck, 0),
		},
		Type:   "lazy",
		Getter: getter,
	}

	return createZodLazyFromDef(def)
}

// createZodLazyFromDef creates a ZodLazy schema from definition
func createZodLazyFromDef(def *ZodLazyDef) core.ZodType[any, any] {
	internals := &ZodLazyInternals{
		ZodTypeInternals: engine.NewBaseZodTypeInternals(def.Type),
		Def:              def,
		Cached:           false,
		Bag:              make(map[string]any),
	}

	// Set up constructor for cloning support
	internals.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any, any] {
		lazyDef := &ZodLazyDef{
			ZodTypeDef: *newDef,
			Type:       "lazy",
			Getter:     def.Getter, // Preserve the original getter function
		}
		return createZodLazyFromDef(lazyDef)
	}

	// Set up parse function
	internals.Parse = func(payload *core.ParsePayload, ctx *core.ParseContext) *core.ParsePayload {
		// Handle Nilable nil values
		if payload.Value == nil && internals.Nilable {
			return payload
		}

		// Run lazy-level checks first
		payload = engine.RunChecks(payload, internals.Checks, ctx)
		if len(payload.Issues) > 0 {
			return payload
		}

		// Get inner type and delegate parsing
		inner := internals.getInnerType()
		if inner == nil {
			issue := issues.CreateInvalidTypeIssue(
				"lazy",
				payload.Value,
			)
			payload.Issues = append(payload.Issues, issue)
			return payload
		}

		// Delegate to inner schema for actual parsing
		return inner.GetInternals().Parse(payload, ctx)
	}

	lazySchema := &ZodLazy{internals: internals}
	engine.InitZodType(any(lazySchema).(core.ZodType[any, any]), &def.ZodTypeDef)
	return any(lazySchema).(core.ZodType[any, any])
}

// getInnerType implements lazy evaluation with caching
func (z *ZodLazyInternals) getInnerType() core.ZodType[any, any] {
	if !z.Cached {
		z.InnerType = z.Def.Getter()
		z.Cached = true
	}
	return z.InnerType
}

// Parse validates and parses input with smart type inference
// Delegates completely to inner schema while maintaining lazy evaluation
func (z *ZodLazy) Parse(input any, ctx ...*core.ParseContext) (any, error) {
	var parseCtx *core.ParseContext
	if len(ctx) > 0 {
		parseCtx = ctx[0]
	}

	payload := &core.ParsePayload{
		Value:  input,
		Issues: make([]core.ZodRawIssue, 0),
	}

	result := z.internals.Parse(payload, parseCtx)
	if len(result.Issues) > 0 {
		config := core.GetConfig()
		finalizedIssues := make([]core.ZodIssue, len(result.Issues))
		for i, rawIssue := range result.Issues {
			finalizedIssues[i] = issues.FinalizeIssue(rawIssue, parseCtx, config)
		}
		return nil, issues.NewZodError(finalizedIssues)
	}

	return result.Value, nil
}

// MustParse validates a value and panics on failure
func (z *ZodLazy) MustParse(input any, ctx ...*core.ParseContext) any {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// TransformAny provides flexible transformation
func (z *ZodLazy) TransformAny(fn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	transform := Transform[any, any](fn)
	return &ZodPipe[any, any]{
		in:  any(z).(core.ZodType[any, any]),
		out: any(transform).(core.ZodType[any, any]),
		def: core.ZodTypeDef{Type: "pipe"},
	}
}

// Pipe creates a validation pipeline
func (z *ZodLazy) Pipe(out core.ZodType[any, any]) core.ZodType[any, any] {
	return &ZodPipe[any, any]{
		in:  any(z).(core.ZodType[any, any]),
		out: out,
		def: core.ZodTypeDef{Type: "pipe"},
	}
}

//////////////////////////
// MODIFIER METHODS
//////////////////////////

// Optional makes the lazy schema optional
func (z *ZodLazy) Optional() core.ZodType[any, any] {
	return Optional(any(z).(core.ZodType[any, any]))
}

// Nilable makes the lazy schema nilable
func (z *ZodLazy) Nilable() core.ZodType[any, any] {
	cloned := engine.Clone(any(z).(core.ZodType[any, any]), func(def *core.ZodTypeDef) {
		// Nilable only changes nil handling
	})
	cloned.(*ZodLazy).internals.Nilable = true
	return cloned
}

// Nullish makes the lazy schema both optional and nilable
func (z *ZodLazy) Nullish() core.ZodType[any, any] {
	return Nullish(any(z).(core.ZodType[any, any]))
}

// Refine adds type-safe custom validation
func (z *ZodLazy) Refine(fn func(any) bool, params ...any) core.ZodType[any, any] {
	return z.RefineAny(fn, params...)
}

// RefineAny adds flexible custom validation
func (z *ZodLazy) RefineAny(fn func(any) bool, params ...any) core.ZodType[any, any] {
	check := checks.NewCustom[any](fn, params...)
	return engine.AddCheck(any(z).(core.ZodType[any, any]), check)
}

// Unwrap resolves and returns the inner schema
func (z *ZodLazy) Unwrap() core.ZodType[any, any] {
	return z.internals.getInnerType()
}

// GetInternals returns the internal structure
func (z *ZodLazy) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// GetZod returns the lazy-specific internals
func (z *ZodLazy) GetZod() *ZodLazyInternals {
	return z.internals
}

// CloneFrom implements the Cloneable interface
func (z *ZodLazy) CloneFrom(source any) {
	if src, ok := source.(*ZodLazy); ok {
		if src.internals.Bag != nil {
			z.internals.Bag = make(map[string]any)
			for k, v := range src.internals.Bag {
				z.internals.Bag[k] = v
			}
		}
	}
}

// Check adds a validation check to the lazy type
func (z *ZodLazy) Check(fn core.CheckFn) core.ZodType[any, any] {
	check := checks.NewCustom[any](func(v any) bool {
		payload := &core.ParsePayload{Value: v, Issues: make([]core.ZodRawIssue, 0)}
		fn(payload)
		return len(payload.Issues) == 0
	}, core.SchemaParams{})
	return engine.AddCheck(any(z).(core.ZodType[any, any]), check)
}

// Transform adds type-safe data transformation
func (z *ZodLazy) Transform(fn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	return z.TransformAny(fn)
}

//////////////////////////
// WRAPPER TYPES
//////////////////////////

// Lazy type uses generic wrappers - no specific wrapper types needed
// The lazy evaluation pattern delegates to inner types

// Default adds default value support for lazy type
func (z *ZodLazy) Default(value any) core.ZodType[any, any] {
	return Default(any(z).(core.ZodType[any, any]), value)
}

// DefaultFunc adds function default value support
func (z *ZodLazy) DefaultFunc(fn func() any) core.ZodType[any, any] {
	return DefaultFunc(any(z).(core.ZodType[any, any]), fn)
}

// Prefault adds fallback value support for lazy type
func (z *ZodLazy) Prefault(value any) core.ZodType[any, any] {
	return Prefault(any(z).(core.ZodType[any, any]), value)
}

// PrefaultFunc adds function fallback value support
func (z *ZodLazy) PrefaultFunc(fn func() any) core.ZodType[any, any] {
	return PrefaultFunc(any(z).(core.ZodType[any, any]), fn)
}

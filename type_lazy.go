package gozod

//////////////////////////
// CORE TYPE DEFINITIONS
//////////////////////////

// ZodLazyDef defines the configuration for lazy validation, enabling deferred schema evaluation for recursive types
type ZodLazyDef struct {
	ZodTypeDef
	Type   string                   // "lazy"
	Getter func() ZodType[any, any] // Schema getter function
}

// ZodLazyInternals contains lazy validator internal state with cached schema resolution
type ZodLazyInternals struct {
	ZodTypeInternals
	Def       *ZodLazyDef            // Definition
	InnerType ZodType[any, any]      // Cached inner schema
	Cached    bool                   // Whether inner schema is cached
	Bag       map[string]interface{} // Additional metadata
}

// ZodLazy represents a lazy validation schema for recursive type definitions
type ZodLazy struct {
	internals *ZodLazyInternals
}

//////////////////////////
// CONSTRUCTOR FUNCTIONS
//////////////////////////

// Lazy creates a lazy schema that defers evaluation until needed, enabling recursive type definitions
func Lazy(getter func() ZodType[any, any]) ZodType[any, any] {
	def := &ZodLazyDef{
		ZodTypeDef: ZodTypeDef{
			Type:   "lazy",
			Checks: make([]ZodCheck, 0),
		},
		Type:   "lazy",
		Getter: getter,
	}

	return createZodLazyFromDef(def)
}

// createZodLazyFromDef creates a ZodLazy schema from definition
func createZodLazyFromDef(def *ZodLazyDef) ZodType[any, any] {
	internals := &ZodLazyInternals{
		ZodTypeInternals: newBaseZodTypeInternals(def.Type),
		Def:              def,
		Cached:           false,
		Bag:              make(map[string]interface{}),
	}

	// Set up constructor for cloning support
	internals.Constructor = func(newDef *ZodTypeDef) ZodType[any, any] {
		lazyDef := &ZodLazyDef{
			ZodTypeDef: *newDef,
			Type:       "lazy",
			Getter:     def.Getter, // Preserve the original getter function
		}
		return createZodLazyFromDef(lazyDef)
	}

	// Set up parse function
	internals.Parse = func(payload *ParsePayload, ctx *ParseContext) *ParsePayload {
		// Handle Nilable nil values
		if payload.Value == nil && internals.Nilable {
			return payload
		}

		// Run lazy-level checks first
		payload = runChecks(payload, internals.Checks, ctx)
		if len(payload.Issues) > 0 {
			return payload
		}

		// Get inner type and delegate parsing
		inner := internals.getInnerType()
		if inner == nil {
			issue := CreateInvalidTypeIssue(
				payload.Value,
				"lazy",
				"failed to resolve lazy schema",
			)
			payload.Issues = append(payload.Issues, issue)
			return payload
		}

		// Delegate to inner schema for actual parsing
		return inner.GetInternals().Parse(payload, ctx)
	}

	lazySchema := &ZodLazy{internals: internals}
	initZodType(any(lazySchema).(ZodType[any, any]), &def.ZodTypeDef)
	return any(lazySchema).(ZodType[any, any])
}

// getInnerType implements lazy evaluation with caching
func (z *ZodLazyInternals) getInnerType() ZodType[any, any] {
	if !z.Cached {
		z.InnerType = z.Def.Getter()
		z.Cached = true
	}
	return z.InnerType
}

// Parse validates and parses input with smart type inference
// Delegates completely to inner schema while maintaining lazy evaluation
func (z *ZodLazy) Parse(input any, ctx ...*ParseContext) (any, error) {
	var parseCtx *ParseContext
	if len(ctx) > 0 {
		parseCtx = ctx[0]
	}

	payload := &ParsePayload{
		Value:  input,
		Issues: make([]ZodRawIssue, 0),
	}

	result := z.internals.Parse(payload, parseCtx)
	if len(result.Issues) > 0 {
		config := GetConfig()
		finalizedIssues := make([]ZodIssue, len(result.Issues))
		for i, rawIssue := range result.Issues {
			finalizedIssues[i] = FinalizeIssue(rawIssue, parseCtx, config)
		}
		return nil, NewZodError(finalizedIssues)
	}

	return result.Value, nil
}

// MustParse validates a value and panics on failure
func (z *ZodLazy) MustParse(input any, ctx ...*ParseContext) any {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// TransformAny provides flexible transformation
func (z *ZodLazy) TransformAny(fn func(any, *RefinementContext) (any, error)) ZodType[any, any] {
	transform := NewZodTransform[any, any](fn)
	return &ZodPipe[any, any]{
		in:  any(z).(ZodType[any, any]),
		out: any(transform).(ZodType[any, any]),
		def: ZodTypeDef{Type: "pipe"},
	}
}

// Pipe creates a validation pipeline
func (z *ZodLazy) Pipe(out ZodType[any, any]) ZodType[any, any] {
	return &ZodPipe[any, any]{
		in:  any(z).(ZodType[any, any]),
		out: out,
		def: ZodTypeDef{Type: "pipe"},
	}
}

//////////////////////////
// MODIFIER METHODS
//////////////////////////

// Optional makes the lazy schema optional
func (z *ZodLazy) Optional() ZodType[any, any] {
	return Optional(any(z).(ZodType[any, any]))
}

// Nilable makes the lazy schema nilable
func (z *ZodLazy) Nilable() ZodType[any, any] {
	cloned := Clone(any(z).(ZodType[any, any]), func(def *ZodTypeDef) {
		// Nilable only changes nil handling
	})
	cloned.(*ZodLazy).internals.Nilable = true
	return cloned
}

// Nullish makes the lazy schema both optional and nilable
func (z *ZodLazy) Nullish() ZodType[any, any] {
	return Nullish(any(z).(ZodType[any, any]))
}

// Refine adds type-safe custom validation
func (z *ZodLazy) Refine(fn func(any) bool, params ...SchemaParams) ZodType[any, any] {
	return z.RefineAny(fn, params...)
}

// RefineAny adds flexible custom validation
func (z *ZodLazy) RefineAny(fn func(any) bool, params ...SchemaParams) ZodType[any, any] {
	check := NewCustom[any](fn, params...)
	return AddCheck(any(z).(ZodType[any, any]), check)
}

// Unwrap resolves and returns the inner schema
func (z *ZodLazy) Unwrap() ZodType[any, any] {
	return z.internals.getInnerType()
}

// GetInternals returns the internal structure
func (z *ZodLazy) GetInternals() *ZodTypeInternals {
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
			z.internals.Bag = make(map[string]interface{})
			for k, v := range src.internals.Bag {
				z.internals.Bag[k] = v
			}
		}
	}
}

// Coerce attempts to coerce input (delegates to inner type)
func (z *ZodLazy) Coerce(input any, ctx ...*ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// Check adds a validation check to the lazy type
func (z *ZodLazy) Check(fn CheckFn) ZodType[any, any] {
	check := NewCustom[any](func(v any) bool {
		payload := &ParsePayload{Value: v, Issues: make([]ZodRawIssue, 0)}
		fn(payload)
		return len(payload.Issues) == 0
	}, SchemaParams{})
	return AddCheck(any(z).(ZodType[any, any]), check)
}

// Transform adds type-safe data transformation
func (z *ZodLazy) Transform(fn func(any, *RefinementContext) (any, error)) ZodType[any, any] {
	return z.TransformAny(fn)
}

//////////////////////////
// WRAPPER TYPES
//////////////////////////

// Lazy type uses generic wrappers - no specific wrapper types needed
// The lazy evaluation pattern delegates to inner types

// Default adds default value support for lazy type
func (z *ZodLazy) Default(value any) ZodType[any, any] {
	return Default(any(z).(ZodType[any, any]), value)
}

// DefaultFunc adds function default value support
func (z *ZodLazy) DefaultFunc(fn func() any) ZodType[any, any] {
	return DefaultFunc(any(z).(ZodType[any, any]), fn)
}

// Prefault adds fallback value support for lazy type
func (z *ZodLazy) Prefault(value any) ZodType[any, any] {
	return Prefault[any, any](any(z).(ZodType[any, any]), value)
}

// PrefaultFunc adds function fallback value support
func (z *ZodLazy) PrefaultFunc(fn func() any) ZodType[any, any] {
	return PrefaultFunc[any, any](any(z).(ZodType[any, any]), fn)
}

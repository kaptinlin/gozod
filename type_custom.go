package gozod

///////////////////////////
////   CUSTOM TYPE DEFINITIONS ////
///////////////////////////

// ZodCustomDef defines the configuration for custom validation schema
type ZodCustomDef struct {
	ZodTypeDef
	Type   string                 // "custom"
	Check  string                 // "custom"
	Path   []interface{}          // PropertyKey[]
	Params map[string]interface{} // Custom parameters
	Fn     interface{}            // Custom validation function (RefineFn or CheckFn)
	FnType string                 // Function type: "refine" or "check"
}

// ZodCustomInternals contains custom validator internal state
type ZodCustomInternals struct {
	ZodTypeInternals
	Def  *ZodCustomDef          // Custom validation definition
	Issc *ZodIssueBase          // The set of issues this check might throw
	Bag  map[string]interface{} // Contains Class and other metadata
}

// ZodCustom represents a custom validation schema
type ZodCustom struct {
	internals *ZodCustomInternals
}

// GetInternals returns the internal state of the schema
func (z *ZodCustom) GetInternals() *ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// GetZod returns the custom-specific internals (type-safe access)
func (z *ZodCustom) GetZod() *ZodCustomInternals {
	return z.internals
}

// CloneFrom implements Cloneable interface for type-specific state copying
func (z *ZodCustom) CloneFrom(source any) {
	if src, ok := source.(*ZodCustom); ok {
		// Deep copy all custom-specific state
		z.internals = &ZodCustomInternals{
			ZodTypeInternals: src.internals.ZodTypeInternals,
			Def:              src.internals.Def,
			Issc:             src.internals.Issc,
			Bag:              make(map[string]interface{}),
		}
		// Copy bag contents
		for k, v := range src.internals.Bag {
			z.internals.Bag[k] = v
		}
	}
}

// Coerce attempts to coerce input (for custom validation, passes through)
func (z *ZodCustom) Coerce(input any, ctx ...*ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// Parse validates input with smart type inference
func (z *ZodCustom) Parse(input any, ctx ...*ParseContext) (any, error) {
	// 1. Unified nil handling - the only effect of Nilable
	if input == nil {
		if !z.internals.Nilable {
			rawIssue := CreateInvalidTypeIssue(input, "custom", "null")
			finalIssue := FinalizeIssue(rawIssue, nil, GetConfig())
			return nil, NewZodError([]ZodIssue{finalIssue})
		}
		return (*any)(nil), nil
	}

	// 2. Execute custom validation - using unified payload pattern
	var parseCtx *ParseContext
	if len(ctx) > 0 {
		parseCtx = ctx[0]
	} else {
		parseCtx = &ParseContext{}
	}

	payload := &ParsePayload{
		Value:  input,
		Issues: make([]ZodRawIssue, 0),
		Path:   make([]interface{}, 0),
	}

	// 3. Run custom check function
	result := z.internals.Parse(payload, parseCtx)
	if len(result.Issues) > 0 {
		return nil, &ZodError{Issues: convertRawIssuesToIssues(result.Issues, parseCtx)}
	}

	// 4. Smart type inference: custom validation preserves original input type
	return result.Value, nil
}

// MustParse validates the input value and panics on failure
func (z *ZodCustom) MustParse(input any, ctx ...*ParseContext) any {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// Transform provides type-safe custom transformation with smart dereferencing support
// Custom validation can handle any type, so Transform accepts any type
func (z *ZodCustom) Transform(fn func(any, *RefinementContext) (any, error)) ZodType[any, any] {
	return z.TransformAny(fn)
}

// TransformAny flexible version of transformation
// Implements ZodType[any, any] interface: TransformAny(fn func(any, *RefinementContext) (any, error)) ZodType[any, any]
func (z *ZodCustom) TransformAny(fn func(any, *RefinementContext) (any, error)) ZodType[any, any] {
	// Create pure Transform
	transform := NewZodTransform[any, any](fn)

	return &ZodPipe[any, any]{
		in:  any(z).(ZodType[any, any]),
		out: any(transform).(ZodType[any, any]),
		def: ZodTypeDef{Type: "pipe"},
	}
}

// Refine provides type-safe custom validation
// Supports smart dereferencing: custom validation can handle any type
func (z *ZodCustom) Refine(fn func(any) bool, params ...SchemaParams) *ZodCustom {
	result := z.RefineAny(fn, params...)
	return result.(*ZodCustom)
}

// RefineAny flexible version of validation that accepts any type
// Implements ZodType[any, any] interface: RefineAny(fn func(any) bool, params ...SchemaParams) ZodType[any, any]
func (z *ZodCustom) RefineAny(fn func(any) bool, params ...SchemaParams) ZodType[any, any] {
	// Follow design principle: add check to current schema instead of creating new Custom schema
	// But ZodCustom itself is a check, so create chained composition
	var schemaParams SchemaParams
	if len(params) > 0 {
		schemaParams = params[0]
	}
	newCustom := NewZodCustom(RefineFn[interface{}](fn), schemaParams)

	// Return Pipe composition
	return &ZodPipe[any, any]{
		in:  any(z).(ZodType[any, any]),
		out: any(newCustom).(ZodType[any, any]),
		def: ZodTypeDef{Type: "pipe"},
	}
}

// Check adds modern check API
// Implements ZodType[any, any] interface
func (z *ZodCustom) Check(fn func(*ParsePayload) error) ZodType[any, any] {
	// Convert error-returning function to CheckFn type
	checkFn := func(payload *ParsePayload) {
		if err := fn(payload); err != nil {
			issue := NewRawIssue("custom", payload.Value, WithOrigin("custom"))
			issue.Message = err.Error()
			payload.Issues = append(payload.Issues, issue)
		}
	}

	newCustom := NewZodCustom(CheckFn(checkFn), SchemaParams{})

	// Return Pipe composition
	return &ZodPipe[any, any]{
		in:  any(z).(ZodType[any, any]),
		out: any(newCustom).(ZodType[any, any]),
		def: ZodTypeDef{Type: "pipe"},
	}
}

// Pipe creates a validation pipeline from this schema to another
func (z *ZodCustom) Pipe(out ZodType[any, any]) ZodType[any, any] {
	return &ZodPipe[any, any]{
		in:  any(z).(ZodType[any, any]),
		out: out,
		def: ZodTypeDef{Type: "pipe"},
	}
}

////////////////////////////
////   WRAPPER METHODS  ////
////////////////////////////

// Optional creates an optional version of this schema
func (z *ZodCustom) Optional() ZodType[any, any] {
	return any(Optional(any(z).(ZodType[any, any]))).(ZodType[any, any])
}

// Nilable implements ZodType[any, any] interface: Nilable() ZodType[any, any]
func (z *ZodCustom) Nilable() ZodType[any, any] {
	return z.setNilable()
}

func (z *ZodCustom) setNilable() ZodType[any, any] {
	// Create a new ZodCustom with Nilable flag set instead of using Clone
	newInternals := &ZodCustomInternals{
		ZodTypeInternals: z.internals.ZodTypeInternals,
		Def:              z.internals.Def,
		Bag:              make(map[string]interface{}),
	}

	// Copy the bag
	for k, v := range z.internals.Bag {
		newInternals.Bag[k] = v
	}

	// Set nilable flag
	newInternals.Nilable = true

	// Copy the parse function
	newInternals.Parse = z.internals.Parse
	newInternals.Constructor = z.internals.Constructor

	return &ZodCustom{internals: newInternals}
}

// Nullish creates a nullish (optional and nilable) version of this schema
func (z *ZodCustom) Nullish() ZodType[any, any] {
	return any(Optional(NewZodNilable(any(z).(ZodType[any, any])))).(ZodType[any, any])
}

////////////////////////////
////   CONSTRUCTOR FUNCTIONS ////
////////////////////////////

// createZodCustomFromDef creates a custom schema from definition
func createZodCustomFromDef(def *ZodCustomDef) *ZodCustom {
	internals := &ZodCustomInternals{
		ZodTypeInternals: newBaseZodTypeInternals("custom"),
		Def:              def,
		Bag:              make(map[string]interface{}),
	}

	// Set up constructor for cloning
	internals.Constructor = func(newDef *ZodTypeDef) ZodType[any, any] {
		customDef := &ZodCustomDef{
			ZodTypeDef: *newDef,
			Type:       def.Type,
			Check:      def.Check,
			Path:       def.Path,
			Params:     def.Params,
			Fn:         def.Fn,
			FnType:     def.FnType,
		}
		return createZodCustomFromDef(customDef)
	}

	// Set up parse function
	internals.Parse = func(payload *ParsePayload, ctx *ParseContext) *ParsePayload {
		input := payload.Value
		switch def.FnType {
		case "refine":
			refineFn := def.Fn.(RefineFn[interface{}])
			if !refineFn(input) {
				path := make([]interface{}, len(payload.Path))
				copy(path, payload.Path)
				if def.Path != nil {
					path = append(path, def.Path...)
				}
				options := []func(*ZodRawIssue){
					WithOrigin("custom"),
					WithPath(path),
				}
				if len(def.Params) > 0 {
					options = append(options, WithParams(def.Params))
				}
				issue := NewRawIssue("custom", input, options...)
				issue.Inst = internals

				// Apply custom error message if provided
				if def.Error != nil {
					errorMap := *def.Error
					customMessage := errorMap(issue)
					if customMessage != "" {
						issue.Message = customMessage
					}
				}

				payload.Issues = append(payload.Issues, issue)
			}
		case "check":
			checkFn := def.Fn.(CheckFn)
			checkFn(payload)
		}
		return payload
	}

	schema := &ZodCustom{internals: internals}
	initZodType(any(schema).(ZodType[any, any]), &def.ZodTypeDef)
	return schema
}

// NewZodCustom creates a new custom validation schema
func NewZodCustom(fn interface{}, params ...SchemaParams) *ZodCustom {
	def := &ZodCustomDef{
		ZodTypeDef: ZodTypeDef{Type: ZodTypeCustom, Checks: make([]ZodCheck, 0)},
		Type:       "custom",
		Check:      "custom",
		Params:     make(map[string]interface{}),
	}

	// Handle function parameter
	if fn != nil {
		// Determine function type and store
		switch f := fn.(type) {
		case CheckFn:
			def.Fn = f
			def.FnType = "check"
		case func(*ParsePayload):
			def.Fn = CheckFn(f)
			def.FnType = "check"
		case RefineFn[interface{}]:
			def.Fn = f
			def.FnType = "refine"
		case func(interface{}) bool:
			def.Fn = RefineFn[interface{}](f)
			def.FnType = "refine"
		default:
			// Try to convert to a generic refine function
			def.Fn = RefineFn[interface{}](func(interface{}) bool { return true })
			def.FnType = "refine"
		}
	} else {
		// Default function that always passes
		def.Fn = RefineFn[interface{}](func(interface{}) bool { return true })
		def.FnType = "refine"
	}

	schema := createZodCustomFromDef(def)

	// Apply parameters
	if len(params) > 0 {
		param := params[0]
		if param.Coerce {
			schema.internals.Bag["coerce"] = true
		}
		if param.Error != nil {
			errorMap := createErrorMap(param.Error)
			if errorMap != nil {
				def.Error = errorMap
				schema.internals.Error = errorMap
			}
		}
		// Handle path parameter
		if param.Path != nil {
			def.Path = make([]interface{}, len(param.Path))
			for i, p := range param.Path {
				def.Path[i] = p
			}
		}
		// Handle params parameter
		if param.Params != nil {
			def.Params = make(map[string]interface{})
			for k, v := range param.Params {
				def.Params[k] = v
			}
		}
	}

	return schema
}

////////////////////////////
////   PUBLIC FUNCTIONS ////
////////////////////////////

// Refine creates a custom refinement validation
func Refine[T any](fn func(T) bool, params ...SchemaParams) *ZodCustom {
	// Convert typed function to interface{} function
	genericFn := func(input interface{}) bool {
		if typed, ok := input.(T); ok {
			return fn(typed)
		}
		return false // Type mismatch fails validation
	}

	return NewZodCustom(RefineFn[interface{}](genericFn), params...)
}

// InstanceOf creates validation for Go type checking
func InstanceOf[T any](params ...SchemaParams) *ZodCustom {
	fn := func(input interface{}) bool {
		_, ok := input.(T)
		return ok
	}

	// Set default error message if not provided
	if len(params) == 0 || params[0].Error == nil {
		params = append(params, SchemaParams{
			Error: "Input is not of the expected type",
		})
	}

	custom := NewZodCustom(RefineFn[interface{}](fn), params...)

	// Store simple type information in the bag for debugging
	custom.internals.Bag["instanceOf"] = true

	return custom
}

// Check creates a low-level custom check function
func Check(fn func(*ParsePayload), params ...SchemaParams) *ZodCustom {
	return NewZodCustom(CheckFn(fn), params...)
}

////////////////////////////
////   HELPER FUNCTIONS ////
////////////////////////////

// Unwrap returns the inner type (for basic types, returns self)
func (z *ZodCustom) Unwrap() ZodType[any, any] {
	return any(z).(ZodType[any, any])
}

// ==================== CUSTOM DEFAULT/PREFAULT WRAPPER TYPES ====================

// ZodCustomDefault is a Default wrapper for Custom type
// Provides perfect type safety and chainable method support
type ZodCustomDefault struct {
	*ZodDefault[*ZodCustom] // Embed concrete pointer to enable method promotion
}

// ZodCustomPrefault is a Prefault wrapper for Custom type
// Provides perfect type safety and chainable method support
type ZodCustomPrefault struct {
	*ZodPrefault[*ZodCustom] // Embed concrete pointer to enable method promotion
}

// ==================== DEFAULT METHOD IMPLEMENTATIONS ====================

// Default adds a default value to the custom schema, returns ZodCustomDefault support chain call
// Compile-time type safety: Custom().Default(value)
func (z *ZodCustom) Default(value any) ZodCustomDefault {
	return ZodCustomDefault{
		&ZodDefault[*ZodCustom]{
			innerType:    z,
			defaultValue: value,
			isFunction:   false,
		},
	}
}

// DefaultFunc adds a default function to the custom schema, returns ZodCustomDefault support chain call
func (z *ZodCustom) DefaultFunc(fn func() any) ZodCustomDefault {
	genericFn := func() any { return fn() }
	return ZodCustomDefault{
		&ZodDefault[*ZodCustom]{
			innerType:   z,
			defaultFunc: genericFn,
			isFunction:  true,
		},
	}
}

// ==================== PREFAULT METHOD IMPLEMENTATIONS ====================

// Prefault adds a prefault value to the custom schema, returns ZodCustomPrefault support chain call
func (z *ZodCustom) Prefault(value any) ZodCustomPrefault {
	baseInternals := z.GetInternals()
	internals := &ZodTypeInternals{
		Version:     baseInternals.Version,
		Type:        ZodTypePrefault,
		Checks:      baseInternals.Checks,
		Coerce:      baseInternals.Coerce,
		Optional:    baseInternals.Optional,
		Nilable:     baseInternals.Nilable,
		Constructor: baseInternals.Constructor,
		Values:      baseInternals.Values,
		Pattern:     baseInternals.Pattern,
		Error:       baseInternals.Error,
		Bag:         baseInternals.Bag,
	}

	return ZodCustomPrefault{
		&ZodPrefault[*ZodCustom]{
			internals:     internals,
			innerType:     z,
			prefaultValue: value,
			prefaultFunc:  nil,
			isFunction:    false,
		},
	}
}

// PrefaultFunc adds a prefault function to the custom schema, returns ZodCustomPrefault support chain call
func (z *ZodCustom) PrefaultFunc(fn func() any) ZodCustomPrefault {
	baseInternals := z.GetInternals()
	internals := &ZodTypeInternals{
		Version:     baseInternals.Version,
		Type:        ZodTypePrefault,
		Checks:      baseInternals.Checks,
		Coerce:      baseInternals.Coerce,
		Optional:    baseInternals.Optional,
		Nilable:     baseInternals.Nilable,
		Constructor: baseInternals.Constructor,
		Values:      baseInternals.Values,
		Pattern:     baseInternals.Pattern,
		Error:       baseInternals.Error,
		Bag:         baseInternals.Bag,
	}

	genericFn := func() any { return fn() }
	return ZodCustomPrefault{
		&ZodPrefault[*ZodCustom]{
			internals:     internals,
			innerType:     z,
			prefaultValue: nil,
			prefaultFunc:  genericFn,
			isFunction:    true,
		},
	}
}

// ==================== ZodCustomDefault CHAINING METHODS ====================

// Refine adds a flexible validation function to the custom schema, returns ZodCustomDefault support chain call
func (s ZodCustomDefault) Refine(fn func(any) bool, params ...SchemaParams) ZodCustomDefault {
	newInner := s.innerType.Refine(fn, params...)
	return ZodCustomDefault{
		&ZodDefault[*ZodCustom]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

// Transform adds data transformation, returns a generic ZodType support transform pipeline
func (s ZodCustomDefault) Transform(fn func(any, *RefinementContext) (any, error)) ZodType[any, any] {
	return s.TransformAny(fn)
}

// Check adds a modern validation function to the custom schema, returns ZodCustomDefault support chain call
func (s ZodCustomDefault) Check(fn func(*ParsePayload) error) ZodCustomDefault {
	// Convert Check to Refine call
	refineFn := func(input any) bool {
		payload := &ParsePayload{Value: input}
		err := fn(payload)
		return err == nil
	}
	newInner := s.innerType.Refine(refineFn)
	return ZodCustomDefault{
		&ZodDefault[*ZodCustom]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

// Optional adds an optional check to the custom schema, returns ZodType support chain call
func (s ZodCustomDefault) Optional() ZodType[any, any] {
	return Optional(any(s).(ZodType[any, any]))
}

// Nilable adds a nilable check to the custom schema, returns ZodType support chain call
func (s ZodCustomDefault) Nilable() ZodType[any, any] {
	return Nilable(any(s).(ZodType[any, any]))
}

// ==================== ZodCustomPrefault CHAINING METHODS ====================

// Refine adds a flexible validation function to the custom schema, returns ZodCustomPrefault support chain call
func (s ZodCustomPrefault) Refine(fn func(any) bool, params ...SchemaParams) ZodCustomPrefault {
	newInner := s.innerType.Refine(fn, params...)
	return ZodCustomPrefault{
		&ZodPrefault[*ZodCustom]{
			internals:     s.internals,
			innerType:     newInner,
			prefaultValue: s.prefaultValue,
			prefaultFunc:  s.prefaultFunc,
			isFunction:    s.isFunction,
		},
	}
}

// Transform adds data transformation, returns a generic ZodType support transform pipeline
func (s ZodCustomPrefault) Transform(fn func(any, *RefinementContext) (any, error)) ZodType[any, any] {
	return s.TransformAny(fn)
}

// Check adds a modern validation function to the custom schema, returns ZodCustomPrefault support chain call
func (s ZodCustomPrefault) Check(fn func(*ParsePayload) error) ZodCustomPrefault {
	// Convert Check to Refine call
	refineFn := func(input any) bool {
		payload := &ParsePayload{Value: input}
		err := fn(payload)
		return err == nil
	}
	newInner := s.innerType.Refine(refineFn)
	return ZodCustomPrefault{
		&ZodPrefault[*ZodCustom]{
			internals:     s.internals,
			innerType:     newInner,
			prefaultValue: s.prefaultValue,
			prefaultFunc:  s.prefaultFunc,
			isFunction:    s.isFunction,
		},
	}
}

// Optional adds an optional check to the custom schema, returns ZodType support chain call
func (s ZodCustomPrefault) Optional() ZodType[any, any] {
	return Optional(any(s).(ZodType[any, any]))
}

// Nilable adds a nilable check to the custom schema, returns ZodType support chain call
func (s ZodCustomPrefault) Nilable() ZodType[any, any] {
	return Nilable(any(s).(ZodType[any, any]))
}

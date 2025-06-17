package gozod

import "fmt"

//////////////////////////
////   DISCRIMINATED UNION TYPES   ////
//////////////////////////

// ZodDiscriminatedUnionDef defines the configuration for discriminated union validation
type ZodDiscriminatedUnionDef struct {
	ZodUnionDef
	Discriminator string // The discriminator field name
	UnionFallback bool   // Whether to fallback to regular union behavior
}

// ZodDiscriminatedUnionInternals contains discriminated union validator internal state
type ZodDiscriminatedUnionInternals struct {
	ZodUnionInternals
	Def        *ZodDiscriminatedUnionDef   // Schema definition
	PropValues map[string]map[any]struct{} // Property values mapping
	DiscMap    map[any]ZodType[any, any]   // Discriminator value to schema mapping
}

// ZodDiscriminatedUnion represents a discriminated union validation schema
type ZodDiscriminatedUnion struct {
	internals *ZodDiscriminatedUnionInternals
}

// GetInternals returns the internal state of the schema
func (z *ZodDiscriminatedUnion) GetInternals() *ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// Parse with smart type inference and optimized performance using discriminator key
func (z *ZodDiscriminatedUnion) Parse(input any, ctx ...*ParseContext) (any, error) {
	// Prioritize Optional/Nilable handling
	if input == nil {
		if z.internals.Nilable || z.internals.Optional {
			return nil, nil
		}
	}

	parseCtx := (*ParseContext)(nil)
	if len(ctx) > 0 {
		parseCtx = ctx[0]
	}

	// Check if input is an object
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		// Not an object, create invalid_type error
		issue := CreateInvalidTypeIssue(input, "object", "object", func(issue *ZodRawIssue) {
			issue.Inst = z.internals
		})
		finalIssue := FinalizeIssue(issue, parseCtx, GetConfig())
		return nil, NewZodError([]ZodIssue{finalIssue})
	}

	// Get the value of the discriminator key
	discriminatorValue, exists := inputMap[z.internals.Def.Discriminator]
	if !exists {
		// Discriminator key does not exist, decide behavior based on unionFallback
		if z.internals.Def.UnionFallback {
			// Fall back to regular union behavior
			return z.parseAsRegularUnion(input, parseCtx)
		}

		// Create invalid_union error
		issue := CreateInvalidUnionIssue(input, nil, func(issue *ZodRawIssue) {
			issue.Properties = map[string]interface{}{
				"note": "No matching discriminator",
			}
			issue.Path = []interface{}{z.internals.Def.Discriminator}
			issue.Inst = z.internals
		})
		finalIssue := FinalizeIssue(issue, parseCtx, GetConfig())
		return nil, NewZodError([]ZodIssue{finalIssue})
	}

	// Find the corresponding schema based on the discriminator value
	targetSchema, found := z.internals.DiscMap[discriminatorValue]
	if !found {
		// No matching discriminator value
		if z.internals.Def.UnionFallback {
			// Fall back to regular union behavior
			return z.parseAsRegularUnion(input, parseCtx)
		}

		// Create invalid_union error
		issue := CreateInvalidUnionIssue(input, nil, func(issue *ZodRawIssue) {
			issue.Properties = map[string]interface{}{
				"note": "No matching discriminator",
			}
			issue.Path = []interface{}{z.internals.Def.Discriminator}
			issue.Inst = z.internals
		})
		finalIssue := FinalizeIssue(issue, parseCtx, GetConfig())
		return nil, NewZodError([]ZodIssue{finalIssue})
	}

	// Validate using the found schema
	result, err := targetSchema.Parse(input, parseCtx)
	if err == nil {
		// Validate success, perform Refine check
		if len(z.internals.Checks) > 0 {
			payload := &ParsePayload{
				Value:  result,
				Issues: make([]ZodRawIssue, 0),
			}
			runChecksOnValue(result, z.internals.Checks, payload, parseCtx)
			if len(payload.Issues) > 0 {
				return nil, &ZodError{Issues: convertRawIssuesToIssues(payload.Issues, parseCtx)}
			}
		}
		return result, nil
	}

	// Validation failed, return error
	return nil, err
}

// parseAsRegularUnion fallback to regular union parsing behavior
func (z *ZodDiscriminatedUnion) parseAsRegularUnion(input any, ctx *ParseContext) (any, error) {
	// Try each option
	for _, option := range z.internals.Options {
		result, err := option.Parse(input, ctx)
		if err == nil {
			// Found matching option
			if len(z.internals.Checks) > 0 {
				payload := &ParsePayload{
					Value:  result,
					Issues: make([]ZodRawIssue, 0),
				}
				runChecksOnValue(result, z.internals.Checks, payload, ctx)
				if len(payload.Issues) > 0 {
					return nil, &ZodError{Issues: convertRawIssuesToIssues(payload.Issues, ctx)}
				}
			}
			return result, nil
		}
	}

	// All options failed, create invalid_union error
	var allErrors []string
	for _, option := range z.internals.Options {
		_, err := option.Parse(input, ctx)
		if err != nil {
			allErrors = append(allErrors, err.Error())
		}
	}

	issue := CreateInvalidUnionIssue(input, nil)
	if len(allErrors) > 0 {
		issue.Properties = map[string]interface{}{
			"unionErrors": allErrors,
		}
	}

	finalIssue := FinalizeIssue(issue, ctx, GetConfig())
	return nil, NewZodError([]ZodIssue{finalIssue})
}

// MustParse validates the input value and panics on failure
func (z *ZodDiscriminatedUnion) MustParse(input any, ctx ...*ParseContext) any {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// Options provides access to the discriminated union's option schemas
func (z *ZodDiscriminatedUnion) Options() []ZodType[any, any] {
	return z.internals.Options
}

// Discriminator returns the discriminator field name
func (z *ZodDiscriminatedUnion) Discriminator() string {
	return z.internals.Def.Discriminator
}

///////////////////////////
////   DISCRIMINATED UNION WRAPPERS ////
///////////////////////////

// Optional makes the discriminated union optional
func (z *ZodDiscriminatedUnion) Optional() ZodType[any, any] {
	return any(Optional(any(z).(ZodType[any, any]))).(ZodType[any, any])
}

// Nilable modifier: only changes nil handling, does not change type inference logic
func (z *ZodDiscriminatedUnion) Nilable() ZodType[any, any] {
	return Clone(z, func(def *ZodTypeDef) {
		// No need to modify def, because Nilable is a runtime flag
	}).(*ZodDiscriminatedUnion).setNilable()
}

// setNilable internal method to set Nilable flag
func (z *ZodDiscriminatedUnion) setNilable() ZodType[any, any] {
	z.internals.Nilable = true
	return z
}

// Pipe creates a validation pipeline
func (z *ZodDiscriminatedUnion) Pipe(out ZodType[any, any]) ZodType[any, any] {
	return &ZodPipe[any, any]{
		in:  any(z).(ZodType[any, any]),
		out: out,
		def: ZodTypeDef{Type: "pipe"},
	}
}

// Transform provides type-safe discriminated union type transformation
func (z *ZodDiscriminatedUnion) Transform(fn func(any, *RefinementContext) (any, error)) ZodType[any, any] {
	return z.TransformAny(fn)
}

// TransformAny flexible version of transformation
func (z *ZodDiscriminatedUnion) TransformAny(fn func(any, *RefinementContext) (any, error)) ZodType[any, any] {
	transform := NewZodTransform[any, any](fn)
	return &ZodPipe[any, any]{
		in:  any(z).(ZodType[any, any]),
		out: any(transform).(ZodType[any, any]),
		def: ZodTypeDef{Type: "pipe"},
	}
}

////////////////////////////
////   DISCRIMINATED UNION REFINE METHODS   ////
////////////////////////////

// Refine adds type-safe custom validation logic to the discriminated union schema
func (z *ZodDiscriminatedUnion) Refine(fn func(any) bool, params ...SchemaParams) *ZodDiscriminatedUnion {
	result := z.RefineAny(func(v any) bool {
		val, isNil, err := extractUnionValue(v)
		if err != nil {
			return false
		}
		if isNil {
			return true
		}
		return fn(val)
	}, params...)
	return result.(*ZodDiscriminatedUnion)
}

// RefineAny adds flexible custom validation logic to the discriminated union schema
func (z *ZodDiscriminatedUnion) RefineAny(fn func(any) bool, params ...SchemaParams) ZodType[any, any] {
	check := NewCustom[any](fn, params...)
	return AddCheck(any(z).(ZodType[any, any]), check)
}

// Check adds modern validation using direct payload access
func (z *ZodDiscriminatedUnion) Check(fn CheckFn) *ZodDiscriminatedUnion {
	check := NewCustom[any](fn, SchemaParams{})
	result := AddCheck(z, check)
	return result.(*ZodDiscriminatedUnion)
}

////////////////////////////
////   DISCRIMINATED UNION DEFAULT WRAPPER ////
////////////////////////////

// ZodDiscriminatedUnionDefault is a Default wrapper for discriminated union type
type ZodDiscriminatedUnionDefault struct {
	*ZodDefault[*ZodDiscriminatedUnion]
}

// Parse ensures correct validation of the inner type
func (s ZodDiscriminatedUnionDefault) Parse(input any, ctx ...*ParseContext) (any, error) {
	return s.ZodDefault.Parse(input, ctx...)
}

// Default adds a default value to the discriminated union schema, returns ZodDiscriminatedUnionDefault support chain call
func (z *ZodDiscriminatedUnion) Default(value any) ZodDiscriminatedUnionDefault {
	return ZodDiscriminatedUnionDefault{
		&ZodDefault[*ZodDiscriminatedUnion]{
			innerType:    z,
			defaultValue: value,
			isFunction:   false,
		},
	}
}

// DefaultFunc adds a default function to the discriminated union schema, returns ZodDiscriminatedUnionDefault support chain call
func (z *ZodDiscriminatedUnion) DefaultFunc(fn func() any) ZodDiscriminatedUnionDefault {
	return ZodDiscriminatedUnionDefault{
		&ZodDefault[*ZodDiscriminatedUnion]{
			innerType:   z,
			defaultFunc: fn,
			isFunction:  true,
		},
	}
}

// Optional adds an optional check to the discriminated union schema, returns ZodType support chain call
func (s ZodDiscriminatedUnionDefault) Optional() ZodType[any, any] {
	// Wrap current wrapper instance instead of underlying type
	return Optional(any(s).(ZodType[any, any]))
}

// Nilable adds a nilable check to the discriminated union schema, returns ZodType support chain call
func (s ZodDiscriminatedUnionDefault) Nilable() ZodType[any, any] {
	// Wrap current wrapper instance instead of underlying type
	return Nilable(any(s).(ZodType[any, any]))
}

////////////////////////////
////   DISCRIMINATED UNION PREFAULT WRAPPER ////
////////////////////////////

// ZodDiscriminatedUnionPrefault is a Prefault wrapper for discriminated union type
type ZodDiscriminatedUnionPrefault struct {
	*ZodPrefault[*ZodDiscriminatedUnion]
}

// Parse ensures correct validation of the inner type
func (u ZodDiscriminatedUnionPrefault) Parse(input any, ctx ...*ParseContext) (any, error) {
	return u.ZodPrefault.Parse(input, ctx...)
}

// Prefault adds a prefault value to the discriminated union schema, returns ZodDiscriminatedUnionPrefault support chain call
func (z *ZodDiscriminatedUnion) Prefault(value interface{}) ZodDiscriminatedUnionPrefault {
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

	return ZodDiscriminatedUnionPrefault{
		&ZodPrefault[*ZodDiscriminatedUnion]{
			internals:     internals,
			innerType:     z,
			prefaultValue: value,
			prefaultFunc:  nil,
			isFunction:    false,
		},
	}
}

// PrefaultFunc adds a prefault function to the discriminated union schema, returns ZodDiscriminatedUnionPrefault support chain call
func (z *ZodDiscriminatedUnion) PrefaultFunc(fn func() interface{}) ZodDiscriminatedUnionPrefault {
	genericFn := func() any { return fn() }

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

	return ZodDiscriminatedUnionPrefault{
		&ZodPrefault[*ZodDiscriminatedUnion]{
			internals:     internals,
			innerType:     z,
			prefaultValue: nil,
			prefaultFunc:  genericFn,
			isFunction:    true,
		},
	}
}

// Optional adds an optional check to the discriminated union schema, returns ZodType support chain call
func (u ZodDiscriminatedUnionPrefault) Optional() ZodType[any, any] {
	// Wrap current wrapper instance instead of underlying type
	return Optional(any(u).(ZodType[any, any]))
}

// Nilable adds a nilable check to the discriminated union schema, returns ZodType support chain call
func (u ZodDiscriminatedUnionPrefault) Nilable() ZodType[any, any] {
	// Wrap current wrapper instance instead of underlying type
	return Nilable(any(u).(ZodType[any, any]))
}

////////////////////////////
////   DISCRIMINATED UNION INTERNAL METHODS ////
////////////////////////////

// CloneFrom implements Cloneable interface for type-specific state copying
func (z *ZodDiscriminatedUnion) CloneFrom(source any) {
	if src, ok := source.(*ZodDiscriminatedUnion); ok {
		// Copy discriminated union-specific fields
		if src.internals.PropValues != nil {
			z.internals.PropValues = make(map[string]map[any]struct{})
			for k, v := range src.internals.PropValues {
				z.internals.PropValues[k] = make(map[any]struct{})
				for val := range v {
					z.internals.PropValues[k][val] = struct{}{}
				}
			}
		}

		if src.internals.DiscMap != nil {
			z.internals.DiscMap = make(map[any]ZodType[any, any])
			for k, v := range src.internals.DiscMap {
				z.internals.DiscMap[k] = v
			}
		}

		// Copy union-specific fields
		if src.internals.Options != nil {
			z.internals.Options = make([]ZodType[any, any], len(src.internals.Options))
			copy(z.internals.Options, src.internals.Options)
		}
	}
}

// Unwrap returns the inner type (for basic types, returns self)
func (z *ZodDiscriminatedUnion) Unwrap() ZodType[any, any] {
	return any(z).(ZodType[any, any])
}

////////////////////////////
////   DISCRIMINATED UNION CONSTRUCTOR ////
////////////////////////////

// createZodDiscriminatedUnionFromDef creates a ZodDiscriminatedUnion from definition
func createZodDiscriminatedUnionFromDef(def *ZodDiscriminatedUnionDef) *ZodDiscriminatedUnion {
	// First build PropValues mapping
	propValues := make(map[string]map[any]struct{})
	discMap := make(map[any]ZodType[any, any])

	// Validate each option and build mapping
	for i, option := range def.Options {
		// Get option's property values (here we need to implement the logic to get the possible values of the schema)
		optionPropValues := extractPropValues(option)
		if len(optionPropValues) == 0 {
			panic(fmt.Sprintf("Invalid discriminated union option at index %d", i))
		}

		// Merge into total propValues
		for prop, values := range optionPropValues {
			if propValues[prop] == nil {
				propValues[prop] = make(map[any]struct{})
			}
			for val := range values {
				propValues[prop][val] = struct{}{}
			}
		}

		// Build discriminator key mapping
		discriminatorValues, exists := optionPropValues[def.Discriminator]
		if !exists || len(discriminatorValues) == 0 {
			panic(fmt.Sprintf("Invalid discriminated union option at index %d: missing discriminator '%s'", i, def.Discriminator))
		}

		for val := range discriminatorValues {
			if _, exists := discMap[val]; exists {
				panic(fmt.Sprintf("Duplicate discriminator value %v", val))
			}
			discMap[val] = option
		}
	}

	internals := &ZodDiscriminatedUnionInternals{
		ZodUnionInternals: ZodUnionInternals{
			ZodTypeInternals: newBaseZodTypeInternals("union"),
			Def:              &def.ZodUnionDef,
			Options:          def.Options,
			Isst:             ZodIssueInvalidUnion{},
			Bag:              make(map[string]interface{}),
		},
		Def:        def,
		PropValues: propValues,
		DiscMap:    discMap,
	}

	// Set simplified constructor for cloning
	internals.Constructor = func(newDef *ZodTypeDef) ZodType[any, any] {
		discUnionDef := &ZodDiscriminatedUnionDef{
			ZodUnionDef: ZodUnionDef{
				ZodTypeDef: *newDef,
				Type:       ZodTypeUnion,
				Options:    def.Options,
			},
			Discriminator: def.Discriminator,
			UnionFallback: def.UnionFallback,
		}
		return any(createZodDiscriminatedUnionFromDef(discUnionDef)).(ZodType[any, any])
	}

	schema := &ZodDiscriminatedUnion{internals: internals}

	// Initialize schema
	initZodType(any(schema).(ZodType[any, any]), &def.ZodTypeDef)

	// Apply schema parameters
	if def.UnionFallback {
		schema.internals.Bag["coerce"] = true
	}

	// Handle schema-level error mapping
	if def.Error != nil {
		errorMap := createErrorMap(def.Error)
		if errorMap != nil {
			def.Error = errorMap
			schema.internals.Error = errorMap
		}
	}

	return schema
}

// extractPropValues extract the property value mapping of the schema
// This function needs to handle Object-type schemas, extracting the possible values of their fields
func extractPropValues(schema ZodType[any, any]) map[string]map[any]struct{} {
	result := make(map[string]map[any]struct{})

	// Try type assertion to ZodObject
	if objSchema, ok := schema.(*ZodObject); ok {
		// Get Object's field definition
		shape := objSchema.Shape()
		// Traverse each field
		for fieldName, fieldSchema := range shape {
			// Recursively extract the possible values of the field
			fieldValues := extractPropValues(fieldSchema)

			// If the field is Literal type, directly get its value
			if fieldInternals := fieldSchema.GetInternals(); fieldInternals != nil && len(fieldInternals.Values) > 0 {
				if result[fieldName] == nil {
					result[fieldName] = make(map[any]struct{})
				}
				for value := range fieldInternals.Values {
					result[fieldName][value] = struct{}{}
				}
			} else {
				// Merge the property values of the field
				for prop, values := range fieldValues {
					if result[prop] == nil {
						result[prop] = make(map[any]struct{})
					}
					for val := range values {
						result[prop][val] = struct{}{}
					}
				}
			}
		}
		return result
	}

	// Try type assertion to ZodDiscriminatedUnion (nested discriminated union)
	if discUnionSchema, ok := schema.(*ZodDiscriminatedUnion); ok {
		// For nested discriminated unions, we need to extract the property values of all options
		for _, option := range discUnionSchema.Options() {
			optionValues := extractPropValues(option)
			for prop, values := range optionValues {
				if result[prop] == nil {
					result[prop] = make(map[any]struct{})
				}
				for val := range values {
					result[prop][val] = struct{}{}
				}
			}
		}
		return result
	}

	// Try type assertion to ZodLiteral
	if litSchema, ok := schema.(*ZodLiteral); ok {
		internals := litSchema.GetInternals()
		if internals != nil && len(internals.Values) > 0 {
			// For Literal, we don't know which property it belongs to, so return empty
			// This value should be handled at the Object level
			return result
		}
	}

	// Check general Values information
	internals := schema.GetInternals()
	if internals != nil && len(internals.Values) > 0 {
		// If there are Values but not in the Object context, we cannot determine the property name
		// Return empty result
		return result
	}

	// If no property values are found, return empty mapping
	// This will cause the discriminated union construction to fail, which is correct behavior
	return result
}

// NewZodDiscriminatedUnion creates a new discriminated union schema
func NewZodDiscriminatedUnion(discriminator string, options []ZodType[any, any], params ...SchemaParams) *ZodDiscriminatedUnion {
	// Validate input parameters
	if len(options) == 0 {
		panic("Discriminated union must have at least one option")
	}
	if discriminator == "" {
		panic("Discriminator field name cannot be empty")
	}

	var finalParams SchemaParams
	if len(params) > 0 {
		finalParams = params[0]
	}

	def := &ZodDiscriminatedUnionDef{
		ZodUnionDef: ZodUnionDef{
			ZodTypeDef: ZodTypeDef{
				Type:   ZodTypeUnion,
				Checks: make([]ZodCheck, 0),
			},
			Type:    ZodTypeUnion,
			Options: options,
		},
		Discriminator: discriminator,
		UnionFallback: finalParams.UnionFallback,
	}

	schema := createZodDiscriminatedUnionFromDef(def)

	// Apply schema parameters
	if finalParams.Coerce {
		schema.internals.Bag["coerce"] = true
	}

	// Handle schema-level error mapping
	if finalParams.Error != nil {
		errorMap := createErrorMap(finalParams.Error)
		if errorMap != nil {
			def.Error = errorMap
			schema.internals.Error = errorMap
		}
	}

	return schema
}

// DiscriminatedUnion creates a new discriminated union validation schema
func DiscriminatedUnion(discriminator string, options []ZodType[any, any], params ...SchemaParams) *ZodDiscriminatedUnion {
	return NewZodDiscriminatedUnion(discriminator, options, params...)
}

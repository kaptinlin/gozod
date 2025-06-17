package gozod

import (
	"errors"
	"fmt"
	"reflect"
)

//////////////////////////////////////////
//////////////////////////////////////////
//////////                      //////////
//////////   ZodIntersection    //////////
//////////                      //////////
//////////////////////////////////////////
//////////////////////////////////////////

// ZodIntersectionDef defines the configuration for intersection validation
type ZodIntersectionDef struct {
	ZodTypeDef
	Type  string            // "intersection"
	Left  ZodType[any, any] // Left schema
	Right ZodType[any, any] // Right schema
}

// ZodIntersectionInternals contains intersection validator internal state
type ZodIntersectionInternals struct {
	ZodTypeInternals
	Def   *ZodIntersectionDef    // Schema definition
	Left  ZodType[any, any]      // Left schema
	Right ZodType[any, any]      // Right schema
	Bag   map[string]interface{} // Additional metadata
}

// ZodIntersection represents an intersection validation schema
type ZodIntersection struct {
	internals *ZodIntersectionInternals
}

//////////////////////////////////////////
//////////     Core Interface    //////////
//////////////////////////////////////////

// GetInternals returns the internal state of the schema
func (z *ZodIntersection) GetInternals() *ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// GetZod returns the intersection-specific internals for framework usage
func (z *ZodIntersection) GetZod() *ZodIntersectionInternals {
	return z.internals
}

// CloneFrom implements Cloneable interface for type-specific state copying
func (z *ZodIntersection) CloneFrom(source any) {
	if src, ok := source.(*ZodIntersection); ok {
		// Copy intersection-specific fields
		z.internals.Left = src.internals.Left
		z.internals.Right = src.internals.Right

		// Copy Bag state
		if len(src.internals.Bag) > 0 {
			if z.internals.Bag == nil {
				z.internals.Bag = make(map[string]interface{})
			}
			for key, value := range src.internals.Bag {
				z.internals.Bag[key] = value
			}
		}
	}
}

// Parse validates the input value using intersection logic
func (z *ZodIntersection) Parse(input any, ctx ...*ParseContext) (any, error) {
	parseCtx := (*ParseContext)(nil)
	if len(ctx) > 0 {
		parseCtx = ctx[0]
	}

	// Parse with left schema
	leftResult, leftErr := z.internals.Left.Parse(input, parseCtx)

	// Parse with right schema
	rightResult, rightErr := z.internals.Right.Parse(input, parseCtx)

	// Collect all errors
	var allErrors []ZodIssue
	if leftErr != nil {
		var zodErr *ZodError
		if errors.As(leftErr, &zodErr) {
			allErrors = append(allErrors, zodErr.Issues...)
		}
	}
	if rightErr != nil {
		var zodErr *ZodError
		if errors.As(rightErr, &zodErr) {
			allErrors = append(allErrors, zodErr.Issues...)
		}
	}

	// If either side failed, return all errors
	if len(allErrors) > 0 {
		return nil, NewZodError(allErrors)
	}

	// Both sides succeeded, merge the results
	merged, mergeErr := mergeValues(leftResult, rightResult)
	if mergeErr != nil {
		// Create unmergeable intersection error
		issue := CreateInvalidIntersectionIssue(input, mergeErr.Error())
		finalIssue := FinalizeIssue(issue, parseCtx, GetConfig())
		return nil, NewZodError([]ZodIssue{finalIssue})
	}

	// Run additional checks if any
	if len(z.internals.Checks) > 0 {
		payload := &ParsePayload{
			Value:  merged,
			Issues: make([]ZodRawIssue, 0),
		}
		runChecksOnValue(merged, z.internals.Checks, payload, parseCtx)
		if len(payload.Issues) > 0 {
			return nil, &ZodError{Issues: convertRawIssuesToIssues(payload.Issues, parseCtx)}
		}
	}

	return merged, nil
}

// MustParse validates the input value and panics on failure
func (z *ZodIntersection) MustParse(input any, ctx ...*ParseContext) any {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

//////////////////////////////////////////
//////////   Wrapper Methods     //////////
//////////////////////////////////////////

// Transform creates a type-safe transformation of intersection values
func (z *ZodIntersection) Transform(fn func(any, *RefinementContext) (any, error)) ZodType[any, any] {
	return z.TransformAny(fn)
}

// TransformAny implements data transformation
func (z *ZodIntersection) TransformAny(fn func(any, *RefinementContext) (any, error)) ZodType[any, any] {
	transform := NewZodTransform[any, any](fn)
	return &ZodPipe[any, any]{
		in:  any(z).(ZodType[any, any]),
		out: any(transform).(ZodType[any, any]),
		def: ZodTypeDef{Type: "pipe"},
	}
}

// Pipe pipes the intersection to another schema
func (z *ZodIntersection) Pipe(schema ZodType[any, any]) ZodType[any, any] {
	return &ZodPipe[any, any]{
		in:  any(z).(ZodType[any, any]),
		out: schema,
		def: ZodTypeDef{Type: "pipe"},
	}
}

//////////////////////////////////////////
//////////   Utility Methods     //////////
//////////////////////////////////////////

// Optional makes the intersection optional
func (z *ZodIntersection) Optional() ZodType[any, any] {
	return any(Optional(any(z).(ZodType[any, any]))).(ZodType[any, any])
}

// Nilable makes the intersection nilable
func (z *ZodIntersection) Nilable() ZodType[any, any] {
	return Nilable(z)
}

// Nullish combines optional and nilable
func (z *ZodIntersection) Nullish() ZodType[any, any] {
	return any(Nullish(any(z).(ZodType[any, any]))).(ZodType[any, any])
}

// Refine adds type-safe custom validation logic to the intersection schema
func (z *ZodIntersection) Refine(fn func(any) bool, params ...SchemaParams) *ZodIntersection {
	result := z.RefineAny(fn, params...)
	return result.(*ZodIntersection)
}

// RefineAny adds flexible custom validation logic to the intersection schema
func (z *ZodIntersection) RefineAny(fn func(any) bool, params ...SchemaParams) ZodType[any, any] {
	check := NewCustom[any](fn, params...)
	return AddCheck(z, check)
}

// Left returns the left schema
func (z *ZodIntersection) Left() ZodType[any, any] {
	return z.internals.Left
}

// Right returns the right schema
func (z *ZodIntersection) Right() ZodType[any, any] {
	return z.internals.Right
}

////////////////////////////
////   INTERSECTION DEFAULT WRAPPER ////
////////////////////////////

// ZodIntersectionDefault is the Default wrapper for intersection types
type ZodIntersectionDefault struct {
	*ZodDefault[*ZodIntersection]
}

// Default creates a default wrapper for intersection
func (z *ZodIntersection) Default(value any) ZodIntersectionDefault {
	return ZodIntersectionDefault{
		&ZodDefault[*ZodIntersection]{
			innerType:    z,
			defaultValue: value,
			isFunction:   false,
		},
	}
}

// DefaultFunc creates a function-based default wrapper for intersection
func (z *ZodIntersection) DefaultFunc(fn func() any) ZodIntersectionDefault {
	return ZodIntersectionDefault{
		&ZodDefault[*ZodIntersection]{
			innerType:   z,
			defaultFunc: fn,
			isFunction:  true,
		},
	}
}

// Refine adds refinement to default wrapper
func (s ZodIntersectionDefault) Refine(fn func(any) bool, params ...SchemaParams) ZodIntersectionDefault {
	newInner := s.innerType.Refine(fn, params...)
	return ZodIntersectionDefault{
		&ZodDefault[*ZodIntersection]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

// Transform adds transformation to default wrapper
func (s ZodIntersectionDefault) Transform(fn func(any, *RefinementContext) (any, error)) ZodType[any, any] {
	return s.TransformAny(fn)
}

// Optional makes the intersection schema optional
func (s ZodIntersectionDefault) Optional() ZodType[any, any] {
	// wrap the current wrapper instance, not the underlying type
	return Optional(any(s).(ZodType[any, any]))
}

// Nilable makes the intersection schema nilable
func (s ZodIntersectionDefault) Nilable() ZodType[any, any] {
	// wrap the current wrapper instance, not the underlying type
	return Nilable(any(s).(ZodType[any, any]))
}

////////////////////////////
////   INTERSECTION PREFAULT WRAPPER ////
////////////////////////////

// ZodIntersectionPrefault is the Prefault wrapper for intersection types
type ZodIntersectionPrefault struct {
	*ZodPrefault[*ZodIntersection]
}

// Prefault creates a prefault wrapper for intersection
func (z *ZodIntersection) Prefault(value any) ZodIntersectionPrefault {
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

	return ZodIntersectionPrefault{
		&ZodPrefault[*ZodIntersection]{
			internals:     internals,
			innerType:     z,
			prefaultValue: value,
			prefaultFunc:  nil,
			isFunction:    false,
		},
	}
}

// PrefaultFunc creates a function-based prefault wrapper for intersection
func (z *ZodIntersection) PrefaultFunc(fn func() any) ZodIntersectionPrefault {
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

	return ZodIntersectionPrefault{
		&ZodPrefault[*ZodIntersection]{
			internals:     internals,
			innerType:     z,
			prefaultValue: nil,
			prefaultFunc:  fn,
			isFunction:    true,
		},
	}
}

// Refine adds refinement to prefault wrapper
func (i ZodIntersectionPrefault) Refine(fn func(any) bool, params ...SchemaParams) ZodIntersectionPrefault {
	newInner := i.innerType.Refine(fn, params...)
	baseInternals := newInner.GetInternals()
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

	return ZodIntersectionPrefault{
		&ZodPrefault[*ZodIntersection]{
			internals:     internals,
			innerType:     newInner,
			prefaultValue: i.prefaultValue,
			prefaultFunc:  i.prefaultFunc,
			isFunction:    i.isFunction,
		},
	}
}

// Transform adds transformation to prefault wrapper
func (i ZodIntersectionPrefault) Transform(fn func(any, *RefinementContext) (any, error)) ZodType[any, any] {
	return i.TransformAny(fn)
}

// Optional makes the intersection schema optional
func (i ZodIntersectionPrefault) Optional() ZodType[any, any] {
	// wrap the current wrapper instance, not the underlying type
	return Optional(any(i).(ZodType[any, any]))
}

// Nilable makes the intersection schema nilable
func (i ZodIntersectionPrefault) Nilable() ZodType[any, any] {
	// wrap the current wrapper instance, not the underlying type
	return Nilable(any(i).(ZodType[any, any]))
}

//////////////////////////////////////////
//////////   Internal Methods    //////////
//////////////////////////////////////////

// createZodIntersectionFromDef creates a ZodIntersection from definition
func createZodIntersectionFromDef(def *ZodIntersectionDef) *ZodIntersection {
	internals := &ZodIntersectionInternals{
		ZodTypeInternals: newBaseZodTypeInternals("intersection"),
		Def:              def,
		Left:             def.Left,
		Right:            def.Right,
		Bag:              make(map[string]interface{}),
	}

	// Set up constructor for cloning
	internals.Constructor = func(newDef *ZodTypeDef) ZodType[any, any] {
		intersectionDef := &ZodIntersectionDef{
			ZodTypeDef: *newDef,
			Type:       "intersection",
			Left:       def.Left,
			Right:      def.Right,
		}
		return createZodIntersectionFromDef(intersectionDef)
	}

	schema := &ZodIntersection{internals: internals}

	// Initialize the schema
	initZodType(any(schema).(ZodType[any, any]), &def.ZodTypeDef)

	return schema
}

// NewZodIntersection creates a new intersection schema
func NewZodIntersection(left, right ZodType[any, any], params ...SchemaParams) *ZodIntersection {
	def := &ZodIntersectionDef{
		ZodTypeDef: ZodTypeDef{
			Type:   "intersection",
			Checks: make([]ZodCheck, 0),
		},
		Type:  "intersection",
		Left:  left,
		Right: right,
	}

	schema := createZodIntersectionFromDef(def)

	// Apply schema parameters
	if len(params) > 0 {
		param := params[0]

		// Handle schema-level error configuration
		if param.Error != nil {
			errorMap := createErrorMap(param.Error)
			if errorMap != nil {
				def.Error = errorMap
				schema.internals.Error = errorMap
			}
		}

		// Handle additional parameters
		if param.Description != "" {
			schema.internals.Bag["description"] = param.Description
		}
		if param.Abort {
			schema.internals.Bag["abort"] = true
		}
	}

	return schema
}

// Intersection creates a new intersection validation schema
func Intersection(left, right ZodType[any, any], params ...SchemaParams) *ZodIntersection {
	return NewZodIntersection(left, right, params...)
}

// Unwrap returns the inner type (for basic types, returns self)
func (z *ZodIntersection) Unwrap() ZodType[any, any] {
	return any(z).(ZodType[any, any])
}

//////////////////////////////////////////
//////////   Merge Logic        //////////
//////////////////////////////////////////

// mergeValues merges two values according to intersection logic
func mergeValues(a, b any) (any, error) {
	// Handle nil values first
	if a == nil {
		return b, nil
	}
	if b == nil {
		return a, nil
	}

	// Use reflection to handle different types
	aVal := reflect.ValueOf(a)
	bVal := reflect.ValueOf(b)

	// If types are different, they can't be merged
	if aVal.Type() != bVal.Type() {
		return nil, fmt.Errorf("%w: %T and %T", ErrCannotMergeDifferentTypes, a, b)
	}

	// Handle maps (objects)
	if aVal.Kind() == reflect.Map && bVal.Kind() == reflect.Map {
		return mergeMaps(a, b)
	}

	// Handle slices (arrays)
	if aVal.Kind() == reflect.Slice && bVal.Kind() == reflect.Slice {
		return mergeSlices(a, b)
	}

	// For other types, use deep equality check
	if reflect.DeepEqual(a, b) {
		return a, nil
	}

	return nil, ErrCannotMergeIncompatibleValues
}

// mergeMaps merges two map values
func mergeMaps(a, b any) (any, error) {
	aMap, aOk := a.(map[string]interface{})
	bMap, bOk := b.(map[string]interface{})

	if !aOk || !bOk {
		// Try to convert using reflection
		aVal := reflect.ValueOf(a)
		bVal := reflect.ValueOf(b)

		if aVal.Kind() != reflect.Map || bVal.Kind() != reflect.Map {
			return nil, ErrExpectedMapsForMerging
		}

		// Convert to map[string]interface{}
		aMap = make(map[string]interface{})
		bMap = make(map[string]interface{})

		for _, key := range aVal.MapKeys() {
			if keyStr, ok := key.Interface().(string); ok {
				aMap[keyStr] = aVal.MapIndex(key).Interface()
			}
		}

		for _, key := range bVal.MapKeys() {
			if keyStr, ok := key.Interface().(string); ok {
				bMap[keyStr] = bVal.MapIndex(key).Interface()
			}
		}
	}

	// Merge the maps
	result := make(map[string]interface{})

	// Copy all keys from a
	for k, v := range aMap {
		result[k] = v
	}

	// Merge keys from b
	for k, v := range bMap {
		if existing, exists := result[k]; exists {
			// Recursively merge if both values exist
			merged, err := mergeValues(existing, v)
			if err != nil {
				return nil, fmt.Errorf("%w '%s': %w", ErrMergeConflictAtKey, k, err)
			}
			result[k] = merged
		} else {
			result[k] = v
		}
	}

	return result, nil
}

// mergeSlices merges two slice values
func mergeSlices(a, b any) (any, error) {
	aVal := reflect.ValueOf(a)
	bVal := reflect.ValueOf(b)

	if aVal.Len() != bVal.Len() {
		return nil, fmt.Errorf("%w: %d vs %d", ErrCannotMergeSlicesDifferentLength, aVal.Len(), bVal.Len())
	}

	// Create result slice
	resultType := aVal.Type()
	result := reflect.MakeSlice(resultType, aVal.Len(), aVal.Len())

	// Merge each element
	for i := 0; i < aVal.Len(); i++ {
		aElem := aVal.Index(i).Interface()
		bElem := bVal.Index(i).Interface()

		merged, err := mergeValues(aElem, bElem)
		if err != nil {
			return nil, fmt.Errorf("%w %d: %w", ErrMergeConflictAtIndex, i, err)
		}

		result.Index(i).Set(reflect.ValueOf(merged))
	}

	return result.Interface(), nil
}

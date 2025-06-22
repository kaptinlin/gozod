package types

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/internal/issues"
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
	core.ZodTypeDef
	Type  string                 // "intersection"
	Left  core.ZodType[any, any] // Left schema
	Right core.ZodType[any, any] // Right schema
}

// ZodIntersectionInternals contains intersection validator internal state
type ZodIntersectionInternals struct {
	core.ZodTypeInternals
	Def   *ZodIntersectionDef    // Schema definition
	Left  core.ZodType[any, any] // Left schema
	Right core.ZodType[any, any] // Right schema
	Bag   map[string]any         // Additional metadata
}

// ZodIntersection represents an intersection validation schema
type ZodIntersection struct {
	internals *ZodIntersectionInternals
}

//////////////////////////////////////////
//////////     Core Interface    //////////
//////////////////////////////////////////

// GetInternals returns the internal state of the schema
func (z *ZodIntersection) GetInternals() *core.ZodTypeInternals {
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
				z.internals.Bag = make(map[string]any)
			}
			for key, value := range src.internals.Bag {
				z.internals.Bag[key] = value
			}
		}
	}
}

// Parse validates the input value using intersection logic
func (z *ZodIntersection) Parse(input any, ctx ...*core.ParseContext) (any, error) {
	parseCtx := (*core.ParseContext)(nil)
	if len(ctx) > 0 {
		parseCtx = ctx[0]
	}

	// Parse with left schema
	leftResult, leftErr := z.internals.Left.Parse(input, parseCtx)

	// Parse with right schema
	rightResult, rightErr := z.internals.Right.Parse(input, parseCtx)

	// Collect all errors
	var allErrors []core.ZodIssue
	if leftErr != nil {
		var zodErr *issues.ZodError
		if errors.As(leftErr, &zodErr) {
			for _, issue := range zodErr.Issues {
				convertedIssue := core.ZodIssue{
					ZodIssueBase: core.ZodIssueBase{
						Code:    issue.Code,
						Input:   issue.Input,
						Path:    issue.Path,
						Message: issue.Message,
					},
				}
				allErrors = append(allErrors, convertedIssue)
			}
		}
	}
	if rightErr != nil {
		var zodErr *issues.ZodError
		if errors.As(rightErr, &zodErr) {
			for _, issue := range zodErr.Issues {
				convertedIssue := core.ZodIssue{
					ZodIssueBase: core.ZodIssueBase{
						Code:    issue.Code,
						Input:   issue.Input,
						Path:    issue.Path,
						Message: issue.Message,
					},
				}
				allErrors = append(allErrors, convertedIssue)
			}
		}
	}

	// If either side failed, return all errors
	if len(allErrors) > 0 {
		return nil, issues.NewZodError(allErrors)
	}

	// Both sides succeeded, merge the results
	merged, mergeErr := mergeValues(leftResult, rightResult)
	if mergeErr != nil {
		// Create unmergeable intersection error
		issue := issues.CreateCustomIssue(mergeErr.Error(), map[string]any{"type": "intersection"}, input)
		finalIssue := issues.FinalizeIssue(issue, parseCtx, core.GetConfig())
		return nil, issues.NewZodError([]core.ZodIssue{finalIssue})
	}

	// Run additional checks if any
	if len(z.internals.Checks) > 0 {
		payload := &core.ParsePayload{
			Value:  merged,
			Issues: make([]core.ZodRawIssue, 0),
		}
		engine.RunChecksOnValue(merged, z.internals.Checks, payload, parseCtx)
		if len(payload.Issues) > 0 {
			finalizedIssues := make([]core.ZodIssue, len(payload.Issues))
			for i, rawIssue := range payload.Issues {
				finalizedIssues[i] = issues.FinalizeIssue(rawIssue, parseCtx, core.GetConfig())
			}
			return nil, issues.NewZodError(finalizedIssues)
		}
	}

	return merged, nil
}

// MustParse validates the input value and panics on failure
func (z *ZodIntersection) MustParse(input any, ctx ...*core.ParseContext) any {
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
func (z *ZodIntersection) Transform(fn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	return z.TransformAny(fn)
}

// TransformAny implements data transformation
func (z *ZodIntersection) TransformAny(fn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	transform := Transform[any, any](fn)
	return &ZodPipe[any, any]{
		in:  any(z).(core.ZodType[any, any]),
		out: any(transform).(core.ZodType[any, any]),
		def: core.ZodTypeDef{Type: "pipe"},
	}
}

// Pipe pipes the intersection to another schema
func (z *ZodIntersection) Pipe(schema core.ZodType[any, any]) core.ZodType[any, any] {
	return &ZodPipe[any, any]{
		in:  any(z).(core.ZodType[any, any]),
		out: schema,
		def: core.ZodTypeDef{Type: "pipe"},
	}
}

//////////////////////////////////////////
//////////   Utility Methods     //////////
//////////////////////////////////////////

// Optional makes the intersection optional
func (z *ZodIntersection) Optional() core.ZodType[any, any] {
	return any(Optional(any(z).(core.ZodType[any, any]))).(core.ZodType[any, any])
}

// Nilable makes the intersection nilable
func (z *ZodIntersection) Nilable() core.ZodType[any, any] {
	return Nilable(z)
}

// Nullish combines optional and nilable
func (z *ZodIntersection) Nullish() core.ZodType[any, any] {
	return any(Nullish(any(z).(core.ZodType[any, any]))).(core.ZodType[any, any])
}

// Refine adds type-safe custom validation logic to the intersection schema
func (z *ZodIntersection) Refine(fn func(any) bool, params ...any) *ZodIntersection {
	result := z.RefineAny(fn, params...)
	return result.(*ZodIntersection)
}

// RefineAny adds flexible custom validation logic to the intersection schema
func (z *ZodIntersection) RefineAny(fn func(any) bool, params ...any) core.ZodType[any, any] {
	check := checks.NewCustom[any](fn, params...)

	// Create new internals with added check
	newInternals := &ZodIntersectionInternals{
		ZodTypeInternals: *z.GetInternals(),
		Def:              z.internals.Def,
		Left:             z.internals.Left,
		Right:            z.internals.Right,
		Bag:              z.internals.Bag,
	}

	// Add the check to the new internals
	newInternals.Checks = append(make([]core.ZodCheck, len(z.internals.Checks)), z.internals.Checks...)
	newInternals.Checks = append(newInternals.Checks, check)

	// Create new intersection with updated internals
	newIntersection := &ZodIntersection{internals: newInternals}

	// Execute onattach callbacks
	if check != nil {
		if checkInternals := check.GetZod(); checkInternals != nil {
			for _, fn := range checkInternals.OnAttach {
				fn(newIntersection)
			}
		}
	}

	return newIntersection
}

// Left returns the left schema
func (z *ZodIntersection) Left() core.ZodType[any, any] {
	return z.internals.Left
}

// Right returns the right schema
func (z *ZodIntersection) Right() core.ZodType[any, any] {
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
func (s ZodIntersectionDefault) Refine(fn func(any) bool, params ...any) ZodIntersectionDefault {
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
func (s ZodIntersectionDefault) Transform(fn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	return s.TransformAny(fn)
}

// Optional makes the intersection schema optional
func (s ZodIntersectionDefault) Optional() core.ZodType[any, any] {
	// wrap the current wrapper instance, not the underlying type
	return Optional(any(s).(core.ZodType[any, any]))
}

// Nilable makes the intersection schema nilable
func (s ZodIntersectionDefault) Nilable() core.ZodType[any, any] {
	// wrap the current wrapper instance, not the underlying type
	return Nilable(any(s).(core.ZodType[any, any]))
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
	internals := &core.ZodTypeInternals{
		Version:     core.Version,
		Type:        core.ZodTypePrefault,
		Checks:      baseInternals.Checks,
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
	internals := &core.ZodTypeInternals{
		Version:     core.Version,
		Type:        core.ZodTypePrefault,
		Checks:      baseInternals.Checks,
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
func (i ZodIntersectionPrefault) Refine(fn func(any) bool, params ...any) ZodIntersectionPrefault {
	newInner := i.innerType.Refine(fn, params...)
	baseInternals := newInner.GetInternals()
	internals := &core.ZodTypeInternals{
		Version:     core.Version,
		Type:        core.ZodTypePrefault,
		Checks:      baseInternals.Checks,
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
func (i ZodIntersectionPrefault) Transform(fn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	return i.TransformAny(fn)
}

// Optional makes the intersection schema optional
func (i ZodIntersectionPrefault) Optional() core.ZodType[any, any] {
	// wrap the current wrapper instance, not the underlying type
	return Optional(any(i).(core.ZodType[any, any]))
}

// Nilable makes the intersection schema nilable
func (i ZodIntersectionPrefault) Nilable() core.ZodType[any, any] {
	// wrap the current wrapper instance, not the underlying type
	return Nilable(any(i).(core.ZodType[any, any]))
}

//////////////////////////////////////////
//////////   Internal Methods    //////////
//////////////////////////////////////////

// createZodIntersectionFromDef creates a ZodIntersection from definition
func createZodIntersectionFromDef(def *ZodIntersectionDef) *ZodIntersection {
	internals := &ZodIntersectionInternals{
		ZodTypeInternals: engine.NewBaseZodTypeInternals("intersection"),
		Def:              def,
		Left:             def.Left,
		Right:            def.Right,
		Bag:              make(map[string]any),
	}

	// Set up constructor for cloning
	internals.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any, any] {
		intersectionDef := &ZodIntersectionDef{
			ZodTypeDef: *newDef,
			Type:       "intersection",
			Left:       def.Left,
			Right:      def.Right,
		}
		return createZodIntersectionFromDef(intersectionDef)
	}

	zodSchema := &ZodIntersection{internals: internals}

	// Initialize the schema
	engine.InitZodType(any(zodSchema).(core.ZodType[any, any]), &def.ZodTypeDef)

	return zodSchema
}

// Intersection creates a new intersection schema
func Intersection(left, right core.ZodType[any, any], params ...any) *ZodIntersection {
	def := &ZodIntersectionDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   "intersection",
			Checks: make([]core.ZodCheck, 0),
		},
		Type:  "intersection",
		Left:  left,
		Right: right,
	}

	schema := createZodIntersectionFromDef(def)

	// Apply schema parameters
	if len(params) > 0 {
		if param, ok := params[0].(core.SchemaParams); ok {
			// Handle schema-level error configuration
			if param.Error != nil {
				errorMap := issues.CreateErrorMap(param.Error)
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
	}

	return schema
}

// Unwrap returns the inner type (for basic types, returns self)
func (z *ZodIntersection) Unwrap() core.ZodType[any, any] {
	return any(z).(core.ZodType[any, any])
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
		return nil, fmt.Errorf("cannot merge different types: %T and %T", a, b)
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

	return nil, fmt.Errorf("cannot merge incompatible values")
}

// mergeMaps merges two map values
func mergeMaps(a, b any) (any, error) {
	aMap, aOk := a.(map[string]any)
	bMap, bOk := b.(map[string]any)

	if !aOk || !bOk {
		// Try to convert using reflection
		aVal := reflect.ValueOf(a)
		bVal := reflect.ValueOf(b)

		if aVal.Kind() != reflect.Map || bVal.Kind() != reflect.Map {
			return nil, fmt.Errorf("expected maps for merging")
		}

		// Convert to map[string]any
		aMap = make(map[string]any)
		bMap = make(map[string]any)

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
	result := make(map[string]any)

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
				return nil, fmt.Errorf("merge conflict at key '%s': %w", k, err)
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
		return nil, fmt.Errorf("cannot merge slices of different length: %d vs %d", aVal.Len(), bVal.Len())
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
			return nil, fmt.Errorf("merge conflict at index %d: %w", i, err)
		}

		result.Index(i).Set(reflect.ValueOf(merged))
	}

	return result.Interface(), nil
}

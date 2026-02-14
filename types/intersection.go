package types

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/internal/utils"
)

// ZodIntersectionDef is the configuration for intersection validation.
type ZodIntersectionDef struct {
	core.ZodTypeDef
	Left  core.ZodSchema
	Right core.ZodSchema
}

// ZodIntersectionInternals holds the internal state of an intersection validator.
type ZodIntersectionInternals struct {
	core.ZodTypeInternals
	Def   *ZodIntersectionDef
	Left  core.ZodSchema
	Right core.ZodSchema
}

// ZodIntersection is an intersection validation schema with dual generic parameters.
// T is the base type (any) and R is the constraint type (any or *any).
type ZodIntersection[T any, R any] struct {
	internals *ZodIntersectionInternals
}

// Internals returns the internal state of the schema.
func (i *ZodIntersection[T, R]) Internals() *core.ZodTypeInternals {
	return &i.internals.ZodTypeInternals
}

// IsOptional reports whether this schema accepts undefined/missing values.
func (i *ZodIntersection[T, R]) IsOptional() bool {
	return i.internals.IsOptional()
}

// IsNilable reports whether this schema accepts nil values.
func (i *ZodIntersection[T, R]) IsNilable() bool {
	return i.internals.IsNilable()
}

// Parse validates input using engine.ParseComplex for unified Default/Prefault handling.
func (i *ZodIntersection[T, R]) Parse(input any, ctx ...*core.ParseContext) (R, error) {
	pc := resolveCtx(ctx)
	result, err := engine.ParseComplex[any](
		input,
		&i.internals.ZodTypeInternals,
		core.ZodTypeIntersection,
		i.extractType,
		i.extractPtr,
		i.validateValue,
		pc,
	)
	if err != nil {
		var zero R
		return zero, err
	}
	return convertToIntersectionConstraintType[T, R](result), nil
}

// extractType extracts the intersection type from input.
func (i *ZodIntersection[T, R]) extractType(input any) (any, bool) {
	return input, true
}

// extractPtr extracts pointer type from input.
func (i *ZodIntersection[T, R]) extractPtr(input any) (*any, bool) {
	if input == nil {
		return nil, true
	}
	return &input, true
}

// collectSchemaIssues extracts ZodIssues from a validation error.
func collectSchemaIssues(err error, input any, ctx *core.ParseContext) []core.ZodIssue {
	if err == nil {
		return nil
	}
	var zErr *issues.ZodError
	if issues.IsZodError(err, &zErr) {
		return zErr.Issues
	}
	issue := issues.NewRawIssue(core.Custom, input, issues.WithMessage(err.Error()))
	return []core.ZodIssue{issues.FinalizeIssue(issue, ctx, core.Config())}
}

// validateValue validates value using intersection logic.
func (i *ZodIntersection[T, R]) validateValue(value any, chks []core.ZodCheck, ctx *core.ParseContext) (any, error) {
	leftResult, leftErr := i.internals.Left.ParseAny(value, ctx)
	rightResult, rightErr := i.internals.Right.ParseAny(value, ctx)

	leftIssues := collectSchemaIssues(leftErr, value, ctx)
	rightIssues := collectSchemaIssues(rightErr, value, ctx)

	// Keys recognized by either side should not be reported as unrecognized (Zod v4).
	mergedIssues := mergeUnrecognizedKeysIssues(leftIssues, rightIssues)
	if len(mergedIssues) > 0 {
		return nil, issues.NewZodError(mergedIssues)
	}

	merged, err := mergeValues(leftResult, rightResult)
	if err != nil {
		iss := issues.CreateCustomIssue(err.Error(), map[string]any{"type": "intersection"}, value)
		return nil, issues.NewZodError([]core.ZodIssue{issues.FinalizeIssue(iss, ctx, core.Config())})
	}

	if len(chks) > 0 {
		return engine.ApplyChecks[any](merged, chks, ctx)
	}
	return merged, nil
}

// mergeUnrecognizedKeysIssues filters unrecognized_keys issues to only include
// keys that both sides reported as unrecognized (Zod v4 behavior).
func mergeUnrecognizedKeysIssues(leftIssues, rightIssues []core.ZodIssue) []core.ZodIssue {
	type side struct{ left, right bool }
	unrec := make(map[string]*side)

	var leftOther []core.ZodIssue
	for _, iss := range leftIssues {
		if iss.Code == core.UnrecognizedKeys {
			for _, k := range iss.Keys {
				if unrec[k] == nil {
					unrec[k] = &side{}
				}
				unrec[k].left = true
			}
		} else {
			leftOther = append(leftOther, iss)
		}
	}

	var rightOther []core.ZodIssue
	for _, iss := range rightIssues {
		if iss.Code == core.UnrecognizedKeys {
			for _, k := range iss.Keys {
				if unrec[k] == nil {
					unrec[k] = &side{}
				}
				unrec[k].right = true
			}
		} else {
			rightOther = append(rightOther, iss)
		}
	}

	// Only keys rejected by both sides are truly unrecognized.
	var bothKeys []string
	for k, s := range unrec {
		if s.left && s.right {
			bothKeys = append(bothKeys, k)
		}
	}

	result := leftOther
	result = append(result, rightOther...)
	if len(bothKeys) > 0 {
		result = append(result, core.ZodIssue{
			ZodIssueBase: core.ZodIssueBase{
				Code:    core.UnrecognizedKeys,
				Message: fmt.Sprintf("Unrecognized key(s) in object: %s", strings.Join(bothKeys, ", ")),
			},
			Keys: bothKeys,
		})
	}
	return result
}

// MustParse panics if Parse returns an error.
func (i *ZodIntersection[T, R]) MustParse(input any, ctx ...*core.ParseContext) R {
	result, err := i.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// ParseAny validates input and returns an untyped result for runtime scenarios.
func (i *ZodIntersection[T, R]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return i.Parse(input, ctx...)
}

// StrictParse validates input with compile-time type safety.
func (i *ZodIntersection[T, R]) StrictParse(input T, ctx ...*core.ParseContext) (R, error) {
	pc := resolveCtx(ctx)
	converted := convertToIntersectionConstraintType[T, R](input)

	leftResult, leftErr := i.internals.Left.ParseAny(converted, pc)
	rightResult, rightErr := i.internals.Right.ParseAny(converted, pc)

	leftIssues := collectSchemaIssues(leftErr, converted, pc)
	rightIssues := collectSchemaIssues(rightErr, converted, pc)

	mergedIssues := mergeUnrecognizedKeysIssues(leftIssues, rightIssues)
	if len(mergedIssues) > 0 {
		var zero R
		return zero, issues.NewZodError(mergedIssues)
	}

	merged, err := mergeValues(leftResult, rightResult)
	if err != nil {
		iss := issues.CreateCustomIssue(err.Error(), map[string]any{"type": "intersection"}, converted)
		var zero R
		return zero, issues.NewZodError([]core.ZodIssue{issues.FinalizeIssue(iss, pc, core.Config())})
	}

	if len(i.internals.Checks) > 0 {
		checked, err := engine.ApplyChecks[any](merged, i.internals.Checks, pc)
		if err != nil {
			var zero R
			return zero, err
		}
		merged = checked
	}

	return convertToIntersectionConstraintType[T, R](merged), nil
}

// MustStrictParse panics if StrictParse returns an error.
func (i *ZodIntersection[T, R]) MustStrictParse(input T, ctx ...*core.ParseContext) R {
	result, err := i.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// Optional returns a schema that accepts T or nil, with constraint type *T.
func (i *ZodIntersection[T, R]) Optional() *ZodIntersection[T, *T] {
	in := i.internals.Clone()
	in.SetOptional(true)
	return i.withPtrInternals(in)
}

// Nilable returns a schema that accepts T or nil, with constraint type *T.
func (i *ZodIntersection[T, R]) Nilable() *ZodIntersection[T, *T] {
	in := i.internals.Clone()
	in.SetNilable(true)
	return i.withPtrInternals(in)
}

// Nullish combines optional and nilable modifiers.
func (i *ZodIntersection[T, R]) Nullish() *ZodIntersection[T, *T] {
	in := i.internals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return i.withPtrInternals(in)
}

// NonOptional removes the optional flag and returns a value constraint (T).
func (i *ZodIntersection[T, R]) NonOptional() *ZodIntersection[T, T] {
	in := i.internals.Clone()
	in.SetOptional(false)
	in.SetNonOptional(true)

	return &ZodIntersection[T, T]{
		internals: &ZodIntersectionInternals{
			ZodTypeInternals: *in,
			Def:              i.internals.Def,
			Left:             i.internals.Left,
			Right:            i.internals.Right,
		},
	}
}

// Default sets a default value, preserving constraint type R.
func (i *ZodIntersection[T, R]) Default(v T) *ZodIntersection[T, R] {
	in := i.internals.Clone()
	in.SetDefaultValue(v)
	return i.withInternals(in)
}

// DefaultFunc sets a default value function, preserving constraint type R.
func (i *ZodIntersection[T, R]) DefaultFunc(fn func() T) *ZodIntersection[T, R] {
	in := i.internals.Clone()
	in.SetDefaultFunc(func() any { return fn() })
	return i.withInternals(in)
}

// Prefault sets a prefault value that goes through the full parsing pipeline.
func (i *ZodIntersection[T, R]) Prefault(v T) *ZodIntersection[T, R] {
	in := i.internals.Clone()
	in.SetPrefaultValue(v)
	return i.withInternals(in)
}

// PrefaultFunc sets a prefault value function, preserving constraint type R.
func (i *ZodIntersection[T, R]) PrefaultFunc(fn func() T) *ZodIntersection[T, R] {
	in := i.internals.Clone()
	in.SetPrefaultFunc(func() any { return fn() })
	return i.withInternals(in)
}

// Meta stores metadata for this schema.
func (i *ZodIntersection[T, R]) Meta(meta core.GlobalMeta) *ZodIntersection[T, R] {
	core.GlobalRegistry.Add(i, meta)
	return i
}

// Describe registers a description in the global registry.
func (i *ZodIntersection[T, R]) Describe(description string) *ZodIntersection[T, R] {
	in := i.internals.Clone()
	existing, ok := core.GlobalRegistry.Get(i)
	if !ok {
		existing = core.GlobalMeta{}
	}
	existing.Description = description
	clone := i.withInternals(in)
	core.GlobalRegistry.Add(clone, existing)
	return clone
}

// Left returns the left schema of the intersection.
func (i *ZodIntersection[T, R]) Left() core.ZodSchema {
	return i.internals.Left
}

// Right returns the right schema of the intersection.
func (i *ZodIntersection[T, R]) Right() core.ZodSchema {
	return i.internals.Right
}

// Transform creates a type-safe transformation pipeline.
func (i *ZodIntersection[T, R]) Transform(fn func(T, *core.RefinementContext) (any, error)) *core.ZodTransform[R, any] {
	wrapper := func(input R, ctx *core.RefinementContext) (any, error) {
		return fn(extractIntersectionValue[T, R](input), ctx)
	}
	return core.NewZodTransform[R, any](i, wrapper)
}

// Pipe creates a validation pipeline to another schema.
func (i *ZodIntersection[T, R]) Pipe(target core.ZodType[any]) *core.ZodPipe[R, any] {
	wrapper := func(input R, ctx *core.ParseContext) (any, error) {
		return target.Parse(extractIntersectionValue[T, R](input), ctx)
	}
	return core.NewZodPipe[R, any](i, target, wrapper)
}

// Refine applies type-safe validation with constraint type R.
func (i *ZodIntersection[T, R]) Refine(fn func(R) bool, params ...any) *ZodIntersection[T, R] {
	wrapper := func(v any) bool {
		if cv, ok := convertToIntersectionConstraintValue[T, R](v); ok {
			return fn(cv)
		}
		return false
	}
	in := i.internals.Clone()
	in.AddCheck(checks.NewCustom[any](wrapper, utils.NormalizeCustomParams(params...)))
	return i.withInternals(in)
}

// RefineAny applies flexible validation without type conversion.
func (i *ZodIntersection[T, R]) RefineAny(fn func(any) bool, params ...any) *ZodIntersection[T, R] {
	in := i.internals.Clone()
	in.AddCheck(checks.NewCustom[any](fn, utils.NormalizeCustomParams(params...)))
	return i.withInternals(in)
}

// And creates an intersection with another schema.
//
// Example:
//
//	schema := gozod.String().Min(3).And(gozod.String().Max(10))
//	result, _ := schema.Parse("hello")
func (i *ZodIntersection[T, R]) And(other any) *ZodIntersection[any, any] {
	return Intersection(i, other)
}

// Or creates a union with another schema.
//
// Example:
//
//	schema := gozod.String().And(gozod.String().Min(3)).Or(gozod.Int())
func (i *ZodIntersection[T, R]) Or(other any) *ZodUnion[any, any] {
	return Union([]any{i, other})
}

func (i *ZodIntersection[T, R]) withPtrInternals(in *core.ZodTypeInternals) *ZodIntersection[T, *T] {
	return &ZodIntersection[T, *T]{internals: &ZodIntersectionInternals{
		ZodTypeInternals: *in,
		Def:              i.internals.Def,
		Left:             i.internals.Left,
		Right:            i.internals.Right,
	}}
}

func (i *ZodIntersection[T, R]) withInternals(in *core.ZodTypeInternals) *ZodIntersection[T, R] {
	return &ZodIntersection[T, R]{internals: &ZodIntersectionInternals{
		ZodTypeInternals: *in,
		Def:              i.internals.Def,
		Left:             i.internals.Left,
		Right:            i.internals.Right,
	}}
}

// CloneFrom copies configuration from another schema.
func (i *ZodIntersection[T, R]) CloneFrom(source any) {
	if src, ok := source.(*ZodIntersection[T, R]); ok {
		i.internals = src.internals
	}
}

// convertToIntersectionConstraintType converts a value to constraint type R.
func convertToIntersectionConstraintType[T any, R any](value any) R {
	if value == nil {
		rt := reflect.TypeFor[R]()
		if rt.Kind() == reflect.Pointer {
			return any((*any)(nil)).(R)
		}
		var zero R
		return zero
	}

	rt := reflect.TypeFor[R]()
	if rt.Kind() == reflect.Pointer {
		if reflect.TypeOf(value).Kind() == reflect.Pointer {
			return any(value).(R) //nolint:unconvert // generic constraint conversion
		}
		return any(new(value)).(R)
	}

	// R is not a pointer; dereference if value is a pointer.
	if reflect.TypeOf(value).Kind() == reflect.Pointer {
		rv := reflect.ValueOf(value)
		if !rv.IsNil() {
			return any(rv.Elem().Interface()).(R) //nolint:unconvert // generic constraint conversion
		}
		var zero R
		return zero
	}
	return any(value).(R) //nolint:unconvert // generic constraint conversion
}

// extractIntersectionValue extracts base type T from constraint type R.
func extractIntersectionValue[T any, R any](value R) T {
	switch v := any(value).(type) {
	case *any:
		if v != nil {
			return any(*v).(T) //nolint:unconvert // generic constraint conversion
		}
		var zero T
		return zero
	default:
		return any(value).(T)
	}
}

// convertToIntersectionConstraintValue converts any value to constraint type R if possible.
func convertToIntersectionConstraintValue[T any, R any](value any) (R, bool) {
	var zero R
	if r, ok := any(value).(R); ok { //nolint:unconvert // generic constraint conversion
		return r, true
	}
	if _, ok := any(zero).(*any); ok {
		if value != nil {
			return any(new(value)).(R), true
		}
		return any((*any)(nil)).(R), true
	}
	return zero, false
}

// mergeValues attempts to merge two validated values.
func mergeValues(a, b any) (any, error) {
	if reflect.DeepEqual(a, b) {
		return a, nil
	}
	if a == nil && b == nil {
		return nil, nil
	}
	if a == nil {
		return b, nil
	}
	if b == nil {
		return a, nil
	}

	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)

	// Different struct types can be merged via map conversion.
	if av.Type() != bv.Type() {
		if av.Kind() != reflect.Struct || bv.Kind() != reflect.Struct {
			return nil, issues.CreateIncompatibleTypesError("incompatible types", a, b, nil, &core.ParseContext{})
		}
	}

	//nolint:exhaustive
	switch av.Kind() {
	case reflect.Map:
		return mergeMaps(a, b)
	case reflect.Slice, reflect.Array:
		return mergeSlices(a, b)
	case reflect.Struct:
		return mergeMaps(structToMap(av), structToMap(bv))
	default:
		return nil, issues.CreateIncompatibleTypesError("different values", a, b, nil, &core.ParseContext{})
	}
}

// mergeMaps merges two map values.
func mergeMaps(a, b any) (any, error) {
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)
	if av.Kind() != reflect.Map || bv.Kind() != reflect.Map {
		return nil, issues.CreateIncompatibleTypesError("both values must be maps", a, b, nil, &core.ParseContext{})
	}

	result := reflect.MakeMap(av.Type())
	for _, key := range av.MapKeys() {
		result.SetMapIndex(key, av.MapIndex(key))
	}
	for _, key := range bv.MapKeys() {
		bVal := bv.MapIndex(key)
		if aVal := av.MapIndex(key); aVal.IsValid() {
			if !reflect.DeepEqual(aVal.Interface(), bVal.Interface()) {
				return nil, issues.CreateIncompatibleTypesError(
					fmt.Sprintf("conflicting values for key %v", key.Interface()),
					aVal.Interface(), bVal.Interface(), nil, &core.ParseContext{},
				)
			}
		} else {
			result.SetMapIndex(key, bVal)
		}
	}
	return result.Interface(), nil
}

// mergeSlices validates that two slices are identical for intersection.
func mergeSlices(a, b any) (any, error) {
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)
	if av.Kind() != reflect.Slice && av.Kind() != reflect.Array {
		return nil, issues.CreateIncompatibleTypesError("first value must be slice or array", a, b, nil, &core.ParseContext{})
	}
	if bv.Kind() != reflect.Slice && bv.Kind() != reflect.Array {
		return nil, issues.CreateIncompatibleTypesError("second value must be slice or array", a, b, nil, &core.ParseContext{})
	}
	if !reflect.DeepEqual(a, b) {
		return nil, issues.CreateIncompatibleTypesError("slice/array values must be identical for intersection", a, b, nil, &core.ParseContext{})
	}
	return a, nil
}

// structToMap converts a struct to map[string]any using exported fields and json tags.
func structToMap(v reflect.Value) map[string]any {
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}
	t := v.Type()
	m := make(map[string]any, t.NumField())
	for i := range t.NumField() {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}
		key := f.Name
		if tag := f.Tag.Get("json"); tag != "" {
			if name := strings.Split(tag, ",")[0]; name != "" && name != "-" {
				key = name
			}
		}
		m[key] = v.Field(i).Interface()
	}
	return m
}

// newZodIntersectionFromDef constructs a new ZodIntersection from a definition.
func newZodIntersectionFromDef[T any, R any](def *ZodIntersectionDef) *ZodIntersection[T, R] {
	in := &ZodIntersectionInternals{
		ZodTypeInternals: engine.NewBaseZodTypeInternals(def.Type),
		Def:              def,
		Left:             def.Left,
		Right:            def.Right,
	}
	in.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any] {
		d := &ZodIntersectionDef{ZodTypeDef: *newDef, Left: def.Left, Right: def.Right}
		return any(newZodIntersectionFromDef[T, R](d)).(core.ZodType[any])
	}
	if def.Error != nil {
		in.Error = def.Error
	}
	for _, c := range def.Checks {
		in.AddCheck(c)
	}
	return &ZodIntersection[T, R]{internals: in}
}

// Intersection creates an intersection schema with value constraint.
func Intersection(left, right any, args ...any) *ZodIntersection[any, any] {
	return IntersectionTyped[any, any](left, right, args...)
}

// IntersectionPtr creates an intersection schema with pointer constraint.
func IntersectionPtr(left, right any, args ...any) *ZodIntersection[any, *any] {
	return IntersectionTyped[any, *any](left, right, args...)
}

// IntersectionTyped creates a typed intersection schema with generic constraints.
func IntersectionTyped[T any, R any](left, right any, args ...any) *ZodIntersection[T, R] {
	p := utils.NormalizeParams(utils.FirstParam(args...))
	leftSchema, err := core.ConvertToZodSchema(left)
	if err != nil {
		panic(fmt.Sprintf("Intersection left schema: %v", err))
	}
	rightSchema, err := core.ConvertToZodSchema(right)
	if err != nil {
		panic(fmt.Sprintf("Intersection right schema: %v", err))
	}
	def := &ZodIntersectionDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeIntersection,
			Checks: []core.ZodCheck{},
		},
		Left:  leftSchema,
		Right: rightSchema,
	}
	utils.ApplySchemaParams(&def.ZodTypeDef, p)
	return newZodIntersectionFromDef[T, R](def)
}

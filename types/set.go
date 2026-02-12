package types

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/internal/utils"
)

// ZodSetDef defines the configuration for a set schema.
type ZodSetDef struct {
	core.ZodTypeDef
	ValueType any
}

// ZodSetInternals holds the internal state of a set validator.
type ZodSetInternals[T comparable] struct {
	core.ZodTypeInternals
	Def       *ZodSetDef
	ValueType any
}

// ZodSet is a type-safe set validation schema.
// T is the element type (must be comparable).
// R is the constraint type (value or pointer).
type ZodSet[T comparable, R any] struct {
	internals *ZodSetInternals[T]
}

// Internals returns the internal state of the schema.
func (z *ZodSet[T, R]) Internals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// IsOptional reports whether this schema accepts undefined or missing values.
func (z *ZodSet[T, R]) IsOptional() bool {
	return z.internals.IsOptional()
}

// IsNilable reports whether this schema accepts nil values.
func (z *ZodSet[T, R]) IsNilable() bool {
	return z.internals.IsNilable()
}

func (z *ZodSet[T, R]) withCheck(c core.ZodCheck) *ZodSet[T, R] {
	in := z.internals.Clone()
	in.AddCheck(c)
	return z.withInternals(in)
}

// Parse validates input using set-specific parsing logic.
func (z *ZodSet[T, R]) Parse(input any, ctx ...*core.ParseContext) (R, error) {
	var zero R

	result, err := engine.ParseComplex[map[T]struct{}](
		input,
		&z.internals.ZodTypeInternals,
		core.ZodTypeSet,
		z.extractForEngine,
		z.extractPtrForEngine,
		z.validateForEngine,
		ctx...,
	)
	if err != nil {
		return zero, err
	}

	switch v := result.(type) {
	case map[T]struct{}:
		return convertToSetConstraintType[T, R](v), nil
	case *map[T]struct{}:
		return convertToSetConstraintType[T, R](v), nil
	case **map[T]struct{}:
		if v != nil {
			return convertToSetConstraintType[T, R](*v), nil
		}
		return convertToSetConstraintType[T, R](nil), nil
	case nil:
		return convertToSetConstraintType[T, R](nil), nil
	default:
		if typed, ok := result.(R); ok {
			return typed, nil
		}
		pc := core.NewParseContext()
		if len(ctx) > 0 && ctx[0] != nil {
			pc = ctx[0]
		}
		return zero, issues.CreateTypeConversionError(
			fmt.Sprintf("%T", result), "set", input, pc,
		)
	}
}

// MustParse validates input and panics on error.
func (z *ZodSet[T, R]) MustParse(input any, ctx ...*core.ParseContext) R {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// StrictParse validates input with compile-time type safety.
func (z *ZodSet[T, R]) StrictParse(input map[T]struct{}, ctx ...*core.ParseContext) (R, error) {
	cv := convertToSetConstraintType[T, R](input)

	return engine.ParseComplexStrict[map[T]struct{}, R](
		cv,
		&z.internals.ZodTypeInternals,
		core.ZodTypeSet,
		z.extractForEngine,
		z.extractPtrForEngine,
		z.validateForEngine,
		ctx...,
	)
}

// MustStrictParse validates input with type safety and panics on error.
func (z *ZodSet[T, R]) MustStrictParse(input map[T]struct{}, ctx ...*core.ParseContext) R {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// ParseAny validates input and returns an untyped result.
func (z *ZodSet[T, R]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// ValueType returns the value schema for this set.
func (z *ZodSet[T, R]) ValueType() any {
	return z.internals.ValueType
}

// Optional returns a schema that accepts nil.
// The constraint type becomes *map[T]struct{}.
func (z *ZodSet[T, R]) Optional() *ZodSet[T, *map[T]struct{}] {
	in := z.internals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// Nilable returns a schema that accepts nil.
// The constraint type becomes *map[T]struct{}.
func (z *ZodSet[T, R]) Nilable() *ZodSet[T, *map[T]struct{}] {
	in := z.internals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish combines optional and nilable modifiers.
func (z *ZodSet[T, R]) Nullish() *ZodSet[T, *map[T]struct{}] {
	in := z.internals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// NonOptional removes the Optional flag and returns base constraint type.
func (z *ZodSet[T, R]) NonOptional() *ZodSet[T, map[T]struct{}] {
	in := z.internals.Clone()
	in.SetOptional(false)
	in.SetNonOptional(true)
	return &ZodSet[T, map[T]struct{}]{internals: &ZodSetInternals[T]{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
		ValueType:        z.internals.ValueType,
	}}
}

// Default sets a value returned when input is nil.
// The default value bypasses validation.
func (z *ZodSet[T, R]) Default(v map[T]struct{}) *ZodSet[T, R] {
	in := z.internals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

// DefaultFunc sets a factory for the default value when input is nil.
func (z *ZodSet[T, R]) DefaultFunc(fn func() map[T]struct{}) *ZodSet[T, R] {
	in := z.internals.Clone()
	in.SetDefaultFunc(func() any { return fn() })
	return z.withInternals(in)
}

// Prefault sets a value that goes through full validation when input is nil.
func (z *ZodSet[T, R]) Prefault(v map[T]struct{}) *ZodSet[T, R] {
	in := z.internals.Clone()
	in.SetPrefaultValue(v)
	return z.withInternals(in)
}

// PrefaultFunc sets a factory for the prefault value.
// The prefault value goes through full validation.
func (z *ZodSet[T, R]) PrefaultFunc(fn func() map[T]struct{}) *ZodSet[T, R] {
	in := z.internals.Clone()
	in.SetPrefaultFunc(func() any { return fn() })
	return z.withInternals(in)
}

// Meta stores metadata for this set schema.
func (z *ZodSet[T, R]) Meta(meta core.GlobalMeta) *ZodSet[T, R] {
	core.GlobalRegistry.Add(z, meta)
	return z
}

// Describe registers a description in the global registry.
func (z *ZodSet[T, R]) Describe(description string) *ZodSet[T, R] {
	in := z.internals.Clone()
	existing, ok := core.GlobalRegistry.Get(z)
	if !ok {
		existing = core.GlobalMeta{}
	}
	existing.Description = description
	clone := z.withInternals(in)
	core.GlobalRegistry.Add(clone, existing)
	return clone
}

// Min sets the minimum number of elements.
func (z *ZodSet[T, R]) Min(minLen int, params ...any) *ZodSet[T, R] {
	sp := utils.NormalizeParams(params...)
	var errMsg any
	if sp.Error != nil {
		errMsg = sp.Error
	}
	return z.withCheck(checks.MinSize(minLen, errMsg))
}

// Max sets the maximum number of elements.
func (z *ZodSet[T, R]) Max(maxLen int, params ...any) *ZodSet[T, R] {
	sp := utils.NormalizeParams(params...)
	var errMsg any
	if sp.Error != nil {
		errMsg = sp.Error
	}
	return z.withCheck(checks.MaxSize(maxLen, errMsg))
}

// Length sets the exact number of elements required.
func (z *ZodSet[T, R]) Length(exactLen int, params ...any) *ZodSet[T, R] {
	sp := utils.NormalizeParams(params...)
	var errMsg any
	if sp.Error != nil {
		errMsg = sp.Error
	}
	return z.withCheck(checks.Size(exactLen, errMsg))
}

// NonEmpty ensures the set has at least one element.
// This is equivalent to Min(1).
func (z *ZodSet[T, R]) NonEmpty(params ...any) *ZodSet[T, R] {
	return z.Min(1, params...)
}

// Transform applies a transformation function to the parsed set value.
func (z *ZodSet[T, R]) Transform(fn func(R, *core.RefinementContext) (any, error)) *core.ZodTransform[R, any] {
	wrap := func(input R, ctx *core.RefinementContext) (any, error) {
		return fn(input, ctx)
	}
	return core.NewZodTransform[R, any](z, wrap)
}

// Overwrite transforms the value while preserving the original type.
func (z *ZodSet[T, R]) Overwrite(transform func(R) R, params ...any) *ZodSet[T, R] {
	fn := func(input any) any {
		if converted, ok := convertToSetType[T, R](input); ok {
			return transform(converted)
		}
		return input
	}
	return z.withCheck(checks.NewZodCheckOverwrite(fn, params...))
}

// Pipe creates a pipeline that passes the parsed result to a target schema.
func (z *ZodSet[T, R]) Pipe(target core.ZodType[any]) *core.ZodPipe[R, any] {
	wrap := func(input R, ctx *core.ParseContext) (any, error) {
		return target.Parse(input, ctx)
	}
	return core.NewZodPipe[R, any](z, target, wrap)
}

// Refine adds a typed custom validation function.
func (z *ZodSet[T, R]) Refine(fn func(R) bool, params ...any) *ZodSet[T, R] {
	wrap := func(v any) bool {
		return fn(convertToSetConstraintType[T, R](v))
	}
	sp := utils.NormalizeParams(params...)
	var errMsg any
	if sp.Error != nil {
		errMsg = sp.Error
	}
	return z.withCheck(checks.NewCustom[any](wrap, errMsg))
}

// RefineAny adds a custom validation function accepting any input.
func (z *ZodSet[T, R]) RefineAny(fn func(any) bool, params ...any) *ZodSet[T, R] {
	sp := utils.NormalizeParams(params...)
	var errMsg any
	if sp.Error != nil {
		errMsg = sp.Error
	}
	return z.withCheck(checks.NewCustom[any](fn, errMsg))
}

// And creates an intersection with another schema.
func (z *ZodSet[T, R]) And(other any) *ZodIntersection[any, any] {
	return Intersection(z, other)
}

// Or creates a union with another schema.
func (z *ZodSet[T, R]) Or(other any) *ZodUnion[any, any] {
	return Union([]any{z, other})
}

func (z *ZodSet[T, R]) withInternals(in *core.ZodTypeInternals) *ZodSet[T, R] {
	return &ZodSet[T, R]{internals: &ZodSetInternals[T]{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
		ValueType:        z.internals.ValueType,
	}}
}

func (z *ZodSet[T, R]) withPtrInternals(in *core.ZodTypeInternals) *ZodSet[T, *map[T]struct{}] {
	return &ZodSet[T, *map[T]struct{}]{internals: &ZodSetInternals[T]{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
		ValueType:        z.internals.ValueType,
	}}
}

// CloneFrom copies configuration from another schema of the same type.
func (z *ZodSet[T, R]) CloneFrom(source any) {
	if src, ok := source.(*ZodSet[T, R]); ok {
		z.internals = src.internals
	}
}

// extractForEngine extracts map[T]struct{} from input for engine.ParseComplex.
func (z *ZodSet[T, R]) extractForEngine(value any) (map[T]struct{}, bool) {
	if s, ok := value.(map[T]struct{}); ok {
		return s, true
	}
	if slice, ok := value.([]T); ok {
		set := make(map[T]struct{}, len(slice))
		for _, elem := range slice {
			set[elem] = struct{}{}
		}
		return set, true
	}
	if anySlice, ok := value.([]any); ok {
		set := make(map[T]struct{}, len(anySlice))
		for _, elem := range anySlice {
			typed, ok := elem.(T)
			if !ok {
				return nil, false
			}
			set[typed] = struct{}{}
		}
		return set, true
	}
	if value == nil {
		return nil, false
	}

	// Fall back to reflection for non-standard map or slice types.
	rv := reflect.ValueOf(value)
	if rv.Kind() == reflect.Map && rv.Type().Elem() == reflect.TypeOf(struct{}{}) {
		set := make(map[T]struct{}, rv.Len())
		for _, key := range rv.MapKeys() {
			typed, ok := key.Interface().(T)
			if !ok {
				return nil, false
			}
			set[typed] = struct{}{}
		}
		return set, true
	}
	if rv.Kind() == reflect.Slice {
		set := make(map[T]struct{}, rv.Len())
		for i := range rv.Len() {
			typed, ok := rv.Index(i).Interface().(T)
			if !ok {
				return nil, false
			}
			set[typed] = struct{}{}
		}
		return set, true
	}
	return nil, false
}

// extractPtrForEngine extracts *map[T]struct{} from input for engine.ParseComplex.
func (z *ZodSet[T, R]) extractPtrForEngine(value any) (*map[T]struct{}, bool) {
	if ptr, ok := value.(*map[T]struct{}); ok {
		return ptr, true
	}
	if s, ok := value.(map[T]struct{}); ok {
		return &s, true
	}
	if s, ok := z.extractForEngine(value); ok {
		return &s, true
	}
	return nil, false
}

// validateForEngine validates set elements and runs checks.
func (z *ZodSet[T, R]) validateForEngine(value map[T]struct{}, chks []core.ZodCheck, ctx *core.ParseContext) (map[T]struct{}, error) {
	transformed, err := engine.ApplyChecks[any](value, chks, ctx)
	if err != nil {
		return nil, err
	}

	switch v := transformed.(type) {
	case map[T]struct{}:
		value = v
	case *map[T]struct{}:
		if v != nil {
			value = *v
		}
	}

	var collected []core.ZodRawIssue
	if z.internals.ValueType != nil {
		for elem := range value {
			collected = z.collectErrors(elem, z.internals.ValueType, elem, ctx, collected)
		}
	}

	if len(collected) > 0 {
		return nil, issues.CreateArrayValidationIssues(collected)
	}
	return value, nil
}

func (z *ZodSet[T, R]) collectErrors(value any, schema any, pathKey any, ctx *core.ParseContext, dst []core.ZodRawIssue) []core.ZodRawIssue {
	err := z.validateDirect(value, schema, ctx)
	if err == nil {
		return dst
	}

	var zodErr *issues.ZodError
	if errors.As(err, &zodErr) {
		for _, issue := range zodErr.Issues {
			dst = append(dst, issues.ConvertZodIssueToRawWithPrependedPath(issue, []any{pathKey}))
		}
	} else {
		raw := issues.CreateIssue(core.Custom, err.Error(), nil, value)
		raw.Path = []any{pathKey}
		dst = append(dst, raw)
	}
	return dst
}

// validateDirect validates a single value against a schema using reflection.
func (z *ZodSet[T, R]) validateDirect(value, schema any, ctx *core.ParseContext) error {
	if schema == nil {
		return nil
	}
	if ctx == nil {
		ctx = core.NewParseContext()
	}

	sv := reflect.ValueOf(schema)
	if !sv.IsValid() || sv.IsNil() {
		return nil
	}

	method := sv.MethodByName("Parse")
	if !method.IsValid() {
		return nil
	}

	mt := method.Type()
	if mt.NumIn() < 1 {
		return nil
	}

	args := []reflect.Value{reflect.ValueOf(value)}
	if mt.NumIn() > 1 && mt.In(1).String() == "*core.ParseContext" {
		args = append(args, reflect.ValueOf(ctx))
	}

	results := method.Call(args)
	if len(results) >= 2 {
		if ei := results[1].Interface(); ei != nil {
			if err, ok := ei.(error); ok {
				return err
			}
		}
	}
	return nil
}

func convertToSetType[T comparable, R any](v any) (R, bool) {
	var zero R

	if set, ok := v.(map[T]struct{}); ok {
		return convertToSetConstraintType[T, R](set), true
	}
	if ptr, ok := v.(*map[T]struct{}); ok {
		return convertToSetConstraintType[T, R](ptr), true
	}
	return zero, false
}

func convertToSetConstraintType[T comparable, R any](value any) R {
	var zero R
	rt := reflect.TypeOf(zero)

	if value == nil {
		return zero
	}

	if rt != nil && rt.Kind() == reflect.Pointer {
		switch v := value.(type) {
		case map[T]struct{}:
			return any(&v).(R)
		case *map[T]struct{}:
			return any(v).(R)
		}
	} else {
		switch v := value.(type) {
		case map[T]struct{}:
			return any(v).(R)
		case *map[T]struct{}:
			if v != nil {
				return any(*v).(R)
			}
		}
	}

	if converted, ok := value.(R); ok {
		return converted
	}
	return zero
}

func newZodSetFromDef[T comparable, R any](def *ZodSetDef) *ZodSet[T, R] {
	in := &ZodSetInternals[T]{
		ZodTypeInternals: engine.NewBaseZodTypeInternals(def.Type),
		Def:              def,
		ValueType:        def.ValueType,
	}

	in.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any] {
		sd := &ZodSetDef{
			ZodTypeDef: *newDef,
			ValueType:  def.ValueType,
		}
		return any(newZodSetFromDef[T, R](sd)).(core.ZodType[any])
	}

	if def.Error != nil {
		in.Error = def.Error
	}
	if len(def.Checks) > 0 {
		for _, c := range def.Checks {
			in.AddCheck(c)
		}
	}
	return &ZodSet[T, R]{internals: in}
}

// Set creates a set schema with element validation.
func Set[T comparable](valueSchema any, paramArgs ...any) *ZodSet[T, map[T]struct{}] {
	return SetTyped[T, map[T]struct{}](valueSchema, paramArgs...)
}

// SetPtr creates a set schema returning a pointer constraint.
func SetPtr[T comparable](valueSchema any, paramArgs ...any) *ZodSet[T, *map[T]struct{}] {
	return SetTyped[T, *map[T]struct{}](valueSchema, paramArgs...)
}

// SetTyped creates a typed set schema with explicit generic constraints.
func SetTyped[T comparable, R any](valueSchema any, paramArgs ...any) *ZodSet[T, R] {
	p := utils.FirstParam(paramArgs...)
	sp := utils.NormalizeParams(p)

	def := &ZodSetDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeSet,
			Checks: []core.ZodCheck{},
		},
		ValueType: valueSchema,
	}
	if sp != nil {
		utils.ApplySchemaParams(&def.ZodTypeDef, sp)
	}

	schema := newZodSetFromDef[T, R](def)

	// Register a pass-through check so the engine invokes validateForEngine
	// for element validation.
	if valueSchema != nil {
		schema.internals.AddCheck(checks.NewCustom[any](func(v any) bool { return true }, core.SchemaParams{}))
	}
	return schema
}

// Check adds a custom validation function that can report multiple issues.
func (z *ZodSet[T, R]) Check(fn func(value R, payload *core.ParsePayload), params ...any) *ZodSet[T, R] {
	wrap := func(payload *core.ParsePayload) {
		if val, ok := payload.Value().(R); ok {
			fn(val, payload)
			return
		}

		// Handle pointer/value mismatch when R is *map but value is map.
		var zero R
		zt := reflect.TypeOf(zero)
		if zt != nil && zt.Kind() == reflect.Pointer {
			et := zt.Elem()
			rv := reflect.ValueOf(payload.Value())
			if rv.IsValid() && rv.Type() == et {
				ptr := reflect.New(et)
				ptr.Elem().Set(rv)
				if casted, ok := ptr.Interface().(R); ok {
					fn(casted, payload)
				}
			}
		}
	}
	return z.withCheck(checks.NewCustom[any](wrap, utils.NormalizeCustomParams(params...)))
}

// With is an alias for Check.
func (z *ZodSet[T, R]) With(fn func(value R, payload *core.ParsePayload), params ...any) *ZodSet[T, R] {
	return z.Check(fn, params...)
}

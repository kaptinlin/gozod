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

var (
	errNilArrayPtr = errors.New("nil pointer to array")
	errNilPtr      = errors.New("nil pointer")
	errNotArray    = errors.New("expected array or slice")
)

// ZodArrayDef holds the schema definition for fixed-length array validation.
type ZodArrayDef struct {
	core.ZodTypeDef
	Items []any // element schemas per position
	Rest  any   // rest schema for variadic elements (nil if none)
}

// ZodArrayInternals contains the internal state for an array schema.
type ZodArrayInternals struct {
	core.ZodTypeInternals
	Def   *ZodArrayDef
	Items []any
	Rest  any
}

// ZodArray is a type-safe fixed-length array validation schema.
// T is the base array type, R is the constraint type (T or *T).
type ZodArray[T any, R any] struct {
	internals *ZodArrayInternals
}

// Internals returns the internal state for framework usage.
func (z *ZodArray[T, R]) Internals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// IsOptional reports whether this schema accepts undefined/missing values.
func (z *ZodArray[T, R]) IsOptional() bool {
	return z.internals.IsOptional()
}

// IsNilable reports whether this schema accepts nil values.
func (z *ZodArray[T, R]) IsNilable() bool {
	return z.internals.IsNilable()
}

// Parse validates input and returns the parsed array value.
func (z *ZodArray[T, R]) Parse(input any, ctx ...*core.ParseContext) (R, error) {
	result, err := engine.ParseComplex[[]any](
		input,
		&z.internals.ZodTypeInternals,
		core.ZodTypeArray,
		z.extractForEngine,
		z.extractPtrForEngine,
		z.validateForEngine,
		ctx...,
	)
	if err != nil {
		var zero R
		return zero, err
	}

	switch v := result.(type) {
	case []any:
		return toConstraint[T, R](v), nil
	case *[]any:
		if v == nil {
			var zero R
			return zero, nil
		}
		return toConstraint[T, R](*v), nil
	case nil:
		var zero R
		return zero, nil
	default:
		if typed, ok := result.(R); ok {
			return typed, nil
		}
		var zero R
		pctx := core.NewParseContext()
		if len(ctx) > 0 && ctx[0] != nil {
			pctx = ctx[0]
		}
		return zero, issues.CreateInvalidTypeError(core.ZodTypeArray, result, pctx)
	}
}

// MustParse validates the input value and panics on failure
func (z *ZodArray[T, R]) MustParse(input any, ctx ...*core.ParseContext) R {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// StrictParse validates input with compile-time type safety.
func (z *ZodArray[T, R]) StrictParse(input T, ctx ...*core.ParseContext) (R, error) {
	converted, ok := toArrayType[T, R](input)
	if !ok {
		var zero R
		if len(ctx) == 0 {
			ctx = []*core.ParseContext{core.NewParseContext()}
		}
		return zero, issues.CreateTypeConversionError(
			fmt.Sprintf("%T", input),
			"array constraint type",
			any(input),
			ctx[0],
		)
	}

	return engine.ParseComplexStrict[[]any, R](
		converted,
		&z.internals.ZodTypeInternals,
		core.ZodTypeArray,
		z.extractForEngine,
		z.extractPtrForEngine,
		z.validateForEngine,
		ctx...,
	)
}

// MustStrictParse validates input with compile-time type safety and panics on failure.
func (z *ZodArray[T, R]) MustStrictParse(input T, ctx ...*core.ParseContext) R {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// ParseAny validates input and returns an untyped result.
func (z *ZodArray[T, R]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// Optional returns a schema that accepts nil values with pointer constraint.
func (z *ZodArray[T, R]) Optional() *ZodArray[T, *T] {
	in := z.internals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// ExactOptional accepts absent keys but rejects explicit nil values.
// Unlike Optional, ExactOptional only accepts absent keys in object fields.
func (z *ZodArray[T, R]) ExactOptional() *ZodArray[T, R] {
	in := z.internals.Clone()
	in.SetExactOptional(true)
	return z.withInternals(in)
}

// Nilable returns a schema that accepts nil values with pointer constraint.
func (z *ZodArray[T, R]) Nilable() *ZodArray[T, *T] {
	in := z.internals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish combines optional and nilable modifiers.
func (z *ZodArray[T, R]) Nullish() *ZodArray[T, *T] {
	in := z.internals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// NonOptional enforces non-nil value constraint, producing a
// "expected = nonoptional" error when input is nil.
func (z *ZodArray[T, R]) NonOptional() *ZodArray[T, T] {
	in := z.internals.Clone()
	in.SetOptional(false)
	in.SetNonOptional(true)

	return &ZodArray[T, T]{
		internals: &ZodArrayInternals{
			ZodTypeInternals: *in,
			Def:              z.internals.Def,
			Items:            z.internals.Items,
			Rest:             z.internals.Rest,
		},
	}
}

// Default sets a default value used when input is nil.
func (z *ZodArray[T, R]) Default(v T) *ZodArray[T, R] {
	in := z.internals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

// DefaultFunc sets a function that provides the default value.
func (z *ZodArray[T, R]) DefaultFunc(fn func() T) *ZodArray[T, R] {
	in := z.internals.Clone()
	in.SetDefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Prefault sets a fallback value used through the full parsing pipeline.
func (z *ZodArray[T, R]) Prefault(v T) *ZodArray[T, R] {
	in := z.internals.Clone()
	in.SetPrefaultValue(v)
	return z.withInternals(in)
}

// PrefaultFunc sets a function that provides the fallback value.
func (z *ZodArray[T, R]) PrefaultFunc(fn func() T) *ZodArray[T, R] {
	in := z.internals.Clone()
	in.SetPrefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Meta stores metadata for this array schema in the global registry.
func (z *ZodArray[T, R]) Meta(meta core.GlobalMeta) *ZodArray[T, R] {
	core.GlobalRegistry.Add(z, meta)
	return z
}

// Describe sets a description for this schema in the global registry.
func (z *ZodArray[T, R]) Describe(description string) *ZodArray[T, R] {
	in := z.internals.Clone()

	meta, ok := core.GlobalRegistry.Get(z)
	if !ok {
		meta = core.GlobalMeta{}
	}
	meta.Description = description

	clone := z.withInternals(in)
	core.GlobalRegistry.Add(clone, meta)

	return clone
}

// Min adds a minimum element count constraint.
func (z *ZodArray[T, R]) Min(n int, args ...any) *ZodArray[T, R] {
	in := z.internals.Clone()
	in.AddCheck(checks.MinLength(n, utils.FirstParam(args...)))
	return z.withInternals(in)
}

// Max adds a maximum element count constraint.
func (z *ZodArray[T, R]) Max(n int, args ...any) *ZodArray[T, R] {
	in := z.internals.Clone()
	in.AddCheck(checks.MaxLength(n, utils.FirstParam(args...)))
	return z.withInternals(in)
}

// Length adds an exact element count constraint.
func (z *ZodArray[T, R]) Length(n int, args ...any) *ZodArray[T, R] {
	in := z.internals.Clone()
	in.AddCheck(checks.Length(n, utils.FirstParam(args...)))
	return z.withInternals(in)
}

// NonEmpty requires at least one element.
func (z *ZodArray[T, R]) NonEmpty(args ...any) *ZodArray[T, R] {
	return z.Min(1, utils.FirstParam(args...))
}

// Element returns the schema at the given index, or nil if out of range.
func (z *ZodArray[T, R]) Element(index int) any {
	if index >= 0 && index < len(z.internals.Items) {
		return z.internals.Items[index]
	}
	return nil
}

// ElementSchemas returns a copy of all element schemas.
func (z *ZodArray[T, R]) ElementSchemas() []any {
	result := make([]any, len(z.internals.Items))
	copy(result, z.internals.Items)
	return result
}

// RestSchema returns the rest parameter schema, or nil if none.
func (z *ZodArray[T, R]) RestSchema() any {
	return z.internals.Rest
}

// Transform applies a transformation function to the parsed value.
func (z *ZodArray[T, R]) Transform(fn func(T, *core.RefinementContext) (any, error)) *core.ZodTransform[R, any] {
	wrapper := func(input R, ctx *core.RefinementContext) (any, error) {
		return fn(baseValue[T, R](input), ctx)
	}
	return core.NewZodTransform[R, any](z, wrapper)
}

// Overwrite transforms the value while preserving the original type.
func (z *ZodArray[T, R]) Overwrite(fn func(R) R, params ...any) *ZodArray[T, R] {
	wrap := func(input any) any {
		converted, ok := toArrayType[T, R](input)
		if !ok {
			return input
		}
		return fn(converted)
	}

	in := z.internals.Clone()
	in.AddCheck(checks.NewZodCheckOverwrite(wrap, params...))
	return z.withInternals(in)
}

// Pipe passes the parsed value to a target schema.
func (z *ZodArray[T, R]) Pipe(target core.ZodType[any]) *core.ZodPipe[R, any] {
	wrapper := func(input R, ctx *core.ParseContext) (any, error) {
		return target.Parse(baseValue[T, R](input), ctx)
	}
	return core.NewZodPipe[R, any](z, target, wrapper)
}

// Refine adds type-safe custom validation.
func (z *ZodArray[T, R]) Refine(fn func(R) bool, params ...any) *ZodArray[T, R] {
	wrapper := func(v any) bool {
		var zero R
		switch any(zero).(type) {
		case *T:
			if v == nil {
				return fn(any((*T)(nil)).(R))
			}
			if val, ok := v.(T); ok {
				cp := val
				return fn(any(&cp).(R))
			}
			return false
		default:
			if v == nil {
				return false
			}
			if val, ok := v.(T); ok {
				return fn(any(val).(R))
			}
			return false
		}
	}

	sp := utils.NormalizeParams(params...)
	var msg any
	if sp.Error != nil {
		msg = sp.Error
	}

	in := z.internals.Clone()
	in.AddCheck(checks.NewCustom[any](wrapper, msg))
	return z.withInternals(in)
}

// RefineAny adds custom validation without type conversion.
func (z *ZodArray[T, R]) RefineAny(fn func(any) bool, params ...any) *ZodArray[T, R] {
	sp := utils.NormalizeParams(params...)
	var msg any
	if sp.Error != nil {
		msg = sp.Error
	}

	in := z.internals.Clone()
	in.AddCheck(checks.NewCustom[any](fn, msg))
	return z.withInternals(in)
}

// And creates an intersection with another schema.
func (z *ZodArray[T, R]) And(other any) *ZodIntersection[any, any] {
	return Intersection(z, other)
}

// Or creates a union with another schema.
func (z *ZodArray[T, R]) Or(other any) *ZodUnion[any, any] {
	return Union([]any{z, other})
}

// withPtrInternals creates a new ZodArray with pointer constraint *T.
func (z *ZodArray[T, R]) withPtrInternals(in *core.ZodTypeInternals) *ZodArray[T, *T] {
	return &ZodArray[T, *T]{internals: &ZodArrayInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
		Items:            z.internals.Items,
		Rest:             z.internals.Rest,
	}}
}

// withInternals creates a new ZodArray keeping the constraint type R.
func (z *ZodArray[T, R]) withInternals(in *core.ZodTypeInternals) *ZodArray[T, R] {
	return &ZodArray[T, R]{internals: &ZodArrayInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
		Items:            z.internals.Items,
		Rest:             z.internals.Rest,
	}}
}

// CloneFrom copies configuration from another schema.
func (z *ZodArray[T, R]) CloneFrom(source any) {
	if src, ok := source.(*ZodArray[T, R]); ok {
		originalChecks := z.internals.Checks
		*z.internals = *src.internals
		z.internals.Checks = originalChecks
	}
}

// toConstraint converts []any to the constraint type R.
func toConstraint[T any, R any](v []any) R {
	if r, ok := any(v).(R); ok {
		return r
	}

	// Pointer types (e.g. *[]any): wrap in pointer.
	var zero R
	if reflect.TypeFor[R]().Kind() == reflect.Pointer {
		if r, ok := any(&v).(R); ok {
			return r
		}
	}

	return zero
}

// toArrayType converts any value to the array constraint type R.
func toArrayType[T any, R any](value any) (R, bool) {
	var zero R

	if value == nil {
		if reflect.TypeFor[R]().Kind() == reflect.Pointer {
			return zero, true
		}
		return zero, false
	}

	var arr []any

	switch v := value.(type) {
	case []any:
		arr = v
	case *[]any:
		if v == nil {
			return zero, false
		}
		arr = *v
	default:
		rv := reflect.ValueOf(value)
		if rv.Kind() != reflect.Slice {
			return zero, false
		}
		arr = make([]any, rv.Len())
		for i := range rv.Len() {
			arr[i] = rv.Index(i).Interface()
		}
	}

	// Convert to target constraint type R.
	rt := reflect.TypeFor[R]()
	//nolint:exhaustive
	switch rt.Kind() {
	case reflect.Slice:
		if reflect.TypeOf(value).AssignableTo(rt) {
			return value.(R), true
		}
	case reflect.Pointer:
		if r, ok := any(&arr).(R); ok {
			return r, true
		}
	}

	return zero, false
}

// baseValue extracts the base value T from constraint type R.
func baseValue[T any, R any](value R) T {
	if ptr, ok := any(value).(*T); ok && ptr != nil {
		return *ptr
	}
	return any(value).(T)
}

// extract converts input to []any.
func (z *ZodArray[T, R]) extract(value any) ([]any, error) {
	switch v := value.(type) {
	case []any:
		return v, nil
	case *[]any:
		if v != nil {
			return *v, nil
		}
		return nil, errNilArrayPtr
	default:
		rv := reflect.ValueOf(value)

		// Handle pointer to slice/array.
		if rv.Kind() == reflect.Pointer {
			if rv.IsNil() {
				return nil, errNilPtr
			}
			rv = rv.Elem()
		}

		if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
			return nil, fmt.Errorf("%w: got %T", errNotArray, value)
		}

		result := make([]any, rv.Len())
		for i := range rv.Len() {
			result[i] = rv.Index(i).Interface()
		}
		return result, nil
	}
}

// validate validates array content and collects all issues.
func (z *ZodArray[T, R]) validate(value []any, chks []core.ZodCheck, ctx *core.ParseContext) ([]any, error) {
	if z.internals == nil {
		return nil, issues.CreateInvalidTypeError(core.ZodTypeArray, value, ctx)
	}

	// Apply checks (including Overwrite transformations) first.
	value, err := engine.ApplyChecks[[]any](value, chks, ctx)
	if err != nil {
		return nil, err
	}

	fixed := len(z.internals.Items)
	actual := len(value)
	hasRest := z.internals.Rest != nil

	// Fail fast on length errors (Zod v4 behavior).
	if hasRest {
		if actual < fixed {
			issue := issues.CreateTooSmallIssue(fixed, true, "array", value)
			issue.Properties["is_rest_param"] = true
			return nil, issues.CreateArrayValidationIssues([]core.ZodRawIssue{issue})
		}
	} else if actual != fixed {
		issue := issues.CreateFixedLengthArrayIssue(fixed, actual, value, actual < fixed)
		return nil, issues.CreateArrayValidationIssues([]core.ZodRawIssue{issue})
	}

	// Validate elements and collect errors.
	var errs []core.ZodRawIssue

	for i := range min(fixed, actual) {
		if err := parseElement(value[i], z.internals.Items[i]); err != nil {
			errs = append(errs, issues.CreateElementValidationIssue(i, "array", value[i], err))
		}
	}

	if hasRest && actual > fixed {
		for i := fixed; i < actual; i++ {
			if err := parseElement(value[i], z.internals.Rest); err != nil {
				errs = append(errs, issues.CreateElementValidationIssue(i, "array rest", value[i], err))
			}
		}
	}

	if len(errs) > 0 {
		return nil, issues.CreateArrayValidationIssues(errs)
	}

	return value, nil
}

// parseElement validates a single element against its schema.
func parseElement(value, schema any) error {
	if schema == nil {
		return nil
	}

	sv := reflect.ValueOf(schema)
	if !sv.IsValid() || sv.IsNil() {
		return nil
	}

	m := sv.MethodByName("Parse")
	if !m.IsValid() || m.Type().NumIn() < 1 {
		return nil
	}

	results := m.Call([]reflect.Value{reflect.ValueOf(value)})
	if len(results) >= 2 {
		if errVal := results[1].Interface(); errVal != nil {
			if err, ok := errVal.(error); ok {
				return err
			}
		}
	}

	return nil
}

// newZodArrayFromDef constructs a new ZodArray from a definition.
func newZodArrayFromDef[T any, R any](def *ZodArrayDef) *ZodArray[T, R] {
	internals := &ZodArrayInternals{
		ZodTypeInternals: core.ZodTypeInternals{
			Type:   def.Type,
			Checks: def.Checks,
			Coerce: def.Coerce,
			Bag:    make(map[string]any),
		},
		Def:   def,
		Items: def.Items,
		Rest:  def.Rest,
	}

	internals.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any] {
		arrayDef := &ZodArrayDef{
			ZodTypeDef: *newDef,
			Items:      def.Items,
		}
		return any(newZodArrayFromDef[T, R](arrayDef)).(core.ZodType[any])
	}

	if def.Error != nil {
		internals.Error = def.Error
	}

	return &ZodArray[T, R]{internals: internals}
}

// Array creates a tuple schema with fixed elements.
//
//	Array()                               - empty tuple
//	Array([]any{String(), Int()})         - fixed length tuple
//	Array([]any{String(), Int()}, Bool()) - tuple with rest parameter
func Array(args ...any) *ZodArray[[]any, []any] {
	if len(args) == 0 {
		return ArrayTyped[[]any, []any](nil)
	}
	if items, ok := args[0].([]any); ok {
		return ArrayTyped[[]any, []any](items, args[1:]...)
	}
	// Treat single non-[]any argument as single-element array.
	return ArrayTyped[[]any, []any]([]any{args[0]}, args[1:]...)
}

// ArrayPtr creates a pointer-capable tuple schema.
func ArrayPtr(args ...any) *ZodArray[[]any, *[]any] {
	if len(args) == 0 {
		return ArrayTyped[[]any, *[]any](nil)
	}
	if items, ok := args[0].([]any); ok {
		return ArrayTyped[[]any, *[]any](items, args[1:]...)
	}
	return ArrayTyped[[]any, *[]any]([]any{args[0]}, args[1:]...)
}

// ArrayTyped is the generic constructor for tuple schemas.
func ArrayTyped[T any, R any](items []any, args ...any) *ZodArray[T, R] {
	if items == nil {
		items = []any{}
	}

	var rest any
	var param any

	for _, arg := range args {
		if v, ok := arg.(core.SchemaParams); ok {
			param = v
		} else if rest == nil {
			rest = arg
		}
	}

	sp := utils.NormalizeParams(param)

	def := &ZodArrayDef{
		ZodTypeDef: core.ZodTypeDef{
			Type: core.ZodTypeArray,
		},
		Items: items,
		Rest:  rest,
	}

	if sp != nil {
		utils.ApplySchemaParams(&def.ZodTypeDef, sp)
	}

	return newZodArrayFromDef[T, R](def)
}

// Check adds a custom validation function that can report multiple issues.
func (z *ZodArray[T, R]) Check(fn func(R, *core.ParsePayload), params ...any) *ZodArray[T, R] {
	wrapper := func(p *core.ParsePayload) {
		if val, ok := p.Value().(R); ok {
			fn(val, p)
			return
		}

		// Pointer type adaptation: wrap value type in pointer for R = *T.
		rt := reflect.TypeFor[R]()
		if rt.Kind() == reflect.Pointer {
			elem := rt.Elem()
			rv := reflect.ValueOf(p.Value())
			if rv.IsValid() && rv.Type() == elem {
				ptr := reflect.New(elem)
				ptr.Elem().Set(rv)
				if v, ok := ptr.Interface().(R); ok {
					fn(v, p)
				}
			}
		}
	}

	in := z.internals.Clone()
	in.AddCheck(checks.NewCustom[any](wrapper, utils.NormalizeCustomParams(params...)))
	return z.withInternals(in)
}

// extractForEngine extracts []any from input for engine.ParseComplex.
func (z *ZodArray[T, R]) extractForEngine(input any) ([]any, bool) {
	result, err := z.extract(input)
	if err != nil {
		return nil, false
	}
	return result, true
}

// extractPtrForEngine extracts pointer to []any from input.
func (z *ZodArray[T, R]) extractPtrForEngine(input any) (*[]any, bool) {
	if ptr, ok := input.(*[]any); ok {
		return ptr, true
	}

	result, err := z.extract(input)
	if err != nil {
		return nil, false
	}
	return &result, true
}

// validateForEngine validates []any for engine.ParseComplex.
func (z *ZodArray[T, R]) validateForEngine(value []any, chks []core.ZodCheck, ctx *core.ParseContext) ([]any, error) {
	return z.validate(value, chks, ctx)
}

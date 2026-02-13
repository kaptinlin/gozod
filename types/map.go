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
	"github.com/kaptinlin/gozod/pkg/mapx"
	"github.com/kaptinlin/gozod/pkg/reflectx"
)

// ZodMapDef is the configuration for a map schema.
type ZodMapDef struct {
	core.ZodTypeDef
	KeyType   any
	ValueType any
}

// ZodMapInternals holds the internal state of a map validator.
type ZodMapInternals struct {
	core.ZodTypeInternals
	Def       *ZodMapDef
	KeyType   any
	ValueType any
}

// ZodMap is a type-safe map validation schema.
// T is the base type, R is the constraint type (value or pointer).
type ZodMap[T any, R any] struct {
	internals *ZodMapInternals
}

// Map creates a map schema with key and value validation returning a value constraint.
func Map(keySchema, valueSchema any, paramArgs ...any) *ZodMap[map[any]any, map[any]any] {
	return MapTyped[map[any]any, map[any]any](keySchema, valueSchema, paramArgs...)
}

// MapPtr creates a map schema returning a pointer constraint.
func MapPtr(keySchema, valueSchema any, paramArgs ...any) *ZodMap[map[any]any, *map[any]any] {
	return MapTyped[map[any]any, *map[any]any](keySchema, valueSchema, paramArgs...)
}

// MapTyped creates a typed map schema with explicit generic constraints.
func MapTyped[T any, R any](keySchema, valueSchema any, paramArgs ...any) *ZodMap[T, R] {
	p := utils.FirstParam(paramArgs...)
	sp := utils.NormalizeParams(p)

	def := &ZodMapDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeMap,
			Checks: []core.ZodCheck{},
		},
		KeyType:   keySchema,
		ValueType: valueSchema,
	}
	if sp != nil {
		utils.ApplySchemaParams(&def.ZodTypeDef, sp)
	}

	schema := newZodMapFromDef[T, R](def)

	// Register a pass-through check so the engine invokes validateMap for key/value validation.
	if keySchema != nil || valueSchema != nil {
		schema.internals.AddCheck(checks.NewCustom[any](func(v any) bool { return true }, core.SchemaParams{}))
	}
	return schema
}

// Internals returns the internal state of the schema.
func (z *ZodMap[T, R]) Internals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// IsOptional reports whether this schema accepts undefined/missing values.
func (z *ZodMap[T, R]) IsOptional() bool {
	return z.internals.IsOptional()
}

// IsNilable reports whether this schema accepts nil values.
func (z *ZodMap[T, R]) IsNilable() bool {
	return z.internals.IsNilable()
}

// KeyType returns the key schema for this map.
func (z *ZodMap[T, R]) KeyType() any {
	return z.internals.KeyType
}

// ValueType returns the value schema for this map.
func (z *ZodMap[T, R]) ValueType() any {
	return z.internals.ValueType
}

// CloneFrom copies configuration from another schema of the same type.
func (z *ZodMap[T, R]) CloneFrom(source any) {
	if src, ok := source.(*ZodMap[T, R]); ok {
		z.internals = src.internals
	}
}

// Parse validates input using map-specific parsing logic.
func (z *ZodMap[T, R]) Parse(input any, ctx ...*core.ParseContext) (R, error) {
	var zero R

	result, err := engine.ParseComplex[map[any]any](
		input,
		&z.internals.ZodTypeInternals,
		core.ZodTypeMap,
		z.extractForEngine,
		z.extractPtrForEngine,
		z.validateForEngine,
		ctx...,
	)
	if err != nil {
		return zero, err
	}

	switch v := result.(type) {
	case map[any]any:
		return convertFromGeneric[T, R](v), nil
	case *map[any]any:
		if v == nil {
			return zero, nil
		}
		return convertFromGeneric[T, R](*v), nil
	case nil:
		return zero, nil
	default:
		pc := core.NewParseContext()
		if len(ctx) > 0 && ctx[0] != nil {
			pc = ctx[0]
		}
		return zero, issues.CreateTypeConversionError(
			fmt.Sprintf("%T", result), "map", input, pc,
		)
	}
}

// MustParse validates input and panics on error.
func (z *ZodMap[T, R]) MustParse(input any, ctx ...*core.ParseContext) R {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// StrictParse validates input with compile-time type safety.
func (z *ZodMap[T, R]) StrictParse(input T, ctx ...*core.ParseContext) (R, error) {
	cv, ok := toConstraintValue[T, R](input)
	if !ok {
		var zero R
		if len(ctx) == 0 {
			ctx = []*core.ParseContext{core.NewParseContext()}
		}
		return zero, issues.CreateTypeConversionError(
			fmt.Sprintf("%T", input), "map constraint type", any(input), ctx[0],
		)
	}

	return engine.ParseComplexStrict[map[any]any, R](
		cv,
		&z.internals.ZodTypeInternals,
		core.ZodTypeMap,
		z.extractForEngine,
		z.extractPtrForEngine,
		z.validateForEngine,
		ctx...,
	)
}

// MustStrictParse validates input with type safety and panics on error.
func (z *ZodMap[T, R]) MustStrictParse(input T, ctx ...*core.ParseContext) R {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// ParseAny validates input and returns an untyped result.
func (z *ZodMap[T, R]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// Optional returns a schema that accepts nil, with constraint type *T.
func (z *ZodMap[T, R]) Optional() *ZodMap[T, *T] {
	in := z.internals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// Nilable returns a schema that accepts nil, with constraint type *T.
func (z *ZodMap[T, R]) Nilable() *ZodMap[T, *T] {
	in := z.internals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish combines optional and nilable modifiers.
func (z *ZodMap[T, R]) Nullish() *ZodMap[T, *T] {
	in := z.internals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// NonOptional removes the Optional flag and returns base constraint type.
func (z *ZodMap[T, R]) NonOptional() *ZodMap[T, T] {
	in := z.internals.Clone()
	in.SetOptional(false)
	in.SetNonOptional(true)
	return &ZodMap[T, T]{internals: &ZodMapInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
		KeyType:          z.internals.KeyType,
		ValueType:        z.internals.ValueType,
	}}
}

// Default sets a value returned when input is nil, bypassing validation.
func (z *ZodMap[T, R]) Default(v T) *ZodMap[T, R] {
	in := z.internals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

// DefaultFunc sets a factory for the default value when input is nil.
func (z *ZodMap[T, R]) DefaultFunc(fn func() T) *ZodMap[T, R] {
	in := z.internals.Clone()
	in.SetDefaultFunc(func() any { return fn() })
	return z.withInternals(in)
}

// Prefault sets a value that goes through full validation when input is nil.
func (z *ZodMap[T, R]) Prefault(v T) *ZodMap[T, R] {
	in := z.internals.Clone()
	in.SetPrefaultValue(v)
	return z.withInternals(in)
}

// PrefaultFunc sets a factory for the prefault value through full validation.
func (z *ZodMap[T, R]) PrefaultFunc(fn func() T) *ZodMap[T, R] {
	in := z.internals.Clone()
	in.SetPrefaultFunc(func() any { return fn() })
	return z.withInternals(in)
}

// Meta stores metadata for this map schema.
func (z *ZodMap[T, R]) Meta(meta core.GlobalMeta) *ZodMap[T, R] {
	core.GlobalRegistry.Add(z, meta)
	return z
}

// Describe registers a description in the global registry.
func (z *ZodMap[T, R]) Describe(description string) *ZodMap[T, R] {
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

// Min sets the minimum number of entries.
func (z *ZodMap[T, R]) Min(minLen int, params ...any) *ZodMap[T, R] {
	return z.withCheck(checks.MinSize(minLen, params...))
}

// Max sets the maximum number of entries.
func (z *ZodMap[T, R]) Max(maxLen int, params ...any) *ZodMap[T, R] {
	return z.withCheck(checks.MaxSize(maxLen, params...))
}

// Length sets the exact number of entries required.
func (z *ZodMap[T, R]) Length(exactLen int, params ...any) *ZodMap[T, R] {
	return z.withCheck(checks.Size(exactLen, params...))
}

// NonEmpty ensures the map has at least one entry.
// Equivalent to Min(1).
func (z *ZodMap[T, R]) NonEmpty(params ...any) *ZodMap[T, R] {
	return z.Min(1, params...)
}

// Transform applies a transformation function to the parsed map value.
func (z *ZodMap[T, R]) Transform(fn func(T, *core.RefinementContext) (any, error)) *core.ZodTransform[R, any] {
	wrap := func(input R, ctx *core.RefinementContext) (any, error) {
		return fn(extractValue[T, R](input), ctx)
	}
	return core.NewZodTransform[R, any](z, wrap)
}

// Overwrite transforms the value while preserving the original type.
func (z *ZodMap[T, R]) Overwrite(transform func(R) R, params ...any) *ZodMap[T, R] {
	fn := func(input any) any {
		converted, ok := toMapType[T, R](input)
		if !ok {
			return input
		}
		return transform(converted)
	}
	return z.withCheck(checks.NewZodCheckOverwrite(fn, params...))
}

// Pipe creates a pipeline that passes the parsed result to a target schema.
func (z *ZodMap[T, R]) Pipe(target core.ZodType[any]) *core.ZodPipe[R, any] {
	wrap := func(input R, ctx *core.ParseContext) (any, error) {
		return target.Parse(extractValue[T, R](input), ctx)
	}
	return core.NewZodPipe[R, any](z, target, wrap)
}

// Refine adds a typed custom validation function.
func (z *ZodMap[T, R]) Refine(fn func(R) bool, params ...any) *ZodMap[T, R] {
	wrap := func(v any) bool {
		var zero R
		switch any(zero).(type) {
		case *map[any]any:
			if v == nil {
				return true
			}
			if cv, ok := toConstraintValue[T, R](v); ok {
				return fn(cv)
			}
			return false
		default:
			if cv, ok := toConstraintValue[T, R](v); ok {
				return fn(cv)
			}
			return false
		}
	}
	return z.withCheck(checks.NewCustom[any](wrap, utils.NormalizeCustomParams(params...)))
}

// RefineAny adds a custom validation function accepting any input.
func (z *ZodMap[T, R]) RefineAny(fn func(any) bool, params ...any) *ZodMap[T, R] {
	return z.withCheck(checks.NewCustom[any](fn, utils.NormalizeCustomParams(params...)))
}

// Check adds a custom validation function that can report multiple issues.
func (z *ZodMap[T, R]) Check(fn func(value R, payload *core.ParsePayload), params ...any) *ZodMap[T, R] {
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
func (z *ZodMap[T, R]) With(fn func(value R, payload *core.ParsePayload), params ...any) *ZodMap[T, R] {
	return z.Check(fn, params...)
}

func (z *ZodMap[T, R]) withCheck(c core.ZodCheck) *ZodMap[T, R] {
	in := z.internals.Clone()
	in.AddCheck(c)
	return z.withInternals(in)
}

func (z *ZodMap[T, R]) withPtrInternals(in *core.ZodTypeInternals) *ZodMap[T, *T] {
	return &ZodMap[T, *T]{internals: &ZodMapInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
		KeyType:          z.internals.KeyType,
		ValueType:        z.internals.ValueType,
	}}
}

func (z *ZodMap[T, R]) withInternals(in *core.ZodTypeInternals) *ZodMap[T, R] {
	return &ZodMap[T, R]{internals: &ZodMapInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
		KeyType:          z.internals.KeyType,
		ValueType:        z.internals.ValueType,
	}}
}

func (z *ZodMap[T, R]) extractType(value any, ctx *core.ParseContext) (map[any]any, error) {
	if m, ok := value.(map[any]any); ok {
		return m, nil
	}
	if ptr, ok := value.(*map[any]any); ok {
		if ptr != nil {
			return *ptr, nil
		}
		return nil, issues.CreateNonOptionalError(ctx)
	}
	if reflectx.IsMap(value) {
		if converted, err := mapx.ToGeneric(value); err == nil && converted != nil {
			return converted, nil
		}
	}
	return nil, issues.CreateInvalidTypeError(core.ZodTypeMap, value, ctx)
}

func (z *ZodMap[T, R]) validateMap(value map[any]any, chks []core.ZodCheck, ctx *core.ParseContext) (map[any]any, error) {
	transformed, err := engine.ApplyChecks[any](value, chks, ctx)
	if err != nil {
		return nil, err
	}

	switch v := transformed.(type) {
	case map[any]any:
		value = v
	case *map[any]any:
		if v != nil {
			value = *v
		}
	default:
		// Keep original value if transformation returned unexpected type.
	}

	var collected []core.ZodRawIssue

	for key, val := range value {
		if z.internals.KeyType != nil {
			collected = z.collectErrors(key, z.internals.KeyType, key, ctx, collected)
		}
		if z.internals.ValueType != nil {
			collected = z.collectErrors(val, z.internals.ValueType, key, ctx, collected)
		}
	}

	if len(collected) > 0 {
		return nil, issues.CreateArrayValidationIssues(collected)
	}
	return value, nil
}

func (z *ZodMap[T, R]) collectErrors(value, schema, pathKey any, ctx *core.ParseContext, dst []core.ZodRawIssue) []core.ZodRawIssue {
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

func (z *ZodMap[T, R]) validateDirect(value, schema any, ctx *core.ParseContext) error {
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

func (z *ZodMap[T, R]) extractForEngine(input any) (map[any]any, bool) {
	result, err := z.extractType(input, core.NewParseContext())
	return result, err == nil
}

func (z *ZodMap[T, R]) extractPtrForEngine(input any) (*map[any]any, bool) {
	if ptr, ok := input.(*map[any]any); ok {
		return ptr, true
	}
	result, err := z.extractType(input, core.NewParseContext())
	if err != nil {
		return nil, false
	}
	return &result, true
}

func (z *ZodMap[T, R]) validateForEngine(value map[any]any, chks []core.ZodCheck, ctx *core.ParseContext) (map[any]any, error) {
	return z.validateMap(value, chks, ctx)
}

func newZodMapFromDef[T any, R any](def *ZodMapDef) *ZodMap[T, R] {
	in := &ZodMapInternals{
		ZodTypeInternals: core.ZodTypeInternals{
			Type:   def.Type,
			Checks: def.Checks,
			Coerce: def.Coerce,
			Bag:    make(map[string]any),
		},
		Def:       def,
		KeyType:   def.KeyType,
		ValueType: def.ValueType,
	}

	in.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any] {
		md := &ZodMapDef{
			ZodTypeDef: *newDef,
			KeyType:    def.KeyType,
			ValueType:  def.ValueType,
		}
		return any(newZodMapFromDef[T, R](md)).(core.ZodType[any])
	}

	if def.Error != nil {
		in.Error = def.Error
	}
	return &ZodMap[T, R]{internals: in}
}

func convertFromGeneric[T any, R any](m map[any]any) R {
	var base T
	switch any(base).(type) {
	case map[string]int:
		converted := make(map[string]int, len(m))
		for k, v := range m {
			if sk, ok := k.(string); ok {
				if iv, ok := v.(int); ok {
					converted[sk] = iv
				}
			}
		}
		base = any(converted).(T)
	case map[string]any:
		converted := make(map[string]any, len(m))
		for k, v := range m {
			if sk, ok := k.(string); ok {
				converted[sk] = v
			}
		}
		base = any(converted).(T)
	default:
		base = any(m).(T)
	}
	return toConstraintType[T, R](base)
}

func toConstraintType[T any, R any](value T) R {
	var zero R
	switch any(zero).(type) {
	case *map[any]any:
		if mv, ok := any(value).(map[any]any); ok {
			cp := mv
			return any(&cp).(R)
		}
		return any((*map[any]any)(nil)).(R)
	default:
		return any(value).(R)
	}
}

func extractValue[T any, R any](value R) T {
	if v, ok := any(value).(T); ok {
		return v
	}
	if rv := reflect.ValueOf(value); rv.Kind() == reflect.Pointer && !rv.IsNil() {
		if v, ok := rv.Elem().Interface().(T); ok {
			return v
		}
	}
	var zero T
	return zero
}

func toMapType[T any, R any](v any) (R, bool) {
	var zero R

	if v == nil {
		if reflect.TypeFor[R]().Kind() == reflect.Pointer {
			return zero, true
		}
		return zero, false
	}

	var m map[any]any
	var ok bool

	switch val := v.(type) {
	case map[any]any:
		m, ok = val, true
	case *map[any]any:
		if val != nil {
			m, ok = *val, true
		}
	case map[string]any:
		m = make(map[any]any, len(val))
		for k, v := range val {
			m[k] = v
		}
		ok = true
	case *map[string]any:
		if val != nil {
			m = make(map[any]any, len(*val))
			for k, v := range *val {
				m[k] = v
			}
			ok = true
		}
	default:
		return zero, false
	}

	if !ok {
		return zero, false
	}

	rt := reflect.TypeFor[R]()
	if rt.Kind() == reflect.Pointer {
		if converted, ok := any(&m).(R); ok {
			return converted, true
		}
	} else {
		if converted, ok := any(m).(R); ok {
			return converted, true
		}
	}

	return zero, false
}

func toConstraintValue[T any, R any](value any) (R, bool) {
	var zero R

	if r, ok := any(value).(R); ok { //nolint:unconvert // Required for generic type constraint conversion
		return r, true
	}

	if _, ok := any(zero).(*map[any]any); ok {
		if mv, ok := value.(map[any]any); ok {
			cp := mv
			return any(&cp).(R), true
		}
	}

	return zero, false
}

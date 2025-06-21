package types

import (
	"errors"
	"fmt"
	"math/big"
	"regexp"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/pkg/coerce"
)

// Error definitions for bigint transformations
var (
	ErrTransformNilBigInt = errors.New("cannot transform nil bigint")
)

//////////////////////////
// CORE TYPE DEFINITIONS
//////////////////////////

// ZodBigIntDef defines the configuration for big.Int validation
type ZodBigIntDef struct {
	core.ZodTypeDef
	Type string // "bigint"
}

// ZodBigIntInternals contains big.Int validator internal state
type ZodBigIntInternals struct {
	core.ZodTypeInternals
	Def     *ZodBigIntDef  // Schema definition
	Pattern *regexp.Regexp // BigInt pattern (if any)
	Bag     map[string]any // Additional metadata (minimum, maximum, coerce flag, etc.)
}

// ZodBigInt represents a big.Int validation schema
type ZodBigInt struct {
	internals *ZodBigIntInternals
}

// GetInternals returns the internal state of the schema
func (z *ZodBigInt) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// GetZod returns the bigint-specific internals for framework usage
func (z *ZodBigInt) GetZod() *ZodBigIntInternals {
	return z.internals
}

// CloneFrom implements Cloneable interface for type-specific state copying
func (z *ZodBigInt) CloneFrom(source any) {
	if src, ok := source.(interface{ GetZod() *ZodBigIntInternals }); ok {
		srcState := src.GetZod()
		tgtState := z.GetZod()

		// Copy Bag state (includes coercion flags, etc.)
		if len(srcState.Bag) > 0 {
			if tgtState.Bag == nil {
				tgtState.Bag = make(map[string]any)
			}
			for key, value := range srcState.Bag {
				tgtState.Bag[key] = value
			}
		}

		// Copy Pattern state
		if srcState.Pattern != nil {
			tgtState.Pattern = srcState.Pattern
		}
	}
}

// Coerce attempts to coerce input to big.Int type
func (z *ZodBigInt) Coerce(input any) (any, bool) {
	coerceFunc := createBigIntCoerceFunc()
	return coerceFunc(input)
}

// Parse validates input with smart type inference
func (z *ZodBigInt) Parse(input any, ctx ...*core.ParseContext) (any, error) {
	parseCtx := (*core.ParseContext)(nil)
	if len(ctx) > 0 {
		parseCtx = ctx[0]
	}

	typeName := "bigint"
	coerceFunc := createBigIntCoerceFunc()

	return engine.ParseType[*big.Int](
		input,
		&z.internals.ZodTypeInternals,
		typeName,
		func(v any) (*big.Int, bool) {
			if bi, ok := v.(*big.Int); ok {
				return bi, true
			}
			return nil, false
		},
		func(v any) (**big.Int, bool) {
			if bi, ok := v.(**big.Int); ok {
				return bi, true
			}
			return nil, false
		},
		validateBigInt,
		coerceFunc,
		parseCtx,
	)
}

// MustParse validates the input value and panics on failure
func (z *ZodBigInt) MustParse(input any, ctx ...*core.ParseContext) any {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

//////////////////////////
// VALIDATION METHODS
//////////////////////////

// Min adds minimum value validation
func (z *ZodBigInt) Min(minimum *big.Int, params ...any) *ZodBigInt {
	check := checks.Gte(minimum, params...)
	result := engine.AddCheck(z, check)
	return result.(*ZodBigInt)
}

// Max adds maximum value validation
func (z *ZodBigInt) Max(maximum *big.Int, params ...any) *ZodBigInt {
	check := checks.Lte(maximum, params...)
	result := engine.AddCheck(z, check)
	return result.(*ZodBigInt)
}

// Gt adds greater than validation (exclusive)
func (z *ZodBigInt) Gt(value *big.Int, params ...any) *ZodBigInt {
	check := checks.Gt(value, params...)
	result := engine.AddCheck(z, check)
	return result.(*ZodBigInt)
}

// Gte adds greater than or equal validation (inclusive)
func (z *ZodBigInt) Gte(value *big.Int, params ...any) *ZodBigInt {
	check := checks.Gte(value, params...)
	result := engine.AddCheck(z, check)
	return result.(*ZodBigInt)
}

// Lt adds less than validation (exclusive)
func (z *ZodBigInt) Lt(value *big.Int, params ...any) *ZodBigInt {
	check := checks.Lt(value, params...)
	result := engine.AddCheck(z, check)
	return result.(*ZodBigInt)
}

// Lte adds less than or equal validation (inclusive)
func (z *ZodBigInt) Lte(value *big.Int, params ...any) *ZodBigInt {
	check := checks.Lte(value, params...)
	result := engine.AddCheck(z, check)
	return result.(*ZodBigInt)
}

// Positive adds positive number validation (> 0)
func (z *ZodBigInt) Positive(params ...any) *ZodBigInt {
	return z.Gt(big.NewInt(0), params...)
}

// Negative adds negative number validation (< 0)
func (z *ZodBigInt) Negative(params ...any) *ZodBigInt {
	return z.Lt(big.NewInt(0), params...)
}

// NonNegative adds non-negative number validation (>= 0)
func (z *ZodBigInt) NonNegative(params ...any) *ZodBigInt {
	return z.Gte(big.NewInt(0), params...)
}

// NonPositive adds non-positive number validation (<= 0)
func (z *ZodBigInt) NonPositive(params ...any) *ZodBigInt {
	return z.Lte(big.NewInt(0), params...)
}

// MultipleOf adds multiple of validation
func (z *ZodBigInt) MultipleOf(value *big.Int, params ...any) *ZodBigInt {
	check := checks.MultipleOf(value, params...)
	result := engine.AddCheck(z, check)
	return result.(*ZodBigInt)
}

// Refine adds a type-safe refinement check for big.Int types
func (z *ZodBigInt) Refine(fn func(*big.Int) bool, params ...any) *ZodBigInt {
	result := z.RefineAny(func(v any) bool {
		val, isNil, err := extractBigIntValue(v)
		if err != nil {
			return false
		}
		if isNil {
			return true // Let upper logic decide
		}
		return fn(val)
	}, params...)
	return result.(*ZodBigInt)
}

// RefineAny adds custom validation to the big.Int schema
func (z *ZodBigInt) RefineAny(fn func(any) bool, params ...any) core.ZodType[any, any] {
	check := checks.NewCustom[any](fn, params...)
	return engine.AddCheck(z, check)
}

// Check adds modern validation using direct payload access
func (z *ZodBigInt) Check(fn core.CheckFn) core.ZodType[any, any] {
	check := checks.NewCustom[any](func(v any) bool {
		payload := &core.ParsePayload{
			Value:  v,
			Issues: make([]core.ZodRawIssue, 0),
			Path:   make([]any, 0),
		}
		fn(payload)
		return len(payload.Issues) == 0
	}, core.SchemaParams{})
	return engine.AddCheck(z, check)
}

//////////////////////////
// TRANSFORM METHODS
//////////////////////////

// Transform creates a transformation pipeline for big.Int types
func (z *ZodBigInt) Transform(fn func(*big.Int, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	return z.TransformAny(func(input any, ctx *core.RefinementContext) (any, error) {
		// Use extractBigIntValue helper function that already exists
		val, isNil, err := extractBigIntValue(input)
		if err != nil {
			return nil, err
		}
		if isNil {
			return nil, ErrTransformNilBigInt
		}
		return fn(val, ctx)
	})
}

// TransformAny creates a transformation pipeline that accepts any input type
func (z *ZodBigInt) TransformAny(fn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	transform := Transform[any, any](fn)

	return &ZodPipe[any, any]{
		in:  any(z).(core.ZodType[any, any]),
		out: any(transform).(core.ZodType[any, any]),
		def: core.ZodTypeDef{Type: "pipe"},
	}
}

// Pipe creates a validation pipeline
func (z *ZodBigInt) Pipe(out core.ZodType[any, any]) core.ZodType[any, any] {
	return &ZodPipe[any, any]{
		in:  any(z).(core.ZodType[any, any]),
		out: out,
		def: core.ZodTypeDef{Type: "pipe"},
	}
}

//////////////////////////
// MODIFIER METHODS
//////////////////////////

// Optional makes the big.Int schema optional
func (z *ZodBigInt) Optional() core.ZodType[any, any] {
	return Optional(z)
}

// Nilable creates a new big.Int schema that accepts nil values
func (z *ZodBigInt) Nilable() core.ZodType[any, any] {
	return Nilable(z)
}

// Nullish makes the big.Int schema both optional and nullable
func (z *ZodBigInt) Nullish() core.ZodType[any, any] {
	return Nullish(z)
}

// Unwrap returns the inner type (for basic types, returns self)
func (z *ZodBigInt) Unwrap() core.ZodType[any, any] {
	return any(z).(core.ZodType[any, any])
}

//////////////////////////
// VALIDATION FUNCTIONS
//////////////////////////

// validateBigInt validates big integer values with checks
func validateBigInt(value *big.Int, checks []core.ZodCheck, ctx *core.ParseContext) error {
	if len(checks) > 0 {
		payload := &core.ParsePayload{
			Value:  value,
			Issues: make([]core.ZodRawIssue, 0),
		}
		engine.RunChecksOnValue(value, checks, payload, ctx)
		if len(payload.Issues) > 0 {
			return issues.NewZodError(issues.ConvertRawIssuesToIssues(payload.Issues, ctx))
		}
	}
	return nil
}

//////////////////////////
// WRAPPER TYPES
//////////////////////////

////////////////////////////
////   DEFAULT METHODS   ////
////////////////////////////

// ZodBigIntDefault is a Default wrapper for bigint type
// Provides perfect type safety and chainable method support
type ZodBigIntDefault struct {
	*ZodDefault[*ZodBigInt] // Embed concrete pointer to enable method promotion
}

// Parse implements ZodType interface
func (s ZodBigIntDefault) Parse(input any, ctx ...*core.ParseContext) (any, error) {
	return s.ZodDefault.Parse(input, ctx...)
}

// Default adds a default value to the bigint schema, returns ZodBigIntDefault support chain call
func (z *ZodBigInt) Default(value *big.Int) ZodBigIntDefault {
	return ZodBigIntDefault{
		&ZodDefault[*ZodBigInt]{
			innerType:    z,
			defaultValue: value,
			isFunction:   false,
		},
	}
}

// DefaultFunc adds a default function to the bigint schema, returns ZodBigIntDefault support chain call
func (z *ZodBigInt) DefaultFunc(fn func() *big.Int) ZodBigIntDefault {
	genericFn := func() any { return fn() }
	return ZodBigIntDefault{
		&ZodDefault[*ZodBigInt]{
			innerType:   z,
			defaultFunc: genericFn,
			isFunction:  true,
		},
	}
}

// ZodBigIntDefault chainable validation methods

func (s ZodBigIntDefault) Min(minimum *big.Int, params ...any) ZodBigIntDefault {
	newInner := s.innerType.Min(minimum, params...)
	return ZodBigIntDefault{
		&ZodDefault[*ZodBigInt]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodBigIntDefault) Max(maximum *big.Int, params ...any) ZodBigIntDefault {
	newInner := s.innerType.Max(maximum, params...)
	return ZodBigIntDefault{
		&ZodDefault[*ZodBigInt]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodBigIntDefault) Gt(value *big.Int, params ...any) ZodBigIntDefault {
	newInner := s.innerType.Gt(value, params...)
	return ZodBigIntDefault{
		&ZodDefault[*ZodBigInt]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodBigIntDefault) Gte(value *big.Int, params ...any) ZodBigIntDefault {
	newInner := s.innerType.Gte(value, params...)
	return ZodBigIntDefault{
		&ZodDefault[*ZodBigInt]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodBigIntDefault) Lt(value *big.Int, params ...any) ZodBigIntDefault {
	newInner := s.innerType.Lt(value, params...)
	return ZodBigIntDefault{
		&ZodDefault[*ZodBigInt]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodBigIntDefault) Lte(value *big.Int, params ...any) ZodBigIntDefault {
	newInner := s.innerType.Lte(value, params...)
	return ZodBigIntDefault{
		&ZodDefault[*ZodBigInt]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodBigIntDefault) Positive(params ...any) ZodBigIntDefault {
	newInner := s.innerType.Positive(params...)
	return ZodBigIntDefault{
		&ZodDefault[*ZodBigInt]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodBigIntDefault) Negative(params ...any) ZodBigIntDefault {
	newInner := s.innerType.Negative(params...)
	return ZodBigIntDefault{
		&ZodDefault[*ZodBigInt]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodBigIntDefault) NonNegative(params ...any) ZodBigIntDefault {
	newInner := s.innerType.NonNegative(params...)
	return ZodBigIntDefault{
		&ZodDefault[*ZodBigInt]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodBigIntDefault) NonPositive(params ...any) ZodBigIntDefault {
	newInner := s.innerType.NonPositive(params...)
	return ZodBigIntDefault{
		&ZodDefault[*ZodBigInt]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodBigIntDefault) MultipleOf(value *big.Int, params ...any) ZodBigIntDefault {
	newInner := s.innerType.MultipleOf(value, params...)
	return ZodBigIntDefault{
		&ZodDefault[*ZodBigInt]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodBigIntDefault) Refine(fn func(*big.Int) bool, params ...any) ZodBigIntDefault {
	newInner := s.innerType.Refine(fn, params...)
	return ZodBigIntDefault{
		&ZodDefault[*ZodBigInt]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodBigIntDefault) Transform(fn func(*big.Int, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	return s.TransformAny(func(input any, ctx *core.RefinementContext) (any, error) {
		// Use extractBigIntValue helper function that already exists
		val, isNil, err := extractBigIntValue(input)
		if err != nil {
			return nil, err
		}
		if isNil {
			return nil, fmt.Errorf("cannot transform nil value")
		}
		return fn(val, ctx)
	})
}

func (s ZodBigIntDefault) Optional() core.ZodType[any, any] {
	return Optional(any(s).(core.ZodType[any, any]))
}

func (s ZodBigIntDefault) Nilable() core.ZodType[any, any] {
	return Nilable(any(s).(core.ZodType[any, any]))
}

////////////////////////////
////   PREFAULT METHODS   ////
////////////////////////////

// ZodBigIntPrefault is a Prefault wrapper for bigint type
// Provides perfect type safety and chainable method support
type ZodBigIntPrefault struct {
	*ZodPrefault[*ZodBigInt] // Embed concrete pointer to enable method promotion
}

// Parse implements ZodType interface
func (s ZodBigIntPrefault) Parse(input any, ctx ...*core.ParseContext) (any, error) {
	return s.ZodPrefault.Parse(input, ctx...)
}

// Prefault adds a prefault value to the bigint schema, returns ZodBigIntPrefault support chain call
func (z *ZodBigInt) Prefault(value *big.Int) ZodBigIntPrefault {
	baseInternals := z.GetInternals()
	internals := &core.ZodTypeInternals{
		Version:     core.Version,
		Type:        core.ZodTypePrefault,
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

	return ZodBigIntPrefault{
		&ZodPrefault[*ZodBigInt]{
			internals:     internals,
			innerType:     z,
			prefaultValue: value,
			prefaultFunc:  nil,
			isFunction:    false,
		},
	}
}

// PrefaultFunc adds a prefault function to the bigint schema, returns ZodBigIntPrefault support chain call
func (z *ZodBigInt) PrefaultFunc(fn func() *big.Int) ZodBigIntPrefault {
	genericFn := func() any { return fn() }

	baseInternals := z.GetInternals()
	internals := &core.ZodTypeInternals{
		Version:     core.Version,
		Type:        core.ZodTypePrefault,
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

	return ZodBigIntPrefault{
		&ZodPrefault[*ZodBigInt]{
			internals:     internals,
			innerType:     z,
			prefaultValue: nil,
			prefaultFunc:  genericFn,
			isFunction:    true,
		},
	}
}

// ZodBigIntPrefault chainable validation methods

func (s ZodBigIntPrefault) Refine(fn func(*big.Int) bool, params ...any) ZodBigIntPrefault {
	newInner := s.innerType.Refine(fn, params...)

	// Construct new internals
	baseInternals := newInner.GetInternals()
	internals := &core.ZodTypeInternals{
		Version:     core.Version,
		Type:        core.ZodTypePrefault,
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

	return ZodBigIntPrefault{
		&ZodPrefault[*ZodBigInt]{
			internals:     internals,
			innerType:     newInner,
			prefaultValue: s.prefaultValue,
			prefaultFunc:  s.prefaultFunc,
			isFunction:    s.isFunction,
		},
	}
}

func (s ZodBigIntPrefault) Transform(fn func(*big.Int, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	return s.TransformAny(func(input any, ctx *core.RefinementContext) (any, error) {
		// Use extractBigIntValue helper function that already exists
		val, isNil, err := extractBigIntValue(input)
		if err != nil {
			return nil, err
		}
		if isNil {
			return nil, fmt.Errorf("cannot transform nil value")
		}
		return fn(val, ctx)
	})
}

func (s ZodBigIntPrefault) Optional() core.ZodType[any, any] {
	return Optional(any(s).(core.ZodType[any, any]))
}

func (s ZodBigIntPrefault) Nilable() core.ZodType[any, any] {
	return Nilable(any(s).(core.ZodType[any, any]))
}

//////////////////////////
// CONSTRUCTOR FUNCTIONS
//////////////////////////

// ZodInt64BigInt wraps ZodBigInt for int64-compatible validation
type ZodInt64BigInt struct {
	*ZodBigInt
}

// ZodUint64BigInt wraps ZodBigInt for uint64-compatible validation
type ZodUint64BigInt struct {
	*ZodBigInt
}

// Parse validates and converts to int64
func (z *ZodInt64BigInt) Parse(input any, ctx ...*core.ParseContext) (any, error) {
	result, err := z.ZodBigInt.Parse(input, ctx...)
	if err != nil {
		return nil, err
	}
	if bigIntResult, ok := result.(*big.Int); ok {
		if bigIntResult.IsInt64() {
			return bigIntResult.Int64(), nil
		} else {
			// Value too large for int64
			return nil, &core.ZodError{
				Issues: []core.ZodRawIssue{{
					Code:    "too_big",
					Message: "BigInt value too large for int64",
					Path:    []any{},
					Input:   input,
				}},
			}
		}
	}
	return result, nil
}

func (z *ZodInt64BigInt) MustParse(input any, ctx ...*core.ParseContext) any {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// Parse validates and converts to uint64
func (z *ZodUint64BigInt) Parse(input any, ctx ...*core.ParseContext) (any, error) {
	result, err := z.ZodBigInt.Parse(input, ctx...)
	if err != nil {
		return nil, err
	}
	if bigIntResult, ok := result.(*big.Int); ok {
		if bigIntResult.IsUint64() {
			return bigIntResult.Uint64(), nil
		} else {
			// Value too large for uint64 or negative
			return nil, &core.ZodError{
				Issues: []core.ZodRawIssue{{
					Code:    "too_big",
					Message: "BigInt value too large for uint64 or negative",
					Path:    []any{},
					Input:   input,
				}},
			}
		}
	}
	return result, nil
}

func (z *ZodUint64BigInt) MustParse(input any, ctx ...*core.ParseContext) any {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// createZodBigIntFromDef creates a ZodBigInt from definition
func createZodBigIntFromDef(def *ZodBigIntDef) *ZodBigInt {
	internals := &ZodBigIntInternals{
		ZodTypeInternals: engine.NewBaseZodTypeInternals(def.Type),
		Def:              def,
		Pattern:          nil,
		Bag:              make(map[string]any),
	}

	// Set up simplified constructor for cloning
	internals.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any, any] {
		bigintDef := &ZodBigIntDef{
			ZodTypeDef: *newDef,
			Type:       "bigint",
		}
		return createZodBigIntFromDef(bigintDef)
	}

	internals.Parse = func(payload *core.ParsePayload, ctx *core.ParseContext) *core.ParsePayload {
		schema := &ZodBigInt{internals: internals}
		result, err := schema.Parse(payload.Value, ctx)
		if err != nil {
			var zodErr *issues.ZodError
			if errors.As(err, &zodErr) {
				for _, issue := range zodErr.Issues {
					rawIssue := core.ZodRawIssue{
						Code:    issue.Code,
						Input:   issue.Input,
						Path:    issue.Path,
						Message: issue.Message,
					}
					payload.Issues = append(payload.Issues, rawIssue)
				}
			}
			return payload
		}
		payload.Value = result
		return payload
	}

	zodSchema := &ZodBigInt{internals: internals}

	// Initialize the schema with proper error handling support
	engine.InitZodType(zodSchema, &def.ZodTypeDef)

	return zodSchema
}

// BigInt creates a new big.Int schema
func BigInt(params ...any) *ZodBigInt {
	def := &ZodBigIntDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   "bigint",
			Checks: make([]core.ZodCheck, 0),
		},
		Type: "bigint",
	}

	schema := createZodBigIntFromDef(def)

	// Apply schema parameters using the same pattern as string.go
	if len(params) > 0 {
		param := params[0]

		// Handle different parameter types
		switch p := param.(type) {
		case string:
			// String parameter becomes Error field
			errorMap := core.ZodErrorMap(func(issue core.ZodRawIssue) string {
				return p
			})
			def.Error = &errorMap
			schema.internals.Error = &errorMap
		case core.SchemaParams:
			// Handle core.SchemaParams
			if p.Coerce {
				schema.internals.Bag["coerce"] = true
				schema.internals.ZodTypeInternals.Bag["coerce"] = true
			}

			if p.Error != nil {
				// Handle string error messages by converting to ZodErrorMap
				if errStr, ok := p.Error.(string); ok {
					errorMap := core.ZodErrorMap(func(issue core.ZodRawIssue) string {
						return errStr
					})
					def.Error = &errorMap
					schema.internals.Error = &errorMap
				} else if errorMap, ok := p.Error.(core.ZodErrorMap); ok {
					def.Error = &errorMap
					schema.internals.Error = &errorMap
				}
			}

			if p.Description != "" {
				schema.internals.Bag["description"] = p.Description
			}
			if p.Abort {
				schema.internals.Bag["abort"] = true
			}
			if len(p.Path) > 0 {
				schema.internals.Bag["path"] = p.Path
			}
		}
	}

	return schema
}

// Int64BigInt creates a new int64-compatible BigInt schema
func Int64BigInt(params ...any) *ZodInt64BigInt {
	baseBigInt := BigInt(params...)

	// Add int64 range validation using individual checks
	minInt64 := big.NewInt(-9223372036854775808) // math.MinInt64
	maxInt64 := big.NewInt(9223372036854775807)  // math.MaxInt64

	// Apply range checks directly via AddCheck
	gteCheck := checks.Gte(minInt64)
	lteCheck := checks.Lte(maxInt64)

	rangedSchema := engine.AddCheck(baseBigInt, gteCheck)
	rangedSchema = engine.AddCheck(rangedSchema, lteCheck)

	if finalBigInt, ok := rangedSchema.(*ZodBigInt); ok {
		return &ZodInt64BigInt{finalBigInt}
	}

	// Fallback - this shouldn't happen but provides safety
	return &ZodInt64BigInt{baseBigInt}
}

// Uint64BigInt creates a new uint64-compatible BigInt schema
func Uint64BigInt(params ...any) *ZodUint64BigInt {
	baseBigInt := BigInt(params...)

	// Add uint64 range validation using individual checks
	minUint64 := big.NewInt(0)
	maxUint64 := new(big.Int).SetUint64(18446744073709551615) // math.MaxUint64

	// Apply range checks directly via AddCheck
	gteCheck := checks.Gte(minUint64)
	lteCheck := checks.Lte(maxUint64)

	rangedSchema := engine.AddCheck(baseBigInt, gteCheck)
	rangedSchema = engine.AddCheck(rangedSchema, lteCheck)

	if finalBigInt, ok := rangedSchema.(*ZodBigInt); ok {
		return &ZodUint64BigInt{finalBigInt}
	}

	// Fallback - this shouldn't happen but provides safety
	return &ZodUint64BigInt{baseBigInt}
}

//////////////////////////
// UTILITY FUNCTIONS
//////////////////////////

// extractBigIntValue extracts big.Int value, handling various input types
func extractBigIntValue(input any) (*big.Int, bool, error) {
	if input == nil {
		return nil, true, nil
	}

	switch v := input.(type) {
	case *big.Int:
		if v == nil {
			return nil, true, nil
		}
		return v, false, nil
	case big.Int:
		return &v, false, nil
	default:
		return nil, false, fmt.Errorf("expected big integer, got %T", input)
	}
}

// createBigIntCoerceFunc creates a coercion function for big.Int
func createBigIntCoerceFunc() func(any) (*big.Int, bool) {
	return func(value any) (*big.Int, bool) {
		result, err := coerce.ToBigInt(value)
		return result, err == nil
	}
}

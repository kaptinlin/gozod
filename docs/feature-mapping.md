# GoZod Feature Mapping Reference

This document provides a comprehensive feature mapping between TypeScript Zod v4 and GoZod validation library, detailing correspondences, unique features, and unimplemented functionality.

## TypeScript Zod v4 Complete Mapping

### Basic Type Mapping

| TypeScript Zod v4 | GoZod Constructor | Return Type | Smart Inference Features | Status |
|-------------------|------------------|-------------|--------------------------|--------|
| `z.string()` | `gozod.String()` | `*ZodString` | ✅ Supports string/\*string inference | ✅ Fully implemented |
| `z.number()` | `gozod.Float64()`, `gozod.Number()` | `*ZodFloatTyped[float64]` | ✅ Supports numeric type inference | ✅ Fully implemented |
| `z.boolean()` | `gozod.Bool()` | `*ZodBool` | ✅ Supports bool/\*bool inference | ✅ Fully implemented |
| `z.bigint()` | `gozod.BigInt()` | `*ZodBigInt` | ✅ Supports big integer inference | ✅ Fully implemented |
| `z.array(T)` | `gozod.Slice(T)` | `*ZodSlice` | ✅ Supports slice inference | ✅ Fully implemented |
| `z.tuple([...])` | `gozod.Array([...])` | `*ZodArray` | ✅ Supports array inference | ✅ Fully implemented |
| `z.object({})` | `gozod.Object({})` | `*ZodObject` | ✅ Supports object inference, smart type preservation | ✅ Fully implemented |
| `z.record(T)` | `gozod.Record(gozod.String(), T)` | `*ZodRecord` | ✅ Supports record inference | ✅ Fully implemented |
| `z.map(K, V)` | `gozod.Map(K, V)` | `*ZodMap` | ✅ Supports map inference | ✅ Fully implemented |
| `z.union([...])` | `gozod.Union([...])` | `*ZodUnion` | ✅ Supports union type inference | ✅ Fully implemented |
| `z.discriminatedUnion()` | `gozod.DiscriminatedUnion()` | `*ZodDiscriminatedUnion` | ✅ Supports discriminated union inference | ✅ Fully implemented |
| `z.intersection()` | `gozod.Intersection()` | `*ZodIntersection` | ✅ Supports intersection type inference | ✅ Fully implemented |
| `z.literal([...])` | `gozod.Literal([...])` | `*ZodLiteral` | ✅ Supports multi-value literal inference | ✅ Fully implemented |
| `z.enum([...])` | `gozod.Enum()`, `gozod.EnumMap()`, `gozod.EnumSlice()` | `*ZodEnum[T]` | ✅ Supports type-safe enum inference | ✅ Fully implemented |
| `z.lazy()` | `gozod.Lazy()` | `ZodType[any, any]` | ✅ Supports recursive pattern inference | ✅ Fully implemented |
| `z.function()` | `gozod.Function()` | `*ZodFunction` | ✅ Supports function type inference | ✅ Fully implemented |
| `z.any()` | `gozod.Any()` | `*ZodAny` | ✅ Supports any type inference | ✅ Fully implemented |
| `z.unknown()` | `gozod.Unknown()` | `*ZodUnknown` | ✅ Supports unknown type inference | ✅ Fully implemented |
| `z.never()` | `gozod.Never()` | `*ZodNever` | ✅ Never matches | ✅ Fully implemented |
| `z.null()` | `gozod.Nil()` | `ZodType[any, any]` | ✅ Supports nil inference | ✅ Fully implemented |
| `z.instanceof(Class)` | `gozod.Struct()` | `*ZodStruct` | ✅ Supports struct instance validation | ✅ Fully implemented |

### Go-Specific Numeric Types

| Go Type | GoZod Constructor | Return Type | Smart Inference Features | Status |
|---------|------------------|-------------|--------------------------|--------|
| `int` | `gozod.Int()` | `*ZodIntegerTyped[int]` | ✅ Supports integer inference | ✅ Fully implemented |
| `int8` | `gozod.Int8()` | `*ZodIntegerTyped[int8]` | ✅ Supports 8-bit integer inference | ✅ Fully implemented |
| `int16` | `gozod.Int16()` | `*ZodIntegerTyped[int16]` | ✅ Supports 16-bit integer inference | ✅ Fully implemented |
| `int32` | `gozod.Int32()`, `gozod.Rune()` | `*ZodIntegerTyped[int32]` | ✅ Supports 32-bit integer inference | ✅ Fully implemented |
| `int64` | `gozod.Int64()` | `*ZodIntegerTyped[int64]` | ✅ Supports 64-bit integer inference | ✅ Fully implemented |
| `uint` | `gozod.Uint()` | `*ZodIntegerTyped[uint]` | ✅ Supports unsigned integer inference | ✅ Fully implemented |
| `uint8` | `gozod.Uint8()`, `gozod.Byte()` | `*ZodIntegerTyped[uint8]` | ✅ Supports 8-bit unsigned integer inference | ✅ Fully implemented |
| `uint16` | `gozod.Uint16()` | `*ZodIntegerTyped[uint16]` | ✅ Supports 16-bit unsigned integer inference | ✅ Fully implemented |
| `uint32` | `gozod.Uint32()` | `*ZodIntegerTyped[uint32]` | ✅ Supports 32-bit unsigned integer inference | ✅ Fully implemented |
| `uint64` | `gozod.Uint64()` | `*ZodIntegerTyped[uint64]` | ✅ Supports 64-bit unsigned integer inference | ✅ Fully implemented |
| `float32` | `gozod.Float32()` | `*ZodFloatTyped[float32]` | ✅ Supports single precision float inference | ✅ Fully implemented |
| `float64` | `gozod.Float64()`, `gozod.Number()` | `*ZodFloatTyped[float64]` | ✅ Supports double precision float inference | ✅ Fully implemented |
| `complex64` | `gozod.Complex64()` | `*ZodComplex[complex64]` | ✅ Supports single precision complex inference | ✅ Fully implemented |
| `complex128` | `gozod.Complex128()` | `*ZodComplex[complex128]` | ✅ Supports double precision complex inference | ✅ Fully implemented |

### Go-Specific Special Types

| Go Feature Type | GoZod Constructor | Return Type | Smart Inference Features | Status |
|-----------------|------------------|-------------|--------------------------|--------|
| String to Boolean Conversion | `gozod.StringBool()` | `*ZodStringBool` | ✅ Supports string to boolean conversion | ✅ Fully implemented |
| File Upload | `gozod.File()` | `*ZodFile` | ✅ Supports file type inference | ✅ Fully implemented |
| IPv4 Address | `gozod.IPv4()` | `*ZodIPv4` | ✅ Supports IPv4 address inference | ✅ Fully implemented |
| IPv6 Address | `gozod.IPv6()` | `*ZodIPv6` | ✅ Supports IPv6 address inference | ✅ Fully implemented |
| IPv4 CIDR | `gozod.CIDRv4()` | `*ZodCIDRv4` | ✅ Supports IPv4 CIDR inference | ✅ Fully implemented |
| IPv6 CIDR | `gozod.CIDRv6()` | `*ZodCIDRv6` | ✅ Supports IPv6 CIDR inference | ✅ Fully implemented |
| Custom Validation | `gozod.Custom()`, `gozod.Check()` | `*ZodCustom` | ✅ Supports custom validation inference | ✅ Fully implemented |
| Struct Validation | `gozod.Struct()` | `*ZodStruct` | ✅ Supports struct inference | ✅ Fully implemented |

### Modifier Mapping

| TypeScript Zod v4 | GoZod Implementation | Wrapper Type | Smart Inference Preservation | Status |
|-------------------|---------------------|-------------|------------------------------|--------|
| `.optional()` | `.Optional()` | `ZodOptional[T]` | ✅ Missing field semantics - returns generic `nil` | ✅ Fully implemented |
| `.nullable()` | `.Nilable()` | `ZodNilable[T]` | ✅ Null value semantics - returns typed `(*T)(nil)` | ✅ Fully implemented |
| `.nullish()` | `.Nullish()` | `ZodOptional[T]` | ✅ Combined semantics | ✅ Fully implemented |
| `.default(value)` | `.Default(value)` | `ZodDefault[T]` | ✅ Default value mechanism - replaces `nil` input | ✅ Fully implemented |
| `.default(() => value)` | `.DefaultFunc(func() T)` | `ZodDefault[T]` | ✅ Function-based default value | ✅ Fully implemented |
| `.catch(value)` | `.Prefault(value)` | `ZodPrefault[T]` | ✅ Fallback value mechanism - replaces any validation failure | ✅ Fully implemented |
| `.catch(() => value)` | `.PrefaultFunc(func() T)` | `ZodPrefault[T]` | ✅ Function-based fallback value | ✅ Fully implemented |

### Validation and Transform Method Mapping

| TypeScript Zod v4 | GoZod Method | Type-Safe Version | Flexible Version | Status |
|-------------------|-------------|------------------|------------------|--------|
| `.refine(fn)` | `.Refine(fn)` | ✅ `func(T) bool` | `RefineAny(func(any) bool)` | ✅ Fully implemented |
| `.transform(fn)` | `.Transform(fn)` | ✅ `func(T, ctx) (any, error)` | `TransformAny(func(any, ctx) (any, error))` | ✅ Fully implemented |
| `.pipe(schema)` | `.Pipe(schema)` | ✅ Type-safe pipeline | - | ✅ Fully implemented |
| `.parse(data)` | `.Parse(data)` | ✅ Smart type inference with pointer identity preservation | - | ✅ Fully implemented |
| `.parseSync(data)` | `.MustParse(data)` | ✅ Returns result or panics | - | ✅ Fully implemented |

### Coercion Method Mapping

| TypeScript Zod v4 | GoZod Method | Status | Description |
|-------------------|-----------------|--------|-------------|
| `z.coerce.string()` | `gozod.Coerce.String()` | ✅ | Force string conversion, supports int, float, bool, etc. |
| `z.coerce.number()` | `gozod.Coerce.Number()` | ✅ | Force numeric conversion, supports string, bool, etc. |
| `z.coerce.boolean()` | `gozod.Coerce.Bool()` | ✅ | Force boolean conversion, supports string, number, etc. |
| `z.coerce.bigint()` | `gozod.Coerce.BigInt()` | ✅ | Force big integer conversion, supports string, number, etc. |
| - | `gozod.Coerce.Complex64()` | ✅ | Go unique: Force single precision complex conversion |
| - | `gozod.Coerce.Complex128()` | ✅ | Go unique: Force double precision complex conversion |

### String Validation Method Mapping

| TypeScript Zod v4 | GoZod Method | Status | Description |
|-------------------|-------------|--------|-------------|
| `.min(n)` | `.Min(n)` | ✅ | Minimum length validation |
| `.max(n)` | `.Max(n)` | ✅ | Maximum length validation |
| `.length(n)` | `.Length(n)` | ✅ | Exact length validation |
| `.email()` | `.Email()` | ✅ | Email format validation |
| `.url()` | `.URL()` | ✅ | URL format validation |
| `.uuid()` | `.UUID()` | ✅ | UUID format validation |
| `.regex(pattern)` | `.Regex(pattern)` | ✅ | Regular expression validation |
| `.includes(str)` | `.Includes(str)` | ✅ | Contains substring validation |
| `.startsWith(str)` | `.StartsWith(str)` | ✅ | Starts with validation |
| `.endsWith(str)` | `.EndsWith(str)` | ✅ | Ends with validation |
| `.json()` | `.JSON()` | ✅ | JSON format validation |
| `.trim()` | `.Trim()` | ✅ | Trim whitespace transformation |
| `.toLowerCase()` | `.ToLowerCase()` | ✅ | Convert to lowercase transformation |
| `.toUpperCase()` | `.ToUpperCase()` | ✅ | Convert to uppercase transformation |

### Numeric Validation Method Mapping

| TypeScript Zod v4 | GoZod Method | Status | Description |
|-------------------|-------------|--------|-------------|
| `.min(n)` | `.Min(n)` | ✅ | Minimum value validation |
| `.max(n)` | `.Max(n)` | ✅ | Maximum value validation |
| `.gt(n)` | `.Gt(n)` | ✅ | Greater than validation |
| `.gte(n)` | `.Gte(n)` | ✅ | Greater than or equal validation |
| `.lt(n)` | `.Lt(n)` | ✅ | Less than validation |
| `.lte(n)` | `.Lte(n)` | ✅ | Less than or equal validation |
| `.positive()` | `.Positive()` | ✅ | Positive number validation |
| `.negative()` | `.Negative()` | ✅ | Negative number validation |
| `.nonnegative()` | `.NonNegative()` | ✅ | Non-negative number validation |
| `.nonpositive()` | `.NonPositive()` | ✅ | Non-positive number validation |
| `.multipleOf(n)` | `.MultipleOf(n)` | ✅ | Multiple validation |
| `.int()` | `.Int()` | ✅ | Integer validation |
| `.finite()` | `.Finite()` | ✅ | Finite number validation |
| `.safe()` | `.Safe()` | ✅ | Safe number validation |

### Array/Slice Validation Method Mapping

| TypeScript Zod v4 | GoZod Method | Status | Description |
|-------------------|-------------|--------|-------------|
| `.min(n)` | `.Min(n)` | ✅ | Minimum length validation |
| `.max(n)` | `.Max(n)` | ✅ | Maximum length validation |
| `.length(n)` | `.Length(n)` | ✅ | Exact length validation |
| `.nonempty()` | `.NonEmpty()` | ✅ | Non-empty validation |

### Object Validation Method Mapping

| TypeScript Zod v4 | GoZod Method | Status | Description |
|-------------------|-------------|--------|-------------|
| `.pick({...})` | `.Pick([...])` | ✅ | Select fields, Default/Prefault auto-filtered |
| `.omit({...})` | `.Omit([...])` | ✅ | Exclude fields, Default/Prefault auto-filtered |
| `.partial()` | `.Partial()` | ✅ | All fields optional |
| `.required()` | `.Required([...])` | ✅ | Specified fields required |
| `.extend({...})` | `.Extend({...})` | ✅ | Extend fields |
| `.merge(schema)` | `.Merge(schema)` | ✅ | Merge schemas |
| `.keyof()` | `.Keyof()` | ✅ | Key type extraction |
| `.strict()` | `gozod.StrictObject({...})` | ✅ | Strict mode (top-level function) |
| `.strip()` | `.Strip()` | ✅ | Strip mode |
| `.passthrough()` | `gozod.LooseObject({...})` | ✅ | Passthrough mode (top-level function) |
| `.catchall(schema)` | `.Catchall(schema)` | ✅ | Catch-all mode |

## GoZod Unique Features

| GoZod Unique Method | Feature Description | TypeScript Equivalent |
|--------------------|---------------------|----------------------|
| `RefineAny(func(any) bool)` | Any-type validator | No equivalent, requires type assertion |
| `TransformAny(func(any, ctx) (any, error))` | Any-type transformer | No equivalent, requires type assertion |
| `DefaultFunc(func() T)` | Function-based default value generation | No equivalent, only supports static values |
| `PrefaultFunc(func() T)` | Function-based fallback value generation | No equivalent, only supports static values |

## Key Behavioral Differences

### Type Inference and Pointer Identity

GoZod provides intelligent type inference that preserves exact input types:

```go
// Pointer identity preservation
input := "hello"
result, _ := gozod.String().Parse(&input)
fmt.Println(result == &input) // true - same memory address
```

### Optional vs Nilable Semantics

Key distinction:

```go
// Optional: "missing field" semantics
optionalResult, _ := gozod.String().Optional().Parse(nil) // returns: nil (generic)

// Nilable: "null value" semantics  
nilableResult, _ := gozod.String().Nilable().Parse(nil)   // returns: (*string)(nil) (typed)
```

### Default vs Prefault Behavior

```go
// Default: Only replaces nil input
defaultSchema := gozod.String().Default("fallback")
result, _ := defaultSchema.Parse(nil)  // "fallback"
_, err := defaultSchema.Parse(123)     // Error: wrong type

// Prefault: Replaces ANY validation failure
prefaultSchema := gozod.String().Prefault("fallback")
result, _ := prefaultSchema.Parse(nil)  // "fallback"
result, _ = prefaultSchema.Parse(123)   // "fallback" (no error!)
```

## TypeScript Zod Features Not Implemented in GoZod

| Feature / Method | Reason | GoZod Notes |
|------------------|--------|-------------|
| Async validation (`.parseAsync()`, `.refine(async fn)`) | Go is a synchronous language | Async validation is not supported due to language and performance model |
| Void type (`z.void()`) | Go lacks a void concept | Go functions return specific types or empty return |
| Undefined type (`z.undefined()`) | Go lacks undefined | Go uses nil or zero values to represent missing data |
| Branded Types (`z.string().brand<"UserId">()`) | Go type system limitations | Can be implemented via custom types |
| Preprocessor (`z.preprocess()`) | Can be implemented in other ways | Use `Transform` and `Pipe` combination instead |
| TypeScript-specific features (`z.infer<>`, `z.readonly()`) | Language feature differences | Go has different type inference and immutability patterns |
| SafeParse (`schema.safeParse()`) | Go error handling pattern | Go uses `(result, error)` return values, so SafeParse is not needed |

---

This mapping demonstrates GoZod's comprehensive compatibility with TypeScript Zod v4 while providing Go-specific enhancements and optimizations. 

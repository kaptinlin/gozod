# GoZod Feature Mapping Reference

This document provides a comprehensive feature mapping between TypeScript Zod v4 and GoZod validation library, detailing correspondences, unique features, and unimplemented functionality.

## TypeScript Zod v4 Complete Mapping

### Basic Type Mapping

| TypeScript Zod v4 | GoZod Constructor | Return Type | Go-Specific Features | Status |
|-------------------|------------------|-------------|---------------------|--------|
| `z.string()` | `gozod.String()` | `string` | ✅ Strict type semantics: requires exact string input | ✅ Fully implemented |
| `z.number()` | `gozod.Float64()`, `gozod.Number()` | `float64` | ✅ Go numeric types: int, float32, float64, etc. | ✅ Fully implemented |
| `z.boolean()` | `gozod.Bool()` | `bool` | ✅ Strict boolean validation | ✅ Fully implemented |
| `z.bigint()` | `gozod.BigInt()` | `*big.Int` | ✅ Go native big.Int support | ✅ Fully implemented |
| `z.date()` | `gozod.Time()` | `time.Time` | ✅ Go native time.Time with timezone support | ✅ Fully implemented |
| `z.array(T)` | `gozod.Slice[T](elementSchema)` | `[]T` | ✅ Type-safe generic slices with element validation | ✅ Fully implemented |
| `z.tuple([...])` | `gozod.Array([...])` | `[]any` | ✅ Fixed-length tuple validation | ✅ Fully implemented |
| `z.object({})` | `gozod.Object({})` | `map[string]any` | ✅ Dynamic object validation for JSON-like data | ✅ Fully implemented |
| `z.record(T)` | `gozod.Record(keySchema, valueSchema)` | `map[string]T` | ✅ Typed key-value record validation | ✅ Fully implemented |
| `z.map(K, V)` | `gozod.Map(K, V)` | `map[K]V` | ✅ Go native map validation with typed keys/values | ✅ Fully implemented |
| `z.union([...])` | `gozod.Union([...])` | `any` | ✅ Type-safe union validation with Go interfaces | ✅ Fully implemented |
| `z.discriminatedUnion(key, [...])` | `gozod.DiscriminatedUnion(key, [...])` | `any` | ✅ Optimized discriminated union with key-based lookup | ✅ Fully implemented |
| `z.intersection(A, B)` | `gozod.Intersection(A, B)` | `any` | ✅ Intersection type validation | ✅ Fully implemented |
| `z.literal(value)` | `gozod.Literal(value)` | `T` | ✅ Type-safe literal value validation | ✅ Fully implemented |
| `z.enum([...])` | `gozod.Enum(...)`, `gozod.EnumMap()`, `gozod.EnumSlice()` | `T` | ✅ Go native enum support with iota constants | ✅ Fully implemented |
| `z.lazy(() => schema)` | `gozod.Lazy(() => schema)` | `T` | ✅ Recursive schema support | ✅ Fully implemented |
| `z.function()` | `gozod.Function()` | `func` | ✅ Go function validation | ✅ Fully implemented |
| `z.any()` | `gozod.Any()` | `any` | ✅ Accept any value | ✅ Fully implemented |
| `z.unknown()` | `gozod.Unknown()` | `any` | ✅ Unknown data validation | ✅ Fully implemented |
| `z.never()` | `gozod.Never()` | - | ✅ Never type validation | ✅ Fully implemented |
| `z.null()` / `z.undefined()` | `gozod.Nil()` | `nil` | ✅ Go nil validation (no undefined in Go) | ✅ Fully implemented |
| - | `gozod.Struct[T]()` | `T` | ✅ **Go-specific**: Native struct validation with generics | ✅ Go enhancement |

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

| TypeScript Zod v4 | GoZod Implementation | Semantic Difference | Go-Specific Behavior | Status |
|-------------------|---------------------|-------------------|-------------------|--------|
| `.optional()` | `.Optional()` | Missing field handling | ✅ Returns `*T` for flexible input, handles nil gracefully | ✅ Fully implemented |
| `.nullable()` | `.Nilable()` | Null value handling | ✅ Returns `*T`, typed nil semantics for Go | ✅ Fully implemented |
| `.nullish()` | `.Nullish()` | Combined optional+nullable | ✅ Handles both missing and nil values | ✅ Fully implemented |
| `.default(value)` | `.Default(value)` | Default fallback | ✅ Short-circuits validation for nil input | ✅ Fully implemented |
| - | `.DefaultFunc(func() T)` | Dynamic defaults | ✅ **Go enhancement**: Function-based default generation | ✅ Go enhancement |
| `.catch(value)` | `.Prefault(value)` | Error fallback | ✅ Pre-parse default with full validation pipeline | ✅ Fully implemented |
| - | `.PrefaultFunc(func() T)` | Dynamic error fallback | ✅ **Go enhancement**: Function-based prefault generation | ✅ Go enhancement |

### Validation and Transform Method Mapping

| TypeScript Zod v4 | GoZod Method | TypeScript Signature | GoZod Signature | Status |
|-------------------|-------------|---------------------|----------------|--------|
| `.refine(fn, message?)` | `.Refine(fn, params?)` | `(val: T) => boolean` | `func(T) bool` + `core.SchemaParams` | ✅ Fully implemented |
| `.transform(fn)` | `.Transform(fn)` | `(val: T) => U` | `func(T, *core.RefinementContext) (any, error)` | ✅ Fully implemented |
| `.pipe(schema)` | `.Pipe(schema)` | `ZodTypeAny` | `core.ZodType[any, any]` | ✅ Fully implemented |
| `.parse(data)` | `.Parse(data)` | `unknown -> T` (throws) | `any -> (T, error)` | ✅ Go error handling |
| `.safeParse(data)` | `.Parse(data)` | `unknown -> SafeParseResult<T>` | `any -> (T, error)` | ✅ Go uses (T, error) pattern |
| `.parseAsync(data)` | *Not implemented* | `Promise<T>` | - | ❌ Go is synchronous |
| - | `.StrictParse(data)` | - | `T -> (T, error)` | ✅ **Go enhancement**: Compile-time type safety |
| - | `.MustParse(data)` | - | `any -> T` (panics on error) | ✅ **Go enhancement**: Panic-based parsing |
| - | `.MustStrictParse(data)` | - | `T -> T` (panics on error) | ✅ **Go enhancement**: Type-safe panic parsing |

### Coercion Method Mapping

| TypeScript Zod v4 | GoZod Method | Status | Description |
|-------------------|-----------------|--------|-------------|
| `z.coerce.string()` | `coerce.String()` | ✅ | Force string conversion from int, float, bool, etc. |
| `z.coerce.number()` | `coerce.Number()`, `coerce.Float64()` | ✅ | Force numeric conversion from string, bool, etc. |
| `z.coerce.boolean()` | `coerce.Bool()` | ✅ | Force boolean conversion from string, number, etc. |
| `z.coerce.bigint()` | `coerce.BigInt()` | ✅ | Force big integer conversion from string, number, etc. |
| `z.coerce.date()` | `coerce.Time()` | ✅ | Force time.Time conversion from string, Unix timestamp, etc. |
| - | `coerce.StringBool()` | ✅ | **Go-specific**: String to boolean coercion ("true"/"false", "1"/"0", etc.) |
| - | `coerce.Int()`, `coerce.Int8()`, `coerce.Int16()`, etc. | ✅ | **Go-specific**: All Go integer types with coercion |
| - | `coerce.Float32()`, `coerce.Complex64()`, `coerce.Complex128()` | ✅ | **Go-specific**: Advanced numeric types with coercion |

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

| GoZod Unique Method | Feature Description | Why Go-Specific | TypeScript Equivalent |
|--------------------|---------------------|----------------|----------------------|
| `StringPtr()`, `IntPtr()`, `BoolPtr()`, etc. | Strict pointer type validation | Go has explicit pointer types | No direct equivalent - TypeScript uses union types |
| `Struct[T]()`, `StructPtr[T]()` | Native Go struct validation with generics | Go's struct system | No direct equivalent - uses `z.object()` for similar purpose |
| `coerce.StringBool()` | String to boolean coercion ("true"/"1"/"yes" → `true`) | Common in Go JSON/config parsing | No direct equivalent |
| `StrictParse(T) -> (T, error)` | Compile-time type checking | Go's strong type system | TypeScript has compile-time inference |
| `RefineAny(func(any) bool)` | Runtime type validation | Go's interface{} system | Uses type guards/predicates |
| `TransformAny(func(any, ctx) (any, error))` | Runtime type transformation | Go's interface{} system | Uses type assertions |
| `DefaultFunc(func() T)` | Function-based default generation | Go functions are first-class | Limited to static defaults |
| `PrefaultFunc(func() T)` | Function-based error fallback | Go error handling patterns | Limited to static catch values |
| `Partial()`, `Required()` on structs | Partial struct validation | Go's zero value semantics | Object-level operations only |

### Complete Strict Type Semantics

GoZod uses complete strict type semantics - all constructors require exact input types:

```go
// Value schemas - strict value input only
stringSchema := gozod.String()      // Only accepts string, returns string
intSchema := gozod.Int()            // Only accepts int, returns int
structSchema := gozod.Struct[T]()   // Only accepts T, returns T

// Pointer schemas - strict pointer input only
stringPtrSchema := gozod.StringPtr()  // Only accepts *string, returns *string
intPtrSchema := gozod.IntPtr()        // Only accepts *int, returns *int
structPtrSchema := gozod.StructPtr[T]()  // Only accepts *T, returns *T

// No automatic conversions - explicit type handling required
value := "hello"
// stringSchema.Parse(&value)  // ❌ Error: requires string, got *string
// stringPtrSchema.Parse(value) // ❌ Error: requires *string, got string

// Use Optional/Nilable for flexible input handling
optionalSchema := gozod.String().Optional()  // Flexible input, *string output
```

## Key Behavioral Differences

### TypeScript Zod vs GoZod: Syntax Comparison

**TypeScript Zod:**
```typescript
import { z } from 'zod';

// Schema definition
const UserSchema = z.object({
  name: z.string().min(2),
  age: z.number().min(0).optional(),
  email: z.string().email(),
});

// Parsing (throws ZodError on failure)
const user = UserSchema.parse(data);

// Safe parsing (returns success/error object)
const result = UserSchema.safeParse(data);
if (result.success) {
  console.log(result.data);
} else {
  console.log(result.error);
}

// Type inference
type User = z.infer<typeof UserSchema>;
```

**GoZod:**
```go
import "github.com/kaptinlin/gozod"

// Struct-based validation (Go-specific)
type User struct {
    Name  string `json:"name"`
    Age   *int   `json:"age,omitempty"`
    Email string `json:"email"`
}

userSchema := gozod.Struct[User](gozod.StructSchema{
    "name":  gozod.String().Min(2),
    "age":   gozod.Int().Min(0).Optional(),
    "email": gozod.String().Email(),
})

// Go error handling pattern
user, err := userSchema.Parse(data)
if err != nil {
    // Handle validation error
    return err
}

// Panic-based parsing (equivalent to TS .parse())
user := userSchema.MustParse(data) // Panics on error

// Type is inferred from generic parameter
// user is of type User
```

### Complete Strict Type Semantics

GoZod enforces strict type requirements with explicit pointer handling:

```go
// Value vs Pointer schemas
input := "hello"

// Value schemas: exact type matching
stringSchema := gozod.String()                   // Requires string input
result1, _ := stringSchema.Parse("hello")        // ✅ Valid
// result1, _ := stringSchema.Parse(&input)       // ❌ Error: wrong type

// Pointer schemas: explicit pointer handling  
stringPtrSchema := gozod.StringPtr()             // Requires *string input
result2, _ := stringPtrSchema.Parse(&input)      // ✅ Valid (preserves pointer identity)
// result2, _ := stringPtrSchema.Parse("hello")   // ❌ Error: wrong type

// Flexible input with Optional/Nilable
optionalSchema := gozod.String().Optional()      // Returns *string
result3, _ := optionalSchema.Parse("hello")      // ✅ string → *string
result4, _ := optionalSchema.Parse(&input)       // ✅ *string → *string (preserves identity)
result5, _ := optionalSchema.Parse(nil)          // ✅ nil → nil
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
// Default: Only replaces nil input (short-circuits, no parsing)
defaultSchema := gozod.String().Default("fallback")
result, _ := defaultSchema.Parse(nil)  // "fallback"
_, err := defaultSchema.Parse(123)     // Error: wrong type

// Prefault: Pre-parse default for nil input (goes through full parsing)
prefaultSchema := gozod.String().Transform(func(s string) string { return strings.ToUpper(s) }).Prefault("fallback")
result, _ := prefaultSchema.Parse(nil)  // "FALLBACK" (nil uses prefault, then transforms)
_, err := prefaultSchema.Parse(123)     // Error: wrong type (no prefault for non-nil)
```

## TypeScript Zod v4 Features vs GoZod Implementation

### Features Not Directly Applicable in Go

| TypeScript Zod v4 Feature | Reason Not Applicable | GoZod Alternative/Notes |
|---------------------------|----------------------|------------------------|
| **Async validation** (`.parseAsync()`, `async` refinements) | Go is synchronous by design | Not applicable - Go uses goroutines for concurrency |
| **Void type** (`z.void()`) | Go has no void concept | Not needed - Go functions return specific types or nothing |
| **Undefined type** (`z.undefined()`) | Go has no undefined | Use `nil` or zero values instead |
| **Branded types** (`z.string().brand<"UserId">()`) | Different type system | Use custom Go types: `type UserId string` |
| **Type inference** (`z.infer<T>`) | Language-level feature | Go generics provide compile-time type safety |
| **Readonly** (`z.readonly()`) | Language-level feature | Go doesn't have readonly - use immutability patterns |
| **Preprocessor** (`z.preprocess()`) | Not needed | Use `.Transform()` and `.Pipe()` for preprocessing |
| **SafeParse result object** | Different error handling | Go uses `(T, error)` pattern - no need for result wrapper |

### TypeScript Zod v4 New Features and GoZod Status

| Zod v4 Feature | GoZod Implementation | Status | Notes |
|----------------|---------------------|--------|-------|
| `z.stringbool()` | `gozod.StringBool()`, `coerce.StringBool()` | ✅ Implemented | String to boolean parsing |
| Enhanced template literal schemas | Template validation via regex/custom | ✅ Available | Use `.Regex()` or custom refinements |
| Improved error messages | Structured error system | ✅ Enhanced | Rich error context with paths |
| `z.strictObject()` | Object validation mode | ✅ Available | Default object behavior is strict |
| Enhanced enum handling | Multiple enum constructors | ✅ Enhanced | `Enum()`, `EnumMap()`, `EnumSlice()` with Go type support |

### Go-Specific Enhancements Not in TypeScript Zod

| GoZod Feature | Description | TypeScript Equivalent |
|---------------|-------------|----------------------|
| **Struct validation** | Native Go struct support with field mapping | No direct equivalent - uses object schemas |
| **Pointer type safety** | Explicit pointer vs value type validation | No equivalent - handled via union types |
| **All Go numeric types** | int8, int16, uint32, complex64, etc. | Limited to `number` and `bigint` |
| **Coercion package** | Separate import for type coercion | Built into `z.coerce.*` |
| **Strict vs flexible parsing** | `Parse()` vs `StrictParse()` methods | Compile-time inference only |
| **Partial struct validation** | Skip zero-value fields in structs | Object-level only |
| **JSON tag mapping** | Automatic field mapping from struct tags | Manual field specification |

---

This mapping demonstrates GoZod's comprehensive compatibility with TypeScript Zod v4 while providing Go-specific enhancements and optimizations.

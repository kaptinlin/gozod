# GoZod Feature Mapping Reference

This document provides a comprehensive feature mapping between TypeScript Zod v4 and GoZod validation library, detailing correspondences, unique features, and implementation status.

## TypeScript Zod v4 Complete Mapping

### Basic Type Mapping

| TypeScript Zod v4 | GoZod Constructor | Return Type | Go-Specific Features | Status |
|-------------------|------------------|-------------|---------------------|--------|
| `z.string()` | `gozod.String()` | `string` | ✅ Complete strict type semantics + `StrictParse()` | ✅ Fully implemented |
| `z.number()` | `gozod.Float64()`, `gozod.Number()` | `float64` | ✅ All Go numeric types with type-safe constructors | ✅ Fully implemented |
| `z.boolean()` | `gozod.Bool()` | `bool` | ✅ Strict boolean validation with pointer variants | ✅ Fully implemented |
| `z.bigint()` | `gozod.BigInt()` | `*big.Int` | ✅ Go native big.Int with full arithmetic support | ✅ Fully implemented |
| `z.date()` | `gozod.Time()` | `time.Time` | ✅ Go native time.Time with timezone and format support | ✅ Fully implemented |
| `z.array(T)` | `gozod.Array(elementSchema)`, `gozod.Slice(elementSchema)` | `[]T` | ✅ Type-safe generic arrays with element validation | ✅ Fully implemented |
| `z.tuple([...])` | `gozod.Tuple([...])` | `[]any` | ✅ Fixed-length tuple validation with type inference | ✅ Fully implemented |
| `z.object({})` | `gozod.Object({})` | `map[string]any` | ✅ Dynamic object validation for JSON-like data | ✅ Fully implemented |
| `z.record(T)` | `gozod.Record(keySchema, valueSchema)` | `map[string]T` | ✅ Typed key-value record validation with generic keys | ✅ Fully implemented |
| `z.partialRecord(K, V)` | `gozod.PartialRecord(keySchema, valueSchema)`, `.Partial()` | `map[string]V` | ✅ Record with optional keys (skips exhaustiveness check) | ✅ Fully implemented |
| `z.looseRecord(K, V)` | `gozod.LooseRecord(keySchema, valueSchema)` | `map[string]V` | ✅ Record that passes through non-matching keys | ✅ Fully implemented |
| `z.map(K, V)` | `gozod.Map(valueSchema)` | `map[string]V` | ✅ Go native map validation with typed values | ✅ Fully implemented |
| `z.set(T)` | `gozod.Set(elementSchema)` | `map[T]struct{}` | ✅ Go idiomatic set pattern with element validation | ✅ Fully implemented |
| `z.union([...])` | `gozod.Union([...])` | `any` | ✅ Type-safe union validation with Go interfaces | ✅ Fully implemented |
| `z.xor([...])` | `gozod.Xor([...])` | `any` | ✅ Exclusive union - exactly one must match | ✅ Fully implemented |
| `z.discriminatedUnion(key, [...])` | `gozod.DiscriminatedUnion(key, [...])` | `any` | ✅ Optimized discriminated union with key-based lookup | ✅ Fully implemented |
| `z.intersection(A, B)` | `gozod.Intersection(A, B)` | `any` | ✅ Intersection type validation with Go type system | ✅ Fully implemented |
| `z.literal(value)` | `gozod.Literal(value)` | `T` | ✅ Type-safe literal value validation | ✅ Fully implemented |
| `z.enum([...])` | `gozod.Enum(...)` | `T` | ✅ Go native enum support with type constraints | ✅ Fully implemented |
| `z.lazy(() => schema)` | `gozod.Lazy(() => schema)` | `T` | ✅ Recursive schema support with automatic circular reference detection | ✅ Fully implemented |
| `z.function()` | `gozod.Function()` | `func` | ✅ Go function type validation | ✅ Fully implemented |
| `z.any()` | `gozod.Any()` | `any` | ✅ Accept any value with type preservation | ✅ Fully implemented |
| `z.unknown()` | `gozod.Unknown()` | `any` | ✅ Unknown data validation | ✅ Fully implemented |
| `z.never()` | `gozod.Never()` | `never` | ✅ Never type validation (always fails) | ✅ Fully implemented |
| `z.null()` / `z.undefined()` | `gozod.Nil()` | `nil` | ✅ Go nil validation (no undefined in Go) | ✅ Fully implemented |
| - | `gozod.Struct[T]()` | `T` | ✅ **Go-specific**: Native struct validation with generics | ✅ Go enhancement |
| - | `gozod.FromStruct[T]()` | `T` | ✅ **Go-specific**: Declarative struct tag validation | ✅ Go enhancement |

### Go-Specific Numeric Types

| Go Type | GoZod Constructor | Return Type | Type Safety Features | Status |
|---------|------------------|-------------|----------------------|--------|
| `int` | `gozod.Int()` | `int` | ✅ Complete strict type semantics with `StrictParse()` | ✅ Fully implemented |
| `int8` | `gozod.Int8()` | `int8` | ✅ 8-bit signed integer with overflow protection | ✅ Fully implemented |
| `int16` | `gozod.Int16()` | `int16` | ✅ 16-bit signed integer with overflow protection | ✅ Fully implemented |
| `int32` | `gozod.Int32()`, `gozod.Rune()` | `int32` | ✅ 32-bit signed integer and rune type support | ✅ Fully implemented |
| `int64` | `gozod.Int64()` | `int64` | ✅ 64-bit signed integer with full range support | ✅ Fully implemented |
| `uint` | `gozod.Uint()` | `uint` | ✅ Platform-dependent unsigned integer | ✅ Fully implemented |
| `uint8` | `gozod.Uint8()`, `gozod.Byte()` | `uint8` | ✅ 8-bit unsigned integer and byte type support | ✅ Fully implemented |
| `uint16` | `gozod.Uint16()` | `uint16` | ✅ 16-bit unsigned integer with overflow protection | ✅ Fully implemented |
| `uint32` | `gozod.Uint32()` | `uint32` | ✅ 32-bit unsigned integer with overflow protection | ✅ Fully implemented |
| `uint64` | `gozod.Uint64()` | `uint64` | ✅ 64-bit unsigned integer with full range support | ✅ Fully implemented |
| `float32` | `gozod.Float32()` | `float32` | ✅ Single precision float with finite validation | ✅ Fully implemented |
| `float64` | `gozod.Float64()`, `gozod.Number()` | `float64` | ✅ Double precision float with NaN/Inf handling | ✅ Fully implemented |
| `complex64` | `gozod.Complex64()` | `complex64` | ✅ Single precision complex number validation | ✅ Fully implemented |
| `complex128` | `gozod.Complex128()` | `complex128` | ✅ Double precision complex number validation | ✅ Fully implemented |

### Modifier Mapping

| TypeScript Zod v4 | GoZod Implementation | Semantic Difference | Go-Specific Behavior | Status |
|-------------------|---------------------|-------------------|-------------------|--------|
| `.optional()` | `.Optional()` | Missing field handling | ✅ Returns `*T` for flexible input, preserves pointer identity | ✅ Fully implemented |
| `.nullable()` | `.Nilable()` | Null value handling | ✅ Returns `*T`, typed nil semantics for Go | ✅ Fully implemented |
| `.nullish()` | `.Nullish()` | Combined optional+nullable | ✅ Handles both missing and nil values with type safety | ✅ Fully implemented |
| `.default(value)` | `.Default(value)` | Default fallback | ✅ Short-circuits validation for nil input (Zod v4 compatible) | ✅ Fully implemented |
| - | `.DefaultFunc(func() T)` | Dynamic defaults | ✅ **Go enhancement**: Function-based default generation | ✅ Go enhancement |
| `.catch(value)` | `.Prefault(value)` | Error fallback | ✅ Pre-parse default with full validation pipeline (Zod v4 compatible) | ✅ Fully implemented |
| - | `.PrefaultFunc(func() T)` | Dynamic error fallback | ✅ **Go enhancement**: Function-based prefault generation | ✅ Go enhancement |

### Parse Method Mapping

| TypeScript Zod v4 | GoZod Method | TypeScript Signature | GoZod Signature | Status |
|-------------------|-------------|---------------------|----------------|--------|
| `.parse(data)` | `.Parse(data)` | `unknown -> T` (throws) | `any -> (T, error)` | ✅ Go error handling pattern |
| `.safeParse(data)` | `.Parse(data)` | `unknown -> SafeParseResult<T>` | `any -> (T, error)` | ✅ Go uses (T, error) pattern |
| `.parseAsync(data)` | *Not applicable* | `Promise<T>` | - | ❌ Go is synchronous by design |
| - | `.StrictParse(data)` | - | `T -> (T, error)` | ✅ **Go enhancement**: Compile-time type safety |
| - | `.MustParse(data)` | - | `any -> T` (panics) | ✅ **Go enhancement**: Panic-based parsing for flexible input |
| - | `.MustStrictParse(data)` | - | `T -> T` (panics) | ✅ **Go enhancement**: Type-safe panic parsing |

### Validation and Transform Method Mapping

| TypeScript Zod v4 | GoZod Method | TypeScript Signature | GoZod Signature | Status |
|-------------------|-------------|---------------------|----------------|--------|
| `.refine(fn, message?)` | `.Refine(fn, params?)` | `(val: T) => boolean` | `func(T) bool` + `core.SchemaParams` | ✅ Fully implemented |
| `.transform(fn)` | `.Transform(fn)` | `(val: T) => U` | `func(T, *core.RefinementContext) (any, error)` | ✅ Go error handling |
| `.pipe(schema)` | `.Pipe(schema)` | `ZodTypeAny` | `core.ZodType[any]` | ✅ Fully implemented |
| - | `.Check(fn)` | - | `func(T, *core.ParsePayload)` | ✅ **Go enhancement**: Multi-issue validation |
| - | `.Overwrite(fn)` | - | `func(T) T` | ✅ **Go enhancement**: In-place transformation |

### Metadata Method Mapping

| TypeScript Zod v4 | GoZod Method | Status | Description |
|-------------------|-------------|--------|-------------|
| `.describe(desc)` | `.Describe(desc)` | ✅ | Instance method on all 26 schema types |
| `z.describe(desc)` | `gozod.Describe(desc)` | ✅ | Check factory for use with `.Check()` |
| `.meta(meta)` | `.Meta(meta)` | ✅ | Instance method on all 26 schema types |
| `z.meta(meta)` | `gozod.Meta(meta)` | ✅ | Check factory for use with `.Check()` |
| `z.fromJSONSchema(schema)` | `gozod.FromJSONSchema(schema)` | ✅ | Create GoZod schema from JSON Schema (supports prefixItems → Tuple) |

### Coercion Method Mapping

| TypeScript Zod v4 | GoZod Method | Status | Description |
|-------------------|-----------------|--------|-------------|
| `z.coerce.string()` | `coerce.String()` | ✅ | Force string conversion from int, float, bool, etc. |
| `z.coerce.number()` | `coerce.Number()`, `coerce.Float64()` | ✅ | Force numeric conversion from string, bool, etc. |
| `z.coerce.boolean()` | `coerce.Bool()` | ✅ | Force boolean conversion from string, number, etc. |
| `z.coerce.bigint()` | `coerce.BigInt()` | ✅ | Force big integer conversion from string, number, etc. |
| `z.coerce.date()` | `coerce.Time()` | ✅ | Force time.Time conversion from string, Unix timestamp, etc. |
| - | All Go numeric types | ✅ | **Go-specific**: `coerce.Int()`, `coerce.Float32()`, etc. |

### String Validation Method Mapping

| TypeScript Zod v4 | GoZod Method | Status | Go-Specific Enhancements |
|-------------------|-------------|--------|-----------------------|
| `.min(n)` | `.Min(n)` | ✅ | Unicode-aware length counting |
| `.max(n)` | `.Max(n)` | ✅ | Unicode-aware length counting |
| `.length(n)` | `.Length(n)` | ✅ | Exact length validation |
| `.email()` | `.Email()` | ✅ | RFC 5322 compliant email validation |
| `.url()` | `.URL()` | ✅ | Full URL validation with scheme support |
| `.uuid()` | `.UUID()` | ✅ | All UUID versions (v1-v5) support |
| `.regex(pattern)` | `.Regex(pattern)` | ✅ | Go regex engine with pre-compilation |
| `.includes(str)` | `.Includes(str)` | ✅ | Substring validation |
| `.startsWith(str)` | `.StartsWith(str)` | ✅ | Prefix validation |
| `.endsWith(str)` | `.EndsWith(str)` | ✅ | Suffix validation |
| `.trim()` | `.Trim()` | ✅ | Whitespace trimming transformation |
| `.toLowerCase()` | `.ToLower()` | ✅ | Unicode-aware case conversion |
| `.toUpperCase()` | `.ToUpper()` | ✅ | Unicode-aware case conversion |
| `.lowercase()` | `.Lowercase()` | ✅ | Validates string has no uppercase letters |
| `.uppercase()` | `.Uppercase()` | ✅ | Validates string has no lowercase letters |
| `.normalize(form?)` | `.Normalize(form?)` | ✅ | Unicode normalization (NFC, NFD, NFKC, NFKD) |

### Network & Format Type Mapping

| TypeScript Zod v4 | GoZod Constructor | Status | Description |
|-------------------|------------------|--------|-------------|
| `z.ipv4()` | `gozod.IPv4()` | ✅ | IPv4 address validation |
| `z.ipv6()` | `gozod.IPv6()` | ✅ | IPv6 address validation |
| `z.hostname()` | `gozod.Hostname()` | ✅ | DNS hostname validation (1-253 chars) |
| `z.mac()` | `gozod.MAC()` | ✅ | MAC address validation (colon/hyphen/dot separators) |
| `z.e164()` | `gozod.E164()` | ✅ | E.164 phone number validation |
| `z.cidr()` | `gozod.CIDRv4()`, `gozod.CIDRv6()` | ✅ | IPv4/IPv6 CIDR notation validation |
| `z.guid()` | `gozod.Guid()` | ✅ | GUID format validation (8-4-4-4-12 hex pattern) |
| - | `gozod.HTTPURL()` | ✅ | **Go enhancement**: HTTP/HTTPS URL only |
| - | `gozod.Hex()` | ✅ | **Go enhancement**: Hexadecimal string validation |

### Hash Validation Checks

| Check Function | Description | Status |
|----------------|-------------|--------|
| `checks.MD5()` | MD5 hash validation (32 hex chars) | ✅ |
| `checks.SHA1()` | SHA1 hash validation (40 hex chars) | ✅ |
| `checks.SHA256()` | SHA256 hash validation (64 hex chars) | ✅ |
| `checks.SHA384()` | SHA384 hash validation (96 hex chars) | ✅ |
| `checks.SHA512()` | SHA512 hash validation (128 hex chars) | ✅ |

### Numeric Validation Method Mapping

| TypeScript Zod v4 | GoZod Method | Status | Go-Specific Features |
|-------------------|-------------|--------|-------------------|
| `.min(n)` | `.Min(n)` | ✅ | Works with all Go numeric types |
| `.max(n)` | `.Max(n)` | ✅ | Works with all Go numeric types |
| `.gt(n)` | `.Gt(n)` | ✅ | Greater than validation |
| `.gte(n)` | `.Gte(n)` | ✅ | Greater than or equal validation |
| `.lt(n)` | `.Lt(n)` | ✅ | Less than validation |
| `.lte(n)` | `.Lte(n)` | ✅ | Less than or equal validation |
| `.positive()` | `.Positive()` | ✅ | Positive number validation |
| `.negative()` | `.Negative()` | ✅ | Negative number validation |
| `.nonnegative()` | `.NonNegative()` | ✅ | Non-negative number validation |
| `.nonpositive()` | `.NonPositive()` | ✅ | Non-positive number validation |
| `.multipleOf(n)` | `.MultipleOf(n)` | ✅ | Multiple validation |
| `.int()` | Built into integer types | ✅ | Native Go integer validation |
| `.finite()` | `.Finite()` | ✅ | Finite number validation (no NaN/Inf) |
| `.safe()` | `.Safe()` | ✅ | Safe number range validation |

### Array/Slice Validation Method Mapping

| TypeScript Zod v4 | GoZod Method | Status | Go-Specific Features |
|-------------------|-------------|--------|-------------------|
| `.min(n)` | `.Min(n)` | ✅ | Minimum length validation |
| `.max(n)` | `.Max(n)` | ✅ | Maximum length validation |
| `.length(n)` | `.Length(n)` | ✅ | Exact length validation |
| `.nonempty()` | `.NonEmpty()` | ✅ | Non-empty validation |

### Object Validation Method Mapping

| TypeScript Zod v4 | GoZod Method | Status | Go-Specific Features |
|-------------------|-------------|--------|-------------------|
| `.pick({...})` | `.Pick([...])` | ✅ | Returns `(*ZodObject, error)` - errors if schema has refinements |
| - | `.MustPick([...])` | ✅ | **Go-specific**: Panics on error |
| `.omit({...})` | `.Omit([...])` | ✅ | Returns `(*ZodObject, error)` - errors if schema has refinements |
| - | `.MustOmit([...])` | ✅ | **Go-specific**: Panics on error |
| `.partial()` | `.Partial()` | ✅ | All fields optional |
| `.required()` | `.Required([...])` | ✅ | Specified fields required |
| `.extend({...})` | `.Extend({...})` | ✅ | Returns `(*ZodObject, error)` - errors if schema has refinements |
| - | `.SafeExtend({...})` | ✅ | **Go-specific**: Extends without refinement check |
| `.merge(schema)` | `.Merge(schema)` | ✅ | Schema merging with conflict resolution |
| `.keyof()` | `.Keyof()` | ✅ | Key type extraction |
| `.strict()` | Default behavior | ✅ | Objects are strict by default |
| `.strip()` | `.Strip()` | ✅ | Strip unknown fields |
| `.passthrough()` | `.Passthrough()` | ✅ | Pass through unknown fields |
| `.catchall(schema)` | `.Catchall(schema)` | ✅ | Validate unknown fields |

## 🏷️ Struct Tag Validation System

GoZod's declarative struct tag system provides a unique approach to validation through Go struct tags. Each tag rule directly corresponds to programmatic API methods:

### Core Tag Rules

| Struct Tag Rule | Programmatic API Equivalent | Example | Status |
|-----------------|----------------------------|---------|--------|
| `gozod:"required"` | Field without `.Optional()` | `Name string \`gozod:"required"\`` | ✅ Implemented |
| `gozod:"min=N"` | `.Min(N)` | `Name string \`gozod:"min=2"\`` | ✅ Implemented |
| `gozod:"max=N"` | `.Max(N)` | `Name string \`gozod:"max=50"\`` | ✅ Implemented |
| `gozod:"length=N"` | `.Length(N)` | `Code string \`gozod:"length=4"\`` | ✅ Implemented |
| `gozod:"email"` | `.Email()` | `Email string \`gozod:"email"\`` | ✅ Implemented |
| `gozod:"url"` | `.URL()` | `Website string \`gozod:"url"\`` | ✅ Implemented |
| `gozod:"uuid"` | `.UUID()` | `ID string \`gozod:"uuid"\`` | ✅ Implemented |
| `gozod:"regex=pattern"` | `.Regex(pattern)` | `Code string \`gozod:"regex=^[A-Z0-9]+$"\`` | ✅ Implemented |
| `gozod:"positive"` | `.Positive()` | `Amount int \`gozod:"positive"\`` | ✅ Implemented |
| `gozod:"negative"` | `.Negative()` | `Debt int \`gozod:"negative"\`` | ✅ Implemented |
| `gozod:"nonnegative"` | `.NonNegative()` | `Count int \`gozod:"nonnegative"\`` | ✅ Implemented |
| `gozod:"nonpositive"` | `.NonPositive()` | `Balance int \`gozod:"nonpositive"\`` | ✅ Implemented |
| `gozod:"nonempty"` | `.NonEmpty()` | `Tags []string \`gozod:"nonempty"\`` | ✅ Implemented |
| `gozod:"-"` | Field exclusion | `Internal string \`gozod:"-"\`` | ✅ Implemented |

### Advanced Tag Features

| Feature | Syntax | Example | Status |
|---------|--------|---------|--------|
| **Multiple rules** | Comma-separated | `\`gozod:"required,min=2,max=50"\`` | ✅ Implemented |
| **Numeric constraints** | `min=N`, `max=N`, `gt=N`, `lt=N` | `\`gozod:"min=0,max=120"\`` | ✅ Implemented |
| **Custom validators** | Custom validator name | `\`gozod:"required,unique_username"\`` | ✅ Implemented |
| **Parameterized validators** | `validator=param` | `\`gozod:"min_age=21"\`` | ✅ Implemented |
| **JSON field mapping** | Works with `json` tags | `Field string \`json:"field_name" gozod:"required"\`` | ✅ Implemented |

### Custom Validator Integration

```go
// Register custom validator
validators.Register(&UniqueUsernameValidator{})

// Use in struct tags
type User struct {
    Username string `gozod:"required,unique_username"`        // Basic validator
    Age      int    `gozod:"required,min_age=21"`           // Parameterized validator
    Email    string `gozod:"required,email"`                // Built-in validator
}

// Generate schema from tags
schema := gozod.FromStruct[User]()
result, err := schema.Parse(user)
```

### Automatic Circular Reference Handling

```go
type User struct {
    Name    string  `gozod:"required,min=2"`
    Email   string  `gozod:"required,email"`
    Friends []*User `gozod:"max=10"`           // Circular reference
}

// GoZod automatically detects and handles circular references
schema := gozod.FromStruct[User]()
result, err := schema.Parse(user) // ✅ No stack overflow
```

## 🎨 Custom Validator System

GoZod provides a flexible custom validator system with registry and struct tag integration:

### Validator Interfaces

```go
// Basic validator
type ZodValidator[T any] interface {
    Name() string
    Validate(value T) bool
    ErrorMessage(ctx *core.ParseContext) string
}

// Parameterized validator
type ZodParameterizedValidator[T any] interface {
    ZodValidator[T]
    ValidateParam(value T, param string) bool
    ErrorMessageWithParam(ctx *core.ParseContext, param string) string
}
```

### Usage Pattern

```go
// 1. Implement validator
type EmailValidator struct{}
func (v *EmailValidator) Name() string { return "custom_email" }
func (v *EmailValidator) Validate(email string) bool { /* validation logic */ }
func (v *EmailValidator) ErrorMessage(ctx *core.ParseContext) string { return "Invalid email" }

// 2. Register validator
validators.Register(&EmailValidator{})

// 3. Use in struct tags or programmatically
type User struct {
    Email string `gozod:"required,custom_email"`
}
```

## GoZod Unique Features

### Complete Strict Type Semantics

GoZod enforces strict type requirements with no automatic conversions:

```go
// Value schemas - require exact value types
stringSchema := gozod.String()      // Only accepts string
intSchema := gozod.Int()            // Only accepts int
structSchema := gozod.Struct[T]()   // Only accepts T

// Pointer schemas - require exact pointer types
stringPtrSchema := gozod.StringPtr()    // Only accepts *string
intPtrSchema := gozod.IntPtr()          // Only accepts *int
structPtrSchema := gozod.StructPtr[T]() // Only accepts *T

// No automatic conversions
value := "hello"
// stringSchema.Parse(&value)     // ❌ Error: requires string, got *string
// stringPtrSchema.Parse(value)   // ❌ Error: requires *string, got string

// Use Optional/Nilable for flexible input
optionalSchema := gozod.String().Optional()  // Flexible input, *string output
```

### Parse vs StrictParse Duality

GoZod provides both runtime flexibility and compile-time type safety:

```go
schema := gozod.String().Min(3)

// Parse - Runtime type checking (flexible input)
result, err := schema.Parse("hello")        // ✅ Works with any input type
result, err = schema.Parse(42)              // ❌ Runtime error: invalid type

// StrictParse - Compile-time type safety (optimal performance)
str := "hello"
result, err = schema.StrictParse(str)       // ✅ Compile-time guarantee
// result, err := schema.StrictParse(42)    // ❌ Compile-time error
```

### Performance Features

| Feature | Description | Benefit |
|---------|-------------|---------|
| **StrictParse** | Compile-time type checking | Eliminates runtime type assertions |
| **Code Generation** | Zero-reflection validation | 5-10x performance improvement |
| **Pointer Identity** | Preserves original pointers | Minimizes memory allocations |
| **Pre-compiled Regex** | Regex patterns cached | Faster validation execution |

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

**GoZod Programmatic:**
```go
import "github.com/kaptinlin/gozod"

// Struct-based validation
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
    return err
}

// Panic-based parsing (equivalent to TS .parse())
user := userSchema.MustParse(data)

// Type-safe parsing with exact input types
validUser, err := userSchema.StrictParse(existingUser)
```

**GoZod Struct Tags:**
```go
type User struct {
    Name  string `json:"name" gozod:"required,min=2"`
    Age   *int   `json:"age" gozod:"min=0"`             // Optional by default
    Email string `json:"email" gozod:"required,email"`
}

// Generate schema from tags
schema := gozod.FromStruct[User]()
user, err := schema.Parse(data)
```

### Default vs Prefault Behavior (Zod v4 Compatible)

```go
// Default: Short-circuits validation (Zod v4 behavior)
defaultSchema := gozod.String().Min(5).Default("fallback")
result, _ := defaultSchema.Parse(nil)  // "fallback" (bypasses Min validation)

// Prefault: Goes through full validation pipeline (Zod v4 behavior)
prefaultSchema := gozod.String().Min(5).Prefault("fallback")
result, _ := prefaultSchema.Parse(nil)  // "fallback" (validates "fallback" >= 5)
```

### Optional vs Nilable Semantics

```go
// Optional: Missing field semantics (returns generic nil)
optionalResult, _ := gozod.String().Optional().Parse(nil) // nil

// Nilable: Null value semantics (returns typed nil pointer)
nilableResult, _ := gozod.String().Nilable().Parse(nil)   // (*string)(nil)

// Nullish: Both missing and null (returns typed nil pointer)
nullishResult, _ := gozod.String().Nullish().Parse(nil)   // (*string)(nil)
```

## Features Not Applicable in Go

| TypeScript Zod v4 Feature | Reason Not Applicable | GoZod Alternative |
|---------------------------|----------------------|------------------|
| **Async validation** | Go is synchronous by design | Use goroutines for concurrent validation |
| **Void type** | Go has no void concept | Functions return specific types or nothing |
| **Undefined type** | Go has no undefined | Use `nil` or zero values |
| **Branded types** | Different type system | Use custom Go types: `type UserId string` |
| **Type inference** | Language-level feature | Go generics provide compile-time type safety |
| **Readonly** | Language-level feature | Use immutability patterns |
| **Preprocessor** | Not needed | Use `.Transform()` and `.Pipe()` |
| **SafeParse result** | Different error handling | Go uses `(T, error)` pattern |

## Go-Specific Enhancements

| GoZod Feature | Description | TypeScript Equivalent |
|---------------|-------------|----------------------|
| **Struct validation** | Native Go struct support | No direct equivalent |
| **🏷️ Struct tags** | Declarative validation with `gozod:"required,min=2"` | No equivalent |
| **Pointer type safety** | Explicit pointer vs value validation | Union types |
| **All Go numeric types** | int8, uint32, complex64, etc. | Limited to number/bigint |
| **Parse vs StrictParse** | Runtime vs compile-time type checking | Compile-time only |
| **Panic-based parsing** | `MustParse()` methods | Throwing behavior |
| **Custom validator system** | Registry-based with struct tag integration | Limited to refinements |
| **Circular reference handling** | Automatic detection with lazy evaluation | Manual lazy schemas |
| **JSON tag mapping** | Automatic field mapping | Manual specification |
| **Apply function** | `gozod.Apply(schema, fn)` for composable schema modifiers | No equivalent |
| **Tuple with Rest** | `gozod.TupleWithRest([items], rest)` for variadic tuples | `z.tuple([...]).rest(schema)` |
| **LooseRecord** | `gozod.LooseRecord(keySchema, valueSchema)` passes non-matching keys | `z.looseRecord()` |
| **Map NonEmpty** | `gozod.Map(...).NonEmpty()` ensures at least one entry | `z.map().nonempty()` |
| **SafeExtend** | `.SafeExtend()` extends object without refinement check | No equivalent |
| **And/Or methods** | `.And()` and `.Or()` fluent methods for schema composition | Use `z.intersection()` / `z.union()` |
| **Hostname validation** | `gozod.Hostname()` DNS hostname validation | `z.hostname()` |
| **MAC validation** | `gozod.MAC()` MAC address validation | `z.mac()` |
| **E164 validation** | `gozod.E164()` E.164 phone validation | `z.e164()` |
| **HTTPURL** | `gozod.HTTPURL()` HTTP/HTTPS only URL validation | No direct equivalent |
| **Hex validation** | `gozod.Hex()` hexadecimal string validation | No direct equivalent |
| **Hash checks** | `checks.MD5()`, `checks.SHA256()`, etc. | No direct equivalent |

## Performance Comparison

### Parse vs StrictParse Performance

| Method | Input Type | Performance | Use Case |
|--------|------------|-------------|----------|
| `Parse()` | `any` | Standard | Unknown input types, API validation |
| `StrictParse()` | `T` | 3-5x faster | Known types, internal validation |
| `MustParse()` | `any` | Standard + panic overhead | Critical failures |
| `MustStrictParse()` | `T` | 3-5x faster + panic overhead | Type-safe critical failures |

### Code Generation Benefits

| Feature | Reflection-based | Generated | Improvement |
|---------|------------------|-----------|-------------|
| **Performance** | Baseline | 5-10x faster | Eliminates reflection overhead |
| **Memory usage** | Baseline | 50-70% reduction | Pre-compiled validation |
| **Type safety** | Runtime | Compile-time | Earlier error detection |

---

This comprehensive mapping demonstrates GoZod's complete compatibility with TypeScript Zod v4 while providing significant Go-specific enhancements including strict type semantics, performance optimizations, declarative struct tags, and a robust custom validator system.
# GoZod API Reference

Complete API documentation for GoZod - a TypeScript Zod v4-inspired validation library for Go.

## 🎯 Overview

GoZod provides comprehensive data validation with:
- **Type Safety**: Full Go generics support with preserved type information
- **Complete Strict Type Semantics**: All methods require exact input types - no automatic conversions
- **Maximum Performance**: Optimized validation pipeline with zero-overhead type handling
- **Composable Schemas**: Chain validations, transformations, and type conversions
- **Rich Validation**: Built-in validators for strings, numbers, objects, arrays, and more
- **Flexible Modifiers**: Optional, Nilable, Default, and Prefault handling for complex scenarios
- **Advanced Types**: Union, Intersection, and Discriminated Union support
- **Custom Validation**: Use `.Refine()` and `.Check()` for custom validation logic
- **Struct Tags**: Declarative validation with `gozod:"required,min=2,email"` syntax

## 🔧 Core Concepts

### Parse vs StrictParse Methods

GoZod provides two parsing approaches for maximum flexibility and performance:

```go
// Parse - Runtime type checking (flexible input)
schema := gozod.String()
result, err := schema.Parse("hello")        // ✅ Works with any input type
result, err = schema.Parse(42)              // ❌ Runtime error: invalid type

// StrictParse - Compile-time type safety (optimal performance)
str := "hello"
result, err := schema.StrictParse(str)      // ✅ Compile-time guarantee
// result, err := schema.StrictParse(42)    // ❌ Compile-time error
```

### Complete Strict Type Semantics

```go
// Value schemas require exact value types
stringSchema := gozod.String()
result, _ := stringSchema.Parse("hello")     // ✅ string → string
// result, _ := stringSchema.Parse(&str)     // ❌ Error: requires string, got *string

// Pointer schemas require exact pointer types
stringPtrSchema := gozod.StringPtr()
result, _ := stringPtrSchema.Parse(&str)     // ✅ *string → *string
// result, _ := stringPtrSchema.Parse("hello") // ❌ Error: requires *string, got string

// For flexible input, use Optional/Nilable modifiers
optionalSchema := gozod.String().Optional()  // Flexible input, *string output
result, _ := optionalSchema.Parse("hello")   // ✅ string → *string (new pointer)
result, _ = optionalSchema.Parse(&str)       // ✅ *string → *string (preserves identity)
```

### Validation Pipeline

```go
// Input → [Coercion] → [Validation] → [Transformation] → Output
schema := gozod.String().
    Min(3).
    Overwrite(strings.ToUpper).
    Transform(func(s string, ctx *core.RefinementContext) (any, error) {
        return fmt.Sprintf("Result: %s", s), nil
    })
```

### Custom Validation

```go
// Single custom validation with Refine
schema := gozod.String().Refine(func(s string) bool {
    return len(s) > 5
}, "String too short")

// Multiple business rules with Check
schema = gozod.Int().Check(func(v int, p *core.ParsePayload) {
    if v%2 != 0 {
        p.AddIssueWithMessage("number must be even")
    }
    if v < 0 {
        p.AddIssueWithCode(core.TooSmall, "number must be positive")
    }
})
```

---

## 📋 Table of Contents

### Core Concepts
- [Core Types](#-core-types)
- [Parse Methods](#-parse-methods)
- [Modifiers](#-modifiers)

### Primitive Types
- [String Validation](#-string-validation)
- [Number Validation](#-number-validation)
- [Boolean Validation](#-boolean-validation)
- [Time Validation](#-time-validation)

### Complex Types
- [Object Validation](#-object-validation)
- [Struct Validation](#-struct-validation)
- [Array/Slice Validation](#-arrayslice-validation)
- [Map Validation](#-map-validation)

### Advanced Types
- [Union Types](#-union-types)
- [Intersection Types](#-intersection-types)
- [Lazy Types](#-lazy-types)
- [Transform Types](#-transform-types)

### Custom Validation
- [Custom Validation](#-custom-validation)
- [Struct Tags](#-struct-tags)

---

## 🧱 Core Types

### ZodType Interface

All GoZod schemas implement the core `ZodType[T]` interface:

```go
type ZodType[T any] interface {
    Parse(input any, ctx ...*ParseContext) (T, error)
    StrictParse(input T, ctx ...*ParseContext) (T, error)  
    MustParse(input any, ctx ...*ParseContext) T
    MustStrictParse(input T, ctx ...*ParseContext) T
    ParseAny(input any, ctx ...*ParseContext) (any, error)
    GetInternals() *ZodTypeInternals
    IsOptional() bool
    IsNilable() bool
}
```

### Constructor Pattern

Every type provides value and pointer constructors:

```go
// Value constructors (return value types)
String()    // ZodString[string]
Int()       // ZodInteger[int]
Bool()      // ZodBoolean[bool]

// Pointer constructors (return pointer types)  
StringPtr() // ZodString[*string]
IntPtr()    // ZodInteger[*int]
BoolPtr()   // ZodBoolean[*bool]
```

---

## 🔍 Parse Methods

### Parse(input any) - Runtime Type Checking

Accepts any input type and performs runtime validation:

```go
schema := gozod.String().Min(3)

// Runtime type checking
result, err := schema.Parse("hello")    // ✅ Valid
result, err = schema.Parse(42)          // ❌ Runtime error
result, err = schema.Parse(nil)         // ❌ Runtime error (unless Optional/Nilable)
```

### StrictParse(input T) - Compile-Time Type Safety

Requires exact input type matching for optimal performance:

```go
schema := gozod.String().Min(3)

// Compile-time type safety
str := "hello"
result, err := schema.StrictParse(str)  // ✅ Optimal performance
// result, err := schema.StrictParse(42) // ❌ Compile-time error
```

### MustParse/MustStrictParse - Panic on Error

Convenience methods that panic instead of returning errors:

```go
schema := gozod.String().Min(3)

// Panic-based versions
result := schema.MustParse("hello")         // ✅ Returns "hello"
result = schema.MustStrictParse("hello")    // ✅ Returns "hello" 
// result := schema.MustParse(42)           // ⚠️ Panics
```

---

## 🎛️ Modifiers

### Optional() - Makes Field Optional

```go
schema := gozod.String().Optional()  // Returns ZodString[*string]

result, _ := schema.Parse("hello")   // ✅ "hello" → &"hello" 
result, _ = schema.Parse(nil)        // ✅ nil → nil
result, _ = schema.Parse(undefined)  // ✅ undefined → nil
```

### Nilable() - Allows nil Values

```go
schema := gozod.String().Nilable()   // Returns ZodString[*string]

result, _ := schema.Parse("hello")   // ✅ "hello" → &"hello"
result, _ = schema.Parse(nil)        // ✅ nil → nil
// result, _ = schema.Parse(undefined) // ❌ undefined still invalid
```

### Nullish() - Optional + Nilable

```go
schema := gozod.String().Nullish()   // Returns ZodString[*string]

result, _ := schema.Parse("hello")   // ✅ "hello" → &"hello"
result, _ = schema.Parse(nil)        // ✅ nil → nil
result, _ = schema.Parse(undefined)  // ✅ undefined → nil
```

### Default() - Provides Default Values

```go
// Default bypasses validation pipeline (Zod v4 compatible)
schema := gozod.String().Min(5).Default("default")

result, _ := schema.Parse("hello world") // ✅ "hello world" (passes validation)
result, _ = schema.Parse("hi")           // ❌ Error: too small
result, _ = schema.Parse(nil)            // ✅ "default" (bypasses validation)
```

### Prefault() - Preprocessing Fallback

```go
// Prefault goes through full validation pipeline (Zod v4 compatible)
schema := gozod.String().Min(5).Prefault("fallback")

result, _ := schema.Parse("hello world") // ✅ "hello world" 
result, _ = schema.Parse("hi")           // ❌ Error: too small
result, _ = schema.Parse(nil)            // ✅ "fallback" (validates "fallback")
```

---

## 🔤 String Validation

### Basic String Schema

```go
// Constructors
stringSchema := gozod.String()          // ZodString[string]
stringPtrSchema := gozod.StringPtr()    // ZodString[*string]

// Usage with exact type requirements
result, err := stringSchema.Parse("hello")        // ✅ string → string
result, err = stringPtrSchema.Parse(&str)         // ✅ *string → *string

// StrictParse for known types
str := "hello"
result, err = stringSchema.StrictParse(str)       // ✅ Optimal performance
```

### Length Validation

```go
schema := gozod.String().
    Min(3).              // Minimum length
    Max(50).             // Maximum length  
    Length(10)           // Exact length

result, err := schema.Parse("hello")     // Length: 5, Min: 3 ✅
result, err = schema.Parse("hi")         // Length: 2, Min: 3 ❌
```

### Format Validation

```go
// Basic string formats  
gozod.Email()                              // Email validation
gozod.URL()                                // URL validation  

// Network addresses (types/network.go)
gozod.IPv4()                               // IPv4: "192.168.1.1"
gozod.IPv6()                               // IPv6: "2001:db8::8a2e:370:7334"
gozod.Hostname()                           // DNS hostname: "example.com"
gozod.MAC()                                // MAC: "00:1A:2B:3C:4D:5E" (default ":")
gozod.MAC("-")                             // MAC: "00-1a-2b-3c-4d-5e"
gozod.E164()                               // E.164 phone: "+14155552671"
gozod.CIDRv4()                             // IPv4 CIDR: "192.168.1.0/24"
gozod.CIDRv6()                             // IPv6 CIDR: "2001:db8::/32"
gozod.HTTPURL()                            // HTTP/HTTPS URL only

// ISO 8601 date/time formats (types/iso.go)
gozod.IsoDate()                            // Date: "2024-12-06"
gozod.IsoTime()                            // Time: "15:30:00"
gozod.IsoDateTime()                        // DateTime: "2024-12-06T15:30:00Z"
gozod.IsoDuration()                        // Duration: "P1Y2M3D", "PT1H30M"

// ISO time with precision
precision := -1  // HH:MM only
gozod.IsoTime(types.IsoTimeOptions{Precision: &precision})

// Unique identifiers (types/ids.go)
gozod.Uuid()                               // UUID format
gozod.Uuid("v4")                           // UUID v4 specific
gozod.Cuid()                               // CUID v1
gozod.Cuid2()                              // CUID v2 
gozod.Ulid()                               // ULID format
gozod.Nanoid()                             // NanoID format

// Text encodings (types/text.go)
gozod.Hex()                                // Hexadecimal string
gozod.Base64()                             // Base64 encoding
gozod.Base64URL()                          // URL-safe Base64
gozod.Emoji()                              // Emoji characters

// Tokens (types/text.go)
gozod.JWT()                                // JWT token structure
gozod.JWT(types.JWTOptions{Algorithm: "HS256"})  // JWT with algorithm

// For custom patterns, still use String
gozod.String().Regex(`^\d{3}-\d{2}-\d{4}$`)
```

### String Transformations

```go
// Built-in transformations
schema := gozod.String().
    Trim().              // Remove whitespace
    ToLower().           // Convert to lowercase
    ToUpper().           // Convert to uppercase
    Slugify()            // URL-friendly: "Hello World!" → "hello-world"

result, err := schema.Parse("  HELLO WORLD  ")  // Result: "hello-world"

// Custom transformation
transformSchema := gozod.String().Transform(func(s string, ctx *core.RefinementContext) (any, error) {
    return strings.ReplaceAll(s, " ", "_"), nil
})
result, err = transformSchema.Parse("hello world")  // Result: "hello_world"
```

### Metadata

```go
// Add description
schema := gozod.Email().Describe("User's email address")

// Add rich metadata
schema = gozod.Uuid().Meta(core.GlobalMeta{
    ID:          "user_id",
    Title:       "User ID",
    Description: "Unique identifier",
    Examples:    []any{"550e8400-e29b-41d4-a716-446655440000"},
})
```

---

## 🔢 Number Validation

### Integer Types

```go
// Different integer types
int8Schema := gozod.Int8()      // ZodInteger[int8]
int16Schema := gozod.Int16()    // ZodInteger[int16] 
int32Schema := gozod.Int32()    // ZodInteger[int32]
int64Schema := gozod.Int64()    // ZodInteger[int64]
intSchema := gozod.Int()        // ZodInteger[int]

// Pointer variants
intPtrSchema := gozod.IntPtr()  // ZodInteger[*int]
```

### Range Validation

```go
schema := gozod.Int().
    Min(0).              // Minimum value (inclusive)
    Max(100).            // Maximum value (inclusive)
    Positive().          // Must be > 0
    NonNegative().       // Must be >= 0
    Negative().          // Must be < 0
    NonPositive()        // Must be <= 0

result, err := schema.Parse(50)   // ✅ Within range
result, err = schema.Parse(-1)    // ❌ Below minimum
```

### Floating Point Types

```go
// Float types
float32Schema := gozod.Float32()  // ZodFloat[float32]
float64Schema := gozod.Float64()  // ZodFloat[float64]

// Validation
floatSchema := gozod.Float64().
    Min(0.0).
    Max(100.0).
    Finite()             // No NaN or Inf

result, err := floatSchema.Parse(3.14)     // ✅
result, err = floatSchema.Parse(math.NaN()) // ❌ Not finite
```

---

## ✅ Boolean Validation

```go
// Boolean schema
boolSchema := gozod.Bool()       // ZodBoolean[bool]
boolPtrSchema := gozod.BoolPtr() // ZodBoolean[*bool]

// Basic usage
result, err := boolSchema.Parse(true)      // ✅ true
result, err = boolSchema.Parse(false)     // ✅ false

// With StrictParse
b := true
result, err = boolSchema.StrictParse(b)    // ✅ Optimal performance
```

---

## ⏰ Time Validation

```go
// Time schema
timeSchema := gozod.Time()       // ZodTime[time.Time]
timePtrSchema := gozod.TimePtr() // ZodTime[*time.Time]

// Basic validation
now := time.Now()
result, err := timeSchema.Parse(now)       // ✅

// Date range validation
schema := gozod.Time().
    After(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)).
    Before(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))

result, err = schema.Parse(time.Now())     // ✅ If within range
```

---

## 📦 Object Validation

Object schemas validate `map[string]any` data (like JSON):

```go
// Object schema for map[string]any
userSchema := gozod.Object(gozod.ObjectSchema{
    "name":  gozod.String().Min(2),
    "age":   gozod.Int().Min(0),
    "email": gozod.Email().Optional(),
})

// Validate JSON-like data
data := map[string]any{
    "name": "Alice",
    "age":  25,
    "email": "alice@example.com",
}

result, err := userSchema.Parse(data)      // ✅ Validated map[string]any
```

---

## 🏗️ Struct Validation

### Basic Struct Validation

```go
type User struct {
    Name  string `json:"name"`
    Age   int    `json:"age"`
    Email string `json:"email"`
}

// Basic struct schema
basicSchema := gozod.Struct[User]()
user := User{Name: "Alice", Age: 25, Email: "alice@example.com"}
result, err := basicSchema.Parse(user)     // ✅ Basic validation
```

### Struct with Field Validation

```go
// Struct schema with field validation
userSchema := gozod.Struct[User](gozod.StructSchema{
    "name":  gozod.String().Min(2).Max(50),
    "age":   gozod.Int().Min(0).Max(120), 
    "email": gozod.Email(),
})

validUser := User{Name: "Bob", Age: 30, Email: "bob@example.com"}
result, err := userSchema.Parse(validUser) // ✅ Full field validation
```

### Struct Tags (Declarative)

```go
type User struct {
    Name     string `gozod:"required,min=2,max=50"`
    Age      int    `gozod:"required,min=0,max=120"`
    Email    string `gozod:"required,email"`
    Bio      string `gozod:"max=500"`           // Optional by default
    Internal string `gozod:"-"`                // Skip validation
}

// Generate schema from tags
schema := gozod.FromStruct[User]()
result, err := schema.Parse(user)              // ✅ Tag-based validation
```

### Nested Structs and Circular References

```go
type User struct {
    Name    string  `gozod:"required,min=2"`
    Friends []*User `gozod:"max=10"`           // Circular reference
}

// Automatically handles circular references with lazy evaluation
schema := gozod.FromStruct[User]()
result, err := schema.Parse(user)              // ✅ No stack overflow
```

---

## 📋 Array/Slice Validation

### Basic Array Validation

```go
// Array of strings
stringArraySchema := gozod.Array(gozod.String())     // ZodArray[string]
result, err := stringArraySchema.Parse([]string{"a", "b", "c"}) // ✅

// Array of validated strings
validatedArraySchema := gozod.Array(gozod.String().Min(2))
result, err = validatedArraySchema.Parse([]string{"hello", "world"}) // ✅
result, err = validatedArraySchema.Parse([]string{"hello", "x"})     // ❌ "x" too short
```

### Slice Validation with Element Schemas

```go
// Slice with element validation
userSliceSchema := gozod.Slice(gozod.Struct[User]())

users := []User{
    {Name: "Alice", Age: 25},
    {Name: "Bob", Age: 30},
}
result, err := userSliceSchema.Parse(users)    // ✅ Validates each element
```

### Array Length Validation

```go
schema := gozod.Array(gozod.String()).
    Min(1).              // At least 1 element
    Max(10).             // At most 10 elements
    Length(5).           // Exactly 5 elements
    NonEmpty()           // Same as Min(1)

result, err := schema.Parse([]string{"a", "b", "c"})  // Length: 3 ✅
```

---

## 📐 Tuple Validation

Tuples represent fixed-length arrays with typed elements at each position:

```go
// Fixed positions with specific types
schema := gozod.Tuple(gozod.String(), gozod.Int(), gozod.Bool())
result, err := schema.Parse([]any{"hello", 42, true})
// result: []any{"hello", 42, true}

// Fails on wrong length
_, err = schema.Parse([]any{"hello"})
// Error: Tuple must have at least 3 element(s)

// Fails on wrong type
_, err = schema.Parse([]any{"hello", "not-int", true})
// Error: Invalid input: expected int, received string
```

### Tuple with Rest Element

```go
// Fixed positions + variadic rest elements
schema := gozod.TupleWithRest(
    []core.ZodSchema{gozod.String(), gozod.Int()},
    gozod.Bool(),  // additional elements must be booleans
)
result, err := schema.Parse([]any{"hello", 42, true, false, true})
// result: []any{"hello", 42, true, false, true}

// Chain with WithRest method
schema = gozod.Tuple(gozod.String(), gozod.Int()).WithRest(gozod.Bool())
```

### Tuple with Optional Elements

```go
// Optional elements at the end
schema := gozod.Tuple(
    gozod.String(),
    gozod.Int().Optional(),
)

// All elements provided
result, _ := schema.Parse([]any{"hello", 42})  // ✅

// Optional element omitted
result, _ = schema.Parse([]any{"hello"})       // ✅
```

### Tuple Methods

| Method | Description |
|--------|-------------|
| `.WithRest(schema)` | Add rest element schema for variadic elements |
| `.Min(n)` | Minimum length validation |
| `.Max(n)` | Maximum length validation |
| `.Length(n)` | Exact length validation |
| `.NonEmpty()` | At least one element |
| `.Optional()` | Returns `*[]any` |
| `.Describe(desc)` | Add description metadata |
| `.Meta(meta)` | Add rich metadata |

---

## 🗺️ Map Validation

```go
// Map validation
stringMapSchema := gozod.Map(gozod.String())  // map[string]string
result, err := stringMapSchema.Parse(map[string]string{
    "key1": "value1",
    "key2": "value2",
})  // ✅

// Map with validated values
userMapSchema := gozod.Map(gozod.Struct[User]())  // map[string]User
userMap := map[string]User{
    "alice": {Name: "Alice", Age: 25},
    "bob":   {Name: "Bob", Age: 30},
}
result, err = userMapSchema.Parse(userMap)         // ✅ Validates each value

// Map with NonEmpty validation
schema := gozod.Map(gozod.String()).NonEmpty()
_, err = schema.Parse(map[string]string{})         // ❌ Map must have at least 1 entry
```

### Record with Key Validation

```go
// Record validates both keys and values
schema := gozod.Record(
    gozod.String().Regex(`^[a-z]+$`),  // lowercase keys only
    gozod.Int(),
)
result, err := schema.Parse(map[string]any{"name": 42})  // ✅
_, err = schema.Parse(map[string]any{"Name": 42})        // ❌ key not lowercase

// LooseRecord passes through non-matching keys unchanged
looseSchema := gozod.LooseRecord(
    gozod.String().Regex(`^S_`),  // only validate keys starting with "S_"
    gozod.String(),
)
result, _ = looseSchema.Parse(map[string]any{
    "S_name": "John",
    "other": 123,  // passed through unchanged
})
// Result: {"S_name": "John", "other": 123}
```

---

## 🔀 Union Types

Union types accept one of several alternative schemas:

```go
// String or Int union
unionSchema := gozod.Union(
    gozod.String(),
    gozod.Int(),
)

result, err := unionSchema.Parse("hello")  // ✅ Matches string
result, err = unionSchema.Parse(42)        // ✅ Matches int  
result, err = unionSchema.Parse(true)      // ❌ No union member matched
```

### Discriminated Unions

```go
type Dog struct {
    Type string `json:"type"`  // "dog"
    Breed string `json:"breed"`
}

type Cat struct {
    Type  string `json:"type"`  // "cat" 
    Lives int    `json:"lives"`
}

// Discriminated union based on "type" field
animalSchema := gozod.DiscriminatedUnion("type", map[string]any{
    "dog": gozod.Struct[Dog](),
    "cat": gozod.Struct[Cat](),
})

dog := Dog{Type: "dog", Breed: "Golden Retriever"}
result, err := animalSchema.Parse(dog)     // ✅ Matches dog schema
```

### Xor (Exclusive Union)

Xor validates that **exactly one** schema matches. Unlike Union which succeeds when any option matches, Xor fails if zero or multiple options match:

```go
// Exactly one must match
schema := gozod.Xor([]any{
    gozod.String().Email(),
    gozod.String().URL(),
})

// ✅ Matches email only
result, err := schema.Parse("user@example.com")

// ✅ Matches URL only
result, err = schema.Parse("https://example.com")

// ❌ Matches neither
_, err = schema.Parse("invalid")
```

Xor schemas convert to JSON Schema `oneOf`:
```json
{
  "oneOf": [
    {"type": "string", "format": "email"},
    {"type": "string", "format": "uri"}
  ]
}
```

---

## ⚡ Intersection Types

Intersection types must satisfy all provided schemas:

```go
// String that satisfies both min length and regex
intersectionSchema := gozod.Intersection(
    gozod.String().Min(3),           // At least 3 chars
    gozod.String().Max(10),          // At most 10 chars  
    gozod.String().Regex(`^[a-z]+$`), // Only lowercase letters
)

result, err := intersectionSchema.Parse("hello")    // ✅ Satisfies all
result, err = intersectionSchema.Parse("hi")        // ❌ Too short
result, err = intersectionSchema.Parse("HELLO")     // ❌ Not lowercase
```

### And/Or Methods

All schema types support fluent `.And()` and `.Or()` methods for composing schemas:

```go
// And() creates an intersection - value must satisfy both schemas
schema := gozod.String().Min(3).And(gozod.String().Max(10))
result, _ := schema.Parse("hello")   // ✅ Satisfies both (len >= 3 AND len <= 10)

// Or() creates a union - value can satisfy either schema
schema = gozod.Int().Or(gozod.String())
result, _ = schema.Parse(42)         // ✅ Matches Int
result, _ = schema.Parse("hello")    // ✅ Matches String

// Chain multiple And/Or calls
schema = gozod.String().Min(1).And(gozod.String().Max(100)).Or(gozod.Nil())
```

**Supported schema types**: String, Integer, Float, Bool, Array, Slice, Object, Union, Intersection

---

## 🔄 Lazy Types

Lazy types enable recursive and circular references:

```go
// Self-referencing type
type Node struct {
    Value    string  `json:"value"`
    Children []*Node `json:"children"`
}

// Lazy schema for recursive validation
var nodeSchema gozod.ZodType[Node]
nodeSchema = gozod.Lazy(func() gozod.ZodType[Node] {
    return gozod.Struct[Node](gozod.StructSchema{
        "value":    gozod.String(),
        "children": gozod.Array(nodeSchema), // Recursive reference
    })
})

// Tree structure
tree := Node{
    Value: "root",
    Children: []*Node{
        {Value: "child1", Children: nil},
        {Value: "child2", Children: nil},
    },
}

result, err := nodeSchema.Parse(tree)      // ✅ Handles recursion
```

---

## 🎭 Transform Types

Transform types convert validated data to different types:

```go
// String to int transformation
transformSchema := gozod.String().
    Regex(`^\d+$`).  // Must be numeric string
    Transform(func(s string, ctx *core.RefinementContext) (any, error) {
        return strconv.Atoi(s)
    })

result, err := transformSchema.Parse("42")  // ✅ "42" → 42 (int)
result, err = transformSchema.Parse("abc")  // ❌ Invalid regex
```

---

## 🎨 Custom Validation

### Using Refine for Custom Logic

Use `.Refine()` for custom validation that returns a boolean:

```go
// Simple custom validation
schema := gozod.String().Refine(func(s string) bool {
    return len(s) > 5 && strings.Contains(s, "@")
}, "Invalid format")

result, err := schema.Parse("user@example.com")  // ✅
result, err = schema.Parse("hi")                  // ❌
```

### Multiple Custom Validations

Chain multiple `.Refine()` calls for complex validation:

```go
schema := gozod.String().
    Min(3).
    Refine(func(s string) bool {
        return !strings.Contains(s, " ")
    }, "No spaces allowed").
    Refine(func(s string) bool {
        return strings.ToLower(s) == s
    }, "Must be lowercase")

result, err := schema.Parse("hello")    // ✅
result, err = schema.Parse("Hello")     // ❌ Not lowercase
```

### Using Check for Detailed Validation

Use `.Check()` when you need to add multiple issues:

```go
schema := gozod.Int().Check(func(v int, p *core.ParsePayload) {
    if v%2 != 0 {
        p.AddIssueWithMessage("number must be even")
    }
    if v < 0 {
        p.AddIssueWithCode(core.TooSmall, "number must be positive")
    }
})
```

### Cross-Field Validation on Structs

```go
type PasswordForm struct {
    Password        string `gozod:"required,min=8"`
    ConfirmPassword string `gozod:"required"`
}

// Create schema with cross-field validation
schema := gozod.FromStruct[PasswordForm]().Refine(func(form PasswordForm) bool {
    return form.Password == form.ConfirmPassword
}, "Passwords must match")

form := PasswordForm{Password: "secret123", ConfirmPassword: "secret123"}
result, err := schema.Parse(form)  // ✅
```

---

## 🏷️ Struct Tags

### Tag Syntax

```go
type User struct {
    // Basic validation
    Name string `gozod:"required,min=2,max=50"`
    
    // Multiple rules
    Email string `gozod:"required,email,max=100"`
    
    // Optional field (default)
    Bio string `gozod:"max=500"`
    
    // Skip validation
    Internal string `gozod:"-"`
    
    // Custom validators
    Username string `gozod:"required,custom_username"`
    Age      int    `gozod:"required,min_age=18"`
}
```

### Available Tag Rules

#### String Rules
- `required` - Field must be present
- `min=N` - Minimum length
- `max=N` - Maximum length  
- `length=N` - Exact length
- `email` - Email format validation
- `url` - URL format validation
- `uuid` - UUID format validation
- `regex=pattern` - Custom regex pattern

#### Numeric Rules
- `min=N` - Minimum value
- `max=N` - Maximum value
- `positive` - Must be > 0
- `negative` - Must be < 0
- `nonnegative` - Must be >= 0
- `nonpositive` - Must be <= 0

#### Array/Slice Rules
- `min=N` - Minimum number of elements
- `max=N` - Maximum number of elements
- `length=N` - Exact number of elements
- `nonempty` - At least one element

### Complex Tag Examples

```go
type CompleteExample struct {
    // String validation
    Name     string   `gozod:"required,min=2,max=50"`
    Username string   `gozod:"required,regex=^[a-zA-Z0-9_]+$"`
    Email    string   `gozod:"required,email"`
    Website  string   `gozod:"url"`
    
    // Numeric validation  
    Age    int     `gozod:"required,min=18,max=120"`
    Score  float64 `gozod:"min=0,max=100"`
    
    // Array validation
    Tags   []string `gozod:"min=1,max=10"`
    
    // Optional fields
    Bio     string  `gozod:"max=500"`
    Avatar  *string `gozod:"url"`
    
    // Skip validation
    Internal string `gozod:"-"`
}

// Generate schema from tags
schema := gozod.FromStruct[CompleteExample]()
result, err := schema.Parse(example)
```

---

## 🛠️ Advanced Usage

### Error Handling

```go
schema := gozod.String().Min(5).Email()
_, err := schema.Parse("hi")

if zodErr, ok := err.(*issues.ZodError); ok {
    // Access structured error information
    for _, issue := range zodErr.Issues {
        fmt.Printf("Path: %v, Code: %s, Message: %s\n", 
            issue.Path, issue.Code, issue.Message)
    }
    
    // Pretty print errors
    fmt.Println(zodErr.PrettifyError())
    
    // Get flattened field errors
    fieldErrors := zodErr.FlattenError()
    fmt.Printf("Field errors: %+v\n", fieldErrors.FieldErrors)
}
```

### Performance Optimization

```go
// Use StrictParse for known input types
str := "hello@example.com"
schema := gozod.Email()
result, err := schema.StrictParse(str)  // Optimal performance

// Use code generation for struct validation
//go:generate gozodgen
type User struct {
    Name  string `gozod:"required,min=2"`
    Email string `gozod:"required,email"`
}

// Generated Schema() method provides zero-reflection validation
schema := gozod.FromStruct[User]()  // Uses generated code automatically
```

### Type Coercion

```go
// Enable coercion for flexible input handling
coercedSchema := gozod.String().Coerce()
result, err := coercedSchema.Parse(42)        // 42 → "42"
result, err = coercedSchema.Parse(true)       // true → "true"

// Time coercion
timeSchema := gozod.Time().Coerce()  
result, err = timeSchema.Parse(1609459200)    // Unix timestamp → time.Time
result, err = timeSchema.Parse("2021-01-01T00:00:00Z")  // ISO string → time.Time
```

---

## 📚 Complete Examples

### API Validation

```go
type CreateUserRequest struct {
    Name     string   `gozod:"required,min=2,max=50"`
    Email    string   `gozod:"required,email"`
    Age      int      `gozod:"required,min=18,max=120"`
    Tags     []string `gozod:"max=10"`
    Website  string   `gozod:"url"`
    IsActive bool     `gozod:"required"`
}

func createUserHandler(w http.ResponseWriter, r *http.Request) {
    var req CreateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }
    
    // Validate with generated schema
    schema := gozod.FromStruct[CreateUserRequest]()
    validatedReq, err := schema.Parse(req)
    if err != nil {
        if zodErr, ok := err.(*issues.ZodError); ok {
            response := map[string]any{
                "error": "Validation failed",
                "issues": zodErr.Issues,
            }
            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(http.StatusBadRequest)
            json.NewEncoder(w).Encode(response)
            return
        }
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    // Use validated data
    createUser(validatedReq)
    w.WriteHeader(http.StatusCreated)
}
```

### Configuration Validation

```go
type DatabaseConfig struct {
    Host     string `gozod:"required"`
    Port     int    `gozod:"required,min=1,max=65535"`
    Database string `gozod:"required,min=1"`
    Username string `gozod:"required"`
    Password string `gozod:"required,min=8"`
    SSL      bool   `gozod:"required"`
    Timeout  int    `gozod:"min=1,max=300"`  // seconds
}

type AppConfig struct {
    Environment string           `gozod:"required,regex=^(development|staging|production)$"`
    Port        int              `gozod:"required,min=1000,max=9999"`
    Database    DatabaseConfig   `gozod:"required"`
    Features    map[string]bool  `gozod:"required"`
}

func loadConfig(filename string) (*AppConfig, error) {
    data, err := os.ReadFile(filename)
    if err != nil {
        return nil, err
    }
    
    var config AppConfig
    if err := json.Unmarshal(data, &config); err != nil {
        return nil, err
    }
    
    // Validate configuration
    schema := gozod.FromStruct[AppConfig]()
    validatedConfig, err := schema.Parse(config)
    if err != nil {
        return nil, fmt.Errorf("configuration validation failed: %w", err)
    }
    
    return &validatedConfig, nil
}
```

This completes the comprehensive API reference for GoZod. The library provides type-safe validation with TypeScript Zod v4 compatibility while embracing Go idioms and providing maximum performance through strict type semantics and optional code generation.

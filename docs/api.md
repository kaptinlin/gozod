# GoZod API Reference

Complete type interface documentation for GoZod - a powerful, type-safe validation library for Go inspired by Zod.

## 🎯 Overview

GoZod provides comprehensive data validation with:
- **Type Safety**: Full Go generics support with preserved type information
- **Complete Strict Semantics**: All methods require exact type input - no automatic conversions
- **Maximum Performance**: Optimized validation pipeline with zero-overhead type handling
- **Composable Schemas**: Chain validations, transformations, and type conversions
- **Rich Validation**: Built-in validators for strings, numbers, objects, arrays, and more
- **Flexible Modifiers**: Optional, Nilable, Default, and Prefault handling for complex scenarios
- **Advanced Types**: Union, Intersection, and Discriminated Union support

## 🔧 Core Concepts

### Validation Pipeline
```go
// Input → [Coercion] → [Validation] → [Transformation] → Output
schema := gozod.String().Min(3).Overwrite(func(s string) string {
    return strings.ToUpper(s)
}).Transform(func(s string, ctx *gozod.RefinementContext) (any, error) {
    return fmt.Sprintf("Result: %s", s), nil
})
```

### Custom Checks
Need multiple business-rules at once? Use `.Check(fn)` to inspect the current value
and push unlimited issues via the provided `ParsePayload`:

```go
schema := gozod.Int().Check(func(v int, p *gozod.ParsePayload) {
    if v%2 != 0 {
        p.AddIssueWithMessage("number must be even")
    }
    if v < 0 {
        p.AddIssueWithCode(gozod.IssueTooSmall, "number must be positive")
    }
})

_, err := schema.Parse(-3) // err contains both issues above
```

For single custom validations, use `.Refine()` with optional `gozod.CustomParams`:

```go
// Simple refinement
schema := gozod.String().Refine(func(s string) bool {
    return len(s) > 5
}, "Too short")

// Advanced refinement with custom parameters
schema = gozod.String().Refine(func(s string) bool {
    return len(s) > 5
}, gozod.CustomParams{
    Error: "String too short",
    Abort: true,
    Path:  []any{"custom", "validation"},
})
```

### Type Preservation and Strict Input Handling
```go
// GoZod uses complete strict type semantics - no automatic conversions
var str = "hello"
var num = 42

// Value schemas: strict value input only
stringSchema := gozod.String()
result1, _ := stringSchema.Parse("hello")     // ✅ string input → string output
// result1, _ := stringSchema.Parse(&str)     // ❌ Error: requires string, got *string

intSchema := gozod.Int()
result2, _ := intSchema.Parse(42)             // ✅ int input → int output  
// result2, _ := intSchema.Parse(&num)        // ❌ Error: requires int, got *int

// Pointer schemas: strict pointer input only
stringPtrSchema := gozod.StringPtr()
result3, _ := stringPtrSchema.Parse(&str)     // ✅ *string input → *string output
// result3, _ := stringPtrSchema.Parse("hello") // ❌ Error: requires *string, got string

// For flexible input handling, use Optional/Nilable modifiers
optionalSchema := gozod.String().Optional()  // Returns *string, flexible input
result4, _ := optionalSchema.Parse("hello")   // ✅ string → *string (new pointer)
result5, _ := optionalSchema.Parse(&str)      // ✅ *string → *string (preserves identity)
```

---

## 📋 Table of Contents

### Core Concepts
- [Core Types](#-core-types)

### Primitive Types
- [String Validation](#-string-validation)
- [Number Validation](#-number-validation)
- [Boolean Validation](#-boolean-validation)
- [Time Validation](#-time-validation)
- [Format Validation](#-format-validation)

### Composite Types
- [Object Validation](#-object-validation)
- [Struct Validation](#-struct-validation)
- [Array & Slice Validation](#-array--slice-validation)
- [Map Validation](#-map-validation)
- [Enum Validation](#-enum-validation)

### Advanced Types
- [Union & Intersection Types](#-union--intersection-types)
- [Any & Unknown Types](#-any--unknown-types)

### Modifiers & Transformations
- [Modifiers & Wrappers](#-modifiers--wrappers)
- [Transform & Pipe](#-transform--pipe)
- [Custom Refinement](#custom-refinement)

### Utilities
- [Type Coercion](#-type-coercion)

### Record Validation
- [Record Validation](#-record-validation)

---

## 🔷 Core Types

### Primitive Types
| Type | Constructor | Description | Key Features |
|------|-------------|-------------|--------------|
| **String** | `gozod.String()` | String validation with format checks | Length, format, pattern validation |
| **Numbers** | `gozod.Int()`, `gozod.Float64()`, `gozod.Number()` | Numeric validation with range checks | Min/max, comparisons, special validations |
| **Boolean** | `gozod.Bool()` | Boolean value validation | Type-safe boolean handling |
| **Time** | `gozod.Time()`, `gozod.Coerce.Time()` | Date and time validation with Go's time.Time | Parsing, coercion, transformations |
| **Email** | `gozod.Email()` | Email format validation | RFC 5322, HTML5, Unicode patterns |
| **URL** | `gozod.URL()` | URL format validation | Hostname/protocol constraints |
| **ISO Formats** | `gozod.IsoDateTime()`, `gozod.IsoDate()`, `gozod.IsoTime()` | ISO 8601 format validation | Precision, timezone, duration support |

### Composite Types
| Type | Constructor | Description | Key Features |
|------|-------------|-------------|--------------|
| **Object** | `gozod.Object()` | Validate structured data with named fields | Field validation, modes, operations |
| **Struct** | `gozod.Struct[T]()`, `gozod.StructPtr[T]()` | Type-safe Go struct validation with field schemas | Strict input types, field validation, JSON tag mapping |
| **Array/Slice** | `gozod.Array()`, `gozod.Slice()` | Collection validation | Length constraints, element validation |
| **Map** | `gozod.Map()` | Key-value pair validation | Type-safe key-value validation |
| **Enum** | `gozod.Enum()`, `gozod.EnumMap()`, `gozod.EnumSlice()` | Type-safe enum validation | Multiple enum formats, simplified design |

### Advanced Types
| Type | Constructor | Description | Key Features |
|------|-------------|-------------|--------------|
| **Union** | `gozod.Union()` | Accept one of multiple types (OR logic) | Type discrimination, flexible validation |
| **Intersection** | `gozod.Intersection()` | Require all types simultaneously (AND logic) | Type combination, comprehensive validation |
| **Any** | `gozod.Any()` | Accept any value type | Universal acceptance, data transformations |
| **Unknown** | `gozod.Unknown()` | Accept any value type with special nil handling | Unknown data validation, sanitization |

---

## 🔤 String Validation

### Basic Usage

```go
// Basic string validation
schema := gozod.String()
result, err := schema.Parse("hello")  // Valid: "hello"

// Length validation
nameSchema := gozod.String().Min(2).Max(50)
result, err = nameSchema.Parse("Alice")  // Valid: "Alice"
_, err = nameSchema.Parse("A")           // Error: too short

// Pattern validation
emailSchema := gozod.String().Email()
result, err = emailSchema.Parse("user@example.com")  // Valid
_, err = emailSchema.Parse("invalid-email")          // Error

// Regular expression
usernameSchema := gozod.String().RegexString(`^[a-zA-Z0-9_]+$`)
result, err = usernameSchema.Parse("user_123")  // Valid
_, err = usernameSchema.Parse("user@123")       // Error
```



### String Methods Reference

| Method | Description | Example | Error Code |
|--------|-------------|---------|------------|
| `.Min(n)` | Minimum length | `gozod.String().Min(3)` | `too_small` |
| `.Max(n)` | Maximum length | `gozod.String().Max(100)` | `too_big` |
| `.Length(n)` | Exact length | `gozod.String().Length(10)` | `invalid_format` |
| `.Email()` | Email format | `gozod.String().Email()` | `invalid_format` |
| `.URL()` | URL format | `gozod.String().URL()` | `invalid_format` |
| `.UUID()` | UUID format | `gozod.String().UUID()` | `invalid_format` |
| `.Regex(pattern)` | Custom pattern | `gozod.String().Regex(pattern)` | `invalid_format` |
| `.RegexString(pattern)` | String pattern (convenience) | `gozod.String().RegexString(\`^\d+$\`)` | `invalid_format` |
| `.Includes(str)` | Contains substring | `gozod.String().Includes("test")` | `invalid_format` |
| `.StartsWith(str)` | Starts with | `gozod.String().StartsWith("prefix")` | `invalid_format` |
| `.EndsWith(str)` | Ends with | `gozod.String().EndsWith("suffix")` | `invalid_format` |



---

## 🔢 Number Validation

### Basic Usage

```go
// Integer validation
ageSchema := gozod.Int().Min(0).Max(120)
result, err := ageSchema.Parse(25)   // Valid: 25
_, err = ageSchema.Parse(-1)         // Error: too small

// Float validation
priceSchema := gozod.Float64().Positive()
result, err = priceSchema.Parse(19.99)  // Valid: 19.99
_, err = priceSchema.Parse(-5.0)        // Error: must be positive

// Range constraints
scoreSchema := gozod.Int().Gte(0).Lt(100)
result, err = scoreSchema.Parse(85)   // Valid: 85
_, err = scoreSchema.Parse(100)       // Error: too big

// Special validations
evenSchema := gozod.Int().MultipleOf(2)
result, err = evenSchema.Parse(4)   // Valid: 4
_, err = evenSchema.Parse(3)        // Error: not multiple of 2
```

### Number Methods Reference

| Method | Description | Example | Error Code |
|--------|-------------|---------|------------|
| `.Min(n)` | Minimum value | `gozod.Int().Min(0)` | `too_small` |
| `.Max(n)` | Maximum value | `gozod.Int().Max(100)` | `too_big` |
| `.Gt(n)` | Greater than | `gozod.Float64().Gt(0.0)` | `too_small` |
| `.Gte(n)` | Greater than or equal | `gozod.Int().Gte(18)` | `too_small` |
| `.Lt(n)` | Less than | `gozod.Float64().Lt(100.0)` | `too_big` |
| `.Lte(n)` | Less than or equal | `gozod.Int().Lte(65)` | `too_big` |
| `.Positive()` | Must be positive | `gozod.Float64().Positive()` | `too_small` |
| `.Negative()` | Must be negative | `gozod.Float64().Negative()` | `too_big` |
| `.MultipleOf(n)` | Multiple of value | `gozod.Int().MultipleOf(5)` | `not_multiple_of` |
| `.Finite()` | Finite number | `gozod.Float64().Finite()` | `not_finite` |

---

## ✅ Boolean Validation

### Basic Usage

```go
// Boolean validation
activeSchema := gozod.Bool()
result, err := activeSchema.Parse(true)   // Valid: true
result, err = activeSchema.Parse(false)  // Valid: false
_, err = activeSchema.Parse("true")       // Error: wrong type

// String-to-Boolean conversion
stringBoolSchema := gozod.StringBool()
result, err = stringBoolSchema.Parse("true")   // Valid: true
result, err = stringBoolSchema.Parse("false")  // Valid: false
result, err = stringBoolSchema.Parse("1")      // Valid: true
result, err = stringBoolSchema.Parse("0")      // Valid: false

// StringBool with prefault (accepts string input type)
stringBoolWithPrefault := gozod.StringBool().Prefault("true")
result, err = stringBoolWithPrefault.Parse(nil)        // Valid: true (uses prefault)
result, err = stringBoolWithPrefault.Parse("false")    // Valid: false
_, err = stringBoolWithPrefault.Parse(123)              // Error: invalid type
```

---

## ⏰ Time Validation

### Basic Usage

```go
// Time validation
timeSchema := Time()
testTime := time.Date(2023, 12, 25, 15, 30, 0, 0, time.UTC)
result, err := timeSchema.Parse(testTime)  // Valid: time.Time

// Coerced time - converts various formats
coercedSchema := Time(core.SchemaParams{Coerce: true})
result, err = coercedSchema.Parse("2023-12-25T15:30:00Z")  // Valid: time.Time
result, err = coercedSchema.Parse(1703517000)              // Valid: Unix timestamp

// Time transformations
utcSchema := Time().Overwrite(func(t time.Time) time.Time {
    return t.UTC()
})
result, err = utcSchema.Parse(testTime)  // Valid: UTC time
```

---

## 🎯 Format Validation

### Basic Usage

```go
// Email validation
emailSchema := gozod.Email()
result, err := emailSchema.Parse("user@example.com")  // Valid

// URL validation
urlSchema := gozod.URL()
result, err = urlSchema.Parse("https://example.com")  // Valid

// IP address validation
ipv4Schema := gozod.IPv4()
result, err = ipv4Schema.Parse("192.168.1.1")  // Valid

// JWT validation
jwtSchema := gozod.JWT()
result, err = jwtSchema.Parse(validToken)  // Valid

// ISO date/time validation
datetimeSchema := gozod.IsoDateTime()
result, err = datetimeSchema.Parse("2023-12-25T15:30:45Z")  // Valid
```

### Format Constructors Reference

| Constructor | Description | Options |
|-------------|-------------|---------|
| `gozod.Email()` | Email validation | `.Html5()`, `.Rfc5322()`, `.Unicode()` |
| `gozod.Emoji()` | Emoji validation | Basic emoji regex |
| `gozod.JWT()` | JWT token validation | `JWTOptions{Algorithm}` for alg check |
| `gozod.URL()` | URL validation | `URLOptions{Hostname, Protocol}` |
| `gozod.IPv4()`, `gozod.IPv6()` | IP address validation | Basic validation |
| `gozod.CIDRv4()`, `gozod.CIDRv6()` | CIDR notation validation | Basic validation |
| `gozod.Cuid()`, `gozod.Uuid()` | ID validation (CUID/UUID etc.) | Version helpers e.g. `Uuidv4()` |
| `gozod.IsoDateTime()` | ISO 8601 datetime | `IsoDatetimeOptions{Precision, Offset, Local}` |
| `gozod.IsoDate()` | ISO 8601 date | Basic validation |
| `gozod.IsoTime()` | ISO 8601 time | `IsoTimeOptions{Precision}` |
| `gozod.IsoDuration()` | ISO 8601 duration | Basic validation |

### Precision Constants

| Constant | Description | Example Format |
|----------|-------------|----------------|
| `PrecisionMinute` | Minute precision | `HH:MM` |
| `PrecisionSecond` | Second precision | `HH:MM:SS` |
| `PrecisionMillisecond` | Millisecond precision | `HH:MM:SS.sss` |
| `PrecisionDecisecond` | Decisecond precision | `HH:MM:SS.s` |
| `PrecisionCentisecond` | Centisecond precision | `HH:MM:SS.ss` |
| `PrecisionMicrosecond` | Microsecond precision | `HH:MM:SS.sssssss` |
| `PrecisionNanosecond` | Nanosecond precision | `HH:MM:SS.sssssssss` |

---

## 🗂️ Object Validation

### Basic Usage

```go
// Object validation
userSchema := gozod.Object(gozod.ObjectSchema{
    "name":  gozod.String().Min(2),
    "age":   gozod.Int().Min(0).Max(120),
    "email": gozod.String().Email().Optional(),
})

userData := map[string]any{
    "name": "Alice",
    "age":  25,
    "email": "alice@example.com",
}
result, err := userSchema.Parse(userData)  // Valid

// Object operations
baseSchema := gozod.Object(gozod.ObjectSchema{
    "name": gozod.String(),
    "age":  gozod.Int(),
})

nameOnly := baseSchema.Pick([]string{"name"})     // Select fields
partial := baseSchema.Partial()                   // Make optional
```

### Object Methods Reference

| Method | Description | Example |
|--------|-------------|---------|
| `.Pick(fields)` | Select specific fields | `schema.Pick([]string{"name", "age"})` |
| `.Omit(fields)` | Exclude specific fields | `schema.Omit([]string{"password"})` |
| `.Partial()` | Make all fields optional | `schema.Partial()` |
| `.Required(fields)` | Make specific fields required | `schema.Required([]string{"name"})` |
| `.Extend(schema)` | Add new fields | `schema.Extend(ObjectSchema{...})` |
| `.Merge(schema)` | Combine with another schema | `schema.Merge(otherSchema)` |
| `.Strip()` | Remove extra fields | `schema.Strip()` |
| `.Catchall(schema)` | Validate extra fields | `schema.Catchall(gozod.String())` |

---

## 🏗️ Struct Validation

### Basic Usage

```go
// Struct validation
type User struct {
    Name  string `json:"name"`
    Age   int    `json:"age"`
    Email string `json:"email"`
}

// Basic struct validation
basicSchema := gozod.Struct[User]()
validUser := User{Name: "John", Age: 30, Email: "john@example.com"}
result, err := basicSchema.Parse(validUser)  // Valid

// Struct with field validation
userSchema := gozod.Struct[User](gozod.StructSchema{
    "name":  gozod.String().Min(2),
    "age":   gozod.Int().Min(0).Max(150),
    "email": gozod.String().Email(),
})
result, err = userSchema.Parse(validUser)  // Valid with field validation
```

### Strict Type Requirements

```go
type User struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}

// Struct[T] - requires exact value type T
valueSchema := gozod.Struct[User](gozod.StructSchema{
    "name":  gozod.String().Min(2),
    "email": gozod.String().Email(),
})

// StructPtr[T] - requires exact pointer type *T
ptrSchema := gozod.StructPtr[User](gozod.StructSchema{
    "name":  gozod.String().Min(2),
    "email": gozod.String().Email(),
})

validUser := User{Name: "John", Email: "john@example.com"}

// Strict input type enforcement
result1, err := valueSchema.Parse(validUser)    // ✅ Valid: User → User
result2, err := ptrSchema.Parse(&validUser)     // ✅ Valid: *User → *User

// Type mismatches are rejected
// result1, err := valueSchema.Parse(&validUser)   // ❌ Error: requires User, got *User
// result2, err := ptrSchema.Parse(validUser)      // ❌ Error: requires *User, got User

// For flexible input handling, use Optional/Nilable modifiers
flexibleSchema := gozod.Struct[User](...).Optional()  // Returns *User, flexible input
```

### JSON Tag Field Mapping

```go
type Person struct {
    ID       int    `json:"id"`
    FullName string `json:"full_name"`
    Active   bool   `json:"active"`
}

// Schema uses JSON tag names for field mapping
personSchema := gozod.Struct[Person](gozod.StructSchema{
    "id":        gozod.Int().Min(1),        // Maps to ID field
    "full_name": gozod.String().Min(2),     // Maps to FullName field
    "active":    gozod.Bool(),              // Maps to Active field
})

validPerson := Person{ID: 123, FullName: "John Doe", Active: true}
result, err := personSchema.Parse(validPerson)  // Valid: JSON tag mapping
```

### Partial and Required Operations

```go
// Partial validation - skips zero value field validation
partialSchema := gozod.Struct[User](gozod.StructSchema{
    "name":  gozod.String().Min(2),
    "email": gozod.String().Email(),
}).Partial()

// Accepts struct with zero values for optional fields
user := User{Name: "John", Age: 0, Email: ""}  // Age and Email are zero values
result, err := partialSchema.Parse(user)  // Valid: zero values skipped

// Selective partial - only specified fields become optional
selectiveSchema := userSchema.Partial([]string{"age"})  // Only age is optional
user = User{Name: "John", Age: 0, Email: "john@example.com"}
result, err = selectiveSchema.Parse(user)  // Valid: age can be zero, email must be valid

// Required operation - make specific fields required in partial schema
requiredSchema := userSchema.Partial().Required([]string{"email"})
user = User{Name: "", Age: 0, Email: "john@example.com"}  // name and age optional, email required
result, err = requiredSchema.Parse(user)  // Valid
```

### Struct Constructors

| Constructor | Description | Input Requirements | Example |
|-------------|-------------|-------------------|---------|
| `gozod.Struct[T]()` | Basic struct validation | **Requires `T` only** | `gozod.Struct[User]()` |
| `gozod.Struct[T](schema)` | Struct with field validation | **Requires `T` only** | `gozod.Struct[User](gozod.StructSchema{...})` |
| `gozod.StructPtr[T]()` | Pointer struct validation | **Requires `*T` only** | `gozod.StructPtr[User]()` |
| `gozod.StructPtr[T](schema)` | Pointer struct with field validation | **Requires `*T` only** | `gozod.StructPtr[User](gozod.StructSchema{...})` |

### Struct Methods Reference

| Method | Description | Example |
|--------|-------------|---------|
| `.Partial()` | Make all fields optional | `schema.Partial()` |
| `.Partial(fields)` | Make specific fields optional | `schema.Partial([]string{"age", "bio"})` |
| `.Required(fields)` | Make specific fields required | `schema.Required([]string{"email"})` |
| `.Extend(schema)` | Add new fields | `schema.Extend(gozod.StructSchema{...})` |

---

## 📋 Array & Slice Validation

### Slice Validation (Dynamic Length)

```go
// Basic slice validation
stringSlice := gozod.Slice[string](gozod.String().Min(2))
result, err := stringSlice.Parse([]string{"hello", "world"})
// result: []string{"hello", "world"}, err: nil

_, err = stringSlice.Parse([]string{"hello", "hi"})
// Error: "hi" is too short

// Size constraints
constrainedSlice := gozod.Slice[string](gozod.String()).Min(1).Max(5)
result, err = constrainedSlice.Parse([]string{"a", "b", "c"})  // Valid
_, err = constrainedSlice.Parse([]string{})                   // Error: too few items
_, err = constrainedSlice.Parse([]string{"a", "b", "c", "d", "e", "f"})  // Error: too many items
```

### Array Validation (Fixed Length Tuples)

```go
// Fixed-length tuple validation
coordinates := gozod.Array([]gozod.ZodType[any, any]{
    gozod.Float64(), // x coordinate
    gozod.Float64(), // y coordinate
})

result, err := coordinates.Parse([]any{3.14, 2.71})
// result: []any{3.14, 2.71}, err: nil

_, err = coordinates.Parse([]any{3.14})
// Error: wrong number of elements

_, err = coordinates.Parse([]any{3.14, 2.71, 1.41})
// Error: too many elements
```

### Array/Slice Methods Reference

| Method | Description | Example |
|--------|-------------|---------|
| `.Min(n)` | Minimum length | `gozod.Slice(gozod.String()).Min(1)` |
| `.Max(n)` | Maximum length | `gozod.Slice(gozod.String()).Max(10)` |
| `.Length(n)` | Exact length | `gozod.Slice(gozod.String()).Length(5)` |
| `.NonEmpty()` | Must not be empty | `gozod.Slice(gozod.String()).NonEmpty()` |

---

## 🗺️ Map Validation

### Basic Map Validation

```go
// Map with typed keys and values
userMap := gozod.Map(gozod.String(), gozod.Int())
result, err := userMap.Parse(map[string]int{
    "alice": 25,
    "bob":   30,
})
// result: map[string]int{"alice": 25, "bob": 30}, err: nil

// Type inference preservation
result, err = userMap.Parse(map[any]any{
    "alice": 25,
    "bob":   30,
})
// result: map[any]any{"alice": 25, "bob": 30}, err: nil
```

### Map Constraints

```go
// Size constraints
sizedMap := gozod.Map(gozod.String(), gozod.Int()).Min(1).Max(5)
result, err := sizedMap.Parse(map[string]int{"key": 42})  // Valid: 1 entry
_, err = sizedMap.Parse(map[string]int{})                // Error: too few entries

// Key and value validation
restrictedMap := gozod.Map(gozod.String().Min(3), gozod.Int().Min(10))
result, err = restrictedMap.Parse(map[string]int{
    "abc": 15,  // Valid: key >= 3 chars, value >= 10
})
_, err = restrictedMap.Parse(map[string]int{
    "ab": 15,   // Error: key too short
})
_, err = restrictedMap.Parse(map[string]int{
    "abc": 5,   // Error: value too small
})
```

### Map Methods Reference

| Method | Description | Example |
|--------|-------------|---------|
| `.Min(n)` | Minimum entries | `gozod.Map(gozod.String(), gozod.Int()).Min(1)` |
| `.Max(n)` | Maximum entries | `gozod.Map(gozod.String(), gozod.Int()).Max(10)` |
| `.Length(n)` | Exact entries | `gozod.Map(gozod.String(), gozod.Int()).Length(5)` |

### 📑 Record Validation

`Record` validates key–value objects where **keys are strings** (e.g. JSON objects).  Compared to `Map`:

1. Keys are always strings. You can supply an `Enum`, `Literal`, or `Union` schema to enforce an exhaustive key set.
2. Unknown keys are rejected by default, ensuring inputs match the expected fields (`PartialRecord` relaxes the missing-key check).
3. `Map` is more generic and accepts any comparable key type; `Record` gives stricter guarantees for object-shaped data.

#### Basic Record Validation

```go
// Strict record: keys must be "id", "name", "email"
userRecord := gozod.Record(
    gozod.Enum("id", "name", "email"), // key schema (exhaustive)
    gozod.String(),                        // value schema
)

validInput := map[string]any{
    "id":    "user-123",
    "name":  "Alice",
    "email": "alice@example.com",
}
_, err := userRecord.Parse(validInput) // ✅
```

#### Partial Records

```go
partial := gozod.PartialRecord(
    gozod.Enum("id", "name", "email"),
    gozod.String(),
)

// "email" can be omitted; unknown keys are still rejected
partial.Parse(map[string]any{"id": "123"})
```

#### Choosing between Record and Map

| Use-case | Recommended |
|----------|-------------|
| Strict JSON objects / fixed fields | **Record** |
| Dynamic or non-string keys | **Map** |

---

## 🏷️ Enum Validation

### Basic Enum Types

```go
// Variadic enum
colorEnum := gozod.Enum("red", "green", "blue")
result, err := colorEnum.Parse("red")    // Valid: "red"
_, err = colorEnum.Parse("yellow")       // Error: not in enum

// Slice-based enum
colorsSlice := gozod.EnumSlice([]string{"red", "green", "blue"})
result, err = colorsSlice.Parse("green") // Valid: "green"

// Map-based enum with value mapping
statusMap := gozod.EnumMap(map[string]string{
    "ACTIVE":   "active",
    "INACTIVE": "inactive",
})
result, err = statusMap.Parse("ACTIVE")  // Valid: "active" (returns mapped value)

// Type-safe enum mapping (returns mapped values)
statusEnum := gozod.EnumMap(map[string]string{
    "SUCCESS": "success",
    "FAILURE": "failure",
})
result, err = statusEnum.Parse("SUCCESS") // Valid: "success"
```

### Go Native Enums

```go
// Go iota enum support - type-safe validation
type Status int
const (
    Active Status = iota
    Inactive
    Pending
)

statusEnum := gozod.Enum(Active, Inactive, Pending)
result, err := statusEnum.Parse(Active)    // Valid: Active
_, err = statusEnum.Parse(Status(99))      // Error: not in enum

// String-based Go enum
type Color string
const (
    Red   Color = "red"
    Green Color = "green"
    Blue  Color = "blue"
)

colorEnum := gozod.Enum(Red, Green, Blue)
result, err = colorEnum.Parse(Red)         // Valid: Red

// Enum with modifiers - simplified design returns zero values
nilableEnum := gozod.Enum("a", "b", "c").Nilable()
result, err = nilableEnum.Parse(nil)       // Valid: "" (zero value for string)
result, err = nilableEnum.Parse("a")       // Valid: "a"

optionalEnum := gozod.Enum(1, 2, 3).Optional()
result, err = optionalEnum.Parse(nil)      // Valid: 0 (zero value for int)
result, err = optionalEnum.Parse(2)        // Valid: 2
```

### Enum Operations

```go
// Get enum options
fishEnum := gozod.Enum("Salmon", "Tuna", "Trout")
options := fishEnum.Options()
// Returns: []string{"Salmon", "Tuna", "Trout"}

// Extract specific values
subset := fishEnum.Extract([]string{"0", "1"}) // Extract by index keys
result, err := subset.Parse("Salmon")  // Valid: included in subset
_, err = subset.Parse("Trout")         // Error: not in subset

// Exclude specific values
remaining := fishEnum.Exclude([]string{"1"}) // Exclude by index key
result, err = remaining.Parse("Salmon")  // Valid: not excluded
_, err = remaining.Parse("Tuna")         // Error: excluded
```

### Enum Methods Reference

| Method | Description | Example |
|--------|-------------|---------|
| `.Options()` | Get all enum values | `enum.Options()` |
| `.Extract(keys)` | Select specific values | `enum.Extract([]string{"0", "1"})` |
| `.Exclude(keys)` | Remove specific values | `enum.Exclude([]string{"1"})` |

---

## 🔀 Union & Intersection Types

### Basic Usage

```go
// Union types (OR logic) - accepts ANY matching schema
stringOrNumber := gozod.Union([]gozod.ZodType[any, any]{
    gozod.String(),
    gozod.Int(),
})
result, err := stringOrNumber.Parse("hello")  // Valid: string
result, err = stringOrNumber.Parse(42)        // Valid: int
_, err = stringOrNumber.Parse(true)           // Error: no match

// Discriminated union (optimized for objects)
apiResponse := gozod.DiscriminatedUnion("status", []gozod.ZodType[any, any]{
    gozod.Object(gozod.ObjectSchema{
        "status": gozod.Literal("success"),
        "data":   gozod.String(),
    }),
    gozod.Object(gozod.ObjectSchema{
        "status": gozod.Literal("error"),
        "error":  gozod.String(),
    }),
})

// Intersection types (AND logic) - requires ALL schemas
personEmployee := gozod.Intersection(
    gozod.Object(gozod.ObjectSchema{
        "name": gozod.String(),
        "age":  gozod.Int().Min(18),
    }),
    gozod.Object(gozod.ObjectSchema{
        "employeeId": gozod.String(),
        "department": gozod.String(),
    }),
)
```

### Union/Intersection Methods Reference

| Method | Logic | Description | Use Case |
|--------|-------|-------------|----------|
| `gozod.Union(schemas)` | OR | Accept if matches ANY schema | Flexible input types |
| `gozod.DiscriminatedUnion(key, schemas)` | OR (optimized) | Union with discriminator field lookup | API responses, tagged objects |
| `gozod.Intersection(schemas...)` | AND | Require ALL schemas to match | Comprehensive validation |

---

## 🌟 Any & Unknown Types

### Basic Usage

```go
// Any type - accepts any value without validation
anySchema := gozod.Any()
result, err := anySchema.Parse("hello")  // Valid: "hello"
result, err = anySchema.Parse(42)        // Valid: 42
result, err = anySchema.Parse(nil)       // Valid: nil

// Unknown type - similar to Any but for untrusted data
unknownSchema := gozod.Unknown()
result, err = unknownSchema.Parse("hello")  // Valid: "hello"
result, err = unknownSchema.Parse(42)       // Valid: 42
result, err = unknownSchema.Parse(nil)      // Valid: nil

// Unknown with validation
safeUnknown := gozod.Unknown().Refine(func(v any) bool {
    return v != nil
}, "Input cannot be nil")
```



---

## 🔧 Modifiers & Wrappers

### Basic Usage

```go
// Optional - allows missing values
optionalEmail := gozod.String().Email().Optional()
result, err := optionalEmail.Parse("user@example.com")  // Valid: email
result, err = optionalEmail.Parse(nil)                  // Valid: nil

// Nilable - handles explicit null by returning a typed nil pointer
nilableAge := gozod.String().Nilable()
result, err = nilableAge.Parse("hello")  // Valid: "hello"
result, err = nilableAge.Parse(nil)      // Valid: (*string)(nil)

// Default values
nameWithDefault := gozod.String().Default("Anonymous")
result, err = nameWithDefault.Parse(nil)    // Valid: "Anonymous"
result, err = nameWithDefault.Parse("Alice") // Valid: "Alice"

// Prefault - pre-parse default for nil input (goes through full parsing)
safeValue := gozod.String().Transform(func(s string) string { return strings.ToUpper(s) }).Prefault("fallback")
result, err = safeValue.Parse("hello")  // Valid: "HELLO"
result, err = safeValue.Parse(nil)      // Valid: "FALLBACK" (nil uses prefault, then transforms)
result, err = safeValue.Parse(123)      // Error: invalid type (no prefault for non-nil)

// Schema introspection
schema := gozod.String().Optional()
isOptional := schema.IsOptional()  // true
isNilable := schema.IsNilable()    // false
```

### Modifier Methods Reference

| Method | Description | Example |
|--------|-------------|---------|
| `.Optional()` | Allow missing values | `gozod.String().Optional()` |
| `.Nilable()` | Handle explicit null | `gozod.String().Nilable()` |
| `.NonOptional()` | Remove optional flag | `gozod.String().Optional().NonOptional()` |
| `.Default(value)` | Static default | `gozod.String().Default("Anonymous")` |
| `.DefaultFunc(fn)` | Dynamic default | `gozod.String().DefaultFunc(func() string {...})` |
| `.Prefault(value)` | Pre-parse default for nil | `gozod.String().Prefault("fallback")` |
| `.IsOptional()` | Check if schema is optional | `schema.IsOptional()` |
| `.IsNilable()` | Check if schema is nilable | `schema.IsNilable()` |

---

## 🔄 Transform & Pipe

### Basic Usage

```go
// Transform - modifies data after validation
upperSchema := gozod.String().Transform(func(s string, ctx *gozod.RefinementContext) (any, error) {
    return strings.ToUpper(s), nil
})
result, err := upperSchema.Parse("hello")  // Valid: "HELLO"

// Pipe - chains validation and transformation
pipeline := gozod.String().
    Transform(func(s string, ctx *gozod.RefinementContext) (any, error) {
        return strings.TrimSpace(s), nil
    }).
    Pipe(gozod.String().Min(3))
result, err = pipeline.Parse("  hello  ")  // Valid: "hello"

// Overwrite - transforms data during validation (preserves type)
overwriteSchema := gozod.String().Min(3).Overwrite(func(s string) string {
    return strings.ToUpper(s)
})
result, err = overwriteSchema.Parse("hi")  // Valid: "HI" (transformed then validated)
```

### Custom Refinement

```go
// Basic refinement with message
passwordSchema := gozod.String().Refine(func(s string) bool {
    return len(s) >= 8
}, "Password must be at least 8 characters")

// Advanced refinement with CustomParams
complexSchema := gozod.String().Refine(func(s string) bool {
    return len(s) >= 8
}, gozod.CustomParams{
    Error: "Password too short",
    Abort: true,  // Stop validation on failure
    Path:  []any{"password", "strength"},  // Custom error path
    When: func(p *gozod.ParsePayload) bool {
        return p.GetIssueCount() == 0  // Only run if no previous errors
    },
})
```

### CustomParams Reference

The `gozod.CustomParams` structure provides advanced control over refinement behavior:

| Field | Type | Description |
|-------|------|-------------|
| `Error` | `any` | Custom error message or error object |
| `Abort` | `bool` | Stop validation chain on failure |
| `Path` | `[]any` | Override error path for better error reporting |
| `When` | `gozod.ZodWhenFn` | Conditional execution predicate |
| `Params` | `map[string]any` | Additional metadata for error handling |

### Transform/Pipe Methods Reference

| Method | Description | Execution Phase | Type Change |
|--------|-------------|----------------|-------------|
| `.Transform(fn)` | Modify data after validation | Post-validation | Yes (string → any) |
| `.Overwrite(fn)` | Transform data during validation | Mid-validation | No (string → string) |
| `.Pipe(schema)` | Chain to another schema | Sequential validation | Depends on target schema |
| `.Refine(fn, params)` | Custom validation with advanced parameters | Validation phase | No (preserves input) |

---

## 🔀 Type Coercion

### Coerce Package

```go
// Basic type coercion using coerce package
import "github.com/kaptinlin/gozod/coerce"

stringSchema := coerce.String()
result, _ := stringSchema.Parse(123)     // "123"
result, _ = stringSchema.Parse(true)     // "true"
result, _ = stringSchema.Parse(3.14)     // "3.14"

numberSchema := coerce.Number()
result, _ = numberSchema.Parse("42")     // 42.0
result, _ = numberSchema.Parse(true)     // 1.0

boolSchema := coerce.Bool()
result, _ = boolSchema.Parse("false")    // false
result, _ = boolSchema.Parse(0)          // false
result, _ = boolSchema.Parse("1")        // true

// Big integer coercion
bigIntSchema := coerce.BigInt()
result, _ = bigIntSchema.Parse("9223372036854775808")  // *big.Int

// Time coercion
timeSchema := coerce.Time()
result, _ = timeSchema.Parse("2023-12-25T15:30:00Z")     // time.Time
result, _ = timeSchema.Parse(1703517000)                 // Unix timestamp to time.Time
result, _ = timeSchema.Parse("1703517000")               // Unix timestamp string to time.Time
result, _ = timeSchema.Parse("2023-12-25")               // Date string to time.Time

// String to boolean coercion
stringBoolSchema := coerce.StringBool()
result, _ = stringBoolSchema.Parse("true")    // true
result, _ = stringBoolSchema.Parse("1")       // true
result, _ = stringBoolSchema.Parse("false")   // false
result, _ = stringBoolSchema.Parse("0")       // false
```

### Schema-Level Coercion

```go
// Enable coercion using coerce package
coerceSchema := coerce.String()
result, _ := coerceSchema.Parse(123)  // "123"

// Coercion with validation
validatedCoerce := coerce.Float64().Min(0.0)
result, _ = validatedCoerce.Parse("3.14")  // 3.14 (coerced then validated)
_, err := validatedCoerce.Parse("-1")      // Error: coerced to -1.0, fails Min(0.0)
```

### Coercion Methods Reference

| Method | Description | Example |
|--------|-------------|---------|
| `coerce.String()` | Coerce to string | `coerce.String().Parse(123)` → `"123"` |
| `coerce.Number()` | Coerce to number | `coerce.Number().Parse("42")` → `42.0` |
| `coerce.Bool()` | Coerce to boolean | `coerce.Bool().Parse("true")` → `true` |
| `coerce.BigInt()` | Coerce to big.Int | `coerce.BigInt().Parse("123")` → `*big.Int` |
| `coerce.Time()` | Coerce to time.Time | `coerce.Time().Parse("2023-12-25T15:30:00Z")` → `time.Time` |
| `coerce.StringBool()` | String to boolean | `coerce.StringBool().Parse("true")` → `true` |
| `coerce.Int()` | Coerce to int | `coerce.Int().Parse("42")` → `42` |
| `coerce.Float64()` | Coerce to float64 | `coerce.Float64().Parse("3.14")` → `3.14` |

---

This API reference provides complete type interface documentation for GoZod. For usage patterns and practical examples, see the [Basics Guide](basics.md). For error handling and customization, see the [Error Customization Guide](error-customization.md).

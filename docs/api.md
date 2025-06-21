# GoZod API Reference

Complete type interface documentation for GoZod validation library.

---

## 📋 Table of Contents

- [Core Types](#-core-types)
- [String Validation](#-string-validation)
- [Number Validation](#-number-validation)
- [Boolean Validation](#-boolean-validation)
- [Object Validation](#-object-validation)
- [Array & Slice Validation](#-array--slice-validation)
- [Map Validation](#-map-validation)
- [Enum Validation](#-enum-validation)
- [Union & Intersection Types](#-union--intersection-types)
- [Modifiers & Wrappers](#-modifiers--wrappers)
- [Transform & Pipe](#-transform--pipe)
- [Type Coercion](#-type-coercion)

---

## 🔷 Core Types

| Type | Constructor | Description | Key Features |
|------|-------------|-------------|--------------|
| **String** | `gozod.String()` | String validation with format checks | Length, format, pattern validation |
| **Numbers** | `gozod.Int()`, `gozod.Float64()`, `gozod.Number()` | Numeric validation with range checks | Min/max, comparisons, special validations |
| **Boolean** | `gozod.Bool()` | Boolean value validation | Type-safe boolean handling |
| **Object** | `gozod.Object()` | Validate object structures | Field validation, modes, operations |
| **Array/Slice** | `gozod.Array()`, `gozod.Slice()` | Collection validation | Length constraints, element validation |
| **Map** | `gozod.Map()` | Key-value pair validation | Type-safe key-value validation |
| **Enum** | `gozod.Enum()`, `gozod.EnumMap()`, `gozod.EnumSlice()` | Type-safe enum validation | Multiple enum formats |
| **Union** | `gozod.Union()` | Accept one of multiple types | Type discrimination |
| **Intersection** | `gozod.Intersection()` | Require all types simultaneously | Type combination |
| **Any** | `gozod.Any()` | Accept any value type | Universal acceptance |

---

## 🔤 String Validation

### Basic Usage with Type Inference

```go
// Basic string validation
schema := gozod.String()
result, err := schema.Parse("hello")
// result: "hello", err: nil

// Pointer input preserves exact pointer identity
input := "world"
result, err = schema.Parse(&input)
// result: &input (same memory address), err: nil
```

### Length Validation

```go
// Minimum length constraint
nameSchema := gozod.String().Min(2)
result, err := nameSchema.Parse("Alice")  // Valid: "Alice"
_, err = nameSchema.Parse("A")            // Error: String must be at least 2 characters

// Maximum length constraint
usernameSchema := gozod.String().Max(20)
result, err = usernameSchema.Parse("user123")  // Valid: "user123"
_, err = usernameSchema.Parse("this_username_is_way_too_long")  // Error: String must be at most 20 characters

// Exact length requirement
codeSchema := gozod.String().Length(6)
result, err = codeSchema.Parse("ABC123")  // Valid: "ABC123"
_, err = codeSchema.Parse("ABC12")        // Error: String must be exactly 6 characters

// Method chaining
validationSchema := gozod.String().Min(5).Max(50)
result, err = validationSchema.Parse("hello world")  // Valid: passes both constraints
```

### Format Validation

```go
// Email validation
emailSchema := gozod.String().Email()
result, err := emailSchema.Parse("user@example.com")  // Valid: "user@example.com"
_, err = emailSchema.Parse("invalid-email")           // Error: Invalid email format

// URL validation
urlSchema := gozod.String().URL()
result, err = urlSchema.Parse("https://example.com")  // Valid: "https://example.com"
_, err = urlSchema.Parse("not-a-url")                 // Error: Invalid URL format

// UUID validation
uuidSchema := gozod.String().UUID()
result, err = uuidSchema.Parse("f47ac10b-58cc-4372-a567-0e02b2c3d479")  // Valid
_, err = uuidSchema.Parse("not-a-uuid")                                 // Error: Invalid UUID format
```

### Pattern Validation

```go
// Substring matching
prefixSchema := gozod.String().StartsWith("user")
result, err := prefixSchema.Parse("user@example.com")  // Valid: starts with "user"
_, err = prefixSchema.Parse("admin@example.com")       // Error: doesn't start with "user"

suffixSchema := gozod.String().EndsWith(".go")
result, err = suffixSchema.Parse("main.go")     // Valid: ends with ".go"
_, err = suffixSchema.Parse("main.py")          // Error: doesn't end with ".go"

containsSchema := gozod.String().Includes("test")
result, err = containsSchema.Parse("testing123")  // Valid: contains "test"
_, err = containsSchema.Parse("production")       // Error: doesn't contain "test"

// Regular expression validation
pattern := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
usernamePattern := gozod.String().Regex(pattern)
result, err = usernamePattern.Parse("user_123")  // Valid: matches pattern
_, err = usernamePattern.Parse("user@123")       // Error: contains invalid characters
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
| `.Includes(str)` | Contains substring | `gozod.String().Includes("test")` | `invalid_format` |
| `.StartsWith(str)` | Starts with | `gozod.String().StartsWith("prefix")` | `invalid_format` |
| `.EndsWith(str)` | Ends with | `gozod.String().EndsWith("suffix")` | `invalid_format` |

---

## 🔢 Number Validation

### Integer Types with Type Inference

```go
// Basic integer validation
ageSchema := gozod.Int().Min(0).Max(120)
result, err := ageSchema.Parse(25)   // Valid: 25 (int)
_, err = ageSchema.Parse(-1)         // Error: Number must be at least 0
_, err = ageSchema.Parse(150)        // Error: Number must be at most 120

// Pointer identity preservation
age := 30
result, err = ageSchema.Parse(&age)  // Valid: &age (exact same pointer)

// Specific integer types with type inference
result1, _ := Int8().Parse(int8(42))        // result1: int8(42)
result2, _ := Int16().Parse(int16(1000))    // result2: int16(1000)
result3, _ := Int32().Parse(int32(100000))  // result3: int32(100000)
result4, _ := Int64().Parse(int64(9223372036854775807)) // result4: int64(...)
result5, _ := Uint8().Parse(uint8(255))     // result5: uint8(255)
```

### Float Types

```go
// Floating point validation with range constraints
priceSchema := gozod.Float64().Min(0.0).Positive()
result, err := priceSchema.Parse(19.99)  // Valid: 19.99 (float64)
_, err = priceSchema.Parse(-5.0)         // Error: Number must be positive

// Different float types preserve exact type
result1, _ := Float32().Parse(float32(3.14))  // result1: float32(3.14)
result2, _ := gozod.Float64().Parse(42.5)           // result2: float64(42.5)
```

### Range Validation

```go
// Range constraints
scoreSchema := gozod.Int().Min(0).Max(100)
result, err := scoreSchema.Parse(85)   // Valid: 85
_, err = scoreSchema.Parse(-10)        // Error: Number must be at least 0
_, err = scoreSchema.Parse(110)        // Error: Number must be at most 100

// Comparison operators
positiveSchema := gozod.Float64().Gt(0.0)      // Greater than (exclusive)
result, err = positiveSchema.Parse(0.1)  // Valid: 0.1
_, err = positiveSchema.Parse(0.0)       // Error: Number must be greater than 0

teenSchema := gozod.Int().Gte(13).Lt(20)       // Range: 13 <= x < 20
result, err = teenSchema.Parse(16)       // Valid: 16
_, err = teenSchema.Parse(12)            // Error: Number must be at least 13
_, err = teenSchema.Parse(20)            // Error: Number must be less than 20
```

### Special Validations

```go
// Sign validations
positiveSchema := gozod.Float64().Positive()     // Must be > 0
negativeSchema := gozod.Float64().Negative()     // Must be < 0
nonNegativeSchema := gozod.Int().NonNegative()   // Must be >= 0
nonPositiveSchema := gozod.Int().NonPositive()   // Must be <= 0

// Mathematical constraints
evenSchema := gozod.Int().MultipleOf(2)          // Must be even
result, err := evenSchema.Parse(4)         // Valid: 4
_, err = evenSchema.Parse(3)               // Error: Number must be multiple of 2

centSchema := gozod.Float64().MultipleOf(0.01)   // Currency precision
result, err = centSchema.Parse(19.99)      // Valid: 19.99
_, err = centSchema.Parse(19.999)          // Error: not multiple of 0.01

// Float-specific validations
finiteSchema := gozod.Float64().Finite()         // No Inf or NaN
result, err = finiteSchema.Parse(42.0)     // Valid: 42.0
_, err = finiteSchema.Parse(math.Inf(1))   // Error: Number must be finite

safeSchema := gozod.Float64().Safe()             // Within safe integer range
result, err = safeSchema.Parse(42.0)       // Valid: 42.0
_, err = safeSchema.Parse(1e100)           // Error: Number exceeds safe range
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
// Basic boolean validation
activeSchema := gozod.Bool()
result, err := activeSchema.Parse(true)   // Valid: true
result, err = activeSchema.Parse(false)   // Valid: false
_, err = activeSchema.Parse("true")       // Error: wrong type
```

### String-to-Boolean Conversion

```go
// Convert strings to booleans
schema := gozod.StringBool(&StringBoolOptions{
    Truthy: []string{"yes", "true", "1", "on"},
    Falsy:  []string{"no", "false", "0", "off"},
    Case:   "insensitive",
})

result, err := schema.Parse("YES")    // Valid: true (bool)
result, err = schema.Parse("false")   // Valid: false (bool)
result, err = schema.Parse("maybe")   // Error: not recognized
```

---

## 🗂️ Object Validation

### Basic Object Schema

```go
// Define object schema
userSchema := gozod.Object(gozod.ObjectSchema{
    "name":  gozod.String().Min(2),
    "age":   gozod.Int().Min(0).Max(120),
    "email": gozod.String().Email().Optional(),
})

// Validate map data
userData := map[string]any{
    "name": "Alice",
    "age":  25,
    "email": "alice@example.com",
}

result, err := userSchema.Parse(userData)
// result: validated map, err: nil

// Missing optional field is okay
minimalData := map[string]any{
    "name": "Bob",
    "age":  30,
    // email is optional, so omitting it is fine
}

result, err = userSchema.Parse(minimalData)
// result: validated map, err: nil
```

### Object Operations

```go
// Base schema for examples
baseSchema := gozod.Object(gozod.ObjectSchema{
    "name":  gozod.String(),
    "age":   gozod.Int(),
    "email": gozod.String().Email(),
})

// Pick specific fields
nameAge := baseSchema.Pick([]string{"name", "age"})
// Result only validates "name" and "age" fields

// Omit fields
publicInfo := baseSchema.Omit([]string{"email"})
// Result validates all fields except "email"

// Make all fields optional
partialUser := baseSchema.Partial()
// All fields become optional

// Make specific fields required (after Partial)
requiredFields := partialUser.Required([]string{"name"})
// "name" becomes required again, others stay optional

// Extend with new fields
extendedUser := baseSchema.Extend(gozod.ObjectSchema{
    "phone": gozod.String(),
    "role":  gozod.String(),
})
// Adds new fields to the schema

// Merge with another schema
otherSchema := gozod.Object(gozod.ObjectSchema{
    "department": gozod.String(),
    "salary":     gozod.Float64(),
})
mergedSchema := baseSchema.Merge(otherSchema)
// Combines both schemas
```

### Object Modes

```go
// Strict mode (no extra fields allowed)
strictSchema := gozod.Object(gozod.ObjectSchema{
    "name": gozod.String(),
}).Strict()

_, err := strictSchema.Parse(map[string]any{
    "name":  "Alice",
    "extra": "field",  // Error: unexpected field
})

// Strip mode (remove extra fields)
stripSchema := gozod.Object(gozod.ObjectSchema{
    "name": gozod.String(),
}).Strip()

result, err := stripSchema.Parse(map[string]any{
    "name":  "Alice",
    "extra": "field",  // Silently removed
})
// result: {"name": "Alice"}

// Passthrough mode (allow extra fields)
passthroughSchema := gozod.Object(gozod.ObjectSchema{
    "name": gozod.String(),
}).Passthrough()

result, err = passthroughSchema.Parse(map[string]any{
    "name":  "Alice",
    "extra": "field",  // Preserved
})
// result: {"name": "Alice", "extra": "field"}

// Catchall (validate extra fields with schema)
catchallSchema := gozod.Object(gozod.ObjectSchema{
    "name": gozod.String(),
}).Catchall(gozod.String())
// Extra fields must be strings
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
| `.Strict()` | Reject extra fields | `schema.Strict()` |
| `.Strip()` | Remove extra fields | `schema.Strip()` |
| `.Passthrough()` | Allow extra fields | `schema.Passthrough()` |
| `.Catchall(schema)` | Validate extra fields | `schema.Catchall(gozod.String())` |

---

## 📋 Array & Slice Validation

### Slice Validation (Dynamic Length)

```go
// Basic slice validation
stringSlice := gozod.Slice(gozod.String().Min(2))
result, err := stringSlice.Parse([]any{"hello", "world"})
// result: []any{"hello", "world"}, err: nil

_, err = stringSlice.Parse([]any{"hello", "hi"})
// Error: "hi" is too short

// Size constraints
constrainedSlice := gozod.Slice(gozod.String()).Min(1).Max(5)
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

// Map-based enum
statusMap := gozod.EnumMap(map[string]string{
    "ACTIVE":   "active",
    "INACTIVE": "inactive",
})
result, err = statusMap.Parse("ACTIVE")  // Valid: "active" (returns mapped value)
```

### Go Native Enums

```go
// Go iota enum support
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

### Union Types (OR Logic)

```go
// Basic union
stringOrNumber := gozod.Union([]gozod.ZodType[any, any]{
    gozod.String(),
    gozod.Int(),
})

result, err := stringOrNumber.Parse("hello")  // Valid: matches string
result, err = stringOrNumber.Parse(42)        // Valid: matches int
_, err = stringOrNumber.Parse(true)           // Error: matches neither
```

### Discriminated Union

```go
// Efficient lookup by discriminator field
apiResponse := gozod.DiscriminatedUnion("status", []gozod.ZodType[any, any]{
    gozod.Object(gozod.ObjectSchema{
        "status": gozod.Literal("success"),
        "data":   gozod.String(),
    }),
    gozod.Object(gozod.ObjectSchema{
        "status": gozod.Literal("error"),
        "error":  gozod.String(),
        "code":   gozod.Int(),
    }),
})

// Efficient validation
successData := map[string]any{
    "status": "success",
    "data":   "Operation completed",
}
result, err := apiResponse.Parse(successData)
// result: validated success response, err: nil

errorData := map[string]any{
    "status": "error",
    "error":  "Something went wrong",
    "code":   500,
}
result, err = apiResponse.Parse(errorData)
// result: validated error response, err: nil
```

### Intersection Types (AND Logic)

```go
// Combine multiple schemas
personEmployee := gozod.Intersection(
    gozod.Object(gozod.ObjectSchema{
        "name": gozod.String(),
        "age":  gozod.Int(),
    }),
    gozod.Object(gozod.ObjectSchema{
        "employeeId": gozod.String(),
        "department": gozod.String(),
    }),
)

// Requires ALL fields from both schemas
validData := map[string]any{
    "name":       "Alice",
    "age":        30,
    "employeeId": "EMP001",
    "department": "Engineering",
}
result, err := personEmployee.Parse(validData)
// result: validated combined object, err: nil

// Missing any required field fails
invalidData := map[string]any{
    "name": "Alice",
    "age":  30,
    // Missing employeeId and department
}
_, err = personEmployee.Parse(invalidData)
// Error: missing required fields
```

### Union/Intersection Methods Reference

| Method | Description | Example |
|--------|-------------|---------|
| `Union(schemas)` | Accept one of multiple types | `gozod.Union([]ZodType{gozod.String(), gozod.Int()})` |
| `gozod.DiscriminatedUnion(key, schemas)` | Union with discriminator field | `gozod.DiscriminatedUnion("type", schemas)` |
| `gozod.Intersection(schemas...)` | Require all schemas | `gozod.Intersection(schema1, schema2)` |

---

## 🔧 Modifiers & Wrappers

### Optional (Allows Missing Values)

```go
// Optional modifier
optionalEmail := gozod.String().Email().Optional()

result, err := optionalEmail.Parse("user@example.com")  // Valid: email
result, err = optionalEmail.Parse(nil)                  // Valid: nil (missing)
_, err = optionalEmail.Parse(123)                       // Error: wrong type
```

### Nilable (Typed Nil Pointers)

```go
// Nilable modifier
nilableAge := gozod.String().Nilable()

result, err := nilableAge.Parse("hello")  // Valid: "hello"
result, err = nilableAge.Parse(nil)       // Valid: (*string)(nil)
_, err = nilableAge.Parse(123)            // Error: wrong type

// Key difference: Nilable returns typed nil pointers
// result type for nil input: (*string)(nil)
// Optional returns generic nil for nil input
```

### Optional vs Nilable Distinction

**Key Semantic Difference**:

```go
// Understanding the difference
nilableSchema := gozod.String().Nilable()
optionalSchema := gozod.String().Optional()

// DIFFERENT nil handling semantics:
nilableResult, _ := nilableSchema.Parse(nil)  // (*string)(nil) - typed nil pointer
optionalResult, _ := optionalSchema.Parse(nil) // nil - generic any nil

// Use cases:
// - Optional: JSON fields that might be absent → returns generic nil
// - Nilable: JSON fields that might be explicitly null → returns typed nil pointer

// Both accept valid values identically
nilableResult, _ = nilableSchema.Parse("hello")  // "hello"
optionalResult, _ = optionalSchema.Parse("hello") // "hello"

// Both reject invalid types consistently
_, nilableErr := nilableSchema.Parse(123)   // Error: Expected string, received number
_, optionalErr := optionalSchema.Parse(123) // Error: Expected string, received number
```

### Default Values

```go
// Static default value
nameWithDefault := gozod.String().Default("Anonymous")
result, err := nameWithDefault.Parse(nil)       // Valid: "Anonymous"
result, err = nameWithDefault.Parse("Alice")    // Valid: "Alice"

// Function-based default
timestampDefault := gozod.String().DefaultFunc(func() string {
    return time.Now().Format(time.RFC3339)
})
result, err = timestampDefault.Parse(nil)  // Valid: current timestamp
```

### Prefault (Fallback on Validation Failure)

```go
// Prefault provides fallback for ANY validation failure
safeValue := gozod.String().Min(5).Prefault("fallback")

result, err := safeValue.Parse("hello world")  // Valid: "hello world" (passes validation)
result, err = safeValue.Parse("hi")            // Valid: "fallback" (too short, uses fallback)
result, err = safeValue.Parse(123)             // Valid: "fallback" (wrong type, uses fallback)
result, err = safeValue.Parse(nil)             // Valid: "fallback" (nil, uses fallback)

// Key behavior: Prefault ALWAYS succeeds by using fallback when validation fails
```

### Default vs Prefault Distinction

```go
// Understanding the difference
prefaultSchema := gozod.String().Prefault("prefault_value")
defaultSchema := gozod.String().Default("default_value")

// For nil input: different behaviors
result1, _ := prefaultSchema.Parse(nil)  // "prefault_value" (nil fails validation, use fallback)
result2, _ := defaultSchema.Parse(nil)   // "default_value" (nil gets default value)

// For invalid type: different behaviors
result3, _ := prefaultSchema.Parse(123)  // "prefault_value" (validation fails, use fallback)
_, err := defaultSchema.Parse(123)       // Error (type validation fails, no fallback for non-nil)
```

### Modifier Methods Reference

| Method | Description | Example |
|--------|-------------|---------|
| `.Optional()` | Allow missing values | `gozod.String().Optional()` |
| `.Nilable()` | Handle explicit null | `gozod.String().Nilable()` |
| `.Default(value)` | Static default | `gozod.String().Default("Anonymous")` |
| `.DefaultFunc(fn)` | Dynamic default | `gozod.String().DefaultFunc(func() string {...})` |
| `.Prefault(value)` | Fallback on failure | `gozod.String().Prefault("fallback")` |

---

## 🔄 Transform & Pipe

### Transform (Data Transformation)

```go
// Transform modifies data after validation
upperSchema := gozod.String().Transform(func(s string, ctx *core.RefinementContext) (any, error) {
    return strings.ToUpper(s), nil
})

result, err := upperSchema.Parse("hello")  // Valid: "HELLO"

// Transform with validation
processedSchema := gozod.String().Min(3).Transform(func(s string, ctx *core.RefinementContext) (any, error) {
    return fmt.Sprintf("processed: %s", s), nil
})

result, err = processedSchema.Parse("hello")  // Valid: "processed: hello"
_, err = processedSchema.Parse("hi")          // Error: fails Min(3) before transform
```

### Pipe (Schema Chaining)

```go
// Pipe chains validation and transformation
pipeline := gozod.String().
    Transform(func(s string, ctx *core.RefinementContext) (any, error) {
        return strings.TrimSpace(s), nil // Clean input first
    }).
    Pipe(gozod.String().Min(3)) // Then validate cleaned string

result, err := pipeline.Parse("  hello  ")  // Valid: "hello"
_, err = pipeline.Parse("  hi  ")           // Error: trimmed "hi" fails Min(3)

// Pipeline example
stringToNumberPipe := gozod.String().
    Transform(func(s string, ctx *core.RefinementContext) (any, error) {
        return strconv.ParseFloat(s, 64)
    }).
    Pipe(gozod.Float64().Min(0)) // Validate as positive number

result, err = stringToNumberPipe.Parse("42.5")  // Valid: 42.5
_, err = stringToNumberPipe.Parse("-10")         // Error: negative number
```

### Transform vs Refine

```go
// Understanding the difference
input := "hello"

// Refine: only validates, never modifies
refineSchema := gozod.String().Refine(func(s string) bool {
    return len(s) > 0
})
refineResult, _ := refineSchema.Parse(input)  // "hello" (unchanged)

// Transform: modifies the data
transformSchema := gozod.String().Transform(func(s string, ctx *core.RefinementContext) (any, error) {
    return strings.ToUpper(s), nil
})
transformResult, _ := transformSchema.Parse(input)  // "HELLO" (modified)

// Key distinction: Refine preserves, Transform modifies
```

### Transform/Pipe Methods Reference

| Method | Description | Example |
|--------|-------------|---------|
| `.Transform(fn)` | Modify data after validation | `gozod.String().Transform(func(s string, ctx *core.RefinementContext) (any, error) {...})` |
| `.Pipe(schema)` | Chain to another schema | `gozod.String().Transform(...).Pipe(gozod.String().Min(3))` |
| `.Refine(fn)` | Custom validation only | `gozod.String().Refine(func(s string) bool {...})` |

---

## 🔀 Type Coercion

### Coerce Namespace

```go
// Basic type coercion
stringSchema := gozod.Coerce.String()
result, _ := stringSchema.Parse(123)     // "123"
result, _ = stringSchema.Parse(true)     // "true"
result, _ = stringSchema.Parse(3.14)     // "3.14"

numberSchema := gozod.Coerce.Number()
result, _ = numberSchema.Parse("42")     // 42.0
result, _ = numberSchema.Parse(true)     // 1.0

boolSchema := gozod.Coerce.Bool()
result, _ = boolSchema.Parse("false")    // false
result, _ = boolSchema.Parse(0)          // false
result, _ = boolSchema.Parse("1")        // true

// Big integer coercion
bigIntSchema := gozod.Coerce.BigInt()
result, _ = bigIntSchema.Parse("9223372036854775808")  // *big.Int
```

### Type-Safe Collection Coercion

```go
// Record coercion - returns type-safe map[string]T
recordSchema := gozod.Coerce.Record(gozod.String(), gozod.Int())
result, _ := recordSchema.Parse(map[string]any{
    "age":   "25", // String coerced to int
    "score": 95,
})
// Result type: map[string]int

// Map coercion - returns type-safe map[K]V
mapSchema := gozod.Coerce.Map(gozod.String(), gozod.Int())
result, _ = mapSchema.Parse(map[any]any{
    "key1": "10", // Coerced to map[string]int
    "key2": 20,
})

// Object coercion - supports struct input
type Person struct {
    Name string `json:"name"`
    Age  int    `json:"age"`
}
objectSchema := gozod.Coerce.gozod.Object(gozod.ObjectSchema{
    "name": gozod.String(),
    "age":  gozod.Int(),
})
result, _ = objectSchema.Parse(Person{Name: "Alice", Age: 25})
// Converts struct to map[string]any
```

### Schema-Level Coercion

```go
// Enable coercion via parameters
coerceSchema := gozod.String(core.SchemaParams{Coerce: true})
result, _ := coerceSchema.Parse(123)  // "123"

// Coercion with validation
validatedCoerce := gozod.Float64(gozod.SchemaParams{Coerce: true}).Min(0.0)
result, _ = validatedCoerce.Parse("3.14")  // 3.14 (coerced then validated)
_, err := validatedCoerce.Parse("-1")      // Error: coerced to -1.0, fails Min(0.0)
```

### Coercion Methods Reference

| Method | Description | Example |
|--------|-------------|---------|
| `gozod.Coerce.String()` | Coerce to string | `gozod.Coerce.String().Parse(123)` → `"123"` |
| `gozod.Coerce.Number()` | Coerce to number | `gozod.Coerce.Number().Parse("42")` → `42.0` |
| `gozod.Coerce.Bool()` | Coerce to boolean | `gozod.Coerce.Bool().Parse("true")` → `true` |
| `gozod.Coerce.BigInt()` | Coerce to big.Int | `gozod.Coerce.BigInt().Parse("123")` → `*big.Int` |
| `gozod.Coerce.Record(K, V)` | Coerce to map[string]T | `gozod.Coerce.Record(gozod.String(), gozod.Int())` |
| `gozod.Coerce.Map(K, V)` | Coerce to map[K]V | `gozod.Coerce.Map(gozod.String(), gozod.Int())` |
| `gozod.Coerce.Object(schema)` | Coerce to object | `gozod.Coerce.gozod.Object(gozod.ObjectSchema{...})` |

---

This API reference provides complete type interface documentation for GoZod. For usage patterns and practical examples, see the [Basics Guide](basics.md). For error handling and customization, see the [Error Customization Guide](error-customization.md).

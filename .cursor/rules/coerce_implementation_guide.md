# GoZod Coerce Mechanism Implementation Guide

## 📋 Overview

This document provides detailed implementation of GoZod Coerce (type coercion) mechanism, including architectural design, file organization, implementation steps, and best practices. The mechanism is based on TypeScript Zod v4 design principles but fully adapted to Go's type system and conventions.

## 🎯 Core Design Principles

### 1. Pre-Parse Type Coercion
- Coercion happens before validation
- Input types are converted to expected types when possible
- Failed coercion falls back to original validation

### 2. TypeScript Compatibility
- Follows TypeScript Zod's coercion behavior exactly
- `z.coerce.string()` corresponds to `gozod.Coerce.String()`
- Maintains same conversion rules and edge cases

### 3. Performance Optimization
- Uses efficient type conversion libraries (spf13/cast)
- Minimal allocation overhead
- Early exit for already-correct types

## 🏗️ Architecture Components

### 1. Coerce Namespace (`coerce.go`)

**Responsibility**: Provides the coercion namespace and factory functions

```go
// CoerceNamespace provides the coerce functionality namespace
// Corresponds to TypeScript Zod's z.coerce namespace
type CoerceNamespace struct{}

// Global instance for the coerce namespace
var Coerce = &CoerceNamespace{}

// String creates a coercive string schema
func (c *CoerceNamespace) String(params ...SchemaParams) *ZodString {
    schema := String(params...)
    internals := schema.GetInternals()
    internals.Coerce = true
    return schema
}

// Bool creates a coercive boolean schema
func (c *CoerceNamespace) Bool(params ...SchemaParams) *ZodBoolean {
    schema := Bool(params...)
    internals := schema.GetInternals()
    internals.Coerce = true
    return schema
}

// Number creates a coercive number schema (float64)
func (c *CoerceNamespace) Number(params ...SchemaParams) *ZodFloat64 {
    schema := Number(params...)
    internals := schema.GetInternals()
    internals.Coerce = true
    return schema
}

// BigInt creates a coercive big integer schema
func (c *CoerceNamespace) BigInt(params ...SchemaParams) *ZodBigInt {
    schema := BigInt(params...)
    internals := schema.GetInternals()
    internals.Coerce = true
    return schema
}

// Date creates a coercive date schema
func (c *CoerceNamespace) Date(params ...SchemaParams) *ZodDate {
    schema := Date(params...)
    internals := schema.GetInternals()
    internals.Coerce = true
    return schema
}
```

### 2. Coercible Interface (`type.go`)

**Responsibility**: Defines the interface for types that support coercion

```go
// Coercible interface for types that support coercion
type Coercible interface {
    Coerce(input interface{}) (output interface{}, success bool)
}
```

### 3. Type-Specific Coercion Implementation

**Responsibility**: Implements coercion for each type

```go
// ZodString coercion implementation
func (z *ZodString) Coerce(input interface{}) (interface{}, bool) {
    if str, ok := coerceToString(input); ok {
        return str, true
    }
    return input, false
}

// ZodBoolean coercion implementation
func (z *ZodBoolean) Coerce(input interface{}) (interface{}, bool) {
    if val, ok := coerceToBool(input); ok {
        return val, true
    }
    return input, false
}

// ZodFloat64 coercion implementation
func (z *ZodFloat64) Coerce(input interface{}) (interface{}, bool) {
    if val, ok := coerceToFloat64(input); ok {
        return val, true
    }
    return input, false
}

// ZodBigInt coercion implementation
func (z *ZodBigInt) Coerce(input interface{}) (interface{}, bool) {
    if val, ok := coerceToBigInt(input); ok {
        return val, true
    }
    return input, false
}
```

### 4. Core Coercion Functions (`utils.go`)

**Responsibility**: Provides the actual type conversion logic

```go
// coerceToString attempts to convert a value to string using spf13/cast
func coerceToString(value interface{}) (string, bool) {
    if value == nil {
        return "", false
    }

    // Use spf13/cast for robust string conversion
    str, err := cast.ToStringE(value)
    if err != nil {
        return "", false
    }

    return str, true
}

// coerceToBool attempts to convert a value to boolean using spf13/cast
func coerceToBool(value interface{}) (bool, bool) {
    if value == nil {
        return false, false
    }

    // Check for nil pointers - reject them
    if reflect.ValueOf(value).Kind() == reflect.Ptr && reflect.ValueOf(value).IsNil() {
        return false, false
    }

    // Handle string values explicitly for better control
    if str, ok := value.(string); ok {
        // Trim whitespace and convert to lowercase for comparison
        str = strings.TrimSpace(strings.ToLower(str))
        // Empty string after trimming should be false
        if str == "" {
            return false, true
        }
        switch str {
        case "true", "1", "yes", "on", "enabled", "y":
            return true, true
        case "false", "0", "no", "off", "disabled", "n":
            return false, true
        default:
            return false, false
        }
    }

    // Use spf13/cast for other types
    val, err := cast.ToBoolE(value)
    if err != nil {
        return false, false
    }

    return val, true
}

// coerceToFloat64 attempts to convert a value to float64 using spf13/cast
func coerceToFloat64(value interface{}) (float64, bool) {
    if value == nil {
        return 0, false
    }

    // Check for nil pointers - reject them
    if reflect.ValueOf(value).Kind() == reflect.Ptr && reflect.ValueOf(value).IsNil() {
        return 0, false
    }

    // Use spf13/cast for robust float64 conversion
    val, err := cast.ToFloat64E(value)
    if err != nil {
        return 0, false
    }

    return val, true
}

// coerceToBigInt attempts to coerce a value to big.Int
func coerceToBigInt(value interface{}) (*big.Int, bool) {
    if value == nil {
        return nil, false
    }

    switch v := value.(type) {
    case *big.Int:
        return new(big.Int).Set(v), true
    case bool:
        // Convert boolean to BigInt (JavaScript BigInt() behavior)
        if v {
            return big.NewInt(1), true
        }
        return big.NewInt(0), true
    case string:
        // Trim whitespace
        v = strings.TrimSpace(v)
        if v == "" {
            return nil, false // Reject empty strings
        }
        // Check if string contains decimal point
        for _, char := range v {
            if char == '.' {
                return nil, false // Reject strings with decimal points
            }
        }
        // Try to convert to big.Int
        if result, ok := new(big.Int).SetString(v, 10); ok {
            return result, true
        }
        return nil, false
    case float32, float64:
        // For floats, only accept if they are whole numbers
        floatVal := toFloat64(v)
        if floatVal == float64(int64(floatVal)) {
            return big.NewInt(int64(floatVal)), true
        }
        return nil, false
    default:
        // Try to convert to int64 first using cast
        if intVal, err := cast.ToInt64E(value); err == nil {
            return big.NewInt(intVal), true
        }

        // Try string conversion for very large numbers
        if strVal, err := cast.ToStringE(value); err == nil {
            if result, ok := new(big.Int).SetString(strVal, 10); ok {
                return result, true
            }
        }

        return nil, false
    }
}
```

## 🔧 Implementation Steps

### Step 1: Implement Coerce Namespace

Create the global namespace that provides coercive constructors for all supported types.

### Step 2: Add Coercible Interface Support

Implement the Coercible interface for all types that support coercion.

### Step 3: Integrate Coercion into Parse Methods

Modify Parse methods to check for coercion flag and apply coercion before validation.

### Step 4: Add Core Conversion Functions

Implement robust type conversion functions using established libraries.

## 🧪 Testing Strategy

### 1. Basic Coercion Tests

```go
func TestCoerceString(t *testing.T) {
    schema := gozod.Coerce.String()
    
    t.Run("coerces number to string", func(t *testing.T) {
        result, err := schema.Parse(123)
        assert.NoError(t, err)
        assert.Equal(t, "123", result)
    })
    
    t.Run("coerces boolean to string", func(t *testing.T) {
        result, err := schema.Parse(true)
        assert.NoError(t, err)
        assert.Equal(t, "true", result)
    })
    
    t.Run("keeps string as string", func(t *testing.T) {
        result, err := schema.Parse("hello")
        assert.NoError(t, err)
        assert.Equal(t, "hello", result)
    })
}

func TestCoerceBool(t *testing.T) {
    schema := gozod.Coerce.Bool()
    
    testCases := []struct {
        name     string
        input    any
        expected bool
        hasError bool
    }{
        {"string true", "true", true, false},
        {"string false", "false", false, false},
        {"string 1", "1", true, false},
        {"string 0", "0", false, false},
        {"number 1", 1, true, false},
        {"number 0", 0, false, false},
        {"boolean true", true, true, false},
        {"boolean false", false, false, false},
        {"invalid string", "maybe", false, true},
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            result, err := schema.Parse(tc.input)
            if tc.hasError {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tc.expected, result)
            }
        })
    }
}
```

### 2. TypeScript Compatibility Tests

```go
func TestTypeScriptCompatibility_Coerce(t *testing.T) {
    t.Run("z.coerce.string() behavior", func(t *testing.T) {
        schema := gozod.Coerce.String()
        
        // Test cases that should match TypeScript Zod exactly
        testCases := []struct {
            input    any
            expected string
        }{
            {123, "123"},
            {123.45, "123.45"},
            {true, "true"},
            {false, "false"},
            {"hello", "hello"},
        }
        
        for _, tc := range testCases {
            result, err := schema.Parse(tc.input)
            assert.NoError(t, err)
            assert.Equal(t, tc.expected, result)
        }
    })
    
    t.Run("z.coerce.number() behavior", func(t *testing.T) {
        schema := gozod.Coerce.Number()
        
        testCases := []struct {
            input    any
            expected float64
        }{
            {"123", 123.0},
            {"123.45", 123.45},
            {true, 1.0},
            {false, 0.0},
            {123, 123.0},
        }
        
        for _, tc := range testCases {
            result, err := schema.Parse(tc.input)
            assert.NoError(t, err)
            assert.Equal(t, tc.expected, result)
        }
    })
}
```

### 3. Edge Case Tests

```go
func TestCoercionEdgeCases(t *testing.T) {
    t.Run("nil values", func(t *testing.T) {
        stringSchema := gozod.Coerce.String()
        _, err := stringSchema.Parse(nil)
        assert.Error(t, err) // Should fail, nil can't be coerced to string
        
        boolSchema := gozod.Coerce.Bool()
        _, err = boolSchema.Parse(nil)
        assert.Error(t, err) // Should fail, nil can't be coerced to bool
    })
    
    t.Run("invalid coercions", func(t *testing.T) {
        boolSchema := gozod.Coerce.Bool()
        _, err := boolSchema.Parse("invalid")
        assert.Error(t, err) // Should fail, "invalid" can't be coerced to bool
        
        numberSchema := gozod.Coerce.Number()
        _, err = numberSchema.Parse("not-a-number")
        assert.Error(t, err) // Should fail, "not-a-number" can't be coerced to number
    })
}
```

## 📖 Usage Examples

### Basic Coercion Usage

```go
// String coercion
stringSchema := gozod.Coerce.String()
result, err := stringSchema.Parse(123) // Returns "123"

// Boolean coercion
boolSchema := gozod.Coerce.Bool()
result, err := boolSchema.Parse("true") // Returns true
result, err := boolSchema.Parse(1)      // Returns true
result, err := boolSchema.Parse("0")    // Returns false

// Number coercion
numberSchema := gozod.Coerce.Number()
result, err := numberSchema.Parse("123.45") // Returns 123.45
result, err := numberSchema.Parse(true)      // Returns 1.0
```

### Coercion with Validation

```go
// Coerce to string then validate length
schema := gozod.Coerce.String().Min(5).Max(10)
result, err := schema.Parse(12345) // Returns "12345"
result, err := schema.Parse(123)   // Error: "123" is too short

// Coerce to number then validate range
schema := gozod.Coerce.Number().Min(0).Max(100)
result, err := schema.Parse("50")  // Returns 50.0
result, err := schema.Parse("150") // Error: 150 is too large
```

### Complex Data Processing

```go
// Process form data with coercion
type FormData struct {
    Name  string `json:"name"`
    Age   int    `json:"age"`
    Email string `json:"email"`
}

schema := gozod.Object(map[string]gozod.ZodType[any, any]{
    "name":  gozod.Coerce.String().Min(1),
    "age":   gozod.Coerce.Number().Int().Min(0).Max(120),
    "email": gozod.Coerce.String().Email(),
})

// Raw form data (all strings)
formInput := map[string]any{
    "name":  "John Doe",
    "age":   "25",      // String will be coerced to number
    "email": "john@example.com",
}

result, err := schema.Parse(formInput)
// Result will have age as number 25, not string "25"
```

This implementation guide provides the complete foundation for implementing the Coerce mechanism in GoZod while maintaining full TypeScript compatibility and Go type safety. 

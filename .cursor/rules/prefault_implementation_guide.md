# GoZod Prefault Mechanism Implementation Guide

## 📋 Overview

This document provides detailed implementation of GoZod Prefault (fallback value) mechanism, including architectural design, file organization, implementation steps, and best practices. The mechanism is based on TypeScript Zod v4 design principles but fully adapted to Go's type system and conventions.

## 🎯 Core Design Principles

### 1. Type Safety First
- Provides compile-time type-safe Prefault methods
- `String().Prefault(10)` should fail compilation, `String().Prefault("hello")` should pass
- Avoids type loss caused by using `any` in interfaces

### 2. Validation-First Mechanism
- Unlike Default (uses default value when nil), Prefault always tries validation first
- Uses fallback value only when validation fails, returns original value when validation succeeds
- For nil input, Prefault should reject and try fallback value

### 3. Wrapper Pattern Design
- Uses dedicated wrapper types to provide chaining support
- Each type has its own Prefault wrapper (e.g., `ZodStringPrefault`)
- Maintains consistency with existing `ZodDefault` wrapper patterns

## 🏗️ Architecture Components

### 1. Core Wrapper (`type_prefault.go`)

**Responsibility**: Provides generic Prefault wrapper implementation

```go
// ZodPrefault represents a validation schema with fallback value (new type-safe version)
// Core design: contains inner type, obtains all its methods through forwarding
// Based on ZodDefault's successful pattern, provides type-safe Prefault functionality
type ZodPrefault[T ZodType[any, any]] struct {
    internals     *ZodTypeInternals // Prefault's own internals, Type = "prefault"
    innerType     T                 // Inner type
    prefaultValue any               // Fallback value
    prefaultFunc  func() any        // Fallback function
    isFunction    bool              // Whether to use function for fallback value
}

// Parse performs smart type inference validation and parsing
// Based on TypeScript Zod v4 mechanism:
// - 🔥 Prefault core mechanism: try validation first, use fallback value on failure
// - Unlike Default (uses default value when nil), Prefault always validates first
// - Important: for nil input, Prefault should reject and try fallback value
func (z *ZodPrefault[T]) Parse(input any, ctx ...*ParseContext) (any, error) {
    var contextToUse *ParseContext
    if len(ctx) > 0 {
        contextToUse = ctx[0]
    }
    
    // 🔥 Core mechanism: try validation first
    result, err := z.innerType.Parse(input, contextToUse)
    if err == nil {
        return result, nil // Validation succeeded, return original value
    }
    
    // Validation failed, use fallback value
    var fallbackValue any
    if z.isFunction && z.prefaultFunc != nil {
        fallbackValue = z.prefaultFunc()
    } else {
        fallbackValue = z.prefaultValue
    }
    
    // Validate fallback value
    fallbackResult, fallbackErr := z.innerType.Parse(fallbackValue, contextToUse)
    if fallbackErr != nil {
        return nil, fallbackErr
    }
    
    return fallbackResult, nil
}
```

### 2. Constructor Functions (`type_prefault.go`)

**Responsibility**: Provides unified Prefault constructor functions

```go
func newZodPrefault[T ZodType[any, any]](innerType T, value any, fn func() any, isFunc bool) *ZodPrefault[T] {
    // Construct Prefault's internals, Type = "prefault", copy inner type's checks/coerce/optional/nilable
    baseInternals := innerType.GetInternals()
    internals := &ZodTypeInternals{
        Version:     baseInternals.Version,
        Type:        ZodTypePrefault,
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
    
    return &ZodPrefault[T]{
        internals:     internals,
        innerType:     innerType,
        prefaultValue: value,
        prefaultFunc:  fn,
        isFunction:    isFunc,
    }
}
```

### 3. Concrete Type Implementation (various type files)

**Responsibility**: Each type provides type-safe Prefault methods

```go
// type_string.go - String type Prefault implementation
type ZodStringPrefault struct {
    *ZodPrefault[*ZodString] // Embed concrete pointer, enables method promotion
}

// Prefault type-safe version - only accepts string type
// Compile-time type safety: String().Prefault(10) will fail compilation
func (z *ZodString) Prefault(value string) ZodStringPrefault {
    // Construct Prefault's internals, Type = "prefault", copy inner type's checks/coerce/optional/nilable
    baseInternals := z.GetInternals()
    internals := &ZodTypeInternals{
        Version:     baseInternals.Version,
        Type:        ZodTypePrefault,
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

    return ZodStringPrefault{
        &ZodPrefault[*ZodString]{
            internals:     internals,
            innerType:     z,
            prefaultValue: value,
            prefaultFunc:  nil,
            isFunction:    false,
        },
    }
}

// PrefaultFunc type-safe version - only accepts functions returning string
func (z *ZodString) PrefaultFunc(fn func() string) ZodStringPrefault {
    genericFn := func() any { return fn() }

    // Construct Prefault's internals, Type = "prefault", copy inner type's checks/coerce/optional/nilable
    baseInternals := z.GetInternals()
    internals := &ZodTypeInternals{
        Version:     baseInternals.Version,
        Type:        ZodTypePrefault,
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

    return ZodStringPrefault{
        &ZodPrefault[*ZodString]{
            internals:     internals,
            innerType:     z,
            prefaultValue: nil,
            prefaultFunc:  genericFn,
            isFunction:    true,
        },
    }
}
```

### 4. Package-Level Functions (`type_prefault.go`)

**Responsibility**: Provides generic Prefault creation functions

```go
// Prefault creates a schema wrapper with fallback value (backward compatible)
func Prefault[In, Out any](innerType ZodType[In, Out], prefaultValue any) ZodType[any, any] {
    return newZodPrefault(any(innerType).(ZodType[any, any]), prefaultValue, nil, false)
}

// PrefaultFunc creates a schema wrapper with function fallback value (backward compatible)
func PrefaultFunc[In, Out any](innerType ZodType[In, Out], fn func() any) ZodType[any, any] {
    return newZodPrefault(any(innerType).(ZodType[any, any]), nil, fn, true)
}
```

## 🔧 Implementation Steps

### Step 1: Define Wrapper Types

Define dedicated Prefault wrappers in type files:

```go
// type_string.go
// ZodStringPrefault is the Prefault wrapper for string type
// Provides perfect type safety and chaining support
type ZodStringPrefault struct {
    *ZodPrefault[*ZodString] // Embed concrete pointer, enables method promotion
}
```

### Step 2: Implement Type-Safe Prefault Methods

```go
// Prefault type-safe version - only accepts string type
// Compile-time type safety: String().Prefault(10) will fail compilation
func (z *ZodString) Prefault(value string) ZodStringPrefault {
    return ZodStringPrefault{
        &ZodPrefault[*ZodString]{
            innerType:     z,
            prefaultValue: value,
            isFunction:    false,
        },
    }
}

// PrefaultFunc type-safe version - only accepts functions returning string
func (z *ZodString) PrefaultFunc(fn func() string) ZodStringPrefault {
    genericFn := func() any { return fn() }
    return ZodStringPrefault{
        &ZodPrefault[*ZodString]{
            innerType:   z,
            prefaultFunc: genericFn,
            isFunction:  true,
        },
    }
}
```

### Step 3: Implement Chaining Methods

Implement all necessary chaining methods for the wrapper:

```go
// Min adds minimum length validation, returns ZodStringPrefault for chaining
func (s ZodStringPrefault) Min(min int, params ...SchemaParams) ZodStringPrefault {
    newInner := s.innerType.Min(min, params...)
    return ZodStringPrefault{
        &ZodPrefault[*ZodString]{
            innerType:     newInner,
            prefaultValue: s.prefaultValue,
            prefaultFunc:  s.prefaultFunc,
            isFunction:    s.isFunction,
        },
    }
}

// Max adds maximum length validation, returns ZodStringPrefault for chaining
func (s ZodStringPrefault) Max(max int, params ...SchemaParams) ZodStringPrefault {
    newInner := s.innerType.Max(max, params...)
    return ZodStringPrefault{
        &ZodPrefault[*ZodString]{
            innerType:     newInner,
            prefaultValue: s.prefaultValue,
            prefaultFunc:  s.prefaultFunc,
            isFunction:    s.isFunction,
        },
    }
}

// Email adds email validation, returns ZodStringPrefault for chaining
func (s ZodStringPrefault) Email(params ...SchemaParams) ZodStringPrefault {
    newInner := s.innerType.Email(params...)
    return ZodStringPrefault{
        &ZodPrefault[*ZodString]{
            innerType:     newInner,
            prefaultValue: s.prefaultValue,
            prefaultFunc:  s.prefaultFunc,
            isFunction:    s.isFunction,
        },
    }
}
```

### Step 4: Implement Interface Methods

Ensure wrapper types implement required interfaces:

```go
// GetInternals implements ZodType interface
func (s ZodStringPrefault) GetInternals() *ZodTypeInternals {
    return s.ZodPrefault.GetInternals()
}

// Parse implements ZodType interface
func (s ZodStringPrefault) Parse(input any, ctx ...*ParseContext) (any, error) {
    return s.ZodPrefault.Parse(input, ctx...)
}

// MustParse implements ZodType interface
func (s ZodStringPrefault) MustParse(input any, ctx ...*ParseContext) any {
    result, err := s.Parse(input, ctx...)
    if err != nil {
        panic(err)
    }
    return result
}
```

## 🧪 Testing Strategy

### 1. Basic Prefault Tests

```go
func TestZodStringPrefault_BasicFunctionality(t *testing.T) {
    schema := gozod.String().Min(5).Prefault("fallback")
    
    t.Run("uses fallback for invalid input", func(t *testing.T) {
        result, err := schema.Parse("hi") // Too short, should use fallback
        assert.NoError(t, err)
        assert.Equal(t, "fallback", result)
    })
    
    t.Run("preserves valid input", func(t *testing.T) {
        result, err := schema.Parse("hello")
        assert.NoError(t, err)
        assert.Equal(t, "hello", result)
    })
    
    t.Run("uses fallback for nil input", func(t *testing.T) {
        result, err := schema.Parse(nil)
        assert.NoError(t, err)
        assert.Equal(t, "fallback", result)
    })
}

func TestZodStringPrefault_FunctionFallback(t *testing.T) {
    counter := 0
    schema := gozod.String().Min(5).PrefaultFunc(func() string {
        counter++
        return fmt.Sprintf("fallback-%d", counter)
    })
    
    t.Run("calls function for each fallback use", func(t *testing.T) {
        result1, err1 := schema.Parse("hi") // Invalid, uses fallback
        assert.NoError(t, err1)
        assert.Equal(t, "fallback-1", result1)
        
        result2, err2 := schema.Parse("x") // Invalid, uses fallback
        assert.NoError(t, err2)
        assert.Equal(t, "fallback-2", result2)
        
        result3, err3 := schema.Parse("hello") // Valid, doesn't use fallback
        assert.NoError(t, err3)
        assert.Equal(t, "hello", result3)
        assert.Equal(t, 2, counter) // Counter should still be 2
    })
}
```

### 2. Type Safety Tests

```go
func TestZodStringPrefault_TypeSafety(t *testing.T) {
    // These should compile successfully
    schema1 := gozod.String().Prefault("hello")
    schema2 := gozod.String().PrefaultFunc(func() string { return "world" })
    
    // These should fail compilation:
    // schema3 := gozod.String().Prefault(123)           // Wrong type
    // schema4 := gozod.String().PrefaultFunc(func() int { return 123 }) // Wrong return type
    
    t.Run("valid prefaults work", func(t *testing.T) {
        result1, err1 := schema1.Parse(nil)
        assert.NoError(t, err1)
        assert.Equal(t, "hello", result1)
        
        result2, err2 := schema2.Parse(nil)
        assert.NoError(t, err2)
        assert.Equal(t, "world", result2)
    })
}
```

### 3. Validation vs Fallback Tests

```go
func TestZodStringPrefault_ValidationVsFallback(t *testing.T) {
    schema := gozod.String().Min(5).Max(10).Prefault("default")
    
    testCases := []struct {
        name     string
        input    any
        expected string
        usesFallback bool
    }{
        {"valid input", "hello", "hello", false},
        {"too short", "hi", "default", true},
        {"too long", "very long string", "default", true},
        {"nil input", nil, "default", true},
        {"wrong type", 123, "default", true},
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            result, err := schema.Parse(tc.input)
            assert.NoError(t, err)
            assert.Equal(t, tc.expected, result)
        })
    }
}
```

### 4. TypeScript Compatibility Tests

```go
func TestTypeScriptCompatibility_Prefault(t *testing.T) {
    // Equivalent to: z.string().min(5).catch("fallback")
    schema := gozod.String().Min(5).Prefault("fallback")
    
    testCases := []struct {
        name     string
        input    any
        expected string
        hasError bool
    }{
        {"valid input", "hello", "hello", false},
        {"invalid input uses fallback", "hi", "fallback", false},
        {"nil uses fallback", nil, "fallback", false},
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

## 📖 Usage Examples

### Basic Prefault Usage

```go
// Static fallback value
schema := gozod.String().Min(5).Prefault("default")
result, err := schema.Parse("hello") // Returns "hello" (valid)
result, err := schema.Parse("hi")    // Returns "default" (too short)
result, err := schema.Parse(nil)     // Returns "default" (nil)

// Function fallback value
schema := gozod.String().Min(5).PrefaultFunc(func() string {
    return "generated-" + uuid.New().String()[:8]
})
result, err := schema.Parse("valid") // Returns "valid"
result, err := schema.Parse("x")     // Returns "generated-12345678"
```

### Prefault vs Default Comparison

```go
// Default: uses default value when input is nil/undefined
defaultSchema := gozod.String().Default("default")
result, err := defaultSchema.Parse(nil)    // Returns "default"
result, err := defaultSchema.Parse("hi")   // Returns "hi" (no validation)

// Prefault: uses fallback value when validation fails
prefaultSchema := gozod.String().Min(5).Prefault("fallback")
result, err := prefaultSchema.Parse(nil)   // Returns "fallback" (validation failed)
result, err := prefaultSchema.Parse("hi")  // Returns "fallback" (validation failed)
result, err := prefaultSchema.Parse("hello") // Returns "hello" (validation passed)
```

### Complex Validation with Fallback

```go
// Email validation with fallback
emailSchema := gozod.String().Email().Prefault("admin@example.com")
result, err := emailSchema.Parse("user@domain.com") // Returns "user@domain.com"
result, err := emailSchema.Parse("invalid-email")   // Returns "admin@example.com"

// Number validation with fallback
numberSchema := gozod.Number().Min(0).Max(100).Prefault(50.0)
result, err := numberSchema.Parse(75)   // Returns 75
result, err := numberSchema.Parse(-10)  // Returns 50 (below minimum)
result, err := numberSchema.Parse(150)  // Returns 50 (above maximum)
```

### Error Handling with Invalid Fallback

```go
// If fallback value also fails validation, return error
schema := gozod.String().Min(10).Prefault("short") // Fallback is too short
result, err := schema.Parse("hi") // Error: fallback "short" also fails validation

// Correct usage with valid fallback
correctSchema := gozod.String().Min(5).Prefault("valid fallback")
result, err := correctSchema.Parse("hi") // Returns "valid fallback"
```

This implementation guide provides the complete foundation for implementing the Prefault mechanism in GoZod while maintaining full TypeScript compatibility and Go type safety.

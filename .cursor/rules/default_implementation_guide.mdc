# GoZod Default Mechanism Implementation Guide

## 📋 Overview

This document provides detailed implementation of GoZod Default (default value) mechanism, including architectural design, file organization, implementation steps, and best practices. The mechanism is based on TypeScript Zod v4 design principles but fully adapted to Go's type system and conventions.

## 🎯 Core Design Principles

### 1. Type Safety First
- Provides compile-time type-safe Default methods
- `String().Default(10)` should fail compilation, `String().Default("hello")` should pass
- Avoids type loss caused by using `any` in interfaces

### 2. Smart Type Inference Preservation
- Does not change existing type inference logic
- `String().Default("x").Parse("hello")` → `string`
- `String().Default("x").Parse(&"hello")` → `*string` (same pointer)

### 3. Wrapper Pattern Design
- Uses dedicated wrapper types to provide chaining support
- Each type has its own Default wrapper (e.g., `ZodStringDefault`)
- Maintains consistency with existing `ZodDefault` and `ZodPrefault` wrapper patterns

## 🏗️ Architecture Components

### 1. Core Wrapper (`type_default.go`)

**Responsibility**: Provides generic Default wrapper implementation

```go
// ZodDefault represents a validation schema with default value
// Core design: contains inner type, obtains all its methods through forwarding
type ZodDefault[T ZodType[any, any]] struct {
    innerType    T          // Inner type (cannot embed type parameters, use field)
    defaultValue any        // Default value
    defaultFunc  func() any // Default function
    isFunction   bool       // Whether to use function for default value
}

// Parse performs smart type inference validation and parsing
// Based on TypeScript Zod v4 mechanism:
// - Uses default value when undefined/nil
// - Otherwise delegates to inner type, preserving smart inference
func (z *ZodDefault[T]) Parse(input any, ctx ...*ParseContext) (any, error) {
    // 🔥 TypeScript mechanism: use default value when undefined, otherwise delegate to inner type
    if input == nil {
        if z.isFunction && z.defaultFunc != nil {
            return z.defaultFunc(), nil
        }
        return z.defaultValue, nil
    }

    // Delegate to inner type's Parse (preserve smart type inference)
    return z.innerType.Parse(input, ctx...)
}
```

### 2. Type-Safe Interface (`type.go`)

**Responsibility**: Defines core interface, removes methods that cannot be type-safe

```go
// ZodType simplified interface - removes methods that cannot be type-safe
type ZodType[In, Out any] interface {
    // Core parsing methods
    Parse(input any, ctx ...*ParseContext) (any, error)
    MustParse(input any, ctx ...*ParseContext) any
    
    // Modifier methods
    Nilable() ZodType[any, any]
    
    // Validation and transformation methods
    RefineAny(fn func(any) bool, params ...SchemaParams) ZodType[any, any]
    TransformAny(fn func(any, *RefinementContext) (any, error)) ZodType[any, any]
    Pipe(out ZodType[any, any]) ZodType[any, any]
    
    // Internal state access
    GetInternals() *ZodTypeInternals
    
    // Removed methods (type-safe versions provided on concrete types):
    // - Default(value any) ZodType[In, Out]
    // - DefaultFunc(fn func() any) ZodType[In, Out]  
    // - Prefault(value any) ZodType[In, Out]
    // - PrefaultFunc(fn func() any) ZodType[In, Out]
}
```

### 3. Concrete Type Implementation (various type files)

**Responsibility**: Each type provides type-safe Default methods

```go
// type_string.go - String type Default implementation
type ZodStringDefault struct {
    *ZodDefault[*ZodString] // Embed concrete pointer, enables method promotion
}

// Type-safe Default method - only accepts string type
func (z *ZodString) Default(value string) ZodStringDefault {
    return ZodStringDefault{
        &ZodDefault[*ZodString]{
            innerType:    z,
            defaultValue: value,
            isFunction:   false,
        },
    }
}

// Type-safe DefaultFunc method - only accepts functions returning string
func (z *ZodString) DefaultFunc(fn func() string) ZodStringDefault {
    genericFn := func() any { return fn() }
    return ZodStringDefault{
        &ZodDefault[*ZodString]{
            innerType:   z,
            defaultFunc: genericFn,
            isFunction:  true,
        },
    }
}
```

### 4. Package-Level Functions (`type_default.go`)

**Responsibility**: Provides generic Default creation functions

```go
// Default creates a schema wrapper with default value (improved version - auto-deduction)
func Default[T interface{ GetInternals() *ZodTypeInternals }](innerType T, defaultValue any) ZodType[any, any] {
    // Use type constraints directly, avoid complex type conversions
    anyInnerType := any(innerType).(ZodType[any, any])
    return &ZodDefault[ZodType[any, any]]{
        innerType:    anyInnerType,
        defaultValue: defaultValue,
        isFunction:   false,
    }
}

// DefaultFunc creates a schema wrapper with function default value (improved version - auto-deduction)
func DefaultFunc[T interface{ GetInternals() *ZodTypeInternals }](innerType T, fn func() any) ZodType[any, any] {
    // Use type constraints directly, avoid complex type conversions
    anyInnerType := any(innerType).(ZodType[any, any])
    return &ZodDefault[ZodType[any, any]]{
        innerType:   anyInnerType,
        defaultFunc: fn,
        isFunction:  true,
    }
}
```

## 🔧 Implementation Steps

### Step 1: Define Wrapper Types

Define dedicated Default wrappers in type files:

```go
// type_string.go
// ZodStringDefault is the Default wrapper for string type
// Provides perfect type safety and chaining support
type ZodStringDefault struct {
    *ZodDefault[*ZodString] // Embed concrete pointer, enables method promotion
}
```

### Step 2: Implement Type-Safe Default Methods

```go
// Default type-safe version - only accepts string type
// Compile-time type safety: String().Default(10) will fail compilation
func (z *ZodString) Default(value string) ZodStringDefault {
    return ZodStringDefault{
        &ZodDefault[*ZodString]{
            innerType:    z,
            defaultValue: value,
            isFunction:   false,
        },
    }
}

// DefaultFunc type-safe version - only accepts functions returning string
func (z *ZodString) DefaultFunc(fn func() string) ZodStringDefault {
    genericFn := func() any { return fn() }
    return ZodStringDefault{
        &ZodDefault[*ZodString]{
            innerType:   z,
            defaultFunc: genericFn,
            isFunction:  true,
        },
    }
}
```

### Step 3: Implement Chaining Methods

Implement all necessary chaining methods for the wrapper:

```go
// Min adds minimum length validation, returns ZodStringDefault for chaining
func (s ZodStringDefault) Min(min int, params ...SchemaParams) ZodStringDefault {
    newInner := s.innerType.Min(min, params...)
    return ZodStringDefault{
        &ZodDefault[*ZodString]{
            innerType:    newInner,
            defaultValue: s.defaultValue,
            defaultFunc:  s.defaultFunc,
            isFunction:   s.isFunction,
        },
    }
}

// Max adds maximum length validation, returns ZodStringDefault for chaining
func (s ZodStringDefault) Max(max int, params ...SchemaParams) ZodStringDefault {
    newInner := s.innerType.Max(max, params...)
    return ZodStringDefault{
        &ZodDefault[*ZodString]{
            innerType:    newInner,
            defaultValue: s.defaultValue,
            defaultFunc:  s.defaultFunc,
            isFunction:   s.isFunction,
        },
    }
}

// Email adds email validation, returns ZodStringDefault for chaining
func (s ZodStringDefault) Email(params ...SchemaParams) ZodStringDefault {
    newInner := s.innerType.Email(params...)
    return ZodStringDefault{
        &ZodDefault[*ZodString]{
            innerType:    newInner,
            defaultValue: s.defaultValue,
            defaultFunc:  s.defaultFunc,
            isFunction:   s.isFunction,
        },
    }
}
```

### Step 4: Implement Interface Methods

Ensure wrapper types implement required interfaces:

```go
// GetInternals implements ZodType interface
func (s ZodStringDefault) GetInternals() *ZodTypeInternals {
    return s.ZodDefault.GetInternals()
}

// Parse implements ZodType interface
func (s ZodStringDefault) Parse(input any, ctx ...*ParseContext) (any, error) {
    return s.ZodDefault.Parse(input, ctx...)
}

// MustParse implements ZodType interface
func (s ZodStringDefault) MustParse(input any, ctx ...*ParseContext) any {
    result, err := s.Parse(input, ctx...)
    if err != nil {
        panic(err)
    }
    return result
}
```

## 🧪 Testing Strategy

### 1. Basic Default Tests

```go
func TestZodStringDefault_BasicFunctionality(t *testing.T) {
    schema := gozod.String().Default("default")
    
    t.Run("uses default for nil input", func(t *testing.T) {
        result, err := schema.Parse(nil)
        assert.NoError(t, err)
        assert.Equal(t, "default", result)
    })
    
    t.Run("preserves provided value", func(t *testing.T) {
        result, err := schema.Parse("hello")
        assert.NoError(t, err)
        assert.Equal(t, "hello", result)
    })
}

func TestZodStringDefault_FunctionDefault(t *testing.T) {
    counter := 0
    schema := gozod.String().DefaultFunc(func() string {
        counter++
        return fmt.Sprintf("default-%d", counter)
    })
    
    t.Run("calls function for each default use", func(t *testing.T) {
        result1, err1 := schema.Parse(nil)
        assert.NoError(t, err1)
        assert.Equal(t, "default-1", result1)
        
        result2, err2 := schema.Parse(nil)
        assert.NoError(t, err2)
        assert.Equal(t, "default-2", result2)
    })
}
```

### 2. Type Safety Tests

```go
func TestZodStringDefault_TypeSafety(t *testing.T) {
    // These should compile successfully
    schema1 := gozod.String().Default("hello")
    schema2 := gozod.String().DefaultFunc(func() string { return "world" })
    
    // These should fail compilation:
    // schema3 := gozod.String().Default(123)           // Wrong type
    // schema4 := gozod.String().DefaultFunc(func() int { return 123 }) // Wrong return type
    
    t.Run("valid defaults work", func(t *testing.T) {
        result1, err1 := schema1.Parse(nil)
        assert.NoError(t, err1)
        assert.Equal(t, "hello", result1)
        
        result2, err2 := schema2.Parse(nil)
        assert.NoError(t, err2)
        assert.Equal(t, "world", result2)
    })
}
```

### 3. Chaining Tests

```go
func TestZodStringDefault_Chaining(t *testing.T) {
    schema := gozod.String().Default("default").Min(5).Max(10).Email()
    
    t.Run("validation applies to default value", func(t *testing.T) {
        // Default "default" should fail email validation
        _, err := schema.Parse(nil)
        assert.Error(t, err)
    })
    
    schema2 := gozod.String().Default("user@example.com").Min(5).Max(20).Email()
    
    t.Run("valid default passes all validations", func(t *testing.T) {
        result, err := schema2.Parse(nil)
        assert.NoError(t, err)
        assert.Equal(t, "user@example.com", result)
    })
}
```

### 4. TypeScript Compatibility Tests

```go
func TestTypeScriptCompatibility_Default(t *testing.T) {
    // Equivalent to: z.string().default("hello")
    schema := gozod.String().Default("hello")
    
    testCases := []struct {
        name     string
        input    any
        expected any
        hasError bool
    }{
        {"undefined behavior", nil, "hello", false},
        {"provided value", "world", "world", false},
        {"invalid type", 123, nil, true},
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

### Basic Default Usage

```go
// Static default value
schema := gozod.String().Default("hello")
result, err := schema.Parse(nil) // Returns "hello"
result, err := schema.Parse("world") // Returns "world"

// Function default value
schema := gozod.String().DefaultFunc(func() string {
    return time.Now().Format("2006-01-02")
})
result, err := schema.Parse(nil) // Returns current date
```

### Default with Validation

```go
// Default value must also pass validation
schema := gozod.String().Default("user@example.com").Email()
result, err := schema.Parse(nil) // Returns "user@example.com"

// Invalid default will cause validation to fail
invalidSchema := gozod.String().Default("invalid").Email()
result, err := invalidSchema.Parse(nil) // Error: invalid email
```

### Complex Default Objects

```go
// Object with default values
userSchema := gozod.Object(map[string]gozod.ZodType[any, any]{
    "name":  gozod.String().Default("Anonymous"),
    "age":   gozod.Number().Default(0),
    "email": gozod.String().Email().Optional(),
})

result, err := userSchema.Parse(map[string]any{
    "email": "user@example.com",
    // name and age will use defaults
})
// Result: {"name": "Anonymous", "age": 0, "email": "user@example.com"}
```

### Dynamic Defaults

```go
// Generate unique IDs
idSchema := gozod.String().DefaultFunc(func() string {
    return uuid.New().String()
})

// Timestamp defaults
timestampSchema := gozod.String().DefaultFunc(func() string {
    return time.Now().UTC().Format(time.RFC3339)
})

result, err := idSchema.Parse(nil) // Returns new UUID
result, err := timestampSchema.Parse(nil) // Returns current timestamp
```

This implementation guide provides the complete foundation for implementing the Default mechanism in GoZod while maintaining full TypeScript compatibility and Go type safety. 

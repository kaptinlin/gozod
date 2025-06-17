# GoZod Schema Type Implementation Guide

## 📋 Overview

This document provides a comprehensive guide for implementing GoZod schema types with TypeScript Zod v4 API compatibility, covering unified design patterns, smart type inference, generic implementation, and complete functionality correspondence.

## 🎯 Core Design Principles

### 1. Smart Type Inference Preservation
- **Basic Type Inference**: `String().Parse("hello")` → `string`
- **Pointer Type Inference**: `String().Parse(&"hello")` → `*string` (same pointer)
- **Nilable Type Inference**: `String().Nilable().Parse(nil)` → `(*string)(nil)`
- **Optional Semantics**: `String().Optional().Parse(nil)` → `nil` (represents missing field)

### 2. Unified Wrapper Pattern
- **Default Wrapper**: `ZodStringDefault` provides type-safe default values
- **Prefault Wrapper**: `ZodStringPrefault` provides type-safe fallback values
- **Optional Wrapper**: `ZodOptional[T]` generic wrapper handles missing field semantics
- **Nilable Wrapper**: `ZodNilable[T]` generic wrapper handles nullable value semantics

### 3. Separation of Concerns Principle
- **Refine**: Only validates data, never modifies data
- **Transform**: Can modify data type and content
- **Coerce**: Performs type coercion before parsing
- **Check**: Low-level validation mechanism, supports complex validation logic

## 🏗️ Schema Type Architecture

### 1. Unified File Organization Pattern

Every GoZod schema type follows the consistent structure from RESTRUCTURE.md:

```go
//////////////////////////
// CORE TYPE DEFINITIONS
//////////////////////////

// Type definitions (ZodXXXDef, ZodXXXInternals, ZodXXX)
type ZodStringDef struct {
    ZodTypeDef
    Type   string     // "string"
    Checks []ZodCheck // String-specific validation checks
}

type ZodStringInternals struct {
    ZodTypeInternals
    Def     *ZodStringDef          // Schema definition
    Checks  []ZodCheck             // Validation checks
    Isst    ZodIssueInvalidType    // Invalid type issue template
    Pattern *regexp.Regexp         // Regex pattern (if any)
    Values  map[string]struct{}    // Allowed string values set
    Bag     map[string]interface{} // Additional metadata (formats, etc.)
}

type ZodString struct {
    internals *ZodStringInternals
}

// Core interface implementations (GetInternals, Coerce, Parse, MustParse)
func (z *ZodString) GetInternals() *ZodTypeInternals { ... }
func (z *ZodString) Coerce(input interface{}) (interface{}, bool) { ... }
func (z *ZodString) Parse(input any, ctx ...*ParseContext) (any, error) { ... }
func (z *ZodString) MustParse(input any, ctx ...*ParseContext) any { ... }

//////////////////////////
// VALIDATION METHODS (or TYPE-SPECIFIC METHODS)
//////////////////////////

// Type-specific validation methods
// Format validators, constraints, etc.

//////////////////////////
// TRANSFORM METHODS
//////////////////////////

// Transform, TransformAny
// Type-specific transformations
// Pipe operations

//////////////////////////
// MODIFIER METHODS
//////////////////////////

// Optional, Nilable, Nullish
// Refine, RefineAny
// Unwrap

//////////////////////////
// WRAPPER TYPES
//////////////////////////

// Default wrappers with all chainable methods
// Prefault wrappers with all chainable methods

//////////////////////////
// CONSTRUCTOR FUNCTIONS
//////////////////////////

// Internal constructors (createZodXXXFromDef)
// Public constructors (NewZodXXX)
// Package-level constructors

//////////////////////////
// UTILITY FUNCTIONS
//////////////////////////

// Helper functions
// Internal utilities
// Type-specific helpers
```

### 2. Required Interface Implementation

Every schema type must implement the following interfaces:

```go
// Core ZodType interface - based on actual definition in type.go
type ZodType[In, Out any] interface {
    // Parse performs smart type inference validation and parsing
    Parse(input any, ctx ...*ParseContext) (any, error)
    
    // MustParse performs validation, panics on failure
    MustParse(input any, ctx ...*ParseContext) any
    
    // GetInternals gets internal state (framework use)
    GetInternals() *ZodTypeInternals
    
    // Nilable modifier: creates generic wrapper to handle nullable value semantics
    Nilable() ZodType[any, any]
    
    // Optional modifier: creates generic wrapper to handle missing field semantics
    Optional() ZodType[any, any]
    
    // Unwrap returns inner type (for wrapper types)
    Unwrap() ZodType[any, any]
    
    // RefineAny corresponds to z.string().refine() - pure validator, never changes data
    RefineAny(fn func(any) bool, params ...SchemaParams) ZodType[any, any]
    
    // TransformAny corresponds to data transformer - can modify data type and content
    TransformAny(fn func(any, *RefinementContext) (any, error)) ZodType[any, any]
    
    // Pipe corresponds to z.string().pipe() - TypeScript Zod v4 pipeline functionality
    Pipe(out ZodType[any, any]) ZodType[any, any]
}

// Optional Coercible interface
type Coercible interface {
    Coerce(input interface{}) (output interface{}, success bool)
}

// Optional Cloneable interface
type Cloneable interface {
    CloneFrom(source any)
}
```

## 🔧 Implementation Patterns by Section

### 1. VALIDATION METHODS Section

Basic validation methods directly return the same type, supporting method chaining:

```go
// Min adds minimum length validation
func (z *ZodString) Min(minLen int, params ...SchemaParams) *ZodString {
    check := NewZodCheckMinLength(minLen, params...)
    result := AddCheck(z, check)
    return result.(*ZodString)
}

// Max adds maximum length validation
func (z *ZodString) Max(maxLen int, params ...SchemaParams) *ZodString {
    check := NewZodCheckMaxLength(maxLen, params...)
    result := AddCheck(z, check)
    return result.(*ZodString)
}

// Email adds email format validation
func (z *ZodString) Email(params ...SchemaParams) *ZodString {
    check := NewZodCheckStringFormat(StringFormatEmail, regexes.Email, params...)
    result := AddCheck(z, check)
    return result.(*ZodString)
}
```

### 2. TRANSFORM METHODS Section

Transform methods return generic ZodType since they may change the type:

```go
// Transform provides type-safe string transformation with smart dereferencing
func (z *ZodString) Transform(fn func(string, *RefinementContext) (any, error)) ZodType[any, any] {
    return z.TransformAny(func(input any, ctx *RefinementContext) (any, error) {
        str, isNil, err := extractStringValue(input)

        if err != nil {
            return nil, err
        }

        if isNil {
            return nil, ErrTransformNilString
        }

        return fn(str, ctx)
    })
}

// TransformAny flexible version of transformation
func (z *ZodString) TransformAny(fn func(any, *RefinementContext) (any, error)) ZodType[any, any] {
    transform := NewZodTransform[any, any](fn)

    return &ZodPipe[any, any]{
        in:  any(z).(ZodType[any, any]),
        out: any(transform).(ZodType[any, any]),
        def: ZodTypeDef{Type: "pipe"},
    }
}

// String Transformations
func (z *ZodString) Trim() ZodType[any, any] {
    return z.createStringTransform(strings.TrimSpace)
}

func (z *ZodString) ToLowerCase() ZodType[any, any] {
    return z.createStringTransform(strings.ToLower)
}

func (z *ZodString) ToUpperCase() ZodType[any, any] {
    return z.createStringTransform(strings.ToUpper)
}

// Pipe operation for pipeline chaining
func (z *ZodString) Pipe(out ZodType[any, any]) ZodType[any, any] {
    return &ZodPipe[any, any]{
        in:  z,
        out: out,
        def: ZodTypeDef{Type: "pipe"},
    }
}
```

### 3. MODIFIER METHODS Section

Modifier methods change the behavior of schemas:

```go
// Optional makes the string optional
func (z *ZodString) Optional() ZodType[any, any] {
    return any(Optional(any(z).(ZodType[any, any]))).(ZodType[any, any])
}

// Nilable makes the string nilable while preserving type inference
func (z *ZodString) Nilable() ZodType[any, any] {
    return Clone(z, func(def *ZodTypeDef) {
        // Nilable is a runtime flag
    }).(*ZodString).setNilable()
}

// Nullish makes the string both optional and nilable
func (z *ZodString) Nullish() ZodType[any, any] {
    return any(Nullish(any(z).(ZodType[any, any]))).(ZodType[any, any])
}

// Refine adds type-safe custom validation logic
func (z *ZodString) Refine(fn func(string) bool, params ...SchemaParams) *ZodString {
    result := z.RefineAny(func(v any) bool {
        str, isNil, err := extractStringValue(v)

        if err != nil {
            return false
        }

        if isNil {
            return true // Let Nilable flag handle nil validation
        }

        return fn(str)
    }, params...)
    return result.(*ZodString)
}

// RefineAny adds flexible custom validation logic
func (z *ZodString) RefineAny(fn func(any) bool, params ...SchemaParams) ZodType[any, any] {
    check := NewCustom[any](fn, params...)
    return AddCheck(z, check)
}

// Unwrap returns the inner type (for basic types, returns self)
func (z *ZodString) Unwrap() ZodType[any, any] {
    return any(z).(ZodType[any, any])
}
```

### 4. WRAPPER TYPES Section

Type-safe wrapper implementations with full method chaining:

```go
// ZodStringDefault is a default value wrapper for string type
type ZodStringDefault struct {
    *ZodDefault[*ZodString]
}

// Type-safe Default methods
func (z *ZodString) Default(value string) ZodStringDefault {
    return ZodStringDefault{
        &ZodDefault[*ZodString]{
            innerType:    z,
            defaultValue: value,
            isFunction:   false,
        },
    }
}

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

// ZodStringDefault chainable validation methods
func (s ZodStringDefault) Min(minLen int, params ...SchemaParams) ZodStringDefault {
    newInner := s.innerType.Min(minLen, params...)
    return ZodStringDefault{
        &ZodDefault[*ZodString]{
            innerType:    newInner,
            defaultValue: s.defaultValue,
            defaultFunc:  s.defaultFunc,
            isFunction:   s.isFunction,
        },
    }
}

// ZodStringPrefault is a prefault value wrapper for string type
type ZodStringPrefault struct {
    *ZodPrefault[*ZodString]
}

// Type-safe Prefault methods
func (z *ZodString) Prefault(value string) ZodStringPrefault {
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
```

### 5. CONSTRUCTOR FUNCTIONS Section

Internal and public constructor implementations:

```go
// createZodStringFromDef creates a ZodString from definition using unified patterns
func createZodStringFromDef(def *ZodStringDef) *ZodString {
    internals := &ZodStringInternals{
        ZodTypeInternals: newBaseZodTypeInternals(def.Type),
        Def:              def,
        Checks:           def.Checks,
        Isst:             ZodIssueInvalidType{Expected: "string"},
        Pattern:          nil,
        Values:           make(map[string]struct{}),
        Bag:              make(map[string]interface{}),
    }

    internals.Constructor = func(newDef *ZodTypeDef) ZodType[any, any] {
        stringDef := &ZodStringDef{
            ZodTypeDef: *newDef,
            Type:       ZodTypeString,
            Checks:     newDef.Checks,
        }
        return any(createZodStringFromDef(stringDef)).(ZodType[any, any])
    }

    schema := &ZodString{internals: internals}
    initZodType(schema, &def.ZodTypeDef)
    return schema
}

// NewZodString creates a new string schema with unified parameter handling
func NewZodString(params ...SchemaParams) *ZodString {
    def := &ZodStringDef{
        ZodTypeDef: ZodTypeDef{
            Type:   ZodTypeString,
            Checks: make([]ZodCheck, 0),
        },
        Type:   ZodTypeString,
        Checks: make([]ZodCheck, 0),
    }

    schema := createZodStringFromDef(def)
    // Parameter handling logic...
    return schema
}

// String creates a new string schema (package-level constructor)
func String(params ...SchemaParams) *ZodString {
    return NewZodString(params...)
}

// CoercedString creates a new string schema with coercion enabled
func CoercedString(params ...SchemaParams) *ZodString {
    var coerceParams SchemaParams
    if len(params) > 0 {
        coerceParams = params[0]
    }
    coerceParams.Coerce = true

    return NewZodString(coerceParams)
}
```

### 6. UTILITY FUNCTIONS Section

Helper functions and internal utilities:

```go
// setNilable sets the Nilable flag internally
func (z *ZodString) setNilable() ZodType[any, any] {
    z.internals.Nilable = true
    return z
}

// GetZod returns the string-specific internals
func (z *ZodString) GetZod() *ZodStringInternals {
    return z.internals
}

// CloneFrom implements Cloneable interface
func (z *ZodString) CloneFrom(source any) {
    if src, ok := source.(interface{ GetZod() *ZodStringInternals }); ok {
        srcState := src.GetZod()
        tgtState := z.GetZod()
        // Clone logic...
    }
}

// createStringTransform creates string transformation helper
func (z *ZodString) createStringTransform(transformFn func(string) string) ZodType[any, any] {
    transform := NewZodTransform[any, any](func(input any, _ctx *RefinementContext) (any, error) {
        str, isNil, err := extractStringValue(input)
        if err != nil {
            return nil, err
        }
        if isNil {
            return nil, ErrTransformNilString
        }
        return transformFn(str), nil
    })
    return &ZodPipe[any, any]{
        in:  any(z).(ZodType[any, any]),
        out: any(transform).(ZodType[any, any]),
        def: ZodTypeDef{Type: "pipe"},
    }
}

// extractStringValue extracts string value from input with smart handling
func extractStringValue(input any) (string, bool, error) {
    switch v := input.(type) {
    case string:
        return v, false, nil
    case *string:
        if v == nil {
            return "", true, nil
        }
        return *v, false, nil
    default:
        return "", false, fmt.Errorf("%w, got %T", ErrExpectedString, input)
    }
}
```

## 📊 Parse Method Implementation

### 1. Standard Parse Structure

```go
func (z *ZodString) Parse(input any, ctx ...*ParseContext) (any, error) {
    parseCtx := (*ParseContext)(nil)
    if len(ctx) > 0 {
        parseCtx = ctx[0]
    }

    return parseType[string](
        input,
        &z.internals.ZodTypeInternals,
        "string",
        func(v any) (string, bool) { str, ok := v.(string); return str, ok },
        func(v any) (*string, bool) { ptr, ok := v.(*string); return ptr, ok },
        validateString,
        coerceToString,
        parseCtx,
    )
}
```

### 2. Coerce Method Implementation

```go
func (z *ZodString) Coerce(input interface{}) (interface{}, bool) {
    return coerceToString(input)
}
```

## 🧪 Testing and Validation

### 1. Comprehensive Test Structure

```go
func TestZodString_Parse(t *testing.T) {
    schema := gozod.String()
    
    t.Run("valid string", func(t *testing.T) {
        result, err := schema.Parse("hello")
        assert.NoError(t, err)
        assert.Equal(t, "hello", result)
    })
    
    t.Run("valid pointer string", func(t *testing.T) {
        input := "hello"
        result, err := schema.Parse(&input)
        assert.NoError(t, err)
        assert.Equal(t, &input, result)
    })
    
    t.Run("invalid type", func(t *testing.T) {
        _, err := schema.Parse(123)
        assert.Error(t, err)
        var zodErr *gozod.ZodError
        assert.True(t, errors.As(err, &zodErr))
    })
}

func TestZodString_Nilable(t *testing.T) {
    schema := gozod.String().Nilable()
    
    t.Run("nil input", func(t *testing.T) {
        result, err := schema.Parse(nil)
        assert.NoError(t, err)
        assert.Nil(t, result)
    })
    
    t.Run("nil pointer input", func(t *testing.T) {
        var input *string = nil
        result, err := schema.Parse(input)
        assert.NoError(t, err)
        assert.Equal(t, input, result)
    })
}
```

### 2. TypeScript Compatibility Tests

```go
func TestTypeScriptCompatibility(t *testing.T) {
    // Equivalent to: z.string().min(5).max(10)
    schema := gozod.String().Min(5).Max(10)
    
    testCases := []struct {
        name     string
        input    any
        expected any
        hasError bool
    }{
        {"valid string", "hello", "hello", false},
        {"too short", "hi", nil, true},
        {"too long", "very long string", nil, true},
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

## 📖 Documentation Requirements

### 1. Method Documentation Template

```go
// Min adds a minimum length validation check to the string schema.
// It corresponds to TypeScript Zod's z.string().min(n) method.
//
// Parameters:
//   - min: The minimum required length (inclusive)
//   - params: Optional validation parameters (error message, etc.)
//
// Returns:
//   - *ZodString: A new string schema with the minimum length check added
//
// Example:
//   schema := gozod.String().Min(5)
//   result, err := schema.Parse("hello") // OK
//   result, err := schema.Parse("hi")    // Error: too short
func (z *ZodString) Min(min int, params ...SchemaParams) *ZodString {
    // Implementation...
}
```

### 2. Type Documentation Template

```go
// ZodString represents a string validation schema, corresponding to 
// TypeScript Zod's z.string() type. It provides validation for string
// values with support for various string-specific checks like length,
// format validation, and pattern matching.
//
// The ZodString type supports smart type inference:
//   - String input: returns string
//   - *string input: returns *string 
//   - nil *string with Nilable(): returns nil
//
// Basic usage:
//   schema := gozod.String()
//   result, err := schema.Parse("hello")
//
// With validation:
//   schema := gozod.String().Min(5).Max(10).Email()
//   result, err := schema.Parse("user@example.com")
type ZodString struct {
    internals *ZodStringInternals
}
```

## 🎨 Pattern Template for Future Files

When implementing remaining schema types, follow this unified template:

```go
package gozod

import (
    // imports
)

//////////////////////////
// CORE TYPE DEFINITIONS
//////////////////////////

// ZodXXXDef defines the configuration
type ZodXXXDef struct {
    ZodTypeDef
    // type-specific fields
}

// ZodXXXInternals contains internal state
type ZodXXXInternals struct {
    ZodTypeInternals
    // type-specific internals
}

// ZodXXX represents the validation schema
type ZodXXX struct {
    internals *ZodXXXInternals
}

// Core interface implementations
func (z *ZodXXX) GetInternals() *ZodTypeInternals { ... }
func (z *ZodXXX) Coerce(input interface{}) (interface{}, bool) { ... }
func (z *ZodXXX) Parse(input any, ctx ...*ParseContext) (any, error) { ... }
func (z *ZodXXX) MustParse(input any, ctx ...*ParseContext) any { ... }

//////////////////////////
// [TYPE-SPECIFIC] METHODS
//////////////////////////

// Type-specific validation methods

//////////////////////////
// TRANSFORM METHODS
//////////////////////////

// Transform, TransformAny, Pipe

//////////////////////////
// MODIFIER METHODS
//////////////////////////

// Optional, Nilable, Nullish, Refine, RefineAny, Unwrap

//////////////////////////
// WRAPPER TYPES
//////////////////////////

// Default and Prefault wrappers with chainable methods

//////////////////////////
// CONSTRUCTOR FUNCTIONS
//////////////////////////

// Internal and public constructors

//////////////////////////
// UTILITY FUNCTIONS
//////////////////////////

// Helper functions and utilities
```

This guide provides the complete foundation for implementing GoZod schema types with full TypeScript Zod v4 compatibility while maintaining Go's type safety and performance characteristics. 

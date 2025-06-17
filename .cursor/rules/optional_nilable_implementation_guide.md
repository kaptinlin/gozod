# GoZod Optional and Nilable Mechanism Implementation Guide

## 📋 Overview

This document provides detailed implementation of GoZod Optional and Nilable mechanisms, including architectural design, file organization, implementation steps, and best practices. The mechanism is based on TypeScript Zod v4 design principles but fully adapted to Go's type system and conventions.

## 🎯 Core Design Principles

### 1. Semantic Distinction Principle
- **Nilable**: Corresponds to TypeScript's `nullable()`, handles explicit "null value" concept, returns typed nil pointer
- **Optional**: Corresponds to TypeScript's `optional()`, handles "missing field" concept, returns generic nil
- **Nullish**: Combines both, can be both missing and null

### 2. Smart Type Inference Preservation
- Nilable preserves smart type inference: `String().Nilable().Parse(nil)` → `(*string)(nil)`
- Optional represents missing: `String().Optional().Parse(nil)` → `nil`
- For non-nil inputs, completely delegates to inner type, preserving smart inference

### 3. Generic Wrapper Pattern Design
- Uses generic wrapper types to provide type-safe chaining support
- Each wrapper has its own generic parameter constraints
- Maintains consistency with existing `ZodDefault` and `ZodPrefault` wrapper patterns
- Uses `innerType` field rather than embedding to avoid generic parameter embedding issues

## 🏗️ Architecture Components

### 1. Nilable Core Implementation (`type_nilable.go`)

**Responsibility**: Provides nullable value validation and handling

```go
// ZodNilableDef defines nullable validation configuration - using generics
type ZodNilableDef[T ZodType[any, any]] struct {
    ZodTypeDef
    Type      string // "nilable"
    InnerType T      // The wrapped type - using generic parameter
}

// ZodNilableInternals contains nullable validator internal state - using generics
type ZodNilableInternals[T ZodType[any, any]] struct {
    ZodTypeInternals
    Def     *ZodNilableDef[T]        // Nilable definition with generic
    Values  map[interface{}]struct{} // Inherited from inner type values
    Pattern *regexp.Regexp           // Inherited from inner type pattern
}

// ZodNilable represents nullable validation schema - using generic design
// Core design: contains inner type, obtains all its methods through forwarding
type ZodNilable[T ZodType[any, any]] struct {
    innerType T                 // Inner type (cannot embed type parameters, use field)
    internals *ZodTypeInternals // Nilable's own internals
}

// GetInternals returns internal state, implements smart nil handling
func (z *ZodNilable[T]) GetInternals() *ZodTypeInternals {
    // Nilable needs its own internals to properly handle nil values
    if z.internals == nil {
        z.internals = &ZodTypeInternals{
            Type: "nilable",
            Parse: func(payload *ParsePayload, ctx *ParseContext) *ParsePayload {
                // Implement Nilable's nil handling logic
                if payload.Value == nil {
                    // Determine inner type and return corresponding typed nil pointer
                    innerTypeInternals := z.innerType.GetInternals()
                    switch innerTypeInternals.Type {
                    case "string":
                        payload.Value = (*string)(nil)
                    case "boolean":
                        payload.Value = (*bool)(nil)
                    case "int", "int8", "int16", "int32", "int64":
                        payload.Value = (*int)(nil)
                    case "uint", "uint8", "uint16", "uint32", "uint64":
                        payload.Value = (*uint)(nil)
                    case "float32", "float64", "number":
                        payload.Value = (*float64)(nil)
                    case "any":
                        payload.Value = (*interface{})(nil)
                    default:
                        // For other types, return generic nil pointer
                        payload.Value = (*interface{})(nil)
                    }
                    return payload
                }

                // Delegate to inner type's Parse
                return z.innerType.GetInternals().Parse(payload, ctx)
            },
        }
    }
    return z.internals
}

// Parse performs smart type inference validation and parsing
// 🔥 Core: only handles nil, completely delegates everything else to inner type
func (z *ZodNilable[T]) Parse(input any, ctx ...*ParseContext) (any, error) {
    // Core: only handles nil, completely delegates everything else to inner type
    if input == nil {
        // Determine inner type and return corresponding typed nil pointer
        innerTypeInternals := z.innerType.GetInternals()
        switch innerTypeInternals.Type {
        case "string":
            return (*string)(nil), nil
        case "boolean":
            return (*bool)(nil), nil
        case "int", "int8", "int16", "int32", "int64":
            return (*int)(nil), nil
        case "uint", "uint8", "uint16", "uint32", "uint64":
            return (*uint)(nil), nil
        case "float32", "float64", "number":
            return (*float64)(nil), nil
        case "any":
            return (*interface{})(nil), nil
        default:
            // For other types, return generic nil pointer
            return (*interface{})(nil), nil
        }
    }

    // Completely delegate to inner type, preserve its smart inference
    return z.innerType.Parse(input, ctx...)
}
```

### 2. Optional Core Implementation (`type_optional.go`)

**Responsibility**: Provides optional field validation and handling

```go
// ZodOptionalDef defines optional validation configuration - using generics
type ZodOptionalDef[T ZodType[any, any]] struct {
    ZodTypeDef
    Type      string // "optional"
    InnerType T      // The wrapped type - using generic parameter
}

// ZodOptionalInternals contains optional validator internal state - using generics
type ZodOptionalInternals[T ZodType[any, any]] struct {
    ZodTypeInternals
    Def     *ZodOptionalDef[T]       // Optional definition with generic
    OptIn   string                   // "optional"
    Values  map[interface{}]struct{} // Inherited from inner type values
    Pattern *regexp.Regexp           // Inherited from inner type pattern
}

// ZodOptional represents optional validation schema - using generic design
// Core design: contains inner type, obtains all its methods through forwarding
type ZodOptional[T ZodType[any, any]] struct {
    innerType T                 // Inner type (cannot embed type parameters, use field)
    internals *ZodTypeInternals // Optional's own internals
}

// GetInternals returns internal state, implements smart nil handling
func (z *ZodOptional[T]) GetInternals() *ZodTypeInternals {
    // Optional needs its own internals to properly handle nil values
    if z.internals == nil {
        z.internals = &ZodTypeInternals{
            Type: "optional",
            Parse: func(payload *ParsePayload, ctx *ParseContext) *ParsePayload {
                // Implement Optional's nil handling logic
                if nullish(payload.Value) {
                    // Optional allows missing values, return generic nil
                    payload.Value = nil
                    return payload
                }

                // Delegate to inner type's Parse
                return z.innerType.GetInternals().Parse(payload, ctx)
            },
        }
    }
    return z.internals
}

// Parse performs smart type inference validation and parsing
// Based on TypeScript Zod v4 mechanism: returns nil when undefined/nil, otherwise delegates to inner type
func (z *ZodOptional[T]) Parse(input any, ctx ...*ParseContext) (any, error) {
    // 🔥 TypeScript mechanism: returns nil when undefined, otherwise delegates to inner type
    if nullish(input) {
        // Optional allows missing values, return generic nil to represent "field can be missing"
        // This differs from Nilable, which returns typed nil to represent "value can be explicitly null"
        return nil, nil
    }

    // Delegate to inner type's Parse (preserve smart type inference)
    return z.innerType.Parse(input, ctx...)
}
```

### 3. Generic Constructor Implementation

**Responsibility**: Provides type-safe constructor functions

```go
// Nilable creates nullable pattern wrapper (generic version - auto-deduction)
func Nilable[T interface{ GetInternals() *ZodTypeInternals }](innerType T, params ...SchemaParams) ZodType[any, any] {
    // Use type constraints directly, avoid complex type conversions
    anyInnerType := any(innerType).(ZodType[any, any])
    return &ZodNilable[ZodType[any, any]]{
        innerType: anyInnerType,
    }
}

// Optional creates optional pattern wrapper (generic version - auto-deduction)
func Optional[T interface{ GetInternals() *ZodTypeInternals }](innerType T, params ...SchemaParams) ZodType[any, any] {
    // Use type constraints directly, avoid complex type conversions
    anyInnerType := any(innerType).(ZodType[any, any])
    return &ZodOptional[ZodType[any, any]]{
        innerType: anyInnerType,
    }
}

// Nullish creates nullish pattern wrapper (combines optional and nullable)
func Nullish[T interface{ GetInternals() *ZodTypeInternals }](innerType T, params ...SchemaParams) ZodType[any, any] {
    // Nullish = Optional(Nilable(innerType))
    nilableWrapper := Nilable(innerType, params...)
    return Optional(nilableWrapper, params...)
}
```

### 4. Type-Safe Method Forwarding

**Responsibility**: Provides method forwarding for wrapper types

```go
// Nilable method forwarding implementation (added to all schema types)
func (z *ZodString) Nilable() ZodType[any, any] {
    return Nilable(any(z).(ZodType[any, any]))
}

// Optional method forwarding implementation (added to all schema types)
func (z *ZodString) Optional() ZodType[any, any] {
    return Optional(any(z).(ZodType[any, any]))
}

// Nullish method forwarding implementation (added to all schema types)
func (z *ZodString) Nullish() ZodType[any, any] {
    return Nullish(any(z).(ZodType[any, any]))
}
```

## 🔧 Implementation Steps

### Step 1: Implement Core Wrapper Types

Create the basic wrapper structures and interfaces for both Optional and Nilable types.

### Step 2: Implement Generic Constructor Functions

Provide type-safe constructor functions that can work with any ZodType.

### Step 3: Add Method Forwarding

Add Nilable(), Optional(), and Nullish() methods to all existing schema types.

### Step 4: Implement Smart Type Inference

Ensure that nil handling preserves type information and follows Go conventions.

### Step 5: Add Comprehensive Testing

Create thorough test suites that verify TypeScript compatibility and Go-specific behavior.

## 🧪 Testing Strategy

### 1. Basic Functionality Tests

```go
func TestNilable_BasicFunctionality(t *testing.T) {
    schema := gozod.String().Nilable()
    
    t.Run("nil input returns typed nil pointer", func(t *testing.T) {
        result, err := schema.Parse(nil)
        assert.NoError(t, err)
        assert.Equal(t, (*string)(nil), result)
    })
    
    t.Run("valid string preserves type", func(t *testing.T) {
        result, err := schema.Parse("hello")
        assert.NoError(t, err)
        assert.Equal(t, "hello", result)
    })
}

func TestOptional_BasicFunctionality(t *testing.T) {
    schema := gozod.String().Optional()
    
    t.Run("nil input returns nil", func(t *testing.T) {
        result, err := schema.Parse(nil)
        assert.NoError(t, err)
        assert.Nil(t, result)
    })
    
    t.Run("valid string preserves type", func(t *testing.T) {
        result, err := schema.Parse("hello")
        assert.NoError(t, err)
        assert.Equal(t, "hello", result)
    })
}
```

### 2. TypeScript Compatibility Tests

```go
func TestTypeScriptCompatibility_Optional(t *testing.T) {
    // Equivalent to: z.string().optional()
    schema := gozod.String().Optional()
    
    testCases := []struct {
        name     string
        input    any
        expected any
        hasError bool
    }{
        {"undefined behavior", nil, nil, false},
        {"valid string", "hello", "hello", false},
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

### 3. Chaining and Composition Tests

```go
func TestOptionalNilable_Chaining(t *testing.T) {
    t.Run("Optional(Nilable(String()))", func(t *testing.T) {
        schema := gozod.String().Nilable().Optional()
        
        // Both nil and typed nil should work
        result1, err1 := schema.Parse(nil)
        assert.NoError(t, err1)
        assert.Nil(t, result1)
        
        var typedNil *string = nil
        result2, err2 := schema.Parse(typedNil)
        assert.NoError(t, err2)
        assert.Equal(t, typedNil, result2)
        
        result3, err3 := schema.Parse("hello")
        assert.NoError(t, err3)
        assert.Equal(t, "hello", result3)
    })
}
```

## 📖 Usage Examples

### Basic Usage

```go
// Create a nilable string schema
nilableString := gozod.String().Nilable()
result, err := nilableString.Parse(nil) // Returns (*string)(nil)

// Create an optional string schema  
optionalString := gozod.String().Optional()
result, err := optionalString.Parse(nil) // Returns nil

// Create a nullish string schema (both optional and nilable)
nullishString := gozod.String().Nullish()
result, err := nullishString.Parse(nil) // Returns nil
```

### Complex Validation Chains

```go
// Complex validation with optional fields
schema := gozod.String().Min(5).Max(10).Email().Optional()
result, err := schema.Parse(nil) // OK, returns nil
result, err := schema.Parse("user@example.com") // OK, returns "user@example.com"
result, err := schema.Parse("invalid") // Error: not a valid email
```

This implementation guide provides the complete foundation for implementing Optional and Nilable mechanisms in GoZod while maintaining full TypeScript compatibility and Go type safety. 

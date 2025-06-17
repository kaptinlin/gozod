# GoZod Checks Mechanism Implementation Guide

## 📋 Overview

This document provides detailed implementation of GoZod Checks (validation checks) mechanism, including architectural design, file organization, implementation steps, and best practices. The mechanism is based on TypeScript Zod v4 design principles but fully adapted to Go's type system and conventions.

## 🎯 Core Design Principles

### 1. Unified Check Interface
- All checks implement the `ZodCheck` interface
- Provides unified `GetZod()` method to get internal state
- Supports conditional execution and error accumulation

### 2. Type Safety First
- Provides compile-time type-safe check functions
- Supports generic constraints and type inference
- Avoids runtime type errors

### 3. Execution Order Mechanism
- Follows TypeScript Zod v4 execution order
- Supports conditional execution (`When` function)
- Supports early abort (`Abort` flag)

## 🏗️ Architecture Components

### 1. Core Interface Definition (`checks.go`)

**Responsibility**: Defines the basic interface for the check mechanism

```go
// ZodCheck represents validation constraint interface
type ZodCheck interface {
    // GetZod returns validator internal state, corresponding to TypeScript's _zod property
    GetZod() *ZodCheckInternals
}

// ZodCheckDef defines the configuration for validation checks
type ZodCheckDef struct {
    Check string       // Check type identifier
    Error *ZodErrorMap // Custom error mapping
    Abort bool         // Whether to abort on validation failure
}

// ZodCheckInternals contains validator internal state and configuration
type ZodCheckInternals struct {
    Def      *ZodCheckDef                     // Check definition
    Issc     *ZodIssueBase                    // The set of issues this check might throw
    Check    ZodCheckFn                       // Validation function
    OnAttach []func(schema interface{})       // Array of attachment callback functions
    When     func(payload *ParsePayload) bool // Conditional function
}

// ZodCheckFn defines the function that executes validation
type ZodCheckFn func(payload *ParsePayload)
```

### 2. Check Execution Mechanism (`type.go`)

**Responsibility**: Provides core logic for check execution

```go
// runChecks executes all checks on a payload synchronously
func runChecks(payload *ParsePayload, checks []ZodCheck, ctx *ParseContext) *ParsePayload {
    isAborted := aborted(*payload, 0)

    for _, check := range checks {
        if check != nil {
            if checkInternals := check.GetZod(); checkInternals != nil {
                // Check conditional execution
                if checkInternals.When != nil && !checkInternals.When(payload) {
                    continue
                }

                // Skip if already aborted
                if isAborted {
                    continue
                }

                currLen := len(payload.Issues)

                // Execute check
                checkInternals.Check(payload)

                // Check if new issues were added and should abort
                if len(payload.Issues) > currLen {
                    if checkInternals.Def.Abort {
                        isAborted = true
                    }
                }
            }
        }
    }

    return payload
}

// aborted checks if parsing should be aborted
func aborted(x ParsePayload, startIndex int) bool {
    for i := startIndex; i < len(x.Issues); i++ {
        if x.Issues[i].Continue != true {
            return true
        }
    }
    return false
}
```

### 3. Check Addition Mechanism (`type.go`)

**Responsibility**: Provides check addition and state management

```go
// AddCheck adds a validation check to any ZodType and returns new instance
func AddCheck[T interface{ GetInternals() *ZodTypeInternals }](schema T, check ZodCheck) ZodType[any, any] {
    internals := schema.GetInternals()

    // Create new type definition with updated checks
    newDef := &ZodTypeDef{
        Type:   internals.Type,
        Error:  internals.Error,
        Checks: append(make([]ZodCheck, len(internals.Checks)), internals.Checks...),
    }
    // Append the new check
    newDef.Checks = append(newDef.Checks, check)

    // Use existing constructor
    if internals.Constructor != nil {
        newSchema := internals.Constructor(newDef)

        // Keep all important state flags
        newInternals := newSchema.GetInternals()
        newInternals.Nilable = internals.Nilable
        newInternals.Optional = internals.Optional
        newInternals.Coerce = internals.Coerce

        // Keep Pattern state
        if internals.Pattern != nil {
            newInternals.Pattern = internals.Pattern
        }

        // Keep Values state
        if len(internals.Values) > 0 {
            newInternals.Values = make(map[interface{}]struct{})
            for k, v := range internals.Values {
                newInternals.Values[k] = v
            }
        }

        // Use Cloneable interface to copy type-specific state
        if cloneable, ok := newSchema.(Cloneable); ok {
            if sourceAny := any(schema); sourceAny != nil {
                if sourceCloneable, ok := sourceAny.(Cloneable); ok {
                    cloneable.CloneFrom(sourceCloneable)
                }
            }
        }

        // Execute onattach callbacks
        if check != nil {
            if checkInternals := check.GetZod(); checkInternals != nil {
                for _, fn := range checkInternals.OnAttach {
                    fn(newSchema)
                }
            }
        }

        return newSchema
    }

    panic(fmt.Sprintf("No constructor found for type: %T", schema))
}
```

### 4. Parameter Processing Tools (`checks.go`)

**Responsibility**: Provides unified parameter processing mechanism

```go
// ApplySchemaParams applies schema parameters to a ZodCheckDef
func ApplySchemaParams(def *ZodCheckDef, params ...SchemaParams) {
    if len(params) == 0 {
        return
    }

    param := params[0]

    // Apply error if provided
    if param.Error != nil {
        switch err := param.Error.(type) {
        case string:
            errorMsg := err
            errorMap := ZodErrorMap(func(ZodRawIssue) string {
                return errorMsg
            })
            def.Error = &errorMap
        case ZodErrorMap:
            def.Error = &err
        case *ZodErrorMap:
            def.Error = err
        case func(ZodRawIssue) string:
            errorMap := ZodErrorMap(err)
            def.Error = &errorMap
        }
    }

    // Apply abort flag if provided
    if param.Abort {
        def.Abort = true
    }
}
```

## 🔧 Implementation Steps

### Step 1: Define Common Check Types

```go
// String length checks
type ZodMinLengthCheck struct {
    internals *ZodCheckInternals
    min       int
}

func (z *ZodMinLengthCheck) GetZod() *ZodCheckInternals {
    return z.internals
}

func NewMinLength(min int, params ...SchemaParams) *ZodMinLengthCheck {
    def := &ZodCheckDef{
        Check: "min",
        Abort: false,
    }
    ApplySchemaParams(def, params...)

    internals := &ZodCheckInternals{
        Def: def,
        Check: func(payload *ParsePayload) {
            if hasLength(payload.Value) {
                length := getLength(payload.Value)
                if length < min {
                    issue := ZodRawIssue{
                        Code:    string(TooSmall),
                        Message: fmt.Sprintf("String must contain at least %d character(s)", min),
                        Path:    payload.Path,
                    }
                    payload.Issues = append(payload.Issues, issue)
                }
            }
        },
    }

    return &ZodMinLengthCheck{
        internals: internals,
        min:       min,
    }
}
```

### Step 2: Implement Custom Check Support

```go
// Generic custom check
type ZodCustomCheck[T any] struct {
    internals *ZodCheckInternals
    fn        func(T) bool
}

func (z *ZodCustomCheck[T]) GetZod() *ZodCheckInternals {
    return z.internals
}

func NewCustom[T any](fn func(T) bool, params ...SchemaParams) *ZodCustomCheck[T] {
    def := &ZodCheckDef{
        Check: "custom",
        Abort: false,
    }
    ApplySchemaParams(def, params...)

    internals := &ZodCheckInternals{
        Def: def,
        Check: func(payload *ParsePayload) {
            if value, ok := payload.Value.(T); ok {
                if !fn(value) {
                    issue := ZodRawIssue{
                        Code:    string(CustomIssue),
                        Message: "Custom validation failed",
                        Path:    payload.Path,
                    }
                    payload.Issues = append(payload.Issues, issue)
                }
            }
        },
    }

    return &ZodCustomCheck[T]{
        internals: internals,
        fn:        fn,
    }
}
```

### Step 3: Add Conditional Execution Support

```go
// Conditional check wrapper
type ZodConditionalCheck struct {
    internals *ZodCheckInternals
    condition func(*ParsePayload) bool
    check     ZodCheck
}

func (z *ZodConditionalCheck) GetZod() *ZodCheckInternals {
    return z.internals
}

func When(condition func(*ParsePayload) bool, check ZodCheck) *ZodConditionalCheck {
    def := &ZodCheckDef{
        Check: "conditional",
        Abort: false,
    }

    internals := &ZodCheckInternals{
        Def: def,
        When: condition,
        Check: func(payload *ParsePayload) {
            if condition(payload) {
                check.GetZod().Check(payload)
            }
        },
    }

    return &ZodConditionalCheck{
        internals: internals,
        condition: condition,
        check:     check,
    }
}
```

### Step 4: Implement Format Validation Checks

```go
// Email format check
type ZodEmailCheck struct {
    internals *ZodCheckInternals
    pattern   *regexp.Regexp
}

func (z *ZodEmailCheck) GetZod() *ZodCheckInternals {
    return z.internals
}

func NewEmail(params ...SchemaParams) *ZodEmailCheck {
    def := &ZodCheckDef{
        Check: "email",
        Abort: false,
    }
    ApplySchemaParams(def, params...)

    emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

    internals := &ZodCheckInternals{
        Def: def,
        Check: func(payload *ParsePayload) {
            if str, ok := payload.Value.(string); ok {
                if !emailRegex.MatchString(str) {
                    issue := ZodRawIssue{
                        Code:    string(InvalidString),
                        Message: "Invalid email format",
                        Path:    payload.Path,
                    }
                    payload.Issues = append(payload.Issues, issue)
                }
            }
        },
    }

    return &ZodEmailCheck{
        internals: internals,
        pattern:   emailRegex,
    }
}
```

## 🧪 Testing Strategy

### 1. Basic Check Tests

```go
func TestMinLengthCheck(t *testing.T) {
    check := NewMinLength(5)
    
    t.Run("passes for valid length", func(t *testing.T) {
        payload := &ParsePayload{
            Value:  "hello",
            Issues: make([]ZodRawIssue, 0),
            Path:   []interface{}{},
        }
        
        check.GetZod().Check(payload)
        assert.Empty(t, payload.Issues)
    })
    
    t.Run("fails for invalid length", func(t *testing.T) {
        payload := &ParsePayload{
            Value:  "hi",
            Issues: make([]ZodRawIssue, 0),
            Path:   []interface{}{},
        }
        
        check.GetZod().Check(payload)
        assert.Len(t, payload.Issues, 1)
        assert.Equal(t, string(TooSmall), payload.Issues[0].Code)
    })
}
```

### 2. Custom Check Tests

```go
func TestCustomCheck(t *testing.T) {
    check := NewCustom(func(s string) bool {
        return strings.Contains(s, "@")
    })
    
    t.Run("passes for valid custom condition", func(t *testing.T) {
        payload := &ParsePayload{
            Value:  "user@example.com",
            Issues: make([]ZodRawIssue, 0),
            Path:   []interface{}{},
        }
        
        check.GetZod().Check(payload)
        assert.Empty(t, payload.Issues)
    })
    
    t.Run("fails for invalid custom condition", func(t *testing.T) {
        payload := &ParsePayload{
            Value:  "invalid",
            Issues: make([]ZodRawIssue, 0),
            Path:   []interface{}{},
        }
        
        check.GetZod().Check(payload)
        assert.Len(t, payload.Issues, 1)
        assert.Equal(t, string(CustomIssue), payload.Issues[0].Code)
    })
}
```

### 3. Integration Tests

```go
func TestCheckIntegration(t *testing.T) {
    schema := gozod.String().Min(5).Max(10).Email()
    
    t.Run("all checks pass", func(t *testing.T) {
        result, err := schema.Parse("user@example.com")
        assert.NoError(t, err)
        assert.Equal(t, "user@example.com", result)
    })
    
    t.Run("first check fails", func(t *testing.T) {
        _, err := schema.Parse("a@b.c")
        assert.Error(t, err)
        var zodErr *gozod.ZodError
        assert.True(t, errors.As(err, &zodErr))
        assert.Len(t, zodErr.Issues, 1)
        assert.Equal(t, string(TooSmall), zodErr.Issues[0].Code)
    })
}
```

## 📖 Usage Examples

### Basic Check Usage

```go
// String length validation
schema := gozod.String().Min(5).Max(10)
result, err := schema.Parse("hello") // OK
result, err := schema.Parse("hi")    // Error: too short
```

### Custom Validation

```go
// Custom validation with error message
schema := gozod.String().Refine(func(s string) bool {
    return strings.Contains(s, "@")
}, gozod.SchemaParams{Error: "Must contain @ symbol"})

result, err := schema.Parse("user@example.com") // OK
result, err := schema.Parse("invalid")          // Error: Must contain @ symbol
```

### Complex Validation Chains

```go
// Multiple validations with different error handling
schema := gozod.String().
    Min(5, gozod.SchemaParams{Error: "Too short"}).
    Max(20, gozod.SchemaParams{Error: "Too long"}).
    Email(gozod.SchemaParams{Error: "Invalid email format"})

result, err := schema.Parse("user@example.com") // OK
result, err := schema.Parse("hi@a.b")           // Error: Too short
result, err := schema.Parse("verylongemailaddress@example.com") // Error: Too long
result, err := schema.Parse("invalid-email")   // Error: Invalid email format
```

This implementation guide provides the complete foundation for implementing the Checks mechanism in GoZod while maintaining full TypeScript compatibility and Go type safety. 

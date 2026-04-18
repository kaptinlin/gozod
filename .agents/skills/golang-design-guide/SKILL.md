---
description: Comprehensive architecture design guide for Golang libraries. Use when designing new Go packages, libraries, or services that need DESIGN.md or CLAUDE.md documentation. Triggers on library design patterns or when creating new Go modules.
name: golang-design-guide
---


# Golang Library Architecture Design Guide

## Overview

This skill provides comprehensive guidance for designing Golang libraries with clear architecture, consistent patterns, and maintainable code. It helps you create DESIGN.md and CLAUDE.md documentation that captures design philosophy, API patterns, and implementation guidelines.

## Library Type Classification

Different types of Go libraries require different design approaches. Identify your library type first:

### 1. Core Types & State Machines
**Examples:** go-fsm, go-command
**Characteristics:** Generic-based, zero-allocation hot paths, immutable runtime, compile-time type safety
**Key Patterns:** Two-phase separation (build vs runtime), internal array lookup, dual declaration styles

### 2. Data Processing & Transformation
**Examples:** jsondiff, jsonmerge, jsonpatch, jsonschema
**Characteristics:** RFC compliance, flat output schemas, zero-copy optimization, type-safe generics
**Key Patterns:** Fast/slow path separation, benchmark-driven performance, minimal API surface

### 3. Infrastructure & Storage
**Examples:** go-cache, go-secrets, go-config
**Characteristics:** Store/Source interfaces, pluggable backends, explicit configuration, lock-free reads
**Key Patterns:** Interface segregation, functional options, policy-based error handling

### 4. Framework-Agnostic Libraries
**Examples:** go-crawler, requests
**Characteristics:** Embeddable core, dual interface (CLI + library), middleware architecture
**Key Patterns:** Two-layer architecture (API + Core), capability interfaces, graceful degradation

### 5. HTTP & Network Libraries
**Examples:** requests
**Characteristics:** Fluent builder pattern, middleware chains, streaming support, retry mechanisms
**Key Patterns:** Builder pattern, buffer pooling, zero-panic policy

**See references/library-types.md for detailed patterns and examples for each type.**

## Design Philosophy Framework

Every library should explicitly state its design philosophy using these principles:

### KISS (Keep It Simple, Stupid)
- Prefer simple solutions over clever abstractions
- Standard library first, custom implementations only when necessary
- Clear code beats micro-optimizations
- Three similar lines beat a single-use helper

### DRY (Don't Repeat Yourself)
- Reuse stdlib packages (slices, maps, cmp, errors)
- Shared patterns across packages (functional options, registry patterns)
- Eliminate duplicate validation, error handling, test setup

### YAGNI (You Aren't Gonna Need It)
- Solve current problems, not hypothetical futures
- Evidence-based optimization (profile before adding sync.Pool)
- No speculative features, unused configurability, or "just in case" abstractions

### Single Responsibility
- Each package has one clear purpose
- Thin orchestration layers with no business logic
- Capability packages depend only on foundation types

### Consumer-First Simplicity (Apple philosophy + Go idioms)
- Design for the 90% path first: most users should import one package and remember one constructor
- Default usage must not require platform checks or understanding internal layering
- Hide discoverer / watcher / prober style assembly behind the root package's best default entry point
- Advanced composition should remain possible, but it must be clearly secondary to the simplest path
- Natural usage should equal correct usage: the shortest path should also be the recommended path

**See references/design-philosophy.md for real-world examples and anti-patterns.**

## API Design Patterns

### Functional Options Pattern
**When to use:** Optional configuration, backward compatibility, sensible defaults

```go
type Config struct {
    Store        Store
    Codec        Codec
    DefaultTTL   time.Duration
}

type Option func(*Config)

func WithTTL(ttl time.Duration) Option {
    return func(c *Config) { c.DefaultTTL = ttl }
}

func New[T any](store Store, opts ...Option) Cache[T] {
    cfg := &Config{DefaultTTL: 5 * time.Minute}
    for _, opt := range opts {
        opt(cfg)
    }
    return &cache[T]{config: cfg}
}
```

### Builder Pattern
**When to use:** Complex object construction, fluent API, method chaining

```go
type Builder[S, E comparable] struct {
    initial S
    rules   []rule[S, E]
}

func New[S, E comparable](initial S) *Builder[S, E]

func (b *Builder[S, E]) From(states ...S) *StateGroup[S, E]
func (b *Builder[S, E]) Final(states ...S) *Builder[S, E]
func (b *Builder[S, E]) Build() (*Machine[S, E], error)
```

### Interface Segregation
**When to use:** Pluggable components, single-method interfaces, optional capabilities

```go
// Core interface - single method
type Source interface {
    Load(ctx context.Context) (map[string]any, error)
}

// Optional capabilities - checked via type assertion
type Watcher interface {
    Watch(ctx context.Context, onChange func(map[string]any)) error
}

type StatusReporter interface {
    Status(fn func(changed bool, err error))
}
```

**See references/api-patterns.md for complete pattern catalog with examples.**

## Error Handling Strategy

### Sentinel Errors
```go
// Define as package-level variables
var (
    ErrNotFound   = errors.New("not found")
    ErrClosed     = errors.New("store closed")
    ErrInvalidKey = errors.New("invalid key")
)

// Or as custom error types for structured errors
type Error struct {
    Code    int
    Message string
    Err     error
}

func (e *Error) Error() string { return e.Message }
func (e *Error) Unwrap() error { return e.Err }
```

### Error Wrapping
```go
// Use %w for error chains
return fmt.Errorf("fetch failed: %w", err)

// Join multiple errors (Go 1.20+)
return errors.Join(ErrValidation, err)

// Check with errors.Is/As
if errors.Is(err, ErrNotFound) { ... }
```

### Policy-Based Error Handling
```go
type Policy int

const (
    PolicyRequired Policy = iota  // Error is fatal
    PolicyOptional                // Error is recorded, source skipped
)

type SkipReporter interface {
    SkipError() error  // Expose absorbed errors
}
```

**See references/error-handling.md for complete error handling patterns.**

## Testing Strategy

### Test Structure
```go
func TestFeature(t *testing.T) {
    tests := []struct {
        name    string
        input   Input
        want    Output
        wantErr error
    }{
        {"success case", validInput, expectedOutput, nil},
        {"error case", invalidInput, Output{}, ErrInvalid},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            got, err := Feature(tt.input)
            if tt.wantErr != nil {
                require.ErrorIs(t, err, tt.wantErr)
                return
            }
            require.NoError(t, err)
            assert.Equal(t, tt.want, got)
        })
    }
}
```

### Benchmark Patterns (Go 1.24+)
```go
func BenchmarkOperation(b *testing.B) {
    setup := prepareData()
    b.ResetTimer()

    for b.Loop() {  // Go 1.24+ b.Loop() instead of for i := 0; i < b.N; i++
        _ = Operation(setup)
    }
}
```

### Coverage Thresholds
- Core packages: 80-90%
- Infrastructure: 70-80%
- Examples: Excluded

**See references/testing-patterns.md for comprehensive testing guide.**

## Go 1.20-1.26 Modern Features

### Recommended Features (Use Liberally)

**Go 1.20:**
- `errors.Join()` - Aggregate multiple errors
- `strings.Cut/CutPrefix/CutSuffix` - String manipulation

**Go 1.21:**
- `min()`, `max()`, `clear()` - Built-in functions
- `slices` package - Sort, Grow, Contains, etc.
- `maps` package - Clone, Copy, Equal
- `cmp` package - Compare for ordered types

**Go 1.22:**
- `for range N` - Integer range loops
- Loop variable scoping - Automatic fix for capture issues

**Go 1.23:**
- `iter.Seq/Seq2` - Custom iterators for lazy evaluation

**Go 1.24:**
- `testing.B.Loop()` - Replace `for i := 0; i < b.N; i++`
- Swiss Tables - Automatic 10-35% map performance boost
- `NewGCMWithRandomNonce` - Automatic nonce management

**Go 1.26:**
- `testing/synctest` - Testing concurrent/time-dependent code
- `sync.WaitGroup.Go()` - Goroutine management

### Features to Avoid/Use Carefully

**❌ sync.Pool - Only Use With Profiling Evidence**
```go
// DON'T use speculatively
// DO use only when profiling shows allocation bottlenecks
```

**❌ Custom Iterators (iter package) - Rarely Needed**
```go
// DON'T add unless there's a clear use case
// Users can write their own for loops when needed
```

**❌ log/slog in Libraries - Let Callers Control Logging**
```go
// DON'T add log/slog dependencies to library packages
// DO provide error returns and let application code handle logging
```

**See references/modern-go-features.md for complete feature guide.**

## Documentation Structure

### DESIGN.md Template
```markdown
# Library Name Design

## Design Principles
- KISS/DRY/YAGNI statements
- Single Responsibility
- Open/Closed

## Architecture
- Core components diagram
- Data flow
- Key types and interfaces

## Rejected Patterns
- Explicitly document what you chose NOT to do and why

## Usage Examples
- Basic usage
- Advanced patterns
- Integration examples
```

### CLAUDE.md Template
```markdown
# CLAUDE.md

## Project Overview
- Module path, Go version, purpose

## Commands
- task test, task lint, task verify

## Architecture
- Package structure
- Dependency rules
- Data flow

## Coding Rules
- Must Follow
- Forbidden
- Go 1.20-1.26 Features

## Testing
- Test structure
- Coverage thresholds
```

**See references/documentation-templates.md for complete templates.**

## Workflow

### 1. Classify Library Type
Identify which library type best matches your use case (see Library Type Classification above).

### 2. Define Design Philosophy
Write explicit KISS/DRY/YAGNI statements with concrete examples from your domain.

### 3. Choose API Patterns
Select appropriate patterns based on library type:
- Core types → Builder + Generics
- Data processing → Functional options + Fast/slow paths
- Infrastructure → Interface segregation + Functional options
- Framework-agnostic → Two-layer architecture + Middleware

### 4. Design Error Handling
Choose sentinel errors, structured errors, or policy-based approach based on error semantics.

### 5. Plan Testing Strategy
Define coverage thresholds, test patterns, and benchmark requirements.

### 6. Document Architecture
Create DESIGN.md with principles, architecture, rejected patterns, and examples.

### 7. Write Implementation Guide
Create CLAUDE.md with commands, coding rules, and testing guidelines.

## References

This skill includes detailed reference files for deep dives:

- **references/library-types.md** - Detailed patterns for each library type with real examples
- **references/design-philosophy.md** - KISS/DRY/YAGNI with anti-patterns and real-world examples
- **references/api-patterns.md** - Complete API pattern catalog with code examples
- **references/error-handling.md** - Comprehensive error handling strategies

Read these references as needed based on your specific design challenges.

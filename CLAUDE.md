# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**GoZod** is a TypeScript Zod v4-inspired validation library for Go, providing strongly-typed, zero-dependency data validation with intelligent type inference. It maintains API compatibility with TypeScript Zod v4 while leveraging Go's type system for compile-time safety and maximum performance.

### Zod v4 Reference Integration

GoZod uses the actual TypeScript Zod v4 source code as reference material located in `.reference/zod/packages/zod/src/v4/`. This ensures accurate API correspondence and semantic compatibility with the upstream TypeScript implementation.

**Key Reference Files:**
- `.reference/zod/packages/zod/src/v4/core/schemas.ts` - Schema type definitions
- `.reference/zod/packages/zod/src/v4/core/checks.ts` - Validation check implementations  
- `.reference/zod/packages/zod/src/v4/core/api.ts` - Public API factory functions
- `.reference/zod/packages/zod/src/v4/core/errors.ts` - Error handling system

### Key Features
- **Complete Strict Type Semantics** - All methods require exact input types, zero automatic conversions
- **Parse vs StrictParse Methods** - Runtime flexibility or compile-time type safety for optimal performance
- **Maximum Performance** - Zero-overhead validation with optional code generation (5-10x faster)
- **TypeScript Zod v4 Compatible API** - Familiar syntax with Go-native optimizations
- **Declarative Struct Tags** - Define validation rules with `gozod:"required,min=2,email"` syntax
- **Custom Validator System** - User-defined validators with registry and struct tag integration
- **Automatic Circular Reference Handling** - Lazy evaluation prevents stack overflow
- **Rich Validation Methods** - Comprehensive built-in validators for all Go types
- **Tuple Types** - Fixed-length arrays with per-position type validation and optional rest elements
- **Exclusive Unions (Xor)** - Exactly-one-match validation for mutually exclusive schemas
- **Schema Metadata** - Built-in `.Describe()` and `.Meta()` methods on all 26 schema types

### Cursor Rules Index

Detailed implementation guides are available in `.cursor/rules/`:

| File | Purpose |
|------|---------|
| `coding-standards.mdc` | Core coding standards, design patterns (Parse/StrictParse, Copy-on-Write, Metadata) |
| `schema_implementation_guide.mdc` | 5-section file layout, type templates, engine integration |
| `schema_test_implementation_guide.mdc` | Test architecture, StrictParse testing, Default/Prefault semantics |
| `checks_implementation_guide.mdc` | Validation check factories, JSON Schema integration |
| `performance-optimization.mdc` | Go 1.25+ optimizations, benchmark patterns (`b.Loop()`) |
| `project-structure.mdc` | Layered architecture, package responsibilities, file organization |
| `naming_guide.md` | Go naming conventions, receiver naming, error naming |
| `module_organization_guide.md` | Package design, dependency injection, testing organization |

## Commands

### Testing
```bash
# Run all tests with race detection
make test

# Run individual package tests  
go test ./types/
go test ./internal/engine/
go test -run TestSpecificFunction ./types/

# Run tests with verbose output
go test -v ./...
```

### Linting and Code Quality
```bash
# Run full lint suite (includes golangci-lint and mod tidy)
make lint

# Run only golangci-lint
make golangci-lint

# Run only module tidy check
make tidy-lint
```

### Building
```bash
# Download dependencies
go mod download

# Build all packages (verify compilation)
go build ./...

# Clean build artifacts
make clean
```

## High-Level Architecture

### Core Design Philosophy
1. **Complete Strict Type Semantics** - All methods require exact input types, zero automatic conversions
2. **Parse vs StrictParse Duality** - Runtime flexibility (`Parse`) or compile-time type safety (`StrictParse`)
3. **Input-Output Symmetry** - Schemas return the same type they accept after validation 
4. **Copy-on-Write Modifiers** - Modifiers like `.Optional()` clone internals and return new instances
5. **TypeScript Zod v4 Compatibility** - Maintains identical API surface and behavior

### Layered Architecture
```
gozod/
├── .reference/    # TypeScript Zod v4 source code for API reference
├── bin/           # Build outputs and tools
├── cmd/           # Command-line tools (gozodgen code generator)
├── coerce/        # Root-level coercion utilities  
├── core/          # Foundation contracts (interfaces, types, constants)
├── docs/          # Documentation and guides
├── examples/      # Example implementations and usage patterns
├── internal/      # Private runtime engine (parser, checks, issues)
├── locales/       # Internationalization bundles
├── pkg/           # Reusable utilities (validate, coerce, reflectx, validators)  
└── types/         # Public schema implementations (String, Array, etc.)
```

#### One-Way Dependency Rule
- `types/` never import each other; all cross-type logic lives in `internal/`, `pkg/`, or `coerce/`
- Core layer contains zero business logic; only defines contracts and helpers
- Strict separation of concerns: parsing, checking, issue creation, and error formatting live in separate packages
- Root-level files: `gozod.go` (main API), `json_schema.go` (JSON Schema generation), `from_json_schema.go` (JSON Schema → GoZod conversion), `Makefile` (build automation)

### Key Components

#### Core Layer (`core/`) - Foundation Contracts
- `interfaces.go` - Primary interfaces (`ZodType[T]`, `ZodSchema`, `Cloneable`)
- `definition.go` - Schema definition primitives (`ZodTypeDef`, `SchemaParams`)
- `parsing.go` - Parse contexts and utilities
- `transform.go` - Transformation & pipeline primitives
- `checks.go` - Generic check contracts & helpers
- `context.go` - Parse context, payload & path helpers
- `issues.go` - Error types (`ZodIssue`, `ZodError`)
- `constants.go` - Library-wide constants (issue codes, type ids)
- `config.go` - Global config & default error map
- `registry.go` - Schema metadata and registration system

#### Types Layer (`types/`) - Schema Implementations
Each file implements one schema type following the structured template:
```go
// Type Constraints
type StringConstraint interface {
    string | *string
}

// Type Definitions  
type ZodStringDef struct {
    core.ZodTypeDef
}

type ZodStringInternals struct {
    core.ZodTypeInternals
    Def *ZodStringDef
}

type ZodString[T StringConstraint] struct {
    internals *ZodStringInternals
}

// Core Methods
func (z *ZodString[T]) Parse(input any, ctx ...*core.ParseContext) (T, error)
func (z *ZodString[T]) StrictParse(input T, ctx ...*core.ParseContext) (T, error)
func (z *ZodString[T]) MustParse(input any, ctx ...*core.ParseContext) T
func (z *ZodString[T]) MustStrictParse(input T, ctx ...*core.ParseContext) T

// Validation Methods
func (z *ZodString[T]) Min(min int, params ...core.SchemaParams) *ZodString[T]
func (z *ZodString[T]) Max(max int, params ...core.SchemaParams) *ZodString[T]

// Modifier Methods
func (z *ZodString[T]) Optional() *ZodString[*string]
func (z *ZodString[T]) Nilable() *ZodString[*string]
```

**File Organization:**
- One schema type per file (e.g., `string.go`, `integer.go`)
- Schema types organized by category:
  - **Primitives**: `string.go`, `bool.go`, `integer.go`, `float.go`, `bigint.go`, `complex.go`
  - **Special Types**: `any.go`, `unknown.go`, `never.go`, `nil.go`
  - **Collections**: `array.go`, `slice.go`, `tuple.go`, `map.go`, `record.go`
  - **Objects**: `object.go`, `struct.go`
  - **Composition**: `union.go` (includes Xor), `discriminated_union.go`, `intersection.go`
  - **Functions**: `function.go`
  - **Formats**: `email.go`, `network.go`, `ids.go`, `iso.go`, `time.go`, `file.go`
  - **Text**: `text.go`, `stringbool.go`
  - **Advanced**: `lazy.go`, `literal.go`, `enum.go`, `transform.go`
- All schema files include comprehensive test counterparts (`*_test.go`)

#### Internal Engine (`internal/`) - Runtime Engine (Private)
- `engine/` - Generic parser, coercion and check runner
  - `parser.go`, `checker.go`, `types.go`, `errors.go`, `params.go`
- `checks/` - Concrete check factories (length, numeric, format, metadata)
  - `checks.go`, `custom.go`, `format.go`, `length.go`, `numeric.go`, `strings.go`, `file.go`, `metadata.go`
- `issues/` - Raw issue builders, finalisers & formatters
  - `creators.go`, `formatters.go`, `finalize.go`, `errors.go`, `types.go`, `raw.go`, `accessors.go`
- `utils/` - Low-level helpers used only by the engine
  - `utils.go`

#### Pkg Layer (`pkg/`) - Utilities
Stateless, allocation-conscious helpers independent of GoZod runtime:
- `validate/` - Atomic, reflection-free validators (length, numeric, regex, ISO 8601)
- `coerce/` - Loss-less type coercion helpers (`ToInt64`, `ToBool`)
- `reflectx/` - Safe reflection helpers, type guards, deref utils (`pointer.go`, `types.go`, `utils.go`, `values.go`)
- `mapx/` - Generic map manipulation (string & generic keys)
- `slicex/` - Generic slice helpers (merge, unique, map)
- `structx/` - Struct ⇄ map conversion without `encoding/json`
- `tagparser/` - Struct tag parsing system (`parser.go`, `parser_test.go`)
- `regexes/` - Pre-compiled, namespaced regex library (emails, networks, datetimes, etc.)
- `validators/` - Custom validator system with registry and struct tag integration

### Type System Design

#### Complete Strict Type Semantics
All schemas require exact input types with zero automatic conversions:
- `String()` requires `string` input, returns `string`
- `StringPtr()` requires `*string` input, returns `*string` 
- No automatic conversions between value and pointer types
- For flexible input handling, use Optional/Nilable modifiers

#### Parse vs StrictParse Duality
```go
schema := gozod.String().Min(3)

// Parse - Runtime type checking (flexible input, any → T)
result, err := schema.Parse("hello")        // ✅ Works with any input type
result, err = schema.Parse(42)              // ❌ Runtime error: invalid type

// StrictParse - Compile-time type safety (exact input, T → T)
str := "hello"
result, err = schema.StrictParse(str)       // ✅ Compile-time guarantee, optimal performance
// result, err := schema.StrictParse(42)    // ❌ Compile-time error
```

#### Modifier Pattern
```go
// Optional/Nilable: flexible input, pointer output
schema := String().Optional()  // Flexible input (string/*string), returns *string

// Default/Prefault: preserve current type
schema := String().Default("hello")  // Returns *ZodString[string]
```

#### Constructor Pattern
Every type has value and pointer constructors:
```go
String()    // Creates ZodString[string]
StringPtr() // Creates ZodString[*string]
Int()       // Creates ZodInteger[int, int]  
IntPtr()    // Creates ZodInteger[*int, *int]
```

#### Root-Level Coercion (`coerce/`)
High-level coercion utilities that bridge between `pkg/coerce` and the schema system:
- `coerce.go` - Type coercion integration with schema validation
- `coerce_test.go` - Tests for coercion functionality

### Custom Validator System (`pkg/validators/`)

#### Architecture Overview
The custom validator system provides a flexible, registry-based approach for user-defined validation:

```go
// Basic validator interface
type ZodValidator[T any] interface {
    Name() string
    Validate(value T) bool
    ErrorMessage(ctx *core.ParseContext) string
}

// Parameterized validator interface
type ZodParameterizedValidator[T any] interface {
    ZodValidator[T]
    ValidateParam(value T, param string) bool
    ErrorMessageWithParam(ctx *core.ParseContext, param string) string
}
```

#### Key Components
- `types.go` - Validator interfaces (`ZodValidator`, `ZodParameterizedValidator`, `ZodDetailedValidator`)
- `registry.go` - Thread-safe validator registry with type-safe registration and retrieval
- `converters.go` - Functions to convert validators to GoZod check functions
- `validators_test.go` - Comprehensive test suite for validator system

#### Usage Pattern
1. **Users implement validators** in their application code
2. **Register validators** using `validators.Register()`
3. **Use in struct tags** or programmatically with converter functions
4. **Struct tag integration** automatically applies registered validators

#### Integration Points
- **Struct Tags**: Custom validators work seamlessly with `gozod:"custom_validator_name"`
- **Programmatic**: Convert validators using `validators.ToRefineFn()` or `validators.ToCheckFn()`
- **Type Safety**: Registry maintains type safety with generic methods

## Development Guidelines

### File Organization Patterns
- **Schema Files**: One schema type per file in `types/` (e.g., `string.go`, `integer.go`)
- **Test Files**: Comprehensive test files mirror source file names (`string_test.go`, `integer_test.go`)
- **All files follow consistent template**: Type Constraints, Type Definitions, Core Methods, Validation Methods, Modifiers
- **Internal Organization**:
  - Check factories in `internal/checks/` grouped by concern (`numeric.go`, `strings.go`, `format.go`, `length.go`)
  - Engine components grouped by function (`parser.go`, `checker.go`, `types.go`, `errors.go`, `params.go`)
  - Issue handling separated by responsibility (`creators.go`, `formatters.go`, `errors.go`, `raw.go`, `accessors.go`, `finalize.go`)
- **Package Naming**: Never use underscores; files use underscores for clarity (`coerce_test.go`)

### Naming Conventions & Code Style
- **Schema constructors**: `String()`, `Int()`, `Bool()`
- **Pointer variants**: `StringPtr()`, `IntPtr()`, `BoolPtr()`
- **Methods follow Go conventions**: `Min()`, `Max()`, `Email()`
- **Go-style naming**: Public types/methods capitalized, meaningful names
- **Code formatting**: Always run `gofmt` and `goimports` before committing
- **Package organization**: Organize by function, avoid circular dependencies, minimize public API

### Default vs Prefault Semantics (Critical Zod v4 Compatibility)

#### Default - Short-Circuit Mechanism
- If input is `nil`, **directly returns default value**
- **Does NOT go through** schema parsing, validation, or transformation
- Default value must be compatible with schema's **output type**
- Zero overhead when triggered

#### Prefault - Preprocessing Mechanism  
- If input is `nil`, uses prefault value for **complete parsing flow**
- Prefault value must be compatible with schema's **input type**
- **Only triggered on `nil` input** - validation failures do NOT trigger prefault
- Goes through full validation and transformation pipeline

```go
// Default example - short-circuit
schema1 := String().Transform(func(s string) int { return len(s) }).Default("default")
result1, _ := schema1.Parse(nil) // => "default" (bypasses transform)

// Prefault example - preprocessing
schema2 := String().Transform(func(s string) int { return len(s) }).Prefault("hello")
result2, _ := schema2.Parse(nil) // => 5 ("hello" goes through transform)
```

### Key Development Patterns
1. **Copy-on-Write Modifiers** - Always clone internals before modification, ensure immutability
2. **Type-Safe Method Chaining** - Each method returns strongly-typed schema
3. **Input-Output Symmetry** - Schemas return same type they accept after validation
4. **Composable Validation Pipeline** - Checks are first-class citizens
5. **Zero-Allocation Fast Paths** - Optimize common validation scenarios
6. **Engine-First Architecture** - All parsing through `engine.ParsePrimitive` or `engine.ParseComplex`
7. **Helper Package Usage** - Never roll custom validation/parsing/coercion; use existing packages
8. **Three-Layer Template** - All schema files follow Definition/Internals/Public API layers

### Testing Approach
- **Unit Tests**: Comprehensive tests for each schema type in `*_test.go` files
- **Integration Tests**: Cross-package validation in `internal/` packages  
- **Performance Benchmarks**: Critical path optimization validation
- **Prefault/Default Testing**: Verify correct Zod v4 semantic behavior
- **Parse/StrictParse Coverage**: 
  - Test both Parse and StrictParse methods with identical validation logic
  - Verify Parse handles any input types correctly with runtime type checking
  - Verify StrictParse accepts exact constraint types with compile-time safety
  - Test error handling consistency between both methods
  - Benchmark performance differences between Parse and StrictParse
- **Panic Method Testing**: Verify MustParse and MustStrictParse panic behavior
- **Custom Validator Testing**: Verify validator registration, retrieval, and integration
- Follow existing test patterns for consistency

### Performance Optimization Guidelines

#### Core Optimization Principles
- **Elegance First**: Prioritize code readability over aggressive optimization
- **Progressive Improvement**: Implement optimizations incrementally
- **Balanced Approach**: 10-20% performance gains with maintained code quality

#### Memory & Runtime Optimization
- **Pre-compile regex patterns** at package level in `pkg/regex`
- **Minimize reflection usage** (relegated to `pkg/reflectx`)
- **Zero-overhead validation paths** with optimal execution
- **Object pools** for frequently allocated objects (`sync.Pool`)
- **String building** with `strings.Builder` instead of concatenation
- **Early return patterns** to avoid unnecessary computations
- **Hot path optimization** for most frequently executed code

#### Compile-Time Optimization
- **Generic constraints** to reduce runtime overhead
- **Type assertion optimization** using type switches
- **Inline functions** for simple operations
- **Pre-extracted constants** for magic numbers

### Struct Tag Development Guidelines

The GoZod struct tag system provides declarative validation through Go struct tags, complementing the programmatic API while maintaining full compatibility and performance.

#### Core Tag Design Principles
1. **API Compatibility**: All tag rules must directly correspond to programmatic API methods
2. **Default Optional**: All fields are optional by default, requiring explicit `required` tag for mandatory fields
3. **Comma Separation**: Multiple rules are comma-separated: `"required,min=2,max=50"`
4. **Parameter Format**: Parameters use equals notation: `"min=5"`, `"enum=red green blue"`
5. **Type Safety**: Tag parsing must maintain full type safety with compile-time validation where possible

#### Tag Usage Examples
```go
type User struct {
    Name     string `gozod:"required,min=2,max=50"`
    Email    string `gozod:"required,email"`
    Age      int    `gozod:"required,min=18,max=120"`
    Bio      string `gozod:"max=500"`           // Optional by default
    Internal string `gozod:"-"`                // Skip validation
}

// Generate schema from tags
schema := gozod.FromStruct[User]()
result, err := schema.Parse(user)
```

#### Custom Validator Integration
Custom validators integrate seamlessly with struct tags:
```go
// Register custom validator
validators.Register(&MyCustomValidator{})

// Use in struct tags
type User struct {
    Username string `gozod:"required,my_custom_validator"`
    Email    string `gozod:"required,email"`
}
```

## Go Language Idioms and Zod v4 Compatibility

### Core Design Philosophy: Semantic Equivalence over Literal Translation

GoZod maintains **semantic compatibility** with Zod v4 while embracing Go language conventions. This principle applies across naming, types, and error handling.

#### Type Naming Convention

**Principle**: Use Go's native type names, not TypeScript's

```go
// ✅ Correct: Go type names
ParsedTypeBool ParsedType = "bool"  // Go's boolean type is "bool"
ParsedTypeNil  ParsedType = "nil"   // Go's zero value is "nil"

// ❌ Incorrect: TypeScript type names
ParsedTypeBoolean ParsedType = "boolean"  // Would be unfamiliar to Go developers
ParsedTypeNull    ParsedType = "null"     // Not Go terminology
```

**Mapping Examples**:
| Zod v4 (TypeScript) | GoZod (Go) | Rationale |
|---------------------|------------|-----------|
| `z.boolean()` → `"boolean"` | `gozod.Bool()` → `"bool"` | Go's type is `bool` |
| `z.null()` → `"null"` | `gozod.Nil()` → `"nil"` | Go uses `nil`, not `null` |
| `z.number()` → `"number"` | `gozod.Number()` → `"number"` | Both languages use "number" |

**Error Message Comparison**:
```
// Zod v4 (JavaScript developers see familiar terms)
"Invalid input: expected string, received boolean"

// GoZod (Go developers see familiar terms)
"Invalid input: expected string, received bool"
```

Both are correct because they use terminology familiar to developers in their respective languages.

#### Go-Specific Type Extensions

GoZod includes types that don't exist in JavaScript but are fundamental to Go:

| Type | Purpose | Example Error |
|------|---------|---------------|
| `ParsedTypeFloat` | Distinguish int from float | `"expected number, received float"` |
| `ParsedTypeSlice` | Distinguish array from slice | `"expected array, received slice"` |
| `ParsedTypeComplex` | Go's complex numbers | `"expected number, received complex"` |
| `ParsedTypeStruct` | Go structs vs generic objects | `"expected object, received struct"` |

These provide **more precise error messages** than generic "number" or "object" would.

#### Error Structure Enhancement

**Zod v4 Approach**:
```typescript
// Stores input, calculates type dynamically during formatting
interface $ZodIssueInvalidType {
  expected: string;
  input?: unknown;  // Type calculated from this when formatting
}
```

**GoZod Approach**:
```go
// Pre-calculates and stores type for better Go error handling
type ZodIssueInvalidType struct {
    Expected core.ZodTypeCode `json:"expected"`
    Received core.ParsedType  `json:"received"`  // Pre-calculated
}
```

**Advantages of GoZod's approach**:
1. **Go Error Convention**: Errors should contain complete information
2. **Performance**: Avoid recalculating type on every format operation
3. **Independence**: Error can be serialized without the original input

### API Semantic Mapping

GoZod ensures **behavioral compatibility** while using Go naming:

```go
// Semantic equivalence maintained
// Zod v4: z.boolean().optional()
// GoZod:  gozod.Bool().Optional()
// Both: Accept bool/nil, return *bool

// Zod v4: z.string().nullable()
// GoZod:  gozod.String().Nilable()
// Both: Accept string/nil, return *string

// Type names reflect runtime behavior
// Zod v4: typeof value === "boolean" → "boolean"
// GoZod:  reflect.TypeOf(value).Kind() == reflect.Bool → "bool"
```

### Documentation Cross-Reference

See `docs/feature-mapping.md` for complete Zod v4 ↔ GoZod API mapping table.

---

## Important Design Decisions

1. **Complete Strict Type Semantics** - All methods require exact input types, zero automatic conversions
2. **Parse vs StrictParse Duality** - Runtime flexibility vs compile-time type safety for performance optimization
3. **Semantic Zod v4 Compatibility** - Maintain identical behavior with Go-native naming conventions
4. **Pointer Identity Preservation** - Optional/Nilable maintain input pointer addresses when appropriate
5. **Immutable Modifiers** - Modifiers create new instances via copy-on-write
6. **Unified Engine Architecture** - ParsePrimitive/ParseComplex for consistent behavior
7. **Go Idioms First** - Error values instead of exceptions, Go type names, interfaces over inheritance
8. **Zero Dependencies** - Pure Go implementation, no external libraries
9. **User-Controlled Custom Validators** - No pre-registered validators, users register their own
10. **Enhanced Error Information** - Pre-calculated type information in errors (Go enhancement)
11. **Schema Metadata on All Types** - Every schema type has `.Describe()` and `.Meta()` methods
12. **Safe Error-Returning APIs** - Pick/Omit/Extend return `(*ZodObject, error)` with Must variants available

## Error Handling

Errors follow structured patterns with rich formatting options:
```go
// Standard validation error
_, err := schema.Parse(input)
if err != nil {
    var zodErr *gozod.ZodError
    if gozod.IsZodError(err, &zodErr) {
        // Handle structured validation errors
        for _, issue := range zodErr.Issues {
            fmt.Printf("Error: %s at %v\n", issue.Message, issue.Path)
        }
    }
}
```

**Error System Features:**
- Structured error handling with custom formatting
- Internationalization support through `locales/` packages  
- Path tracking for nested validation failures
- Multiple output formats for different use cases

## Integration Points

- **JSON Schema Generation** - Convert GoZod schemas to JSON Schema format (supports Tuple, Xor, all types)
- **Internationalization** - Built-in i18n bundles in `locales/` packages
- **Custom Error Formatting** - Flexible error output via `internal/issues`
- **Metadata System** - Built-in `.Describe()` and `.Meta()` methods on all schema types, plus Registry API
- **Transform & Pipeline Support** - Composable data transformation
- **Custom Validator System** - User-defined validators with registry and struct tag integration
- **Code Generation** - Optional zero-reflection performance optimization
- **Apply Function** - Composable schema modifiers for clean method chaining

## API Consistency Requirements

**All schema types must implement:**
- `Parse(input any, ctx ...*ParseContext) (T, error)` - Runtime type checking with flexible input handling
- `StrictParse(input T, ctx ...*ParseContext) (T, error)` - Compile-time type safety with exact type matching
- `MustParse(input any, ctx ...*ParseContext) T` - Panic-based Parse for critical error scenarios
- `MustStrictParse(input T, ctx ...*ParseContext) T` - Panic-based StrictParse for type-safe critical scenarios
- `ParseAny(input any, ctx ...*ParseContext) (any, error)` - Untyped result for runtime scenarios
- `GetInternals() *core.ZodTypeInternals` - Access internal schema state
- `IsOptional() bool` - Check if schema accepts missing values
- `IsNilable() bool` - Check if schema allows explicit nil values
- `Describe(description string) *Schema` - Add description metadata (returns same schema type)
- `Meta(meta GlobalMeta) *Schema` - Add full metadata object (returns same schema type)

**Method Implementation Requirements:**
- Parse and StrictParse must use identical validation pipelines
- StrictParse should optimize for known input types when possible
- Both methods must handle modifiers (Optional, Default, etc.) consistently
- Error messages and paths should be identical between Parse and StrictParse
- MustParse/MustStrictParse should panic with the same error as their non-panic versions

**Engine API Usage Patterns:**
- **Primitive types**: Use `engine.ParsePrimitive` and `engine.ParsePrimitiveStrict`
- **Complex types**: Use `engine.ParseComplex` and `engine.ParseComplexStrict`
- **Never bypass engine APIs** - maintains consistent modifier handling
- **Type constraints**: Use proper generic constraints (`StringConstraint interface { string | *string }`)

## Code Quality Standards

### Documentation Requirements
- **Every public type**: Complete GoDoc comments
- **Every public method**: Clear usage examples and parameter descriptions
- **Complex logic**: Inline comments explaining reasoning
- **API changes**: Update corresponding documentation

### Testing Requirements
- **Unit tests**: Every public method with comprehensive coverage
- **Integration tests**: Complex workflows and cross-package validation
- **Performance tests**: Benchmark tests for critical paths
- **Race detection**: All tests must pass with `-race` flag

### Code Review Standards
- **Type safety**: All type assertions and conversions must be safe
- **Error handling**: All error cases properly handled following Go idioms
- **Performance**: No unnecessary allocations or computations
- **Compatibility**: Changes must not break TypeScript Zod v4 compatibility

## Development Workflow

### TypeScript Zod v4 Reference Workflow
1. **Consult reference implementation** in `.reference/zod/packages/zod/src/v4/`
2. **Map TypeScript patterns** to Go idioms while preserving semantic behavior
3. **Maintain API correspondence** with upstream Zod v4 factory functions
4. **Document TypeScript equivalents** in code comments for traceability

### Implementation Order
1. **Define type constraints** (Go union types for flexibility)
2. **Define type structure** (Def, Internals, main generic type)
3. **Implement core interfaces** (ZodType, Parse, StrictParse, MustParse, MustStrictParse methods)
4. **Add validation methods** (Min, Max, specific checks following Zod v4 check system)
5. **Implement modifiers** (Optional, Nilable, Default, Prefault)
6. **Add advanced features** (Refine, Transform, Pipe)
7. **Write comprehensive tests** with both Parse and StrictParse coverage
8. **Add performance benchmarks** comparing Parse vs StrictParse
9. **Add documentation and examples** showing both parsing approaches

### Build System Integration
```bash
# Standard development commands
make test     # Run all tests with race detection
make lint     # Run golangci-lint and mod tidy checks  
make build    # Build all packages
make clean    # Clean build artifacts
```

This architecture ensures type safety, performance, and semantic compatibility with TypeScript Zod v4 while embracing Go language idioms and providing maximum developer productivity.

## TypeScript Zod v4 Correspondence

### API Mapping
GoZod maintains close correspondence with Zod v4's API design:

**TypeScript Zod v4 Pattern:**
```typescript
// From .reference/zod/packages/zod/src/v4/core/api.ts
export function _string<T extends schemas.$ZodString>(
  Class: util.SchemaClass<T>,
  params?: string | $ZodStringParams
): T
```

**GoZod Go Pattern:**
```go
// Equivalent factory function in gozod.go
func String(params ...core.SchemaParams) *types.ZodString[string] {
    return types.NewZodString[string](params...)
}
```

### Check System Correspondence
GoZod's check system directly mirrors Zod v4's validation architecture:

**TypeScript Zod v4:**
```typescript
// From .reference/zod/packages/zod/src/v4/core/checks.ts
export interface $ZodCheckDef {
  check: string;
  error?: errors.$ZodErrorMap<never> | undefined;
  abort?: boolean | undefined;
}
```

**GoZod Go:**
```go
// From core/checks.go  
type ZodCheckDef struct {
    Check string       // Check type identifier
    Error *ZodErrorMap // Custom error mapping
    Abort bool         // Whether to abort on validation failure
}
```

### Schema Definition Correspondence
**TypeScript Zod v4:**
```typescript
// From .reference/zod/packages/zod/src/v4/core/schemas.ts
export interface $ZodTypeDef {
  type: "string" | "number" | "int" | ...;
  error?: errors.$ZodErrorMap<never> | undefined;
  checks?: checks.$ZodCheck<never>[];
}
```

**GoZod Go:**
```go
// From core/definition.go
type ZodTypeDef struct {
    Type     ZodTypeCode  // Type name using type-safe constants
    Error    *ZodErrorMap // Custom error handler  
    Checks   []ZodCheck   // Validation checks
}
```

This ensures that GoZod behavior remains consistent with TypeScript Zod v4 while leveraging Go's type system for enhanced safety and performance.
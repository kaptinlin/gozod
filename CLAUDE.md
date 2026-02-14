# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**GoZod** is a TypeScript Zod v4-inspired validation library for Go, providing strongly-typed, zero-dependency data validation with intelligent type inference. It maintains API compatibility with TypeScript Zod v4 while leveraging Go's type system for compile-time safety and maximum performance.

### Zod v4 Reference

TypeScript Zod v4 source code is located in `.reference/zod/packages/zod/src/v4/` for API correspondence. Key files: `core/schemas.ts`, `core/checks.ts`, `core/api.ts`, `core/errors.ts`.

### Agent Rules Index

Detailed implementation guides in `.agents/rules/`:

| File | Purpose |
|------|---------|
| `coding-standards.mdc` | Core coding standards, design patterns (Parse/StrictParse, Copy-on-Write, Metadata) |
| `schema_implementation_guide.mdc` | 5-section file layout, type templates, engine integration |
| `schema_test_implementation_guide.mdc` | Test architecture, StrictParse testing, Default/Prefault semantics |
| `checks_implementation_guide.mdc` | Validation check factories, JSON Schema integration |
| `performance-optimization.mdc` | Go 1.26+ optimizations, benchmark patterns (`b.Loop()`) |
| `project-structure.mdc` | Layered architecture, package responsibilities, file organization |
| `naming_guide.md` | Go naming conventions, receiver naming, error naming |
| `module_organization_guide.md` | Package design, dependency injection, testing organization |

## Commands

```bash
make test           # Run all tests with race detection
make lint           # Run golangci-lint + mod tidy
make golangci-lint  # Run only golangci-lint
make tidy-lint      # Run only module tidy check
go build ./...      # Verify compilation
make clean          # Clean build artifacts

# Individual package tests
go test ./types/
go test -run TestSpecificFunction ./types/
```

## Architecture

```
gozod/
├── .reference/    # TypeScript Zod v4 source (read-only)
├── .agents/rules/ # Implementation guides for AI agents
├── cmd/           # Command-line tools (gozodgen)
├── coerce/        # Root-level coercion utilities
├── core/          # Foundation contracts (interfaces, types, constants)
├── docs/          # User-facing documentation
├── examples/      # Example implementations
├── internal/      # Private runtime (engine, checks, issues, utils)
├── jsonschema/    # JSON Schema conversion (to/from)
├── locales/       # Internationalization bundles
├── pkg/           # Reusable utilities (validate, coerce, reflectx, mapx, regex, slicex, structx, tagparser, transform)
└── types/         # Public schema implementations (one type per file)
```

### One-Way Dependency Rule

- `types/` never import each other; cross-type logic lives in `internal/`, `pkg/`, or `coerce/`
- Core layer contains zero business logic; only defines contracts
- Root-level files: `gozod.go` (main API re-exports all types, constructors, and JSON Schema conversion from subpackages)

### Schema Type Categories (`types/`)

- **Primitives**: `string.go`, `bool.go`, `integer.go`, `float.go`, `bigint.go`, `complex.go`
- **Special**: `any.go`, `unknown.go`, `never.go`, `nil.go`
- **Collections**: `array.go`, `slice.go`, `tuple.go`, `map.go`, `record.go`, `set.go`
- **Objects**: `object.go`, `struct.go`
- **Composition**: `union.go`, `xor.go`, `discriminated_union.go`, `intersection.go`
- **Functions**: `function.go`
- **Formats**: `email.go`, `network.go`, `ids.go`, `iso.go`, `time.go`, `file.go`
- **Text**: `text.go`, `stringbool.go`
- **Advanced**: `lazy.go`, `literal.go`, `enum.go`

## Core Design Principles

1. **Complete Strict Type Semantics** - All methods require exact input types, zero automatic conversions
2. **Parse vs StrictParse Duality** - `Parse(any)` for runtime flexibility, `StrictParse(T)` for compile-time safety
3. **Input-Output Symmetry** - Schemas return the same type they accept
4. **Copy-on-Write Modifiers** - `.Optional()`, `.Nilable()` clone internals and return new instances
5. **Engine-First Architecture** - All parsing through `engine.ParsePrimitive` or `engine.ParseComplex`
6. **Semantic Zod v4 Compatibility** - Identical behavior with Go-native naming (`"bool"` not `"boolean"`, `"nil"` not `"null"`)
7. **Go Idioms First** - Error values, Go type names, interfaces over inheritance
8. **Zero Dependencies** - Pure Go implementation, no external libraries

### Default vs Prefault Semantics

- **Default**: Short-circuit. If input is `nil`, directly returns default value (bypasses validation/transform).
- **Prefault**: Preprocessing. If input is `nil`, uses prefault value through the full parsing pipeline.

### Constructor Pattern

Every type has value and pointer constructors: `String()` / `StringPtr()`, `Int()` / `IntPtr()`, etc.

## API Consistency Requirements

All schema types must implement: `Parse`, `StrictParse`, `MustParse`, `MustStrictParse`, `ParseAny`, `GetInternals`, `IsOptional`, `IsNilable`, `Describe`, `Meta`.

Engine API usage:
- Primitive types: `engine.ParsePrimitive` / `engine.ParsePrimitiveStrict`
- Complex types: `engine.ParseComplex` / `engine.ParseComplexStrict`
- Never bypass engine APIs

## Error Handling

```go
_, err := schema.Parse(input)
if err != nil {
    var zodErr *gozod.ZodError
    if gozod.IsZodError(err, &zodErr) {
        for _, issue := range zodErr.Issues {
            fmt.Printf("Error: %s at %v\n", issue.Message, issue.Path)
        }
    }
}
```

## Quality Standards

- **Testing**: >90% coverage, all tests pass with `-race` flag, use `for b.Loop()` in benchmarks
- **Linting**: golangci-lint v2.9.0 clean, `gofmt`/`goimports` before committing
- **Type safety**: All type assertions must be safe; use `var zero R` for zero values (not `*new(R)`)
- **Compatibility**: Changes must not break Zod v4 semantic compatibility

## Documentation Cross-Reference

- `docs/feature-mapping.md` - Complete Zod v4 <> GoZod API mapping
- `docs/api.md` - Full API reference
- `docs/tags.md` - Struct tag validation guide
- `docs/json-schema.md` - JSON Schema integration

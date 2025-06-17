---
description: GoZod project structure and file organization guide
globs:
alwaysApply: true
---

# GoZod Project Structure and File Organization Guide

This document provides a comprehensive guide to GoZod project structure and file organization, based on actual implementation analysis. It explains architectural design, file purposes, and relationships between components, helping developers understand and navigate the codebase.

## 🎯 Core Design Philosophy

Based on smart type inference design principles:

1. **Input-Output Consistency** - Output the same type as input
2. **Unified Wrapper Pattern** - Consistent design for Default, Prefault, Optional, Nilable
3. **Separation of Concerns Principle** - Clear division of Refine, Transform, Coerce, Check
4. **Full TypeScript Zod v4 API Compatibility** - Precise implementation providing equivalent functionality

## 🏗️ Core Architecture

GoZod implements a **generic type-safe architecture** based on the `ZodType[In, Out any]` interface, providing compile-time type safety while maintaining input flexibility.

### Three-Layer Architecture Pattern

Every GoZod schema type follows a unified three-layer architecture:

1. **Definition Layer** - Basic type definitions and configuration information
2. **Internals Layer** - Internal type implementation, including parse functions, check lists, etc.
3. **Public Interface Layer** - Provides public APIs for user use

## 📁 Core File Structure

### Foundation Architecture Files

| File | Primary Responsibility | Description |
|------|----------------------|-------------|
| **type.go** | Core interfaces and basic type system | `ZodType` interface, `ZodTypeInternals`, `SchemaParams` |
| **parse.go** | Parse logic and context management | `ParseContext`, `ParsePayload`, global parse functions |
| **checks.go** | Validation check system | `ZodCheck`, validation functions, `AddCheck` mechanism |
| **issue.go** | Validation error definitions | `ZodIssue`, `ZodRawIssue`, error type definitions |
| **error.go** | Error handling and formatting | `ZodError`, error aggregation, formatting tools |
| **utils.go** | Common utility functions | Type conversion, helper functions, common logic |

### Functional Module Files

| File | Function | Description |
|------|----------|-------------|
| **transform.go** | Data transformation pipeline | `ZodTransform`, transformation logic, pipeline connection |
| **coerce.go** | Type coercion | Coercion namespace, factory functions |
| **config.go** | Global configuration | `ZodConfig`, global settings, localization configuration |

### Wrapper Type Files

| File | Wrapper Type | Description |
|------|--------------|-------------|
| **type_default.go** | Default wrapper | `ZodDefault`, default value mechanism |
| **type_prefault.go** | Prefault wrapper | `ZodPrefault`, fallback value mechanism |
| **type_optional.go** | Optional wrapper | `ZodOptional`, optional field handling |
| **type_nilable.go** | Nilable wrapper | `ZodNilable`, nullable value handling |

## 📊 Schema Type Implementation Files

### Basic Types

| File | Go Type | TypeScript Equivalent |
|------|---------|----------------------|
| **type_string.go** | `string` | `z.string()` |
| **type_boolean.go** | `bool` | `z.boolean()` |
| **type_integer.go** | `int`, `int8`, `int16`, `int32`, `int64`, `uint*` | `z.number().int()` |
| **type_float.go** | `float32`, `float64` | `z.number()` |
| **type_bigint.go** | `*big.Int` | `z.bigint()` |
| **type_complex.go** | `complex64`, `complex128` | No direct equivalent |

### Composite Types

| File | Go Type | TypeScript Equivalent |
|------|---------|----------------------|
| **type_slice.go** | `[]T` | `z.array()` |
| **type_array.go** | `[N]T` | `z.tuple()` |
| **type_map.go** | `map[K]V` | `z.map()` |
| **type_record.go** | `map[string]T` | `z.record()` |
| **type_object.go** | `map[string]interface{}` | `z.object()` |
| **type_struct.go** | Go structs | `z.object()` struct mapping |

### Special Types

| File | Purpose | TypeScript Equivalent |
|------|---------|----------------------|
| **type_enum.go** | Enumeration types | `z.enum()` |
| **type_literal.go** | Literal types | `z.literal()` |
| **type_union.go** | Union types | `z.union()` |
| **type_any.go** | Any type | `z.any()` |
| **type_never.go** | Never type | `z.never()` |
| **type_lazy.go** | Lazy pattern | `z.lazy()` |
| **type_function.go** | Function types | `z.function()` |

### Specialized Validation Types

| File | Specialized Domain | Description |
|------|-------------------|-------------|
| **type_network.go** | Network type validation | IP, MAC, port and other network-related validation |
| **type_file.go** | File type validation | File operations, MIME type validation |
| **type_iso.go** | ISO standard validation | ISO dates, country codes, currency codes |

## 📂 Support Directories and Files

### Internationalization Support

```
locales/
├── locales.go          # Localization framework
├── en.go              # English error messages
└── zh_cn.go           # Chinese error messages
```

### Regular Expression Library

```
regexes/
├── regexes.go         # Pre-compiled regular expressions
└── patterns.go        # Common pattern definitions
```

### Reference and Documentation

```
.reference/
└── zod/               # TypeScript Zod v4 API reference (Git submodule)

examples/              # Usage examples
docs/                  # Additional documentation
```

## 🔄 Dependency Relationship Diagram

### Core Dependency Chain

```
type.go (ZodType interface)
    ↓
parse.go (parse logic)
    ↓
checks.go (validation system)
    ↓
type_*.go (concrete type implementations)
    ↓
utils.go (utility functions)
```

### Error Handling Flow

```
type_*.go → checks.go → issue.go → error.go
```

### Wrapper Dependencies

```
type_*.go → type_default.go
          → type_prefault.go
          → type_optional.go
          → type_nilable.go
```

## 📋 File Naming Conventions

### Type File Naming

- **Basic types**: `type_<typename>.go` (e.g., `type_string.go`)
- **Wrapper types**: `type_<wrapper>.go` (e.g., `type_default.go`)
- **Special functionality**: `type_<feature>.go` (e.g., `type_network.go`)

### Core File Naming

- **Interface definitions**: `type.go`
- **Functional modules**: `<feature>.go` (e.g., `parse.go`, `checks.go`)
- **Utility functions**: `utils.go`
- **Configuration management**: `config.go`

## 🎯 File Responsibility Boundaries

### Strictly Separated Responsibilities

1. **type.go** - Only defines interfaces and basic types, no concrete implementations
2. **parse.go** - Only handles parse logic, no type-specific validation
3. **checks.go** - Only handles validation checks, no type definitions
4. **type_*.go** - Only implements specific types, no general logic
5. **utils.go** - Only provides utility functions, no business logic

### Shared Components

- **issue.go** and **error.go** - Shared by all type files
- **utils.go** - Provides cross-type general functionality
- **config.go** - Provides global configuration support

## 🏆 Architectural Advantages

1. **Modular Design** - Clear separation of concerns
2. **Type Safety** - Compile-time type checking
3. **Extensibility** - Easy to add new types and validation rules
4. **Performance** - Optimized for minimal allocations and fast execution
5. **Maintainability** - Well-organized code structure with clear dependencies

## 🔧 Development Guidelines

### Adding New Schema Types

1. **Create type file**: `type_newtype.go`
2. **Define type structures**: Def, Internals, main type
3. **Implement core interfaces**: ZodType interface methods
4. **Add validation methods**: Type-specific validation functions
5. **Implement modifiers**: Optional, Nilable, Default support
6. **Write comprehensive tests**: Unit and integration tests
7. **Add documentation**: GoDoc comments and examples

### Modifying Existing Types

1. **Maintain backward compatibility**: Don't break existing APIs
2. **Follow established patterns**: Use consistent design patterns
3. **Update documentation**: Keep docs synchronized with changes
4. **Add tests**: Ensure new functionality is properly tested
5. **Performance consideration**: Profile changes that might affect performance

This guide serves as the definitive reference for understanding and working with the GoZod codebase structure. 

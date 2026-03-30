# Documentation Templates

This document provides complete templates for DESIGN.md and CLAUDE.md with real-world examples.

## Table of Contents

1. [DESIGN.md Template](#designmd-template)
2. [CLAUDE.md Template](#claudemd-template)
3. [Real-World Examples](#real-world-examples)

---

## DESIGN.md Template

DESIGN.md captures the architectural decisions, design philosophy, and implementation patterns for a library. It's the "why" behind the code.

### Template Structure

```markdown
# [Library Name] Design

[One-paragraph overview of what the library does and its key characteristics]

## Design Principles

### KISS (Keep It Simple, Stupid)
- [Specific KISS statement for this library]
- [Example of simplicity choice]

### DRY (Don't Repeat Yourself)
- [Specific DRY statement for this library]
- [Example of code reuse]

### YAGNI (You Aren't Gonna Need It)
- [Specific YAGNI statement for this library]
- [Example of rejected feature]

### [Other Principles]
- Single Responsibility
- Open/Closed
- Interface Segregation
- [Library-specific principles]

## Architecture

### Core Components

```
[ASCII diagram of package structure]
```

[Description of each component and its responsibility]

### Data Flow

```
[ASCII diagram of data flow through the system]
```

[Description of how data moves through the system]

### Key Types and Interfaces

```go
// [Core interface or type]
type [Name] interface {
    [Method](ctx context.Context, ...) ([Return], error)
}
```

[Description of design decisions for this type]

## Core Interfaces

### [Interface Name]

```go
type [Interface] interface {
    [Methods]
}
```

**Specifications:**
- [Specification 1]
- [Specification 2]
- [Error semantics]
- [Concurrency guarantees]

## [Implementation Patterns]

### [Pattern Name]

```go
[Code example]
```

**Benefits:**
- [Benefit 1]
- [Benefit 2]

**Trade-offs:**
- [Trade-off 1]
- [Trade-off 2]

## Rejected Patterns

The following patterns are explicitly rejected to maintain simplicity:

| Pattern | Why Rejected |
|---------|--------------|
| [Pattern 1] | [Reason] |
| [Pattern 2] | [Reason] |

## Configuration

```go
type Config struct {
    [Fields]
}

type Option func(*Config)

func With[Option]([params]) Option
```

**Features:**
- [Feature 1]
- [Feature 2]

## Advanced Features

### [Feature Name]

```go
[Code example]
```

[Description and use cases]

## Concurrency and Thread Safety

- [Component 1]: [Concurrency model]
- [Component 2]: [Concurrency model]
- [Context handling]
- [Cancellation support]

## Error Handling and Observability

**Error Context:**
- [Error handling strategy]
- [Error wrapping approach]

**Structured Logging:**
- [Logging approach if applicable]

**Metrics:**
- [Metrics approach if applicable]

**Resource Management:**
- [Cleanup strategy]
- [Idempotent operations]

## Usage Examples

### Basic Usage

```go
[Simple example]
```

### Advanced Usage

```go
[Complex example]
```

### Integration Example

```go
[Real-world integration example]
```

## Performance Optimization

### [Optimization Strategy]

[Description and benchmarks]

### Performance Targets

| Scenario | Target |
|----------|--------|
| [Scenario 1] | [Target] |
| [Scenario 2] | [Target] |

## Design Goals and Boundaries

### What We Do
- [Goal 1]
- [Goal 2]

### What We Don't Do (YAGNI)

| Feature | Why Not |
|---------|---------|
| [Feature 1] | [Reason] |
| [Feature 2] | [Reason] |
```

---

## CLAUDE.md Template

CLAUDE.md is the implementation guide for AI coding assistants. It's the "how" for writing code.

### Template Structure

```markdown
# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Module**: `[module path]`
**Go Version**: [version]
**Purpose**: [One-sentence description]

[Key characteristics bullet list]

## Commands

This package uses [Task](https://taskfile.dev/) for build automation.

```bash
# View all available tasks
task --list

# Common tasks
task test          # Run all tests with race detection
task lint          # Run all linters (golangci-lint + tidy-lint)
task verify        # Run all verification steps (deps, fmt, vet, lint, build, test)
task clean         # Clean build artifacts and caches
task deps          # Download and tidy Go module dependencies
task build         # Build all packages
task cover         # Run tests with coverage
```

**Before committing, always run:**
```bash
task verify
```

## Architecture

### [Package Structure]

```
[package]/
├── [file1.go]      # [Description]
├── [file2.go]      # [Description]
└── [subpackage]/   # [Description]
```

### [Dependency Rules]

- [Rule 1]
- [Rule 2]

### [Data Flow]

```
[ASCII diagram]
```

## Key Types and Interfaces

**[Type Name]**
```go
type [Name] struct {
    [Fields]
}
```

[Description]

## Coding Rules

### Must Follow

- Go [version], using [features]
- All public functions accept `context.Context` as first parameter
- [Rule 1]
- [Rule 2]
- [Library-specific rules]

### Forbidden

- No `any` for [specific types]
- No `reflect` [where]
- No `panic` (all errors returned via error)
- No heap allocations in [hot paths]
- [Library-specific forbidden patterns]

### Go 1.20-1.26 Features

**Recommended:**
- Use `errors.Join()` for error aggregation
- Use `slices` package for slice operations
- Use `maps` package for map operations
- Use `for range N` for integer loops
- Use `testing.B.Loop()` in benchmarks

**Avoid:**
- No `sync.Pool` without profiling evidence
- No custom iterators unless truly needed
- No `log/slog` in library packages

### Testing

- Use `github.com/stretchr/testify` (`assert`/`require`)
- [Test file naming conventions]
- Use `t.Parallel()` in all tests unless they modify shared state
- Use `t.Context()` (Go 1.24) — never `context.Background()` in tests
- Use `b.Loop()` (Go 1.24) — never `for i := 0; i < b.N; i++` in benchmarks

### Package Organization

**Import order** (enforced by goimports):
1. Standard library
2. External dependencies
3. Internal packages ([module path]/...)

**Naming conventions:**
- Interfaces: noun ([examples])
- Implementations: noun + adjective ([examples])
- Errors: Err prefix ([examples])
- Options: With prefix ([examples])
- Receiver names: one or two letter abbreviation, consistent across methods
- Acronyms: consistent casing (URL/url, ID/id, never Url or Id)
- Constants: MixedCaps only (no ALL_CAPS or k-prefix)

**Documentation:**
- All exported names must have doc comments starting with the symbol name
- Package-private types and functions should have doc comments explaining their purpose
- Compile-time interface checks should have explanatory comments

## Testing

### Coverage Thresholds

Enforced by `task cover`:
- [Package 1]: [threshold]%
- [Package 2]: [threshold]%

### Test Types

**Unit tests** — Test individual components in isolation:
```bash
task test
```

**Integration tests** — Test full pipeline:
```bash
task test-integration
```

**Benchmark tests** — Performance verification:
```bash
go test -bench=. -benchmem ./...
```

## Dependencies

### Production Dependencies

| Dependency | Purpose |
|------------|---------|
| [dep1] | [purpose] |
| [dep2] | [purpose] |

### Test Dependencies

| Dependency | Purpose |
|------------|---------|
| [dep1] | [purpose] |

### Dependency Notes

- [Note 1]
- [Note 2]

## Error Handling

**Sentinel errors:**
```go
var (
    [ErrName] = errors.New("[message]")
)
```

**Error wrapping:**
```go
return fmt.Errorf("[context]: %w", err)
```

**Error checking:**
```go
if errors.Is(err, [ErrName]) { ... }
```

## Performance

### Optimization Strategies

- [Strategy 1]
- [Strategy 2]

### Benchmarking

```bash
go test -bench=. -benchmem ./...
```

## Design Philosophy

### KISS, DRY, YAGNI

- [Library-specific statements]

### [Other Principles]

- [Principle statements]

## Before Committing

1. Read `docs/` folder to ensure changes align with design decisions
2. Run `task verify` (deps → fmt → vet → lint → build → test)
3. Ensure all tests pass with race detection
4. Check that coverage thresholds are met
5. Verify no linter warnings

## Agent Skills

This repository includes agent skills in `.claude/skills/` (symlinked to `.agents/skills/`):

- **agent-md-creating** — Generate CLAUDE.md for Go projects
- **code-simplifying** — Refine and simplify recently written Go code
- **committing** — Create conventional commits for Go packages
- **dependency-selecting** — Select Go dependencies from vetted libraries
- **go-best-practices** — Google Go coding best practices and style guide
- **golang-taskfile** — Create and manage Taskfiles for Go projects
- **linting** — Set up and run golangci-lint v2
- **modernizing** — Go code modernization guide (Go 1.20-1.26 features)
- **ralphy-initializing** — Initialize Ralphy AI coding loop configuration
- **ralphy-todo-creating** — Create Ralphy TODO.yaml task files
- **readme-creating** — Generate README.md for Go libraries
- **releasing** — Guide release process for Go packages
- **testing** — Write Go tests following best practices

Use these skills when relevant to your task.
```

---

## Real-World Examples

### Example 1: go-fsm DESIGN.md

```markdown
# Go FSM Library

A Go generics-based state machine library, guided by KISS/DRY/YAGNI.

## Design Philosophy

Three iron rules that drive every decision:

**KISS — Each concept has exactly one representation.** The core of a state machine is states + events + transitions. No need for five kinds of trigger behaviour, no need to distinguish transitioning / reentry / internal / ignored / dynamic. Keep the implementation lean, achieving nanosecond-level zero-allocation performance.

**DRY — Rules are declared once; runtime, validation, and visualization share the same data.** The transition table serves `Fire()` dispatch, `Build()` validation, and `Mermaid()`/`ASCII()` export simultaneously. No separate graph structure is maintained for visualization.

**YAGNI — Only implement what's truly needed.** No sub-states (orders don't need them). No parallel states (that's a workflow engine's job). No Actor model. No JSON declarative config. Add things when needed — the cost of adding later is far lower than maintaining unused features.

## Architecture

### Two-Phase Separation: Build vs Runtime

- **Build phase** (Builder): register rules, hooks, validation — mutable
- **Runtime** (Machine): immutable after `Build()`, concurrency-safe, zero-allocation hot path
- Two declaration styles coexist: Builder method chaining (Style A) and `[]Rule` transition table (Style B), sharing the same internal builder

Core value of this separation:

- `Build()` performs one-time validation: initial state exists, final states have no outgoing edges, guard mutual exclusivity check
- Runtime `Machine` is immutable, inherently concurrency-safe (concurrent state storage is the `StateStore`'s responsibility)
- Lookup tables are pre-allocated at `Build()` time, zero allocation at runtime

### Generic Parameters

```go
type Machine[S comparable, E comparable] struct { ... }
```

- S = state type, E = event type, both constrained to `comparable`
- No `any`, no interface boxing — every `Fire()` call avoids boxing and type assertion overhead
- Users are recommended to use `int` iota enums to define states and events

## Rejected Patterns

| Feature | Why Not |
|---------|---------|
| Sub-states / hierarchical states | Order scenarios don't need them, adds recursive Enter/Exit logic |
| Parallel states | Workflow engine's responsibility, not FSM's |
| Async transitions | Two-phase commit adds massive complexity, users handle async in callbacks |
| Dynamic targets | Breaks declarative predictability, use Guard + multiple rules instead |

## Performance Targets

| Scenario | Target |
|----------|--------|
| Pure transition (no guards, no callbacks) | < 60 ns/op, 0 allocs |
| With 1 guard | < 100 ns/op, 0 allocs |
| With Entry callback | < 120 ns/op, 0 allocs |
```

### Example 2: go-cache DESIGN.md

```markdown
# Cache Library Design

Framework-agnostic, store-driven cache library for Go that unifies local and cloud cache stores behind a single generic interface.

## Design Principles

- **KISS** — Simple API (5-10 core methods), minimal abstractions. Three similar lines beat a single-use helper.
- **DRY** — Store interface shared across all backends, codec abstraction eliminates serialization duplication
- **YAGNI** — No speculative features, implement only what's currently needed. No premature optimization.
- **Single Responsibility** — Store handles storage, Codec handles serialization, Cache handles type safety
- **Open/Closed** — Extend via new Store implementations and Codec types, not modification
- **Interface Segregation** — Small focused interfaces (Store: 5 methods, Codec: 2 methods, Cache[T]: 5 methods)

## Rejected Patterns

The following patterns are explicitly rejected to maintain simplicity:

- **Middleware/Hooks System** — Adds complexity without clear use case. Users can wrap Cache interface if needed.
- **Compression/Encryption Codecs** — Users can implement custom Codec interface. No need for built-in wrappers.
- **Adaptive Cleanup Frequency** — Premature optimization. Fixed intervals with configuration are sufficient.
- **Batch Size Enforcement** — Documentation is sufficient. Hard limits add complexity and may not fit all use cases.
- **Reflection in Hot Paths** — Use generics for type safety and performance.

## Core Interfaces

### Store Interface (Low-Level, Byte-Oriented)

```go
type Store interface {
    Get(ctx context.Context, key string) ([]byte, error)
    Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
    Delete(ctx context.Context, key string) error
    Clear(ctx context.Context) error
    Close() error
}
```

**Specifications**:
- Byte-oriented for maximum flexibility
- Context support for cancellation and timeouts
- Thread-safe for concurrent access
- TTL semantics: `TTL > 0` expires, `TTL = 0` uses default, `TTL < 0` never expires
- Returns `ErrNotFound` for missing keys
- Returns `ErrClosed` after Close() is called
```

### Example 3: go-config CLAUDE.md

```markdown
# Go Config Library Design

A type-safe configuration library for Go 1.26+ with a minimal `Source + Pipeline` API.

## Commands

```bash
task test       # Run root module tests with race detector
task lint       # Run golangci-lint v2 + go mod tidy check (root)
task test:all   # Run tests for root + all sub-modules
task lint:all   # Run linters for root + all sub-modules
```

## Architecture

### Multi-Module Structure

This repository uses two Go modules to isolate heavy dependencies:

| Module | Path | Heavy Deps |
|--------|------|------------|
| Core | `github.com/agentable/go-config` | stdlib only |
| Formats | `github.com/agentable/go-config/format/*` | yaml/toml parsers |

## Coding Rules

### Must Follow

- Go 1.26 — use modern language features where they simplify code
- Follow [Google Go Best Practices](https://google.github.io/go-style/best-practices)
- KISS/DRY/YAGNI — no premature abstractions, no unused features
- All public functions accept `context.Context` as first parameter
- `Value()` must be lock-free and allocation-free
- Functional options return `error` (not void)
- Nil guards: `Load()` returns error for nil sources
- Sentinel errors for expected failures; callers use `errors.Is`

### Forbidden

- No plaintext in logs
- No `panic` in production code
- No framework dependencies in core packages
- No global mutable state
- No Watch/real-time push (configs change infrequently)

### Go 1.20-1.26 Features

| Feature | Where Used |
|---------|-----------|
| `iter.Seq2` (1.23) | `Source.List` returns lazy iterator |
| `clear()` (1.21) | Map clearing in merge operations |
| `errors.Join` (1.20) | Aggregating source errors |
| `atomic.Pointer[T]` (1.19) | Lock-free config snapshots |
```

---

## Documentation Checklist

### DESIGN.md Checklist

- [ ] One-paragraph overview at the top
- [ ] Explicit KISS/DRY/YAGNI statements with examples
- [ ] ASCII diagrams for architecture and data flow
- [ ] Core interfaces with specifications
- [ ] Rejected patterns section with rationale
- [ ] Usage examples (basic, advanced, integration)
- [ ] Performance targets if applicable
- [ ] Design goals and boundaries (what we do/don't do)

### CLAUDE.md Checklist

- [ ] Project overview with module path and Go version
- [ ] Commands section with task runner usage
- [ ] Architecture section with package structure
- [ ] Coding rules (must follow, forbidden, Go features)
- [ ] Testing section with coverage thresholds
- [ ] Dependencies section with rationale
- [ ] Error handling patterns
- [ ] Performance optimization strategies
- [ ] Before committing checklist
- [ ] Agent skills list

---

## Tips for Writing Documentation

### DESIGN.md Tips

1. **Start with "why"** - Explain design decisions, not just what the code does
2. **Be explicit about trade-offs** - Document what you chose NOT to do
3. **Use real code examples** - Show actual API usage, not pseudocode
4. **Document rejected patterns** - Prevent future contributors from repeating mistakes
5. **Keep it updated** - Update DESIGN.md when making architectural changes

### CLAUDE.md Tips

1. **Start with commands** - Make it easy to run tests and verify changes
2. **Be prescriptive** - Tell AI exactly what to do and what not to do
3. **Include examples** - Show correct patterns for common tasks
4. **Document conventions** - Naming, error handling, testing patterns
5. **Link to skills** - Reference available agent skills for common tasks

### Common Mistakes to Avoid

1. **Too generic** - "Write clean code" is not helpful
2. **Too verbose** - Keep it concise, focus on non-obvious information
3. **Outdated examples** - Ensure code examples actually work
4. **Missing rationale** - Explain WHY, not just WHAT
5. **No rejected patterns** - Document what you chose NOT to do

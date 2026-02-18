# CLAUDE.md Template for Go Projects

This template defines the structure for generating CLAUDE.md files. Include sections based on the rules in SKILL.md. Adapt section depth and detail to project complexity.

---

## Development Philosophy Defaults

Every generated CLAUDE.md must embed these principles in the Design Philosophy and Coding Rules sections. These are the baseline — project-specific analysis may add to them, but never contradict them.

### Core Design Principles

Apply these principles when writing the **Design Philosophy** section. Select the ones most relevant to the project, and add a project-specific one-sentence explanation for each:

| Principle | Guideline for CLAUDE.md |
|-----------|------------------------|
| **KISS** | Each concept has exactly one representation. Prefer the simplest solution that works. No premature abstractions. |
| **DRY** | Rules declared once, shared across runtime/validation/visualization. No copy-paste of logic. |
| **YAGNI** | Only implement what's currently needed. The cost of adding later is lower than maintaining unused features. |
| **Single Responsibility (SRP)** | Each type/package/file has one reason to change. Split when responsibilities diverge. |
| **Open-Closed (OCP)** | Extend behavior through interfaces, functional options, or plugins — not by modifying existing code. |
| **Liskov Substitution (LSP)** | Interface implementations must be fully substitutable. No surprises when swapping concrete types. |
| **Interface Segregation (ISP)** | Small, focused interfaces (1-3 methods). Consumers depend only on what they use. |
| **Dependency Inversion (DIP)** | Core packages depend on abstractions (interfaces), not concretions. Inject dependencies via constructors or functional options. |

**Usage rule:** Include 2-5 principles in the generated CLAUDE.md. Pick the ones the project actually demonstrates. Write project-specific explanations, not generic definitions.

**Example output for a FSM library:**
```markdown
## Design Philosophy

- **KISS** — Each concept has exactly one representation. No sub-states, no parallel states, no five kinds of trigger behaviour.
- **DRY** — The transition table serves Fire() dispatch, Build() validation, and Mermaid()/ASCII() export simultaneously.
- **YAGNI** — No Actor model, no JSON config, no Graphviz. Add when needed — the cost of adding later is far lower than maintaining unused features.
- **OCP** — Extend via StateStore interface and Guard/Action callbacks, not by modifying Machine internals.
- **ISP** — StateStore has exactly 2 methods (Get, Set). No bloated storage interface.
```

### Go 1.26+ Features and Idioms

When generating Coding Rules, incorporate applicable Go 1.26+ features detected in the project source. Use this reference to identify and document them:

#### Language Features

| Feature | Since | Usage Pattern | Detect By |
|---------|-------|---------------|-----------|
| **Type parameter constraints** | 1.18 | `[S comparable]`, `[T any]` | Generic type declarations |
| **Self-referential generic constraints** | 1.26 | `type Describable[S Describable[S]]` | Constraint referencing its own type parameter |
| **`new(expr)`** | 1.26 | `new(myFunc())` instead of `tmp := myFunc(); &tmp` | Pointer creation from expression |
| **Range over integer** | 1.22 | `for range N` instead of `for i := 0; i < N; i++` | `for range` with integer operand |
| **Range over function** | 1.23 | Iterator functions `func(yield func(T) bool)` | `for v := range iterFunc` |
| **`slices` package** | 1.21 | `slices.Clone`, `slices.Sort`, `slices.Contains` | `slices.` calls |
| **`maps` package** | 1.21 | `maps.Keys`, `maps.Values`, `maps.Clone` | `maps.` calls |
| **`clear()` builtin** | 1.21 | `clear(myMap)`, `clear(mySlice)` | `clear(` calls |
| **`log/slog`** | 1.21 | Structured logging | `slog.` imports |
| **`iter.Seq` / `iter.Seq2`** | 1.23 | `func() iter.Seq2[K, V]` | `iter.Seq` in signatures |

#### Testing Features

| Feature | Since | Usage Pattern | Detect By |
|---------|-------|---------------|-----------|
| **`testing.B.Loop()`** | 1.24 | `for b.Loop() { }` instead of `for i := 0; i < b.N; i++` — does not prevent inlining | `b.Loop()` in benchmarks |
| **`t.Context()`** | 1.24 | `t.Context()` instead of `context.Background()` in tests | `t.Context()` calls |
| **`testing/synctest`** | 1.25 | Virtualized time for concurrent tests | `synctest.` imports |

#### Runtime and Toolchain

| Feature | Since | Impact |
|---------|-------|--------|
| **Green Tea GC** | 1.26 | Reduced GC pause, benefits small struct-heavy code |
| **Compiler slice stack allocation** | 1.26 | Small slices may avoid heap allocation |
| **`errors.AsType[T]()`** | 1.26 | Type-safe error extraction without manual type assertion |
| **`crypto/rand.Text`** | 1.24 | Random string generation without manual base64 |
| **`os.Root`** | 1.25 | Sandboxed file access (chroot-like) |
| **`sync.OnceValue[T]`** | 1.21 | Lazy initialization |

**Usage rule:** Scan the project source for these patterns. List only the features actually used in the Coding Rules section. Group under a "Go {version} Features" subsection if 3+ features are present.

**Example output:**
```markdown
### Go 1.26 Features

| Feature | Where Used |
|---------|-----------|
| `testing.B.Loop()` | All benchmarks — does not prevent inlining |
| `slices.Clone` | `Clone()` method for entry copying |
| `clear()` | `UnmarshalJSONFrom` for GC of old references |
| `for range N` | Iteration patterns throughout |
| `new(expr)` | Pointer creation from expressions |
```

### Standard Coding Rules Baseline

Always include these in the **Coding Rules > Must Follow** section (adapt wording to project):

```markdown
### Must Follow

- Go {version} — use modern language features where they simplify code
- Follow Google Go Best Practices: https://google.github.io/go-style/best-practices
- Follow Google Go Style Decisions: https://google.github.io/go-style/decisions
- KISS/DRY/YAGNI — no premature abstractions, no unused features, no duplicated logic
- Small interfaces — 1-3 methods per interface, consumers depend only on what they use
- Extend via interfaces and composition, not by modifying existing types
- Explicit error handling — return errors, wrap with context via `fmt.Errorf("%w")`
```

Then add project-specific rules detected from source code analysis (zero-alloc requirements, context.Context conventions, no-panic policy, generic constraints, etc.).

### Standard Forbidden Rules Baseline

Always include these in the **Coding Rules > Forbidden** section (adapt to project):

```markdown
### Forbidden

- No `panic` in production code (all errors returned via `error`)
- No premature abstraction — three similar lines are better than a helper used once
- No feature creep — only implement what's currently needed
- No god objects — split types that have more than one reason to change
```

Then add project-specific forbidden rules (no reflect, no framework deps in core, no heap allocs in hot path, etc.).

### Content Exclusion Rules

**Never include in CLAUDE.md:**

- Implementation phases, migration plans, or roadmaps (use DESIGN.md or PLAN.md)
- TODO lists, task tracking, or work-in-progress notes (use issue trackers)
- Step-by-step build/setup instructions for new contributors (use README.md)
- Historical context ("we used to do X, now we do Y") unless it prevents a common mistake
- Aspirational rules not yet enforced ("we should eventually...")

CLAUDE.md describes **what the project is now** and **how to work with it correctly**. It is not a planning document.

### Agent Skills Index

**Always include** this section in every generated CLAUDE.md. It indexes the shared skill library at `.agents/skills/` so AI agents know what specialized capabilities are available.

The skills directory is located at `.agents/skills/` in the project root (typically a symlink to the shared skills repository).

**Always-include skills table:**

```markdown
## Agent Skills

Specialized skills available in `.agents/skills/`:

| Skill | When to Use |
|-------|------------|
| [testing](.agents/skills/testing/) | Writing or reviewing Go tests — testify patterns, table-driven tests, mocking, concurrency testing, benchmarks |
| [linting](.agents/skills/linting/) | Setting up or running golangci-lint v2, fixing lint errors, configuring linters |
| [modernizing](.agents/skills/modernizing/) | Adopting Go 1.20-1.26 new features — generics, iterators, error handling, stdlib collections |
| [committing](.agents/skills/committing/) | Creating conventional commit messages for Go packages |
| [releasing](.agents/skills/releasing/) | Releasing a Go package — semantic versioning, tagging, dependency upgrades |
| [code-simplifying](.agents/skills/code-simplifying/) | Refining recently written Go code for clarity and consistency without changing functionality |
| [go-best-practices](.agents/skills/go-best-practices/) | Applying Google Go style guide — naming, error handling, interfaces, testing, concurrency |
```

**Usage rules:**
- Always include this section in every generated CLAUDE.md
- Only list skills that are actually present in `.agents/skills/` (verify by checking the directory)
- Use relative paths from project root
- Keep descriptions to one sentence focusing on the **trigger scenario** ("When to Use")

---

## Template

```markdown
# {Project Name}

{One to three sentence description: what it is, what problem it solves, key characteristic (e.g., zero-dependency, generics-based, plugin-driven).}

{If there is a reference implementation, note it here:}
- **Reference implementation:** {name} ({language}) — {relationship, e.g., "output compatibility target", "architecture equivalence required"}

## Commands

{Detect from Makefile. If no Makefile, use Go defaults. Only list commands that actually exist.}

```bash
task test          # Run all tests with race detection
task lint          # Run golangci-lint + go mod tidy check
make fmt           # Format code
make vet           # Run go vet
task verify        # Full verification: deps, fmt, vet, lint, test
make bench         # Run benchmarks
task deps          # Download and tidy dependencies
task clean         # Clean build artifacts
```

{If no Makefile:}

```bash
go test -race ./...       # Run all tests
go vet ./...              # Static analysis
go build ./...            # Build all packages
golangci-lint run ./...   # Lint (if configured)
```

## Architecture

{Show directory structure. Include only key packages — not every file.}

```
{module-name}/
├── {pkg1}/          # {one-line description}
├── {pkg2}/          # {one-line description}
├── internal/        # {private packages}
│   ├── {sub}/       # {description}
│   └── {sub}/       # {description}
└── cmd/             # {CLI entry points, if any}
```

{For single-file libraries, describe the core type instead of directory tree.}

### Key Types

{List the main exported types and their relationships. Use a code block for type signatures if they show project-specific patterns (generics, interface hierarchies, functional options).}

```go
// Core type and constructor
type {MainType} struct { ... }
func New(opts ...Option) *{MainType}

// Key interface (if any)
type {Interface} interface {
    {Method signatures}
}
```

### Design Patterns

{Only list patterns actually used in the project. Common ones for Go projects:}

- **{Pattern Name}** — {where it's used and why}

{Examples: Functional Options, Builder, Strategy, Pipeline/AST-centric, Two-Phase (Build/Runtime), Plugin Architecture, Registry Pattern}

## Design Philosophy

{Include only if the project has explicit design principles that affect implementation decisions. Skip for straightforward CRUD/utility projects.}

- **{Principle}** — {one sentence explaining how it constrains implementation}

## Coding Rules

### Must Follow

- Go {version from go.mod}
- {Project-specific rules only — not standard Go practices}
- {Examples: "All public functions accept context.Context as first parameter", "Zero heap allocations in hot path", "All APIs return errors instead of panicking"}

### Forbidden

- {Things explicitly avoided in this project}
- {Examples: "No reflect", "No panic in production code", "No framework dependencies in core"}

## Testing

{Describe the testing approach and how to run specific tests.}

- {Testing framework: testify, stdlib, or other}
- {Test patterns: t.Parallel(), golden tests, table-driven, property-based}
- {Benchmark conventions: b.Loop() vs b.N, benchstat}

```bash
# Run specific package tests
go test -race ./pkg/{package}/

# Run specific test function
go test -race -run {TestName} ./{package}/

# Run benchmarks
go test -bench=. -benchmem ./{package}/
```

## Dependencies

{Only if there are non-test external dependencies. List with purpose.}

| Dependency | Purpose |
|------------|---------|
| `{module}` | {one-line description} |

## Error Handling

{Only if the project has specific error patterns beyond standard Go.}

- {Examples: "Sentinel errors: ErrNotFound, ErrDenied, ...", "All errors wrapped with fmt.Errorf(\"%w\")", "ParseError type with position info"}

## Performance

{Only if the project has benchmarks or performance requirements.}

| Scenario | Target |
|----------|--------|
| {operation} | {target, e.g., "< 60 ns/op, 0 allocs"} |

## Linting

{Only if non-default lint configuration.}

golangci-lint {version}. Config in `.golangci.yml`.

## CI

{Only if CI is configured.}

{Brief description of CI pipeline and triggers.}

## Agent Skills

{Always include. Verify which skills exist in .agents/skills/ and list only those present.}

Specialized skills available in `.agents/skills/`:

| Skill | When to Use |
|-------|------------|
| [testing](.agents/skills/testing/) | Writing or reviewing Go tests — testify patterns, mocking, concurrency testing, benchmarks |
| [linting](.agents/skills/linting/) | Setting up golangci-lint v2, fixing lint errors, configuring linters |
| [modernizing](.agents/skills/modernizing/) | Adopting Go 1.20-1.26 features — generics, iterators, error handling, stdlib collections |
| [committing](.agents/skills/committing/) | Creating conventional commit messages for Go packages |
| [releasing](.agents/skills/releasing/) | Releasing a Go package — semantic versioning, tagging, dependency upgrades |
| [code-simplifying](.agents/skills/code-simplifying/) | Refining recently written Go code for clarity and consistency |
| [go-best-practices](.agents/skills/go-best-practices/) | Applying Google Go style guide — naming, error handling, interfaces, concurrency |
```

---

## Section Sizing Guide

| Project Complexity | Target CLAUDE.md Lines | Approach |
|--------------------|----------------------|----------|
| Single-file library | 30-60 lines | Overview, Commands, Architecture (brief), Testing |
| Small library (2-5 packages) | 60-120 lines | Add Coding Rules, Dependencies, Design Patterns |
| Medium library (5-15 packages) | 120-200 lines | Add Error Handling, Performance, detailed Architecture |
| Large project (15+ packages) | 150-250 lines + `.claude/` | Progressive disclosure: link to detail files |

## Adaptation Examples

### Minimal (single-file library like orderedobject)

```markdown
# orderedobject

Generic ordered JSON object for Go preserving insertion order. Uses go-json-experiment/json v2 streaming API.

## Commands

\```bash
task test    # Run all tests
task lint    # Run golangci-lint and go mod tidy check
task verify  # Run all: deps, fmt, vet, lint, test
\```

## Architecture

Single-file library (`object.go`) with one core type:
- `Object[V any]` — ordered key-value store backed by `[]Entry[V]`
- JSON marshalling uses streaming API for zero-intermediate-allocation output

## Key Design Decisions

- Linear scan key lookup (no map index) — simple and correct for typical JSON sizes
- Shallow clone — `Clone()` copies entries slice, not values
- Duplicate key rejection via go-json-experiment default behavior

## Testing

- All tests use `t.Parallel()`, stdlib assertions only
- Benchmarks use `b.Loop()` (Go 1.24+)
- Run with race detector: `go test -race ./...`
```

### Comprehensive (multi-package library)

Should include all sections, with detailed Architecture showing two-phase design, callback order, error classification, and API type chain. Code examples justified by complex generic patterns.

### Reference-driven (like jsoncrdt with json-joy reference)

Add dedicated section for reference implementation requirements: file organization rules, mandatory analysis process, TypeScript-to-Go mapping patterns.

---
description: Use when writing, reviewing, or refactoring Go code for idiomatic API design, naming, error handling, testing structure, testify usage boundaries, concurrency patterns, formatting, and documentation. Pair with modernizing when adopting Go 1.20+ stdlib and language improvements.
name: go-best-practices
---


# Google Go Best Practices

Comprehensive Go coding guide based on Google's internal style guide, tuned for modern Go codebases. Use it for durable judgment first: API shape, naming, errors, tests, readability, and maintainability. Then layer in version-gated modernization when newer Go features make the code simpler.

## When to Apply

Reference these guidelines when:
- Writing new Go packages, functions, or methods
- Reviewing Go code for style and correctness
- Refactoring existing Go code
- Designing Go APIs (interfaces, option patterns, error types)
- Writing or improving Go tests

## Relationship to `modernizing`

Use this skill for durable rules that stay true across Go versions:
- API shape and package boundaries
- naming and documentation
- error contracts
- test readability and maintainability
- concurrency ownership and lifecycle

Use `modernizing` for version-gated upgrades:
- `errors.Join`, `errors.AsType`, `context.WithCancelCause`
- `sync.WaitGroup.Go`, `sync.OnceValue`
- `b.Loop`, `t.Context`, `testing/synctest`
- `slices`, `maps`, `cmp`, `min`, `max`, `clear`

When both skills apply:
1. Keep the durable rule from this skill.
2. Prefer the newer stdlib API when it reduces code and keeps intent obvious.
3. Do not modernize mechanically if it conflicts with repository convention or makes the code harder to read.

## Rule Categories by Priority

| Priority | Category | Impact | Guide |
|----------|----------|--------|-------|
| 1 | Naming | CRITICAL | [rules/naming.md](rules/naming.md) |
| 2 | Error Handling | CRITICAL | [rules/error.md](rules/error.md) |
| 3 | Design Patterns | HIGH | [rules/design.md](rules/design.md) |
| 4 | Formatting | HIGH | [rules/format.md](rules/format.md) |
| 5 | Documentation | MEDIUM | [rules/doc.md](rules/doc.md) |
| 6 | Testing | MEDIUM | [rules/testing.md](rules/testing.md) |
| 7 | Concurrency | MEDIUM | [rules/concurrency.md](rules/concurrency.md) |
| 8 | Performance | LOW-MEDIUM | [rules/perf.md](rules/perf.md) |

## Quick Reference

### 1. Naming (CRITICAL) — See [rules/naming.md](rules/naming.md)

- Avoid Redundant Naming — don't repeat package, receiver, parameter, or return type info
- Package Naming — short, lowercase, no underscores; avoid `util`, `helper`, `common`
- Receiver Naming — one or two letter abbreviation, consistent across methods
- Constant Naming — MixedCaps only; no ALL_CAPS or k-prefix; name by role not value
- Acronym Casing — consistent: `URL`/`url`, `ID`/`id`, never `Url` or `Id`
- No Get Prefix — use nouns for accessors, verbs for actions
- Variable Naming — length proportional to scope; omit type info
- Function Naming — nouns for return-value functions, verbs for actions

### 2. Error Handling (CRITICAL) — See [rules/error.md](rules/error.md)

- Structured Errors — use sentinel values or typed errors, not string matching
- Add Non-Redundant Context — meaningful wrapping without duplication
- %v vs %w — `%v` at boundaries, `%w` for programmatic inspection
- %w Position — place at end: `"context: %w"`
- Modern Error APIs — prefer `errors.Join`, `errors.AsType`, and cause-aware contexts when they simplify the code
- Return error Interface — not concrete types
- Handle Errors Explicitly — never silently discard with `_`
- Indent Error Flow — handle errors first, keep success path unindented
- Avoid In-Band Errors — use multiple returns instead of special values
- Error Logging — don't double-log; guard expensive log calls

### 3. Design Patterns (HIGH) — See [rules/design.md](rules/design.md)

- Interfaces Belong to Consumers — define in consumer, return concrete from producer
- Option Structs — for many callers needing many params
- Variadic Options — functional options when most callers need no config
- Avoid Global State — provide instance-based APIs
- Pass Values — not pointers for small fixed-size types
- Receiver Types — pointer for mutation/large; value for small immutable
- Generics — use only when genuinely needed
- Context Conventions — always first param, never in structs

### 4. Formatting (HIGH) — See [rules/format.md](rules/format.md)

- Always gofmt — use `gofmt` or `goimports`
- Import Grouping — stdlib, third-party, proto, side-effect
- Import Renaming — only for conflicts; proto uses `pb` suffix
- Struct Literal Fields — use field names; omit zero values
- Nil Slices — prefer `var t []string` over `t := []string{}`
- Function Formatting — keep signatures on one line; extract locals
- Variable Declarations — `:=` for non-zero, `var` for zero, `new()` for pointers, `new(expr)` when inline pointer creation is clearer
- Conditions — extract complex conditions; no Yoda; no redundant `break`

### 5. Documentation (MEDIUM) — See [rules/doc.md](rules/doc.md)

- Doc Comments — exported names start with symbol name as complete sentence
- Package Comments — one per package above `package` clause
- Parameter Docs — only document non-obvious parameters
- Cleanup Docs — document cleanup requirements and error sentinels
- Signal Boosting — add comments for code that looks standard but isn't

### 6. Testing (MEDIUM) — See [rules/testing.md](rules/testing.md)

- Table-Driven Tests — with named fields and descriptions
- Assertion Style — prefer `testing` and `cmp`; `testify` is acceptable when repo conventions or clarity justify it
- Got Before Want — format: `Func(%v) = %v, want %v`
- Test Helpers — call `t.Helper()`; prefix must-succeed with `must`
- Modern Test APIs — use `b.Loop`, `t.Context`, `t.Chdir`, and `testing/synctest` when Go version and test shape fit
- Scoped Setup — explicit per test; no package-level `init()`
- Error Semantics — test with `errors.Is`, not strings
- Goroutine Fatal — use `t.Error` not `t.Fatal` from goroutines

### 7. Concurrency (MEDIUM) — See [rules/concurrency.md](rules/concurrency.md)

- Goroutine Lifetimes — use WaitGroup to bound lifetimes
- Modern Sync Helpers — prefer `WaitGroup.Go`, `OnceValue`, and newer timer semantics when they remove boilerplate
- Synchronous Functions — prefer sync; callers add concurrency
- Channel Direction — specify `<-chan` or `chan<-` in signatures
- No Copy — never copy `sync.Mutex` or types with pointer methods
- No Panic — use errors for normal failures; panic only for invariants
- Variable Shadowing — watch for `:=` shadowing in inner scopes

### 8. Performance (LOW-MEDIUM) — See [rules/perf.md](rules/perf.md)

- String Concatenation — `+` for simple, `Sprintf` for format, `Builder` for loops
- Modern Stdlib Helpers — prefer `slices`, `maps`, `cmp`, `bytes.Clone`, `min`, `max`, and `clear` before hand-rolled helpers
- Size Hints — pre-allocate with justified hints only
- %q Format — use for readable string output
- crypto/rand — for keys, never `math/rand`
- Use any — instead of `interface{}` in new code

## Full Compiled Document

For the complete guide with all rules expanded: [AGENTS.md](AGENTS.md)

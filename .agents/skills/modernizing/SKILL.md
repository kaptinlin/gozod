---
name: modernizing
description: Go code modernization guide covering Go 1.20-1.26 new features. Use when writing, reviewing, or refactoring Go code to adopt modern idioms — generics, iterators, error handling, concurrency patterns, stdlib collections, HTTP routing, testing, security APIs, and tooling. Triggers on Go code optimization, modernization, or upgrade tasks.
---

# Go Modernization Guide (1.20 → 1.26)

Actionable guide for modernizing Go code with new language features and standard library APIs introduced in Go 1.20 through 1.26. Organized by functional module with clear when-to-use and when-NOT-to-use guidance.

## When to Apply

Reference these guidelines when:
- Writing new Go code (use modern idioms from the start)
- Reviewing Go code for modernization opportunities
- Refactoring existing Go code to use newer APIs
- Upgrading a project's Go version
- Running `go fix ./...` and understanding the changes it applies

## Guide Categories by Priority

| Priority | Category | Key Features | Guide |
|----------|----------|-------------|-------|
| 1 | Stdlib Collections | slices, maps, cmp, min/max/clear | [guides/stdlib-collections.md](guides/stdlib-collections.md) |
| 2 | Error Handling | errors.Join, errors.AsType, context cause | [guides/error-handling.md](guides/error-handling.md) |
| 3 | Iterators & Loops | range over int, function iterators, iter pkg | [guides/iterators-loops.md](guides/iterators-loops.md) |
| 4 | Generics | type aliases, constraints, Null[T], TypeFor | [guides/generics.md](guides/generics.md) |
| 5 | Concurrency | OnceValue, WaitGroup.Go, timer changes | [guides/concurrency.md](guides/concurrency.md) |
| 6 | HTTP & Networking | enhanced routing, ResponseController, CSRF | [guides/http-networking.md](guides/http-networking.md) |
| 7 | Testing | b.Loop, t.Context, synctest, t.Chdir | [guides/testing.md](guides/testing.md) |
| 8 | Security & Crypto | os.Root, rand.Text, HPKE, ML-KEM | [guides/security.md](guides/security.md) |
| 9 | Tooling & Build | go fix, PGO, tool directives, omitzero, slog | [guides/tooling-build.md](guides/tooling-build.md) |

## Quick Modernization Checklist

### Immediate wins (safe, mechanical replacements)

**Collections & builtins (Go 1.20-1.22)**:
- Custom `min`/`max` helpers → built-in `min`/`max` (1.21+)
- `for k := range m { delete(m, k) }` → `clear(m)` (1.21+)
- Custom `contains`/`indexOf` → `slices.Contains`/`slices.Index` (1.21+)
- `sort.Slice(s, less)` → `slices.SortFunc(s, cmp)` (1.21+)
- Manual map cloning loop → `maps.Clone(m)` (1.21+)
- `strings.HasPrefix` + `TrimPrefix` → `strings.CutPrefix` (1.20+)
- `make([]byte, n); copy(dst, src)` → `bytes.Clone(b)` (1.20+)
- Default value fallback chains → `cmp.Or(a, b, fallback)` (1.22+)
- `slices.Concat(a, b, c)` for joining slices (1.22+)

**Loops (Go 1.22+)**:
- `for i := 0; i < n; i++` → `for i := range n` (1.22+)
- Remove `v := v` workarounds in loop closures (1.22+ auto-scoping)

**Error handling (Go 1.20-1.26)**:
- Manual error list accumulation → `errors.Join(errs...)` (1.20+)
- `fmt.Errorf` supports multiple `%w` verbs (1.20+)
- `var t *T; errors.As(err, &t)` → `errors.AsType[*T](err)` (1.26+)

**Concurrency (Go 1.21-1.25)**:
- `sync.Once` + separate var → `sync.OnceValue(func() T)` (1.21+)
- `wg.Add(1); go func() { defer wg.Done() ... }()` → `wg.Go(func() { ... })` (1.25+)
- Timer/Ticker drain-after-stop boilerplate → remove (1.23+ guarantees no stale values)

**Reflection & types (Go 1.22+)**:
- `reflect.TypeOf((*T)(nil)).Elem()` → `reflect.TypeFor[T]()` (1.22+)
- `sql.NullString` / `sql.NullInt64` → `sql.Null[T]` (1.22+)

**Testing (Go 1.24+)**:
- `for i := 0; i < b.N; i++` → `for b.Loop()` (1.24+)
- `ctx, cancel := context.WithCancel(ctx); t.Cleanup(cancel)` → `t.Context()` (1.24+)
- Manual `os.Chdir` + restore → `t.Chdir("testdata")` (1.24+)

**Security & encoding (Go 1.24+)**:
- `rand.Read(b); base64.Encode(b)` → `rand.Text()` (1.24+)
- Manual AES-GCM nonce → `cipher.NewGCMWithRandomNonce` (1.24+)
- `json:"field,omitempty"` on `time.Time` → `json:"field,omitzero"` (1.24+)
- Remove error checks on `crypto/rand.Read` (always nil since 1.24)
- `filepath.Join(base, userInput)` → `os.OpenInRoot(base, userInput)` (1.24+)

**Tooling (Go 1.24-1.26)**:
- `tools.go` with build tag → `go get -tool <pkg>` in go.mod (1.24+)
- Temp variable for pointer: `age := x; &age` → `new(x)` (1.26+)
- Run `go fix ./...` to auto-apply many of the above (1.26+)

### Requires judgment (read the guide first)

**Generics** — see [guides/generics.md](guides/generics.md):
- Don't use generics to replace interfaces for behavior abstraction
- Don't add type parameters speculatively — wait for actual duplicate code
- Be aware of GCShape stenciling performance for pointer types
- Generic type aliases (1.24+) and self-referential constraints (1.26+) are advanced features

**Iterators** — see [guides/iterators-loops.md](guides/iterators-loops.md):
- Use for lazy sequences, composable pipelines, encapsulated traversal
- Don't use for simple slice iteration or when caller will `Collect` immediately
- `strings.SplitSeq` / `strings.Lines` avoid allocating `[]string` (1.23+)

**HTTP routing** — see [guides/http-networking.md](guides/http-networking.md):
- Stdlib now handles method matching + wildcards (1.22+)
- Use third-party router only if you need middleware chaining or regex constraints
- `ReverseProxy.Director` deprecated — use `Rewrite` (1.20+, deprecated 1.25)
- `math/rand` → `math/rand/v2` (1.22+)

**Structured logging** — see [guides/tooling-build.md](guides/tooling-build.md):
- `log/slog` for new projects — JSONHandler for production, TextHandler for dev
- Use `LogValuer` to redact sensitive data
- Use `LogAttrs` on hot paths to minimize allocations
- Don't migrate from zap/zerolog unless it simplifies the codebase

**Concurrency patterns** — see [guides/concurrency.md](guides/concurrency.md):
- `OnceValues` caches errors permanently — use custom retry pattern if needed
- `synctest.Test` for deterministic concurrent tests (1.25+)
- `unique.Make` for string interning in memory-constrained scenarios (1.23+)

**Security** — see [guides/security.md](guides/security.md):
- `os.Root` prevents path traversal — use for any user-provided file paths
- `crypto/hpke` for hybrid public key encryption (1.26+)
- `crypto/mlkem` for post-quantum key exchange (1.24+)

**Runtime & performance** — see [guides/tooling-build.md](guides/tooling-build.md):
- PGO with `default.pgo` for 2-14% runtime improvement (1.21+)
- Green Tea GC enabled by default (1.26+) — 10-40% GC overhead reduction
- Container-aware GOMAXPROCS auto-adjusts on Linux (1.25+)
- `trace.FlightRecorder` for production debugging (1.25+)

## How to Use

Read individual guide files for detailed explanations, code examples, and when-to-use/when-NOT-to-use guidance:

```
guides/stdlib-collections.md    — slices, maps, cmp, min/max/clear, CutPrefix
guides/error-handling.md        — errors.Join, AsType, context cause, AfterFunc
guides/iterators-loops.md       — range int, func iterators, custom iterators
guides/generics.md              — when to use/avoid, constraints, performance
guides/concurrency.md           — OnceValue, WaitGroup.Go, timer, synctest
guides/http-networking.md       — enhanced routing, ResponseController, CSRF
guides/testing.md               — b.Loop, t.Context, synctest, ArtifactDir
guides/security.md              — os.Root, rand.Text, HPKE, omitzero
guides/tooling-build.md         — go fix, PGO, slog, tool directives, new()
```

Each guide contains:
- Feature overview with minimum Go version required
- Code examples (old pattern → new pattern)
- When to use (concrete scenarios)
- When NOT to use (anti-patterns and pitfalls)
- Migration strategy (step-by-step adoption)

## Key Migration Patterns (Quick Code Reference)

```go
// min/max/clear (1.21+)
x := min(a, b)
v = min(max(v, lo), hi) // clamp
clear(m)                // delete all map entries

// slices/maps (1.21+)
slices.Contains(s, v)
slices.SortFunc(items, func(a, b T) int { return cmp.Compare(a.Name, b.Name) })
keys := slices.Sorted(maps.Keys(m))             // 1.23+
port := cmp.Or(cfg.Port, envPort, 8080)          // 1.22+

// Loops (1.22+)
for i := range 100 { ... }
for line := range strings.Lines(text) { ... }    // 1.23+

// Error handling (1.20+)
err = errors.Join(err, f.Close())                // defer pattern
cancel(fmt.Errorf("shutting down: %w", reason))  // WithCancelCause

// errors.AsType (1.26+)
if pathErr, ok := errors.AsType[*fs.PathError](err); ok { ... }

// Concurrency (1.21-1.25)
var getClient = sync.OnceValue(func() *Client { return newClient() })
wg.Go(func() { work() })                        // 1.25+

// HTTP routing (1.22+)
mux.HandleFunc("GET /items/{id}", handler)
id := r.PathValue("id")

// Testing (1.24+)
for b.Loop() { foo() }
ctx := t.Context()
t.Chdir("testdata")

// Security (1.24+)
f, _ := os.OpenInRoot("/var/data", userPath)
token := rand.Text()
```

## Version Quick Reference

| Go Version | Key Additions |
|------------|--------------|
| **1.20** | `errors.Join`, `context.WithCancelCause`, `CutPrefix`/`CutSuffix`, `ResponseController`, `bytes.Clone`, `time.DateTime`/`DateOnly`/`TimeOnly`, `errors.ErrUnsupported` |
| **1.21** | `min`/`max`/`clear` builtins, `slices`/`maps`/`cmp` packages, `log/slog`, `sync.OnceValue`/`OnceFunc`, `context.AfterFunc`/`WithoutCancel`, PGO stable |
| **1.22** | Loop variable scoping fix, `range` over integers, enhanced HTTP routing, `math/rand/v2`, `reflect.TypeFor[T]`, `cmp.Or`, `sql.Null[T]` |
| **1.23** | Range over function iterators, `iter` package, `unique` package, timer/ticker unbuffered channels, `slices.Chunk`/`Sorted`, `strings.Lines`/`SplitSeq` |
| **1.24** | Generic type aliases, `os.Root`, `testing.B.Loop`, `t.Context`/`t.Chdir`, `rand.Text`, `omitzero`, tool directives, `weak` package, `crypto/mlkem` |
| **1.25** | `sync.WaitGroup.Go`, `testing/synctest`, `encoding/json/v2` (experimental), container-aware GOMAXPROCS, Green Tea GC (experimental), `FlightRecorder` |
| **1.26** | `new(expr)`, self-referential constraints, `errors.AsType[T]`, `crypto/hpke`, `go fix` modernizers, Green Tea GC default, `slog.NewMultiHandler`, `reflect` iterators |

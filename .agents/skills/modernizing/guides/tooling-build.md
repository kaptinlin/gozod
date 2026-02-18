# Tooling & Build

Modern Go improves the development workflow: automatic code modernization, profile-guided optimization, structured logging, dependency management for tools, and better diagnostics.

## Contents
- go fix modernizers (1.26+)
- Profile-Guided Optimization / PGO (1.21+)
- Tool directives in go.mod (1.24+)
- log/slog structured logging (1.21+)
- go mod tidy -diff (1.23+)
- go vet new analyzers
- new() with expressions (1.26+)
- Runtime improvements

---

## go fix modernizers (Go 1.26+)

`go fix` has been revamped with dozens of modernizers that automatically update code to use newer Go idioms and APIs. Push-button way to adopt new features.

### When to use
- After upgrading Go version — run `go fix ./...` to apply all applicable modernizations
- Before code review — auto-apply mechanical improvements
- As part of CI — ensure codebase uses modern patterns

### What it modernizes (partial list)
- `for i := 0; i < b.N; i++` → `for b.Loop()`
- `errors.As(err, &target)` → `errors.AsType[T](err)`
- `sort.Slice` → `slices.SortFunc`
- Custom min/max → built-in `min`/`max`
- Old loop patterns → `for range n`
- And many more

```bash
# Apply all modernizers
go fix ./...

# Preview changes without applying
go fix -diff ./...
```

---

## Profile-Guided Optimization / PGO (Go 1.21+)

Compile using real-world CPU profiles to optimize hot paths. 2-14% runtime improvement in typical workloads.

### When to use
- Production services where you can collect CPU profiles
- After establishing performance baselines
- Auto-enabled when `default.pgo` file exists in the main package directory

### When NOT to use
- Development/test builds where compilation speed matters more
- Short-lived CLI tools where startup time dominates
- Without representative production profiles — bad profiles can hurt

### Workflow

```bash
# 1. Collect production CPU profiles
curl -o cpu.pprof 'http://localhost:6060/debug/pprof/profile?seconds=30'

# 2. Merge profiles from different instances/times for representativeness
go tool pprof -proto profile1.pprof profile2.pprof > merged.pprof

# 3. Place in main package and commit to repo
cp merged.pprof cmd/server/default.pgo
git add cmd/server/default.pgo

# 4. Build — PGO is auto-enabled with default.pgo
go build ./cmd/server
```

**Key details**:
- PGO devirtualizes interface calls when profiling shows a dominant concrete type (2-14% improvement)
- `-pgo=auto` is the default since Go 1.21 — just add `default.pgo`
- **Commit `default.pgo` to repo** — all developers and CI benefit automatically
- An unrepresentative profile won't make code slower — at worst, no benefit
- Use production profiles, not microbenchmarks — they cover too little of the app
- Re-collect profiles after significant code changes to limit source skew

---

## Tool directives in go.mod (Go 1.24+)

Declare executable tool dependencies directly in `go.mod`. Replaces the `tools.go` hack.

### When to use
- All tool dependencies (linters, code generators, etc.)
- Replacing `tools.go` with `//go:build tools` workaround

### When NOT to use
- Library dependencies — use regular `require` directives

```bash
# Old — tools.go workaround
# //go:build tools
# package tools
# import _ "golang.org/x/tools/cmd/stringer"

# New (Go 1.24+)
go get -tool golang.org/x/tools/cmd/stringer

# go.mod now contains:
# tool golang.org/x/tools/cmd/stringer

# Run the tool
go tool stringer -type=Color
```

---

## log/slog structured logging (Go 1.21+)

Standard library structured logging with levels, handlers, and context integration.

### When to use
- New projects — start with slog instead of `log` package
- When you need structured (JSON/logfmt) log output
- When you need log levels (Debug, Info, Warn, Error)
- Replacing `log.Printf` with structured fields

### When NOT to use
- Simple scripts where `fmt.Println` suffices
- If your existing structured logging library (zerolog, zap) is deeply integrated and working well — migration cost may not be worth it
- **Don't mix** `log` and `slog` in the same package — pick one

```go
// Basic usage
slog.Info("request handled",
    "method", r.Method,
    "path", r.URL.Path,
    "status", status,
    "duration", duration,
)

// JSON handler for production
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelInfo,
}))
slog.SetDefault(logger)

// With context (request-scoped fields)
logger = logger.With("request_id", requestID)
logger.InfoContext(ctx, "processing order", "order_id", orderID)

// Discard handler for tests (Go 1.25+)
slog.SetDefault(slog.New(slog.DiscardHandler))

// Multi-handler (Go 1.26+)
handler := slog.NewMultiHandler(jsonHandler, metricsHandler)
```

### Performance: use LogAttrs on hot paths

```go
// Faster — typed attributes, fewer allocations
slog.LogAttrs(ctx, slog.LevelInfo, "processed",
    slog.String("user", user),
    slog.Int("count", n))
```

### Redact sensitive data with LogValuer

```go
type Secret string

func (s Secret) LogValue() slog.Value {
    return slog.StringValue("***REDACTED***")
}

type User struct {
    Name     string
    Password Secret
}
// slog.Info("login", "user", user) → password is redacted
```

### slog with groups

```go
slog.Info("request",
    slog.Group("http",
        slog.String("method", "GET"),
        slog.Int("status", 200),
    ),
    slog.Group("user",
        slog.String("id", userID),
    ),
)
// JSON: {"msg":"request","http":{"method":"GET","status":200},"user":{"id":"..."}}
```

### Common mistakes
- Mismatched key-value pairs: `slog.Info("msg", "k1", v1, "k2")` — use `go vet` to catch
- Not using context: pass `ctx` to `InfoContext`/`ErrorContext` for trace correlation
- Over-logging at INFO: use DEBUG for diagnostics, INFO for business events

---

## go mod tidy -diff (Go 1.23+)

Shows what `go mod tidy` would change without modifying files. Useful for CI checks.

```bash
# Check if go.mod/go.sum are tidy (exit code 1 if not)
go mod tidy -diff

# In CI
go mod tidy -diff || (echo "run 'go mod tidy'" && exit 1)
```

---

## go vet new analyzers

### stdversion (Go 1.23+)
Flags references to stdlib symbols that are too new for the module's `go` directive.

### waitgroup (Go 1.25+)
Detects misplaced `WaitGroup.Add` calls (e.g., inside the goroutine instead of before it).

### hostport (Go 1.25+)
Detects IPv6-unsafe address construction (e.g., `host + ":" + port` instead of `net.JoinHostPort`).

```bash
go vet ./...  # runs all analyzers including new ones
```

---

## new() with expressions (Go 1.26+)

`new(expr)` now accepts an expression as operand, returning a pointer to a value initialized with that expression. Useful for optional pointer fields.

### When to use
- Optional pointer fields in struct literals (protobuf-style APIs)
- Inline pointer creation without a temporary variable

### When NOT to use
- When the expression is complex — readability matters, use a variable
- When a helper function like `ptr(v)` already exists in your codebase

```go
// Old — need a temporary variable
age := yearsSince(born)
p := Person{Name: name, Age: &age}

// New (Go 1.26+) — inline pointer
p := Person{Name: name, Age: new(yearsSince(born))}

// Old — pointer to literal
s := "hello"
p := &s

// New (Go 1.26+)
p := new("hello")
```

---

## Runtime improvements worth knowing

### Container-aware GOMAXPROCS (Go 1.25+)
On Linux, the runtime now considers cgroup CPU bandwidth limits and periodically adjusts GOMAXPROCS. No code changes needed — just deploy.

### Green Tea GC (Go 1.25 experimental, Go 1.26 default)
10-40% GC overhead reduction. Enabled by default in Go 1.26. Disable with `GOEXPERIMENT=nogreenteagc` if issues arise.

### Cgo calls ~30% faster (Go 1.26+)
Reduced baseline overhead for C function calls from Go.

### Flight recorder (Go 1.25+)
Lightweight continuous trace capture with snapshot-on-demand for production debugging.

```go
fr := trace.NewFlightRecorder()
fr.Start()

// On critical error, capture trace
if criticalError {
    f, _ := os.Create("trace.out")
    fr.WriteTo(f)
    f.Close()
}
```

---

## Migration strategy

1. Run `go fix ./...` after upgrading to Go 1.26+ — free mechanical improvements
2. Replace `tools.go` with `go get -tool` directives (Go 1.24+)
3. Add `default.pgo` to production services for PGO benefits
4. Add `go mod tidy -diff` to CI pipeline
5. Migrate from `log` to `slog` for new code; migrate existing code gradually
6. Run `go vet ./...` to catch new analyzer warnings

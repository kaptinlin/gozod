# Pattern: Options Only

## Signature

```go
func New(opts ...Option) *T
// or with required runtime dependency:
func New(required Dep, opts ...Option) *T
```

## When to Use

A package has **no serializable data parameters** — only runtime dependencies that cannot appear in a config file.

**Indicators:**
- All parameters are interfaces (`slog.Handler`, `trace.Tracer`)
- All parameters are functions (`func(string) string`, callbacks)
- All parameters are runtime objects (`*slog.Logger`, `io.Writer`)
- Nothing meaningful to write in YAML/JSON

## Design Rules

### 1. Required Dependencies as Positional Parameters

If a dependency is mandatory and has no meaningful default, make it a positional parameter:

```go
// Good: handler is required, no sensible default
func New(handler slog.Handler, opts ...Option) *Audit

// Good: resolve function is required
func New(resolve ResolveFunc, opts ...Option) *Proxy
```

### 2. Optional Dependencies as Options

```go
type Option func(*Plugin) error

func WithLogger(l *slog.Logger) Option {
    return func(p *Plugin) error { p.logger = l; return nil }
}

func WithTracer(t trace.Tracer) Option {
    return func(p *Plugin) error { p.tracer = t; return nil }
}
```

### 3. No Config Struct

Don't create a Config struct that holds interfaces or functions:

```go
// Bad: Config with non-serializable fields
type Config struct {
    Logger  *slog.Logger   // can't serialize
    Handler slog.Handler   // can't serialize
    OnError func(error)    // can't serialize
}

// Good: use Options
func WithLogger(l *slog.Logger) Option { ... }
func WithHandler(h slog.Handler) Option { ... }
func WithOnError(fn func(error)) Option { ... }
```

### 4. Option Signature Consistency

Match the parent framework's Option pattern:

```go
// If framework uses func(*T) error:
type Option func(*Backend) error

// If package never returns constructor error and options can't fail:
type Option func(*Guard)  // simpler, acceptable
```

## When Options-Only Becomes Config + Options

If serializable parameters are added later, migrate to `New(cfg Config, opts ...Option)`:

**Before:**
```go
// Only runtime deps
func New(opts ...Option) *Guard
```

**After:**
```go
// Gained serializable parameters
type Config struct {
    DenyPatterns []string
    DetectLevel  DetectLevel
}

func New(cfg Config, opts ...Option) *Guard
```

**Trigger:** When you find yourself writing `With*` options for `string`, `int`, `bool`, or `[]string` values that a user might want in a config file.

## Real Examples

| Package | Constructor | Required Dep | Optional Deps |
|---------|-------------|-------------|---------------|
| `seatbelt` | `New(opts ...Option)` | — | `WithLogger` |
| `landlock` | `New(opts ...Option)` | — | `WithLogger` |
| `audit` | `New(handler, opts...)` | `slog.Handler` | — |
| `otel` | `New(opts ...Option)` | — | `WithTracer` |
| `secrets` | `New(resolve, opts...)` | `ResolveFunc` | `WithSecrets`, `WithLogger` |

## Framework / Orchestrator Pattern

The top-level orchestrator (e.g., `sandbox.New`) typically uses Options-Only because it accepts composed types:

```go
// Orchestrator: composes backends, policies, plugins
func New(opts ...Option) (*Sandbox, error)

func WithBackend(b Backend) Option { ... }       // interface
func WithPolicy(p Policy) Option { ... }         // struct (but set programmatically)
func WithPlugin(p ...Plugin) Option { ... }      // interface
func WithLogger(l *slog.Logger) Option { ... }   // runtime dep
```

The orchestrator doesn't have a Config struct because its "configuration" is the composition of other components — each with their own Config.

---
description: Design configuration strategy for Go projects — choose between go-config, Config struct, and With* functional options. Use when designing new Go packages, libraries, services, or applications that need configuration. Triggers on config design, constructor patterns, or option patterns.
name: config-designing
---


# Go Configuration Design Guide

Select the right configuration pattern for Go projects. Three mechanisms serve different purposes and compose together.

## Decision Flowchart

```
What are you building?
│
├─ Application / Service (has main())
│  └─ Use go-config to load from multiple sources
│     └─ Config structs from libraries compose into your AppConfig
│
├─ Library / Package (imported by others)
│  │
│  ├─ Has serializable parameters? (string, int, bool, []string, Duration)
│  │  └─ New(cfg Config, opts ...Option) — Config holds data, Option holds runtime deps
│  │
│  └─ Only runtime dependencies? (logger, tracer, callback, interface)
│     └─ New(opts ...Option) — no Config struct needed
│
└─ Framework / Orchestrator (composes other packages)
   └─ New(opts ...Option) — accept composed types via With* functions
```

## Three Mechanisms

| Mechanism | Purpose | Holds | Serializable | Example |
|-----------|---------|-------|--------------|---------|
| `Config` struct | Data parameters | `string`, `int`, `bool`, `[]string`, `time.Duration` | Yes | `docker.Config{Image: "node:20"}` |
| `With*` Option | Runtime dependencies | `*slog.Logger`, `func(...)`, interfaces | No | `WithLogger(logger)` |
| `go-config` | Source loading | Files, env vars, flags, secrets → `Config` structs | Yes | `config.Load[AppConfig](sources)` |

**Key rule:** Config holds what can be written in a config file. Option holds what can only exist at runtime.

## Pattern 1: Config Struct + Options — [details](references/config-struct.md)

Use when a package has serializable configuration parameters.

```go
// Constructor signature
func New(cfg Config, opts ...Option) (*T, error)
```

**When to use:**
- Package has `string`/`int`/`bool`/`[]string`/`Duration` parameters
- Users may want to load these from files, env vars, or flags
- Zero-value Config should produce valid defaults

**Design rules:**

| Rule | Rationale |
|------|-----------|
| Value receiver `Config` (not pointer) | `Config{}` is valid zero value, no nil check needed |
| `cmp.Or(cfg.Field, default)` for defaults | Zero value = unset = use default |
| No yaml/toml tags on Config | Library doesn't dictate config format; use `json` tags only if needed |
| Enum types implement `TextMarshaler`/`TextUnmarshaler` | Any config library (go-config, JSON, YAML) auto-converts `"deny"` → `NetworkDeny` |
| Option returns `error` | `func(*T) error` — consistent, validates at apply time |
| Config holds data, Option holds runtime deps | Logger, tracer, callback → Option. Port, host, timeout → Config |

**Example:**

```go
type Config struct {
    Image       string
    MemoryLimit string
    CPUQuota    float64
    PidsLimit   int
}

type Option func(*Backend) error

func WithLogger(l *slog.Logger) Option {
    return func(b *Backend) error { b.logger = l; return nil }
}

func WithCLIPath(path string) Option {
    return func(b *Backend) error { b.cliPath = path; return nil }
}

func New(cfg Config, opts ...Option) (*Backend, error) {
    b := &Backend{
        image:       cmp.Or(cfg.Image, "ubuntu:24.04"),
        memoryLimit: cmp.Or(cfg.MemoryLimit, "512m"),
        cpuQuota:    cmp.Or(cfg.CPUQuota, 1.0),
        pidsLimit:   cmp.Or(cfg.PidsLimit, 256),
        logger:      slog.Default(),
    }
    for _, opt := range opts {
        if err := opt(b); err != nil {
            return nil, err
        }
    }
    return b, nil
}
```

## Pattern 2: Options Only — [details](references/options-only.md)

Use when a package has no serializable data — only runtime dependencies.

```go
// Constructor signature
func New(opts ...Option) *T
```

**When to use:**
- All parameters are interfaces, functions, or runtime objects
- Nothing meaningful to put in a config file
- Package is a thin wrapper around runtime behavior

**Examples:**

| Package | Why Options Only |
|---------|-----------------|
| `seatbelt.New(opts...)` | OS-level backend, no serializable config |
| `audit.New(handler, opts...)` | Parameter is `slog.Handler` (runtime) |
| `otel.New(opts...)` | Parameter is `trace.Tracer` (runtime) |
| `secrets.New(resolve, opts...)` | Parameter is `func(string) string` (runtime) |

## Pattern 3: go-config for Applications — [details](references/go-config-usage.md)

Use in applications (has `main()`) to compose library Config structs into a unified config.

```go
// Application-level config composes library configs
type AppConfig struct {
    Server ServerConfig       `json:"server"`
    Docker docker.Config      `json:"docker"`
    Guard  commandguard.Config `json:"guard"`
    Filter domainfilter.Config `json:"filter"`
}

cfg, err := config.Load[AppConfig]([]config.Source{
    static.New(defaults),
    file.New("config.yaml"),
    env.New(env.WithPrefix("APP_")),
}, config.WithDecodeHook(func(s string) (time.Duration, error) {
    return time.ParseDuration(s)
}))
```

**When to use go-config:**
- Application with `main()` that assembles components
- Need multi-source config (file + env + flags + secrets)
- Want hot reload, explainability, or source composition

**When NOT to use go-config:**
- Inside a library package — libraries expose `Config` structs, consumers choose how to populate them
- Simple CLI tool with only env vars — use `caarlos0/env/v11`
- Only loading `.env` files — use `joho/godotenv`

## TextMarshaler / TextUnmarshaler for Enum Types

Any enum type that appears in a `Config` struct MUST implement `encoding.TextMarshaler` and `encoding.TextUnmarshaler`. This enables automatic conversion by go-config, `encoding/json`, and any config library.

```go
type NetworkPolicy int

const (
    NetworkDeny NetworkPolicy = iota
    NetworkAllow
    NetworkAllowList
)

func (n NetworkPolicy) MarshalText() ([]byte, error) {
    switch n {
    case NetworkDeny:
        return []byte("deny"), nil
    case NetworkAllow:
        return []byte("allow"), nil
    case NetworkAllowList:
        return []byte("allowlist"), nil
    default:
        return nil, fmt.Errorf("%w: %d", ErrUnknownNetworkPolicy, n)
    }
}

func (n *NetworkPolicy) UnmarshalText(text []byte) error {
    switch string(text) {
    case "deny", "":
        *n = NetworkDeny
    case "allow":
        *n = NetworkAllow
    case "allowlist":
        *n = NetworkAllowList
    default:
        return fmt.Errorf("%w: %q", ErrUnknownNetworkPolicy, string(text))
    }
    return nil
}
```

**Result:** Config file `network: deny` automatically maps to `NetworkDeny`. No intermediate config types needed.

## Anti-Patterns

| Anti-Pattern | Problem | Fix |
|-------------|---------|-----|
| Library reads config files directly | Couples library to specific format/location | Expose `Config` struct, let application load |
| YAML/TOML tags on library types | Forces format dependency on consumers | Use `json` tags only, or no tags |
| `LoadConfig()` in library | Library shouldn't know about env vars or file paths | Delete; application uses go-config |
| `*Config` pointer parameter | Forces nil check, `New(nil)` ambiguity | Value receiver `Config`, `New(Config{})` is zero-value |
| `Default` prefix on Config fields | `Config.DefaultImage` — "Default" is implied | `Config.Image` |
| Conflicting defaults across layers | Library has 10MB max output, framework has 1MB | Single source of truth; trust framework defaults |
| Option for serializable data | `WithImage("node:20")` duplicates what Config does | Config field. Option only for runtime deps |
| Config for runtime dependency | `Config{Logger: slog.Logger{}}` — not serializable | `WithLogger(l)` Option |

## Quick Classification Checklist

For each parameter in a new package:

| Question | Yes → | No → |
|----------|-------|------|
| Can it be written in YAML/JSON? | Config field | Option |
| Is it `string`, `int`, `bool`, `[]string`? | Config field | — |
| Is it `*slog.Logger`, `io.Writer`, `func(...)`? | Option | — |
| Is it an `interface` the caller provides? | Option | — |
| Is it `time.Duration`? | Config field (go-config handles via decode hook) | — |
| Is it an enum (`int` with named constants)? | Config field + TextUnmarshaler | — |
| Does zero value mean "use default"? | Config field + `cmp.Or` | — |
| Does it detect runtime environment? | Option (`WithCLIPath`) | — |

## Complete Integration Example

See [references/integration-example.md](references/integration-example.md) for a full application wiring library Config structs through go-config.

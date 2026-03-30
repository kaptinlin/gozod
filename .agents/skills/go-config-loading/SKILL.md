---
description: Load and manage typed application configuration using agentable/go-config with multi-source composition, hot reload, and explainability. Use when loading config from files, environment variables, flags, or secrets, setting up config hot reload, or debugging config provenance.
name: go-config-loading
---


# go-config Library Usage Guide

Load typed configuration from multiple sources using `github.com/agentable/go-config`. For design patterns (Config struct vs Options vs go-config), see the **config-designing** skill.

## Core API

```go
import "github.com/agentable/go-config"

// Load -- primary entry point, returns populated Config[T]
cfg, err := config.Load[AppConfig](sources, opts...)

// MustLoad -- panics on error (use in main() or tests only)
cfg := config.MustLoad[AppConfig](sources, opts...)

// New + Load -- two-phase construction (needed for Watch)
cfg := config.New[AppConfig](opts...)
err := cfg.Load(ctx, sources...)

// Read -- lock-free atomic snapshot
v := cfg.Value()

// Decode -- standalone map-to-struct conversion
var v AppConfig
err := config.Decode(src, &v)
```

`Load` applies sources in order (later wins), deep-merges maps, replaces scalars and slices, interpolates `${env:*}`/`${secret:*}`/`${path}` references, decodes into `T`, and stores atomically.

## Source Composition

Sources are applied in order; later sources override earlier ones:

```go
cfg, err := config.Load[AppConfig]([]config.Source{
    static.New(defaults),                                    // 1. defaults
    file.New("config.yaml"),                                 // 2. config file
    file.New("config.local.yaml", file.WithPolicy(config.PolicyOptional)), // 3. local overrides
    env.New(env.WithPrefix("APP_")),                         // 4. environment
    secretsSrc,                                              // 5. secrets
    flags,                                                   // 6. CLI flags (highest)
})
```

## Provider Selection

| Provider | Import | Watch | Use Case |
|----------|--------|-------|----------|
| `file` | `provider/file` | Yes | OS files with auto-format detection by extension |
| `file.Discovery` | `provider/file` | Yes | Walk-up file discovery (find nearest config file) |
| `fs` | `provider/fs` | No | `embed.FS` / `io/fs.FS` (immutable) |
| `env` | `provider/env` | No | Environment variables with prefix/delimiter mapping |
| `flag` | `provider/flag` | No | CLI flags from struct tags via `spf13/pflag` |
| `static` | `provider/static` | No | In-memory defaults from structs or maps |
| `secrets` | `provider/secrets` | No | Secrets from `agentable/go-secrets` |

Provider details: [references/providers.md](references/providers.md)

## Format Registration

JSON is built-in. YAML and TOML require blank imports:

```go
import (
    _ "github.com/agentable/go-config/format/yaml" // .yaml, .yml
    _ "github.com/agentable/go-config/format/toml" // .toml
)
```

File and FS providers auto-detect format by extension. No manual decoder setup needed.

## Options

```go
// Custom type conversion (e.g., string -> time.Duration)
config.WithDecodeHook(func(s string) (time.Duration, error) {
    return time.ParseDuration(s)
})

// Secret interpolation resolver for ${secret:KEY} in config files
config.WithSecretResolver(resolver)

// Status monitoring for metrics/observability
config.WithOnStatus(func(s config.Source, changed bool, err error) {
    if err != nil { metrics.RecordError(s, err) }
})
```

## Default Tags

Struct fields support `default` tags -- applied when the field is absent from all sources:

```go
type Config struct {
    Host    string        `json:"host"    default:"localhost"`
    Port    int           `json:"port"    default:"8080"`
    Timeout time.Duration `json:"timeout" default:"30s"`
}
```

## Interpolation

Config file values support variable references:

```yaml
server:
  host: "${env:SERVER_HOST}"
  dsn: "postgres://${env:DB_USER}:${secret:DB_PASS}@${server.host}:5432/mydb"
```

| Syntax | Resolves From |
|--------|---------------|
| `${env:VAR}` | Environment variable |
| `${secret:KEY}` | SecretResolver (requires `WithSecretResolver`) |
| `${path.to.key}` | Cross-reference another config value |

Circular references are detected and return an error.

## Watch and Hot Reload

```go
// Register callbacks before starting watch
cfg.OnChange(func(c *config.Config[AppConfig]) {
    v := c.Value()
    updateComponent(v.Database)
}, "database.host", "database.port") // fire only on specific path changes

// Typed callback -- uses == instead of reflect.DeepEqual
config.OnChangeFunc(cfg,
    func(c AppConfig) string { return c.Server.Host },
    func(host string) { rebind(host) },
)

// Start watching (blocks until ctx is done)
go cfg.Watch(ctx)
```

Only sources implementing `Watcher` (currently: `file`, `file.Discovery`) trigger reload. All sources are reloaded atomically on any change.

Details: [references/watch-patterns.md](references/watch-patterns.md)

## Explainability

```go
exp := cfg.Explain("database.password")
// exp.Value    -- final value (sensitive values auto-blurred)
// exp.Source   -- source name that set it, e.g., "env:APP_*"
// exp.Overrides -- values replaced by higher-priority sources
// exp.Skipped  -- optional sources that were skipped
// exp.Failed   -- sources that failed to load (with policy)
```

Sensitive values are automatically blurred:
- **Path-based**: fields with names containing `password`, `secret`, `token`, `key`, `cred`, `bearer`, `private_key`, etc. show `******`
- **Pattern-based**: values matching known formats (AWS keys, GitHub tokens, Stripe keys, JWTs, PEM keys, etc.) show the pattern name (e.g., `"AWS API key"`)
- Path-based blurring takes priority over pattern-based for high-sensitivity paths (`password`, `secret`, `private_key`)

Use `Explain` in debug endpoints, startup logs, or troubleshooting scripts.

## Error Handling

Load attempts all sources and aggregates errors via `errors.Join`.

| Policy | Behavior on Error |
|--------|-------------------|
| `PolicyRequired` (default) | Fatal -- error returned to caller |
| `PolicyOptional` | Recorded in `Explain()`, source skipped |

```go
// Optional config file -- won't fail if missing
file.New("config.local.yaml", file.WithPolicy(config.PolicyOptional))
```

Error types:

| Type | Description |
|------|-------------|
| `*SourceError` | Wraps source-specific error with source name |
| `*DecodeError` | Aggregates field conversion errors |
| `*FieldError` | Single field conversion failure (path, type, value) |
| `*ErrNoFormat` | No decoder registered for extension (suggests import) |
| `ErrNoSources` | No sources passed to Load |
| `ErrAlreadyWatching` | Watch called twice on same Config |

## Custom Providers

Embed `config.PolicyMixin` to get policy/skip-error handling for free:

```go
type MyProvider struct {
    config.PolicyMixin
    // ... your fields
}

func (p *MyProvider) Load(ctx context.Context) (map[string]any, error) {
    p.ClearSkipError()
    data, err := fetchData(ctx)
    if err != nil {
        return p.HandleError(err) // respects PolicyOptional
    }
    return data, nil
}

func (p *MyProvider) String() string { return "my-provider" }
```

Implement `fmt.Stringer` for meaningful names in `Explain` output.

## OptionalBool

Three-state boolean distinguishing "not set" from "explicitly false":

```go
type AppConfig struct {
    Debug config.OptionalBool `json:"debug"`
}

v := cfg.Value()
v.Debug.IsSet()   // true if explicitly provided
v.Debug.IsTrue()  // true if set to true
v.Debug.IsFalse() // true if set to false
```

Supports text unmarshaling: `true/1/yes/on` and `false/0/no/off`.

## Testing with go-config

```go
// Use static provider for deterministic tests
cfg, err := config.Load[AppConfig]([]config.Source{
    static.New(AppConfig{Server: Server{Port: 9090}}),
})

// Use env provider with custom source for test isolation
cfg, err := config.Load[AppConfig]([]config.Source{
    env.New(
        env.WithPrefix("APP_"),
        env.WithSource(map[string]string{"APP_SERVER_PORT": "9090"}),
    ),
})

// Use Decode directly for unit tests
var cfg AppConfig
err := config.Decode(map[string]any{"port": 8080}, &cfg)
```

## Anti-Patterns

| Anti-Pattern | Problem | Fix |
|-------------|---------|-----|
| Library imports go-config | Couples library to config loading mechanism | Library exposes `Config` struct; app uses go-config |
| Missing format blank import | `ErrNoFormat` at runtime for YAML/TOML files | Add `_ "github.com/agentable/go-config/format/yaml"` |
| `PolicyRequired` on local overrides file | App crashes if `config.local.yaml` missing | `file.WithPolicy(config.PolicyOptional)` |
| Calling `Watch` without `OnChange` | Watch runs but nothing reacts to changes | Register callbacks before calling `Watch` |
| Blocking on `cfg.Watch(ctx)` in main goroutine | App hangs -- Watch blocks until ctx done | `go cfg.Watch(ctx)` in separate goroutine |
| Ignoring `Load` error | Silent misconfiguration, zero-value struct used | Always check `err` from `Load` |
| Using `MustLoad` in library code | Panic in library is unrecoverable | `MustLoad` only in `main()` or test setup |
| Duplicate decode hooks for same type | Last hook wins, confusing behavior | Register one hook per type conversion |

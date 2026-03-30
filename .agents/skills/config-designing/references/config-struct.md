# Pattern: Config Struct + Options

## Signature

```go
func New(cfg Config, opts ...Option) (*T, error)
```

## When to Use

A package has **serializable data parameters** — values that make sense in a config file.

**Indicators:**
- Constructor takes `string`, `int`, `bool`, `[]string`, `float64`, `time.Duration`
- Users want to load parameters from YAML/JSON/env vars
- Zero-value `Config{}` should produce a working instance with defaults

## Design Rules

### 1. Value Receiver Config (Not Pointer)

```go
// Good: zero value is valid, no nil check
func New(cfg Config, opts ...Option) (*Backend, error)

b, _ := docker.New(docker.Config{})  // all defaults
b, _ := docker.New(appCfg.Docker)    // from config file

// Bad: requires nil check, ambiguous
func New(cfg *Config, opts ...Option) (*Backend, error)

b, _ := docker.New(nil)  // what does nil mean?
```

### 2. Zero Value = Default via cmp.Or

```go
type Config struct {
    Image       string   // "" → "ubuntu:24.04"
    MemoryLimit string   // "" → "512m"
    CPUQuota    float64  // 0  → 1.0
    PidsLimit   int      // 0  → 256
}

func New(cfg Config, opts ...Option) (*Backend, error) {
    b := &Backend{
        image:       cmp.Or(cfg.Image, "ubuntu:24.04"),
        memoryLimit: cmp.Or(cfg.MemoryLimit, "512m"),
        cpuQuota:    cmp.Or(cfg.CPUQuota, 1.0),
        pidsLimit:   cmp.Or(cfg.PidsLimit, 256),
    }
    // ...
}
```

**Note:** `cmp.Or` returns the first non-zero value. For `int`/`float64`, zero means "use default". If zero is a valid explicit value, use a pointer or a sentinel.

### 3. No YAML/TOML Tags

```go
// Good: library-neutral, application chooses format
type Config struct {
    Image       string
    MemoryLimit string
    CPUQuota    float64
    PidsLimit   int
}

// Acceptable: json tags for API serialization
type Config struct {
    Image       string  `json:"image,omitzero"`
    MemoryLimit string  `json:"memory_limit,omitzero"`
    CPUQuota    float64 `json:"cpu_quota,omitzero"`
    PidsLimit   int     `json:"pids_limit,omitzero"`
}

// Bad: forces yaml dependency on consumers
type Config struct {
    Image string `yaml:"image" json:"image"`
}
```

### 4. Config vs Option Split

```go
type Config struct {
    // Serializable data → Config
    Image       string
    MemoryLimit string
    CPUQuota    float64
    PidsLimit   int
}

type Option func(*Backend) error

// Runtime dependency → Option
func WithLogger(l *slog.Logger) Option {
    return func(b *Backend) error { b.logger = l; return nil }
}

// Runtime environment detection → Option
func WithCLIPath(path string) Option {
    return func(b *Backend) error { b.cliPath = path; return nil }
}
```

### 5. Option Returns Error

```go
// Good: consistent, validates at apply time
type Option func(*Backend) error

func WithCLIPath(path string) Option {
    return func(b *Backend) error {
        if path == "" {
            return errors.New("empty CLI path")
        }
        b.cliPath = path
        return nil
    }
}

// Acceptable for plugins without error cases:
type Option func(*Guard)  // only if New() never returns error
```

### 6. No "Default" Prefix

```go
// Good: Config fields are defaults by definition
type Config struct {
    Image       string
    MemoryLimit string
}

// Bad: redundant prefix
type Config struct {
    DefaultImage       string
    DefaultMemoryLimit string
}
```

## Enum Fields: TextMarshaler/TextUnmarshaler

Any enum type in a Config struct implements `encoding.TextMarshaler` and `encoding.TextUnmarshaler`:

```go
type DetectLevel int

const (
    DetectNone DetectLevel = iota
    DetectBasic
    DetectDeep
)

func (d DetectLevel) MarshalText() ([]byte, error) {
    switch d {
    case DetectNone:  return []byte("none"), nil
    case DetectBasic: return []byte("basic"), nil
    case DetectDeep:  return []byte("deep"), nil
    default:          return nil, fmt.Errorf("%w: %d", errUnknownDetectLevel, d)
    }
}

func (d *DetectLevel) UnmarshalText(text []byte) error {
    switch string(text) {
    case "none":       *d = DetectNone
    case "basic", "":  *d = DetectBasic  // empty = default
    case "deep":       *d = DetectDeep
    default:           return fmt.Errorf("%w: %q", errUnknownDetectLevel, string(text))
    }
    return nil
}

type Config struct {
    DenyPatterns []string
    DetectLevel  DetectLevel  // YAML "deep" → DetectDeep automatically
}
```

## Duration Fields

`time.Duration` is already supported by go-config via decode hook:

```go
// In Config struct — just use time.Duration
type ResourceLimits struct {
    Timeout time.Duration
}

// Application registers decode hook once
cfg, _ := config.Load[AppConfig](sources,
    config.WithDecodeHook(func(s string) (time.Duration, error) {
        return time.ParseDuration(s)
    }),
)
```

Config file: `timeout: 30s` → `30 * time.Second`.

## Real Examples

### Has Config Struct

| Package | Config Fields | Options |
|---------|--------------|---------|
| `docker` | Image, MemoryLimit, CPUQuota, PidsLimit | WithLogger, WithCLIPath |
| `wasm` | CacheDir, MemoryLimitPages | WithLogger |
| `commandguard` | DenyPatterns, DetectLevel | WithDenyPatterns, WithDetectLevel |
| `domainfilter` | AllowedDomains, DeniedDomains | WithLogger |

### No Config Struct (Options Only)

| Package | Why | Constructor |
|---------|-----|-------------|
| `seatbelt` | No serializable params | `New(opts ...Option)` |
| `landlock` | No serializable params | `New(opts ...Option)` |
| `audit` | Param is `slog.Handler` | `New(handler, opts...)` |
| `otel` | Param is `trace.Tracer` | `New(opts ...Option)` |
| `secrets` | Param is `func(string) string` | `New(resolve, opts...)` |

## Migration: Adding Config to Existing Package

If a package currently uses `New(opts ...Option)` and gains serializable parameters:

1. Create `Config` struct with the new fields
2. Change signature: `New(opts ...Option)` → `New(cfg Config, opts ...Option)`
3. Existing `With*` options for data fields can stay for programmatic overrides
4. `Config{}` (zero value) preserves backward compatibility in behavior

```go
// Before
g := commandguard.New(
    commandguard.WithDenyPatterns("rm -rf", "shutdown"),
    commandguard.WithDetectLevel(commandguard.DetectDeep),
)

// After — both work
g := commandguard.New(commandguard.Config{
    DenyPatterns: []string{"rm -rf", "shutdown"},
    DetectLevel:  commandguard.DetectDeep,
})

g := commandguard.New(commandguard.Config{},
    commandguard.WithDenyPatterns("rm -rf", "shutdown"),
    commandguard.WithDetectLevel(commandguard.DetectDeep),
)
```

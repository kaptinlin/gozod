# Configuration & CLI

## KISS First

- `.env` loading and env parsing are different concerns.
- Start with the smallest layer that solves the real problem.
- Avoid multi-source config frameworks for tiny env-only cases.

## `github.com/agentable/go-dotenv` — `.env` Loader

- Loads `.env` files with variable expansion and multi-environment discovery
- Supports `${secret:name}` resolution via `agentable/go-secrets`
- Can inject values into `os.Environ` for downstream config providers

**When to use:** Local dev, examples, tests, and CI that rely on `.env` files, especially when secrets resolution should stay in our ecosystem.

**When NOT to use:** As a replacement for full application config composition.

## `github.com/agentable/go-config` — Application Configuration

Core configuration library with zero external dependencies.

**Architecture:** Type-safe `Config[T]` with generics, unified `Source` interface, lock-free reads via `atomic.Pointer`, hot reload support.

**Module isolation keeps deps minimal:**
```
go-config core                    → 0 external deps (stdlib only)
+ format/json                     → +go-json-experiment/json (built-in)
+ format/yaml                     → +goccy/go-yaml
+ format/toml                     → +pelletier/go-toml/v2
+ provider/file                   → +fsnotify/fsnotify (watch support)
+ provider/fs                     → 0 external deps (embed.FS, no watch)
+ provider/env                    → 0 external deps
+ provider/flag                   → +spf13/pflag
+ provider/static                 → 0 external deps
+ provider/secre             → requires SecretClient implementation
```

Users only pull deps for features they use.

**Key Features:**
- Lock-free reads: `Value()` uses `atomic.Pointer[T]` for zero-overhead concurrent access
- Source composition: Files, env vars, flags, secrets, and static values share single `Source` interface
- Live reload: `Watch()` detects changes and reloads all sources atomically
- Typed callbacks: `OnChange()` and `OnChangeFunc()` fire on specific path or value changes
- Explainability: `Explain()` traces which source provided each value with full override history
- Policy control: `PolicyRequired` vs `PolicyOptional` for flexible error handling
- Auto-detect formats: JSON built-in; YAML and TOML via blank import

**When to use:** Any application needing multi-source config (env + file + flags), hot reload, or type-safe config structs.

### Format selection

| Format | Library | Why |
|--------|---------|-----|
| JSON (built-in) | `go-json-experiment/json` | 2.5-10x faster than stdlib, `omitzero`, auto-registered |
| YAML | `goccy/go-yaml` | yaml.v3 **archived 2025-04**; supports `json` tags, better spec (355/402 vs 295/402) |
| TOML | `pelletier/go-toml/v2` | 2-5x faster than BurntSushi, zero deps, `encoding/json`-style API |

### Provider selection

| Provider | Use Case | Watch Support | Dependencies |
|----------|----------|---------------|--------------|
| `provider/file` | OS files with hot reload | ✅ Yes | fsnotify |
| `provider/fs` | embed.FS, immutable configs | ❌ No | None |
| `provider/env` | Environment variables | ❌ No | None |
| `provider/flag` | CLI flags (POSIX) | ❌ No | pflag |
| `provider/static` | In-memory defaults | ❌ No | None |
| `provider/secrets` | Third-party secret services | ❌ No | Custom client |

### Example Usage

```go
import (
    "github.com/agentable/go-config"
    "github.com/agentable/go-config/provider/file"
    "github.com/agentable/go-config/provider/env"
    _ "github.com/agentable/go-config/format/yaml" // auto-register
)

type AppConfig struct {
    Server struct {
        Host string `json:"host"`
        Port int    `json:"port"`
    } `json:"server"`
}

cfg, err := config.Load[AppConfig]([]config.Source{
    file.New("config.yaml"),
    env.New(env.WithPrefix("APP_")),
})

v := cfg.Value() // lock-free atomic read
fmt.Printf("server=%s:%d\n", v.Server.Host, v.Server.Port)

// Watch for changes
go cfg.Watch(ctx)
```

## `github.com/agentable/go-command` — CLI Framework

- Web-like command API with middleware, typed binding, help generation, and mode-aware output
- Supports app-level options, struct binding, examples, shell completion, and schema export
- Works well with `go-config` and `go-dotenv` for CLI-first applications

**When to use:** Building full CLI applications with subcommands, middleware, typed option binding, and agent/CI-aware output.

**When NOT to use:** If you only need config loading without a CLI surface.

## `github.com/agentable/go-secrets` — Secrets Management

- Plugin-driven architecture
- Zero-config default: file store + AES-256-GCM encryption
- Scope-based isolation

**When to use:** Sensitive config values referenced via `${secret:NAME}` in configuration.

## `github.com/agentable/go-version` — Version Metadata

- Single source of truth for version, commit, branch, build date, and runtime metadata
- Stable simple/text/JSON/header output for CLIs, services, and diagnostics
- Build-time injection via `-ldflags -X`

**When to use:** Exposing version/build identity in CLIs or HTTP services without ad-hoc formatting code.

**When NOT to use:** As a replacement for app configuration or command routing.

## Decision Tree

```
Need configuration handling?
├── Only load `.env` / resolve `${secret:...}`?
│   └── agentable/go-dotenv
├── Multi-source config (file + env + flags + secrets), optional hot reload?
│   └── agentable/go-config
└── Full CLI app with commands, middleware, and typed binding?
    └── agentable/go-command
```

## Do NOT Use

| Library | Reason | Use Instead |
|---------|--------|-------------|
| `spf13/viper` | We have go-config | `agentable/go-config` |
| `knadh/koanf/v2` | We have go-config | `agentable/go-config` |
| `joho/godotenv` | Prefer our `.env` loader with secret resolution | `github.com/agentable/go-dotenv` |
| `spf13/cobra` | Prefer our CLI framework for new apps | `github.com/agentable/go-command` |
| `gopkg.in/yaml.v3` | **Archived 2025-04** | `github.com/goccy/go-yaml` |
| `BurntSushi/toml` | 2-5x slower | `pelletier/go-toml/v2` |
| `hashicorp/hcl` | YAGNI — Terraform-specific | — |
| `gopkg.in/ini.v1` | YAGNI — INI is legacy | — |

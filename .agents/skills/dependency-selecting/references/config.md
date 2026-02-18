# Configuration & CLI

## KISS First

- `.env` loading and env parsing are different concerns.
- Start with the smallest layer that solves the real problem.
- Avoid multi-source config frameworks for tiny env-only cases.

## `github.com/joho/godotenv` — `.env` Loader

- Loads `.env` files into process environment variables
- Tiny API surface (`Load`, `Overload`, `Read`)

**When to use:** Local dev, examples, tests, and CI that rely on `.env` files.

**When NOT to use:** As a replacement for env-to-struct parsing/validation.

## `github.com/agentable/go-config` — Application Configuration

Core configuration library with zero external dependencies.

**Architecture:** Type-safe `Config[T]` with generics, multi-source providers, hot reload, codec-agnostic.

**Module isolation keeps deps minimal:**
```
go-config core                    → 0 external deps
+ codec/json2codec                → +go-json-experiment/json
+ codec/yamlcodec                 → +goccy/go-yaml
+ codec/tomlcodec                 → +pelletier/go-toml/v2
+ provider/file                   → +fsnotify/fsnotify
+ provider/flag                   → +spf13/pflag
+ configsecrets                   → +agentable/go-secrets
```

Users only pull deps for features they use.

**When to use:** Any application needing multi-source config (env + file + flags), hot reload, or type-safe config structs.

### Codec selection

| Format | Library | Why |
|--------|---------|-----|
| JSON (standard) | `encoding/json` | Default, zero deps |
| JSON (high-perf) | `go-json-experiment/json` | 2.5-10x faster, `omitzero` |
| YAML | `goccy/go-yaml` | yaml.v3 **archived 2025-04**; supports `json` tags, better spec (355/402 vs 295/402) |
| TOML | `pelletier/go-toml/v2` | 2-5x faster than BurntSushi, zero deps, `encoding/json`-style API |

## `github.com/caarlos0/env/v11` — Environment Variables

- Zero dependencies, struct tag based: `env:"VAR_NAME"`
- Supports all Go types + `time.Duration`, `url.URL`, pointers, slices, maps
- Tag options: `,expand`, `,file`, `,required`, `,notEmpty`
- Feature-complete (stable, mature)

**When to use:** env-to-struct binding with typed defaults/validation in env-driven apps.

**When NOT to use:** If you only need `.env` loading (use `godotenv`) or need multi-source config (use `agentable/go-config`).

## `github.com/agentable/go-secrets` — Secrets Management

- Plugin-driven architecture
- Zero-config default: file store + AES-256-GCM encryption
- Scope-based isolation

**When to use:** Sensitive config values referenced via `${secret:NAME}` in configuration.

## `github.com/spf13/cobra` — CLI Framework

- Hierarchical subcommands: `app server`, `app migrate`
- Auto-generated help, man pages, shell completions (bash/zsh/fish/PowerShell)
- POSIX flags via pflag
- Used by Kubernetes, Docker, Hugo, GitHub CLI

**When to use:** Building full CLI applications with subcommands.

**When NOT to use:** If you only need CLI flags for config, use `spf13/pflag` directly via go-config/provider/flag.

## Decision Tree

```
Need configuration handling?
├── Only load `.env` into process env?
│   └── joho/godotenv
├── Bind env vars to typed struct with defaults/required?
│   └── caarlos0/env/v11
└── Multi-source config (file + env + flags), optional hot reload?
    └── agentable/go-config
```

## Do NOT Use

| Library | Reason | Use Instead |
|---------|--------|-------------|
| `spf13/viper` | We have go-config | `agentable/go-config` |
| `knadh/koanf/v2` | We have go-config | `agentable/go-config` |
| `gopkg.in/yaml.v3` | **Archived 2025-04** | `github.com/goccy/go-yaml` |
| `BurntSushi/toml` | 2-5x slower | `pelletier/go-toml/v2` |
| `hashicorp/hcl` | YAGNI — Terraform-specific | — |
| `gopkg.in/ini.v1` | YAGNI — INI is legacy | — |

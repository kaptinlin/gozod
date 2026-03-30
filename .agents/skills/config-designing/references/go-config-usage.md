# Pattern: go-config for Applications

## When to Use

`github.com/agentable/go-config` is for **applications** (has `main()`) that assemble library components. Libraries never use go-config directly — they expose Config structs.

## Boundary Rule

```
┌──────────────────────────────────────────────┐
│ Application (main.go)                         │
│                                               │
│  Uses go-config to:                           │
│  - Load from file, env, flags, secrets        │
│  - Compose library Config structs             │
│  - Watch for changes, explain provenance      │
│                                               │
├───────────────────────────────────────────────┤
│ Library Packages                              │
│                                               │
│  Expose Config struct + New(cfg, opts...)      │
│  Never import go-config                       │
│  Never read files or env vars directly        │
│                                               │
└───────────────────────────────────────────────┘
```

## Application Config Struct

Compose library Config types into a single application config:

```go
type AppConfig struct {
    Server struct {
        Host string `json:"host"`
        Port int    `json:"port"`
    } `json:"server"`

    Sandbox struct {
        Policy sandbox.Policy       `json:"policy"`
        Docker docker.Config        `json:"docker"`
        Guard  commandguard.Config  `json:"guard"`
        Filter domainfilter.Config  `json:"filter"`
    } `json:"sandbox"`

    Log struct {
        Level  string `json:"level"`
        Format string `json:"format"`
    } `json:"log"`
}
```

**Key:** Library Config types are embedded directly. go-config handles the mapping from sources to these structs.

## Source Composition

```go
import (
    "github.com/agentable/go-config"
    "github.com/agentable/go-config/provider/static"
    "github.com/agentable/go-config/provider/file"
    "github.com/agentable/go-config/provider/env"
    _ "github.com/agentable/go-config/format/yaml"
)

cfg, err := config.Load[AppConfig]([]config.Source{
    static.New(AppConfig{                           // 1. defaults (lowest)
        Server: struct{ Host string; Port int }{
            Host: "0.0.0.0", Port: 8080,
        },
    }),
    file.New("config.yaml"),                        // 2. config file
    file.New("config.local.yaml",                   // 3. local overrides (optional)
        file.WithPolicy(config.PolicyOptional)),
    env.New(env.WithPrefix("APP_")),                // 4. env vars (highest)
}, config.WithDecodeHook(func(s string) (time.Duration, error) {
    return time.ParseDuration(s)
}))
```

**Precedence:** Later sources override earlier. `APP_SANDBOX_DOCKER_IMAGE=node:20` overrides `config.yaml`.

## Wiring: Config → Library Constructors

```go
func buildSandbox(cfg AppConfig, logger *slog.Logger) (*sandbox.Sandbox, error) {
    sc := cfg.Sandbox

    backend, err := docker.New(sc.Docker, docker.WithLogger(logger))
    if err != nil {
        return nil, err
    }

    return sandbox.New(
        sandbox.WithBackend(backend),
        sandbox.WithPolicy(sc.Policy),
        sandbox.WithPlugin(commandguard.New(sc.Guard)),
        sandbox.WithPlugin(domainfilter.New(sc.Filter, domainfilter.WithLogger(logger))),
        sandbox.WithLogger(logger),
    )
}
```

**Pattern:** Config structs flow from go-config → library constructors. Runtime deps (logger) are wired separately via Options.

## Config File Example

```yaml
server:
  host: 0.0.0.0
  port: 8080

sandbox:
  policy:
    network: deny
    writable_paths: [/tmp, /app/data]
    inherit_env: true
    resources:
      timeout: 30s
      max_memory_mb: 512
  docker:
    image: node:20
    memory_limit: 1g
  guard:
    deny_patterns: ["rm -rf /", "shutdown"]
    detect_level: deep
  filter:
    allowed_domains: ["*.example.com", "api.openai.com"]
```

`network: deny` → `sandbox.NetworkDeny` (via TextUnmarshaler).
`detect_level: deep` → `commandguard.DetectDeep` (via TextUnmarshaler).
`timeout: 30s` → `30 * time.Second` (via decode hook).

## Hot Reload

```go
cfg.OnChange(func(c *config.Config[AppConfig]) {
    v := c.Value()
    guard.UpdateDenyPatterns(v.Sandbox.Guard.DenyPatterns)
    filter.UpdateAllowedDomains(v.Sandbox.Filter.AllowedDomains)
}, "sandbox.guard", "sandbox.filter")

go cfg.Watch(ctx)
```

## When NOT to Use go-config

| Scenario | Use Instead |
|----------|-------------|
| Inside a library package | Expose `Config` struct |
| Simple env-only tool | `caarlos0/env/v11` |
| Only loading `.env` files | `joho/godotenv` |
| CLI with subcommands | `spf13/cobra` (can feed into go-config via flag provider) |
| Unit test configuration | `static.New()` or direct struct literal |

## Decision: go-config vs caarlos0/env

| Need | go-config | caarlos0/env |
|------|-----------|-------------|
| Multi-source (file + env + flags) | Yes | No (env only) |
| Hot reload / watch | Yes | No |
| Explainability / provenance | Yes | No |
| Config file support | Yes (YAML, JSON, TOML) | No |
| Zero external deps (core) | Yes | Yes |
| Minimal env-only binding | Overkill | Perfect fit |

# go-config Integration

Two modes for integrating go-secrets with go-config: **Source mode** (load all secrets into config) and **Interpolation mode** (resolve `${secret:NAME}` references in config files).

## Source Mode

Secrets loaded as a config source, merged into `AppConfig` via the standard source priority chain.

```go
import (
    "github.com/agentable/go-config"
    "github.com/agentable/go-config/provider/env"
    "github.com/agentable/go-config/provider/file"
    "github.com/agentable/go-config/provider/static"
    cfgsecrets "github.com/agentable/go-config/provider/secrets"
    _ "github.com/agentable/go-config/format/yaml"
)

secretsSrc := cfgsecrets.New(s, "prod")

cfg, err := config.Load[AppConfig]([]config.Source{
    static.New(defaults),          // 1. defaults (lowest)
    file.New("config.yaml"),       // 2. config file
    env.New(env.WithPrefix("APP_")), // 3. env vars
    secretsSrc,                    // 4. secrets (highest)
})
```

### How Source Mode Works

1. `cfgsecrets.New(s, "prod")` creates a source bound to scope `"prod"`
2. On `Load`, calls `s.Collect(ctx, "prod")` — retrieves all secrets as `map[string]string`
3. Dot-notation keys become nested maps: `"database.password"` → `{database: {password: "..."}}`
4. Merged into the config map following standard priority (later sources override earlier)

### Secret Naming Convention for Source Mode

Secret names map directly to config struct fields via dot notation:

```go
type AppConfig struct {
    Database struct {
        Password string `json:"password"`
        Host     string `json:"host"`
    } `json:"database"`
    APIKey string `json:"api_key"`
}
```

Set secrets with matching paths:
```go
s.Set(ctx, "prod", "database.password", []byte("s3cret"))
s.Set(ctx, "prod", "database.host", []byte("db.example.com"))
s.Set(ctx, "prod", "api_key", []byte("key456"))
```

Result: `cfg.Value().Database.Password == "s3cret"`

## Interpolation Mode

Config files reference secrets by name using `${secret:NAME}` syntax. Resolved at load time via a `SecretResolver`.

```go
secretsSrc := cfgsecrets.New(s, "prod")

cfg, err := config.Load[AppConfig]([]config.Source{
    static.New(defaults),
    file.New("config.yaml"),
    env.New(env.WithPrefix("APP_")),
}, config.WithSecretResolver(secretsSrc.Resolver()))
```

### Config File with Interpolation

```yaml
database:
  host: db.example.com
  dsn: "postgres://app:${secret:db_password}@${database.host}:5432/mydb"

redis:
  url: "redis://:${secret:redis_password}@redis:6379"

api:
  key: "${secret:api_key}"
```

### How Interpolation Works

1. `secretsSrc.Resolver()` returns a `config.SecretResolverFunc`
2. Config engine detects `${secret:NAME}` patterns in string values
3. For each match, calls `resolver.Resolve(ctx, "NAME")`
4. Resolver calls `s.Get(ctx, scope, name)`, returns `val.ExposeString()`
5. Properly closes the Value to zero memory
6. Missing secrets return empty string (reference left as-is)

### Three Interpolation Types

go-config supports three reference types in the same value:

```yaml
dsn: "postgres://app:${secret:db_password}@${env:DB_HOST}:${database.port}/mydb"
#                       ^^^^^^^^^^^^^^^^     ^^^^^^^^^^^^   ^^^^^^^^^^^^^^^^
#                       secret reference     env var ref    config path ref
```

| Syntax | Resolves To |
|--------|------------|
| `${secret:NAME}` | Secret from go-secrets via SecretResolver |
| `${env:VAR}` | Environment variable |
| `${path.to.key}` | Another config value at the given path |

## Combined Mode

Use both modes together — secrets loaded as a source AND available for interpolation:

```go
secretsSrc := cfgsecrets.New(s, "prod")

cfg, err := config.Load[AppConfig]([]config.Source{
    static.New(defaults),
    file.New("config.yaml"),             // contains ${secret:NAME} references
    env.New(env.WithPrefix("APP_")),
    secretsSrc,                          // source mode: loads all secrets
}, config.WithSecretResolver(secretsSrc.Resolver()))  // interpolation mode
```

When to use each:

| Use Case | Mode |
|----------|------|
| Secret maps directly to a config field | Source |
| Secret embedded in a connection string | Interpolation |
| Secret used as part of a URL template | Interpolation |
| Bulk loading all credentials | Source |
| Mixed: some fields, some templates | Combined |

## Optional Source Policy

By default, secrets source is required — load fails if secrets are unavailable. Use `PolicyOptional` for graceful degradation:

```go
secretsSrc := cfgsecrets.New(s, "prod",
    cfgsecrets.WithPolicy(config.PolicyOptional),
)
```

Check if the source was skipped:
```go
if err := secretsSrc.SkipError(); err != nil {
    logger.Warn("secrets source skipped", slog.Any("error", err))
}
```

## Config Provenance

Trace where a config value came from:

```go
cfg.Explain("database.password")
// → "database.password: secrets://prod (set by secrets source)"
```

This identifies which source provided each value — file, env, or secrets.

## Full Wiring Example

```go
package main

import (
    "context"
    "log"
    "log/slog"
    "os"

    "github.com/agentable/go-config"
    "github.com/agentable/go-config/provider/env"
    "github.com/agentable/go-config/provider/file"
    "github.com/agentable/go-config/provider/static"
    cfgsecrets "github.com/agentable/go-config/provider/secrets"
    _ "github.com/agentable/go-config/format/yaml"

    "github.com/agentable/go-secrets"
    "github.com/agentable/go-secrets/store/filestore"
    "github.com/agentable/go-secrets/cipher/envelope"
    "github.com/agentable/go-secrets/masterkey/envkey"
    "github.com/agentable/go-secrets/plugin/audit"
)

type AppConfig struct {
    Server struct {
        Host string `json:"host"`
        Port int    `json:"port"`
    } `json:"server"`
    Database struct {
        DSN string `json:"dsn"`
    } `json:"database"`
    Redis struct {
        URL string `json:"url"`
    } `json:"redis"`
}

func main() {
    logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

    // 1. Set up go-secrets
    mk, err := envkey.New("")
    if err != nil {
        log.Fatal(err)
    }
    enc, err := envelope.New(mk)
    if err != nil {
        log.Fatal(err)
    }
    store, err := filestore.New("./secrets.enc")
    if err != nil {
        log.Fatal(err)
    }
    s, err := secrets.New(
        secrets.WithStore(store),
        secrets.WithCipher(enc),
        secrets.WithPlugin(audit.New(logger.Handler())),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer s.Close()

    // 2. Create secrets source for go-config
    secretsSrc := cfgsecrets.New(s, "prod")

    // 3. Load config with secrets integration
    cfg, err := config.Load[AppConfig]([]config.Source{
        static.New(AppConfig{
            Server: struct {
                Host string `json:"host"`
                Port int    `json:"port"`
            }{Host: "0.0.0.0", Port: 8080},
        }),
        file.New("config.yaml"),
        env.New(env.WithPrefix("APP_")),
        secretsSrc,
    }, config.WithSecretResolver(secretsSrc.Resolver()))
    if err != nil {
        log.Fatal(err)
    }

    v := cfg.Value()
    logger.Info("config loaded",
        slog.String("server", v.Server.Host),
        slog.Int("port", v.Server.Port),
    )
}
```

Config file (`config.yaml`):
```yaml
server:
  host: 0.0.0.0
  port: 8080

database:
  dsn: "postgres://app:${secret:db_password}@db.example.com:5432/mydb"

redis:
  url: "redis://:${secret:redis_password}@redis:6379"
```

## When NOT to Use go-config Integration

| Scenario | Use Instead |
|----------|------------|
| Simple env-only secrets | `envstore` directly, no go-config |
| Single secret lookup at runtime | `s.Get(ctx, scope, name)` directly |
| Secrets in context (middleware) | `secrets.FromContext(ctx, name)` |
| Unit tests | `memstore` + `WithNoEncryption()` |

---
description: Manage secrets securely in Go applications using agentable/go-secrets with go-config integration. Use when designing secret storage, loading credentials in config, setting up encryption, or integrating secrets into application startup.
name: go-secrets-managing
---


# Go Secrets Management Guide

Securely manage credentials in Go applications using `github.com/agentable/go-secrets` and its integration with `github.com/agentable/go-config`.

## Decision Flowchart

```
How do secrets reach your application?
│
├─ Secrets in encrypted file (committed to repo, git-diff friendly)
│  └─ go-secrets + filestore + envelope cipher
│     └─ Production: envkey master key from env var / KMS
│
├─ Secrets in environment variables (K8s, Docker, CI/CD)
│  └─ go-secrets + envstore (read-only, no encryption needed)
│     └─ Or: go-config env provider directly (simpler)
│
├─ Secrets in external vault (HashiCorp Vault, AWS SM, GCP SM)
│  └─ go-secrets + custom Store implementation
│     └─ Or: go-config secrets provider with custom SecretClient
│
└─ Config file references secrets by name
   └─ go-config interpolation: ${secret:DB_PASSWORD}
      └─ SecretResolver backed by go-secrets
```

## Architecture Overview

```
┌──────────────────────────────────────────────────────┐
│ Application (main.go)                                 │
│                                                       │
│  go-config loads AppConfig from:                      │
│    file → env → secrets (source) → flags              │
│                                                       │
│  Interpolation resolves ${secret:KEY} via resolver    │
├───────────────────────────────────────────────────────┤
│ go-secrets                                            │
│  ┌─────────┐   ┌──────────┐   ┌───────────────┐     │
│  │  Store   │   │  Cipher  │   │  MasterKey    │     │
│  │filestore │   │ envelope │   │ envkey / KMS  │     │
│  │memstore  │   │ aesgcm   │   │ multi (rotate)│     │
│  │envstore  │   └──────────┘   └───────────────┘     │
│  └─────────┘                                          │
├───────────────────────────────────────────────────────┤
│ Plugins: audit (slog), rotation (age monitoring)      │
└──────────────────────────────────────────────────────┘
```

## Production Setup — [details](references/production-setup.md)

Recommended production configuration with envelope encryption and filestore:

```go
import (
    "github.com/agentable/go-secrets"
    "github.com/agentable/go-secrets/store/filestore"
    "github.com/agentable/go-secrets/cipher/envelope"
    "github.com/agentable/go-secrets/masterkey/envkey"
    "github.com/agentable/go-secrets/plugin/audit"
)

mk, err := envkey.New("")  // reads SECRETS_MASTER_KEY
enc, err := envelope.New(mk)
store, err := filestore.New("./secrets.enc")

s, err := secrets.New(
    secrets.WithStore(store),
    secrets.WithCipher(enc),
    secrets.WithPlugin(audit.New(slogHandler)),
)
defer s.Close()
```

**Why envelope encryption:** Per-secret DEK derived via HKDF. Master key wraps IKM, never touches plaintext. Key rotation changes only the wrapping layer.

## go-config Integration — [details](references/config-integration.md)

Two integration modes:

| Mode | How | When |
|------|-----|------|
| **Source** | Secrets loaded as config source, merged into AppConfig | Secrets map directly to config fields |
| **Interpolation** | `${secret:KEY}` in config files resolved via SecretResolver | Secrets referenced inline within other config values |

```go
import cfgsecrets "github.com/agentable/go-config/provider/secrets"

secretsSrc := cfgsecrets.New(s, "prod")

cfg, err := config.Load[AppConfig]([]config.Source{
    static.New(defaults),
    file.New("config.yaml"),
    env.New(env.WithPrefix("APP_")),
    secretsSrc,                                          // source mode
}, config.WithSecretResolver(secretsSrc.Resolver()))     // interpolation mode
```

Config file with interpolation:
```yaml
database:
  host: db.example.com
  dsn: "postgres://app:${secret:db_password}@${database.host}:5432/mydb"
```

## Security Checklist

| Requirement | Solution | Reference |
|-------------|----------|-----------|
| Encrypt secrets at rest | `envelope` cipher + `filestore` | [production-setup.md](references/production-setup.md) |
| Master key not in code | `envkey` reads from env var | [master-key.md](references/master-key.md) |
| Per-secret key isolation | Envelope HKDF derives unique DEK per secret | [encryption-model.md](references/encryption-model.md) |
| Cross-scope replay prevention | AAD binds ciphertext to `scope/name` | [encryption-model.md](references/encryption-model.md) |
| Zero plaintext in memory after use | `Value.Close()` uses `clear()` | [memory-safety.md](references/memory-safety.md) |
| Audit trail | `audit` plugin logs all operations via slog | [production-setup.md](references/production-setup.md) |
| Key rotation without downtime | `multi` master key wraps with old + new | [master-key.md](references/master-key.md) |
| Secret age monitoring | `rotation` plugin triggers callback on stale secrets | [production-setup.md](references/production-setup.md) |
| Config provenance | `cfg.Explain("database.password")` shows source | [config-integration.md](references/config-integration.md) |
| No plaintext in logs | `Value.String()` returns `[REDACTED]` | [memory-safety.md](references/memory-safety.md) |

## Store Selection

| Store | Use Case | Encryption | Write Support |
|-------|----------|-----------|---------------|
| `filestore` | Encrypted file committed to repo | Required (Cipher) | Yes |
| `envstore` | K8s/Docker/CI secrets in env vars | Not needed (platform handles) | Read-only |
| `memstore` | Unit tests | Optional (`WithNoEncryption`) | Yes |
| Custom `Store` | Vault, AWS SM, GCP SM, Azure KV | Depends on provider | Depends |

## Cipher Selection

| Cipher | Use Case | Key Management |
|--------|----------|---------------|
| `envelope` | **Production** — per-secret DEK, master key wrapping | MasterKey interface |
| `aesgcm` | Simple direct encryption, single key for all secrets | Direct 32-byte key |
| `WithNoEncryption()` | **Tests only** — pass-through, no encryption | None |

## Anti-Patterns

| Anti-Pattern | Problem | Fix |
|-------------|---------|-----|
| Hardcoded secrets in source code | Secrets in git history forever | `filestore` + `envelope` encryption |
| `WithNoEncryption()` in production | Plaintext secrets on disk | Always use `envelope` or `aesgcm` |
| Master key in config file | Config file readable = secrets readable | `envkey` from env var or KMS |
| Ignoring `Value.Close()` | Plaintext lingers in memory | `defer val.Close()` immediately after Get |
| Logging `val.ExposeString()` | Plaintext in log files | Use `val.String()` → `[REDACTED]` |
| Same scope for all environments | Prod secrets accessible from dev | Separate scopes: `"prod"`, `"staging"`, `"dev"` |
| Skipping audit in production | No trail of secret access | Always add `audit.New(handler)` plugin |

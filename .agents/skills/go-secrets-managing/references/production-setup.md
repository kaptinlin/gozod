# Production Setup

Complete production configuration with envelope encryption, filestore, audit logging, and secret rotation monitoring.

## Minimal Production Setup

```go
import (
    "log/slog"
    "os"

    "github.com/agentable/go-secrets"
    "github.com/agentable/go-secrets/store/filestore"
    "github.com/agentable/go-secrets/cipher/envelope"
    "github.com/agentable/go-secrets/masterkey/envkey"
    "github.com/agentable/go-secrets/plugin/audit"
)

func newSecrets(logger *slog.Logger) (*secrets.Secrets, error) {
    mk, err := envkey.New("")  // reads SECRETS_MASTER_KEY env var
    if err != nil {
        return nil, fmt.Errorf("master key: %w", err)
    }

    enc, err := envelope.New(mk)
    if err != nil {
        return nil, fmt.Errorf("cipher: %w", err)
    }

    store, err := filestore.New("./secrets.enc")
    if err != nil {
        return nil, fmt.Errorf("store: %w", err)
    }

    return secrets.New(
        secrets.WithStore(store),
        secrets.WithCipher(enc),
        secrets.WithPlugin(audit.New(logger.Handler())),
    )
}
```

## Full Production Setup (with rotation + MAC)

```go
import (
    "context"
    "log/slog"
    "time"

    "github.com/agentable/go-secrets"
    "github.com/agentable/go-secrets/store/filestore"
    "github.com/agentable/go-secrets/cipher/envelope"
    "github.com/agentable/go-secrets/masterkey/envkey"
    "github.com/agentable/go-secrets/plugin/audit"
    "github.com/agentable/go-secrets/plugin/rotation"
)

func newSecrets(logger *slog.Logger) (*secrets.Secrets, error) {
    mk, err := envkey.New("")
    if err != nil {
        return nil, fmt.Errorf("master key: %w", err)
    }

    enc, err := envelope.New(mk)
    if err != nil {
        return nil, fmt.Errorf("cipher: %w", err)
    }

    macKey := []byte("...32-byte-mac-key...")  // from env or KMS
    store, err := filestore.New("./secrets.enc",
        filestore.WithMACKey(macKey),  // HMAC-SHA256 file integrity
    )
    if err != nil {
        return nil, fmt.Errorf("store: %w", err)
    }

    rotationPlugin, err := rotation.New(
        90*24*time.Hour,  // 90-day max age
        func(ctx context.Context, scope, name string, meta *secrets.Meta) error {
            logger.WarnContext(ctx, "secret needs rotation",
                slog.String("scope", scope),
                slog.String("name", name),
                slog.Time("last_updated", meta.UpdatedAt),
            )
            return nil
        },
    )
    if err != nil {
        return nil, fmt.Errorf("rotation plugin: %w", err)
    }

    return secrets.New(
        secrets.WithStore(store),
        secrets.WithCipher(enc),
        secrets.WithPlugin(audit.New(logger.Handler())),
        secrets.WithPlugin(rotationPlugin),
    )
}
```

## FileStore Details

SOPS-style encrypted JSON file, safe to commit to version control.

```json
{
    "entries": {
        "prod/db_password": {
            "value": "base64-encoded-ciphertext",
            "meta": {
                "created_at": "2025-01-15T10:00:00Z",
                "updated_at": "2025-01-15T10:00:00Z",
                "created_by": "deploy-bot",
                "version": 1
            }
        }
    },
    "_mac": "base64-hmac-sha256"
}
```

Key characteristics:
- **Atomic writes**: temp file → fsync → rename (no partial writes)
- **Git-diff friendly**: scope/name keys in plaintext, values encrypted independently
- **Path traversal protection**: uses `os.OpenRoot` (Go 1.24+)
- **Optional MAC**: HMAC-SHA256 over sorted keys and metadata for tamper detection
- **Thread-safe**: `sync.RWMutex` for concurrent access

```go
// FileStore constructor
store, err := filestore.New("./secrets.enc")

// With MAC integrity protection (32-byte key required)
store, err := filestore.New("./secrets.enc",
    filestore.WithMACKey(macKey),
)
```

## Audit Plugin

Decorates the Store to log all secret operations via `slog.Handler`.

```go
auditPlugin := audit.New(logger.Handler())
// Returns nil if handler is nil (safe to use unconditionally)
```

Logged operations and fields:
| Operation | Fields |
|-----------|--------|
| `Get` | action, scope, name, actor (from context), error |
| `Set` | action, scope, name, actor, error |
| `Delete` | action, scope, name, actor, error |
| `List` | action, scope, actor |

Attach actor identity to context:
```go
ctx = audit.WithActor(ctx, "deploy-bot")
val, err := s.Get(ctx, "prod", "db_password")
// Audit log includes: actor="deploy-bot"
```

Audit failures never block operations — the audit plugin silently ignores logging errors.

## Rotation Plugin

Monitors secret age on `Get` and invokes a callback when secrets exceed a configured max age.

```go
rotationPlugin, err := rotation.New(
    90*24*time.Hour,  // max age before triggering callback
    func(ctx context.Context, scope, name string, meta *secrets.Meta) error {
        // Send alert, trigger rotation pipeline, etc.
        alerting.Send("secret %s/%s needs rotation", scope, name)
        return nil
    },
)
```

Key characteristics:
- Checks age on every `Get` call (uses `meta.UpdatedAt`)
- Callback errors do not block the `Get` operation
- Callback receives full metadata for decision-making
- Uses `WithNowFunc` option for testing time-dependent behavior

## Store Read/Write Operations

```go
ctx := context.Background()

// Set with metadata
err := s.Set(ctx, "prod", "db_password", []byte("s3cret"),
    secrets.WithMeta(&secrets.Meta{
        Labels:    map[string]string{"team": "platform"},
        ExpiresAt: time.Now().Add(90 * 24 * time.Hour),
        CreatedBy: "deploy-bot",
    }),
)

// Get (always defer Close)
val, err := s.Get(ctx, "prod", "db_password")
if err != nil {
    return err
}
defer val.Close()
dsn := fmt.Sprintf("postgres://app:%s@db:5432/mydb", val.ExposeString())

// Check existence without decryption
exists, err := s.Exists(ctx, "prod", "db_password")

// Delete
err = s.Delete(ctx, "prod", "db_password")

// List (iterator)
for ref, err := range s.List(ctx, "prod") {
    if err != nil {
        return err
    }
    fmt.Printf("%s/%s (version %d)\n", ref.Scope, ref.Name, ref.Meta.Version)
}

// Collect all secrets as map[string]string
all, err := s.Collect(ctx, "prod")
// all = {"db_password": "s3cret", "api_key": "key123"}
```

## Environment Setup

```bash
# Generate a master key
go run -mod=mod github.com/agentable/go-secrets/cmd/secrets keygen

# Set the master key (base64-encoded 32 bytes)
export SECRETS_MASTER_KEY="base64-encoded-32-byte-key"

# In Kubernetes
kubectl create secret generic app-secrets \
    --from-literal=SECRETS_MASTER_KEY="$(cat master.key)"

# In Docker
docker run -e SECRETS_MASTER_KEY="..." myapp
```

## Deployment Checklist

| Item | How |
|------|-----|
| Master key from env var, not file | `envkey.New("")` reads `SECRETS_MASTER_KEY` |
| Encrypted secrets file in repo | `filestore.New("./secrets.enc")` with envelope cipher |
| MAC integrity check | `filestore.WithMACKey(key)` for tamper detection |
| Audit logging enabled | `audit.New(logger.Handler())` |
| Rotation monitoring | `rotation.New(maxAge, callback)` |
| Separate scopes per environment | `"prod"`, `"staging"`, `"dev"` |
| Value.Close() called | `defer val.Close()` after every Get |
| No plaintext in logs | Use `val.String()` (returns `[REDACTED]`) |

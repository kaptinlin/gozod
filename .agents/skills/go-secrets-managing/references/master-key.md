# Master Key Management

The MasterKey interface wraps and unwraps per-secret encryption keys. Production deployments use `envkey` (reads from environment variable) or `multi` (key rotation).

## MasterKey Interface

```go
type MasterKey interface {
    Encrypt(dataKey []byte) ([]byte, error)   // wrap a DEK
    Decrypt(encryptedKey []byte) ([]byte, error) // unwrap a DEK
    NeedsRotation() bool                      // signal rotation needed
    ID() string                               // unique identifier
}
```

## envkey — Environment Variable Master Key

Reads a base64-encoded 32-byte AES-256 key from an environment variable.

```go
import "github.com/agentable/go-secrets/masterkey/envkey"

// Default: reads SECRETS_MASTER_KEY
mk, err := envkey.New("")

// Custom env var name
mk, err := envkey.New("MY_APP_MASTER_KEY")
```

### How It Works

1. Constructor reads the environment variable
2. Base64-decodes to get 32 raw bytes
3. Creates AES-256-GCM AEAD from the key
4. Key material cleared from memory after AEAD creation (`clear()`)
5. All subsequent Encrypt/Decrypt use the AEAD

### Generating a Master Key

```go
key := envkey.GenerateKey()
// Returns: base64-encoded 32 cryptographically random bytes
// Example: "dGhpcyBpcyBhIHRlc3Qga2V5IGZvciBkZW1v..."
```

```bash
# Set in environment
export SECRETS_MASTER_KEY="$(go run github.com/agentable/go-secrets/cmd/secrets keygen)"

# Kubernetes secret
kubectl create secret generic app-master-key \
    --from-literal=SECRETS_MASTER_KEY="$(go run github.com/agentable/go-secrets/cmd/secrets keygen)"

# Docker
docker run -e SECRETS_MASTER_KEY="..." myapp
```

### Error Cases

| Error | Cause |
|-------|-------|
| `ErrEnvNotSet` | Environment variable not set |
| `ErrInvalidKey` | Not valid base64 or not exactly 32 bytes |

### Properties

- `NeedsRotation()` → always `false` (rotation handled externally)
- `ID()` → `"envkey:SECRETS_MASTER_KEY"` (or custom var name)

## multi — Key Rotation

Wraps multiple MasterKeys for zero-downtime key rotation and cross-region deployments.

```go
import "github.com/agentable/go-secrets/masterkey/multi"

mk, err := multi.New(newKey, oldKey)
```

### How It Works

**Encrypt:** wraps the DEK with ALL underlying keys. Wire format:
```
[1-byte count][2-byte len₁][wrapped₁][2-byte len₂][wrapped₂]...
```

**Decrypt:** tries each wrapped DEK in order. First successful decryption wins.

This design ensures:
- Secrets encrypted with old key can still be decrypted
- Secrets encrypted with new key are wrapped by both keys
- Re-encryption with new key happens naturally on next `Set`

### Key Rotation Workflow

```
Step 1: Deploy with old key only
    mk := envkey.New("OLD_KEY")

Step 2: Generate new key, deploy with both
    old := envkey.New("OLD_KEY")
    new := envkey.New("NEW_KEY")
    mk := multi.New(new, old)  // new key first = primary

Step 3: Re-encrypt all secrets (reads with old, writes with new+old)
    for ref, err := range s.List(ctx, "prod") {
        val, _ := s.Get(ctx, ref.Scope, ref.Name)
        s.Set(ctx, ref.Scope, ref.Name, val.Expose())
        val.Close()
    }

Step 4: Remove old key from deployment
    mk := envkey.New("NEW_KEY")
```

### Properties

- `NeedsRotation()` → `true` if any underlying key returns `true`
- `ID()` → `"multi:[envkey:NEW_KEY,envkey:OLD_KEY]"`
- Minimum 1 key, maximum 255 keys
- Returns `ErrNoKeys` if no keys provided
- Returns `ErrTooMany` if more than 255 keys
- Returns `ErrAllFailed` (with joined errors) if all decryption attempts fail

## Custom MasterKey (KMS)

Implement the `MasterKey` interface for cloud KMS integration:

```go
type kmsMasterKey struct {
    client    kmsClient
    keyID     string
    rotateAt  time.Time
}

func (k *kmsMasterKey) Encrypt(dataKey []byte) ([]byte, error) {
    return k.client.Encrypt(context.Background(), k.keyID, dataKey)
}

func (k *kmsMasterKey) Decrypt(encryptedKey []byte) ([]byte, error) {
    return k.client.Decrypt(context.Background(), k.keyID, encryptedKey)
}

func (k *kmsMasterKey) NeedsRotation() bool {
    return time.Now().After(k.rotateAt)
}

func (k *kmsMasterKey) ID() string {
    return fmt.Sprintf("kms:%s", k.keyID)
}
```

Use with envelope cipher:
```go
mk := &kmsMasterKey{client: awsKMS, keyID: "arn:aws:kms:..."}
enc, err := envelope.New(mk)
```

## Where the Master Key Lives

| Platform | Method | Notes |
|----------|--------|-------|
| Local dev | `.env` file (gitignored) | `SECRETS_MASTER_KEY=...` |
| Kubernetes | Secret resource | Mounted as env var in pod spec |
| Docker | `-e` flag or Docker secrets | `docker run -e SECRETS_MASTER_KEY=...` |
| AWS | Parameter Store / Secrets Manager | Injected via sidecar or init container |
| GCP | Secret Manager | Accessed via workload identity |
| CI/CD | Pipeline secret variables | GitHub Actions secrets, GitLab CI vars |

## Security Rules

- Master key MUST come from environment variable or KMS — never from config files
- Master key MUST be 32 bytes (AES-256)
- Master key is cleared from memory after AEAD construction
- Each environment (prod, staging, dev) MUST use a different master key
- Rotation requires deploying both old and new keys simultaneously via `multi`

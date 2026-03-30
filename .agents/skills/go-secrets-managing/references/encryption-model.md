# Encryption Model

go-secrets provides two cipher implementations: `envelope` (production) and `aesgcm` (simple). Both implement the `Cipher` interface with AAD (Additional Authenticated Data) binding.

## Cipher Interface

```go
type Cipher interface {
    Encrypt(plaintext, aad []byte) ([]byte, error)
    Decrypt(ciphertext, aad []byte) ([]byte, error)
}
```

**AAD** = `"scope/name"` (e.g., `"prod/db_password"`). Binds ciphertext to its storage key, preventing cross-scope replay attacks.

## Envelope Encryption (Production)

Per-secret random DEK (Data Encryption Key) wrapped by a MasterKey. The master key never touches plaintext data.

### Encryption Flow

```
plaintext + AAD ("prod/db_password")
        │
        ▼
1. Generate random 32-byte IKM (Input Key Material)
        │
        ▼
2. Derive DEK via HKDF-SHA256
   DEK = HKDF(IKM, salt=AAD, info="secrets-envelope-dek", len=32)
        │
        ▼
3. Encrypt plaintext with AES-256-GCM(DEK, AAD)
   ciphertext = AES-GCM-Encrypt(DEK, plaintext, AAD)
        │
        ▼
4. Wrap IKM with MasterKey
   wrappedIKM = MasterKey.Encrypt(IKM)
        │
        ▼
5. Clear IKM and DEK from memory
        │
        ▼
6. Return wire format
```

### Decryption Flow

```
wire format + AAD ("prod/db_password")
        │
        ▼
1. Parse wire format → wrappedIKM + ciphertext
        │
        ▼
2. Unwrap IKM via MasterKey
   IKM = MasterKey.Decrypt(wrappedIKM)
        │
        ▼
3. Re-derive DEK via HKDF-SHA256
   DEK = HKDF(IKM, salt=AAD, info="secrets-envelope-dek", len=32)
        │
        ▼
4. Decrypt with AES-256-GCM(DEK, AAD)
   plaintext = AES-GCM-Decrypt(DEK, ciphertext, AAD)
        │
        ▼
5. Clear IKM and DEK from memory
        │
        ▼
6. Return plaintext
```

### Wire Format

```
[2-byte BE wrappedIKM length][wrappedIKM][AES-GCM ciphertext]
 └── uint16 big-endian ──┘   └── var ──┘  └── var length ──┘
```

Constants:
```go
const (
    dekSize    = 32    // AES-256 key size
    ikmSize    = 32    // Input Key Material for HKDF
    lenSize    = 2     // uint16 big-endian prefix
    maxWrapped = 1024  // sanity bound on wrapped IKM size
    hkdfInfo   = "secrets-envelope-dek"
)
```

### Why Envelope Encryption

| Property | Benefit |
|----------|---------|
| Per-secret random IKM | Compromise of one secret's DEK doesn't affect others |
| HKDF key derivation | DEK is cryptographically bound to scope/name via salt=AAD |
| Master key wraps IKM, not plaintext | Master key never touches sensitive data directly |
| Key rotation changes wrapping only | Re-encrypt IKM wrapper without re-deriving DEK |
| AAD in both HKDF and AES-GCM | Double binding prevents cross-key replay attacks |

### Cross-Scope Replay Prevention

AAD serves dual purpose:
1. **HKDF salt**: Different scope/name → different DEK (even with same IKM)
2. **AES-GCM AAD**: Decryption fails if ciphertext moved to different scope/name

Example: copying `prod/db_password` ciphertext to `dev/db_password` fails because:
- HKDF derives different DEK (salt differs)
- AES-GCM authentication check fails (AAD mismatch)

### Constructor

```go
import "github.com/agentable/go-secrets/cipher/envelope"

enc, err := envelope.New(mk)  // mk implements MasterKey interface
```

- Validates MasterKey lazily on first operation (`sync.OnceValue`)
- Returns `ErrMasterKeyRequired` if mk is nil
- Returns `ErrMalformed` on invalid wire format during decryption
- Returns `ErrWrappedIKMTooLarge` if wrapped IKM exceeds 1024 bytes

### Cryptographic Primitives

| Primitive | Usage |
|-----------|-------|
| `crypto/rand` | Generate random IKM (32 bytes) |
| `crypto/hkdf` | Derive DEK from IKM + AAD salt |
| `crypto/aes` + `cipher.NewGCMWithRandomNonce` | AES-256-GCM encryption |
| `clear()` | Zero IKM and DEK after use |

## AES-256-GCM (Simple)

Direct encryption with a single 32-byte key. All secrets share the same key.

```go
import "github.com/agentable/go-secrets/cipher/aesgcm"

key := make([]byte, 32)  // 32 bytes for AES-256
crypto.Read(key)

enc, err := aesgcm.New(key)
```

### How It Works

1. Single AES-256-GCM AEAD created at construction
2. Uses `crypto/cipher.NewGCMWithRandomNonce` (Go 1.24+) for automatic nonce management
3. AAD bound to scope/name for replay prevention
4. Same key encrypts all secrets

### When to Use

- Development environments where envelope complexity isn't needed
- Single-key scenarios without rotation requirements
- When the key is already managed by an external system

### Error Cases

| Error | Cause |
|-------|-------|
| `ErrInvalidKeySize` | Key is not exactly 32 bytes |
| `ErrDecryptFailed` | Decryption failed (wrong key, corrupted, or AAD mismatch) |

## WithNoEncryption (Tests Only)

Pass-through cipher that clones bytes without encryption. Use only in tests.

```go
s, err := secrets.New(
    secrets.WithStore(memstore.New()),
    secrets.WithNoEncryption(),  // pass-through, no encryption
)
```

Internally implements Cipher as:
```go
type noEncryptionCipher struct{}

func (noEncryptionCipher) Encrypt(plaintext, _ []byte) ([]byte, error) {
    return bytes.Clone(plaintext), nil
}

func (noEncryptionCipher) Decrypt(ciphertext, _ []byte) ([]byte, error) {
    return bytes.Clone(ciphertext), nil
}
```

## Cipher Selection Guide

```
Is this production?
│
├─ Yes → Do you need key rotation?
│        │
│        ├─ Yes → envelope + multi master key
│        └─ No  → envelope + envkey
│
└─ No → Is this a test?
         │
         ├─ Yes → WithNoEncryption() + memstore
         └─ No (dev) → aesgcm or envelope + envkey
```

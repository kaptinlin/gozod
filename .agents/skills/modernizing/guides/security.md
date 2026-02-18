# Security & Crypto

Modern Go strengthens security defaults and adds new cryptographic primitives: sandboxed file access, secure token generation, post-quantum key exchange, and hybrid encryption.

## Contents
- os.Root sandboxed filesystem (1.24+)
- crypto/rand.Text (1.24+)
- crypto/cipher.NewGCMWithRandomNonce (1.24+)
- crypto/mlkem post-quantum KEM (1.24+)
- crypto/hpke hybrid encryption (1.26+)
- errors.AsType for secure error handling (1.26+)
- encoding.TextAppender / BinaryAppender (1.24+)
- weak package (1.24+)

---

## os.Root — sandboxed filesystem (Go 1.24+)

Restricts file operations to a specific directory. Prevents path traversal attacks (`../../../etc/passwd`).

### When to use
- Any server that accesses user-specified file paths
- Sandboxed plugin or template systems
- Processing uploaded files within a restricted directory
- Replacing manual path sanitization (`filepath.Clean`, `filepath.Rel` checks)

### When NOT to use
- When you need access to arbitrary filesystem paths (admin tools, CLI utilities)
- If the directory doesn't exist yet — `OpenRoot` requires an existing directory

```go
// Old — manual path sanitization (error-prone)
func serveFile(base, userPath string) error {
    clean := filepath.Clean(userPath)
    full := filepath.Join(base, clean)
    if !strings.HasPrefix(full, base) {
        return errors.New("path traversal detected")
    }
    data, err := os.ReadFile(full)
    // ...
}

// New (Go 1.24+) — sandboxed by design
root, err := os.OpenRoot("/var/app/data")
if err != nil { return err }
defer root.Close()

f, err := root.Open(userPath) // cannot escape /var/app/data
if err != nil { return err }
defer f.Close()

// One-liner convenience
f, err := os.OpenInRoot("/var/app/data", userPath)

// Go 1.25+ adds more methods
data, err := root.ReadFile("config.json")
root.WriteFile("output.txt", data, 0o644)
root.MkdirAll("subdir/nested", 0o755)
```

**Rule of thumb**: if your code does `os.Open(filepath.Join(baseDir, userInput))`, replace with `os.OpenInRoot(baseDir, userInput)`.

**Pre-validation**: use `filepath.IsLocal(path)` (Go 1.20+) as a quick check — rejects absolute paths, `..`, and reserved names.

---

## crypto/rand.Text (Go 1.24+)

Generates cryptographically random text strings. Replaces the common pattern of `rand.Read` + base64 encoding.

### When to use
- API tokens, session IDs, CSRF tokens, nonces
- Any random string for authentication/authorization

### When NOT to use
- When you need raw random bytes — use `crypto/rand.Read`
- When you need a specific character set or length — `rand.Text` uses a fixed base32-like alphabet

```go
// Old
b := make([]byte, 32)
crypto_rand.Read(b)
token := base64.URLEncoding.EncodeToString(b)

// New (Go 1.24+)
token := rand.Text()
```

---

## crypto/cipher.NewGCMWithRandomNonce (Go 1.24+)

AES-GCM with automatically generated random nonces. Eliminates the #1 AES-GCM footgun (nonce reuse).

### When to use
- Encrypting data at rest (files, database fields)
- Any AES-GCM encryption where you don't need to control the nonce

### When NOT to use
- When you need deterministic encryption (same plaintext → same ciphertext)
- When you must control the nonce for protocol compliance
- When the nonce comes from an external source (TLS, protocol-specific)

```go
// Old — manual nonce management (easy to get wrong)
block, _ := aes.NewCipher(key)
gcm, _ := cipher.NewGCM(block)
nonce := make([]byte, gcm.NonceSize())
io.ReadFull(crypto_rand.Reader, nonce)
ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

// New (Go 1.24+) — nonce handled automatically
block, _ := aes.NewCipher(key)
gcm, _ := cipher.NewGCMWithRandomNonce(block)
ciphertext := gcm.Seal(nil, nil, plaintext, nil) // nonce prepended automatically
```

---

## crypto/mlkem — post-quantum KEM (Go 1.24+)

ML-KEM (FIPS 203) key encapsulation mechanism. Post-quantum secure key exchange.

### When to use
- Forward-looking crypto that needs post-quantum resistance
- When compliance requires FIPS 203 (ML-KEM)

### When NOT to use
- Unless your threat model includes quantum attacks or compliance requires it
- For general-purpose encryption — use standard AES-GCM or HPKE

```go
// Generate key pair
dk, err := mlkem.GenerateKey768()
ek := dk.EncapsulationKey()

// Encapsulate (sender)
sharedKey, ciphertext := ek.Encapsulate()

// Decapsulate (receiver)
sharedKey2, err := dk.Decapsulate(ciphertext)
```

---

## crypto/hpke — hybrid encryption (Go 1.26+)

Hybrid Public Key Encryption (RFC 9180) with post-quantum hybrid KEM support.

### When to use
- Encrypting messages to a public key without a pre-shared secret
- Replacing custom RSA/ECDH + AES envelope encryption schemes
- When you need authenticated encryption with associated data (AEAD) and key encapsulation in one API

### When NOT to use
- Simple symmetric encryption — use AES-GCM directly
- When you already have a shared secret — use HKDF + AES-GCM

---

## crypto/rand.Read never fails (Go 1.24+)

`crypto/rand.Read` is now guaranteed to never return an error. The error return is always nil.

### Implication
- You can safely ignore the error: `crypto_rand.Read(b)` without checking err
- **Breaking change**: if you override `crypto/rand.Reader` with a custom reader that fails, the program **crashes**

```go
// Old
b := make([]byte, 32)
if _, err := crypto_rand.Read(b); err != nil {
    return err // this was always dead code in practice
}

// New (Go 1.24+) — error is always nil
b := make([]byte, 32)
crypto_rand.Read(b)
```

---

## encoding.TextAppender / BinaryAppender (Go 1.24+)

New interfaces that append to an existing byte slice instead of allocating a new one. Implemented by many stdlib types.

### When to use
- High-throughput encoding paths where allocation matters
- Building formatted output incrementally

### When NOT to use
- Normal code where allocation is not a bottleneck

```go
// Types implementing TextAppender: net.IP, time.Time, etc.
var buf []byte
buf, err := ip.AppendText(buf)
buf, err = ts.AppendText(buf)
```

---

## omitzero struct tag (Go 1.24+)

`json:"field,omitzero"` omits fields whose value is the zero value for their type. Works correctly with `time.Time` (unlike `omitempty`).

### omitzero vs omitempty

| Value | `omitempty` | `omitzero` |
|-------|-----------|----------|
| `time.Time{}` | NOT omitted | Omitted |
| Empty struct `Foo{}` | NOT omitted | Omitted |
| `nil` slice/map | Omitted | Omitted |
| Empty `[]int{}` | Omitted | NOT omitted |
| `false`, `0`, `""` | Omitted | Omitted |

### When to use
- `time.Time` fields — `omitempty` never omits them
- Nested struct fields — `omitempty` emits `{}` for empty structs
- Types with `IsZero() bool` method — `omitzero` calls it
- Combine both for maximum omission: `json:"field,omitempty,omitzero"`

### When NOT to use
- When you want empty (non-nil) slices omitted — use `omitempty`
- Targeting Go < 1.24 — `omitzero` is not recognized

```go
// Old — broken for time.Time
type Event struct {
    CreatedAt time.Time `json:"created_at,omitempty"` // always emitted!
}

// New (Go 1.24+)
type Event struct {
    CreatedAt time.Time `json:"created_at,omitzero"` // omitted when zero
    Tags      []string  `json:"tags,omitempty"`       // omit empty/nil slice
    Location  Address   `json:"location,omitzero"`    // omit zero struct
}
```

---

## Migration strategy

1. Replace manual path sanitization with `os.Root` for user-facing file access
2. Replace `rand.Read` + `base64.Encode` with `rand.Text()` for token generation
3. Replace manual nonce management with `NewGCMWithRandomNonce`
4. Replace `omitempty` with `omitzero` for `time.Time` and struct fields in JSON
5. Remove error checks on `crypto/rand.Read` (always nil since Go 1.24)

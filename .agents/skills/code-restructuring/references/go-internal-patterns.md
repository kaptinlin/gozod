# Go Internal Patterns

How and when to use `internal/` packages in Go 1.26+. Compiler-enforced visibility for API boundary control.

## internal/ Package Rules

The Go toolchain enforces a hard rule: code in `internal/` can only be imported by code in the parent directory tree.

```
mymodule/
├── internal/
│   └── secret/
│       └── secret.go        # Can be imported by anything in mymodule/
├── pkg/
│   └── client/
│       └── client.go        # CAN import internal/secret (same module root)
└── cmd/
    └── app/
        └── main.go          # CAN import internal/secret (same module root)
```

```
othermodule/
└── main.go                  # CANNOT import mymodule/internal/secret
                             # Compiler error: use of internal package not allowed
```

Nested `internal/` scopes visibility further:

```
mymodule/
├── pkg/
│   ├── auth/
│   │   ├── internal/
│   │   │   └── hasher/
│   │   │       └── bcrypt.go  # Only pkg/auth/ and its children can import
│   │   ├── auth.go            # CAN import auth/internal/hasher
│   │   └── token.go           # CAN import auth/internal/hasher
│   └── user/
│       └── user.go            # CANNOT import auth/internal/hasher
└── cmd/
    └── app/
        └── main.go            # CANNOT import auth/internal/hasher
```

## When to Use internal/

### 1. Hide Implementation from Library Consumers

```
Before:
mylib/
├── parser.go                # Public API
├── optimizer.go             # Public -- but consumers shouldn't use directly
└── codegen.go               # Public -- but consumers shouldn't use directly

After:
mylib/
├── parser.go                # Public API: Parse(), Format()
└── internal/
    ├── optimizer/
    │   └── optimizer.go     # Hidden: consumers can't import
    └── codegen/
        └── codegen.go       # Hidden: consumers can't import
```

Why: Consumers depend only on `parser.go`. You can refactor `optimizer` and `codegen` freely without breaking anyone.

### 2. Share Code Across Packages Without Exposing It

```
Before:
internal/
├── user/
│   └── validate.go          # Duplicated validation logic
├── order/
│   └── validate.go          # Same validation, copy-pasted
└── product/
    └── validate.go          # Same validation, copy-pasted

After:
internal/
├── validate/                # Shared validation, not public
│   └── rules.go
├── user/
│   └── service.go           # imports internal/validate
├── order/
│   └── service.go           # imports internal/validate
└── product/
    └── service.go           # imports internal/validate
```

Why: DRY without exposing validation rules to external consumers.

### 3. Prevent Import of Unstable Experimental Code

```
internal/
├── stable/
│   └── api.go               # Stable, well-tested
└── experimental/
    └── newfeature.go         # Not ready for public use
```

Even within a large org, `internal/experimental/` signals "do not depend on this."

## When NOT to Use internal/

### 1. Single-Module Projects -- Just Unexport

```
# DON'T do this:
myapp/
├── internal/
│   └── config/
│       └── config.go        # Only used by cmd/myapp

# DO this:
myapp/
├── config/
│   └── config.go            # unexported functions are already private
```

If you have one module and one binary, `internal/` adds directory depth with no practical benefit. Use unexported identifiers (`lowercase`) instead.

### 2. When You Might Export Later

```
# DON'T start here:
mylib/
└── internal/
    └── parser/
        └── parser.go        # "Maybe we'll make this public someday"

# DO start here:
mylib/
└── parser/
    └── parser.go            # Public from day one, or keep unexported symbols
```

Moving from `internal/` to public is a breaking change for your internal structure. If there's a reasonable chance of exporting, start public with unexported symbols.

### 3. Excessive internal/ Nesting

```
# DON'T do this:
pkg/auth/
├── internal/
│   └── token/
│       └── internal/
│           └── encoder/
│               └── base64.go   # 4 levels deep, unnavigable

# DO this:
pkg/auth/
├── internal/
│   └── tokenenc/
│       └── base64.go           # Flat internal, clear purpose
```

One level of `internal/` per package boundary is enough.

## internal/ Directory Placement

### Project-Level internal/ (Most Common)

```
myproject/
├── cmd/
│   ├── api/main.go
│   └── worker/main.go
├── internal/                # Shared across all cmd/ binaries
│   ├── auth/
│   ├── user/
│   ├── order/
│   └── platform/
│       ├── database.go
│       └── logger.go
└── go.mod
```

Use when: Multiple binaries share application code that should not be imported by other modules.

### Package-Level internal/ (Library Pattern)

```
mylib/
├── client.go                # Public API
├── client_test.go
├── internal/                # Hidden from library consumers
│   ├── pool/
│   │   └── connection.go
│   └── codec/
│       └── binary.go
└── go.mod
```

Use when: A public library needs hidden implementation details.

### Sub-Package internal/ (Rare, Use Sparingly)

```
internal/
├── auth/
│   ├── auth.go
│   ├── token.go
│   └── internal/            # Only auth/ can import
│       └── hasher/
│           └── bcrypt.go
└── user/
    └── service.go           # CANNOT import auth/internal/hasher
```

Use when: A package has implementation details that even sibling packages should not access. Rare in practice.

## API Boundary Design

### Public API in Root, Implementation in internal/

```
mylib/
├── client.go                # Public: NewClient(), Client.Do()
├── option.go                # Public: WithTimeout(), WithRetry()
├── errors.go                # Public: ErrTimeout, ErrNotFound
├── internal/
│   ├── transport/
│   │   ├── http.go          # Hidden: HTTP transport details
│   │   └── retry.go         # Hidden: retry logic
│   ├── pool/
│   │   └── pool.go          # Hidden: connection pooling
│   └── codec/
│       ├── json.go          # Hidden: serialization
│       └── proto.go         # Hidden: serialization
├── doc.go
└── go.mod
```

**Public surface:** `client.go`, `option.go`, `errors.go` -- 3 files consumers need to read.
**Implementation:** `internal/` -- 5 files consumers never see.

### Layered API with internal/

```
mylib/
├── client.go                # High-level API
├── advanced/                # Public sub-package for power users
│   └── batch.go
├── internal/
│   └── engine/              # Shared engine used by both client and advanced
│       ├── engine.go
│       └── scheduler.go
└── go.mod
```

Both `client.go` and `advanced/` import `internal/engine/`. External consumers import only `mylib` or `mylib/advanced`.

## Migration Patterns

### Exporting: internal/ to Public

When you decide to make internal code public:

```
Step 1 (before):
mylib/
└── internal/
    └── parser/
        └── parser.go        # func Parse(input string) (*AST, error)

Step 2 (create public wrapper):
mylib/
├── parser/
│   └── parser.go            # Public: calls internal, re-exports types
└── internal/
    └── parser/
        └── parser.go        # Still exists, still works

Step 3 (migrate and remove internal):
mylib/
└── parser/
    └── parser.go            # All code moved here, internal/ removed
```

**Key:** Add the public package first, delegate to internal, then migrate. Never break existing internal consumers.

### Hiding: Public to internal/ (Deprecation)

When you need to hide previously public code:

```
Step 1 (before):
mylib/
└── codec/
    └── binary.go            # Public, but consumers shouldn't use it

Step 2 (deprecate):
mylib/
├── codec/
│   └── binary.go            # Add: // Deprecated: Use mylib.Encode instead.
└── internal/
    └── codec/
        └── binary.go        # Copy implementation here

Step 3 (next major version):
mylib/  # v2
└── internal/
    └── codec/
        └── binary.go        # Public codec/ removed in v2
```

**Key:** Deprecate first, provide alternatives, remove in next major version. Follow semver.

### Restructuring internal/ Itself

When `internal/` becomes a flat mega-package:

```
Before:
internal/
├── auth.go
├── cache.go
├── config.go
├── database.go
├── email.go
├── logger.go
├── metrics.go
├── queue.go
├── ... (20 more files)

After:
internal/
├── auth/
│   └── auth.go
├── platform/                # Infrastructure concerns
│   ├── database.go
│   ├── cache.go
│   ├── queue.go
│   └── logger.go
├── notify/                  # Notification concerns
│   └── email.go
├── observe/                 # Observability
│   └── metrics.go
└── config/
    └── config.go
```

Safe to restructure because `internal/` packages are only imported within the module -- no external consumers to break.

## Decision Checklist

Before placing code in `internal/`:

```
1. Will external modules import this?
   YES -> pkg/ or module root    NO -> internal/ is appropriate

2. Is this a single-module, single-binary project?
   YES -> probably don't need internal/, use unexported symbols
   NO  -> internal/ protects cross-module boundaries

3. Might this become public later?
   YES -> start public with unexported symbols, export when ready
   NO  -> internal/ is safe

4. Is this shared across multiple packages in the module?
   YES -> project-level internal/
   NO  -> package-level internal/ or keep in the package itself

5. Am I nesting internal/ more than one level?
   YES -> flatten it
   NO  -> fine
```

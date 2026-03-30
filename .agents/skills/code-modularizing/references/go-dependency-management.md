# Go Dependency Management

Dependency management best practices for Go 1.26+ modules.

## go.mod Best Practices

### Require Directives

**❌ Pinning to Exact Commits**
```go
require github.com/org/cache v0.0.0-20260101120000-abc123def456
```

**✅ Pin to Tagged Versions**
```go
require github.com/org/cache v1.4.2
```

**Benefits**: Readable versions, semver guarantees, `go get -u` works correctly

### Replace for Local Development

**❌ Committing replace Directives**
```go
// go.mod — checked into git, breaks CI and other developers
module github.com/org/myapp

require github.com/org/cache v1.4.2

replace github.com/org/cache => ../cache  // DO NOT COMMIT
```

**✅ Use go.work Instead (Gitignored)**
```go
// go.work — local only, gitignored
go 1.26

use (
    .
    ../cache
)
```

```bash
# .gitignore
go.work
go.work.sum
```

**Benefits**: No accidental commits, no broken CI, works for multi-module repos

### Retract for Bad Versions

**❌ Deleting Git Tags**
```bash
git tag -d v1.3.0
git push origin :refs/tags/v1.3.0
# Consumers who already cached v1.3.0 still have the broken version
```

**✅ Retract in go.mod**
```go
module github.com/org/cache

go 1.26

retract (
    v1.3.0 // data corruption in concurrent Set()
    [v1.1.0, v1.1.3] // incorrect TTL calculation
)
```

**Benefits**: `go get` warns consumers, proxy caches respect retract, auditable history

## Dependency Minimization

### Stdlib First

**❌ Third-Party for Stdlib-Equivalent Tasks**
```go
import "github.com/pkg/errors"

func process() error {
    return errors.Wrap(err, "processing failed")
}
```

**✅ Use stdlib — errors, fmt, slog**
```go
import "fmt"

func process() error {
    return fmt.Errorf("processing failed: %w", err)
}
```

**Benefits**: Zero dependency, no supply chain risk, always up to date with Go version

### Fewer Deps = Less Risk

**❌ Heavy Framework for Simple HTTP**
```go
import "github.com/some/mega-framework"

func main() {
    app := mega.New()
    app.GET("/health", func(c mega.Context) error {
        return c.JSON(200, map[string]string{"status": "ok"})
    })
    app.Start(":8080")
}
```

**✅ net/http for Simple Services**
```go
import "net/http"

func main() {
    http.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.Write([]byte(`{"status":"ok"}`))
    })
    http.ListenAndServe(":8080", nil)
}
```

**Benefits**: Go 1.22+ routing patterns cover most needs, zero transitive deps, smaller binary

### Evaluate Before Adding

```bash
# Check what a dependency pulls in
go mod graph | grep github.com/some/dep

# Count transitive dependencies
go list -m all | wc -l

# Check module size
go mod download -x github.com/some/dep@latest 2>&1 | tail -1
```

**Benefits**: Informed decisions, avoid dependency bloat

## Vendoring Strategy

### When to Vendor

**❌ Vendoring Everything by Default**
```bash
go mod vendor  # 200MB of vendored code in every repo
```

**✅ Vendor When Reproducibility Matters**
```bash
# Vendor for: production services, air-gapped builds, CI stability
go mod vendor

# Verify vendor matches go.sum
go mod verify
```

```go
// go.mod — nothing special needed
module github.com/org/myapp

go 1.26
```

```bash
# Build with vendor
go build -mod=vendor ./...

# Test with vendor
go test -mod=vendor ./...
```

**Benefits**: Hermetic builds, survives registry outages, required for some compliance

### Keep Vendor Updated

```bash
# Update vendor after go.mod changes
go mod tidy && go mod vendor

# Verify consistency
go mod verify
```

**Benefits**: Vendor stays in sync, no stale dependencies

## Dependency Auditing

### Security and Vulnerability Scanning

**❌ Never Checking for Vulnerabilities**
```bash
# Ship it! What could go wrong?
go build ./...
```

**✅ Regular Vulnerability Checks**
```bash
# Check for known vulnerabilities (official Go tool)
govulncheck ./...

# Check only dependencies (faster, no source analysis)
govulncheck -mode=binary ./cmd/myapp

# Check in CI
govulncheck -format=json ./... | jq '.findings[]'
```

**Benefits**: Catches known CVEs, official Go toolchain integration, actionable reports

### Dependency Graph Analysis

```bash
# Full dependency graph
go mod graph

# Why is a specific dependency needed?
go mod why github.com/some/dep

# All direct and indirect dependencies as JSON
go list -m -json all

# Find which package imports a dependency
go list -deps ./... | grep github.com/some/dep
```

**Benefits**: Understand dependency chains, identify unnecessary transitive deps

## Upgrade Strategy

### Patch Versions — Auto-Update

**❌ Ignoring Patches**
```bash
# Staying on v1.4.0 when v1.4.3 has security fixes
```

**✅ Regular Patch Updates**
```bash
# Update all dependencies to latest patch
go get -u=patch ./...
go mod tidy
go test ./...
```

**Benefits**: Security fixes, bug fixes, no API changes

### Minor Versions — Test After Update

```bash
# Update specific dependency to latest minor
go get github.com/org/cache@latest
go mod tidy
go test ./...

# Update all to latest minor
go get -u ./...
go mod tidy
go test ./...
```

**Benefits**: New features, backwards compatible, verify with tests

### Major Versions — Audit API Changes

**❌ Blind Major Upgrade**
```bash
go get github.com/org/cache/v2@latest  # hope nothing breaks
```

**✅ Audit Breaking Changes**
```bash
# Read changelog and migration guide first
# Then update
go get github.com/org/cache/v2@latest
go mod tidy

# Fix compilation errors — each one is a breaking change
go build ./...

# Run full test suite
go test ./...
```

**Benefits**: Controlled migration, understand what changed, no surprise breakage

## Private Modules

### Configure Go for Private Repos

**❌ Broken go get for Private Modules**
```bash
go get github.com/private-org/internal-lib@latest
# 404 — Go tries proxy.golang.org which can't access private repos
```

**✅ Set GOPRIVATE**
```bash
# In shell profile or CI config
export GOPRIVATE="github.com/private-org/*"

# For multiple orgs
export GOPRIVATE="github.com/private-org/*,github.com/other-org/*"

# Skip checksum database for private modules
export GONOSUMDB="github.com/private-org/*"

# Skip sum verification entirely (air-gapped)
export GONOSUMCHECK="github.com/private-org/*"
```

**Benefits**: `go get` works for private repos, no leaking module paths to public proxy

### Private Module Authentication

```bash
# Git credential for HTTPS
git config --global url."https://oauth2:${TOKEN}@github.com/private-org".insteadOf \
    "https://github.com/private-org"

# Or use SSH
git config --global url."ssh://git@github.com/private-org".insteadOf \
    "https://github.com/private-org"

# Verify it works
GOPRIVATE="github.com/private-org/*" go get github.com/private-org/internal-lib@latest
```

**Benefits**: Seamless authentication, works in CI and local development

### Private Module Proxy (Enterprise)

```bash
# Run a private Go module proxy (Athens, Artifactory, etc.)
export GOPROXY="https://athens.internal.example.com,https://proxy.golang.org,direct"
export GONOSUMDB="github.com/private-org/*"

# All modules route through internal proxy first
go get github.com/private-org/internal-lib@latest
```

**Benefits**: Caching, access control, audit trail, survives upstream outages

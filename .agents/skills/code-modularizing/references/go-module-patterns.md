# Go Module Patterns

Module extraction and organization patterns for Go 1.26+.

## Module Extraction Decision

### When to Create a Separate go.mod

**❌ Single Consumer — Keep In-Package**
```go
// myapp/internal/cache/cache.go
// Only used by myapp — no reason to extract
package cache

type Store struct {
    data map[string]any
}

func New() *Store {
    return &Store{data: make(map[string]any)}
}
```

**✅ 2+ Independent Consumers — Extract Module**
```go
// github.com/org/go-cache/cache.go
// Used by myapp, api-gateway, and worker-service
module github.com/org/go-cache

package cache

type Store struct {
    data map[string]any
}

func New() *Store {
    return &Store{data: make(map[string]any)}
}
```

**Benefits**: Clear ownership boundary, independent versioning, each consumer pins its own version

### When to Use internal/

**❌ Exposing Implementation Details as Public API**
```go
// github.com/org/mylib/convert/reflect.go
package convert

// Exposed publicly but only used by convert package
func ReflectFields(v any) []string { /* ... */ }
```

**✅ Hide Implementation with internal/**
```go
// github.com/org/mylib/internal/reflect/fields.go
package reflect

// Only accessible within mylib module
func Fields(v any) []string { /* ... */ }
```

**Benefits**: Prevents accidental coupling, free to refactor without breaking consumers

## go.work for Multi-Module Development

### Workspace Setup

**❌ Using replace Directives in go.mod**
```go
// go.mod — pollutes the module file, easy to forget before commit
module github.com/org/myapp

require github.com/org/go-cache v1.2.0

replace github.com/org/go-cache => ../go-cache
```

**✅ Using go.work for Local Development**
```go
// go.work — gitignored, local development only
go 1.26

use (
    ./myapp
    ./go-cache
    ./go-config
)
```

```bash
# Initialize workspace
go work init ./myapp ./go-cache ./go-config

# Add a module to the workspace
go work use ./new-module

# Sync workspace dependencies
go work sync
```

**Benefits**: No accidental replace commits, workspace-scoped, multiple modules developed in parallel

### Local Development Workflow

**❌ Tag-and-Update Cycle for Every Change**
```bash
# Slow: tag go-cache, push, update myapp go.mod
cd go-cache && git tag v1.2.1 && git push --tags
cd ../myapp && go get github.com/org/go-cache@v1.2.1
```

**✅ Workspace-Based Iteration**
```bash
# Fast: edit go-cache, changes visible immediately in myapp
cd go-cache
# edit code...
cd ../myapp
go test ./...  # uses local go-cache automatically via go.work
```

**Benefits**: Instant feedback loop, no version churn during development

## Version Strategy

### v0 — Unstable API

**❌ Tagging v1.0.0 Before API Stabilizes**
```go
// go.mod
module github.com/org/go-cache
// Tagged v1.0.0 but API still changing weekly
```

**✅ Stay on v0 Until API Is Stable**
```go
// go.mod
module github.com/org/go-cache
// v0.x.y — no stability promise, free to break API
```

```bash
git tag v0.1.0  # initial release
git tag v0.2.0  # breaking change — fine, it's v0
git tag v0.3.0  # another breaking change — still fine
```

**Benefits**: Freedom to iterate on API design without breaking semver promises

### v1 — Stable API

```go
// go.mod
module github.com/org/go-cache
// v1.x.y — backwards-compatible changes only
```

```bash
git tag v1.0.0  # stable release
git tag v1.1.0  # new feature, backwards compatible
git tag v1.1.1  # bug fix
```

**Benefits**: Consumers trust stability, go get upgrades safely within v1

### v2+ — Breaking Changes with /v2 Path

**❌ Breaking API in v1**
```go
// v1 consumer code breaks silently after update
cache := gocache.New("redis://localhost")  // was New(config)
```

**✅ Major Version Suffix for Breaking Changes**
```go
// go.mod
module github.com/org/go-cache/v2

// v2/cache.go
package cache

// New API — incompatible with v1
func New(opts ...Option) *Store { /* ... */ }
```

```go
// Consumer can use both versions during migration
import (
    cachev1 "github.com/org/go-cache"
    cachev2 "github.com/org/go-cache/v2"
)
```

**Benefits**: Consumers migrate at their own pace, v1 and v2 coexist

## Minimal Dependencies

### Avoid Transitive Dependency Bloat

**❌ Heavy Dependencies for Simple Tasks**
```go
import "github.com/some/huge-framework/utils"

func ToPtr[T any](v T) *T {
    return utils.Ptr(v)  // pulled in a 50-package framework for one function
}
```

**✅ Stdlib or Inline for Simple Logic**
```go
func ToPtr[T any](v T) *T {
    return &v  // zero dependencies
}
```

**Benefits**: Smaller binary, fewer supply chain risks, faster builds

### Audit and Clean Dependencies

```bash
# Remove unused dependencies
go mod tidy

# Understand why a dependency exists
go mod why github.com/some/dep

# View full dependency graph
go mod graph

# Check for known vulnerabilities
govulncheck ./...

# List all dependencies as JSON
go list -m -json all
```

**Benefits**: Clean go.mod, no phantom dependencies, vulnerability awareness

## Module Boundaries

### One Concern Per Module

**❌ Kitchen-Sink Module**
```go
// github.com/org/go-toolkit — does everything poorly
package toolkit

func ParseJSON(data []byte) (any, error)      { /* ... */ }
func SendEmail(to, body string) error          { /* ... */ }
func HashPassword(pw string) (string, error)   { /* ... */ }
func ResizeImage(img []byte, w, h int) []byte  { /* ... */ }
```

**✅ Focused Modules**
```go
// github.com/org/go-jsonutil — JSON utilities only
package jsonutil
func Parse(data []byte) (any, error) { /* ... */ }

// github.com/org/go-mailer — email only
package mailer
func Send(to, body string) error { /* ... */ }
```

**Benefits**: Consumers import only what they need, independent release cycles

### Clear Public API Surface

**❌ Exposing Internal Types**
```go
package cache

// All types exported — consumers couple to implementation
type ShardedMap struct {
    Shards    []*Shard
    ShardFunc func(string) int
}

type Shard struct {
    Data map[string]any
    Mu   sync.RWMutex
}
```

**✅ Minimal Exported API**
```go
package cache

// Only export what consumers need
type Store struct {
    shards    []*shard
    shardFunc func(string) int
}

func New(opts ...Option) *Store { /* ... */ }
func (s *Store) Get(key string) (any, bool) { /* ... */ }
func (s *Store) Set(key string, val any)    { /* ... */ }
```

**Benefits**: Free to change internals, smaller API surface to maintain, clearer documentation

## Retract Bad Versions

**❌ Telling Consumers to Avoid a Version Via Slack/Email**
```
"Hey everyone, don't use v1.3.0, it has a data corruption bug"
```

**✅ Using retract in go.mod**
```go
// go.mod
module github.com/org/go-cache

go 1.26

retract (
    v1.3.0 // data corruption bug in Set()
    [v1.1.0, v1.1.3] // incorrect cache invalidation
)
```

**Benefits**: `go get` warns consumers, `go list` hides retracted versions, machine-readable

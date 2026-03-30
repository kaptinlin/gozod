# Go Package Design

Package organization, naming, and API design patterns for Go 1.26+.

## Package Organization

### Group by Concern, Not by Type

**❌ Grouped by Type — util/model/handler**
```go
models/
    user.go         // type User struct
    order.go        // type Order struct
handlers/
    user_handler.go // func HandleUser
    order_handler.go
services/
    user_service.go
    order_service.go
```

**✅ Grouped by Concern — user/order**
```go
user/
    user.go         // type User struct + Service + Handler
    store.go        // persistence
order/
    order.go        // type Order struct + Service + Handler
    store.go        // persistence
```

**Benefits**: Related code changes together, clear ownership, reduces cross-package imports

### Avoid utils/common/helpers Packages

**❌ Dumping Ground Package**
```go
package utils

func FormatTime(t time.Time) string     { /* ... */ }
func HashPassword(pw string) string     { /* ... */ }
func ParseCSV(data []byte) [][]string   { /* ... */ }
func SendEmail(to, body string) error   { /* ... */ }
```

**✅ Purpose-Named Packages or Inline**
```go
// Option A: inline simple helpers where used
func (s *UserService) formatCreatedAt() string {
    return s.CreatedAt.Format(time.RFC3339)
}

// Option B: purpose-named package for reusable logic
package csvutil
func Parse(data []byte) [][]string { /* ... */ }
```

**Benefits**: Package name communicates intent, no grab-bag accumulation

## Export Strategy

### Export Only What Consumers Need

**❌ Everything Exported**
```go
package auth

type TokenValidator struct {
    SecretKey    []byte          // implementation detail
    Clock        func() time.Time // implementation detail
    ParseFunc    func(string) (*jwt.Token, error) // implementation detail
}

func (v *TokenValidator) Validate(token string) (*Claims, error) { /* ... */ }
func (v *TokenValidator) RefreshCache() { /* ... */ }  // internal maintenance
```

**✅ Minimal Exports**
```go
package auth

type TokenValidator struct {
    secretKey []byte
    clock     func() time.Time
    parseFunc func(string) (*jwt.Token, error)
}

func NewTokenValidator(secret []byte, opts ...Option) *TokenValidator { /* ... */ }
func (v *TokenValidator) Validate(token string) (*Claims, error) { /* ... */ }
```

**Benefits**: Smaller API surface, freedom to refactor internals, clearer godoc

### Use internal/ for Implementation

**❌ Public Package Used Only Internally**
```go
// pkg/reflectutil/fields.go — exported but only used by this project
package reflectutil

func StructFields(v any) []reflect.StructField { /* ... */ }
```

**✅ internal/ Prevents External Use**
```go
// internal/reflectutil/fields.go — compiler enforces internal-only access
package reflectutil

func StructFields(v any) []reflect.StructField { /* ... */ }
```

**Benefits**: Compiler-enforced boundary, no accidental public API commitments

## Package Naming

### Short, Clear, No Stuttering

**❌ Stuttering Names**
```go
package user

type UserService struct{}       // user.UserService — "user" repeated
type UserRepository struct{}    // user.UserRepository — "user" repeated

func NewUserService() *UserService { return &UserService{} }
```

**✅ Context-Aware Names**
```go
package user

type Service struct{}           // user.Service — clear and concise
type Repository struct{}        // user.Repository

func NewService() *Service { return &Service{} }
```

**Benefits**: `user.Service` reads naturally, less visual noise, idiomatic Go

### Package Name as API Prefix

**❌ Redundant Package Name**
```go
package cachemanager

func New() *CacheManager { /* ... */ }
// Usage: cachemanager.New() — tells you nothing new
```

**✅ Package Name Carries Meaning**
```go
package cache

func NewManager() *Manager { /* ... */ }
// Usage: cache.NewManager() — reads as sentence
```

**Benefits**: API reads like documentation, shorter import aliases

### Singular Package Names

**❌ Plural Package Names**
```go
package users    // users.New() — awkward
package configs  // configs.Load() — awkward
```

**✅ Singular Package Names**
```go
package user    // user.New() — natural
package config  // config.Load() — natural
```

**Benefits**: Consistent with stdlib convention (`strings` is the exception, not the rule)

## Package Documentation

### doc.go Convention

**❌ No Package Documentation**
```go
package cache

// Jump straight into code with no context
type Store struct { /* ... */ }
```

**✅ doc.go with Package-Level Comment**
```go
// Package cache provides an in-memory key-value store with TTL support
// and optional sharding for concurrent access.
//
// Basic usage:
//
//	s := cache.New(cache.WithTTL(5 * time.Minute))
//	s.Set("key", "value")
//	v, ok := s.Get("key")
//
// For high-concurrency workloads, enable sharding:
//
//	s := cache.New(cache.WithShards(16))
package cache
```

**Benefits**: `go doc cache` shows useful overview, godoc renders examples

## Interface Location

### Define Interfaces in Consumer, Not Provider

**❌ Provider Defines Interface**
```go
// github.com/org/cache — provider defines interface
package cache

type Cache interface {         // only 1 implementation exists
    Get(key string) (any, error)
    Set(key string, val any) error
}

type MemoryCache struct{}
func (m *MemoryCache) Get(key string) (any, error)     { /* ... */ }
func (m *MemoryCache) Set(key string, val any) error   { /* ... */ }
```

**✅ Consumer Defines Interface**
```go
// github.com/org/cache — provider exports concrete type
package cache

type MemoryCache struct{}
func New() *MemoryCache { return &MemoryCache{} }
func (m *MemoryCache) Get(key string) (any, error)     { /* ... */ }
func (m *MemoryCache) Set(key string, val any) error   { /* ... */ }

// github.com/org/myapp/internal/app — consumer defines what it needs
package app

type Getter interface {
    Get(key string) (any, error)
}

func Process(cache Getter) error { /* ... */ }
```

**Benefits**: Consumer imports only what it needs, provider free to add methods, implicit satisfaction

### Accept Interfaces, Return Structs

**❌ Return Interface**
```go
func NewStore() Store {          // returns interface — hides concrete type
    return &memoryStore{}
}
```

**✅ Return Concrete Type**
```go
func NewStore() *MemoryStore {   // returns struct — consumer decides abstraction
    return &MemoryStore{}
}

// Consumer side — define interface if needed
type Getter interface {
    Get(key string) (any, error)
}

func handler(store Getter) { /* ... */ }  // accepts interface
```

**Benefits**: Consumers see full API, can choose which methods to abstract, no hidden capabilities

## API Design Patterns

### Constructor Functions

**❌ Direct Struct Init with Required Setup**
```go
store := &cache.Store{}
store.Init()  // easy to forget, panics without it
```

**✅ Constructor Enforces Valid State**
```go
store := cache.New()  // always returns valid, ready-to-use Store
```

**Benefits**: Impossible to create invalid state, clear entry point

### Functional Options

**❌ Config Struct with Many Optional Fields**
```go
type Config struct {
    MaxSize    int
    TTL        time.Duration
    Shards     int
    Logger     *slog.Logger
    OnEvict    func(key string)
}

store := cache.New(cache.Config{
    MaxSize: 1000,
    // which fields are required? what are defaults?
})
```

**✅ Functional Options Pattern**
```go
type Option func(*config)

func WithMaxSize(n int) Option {
    return func(c *config) { c.maxSize = n }
}

func WithTTL(d time.Duration) Option {
    return func(c *config) { c.ttl = d }
}

func WithLogger(l *slog.Logger) Option {
    return func(c *config) { c.logger = l }
}

// Usage — only specify what you need
store := cache.New(
    cache.WithMaxSize(1000),
    cache.WithTTL(5 * time.Minute),
)
```

**Benefits**: Self-documenting, backwards-compatible additions, clear defaults

### Compile-Time Interface Checks

**❌ Runtime Panic When Interface Not Satisfied**
```go
// Discovered at runtime that MemoryStore doesn't implement Closer
var s Store = &MemoryStore{}  // compiles but may fail later
```

**✅ Compile-Time Assertion**
```go
var _ Store = (*MemoryStore)(nil)   // fails at compile time if unsatisfied
var _ io.Closer = (*MemoryStore)(nil)
```

**Benefits**: Immediate feedback, no runtime surprises, documents intent

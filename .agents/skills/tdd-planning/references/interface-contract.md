---
name: golang-interface-contract
description: Go-specific interface contract syntax for TDD planning
---

# Go Interface Contracts

## Type Definitions

```go
type Token struct {
    UserID    string
    ExpiresAt time.Time
}
```

## Error Definitions

```go
var (
    ErrExpired = errors.New("token expired")
    ErrInvalid = errors.New("token invalid")
    ErrNotFound = errors.New("not found")
)
```

## Function Signatures

```go
func ParseToken(raw string) (Token, error)
func ValidateToken(tok Token, secret []byte) error
```

## Interface Definitions

```go
type Storer interface {
    Get(ctx context.Context, key string) (string, error)
    Set(ctx context.Context, key, value string) error
    Delete(ctx context.Context, key string) error
}

type TokenParser interface {
    Parse(raw string) (Token, error)
    Validate(tok Token) error
}
```

## Constructor Signatures

```go
// Functional options pattern
type Option func(*Store)

func WithTimeout(d time.Duration) Option
func NewStore(opts ...Option) (*Store, error)

// Config struct pattern
type Config struct {
    Timeout   time.Duration
    MaxRetries int
}

func NewStore(cfg Config) (*Store, error)
```

## Interface Compliance Check

```go
// Compile-time check - place at package level
var _ Storer = (*Store)(nil)
var _ TokenParser = (*TokenService)(nil)
```

## Method Signatures

```go
func (s *Store) Get(ctx context.Context, key string) (string, error)
func (s *Store) Set(ctx context.Context, key, value string) error
func (s *Store) Close() error
```

## Context Usage

```go
func (s *Store) Process(ctx context.Context, input Input) (Output, error) {
    select {
    case <-ctx.Done():
        return "", ctx.Err()
    default:
        // Process
    }
}
```

## Return Value Patterns

```go
// Single value
func GetValue() string

// Value + error
func GetValue() (string, error)

// Multiple values
func Parse(input string) (Token, error)

// Interface return
func NewStore() Storer
```

## Channel Patterns

```go
func (s *Store) Stream(ctx context.Context) <-chan Result
func (s *Store) Consume(stream <-chan Input) error
```

## Concurrency Primitives

```go
type Store struct {
    closed atomic.Bool
    mu     sync.RWMutex
    once   sync.Once
}
```

## Example Complete Contract

```go
package store

import (
    "context"
    "sync"
    "time"
)

// Errors
var (
    ErrNotFound    = errors.New("not found")
    ErrAlreadyClosed = errors.New("already closed")
)

// Token represents an authentication token
type Token struct {
    UserID    string
    ExpiresAt time.Time
}

// Store defines the storage interface
type Store interface {
    Get(ctx context.Context, key string) (string, error)
    Set(ctx context.Context, key, value string) error
    Delete(ctx context.Context, key string) error
    Close() error
}

// Config holds store configuration
type Config struct {
    Timeout time.Duration
}

// Option implements functional options pattern
type Option func(*store)

func WithTimeout(d time.Duration) Option {
    return func(s *store) { s.timeout = d }
}

// store implements Store interface
var _ Store = (*store)(nil)

type store struct {
    timeout time.Duration
    closed  atomic.Bool
    mu      sync.RWMutex
    data    map[string]string
}

func NewStore(opts ...Option) (Store, error) {
    s := &store{
        timeout: 30 * time.Second,
        data:    make(map[string]string),
    }
    for _, opt := range opts {
        opt(s)
    }
    return s, nil
}
```

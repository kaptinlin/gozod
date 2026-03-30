---
description: Implement caching in Go applications using agentable/go-cache with type-safe generics, pluggable stores, and multi-layer support. Use when adding caching to a service, configuring cache stores (memory, Redis, SQLite, PostgreSQL), setting up multi-layer caching, implementing loader patterns, or choosing codecs.
name: go-cache-caching
---


# Go Caching Guide

Implement caching in Go applications using `github.com/agentable/go-cache` -- a generic, store-driven cache library with type-safe `Cache[T]` interface, pluggable backends, and multi-layer support.

## Decision Flowchart

```
What caching pattern does the application need?
|
+- Simple key-value cache (single store)
|  |
|  +- Hot data, same process, ephemeral
|  |  -> memory store (Ristretto, <100ns)
|  |
|  +- Shared across services, distributed
|  |  -> redis store (Rueidis, ~40us, client-side caching optional)
|  |
|  +- Local persistent cache, survives restarts
|  |  -> sqlite store (pure Go, ~8us, WAL mode)
|  |
|  +- Distributed persistent, queryable
|     -> postgres store (~10ms, existing infrastructure)
|
+- Multi-layer cache (L1 fast + L2 durable)
|  -> NewChain[T](l1Cache, l2Cache) with automatic backfill
|
+- Cache-aside with automatic loading on miss
|  -> NewLoader[T](store) + GetOrLoad with single-flight dedup
|
+- Bulk reads/writes to reduce round-trips
   -> NewBatch[T](store) + GetMany/SetMany/DeleteMany
```

## Architecture Overview

```
+-----------------------------------------------------------+
| Application                                                |
|                                                            |
|  Cache[T] / BatchCache[T] / LoaderCache[T] / ChainCache[T]|
|    - Type-safe generic API                                 |
|    - Functional options: WithTTL, WithCodec, WithMetrics   |
|    - Automatic serialization via Codec interface            |
+-----------------------------------------------------------+
| Store interface (byte-oriented, 5 methods)                 |
|  +----------+  +---------+  +----------+  +-----------+   |
|  | memory/  |  | redis/  |  | sqlite/  |  | postgres/ |   |
|  | Ristretto|  | Rueidis |  | pure Go  |  | pgx/sql   |   |
|  +----------+  +---------+  +----------+  +-----------+   |
+-----------------------------------------------------------+
| Codec interface (Encode/Decode)                            |
|  JSON (default) | MessagePack (prod) | Gob (Go-specific)  |
+-----------------------------------------------------------+
```

## Quick Setup

```go
import (
    "github.com/agentable/go-cache"
    "github.com/agentable/go-cache/store/memory"
)

// 1. Create store
store, err := memory.New(memory.Config{MaxCost: 100 << 20}) // 100MB
if err != nil { return err }
defer store.Close()

// 2. Create type-safe cache
userCache := cache.New[User](store, cache.WithTTL(5*time.Minute))
defer userCache.Close()

// 3. Use it
userCache.Set(ctx, "user:123", user, 0)           // default TTL
userCache.Set(ctx, "user:456", user, 1*time.Minute) // custom TTL
retrieved, err := userCache.Get(ctx, "user:123")
```

## Core Interfaces

| Constructor | Interface | Methods | Use Case |
|-------------|-----------|---------|----------|
| `cache.New[T](store, opts...)` | `Cache[T]` | Get, Set, Delete, Clear, Close | Basic caching |
| `cache.NewBatch[T](store, opts...)` | `BatchCache[T]` | Cache[T] + GetMany, SetMany, DeleteMany | Bulk operations |
| `cache.NewLoader[T](store, opts...)` | `LoaderCache[T]` | Cache[T] + GetOrLoad | Cache-aside with single-flight |
| `cache.NewChain[T](layers...)` | `ChainCache[T]` | Cache[T] (multi-layer) | L1/L2/L3 with backfill |
| `cache.NewIterable[T](store, opts...)` | `IterableCache[T]` | Cache[T] + All(ctx) iter.Seq2 | Iteration (SQLite, Postgres only) |

## Options

```go
cache.WithTTL(5 * time.Minute)          // default TTL (0 = no expiration)
cache.WithCodec(codec.MessagePackCodec{}) // serialization (default: JSON)
cache.WithMetrics(&cache.Metrics{})      // hit/miss/error tracking
cache.WithLogger(slog.Default())         // structured logging
```

## Multi-Layer Caching

```go
// L1: Fast in-memory (small, short TTL)
l1Store, _ := memory.New(memory.Config{MaxCost: 1 << 20})
l1 := cache.New[User](l1Store, cache.WithTTL(1*time.Minute))

// L2: Distributed Redis (larger, longer TTL)
l2Store, _ := redis.New(redis.Config{Client: redisClient})
l2 := cache.New[User](l2Store, cache.WithTTL(10*time.Minute))

// Chain: Get checks L1 -> L2, backfills L1 on L2 hit
chain := cache.NewChain[User](l1, l2)
defer chain.Close()
```

Backfill is asynchronous (background context) -- does not block the Get caller. Set/Delete/Clear propagate to all layers.

## Loader Pattern (Cache-Aside)

```go
loader := cache.NewLoader[User](store, cache.WithTTL(5*time.Minute))
defer loader.Close()

user, err := loader.GetOrLoad(ctx, "user:123", func(ctx context.Context, key string) (User, error) {
    return db.GetUser(ctx, key) // called only on cache miss
})
```

Single-flight deduplication: 100 concurrent requests for the same key trigger only 1 loader call. Uses `context.WithoutCancel` so one caller's cancellation does not abort the shared load.

## Batch Operations

```go
batch := cache.NewBatch[User](store, cache.WithTTL(5*time.Minute))

// SetMany: all-or-nothing semantics
batch.SetMany(ctx, map[string]User{
    "user:100": alice,
    "user:101": bob,
}, 0)

// GetMany: returns only found keys (missing keys are not errors)
found, err := batch.GetMany(ctx, []string{"user:100", "user:101", "user:999"})

// DeleteMany: idempotent
batch.DeleteMany(ctx, []string{"user:100", "user:101"})
```

## Metrics

```go
metrics := &cache.Metrics{}
c := cache.New[User](store, cache.WithMetrics(metrics), cache.WithTTL(5*time.Minute))

// After operations:
metrics.Hits.Load()        // int64
metrics.Misses.Load()      // int64
metrics.Errors.Load()      // int64
metrics.SharedLoads.Load() // int64 (loader dedup count)
metrics.HitRate()          // float64 (0.0-1.0)
```

## Error Handling

```go
import "errors"

val, err := c.Get(ctx, key)
if errors.Is(err, cache.ErrNotFound) { /* cache miss */ }
if errors.Is(err, cache.ErrClosed)   { /* cache was closed */ }
if errors.Is(err, cache.ErrInvalidKey)   { /* empty key */ }
if errors.Is(err, cache.ErrInvalidValue) { /* negative TTL or nil value */ }
```

## Store Selection -- [details](references/store-configuration.md)

| Store | Latency | Persistence | Distribution | Best For |
|-------|---------|-------------|--------------|----------|
| `memory` | <100ns | No | Single-server | Hot data, sessions, rate limiting |
| `redis` | ~40us | Optional | Multi-server | Shared cache, distributed systems |
| `sqlite` | ~8us | Yes | Single-server | Local persistent, edge deployments |
| `postgres` | ~10ms | Yes | Multi-server | Distributed persistent, queryable |

## Codec Selection

| Codec | Round-Trip | Size | Use Case |
|-------|-----------|------|----------|
| `codec.MessagePackCodec{}` | ~800ns | Compact | **Production (recommended)** |
| `codec.JSONCodec{}` | ~1.8us | Medium | Development, debugging |
| `codec.GobCodec{}` | ~70us | Large | Go-specific complex types |

## Anti-Patterns

| Anti-Pattern | Problem | Fix |
|-------------|---------|-----|
| No `defer cache.Close()` / `defer store.Close()` | Resource leak, goroutine leak | Always defer Close on both store and cache |
| Sharing one `Cache[T]` for different types | Type mismatch on Get, silent decode errors | Create separate `Cache[T]` per type |
| Negative TTL in Set | Returns `ErrInvalidValue` | Use 0 for default TTL, positive for custom |
| Calling Set inside LoaderFunc | Deadlock -- loader is called inside singleflight | Return value from loader; cache stores it automatically |
| No key prefix in shared Redis | Key collisions across services | Use `redis.Config{KeyPrefix: "svc:"}` |
| Using GobCodec in production | 40-90x slower than MessagePack | Use `codec.MessagePackCodec{}` for production |
| Ignoring context in loader functions | Leaked goroutines, slow shutdown | Respect `ctx.Err()` and return early |
| Storing pointers in cache | Cached value mutated by caller, race conditions | Store value types; the codec serializes a copy |

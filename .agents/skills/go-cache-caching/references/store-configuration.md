# Store Configuration Reference

Detailed configuration for each `github.com/agentable/go-cache` store backend.

## Contents
- Memory store (Ristretto)
- Redis store (Rueidis)
- SQLite store (pure Go)
- PostgreSQL store
- Iteration support
- Store feature matrix

## Memory Store

```go
import "github.com/agentable/go-cache/store/memory"

store, err := memory.New(memory.Config{
    MaxCost:     100 << 20, // REQUIRED: max cache size in bytes (100MB)
    NumCounters: 0,         // optional: TinyLFU counters (default: 10 * MaxCost / 1024)
    BufferItems: 64,        // optional: async buffer size (default: 64)
    Metrics:     false,     // optional: Ristretto-level metrics (default: false)
})
defer store.Close()
```

**Key behavior:**
- Set is asynchronous in Ristretto -- value may not be immediately visible after Set returns
- In tests, call `store.Wait()` after Set to ensure visibility before Get
- TinyLFU eviction: frequently accessed items survive longer
- Cost is `len(value)` in bytes after codec encoding
- Does not implement `IterableStore` (Ristretto does not expose key iteration)

**Sizing guide:**
- `MaxCost` = estimated working set size in bytes
- For 10,000 items averaging 1KB each: `MaxCost: 10 << 20` (10MB)
- Over-provisioning by 20-50% improves hit ratio

## Redis Store

```go
import (
    "github.com/agentable/go-cache/store/redis"
    "github.com/redis/rueidis"
)

client, err := rueidis.NewClient(rueidis.ClientOption{
    InitAddress: []string{"localhost:6379"},
})
if err != nil { return err }

store, err := redis.New(redis.Config{
    Client:            client, // REQUIRED: Rueidis client
    KeyPrefix:         "myapp:", // optional: prefix for all keys (default: "")
    ClientSideCaching: false,   // optional: enable client-side caching (default: false)
})
defer store.Close()
```

**Key behavior:**
- Client-side caching: Redis server-assisted, invalidation-based local cache (RESP3)
- KeyPrefix: isolate keys per service in shared Redis instances
- Clear with prefix uses SCAN (iterative, context-cancellable); without prefix uses FLUSHDB
- Rueidis provides auto-pipelining for concurrent commands

**When to enable client-side caching:**
- Read-heavy workloads with infrequent writes
- When network latency to Redis is significant
- Not useful if data changes frequently (invalidation overhead)

## SQLite Store

```go
import "github.com/agentable/go-cache/store/sqlite"

store, err := sqlite.New(sqlite.Config{
    Path:            "/var/cache/myapp.db", // REQUIRED: file path or ":memory:"
    MaxOpenConns:    1,                     // optional: max connections (default: 1)
    CleanupInterval: 1 * time.Minute,       // optional: expired entry cleanup (default: 1m)
    WALMode:         nil,                   // optional: WAL mode (default: true)
})
defer store.Close()
```

**Key behavior:**
- WAL mode (default enabled): allows concurrent reads during writes
- `MaxOpenConns: 1` recommended for SQLite to avoid SQLITE_BUSY errors
- Background cleanup goroutine removes expired entries at `CleanupInterval`
- Set `CleanupInterval` to negative value to disable automatic cleanup
- Implements `IterableStore` -- use with `cache.NewIterable[T]()` for iteration
- Pure Go driver (`modernc.org/sqlite`) -- no CGO required
- Use `":memory:"` for in-memory database in tests

## PostgreSQL Store

```go
import (
    "database/sql"
    "github.com/agentable/go-cache/store/postgres"
)

// Option A: Provide DSN (store manages the connection)
store, err := postgres.New(postgres.Config{
    DSN:              "postgres://user:pass@localhost:5432/cache?sslmode=disable",
    TableName:        "cache_entries",  // optional (default: "cache_entries")
    MaxOpenConns:     25,               // optional (default: 25)
    MaxIdleConns:     5,                // optional (default: 5)
    ConnMaxLifetime:  5 * time.Minute,  // optional (default: 5m)
    CleanupInterval:  1 * time.Minute,  // optional (default: 1m)
    CleanupBatchSize: 1000,             // optional (default: 1000)
})

// Option B: Provide existing *sql.DB (caller manages the connection)
db, _ := sql.Open("pgx", dsn)
store, err := postgres.New(postgres.Config{
    DB:        db,
    TableName: "cache_entries",
})
defer store.Close()
```

**Key behavior:**
- Auto-creates the cache table on first use (schema migration)
- Background cleanup deletes expired entries in batches (`CleanupBatchSize`)
- Set `CleanupInterval` to negative value to disable automatic cleanup
- Implements `IterableStore` -- use with `cache.NewIterable[T]()` for iteration
- MVCC provides concurrent read/write without locking

## Iteration (SQLite + PostgreSQL only)

```go
// Only for stores implementing IterableStore (SQLite, PostgreSQL)
iterCache := cache.NewIterable[User](sqliteStore, cache.WithTTL(5*time.Minute))

for key, user := range iterCache.All(ctx) {
    fmt.Printf("%s: %s\n", key, user.Name)
}
```

- Yields only non-expired entries
- Decode failures are silently skipped
- Supports early termination (break from range)
- Respects context cancellation
- Not a snapshot: concurrent modifications may or may not be visible
- Memory store does NOT support iteration

## Store Feature Matrix

| Feature | Memory | Redis | SQLite | PostgreSQL |
|---------|--------|-------|--------|------------|
| Type-safe API | yes | yes | yes | yes |
| Batch operations | yes | yes | yes | yes |
| TTL expiration | yes | yes | yes | yes |
| Auto eviction | yes (TinyLFU) | yes (Redis LRU) | yes (cleanup) | no |
| Concurrent reads | yes | yes | yes | yes |
| Concurrent writes | yes | yes | limited (WAL helps) | yes |
| Survives restarts | no | optional (AOF/RDB) | yes | yes |
| Client-side caching | no | yes (RESP3) | no | no |
| Auto-pipelining | no | yes | no | no |
| Iteration support | no | no | yes | yes |
| Background cleanup | no | no (Redis handles) | yes (goroutine) | yes (goroutine) |
| Pure Go | yes | yes | yes | yes (pgx) |

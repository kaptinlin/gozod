# Caching

## `github.com/samber/hot` — In-Memory Cache

Best for read-heavy, single-instance applications needing advanced eviction.

**Key features:**
- 9 eviction algorithms: LRU, LFU, TinyLFU, W-TinyLFU, S3FIFO, ARC, 2Q, SIEVE, FIFO
- TTL with jitter (prevents cache stampedes)
- Stale-while-revalidate pattern (background refresh)
- Missing key caching (avoids repeated lookups for absent keys)
- Sharded architecture (reduced lock contention)
- Loader chains with concurrent request deduplication
- Type-safe generics, Prometheus metrics

**When to use:**
- Single-process app with hot data
- Need fine-grained eviction control
- Read-heavy workloads
- Need stale-while-revalidate

## `github.com/eko/gocache` — Multi-Backend Cache

Best for applications needing cache abstraction across multiple backends.

**Key features:**
- Unified interface: in-memory (Bigcache, Ristretto, Go-cache), distributed (Redis, Memcache, Hazelcast)
- Chain caching: L1 memory → L2 Redis with automatic fallback
- Loadable caching: auto-populate on miss via callback
- Built-in Prometheus metrics
- Tag-based invalidation
- Type-safe generics

**When to use:**
- Multi-tier caching (memory + Redis)
- Need backend abstraction
- Tag-based invalidation needed
- May switch cache backends in the future

## Decision Tree

```
Need caching?
├── Single process, in-memory only?
│   ├── Advanced eviction / stale-while-revalidate → samber/hot
│   └── Simple TTL enough → sync.Map + time.AfterFunc (stdlib)
├── Multi-tier (memory + Redis/Memcache)?
│   └── eko/gocache (chain cache)
└── Just Redis/Memcache client?
    └── Use the client directly, no cache wrapper
```

## Do NOT Use

| Library | Reason |
|---------|--------|
| `patrickmn/go-cache` | Unmaintained, no generics |
| Rolling your own LRU | Use `samber/hot` with LRU algorithm |

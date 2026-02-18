# Concurrency

## `github.com/panjf2000/ants/v2` — Goroutine Pool

Prevents unbounded goroutine creation with a fixed-capacity pool.

**Key features:**
- Automatic goroutine lifecycle management and recycling
- Periodic idle goroutine cleanup
- Dynamic pool capacity adjustment at runtime
- Graceful panic handling
- Non-blocking task submission
- Configurable queue structures (ring buffers, stacks, queues)
- Three pool types: generic, function-specific, multi-pools

**When to use:**
- High-concurrency scenarios where goroutine explosion is a risk
- Need bounded parallelism for batch processing
- Want to limit memory/GC pressure from goroutines

**Production users:** Tencent, ByteDance, Baidu, Milvus, TDengine, Apache DevLake.

```go
pool, _ := ants.NewPool(1000)
defer pool.Release()

for _, task := range tasks {
    task := task
    pool.Submit(func() {
        process(task)
    })
}
```

## When NOT to Use

- **Small concurrency (< 100 goroutines):** Just use `go` + `sync.WaitGroup` or `wg.Go()` (Go 1.25+)
- **Structured concurrency:** Use `golang.org/x/sync/errgroup` with `SetLimit()` for bounded parallelism with error propagation
- **Simple fan-out/fan-in:** `errgroup` or channels suffice

## Decision Tree

```
Need concurrent execution?
├── < 100 goroutines, simple case?
│   └── sync.WaitGroup.Go() or x/sync/errgroup
├── Bounded parallelism with error handling?
│   └── x/sync/errgroup.SetLimit(n)
└── High-concurrency, need pool management, recycling?
    └── panjf2000/ants/v2
```

# Benchmarks

## b.Loop() (Go 1.24+)

Replace `for i := 0; i < b.N; i++` with `for b.Loop()`. Since Go 1.26, `b.Loop()` does **not prevent inlining** in the loop body, giving more accurate measurements.

```go
// Old
func BenchmarkParse(b *testing.B) {
    for i := 0; i < b.N; i++ {
        Parse(input)
    }
}

// New (Go 1.24+)
func BenchmarkParse(b *testing.B) {
    for b.Loop() {
        Parse(input)
    }
}
```

**`b.Loop()` keeps results alive** — no need for `runtime.KeepAlive` or assigning to a package-level `sink` variable:

```go
// Old — needed to prevent dead-code elimination
var sink string

func BenchmarkFormat(b *testing.B) {
    for i := 0; i < b.N; i++ {
        sink = Format(data)
    }
}

// New — b.Loop() handles this automatically
func BenchmarkFormat(b *testing.B) {
    for b.Loop() {
        Format(data)
    }
}
```

### Migration

Run `go fix ./...` (Go 1.26+) to auto-convert existing benchmarks.

---

## Benchmark Structure

### Basic benchmark

```go
func BenchmarkEncode(b *testing.B) {
    data := generateTestData() // Setup outside loop
    b.ResetTimer()             // Exclude setup time

    for b.Loop() {
        Encode(data)
    }
}
```

### With allocation reporting

```go
func BenchmarkEncode(b *testing.B) {
    data := generateTestData()
    b.ResetTimer()
    b.ReportAllocs() // Report allocations per operation

    for b.Loop() {
        Encode(data)
    }
}
```

Run with: `go test -bench=BenchmarkEncode -benchmem ./...`

Output:
```
BenchmarkEncode-8    5000000    240 ns/op    64 B/op    2 allocs/op
```

### Multiple sizes

```go
func BenchmarkEncode(b *testing.B) {
    sizes := []struct {
        name string
        size int
    }{
        {"small", 100},
        {"medium", 10_000},
        {"large", 1_000_000},
    }

    for _, s := range sizes {
        b.Run(s.name, func(b *testing.B) {
            data := generateData(s.size)
            b.ResetTimer()
            b.ReportAllocs()

            for b.Loop() {
                Encode(data)
            }
        })
    }
}
```

### Bytes throughput

For I/O-bound operations, report throughput:

```go
func BenchmarkCompress(b *testing.B) {
    data := loadTestData()
    b.SetBytes(int64(len(data))) // Report MB/s
    b.ResetTimer()

    for b.Loop() {
        Compress(data)
    }
}
```

Output:
```
BenchmarkCompress-8    100000    12000 ns/op    850.00 MB/s
```

---

## Context in Benchmarks

Use `b.Context()` (Go 1.24+):

```go
func BenchmarkFetch(b *testing.B) {
    svc := newTestService(b)
    b.ResetTimer()

    for b.Loop() {
        svc.Fetch(b.Context(), "key")
    }
}
```

---

## benchstat Comparison

Compare benchmark results across code changes:

```bash
# Before change
go test -bench=. -benchmem -count=10 ./... > old.txt

# After change
go test -bench=. -benchmem -count=10 ./... > new.txt

# Compare
benchstat old.txt new.txt
```

Output:
```
          │  old.txt  │           new.txt           │
          │  sec/op   │  sec/op    vs base          │
Encode-8   240.0n ± 2%  180.0n ± 1%  -25.00% (p=0.000 n=10)
```

**Rules:**
- Always use `-count=10` (or more) for statistical significance
- Use `-benchtime=2s` for short benchmarks to reduce noise
- `benchstat` reports p-value — reject changes with p > 0.05

---

## Avoiding Benchmark Pitfalls

### Do not include setup in the measured loop

```go
// Wrong — measures setup + work
func BenchmarkProcess(b *testing.B) {
    for b.Loop() {
        data := loadData()   // This is measured!
        Process(data)
    }
}

// Correct — measures only work
func BenchmarkProcess(b *testing.B) {
    data := loadData()
    b.ResetTimer()
    for b.Loop() {
        Process(data)
    }
}
```

### Do not share mutable state across iterations

```go
// Wrong — buffer grows across iterations
func BenchmarkAppend(b *testing.B) {
    buf := make([]byte, 0, 1024)
    for b.Loop() {
        buf = append(buf, data...)  // Accumulates!
    }
}

// Correct — fresh buffer each iteration
func BenchmarkAppend(b *testing.B) {
    for b.Loop() {
        buf := make([]byte, 0, 1024)
        buf = append(buf, data...)
    }
}
```

### Do not ignore errors in benchmarks

```go
// Wrong — error path may be faster/slower
func BenchmarkParse(b *testing.B) {
    for b.Loop() {
        Parse(input) // Ignoring error
    }
}

// Correct
func BenchmarkParse(b *testing.B) {
    for b.Loop() {
        _, err := Parse(input)
        if err != nil {
            b.Fatal(err) // Stop benchmark on error
        }
    }
}
```

---

## Profiling

### CPU profiling

```bash
go test -bench=BenchmarkTarget -cpuprofile=cpu.prof ./pkg/...
go tool pprof -http=:8080 cpu.prof
```

### Memory profiling

```bash
go test -bench=BenchmarkTarget -memprofile=mem.prof ./pkg/...
go tool pprof -http=:8080 mem.prof
```

### Allocation analysis

```bash
go test -bench=BenchmarkTarget -benchmem -memprofile=mem.prof ./pkg/...
go tool pprof -alloc_objects mem.prof  # Count of allocations
go tool pprof -alloc_space mem.prof    # Size of allocations
```

### Trace

```bash
go test -bench=BenchmarkTarget -trace=trace.out ./pkg/...
go tool trace trace.out
```

---

## Performance Regression Testing

Embed benchmark baselines in CI:

```go
func BenchmarkFire(b *testing.B) {
    m := buildTestMachine(b)
    b.ResetTimer()
    b.ReportAllocs()

    for b.Loop() {
        m.Fire(b.Context(), EventPay)
    }
}
```

In CI pipeline:

```bash
go test -bench=. -benchmem -count=10 ./... > bench.txt
benchstat baseline.txt bench.txt | grep -E '\+[0-9]' && echo "REGRESSION" && exit 1
```

Accept no regressions > 5% without justification.

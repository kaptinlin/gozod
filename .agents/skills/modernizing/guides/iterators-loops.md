# Iterators & Loop Patterns

Go 1.22-1.23 introduced major loop improvements: per-iteration variable scoping, range over integers, and user-defined function iterators.

## Contents
- Loop variable scoping fix (1.22+)
- Range over integers (1.22+)
- Range over function iterators (1.23+)
- iter package (1.23+)
- Iterator-powered stdlib functions (1.23+)
- Custom iterator patterns

---

## Loop variable scoping fix (Go 1.22+)

Each iteration of `for` now creates **new** variables. The old behavior (shared variable across iterations) caused notorious goroutine/closure bugs.

### When to use
- This is automatic when `go 1.22` or later is in `go.mod` — no code change needed
- You can **remove** workarounds like `v := v` inside loops

### When NOT to use
- If your `go.mod` still says `go 1.21` or earlier, the old behavior applies
- Code that intentionally relies on shared loop variables (rare, but check before upgrading)

```go
// Old — bug: all goroutines capture the same `v`
for _, v := range values {
    go func() {
        fmt.Println(v) // always prints last element
    }()
}

// Old workaround
for _, v := range values {
    v := v // shadow with new variable
    go func() {
        fmt.Println(v)
    }()
}

// Go 1.22+ — just works, each iteration has its own `v`
for _, v := range values {
    go func() {
        fmt.Println(v) // correct
    }()
}
```

---

## Range over integers (Go 1.22+)

`for i := range n` iterates `i` from `0` to `n-1`.

### When to use
- Replace `for i := 0; i < n; i++` when you just need a counter
- Cleaner and less error-prone (no off-by-one in the condition)

### When NOT to use
- When starting from a non-zero value — use traditional `for i := start; i < end; i++`
- When the step is not 1 — use traditional loop
- When you need to modify `i` inside the loop body

```go
// Old
for i := 0; i < 100; i++ {
    process(i)
}

// New (Go 1.22+)
for i := range 100 {
    process(i)
}

// Repeat N times (ignore index)
for range 5 {
    retry()
}
```

---

## Range over function iterators (Go 1.23+)

Functions matching `func(yield func(V) bool)` or `func(yield func(K, V) bool)` can be used as range expressions. The `iter` package defines `Seq[V]` and `Seq2[K, V]` types for these.

### When to use
- **Lazy sequences** — generate values on demand without allocating a full slice
- **Composable pipelines** — chain filter/map/take operations
- **Encapsulated iteration** — hide complex traversal logic (trees, graphs, pagination)
- When you want to support `break` / early termination in the consumer

### When NOT to use
- **Simple slice iteration** — `for _, v := range slice` is clearer and faster
- **Small, known collections** — building a slice and ranging over it is simpler
- **Performance-critical hot paths** — iterator function calls have overhead compared to direct slice indexing; benchmark if it matters
- **When the full collection is needed anyway** — if the caller will `slices.Collect` it immediately, just return a slice

```go
// Lazy Fibonacci sequence
func Fibonacci() iter.Seq[int] {
    return func(yield func(int) bool) {
        a, b := 0, 1
        for yield(a) {
            a, b = b, a+b
        }
    }
}

for n := range Fibonacci() {
    if n > 1000 { break }
    fmt.Println(n)
}

// Two-value iterator (key-value)
func Enumerate[T any](s []T) iter.Seq2[int, T] {
    return func(yield func(int, T) bool) {
        for i, v := range s {
            if !yield(i, v) { return }
        }
    }
}

for i, name := range Enumerate(names) {
    fmt.Printf("%d: %s\n", i, name)
}
```

### Iterator protocol rules

1. `yield` returns `false` when the consumer breaks — you **must** stop iterating
2. **Never** call `yield` after it returns `false` — causes runtime panic
3. **Never** call `yield` after the iterator function returns — crashes the program
4. Don't call `yield` from other goroutines without synchronization
5. Iterators can be single-use or reusable — document which
6. Always use `for range`, never call the iterator function directly

### Error handling with Seq2

For fallible iteration (I/O, database, API calls), use `iter.Seq2[V, error]`:

```go
func ReadLines(path string) iter.Seq2[string, error] {
    return func(yield func(string, error) bool) {
        f, err := os.Open(path)
        if err != nil {
            yield("", err)
            return
        }
        defer f.Close()
        scanner := bufio.NewScanner(f)
        for scanner.Scan() {
            if !yield(scanner.Text(), nil) { return }
        }
        if err := scanner.Err(); err != nil {
            yield("", err)
        }
    }
}

for line, err := range ReadLines("/etc/hosts") {
    if err != nil { return err }
    process(line)
}
```

### Container convention: All() method

Custom containers should provide an `All()` method returning `iter.Seq`:

```go
func (s *Set[E]) All() iter.Seq[E] {
    return func(yield func(E) bool) {
        for v := range s.m {
            if !yield(v) { return }
        }
    }
}

for v := range mySet.All() { ... }
```

### Anti-pattern: channel-based iteration

Don't use goroutines + channels for iteration — use `iter.Seq` instead. Channels add scheduler overhead, leak risks, and unnecessary concurrency complexity.

---

## Iterator-powered stdlib (Go 1.23+)

Many stdlib functions now accept or return iterators.

### slices iterator functions
```go
// Collect iterator into slice
names := slices.Collect(maps.Keys(m))

// Sorted keys of a map
keys := slices.Sorted(maps.Keys(m))

// Append iterator results to existing slice
all = slices.AppendSeq(all, moreItems)

// Chunk a slice into groups
for chunk := range slices.Chunk(items, 100) {
    processBatch(chunk)
}

// All / Values / Backward — iterate slices as Seq/Seq2
for i, v := range slices.All(s) { ... }
for v := range slices.Values(s) { ... }
for i, v := range slices.Backward(s) { ... }
```

### maps iterator functions
```go
// Iterate all key-value pairs
for k, v := range maps.All(m) { ... }

// Collect iterator into map
m := maps.Collect(kvIterator)

// Insert key-value pairs from an iterator
maps.Insert(m, kvIterator)
```

### strings/bytes iterator functions
```go
// Iterate lines (Go 1.23+)
for line := range strings.Lines(text) {
    process(line)
}

// Split as iterator (no allocation of []string)
for part := range strings.SplitSeq(s, ",") {
    process(part)
}

// Fields as iterator
for field := range strings.FieldsSeq(s) {
    process(field)
}
```

---

## Common iterator patterns

### Filter
```go
func Filter[V any](seq iter.Seq[V], pred func(V) bool) iter.Seq[V] {
    return func(yield func(V) bool) {
        for v := range seq {
            if pred(v) {
                if !yield(v) { return }
            }
        }
    }
}
```

### Map / Transform
```go
func Map[In, Out any](seq iter.Seq[In], f func(In) Out) iter.Seq[Out] {
    return func(yield func(Out) bool) {
        for v := range seq {
            if !yield(f(v)) { return }
        }
    }
}
```

### Take
```go
func Take[V any](seq iter.Seq[V], n int) iter.Seq[V] {
    return func(yield func(V) bool) {
        i := 0
        for v := range seq {
            if i >= n { return }
            if !yield(v) { return }
            i++
        }
    }
}
```

### Pull-based iteration with iter.Pull

Convert push iterators to pull-based when you need caller-controlled flow (e.g., comparing two sequences):

```go
func EqualSeq[E comparable](s1, s2 iter.Seq[E]) bool {
    next1, stop1 := iter.Pull(s1)
    defer stop1() // always defer stop — prevents resource leaks
    next2, stop2 := iter.Pull(s2)
    defer stop2()
    for {
        v1, ok1 := next1()
        v2, ok2 := next2()
        if !ok1 || !ok2 { return ok1 == ok2 }
        if v1 != v2 { return false }
    }
}
```

**Pull iterators are slower** than push — they use runtime coroutines. Only use when you genuinely need caller-controlled flow.

---

## Migration strategy

1. Remove `v := v` loop variable workarounds after upgrading to Go 1.22+
2. Replace `for i := 0; i < n; i++` with `for i := range n`
3. Replace `keys := make([]K, 0, len(m)); for k := range m { ... }; sort.Strings(keys)` with `slices.Sorted(maps.Keys(m))`
4. Replace `strings.Split` with `strings.SplitSeq` when iterating (avoids `[]string` allocation)
5. Consider iterators for tree/graph traversals, paginated API results, and lazy pipelines

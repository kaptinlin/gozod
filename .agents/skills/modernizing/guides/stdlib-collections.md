# Standard Library Collections

Modern Go replaces hand-written helpers with generic packages: `slices`, `maps`, `cmp`, and built-in `min`/`max`/`clear`.

## Contents
- Built-in min/max/clear (1.21+)
- slices package (1.21+)
- maps package (1.21+)
- cmp package (1.21+)
- strings/bytes CutPrefix/CutSuffix (1.20+)
- cmp.Or for default values (1.22+)

---

## Built-in min, max, clear (Go 1.21+)

### When to use
- **Always** replace custom `min`/`max` helpers — the builtins work with any ordered type
- Use `clear(m)` to delete all map entries instead of a `for`/`delete` loop
- Use `clear(s)` to zero all elements of a slice (length preserved)
- Value clamping: `v = min(max(v, lo), hi)`

### When NOT to use
- `clear(s)` does **not** shrink a slice — if you need to reset length, use `s = s[:0]`
- `min`/`max` with mixed types requires explicit conversion — they don't auto-promote

```go
// Old
func minInt(a, b int) int {
    if a < b { return a }
    return b
}
x := minInt(a, b)

// New (Go 1.21+)
x := min(a, b)
x = min(a, b, c, d) // variadic

// Old — clear a map
for k := range m { delete(m, k) }

// New (Go 1.21+)
clear(m)

// Clamp a value to [lo, hi]
v = min(max(v, lo), hi)
```

---

## slices package (Go 1.21+)

### When to use
- Replace all custom `contains`, `indexOf`, `remove`, `unique` helpers
- Replace `sort.Slice` / `sort.SliceStable` with `slices.SortFunc` / `slices.SortStableFunc`
- Use `slices.Concat` (1.22+) to join multiple slices
- Use `slices.Repeat` (1.23+) to repeat a slice N times

### When NOT to use
- If you need stable element order during removal, `slices.Delete` reorders — verify behavior
- `slices.Delete`, `slices.Compact`, `slices.Replace` zero removed elements (Go 1.22+) — this is safe but may differ from older hand-written code
- `slices.Insert` panics if index is out of range (Go 1.22+) — validate index first

```go
// Old
func contains(s []string, v string) bool {
    for _, item := range s {
        if item == v { return true }
    }
    return false
}

// New (Go 1.21+)
slices.Contains(s, v)

// Old — sort with custom comparison
sort.Slice(items, func(i, j int) bool {
    return items[i].Name < items[j].Name
})

// New (Go 1.21+)
slices.SortFunc(items, func(a, b Item) int {
    return cmp.Compare(a.Name, b.Name)
})

// Binary search
i, found := slices.BinarySearchFunc(items, target, func(a, b Item) int {
    return cmp.Compare(a.ID, b.ID)
})

// Concatenate slices (Go 1.22+)
all := slices.Concat(a, b, c)
```

---

## maps package (Go 1.21+)

### When to use
- `maps.Clone(m)` — shallow copy of a map (replaces manual loop)
- `maps.Copy(dst, src)` — merge src into dst
- `maps.Equal(a, b)` — compare two maps
- `maps.Keys(m)` / `maps.Values(m)` — returns iterators (Go 1.23+)

### When NOT to use
- `maps.Clone` is a **shallow** copy — nested maps/slices share references
- `maps.Equal` uses `==` — for deep comparison of values, use `maps.EqualFunc`

```go
// Old — clone a map
clone := make(map[string]int, len(m))
for k, v := range m {
    clone[k] = v
}

// New (Go 1.21+)
clone := maps.Clone(m)

// Sorted keys (Go 1.23+)
keys := slices.Sorted(maps.Keys(m))
```

---

## cmp package (Go 1.21+)

### When to use
- `cmp.Compare(a, b)` in sort functions — returns -1, 0, or +1
- `cmp.Ordered` as a type constraint for ordered types
- `cmp.Or(values...)` (Go 1.22+) — returns first non-zero value, useful for defaults

### When NOT to use
- Don't use `cmp.Compare` for floating-point if you need custom NaN handling — it treats NaN as less than other values
- `cmp.Or` short-circuits on the first non-zero value — don't rely on side effects of later arguments

```go
// Default values with cmp.Or (Go 1.22+)
port := cmp.Or(cfg.Port, envPort, 8080)       // first non-zero wins
name := cmp.Or(user.DisplayName, user.Email, "anonymous")

// Multi-field sort
slices.SortFunc(people, func(a, b Person) int {
    return cmp.Or(
        cmp.Compare(a.LastName, b.LastName),
        cmp.Compare(a.FirstName, b.FirstName),
        cmp.Compare(a.Age, b.Age),
    )
})
```

---

## strings/bytes CutPrefix, CutSuffix (Go 1.20+)

### When to use
- Replace `strings.HasPrefix` + `strings.TrimPrefix` pairs
- Cleaner than manual index slicing

### When NOT to use
- If you only need to check (not extract), `strings.HasPrefix` alone is sufficient

```go
// Old
if strings.HasPrefix(s, "Bearer ") {
    token := strings.TrimPrefix(s, "Bearer ")
    use(token)
}

// New (Go 1.20+)
if token, ok := strings.CutPrefix(s, "Bearer "); ok {
    use(token)
}
```

---

## bytes.Clone (Go 1.20+)

### When to use
- When you need to keep a copy of a byte slice that may be reused by the caller (e.g., from `bufio.Scanner`)

```go
// Old
cpy := make([]byte, len(b))
copy(cpy, b)

// New (Go 1.20+)
cpy := bytes.Clone(b)
```

---

## Migration strategy

1. Run `go fix ./...` (Go 1.26+) to auto-apply many of these replacements
2. Search for custom `min`, `max`, `contains`, `indexOf` helpers and replace
3. Replace `sort.Slice`/`sort.Sort` with `slices.SortFunc`
4. Replace manual map cloning with `maps.Clone`
5. Replace `HasPrefix`+`TrimPrefix` pairs with `CutPrefix`

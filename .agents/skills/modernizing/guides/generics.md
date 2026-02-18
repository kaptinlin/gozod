# Generics

Go generics (type parameters) enable type-safe, reusable code. This guide covers when to use them effectively and when to avoid them.

## Contents
- Core principles: when to use generics
- When NOT to use generics
- Type constraints (comparable, cmp.Ordered)
- Generic type aliases (1.24+)
- Self-referential type constraints (1.26+)
- database/sql.Null[T] (1.22+)
- reflect.TypeFor[T] (1.22+)
- Performance considerations

---

## Core principles: when to use generics

Use generics when you find yourself writing **the exact same logic** for different types and the only difference is the type itself.

### Good use cases

1. **Data structures** — generic collections (stacks, queues, trees, ordered maps) that work with any element type
2. **Utility functions on collections** — filter, map, reduce, contains, groupBy (most are now in `slices`/`maps` packages)
3. **Type-safe wrappers** — `sync.Pool`-like wrappers, result types, optional values
4. **Reducing `interface{}` / `any`** — replace `any` fields/returns with type parameters for compile-time safety

```go
// Good: generic data structure
type Stack[T any] struct {
    items []T
}

func (s *Stack[T]) Push(v T) { s.items = append(s.items, v) }
func (s *Stack[T]) Pop() (T, bool) {
    if len(s.items) == 0 {
        var zero T
        return zero, false
    }
    v := s.items[len(s.items)-1]
    s.items = s.items[:len(s.items)-1]
    return v, true
}

// Good: utility function (before slices package existed)
func Contains[T comparable](s []T, v T) bool {
    for _, item := range s {
        if item == v { return true }
    }
    return false
}
// Now use slices.Contains instead
```

---

## When NOT to use generics

### 1. Don't replace interfaces with generics for behavior abstraction
If different types share **behavior** (methods), use interfaces. Generics are for shared **data shape**.

```go
// Bad — generics for behavior
func Process[T interface{ Process() error }](item T) error {
    return item.Process()
}

// Good — interface
func Process(item Processor) error {
    return item.Process()
}
```

### 2. Don't add type parameters "just in case"
Start with concrete types. Add generics only when you have **actual** duplicate code.

```go
// Bad — premature generics
func NewUserService[T UserRepository]() *Service[T] { ... }

// Good — concrete type until proven otherwise
func NewUserService(repo UserRepository) *Service { ... }
```

### 3. Don't use generics for de-virtualization performance
Go's generic implementation uses **GCShape stenciling with dictionaries**, not full monomorphization. All pointer types share the same shape, so method calls on pointer-typed generic parameters use dictionary-based dispatch — potentially **slower** than direct interface calls.

```go
// Bad — slower than interface for pointer types
func CallMethod[T interface{ DoWork() }](v T) { v.DoWork() }

// Good — just use the interface
func CallMethod(v Worker) { v.DoWork() }
```

### 4. Don't over-constrain
Use the **narrowest** constraint that works. Prefer `any` when no operations are needed on the type, `comparable` for map keys and equality checks, `cmp.Ordered` for sorting/comparison.

```go
// Bad — over-constrained
func First[T cmp.Ordered](s []T) T { return s[0] }

// Good — any is sufficient
func First[T any](s []T) T { return s[0] }
```

### 5. Don't use generics when stdlib already provides it
Check `slices`, `maps`, `cmp`, `sync` packages first. Most common generic helpers are already there.

---

## Type constraints

### comparable (built-in)
Use for map keys and `==`/`!=` comparisons. Since Go 1.20, all comparable types satisfy this constraint (including interfaces), but comparison **may panic at runtime** for interface values containing uncomparable types.

### cmp.Ordered (Go 1.21+)
Use for types supporting `<`, `>`, `<=`, `>=`. Covers all integer, float, and string types.

```go
func Max[T cmp.Ordered](a, b T) T {
    if a > b { return a }
    return b
}
// But prefer the built-in max(a, b) instead (Go 1.21+)
```

### Custom constraints
```go
// Union constraint
type Number interface {
    ~int | ~int8 | ~int16 | ~int32 | ~int64 |
    ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
    ~float32 | ~float64
}

// Method + type constraint
type Stringer interface {
    comparable
    String() string
}
```

---

## Generic type aliases (Go 1.24+)

Type aliases can now have type parameters.

### When to use
- Gradually migrating types between packages without breaking callers
- Creating shorter names for verbose generic instantiations

### When NOT to use
- Don't use aliases to "rename" types across packages unless you're doing a migration
- Aliases share identity with their target — `MyList[int]` **is** `[]int`, not a distinct type

```go
// Generic type alias (Go 1.24+)
type List[T any] = []T
type Pair[A, B any] = struct{ First A; Second B }

// Package migration alias
type OldMap[K comparable, V any] = newpkg.Map[K, V]
```

---

## Self-referential type constraints (Go 1.26+)

Generic types can reference themselves in constraints — enables patterns like the "self type" or "CRTP" (curiously recurring template pattern).

### When to use
- Fluent builder APIs where methods return the concrete type
- Algebraic type patterns (e.g., `Add(Self) Self`)

### When NOT to use
- Simple cases — this is an advanced pattern that adds complexity
- If an interface with concrete types works, prefer that

```go
// Self-referential constraint (Go 1.26+)
type Adder[A Adder[A]] interface {
    Add(A) A
}

type Vector2D struct{ X, Y float64 }

func (v Vector2D) Add(other Vector2D) Vector2D {
    return Vector2D{v.X + other.X, v.Y + other.Y}
}

func Sum[T Adder[T]](items []T) T {
    var result T
    for _, item := range items {
        result = result.Add(item)
    }
    return result
}
```

---

## database/sql.Null[T] (Go 1.22+)

Generic nullable type replacing `NullString`, `NullInt64`, etc.

### When to use
- All new nullable database fields — cleaner and type-safe
- Any nullable type (not just strings and ints)

### When NOT to use
- Existing code using `NullString` etc. works fine — migrate gradually

```go
// Old
var name sql.NullString
var age sql.NullInt64
err := row.Scan(&name, &age)

// New (Go 1.22+)
var name sql.Null[string]
var age sql.Null[int]
err := row.Scan(&name, &age)

// Custom types work too
var ts sql.Null[time.Time]
```

---

## reflect.TypeFor[T] (Go 1.22+)

Replaces the `reflect.TypeOf((*T)(nil)).Elem()` idiom.

### When to use
- Always prefer over the old idiom — cleaner and avoids the nil-pointer trick

```go
// Old
t := reflect.TypeOf((*MyInterface)(nil)).Elem()

// New (Go 1.22+)
t := reflect.TypeFor[MyInterface]()
```

---

## Decision flowchart

```
Is the code identical except for types?
  YES → Operating on built-in containers or data structures?
          YES → Use generics
          NO  → Only calling methods on values?
                  YES → Use an interface
                  NO  → Behavior differs per type?
                          YES → Use interfaces (or reflection if types lack methods)
                          NO  → Use generics
  NO  → Use concrete types or interfaces
```

---

## Performance considerations

Go generics use **GCShape stenciling** — all pointer types share one implementation, non-pointer types get separate implementations. This means:

- **Value types** (int, float64, structs): fully monomorphized, performance comparable to hand-written code
- **Pointer types**: method calls go through a runtime dictionary (indirect call) — may be slower than interface calls
- **Interface arguments**: passing an interface to a generic function triggers expensive `runtime.assertI2I` lookups

### Measured overhead (PlanetScale benchmarks)

| Approach | Latency |
|----------|---------|
| Direct (monomorphized) | 5.06 us |
| Interface-based | 6.85 us |
| Generic with pointer arg | 7.18 us |
| Generic with exact interface constraint | 9.68 us |
| Generic with super-interface constraint | 17.6 us |

### Guidance
- For hot paths with pointer types, **benchmark** generic vs interface versions
- For data structures storing values (not pointers), generics perform well
- For simple functional helpers that inline, generics are fine
- Don't use generics to try to "speed up" interface method dispatch — it doesn't work that way
- Prefer functions over methods for comparisons — requiring a `Less()` method forces users to wrap simple types like `int`

---

## Migration strategy

1. Replace custom collection helpers with `slices`/`maps` packages (Go 1.21+)
2. Replace `sql.NullString` etc. with `sql.Null[T]` (Go 1.22+)
3. Replace `reflect.TypeOf((*T)(nil)).Elem()` with `reflect.TypeFor[T]()` (Go 1.22+)
4. Introduce generics for **actual** duplicate code, not speculatively
5. Prefer stdlib generics over rolling your own

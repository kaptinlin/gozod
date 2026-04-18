# Go Internal Package Refactoring

The `internal/` directory is one of Go's most powerful boundary tools, but it
is easy to misuse. These patterns surface repeatedly when refactoring libraries
that have grown adapter layers, sub-packages, or protocol bindings.

## First Principles

1. **`internal/` exists to keep things out of someone else's API**, not to
   organize code by file size.
2. **Imports must flow in one direction**: parent → internal, never the
   reverse. The compiler does not enforce this; the discipline does.
3. **Adapters belong inside `internal/`**. Every third-party library binding
   should sit behind an adapter package whose only public API is your own
   domain types.
4. **One owner per concept**. If a type, sentinel, or interface lives in two
   packages, drift is inevitable.

---

## Anti-Pattern: Reverse Imports from `internal/`

```go
// internal/dns/resolver.go — internal package importing the parent
package dns

import "github.com/example/mylib" // ← reverse import

func Lookup(name string) (mylib.Status, error) {
    // ...
}
```

This compiles, but it inverts the dependency direction. The parent package
defines `mylib.Status`; the internal package consumes it. When the parent
later wants to depend on the internal package (which is the whole point of
having an internal package), you have a cycle waiting to happen, and you
can never split the parent's `Status` definition without touching every
internal subpackage.

**Symptom**: `import "github.com/example/mylib"` appears anywhere under
`internal/`. Or worse, the same parent type is referenced by 4 different
internal subpackages, each forming an independent dependency edge back to
the parent.

**Fix**: introduce a shared core package that both sides can depend on.

```go
// internal/core/types.go — leaf package, depends on nothing
package core

type Status int

const (
    StatusOK Status = iota
    StatusFail
)

var ErrNotFound = errors.New("mylib: not found")
```

```go
// mylib.go — re-exports from core
package mylib

import "github.com/example/mylib/internal/core"

type Status = core.Status

const (
    StatusOK   = core.StatusOK
    StatusFail = core.StatusFail
)

var ErrNotFound = core.ErrNotFound
```

```go
// internal/dns/resolver.go — depends on core, not the parent
package dns

import "github.com/example/mylib/internal/core"

func Lookup(name string) (core.Status, error) {
    // ...
}
```

Now the dependency graph is a tree:

```
mylib (parent)
   ↓
internal/core (leaf — no internal imports)
   ↑
internal/dns, internal/cache, internal/adapter (siblings, no edges between them)
```

Public API is preserved via type aliases. Callers see `mylib.Status` exactly
as before; the parent re-export is invisible at the call site.

---

## Anti-Pattern: Mirror Tables Across Layers

A common drift source: the public package and an `internal/` adapter each
define their own version of "the same thing".

```go
// internal/adapter/errors.go
var ErrTimeout = errors.New("adapter: timeout")

// errors.go (public)
var ErrTimeout = errors.New("mylib: timeout")

// dispatch.go
func toPublic(err error) error {
    if errors.Is(err, internal.ErrTimeout) {
        return ErrTimeout
    }
    // ... 11 more cases
}
```

Two declarations of "timeout" means two messages, two prefixes, two test
fixtures, and a remap function that exists only to translate between them.

**Fix**: pick a single owner — usually the leaf package — and re-export.

```go
// internal/core/errors.go — single owner
var ErrTimeout = errors.New("mylib: timeout")

// errors.go (public)
var ErrTimeout = core.ErrTimeout

// dispatch.go
func toPublic(err error) error {
    return err // identity — sentinels already match
}
```

The dispatch function disappears. `errors.Is` works across the boundary
because the sentinels are literally the same variable, not two equal-looking
copies.

This applies equally to status enums, configuration types, option types, and
any value-level identity that crosses the internal boundary.

---

## Pattern: Adapter Package per Third-Party Binding

A library that depends on `github.com/foo/parser` should never expose
`parser.Document` in its public API. Wrap the binding inside an internal
adapter so the third-party type is invisible to callers.

```
mylib/
├── doc.go            # Public Document type, owned by mylib
├── parse.go          # Public Parse function
└── internal/
    └── parseradapter/
        ├── adapter.go  # Translates parser.Document → mylib.Document
        └── errors.go   # Translates parser errors → mylib sentinels
```

```go
// internal/parseradapter/adapter.go
package parseradapter

import (
    "github.com/foo/parser"
    "github.com/example/mylib/internal/core"
)

func Parse(input []byte) (core.Document, error) {
    raw, err := parser.Parse(input)
    if err != nil {
        return core.Document{}, classifyErr(err)
    }
    return convert(raw), nil
}

// convert is the only place that touches parser.Document.
func convert(raw *parser.Document) core.Document {
    return core.Document{
        Title:   raw.Title,
        Body:    raw.Body,
        Updated: raw.Modified,
    }
}
```

**Benefits**
- Replacing `github.com/foo/parser` with a different library is one file change.
- Upstream type breakage cannot reach your callers.
- Tests for the adapter live next to it and exercise the conversion in
  isolation.
- Bug reports against the upstream library are localized — you have one
  file to point at when filing issues.

**Discipline**: the adapter package's exported surface should contain *only*
your domain types. If `parser.Document` appears in any exported function
signature inside `internal/parseradapter/`, the abstraction is leaking.

---

## Pattern: `internal/core` for Shared Vocabulary

Once a library has more than one internal subpackage, the question of where
to put shared types becomes a design decision. The answer is almost always:
in a leaf `internal/core` package that depends on nothing.

```
mylib/
├── api.go              # imports internal/core, internal/foo, internal/bar
└── internal/
    ├── core/           # leaf — no internal imports, no parent imports
    │   ├── status.go   # Status enum
    │   ├── errors.go   # Sentinel errors
    │   └── types.go    # Shared value types
    ├── foo/            # imports internal/core
    └── bar/            # imports internal/core
```

**Why a separate package and not just files in `internal/`?**

Files in `internal/` form a single package. That single package becomes the
import target for both `internal/foo` and `internal/bar`, which is fine until
that package needs to import its own subpackage — then you have a cycle.
A leaf `internal/core` package keeps the dependency graph acyclic by
construction.

**Naming**: `core`, `common`, `shared`, `types` — pick one and use it
consistently. Avoid `util` (vacuous) and `internal/internal` (silly).

---

## Anti-Pattern: `internal/` as a Junk Drawer

```
mylib/
└── internal/
    ├── helpers.go      # 800 lines of unrelated functions
    ├── stuff.go        # more unrelated functions
    └── misc.go         # the things that didn't fit
```

A flat `internal/` package with thousands of lines is a sign that nothing
inside has a reason to live there beyond "we didn't want to export it."
This package becomes the bottleneck for every refactor — touching one helper
forces re-running all the package's tests.

**Fix**: split by responsibility, the same way you would split a public
package. Each `internal/X/` should have a one-sentence package doc that
describes its single job.

```
internal/
├── auth/        # SASL mechanism implementations
├── tls/         # TLS config helpers
├── retry/       # backoff and classification
└── parser/      # protocol response parsing
```

If you cannot write a single sentence describing what an internal package
does, it doesn't deserve to be a package.

---

## Pattern: Re-Export via Type Alias

When the parent package wants to expose a type defined in `internal/core`,
use Go's type alias syntax (`type X = core.X`), not a wrapper struct.

```go
// Good
type Status = core.Status

const (
    StatusOK   = core.StatusOK
    StatusFail = core.StatusFail
)
```

```go
// Bad — wrapper struct breaks identity
type Status struct {
    inner core.Status
}
```

The alias preserves *value identity*: `mylib.Status(0) == core.Status(0)`,
`errors.Is(myerr, mylib.ErrFoo) == errors.Is(myerr, core.ErrFoo)`, and
`fmt.Sprintf("%T", val)` shows the alias name. The wrapper struct creates
a parallel type with no relationship to the original, breaking every
`errors.Is` and `==` comparison that crosses the boundary.

**Constraint**: type aliases are not the same as defined types. You cannot
attach methods to an alias from outside the original package. This is fine
for sentinels and value types; if you need methods, the type must be defined
in the package that owns the methods.

---

## Verification Checklist

After any internal-package refactor, run through this list:

- [ ] No file under `internal/` imports the parent module path.
      (`grep -rn 'import .*github.com/example/mylib"' internal/`)
- [ ] Every third-party library is referenced from exactly one
      `internal/<name>adapter/` package.
- [ ] No third-party type appears in any exported signature outside of
      `internal/`.
- [ ] Sentinels and shared types live in exactly one place; the public
      package re-exports via `var` or `type X = ...`.
- [ ] Each `internal/X/` package has a one-sentence doc comment describing
      its single responsibility.
- [ ] No internal package contains a `helpers.go` or `util.go` file with
      unrelated functions.
- [ ] The dependency graph from public package down to `internal/core` is
      acyclic and forms a tree, not a mesh.

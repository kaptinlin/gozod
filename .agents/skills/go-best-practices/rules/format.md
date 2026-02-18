# Formatting

Consistent formatting via gofmt, import ordering, and literal conventions ensure uniform code appearance and reduce review friction.

## Contents
- Always Use gofmt
- Import Grouping and Ordering
- Import Renaming Conventions
- Struct Literal Field Names
- Prefer Nil Slices
- Function Formatting
- Variable Declaration Conventions
- Conditions and Control Flow

---

## Always Use gofmt

All Go source files must be formatted with `gofmt` or `goimports`. Use tabs for indentation. No exceptions. Generated code should also be formatted.

```bash
gofmt -w .
goimports -w .
```

---

## Import Grouping and Ordering

Organize imports in groups separated by blank lines: (1) standard library, (2) third-party/project packages, (3) protocol buffer imports, (4) side-effect imports.

**Incorrect:**

```go
import (
    "fmt"
    "github.com/dsnet/compress/flate"
    foopb "myproj/foo/proto/proto"
    "os"
    _ "myproj/rpc/protocols/dial"
    "golang.org/x/text/encoding"
)
```

**Correct:**

```go
import (
    "fmt"
    "hash/adler32"
    "os"

    "github.com/dsnet/compress/flate"
    "golang.org/x/text/encoding"
    "google.golang.org/protobuf/proto"

    foopb "myproj/foo/proto/proto"

    _ "myproj/rpc/protocols/dial"
    _ "myproj/security/auth/authhooks"
)
```

Within each group, imports are alphabetically sorted (as `goimports` does).

---

## Import Renaming Conventions

Only rename imports to avoid conflicts or fix uninformative names. Proto packages use `pb` suffix. Never use dot imports.

```go
// Proto renaming:
import (
    foopb "path/to/package/foo_service_go_proto"
    foogrpc "path/to/package/foo_service_go_grpc"
)

// Conflict resolution:
import (
    urlpkg "net/url" // when 'url' clashes with local variable
)
```

**Incorrect (dot import hides origin):**

```go
import . "foo"
var myThing = Bar() // Where does Bar come from?
```

Side-effect imports (`import _ "pkg"`) are only allowed in `main` packages or tests.

---

## Struct Literal Field Names

Always specify field names for struct literals of types defined outside the current package. Omit zero-value fields when it doesn't hurt clarity.

**Incorrect:**

```go
r := csv.Reader{',', '#', 4, false, false, false, false}

ldb := leveldb.Open("/my/table", &db.Options{
    BlockSize: 1<<16, ErrorIfDBExists: true,
    BlockRestartInterval: 0, Comparer: nil, Compression: nil, MaxOpenFiles: 0,
})
```

**Correct:**

```go
r := csv.Reader{
    Comma:           ',',
    Comment:         '#',
    FieldsPerRecord: 4,
}

ldb := leveldb.Open("/my/table", &db.Options{
    BlockSize:       1<<16,
    ErrorIfDBExists: true,
})
```

Omit repeated element types in slice/map literals:

```go
items := []*Type{{A: 42}, {A: 43}}
```

---

## Prefer Nil Slices

Use nil initialization for empty slices. Don't force callers to distinguish nil from empty. Use `len()` to check emptiness.

**Incorrect:**

```go
t := []string{}

if s == nil { return }
```

**Correct:**

```go
var t []string

if len(s) == 0 { return }
```

`nil` slices work with `len()`, `cap()`, `append()`, and `range`.

---

## Function Formatting

Keep function signatures on one line. When calls are long, extract local variables rather than splitting arguments across lines. Don't add inline comments to call arguments.

**Incorrect:**

```go
func (r *SomeType) SomeLongFunctionName(foo1, foo2, foo3 string,
    foo4, foo5, foo6 int) {
    foo7 := bar(foo1)
}

bad := server.New(
    ctx,
    42, // Port
)
```

**Correct:**

```go
local := helper(some, parameters, here)
good := foo.Call(list, of, parameters, local)

good := server.New(ctx, server.Options{Port: 42})
```

---

## Variable Declaration Conventions

Use `:=` for non-zero initialization. Use `var` for zero-value declarations. Use `new()` for pointer-to-zero-value.

```go
// Non-zero — use :=
i := 42

// Zero-value — use var
var (
    coords Point
    magic  [4]byte
    primes []int
)

// Pointer to zero value
buf := new(bytes.Buffer)
msg := new(pb.Message)
```

Maps must be explicitly initialized before modification, but reading from a zero-value map is safe.

---

## Conditions and Control Flow

Don't wrap `if` conditions across multiple lines. Extract complex conditions into named local variables. No Yoda conditions. No redundant `break` in `switch`.

**Incorrect:**

```go
if db.CurrentStatusIs(db.InTransaction) && db.ValuesEqual(db.TransactionKey(), row.Key()) {
}

if "foo" == result { ... }

case "A":
    buf.WriteString(x)
    break // redundant in Go
```

**Correct:**

```go
inTransaction := db.CurrentStatusIs(db.InTransaction)
keysMatch := db.ValuesEqual(db.TransactionKey(), row.Key())
if inTransaction && keysMatch {
}

if result == "foo" { ... }
```

Go `switch` cases auto-break. Use `fallthrough` explicitly when C-style fall-through is needed.

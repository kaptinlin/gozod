# Documentation

Effective documentation through godoc comments, parameter docs, and signal boosting reduces misuse and speeds onboarding.

## Contents
- Documentation Comment Conventions
- Package Comments
- Documenting Parameters
- Document Cleanup and Error Types
- Signal Boosting for Subtle Code

---

## Documentation Comment Conventions

All exported names must have doc comments starting with the symbol name as a complete sentence. Unexported types/functions with non-obvious behavior should too.

**Incorrect:**

```go
// This function returns the absolute value of x.
func Abs(x float64) float64
```

**Correct:**

```go
// Abs returns the absolute value of x.
func Abs(x float64) float64
```

**Correct (struct documentation with field groups):**

```go
// Options configure the group management service.
type Options struct {
    // General setup:
    Name  string
    Group *FooGroup

    // Dependencies:
    DB *sql.DB

    // Customization:
    LargeGroupThreshold int // optional; default: 10
    MinimumMembers      int // optional; default: 2
}
```

Keep comment lines readable at ~80 columns. Full sentences get capitalization and punctuation; sentence fragments do not.

---

## Package Comments

Every package must have exactly one package comment, directly above the `package` clause with no blank line between.

```go
// Package math provides basic constants and mathematical functions.
//
// This package does not guarantee bit-identical results across architectures.
package math
```

```go
// The seed_generator command is a utility that generates a Finch seed file
// from a set of JSON study configs.
package main
```

For multi-file packages, the package comment only needs to appear in one file (usually the primary file).

---

## Documenting Parameters

Only document parameters that are error-prone or non-obvious. Don't restate what's already clear from types. Don't restate that context cancellation returns an error — it's implied.

**Incorrect (restates the obvious):**

```go
// Sprintf formats and returns the result string.
//
// format is the format, data is the interpolation data.
func Sprintf(format string, data ...any) string
```

**Correct (documents non-obvious behavior):**

```go
// Sprintf formats and returns the result string.
//
// The provided data is used to interpolate the format string. If data does not
// match the expected format verbs, the function will inline format error
// warnings in the output string.
func Sprintf(format string, data ...any) string
```

**Incorrect (restates context cancellation):**

```go
// Run executes the worker's run loop.
//
// The method will process work until the context is canceled and return an error accordingly.
func (Worker) Run(ctx context.Context) error
```

**Correct:**

```go
// Run executes the worker's run loop.
func (Worker) Run(ctx context.Context) error
```

---

## Document Cleanup and Error Types

Always document explicit cleanup requirements. Document error sentinel values and types callers can handle.

```go
// NewTicker returns a new Ticker containing a channel that will send the
// current time on the channel after each tick.
//
// Call Stop when done to release the Ticker's associated resources.
func NewTicker(d Duration) *Ticker

// Read reads up to len(b) bytes from the File and stores them in b.
//
// At end of file, Read returns 0, io.EOF.
func (*File) Read(b []byte) (n int, err error)

// Chdir changes the current working directory to the named directory.
//
// If there is an error, it will be of type *PathError.
func Chdir(dir string) error
```

---

## Signal Boosting for Subtle Code

Add comments to draw attention when code looks standard but behaves differently. The classic example is `err == nil` (not `err != nil`).

```go
if err := doSomething(); err == nil { // if NO error
    // ...
}
```

Also boost non-obvious assignments:

```go
// Gregorian leap years are not just year%4 == 0.
// See https://en.wikipedia.org/wiki/Leap_year#Algorithm.
var (
    leap4   = year%4 == 0
    leap100 = year%100 == 0
    leap400 = year%400 == 0
)
leap := leap4 && (!leap100 || leap400)
```

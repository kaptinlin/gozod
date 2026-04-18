# Go Error Handling Refactoring

Error handling patterns and anti-patterns for Go 1.20+. These principles surface
repeatedly across long-lived libraries; treat them as a checklist when reviewing
or refactoring any package that returns errors.

## First Principles

1. **Errors are an API**. Their construction, classification, and identity are
   public contract — treat them with the same care as exported types.
2. **`errors.Is` is the only stable form of equality**. If a sentinel exists, a
   call site must be able to match it. Otherwise the sentinel is a lie.
3. **Honesty over convenience**. An error message must accurately describe
   *where* the failure happened. Subsystem labels are not decoration.
4. **One way to express each fact**. The same failure should not be wrapped
   twice, classified twice, or represented in two parallel sentinel tables.

---

## Anti-Pattern: Dead Sentinels

A sentinel that is *defined* and *exported* but never *produced* is a contract
violation. Callers see it in the docs and write `errors.Is(err, ErrFoo)`, which
silently always returns false.

**Symptoms**
- Test asserts `errors.Is(err, ErrFoo)` but fails on real input.
- Grep finds `ErrFoo` only in `errors.go` and tests, never in production paths.
- Doc comment says "returned when X happens" but no code produces it.

**Fix**
- Either *activate* the sentinel — find the path that should produce it and
  wire it up — or *delete* it. There is no third option.
- Add a coverage test that asserts the sentinel is reachable from at least one
  call site, ideally driven by table data.

```go
// reachability test — fails if anyone deletes the production path
func TestErrFooReachable(t *testing.T) {
    err := triggerKnownFooFailure()
    if !errors.Is(err, ErrFoo) {
        t.Fatalf("ErrFoo no longer reachable: %v", err)
    }
}
```

---

## Anti-Pattern: Sentinel Mirror Tables

When an `internal/` adapter package and the public package each define their
own copy of "the same" sentinel, drift is guaranteed. The dispatch layer ends
up doing `errors.Is(err, internal.ErrX) → return ErrX`, the same fact stated
twice, with a third stage of wrapping that says it once more.

**Anti-pattern**
```go
// internal/provider.go
var ErrAuthFailed = errors.New("provider: auth failed")

// errors.go (public)
var ErrAuthFailed = errors.New("mylib: auth failed")

// session.go
func dispatch(err error) error {
    switch {
    case errors.Is(err, internal.ErrAuthFailed):
        return ErrAuthFailed
    // ... 11 more cases ...
    }
}
```

**Fix**: Pick a single owner. The public package re-exports via `var`:

```go
// internal/provider.go — single source of truth
var ErrAuthFailed = errors.New("mylib: auth failed")

// errors.go (public)
var ErrAuthFailed = internal.ErrAuthFailed
```

Now `errors.Is` traverses adapter boundaries unchanged. The dispatch
remap disappears. The two sentinels are literally the same variable.

**Why this works**: Go's `var X = Y` creates an alias at the value level, not
a new error. `errors.Is` compares pointers (via the underlying `errors.errorString`
identity), so the alias is indistinguishable from the original.

---

## Anti-Pattern: Reverse Wrapping

```go
// Anti-pattern: sentinel-first wrapping reads backwards
return fmt.Errorf("%w: reading config file %q", ErrConfigInvalid, path)
```

This produces `mylib: invalid config: reading config file "x"`, which puts
the *cause* after the *effect* and forces grep to scan past the sentinel
prefix to find context.

**Convention**
```go
return fmt.Errorf("config: read %q: %w", path, ErrConfigInvalid)
```

**Rules**
- Subsystem label first.
- Specific context (operation, target) next.
- Wrapped sentinel last, with `%w`.
- One `%w` per `Errorf` call. Never two (see below).

---

## Pattern: Multi-`%w` is for Genuine Multi-Cause, Not Diagnostic Context

`fmt.Errorf` has supported multiple `%w` since Go 1.20. The feature exists for a
specific use case: combining two errors that *both* belong in the chain because
both represent real failures the caller might want to match.

**Legitimate use** — operation and cleanup both fail, both are public API:

```go
// Both errors deserve to be in the chain. Caller may want to branch on either.
return fmt.Errorf("write %q: %w; cleanup also failed: %w", path, writeErr, closeErr)
```

**Misuse** — one of the `%w` is just diagnostic context that doesn't belong in
the chain:

```go
// Anti-pattern: ErrParseFail is the public sentinel; upstreamErr is third-party
// noise that should not become part of your public API surface.
return fmt.Errorf("%w: %w", ErrParseFail, upstreamErr)
```

The misuse pattern surfaces upstream library types into your `errors.Is` chain,
makes wrapping order load-bearing for `errors.Is` traversal, and forces every
caller to discover that they can now match against errors from a library they
never imported.

**Decision rule** for each `%w` candidate: would a caller reasonably want to
`errors.Is(err, candidate)`? If yes, `%w` is correct. If no, use `%v` or `%s`
to keep the value as a string only.

```go
// Strategy A: upstream is opaque — flatten with %v
return fmt.Errorf("parse: decode body: %v: %w", upstreamErr, ErrParseFail)

// Strategy B: upstream is part of the public chain — single %w
return fmt.Errorf("parse: decode body: %w", upstreamErr)

// Strategy C: both are public — multi-%w is the intended use
return fmt.Errorf("commit: tx %w; rollback also failed: %w", txErr, rollbackErr)
```

**Use `errors.Join` instead** when accumulating an unbounded list of errors
(validation loop, parallel goroutines, batch processing) — `errors.Join` is
the right tool for "list of independent errors", not `fmt.Errorf` with N
`%w` verbs.

```go
var errs []error
for _, item := range items {
    if err := validate(item); err != nil {
        errs = append(errs, err)
    }
}
return errors.Join(errs...) // returns nil if errs is empty or all-nil
```

---

## Anti-Pattern: Stdlib Internal Pollution

Wrapping `net/mail`, `encoding/json`, `crypto/x509`, `pem`, or any stdlib
parser error with `%w` exposes its internal error types as part of your
public `errors.Is` chain. Callers start matching against unexported stdlib
sentinels that may change between Go versions.

**Anti-pattern**
```go
addr, err := mail.ParseAddress(raw)
if err != nil {
    return fmt.Errorf("invalid address: %w", err) // leaks net/mail internals
}
```

**Fix**: format as `%v` and carry the raw input as diagnostic context.

```go
addr, err := mail.ParseAddress(raw)
if err != nil {
    return fmt.Errorf("address: parse %q: %v", raw, err)
}
```

The caller still sees the cause in the error string, but the chain is now
yours to control.

**Exception**: stdlib errors that are *intentionally* part of public Go
contract — `io.EOF`, `context.Canceled`, `context.DeadlineExceeded`,
`fs.ErrNotExist`, `*net.OpError`, `*net.DNSError` — should pass through with
`%w` because callers reasonably depend on matching them.

---

## Anti-Pattern: Cause Loss in Fallback

A wrapper helper that constructs a typed error from an unknown input often
forgets to preserve the cause on the fallback branch.

```go
func wrapTemporary(err error) *MyError {
    var typed *MyError
    if errors.As(err, &typed) {
        return typed // good
    }
    return &MyError{
        Kind:    Temporary,
        Message: err.Error(), // BUG: loses Cause
    }
}
```

After this, `errors.Is(wrapped, context.DeadlineExceeded)` returns false even
though the underlying error *was* a deadline. Every chain-walking function
breaks silently.

**Fix**
```go
return &MyError{
    Kind:    Temporary,
    Message: err.Error(),
    Cause:   err, // preserve the chain
}
```

If `*MyError` implements `Unwrap() error` returning `Cause`, the chain stays
intact and `errors.Is` walks through.

---

## Anti-Pattern: Closure Nil-Check Trap

When a `Hooks` struct has optional callback fields, a "skip if nil" helper
that takes a closure cannot tell whether the closure's body is a no-op.

```go
// Anti-pattern: nil check happens too late
func invokeHook(fn func()) {
    defer func() { recover() }()
    if fn == nil { return } // always non-nil — fn is a closure literal
    fn()
}

func (h *Hooks) fireBeforeSend(info Info) {
    invokeHook(func() { h.BeforeSend(info) }) // closure is non-nil even if h.BeforeSend is nil
}
```

This silently invokes a nil method value, which panics, which `recover` swallows.
With panic logging enabled, every Send produces three "recovered panic" log
lines on a zero-config caller.

**Fix**: nil-check the *field*, before constructing the closure.

```go
func (h *Hooks) fireBeforeSend(info Info) {
    if h.BeforeSend == nil {
        return
    }
    defer func() {
        if r := recover(); r != nil {
            slog.Error("hook panic recovered", "hook", "BeforeSend", "panic", r)
        }
    }()
    h.BeforeSend(info)
}
```

The wrapper now correctly does nothing for the zero-config case, at zero cost.

---

## Anti-Pattern: Silent `recover`

Bare `defer func() { recover() }()` swallows panics with no trace. The
behavior of "hooks must not crash the caller" is correct, but turning a
real bug into invisible silence is worse than the panic.

**Fix**: log every recovered panic with structured context.

```go
defer func() {
    if r := recover(); r != nil {
        slog.Error("mylib: hook panic recovered",
            "hook", name,
            "panic", r,
            "stack", string(debug.Stack()),
        )
    }
}()
```

Operators can suppress these logs at the slog level if they really want them
gone, but the default surfaces the bug.

---

## Anti-Pattern: Subsystem Label Drift

```go
// parse.go
return fmt.Errorf("mylib: mylib: invalid input: %w", err)
//                  ^^^^^^^^^^^^^^ double prefix
```

Or worse:

```go
// http_handler.go
return fmt.Errorf("database: %w", upstreamErr) // upstream is HTTP, not DB
```

These happen when refactoring moves code between subsystems and labels lag
behind. The user sees a database error from an HTTP handler and wastes hours
chasing the wrong subsystem.

**Rules**
- The subsystem label must name the package or component *currently* doing
  the work, not the eventual destination or the historical owner.
- Never include the parent package's prefix in a wrap — `Errorf` adds the
  prefix exactly once per `errors.Errorf` call, and the wrapped sentinel
  already carries its own prefix.
- After a refactor that moves a function across packages, grep for the old
  prefix. It will haunt the new file.

---

## Pattern: Context Cancellation Classification

`IsTemporary(error)` and `IsPermanent(error)` helpers must classify
`context.Canceled` and `context.DeadlineExceeded` correctly, or callers that
wrap your library in their own retry loop will mis-classify cancellations as
permanent failures and stop retrying user-initiated cancels (which is fine)
or stop retrying genuine deadline exceeds (which silently breaks).

```go
func IsTemporary(err error) bool {
    if err == nil {
        return false
    }
    if errors.Is(err, context.DeadlineExceeded) {
        return true // deadline is retryable with a fresh ctx
    }
    if errors.Is(err, context.Canceled) {
        return false // user cancellation is not retryable
    }
    var t Classifier
    if errors.As(err, &t) {
        return t.Temporary()
    }
    return false
}
```

**Why both directions matter**: `context.Canceled` is *not* temporary — the
caller asked to stop. Treating it as retryable creates infinite loops on
HTTP cancels. `context.DeadlineExceeded` *is* temporary — the deadline was
the caller's policy, not a permanent server state.

---

## Pattern: Constructor Functions for Typed Errors

Scattered `&MyError{Kind: ..., Message: ...}` literals across 30 call sites
guarantee drift. Some forget to set `Cause`, some forget `Sentinel`, some use
the wrong `Kind`.

**Anti-pattern**
```go
return &MyError{Kind: KindValidation, Message: "address required"}
return &MyError{Message: "address required", Kind: KindValidation} // field order drift
return &MyError{Kind: ValidationKind, Message: fmt.Sprintf("%s required", field)} // typo, never compiles, but the literal pattern enables it
```

**Fix**: small constructors per category.

```go
func newValidationError(msg string) *MyError {
    return &MyError{Kind: KindValidation, Message: msg}
}

func newAuthError(cause error) *MyError {
    return &MyError{
        Kind:     KindConfiguration,
        Message:  "authentication failed",
        Cause:    cause,
        Sentinel: ErrAuthFailed,
    }
}
```

Now adding a new field to `MyError` is one edit per constructor, not 30 edits
across the codebase. The constructors also become the place where invariants
live ("auth errors always carry the AuthFailed sentinel").

Constructors should be unexported by default. Export only when sub-packages
need to participate in the same sentinel scheme — this is the case for
provider adapter packages that must produce sentinels matching the parent
package's API.

**Use `errors.AsType[T]` (Go 1.26+) when extracting typed errors** in the
constructor's caller. The generic form is type-safe at compile time, faster
than reflection-based `errors.As`, and migrates automatically via `go fix`:

```go
// Go 1.26+
if myErr, ok := errors.AsType[*MyError](err); ok {
    if myErr.Kind == KindValidation {
        // ...
    }
}

// Pre-1.26 — fall back to errors.As
var myErr *MyError
if errors.As(err, &myErr) {
    // ...
}
```

---

## Pattern: Add Non-Redundant Context

When wrapping an error, add information the caller's context provides that the
underlying error doesn't already say. Avoid restating what the error already
contains.

**Anti-pattern** — duplicates the path that `os.Open`'s error already includes:

```go
if err := os.Open("settings.txt"); err != nil {
    return fmt.Errorf("could not open settings.txt: %w", err)
}
// Output: could not open settings.txt: open settings.txt: no such file or directory
//                                       ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
//                                       already in the error
```

**Fix** — add domain meaning, not file-name echo:

```go
if err := os.Open("settings.txt"); err != nil {
    return fmt.Errorf("launch codes unavailable: %w", err)
}
// Output: launch codes unavailable: open settings.txt: no such file or directory
```

**Rule of thumb**: if the wrapping prefix could be deleted with no information
loss, the wrap is wrong. Either delete the wrap entirely (just `return err`) or
replace the prefix with something that explains *why this code path cared* that
the operation failed.

**The simplest correct wrap is no wrap**: `return err` is better than
`fmt.Errorf("failed: %w", err)`. The latter adds noise and zero information.

---

## Pattern: Diagnostic Context in Stage-Wise Wrapping

A multi-step operation should label *which step* failed, not just *that*
something failed.

**Anti-pattern**
```go
func fetchAndParse(url string) (*Doc, error) {
    resp, err := http.Get(url)
    if err != nil {
        return nil, fmt.Errorf("fetch: %w", err)
    }
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("fetch: %w", err) // same label for two stages
    }
    doc, err := parse(body)
    if err != nil {
        return nil, fmt.Errorf("fetch: %w", err) // wrong label
    }
    return doc, nil
}
```

**Fix**
```go
func fetchAndParse(url string) (*Doc, error) {
    resp, err := http.Get(url)
    if err != nil {
        return nil, fmt.Errorf("doc: http get %q: %w", url, ErrTransport)
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("doc: read body: %w", ErrTransport)
    }

    parsed, err := parse(body)
    if err != nil {
        return nil, fmt.Errorf("doc: parse: %w", ErrInvalidContent)
    }
    return parsed, nil
}
```

Each step has its own label *and* its own sentinel where the failure mode
differs. Operators reading logs can immediately tell whether the network
or the content was the problem, without firing up the debugger.

---

## Pattern: `Result.Issues` for Non-Fatal Problems

A parser or normalizer that finds *some* problems but produces *some* output
should not return error. Errors are for "no result"; non-fatal issues belong
on the result.

```go
type Result struct {
    Records []Record
    Issues  []Issue // warnings, malformed entries, recoverable parse errors
}

type Issue struct {
    Path        string // location of the issue (line, field, sub-document)
    Description string
    Severity    Severity // Warning, Severe
    Err         error    // optional, for sentinel matching
}
```

**Rules**
- Fatal failure (no records produced) → return `Result{}, err`.
- Partial success (some records, some problems) → return `Result{Records:..., Issues:...}, nil`.
- Never mix the two channels. A non-nil error with a populated `Records` slice
  forces every caller to handle "did anything come through?" twice.
- Issues are *not* errors in the Go sense. Don't make `Issue` implement
  `error` — it tempts callers to `return issue` from a function that should
  succeed.

This pattern is essential for parsers consuming user-supplied data (mail,
config files, RSS feeds, CSV imports) where halting on the first problem
is worse than reporting all problems and continuing.

---

## Pattern: Sentinel Inventory Discipline

The set of exported sentinels is API. New ones are easy to add and impossible
to remove without breaking callers. Defend the inventory.

**Rules**
- A new sentinel must show a use case where a caller needs to branch on
  *this specific* failure differently from any other failure in the same
  category. "It might be useful someday" is not a use case.
- Document each sentinel with the exact condition that produces it and at
  least one call-site reference.
- Periodically grep `errors.Is(.*, ErrX)` across consumer code (your own
  tests count). If a sentinel has no consumer matches, it is dead and should
  be deleted.
- Resist the urge to split a sentinel into sub-variants. `ErrAuthFailed`
  is more useful than `ErrAuthFailedBadUsername` + `ErrAuthFailedBadPassword`
  + `ErrAuthFailedExpiredToken` because most callers want the same branch
  for all three.

---

## Verification Checklist

After any error-handling refactor, run through this list:

- [ ] Every exported sentinel is reachable from at least one call site
      (verified by a test).
- [ ] Every `fmt.Errorf` call's `%w` argument is a value the caller would
      reasonably want to match. Diagnostic-only context uses `%v` or `%s`.
- [ ] No subsystem label appears twice in any single error message
      (`grep -E '(\w+: ){2}'`).
- [ ] Wraps add non-redundant context — no `fmt.Errorf("foo failed: %w", err)`
      where the error already says "foo".
- [ ] `IsTemporary` / `IsPermanent` (if defined) classify
      `context.DeadlineExceeded` and `context.Canceled` correctly.
- [ ] No `recover()` is bare — all recoveries log structured context.
- [ ] `Hooks`-style optional callbacks nil-check the *field*, not the
      enclosing closure.
- [ ] No `*Error` struct literal exists outside a constructor function
      (after the constructor refactor lands).
- [ ] Internal-package and public-package sentinel tables are unified
      (one declared, the other re-exported via `var`).
- [ ] Stdlib parser errors (`net/mail`, `json`, `x509`, `pem`) are wrapped
      with `%v`, not `%w`, unless the caller-facing chain demands them.
- [ ] `Result{Records, Issues}` types do not also return non-nil error on
      partial success.
- [ ] On Go 1.26+, `errors.As(err, &target)` calls have been migrated to
      `errors.AsType[T](err)`. Run `go fix ./...` to automate.
- [ ] Loops accumulating errors use `errors.Join`, not manual list-of-errors
      types.

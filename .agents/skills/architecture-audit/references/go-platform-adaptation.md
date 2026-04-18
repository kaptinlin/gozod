# Go Platform Adaptation

This reference applies only to Go projects.

Guidance for auditing cross-platform Go code when platform differences should be handled by Go's native OS selection model rather than a heavy runtime adapter/factory layer.

## Core Principle

When behavior truly differs by operating system, prefer Go's platform adaptation mechanisms:

- file naming by GOOS / GOARCH (`foo_darwin.go`, `foo_linux.go`, `foo_windows.go`)
- build tags when naming alone is insufficient
- per-platform packages only when they represent genuinely different implementations

Do not default to a runtime adapter/provider/factory pattern when the choice is already known at build time.

## What to Audit

### 1. Compile-time vs runtime selection

Check whether OS-specific behavior is selected at the right layer.

**Prefer:**
- compile-time file selection for OS-specific mechanics
- one stable cross-platform API surface above those files
- shared code only for semantics that are actually common

**Flag as architecture issues:**
- runtime registries/providers whose main job is just `switch runtime.GOOS`
- interface layers created only to dispatch among known OS implementations
- "platform abstraction" that duplicates what Go file selection already provides

### 2. Package and file layout

Good shapes:

```text
launch.go
launch_darwin.go
launch_linux.go
launch_windows.go
```

or

```text
platform/
  open_darwin.go
  open_linux.go
  open_windows.go
```

Acceptable when platform code is meaningfully large:

```text
platform/darwin/
platform/linux/
platform/windows/
```

Flag when:
- platform directories exist but are only placeholders
- real behavior lives in one shared file while per-platform files are empty
- runtime indirection hides that packages actually compile into different binaries

### 3. Semantic unification vs mechanism unification

Cross-platform design should unify **intent and outcomes**, not force identical mechanisms.

Audit for:
- stable domain vocabulary above platform code
- platform-specific command/process details staying below the boundary
- result semantics shared across OSes even if launch mechanics differ

Flag when:
- public APIs expose `open`, `xdg-open`, `cmd.exe`, bundle IDs, desktop-file details, etc.
- code tries to pretend platforms are identical by flattening meaningful differences
- common code becomes an abstraction sink for unrelated per-OS mechanics

### 4. Packaging reality

Different target OSes produce different build artifacts. Architecture should acknowledge that reality.

Audit for:
- code organization that assumes separate platform builds are normal
- platform-specific imports isolated to matching files
- no accidental cross-OS dependency bleed

Flag when:
- one package imports OS-specific APIs for multiple platforms in the same file
- platform conditionals dominate runtime flow instead of file-level separation
- the architecture treats packaging differences as an implementation smell rather than a normal Go constraint

## Review Questions

Use these during architecture audit:

1. Is this runtime abstraction doing work that Go's build system should do instead?
2. Are per-OS mechanics isolated in per-OS files/packages?
3. Does the public API stay stable while each OS keeps its own implementation truth?
4. Are platform placeholders hiding missing implementation?
5. Does the design acknowledge that Darwin/Linux/Windows builds are different artifacts?

## Good Pattern

```go
// launch.go
package launch

func Launch(target Target) (Result, error) {
    return launchCurrentPlatform(target)
}
```

```go
// launch_darwin.go
package launch

func launchCurrentPlatform(target Target) (Result, error) {
    // Darwin implementation
}
```

```go
// launch_linux.go
package launch

func launchCurrentPlatform(target Target) (Result, error) {
    // Linux implementation
}
```

## Anti-Pattern

```go
type Adapter interface {
    Launch(Target) (Result, error)
}

type Provider struct {
    adapters map[string]Adapter
}

func (p Provider) Current() Adapter {
    return p.adapters[runtime.GOOS]
}
```

If the only real selection input is the current OS, this is often unnecessary runtime indirection.

## Architecture Audit Notes

When reporting findings:
- treat placeholder per-platform packages as a real implementation gap
- distinguish valid shared semantics from unnecessary runtime dispatch
- recommend moving OS-specific mechanics into build-selected files when that simplifies the design
- preserve progressive disclosure at the API layer while reducing internal platform ceremony

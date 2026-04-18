# Utilities

## `github.com/google/uuid` — UUID Generation

- RFC 9562 and DCE 1.1 compliant
- Versions 1, 4, 6, 7
- 16-byte array representation (not slice)
- SQL scanning, JSON marshaling, null UUID handling
- Zero dependencies, 6K+ stars, 448K+ dependents

**When to use:** Unique identifiers for database PKs, distributed systems, request tracing. Prefer **v7** for time-ordered sortable IDs, **v4** for random IDs.

```go
id := uuid.New()        // v4 random
id := uuid.Must(uuid.NewV7()) // v7 time-ordered
```

## `github.com/agentable/go-time` — Date/Time Semantics

- Human time parsing, immutable value objects, locale-aware formatting, and timezone-aware semantics
- Explicit resolved/ambiguous/invalid parse status instead of silent guessing
- Covers natural language time, intervals, relative time, and value-object style APIs

**When to use:** Time-heavy business logic, human time input, locale-aware formatting, timezone-sensitive scheduling, and interval semantics.

**When NOT to use:** Simple timestamp parse/format or basic duration arithmetic that stdlib `time` already handles cleanly.

## `github.com/agentable/go-humanize` — Human-Readable Formatting

- Bytes, numbers, durations, relative time, ordinals, pluralization, and percentages
- Pure functions, zero dependencies

**When to use:** Presenting machine values to humans in CLIs, logs, UIs, and diagnostics.

## `github.com/agentable/go-color` — Color Processing

- Parse, convert, analyze, and output colors across 11 color spaces
- Includes WCAG contrast checks, palette generation, and design-token-friendly output

**When to use:** Design systems, theming, color accessibility checks, palette tooling, and color conversion workflows.

## `github.com/agentable/go-version` — Build/Runtime Version Metadata

- Single source of truth for version/build/runtime metadata
- Stable simple/text/JSON/header output without framework lock-in

**When to use:** CLIs and services that need reliable version/build reporting.

## `github.com/agentable/go-filesystem` — Filesystem Abstraction

- Capability-based filesystem interfaces with local/memory/cloud drivers and middleware
- Good fit for mountable, testable, multi-backend file operations

**When to use:** Abstracting over local/object/virtual filesystems in application code.

## `github.com/agentable/go-test` — Scenario / Protocol Test Orchestration

- YAML-driven HTTP, gRPC, WebSocket, and CLI test orchestration with native `testing.T` integration
- Useful when `testify` assertions alone are not enough

**When to use:** Multi-step API/CLI/protocol testing scenarios and reusable scenario suites.

## `github.com/bojanz/currency` — Currency Handling

- All currency codes, numeric codes, fraction digits
- Locale-aware formatting (~370 locales, CLDR v48)
- Country-to-currency mapping
- Fowler's Money pattern with value semantics
- Arbitrary-precision decimals via `cockroachdb/apd` (no float errors)
- Direct PostgreSQL composite type support

**When to use:** Financial applications, e-commerce, invoicing — any code handling money.

**Critical rule:** Never use `float64` for currency. Use this library.

```go
amount, _ := currency.NewAmount("99.99", "USD")
locale := currency.NewLocale("en-US")
formatter := currency.NewFormatter(locale)
fmt.Println(formatter.Format(amount)) // "$99.99"
```

## `github.com/kaptinlin/gozod` — Data Validation

- Zod v4-inspired validation for Go
- Strongly-typed, zero-dependency
- Intelligent type inference

**When to use:** API input validation, form validation, data sanitization — when jsonschema is overkill.

## `github.com/kaptinlin/deepclone` — Deep Cloning

- High-performance deep copy of any Go value
- Safe, handles cycles

**When to use:** Copying complex structs before mutation, snapshot patterns, test fixtures.

## `github.com/stretchr/testify` — Testing Assertions

- `assert`/`require` for expressive test assertions
- 25K+ stars, most widely used Go test library
- Test-only dependency (doesn't affect user's dep tree)

**When to use:** All test files. Always use as test dependency.

```go
assert.Equal(t, expected, actual)
require.NoError(t, err)  // stops test on failure
```

## `github.com/cerbos/cerbos-sdk-go` — Authorization

- Go SDK for Cerbos policy decision platform
- Simple `IsAllowed()` API
- Flexible deployment: microservice or sidecar
- TLS support, test utilities

**When to use:** Externalized access control, policy-based authorization, RBAC/ABAC.

## `github.com/fsnotify/fsnotify` — File System Notifications

- Cross-platform file system event notifications (create, write, remove, rename, chmod)
- Backends: inotify (Linux), kqueue (macOS/BSD), ReadDirectoryChangesW (Windows), FEN (illumos)
- Non-recursive — subdirectories require separate `Add()` calls
- 10K+ stars, widely used across Go ecosystem

**When to use:** Config file hot-reload, dev tools (live reload, asset pipeline), log rotation detection, file sync — any time you need to react to file system changes.

**When NOT to use:** Recursive directory trees with thousands of entries (consider polling instead). One-shot file existence checks (use `os.Stat`).

```go
watcher, _ := fsnotify.NewWatcher()
defer watcher.Close()

go func() {
    for {
        select {
        case event := <-watcher.Events:
            if event.Has(fsnotify.Write) {
                log.Println("modified:", event.Name)
            }
        case err := <-watcher.Errors:
            log.Println("error:", err)
        }
    }
}()

watcher.Add("/path/to/watch")
```

## `github.com/agentable/vfs` — Virtual Filesystem

- Unix-like commands, mountable backends
- Permission-aware access
- Shell semantics

**When to use:** AI agent file operations, sandboxed environments, multi-backend file abstraction.

## `github.com/agentable/unifmsg` — Unified Messaging

- One interface for all platforms: Telegram, Discord, Slack, WhatsApp, Line, Twitter, WeChat
- Multimodal messages, command routing

**When to use:** Multi-platform chatbot, notification delivery across platforms, unified messaging gateway.

## `github.com/agentable/go-bashrepair` — Bash Command Repair

- Auto-fixes common bash syntax errors
- Designed for AI-generated shell commands

**When to use:** Processing LLM-generated bash commands, CLI input sanitization.

## `github.com/agentable/gendog` — Code Generation

- Schema-driven templates
- Intelligent variable collection
- kubectl/helm-like CLI patterns

**When to use:** Scaffolding projects, generating boilerplate, schema-to-code workflows.

## Do NOT Use

| Library | Reason |
|---------|--------|
| `jinzhu/now` | Narrower API, prefer `agentable/go-time` for richer time semantics |
| `araddon/dateparse` as primary time toolkit | Useful parser only; not a full time semantics toolkit |
| `dustin/go-humanize` | Prefer our ecosystem humanization helpers | `github.com/agentable/go-humanize` |
| `dromara/carbon/v2` | Prefer our time semantics/value-object toolkit | `github.com/agentable/go-time` |

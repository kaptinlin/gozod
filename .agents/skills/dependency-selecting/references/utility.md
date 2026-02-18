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

## `github.com/dromara/carbon/v2` — Date/Time Toolkit

- Rich date/time API (parse, format, add/sub, boundaries like start/end of day)
- Timezone helpers and locale-aware formatting
- Human-friendly helpers for common business time logic

**When to use:** Time-heavy business code (report windows, billing cycles, schedule windows, timezone conversions) where stdlib `time` code becomes repetitive.

**When NOT to use:** Simple timestamp parse/format or basic duration arithmetic that stdlib `time` already handles cleanly.

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

## `github.com/agentable/vfs` — Virtual Filesystem

- Unix-like commands, mountable backends
- Permission-aware access
- Shell semantics

**When to use:** AI agent file operations, sandboxed environments, multi-backend file abstraction.

## `github.com/agentable/unifmsg` — Unified Messaging

- One interface for all platforms: Telegram, Discord, Slack, WhatsApp, Line, Twitter, WeChat
- Multimodal messages, command routing

**When to use:** Multi-platform chatbot, notification delivery across platforms, unified messaging gateway.

## `github.com/agentable/bashrepair` — Bash Command Repair

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
| `jinzhu/now` | Narrower API, prefer `carbon/v2` for richer time operations |
| `araddon/dateparse` as primary time toolkit | Useful parser only; not a full date-time operation toolkit |

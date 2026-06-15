# 001 Core Validation Language

## Contract

GoZod is a Go-native Zod v4-inspired validation language. The public language is
the root `gozod` package, fluent schema methods, and the `Parse(any)` /
`StrictParse(T)` pair. Subpackages exist for focused work, but user code should
start from root constructors until it needs a lower-level contract.

Go type identity is strict by default. Coercion belongs behind explicit
constructors such as `coerce` or `Coerced*`. Lossy conversion and compatibility
surface should be explicit, not accidental.

Canonical public constructor spelling follows the current root facade:
`UUID`, `UUIDv4`, `UUIDv6`, `UUIDv7`, `CUID`, `CUID2`, `ULID`, `KSUID`,
`NanoID`, and related uppercase initialism forms. Documentation should use real
exports instead of preserving stale spellings.

## Modifier Semantics

Presence and default behavior is ordered. Fluent calls such as
`.optional().default(x)` and `.default(x).optional()` are different parse
intentions and must not be collapsed into unordered flags.

`core.ZodTypeInternals.Modifiers` is the semantic spine for ordered presence and
default behavior:

- optional
- nilable
- nonoptional
- default
- prefault

Parsing walks that ordered spine from the outside in. Default short-circuits nil
input only when it is the outer modifier that claims the input. Prefault supplies
fallback input that still flows through validation. Nonoptional rejects missing
or nil values according to its position in the chain.

Legacy summary fields such as `Optional`, `Nilable`, `DefaultValue`, and
`PrefaultValue` may remain only as compatibility or fast-path state while real
behavior is protected by the ordered modifier spine. No second priority system
should be reintroduced beside `Modifiers`; the slice order is the call order.

All fluent modifier methods must remain copy-on-write. The original schema must
not change when a modifier returns a configured clone.

## Checks

First-party checks are semantic runtime units. Their `core.ZodCheckDef.Params`
are the source of truth for:

- runtime issue creation
- JSON Schema projection
- tests that tie fluent calls to produced constraints

Schema bags may cache projection hints, but they are not the semantic owner of a
check. JSON Schema conversion must be able to project first-party checks from
check definitions even when bag state is absent.

## Transforms And Pipes

`core.ZodTransform` and `core.ZodPipe` stay structural. Modifier ordering should
compose with transforms and pipes instead of being folded into special parse
paths. Transform output must continue to honor Go type identity for strict
parsing.

## Internal Ownership

`core.ZodTypeInternals` is allowed to be broad only where no better tested owner
exists yet. Fields may move out of it only after behavior tests prove the new
semantic owner:

- modifier order belongs to `Modifiers`
- first-party JSON Schema projection belongs to check definitions
- conversion caches belong to converters, not to public contract language

Package topology should not be created for appearance. New internal packages or
public interfaces need a real caller and a behavior contract.

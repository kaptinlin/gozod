# 001 Core Validation Language

## Contract

GoZod is a Go-native Zod v4-inspired validation language. The public language is
the root `gozod` package, fluent schema methods, and the `Parse(any)` /
`StrictParse(T)` pair. Subpackages exist for focused work, but user code should
start from root constructors until it needs a lower-level contract.

The root facade also exposes the shared schema contract as `gozod.ZodSchema`.
User code should be able to name the schema abstraction without importing
`core` unless it needs lower-level internals.

Exported facade operations are declared functions. Constructors, configuration,
error presentation, and JSON Schema conversion cannot be rebound as mutable
package state. Sentinel errors, registries, constants, and precision values
remain data declarations because callers inspect or configure them as data.

Go type identity is strict by default. Coercion belongs behind explicit
constructors such as `coerce` or `Coerced*`. Lossy conversion and compatibility
surface should be explicit, not accidental.

## Structural Output

Successful parse output must match the schema-described output, not merely the
original input shape. Object fields and catchall entries write parsed child
values into returned maps. This includes child defaults, prefaults, overwrites,
transforms, and explicit coercion.

Missing object fields with child defaults or prefaults run the child schema
instead of being silently omitted. Optional and exact optional modifiers still
decide absent key acceptance; explicit nil remains distinct from absence for
exact optional fields.

> **Why**: a schema is a data constructor as much as a validator. Once child
> schemas can transform, overwrite, coerce, or synthesize defaults, returning
> raw input values makes object parsing lie about the output type.
>
> **Rejected**: object parsing that only validates child schemas and then
> returns original field values.

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

No summary flags or fallback fields may form a second priority system beside
`Modifiers`; the slice order is the call order.

All fluent modifier methods must remain copy-on-write. The original schema must
not change when a modifier returns a configured clone.

## Checks

First-party checks are semantic runtime units. Attachment materializes their
JSON Schema-relevant constraints into the schema bag through the check's
`OnAttach` behavior. The exporter reads that materialized bag; it does not run a
second check-definition projector. `core.ZodCheckDef.Params` remains semantic
introspection data for checks and tests, not a parallel export implementation.

The check runner owns runtime issue production. It stamps issue instances,
honors per-check `Abort` and `When` behavior, applies custom error functions,
and returns accumulated issues through one path. Structural checks should not
hide in documentation-only assertions or converter-only metadata.

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
- first-party JSON Schema projection is materialized by check attachment and read from the schema bag
- conversion caches belong to converters, not to public contract language

Package topology should not be created for appearance. New internal packages or
public interfaces need a real caller and a behavior contract.

## Acceptance Criteria

- Object parsing tests prove parsed child output for fields and catchall values,
  including defaults, prefaults, overwrites, transforms, and explicit coercion.
- Modifier-order tests prove optional, nilable, nonoptional, default, and
  prefault chains follow call order and preserve copy-on-write behavior.
- Check runner tests prove first-party checks produce the expected issue
  instance, continue, abort, and custom-error behavior.
- Composite constructor tests prove root constructor spelling and `gozod.ZodSchema`
  are real exported API, not only documentation claims.
- `task contractcheck` passes after schema interface or family changes.

# 002 JSON Schema Contract

## Contract

GoZod and JSON Schema are different validation languages. Conversion should be
honest, deterministic, and fail-closed. GoZod does not try to import or export
the entire JSON Schema universe.

`ToJSONSchema` emits Draft 2020-12-compatible schemas for supported GoZod
constructs. Generated schemas should be deterministic: repeated conversion of
the same schema should change only when the source schema or converter logic
changes.

Single-schema and batch export are distinct typed operations:
`ToJSONSchema(core.ZodSchema)` exports one schema, while
`ToJSONSchemaRegistry(*core.Registry[core.GlobalMeta])` exports a registry
bundle. Invalid input categories are compile-time errors rather than runtime
dispatch failures.

Batch registry export requires every top-level entry to have a non-empty,
unique registry `ID`. Missing and duplicate IDs fail with
`ErrInvalidRegistrySchemaID` before schema conversion or override callbacks.
Nil registries fail with the same sentinel instead of panicking.
Conversion uses one frozen schema-and-metadata snapshot and visits entries in ID
order, so concurrent registry replacement cannot mix metadata generations and
the first conversion error is stable.

Draft 2020-12 is the fixed export dialect. The API exposes no target option
until another dialect is implemented faithfully.

`FromJSONSchema` imports a defined subset into GoZod schemas and fails on
unsupported semantics. `FromJSONSchemaLossy` is the explicit partial-import
entry point; it returns the schema, a stable location-aware `[]ImportLossError`, and
any fatal conversion error. Each loss identifies the omitted keyword, its RFC
6901 JSON Pointer, and an inspectable cause.

## Supported Import Language

The import subset intentionally covers the overlap GoZod owns:

- boolean schemas
- `$ref` when already resolved by `github.com/kaptinlin/jsonschema`
- primitive `type`: `string`, `number`, `integer`, `boolean`, `null`, `array`,
  `object`
- multi-type unions built from supported primitive types
- string constraints: `format`, `minLength`, `maxLength`, `pattern`
- numeric constraints: `minimum`, `maximum`, `exclusiveMinimum`,
  `exclusiveMaximum`, `multipleOf`
- array constraints: `items`, `prefixItems`, `minItems`, `maxItems`
- object shape: `properties`, `required`, `additionalProperties`
- composition: `allOf`, `anyOf`, `oneOf`
- constants and enums: `const`, `enum`
- metadata: `$id`, `title`, `description`, `examples`

Unknown string formats fall back to ordinary string validation instead of
pretending GoZod knows that format.

JSON Schema `integer` imports target Go's platform-sized `int` domain.
Rational bounds are rounded to the equivalent inclusive integer bound. A
constraint operand or exclusive-bound adjustment outside that domain is a
fatal `ErrInvalidJSONSchema` with the keyword's JSON Pointer; lossy import does
not discard or approximate it.

## Fail-Closed Keywords

These JSON Schema keywords are not imported by default because accepting them
silently would drop validation semantics:

| Keyword | Import Behavior |
|---------|-----------------|
| `if` / `then` / `else` | Strict error; lossy import records `if/then/else`. |
| `patternProperties` | Strict error unless it is the supported one-pattern record shape; otherwise lossy import records `patternProperties`. |
| `$dynamicRef` | Strict error; lossy import records `$dynamicRef`. |
| `unevaluatedProperties` | Strict error; lossy import records `unevaluatedProperties`. |
| `unevaluatedItems` | Strict error; lossy import records `unevaluatedItems`. |
| `not` | Strict error; lossy import records `not`. |
| `uniqueItems: true` | Strict error; lossy import records `uniqueItems`. |
| `dependentSchemas` | Strict error; lossy import records `dependentSchemas`. |
| `propertyNames` | Strict error unless it is the supported record-key shape; otherwise lossy import records `propertyNames`. |
| `contains` / `minContains` / `maxContains` | Strict error; lossy import records `contains`. |

The converter should keep the fail-closed keyword set as executable data, not as
scattered prose. Tests must prove each listed keyword fails by default and is
recorded in lossy mode.

## Option Modes

JSON Schema conversion modes are typed protocol values, not stringly typed
configuration. Public callers should use exported constants for:

- unrepresentable schemas
- cycle handling
- reused schema handling
- input/output projection

Invalid option values fail with `ErrInvalidJSONSchemaOption`.

## Export Language

Export uses JSON Schema as the projection of GoZod semantics, not as an
alternate source of truth. First-party check attachment materializes exportable
constraints into the schema bag. The exporter reads that bag and does not own a
second projector from `ZodCheckDef.Params`.

Regex-backed format checks emit their authoritative `pattern` without also
emitting `format`. Combining both keywords would intersect independently
defined accepted sets and can narrow GoZod semantics when format assertion is
enabled. Check definitions may retain the format name as introspection data;
only attachment metadata controls export. Patterns are emitted from Go's RE2
dialect and parity is verified with the owned Go JSON Schema validator. This
does not claim character-class equivalence with ECMAScript engines; for
example, RE2 `\s` and ECMAScript `\s` are different sets.

Unrepresentable GoZod schemas fail by default.
`Options{Unrepresentable: JSONSchemaUnrepresentableAny}` may emit `{}` only when
the caller explicitly chooses that fallback.

Round-trip tests should cover only the supported overlap. They should not imply
that arbitrary JSON Schema documents can be imported and re-emitted unchanged.

Integer constraints are emitted exactly. Native signed and unsigned integer
values reach the JSON Schema rational representation without a `float64`
intermediary, and each integer schema node carries its Go type range regardless
of whether it appears at the root, in a property or item, in a union, or in
`$defs`. An explicit lower or upper bound replaces only that side of the type
range.

Lazy schemas use the same fail-closed policy as other unrepresentable schemas.
A lazy getter returning nil or a non-schema value returns
`ErrUnrepresentableType` without panicking; only explicit
`JSONSchemaUnrepresentableAny` may project it as `{}`.

> **Why**: numeric meaning belongs to the schema node, not traversal depth, and
> an invalid lazy target is an unrepresentable schema rather than an empty
> contract.
>
> **Rejected**: approximate integer constraints, root-only type bounds, or an
> implicit `{}` fallback for invalid lazy values.

## Metadata Ownership

Standard fluent metadata belongs to the schema value, not ambient global state.
`ToJSONSchema` reads schema-owned metadata by default. When an explicit metadata
registry contains the schema, that whole record is authoritative for the
conversion; a missing entry falls back to schema-owned metadata without
field-by-field merge.

`FromJSONSchemaOptions.Metadata` chooses the registry that imported metadata is
attached to. With no explicit registry, metadata belongs to the returned schema.
An explicit registry is the sole destination; import never writes
`core.GlobalRegistry` implicitly.

Import snapshots `$id`, title, description, and nested examples before attaching
them to either destination. Export also detaches nested examples from registry
records before returning a JSON Schema document. Mutating the input document,
destination metadata, or exported document must not cross those boundaries.

Generic registries retain shallow caller-owned value semantics. JSON Schema
import/export owns these deep snapshots because it is the external data
boundary; `Registry[M]` does not impose a universal clone policy on arbitrary
metadata types.

## Acceptance Criteria

- Option validation tests prove invalid mode values fail with
  `ErrInvalidJSONSchemaOption`; compile-negative coverage proves `Options` has
  no `Target` field.
- Export tests prove first-party constraints materialized by check attachment
  are emitted without a second definition projector.
- Numeric export tests prove all native integer kinds serialize exact explicit
  constraints and position-independent default ranges.
- Lazy export tests prove nil and non-schema targets fail without panic unless
  the caller explicitly selects the any fallback, while recursive cycle modes
  retain their defined behavior.
- Import tests prove every fail-closed keyword errors through `FromJSONSchema`
  and is returned with location and cause through `FromJSONSchemaLossy`.
- Metadata tests prove schema-owned and explicit-registry imports snapshot nested
  examples, round-trip supported annotations, and never leak into the global registry.
- Export tests prove default output ignores process-global registry history and
  explicit registry entries use whole-record precedence with detached examples.
- Batch tests prove missing or duplicate IDs fail before conversion, registry
  mutation cannot tear a conversion snapshot, and callbacks/errors follow ID order.
- External JSON Schema validator tests cover emitted Draft 2020-12 schemas for
  supported string, number, array, object, enum, union, and metadata cases.
  Format-asserting tests prove regex-backed Email and URL exports accept the
  runtime-valid edge cases that their standard formats reject.

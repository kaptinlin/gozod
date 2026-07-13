# 002 JSON Schema Contract

## Contract

GoZod and JSON Schema are different validation languages. Conversion should be
honest, deterministic, and fail-closed. GoZod does not try to import or export
the entire JSON Schema universe.

`ToJSONSchema` emits Draft 2020-12-compatible schemas for supported GoZod
constructs. Generated schemas should be deterministic: repeated conversion of
the same schema should change only when the source schema or converter logic
changes.

Batch registry export requires every top-level entry to have a non-empty,
unique registry `ID`. Missing and duplicate IDs fail with
`ErrInvalidRegistrySchemaID` before schema conversion or override callbacks.
Conversion uses one frozen schema-and-metadata snapshot and visits entries in ID
order, so concurrent registry replacement cannot mix metadata generations and
the first conversion error is stable.

Draft 2020-12 is the supported target. A requested target outside the supported
set fails with `ErrUnsupportedJSONSchemaTarget` instead of silently emitting a
schema with ambiguous semantics.

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
- target dialect
- input/output projection

Invalid option values fail with `ErrInvalidJSONSchemaOption`. Unsupported target
dialects fail with `ErrUnsupportedJSONSchemaTarget`.

## Export Language

Export uses JSON Schema as the projection of GoZod semantics, not as an
alternate source of truth. First-party check attachment materializes exportable
constraints into the schema bag. The exporter reads that bag and does not own a
second projector from `ZodCheckDef.Params`.

Unrepresentable GoZod schemas fail by default.
`Options{Unrepresentable: JSONSchemaUnrepresentableAny}` may emit `{}` only when
the caller explicitly chooses that fallback.

Round-trip tests should cover only the supported overlap. They should not imply
that arbitrary JSON Schema documents can be imported and re-emitted unchanged.

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
  `ErrInvalidJSONSchemaOption` and unsupported targets fail with
  `ErrUnsupportedJSONSchemaTarget`.
- Export tests prove first-party constraints materialized by check attachment
  are emitted without a second definition projector.
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

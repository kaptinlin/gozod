# 002 JSON Schema Contract

## Contract

GoZod and JSON Schema are different validation languages. Conversion should be
honest, deterministic, and fail-closed. GoZod does not try to import or export
the entire JSON Schema universe.

`ToJSONSchema` emits Draft 2020-12-compatible schemas for supported GoZod
constructs. Generated schemas should be deterministic: repeated conversion of
the same schema should change only when the source schema or converter logic
changes.

Draft 2020-12 is the supported target. A requested target outside the supported
set fails with `ErrUnsupportedJSONSchemaTarget` instead of silently emitting a
schema with ambiguous semantics.

`FromJSONSchema` imports a defined subset into GoZod schemas. Unsupported
keywords fail by default. `AllowLossy` permits partial import only when the
caller explicitly accepts semantic loss, and `LossyKeywords` records every
unsupported keyword ignored by that import.

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
| `if` / `then` / `else` | Error unless `AllowLossy` records `if/then/else`. |
| `patternProperties` | Error unless `AllowLossy` records `patternProperties`. |
| `$dynamicRef` | Error unless `AllowLossy` records `$dynamicRef`. |
| `unevaluatedProperties` | Error unless `AllowLossy` records `unevaluatedProperties`. |
| `unevaluatedItems` | Error unless `AllowLossy` records `unevaluatedItems`. |
| `not` | Error unless `AllowLossy` records `not`. |
| `uniqueItems: true` | Error unless `AllowLossy` records `uniqueItems`. |
| `dependentSchemas` | Error unless `AllowLossy` records `dependentSchemas`. |
| `propertyNames` | Error unless `AllowLossy` records `propertyNames`. |
| `contains` / `minContains` / `maxContains` | Error unless `AllowLossy` records `contains`. |

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
alternate source of truth. First-party checks project from their check
definitions. Bag state can cache hints, but conversion must not depend on bag
mutation for first-party checks.

Unrepresentable GoZod schemas fail by default.
`Options{Unrepresentable: JSONSchemaUnrepresentableAny}` may emit `{}` only when
the caller explicitly chooses that fallback.

Round-trip tests should cover only the supported overlap. They should not imply
that arbitrary JSON Schema documents can be imported and re-emitted unchanged.

## Metadata Ownership

Metadata belongs to a registry, not to ambient global state hidden inside
conversion. `ToJSONSchema` reads metadata from its configured registry when
available. `FromJSONSchemaOptions.Metadata` chooses the registry that imported
metadata is attached to; nil means the package global registry.

Tests that import JSON Schema metadata must prove custom registries stay
isolated from the global registry.

## Acceptance Criteria

- Option validation tests prove invalid mode values fail with
  `ErrInvalidJSONSchemaOption` and unsupported targets fail with
  `ErrUnsupportedJSONSchemaTarget`.
- Export tests prove first-party constraints project from check definitions even
  when bag projection hints are absent.
- Import tests prove every fail-closed keyword errors by default and is recorded
  when `AllowLossy` is enabled.
- Metadata tests prove custom import registries receive imported metadata without
  leaking into the global registry.
- External JSON Schema validator tests cover emitted Draft 2020-12 schemas for
  supported string, number, array, object, enum, union, and metadata cases.

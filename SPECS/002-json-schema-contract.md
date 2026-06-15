# 002 JSON Schema Contract

## Contract

GoZod and JSON Schema are different validation languages. Conversion should be
honest, deterministic, and fail-closed. GoZod does not try to import or export
the entire JSON Schema universe.

`ToJSONSchema` emits Draft 2020-12-compatible schemas for supported GoZod
constructs. Generated schemas should be deterministic: repeated conversion of
the same schema should change only when the source schema or converter logic
changes.

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

## Export Language

Export uses JSON Schema as the projection of GoZod semantics, not as an
alternate source of truth. First-party checks project from their check
definitions. Bag state can cache hints, but conversion must not depend on bag
mutation for first-party checks.

Unrepresentable GoZod schemas fail by default. `Options{Unrepresentable: "any"}`
may emit `{}` only when the caller explicitly chooses that fallback.

Round-trip tests should cover only the supported overlap. They should not imply
that arbitrary JSON Schema documents can be imported and re-emitted unchanged.

# JSON Ecosystem

## `github.com/go-json-experiment/json` — High-Performance JSON

- 2.5-10x faster Unmarshal than `encoding/json` v1, significantly less memory allocation
- Supports `omitzero` (correct `time.Time` zero value handling)
- Native `encoding.TextUnmarshaler` support (`time.Duration`, `net.IP` auto-convert)
- Handles `map[string]any` → struct conversion (replaces mapstructure)
- Future `encoding/json/v2` standard library (Go 1.25/1.26 via `GOEXPERIMENT=jsonv2`)

**When to use:** Performance-sensitive code, `map[string]any` → struct pipelines, `omitzero` needed.

**Risk:** API not yet stable. Isolate in separate module (e.g., `codec/json2codec`). Use `encoding/json` as default codec.

## `github.com/kaptinlin/jsonschema` — Schema Validation

- High-performance JSON Schema validator
- Direct struct validation (no marshal→validate roundtrip)
- Smart unmarshaling with defaults
- Separated validation workflow

**When to use:** API request/response validation, configuration validation, form input validation.

## `github.com/kaptinlin/jsonpatch` — RFC 6902 JSON Patch

- Full RFC 6902 compliance, type-safe, generic support
- Ported from json-joy (battle-tested reference)

**When to use:** Incremental document updates, PATCH API endpoints.

## `github.com/kaptinlin/jsonmerge` — RFC 7386 JSON Merge Patch

- Type-safe, RFC 7386 compliant, generic support

**When to use:** Simple document merging where `null` means "delete field".

## `github.com/kaptinlin/jsonpointer` — RFC 6901 JSON Pointer

- Fast JSON Pointer (RFC 6901) implementation

**When to use:** Navigating nested JSON documents by path.

## `github.com/kaptinlin/jsonrepair` — Fix Malformed JSON

- Port of popular JavaScript `jsonrepair`
- Fixes missing quotes, escape chars, trailing commas
- Handles LLM-generated JSON

**When to use:** Processing LLM output, user-submitted JSON, untrusted JSON sources.

## `github.com/agentable/jsoncrdt` — Conflict-Free Replicated Data Types

- 100% compatible with json-joy TypeScript reference
- Distributed collaborative editing

**When to use:** Real-time collaboration, offline-first sync, multi-user document editing.

## `github.com/agentable/jsondiff` — JSON Diffing

- High-performance, generic diff computation
- Flat serializable change lists

**When to use:** Change tracking, audit logs, document version comparison.

## `github.com/kaptinlin/orderedobject` — Ordered JSON Objects

- Preserves insertion order
- Designed for `go-json-experiment/json` v2 compatibility

**When to use:** When JSON key order matters (config files, deterministic output, schema defs).

## Do NOT Use

| Library | Reason |
|---------|--------|
| `go-viper/mapstructure/v2` | json/v2 handles map→struct natively |
| `goccy/go-json` | We use `go-json-experiment/json` |
| `tidwall/sjson` | We don't need JSON path set operations |
| `tidwall/gjson` | Use jsonpointer for path navigation |

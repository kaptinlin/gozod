# GoZod Specs

These specs hold durable product and implementation contracts for GoZod. They
record semantic decisions that should stay true across releases; they do not
record execution plans, agent runs, or one-off migration notes.

## Index

| Spec | Owns |
|------|------|
| [001 Core Validation Language](001-core-validation-language.md) | Root API language, strict parsing, modifier semantics, check ownership, and internal ownership rules. |
| [002 JSON Schema Contract](002-json-schema-contract.md) | Supported JSON Schema import/export language, fail-closed import, lossy recording, and round-trip boundaries. |
| [003 Struct Tags And Generation](003-struct-tags-and-generation.md) | Struct tag semantic compilation, runtime/generated parity, and deterministic generated artifacts. |

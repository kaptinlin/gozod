# GoZod

GoZod is a TypeScript Zod v4-inspired validation library for Go 1.26.5. The root package is a user-facing facade over concrete schemas in `types/`, shared contracts in `core/`, runtime machinery in `internal/`, struct-tag helpers in `pkg/tagparser`, and JSON Schema conversion in `jsonschema/`.

- **Reference implementation:** TypeScript Zod v4 in [`.reference/zod/`](.reference/zod/) when that submodule is initialized.
- **User-facing usage:** see [README.md](README.md), [docs/](docs/), and [examples/](examples/).

## Commands

Run these from the repository root.

```bash
task test                           # Run go test -race ./...
task test:race                      # Run race tests for core and lightweight utility packages
task lint                           # Run golangci-lint and tidy-lint
task golangci-lint                  # Run golangci-lint v2 only
task tidy-lint                      # Verify go.mod and go.sum stay tidy
task contractcheck                  # Run go test -tags=contractcheck ./types
task docs:integrity                 # Run public documentation stale-claim checks
task bench:smoke                    # Compile and smoke-run benchmarks with tiny benchtime
task fmt                            # Run golangci-lint fmt ./...
task vet                            # Run go vet ./...
task vuln                           # Run govulncheck ./...
task verify:ci                      # Run lint, contractcheck, docs, race tests, benchmark smoke, and vuln
task verify                         # Run deps, fmt, vet, lint, contracts, docs, tests, benchmark smoke, and vuln
go build ./...                      # Verify all packages compile
```

## Architecture

```text
gozod/
├── constructors_*.go # Root facade constructor functions
├── core.go           # Root aliases for shared contracts
├── errors.go         # Root error presentation functions and sentinels
├── jsonschema.go     # Root JSON Schema conversion functions and protocol aliases
├── metadata.go       # Root registry and metadata aliases
├── structs.go        # Root FromStruct helpers
├── cmd/gozodgen/     # Code generator for struct-tag schemas
├── coerce/           # z.coerce-style constructors for automatic input conversion
├── core/             # Public contracts, config, issue codes, shared types
├── docs/             # User documentation
├── examples/         # Runnable examples
├── internal/         # Engine, checks, issues, and runtime helpers
├── jsonschema/       # To/From JSON Schema conversion
├── locales/          # Localized validation messages
├── pkg/              # Shared helpers for clone, coercion, maps, regex, reflection, tags, transforms
├── types/            # Concrete schema implementations and fluent APIs
├── .agents/rules/    # Project rules for schemas, checks, naming, tests, performance, and structure
└── .agents/skills/   # Local agent skills available in this repository
```

## Agent Operating Rules

- **Read before writing** - Inspect the relevant source, docs, and `.agents/rules/` guide before changing behavior.
- **Think in contracts** - Preserve public semantics before polishing implementation shape.
- **Keep edits surgical** - Change the smallest module boundary that actually owns the behavior.
- **Prefer simple truth** - A direct implementation beats a helper that only hides one call site.
- **Verify claims** - Do not document commands, files, exports, examples, or reference paths you have not checked.
- **Test behavior** - Tests should prove user-visible behavior, public contracts, regressions, and generated artifacts.
- **Fail loud** - Return errors for ordinary failures and expose unsupported conversions explicitly.
- **Respect context budgets** - Link to detailed specs and docs instead of duplicating large design material.

## Agent Workflow

- Read the relevant guide in [`.agents/rules/`](.agents/rules/) before changing schema internals, checks, struct tags, naming, tests, performance-sensitive paths, or package boundaries.
- Read the relevant durable contract in [SPECS/](SPECS/) before changing core validation semantics, JSON Schema conversion, or struct-tag/codegen behavior.
- Use [`.reference/zod/`](.reference/zod/) for parity work only after the submodule is initialized and the relevant TypeScript Zod v4 behavior is inspected.
- Keep `README.md` user-facing, `CLAUDE.md` development-facing, and `.agents/rules/` contract-facing.
- When public APIs or docs change, update examples and docs-integrity expectations together.
- Work with the current dirty tree; never revert unrelated user changes.

## SPECS Index

| Spec | Use When |
|------|----------|
| [SPECS/001-core-validation-language.md](SPECS/001-core-validation-language.md) | Changing root API language, strict parsing, object output, modifiers, checks, or internal ownership rules |
| [SPECS/002-json-schema-contract.md](SPECS/002-json-schema-contract.md) | Changing JSON Schema import/export modes, fail-closed import, metadata registries, or parity tests |
| [SPECS/003-struct-tags-and-generation.md](SPECS/003-struct-tags-and-generation.md) | Changing `gozod` tags, `pkg/tagparser`, `FromStruct`, `gozodgen`, generated fixtures, or codegen docs |

## Rules Index

| Rule | Use When |
|------|----------|
| [`.agents/rules/schema_implementation_guide.mdc`](.agents/rules/schema_implementation_guide.mdc) | Adding or changing schema implementations in `types/` |
| [`.agents/rules/checks_implementation_guide.mdc`](.agents/rules/checks_implementation_guide.mdc) | Adding or changing validation checks or JSON Schema check metadata |
| [`.agents/rules/schema_test_implementation_guide.mdc`](.agents/rules/schema_test_implementation_guide.mdc) | Writing schema tests, strict parsing tests, modifier tests, or benchmarks |
| [`.agents/rules/coding-standards.mdc`](.agents/rules/coding-standards.mdc) | Applying GoZod-specific coding and Zod v4 parity rules |
| [`.agents/rules/project-structure.mdc`](.agents/rules/project-structure.mdc) | Checking package layout and one-way dependency boundaries |
| [`.agents/rules/module_organization_guide.md`](.agents/rules/module_organization_guide.md) | Moving code across packages or creating new packages |
| [`.agents/rules/naming_guide.md`](.agents/rules/naming_guide.md) | Renaming exported APIs, files, packages, errors, or helpers |
| [`.agents/rules/performance-optimization.mdc`](.agents/rules/performance-optimization.mdc) | Optimizing parse paths, benchmarks, allocation behavior, or hot code |

## Design Philosophy

- **KISS** - Keep the public surface centered on root constructors, fluent modifiers, and the `Parse` / `StrictParse` pair.
- **DRY** - Share rule interpretation through `pkg/tagparser` and shared checks instead of duplicating validation semantics in reflection, generation, and JSON Schema paths.
- **Single Responsibility** - `core` defines contracts, `types` own schema behavior, `internal` owns execution, and `jsonschema` owns translation.
- **Open-Closed** - Extend through checks, refinements, metadata, schema families, and conversion options instead of special-casing existing schemas.
- **APIs as language** - Chains such as `gozod.String().Min(2).Email()` and `gozod.FromStruct[T]()` should read like the validation intent.
- **Beauty is structural** - The root package stays a thin facade while dependency direction remains one-way toward runtime internals.
- **Never:** accidental complexity, feature gravity, abstraction theater, configurability cope.

## API Design Principles

- **Progressive Disclosure:** Start with `gozod` root constructors and `FromStruct[T]()`; use `MustFromStruct[T]()` only when construction failure is an invariant breach.
- **Strict by default:** Preserve Go type identity unless callers explicitly opt into coercion through `coerce/` or `Coerced*` constructors.
- **Loss explicitness:** JSON Schema import fails on unsupported keywords by default; use `FromJSONSchemaLossy` only when dropping information is intentional and inspect every returned location-aware loss.

## Coding Rules

### Must Follow

- Use Go 1.26.5 and modern language features where they simplify code.
- Follow Google Go Best Practices: <https://google.github.io/go-style/best-practices>
- Follow Google Go Style Decisions: <https://google.github.io/go-style/decisions>
- Preserve the two parsing modes: `Parse(any)` for dynamic inputs and `StrictParse(T)` for compile-time constrained inputs.
- Preserve schema-described structural output: object fields and catchall entries return parsed child values, including defaults, prefaults, transforms, overwrites, and coercion when explicitly selected.
- Keep modifier methods copy-on-write. `Optional`, `Nilable`, `Default`, `Prefault`, `Describe`, `Meta`, and similar fluent modifiers must return a new configured schema.
- Route parsing through `internal/engine` helpers. Do not reimplement parsing pipelines in root facade files.
- Build validations through `internal/checks` and shared helpers; do not duplicate validation logic in schema methods.
- Materialize first-party JSON Schema constraints through check `OnAttach` behavior and read the schema bag in the exporter; keep `core.ZodCheckDef.Params` as introspection data, not a second projector.
- Keep struct-tag behavior shared between runtime reflection and `gozodgen` through `pkg/tagparser`.
- Keep root facade operations as declared functions; only sentinels, registries, constants, and other genuine data remain exported variables.
- Keep JSON Schema protocol modes typed and fail loud: invalid option values return errors, unsupported targets do not silently emit Draft 2020-12, and invalid batch IDs fail before conversion.
- Snapshot batch registry entries before conversion, visit them in ID order, and detach imported/exported nested examples at the JSON Schema boundary.
- Store imported metadata on the returned schema by default; an explicit registry is the sole destination and `GlobalRegistry` is never an implicit fallback.
- Keep dependency direction one-way. Schema files in `types/` do not import each other; shared behavior belongs in `internal/`, `pkg/`, `core/`, or `coerce/`.
- Update [types/constraints_verify.go](types/constraints_verify.go) when changing schema interfaces or adding new schema families.
- Keep docs, examples, generated fixtures, and task commands aligned with real exports and real files.

### Go 1.26 Features

| Feature | Where Used |
|---------|------------|
| Self-referential generic constraints | [core/constraints.go](core/constraints.go) defines fluent schema contracts |
| `new(expr)` | Pointer construction across parser, schema, locale, and JSON Schema helpers |
| `maps.Clone` and `slices.Clone` | Defensive copies for internals, payloads, object shapes, and JSON Schema bags |
| `slices.Sorted`, `maps.Keys`, `slices.SortFunc` | Deterministic JSON Schema and enum output |
| `errors.AsType[T]` | Typed error extraction in engine, issues, and composite schemas |
| `testing.B.Loop()` | Benchmarks across `core/`, `internal/`, `pkg/`, and root tag tests |

### Forbidden

- No hand-written parsing fast paths that bypass `internal/engine`.
- No new cross-imports between schema files in `types/`.
- No panic-based normal validation flow. Return errors for ordinary failures; reserve panics for `Must*` APIs and explicit invariant breaches.
- No stale doc claims. Do not mention commands, files, exports, examples, or reference paths without verifying them.
- No working around dependency bugs. If a bug or limitation is in a dependency library, create a report in [reports/](reports/) instead of reimplementing the dependency's functionality.
- No documentation masquerading as code. Keep contracts and explanations in docs and rules; do not encode prose into values no runtime path reads.
- No policy-only gates that only restate docs or rules. Enforce truth through behavior tests, generated-doc checks, validators, or lint rules when a program consumes the rule.
- No spec mirror tests that merely assert a rule sentence exists after stronger behavior coverage already proves the invariant.

## Testing

- Use the standard `testing` package with `testify/assert`, `testify/require`, and `go-cmp` where that is the local convention.
- Prefer focused subtests. Add `t.Parallel()` only when state isolation is explicit.
- Test both `Parse(any)` and `StrictParse(T)` for schema behavior.
- Test copy-on-write modifiers by proving the original schema remains unchanged.
- Write benchmarks with `b.Loop()`.
- Run `go test -tags=contractcheck ./types` when auditing compile-time schema coverage.
- Run `task docs:integrity` whenever README, docs, examples, `CLAUDE.md`, or public API names change.
- Run `task bench:smoke` to prove benchmark entry points compile without turning benchmark numbers into release criteria.

## Dependency Issue Reporting

When you encounter a bug, limitation, or unexpected behavior in a dependency library:

1. **Do NOT** work around it by reimplementing the dependency's functionality.
2. **Do NOT** skip or ignore the dependency and write your own version.
3. **Do** create a report file: `reports/<dependency-name>.md`.
4. **Do** include the dependency name and version, problem description, trigger scenario, expected behavior, actual behavior, relevant errors, and any suggested workaround.
5. **Do** continue with other tasks that do not depend on the broken functionality.

The `reports/` directory is checked after work cycles and routed to the appropriate dependency maintainer.

## Error Handling

- Return validation failures as `error`.
- Use `gozod.IsZodError` or `errors.AsType[*gozod.ZodError]` for typed inspection.
- Use `PrettifyError`, `FlattenError`, and `TreeifyError` for presentation instead of ad-hoc formatting.
- Keep `Must*` methods opt-in only. The default caller path should stay error-returning.

## Dependencies

- `github.com/go-json-experiment/json` - JSON v2 support used by struct parsing, validation helpers, examples, and generated code.
- `github.com/kaptinlin/jsonschema` - JSON Schema Draft 2020-12 conversion target and validation companion.
- `github.com/kaptinlin/deepclone` - deep cloning for copy-on-write schema internals.
- `github.com/golang-jwt/jwt/v5` - JWT structure validation.
- `golang.org/x/text` - Unicode normalization for string validation.
- `github.com/stretchr/testify` and `github.com/google/go-cmp` - test assertions and structural diffs.

## Agent Skills

Specialized skills in [`.agents/skills/`](.agents/skills/):

| Skill | When to Use |
|-------|-------------|
| [`go-best-practices`](.agents/skills/go-best-practices/) | Review or write Go APIs, naming, errors, concurrency, and tests |
| [`modernizing`](.agents/skills/modernizing/) | Adopt Go 1.20-1.26 language and standard-library improvements |
| [`golangci-linting`](.agents/skills/golangci-linting/) | Configure or fix golangci-lint v2 issues |
| [`library-test-covering`](.agents/skills/library-test-covering/) | Extend test coverage while staying consistent with existing patterns |
| [`taskfile-configuring`](.agents/skills/taskfile-configuring/) | Update Taskfile targets or command orchestration |
| [`library-code-optimizing`](.agents/skills/library-code-optimizing/) | Remove dead code and improve code quality while preserving behavior |
| [`library-code-simplifying`](.agents/skills/library-code-simplifying/) | Simplify library internals without changing public behavior |
| [`library-legacy-pruning`](.agents/skills/library-legacy-pruning/) | Delete deprecated APIs, legacy shims, and stale compatibility surface |
| [`library-docs-maintaining`](.agents/skills/library-docs-maintaining/) | Refresh `CLAUDE.md`, `AGENTS.md`, and `README.md` together |
| [`agent-md-writing`](.agents/skills/agent-md-writing/) | Regenerate `CLAUDE.md` and verify the `AGENTS.md` symlink |
| [`readme-writing`](.agents/skills/readme-writing/) | Regenerate the human-facing usage guide |
| [`gozod-validating`](.agents/skills/gozod-validating/) | Design or review validation flows built on GoZod |
| [`jsonschema-validating`](.agents/skills/jsonschema-validating/) | Work on JSON Schema validation or GoZod JSON Schema interoperability |

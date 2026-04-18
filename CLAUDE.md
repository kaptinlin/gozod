# GoZod

GoZod is a TypeScript Zod v4-inspired validation library for Go 1.26+. The root package is a user-facing facade over concrete schemas in `types/`, shared contracts in `core/`, runtime machinery in `internal/`, and JSON Schema translation in `jsonschema/`.

- **Reference implementation:** TypeScript Zod v4 in [`.reference/zod/`](.reference/zod/) when that submodule is initialized.
- **For user-facing examples and installation:** see [README.md](README.md).

## Commands

Run these from the repository root.

```bash
task test                         # Run the default test suite
task test:race                    # Run race-enabled tests for lightweight packages
task lint                         # Run golangci-lint and the tidy check
task golangci-lint                # Run golangci-lint only
task tidy-lint                    # Verify go.mod and go.sum stay tidy
task fmt                          # Run go fmt ./...
task vet                          # Run go vet ./...
task verify                       # Run deps, fmt, vet, lint, test, and govulncheck
go build ./...                    # Verify all packages compile
go test -tags=contractcheck ./types # Audit compile-time schema contracts
```

## Architecture

```text
gozod/
├── gozod_*.go       # Root facade: constructors, aliases, errors, metadata, JSON Schema entry points
├── cmd/gozodgen/    # Code generator for struct-tag schemas
├── coerce/          # z.coerce-style constructors for automatic input conversion
├── core/            # Public contracts, config, issue codes, shared types
├── docs/            # User documentation
├── examples/        # Runnable examples
├── internal/        # Engine, checks, issues, and runtime helpers
├── jsonschema/      # To/From JSON Schema conversion
├── locales/         # Localized validation messages
├── pkg/             # Shared helpers (clone, coercion, maps, regex, reflection, tags, transforms)
├── types/           # Concrete schema implementations and fluent APIs
├── .agents/rules/   # Schema, check, testing, naming, and structure guides
└── .reference/zod/  # Optional TypeScript Zod reference checkout
```

## Agent Workflow

- Read the relevant guides in [`.agents/rules/`](.agents/rules/) before changing schema internals, checks, naming, or tests.
- Verify every user-facing claim against the current repo state. Do not document commands, files, examples, or reference paths you have not checked.
- Initialize [`.reference/zod/`](.reference/zod/) before relying on it for parity work against TypeScript Zod.

## Design Philosophy

- **KISS** — Keep the public surface centered on root constructors, fluent modifiers, and the `Parse` / `StrictParse` pair.
- **Single Responsibility** — `core` defines contracts, `types` own schema behavior, `internal` owns execution, and `jsonschema` owns translation.
- **Open-Closed** — Extend behavior through checks, refinements, metadata, and new schema types instead of special-casing existing ones.
- **APIs as language** — Chains such as `gozod.String().Min(2).Email()` and `gozod.FromStruct[T]()` should read like the validation intent.
- **Beauty is structural** — The root package stays a thin facade while dependency direction remains one-way toward runtime internals.
- **Never:** accidental complexity, feature gravity, abstraction theater, configurability cope.

## API Design Principles

- **Progressive Disclosure**: Start from the root `gozod` package and `FromStruct[T]()`. Drop to `coerce/`, `types/`, `jsonschema/`, or `cmd/gozodgen` only when the simple path stops fitting.

## Coding Rules

### Must Follow

- Go 1.26.2 — use modern language features where they simplify code.
- Follow Google Go Best Practices: <https://google.github.io/go-style/best-practices>
- Follow Google Go Style Decisions: <https://google.github.io/go-style/decisions>
- KISS/DRY/YAGNI — no speculative helpers, no duplicated validation logic, no unused API surface.
- Preserve the two parsing modes: `Parse(any)` for dynamic inputs and `StrictParse(T)` for compile-time constrained inputs.
- Keep modifier methods copy-on-write. `Optional`, `Nilable`, `Default`, `Prefault`, `Describe`, `Meta`, and similar fluent modifiers must return a new configured schema.
- Route parsing through `internal/engine` helpers. Do not reimplement parsing pipelines in root facade files.
- Keep dependency direction one-way. Schema files in `types/` do not import each other; shared behavior belongs in `internal/`, `pkg/`, or `coerce/`.
- Update [types/constraints_verify.go](types/constraints_verify.go) when changing schema interfaces or adding new schema families.
- Keep docs, examples, and task commands aligned with real exports and real files.

### Go 1.26 Features

| Feature | Where Used |
|---------|------------|
| Self-referential generic constraints | [core/constraints.go](core/constraints.go) defines fluent schema contracts |
| `new(expr)` | Pointer construction across parser, schema, locale, and JSON Schema helpers |
| `maps.Clone` | [core/interfaces.go](core/interfaces.go) clones map-backed internals safely |
| `testing.B.Loop()` | Benchmarks across `core/`, `internal/`, `pkg/`, and root tag tests |

### Forbidden

- No hand-written parsing fast paths that bypass `internal/engine`.
- No new cross-imports between schema files in `types/`.
- No panic-based normal validation flow — return errors for ordinary failures. Reserve panics for `Must*` APIs and explicit invariant breaches.
- No stale doc claims — do not mention commands, files, or reference paths without verifying them.
- No working around dependency bugs — if a bug or limitation is in a dependency library, do not bypass it by reimplementing the dependency's functionality. Create a report in [reports/](reports/) instead.
- No documentation masquerading as code — keep contracts and explanations in docs; do not encode prose into values no runtime path reads.

## Testing

- Use the standard `testing` package with `testify/assert` and `testify/require`, which is the repository's current testing convention.
- Prefer focused subtests. Add `t.Parallel()` only when state isolation is explicit.
- Write benchmarks with `b.Loop()`.
- When public APIs or docs change, update examples and docs-integrity checks together.
- Run `go test -tags=contractcheck ./types` when auditing compile-time schema coverage.

## Dependency Issue Reporting

When you encounter a bug, limitation, or unexpected behavior in a dependency library:

1. **Do NOT** work around it by reimplementing the dependency's functionality.
2. **Do NOT** skip or ignore the dependency and write your own version.
3. **Do** create a report file: `reports/<dependency-name>.md`.
4. **Do** include in the report:
   - Dependency name and version
   - Problem description
   - Trigger scenario
   - Expected behavior vs actual behavior
   - Relevant error messages or stack traces
   - Workaround suggestion, if any, without implementing it
5. **Do** continue with other tasks that do not depend on the broken functionality.

The `reports/` directory is checked after work cycles and routed to the appropriate dependency maintainer.

## Error Handling

- Return validation failures as `error`.
- Use `gozod.IsZodError` or `errors.As` against `*gozod.ZodError` for typed inspection.
- Use `PrettifyError`, `FlattenError`, and `TreeifyError` for presentation instead of ad-hoc formatting.
- Keep `Must*` methods opt-in only. The default caller path should stay error-returning.

## Dependencies

- `github.com/go-json-experiment/json` — JSON v2 support used in examples and schema conversion flows.
- `github.com/kaptinlin/jsonschema` — JSON Schema Draft 2020-12 conversion target and validation companion.
- `github.com/kaptinlin/deepclone` — deep cloning for copy-on-write schema internals.
- `github.com/kaptinlin/go-i18n` and `golang.org/x/text` — localized validation messages.

## Agent Skills

Specialized skills in [`.claude/skills/`](.claude/skills/):

| Skill | When to Use |
|-------|-------------|
| [`go-best-practices`](.claude/skills/go-best-practices/) | Review or write Go APIs, naming, errors, concurrency, and tests |
| [`modernizing`](.claude/skills/modernizing/) | Adopt Go 1.20-1.26 language and stdlib improvements |
| [`golangci-linting`](.claude/skills/golangci-linting/) | Configure or fix golangci-lint v2 issues |
| [`library-test-covering`](.claude/skills/library-test-covering/) | Extend test coverage while staying consistent with existing patterns |
| [`taskfile-configuring`](.claude/skills/taskfile-configuring/) | Update Taskfile targets or command orchestration |
| [`committing`](.claude/skills/committing/) | Create conventional commits for completed changes |
| [`library-docs-maintaining`](.claude/skills/library-docs-maintaining/) | Refresh `CLAUDE.md`, `AGENTS.md`, and `README.md` together |
| [`agent-md-writing`](.claude/skills/agent-md-writing/) | Regenerate `CLAUDE.md` and the `AGENTS.md` symlink |
| [`readme-writing`](.claude/skills/readme-writing/) | Regenerate the human-facing usage guide |
| [`gozod-validating`](.claude/skills/gozod-validating/) | Design or review validation flows built on GoZod |
| [`jsonschema-validating`](.claude/skills/jsonschema-validating/) | Work on JSON Schema validation or GoZod JSON Schema interoperability |

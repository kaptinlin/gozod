# library-code-optimizing — gozod pass 2

## Baseline
- `go test ./...`: pass from module root.
- `golangci-lint run`: 22 existing findings: 2 gosec G115, 2 unused nolint directives, 18 staticcheck SA4006 unused assigned values.
- `.cov.before` total statements: 44.0%.

## Package order
1. pkg/coerce
2. cmd/gozodgen
3. internal/engine
4. jsonschema
5. types

## Per-package items
### pkg/coerce
- Other simplifying / go-best-practices targets: `pkg/coerce/coerce.go:198`, `pkg/coerce/coerce.go:376` — make the checked uint-to-int64 conversions explicit enough for gosec without changing overflow behavior.

### cmd/gozodgen
- Exported-symbol exclusions: `cmd/gozodgen/analyzer.go:62` retains the `staticcheck` nolint because the project-pinned linter still reports SA1019.

### internal/engine
- Exported-symbol exclusions: `internal/engine/checker.go:71` retains the `gosec` nolint because the project-pinned linter still reports G602.
- Other simplifying / go-best-practices targets: `internal/engine/parser.go:1005` — remove unused type-assertion binding while preserving pointer conversion behavior.

### jsonschema
- Other simplifying / go-best-practices targets: `jsonschema/to.go:565`, `jsonschema/to.go:570`, `jsonschema/to.go:919`, `jsonschema/to.go:923`, `jsonschema/to.go:939`, `jsonschema/to.go:943`, `jsonschema/to.go:947`, `jsonschema/to.go:951`, `jsonschema/to.go:955`, `jsonschema/to.go:959`, `jsonschema/to.go:963`, `jsonschema/to.go:1318`, `jsonschema/to.go:1324`, `jsonschema/to.go:1329` — avoid assigning conversion results that are only passed to `new`.
- Other simplifying / go-best-practices targets: `jsonschema/to.go:801` — remove unused `lastRequired` assignment by computing tuple `minItems` directly.

### types
- Other simplifying / go-best-practices targets: `types/file_test.go:509`, `types/function_test.go:570` — remove unused type-assertion bindings in tests.

## Exported-symbol exclusions
- No exported symbols are in scope for deletion or rename in this pass.

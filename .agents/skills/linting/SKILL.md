---
name: linting
description: Sets up and runs golangci-lint v2 for Go projects. Use when adding linting to a Go package, configuring golangci-lint, fixing lint errors, or integrating linters into Makefile and CI. Triggers on lint setup, golangci-lint configuration, or lint fix requests in Go projects.
---

# Go Linting with golangci-lint v2

Set up, configure, and run golangci-lint v2 in Go projects. Follow these patterns strictly.

## Quick Reference

```bash
# Run linters
golangci-lint run ./...

# Run formatters
golangci-lint fmt ./...

# Auto-fix where supported
golangci-lint run --fix ./...

# Specific directories
golangci-lint run dir1/... dir2/...
```

## Project Setup

Every Go package needs three linting files:

### 1. `.golangci.version` — Pin the version

```
2.9.0
```

Single line, no newline. This file is the source of truth for the golangci-lint version used by the project.

### 2. `.golangci.yml` — Linter configuration

See `references/recommended-config.md` for the recommended configuration template.

Key rules:
- Always set `version: "2"` (required for v2)
- Enable linters explicitly under `linters.enable`
- Exclude `gosec`, `noctx`, and `revive` in test files
- Set `max-issues-per-linter: 0` and `max-same-issues: 0` to show all issues

### 3. Makefile targets

See `references/makefile-integration.md` for standard Makefile targets and CI workflow.

The Makefile must:
- Auto-install golangci-lint from `.golangci.version`
- Install to local `./bin/` (not global `$GOPATH/bin`)
- Include `lint`, `golangci-lint`, `tidy-lint`, `fmt`, and `vet` targets
- Support multi-module repositories via `MODULE_DIRS`

## Fixing Lint Issues

When fixing lint errors, follow this workflow:

1. Run `task lint` to see all issues
2. Fix issues by category, not file-by-file
3. Never suppress warnings with `//nolint` unless there is a justified reason (add explanation: `//nolint:lintername // reason`)
4. After fixing, run `task lint` again to confirm zero issues
5. Run `task test` to ensure fixes don't break behavior

### Common Fix Patterns

**errcheck** — Handle or explicitly ignore errors:
```go
// Bad
json.Unmarshal(data, &v)

// Good
if err := json.Unmarshal(data, &v); err != nil {
    return fmt.Errorf("unmarshal config: %w", err)
}
```

**err113** — Use sentinel errors or wrap with `%w`:
```go
// Bad
return errors.New("not found")

// Good (sentinel)
var ErrNotFound = errors.New("not found")
return ErrNotFound

// Good (wrap)
return fmt.Errorf("lookup %s: %w", key, ErrNotFound)
```

**errorlint** — Use `errors.Is`/`errors.As` instead of `==`:
```go
// Bad
if err == io.EOF { ... }

// Good
if errors.Is(err, io.EOF) { ... }
```

**gosec** — Address security warnings:
```go
// Bad: G404 - weak random
val := rand.Intn(100)

// Good: use crypto/rand
val, err := rand.Int(rand.Reader, big.NewInt(100))
```

**prealloc** — Pre-allocate slices when length is known:
```go
// Bad
var results []string
for _, item := range items {
    results = append(results, item.Name)
}

// Good
results := make([]string, 0, len(items))
for _, item := range items {
    results = append(results, item.Name)
}
```

**ineffassign** — Remove or use assigned values:
```go
// Bad
x := computeValue()
x = otherValue() // previous assignment never used

// Good
_ = computeValue() // if side effects needed
x := otherValue()
```

**nakedret** — Use explicit returns in long functions:
```go
// Bad (function > 5 lines with named returns)
func parse(s string) (result int, err error) {
    // ... many lines ...
    return // naked return
}

// Good
func parse(s string) (int, error) {
    // ... many lines ...
    return result, nil
}
```

## Adding golangci-lint to an Existing Project

1. Create `.golangci.version` with the latest stable version
2. Create `.golangci.yml` from the recommended config template
3. Add Makefile targets from the integration template
4. Add CI workflow (GitHub Actions)
5. Run `task lint` and fix all issues before committing
6. Commit all three config files together:
   ```bash
   git add .golangci.version .golangci.yml Makefile
   git commit -m "build(lint): add golangci-lint v2 configuration"
   ```

## Linter Selection Guide

**Always enable** (default set):
- `errcheck`, `govet`, `ineffassign`, `staticcheck`, `unused`

**Recommended additions**:
- Error handling: `err113`, `errorlint`, `nilerr`
- Code quality: `misspell`, `revive`, `gocritic`, `unconvert`
- Style: `whitespace`, `nakedret`, `dogsled`, `copyloopvar`
- Security: `gosec`
- Best practices: `exhaustive`, `noctx`, `nolintlint`, `prealloc`

**Do NOT enable** unless project-specific need:
- `gochecknoglobals` — too restrictive for most projects
- `wrapcheck` — excessive wrapping in internal code
- `funlen` / `cyclop` — arbitrary limits, use code review instead

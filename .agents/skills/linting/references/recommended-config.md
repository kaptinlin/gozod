# Recommended golangci-lint v2 Configuration

## `.golangci.yml`

```yaml
version: 2
run:
  timeout: 5m
  tests: true

linters:
  enable:
    - errcheck
    - govet
    - ineffassign
    - staticcheck
    - unused
    - misspell
    - revive
    - whitespace
    - err113
    - errorlint
    - nilerr
    - gocritic
    - nakedret
    - unconvert
    - dogsled
    - copyloopvar
    - prealloc
    - gosec
    - exhaustive
    - noctx
    - nolintlint
    - promlinter

issues:
  max-issues-per-linter: 0
  max-same-issues: 0
  exclude-rules:
    - path: _test\.go
      linters:
        - gosec
        - noctx
        - revive
```

## Configuration Reference

### `version`
Required. Must be `"2"` for golangci-lint v2.

### `run`
- `timeout`: Maximum time for analysis. `5m` is recommended; increase for large codebases.
- `tests`: Set to `true` to lint test files too.

### `linters.enable`
Explicit list of enabled linters. Using explicit enable (not `enable-all`) ensures predictable behavior across version upgrades.

### `issues`
- `max-issues-per-linter: 0` — No limit, show all issues from each linter.
- `max-same-issues: 0` — No limit on duplicate issues.
- `exclude-rules` — Relax rules in test files where strict security or HTTP context checking is unnecessary.

### `.golangci.version`
Separate file containing only the version number (e.g., `2.9.0`). Used by Makefile for automatic installation. Keeping it separate from `.golangci.yml` allows tooling to read the version without parsing YAML.

## Minimal Configuration

For small projects or getting started:

```yaml
version: 2
run:
  timeout: 5m

linters:
  enable:
    - errcheck
    - govet
    - ineffassign
    - staticcheck
    - unused
    - misspell
    - errorlint
    - gosec
```

## Adding Linter-Specific Settings

```yaml
version: 2
run:
  timeout: 5m
  tests: true

linters:
  enable:
    - errcheck
    - govet
    - revive
    # ... other linters

  settings:
    revive:
      rules:
        - name: exported
          arguments:
            - disableStutteringCheck
    exhaustive:
      default-signifies-exhaustive: true
    nakedret:
      max-func-lines: 30
    gocritic:
      enabled-tags:
        - diagnostic
        - style
        - performance
```

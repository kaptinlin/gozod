# Go Toolchain for Code Refactoring

Tools, configurations, and commands for maintaining Go code quality. Go 1.26+.

## golangci-lint Configuration

Recommended `.golangci.yml` for refactoring-oriented projects:

```yaml
run:
  timeout: 5m
  go: "1.26"

linters:
  enable:
    - gofmt          # Format correctness
    - goimports      # Import grouping and ordering
    - govet          # Suspicious constructs (printf args, struct tags, etc.)
    - staticcheck    # Advanced static analysis (SA*, S*, ST*, QF* checks)
    - gosimple       # Simplification suggestions (S1000-S1040)
    - unused         # Unused code detection
    - errcheck       # Unchecked error returns
    - ineffassign    # Ineffectual assignments
    - misspell       # Common spelling mistakes in comments/strings
    - revive         # Extensible linter (replaces golint)

linters-settings:
  revive:
    rules:
      - name: exported
        arguments: [checkPrivateReceivers]
      - name: unused-parameter
      - name: blank-imports
      - name: context-as-argument
      - name: error-return
      - name: error-naming
      - name: if-return
      - name: increment-decrement
      - name: var-declaration
      - name: range
      - name: receiver-naming
      - name: time-naming
      - name: unexported-return
      - name: indent-error-flow
      - name: errorf
      - name: empty-block
      - name: superfluous-else
      - name: unreachable-code
  staticcheck:
    checks: ["all"]
  govet:
    enable-all: true

issues:
  exclude-rules:
    - path: _test\.go
      linters: [errcheck]
  max-issues-per-linter: 0
  max-same-issues: 0
```

Run:
```bash
# Full lint
golangci-lint run ./...

# Specific linters only
golangci-lint run --enable-only errcheck,unused ./...

# Auto-fix where possible
golangci-lint run --fix ./...

# Show suggested fixes
golangci-lint run --out-format line-number ./...
```

## Static Analysis Tools

### go vet

Built-in. Catches common mistakes the compiler misses.

```bash
# All packages
go vet ./...

# Specific checks
go vet -printf ./...       # Printf format string mismatches
go vet -structtag ./...    # Malformed struct tags
go vet -copylocks ./...    # Passing locks by value
```

Key checks: printf args, struct tags, unreachable code, copy of locks, nil function comparison, shift overflow.

### staticcheck

Most comprehensive Go static analyzer. Catches bugs, suggests simplifications.

```bash
# Install
go install honnef.co/go/tools/cmd/staticcheck@latest

# Run all checks
staticcheck ./...

# Specific check categories
staticcheck -checks "SA*" ./...   # Bug detection
staticcheck -checks "S*" ./...    # Simplifications
staticcheck -checks "ST*" ./...   # Style issues
staticcheck -checks "QF*" ./...   # Quick fixes
```

Useful checks for refactoring:
- `SA4006` — assigned value never used
- `SA9003` — empty branch
- `S1000` — single-case select → direct receive
- `S1002` — `bool == true/false` → simplify
- `S1005` — unnecessary blank identifier
- `S1008` — `if/else` returning bool → return condition

### gosec (Security)

Finds security vulnerabilities.

```bash
# Install
go install github.com/securego/gosec/v2/cmd/gosec@latest

# Run
gosec ./...

# Exclude specific rules
gosec -exclude=G104 ./...  # Skip unchecked errors (if errcheck covers it)

# JSON output for CI
gosec -fmt json -out results.json ./...
```

Key rules: G101 (hardcoded credentials), G104 (unchecked errors), G201 (SQL injection), G301 (file permissions), G401 (weak crypto).

### govulncheck

Checks dependencies for known vulnerabilities.

```bash
# Install
go install golang.org/x/vuln/cmd/govulncheck@latest

# Check all dependencies
govulncheck ./...

# Check binary
govulncheck -mode=binary ./cmd/myapp
```

## Code Quality Tools

### gofumpt (Stricter gofmt)

Enforces stricter formatting rules on top of `gofmt`.

```bash
# Install
go install mvdan.cc/gofumpt@latest

# Format
gofumpt -w .

# Check without modifying
gofumpt -d .

# Extra strict
gofumpt -extra -w .
```

Rules beyond gofmt: no empty lines at start/end of blocks, grouped declarations, consistent composite literal formatting.

### golines (Line Length)

Enforces maximum line length by reflowing long lines.

```bash
# Install
go install github.com/segmentio/golines@latest

# Reformat lines over 120 chars
golines -w --max-len=120 .

# Preview changes
golines --max-len=120 --dry-run .

# Reformat with goimports
golines -w --max-len=120 --base-formatter=goimports .
```

### gci (Import Ordering)

Enforces consistent import grouping: stdlib, external, local.

```bash
# Install
go install github.com/daixiang0/gci@latest

# Format imports
gci write --section Standard --section Default --section "Prefix(github.com/yourorg)" .

# Check without modifying
gci diff --section Standard --section Default --section "Prefix(github.com/yourorg)" .
```

Result:
```go
import (
    // Standard library
    "context"
    "fmt"

    // External
    "github.com/go-chi/chi/v5"
    "go.uber.org/zap"

    // Local
    "github.com/yourorg/myapp/internal/user"
)
```

## Performance Analysis

### pprof (CPU and Memory Profiling)

```bash
# CPU profile from test
go test -cpuprofile=cpu.prof -bench=BenchmarkProcess ./...
go tool pprof cpu.prof

# Memory profile from test
go test -memprofile=mem.prof -bench=BenchmarkProcess ./...
go tool pprof mem.prof

# HTTP server profiling (add import _ "net/http/pprof" to main)
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30
go tool pprof http://localhost:6060/debug/pprof/heap

# Common pprof commands
# (pprof) top 20            # Top 20 CPU consumers
# (pprof) list FunctionName  # Line-by-line breakdown
# (pprof) web               # Open flame graph in browser
```

### trace

Visualize goroutine scheduling, GC pauses, and syscalls.

```bash
# Generate trace from test
go test -trace=trace.out ./...

# View in browser
go tool trace trace.out
```

### benchstat

Compare benchmark results across runs to validate refactoring performance.

```bash
# Install
go install golang.org/x/perf/cmd/benchstat@latest

# Before refactoring
go test -bench=. -count=10 ./... > before.txt

# After refactoring
go test -bench=. -count=10 ./... > after.txt

# Compare
benchstat before.txt after.txt
```

Output:
```
name           old time/op  new time/op  delta
Process-8      45.2ms ± 3%  38.1ms ± 2%  -15.71% (p=0.000 n=10+10)
```

## Testing Tools

### go test -race

Detect data races. Run during refactoring to ensure concurrency safety.

```bash
# Race detection on all packages
go test -race ./...

# Race detection with coverage
go test -race -cover ./...

# Race detection on specific package
go test -race ./internal/worker/...
```

### go test -cover

Measure test coverage. Useful to verify refactored code is still covered.

```bash
# Coverage summary
go test -cover ./...

# Coverage profile
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Coverage by function
go tool cover -func=coverage.out

# Coverage for specific package
go test -coverprofile=coverage.out -coverpkg=./internal/... ./...
```

### gotestsum

Better test output formatting. Useful during refactoring to spot failures fast.

```bash
# Install
go install gotest.tools/gotestsum@latest

# Run with summary
gotestsum ./...

# Show only failures
gotestsum --format short ./...

# JUnit XML output for CI
gotestsum --junitfile results.xml ./...

# Re-run failures
gotestsum --rerun-fails=3 ./...
```

## Refactoring-Specific Commands

Quick commands for common refactoring tasks:

```bash
# Find unused exports across the module
# (staticcheck S1023 + unused linter catch most cases)
golangci-lint run --enable-only unused ./...

# List all packages and their import counts
go list -f '{{.ImportPath}}: {{len .Imports}} imports' ./...

# Find import cycles (will fail with error if cycle exists)
go build ./...

# Check for deprecated API usage
staticcheck -checks "SA1019" ./...

# Find functions with too many parameters (revive)
golangci-lint run --enable-only revive ./...

# Analyze binary size (find bloat)
go build -o app ./cmd/myapp && go tool nm -size app | sort -n -r | head -20
```

## References

- [golangci-lint](https://golangci-lint.run/)
- [staticcheck](https://staticcheck.dev/)
- [gosec](https://github.com/securego/gosec)
- [govulncheck](https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck)
- [pprof](https://pkg.go.dev/net/http/pprof)

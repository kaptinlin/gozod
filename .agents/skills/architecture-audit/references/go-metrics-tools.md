# Go Metrics Tools

Essential tools for Phase 5 (Quantitative Metrics) in architecture audit.

## Package-Level Metrics

### Lines of Code per Package

```bash
# Count LOC per package (excluding tests)
for pkg in pkg/*; do
  if [ -d "$pkg" ]; then
    loc=$(find "$pkg" -name "*.go" -not -name "*_test.go" -exec wc -l {} + 2>/dev/null | tail -1 | awk '{print $1}')
    echo "$(basename $pkg) $loc"
  fi
done | sort -k2 -rn
```

**Threshold:** Max 2000 LOC per package

### Package Coupling (Import Count)

```bash
# Count imports per package
go list -f '{{.ImportPath}} {{join .Imports " "}}' ./pkg/... 2>/dev/null | \
  awk '{print $1, NF-1}' | sort -k2 -rn
```

**Threshold:** Max 15 imports per package

## Test Coverage

### Per-Package Coverage

```bash
# Run tests with coverage
go test -cover ./... 2>&1 | grep -E "coverage:|ok"
```

**Threshold:** Min 80% coverage per package

### Coverage Report

```bash
# Generate detailed coverage report
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | grep -v "total:"
```

## Using golangci-lint

**Note:** Function-level metrics (cyclomatic complexity, dead code) are handled by `task lint` (golangci-lint).

```bash
# Run linter (includes complexity, unused code, etc.)
task lint

# Or directly
golangci-lint run --timeout=10m
```

**Linters included:**
- `gocyclo` - cyclomatic complexity
- `unused` - dead code detection
- `staticcheck` - static analysis
- `gosec` - security issues
- Many more (see .golangci.yml)

## Thresholds Summary

| Metric | Threshold | Tool |
|--------|-----------|------|
| Package LOC | ≤2000 | Manual (wc -l) |
| Package imports | ≤15 | Manual (go list) |
| Test coverage | ≥80% | go test -cover |
| Cyclomatic complexity | ≤15 | golangci-lint (gocyclo) |
| Dead code | 0 | golangci-lint (unused) |

## Quick Metrics Script

Save as `collect-metrics.sh`:

```bash
#!/bin/bash
# Quick metrics collection for architecture audit

echo "=== Package Sizes (LOC) ==="
for pkg in pkg/*; do
  if [ -d "$pkg" ]; then
    loc=$(find "$pkg" -name "*.go" -not -name "*_test.go" -exec wc -l {} + 2>/dev/null | tail -1 | awk '{print $1}')
    echo "$(basename $pkg) $loc"
  fi
done | sort -k2 -rn | head -10

echo ""
echo "=== Package Coupling (Imports) ==="
go list -f '{{.ImportPath}} {{join .Imports " "}}' ./pkg/... 2>/dev/null | \
  awk '{print $1, NF-1}' | sort -k2 -rn | head -10

echo ""
echo "=== Test Coverage ==="
go test -cover ./... 2>&1 | grep -E "coverage:" | head -10
```

**Usage:**
```bash
chmod +x collect-metrics.sh
./collect-metrics.sh
```

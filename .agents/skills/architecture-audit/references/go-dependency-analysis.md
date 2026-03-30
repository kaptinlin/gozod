# Go Dependency Analysis

Tools and patterns for analyzing dependencies in Go 1.26+ projects.

## Circular Dependency Detection

### Using go mod graph

```bash
# Basic circular dependency check
go mod graph | awk '{print $1}' | sort -u > deps.txt
go mod graph | awk '{print $2}' | sort -u >> deps.txt
sort deps.txt | uniq -d

# Using goda (more powerful)
go install github.com/loov/goda@latest
goda graph "reach(./...)" | goda cut "reach(./pkg/...):reach(./internal/...)"
```

### Manual Package Import Check

```bash
# Check if pkg imports internal (should be empty)
grep -r "import.*internal" pkg/

# Check import cycles within packages
go list -f '{{.ImportPath}} {{join .Imports " "}}' ./... | \
  awk '{for(i=2;i<=NF;i++) if($i ~ /^github.com\/yourorg\/yourrepo/) print $1, $i}'
```

## Layer Boundary Validation

### Standard Go Project Layers

```
cmd/          # Entry points (imports pkg, internal)
  └─ pkg/     # Public libraries (imports internal only)
      └─ internal/  # Private code (no external imports from project)
```

### Validation Commands

```bash
# Verify cmd doesn't import other cmd packages
find cmd -name "*.go" -exec grep -l "import.*cmd/" {} \;

# Verify pkg doesn't import cmd
find pkg -name "*.go" -exec grep -l "import.*cmd/" {} \;

# Check internal is only imported by same-parent packages
go list -f '{{.ImportPath}}: {{join .Imports "\n"}}' ./... | \
  grep "internal" | grep -v "$(pwd | xargs basename)"
```

## Package Coupling Analysis

### Dependency Graph Visualization

```bash
# Generate visual dependency graph
go install golang.org/x/exp/cmd/modgraphviz@latest
go mod graph | modgraphviz | dot -Tsvg -o deps.svg

# Using go-mod-graph-chart
go install github.com/nikolaydubina/go-mod-graph-chart@latest
go mod graph | go-mod-graph-chart -o deps.svg
```

### Package Dependency Count

```bash
# Count internal dependencies per package
go list -f '{{.ImportPath}} {{len .Imports}}' ./... | \
  grep "$(go list -m)" | sort -k2 -n

# List packages with most dependencies (coupling hotspots)
go list -f '{{.ImportPath}} {{len .Imports}}' ./... | \
  grep "$(go list -m)" | sort -k2 -rn | head -10
```

## Anti-Patterns and Solutions

### Anti-Pattern: Circular Import

```go
// pkg/a/a.go
package a
import "project/pkg/b"  // A imports B

// pkg/b/b.go
package b
import "project/pkg/a"  // B imports A ❌ Circular
```

### Solution 1: Extract Common Interface

```go
// pkg/common/interface.go
package common
type Service interface { Do() }

// pkg/a/a.go
package a
import "project/pkg/common"
func Use(s common.Service) { s.Do() }

// pkg/b/b.go
package b
import "project/pkg/common"
type Impl struct{}
func (i Impl) Do() {}  // Implements common.Service
```

### Solution 2: Dependency Inversion

```go
// pkg/a/a.go
package a
type Dependency interface { Helper() }  // A defines interface
func Process(d Dependency) { d.Helper() }

// pkg/b/b.go
package b
import "project/pkg/a"
type BImpl struct{}
func (b BImpl) Helper() {}  // B implements A's interface
```

### Solution 3: Move to Lower Layer

```go
// If both A and B need shared code, move it down
// pkg/internal/shared/shared.go
package shared
func CommonLogic() {}

// pkg/a/a.go
package a
import "project/pkg/internal/shared"

// pkg/b/b.go
package b
import "project/pkg/internal/shared"
```

## Go 1.26 Specific Features

### Using go.work for Multi-Module Projects

```bash
# Initialize workspace
go work init ./module1 ./module2

# Check workspace dependencies
go work sync
go list -m all
```

### Module Graph Analysis

```bash
# Show why a module is needed
go mod why -m github.com/some/module

# Show module dependency graph
go mod graph | grep "^$(go list -m)"

# Prune unused dependencies
go mod tidy
```

## Thresholds and Metrics

| Metric | Threshold | Action |
|--------|-----------|--------|
| Package imports | <10 | Good |
| Package imports | 10-15 | Review coupling |
| Package imports | >15 | Refactor (likely god package) |
| Circular deps | 0 | Required |
| Layer violations | 0 | Required |

## Common Issues

### Issue: Import Cycle Not Allowed

```
package project/pkg/a
    imports project/pkg/b
    imports project/pkg/a: import cycle not allowed
```

**Fix:** Apply one of the three solutions above (extract interface, DIP, or move to lower layer).

### Issue: Package Coupling Too High

**Symptom:** Package imports 15+ other packages

**Fix:**
1. Identify cohesive subsets of functionality
2. Extract to separate packages
3. Use interfaces to reduce coupling
4. Apply dependency inversion where appropriate

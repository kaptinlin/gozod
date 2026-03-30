# Go Refactoring Patterns

Real-world refactoring patterns for Go codebases based on code-modularizing principles.

## Pattern 1: Package Overload → Layered Extraction

### Problem
Single package accumulates 20+ files mixing multiple concerns.

### Symptoms
```go
// validator/ package (20 files)
validator/
├── schema.go              // Schema validation
├── schema_error.go
├── document.go            // Document model
├── loader.go              // File loading
├── index.go               // Entity indexing
├── runner.go              // Validation engine
├── registry.go
└── rules/                 // 10 rule files
```

### Analysis
```bash
# Count files
find validator -name "*.go" | wc -l
# Output: 35

# Identify concerns by file patterns
ls validator/*schema*.go    # Schema validation cluster
ls validator/*document*.go  # Document model cluster
ls validator/*index*.go     # Entity indexing cluster
```

### Solution: Extract by Layer

**Step 1: Identify concerns**
1. Schema validation (3 files)
2. Document model (2 files)
3. Entity indexing (1 file)
4. Validation engine (core, keep)
5. Domain rules (keep)

**Step 2: Extract in dependency order**

```go
// Before: validator.Document used everywhere
package validator
type Document struct {
    FilePath string
    Data     map[string]any
    Content  string
}

// After: Extract to pkg/document
package document
type Document struct {
    FilePath string
    Data     map[string]any
    Content  string
}

// validator/ now imports pkg/document
package validator
import "yourmodule/pkg/document"

func (r *Runner) Validate(ctx context.Context, docs []*document.Document) []Issue
```

**Step 3: Extract schema validation**

```go
// Before: validator.SchemaValidator
package validator
type SchemaValidator struct {
    schemas map[string]*Schema
}

// After: Extract to pkg/schema
package schema
type Validator struct {
    schemas map[string]*Schema
}

// validator/rules uses pkg/schema
package rules
import "yourmodule/pkg/schema"

func (r *SchemaRule) Check(ctx context.Context, docs []*document.Document) []validator.Issue {
    v := schema.NewValidator()
    // ...
}
```

**Result:**
- validator/ reduced from 35 files to 10 files
- 3 new reusable packages: pkg/document, pkg/schema, pkg/entity
- Clear separation of concerns

---

## Pattern 2: Duplicate Packages → Consolidation

### Problem
Two packages with same name in different locations.

### Symptoms
```go
// Both exist with overlapping functionality
internal/config/  (25 files, 3000 lines)
pkg/config/       (5 files, 200 lines)

// Import confusion
import "yourmodule/internal/config"  // 12 locations
import "yourmodule/pkg/config"       // 18 locations
```

### Analysis
```bash
# Count imports
grep -r "internal/config" --include="*.go" | wc -l
# Output: 12

grep -r "pkg/config" --include="*.go" | wc -l
# Output: 18

# Compare sizes
find internal/config -name "*.go" | wc -l  # 25 files
find pkg/config -name "*.go" | wc -l       # 5 files
```

### Decision Tree

**Question 1: Which has more complete implementation?**
- internal/config: Full implementation (loader, parser, validator)
- pkg/config: Minimal stub (just types)

**Question 2: Which should be public?**
- Types and loader should be public (reusable)
- Parser internals can stay internal (implementation detail)

### Solution: Split and Merge

```go
// Step 1: Keep core types in pkg/config
package config
type Config struct { ... }
type Loader struct { ... }
func Load(path string) (*Config, error)

// Step 2: Extract parser to internal/parser
package parser
import "yourmodule/pkg/config"
func parse(data []byte) (*config.Config, error)

// Step 3: Remove duplicate pkg/config stub
// Merge any unique functionality into main pkg/config
```

**Result:**
- pkg/config: Complete public API
- internal/parser: Implementation details
- No duplicate package names
- Clear public/private boundary

---

## Pattern 3: Three-Layer Architecture

### Problem
Package mixes file I/O, type system, and business logic.

### Symptoms
```go
// internal/processor/ mixing everything
internal/processor/
├── reader.go          // File I/O
├── types.go           // Type definitions
├── registry.go        // Type system
├── handler.go         // Business logic
└── pipeline.go        // Orchestration
```

### Solution: Strict Layering

```go
// Layer 1: File I/O (zero dependencies)
package io
func ReadFile(path string) ([]byte, error)
func WriteFile(path string, data []byte) error

// Layer 2: Type system (zero dependencies)
package types
type Entity string
type Descriptor struct { ... }
func Register(d *Descriptor) error

// Layer 3: Business logic (depends on Layer 1 + 2)
package processor
import (
    "yourmodule/internal/io"
    "yourmodule/pkg/types"
)
type Handler struct { ... }
```

**Dependency flow:**
```
Layer 3 (business) → Layer 2 (types) → Layer 1 (I/O)
                  ↘                  ↗
                    Never reverse!
```

---

## Pattern 4: internal/ → pkg/ Promotion

### Problem
Reusable utilities hidden in internal/.

### Symptoms
```go
// internal/fileutil/ - generic file utilities
package fileutil
func Walk(dir string, pattern string) ([]string, error)
func FindByName(dir, name string) (string, error)
func EnsureDir(path string) error

// Used by cmd/ but blocked from external use
```

### Solution: Promote to pkg/

```bash
# Move to public API
git mv internal/fileutil pkg/fileutil

# Update imports (5 locations)
# Before: import "yourmodule/internal/fileutil"
# After:  import "yourmodule/pkg/fileutil"
```

**Decision criteria:**
- ✅ Generic, reusable logic → pkg/
- ✅ No business logic → pkg/
- ❌ Project-specific → internal/
- ❌ Implementation details → internal/

---

## Pattern 5: Dead Code Detection

### Problem
Package with zero imports.

### Analysis
```bash
# Check if package is imported
grep -r "internal/legacy" --include="*.go" | wc -l
# Output: 0

# Package exists but unused
internal/legacy/
├── converter.go   # 150 lines
├── parser.go
└── converter_test.go
```

### Solution
```go
// Option 1: Delete if truly unused
git rm -r internal/legacy/

// Option 2: If needed later, document why it's kept
// Add README explaining future plans
```

---

## Refactoring Checklist

### Before Starting
- [ ] Count files per package
- [ ] Measure import usage
- [ ] Identify duplicate packages
- [ ] Map dependency graph
- [ ] List all concerns per package

### During Refactoring
- [ ] Extract in dependency order (zero-dep first)
- [ ] Update imports incrementally
- [ ] Run tests after each extraction
- [ ] Check for import cycles: `go build ./...`
- [ ] Verify no circular dependencies

### After Completion
- [ ] All tests pass: `go test ./...`
- [ ] No import cycles: `go build ./...`
- [ ] Linter passes: `golangci-lint run`
- [ ] Documentation updated
- [ ] Clear package responsibilities

---

## Common Mistakes

### ❌ Extracting Too Early
```go
// Only 1 consumer, don't extract yet
internal/helper/  // Used only by cmd/
```

### ❌ Breaking Dependency Flow
```go
// Lower layer importing higher layer
pkg/types/     → imports → pkg/pipeline/  // WRONG!
```

### ❌ Leaving Duplicates
```go
// Same logic in 3 places
pkg/parser/parse.go
internal/reader/parse.go
cmd/tool/parse.go
```

### ✅ Correct Approach
```go
// Consolidate into one canonical location
pkg/parser/parse.go  // Single source of truth
```

---

## Tools and Commands

### Analysis
```bash
# Count files per package
find pkg/ -name "*.go" | cut -d/ -f1-2 | sort | uniq -c

# Count lines per package
find pkg/ -name "*.go" -exec wc -l {} + | sort -n

# Find import usage
grep -r "import.*internal/config" --include="*.go" | wc -l

# Detect import cycles
go build ./...
```

### Refactoring
```bash
# Move package
git mv internal/fileutil pkg/fileutil

# Update imports (use gofmt or goimports)
goimports -w .

# Run tests
go test ./...

# Check for issues
golangci-lint run
```

# Go Dependency Inversion Patterns

## Core Principle

**Dependencies flow inward:** Outer layers (CLI, HTTP, infrastructure) depend on inner layers (domain, business logic). Inner layers never import outer layers.

## Layer Architecture

```
cmd/                          ← Presentation (CLI, formatters, I/O)
    ↓ imports
internal/orchestration/       ← Orchestration (workflows, use cases)
    ↓ imports
pkg/domain/                   ← Domain (business logic, rules)
    ↓ imports
pkg/interfaces/               ← Abstractions (interfaces only)
```

## Anti-Pattern: Presentation in Core

### ❌ Wrong: Core package handles formatting

```go
// pkg/validation/validator.go
package validation

import "fmt"

type Validator struct{}

func (v *Validator) Validate(data Data) string {
    if data.Name == "" {
        return "ERROR: name is required\n"  // ❌ Formatting in core
    }
    return "OK\n"
}
```

### ✅ Correct: Core returns structured data

```go
// pkg/validation/validator.go
package validation

type Result struct {
    Valid  bool
    Errors []Error
}

type Error struct {
    Field   string
    Message string
}

func (v *Validator) Validate(data Data) Result {
    var errs []Error
    if data.Name == "" {
        errs = append(errs, Error{Field: "name", Message: "required"})
    }
    return Result{Valid: len(errs) == 0, Errors: errs}
}

// cmd/app/formatter.go
package main

import "fmt"

func formatResult(r validation.Result) string {
    if r.Valid {
        return "OK\n"
    }
    return fmt.Sprintf("ERROR: %s\n", r.Errors[0].Message)
}
```

## Anti-Pattern: I/O in Pure Logic

### ❌ Wrong: Algorithm reads files directly

```go
// pkg/analyzer/analyzer.go
package analyzer

import "os"

func Analyze(path string) (Report, error) {
    data, err := os.ReadFile(path)  // ❌ I/O in core
    if err != nil {
        return Report{}, err
    }
    return process(data), nil
}
```

### ✅ Correct: Inject I/O via interface

```go
// pkg/analyzer/analyzer.go
package analyzer

type Reader interface {
    Read(path string) ([]byte, error)
}

func Analyze(r Reader, path string) (Report, error) {
    data, err := r.Read(path)
    if err != nil {
        return Report{}, err
    }
    return process(data), nil
}

// cmd/app/main.go
package main

type FileReader struct{}

func (f FileReader) Read(path string) ([]byte, error) {
    return os.ReadFile(path)
}

func main() {
    analyzer := analyzer.New()
    report, _ := analyzer.Analyze(FileReader{}, "data.json")
}
```

## Anti-Pattern: Framework Depends on Application

### ❌ Wrong: Validation framework imports CLI formatter

```go
// pkg/lint/lint.go
package lint

import "myapp/cmd/cli"  // ❌ Core imports presentation

type Runner struct {
    formatter *cli.Formatter  // ❌ Presentation type in core
}

func (r *Runner) Check(docs []Doc) string {
    diags := r.validate(docs)
    return r.formatter.Format(diags)  // ❌ Formatting in core
}
```

### ✅ Correct: Return structured data, format in CLI

```go
// pkg/lint/lint.go
package lint

type Diagnostic struct {
    File     string
    Line     int
    Severity string
    Message  string
}

type Runner struct{}

func (r *Runner) Check(docs []Doc) []Diagnostic {
    return r.validate(docs)  // Returns data, not formatted output
}

// cmd/cli/formatter.go
package cli

import "myapp/pkg/lint"

type Formatter struct{}

func (f *Formatter) Format(diags []lint.Diagnostic) string {
    // Presentation logic here
}
```

## Anti-Pattern: Snapshot Storage in Core

### ❌ Wrong: Core package writes to filesystem

```go
// pkg/lint/snapshot.go
package lint

import "os"

type SnapshotStore struct {
    rootDir string
}

func (s *SnapshotStore) Save(snapshot Snapshot) error {
    return os.WriteFile(s.path(), data, 0600)  // ❌ I/O in core
}
```

### ✅ Correct: Move persistence to orchestration/CLI

```go
// pkg/lint/lint.go
package lint

// Core only defines data structures
type Snapshot struct {
    Timestamp time.Time
    Issues    []Diagnostic
}

// internal/orchestration/snapshot.go
package orchestration

import (
    "os"
    "myapp/pkg/lint"
)

type SnapshotStore struct {
    rootDir string
}

func (s *SnapshotStore) Save(snapshot lint.Snapshot) error {
    return os.WriteFile(s.path(), data, 0600)  // ✅ I/O in orchestration
}
```

## Detection Checklist

Run these checks on core/domain packages:

```bash
# Check for presentation imports
grep -r "fmt\\.Sprintf\\|fmt\\.Printf" pkg/*/

# Check for I/O imports
grep -r "\"os\"\\|\"io\"\\|\"net/http\"" pkg/*/

# Check for upward imports
go list -f '{{.ImportPath}}: {{join .Imports "\n"}}' ./pkg/... | grep cmd/

# Check for CLI/framework confusion
find pkg/ -name "*formatter*" -o -name "*renderer*" -o -name "*snapshot*"
```

## Refactoring Strategy

1. **Identify the violation** - core package imports presentation/I/O
2. **Extract interface** - define abstraction in core if needed
3. **Move implementation up** - move concrete I/O/formatting to outer layer
4. **Return structured data** - core returns data structures, not formatted output
5. **Inject dependencies** - pass I/O capabilities via constructor or method parameters

## When to Use Each Pattern

| Scenario | Pattern | Example |
|----------|---------|---------|
| Core needs to read files | Inject Reader interface | `Analyze(r Reader, path string)` |
| Core needs to format output | Return structured data | Return `Result`, format in CLI |
| Core needs to persist state | Move persistence to orchestration | `SnapshotStore` in `internal/` |
| Core needs logging | Inject logger interface | `New(logger Logger)` |
| Core needs HTTP client | Inject client interface | `Fetch(client HTTPClient, url string)` |

## Summary

- Core packages return **data structures**, not formatted strings
- Core packages accept **interfaces**, not concrete I/O types
- Presentation logic (formatters, renderers) lives in **cmd/** or **internal/app/**
- Persistence logic (file/DB operations) lives in **internal/orchestration/** or **cmd/**
- Dependencies flow **inward**: outer layers import inner layers, never the reverse

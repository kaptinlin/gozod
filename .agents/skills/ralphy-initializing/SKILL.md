---
description: Initializes Ralphy AI coding loop configuration. Creates .ralphy/config.yaml and updates .gitignore. Use when setting up Ralphy in a repository, initializing ralphy config, or when the user mentions ralphy init or setup.
name: ralphy-initializing
---


# Ralphy Initialization for Go Projects

Set up Ralphy configuration in a Go project. Follow these steps in order.

## Step 1: Detect Project Info

Before creating config, gather project context:

1. Read `go.mod` → module name, Go version
2. Read `Taskfile.yml` (if exists) → all available task names
3. Read `CLAUDE.md` or `README.md` → project description, existing conventions
4. Check for existing `.ralphy/` directory
5. Detect project structure:
   - `.references/` or `.reference/` directory
   - `SPECS/`, `DESIGNS/`, `designs/`, `DESIGN.md` directories/files
   - `.agents/skills/` or `.claude/skills/` directory
   - `research/` directory
   - `bin/` directory

## Step 2: Create `.ralphy/config.yaml`

```bash
mkdir -p .ralphy
```

Use this template, filling in detected values:

```yaml
# Ralphy Configuration
# https://github.com/michaelshimeles/ralphy

# Project info
project:
  name: "<module-name>"
  language: "Go"
  framework: "<framework-if-any>"
  description: "<one-line project description>"

# Commands (from Taskfile or defaults)
commands:
  test: "task test"
  lint: "task lint"
  build: "go build ./..."
  verify: "task verify"
  fmt: "task fmt"
  vet: "task vet"

# Rules - instructions the AI MUST follow
rules:
  # ── Go Coding Conventions ──
  - "Go <version> — use modern features: slices/maps packages, for range N, errors.AsType[T]()."
  - "Follow Google Go Best Practices and Style Decisions."
  - "Follow KISS, DRY, and YAGNI principles. No premature abstractions."
  - "Small interfaces — 1-3 methods per interface. Consumers depend only on what they use."
  - "Explicit error handling — return errors, wrap with fmt.Errorf(\"%w\"). No panic in production."
  - "All I/O functions take context.Context as first parameter."
  - "Exported methods check nil receiver."

  # ── Change Discipline ──
  - "Keep changes focused and minimal. Do not refactor unrelated code."
  - "One logical change per commit. Break large tasks into subtasks."
  - "Don't leave dead code. Delete unused code completely."

  # ── Design References ──
  - "Read SPECS/*.md for system contracts and data format specs before implementation."
  - "Read DESIGNS/*.md for architecture decisions before making structural changes."
  - "Consult research/ reports for domain knowledge and capability boundaries."
  - "Before writing code, read at least two relevant reference implementations from .references/ to understand design trade-offs."

  # ── Skill Usage ──
  - "When a task matches a skill in .agents/skills/, invoke that skill instead of handling it manually."
  - "Use dependency-selecting skill when choosing Go libraries."
  - "Use testing skill when writing tests."
  - "Use committing skill when creating commits."

  # ── Testing & Verification ──
  - "Write tests for all new features and bug fixes."
  - "All tests use t.Parallel(). Prefer table-driven tests."
  - "Use testify for assertions. Mock via function injection, not mock frameworks."
  - "Run 'task test' before committing."
  - "Run 'task lint' and fix all issues before committing."

  # ── Error Handling ──
  - "Sentinel errors: each package defines var Err* = errors.New(...)."
  - "Error wrapping: fmt.Errorf(\"%w: context\", err). Use errors.Is/As for matching."
  - "Input validation at function entry. Trust already-validated data internally."

  # ── Documentation ──
  - "Code comments and API documentation in English."
  - "Update CLAUDE.md when adding new patterns or conventions."

# Boundaries - files/folders the AI should not modify
boundaries:
  never_touch:
    - "go.sum"
    - ".references/**"
    - ".agents/skills/**"
    - ".ralphy/progress.txt"
    - "bin/**"
```

### Adaptation Rules

#### `project.name`
Extract from `go.mod` module path (last segment).

#### `project.framework`
Leave empty string if no framework. Fill in if detected (e.g., `chi`, `gin`, `echo`).

#### `commands`
- Match actual Taskfile tasks. Only include commands that exist.
- If no Taskfile, use Go defaults:
  ```yaml
  test: "go test ./..."
  lint: "go vet ./..."
  build: "go build ./..."
  ```
- Scan Taskfile for additional tasks (`verify`, `fmt`, `vet`, `generate`, etc.) and include all that exist.

#### `rules`
Rules are organized by category. Include or exclude per category based on detection:

| Category | Condition | Action |
|----------|-----------|--------|
| Go Coding Conventions | Always | Include. Replace `<version>` with Go version from `go.mod` |
| Change Discipline | Always | Include as-is |
| Design References | Per-item conditional | Include `.references/` rule only if `.references/` or `.reference/` exists. Include `SPECS/` rule only if `SPECS/` exists. Include `DESIGNS/` rule only if `DESIGNS/` or `designs/` or `DESIGN.md` exists. Include `research/` rule only if `research/` exists |
| Skill Usage | `.agents/skills/` or `.claude/skills/` exists | Include. Adjust path to match actual directory. Add skill-specific rules for skills that exist (dependency-selecting, testing, committing) |
| Testing & Verification | Always | Include. Adjust test/lint commands to match actual `commands` section |
| Error Handling | Always | Include as-is |
| Documentation | Always | Include as-is |

Add project-specific rules extracted from `CLAUDE.md` if present (e.g., library preferences, naming conventions).

#### `boundaries.never_touch`
- Always include `go.sum`
- Add each only if the directory exists: `.references/**`, `.agents/skills/**`, `bin/**`, `vendor/**`
- Always include `.ralphy/progress.txt`

## Step 3: Update `.gitignore`

Append these entries to `.gitignore` if not already present:

```
# Ralphy
.ralphy/*.json

# AI task/planning files
TODO.yml
TODO.yaml
TODO.md
PLAN.md
REFACTOR.md
```

Check each line before appending to avoid duplicates.

## Workflow Summary

1. Detect project info from `go.mod`, `Taskfile.yml`, `CLAUDE.md`, and project structure
2. Create `.ralphy/config.yaml` with detected values, adapting rules by category
3. Update `.gitignore` with ralphy and planning file entries

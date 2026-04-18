---
description: Initializes Ralphy AI coding loop configuration. Creates .ralphy/config.yaml and updates .gitignore. Use when setting up Ralphy in a repository, initializing ralphy config, or when the user mentions ralphy init or setup.
name: ralphy-initializing
---


# Ralphy Initialization for Go Projects

Set up Ralphy configuration in a Go project. Follow these steps in order.

## Step 1: Detect Project Info

Before creating config, gather project context:

1. Read `go.mod` → module name, Go version
2. Read `Taskfile.yml` or `Taskfile.yaml` (if exists) → all available task names
3. Read `CLAUDE.md` or `README.md` → project description, existing conventions
4. Read `LICENSE` → license type (MIT, agentable Commercial, Apache, etc.)
5. Check for existing `.ralphy/` directory (update if exists, create if not)
6. Detect project structure:
   - `.references/` or `.reference/` directory
   - `SPECS/`, `DESIGNS/`, `designs/`, `DESIGN.md` directories/files
   - `.agents/skills/` or `.claude/skills/` directory
   - `.research/` or `research/` directory
   - `DECISIONS.md`, `AGENTS.md` files
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
  # ── Craft Philosophy ──
  - "Simplicity as art — maximize work not done. Every line earns its place."
  - "Precision over cleverness — the right name, the right type, the right boundary."
  - "Elegance through reduction — if it doesn't spark clarity, remove it."
  - "Craft over shipping — the difference between a tool and an instrument is love."
  - "APIs as language — commands should feel like natural speech, not configuration."
  - "Errors as teachers — every failure message should illuminate the path forward."
  - "Beauty is structural — beautiful code is beautiful architecture made visible."
  - "Never: accidental complexity, feature gravity, abstraction theater, configurability cope."

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
  - "Use committing skill when creating commits."

  # ── Testing & Verification ──
  - "Write tests for all new features and bug fixes."
  - "All tests use t.Parallel(). Prefer table-driven tests."
  - "Run 'task test' before committing."
  - "Run 'task lint' and fix all issues before committing."

  # ── Error Handling ──
  - "Sentinel errors: each package defines var Err* = errors.New(...)."
  - "Error wrapping: fmt.Errorf(\"%w: context\", err). Use errors.Is/As for matching."
  - "Input validation at function entry. Trust already-validated data internally."

  # ── Documentation ──
  - "Code comments and API documentation in English."
  - "Update CLAUDE.md when adding new patterns or conventions."

  # ── Dependency Issue Reporting ──
  - "When a dependency library has a bug or limitation, do NOT work around it by reimplementing the dependency's functionality."
  - "Instead, create a report file at reports/<dependency-name>.md describing: dependency name and version, problem description, trigger scenario, expected vs actual behavior, error messages, and workaround suggestion (without implementing it)."
  - "Continue with other tasks that don't depend on the broken functionality."

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

#### `project.description`
Extract from `CLAUDE.md` or `README.md` first line/heading that describes the project. Do not use placeholders like "A Go library" — extract the actual description.

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
- Scan Taskfile for **all** defined tasks (`verify`, `fmt`, `vet`, `generate`, `deps`, `clean`, `deps:update`, etc.) and include every one that exists.

#### `rules`
Rules are organized by category. Include or exclude per category based on detection:

| Category | Condition | Action |
|----------|-----------|--------|
| Craft Philosophy | Always | Include as-is. These principles guide the spirit of all design decisions |
| Go Coding Conventions | Always | Include. Replace `<version>` with Go version from `go.mod` |
| Change Discipline | Always | Include as-is |
| Design References | Per-item conditional | See table below |
| Skill Usage | `.agents/skills/` or `.claude/skills/` exists | Include. Adjust path to match actual directory |
| Testing & Verification | Always | Include. Adjust test/lint commands to match actual `commands` section |
| Error Handling | Always | Include as-is |
| Documentation | Always | Include as-is |

**Design References — per-item rules:**

| Rule | Include if... |
|------|---------------|
| `.references/` rule | `.references/` or `.reference/` directory exists |
| `SPECS/` rule | `SPECS/` directory exists |
| `DESIGNS/` rule | `DESIGNS/` or `designs/` or `DESIGN.md` exists |
| `research/` rule | `.research/` or `research/` directory exists |
| `DECISIONS.md` rule | `DECISIONS.md` file exists. Add: "Read DECISIONS.md for design decisions before making structural changes." |
| `AGENTS.md` rule | `AGENTS.md` file exists. Add: "Read AGENTS.md for the full design specification and implementation plan." |

**Project-specific rules:**

After generating the standard rules, scan `CLAUDE.md` for project-specific conventions that should be encoded as rules:
- Zero-dependency constraints
- Specific testing patterns (e.g., "Use Clock interface — no time.Sleep")
- Named error conventions (e.g., "Six sentinel errors only: ...")
- Library/tool preferences

Add these as a `# ── Project-Specific ──` category at the end.

#### `boundaries.never_touch`
- Always include `go.sum`
- Always include `.ralphy/progress.txt`
- Add each only if the directory exists: `.references/**`, `.agents/skills/**`, `.claude/skills/**`, `.research/**`, `bin/**`, `vendor/**`, `reports/**`

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
.plans/*.md
```

Check each line individually before appending to avoid duplicates. Some projects may already have partial entries (e.g., `TODO.yaml` but not `TODO.yml`).

## Workflow Summary

1. Detect project info from `go.mod`, `Taskfile.yml`, `CLAUDE.md`, `LICENSE`, and project structure
2. Create `.ralphy/config.yaml` with detected values, adapting rules by category
3. Extract project-specific rules from `CLAUDE.md`
4. Update `.gitignore` with ralphy and planning file entries

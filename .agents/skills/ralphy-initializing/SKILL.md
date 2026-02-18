---
name: ralphy-initializing
description: Initializes Ralphy AI coding loop configuration for Go projects. Creates .ralphy/config.yaml and updates .gitignore. Use when setting up Ralphy in a Go repository, initializing ralphy config, or when the user mentions ralphy init or ralphy setup in a Go project.
---

# Ralphy Initialization for Go Projects

Set up Ralphy configuration in a Go project. Follow these steps in order.

## Step 1: Detect Project Info

Before creating config, gather project context:

1. Read `go.mod` for module name
2. Read `Makefile` (if exists) for available commands
3. Read `CLAUDE.md` or `README.md` for project description
4. Check for existing `.ralphy/` directory

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

# Commands (from Makefile or defaults)
commands:
  test: "task test"
  lint: "task lint"
  build: "go build ./..."

# Rules - instructions the AI MUST follow
rules:
  - "Before writing code, read at least two relevant reference implementations from .reference/ to understand design trade-offs."
  - "Read DESIGN.md or docs in designs/ for full architecture decisions before making structural changes."
  - "Follow KISS, DRY, and YAGNI principles."
  - "Follow Google Go Best Practices: https://google.github.io/go-style/best-practices"
  - "Follow Google Go Decisions: https://google.github.io/go-style/decisions"
  - "When a task matches a skill in .agents/skills/, invoke that skill instead of handling it manually."

# Boundaries - files/folders the AI should not modify
boundaries:
  never_touch:
    - "go.sum"
    - "vendor/**"
```

### Adaptation Rules

- **`project.name`**: Extract from `go.mod` module path (last segment)
- **`commands`**: Match actual Makefile targets. If no Makefile, use Go defaults (`go test ./...`, `go vet ./...`, `go build ./...`)
- **`rules`**:
  - Always include Google Go style guides and KISS/DRY/YAGNI
  - Include `.reference/` rule only if `.reference/` directory exists
  - Include `DESIGN.md` / `designs/` rule only if `DESIGN.md` or `designs/` directory exists
  - Include project-specific rules from `CLAUDE.md` if present
- **`boundaries.never_touch`**: Add `.reference/**`, `bin/**` if those directories exist

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

1. Detect project info from `go.mod`, `Makefile`, `CLAUDE.md`
2. Create `.ralphy/config.yaml` with detected values
3. Update `.gitignore` with ralphy and planning file entries

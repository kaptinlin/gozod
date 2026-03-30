---
description: Initialize a Go library as multi-module, or convert an existing single-module library to multi-module. Use when creating a multi-module Go repo, splitting a package into sub-modules, or setting up go.work.
name: multimodule-initializing
---


# Multi-Module Go Library Initialization

Set up or convert a Go library to multi-module structure with go.work, sub-module go.mod files, Taskfile, CI, and dependabot.

## When to Use

- Initializing a new multi-module Go library from scratch
- Converting an existing single-module Go library to multi-module
- Adding a new sub-module to an existing multi-module repo

## Pre-Flight Check: Should You Split?

**Default: keep everything in a single module.** Splitting into multiple modules adds maintenance overhead (coordinated releases, multiple go.mod files, dependabot entries, CI complexity). Only split when the dependency cost to consumers is clearly unjustified.

### Simple Decision Rule

Split a package into its own module ONLY if ALL three conditions are true:

1. **Heavy dependencies**: Package adds >5 MB of external dependencies (or >1 MB if moderate)
2. **Not core functionality**: Most library users don't need this package
3. **Clearly separable**: Package can function independently from other packages

**Examples:**
- ✅ Split: `postgres/` with pgx+pq (~8 MB) that only PostgreSQL users need
- ✅ Split: `cli/` with cobra+viper (~3 MB) that only CLI users need
- ❌ Keep in root: `validator/` with stdlib only (0 MB) even if optional
- ❌ Keep in root: `cache/` with small deps (~500 KB) that most users need

### Quick Dependency Check

Check dependency sizes before splitting:

```bash
# Check what dependencies a package pulls in
go list -m all

# Estimate size (rough approximation)
go mod download && du -sh $GOPATH/pkg/mod/cache/download
```

**Example analysis:**

| Package | Dependencies | Size | Core Users Need? | Decision |
|---------|-------------|------|------------------|----------|
| `core/` | stdlib only | 0 | Yes | ✅ Keep in root |
| `validator/` | stdlib only | 0 | Yes | ✅ Keep in root |
| `cache/` | small deps | ~500 KB | Yes | ✅ Keep in root |
| `postgres/` | pgx + pq | ~8 MB | No | ⚠️ **Split** |
| `cli/` | cobra + viper | ~3 MB | No | ⚠️ **Split** |

**Result:** Most packages stay in root, only 1-2 heavy packages become separate modules.

## What to Keep in Root

**Keep a package in root if ANY of these apply:**

- Uses only stdlib (no external dependencies)
- Has small external deps (<1 MB)
- Most users need it (core functionality)
- Internal package (`internal/`)

**Examples:** `core/`, `validator/`, `cache/`, `internal/` — all stay in root.

## Anti-Pattern: Over-Modularization

**❌ WRONG — Too many modules:**
```
github.com/org/repo/validator         ← separate module
github.com/org/repo/cache             ← separate module
github.com/org/repo/logger            ← separate module
```
Problems: version hell, no dependency savings (stdlib only), release friction.

**✅ CORRECT — Minimal modules:**
```
github.com/org/repo                   ← root: core + validator + cache + internal
github.com/org/repo/postgres          ← sub-module: pgx + pq (~8 MB)
github.com/org/repo/cli               ← sub-module: cobra + viper (~3 MB)
```

**Typical pattern: 1-2 sub-modules for heavy dependencies, everything else in root.**

## Consumer Impact

Show the before/after to justify the split:

| Use Case | Before | After | Savings |
|----------|--------|-------|---------|
| Library user (core only) | 15 MB | 500 KB | 30x smaller |
| PostgreSQL user | 15 MB | 8.5 MB | 1.8x smaller |
| CLI user | 15 MB | 3.5 MB | 4.3x smaller |

## Process

### Step 1: Audit Dependencies

Check which packages have heavy dependencies:

```bash
# List all dependencies
go list -m all

# Check what each package imports
go list -f '{{.ImportPath}}: {{join .Imports ", "}}' ./...
```

Build a simple table (see Quick Dependency Check above). Only proceed if you find packages with >5 MB dependencies that most users don't need.

### Step 2: Create Sub-Module go.mod Files

For each sub-module directory, create a `go.mod` file:

**Template:**
```go
module github.com/<org>/<repo>/<subdir>

go <version>

require github.com/<org>/<repo> v0.0.0

replace github.com/<org>/<repo> => <relative-path>
```

**Example 1 — PostgreSQL driver sub-module:**
```go
// postgres/go.mod
module github.com/org/repo/postgres

go 1.26

require (
    github.com/org/repo v0.0.0
    github.com/jackc/pgx/v5 v5.5.0
)

replace github.com/org/repo => ..
```

**Example 2 — CLI sub-module:**
```go
// cli/go.mod
module github.com/org/repo/cli

go 1.26

require (
    github.com/org/repo v0.0.0
    github.com/spf13/cobra v1.10.2
    github.com/spf13/viper v1.19.0
)

replace github.com/org/repo => ..
```

After creating each go.mod, run `go mod tidy` in that directory.

### Step 3: Trim Root go.mod

Remove dependencies now isolated in sub-modules:

**Before:**
```go
// go.mod (root)
require (
    github.com/jackc/pgx/v5 v5.5.0        // ← REMOVE
    github.com/spf13/cobra v1.10.2        // ← REMOVE
    github.com/spf13/viper v1.19.0        // ← REMOVE
    github.com/stretchr/testify v1.9.0    // KEEP
)
```

**After:**
```go
// go.mod (root)
require (
    github.com/stretchr/testify v1.9.0    // KEEP
)
```

Run `go mod tidy` in root after trimming.

### Step 4: Create go.work

Create a workspace file listing all modules:

```bash
go work init . ./postgres ./cli
```

Result (`go.work`):
```go
go 1.26.0

use (
    .
    ./postgres
    ./cli
)
```

### Step 5: Update .gitignore

Append if not already present:

```gitignore
# Go workspace (developer-local, not published)
go.work
go.work.sum
```

### Step 6: Update Taskfile

Use auto-discovered `MODULES` variable and multi-module tasks. See the `taskfile-configuring` skill for the complete multi-module Taskfile template.

**MODULES auto-discovery** — no manual list to maintain:

```yaml
vars:
  MODULES:
    sh: find . -mindepth 2 -name go.mod -not -path '*/vendor/*' -not -path '*/.*' -exec dirname {} \; | sed 's|^\./||' | sort
```

Excludes `vendor/`, `.references/`, `.git/`, and other hidden directories. Adding a new sub-module only requires creating its `go.mod` — no Taskfile edit needed.

**Required multi-module tasks:**

```yaml
tasks:
  test:
    desc: Run tests (root module)
    cmds:
      - go test -race ./...

  test:all:
    desc: Run tests for all modules
    cmds:
      - go test -race ./...
      - for: { var: MODULES }
        cmd: |
          echo "Running tests in {{.ITEM}}..."
          cd {{.ITEM}} && go test -race ./...

  lint:all:
    desc: Run linter for all modules
    cmds:
      - golangci-lint run
      - for: { var: MODULES }
        cmd: cd {{.ITEM}} && golangci-lint run

  tidy:all:
    desc: Run go mod tidy for all modules
    cmds:
      - go mod tidy
      - for: { var: MODULES }
        cmd: |
          echo "Tidying {{.ITEM}}..."
          cd {{.ITEM}} && go mod tidy

  deps:update:
    desc: Update all dependencies across root and all sub-modules
    cmds:
      - echo "Updating root module dependencies..."
      - GOWORK=off go get -u ./... && GOWORK=off go mod tidy
      - for: { var: MODULES }
        cmd: |
          if [ ! -f {{.ITEM}}/go.mod ]; then exit 0; fi
          echo "Updating dependencies in {{.ITEM}}..."
          DEPS=$(sed -n '/^require (/,/^)/p' {{.ITEM}}/go.mod | grep -v '// indirect\|require\|)' | awk '{print $1}' | while read -r dep; do grep -q "^replace $dep " {{.ITEM}}/go.mod || echo "$dep"; done | tr '\n' ' ')
          if [ -n "$DEPS" ]; then
            cd {{.ITEM}} && GOWORK=off go get -u $DEPS
          fi
      - echo "Syncing workspace..."
      - go work sync
```

**Why `deps:update` needs special handling:** Replace directives are module-local and not transitive. When `go get -u ./...` runs in a sub-module, it follows the replace to the root, but the root's replace directives for sibling modules don't apply — Go tries to fetch them from the proxy at `v0.0.0` and fails. The solution: update only direct external (non-replaced) dependencies in each sub-module, then sync the workspace.

### Step 7: Create Dependabot Config

Add one entry per module:

```yaml
# .github/dependabot.yml
version: 2
updates:
  - package-ecosystem: "gomod"
    directory: "/"
    schedule:
      interval: "weekly"
  - package-ecosystem: "gomod"
    directory: "/postgres"
    schedule:
      interval: "weekly"
  - package-ecosystem: "gomod"
    directory: "/cli"
    schedule:
      interval: "weekly"
```

### Step 8: Set Up CI

Use the `github-actions-configuring` skill with `private-package.yml` template. Key settings:

- Set `GOWORK: "off"` in env (validates published go.mod files)
- Run `task test` and `task lint` for root module only
- Don't use matrix per-module CI before v1.0.0 release

### Step 9: Document Decisions

Add a section in REFACTOR.md or ARCHITECTURE.md:

```markdown
## Multi-Module Structure

**Why split:**
- Library users: 30x smaller (15 MB → 500 KB)
- Isolated pgx+pq (~8 MB) to postgres/
- Isolated cobra+viper (~3 MB) to cli/

**Why packages stay in root:**
- core/, validator/, cache/ use stdlib only
- Splitting adds maintenance without dependency savings
```

### Step 10: Verify

Test the setup:

```bash
# Tidy all modules
task tidy:all

# Test root without workspace (simulates CI)
GOWORK=off task test

# Test all modules locally
task test:all
```

### Step 11: Set Up Release Automation

Multi-module repos need coordinated release tagging. The release scripts pin cross-module `v0.0.0` references to the release version before creating tags.

**Copy release scripts:**

```bash
mkdir -p scripts
cp /path/to/.agents/skills/releasing/scripts/tag_release.sh scripts/
cp /path/to/.agents/skills/releasing/scripts/check_changes.sh scripts/
chmod +x scripts/*.sh
```

**Add release tasks to Taskfile.yml:**

```yaml
tasks:
  release:check:
    desc: Check if a new release tag is needed
    cmds:
      - bash scripts/check_changes.sh

  release:tag:
    desc: Tag a new release (usage — task release:tag -- 0.2.0)
    cmds:
      - bash scripts/tag_release.sh {{.CLI_ARGS}}
```

**Add CI release workflow** (`.github/workflows/release.yml`):

Use the `release-library.yml` template from the `github-actions-configuring` skill. This automates the full release flow:

```bash
# Developer workflow — CI handles the rest:
git tag v0.1.0
git push origin v0.1.0
# CI: verify → pin go.mod → sub-module tags → push
```

**Key design decisions:**
- `go mod edit` only (no `go mod tidy`) — replace directives are not transitive
- grep pattern `"${mod_path} v"` matches both block and single-line require formats
- CI force-moves root tag to pinned commit, creates sub-module tags
- `GITHUB_TOKEN` for push prevents workflow re-trigger

## Adding a New Sub-Module

To add a new sub-module to an existing multi-module repo:

1. Create `<subdir>/go.mod` with replace directive (see Step 2)
2. Add to workspace: `go work use ./<subdir>`
3. Add entry in `.github/dependabot.yml`
4. Run `go mod tidy` in the new directory
5. Verify: `task test:all` (module is auto-discovered via MODULES variable)

## Common Mistakes

| Mistake | Fix |
|---------|-----|
| Splitting stdlib-only packages | Keep in root — no dependency savings |
| Module path doesn't match directory | `github.com/org/repo/postgres` must be in `postgres/` |
| Missing replace directive | Add `replace github.com/org/repo => ..` |
| Using real version instead of v0.0.0 | Always use `v0.0.0` for local modules |
| Forgetting go.work in .gitignore | Always gitignore for libraries |
| Running `go mod tidy` without workspace | Use go.work locally, `GOWORK=off` in CI |
| Too many modules | 1-2 sub-modules is typical; 5+ is too many |
| Running `go get -u ./...` in sub-modules | Replace directives aren't transitive — use selective external dep update (see `deps:update` task) |
| Hardcoding MODULES list in Taskfile | Use `find`-based auto-discovery so new modules are picked up automatically |
| Forgetting `go work sync` after updating deps | Always sync workspace after updating any module's deps |

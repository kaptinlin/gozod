---
description: Initialize a Go library as multi-module, or convert an existing single-module library to multi-module. Use when creating a multi-module Go repo, splitting a package into sub-modules, or setting up go.work.
name: multimodule-initializing
---


# Multi-Module Go Library Initialization

Set up or convert a Go library to multi-module structure with `go.work`, sub-module `go.mod` files, Taskfile, CI, and dependabot.

## Coupling Model First

Before choosing the exact setup, decide which of these two models you have:

### Model A — Strongly Coupled Monorepo Modules (Default)

Use this when sub-modules are developed, tested, and released together.

Typical signs:

- sub-modules are framework adapters, provider integrations, or companion tooling for the same core library
- cross-module changes often happen in the same PR
- versioning is coordinated across modules
- the modules are not intended to behave like a public extension marketplace

**Best practice for this model:**

- keep local `replace` directives between sibling modules
- also use `go.work` for local developer convenience
- require `GOWORK=off` verification in CI and release validation

This is the common case for internal or tightly coupled Go monorepos.

### Model B — Independently Consumable Extension Modules

Use this only when sub-modules are meant to behave like independently consumable extensions.

Typical signs:

- modules should be usable and versioned more independently
- local development should not be encoded into published `go.mod` files
- users consume extension modules like a separate product surface

**Best practice for this model:**

- prefer `go.work` for local development
- avoid local `replace` directives by default
- publish and tag each module as a true independently consumable module

**Default assumption:** Most multi-module Go libraries are **Model A**, not Model B.

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

### Step 2: Create Sub-Module `go.mod` Files

For each sub-module directory, create a `go.mod` file.

Choose the template based on the coupling model.

#### Model A Template — Strongly Coupled Modules (Default)

For strongly coupled modules, keep the local `replace`. This makes each sub-module directly testable and developable even without a workspace, while `GOWORK=off` still verifies published-module behavior.

```go
module github.com/<org>/<repo>/<subdir>

go <version>

require github.com/<org>/<repo> v0.0.0

replace github.com/<org>/<repo> => <relative-path>
```

#### Model B Template — Independently Consumable Extensions

For independently consumable modules, prefer a publishable `go.mod` without local `replace`, and rely on `go.work` for local multi-module development.

```go
module github.com/<org>/<repo>/<subdir>

go <version>

require github.com/<org>/<repo> v0.0.0
```

**Model A Example — CLI sub-module:**
```go
// cli/go.mod
module github.com/org/repo/cli

go 1.26

require (
    github.com/org/repo v0.0.0
    github.com/spf13/cobra v1.10.2
)

replace github.com/org/repo => ..
```

**Model B Example — independently consumable extension:**
```go
// echoext/go.mod
module github.com/org/repo/echoext

go 1.26

require (
    github.com/org/repo v0.0.0
    github.com/labstack/echo/v4 v4.13.4
)
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

### Step 4: Create `go.work`

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

For existing repositories with several modules, prefer auto-discovery:

```bash
go work init
go work use -r .
```

**Workspace rules:**

- `go.work` is for local multi-module development, not for publishing module relationships
- the `go` version in `go.work` must be greater than or equal to every module listed in `use`
- in Model A, `go.work` complements local `replace`
- in Model B, `go.work` usually replaces local `replace`

### Step 5: Update `.gitignore`

Append if not already present:

```gitignore
# Go workspace (developer-local, not published)
go.work
go.work.sum
```

**Best practice:** for libraries, do not commit `go.work` by default. A checked-in workspace can override a developer's parent workspace and can cause CI to test the wrong dependency graph. Treat `go.work.sum` the same way: ignore it and keep it developer-local.

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

**Why `deps:update` needs special handling:** Multi-module repos must keep every module independently resolvable. Run dependency updates per module, keep local development wiring explicit, and verify each module with `GOWORK=off`.

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

- Set `GOWORK: "off"` in env (validates published `go.mod` files instead of local workspace wiring)
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

# Test root without workspace (simulates CI and consumer resolution)
GOWORK=off task test

# Test all modules locally
task test:all

# Test each published module independently
for mod in . ./postgres ./cli; do
  (cd "$mod" && GOWORK=off go test ./...)
done
```

**Verification rule by model:**

- Model A: local `replace` is allowed, but every module must still pass `GOWORK=off go test ./...`
- Model B: avoid local `replace`, and every module must pass `GOWORK=off go test ./...`

### Step 11: Set Up Release Automation

Multi-module repos need coordinated release tagging. The release scripts pin cross-module `v0.0.0` references to the release version before creating tags.

**Tagging rule:** sub-directory modules must be tagged with the directory prefix:

```text
v0.3.0              # root module
postgres/v0.3.0     # postgres sub-module
cli/v0.3.0          # cli sub-module
```

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
- `go mod edit` only (no `go mod tidy`) — release pinning should update module requirements without relying on local workspace state
- grep pattern `"${mod_path} v"` matches both block and single-line require formats
- CI force-moves root tag to pinned commit, creates sub-module tags
- `GITHUB_TOKEN` for push prevents workflow re-trigger

## Adding a New Sub-Module

To add a new sub-module to an existing multi-module repo:

1. Decide whether the new module is Model A or Model B
2. Create `<subdir>/go.mod` using the matching template (see Step 2)
2. Add to workspace: `go work use ./<subdir>`
3. Add entry in `.github/dependabot.yml`
4. Run `go mod tidy` in the new directory
5. Verify locally: `task test:all`
6. Verify publishability: `cd <subdir> && GOWORK=off go test ./...`
7. When releasing, tag with `<subdir>/vX.Y.Z`

## Common Mistakes

| Mistake | Fix |
|---------|-----|
| Splitting stdlib-only packages | Keep in root — no dependency savings |
| Module path doesn't match directory | `github.com/org/repo/postgres` must be in `postgres/` |
| Treating all sub-modules like independent extensions | Most repos are Model A — use local `replace` when modules are strongly coupled |
| Removing local `replace` from strongly coupled modules too early | Keep `replace` for Model A; rely on `GOWORK=off` to validate publishability |
| Using local `replace` in a module meant for independent consumption | Prefer Model B — use `go.work` for local development and keep `go.mod` publishable |
| Using real version instead of v0.0.0 | Always use `v0.0.0` for local modules |
| Forgetting `go.work` and `go.work.sum` in `.gitignore` | Always ignore both for libraries unless you have a deliberate repo-wide reason to commit them |
| Only testing with `go.work` enabled | Also run `GOWORK=off go test ./...` in each module |
| Too many modules | 1-2 sub-modules is typical; 5+ is too many |
| Tagging sub-modules with root-style tags | Use `<subdir>/vX.Y.Z` for sub-directory modules |
| Hardcoding MODULES list in Taskfile | Use `find`-based auto-discovery so new modules are picked up automatically |
| Committing `go.work` for a library by default | Keep it developer-local unless you have a specific repo-level reason to commit it |
| Forgetting `go.work` version constraints | Keep `go.work` `go` version >= every module's `go` version |

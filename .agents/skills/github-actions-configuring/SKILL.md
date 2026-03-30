---
description: Configure GitHub Actions CI/CD workflows for Go packages in monorepo and single-module layouts. Use when setting up CI/CD pipelines, fixing submodule authentication issues, configuring workflows for open-source vs private packages, or managing PAT_TOKEN for private Go module dependencies.
name: github-actions-configuring
---


# GitHub Actions for Go Packages

Configure GitHub Actions CI/CD workflows for Go packages with proper handling of submodules and private dependencies.

## Quick Decision Tree

**Step 1: Detect project type**

```bash
MODULE_COUNT=$(find . -name go.mod -not -path '*/vendor/*' -not -path '*/.*' | wc -l)
# 1 = single-module, >1 = multi-module
```

**Step 2: Select template**

```
Single-module (1 go.mod)
  ├─ Has private deps (github.com/agentable/*)?
  │   └─ Yes → assets/private-package.yml
  │   └─ No  → Has test submodules (.gitmodules with tests/)?
  │       ├─ Yes → assets/opensource-with-tests.yml
  │       └─ No  → assets/opensource-simple.yml
  │
Multi-module (>1 go.mod, has go.work)
  └─ CI:      assets/private-package.yml with GOWORK: "off"
  └─ Release: assets/release-library.yml (tag-triggered, pins + sub-tags)
      └─ Pre-release (v0.0.0 sub-modules): task test + task lint (root only)
      └─ Post-release (published versions):  task test:all + task lint:all (optional)

Single-module CLI/binary
  └─ Release: GoReleaser release.yml (goreleaser-action)

Single-module library
  └─ Release: Optional verify-on-tag workflow (no pinning needed)
```

| Project Type | CI Template | Release Template | GOWORK | Test Command | Lint Command |
|-------------|------------|-----------------|--------|-------------|-------------|
| Single-module, open-source | `opensource-simple.yml` | — | (default) | `task test` | `task lint` |
| Single-module, with test submodules | `opensource-with-tests.yml` | — | (default) | `task test` | `task lint` |
| Single-module, private deps | `private-package.yml` | — | (default) | `task test` | `task lint` |
| Multi-module, pre-release | `private-package.yml` | `release-library.yml` | `"off"` | `task test` | `task lint` |
| Multi-module, post-release | `private-package.yml` | `release-library.yml` | `"off"` | `task test` or `task test:all` | `task lint` or `task lint:all` |
| Single-module CLI/binary | `private-package.yml` | GoReleaser `release.yml` | (default) | `task test` | `task lint` |

## Core Principles

### Submodule Strategy

**NEVER initialize `.agents/skills` submodule in CI/CD workflows.** This private submodule causes authentication failures in open-source packages and is unnecessary for CI/CD.

**Key rules:**
- Open-source packages: No submodule initialization needed
- Packages with test submodules: Initialize ONLY specific test submodules
- Private packages: Use PAT_TOKEN for Go modules, NOT for submodules

### Workflow Structure

Every Go package workflow should have two jobs:

1. **test** — Run `task test` with race detection
2. **lint** — Run `task lint` with golangci-lint

Both jobs should:
- Use `actions/checkout@v6` without `submodules: recursive`
- Use `actions/setup-go@v6` with `go-version-file: go.mod`
- Enable Go module caching with `cache: true`
- Use `go-task/setup-task@v1` to install Task runner

## Multi-Module Repositories

### The Core Problem: Replace Directives Are Not Transitive

In multi-module repos, sub-modules use `replace` directives to point to the root. But the root's replace directives for sibling modules don't apply to sub-modules:

```
Root go.mod:
  require provider/file v0.0.0
  replace provider/file => ./provider/file     ← only applies to root

Sub-module format/json/go.mod:
  require go-config v0.0.0
  replace go-config => ../..                   ← resolves root
  (no replace for provider/file)               ← root's replace NOT inherited
```

When CI runs `go mod tidy` or `go get -u ./...` in a sub-module with `GOWORK=off`:
1. Follows replace to root -> reads root go.mod
2. Root requires `provider/file v0.0.0` -> root's replace doesn't apply
3. Tries to fetch `provider/file@v0.0.0` from proxy -> **fails** (unpublished)

### Why Matrix CI Fails Pre-Release

Do **NOT** use dynamic module detection + matrix CI (one job per module) for multi-module repos before the first release:

1. Sub-modules depend on root via `replace` with `v0.0.0` placeholders
2. `go mod tidy` with `GOWORK=off` traverses the full transitive graph, including test deps of replaced modules
3. Sibling modules at `v0.0.0` aren't published — every per-module job fails
4. Result: N modules x jobs = dozens of guaranteed failures

### Correct Approach: Pre-Release

Test root module only via Taskfile. The root module's `replace` directives resolve all local sub-module deps correctly with `GOWORK=off`. Tidy check is included in `task lint` (via `tidy-lint` dependency).

For multi-module repos, use the `private-package.yml` template with `GOWORK: "off"` in the env block. The template already includes this setting.

Key differences from single-module CI:

| Aspect | Single-Module | Multi-Module (pre-release) |
|--------|--------------|---------------------------|
| `GOWORK` | Not set (default) | `"off"` (validate published go.mod) |
| Test command | `task test` | `task test` (root only) |
| Lint command | `task lint` | `task lint` (root only) |
| Matrix per module | N/A | **No** — sibling modules unpublished |
| Tidy check | Via `tidy-lint` | Via `tidy-lint` (root only) |

### Correct Approach: Post-Release

After the first release (all modules tagged with real versions), per-module CI becomes viable:

- **`task test:all`** — tests root + all sub-modules (each resolves published versions)
- **`task lint:all`** — lints root + all sub-modules
- Matrix per module is possible but root-only approach remains simpler

Recommended: stay with root-only CI unless sub-modules have divergent test requirements (e.g., different build tags, external services).

### Dependency Update in CI

For multi-module repos, validating dependency updates in CI requires the three-phase approach from the `taskfile-configuring` skill. The `deps:update` Taskfile task handles this correctly:
1. Root module: `GOWORK=off go get -u ./...` (has all replace directives)
2. Sub-modules: only update external (non-replaced) deps
3. `go work sync` to propagate version changes

## Template Selection Guide

### Template 1: Open-Source Simple

**Use for:** Single-module open-source packages with no test submodules and no private Go dependencies

**Characteristics:**
- No submodule initialization
- No PAT_TOKEN required
- Standard test and lint jobs
- Fast checkout and execution

**Examples:** deepclone, defuddle-go, emitter, filter, go-i18n, gozod, jsonmerge, jsonpatch, jsonpointer, jsonrepair, orderedobject, queue, requests, template

**Template:** See `assets/opensource-simple.yml`

### Template 2: Open-Source with Test Submodules

**Use for:** Single-module open-source packages that require specific public submodules for testing

**Characteristics:**
- Selective submodule initialization (test submodules only)
- No PAT_TOKEN required
- Custom submodule init step before tests

**Examples:**
- jsonschema → needs `tests/testdata/JSON-Schema-Test-Suite`
- messageformat-go → needs `tests/messageformat-wg` and `.reference/messageformat`

**Template:** See `assets/opensource-with-tests.yml`

**Customization required:** Update the submodule paths in the "Initialize test submodules" step.

### Template 3: Private/Internal Packages

**Use for:** Packages with private Go module dependencies from github.com/agentable/*, including multi-module repos

**Characteristics:**
- No submodule initialization
- Requires PAT_TOKEN secret for Go module access
- GOPRIVATE environment variables
- GOWORK: "off" for consistent module resolution
- Git URL rewriting for authentication

**Examples:** agentstack, aster, bashrepair, condeval, filterschema, gendog, go-config, go-fsm, go-sandbox, go-secrets, godocx, jsoncrdt, jsondiff, knora, mathconv, openapi-generator, openapi-request, pdfkit, polytrans, queryparse, queryschema, unifai, unifmsg, vfs

**Template:** See `assets/private-package.yml`

**Prerequisites:** Repository must have `PAT_TOKEN` secret configured in GitHub Settings → Secrets and variables → Actions.

## Release Workflow for Multi-Module Libraries

Multi-module repos need a release workflow that pins cross-module dependencies and creates sub-module tags. The developer pushes a single root tag, and CI handles everything else.

**Template:** `assets/release-library.yml`

**Developer workflow:**

```bash
git tag v0.1.2
git push origin v0.1.2
# CI: verify → pin go.mod → sub-module tags → push
```

**What CI does:**
1. Runs `task verify` (test + lint)
2. Pins all cross-module `v0.0.0` refs to `v0.1.2` via `go mod edit` (NO `go mod tidy`)
3. Commits the pinned go.mod files
4. Force-moves root tag to the pinned commit
5. Creates sub-module tags (`provider/file/v0.1.2`, etc.)
6. Pushes using `GITHUB_TOKEN` (prevents re-trigger)

**Critical design decisions:**
- **No `go mod tidy`**: Replace directives are NOT transitive. Sub-module tidy fails because sibling tags don't exist on the proxy yet. `go mod edit -require` is sufficient.
- **grep pattern `"${mod_path} v"`**: Matches both block-style (TAB-indented) and single-line require formats in go.mod.
- **`GITHUB_TOKEN` for push**: Pushes via `GITHUB_TOKEN` don't re-trigger workflows. Sub-module tags (`provider/file/v0.1.2`) don't match `v*` pattern either.
- **Pinned commit not on main**: The pinned commit is reachable via tags but not on any branch. Main stays clean with `replace` directives for development.

**Setup:**
1. Copy `assets/release-library.yml` to `.github/workflows/release.yml`
2. Copy `scripts/tag_release.sh` and `scripts/check_changes.sh` from the `releasing` skill
3. Add `release:check` and `release:tag` tasks to Taskfile (see `taskfile-configuring` skill)

## Implementation Workflow

### Step 1: Identify Project Type

Check the project structure:

```bash
# Detect project type
MODULE_COUNT=$(find . -name go.mod -not -path '*/vendor/*' -not -path '*/.*' | wc -l)
echo "Modules: $MODULE_COUNT"  # 1 = single-module, >1 = multi-module

# Check for private dependencies
grep "github.com/agentable" go.mod

# Check for test submodules
grep "tests/" .gitmodules 2>/dev/null

# Check for go.work (multi-module indicator)
test -f go.work && echo "Multi-module (has go.work)"
```

### Step 2: Select Template

Based on Step 1 results:
- Multi-module (>1 go.mod) → `private-package.yml` with `GOWORK: "off"`
- Single-module, has private deps → `private-package.yml`
- Single-module, no private deps, has test submodules → `opensource-with-tests.yml`
- Single-module, no private deps, no test submodules → `opensource-simple.yml`

### Step 3: Create Workflow File

```bash
mkdir -p .github/workflows
cp /path/to/template .github/workflows/ci.yml
```

### Step 4: Customize (if needed)

For packages with test submodules, update the submodule paths:

```yaml
- name: Initialize test submodules
  run: |
    git submodule update --init tests/testdata/YOUR-SUBMODULE
    git submodule update --init .reference/YOUR-REFERENCE
```

### Step 5: Verify and Test

```bash
# Commit and push
git add .github/workflows/ci.yml
git commit -m "ci: add GitHub Actions workflow"
git push

# Monitor the workflow run in GitHub Actions tab
```

## Common Issues and Solutions

### Issue 1: "fatal: could not read Username for 'https://github.com'"

**Cause:** Workflow is trying to initialize `.agents/skills` submodule (private repo)

**Solution:** Remove `submodules: recursive` from checkout step and any `git submodule update --init --recursive` commands

### Issue 2: "go: github.com/agentable/package@version: invalid version: git ls-remote failed"

**Cause:** Private Go module dependencies require authentication

**Solution:** Use the private-package.yml template with PAT_TOKEN configuration

### Issue 3: Tests fail with "submodule not found"

**Cause:** Test submodules are not initialized

**Solution:** Use opensource-with-tests.yml template and specify the exact test submodule paths

### Issue 4: "missing go.sum entry" or "unknown revision v0.0.0" in multi-module CI

**Cause:** Matrix CI runs `go mod tidy` or `go test` per sub-module with `GOWORK=off`. Sub-modules reference sibling modules at `v0.0.0` which aren't published.

**Solution:** Do not use matrix per-module CI pre-release. Use the root-only approach: `task test` + `task lint` from the repository root. See "Multi-Module Repositories" section.

### Issue 5: Workflow is slow

**Cause:** Unnecessary submodule cloning or missing Go module cache

**Solution:**
- Remove all submodule initialization except required test submodules
- Ensure `cache: true` is set in `actions/setup-go@v6`

## Reference Documentation

For detailed workflow patterns and examples, see:
- `references/workflow-patterns.md` — Complete workflow examples and variations
- `references/troubleshooting.md` — Detailed troubleshooting guide

## Validation Checklist

After creating the workflow, verify:

**All project types:**
- [ ] Workflow file is at `.github/workflows/ci.yml`
- [ ] No `submodules: recursive` in checkout step
- [ ] No `git submodule update --init --recursive` commands
- [ ] `go-version-file: go.mod` is set
- [ ] `cache: true` is enabled
- [ ] Both test and lint jobs are present
- [ ] Workflow triggers on push to main and pull requests

**Single-module specific:**
- [ ] Test submodules (if any) are initialized with specific paths
- [ ] PAT_TOKEN is configured (for private packages only)

**Multi-module specific:**
- [ ] `GOWORK: "off"` is set in env
- [ ] PAT_TOKEN is configured (always required — multi-module repos typically have private deps)
- [ ] Uses `task test` (root only), NOT `task test:all` (pre-release)
- [ ] Uses `task lint` (root only), NOT `task lint:all` (pre-release)
- [ ] No matrix per module (pre-release)
- [ ] Taskfile has `deps:update` task with three-phase strategy (see `taskfile-configuring` skill)

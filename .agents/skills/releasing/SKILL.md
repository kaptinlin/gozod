---
description: Release workflow for a single Go repository following semantic versioning. Use when releasing a Go package, creating a new version tag, or preparing to publish a Go module.
name: releasing
---


# Go Package Release Guide

Sequential release process for a single Go repository. Follow steps in order.

## Prerequisites

- Go 1.26+ installed
- golangci-lint v2.9.0+ (auto-installed via `task`)
- `gh` CLI (optional, for GitHub release notes)

## Scripts

| Script | Purpose |
|--------|---------|
| `scripts/check_changes.sh` | Detect `.go`/`go.mod`/`go.sum` changes since latest tag (exit 0 = tag needed) |
| `scripts/tag_release.sh <VERSION>` | Pin sub-module deps, create root + sub-module tags, push |

## Release Workflow

Copy this checklist and track progress:

```
Release Progress:
- [ ] Step 1: Determine version
- [ ] Step 2: Fix all issues and pass checks
- [ ] Step 3: Update documentation
- [ ] Step 4: Stage and commit
- [ ] Step 5: Tag and push
- [ ] Step 6: Verify release
```

### Step 1: Determine Version

Follow [Semantic Versioning](https://semver.org/):

| Change Type | Version Bump | Example |
|---|---|---|
| Breaking API changes (removed/renamed exports, changed signatures) | **MAJOR** `vX.0.0` | `v1.0.0` -> `v2.0.0` |
| New features, new exports (backward compatible) | **MINOR** `vX.Y.0` | `v1.2.0` -> `v1.3.0` |
| Bug fixes, performance improvements, internal refactors | **PATCH** `vX.Y.Z` | `v1.2.3` -> `v1.2.4` |

**Rules:**
- `v0.x.x` packages: breaking changes only bump MINOR (`v0.1.0` -> `v0.2.0`)
- `v2+` modules **must** have `/v2` suffix in `go.mod` module path
- First release: `v0.1.0`

Check existing tags:

```bash
git tag --list 'v*' --sort=-v:refname | head -5
```

### Step 2: Fix All Issues and Pass Checks

```bash
go fmt .
task lint
task test
```

Then run package-wide formatting if needed:

```bash
go fmt ./...
```

**Hard gate before commit:**
- Fix all compile, lint, and test issues first.
- `go fmt .`, `task lint`, and `task test` must all pass.
- Do not proceed to staging/commit while any check is failing.

### Step 3: Update Documentation

**README.md** - Update if applicable:
- Version badge / installation instructions
- New features or API changes in usage examples
- Changelog / release notes section
- Compatibility matrix (Go version, dependency versions)

**CLAUDE.md** - Update if there are:
- New public types, functions, or architectural changes
- New development commands or build targets
- Changes to testing or CI requirements

### Step 4: Stage and Commit

Use conventional commit format directly in this skill:

- Format: `<type>(<scope>): <description>`
- `type`: `feat`, `fix`, `refactor`, `perf`, `docs`, `test`, `build`, `ci`, `chore`
- `<scope>` must describe changed area (for example `api`, `json`, `lint`, `release`)
- `<scope>` is **not** the project or repository name
- Commit message must not mention `claude` (any casing)
- Commit message must not include collaborator/assistant attribution lines (for example `Co-authored-by`)

Update git submodules to latest before staging:

```bash
git submodule update --remote --merge
```

```bash
git add .
git reset TODO.md PLAN.md IMPROVE.md REFACTOR.md 2>/dev/null || true
git commit -m "<type>(<scope>): <description>"
```

**Staging rules:**
- Stage all release-related changes: Go source files, `go.mod`, `go.sum`, git submodule updates, documentation.
- Exclude temporary development files: `TODO.md`, `PLAN.md`, `IMPROVE.md`, `REFACTOR.md`, scratch notes, local plan files.
- Only run `git add .` after Step 2 checks are fully green.

**Before tagging**, check recent commits for package-name scopes and rewrite if needed:

```bash
git log --oneline -10
```

### Step 5: Tag and Push

Run `scripts/check_changes.sh` to determine whether a new tag is needed:

```bash
bash scripts/check_changes.sh
```

The script compares against the latest `v*` tag and checks:
- `.go` file changes
- Root `go.mod` dependency changes (ignoring `go`/`toolchain`/`module` directives)
- Sub-module `go.mod`/`go.sum` changes (multi-module repos)

**Exit code:** `0` = tag needed, `1` = no tag needed.

**Tag and push** (when tag needed):

```bash
bash scripts/tag_release.sh <VERSION>
```

The script handles:
1. **Pin ALL cross-module dependencies** — updates go.mod in all directions (sub→root, root→sub, sub→sub) from development versions to `v<VERSION>` using `go mod edit` (no `go mod tidy` — replace directives are not transitive), commits
2. **Root tag** — `v<VERSION>`
3. **Sub-module tags** — `<subdir>/v<VERSION>` (same version as root)
4. **Push** — `git push origin <branch> --tags`

**CI Automation:** For repos with `.github/workflows/release.yml`, the developer only needs to push a root tag. CI handles pinning, sub-module tags, and push automatically. See the `github-actions-configuring` skill for the release workflow template.

**Push only** (when no tag needed):

```bash
git push origin main
```

For `v2+` module paths:

```bash
git tag -a v2.1.0 -m "v2.1.0"
git push origin v2.1.0
```

### Step 6: Verify Release

```bash
GOPROXY=direct go list -m <module-path>@v<VERSION>
```

Example:

```bash
GOPROXY=direct go list -m github.com/kaptinlin/requests@v1.3.0
```

## Quick Reference

Release checklist:

```bash
cd <package>
task deps:update                                   # upgrade deps (skip replaced)
go fmt .                                           # format (gate 1)
go fmt ./...                                       # format package-wide
task lint                                          # lint
task test                                          # test
# ... fix all issues, update README.md / CLAUDE.md if needed ...
git submodule update --remote --merge              # update submodules to latest
git add .
git reset TODO.md PLAN.md IMPROVE.md REFACTOR.md 2>/dev/null || true
git commit -m "<type>(<scope>): <description>"
# Check and tag
if bash scripts/check_changes.sh; then
  bash scripts/tag_release.sh <VERSION>
else
  git push origin main
fi
```

---
name: releasing
description: Guides the release process for individual Go packages following semantic versioning. Use when releasing a Go package, creating a new version tag, upgrading dependencies before release, or when the user asks to prepare a release, tag a version, or publish a Go module. Triggers on release, version bump, tagging, or publishing Go packages.
---

# Go Package Release Guide

Sequential release process for individual Go packages. Follow steps in order.

## Prerequisites

- Go 1.26+ installed
- golangci-lint v2.9.0 (auto-installed via `make`)
- `gh` CLI (optional, for GitHub release notes)

## Release Workflow

Copy this checklist and track progress:

```
Release Progress:
- [ ] Step 1: Determine version
- [ ] Step 2: Check dependency order
- [ ] Step 3: Upgrade dependencies
- [ ] Step 4: Fix all issues and pass checks
- [ ] Step 5: Update documentation
- [ ] Step 6: Stage and commit
- [ ] Step 7: Tag and push
- [ ] Step 8: Verify release
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

### Step 2: Check Dependency Order

If releasing multiple packages, check dependency levels in `packages.md`:

```
Level 0 -> Level 1 -> Level 2 -> Level 3 -> Level 4 -> Level 5
```

Release lower-level dependencies **first**, then update and release higher-level packages.

### Step 3: Upgrade Dependencies

```bash
cd <package-dir>
go get -u ./...
go mod tidy
```

Review changes:

```bash
git diff go.mod go.sum
```

If a dependency upgrade introduces breaking changes, fix the code before proceeding.

### Step 4: Fix All Issues and Pass Checks

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

### Step 5: Update Documentation

**README.md** - Update if applicable:
- Version badge / installation instructions
- New features or API changes in usage examples
- Changelog / release notes section
- Compatibility matrix (Go version, dependency versions)

**CLAUDE.md** - Update if there are:
- New public types, functions, or architectural changes
- New development commands or build targets
- Changes to testing or CI requirements

### Step 6: Stage and Commit

Use conventional commit format directly in this skill:

- Format: `<type>(<scope>): <description>`
- `type`: `feat`, `fix`, `refactor`, `perf`, `docs`, `test`, `build`, `ci`, `chore`
- `<scope>` must describe changed area (for example `api`, `json`, `lint`, `release`)
- `<scope>` is **not** the project or repository name
- Commit message must not mention `claude` (any casing)
- Commit message must not include collaborator/assistant attribution lines (for example `Co-authored-by`)

```bash
git add .
git reset TODO.md PLAN.md 2>/dev/null || true
git commit -m "<type>(<scope>): <description>"
```

**Staging rules:**
- Stage all release-related changes.
- Exclude temporary development files (for example: `TODO.md`, `PLAN.md`, scratch notes, local plan files).
- Only run `git add .` after Step 4 checks are fully green.

**Before tagging**, check recent commits for package-name scopes and rewrite if needed:

```bash
git log --oneline -10
```

### Step 7: Tag and Push

```bash
git tag -a v<VERSION> -m "v<VERSION>"
git push origin main
git push origin v<VERSION>
```

For `v2+` module paths:

```bash
git tag -a v2.1.0 -m "v2.1.0"
git push origin v2.1.0
```

### Step 8: Verify Release

```bash
GOPROXY=direct go list -m <module-path>@v<VERSION>
```

Example:

```bash
GOPROXY=direct go list -m github.com/kaptinlin/requests@v1.3.0
```

## Quick Reference

Single-package release checklist:

```bash
cd <package>
git tag --list 'v*' --sort=-v:refname | head -5  # check current version
go get -u ./...                                    # upgrade deps
go mod tidy                                        # tidy modules
go fmt .                                           # format (gate 1)
go fmt ./...                                       # format package-wide
task lint                                          # lint
task test                                          # test
# ... fix all issues, update README.md / CLAUDE.md if needed ...
git add .
git reset TODO.md PLAN.md 2>/dev/null || true      # exclude temp dev files
git commit -m "<type>(<scope>): <description>"
git tag -a v<VERSION> -m "v<VERSION>"
git push origin main
git push origin v<VERSION>
```

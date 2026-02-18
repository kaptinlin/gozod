---
name: committing
description: Creates conventional commits for Go packages following project commit conventions. Use when committing Go code changes, creating commit messages, or when the user asks to commit staged changes in a Go repository. Triggers on commit, git commit, or staged changes in Go projects.
---

# Go Commit Guide

Create conventional commits for Go packages. Follow these rules strictly.

## Commit Message Format

```
<type>(<scope>): <description>

[optional body]
```

### Types

`feat`, `fix`, `perf`, `refactor`, `docs`, `chore`, `test`, `build`, `ci`

### Scope Rules

- Scope is **optional** and describes the **nature of the change** (e.g., `deps`, `config`, `docs`, `lint`, `test`)
- **NEVER use the package name as scope** (e.g., never `feat(jsonschema):` or `chore(deepclone):`)
- Breaking changes: append `!` after type (e.g., `feat!:`) or add `BREAKING CHANGE:` in body

### Description Rules

- Imperative mood, lowercase, no trailing period
- Body explains **why**, not **what** (optional)
- **No `Co-Authored-By` trailer**

## Examples

```bash
# New feature
git commit -m "feat: add streaming retry support"

# Bug fix
git commit -m "fix: handle nullable anyOf validation"

# Dependency upgrade
git commit -m "chore(deps): upgrade dependencies and fix lint warnings"

# Breaking change
git commit -m "feat!: redesign validation pipeline API"

# With body
git commit -m "fix: handle nullable anyOf validation

anyOf with null type was not generating pointer fields correctly"
```

## Pre-Commit Checks

Before committing, run verification:

```bash
go fmt ./...
task lint
task test
```

Or all at once:

```bash
task verify
```

Fix all issues before committing. Do not skip lint or test failures.

## Workflow

1. Stage changes: `git add .` (or specific files)
2. Review staged diff: `git diff --cached`
3. Determine the appropriate type and optional scope
4. Write commit message following the format above
5. Commit

## Clean Up Commit Messages

Before tagging a release, check recent commits for package-name scopes:

```bash
git log --oneline -10
```

If any commit uses the package name as scope (e.g., `chore(deepclone):`, `feat(jsonschema):`), rewrite with interactive rebase:

```bash
git rebase -i HEAD~N
# Mark offending commits as "reword", then fix the scope
# e.g., chore(deepclone): ... -> chore(docs): ...
#        feat(requests): ...  -> feat: ...
```

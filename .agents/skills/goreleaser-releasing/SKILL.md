---
description: Set up GoReleaser for Go CLI/binary projects. Use when introducing binary releases, goreleaser config, release workflows, or Taskfile release tasks.
name: goreleaser-releasing
---


# GoReleaser Setup for Go Projects

Set up GoReleaser to automate cross-platform binary builds and GitHub Releases for Go CLI tools.

## Prerequisites

- Go 1.22+ (or version specified in `go.mod`)
- [Task](https://taskfile.dev/) runner
- GitHub repository with tag-based releases
- Project has a `cmd/` directory with a `main.go` entry point

## Quick Decision Tree

**Step 1: Detect project type**

```bash
# Check for cmd/ directory (binary project)
ls cmd/*/main.go 2>/dev/null

# Check for private dependencies
grep "github.com/agentable" go.mod

# Check for CGO dependencies
grep "CGO_ENABLED" Taskfile.yml Makefile 2>/dev/null
```

**Step 2: Select templates**

```
Has cmd/ directory with main.go?
  ├─ Yes → GoReleaser applicable
  │   ├─ Has private deps (github.com/agentable/*)?
  │   │   └─ Yes → assets/release-private.yml + assets/goreleaser.yml
  │   │   └─ No  → assets/release-opensource.yml + assets/goreleaser.yml
  │   └─ Requires CGO?
  │       └─ Yes → Add CGO config to goreleaser.yml (see references/configuration-guide.md)
  └─ No  → GoReleaser not needed (library-only package)
```

| Project Type | GoReleaser Config | Release Workflow |
|-------------|------------------|-----------------|
| Open-source CLI | `assets/goreleaser.yml` | `assets/release-opensource.yml` |
| Private deps CLI | `assets/goreleaser.yml` | `assets/release-private.yml` |
| CGO required | Customize `goreleaser.yml` | Add system deps step |

## Setup Workflow

### Step 1: Create `.goreleaser.yaml`

Copy the template and customize:

```bash
cp assets/goreleaser.yml .goreleaser.yaml
```

**Required customizations:**

1. **`builds[].main`** — Path to main package (e.g., `./cmd/myapp`)
2. **`builds[].binary`** — Output binary name
3. **`builds[].ldflags`** — Version injection variables (match your `internal/version` package)
4. **`release.github.owner`** and **`release.github.name`** — GitHub org/repo
5. **`release.header`** — Customize installation instructions

**Optional customizations:**

- **`builds[].ignore`** — Skip platform combinations (e.g., windows/arm64)
- **`archives[].files`** — Extra files to include (README, LICENSE, config examples)
- **`before.hooks`** — Pre-build commands (e.g., `go generate ./...`)

### Step 2: Create Release Workflow

```bash
mkdir -p .github/workflows
cp assets/release-opensource.yml .github/workflows/release.yml
# OR for private deps:
cp assets/release-private.yml .github/workflows/release.yml
```

### Step 3: Add Taskfile Tasks

Add these tasks to `Taskfile.yml`:

```yaml
  install-goreleaser:
    desc: Install goreleaser if not present
    cmds:
      - |
        if ! command -v goreleaser >/dev/null 2>&1; then
          echo "Installing GoReleaser..."
          go install github.com/goreleaser/goreleaser@latest
        fi
    status:
      - command -v goreleaser >/dev/null 2>&1

  release-check:
    desc: Check GoReleaser configuration
    deps:
      - install-goreleaser
    cmds:
      - echo "Checking GoReleaser configuration..."
      - goreleaser check

  snapshot:
    desc: Create a snapshot release (without publishing)
    deps:
      - install-goreleaser
    cmds:
      - echo "Creating snapshot release..."
      - goreleaser release --snapshot --clean --skip=publish

  build-all:
    desc: Build binaries for all platforms using GoReleaser
    deps:
      - install-goreleaser
    cmds:
      - echo "Building for all platforms..."
      - goreleaser build --snapshot --clean

  release:
    desc: Create a full release (requires git tag)
    deps:
      - install-goreleaser
    cmds:
      - echo "Creating release..."
      - goreleaser release --clean
```

### Step 4: Validate Configuration

```bash
# Check config syntax
task release-check

# Test a snapshot build (no publish)
task snapshot

# Verify built artifacts
ls dist/
```

### Step 5: Test the Release Flow

```bash
# Create a test tag
git tag -a v0.1.0 -m "v0.1.0"
git push origin v0.1.0

# Monitor the GitHub Actions release workflow
gh run watch
```

## GoReleaser Config Anatomy

Key sections in `.goreleaser.yaml`:

| Section | Purpose |
|---------|---------|
| `version` | Config format version (always `2`) |
| `before.hooks` | Pre-build commands (`go mod tidy`, `go generate`) |
| `builds` | Build matrix: platforms, flags, ldflags |
| `archives` | Archive format and naming |
| `checksum` | Checksum file generation |
| `changelog` | Changelog from conventional commits |
| `release` | GitHub Release configuration and notes |
| `snapshot` | Snapshot version template for local builds |

## Version Injection Pattern

Standard pattern for injecting version info via ldflags:

**In `internal/version/version.go`:**

```go
package version

var (
    Number string = "dev"
    Commit string = "unknown"
    Date   string = "unknown"
)
```

**In `cmd/myapp/main.go`:**

```go
package main

var (
    version = "dev"
    commit  = "unknown"
    date    = "unknown"
)
```

**In `.goreleaser.yaml`:**

```yaml
builds:
  - ldflags:
      - -s -w
      - -X main.version={{.Version}}
      - -X main.commit={{.Commit}}
      - -X main.date={{.Date}}
```

Choose one pattern. The `main` package pattern is simpler. The `internal/version` pattern is better for libraries that also ship a CLI.

## Changelog with Conventional Commits

GoReleaser generates changelogs from git history. Configure grouping for conventional commits:

```yaml
changelog:
  sort: asc
  use: github
  filters:
    exclude:
      - '^docs:'
      - '^test:'
      - '^chore:'
      - '^style:'
      - Merge pull request
      - Merge branch
  groups:
    - title: Features
      regexp: '^.*?feat(\(.+\))??!?:.+$'
      order: 0
    - title: Bug fixes
      regexp: '^.*?fix(\(.+\))??!?:.+$'
      order: 1
    - title: Performance improvements
      regexp: '^.*?perf(\(.+\))??!?:.+$'
      order: 2
    - title: Others
      order: 999
```

## Common Issues

### Issue 1: "dirty" suffix in version

**Cause:** Uncommitted changes when running GoReleaser.

**Solution:** Commit or stash all changes before tagging.

### Issue 2: Missing `fetch-depth: 0` in CI

**Cause:** Shallow clone cannot generate changelog.

**Solution:** Set `fetch-depth: 0` in `actions/checkout`.

### Issue 3: `go mod tidy` fails during build

**Cause:** Dependencies not downloaded before parallel builds.

**Solution:** Add `go mod tidy` and `go mod download` to `before.hooks`.

### Issue 4: Private dependency auth in CI

**Cause:** GoReleaser build needs access to private Go modules.

**Solution:** Use `release-private.yml` template with PAT_TOKEN.

## Advanced Configuration

For CGO builds, multiple binaries, Docker images, Homebrew taps, and other advanced scenarios, see [references/configuration-guide.md](references/configuration-guide.md).

## Validation Checklist

After setup, verify:

- [ ] `.goreleaser.yaml` exists at project root
- [ ] `goreleaser check` passes
- [ ] `task snapshot` builds successfully
- [ ] `.github/workflows/release.yml` triggers on `v*` tags
- [ ] Release workflow has `fetch-depth: 0` in checkout
- [ ] Release workflow has `contents: write` permission
- [ ] Version ldflags match your version package variables
- [ ] Changelog filters exclude non-user-facing commits
- [ ] Archive naming produces readable filenames

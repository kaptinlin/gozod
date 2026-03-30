# GoReleaser Configuration Guide

Detailed reference for `.goreleaser.yaml` customization beyond the basic template.

## Table of Contents

- [Multiple Binaries](#multiple-binaries)
- [CGO Builds](#cgo-builds)
- [Version Injection Patterns](#version-injection-patterns)
- [Archive Customization](#archive-customization)
- [Release Notes Customization](#release-notes-customization)
- [Platform Matrix](#platform-matrix)
- [Build Hooks](#build-hooks)
- [Reproducible Builds](#reproducible-builds)

## Multiple Binaries

Build multiple binaries from different `cmd/` directories:

```yaml
builds:
  - id: cli
    main: ./cmd/cli
    binary: cli
    env:
      - CGO_ENABLED=0
    goos: [linux, darwin, windows]
    goarch: [amd64, arm64]
    ldflags:
      - -s -w -X main.version={{.Version}}
    flags:
      - -trimpath

  - id: worker
    main: ./cmd/worker
    binary: worker
    env:
      - CGO_ENABLED=0
    goos: [linux]
    goarch: [amd64, arm64]
    ldflags:
      - -s -w -X main.version={{.Version}}
    flags:
      - -trimpath

archives:
  - id: cli
    ids: [cli]
    name_template: "cli_{{ .Os }}_{{ .Arch }}"

  - id: worker
    ids: [worker]
    name_template: "worker_{{ .Os }}_{{ .Arch }}"
```

## CGO Builds

When CGO is required (e.g., SQLite, libvips):

```yaml
builds:
  - id: myapp
    main: ./cmd/myapp
    binary: myapp
    env:
      - CGO_ENABLED=1
    goos: [linux, darwin]
    goarch: [amd64, arm64]
    ldflags:
      - -s -w -X main.version={{.Version}}
    overrides:
      - goos: darwin
        goarch: amd64
        env:
          - CC=o64-clang
      - goos: darwin
        goarch: arm64
        env:
          - CC=aarch64-apple-darwin20.2-clang
```

**CI workflow addition for system dependencies:**

```yaml
      - name: Install system dependencies
        run: |
          sudo apt-get update
          sudo apt-get install -y build-essential pkg-config libsqlite3-dev
```

## Version Injection Patterns

### Pattern 1: Main package variables (simple)

```go
// cmd/myapp/main.go
package main

var (
    version = "dev"
    commit  = "unknown"
    date    = "unknown"
)
```

```yaml
ldflags:
  - -s -w
  - -X main.version={{.Version}}
  - -X main.commit={{.Commit}}
  - -X main.date={{.Date}}
```

### Pattern 2: Internal version package (reusable)

```go
// internal/version/version.go
package version

var (
    Number string = "dev"
    Commit string = "unknown"
    Date   string = "unknown"
)

func String() string {
    return Number + " (" + Commit + ") " + Date
}
```

```yaml
ldflags:
  - -s -w
  - -X 'github.com/OWNER/REPO/internal/version.Number={{.Version}}'
  - -X 'github.com/OWNER/REPO/internal/version.Commit={{.Commit}}'
  - -X 'github.com/OWNER/REPO/internal/version.Date={{.Date}}'
```

## Archive Customization

### Human-readable archive names

```yaml
archives:
  - name_template: >-
      {{ .ProjectName }}_
      {{- title .Os }}_
      {{- if eq .Arch "amd64" }}x86_64
      {{- else if eq .Arch "386" }}i386
      {{- else }}{{ .Arch }}{{ end }}
      {{- if .Arm }}v{{ .Arm }}{{ end }}
    files:
      - README.md
      - LICENSE*
      - docs/*
    format_overrides:
      - goos: windows
        formats: [ zip ]
```

### Include additional files

```yaml
archives:
  - files:
      - README.md
      - LICENSE*
      - completions/*
      - manpages/*
```

## Release Notes Customization

### Custom header with install instructions

```yaml
release:
  header: |
    ## {{ .ProjectName }} {{ .Tag }}

    Short description of the release.

    ### Installation

    #### Homebrew
    ```bash
    brew install OWNER/tap/myapp
    ```

    #### Download
    Download the appropriate binary from the assets below.

    #### Go Install
    ```bash
    go install github.com/OWNER/REPO/cmd/myapp@{{ .Tag }}
    ```
  footer: |

    **Full Changelog**: https://github.com/OWNER/REPO/compare/{{ .PreviousTag }}...{{ .Tag }}
```

### Changelog grouping for conventional commits

```yaml
changelog:
  sort: asc
  use: github
  format: "{{.SHA}}: {{.Message}}"
  filters:
    exclude:
      - '^docs:'
      - '^test:'
      - '^chore:'
      - '^style:'
      - 'merge conflict'
      - Merge pull request
      - Merge branch
  groups:
    - title: Breaking Changes
      regexp: '^.*?!:.+$'
      order: 0
    - title: Features
      regexp: '^.*?feat(\(.+\))??!?:.+$'
      order: 1
    - title: Bug fixes
      regexp: '^.*?fix(\(.+\))??!?:.+$'
      order: 2
    - title: Performance
      regexp: '^.*?perf(\(.+\))??!?:.+$'
      order: 3
    - title: Others
      order: 999
```

## Platform Matrix

### Default targets

```yaml
builds:
  - goos: [linux, darwin, windows]
    goarch: [amd64, arm64]
```

### Extended targets with ARM

```yaml
builds:
  - goos: [linux, darwin, windows]
    goarch: ["386", amd64, arm, arm64]
    goarm: ["7"]
    ignore:
      - goos: windows
        goarch: arm
      - goos: windows
        goarch: arm64
      - goos: darwin
        goarch: "386"
```

### Minimal (server-only)

```yaml
builds:
  - goos: [linux]
    goarch: [amd64, arm64]
```

## Build Hooks

### Pre-build hooks

```yaml
before:
  hooks:
    - go mod tidy
    - go generate ./...
    - go fmt ./...
```

### Per-build hooks

```yaml
builds:
  - hooks:
      pre:
        - cmd: go generate ./...
          dir: "{{ dir .Dist }}"
      post:
        - upx "{{ .Path }}"
```

## Reproducible Builds

For deterministic, reproducible releases:

```yaml
builds:
  - ldflags:
      - -s -w
      - -X main.version={{.Version}}
      - -X main.commit={{.Commit}}
      - -X main.date={{.CommitDate}}
    mod_timestamp: "{{ .CommitTimestamp }}"
    flags:
      - -trimpath
```

Key changes from default:
- Use `{{.CommitDate}}` instead of `{{.Date}}` for build date
- Set `mod_timestamp` to `{{.CommitTimestamp}}`
- Always use `-trimpath` flag

## Taskfile Integration

Complete set of GoReleaser tasks for `Taskfile.yml`:

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

  tag:
    desc: Create and push a new tag (usage task tag -- v0.1.0)
    cmds:
      - |
        VERSION="{{.CLI_ARGS}}"
        if [ -z "$VERSION" ]; then
          echo "Usage: task tag -- v0.1.0"
          exit 1
        fi
        echo "Creating tag $VERSION..."
        git tag -a "$VERSION" -m "Release $VERSION"
        git push origin "$VERSION"
        echo "Tag $VERSION created and pushed"
```

## Library-Only Projects

If your project is a library with no CLI binary, GoReleaser is not needed. Use the `releasing` skill instead for tag-only releases.

To skip builds in GoReleaser (if you still want automated release notes):

```yaml
builds:
  - skip: true

release:
  github:
    owner: OWNER
    name: REPO
```

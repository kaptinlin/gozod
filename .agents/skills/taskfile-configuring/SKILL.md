---
description: Create and manage Taskfiles for Go projects using the Task task runner. Use when setting up build automation, creating or modifying Taskfile.yml files, configuring task dependencies, or structuring multi-module task workflows.
name: taskfile-configuring
---


# Golang Taskfile

Create and manage Taskfiles for Go projects using the Task task runner.

## Scripts

| Script | Purpose |
|--------|---------|
| `scripts/init_taskfile.sh` | Generate Taskfile.yml — auto-detects single vs multi-module layout |
| `scripts/init_taskfile.sh <path>` | Generate in a specific project directory |

```bash
bash scripts/init_taskfile.sh              # current directory
bash scripts/init_taskfile.sh /path/to/project
```

## Quick Start

Basic Taskfile structure:

```yaml
version: '3'

tasks:
  build:
    desc: Build the Go binary
    cmds:
      - go build -v -o ./app{{exeExt}} .

  test:
    desc: Run tests
    cmds:
      - go test -v ./...

  run:
    desc: Run the application
    deps: [build]
    cmds:
      - ./app{{exeExt}}
```

## Common Patterns

### Build and Test Workflow

```yaml
version: '3'

tasks:
  default:
    desc: Run tests and build
    cmds:
      - task: test
      - task: build

  build:
    desc: Build the binary
    sources:
      - '**/*.go'
    generates:
      - ./bin/app{{exeExt}}
    cmds:
      - go build -o ./bin/app{{exeExt}} .

  test:
    desc: Run all tests
    cmds:
      - go test -race -v ./...

  lint:
    desc: Run linter
    cmds:
      - golangci-lint run
```

### Variables and Environment

```yaml
version: '3'

vars:
  BINARY_NAME: myapp
  VERSION:
    sh: git describe --tags --always --dirty

env:
  CGO_ENABLED: 0

tasks:
  build:
    cmds:
      - go build -ldflags="-X main.Version={{.VERSION}}" -o {{.BINARY_NAME}}{{exeExt}}
```

### Cross-Platform Builds

```yaml
version: '3'

tasks:
  build-all:
    desc: Build for all platforms
    cmds:
      - for:
          matrix:
            OS: [linux, darwin, windows]
            ARCH: [amd64, arm64]
        cmd: GOOS={{.ITEM.OS}} GOARCH={{.ITEM.ARCH}} go build -o bin/{{.BINARY_NAME}}-{{.ITEM.OS}}-{{.ITEM.ARCH}}{{if eq .ITEM.OS "windows"}}.exe{{end}}
```

### Dependencies and Ordering

```yaml
version: '3'

tasks:
  deploy:
    desc: Deploy application
    deps: [test, build]  # Run in parallel
    cmds:
      - echo "Deploying..."

  setup:
    desc: Setup development environment
    cmds:
      - task: install-deps
      - task: generate-code  # Sequential execution
      - task: build
```

### File Watching

```yaml
version: '3'

tasks:
  dev:
    desc: Watch and rebuild on changes
    watch: true
    sources:
      - '**/*.go'
    cmds:
      - go build -o ./bin/app{{exeExt}}
      - ./bin/app{{exeExt}}
```

### Conditional Execution

```yaml
version: '3'

tasks:
  deploy-prod:
    desc: Deploy to production
    preconditions:
      - test -f .env
      - sh: '[ "$ENV" = "production" ]'
        msg: "ENV must be set to production"
    cmds:
      - echo "Deploying to production..."
```

### Including Other Taskfiles

```yaml
version: '3'

includes:
  docker:
    taskfile: ./docker/Taskfile.yml
    dir: ./docker

  tools:
    taskfile: ./tools/Taskfile.yml

tasks:
  build:
    cmds:
      - go build .
      - task: docker:build
```

## Advanced Features

### Dynamic Variables

```yaml
version: '3'

tasks:
  release:
    vars:
      GIT_COMMIT:
        sh: git rev-parse --short HEAD
      BUILD_TIME:
        sh: date -u +"%Y-%m-%dT%H:%M:%SZ"
    cmds:
      - go build -ldflags="-X main.Commit={{.GIT_COMMIT}} -X main.BuildTime={{.BUILD_TIME}}"
```

### Looping Over Files

```yaml
version: '3'

tasks:
  test-packages:
    vars:
      PACKAGES:
        sh: go list ./...
    cmds:
      - for: { var: PACKAGES }
        cmd: go test -v {{.ITEM}}
```

### Status Checks (Skip if Up-to-Date)

```yaml
version: '3'

tasks:
  generate:
    desc: Generate code from protobuf
    sources:
      - 'proto/**/*.proto'
    generates:
      - 'pkg/pb/**/*.go'
    cmds:
      - protoc --go_out=. proto/**/*.proto
```

### Cleanup with Defer

```yaml
version: '3'

tasks:
  test-with-db:
    cmds:
      - docker run -d --name test-db postgres:15
      - defer: docker rm -f test-db
      - go test -v ./...
```

## Project Type Detection

Before creating a Taskfile, determine the project type:

```bash
MODULE_COUNT=$(find . -name go.mod -not -path '*/vendor/*' -not -path '*/.*' | wc -l)
```

| Count | Type | Template |
|-------|------|----------|
| 1 | Single-module | Use "Single-Module Taskfile" |
| >1 | Multi-module | Use "Multi-Module Taskfile" |

**Key differences:**

| Feature | Single-Module | Multi-Module |
|---------|--------------|--------------|
| MODULES variable | Not needed | Auto-discovered via `find` |
| `test` | `go test -race ./...` | Root only |
| `test:all` | Not needed | Root + all sub-modules |
| `lint:all` | Not needed | Root + all sub-modules |
| `tidy:all` | Not needed | Root + all sub-modules |
| `deps:update` | `go get -u ./... && go mod tidy` | Three-phase: root, sub-modules (external only), `go work sync` |
| `deps` | `go mod download && go mod tidy` | Same (root only) |

## Go-Specific Patterns

### Tool Version Management

Use a `.golangci.version` file to pin the linter version:

```bash
echo "2.9.0" > .golangci.version
```

Then reference it in your Taskfile:

```yaml
vars:
  GOBIN: '{{.PROJECT_ROOT}}/bin'
  GOLANGCI_LINT_BINARY: '{{.GOBIN}}/golangci-lint'
  REQUIRED_GOLANGCI_LINT_VERSION:
    sh: cat .golangci.version 2>/dev/null || echo "2.9.0"
  GOLANGCI_LINT_VERSION:
    sh: '{{.GOLANGCI_LINT_BINARY}} version --format short 2>/dev/null || echo "not-installed"'

tasks:
  install-golangci-lint:
    desc: Install golangci-lint with the required version
    cmds:
      - mkdir -p {{.GOBIN}}
      - |
        if [ "{{.GOLANGCI_LINT_VERSION}}" != "{{.REQUIRED_GOLANGCI_LINT_VERSION}}" ]; then
          curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b {{.GOBIN}} v{{.REQUIRED_GOLANGCI_LINT_VERSION}}
        fi
    status:
      - test "{{.GOLANGCI_LINT_VERSION}}" = "{{.REQUIRED_GOLANGCI_LINT_VERSION}}"
```

The `status:` field ensures the installation is skipped if the correct version is already installed.

### Git Submodules

For packages that require git submodules:

```yaml
tasks:
  submodules:
    desc: Initialize and update git submodules
    cmds:
      - echo "Initializing git submodules..."
      - git submodule update --init --depth 1 || echo "Submodule init completed with warnings"

  test:
    desc: Run all tests
    deps:
      - submodules  # Ensure submodules are initialized before tests
    cmds:
      - go test -race ./...
```

### Running Examples

For packages with examples:

```yaml
tasks:
  examples:
    desc: Run all examples
    cmds:
      - echo "Running examples..."
      - for: { var: EXAMPLE_DIRS }
        cmd: cd {{.ITEM}} && go run main.go
    vars:
      EXAMPLE_DIRS:
        sh: find examples -name main.go -exec dirname {} \;
```

### Module Management (Single-Module)

```yaml
version: '3'

tasks:
  deps:
    desc: Download dependencies
    cmds:
      - go mod download

  tidy:
    desc: Tidy go.mod
    cmds:
      - go mod tidy

  deps:update:
    desc: Update all dependencies
    cmds:
      - go get -u ./... && go mod tidy

  vendor:
    desc: Vendor dependencies
    cmds:
      - go mod vendor
```

### Module Management (Multi-Module)

Multi-module repos need auto-discovery, per-module iteration, and a replace-aware dependency update strategy.

**MODULES auto-discovery** — no manual list to maintain:

```yaml
vars:
  MODULES:
    sh: find . -mindepth 2 -name go.mod -not -path '*/vendor/*' -not -path '*/.*' -exec dirname {} \; | sed 's|^\./||' | sort
```

**Per-module tasks:**

```yaml
tasks:
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
    deps:
      - install-golangci-lint
    cmds:
      - '{{.GOLANGCI_LINT_BINARY}} run --timeout=10m --path-prefix .'
      - for: { var: MODULES }
        cmd: |
          echo "Running golangci-lint in {{.ITEM}}..."
          cd {{.ITEM}} && {{.GOLANGCI_LINT_BINARY}} run --timeout=10m --path-prefix .

  tidy:all:
    desc: Run go mod tidy for all modules
    cmds:
      - go mod tidy
      - for: { var: MODULES }
        cmd: |
          echo "Tidying {{.ITEM}}..."
          cd {{.ITEM}} && go mod tidy
```

**deps:update** — three-phase strategy required because replace directives are not transitive:

```yaml
tasks:
  deps:update:
    desc: Update all dependencies across root and all sub-modules
    cmds:
      # Phase 1: Root module (has all replace directives, so go get -u ./... works)
      - GOWORK=off go get -u ./... && GOWORK=off go mod tidy
      # Phase 2: Sub-modules (only update external, non-replaced deps)
      - for: { var: MODULES }
        cmd: |
          if [ ! -f {{.ITEM}}/go.mod ]; then exit 0; fi
          echo "Updating dependencies in {{.ITEM}}..."
          REPLACED=$(awk '
            /^replace[[:space:]]+[^(]/ { print $2 }
            /^replace[[:space:]]*\(/ { f=1; next }
            f && /^\)/ { f=0; next }
            f && /=>/ { print $1 }
          ' {{.ITEM}}/go.mod)
          DEPS=$(awk '
            /^require[[:space:]]+[^(]/ { print $2 }
            /^require[[:space:]]*\(/ { f=1; next }
            f && /^\)/ { f=0; next }
            f && !/\/\/ indirect/ && NF >= 2 { print $1 }
          ' {{.ITEM}}/go.mod | while read -r dep; do
            echo "$REPLACED" | grep -qxF "$dep" || echo "$dep"
          done | tr '\n' ' ')
          if [ -n "$DEPS" ]; then
            cd {{.ITEM}} && GOWORK=off go get -u $DEPS
          fi
      # Phase 3: Sync workspace to propagate version changes
      - go work sync
```

**Why the three-phase approach?** `go get -u ./...` in a sub-module follows replace directives to the root, but the root's replace directives for sibling modules don't apply transitively. Go tries to fetch siblings at `v0.0.0` from the proxy and fails. The awk-based filter parses both single-line (`replace module => path`) and block (`replace ( ... )`) formats from `go.mod`, extracts each direct dependency, and skips any with a corresponding `replace` directive. Do not run `go mod tidy` in sub-modules with `GOWORK=off` — it fails for the same transitive reason; `go work sync` handles consistency instead.

### Testing Patterns

```yaml
version: '3'

tasks:
  test:
    desc: Run all tests
    cmds:
      - go test -v -race -coverprofile=coverage.out ./...

  test-short:
    desc: Run short tests
    cmds:
      - go test -v -short ./...

  coverage:
    desc: Show test coverage
    deps: [test]
    cmds:
      - go tool cover -html=coverage.out
```

### Linting and Formatting

```yaml
version: '3'

tasks:
  fmt:
    desc: Format code
    cmds:
      - go fmt ./...
      - goimports -w .

  lint:
    desc: Run linters
    cmds:
      - golangci-lint run --fix

  vet:
    desc: Run go vet
    cmds:
      - go vet ./...
```

## Best Practices

1. **Use descriptive task names**: `build`, `test`, `deploy` are clear
2. **Add descriptions**: Use `desc:` for `task --list` output
3. **Leverage dependencies**: Use `deps:` for parallel execution
4. **Use sources/generates**: Skip tasks when files haven't changed
5. **Use status checks**: Skip expensive operations when not needed
6. **Set appropriate working directories**: Use `dir:` when needed
7. **Use variables for reusability**: Define common values in `vars:`
8. **Platform-specific tasks**: Use `platforms:` to restrict tasks
9. **Silent mode for clean output**: Use `silent: true` for cleaner logs
10. **Version management**: Use external files (`.golangci.version`) for tool versions
11. **GOBIN management**: Always set GOBIN to `./bin` for reproducible builds
12. **Race detection**: Always use `-race` flag for tests
13. **Error handling**: Use `|| true` for commands that may fail gracefully
14. **Multi-module: auto-discover MODULES**: Use `find`-based `sh:` variable instead of hardcoded list
15. **Multi-module: never `go get -u ./...` in sub-modules**: Replace directives aren't transitive — use replace-aware selective update
16. **Multi-module: always `go work sync`**: After updating any module's deps, sync the workspace

## Single-Module Taskfile

For projects with a single `go.mod`:

```yaml
version: '3'

vars:
  PROJECT_ROOT:
    sh: pwd
  GOBIN: '{{.PROJECT_ROOT}}/bin'
  GOLANGCI_LINT_BINARY: '{{.GOBIN}}/golangci-lint'
  REQUIRED_GOLANGCI_LINT_VERSION:
    sh: cat .golangci.version 2>/dev/null || echo "2.9.0"
  GOLANGCI_LINT_VERSION:
    sh: '{{.GOLANGCI_LINT_BINARY}} version --format short 2>/dev/null || {{.GOLANGCI_LINT_BINARY}} version --short 2>/dev/null || echo "not-installed"'

env:
  GOBIN: '{{.GOBIN}}'

tasks:
  default:
    desc: Run lint and test
    cmds:
      - task: lint
      - task: test

  help:
    desc: Show this help message
    cmds:
      - echo "Available targets:"
      - task --list
    silent: true

  clean:
    desc: Clean build artifacts and caches
    cmds:
      - echo "Cleaning build artifacts..."
      - rm -rf {{.GOBIN}}
      - go clean -cache -testcache || true

  deps:
    desc: Download Go module dependencies
    cmds:
      - echo "Downloading dependencies..."
      - go mod download
      - go mod tidy

  deps:update:
    desc: Update all dependencies
    cmds:
      - echo "Updating dependencies..."
      - go get -u ./... && go mod tidy

  test:
    desc: Run all tests with race detection
    cmds:
      - echo "Running all tests..."
      - go test -race ./...

  bench:
    desc: Run benchmarks
    cmds:
      - echo "Running benchmarks..."
      - go test -bench=. -benchmem ./...

  lint:
    desc: Run all linters
    deps:
      - golangci-lint
      - tidy-lint

  install-golangci-lint:
    desc: Install golangci-lint with the required version
    cmds:
      - mkdir -p {{.GOBIN}}
      - |
        if [ "{{.GOLANGCI_LINT_VERSION}}" != "{{.REQUIRED_GOLANGCI_LINT_VERSION}}" ]; then
          echo "Installing golangci-lint v{{.REQUIRED_GOLANGCI_LINT_VERSION}} (current: {{.GOLANGCI_LINT_VERSION}})..."
          curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b {{.GOBIN}} v{{.REQUIRED_GOLANGCI_LINT_VERSION}}
          echo "golangci-lint v{{.REQUIRED_GOLANGCI_LINT_VERSION}} installed successfully"
        fi
    status:
      - test "{{.GOLANGCI_LINT_VERSION}}" = "{{.REQUIRED_GOLANGCI_LINT_VERSION}}"

  golangci-lint:
    desc: Run golangci-lint
    deps:
      - install-golangci-lint
    cmds:
      - '{{.GOLANGCI_LINT_BINARY}} version'
      - echo "Running golangci-lint..."
      - '{{.GOLANGCI_LINT_BINARY}} run --timeout=10m --path-prefix .'

  tidy-lint:
    desc: Check if go.mod and go.sum are tidy
    cmds:
      - echo "Checking go.mod and go.sum are tidy..."
      - go mod tidy
      - git diff --exit-code -- go.mod go.sum

  fmt:
    desc: Format Go code
    cmds:
      - echo "Formatting Go code..."
      - go fmt ./...

  vet:
    desc: Run go vet
    cmds:
      - echo "Running go vet..."
      - go vet ./...

  verify:
    desc: Run all verification steps (deps, format, vet, lint, test)
    cmds:
      - task: deps
      - task: fmt
      - task: vet
      - task: lint
      - task: test
      - echo "All verification steps completed successfully"
```

## Multi-Module Taskfile

For projects with multiple `go.mod` files and a `go.work` workspace:

```yaml
version: '3'

vars:
  PROJECT_ROOT:
    sh: pwd
  GOBIN: '{{.PROJECT_ROOT}}/bin'
  GOLANGCI_LINT_BINARY: '{{.GOBIN}}/golangci-lint'
  REQUIRED_GOLANGCI_LINT_VERSION:
    sh: cat .golangci.version 2>/dev/null || echo "2.9.0"
  GOLANGCI_LINT_VERSION:
    sh: '{{.GOLANGCI_LINT_BINARY}} version --format short 2>/dev/null || {{.GOLANGCI_LINT_BINARY}} version --short 2>/dev/null || echo "not-installed"'
  # Auto-discover sub-modules (excludes vendor/ and hidden dirs like .references/)
  MODULES:
    sh: find . -mindepth 2 -name go.mod -not -path '*/vendor/*' -not -path '*/.*' -exec dirname {} \; | sed 's|^\./||' | sort

env:
  GOBIN: '{{.GOBIN}}'

tasks:
  default:
    desc: Run lint and test
    cmds:
      - task: lint
      - task: test

  help:
    desc: Show this help message
    cmds:
      - echo "Available targets:"
      - task --list
    silent: true

  clean:
    desc: Clean build artifacts and caches
    cmds:
      - echo "Cleaning build artifacts..."
      - rm -rf {{.GOBIN}}
      - go clean -cache -testcache || true

  deps:
    desc: Download Go module dependencies
    cmds:
      - echo "Downloading dependencies..."
      - go mod download
      - go mod tidy

  test:
    desc: Run all tests with race detection (root module only)
    cmds:
      - echo "Running root module tests..."
      - go test -race ./...

  test:all:
    desc: Run tests for root and all sub-modules
    cmds:
      - echo "Running root module tests..."
      - go test -race ./...
      - for: { var: MODULES }
        cmd: |
          echo "Running tests in {{.ITEM}}..."
          cd {{.ITEM}} && go test -race ./...

  bench:
    desc: Run benchmarks
    cmds:
      - echo "Running benchmarks..."
      - go test -bench=. -benchmem ./...

  lint:
    desc: Run all linters (root module only)
    deps:
      - golangci-lint
      - tidy-lint

  lint:all:
    desc: Run linter for root and all sub-modules
    deps:
      - install-golangci-lint
    cmds:
      - echo "Running golangci-lint on root..."
      - '{{.GOLANGCI_LINT_BINARY}} run --timeout=10m --path-prefix .'
      - for: { var: MODULES }
        cmd: |
          echo "Running golangci-lint in {{.ITEM}}..."
          cd {{.ITEM}} && {{.GOLANGCI_LINT_BINARY}} run --timeout=10m --path-prefix .

  tidy:all:
    desc: Run go mod tidy for root and all sub-modules
    cmds:
      - echo "Tidying root module..."
      - go mod tidy
      - for: { var: MODULES }
        cmd: |
          echo "Tidying {{.ITEM}}..."
          cd {{.ITEM}} && go mod tidy

  # Three-phase update: replace directives are not transitive,
  # so go get -u ./... fails in sub-modules. Instead:
  # 1. Root module (has all replace directives)
  # 2. Sub-modules: only update external (non-replaced) deps
  # 3. go work sync to propagate version changes
  deps:update:
    desc: Update all dependencies across root and all sub-modules
    cmds:
      - echo "Updating root module dependencies..."
      - GOWORK=off go get -u ./... && GOWORK=off go mod tidy
      - for: { var: MODULES }
        cmd: |
          if [ ! -f {{.ITEM}}/go.mod ]; then exit 0; fi
          echo "Updating dependencies in {{.ITEM}}..."
          REPLACED=$(awk '
            /^replace[[:space:]]+[^(]/ { print $2 }
            /^replace[[:space:]]*\(/ { f=1; next }
            f && /^\)/ { f=0; next }
            f && /=>/ { print $1 }
          ' {{.ITEM}}/go.mod)
          DEPS=$(awk '
            /^require[[:space:]]+[^(]/ { print $2 }
            /^require[[:space:]]*\(/ { f=1; next }
            f && /^\)/ { f=0; next }
            f && !/\/\/ indirect/ && NF >= 2 { print $1 }
          ' {{.ITEM}}/go.mod | while read -r dep; do
            echo "$REPLACED" | grep -qxF "$dep" || echo "$dep"
          done | tr '\n' ' ')
          if [ -n "$DEPS" ]; then
            echo "  Upgrading: $DEPS"
            cd {{.ITEM}} && GOWORK=off go get -u $DEPS
          else
            echo "  No external dependencies to upgrade."
          fi
      - echo "Syncing workspace..."
      - go work sync

  install-golangci-lint:
    desc: Install golangci-lint with the required version
    cmds:
      - mkdir -p {{.GOBIN}}
      - |
        if [ "{{.GOLANGCI_LINT_VERSION}}" != "{{.REQUIRED_GOLANGCI_LINT_VERSION}}" ]; then
          echo "Installing golangci-lint v{{.REQUIRED_GOLANGCI_LINT_VERSION}} (current: {{.GOLANGCI_LINT_VERSION}})..."
          curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b {{.GOBIN}} v{{.REQUIRED_GOLANGCI_LINT_VERSION}}
          echo "golangci-lint v{{.REQUIRED_GOLANGCI_LINT_VERSION}} installed successfully"
        fi
    status:
      - test "{{.GOLANGCI_LINT_VERSION}}" = "{{.REQUIRED_GOLANGCI_LINT_VERSION}}"

  golangci-lint:
    desc: Run golangci-lint
    deps:
      - install-golangci-lint
    cmds:
      - '{{.GOLANGCI_LINT_BINARY}} version'
      - echo "Running golangci-lint..."
      - '{{.GOLANGCI_LINT_BINARY}} run --timeout=10m --path-prefix .'

  tidy-lint:
    desc: Check if go.mod and go.sum are tidy
    cmds:
      - echo "Checking go.mod and go.sum are tidy..."
      - go mod tidy
      - git diff --exit-code -- go.mod go.sum

  fmt:
    desc: Format Go code
    cmds:
      - echo "Formatting Go code..."
      - go fmt ./...

  vet:
    desc: Run go vet
    cmds:
      - echo "Running go vet..."
      - go vet ./...

  verify:
    desc: Run all verification steps across all modules
    cmds:
      - task: deps
      - task: fmt
      - task: vet
      - task: lint:all
      - task: test:all
      - echo "All verification steps completed successfully"

  release:check:
    desc: Check if a new release tag is needed
    cmds:
      - bash scripts/check_changes.sh

  release:tag:
    desc: Tag a new release (usage — task release:tag -- 0.2.0)
    cmds:
      - bash scripts/tag_release.sh {{.CLI_ARGS}}
```

## Reference

For comprehensive Taskfile documentation, see [references/guide.md](references/guide.md).

## Troubleshooting

### Task not found
```bash
task --list  # List all available tasks
```

### Variables not expanding
Ensure you're using the correct syntax: `{{.VAR}}` not `$VAR`

### golangci-lint version mismatch
```bash
task install-golangci-lint  # Force reinstall
```

### Submodules not initialized
```bash
task submodules  # Initialize submodules manually
```

### Tests failing with "package not found"
```bash
task deps  # Ensure dependencies are downloaded
```

### Task runs every time (not skipping)
Use `status:` or `sources:`/`generates:` to enable smart skipping:

```yaml
tasks:
  build:
    sources:
      - '**/*.go'
    generates:
      - ./bin/app
    cmds:
      - go build -o ./bin/app
```


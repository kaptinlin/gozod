#!/usr/bin/env bash
# Generate a Taskfile.yml for Go projects.
# Auto-detects single-module vs multi-module layout.
#
# Usage:
#   bash scripts/init_taskfile.sh                  # generate in current directory
#   bash scripts/init_taskfile.sh /path/to/project  # generate in specified directory

set -euo pipefail

DIR="${1:-.}"
TASKFILE="${DIR}/Taskfile.yml"

if [ ! -f "${DIR}/go.mod" ]; then
  echo "Error: no go.mod found in ${DIR}" >&2
  exit 1
fi

if [ -f "$TASKFILE" ]; then
  echo "Error: Taskfile.yml already exists in ${DIR}" >&2
  echo "Remove it first if you want to regenerate." >&2
  exit 1
fi

# Detect project type
MODULE_COUNT=$(find "$DIR" -name go.mod -not -path '*/vendor/*' -not -path '*/.*' | wc -l | tr -d ' ')
PROJECT_NAME=$(head -1 "${DIR}/go.mod" | awk '{print $2}' | sed 's|.*/||')

echo "Project: ${PROJECT_NAME}"
echo "Modules: ${MODULE_COUNT}"

if [ "$MODULE_COUNT" -gt 1 ]; then
  echo "Type: multi-module"
  MODE="multi"
else
  echo "Type: single-module"
  MODE="single"
fi

# --- Single-module template ---
generate_single() {
  cat <<'TASKFILE_EOF'
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
TASKFILE_EOF
}

# --- Multi-module template ---
generate_multi() {
  cat <<'TASKFILE_EOF'
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
TASKFILE_EOF
}

# Generate
if [ "$MODE" = "multi" ]; then
  generate_multi > "$TASKFILE"
else
  generate_single > "$TASKFILE"
fi

# Create .golangci.version if missing
if [ ! -f "${DIR}/.golangci.version" ]; then
  echo "2.9.0" > "${DIR}/.golangci.version"
  echo "Created .golangci.version (2.9.0)"
fi

echo ""
echo "Generated ${TASKFILE} (${MODE}-module)"
echo ""
echo "Verify with:"
echo "  task --list"
echo "  task test"

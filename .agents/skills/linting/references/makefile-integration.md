# Makefile and CI Integration

## Makefile Targets

Add these targets to your project Makefile. Adjust the first comment line and `MODULE_DIRS` for your project.

```makefile
# Project Name
# Set up GOBIN so that our binaries are installed to ./bin instead of $GOPATH/bin.
PROJECT_ROOT = $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
export GOBIN = $(PROJECT_ROOT)/bin

GOLANGCI_LINT_BINARY := $(GOBIN)/golangci-lint
GOLANGCI_LINT_VERSION := $(shell $(GOLANGCI_LINT_BINARY) version --format short 2>/dev/null || $(GOLANGCI_LINT_BINARY) version --short 2>/dev/null || echo "not-installed")
REQUIRED_GOLANGCI_LINT_VERSION := $(shell cat .golangci.version 2>/dev/null || echo "2.9.0")

# Directories containing independent Go modules.
# For single-module repos: MODULE_DIRS = .
# For multi-module repos: MODULE_DIRS = . ./submodule1 ./submodule2
MODULE_DIRS = .

.PHONY: all
all: lint test

.PHONY: lint
lint: golangci-lint tidy-lint ## Run all linters

# Install golangci-lint with the required version in GOBIN if not already installed.
.PHONY: install-golangci-lint
install-golangci-lint:
	@mkdir -p $(GOBIN)
	@if [ "$(GOLANGCI_LINT_VERSION)" != "$(REQUIRED_GOLANGCI_LINT_VERSION)" ]; then \
		echo "Installing golangci-lint v$(REQUIRED_GOLANGCI_LINT_VERSION) (current: $(GOLANGCI_LINT_VERSION))..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOBIN) v$(REQUIRED_GOLANGCI_LINT_VERSION); \
		echo "golangci-lint v$(REQUIRED_GOLANGCI_LINT_VERSION) installed successfully"; \
	fi

.PHONY: golangci-lint
golangci-lint: install-golangci-lint ## Run golangci-lint
	@echo "[lint] $(shell $(GOLANGCI_LINT_BINARY) version)"
	@$(foreach mod,$(MODULE_DIRS), \
		(cd $(mod) && \
		echo "[lint] golangci-lint: $(mod)" && \
		$(GOLANGCI_LINT_BINARY) run --timeout=10m --path-prefix $(mod)) &&) true

.PHONY: tidy-lint
tidy-lint: ## Check if go.mod and go.sum are tidy
	@$(foreach mod,$(MODULE_DIRS), \
		(cd $(mod) && \
		echo "[lint] mod tidy: $(mod)" && \
		go mod tidy && \
		git diff --exit-code -- go.mod go.sum) &&) true

.PHONY: fmt
fmt: ## Format Go code
	@echo "[fmt] Formatting Go code..."
	@go fmt ./...

.PHONY: vet
vet: ## Run go vet
	@echo "[vet] Running go vet..."
	@go vet ./...

.PHONY: test
test: ## Run all tests with race detection
	@echo "[test] Running all tests..."
	@$(foreach mod,$(MODULE_DIRS),(cd $(mod) && go test -race ./...) &&) true

.PHONY: verify
verify: deps fmt vet lint test ## Run full verification pipeline
	@echo "[verify] All verification steps completed successfully"

.PHONY: deps
deps: ## Download Go module dependencies
	@echo "[deps] Downloading dependencies..."
	@go mod download
	@go mod tidy

.PHONY: clean
clean: ## Clean build artifacts and caches
	@echo "[clean] Cleaning build artifacts..."
	@rm -rf $(GOBIN)
	@go clean -cache -testcache
```

### Key Design Decisions

- **Local `./bin/`**: golangci-lint installs to project-local `./bin/`, not global paths. Add `bin/` to `.gitignore`.
- **Version pinning**: `.golangci.version` file is the single source of truth. Changing this file triggers automatic reinstall on next `task lint`.
- **Multi-module support**: `MODULE_DIRS` variable allows linting across multiple Go modules in a monorepo.
- **Tidy check**: `tidy-lint` ensures `go.mod`/`go.sum` are committed in their tidy state.

## GitHub Actions CI Workflow

Save as `.github/workflows/ci.yml`:

```yaml
name: Go

on:
  push:
    branches:
      - main
  pull_request:

jobs:
  test:
    name: Test
    runs-on: ubuntu-latest
    steps:
      - name: Check out repository
        uses: actions/checkout@v6

      - name: Set up Go
        uses: actions/setup-go@v6
        with:
          go-version-file: go.mod
          cache: true
          cache-dependency-path: go.sum

      - name: Install dependencies
        run: task deps

      - name: Run unit tests
        run: task test

  lint:
    name: Lint
    runs-on: ubuntu-latest
    steps:
      - name: Check out repository
        uses: actions/checkout@v6

      - name: Set up Go
        uses: actions/setup-go@v6
        with:
          go-version-file: go.mod
          cache: true
          cache-dependency-path: go.sum

      - name: Install dependencies
        run: task deps

      - name: Run linters
        run: task lint
```

### CI Notes

- Uses `go-version-file: go.mod` so Go version is always in sync with the project.
- Lint and test run as separate jobs for parallel execution.
- `task lint` handles golangci-lint installation automatically via the Makefile.
- No need for the `golangci-lint-action` GitHub Action — the Makefile approach gives identical behavior locally and in CI.

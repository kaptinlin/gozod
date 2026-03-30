# GitHub Actions Troubleshooting Guide

This document provides solutions to common GitHub Actions issues for Go packages.

## Authentication Issues

### Issue: "fatal: could not read Username for 'https://github.com'"

**Full error:**
```
fatal: could not read Username for 'https://github.com': No such device or address
fatal: clone of 'https://github.com/agentable/golang-skills.git' into submodule path '.agents/skills' failed
```

**Cause:** Workflow is attempting to initialize the `.agents/skills` submodule, which is a private repository.

**Solution:**

1. Remove `submodules: recursive` from checkout step:
```yaml
# Bad
- name: Check out repository
  uses: actions/checkout@v6
  with:
    submodules: recursive

# Good
- name: Check out repository
  uses: actions/checkout@v6
```

2. Remove any `git submodule update --init --recursive` commands:
```yaml
# Bad
- name: Initialize submodules
  run: git submodule update --init --recursive

# Good - only initialize specific test submodules if needed
- name: Initialize test submodules
  run: git submodule update --init tests/testdata/JSON-Schema-Test-Suite
```

### Issue: "go: github.com/agentable/package@version: invalid version: git ls-remote failed"

**Full error:**
```
go: github.com/agentable/unifai@v0.1.0: invalid version: git ls-remote -q origin in /go/pkg/mod/cache/vcs/...: exit status 128:
fatal: could not read Username for 'https://github.com': terminal prompts disabled
```

**Cause:** Package has private Go module dependencies that require authentication.

**Solution:**

1. Add PAT_TOKEN secret to repository (Settings → Secrets and variables → Actions)

2. Configure Git URL rewriting in workflow:
```yaml
env:
  GOPRIVATE: github.com/agentable/*
  GONOPROXY: github.com/agentable/*
  GOPROXY: direct

steps:
  - name: Configure Git for private modules
    env:
      TOKEN: ${{ secrets.PAT_TOKEN }}
    run: |
      git config --global url."https://x-access-token:${TOKEN}@github.com/".insteadOf "https://github.com/"
```

3. Use the `private-package.yml` template.

### Issue: "Error: Input required and not supplied: token"

**Cause:** PAT_TOKEN secret is not configured in repository settings.

**Solution:**

1. Generate a Personal Access Token (classic) with `repo` scope
2. Add it to repository: Settings → Secrets and variables → Actions → New repository secret
3. Name: `PAT_TOKEN`
4. Value: Your generated token

## Submodule Issues

### Issue: Tests fail with "no such file or directory" for test data

**Cause:** Test submodules are not initialized.

**Solution:**

1. Identify required test submodules:
```bash
grep "tests/" .gitmodules
```

2. Add selective initialization:
```yaml
- name: Initialize test submodules
  run: |
    git submodule update --init tests/testdata/JSON-Schema-Test-Suite
```

3. Use the `opensource-with-tests.yml` template.

### Issue: "Submodule 'tests/...' registered for path but not initialized"

**Cause:** Submodule path exists in `.gitmodules` but wasn't initialized.

**Solution:**

Add the specific submodule to initialization step:
```yaml
- name: Initialize test submodules
  run: |
    git submodule update --init tests/testdata/YOUR-SUBMODULE
```

## Go Module Issues

### Issue: "missing go.sum entry" or "unknown revision v0.0.0" in multi-module repo

**Full error:**
```
go: github.com/kr/text@v0.2.0: missing go.sum entry for go.mod file; to add it:
    go mod download github.com/kr/text
```
or:
```
go: reading github.com/org/repo/provider/file/go.mod at revision provider/file/v0.0.0: unknown revision provider/file/v0.0.0
```

**Cause:** Multi-module repo uses matrix CI (per-module jobs) with `GOWORK=off`. Sub-modules depend on sibling modules via `replace` with `v0.0.0` placeholders. When `go mod tidy` or `go test` runs per sub-module, it walks the full transitive graph and tries to fetch unpublished sibling modules from the network.

**Solution:**

Do not use matrix per-module CI for multi-module repos pre-release. Test root module only:

```yaml
jobs:
  test:
    env:
      GOWORK: "off"
    steps:
      - run: task test   # Root module only, not task test:all
  lint:
    env:
      GOWORK: "off"
    steps:
      - run: task lint    # Root module only
```

The root module's `replace` directives resolve local sub-module deps correctly. This problem disappears after the first release when real versions are tagged.

### Issue: "go: module cache not found"

**Cause:** Go module cache is not properly configured.

**Solution:**

Ensure `cache: true` is set:
```yaml
- name: Set up Go
  uses: actions/setup-go@v6
  with:
    go-version-file: go.mod
    cache: true
```

### Issue: "go.mod file not found"

**Cause:** Workflow is running in wrong directory or go.mod doesn't exist.

**Solution:**

1. Verify go.mod exists in repository root
2. If in subdirectory, add working-directory:
```yaml
- name: Run tests
  working-directory: ./subdir
  run: make test
```

### Issue: "go: cannot find main module"

**Cause:** Working directory doesn't contain go.mod.

**Solution:**

Add `working-directory` to all Go-related steps:
```yaml
- name: Set up Go
  uses: actions/setup-go@v6
  with:
    go-version-file: ./subdir/go.mod
    cache: true

- name: Run tests
  working-directory: ./subdir
  run: make test
```

## Makefile Issues

### Issue: "make: *** No rule to make target 'test'"

**Cause:** Makefile doesn't have a `test` target.

**Solution:**

1. Add test target to Makefile:
```makefile
.PHONY: test
test:
	go test -race -v ./...
```

2. Or run go test directly:
```yaml
- name: Run tests
  run: go test -race -v ./...
```

### Issue: "make: *** No rule to make target 'lint'"

**Cause:** Makefile doesn't have a `lint` target or golangci-lint is not installed.

**Solution:**

1. Add lint target to Makefile (see linting skill for complete setup):
```makefile
.PHONY: lint
lint: golangci-lint
	./bin/golangci-lint run ./...

.PHONY: golangci-lint
golangci-lint:
	@if [ ! -f bin/golangci-lint ]; then \
		mkdir -p bin; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ./bin v$$(cat .golangci.version); \
	fi
```

### Issue: "make: deps: No such file or directory"

**Cause:** Makefile doesn't have a `deps` target.

**Solution:**

1. Add deps target to Makefile:
```makefile
.PHONY: deps
deps:
	go mod download
	go mod verify
```

2. Or remove the deps step from workflow if not needed:
```yaml
# Remove this step if Makefile doesn't have deps target
- name: Install dependencies
  run: make deps
```

## Performance Issues

### Issue: Workflow is slow (>5 minutes)

**Causes and solutions:**

1. **Unnecessary submodule cloning:**
   - Remove all submodule initialization except required test submodules
   - Verify no `submodules: recursive` in checkout

2. **Missing Go module cache:**
   - Add `cache: true` to setup-go step
   - For multi-module repos, add `cache-dependency-path: go.sum`

3. **Downloading dependencies every time:**
   - Ensure `cache: true` is working
   - Check that go.sum is committed to repository

4. **Running unnecessary steps:**
   - Remove unused steps (e.g., deps if not needed)
   - Use conditional execution for optional steps

### Issue: Cache not working

**Cause:** Cache key is not properly configured.

**Solution:**

1. Ensure go.sum is committed:
```bash
git add go.sum
git commit -m "chore: add go.sum for caching"
```

2. For multi-module repos, specify all go.sum files:
```yaml
- name: Set up Go
  uses: actions/setup-go@v6
  with:
    go-version-file: go.mod
    cache: true
    cache-dependency-path: |
      go.sum
      submodule1/go.sum
      submodule2/go.sum
```

## Test Failures

### Issue: Tests pass locally but fail in CI

**Common causes:**

1. **Race conditions:**
   - CI runs with `-race` flag
   - Fix race conditions in code
   - Use `synctest` for concurrent tests (Go 1.25+)

2. **Missing test dependencies:**
   - Test submodules not initialized
   - External services not available
   - Add service containers if needed

3. **Environment differences:**
   - Different Go version
   - Different OS (Linux vs macOS/Windows)
   - Different timezone or locale

**Solutions:**

1. Run tests locally with race detector:
```bash
go test -race ./...
```

2. Test with same Go version as CI:
```bash
go test -race ./...
```

3. Add debug output:
```yaml
- name: Run tests
  run: |
    go version
    go env
    make test
```

### Issue: "panic: test timed out"

**Cause:** Test is hanging or taking too long.

**Solution:**

1. Add timeout to test command:
```yaml
- name: Run tests
  run: go test -race -timeout 10m ./...
```

2. Identify slow tests:
```bash
go test -v -race ./... | grep -E "PASS|FAIL"
```

3. Use `testing.Short()` for slow tests:
```go
func TestSlow(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping slow test")
    }
    // slow test code
}
```

## Lint Failures

### Issue: Linter errors that don't appear locally

**Cause:** Different golangci-lint version or configuration.

**Solution:**

1. Pin golangci-lint version in `.golangci.version`:
```
2.9.0
```

2. Ensure `.golangci.yml` is committed:
```bash
git add .golangci.yml
git commit -m "build(lint): add golangci-lint config"
```

3. Run same version locally:
```bash
make lint
```

### Issue: "golangci-lint: command not found"

**Cause:** golangci-lint is not installed in CI.

**Solution:**

Use Makefile target that auto-installs golangci-lint:
```makefile
.PHONY: golangci-lint
golangci-lint:
	@if [ ! -f bin/golangci-lint ]; then \
		mkdir -p bin; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ./bin v$$(cat .golangci.version); \
	fi

.PHONY: lint
lint: golangci-lint
	./bin/golangci-lint run ./...
```

## Workflow Syntax Issues

### Issue: "Invalid workflow file"

**Cause:** YAML syntax error.

**Solution:**

1. Validate YAML syntax:
```bash
yamllint .github/workflows/ci.yml
```

2. Check indentation (use spaces, not tabs)

3. Verify required fields are present:
```yaml
name: Go CI  # Required
on: ...      # Required
jobs: ...    # Required
```

### Issue: "Unexpected value 'submodules'"

**Cause:** Typo or incorrect indentation.

**Solution:**

Ensure proper indentation under `with`:
```yaml
# Bad
- uses: actions/checkout@v6
submodules: recursive

# Good
- uses: actions/checkout@v6
  with:
    submodules: recursive
```

## Secret Issues

### Issue: "Secret PAT_TOKEN not found"

**Cause:** Secret is not configured or has wrong name.

**Solution:**

1. Verify secret exists: Repository Settings → Secrets and variables → Actions
2. Check secret name matches exactly: `PAT_TOKEN` (case-sensitive)
3. Ensure secret is available to workflow (not organization-level only)

### Issue: "Token doesn't have required permissions"

**Cause:** PAT token lacks necessary scopes.

**Solution:**

1. Generate new token with `repo` scope (full control of private repositories)
2. For fine-grained tokens, ensure these permissions:
   - Contents: Read
   - Metadata: Read
   - Pull requests: Read (if needed)

## Debugging Tips

### Enable debug logging

Add to workflow:
```yaml
env:
  ACTIONS_STEP_DEBUG: true
  ACTIONS_RUNNER_DEBUG: true
```

### Add diagnostic steps

```yaml
- name: Debug environment
  run: |
    echo "Go version: $(go version)"
    echo "GOPATH: $GOPATH"
    echo "GOPRIVATE: $GOPRIVATE"
    echo "PWD: $(pwd)"
    ls -la
    cat go.mod

- name: Debug Git config
  run: |
    git config --list
    git submodule status
```

### Test workflow locally

Use [act](https://github.com/nektos/act):
```bash
# Install act
brew install act

# Run workflow locally
act -j test
```

## Common Workflow Mistakes

### Mistake 1: Using submodules: recursive

**Problem:** Initializes all submodules including private `.agents/skills`

**Fix:** Remove `submodules: recursive` entirely

### Mistake 2: Missing cache configuration

**Problem:** Downloads dependencies every run

**Fix:** Add `cache: true` to setup-go

### Mistake 3: Not pinning golangci-lint version

**Problem:** Lint results differ between local and CI

**Fix:** Create `.golangci.version` file

### Mistake 4: Hardcoding Go version

**Problem:** Version mismatch with go.mod

**Fix:** Use `go-version-file: go.mod`

### Mistake 5: Running deps for open-source packages

**Problem:** Unnecessary step that slows down CI

**Fix:** Remove `make deps` step for packages without private dependencies

## Getting Help

If issues persist after trying these solutions:

1. Check GitHub Actions logs for detailed error messages
2. Verify the workflow file matches one of the provided templates
3. Compare with working workflows in similar packages
4. Check if the issue is specific to the package or affects all packages
5. Review recent changes to workflow files or dependencies

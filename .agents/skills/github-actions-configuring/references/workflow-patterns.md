# GitHub Actions Workflow Patterns

This document provides detailed workflow patterns and variations for Go packages.

## Standard Workflow Structure

All workflows follow this basic structure:

```yaml
name: Go CI

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]
  workflow_dispatch:

jobs:
  test:
    # Test job configuration

  lint:
    # Lint job configuration
```

## Trigger Patterns

### Pattern 1: Standard (Recommended)

```yaml
on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]
  workflow_dispatch:
```

**Use for:** Most packages. Runs on pushes to main, all PRs, and manual triggers.

### Pattern 2: Multiple Branches

```yaml
on:
  push:
    branches: [ main, develop, release/* ]
  pull_request:
    branches: [ main, develop ]
  workflow_dispatch:
```

**Use for:** Packages with multiple active branches.

### Pattern 3: Scheduled Runs

```yaml
on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]
  schedule:
    - cron: '0 0 * * 0'  # Weekly on Sunday
  workflow_dispatch:
```

**Use for:** Packages that need periodic testing (e.g., integration tests with external services).

## Job Patterns

### Pattern 1: Basic Test Job (Open-Source)

```yaml
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

    - name: Install Task
      uses: go-task/setup-task@v1
      with:
        version: 3.x
        repo-token: ${{ secrets.GITHUB_TOKEN }}

    - name: Run tests
      run: task test
```

### Pattern 2: Test Job with Coverage

```yaml
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

    - name: Install Task
      uses: go-task/setup-task@v1
      with:
        version: 3.x
        repo-token: ${{ secrets.GITHUB_TOKEN }}

    - name: Run tests with coverage
      run: |
        task test
        go test -race -coverprofile=coverage.out -covermode=atomic ./...

    - name: Upload coverage to Codecov
      uses: codecov/codecov-action@v4
      with:
        file: ./coverage.out
        token: ${{ secrets.CODECOV_TOKEN }}
```

### Pattern 3: Matrix Testing (Multiple Go Versions)

```yaml
test:
  name: Test
  runs-on: ubuntu-latest
  strategy:
    matrix:
      go-version: ['1.23', '1.24', '1.25', '1.26']
  steps:
    - name: Check out repository
      uses: actions/checkout@v6

    - name: Set up Go ${{ matrix.go-version }}
      uses: actions/setup-go@v6
      with:
        go-version: ${{ matrix.go-version }}
        cache: true

    - name: Install Task
      uses: go-task/setup-task@v1
      with:
        version: 3.x
        repo-token: ${{ secrets.GITHUB_TOKEN }}

    - name: Run tests
      run: task test
```

### Pattern 4: Matrix Testing (Multiple OS)

```yaml
test:
  name: Test
  strategy:
    matrix:
      os: [ubuntu-latest, macos-latest, windows-latest]
  runs-on: ${{ matrix.os }}
  steps:
    - name: Check out repository
      uses: actions/checkout@v6

    - name: Set up Go
      uses: actions/setup-go@v6
      with:
        go-version-file: go.mod
        cache: true

    - name: Run tests
      run: make test
```

### Pattern 5: Test Job with Services (Database, Redis, etc.)

```yaml
test:
  name: Test
  runs-on: ubuntu-latest
  services:
    postgres:
      image: postgres:15
      env:
        POSTGRES_PASSWORD: postgres
        POSTGRES_DB: testdb
      options: >-
        --health-cmd pg_isready
        --health-interval 10s
        --health-timeout 5s
        --health-retries 5
      ports:
        - 5432:5432

    redis:
      image: redis:7
      options: >-
        --health-cmd "redis-cli ping"
        --health-interval 10s
        --health-timeout 5s
        --health-retries 5
      ports:
        - 6379:6379

  steps:
    - name: Check out repository
      uses: actions/checkout@v6

    - name: Set up Go
      uses: actions/setup-go@v6
      with:
        go-version-file: go.mod
        cache: true

    - name: Run integration tests
      env:
        DATABASE_URL: postgres://postgres:postgres@localhost:5432/testdb
        REDIS_URL: redis://localhost:6379
      run: make test-integration
```

## Submodule Patterns

### Pattern 1: No Submodules (Recommended for Most Packages)

```yaml
- name: Check out repository
  uses: actions/checkout@v6
  # No submodule configuration
```

### Pattern 2: Selective Test Submodules

```yaml
- name: Check out repository
  uses: actions/checkout@v6

- name: Initialize test submodules
  run: |
    git submodule update --init tests/testdata/JSON-Schema-Test-Suite
    git submodule update --init tests/messageformat-wg
```

### Pattern 3: Conditional Submodule Initialization

```yaml
- name: Check out repository
  uses: actions/checkout@v6

- name: Initialize test submodules
  if: github.event_name == 'push' || contains(github.event.pull_request.labels.*.name, 'full-test')
  run: |
    git submodule update --init tests/testdata/JSON-Schema-Test-Suite
```

## Private Module Patterns

### Pattern 1: Standard Private Module Access

```yaml
test:
  name: Test
  runs-on: ubuntu-latest
  env:
    GOPRIVATE: github.com/agentable/*
    GONOPROXY: github.com/agentable/*
    GOPROXY: direct
  steps:
    - name: Check out repository
      uses: actions/checkout@v6

    - name: Configure Git for private modules
      env:
        TOKEN: ${{ secrets.PAT_TOKEN }}
      run: |
        git config --global url."https://x-access-token:${TOKEN}@github.com/".insteadOf "https://github.com/"

    - name: Set up Go
      uses: actions/setup-go@v6
      with:
        go-version-file: go.mod
        cache: true
        cache-dependency-path: go.sum

    - name: Install dependencies
      run: make deps

    - name: Run tests
      run: make test
```

### Pattern 2: Multiple Private Module Sources

```yaml
env:
  GOPRIVATE: github.com/agentable/*,github.com/yourorg/*
  GONOPROXY: github.com/agentable/*,github.com/yourorg/*
  GOPROXY: direct

steps:
  - name: Configure Git for private modules
    env:
      TOKEN: ${{ secrets.PAT_TOKEN }}
    run: |
      git config --global url."https://x-access-token:${TOKEN}@github.com/".insteadOf "https://github.com/"
```

## Caching Patterns

### Pattern 1: Basic Go Module Cache

```yaml
- name: Set up Go
  uses: actions/setup-go@v6
  with:
    go-version-file: go.mod
    cache: true
```

### Pattern 2: Multi-Module Cache

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

### Pattern 3: Custom Cache Key

```yaml
- name: Cache Go modules
  uses: actions/cache@v4
  with:
    path: |
      ~/.cache/go-build
      ~/go/pkg/mod
    key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
    restore-keys: |
      ${{ runner.os }}-go-
```

## Advanced Patterns

### Pattern 1: Conditional Job Execution

```yaml
jobs:
  test:
    if: github.event_name == 'pull_request' || github.ref == 'refs/heads/main'
    # ... job configuration

  lint:
    if: github.event_name == 'pull_request'
    # ... job configuration
```

### Pattern 2: Job Dependencies

```yaml
jobs:
  test:
    # ... test configuration

  lint:
    # ... lint configuration

  build:
    needs: [test, lint]
    runs-on: ubuntu-latest
    steps:
      - name: Check out repository
        uses: actions/checkout@v6

      - name: Set up Go
        uses: actions/setup-go@v6
        with:
          go-version-file: go.mod
          cache: true

      - name: Build
        run: make build
```

### Pattern 3: Artifact Upload

```yaml
- name: Build binary
  run: make build

- name: Upload binary
  uses: actions/upload-artifact@v4
  with:
    name: binary-${{ runner.os }}-${{ github.sha }}
    path: bin/
    retention-days: 7
```

### Pattern 4: Benchmark Comparison

```yaml
benchmark:
  name: Benchmark
  runs-on: ubuntu-latest
  steps:
    - name: Check out repository
      uses: actions/checkout@v6
      with:
        fetch-depth: 0

    - name: Set up Go
      uses: actions/setup-go@v6
      with:
        go-version-file: go.mod
        cache: true

    - name: Run benchmarks
      run: |
        go test -bench=. -benchmem -run=^$ ./... | tee benchmark.txt

    - name: Compare with main
      run: |
        git checkout main
        go test -bench=. -benchmem -run=^$ ./... | tee benchmark-main.txt
        benchstat benchmark-main.txt benchmark.txt
```

## Complete Examples

### Example 1: Simple Open-Source Package

```yaml
name: Go CI

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]
  workflow_dispatch:

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

      - name: Run tests
        run: make test

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

      - name: Run linters
        run: make lint
```

### Example 2: Package with Test Submodules

```yaml
name: Go CI

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]
  workflow_dispatch:

jobs:
  test:
    name: Test
    runs-on: ubuntu-latest
    steps:
      - name: Check out repository
        uses: actions/checkout@v6

      - name: Initialize test submodules
        run: |
          git submodule update --init tests/testdata/JSON-Schema-Test-Suite

      - name: Set up Go
        uses: actions/setup-go@v6
        with:
          go-version-file: go.mod
          cache: true

      - name: Run tests
        run: make test

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

      - name: Run linters
        run: make lint
```

### Example 3: Private Package with Full CI/CD

```yaml
name: Go CI

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]
  workflow_dispatch:

jobs:
  test:
    name: Test
    runs-on: ubuntu-latest
    env:
      GOPRIVATE: github.com/agentable/*
      GONOPROXY: github.com/agentable/*
      GOPROXY: direct
    steps:
      - name: Check out repository
        uses: actions/checkout@v6

      - name: Configure Git for private modules
        env:
          TOKEN: ${{ secrets.PAT_TOKEN }}
        run: |
          git config --global url."https://x-access-token:${TOKEN}@github.com/".insteadOf "https://github.com/"

      - name: Set up Go
        uses: actions/setup-go@v6
        with:
          go-version-file: go.mod
          cache: true
          cache-dependency-path: go.sum

      - name: Install dependencies
        run: make deps

      - name: Run tests
        run: make test

  lint:
    name: Lint
    runs-on: ubuntu-latest
    env:
      GOPRIVATE: github.com/agentable/*
      GONOPROXY: github.com/agentable/*
      GOPROXY: direct
    steps:
      - name: Check out repository
        uses: actions/checkout@v6

      - name: Configure Git for private modules
        env:
          TOKEN: ${{ secrets.PAT_TOKEN }}
        run: |
          git config --global url."https://x-access-token:${TOKEN}@github.com/".insteadOf "https://github.com/"

      - name: Set up Go
        uses: actions/setup-go@v6
        with:
          go-version-file: go.mod
          cache: true
          cache-dependency-path: go.sum

      - name: Install dependencies
        run: make deps

      - name: Run linters
        run: make lint
```

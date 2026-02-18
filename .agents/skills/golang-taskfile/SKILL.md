---
name: golang-taskfile
description: Create and manage Taskfiles for Go projects using Task (task runner). Use when users mention Taskfile, task runner, build automation, or ask to create/modify Taskfile.yml files. Provides patterns for tasks, variables, dependencies, loops, and workflows.
---

# Golang Taskfile

Create and manage Taskfiles for Go projects using the Task task runner.

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

## Go-Specific Patterns

### Module Management

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

  vendor:
    desc: Vendor dependencies
    cmds:
      - go mod vendor
```

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
5. **Set appropriate working directories**: Use `dir:` when needed
6. **Use variables for reusability**: Define common values in `vars:`
7. **Platform-specific tasks**: Use `platforms:` to restrict tasks
8. **Silent mode for clean output**: Use `silent: true` for cleaner logs

## Common Go Project Taskfile

```yaml
version: '3'

vars:
  BINARY_NAME: myapp
  VERSION:
    sh: git describe --tags --always --dirty
  LDFLAGS: -ldflags="-X main.Version={{.VERSION}}"

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
      - go.mod
      - go.sum
    generates:
      - ./bin/{{.BINARY_NAME}}{{exeExt}}
    cmds:
      - go build {{.LDFLAGS}} -o ./bin/{{.BINARY_NAME}}{{exeExt}} .

  test:
    desc: Run tests with coverage
    cmds:
      - go test -v -race -coverprofile=coverage.out ./...

  lint:
    desc: Run linters
    cmds:
      - golangci-lint run

  fmt:
    desc: Format code
    cmds:
      - go fmt ./...
      - goimports -w .

  clean:
    desc: Clean build artifacts
    cmds:
      - rm -rf ./bin
      - rm -f coverage.out

  deps:
    desc: Download and tidy dependencies
    cmds:
      - go mod download
      - go mod tidy

  run:
    desc: Run the application
    deps: [build]
    cmds:
      - ./bin/{{.BINARY_NAME}}{{exeExt}}

  install:
    desc: Install the binary
    cmds:
      - go install {{.LDFLAGS}}
```

## Reference

For comprehensive Taskfile documentation, see [references/guide.md](references/guide.md).

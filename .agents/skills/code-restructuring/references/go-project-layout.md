# Go Project Layout

Concrete directory structures for Go 1.26+ projects. Choose the layout that matches your project type, then adapt.

## Standard Project Structure

```
myproject/
├── cmd/                    # Executable entry points
│   ├── api/
│   │   └── main.go         # API server binary
│   └── worker/
│       └── main.go         # Background worker binary
├── internal/               # Private application code (compiler-enforced)
│   ├── auth/
│   │   ├── token.go
│   │   └── middleware.go
│   ├── order/
│   │   ├── service.go
│   │   └── repository.go
│   └── platform/           # Shared infrastructure
│       ├── database.go
│       └── logger.go
├── pkg/                    # Public library code (importable by others)
│   ├── httpclient/
│   │   └── client.go
│   └── validation/
│       └── rules.go
├── api/                    # API definitions
│   ├── openapi.yaml
│   └── proto/
│       └── service.proto
├── web/                    # Web assets
│   ├── static/
│   └── templates/
├── configs/                # Configuration files
│   ├── config.yaml
│   └── config.prod.yaml
├── scripts/                # Build and CI scripts
│   ├── migrate.sh
│   └── seed.sh
├── build/                  # Packaging and CI
│   ├── Dockerfile
│   └── docker-compose.yaml
├── deployments/            # Deployment configurations
│   ├── kubernetes/
│   └── terraform/
├── go.mod
├── go.sum
└── README.md
```

## When to Use Each Directory

| Directory | Purpose | Rule |
|-----------|---------|------|
| `cmd/` | One subdirectory per binary | Only `main.go` + minimal wiring |
| `internal/` | Private application code | Cannot be imported outside module |
| `pkg/` | Public libraries | Only if you intend external consumers |
| `api/` | API contracts (OpenAPI, proto) | Machine-readable definitions |
| `web/` | Static assets, templates | Embedded via `//go:embed` in Go 1.26+ |
| `configs/` | Config files, not Go code | YAML, TOML, env templates |
| `scripts/` | Automation scripts | Migration, seeding, CI helpers |
| `build/` | Dockerfiles, CI configs | Build-time artifacts only |
| `deployments/` | Infra-as-code | K8s manifests, Terraform, Helm |

## File Naming Conventions

```
token.go                   # Main implementation
token_test.go              # Tests for token.go
token_integration_test.go  # Integration tests (use build tags)
doc.go                     # Package-level documentation
example_token_test.go      # Testable examples (appear in godoc)
token_unix.go              # Platform-specific (with //go:build)
token_windows.go           # Platform-specific (with //go:build)
```

**Rules:**
- `snake_case.go` always -- never camelCase or kebab-case
- `*_test.go` for all test files (Go toolchain convention)
- `doc.go` for package documentation when the package comment is long
- `example_*_test.go` for testable examples in `package foo_test`

## Build Tags (Go 1.26+)

```go
//go:build linux && amd64

package mypackage
```

Use `//go:build` (not the legacy `// +build`). Common patterns:

```go
//go:build integration          // Integration tests
//go:build !windows             // Everything except Windows
//go:build linux || darwin      // Unix-like systems
```

## Layout by Project Type

### 1. Single Binary Application

```
myapp/
├── cmd/
│   └── myapp/
│       └── main.go
├── internal/
│   ├── server/
│   │   ├── server.go
│   │   ├── routes.go
│   │   └── middleware.go
│   ├── domain/
│   │   ├── user.go
│   │   └── order.go
│   └── store/
│       ├── postgres.go
│       └── redis.go
├── go.mod
└── go.sum
```

No `pkg/` -- single binaries rarely need public libraries.

### 2. Multiple Binaries (Monorepo)

```
platform/
├── cmd/
│   ├── api/
│   │   └── main.go          # HTTP API server
│   ├── worker/
│   │   └── main.go          # Async job processor
│   ├── migrate/
│   │   └── main.go          # Database migrations
│   └── cli/
│       └── main.go          # Developer CLI tool
├── internal/
│   ├── api/                  # API-specific code
│   │   ├── handler.go
│   │   └── router.go
│   ├── worker/               # Worker-specific code
│   │   ├── processor.go
│   │   └── scheduler.go
│   ├── domain/               # Shared domain logic
│   │   ├── user.go
│   │   ├── order.go
│   │   └── notification.go
│   └── platform/             # Shared infrastructure
│       ├── database.go
│       ├── cache.go
│       └── pubsub.go
├── go.mod
└── go.sum
```

Each `cmd/` binary imports from `internal/`. Shared domain and infra live in `internal/domain/` and `internal/platform/`.

### 3. Public Library

```
myhttplib/
├── client.go                 # Top-level public API
├── client_test.go
├── request.go
├── response.go
├── option.go                 # Functional options
├── transport/                # Sub-package for transport layer
│   ├── http2.go
│   └── retry.go
├── internal/                 # Hidden implementation details
│   └── pool/
│       └── connection.go
├── example_test.go
├── doc.go
├── go.mod
└── go.sum
```

No `cmd/`, no `pkg/`. The module root IS the package. `internal/` hides implementation.

### 4. Library with CLI Tool

```
mylib/
├── mylib.go                  # Public library API
├── mylib_test.go
├── parser/                   # Public sub-package
│   ├── parser.go
│   └── parser_test.go
├── cmd/
│   └── mylib-cli/
│       └── main.go           # CLI that uses the library
├── internal/
│   └── codegen/
│       └── generator.go
├── go.mod
└── go.sum
```

### 5. Microservice

```
usersvc/
├── cmd/
│   └── usersvc/
│       └── main.go
├── internal/
│   ├── handler/              # HTTP/gRPC handlers
│   │   ├── user.go
│   │   └── health.go
│   ├── service/              # Business logic
│   │   └── user.go
│   ├── repository/           # Data access
│   │   ├── postgres.go
│   │   └── cache.go
│   └── model/                # Domain types
│       └── user.go
├── api/
│   └── proto/
│       └── user.proto
├── configs/
│   └── config.yaml
├── build/
│   └── Dockerfile
├── go.mod
└── go.sum
```

## Anti-Patterns

### Before: Flat Everything

```
myapp/
├── main.go
├── user.go
├── user_test.go
├── order.go
├── order_test.go
├── db.go
├── cache.go
├── handler_user.go
├── handler_order.go
├── middleware.go
├── config.go
├── util.go              # Grab-bag utilities
└── helper.go            # More grab-bag
```

### After: Organized by Concern

```
myapp/
├── cmd/
│   └── myapp/
│       └── main.go
├── internal/
│   ├── user/
│   │   ├── handler.go
│   │   ├── service.go
│   │   └── repository.go
│   ├── order/
│   │   ├── handler.go
│   │   ├── service.go
│   │   └── repository.go
│   ├── server/
│   │   ├── server.go
│   │   └── middleware.go
│   └── platform/
│       ├── database.go
│       ├── cache.go
│       └── config.go
├── go.mod
└── go.sum
```

No `util.go`, no `helper.go`. Every file has a clear domain home.

### Before: Premature pkg/

```
myapp/
├── cmd/myapp/main.go
├── pkg/
│   ├── models/user.go        # Only used internally
│   ├── handlers/user.go      # Only used internally
│   └── utils/string.go       # Grab-bag
└── go.mod
```

### After: internal/ Until Proven Public

```
myapp/
├── cmd/myapp/main.go
├── internal/
│   ├── user/
│   │   ├── model.go
│   │   └── handler.go
│   └── stringutil/
│       └── format.go
└── go.mod
```

Use `pkg/` only when you explicitly intend external consumers. Default to `internal/`.

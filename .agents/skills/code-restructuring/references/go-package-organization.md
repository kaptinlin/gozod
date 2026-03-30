# Go Package Organization

Domain-driven package design for Go 1.26+. Group by what code does, not what it is.

## Group by Domain, Not by Type

### Before: Grouped by Technical Layer

```
internal/
├── models/
│   ├── user.go
│   ├── order.go
│   ├── product.go
│   └── payment.go
├── handlers/
│   ├── user.go
│   ├── order.go
│   ├── product.go
│   └── payment.go
├── services/
│   ├── user.go
│   ├── order.go
│   ├── product.go
│   └── payment.go
└── repositories/
    ├── user.go
    ├── order.go
    ├── product.go
    └── payment.go
```

Problem: Adding "order returns" requires editing 4 packages. Related code is scattered.

### After: Grouped by Domain

```
internal/
├── user/
│   ├── model.go
│   ├── handler.go
│   ├── service.go
│   └── repository.go
├── order/
│   ├── model.go
│   ├── handler.go
│   ├── service.go
│   ├── repository.go
│   └── returns.go          # New feature: one package
├── product/
│   ├── model.go
│   ├── handler.go
│   └── service.go
└── payment/
    ├── model.go
    ├── handler.go
    ├── service.go
    └── gateway.go
```

Adding "order returns" touches only `internal/order/`. High cohesion.

## Flat Over Nested

### Before: Over-Nested

```
internal/
├── services/
│   └── user/
│       └── impl/
│           └── service.go
├── repositories/
│   └── user/
│       └── postgres/
│           └── repository.go
└── handlers/
    └── http/
        └── v1/
            └── user/
                └── handler.go
```

Import path: `internal/handlers/http/v1/user` -- 5 levels deep, hard to navigate.

### After: Flat Domain Packages

```
internal/
├── user/
│   ├── service.go
│   ├── repository.go       # Postgres implementation
│   └── handler.go          # HTTP handler
└── platform/
    └── postgres.go          # Shared DB connection
```

Import path: `internal/user` -- direct and obvious.

## Avoid Package Explosion

### Before: Too Many Tiny Packages

```
internal/
├── auth/
│   └── auth.go              # 1 file, 40 lines
├── authtoken/
│   └── token.go             # 1 file, 30 lines
├── authmiddleware/
│   └── middleware.go         # 1 file, 25 lines
├── authsession/
│   └── session.go           # 1 file, 35 lines
└── authvalidator/
    └── validator.go          # 1 file, 20 lines
```

5 packages, ~150 lines total. Every call requires a cross-package import.

### After: One Cohesive Package

```
internal/
└── auth/
    ├── auth.go              # Core types and logic
    ├── token.go             # Token generation/validation
    ├── middleware.go         # HTTP middleware
    ├── session.go           # Session management
    └── validator.go         # Input validation
```

One package, same ~150 lines. No cross-package imports needed.

**Rule of thumb:** If a package has 1-2 files and under 200 lines, consider merging it with a related package.

## Package Naming Anti-Patterns

### Before: Generic Names

```
internal/
├── utils/                   # What kind of utils?
│   ├── string.go
│   ├── time.go
│   ├── http.go
│   └── crypto.go
├── common/                  # Common to what?
│   ├── errors.go
│   └── constants.go
├── helpers/                 # Helpers for what?
│   ├── format.go
│   └── convert.go
├── base/                    # Base of what?
│   └── types.go
└── shared/                  # Shared by whom?
    └── config.go
```

### After: Descriptive Names

```
internal/
├── stringutil/              # String manipulation
│   └── format.go
├── timeutil/                # Time helpers
│   └── parse.go
├── httpclient/              # HTTP client wrapper
│   └── client.go
├── crypto/                  # Cryptographic operations
│   └── hash.go
├── apperror/                # Application error types
│   └── errors.go
└── config/                  # Configuration loading
    └── config.go
```

**Banned package names:** `utils`, `common`, `helpers`, `base`, `shared`, `misc`, `core`, `lib`, `types` (standalone), `models` (standalone at root).

**Naming rule:** A package name should tell you what it does, not that it's "shared."

## Dependency Direction

Dependencies flow one way: higher-level imports lower-level.

### Before: Circular Dependencies

```
internal/
├── user/
│   └── service.go           # imports internal/order (to check orders)
├── order/
│   └── service.go           # imports internal/user (to validate user)
└── ...
```

`user` -> `order` -> `user` = cycle. Does not compile.

### After: Dependency Inversion

**Option A: Extract shared interface**

```
internal/
├── user/
│   └── service.go           # imports internal/domain
├── order/
│   └── service.go           # imports internal/domain
└── domain/
    └── types.go             # User and Order interfaces/types
```

Both depend on `domain/`, neither depends on each other.

**Option B: Accept interface, return struct**

```
internal/
├── user/
│   ├── service.go           # defines UserForOrder interface
│   └── handler.go
└── order/
    └── service.go           # accepts user.UserForOrder interface
```

`order` imports `user` for the interface. `user` never imports `order`. One-way dependency.

**Option C: Event-driven decoupling**

```
internal/
├── user/
│   └── service.go           # publishes UserCreated event
├── order/
│   └── service.go           # subscribes to UserCreated event
└── event/
    └── bus.go               # Event bus (both import this)
```

No direct dependency between `user` and `order`.

## Circular Dependency Resolution

### Pattern 1: Extract Shared Types

**Before (cycle):**
```
pkg/auth/
├── user.go                  # type User struct; uses Token
└── token.go                 # type Token struct; uses User
```

`auth.User` references `auth.Token`, `auth.Token` references `auth.User` -- works within one package. But if split:

```
pkg/auth/user/user.go        # imports pkg/auth/token
pkg/auth/token/token.go      # imports pkg/auth/user  -> CYCLE
```

**After (resolved):**
```
pkg/auth/
├── types.go                 # User and Token types (same package, no cycle)
├── user_service.go          # User business logic
└── token_service.go         # Token business logic
```

Keep coupled types in the same package. Split logic, not types.

### Pattern 2: Interface at Consumer

**Before (cycle):**
```
internal/notification/
└── sender.go                # imports internal/user to get User
internal/user/
└── service.go               # imports internal/notification to send
```

**After (resolved):**
```
internal/notification/
├── sender.go                # defines Recipient interface locally
└── types.go                 # type Recipient interface { Email() string }
internal/user/
└── service.go               # User implements notification.Recipient implicitly
```

`user` does not import `notification`. `notification` defines what it needs.

### Pattern 3: Dependency Inversion with Registry

**Before (cycle):**
```
internal/plugin/
└── manager.go               # imports internal/core to access Core
internal/core/
└── app.go                   # imports internal/plugin to load plugins
```

**After (resolved):**
```
internal/plugin/
├── plugin.go                # defines Plugin interface
└── manager.go               # manages []Plugin, no core import
internal/core/
└── app.go                   # imports plugin, registers plugins
```

`core` depends on `plugin`. `plugin` depends on nothing.

### Pattern 4: Extract to Third Package

**Before (cycle):**
```
internal/billing/
└── invoice.go               # imports internal/inventory for stock check
internal/inventory/
└── stock.go                 # imports internal/billing for pricing
```

**After (resolved):**
```
internal/billing/
└── invoice.go               # imports internal/pricing
internal/inventory/
└── stock.go                 # imports internal/pricing
internal/pricing/
└── pricing.go               # shared pricing logic (new package)
```

Extract the shared concern into its own package. Both depend on it, neither depends on each other.

## Scaling: When Flat Becomes Flat-Mega

### Before: 40+ Files in One Package

```
internal/order/
├── order.go
├── order_test.go
├── line_item.go
├── discount.go
├── tax.go
├── shipping.go
├── returns.go
├── refund.go
├── fulfillment.go
├── tracking.go
├── notification.go
├── invoice.go
├── payment.go
├── validation.go
├── ... (26 more files)
```

### After: Sub-Domains

```
internal/order/
├── order.go                 # Core Order type and creation
├── line_item.go
├── validation.go
├── fulfillment/             # Sub-domain: fulfillment
│   ├── fulfillment.go
│   ├── tracking.go
│   └── shipping.go
├── billing/                 # Sub-domain: billing
│   ├── invoice.go
│   ├── payment.go
│   ├── tax.go
│   └── discount.go
└── returns/                 # Sub-domain: returns
    ├── returns.go
    └── refund.go
```

Split when a package crosses ~30 files AND has clear sub-domains. Keep the core types (`order.go`, `line_item.go`) at the package root to avoid import cycles.

## Package Design Checklist

Before creating or restructuring a package, verify:

1. **Name describes behavior** -- not `utils`, `common`, or `helpers`
2. **Single clear purpose** -- can you describe it in one sentence?
3. **Dependencies flow downward** -- higher-level imports lower-level
4. **No circular imports** -- verified with `go build ./...`
5. **Right size** -- not 1 file (merge up) and not 40+ files (split by sub-domain)
6. **Flat unless proven otherwise** -- only nest when you have 30+ files with clear sub-domains
7. **Types stay with their logic** -- do not separate `types/` from implementation
8. **internal/ by default** -- use `pkg/` only for intentional public API

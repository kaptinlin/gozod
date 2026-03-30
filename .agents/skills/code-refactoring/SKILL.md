---
description: Pragmatic architecture review and refactoring — eliminates redundancy, maximizes code reuse, pursues simplicity and elegance. Analyzes recently changed packages via git diff, produces actionable refactoring plan, then implements approved changes concurrently. Use after a batch of feature work, when code smells accumulate, or when the user says "refactor", "review architecture", "simplify code", "clean up".
name: code-refactoring
---


# Pragmatic Code Refactoring

Review recently changed code for architecture improvements. Eliminate redundancy, maximize code reuse, pursue simplicity and elegance. Produce an actionable refactoring plan, then implement approved changes.

**Announce at start:** "I'm using the code-refactoring skill."

## Core Beliefs

1. **Zero tolerance for redundancy** — if the same logic exists twice, extract it. Three similar lines are three too many when a shared helper serves the same purpose.
2. **Maximize code reuse** — every exported type, function, and constant should earn its existence. If two packages solve the same problem differently, converge on one solution.
3. **Simplicity is elegance** — the best code is the code you don't write. Fewer lines, fewer types, fewer packages. Complexity is a cost, not a feature.
4. **No documentation masquerading as code** — specs, PRDs, and architecture docs are the SSOT for rules and constraints. Code executes rules, it does not redescribe them. Prose fields no code reads, identity factories, duplicate contracts, and governance data in runtime libs belong in docs, not code.
5. **Refactoring is not rewriting** — preserve behavior exactly. Change structure, not semantics. Every refactoring must pass existing tests unchanged.

## Anti-Patterns We Reject

| Anti-Pattern | Why It's Wrong | What We Do Instead |
|---|---|---|
| Over-splitting | 10 packages with 1 file each is harder to navigate than 3 packages with 4 files | Merge related concerns until each package has a clear, substantial reason to exist |
| Over-abstracting | An interface with 1 implementation is ceremony, not design | Only extract interfaces when >=2 consumers exist or testing demands it |
| Premature generics | Generic code when only one type is used | Use concrete types until the second use case appears |
| Wrapper fever | Multiple layers of wrappers add indirection without value | Inline small wrappers; use composition only when layers have distinct responsibilities |
| Comment compensation | Long comments explaining why the code is structured a certain way | Restructure so the code is self-evident |
| Scope creep | A refactoring that touches unrelated packages "while we're at it" | One concern per finding; stay within scope |

## What to Look For

### Priority 0: Naming (Industry Alignment)

Naming is the first thing readers see. Poor names create cognitive load before any logic is read. Good names align with industry standards, eliminate redundancy, and make code self-documenting.

**Redundant prefixes/suffixes:**
- **Package context redundancy** — when package name already provides context, don't repeat it in type names
  - `validation.ValidationConfig` → `validation.Config` (package already says "validation")
  - `user.UserService` → `user.Service` (package already says "user")
  - `cache.CacheManager` → `cache.Manager` or just `cache.Cache`
- **Banned suffixes** — avoid generic suffixes that add no information
  - `*Manager`, `*Handler`, `*Helper`, `*Util` when the type does one specific thing
  - `SessionManager` → `Sessions` (if it's just a collection)
  - `ConfigHandler` → `Config` (if it just holds config)
- **Type-in-name redundancy** — don't encode the type in the name
  - `ProcessingDocument` → `Document` (the type already says it's a doc)
  - `ValidationError` → `Error` in a validation package

**Industry standard terminology:**
- **Check domain conventions** — use terms the industry recognizes
  - Lint domain: `Rule` not `Checker`, `Check()` not `Analyze()`, `Diagnostic` not `Issue`
  - HTTP domain: `Handler` not `Processor`, `Middleware` not `Interceptor`
  - Database domain: `Repository` not `Store`, `Transaction` not `Session`
- **Avoid inventing terms** — if the industry has a standard term, use it
  - Don't create `Examiner` when everyone says `Validator`
  - Don't create `Transformer` when everyone says `Converter`

**Method naming precision:**
- **Domain-specific verbs** — use verbs that match the domain
  - `Validate()` > `Process()` for validators (more precise)
  - `Build()` > `Make()` for builders (more idiomatic)
  - `Parse()` > `Convert()` for parsers (more standard)
- **Constructor conventions** — follow language idioms
  - Go: `New()` or `NewX()`, not `makeX()` or `createX()`
  - TypeScript: `create()` or constructor, not `make()`

**Interface naming simplicity:**
- **Shortest unambiguous name** — remove all unnecessary words
  - `FilterItemsByCategory()` → `ByCategory()` when the receiver already says "items"
  - `GetUserByID()` → `ByID()` when the receiver is `Users`
- **No type information** — the type signature already documents types
  - `ParseStringToInt()` → `Parse()` (signature shows string → int)

**Decision framework for renaming:**

```
1. Does the current name repeat information from context (package/type/signature)?
   YES -> rename to remove redundancy    NO -> keep

2. Is there an industry-standard term for this concept?
   YES -> use standard term    NO -> keep current name

3. Would a domain expert recognize this term immediately?
   YES -> keep    NO -> consider standard terminology

4. Does the name use a banned suffix (*Manager, *Handler, *Helper, *Util)?
   YES -> rename to specific noun or verb    NO -> keep

5. Is this a public API with external callers?
   YES -> higher bar for renaming    NO -> rename freely

6. Does renaming require changing >20 call sites?
   YES -> batch with other refactorings    NO -> rename immediately
```

**When NOT to rename:**
- The name is already industry-standard (even if verbose)
- The name is part of a stable public API with external consumers
- The name matches a spec/RFC/standard document exactly
- The verbosity adds clarity in a complex domain
- The rename would break more than it fixes

### Priority 1: Redundancy (DRY violations)

- **Duplicate logic** across packages — extract to shared internal package or common function
- **Copy-paste structs** — unify type definitions, use type aliases only when semantics differ
- **Repeated error handling patterns** — extract sentinel errors or helper functions
- **Parallel validation** — two packages validating the same data in different ways

### Priority 2: Responsibility Misplacement (SRP violations)

- **God types** — a struct with 10+ methods doing unrelated things -> split by concern
- **I/O mixed with logic** — pure algorithms depending on I/O libraries -> extract pure core
- **Package doing two jobs** — if the package doc needs "and" to describe it, consider splitting
- **Technical layer organization** — `internal/service/`, `internal/repository/` mixing all domains -> organize by domain (`internal/user/`, `internal/order/`)

### Priority 3: Dependency Inversion (DIP violations)

Core/framework packages should not depend on upper layers (presentation, orchestration, infrastructure). Dependencies should flow inward toward stable abstractions.

**Upward dependencies (core → upper layers):**
- **Presentation logic in core** — formatting, rendering, UI concerns in business logic -> move to presentation layer
- **I/O in pure logic** — file system, network, database operations in algorithms -> inject via interface or move to orchestration layer
- **CLI/API concerns in domain** — command parsing, HTTP handlers, request/response types in domain models -> move to application layer
- **Infrastructure leaking down** — logging, metrics, tracing in pure functions -> inject via interface or move to caller

**Layer confusion:**
- **Framework depending on application** — reusable validation framework importing CLI formatters -> move formatters to CLI layer
- **Library depending on binary** — shared package importing main package types -> invert dependency or extract shared types
- **Domain depending on delivery** — business logic importing REST/GraphQL/gRPC types -> use domain types, map at boundary

**Correct dependency flow:**
```
Presentation Layer (CLI, Web, API)
    ↓ depends on
Orchestration Layer (use cases, workflows)
    ↓ depends on
Domain Layer (business logic, rules)
    ↓ depends on
Infrastructure Interfaces (abstractions only)
```

**Detection heuristics:**
1. Core package imports `fmt` for formatting output -> presentation concern
2. Validation framework has `Formatter` type -> belongs in CLI
3. Business logic imports `os`, `io`, `net/http` -> I/O should be injected
4. Library package imports from `cmd/` or `internal/app/` -> inverted dependency
5. Pure algorithm returns formatted strings instead of structured data -> mixing concerns

**Refactoring strategies:**
- **Move presentation up** — formatters, renderers, templates belong in presentation layer
- **Move I/O to edges** — file/network operations belong in orchestration or infrastructure layer
- **Inject dependencies** — pass I/O capabilities via interfaces or function parameters
- **Return structured data** — core returns data structures, caller handles formatting
- **Extract shared types** — if both layers need a type, extract to shared package both can import

### Priority 4: Rigidity (OCP violations)

- **Switch on type** — growing switch statements -> use interface dispatch or registry
- **Hardcoded variants** — if adding a new case requires modifying 3 files -> use registration pattern
- **Concrete dependencies** — core logic importing specific implementations -> inject via interface/function

### Priority 4: Unnecessary Complexity (KISS violations)

- **Documentation masquerading as code** — data with no runtime consumer, identity operations, redundant contracts, governance metadata in runtime code -> delete or move to docs
- **Unused exports** — exported types/functions with zero external callers -> unexport or delete
- **Dead code** — unreachable branches, commented-out code, TODO-marked stubs -> delete
- **Over-engineered helpers** — a 30-line generic helper used once -> inline it

## Workflow

### Phase 1: Analyze

Start from recent changes, then expand to global cross-references to catch missed reuse opportunities.

**Step 1 — Scope recent changes:**

Identify what changed recently and which packages were touched.

**Step 2 — Analyze touched packages in depth:**

For each touched package:
1. Read all source files (skip test files for analysis, but ensure tests exist)
2. List exported types, functions, interfaces, sentinel errors
3. Check for the anti-patterns listed above

**Step 3 — Cross-reference against the entire codebase:**

For each touched package, ask:
1. **Does this new code duplicate logic in an untouched package?** -> extract shared helper or reuse existing one
2. **Does an untouched package have logic that could reuse something we just wrote?** -> refactor the untouched package to call the new code
3. **Are there parallel type definitions, error sentinels, or validation patterns across packages?** -> converge on one

### Phase 2: Generate Refactoring Plan

Write `.plans/YYYY-MM-DD-RN-pragmatic-refactor.md` with this structure:

```markdown
# Pragmatic Refactor Review RN

## Scope
Packages analyzed: {list}
Git range: HEAD~4..HEAD
Date: YYYY-MM-DD

## Findings

### 1. {Finding title}
**Category:** DRY / SRP / OCP / KISS
**Severity:** HIGH / MEDIUM / LOW
**Location:** {file}:{line range}
**Current:** {what it does now}
**Proposed:** {what it should do}
**Justification:** {why this change improves the codebase}

### 2. ...

## Rejected Suggestions
{List anything you considered but decided against, with reasoning}
```

**Rules for findings:**
- Every finding must have a concrete code location
- Every finding must explain WHY, not just WHAT
- Severity: HIGH = affects correctness or API; MEDIUM = affects readability; LOW = cosmetic
- If a finding is "maybe" — leave it out. Only include what you're confident about.
- **Never suggest adding code just to "improve coverage"** — refactoring reduces, it doesn't inflate

### Phase 3: Execute Approved Refactorings

After user review of the plan:

1. Remove findings the user rejects
2. Split into independent batches by package
3. Implement batches concurrently where safe
4. Run tests after each batch
5. If any test breaks, revert that specific batch and flag it
6. Validate with linter after all batches complete

### Phase 4: Polish

Final pass on changed code:
- **Tighten** — remove dead imports, unused variables, redundant comments left by the refactoring
- **Consistency** — ensure naming, error handling, and formatting patterns are uniform across touched files
- **Flow** — does each function/method lead naturally to the next? Reorder if needed
- **Cross-references** — verify all internal calls, type references, and package imports still resolve correctly

## Decision Framework

Before suggesting any refactoring, ask yourself:

```
1. Does this change reduce total lines of code?
   YES -> likely good    NO -> needs strong justification

2. Does this change reduce the number of concepts a reader must understand?
   YES -> likely good    NO -> probably not worth it

3. Would a new team member find the BEFORE or AFTER code easier to understand?
   AFTER -> proceed    BEFORE -> don't refactor

4. Does this change break any existing tests?
   NO -> proceed       YES -> reconsider the approach

5. Is there a simpler way to achieve the same improvement?
   YES -> do that instead    NO -> proceed with current plan

6. Is this code redescribing what a spec/doc already says?
   YES -> delete it          NO -> proceed
```

## Language-Specific Guidelines

This skill provides language-agnostic principles. For language-specific patterns and concrete examples, see the language-specific template references:

**Go**: `references/go-dependency-inversion.md`, `references/go-refactoring-patterns.md`, `references/go-design-principles.md`, `references/go-toolchain.md`

**TypeScript**: `references/ts-dependency-inversion.md`, `references/ts-refactoring-patterns.md`, `references/ts-design-principles.md`, `references/ts-toolchain.md`

## Remember

- **Refactoring reduces** — if your diff adds more code than it removes, reconsider
- **Tests are sacred** — if existing tests break, the refactoring is wrong, not the tests
- **One concern per finding** — don't bundle unrelated changes into one suggestion
- **Confidence threshold** — only suggest changes you're sure about. "Maybe" is not a finding.
- **No cosmetic-only changes** — renaming a variable for style alone is not worth a review cycle
- **Preserve public API** — don't break callers unless the finding is about a wrong API design
- **Preserve context** — don't make library code look like application code, or vice versa; respect the codebase's idioms

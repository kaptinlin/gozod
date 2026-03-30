---
description: Analyze gaps between SPECS and code, generate root ANALYSIS.md defining implementation scope. Use when planning code implementation, identifying missing features, or analyzing spec-to-code alignment.
name: spec-gap-analyzing
---


# Spec Gap Analyzing

Systematically analyze the gap between specification documents (`SPECS/`) and implementation code, then generate `ANALYSIS.md` at project root to define implementation scope, priorities, and structure.

**Announce at start:** "I'm using the spec-gap-analyzing skill."

## When to Use

- User asks to analyze gaps between specs and code
- User wants to plan implementation work from specifications
- User says "分析实施差距", "规划代码实施", "analyze implementation gaps", "plan code implementation"
- Before generating implementation tasks via `spec-gap-tasking`
- When creating an implementation roadmap from design documents

## Workflow

```
┌─────────────┐     ┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│  1. Survey   │────▶│  2. Analyze   │────▶│  3. Categorize│────▶│  4. Generate  │
│  SPECS + Code│     │  Gaps         │     │  & Prioritize │     │  ANALYSIS.md │
└─────────────┘     └──────────────┘     └──────────────┘     └──────────────┘
```

### Phase 1: Survey

Read both sides — specifications and implementation — to build a complete picture.

**Spec side:**
1. Read every `SPECS/*.md` file
2. Extract: contracts, data formats, exported interfaces, behavioral rules, error handling requirements
3. Note complexity: simple / medium / complex

**Implementation side:**
1. Survey implementation directories (e.g., `pkg/*/`, `src/*/`, `lib/*/`)
2. Identify: exported types, functions, classes, interfaces, error definitions
3. Estimate completeness per module

**Output:** A mental model of "what SPECS require" vs "what code provides."

### Phase 2: Analyze Gaps

For each spec, compare its requirements against the corresponding implementation:

| Gap Type | Description | Example |
|----------|-------------|---------|
| **Missing** | Spec defines it, code doesn't exist | Spec requires a retry mechanism, code has none |
| **Partial** | Code exists but doesn't cover all spec requirements | Spec defines 5 validation rules, code implements 3 |
| **Divergent** | Code contradicts spec | Spec says status `warning`, code uses `warn` |
| **Untested** | Code exists, matches spec, but lacks tests | Function exists but no test coverage |
| **Integration** | Individual modules work, cross-module flow untested | Auth → Storage → API roundtrip never tested end-to-end |

**CRITICAL: Distinguish Runtime Behavior from Documentation Constraints**

Before marking something as a gap, ask: **Does a program need to read this value to drive behavior?**

The deletion test: If you delete this code, does the program's behavior change?
- No → It's documentation disguised as code. Keep it in SPEC, enforce via tests/lint.
- Yes → It's runtime behavior. Implement it.

**NOT gaps (keep in SPEC, enforce via tests/lint):**
- Architecture principles, design philosophy, system overview
- Responsibility descriptions, scope definitions, governance rules
- "Why we chose X" rationale, trade-off discussions
- Ownership matrices, team boundaries, approval workflows
- Style guides, naming conventions (unless enforced by linter config)
- **Constraint lists expressed as type unions** — e.g., `type Forbidden = 'A' | 'B' | 'C'`
- **Compile-time assertion types** — e.g., `AssertValid<T>` that no runtime code reads
- **Identity function "validators"** — types that just echo input, admit "for docs"

**ARE gaps (implement as runtime code):**
- Data structures that code must create, validate, or transform
- Algorithms, business logic, validation rules that execute at runtime
- API contracts, function signatures, error codes that programs consume
- Configuration schemas that programs parse to drive behavior
- State machines, workflows that code must implement
- **Runtime validation functions** — actual code that checks inputs and throws/errors
- **Test assertions** — verify rules via executable test suites

### Phase 3: Categorize & Prioritize

Group gaps into categories by domain. Each category becomes a section in `ANALYSIS.md`.

**Categorization principles:**
- Group by functional domain (e.g., Authentication, Data Model, API Layer)
- Keep categories focused (3-8 items per category)
- Separate foundational work from feature work
- Identify integration test needs

**Prioritization:**
1. **Core contracts first** (foundations, config, auth) — everything depends on these
2. **Domain logic second** (business logic, data processing) — the main value
3. **Interface layer third** (CLI, API, UI) — depends on domain logic
4. **Advanced systems fourth** (optimization, caching) — independent enhancements
5. **Integration tests last** — validates the whole stack

### Phase 4: Generate ANALYSIS.md

Create `ANALYSIS.md` at project root with three sections:

#### Section 1: Scope

List all areas requiring implementation work:

```markdown
## Scope

- Implement SPECS/02 authentication (session management, permissions, token refresh)
- Implement SPECS/10 data model (schema, validation, migrations)
- Implement SPECS/20 API layer (endpoints, request handling, error responses)
- Add integration tests for auth → data → API roundtrip
```

#### Section 2: Priority

Order categories by dependency and importance:

```markdown
## Priority

1. Authentication (Category A) — foundation for all other features
2. Data Model (Category B) — core domain logic
3. API Layer (Category C) — depends on A and B
4. Integration Tests (Category IT) — validates complete flows
```

#### Section 3: Structure

Break down each category into numbered items with brief descriptions. Include From/To fields to trace back to specs and forward to implementation targets:

```markdown
## Structure

### Category A: Authentication (SPECS/02)
**From**: `SPECS/02-auth.md` (Sections 1-3) + `pkg/auth/` (current state)
**To**: `pkg/auth/` enhancements
**Priority**: High — foundation for all other features
**Items**: 3 gaps — session management, permissions, token refresh
**Gap Summary**: Missing session storage, partial permission model

- A1: Session management and storage
- A2: Permission and role definitions
- A3: Token refresh and expiry

### Category B: Data Model (SPECS/10)
**From**: `SPECS/10-data-model.md` (Sections 1-2) + `pkg/data/` (current state)
**To**: `pkg/data/` enhancements
**Priority**: Medium — core domain logic, depends on Category A
**Items**: 3 gaps — schema definitions, validation, migrations
**Gap Summary**: Partial schema, missing validation rules

- B1: Schema definitions and types
- B2: Validation rules
- B3: Migration scripts

### Category C: API Layer (SPECS/20)
**From**: `SPECS/20-api-layer.md` (Sections 1-3) + `pkg/api/` (current state)
**To**: `pkg/api/` enhancements
**Priority**: Medium — depends on A and B
**Items**: 3 gaps — routing, serialization, error handling
**Gap Summary**: Missing error handling, partial serialization

- C1: Endpoint routing and handlers
- C2: Request/response serialization
- C3: Error handling and status codes

### Category IT: Integration Tests
**From**: `SPECS/02-auth.md`, `SPECS/10-data-model.md`, `SPECS/20-api-layer.md`
**To**: integration tests
**Priority**: Low — validate after implementation
**Items**: 2 flows — auth-data-API roundtrip, schema-validation-persistence

- IT1: Auth → data access → API response
- IT2: Schema update → validation → persistence

## Dependencies

- Category B → Category A — data model needs auth context
- Category C → Category A, B — API layer needs both auth and data
- Category IT → Category A, B, C — integration tests need complete functionality
```

## ANALYSIS.md Template

```markdown
# Gap Analysis

## Context

**Previous Stage**: `SPECS/*.md` (specifications) + implementation code (current state)
**Current Goal**: Identify and close gaps between specifications and implementation
**Scope**: {scope boundary}

## Structure

### Category {ID}: {Domain Name}
**Focus**: {what gaps this category addresses}
**From**: `SPECS/{spec-file}` (Sections X-Y) + `{module/}` (current state)
**To**: `{module/}` enhancements
**Priority**: High|Medium|Low — {rationale}
**Items**: N gaps — {brief list}
**Gap Summary**: {types of gaps: Missing/Partial/Divergent/Untested}

---

### Category IT: Integration Tests
**Focus**: End-to-end cross-module validation
**From**: `SPECS/{spec1}`, `SPECS/{spec2}`
**To**: integration tests
**Priority**: Low — validate after implementation
**Items**: N flows — {brief list}

## Priorities
1. **Category {ID}** — {rationale}

## Dependencies
- Category {ID} → Category {ID} — {dependency description}

## Notes
{overall direction, risks, suggestions}
```

## Integration Test Identification

Integration tests deserve special attention. Identify integration test candidates when:

- A spec references multiple modules working together
- Data flows through a pipeline (parse → validate → transform → output)
- An I/O boundary connects two core modules
- A command orchestrates multiple backend services

**Integration test structure:**
- List all modules in the flow
- Define the end-to-end scenario (input → expected output)
- Identify mock boundaries (external APIs, file system, network)

## Inputs

Required:
- `SPECS/` directory with specification documents
- Implementation code directories (e.g., `pkg/`, `src/`, `lib/`)

Optional:
- Existing `ANALYSIS.md` to update (incremental analysis)
- Specific spec file to focus on (single-spec mode)
- Category filter (e.g., "only Category C")

## Modes

### Full Analysis Mode

Survey all SPECS, analyze all code, generate complete ANALYSIS.md.

```
User: "分析实施差距"
→ Read all SPECS/*.md
→ Survey all implementation directories
→ Generate ANALYSIS.md with all categories
```

### Single Spec Mode

Focus on one spec, analyze corresponding code.

```
User: "Analyze gaps for SPECS/20-api-layer.md"
→ Read that spec
→ Survey relevant implementation modules
→ Generate ANALYSIS.md with focused categories
```

### Incremental Mode

Update existing ANALYSIS.md after code changes.

```
User: "Update gap analysis — we implemented A1 and A2"
→ Read existing ANALYSIS.md
→ Re-survey implementation
→ Update ANALYSIS.md with remaining gaps
```

## Output Location

**Always write to project root:** `ANALYSIS.md`

This makes it easy to find and serves as the single source of truth for implementation planning.

## Next Steps

After generating `ANALYSIS.md`:

1. **Human review** (optional): Team reviews scope, priorities, and structure
2. **Generate tasks**: Use `spec-gap-tasking` skill to convert `ANALYSIS.md` into `TODO.yaml`
3. **Execute tasks**: Follow `TODO.yaml` to implement features

## Remember

- This skill **analyzes and plans** — it does not generate tasks or implement code
- `ANALYSIS.md` is a high-level analysis document, not a detailed implementation plan
- Keep categories focused (3-8 items per category)
- Integration tests are first-class citizens, not afterthoughts
- Prioritize by dependency order: foundations → domain → interface → advanced → integration
- After identifying gaps, proceed directly to generate ANALYSIS.md

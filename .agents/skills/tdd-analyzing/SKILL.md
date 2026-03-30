---
description: Analyzes SPECS documents and existing code to determine TDD implementation scope, priorities, and structure, producing an ANALYSIS.md at project root. Use when starting Code phase, planning implementation work, or when asked to analyze implementation scope.
name: tdd-analyzing
---


# TDD Analyzing

Systematically analyze specification documents (`SPECS/`) and existing code to determine implementation scope, then generate `ANALYSIS.md` at project root to define implementation priorities and structure for TDD workflow.

**Used in:** Code phase (workflow stage 3) - Analysis step

**Announce at start:** "I'm using the tdd-analyzing skill."

## Pipeline Position

```
Code Phase Workflow:

SPECS/*.md + existing code
    ↓
tdd-analyzing → ANALYSIS.md (strategic layer: scope, priority, structure)
    ↓
tdd-tasking → TODO.yaml (tactical layer: Plan+Impl task pairs)
    ↓
Ralphy loop → implementation
```

## When to Use

- Starting Code phase after Spec phase completes
- User asks to analyze implementation scope
- User wants to plan implementation work from specifications
- User says "analyze implementation scope", "plan code implementation", "determine what to build"
- Before generating implementation tasks via `tdd-tasking`

## Workflow

```
┌─────────────┐     ┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│  1. Survey   │────▶│  2. Analyze   │────▶│  3. Categorize│────▶│  4. Generate  │
│  SPECS + Code│     │  Scope        │     │  & Prioritize │     │  ANALYSIS.md │
└─────────────┘     └──────────────┘     └──────────────┘     └──────────────┘
```

### Phase 1: Survey

Read both sides — specifications and implementation — to build a complete picture.

**Spec side:**
1. Read every `SPECS/*.md` file
2. Extract: contracts, data formats, exported interfaces, behavioral rules, error handling requirements
3. Note complexity: simple / medium / complex

**Implementation side:**
1. Survey implementation directories (e.g., `pkg/*/`, `src/*/`, `lib/*/`, `internal/*/`)
2. Identify: existing modules, exported types, functions, interfaces, error definitions
3. Estimate what exists vs what needs to be built

**Output:** A mental model of "what SPECS require" vs "what needs to be implemented."

### Phase 2: Analyze Scope

For each spec, determine what needs to be implemented:

| Scope Type | Description | Example |
|------------|-------------|---------|
| **New module** | Spec defines new functionality, no code exists | Spec requires auth system, no auth code exists |
| **Module extension** | Code exists but needs new features | Existing parser needs validation rules from spec |
| **Refactor** | Code exists but needs restructuring per spec | Spec defines new interface, existing code needs adaptation |
| **Integration** | Individual modules exist, need integration | Auth + Storage + API need end-to-end flow |

### Phase 3: Categorize & Prioritize

Group implementation work into categories by domain. Each category becomes a section in `ANALYSIS.md`.

**Categorization principles:**
- Group by functional domain (e.g., Authentication, Data Model, API Layer)
- Keep categories focused (3-8 items per category)
- Separate foundational work from feature work
- Identify integration test needs

**Prioritization:**
1. **Core contracts first** (foundations, config, types) — everything depends on these
2. **Domain logic second** (business logic, data processing) — the main value
3. **Interface layer third** (CLI, API, UI) — depends on domain logic
4. **Advanced systems fourth** (optimization, caching) — independent enhancements
ntegration tests last** — validates the whole stack

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
## Prioriuthentication (Category A) — foundation for all other features
2. Data Model (Category B) — core domain logic
3. API Layer (Category C) — depends on A and B
4. Integration Tests (Category IT) — validates complete flows
```

#### Section 3: Structure

Break down each category into numbered items with brief descriptions. Include From/To fields to trace back to specs and forward to implementation targets:

```markdown
## Structure

### Category A: Authentication (SPECS/02)
**From**: `SPECS/02-auth.md` (Sections 1-3)
**To**: `pkg/auth/` (new module)
**Priority**: High — foundation for all other features
**Items**: 3 features — session management, permissions, token refresh
**Complexity**: Medium — requires external dependencies (Redis, JWT library)

- A1: Session management and storage
- A2: Permission and role definitions
- A3: Token refresh and expiry

### Category B: Data Model (SPECS/10)
**From**: `SPECS/10-data-model.md` (Sections 1-2)
**To**: `pkg/data/` (new module)
**Priority**: Medium — core domain logic, depends on Category A
**Items**: 3 features — schema definitions, validation, migrations
**Complexity**: Medium — databagration

- B1: Schema definitions and ty Validation rules
- B3: Migration scripts

### Category C: API Layer (SPECS/20)
**From**: `SPECS/20-api-layer.md` (Sections 1-3)
**To**: `pkg/api/` (new module)
**Priority**: Medium — depends on A and B
**Items**: 3 features — routing, serialization, error handling
**Complexity**: Low — standard HTTP patterns

- C1: Endpoint routing and handlers
- C2: Request/response serialization
- C3: Error handling and status codes

### Category IT: Integration Tests
**From**: `SPECS/02-auth.md`, `SPECS/10-data-model.md`, `SPECS/20-api-layer.md`
**To**: integration tests
**Priority**: Low — validate after implementation
**Itemflows — auth-data-API roundtrip, schema-validation-persistence

- IT1: Auth → data access → API response
- IT2: Schema validation → persistence → retrieval
```

## ANALYSIS.md Format

```markdown
# Implementation Analysis

## Scope
[List all implementation areas]

## Priority
[Order categories by dependency]

## Structure
[Break down each category with From/To/Priority/Items/Complexity]

### Category X: [Name] (SPECS/NN)
**From**: `SPECS/NN-xxx.md` (Sections X-Y)
**To**: `pkg/xxx/` or `src/xxx/` (target location)
**Priority**: HiLow — [reason]
**Items**: N features — [brief list]
**Complexity**: Low/Medium/High — [key challenges]

- X1: [Feature name]
- X2: [Feature name]
- X3: [Feature name]
```

## Abstraction Level

Keep analysis at **category level**, not task level:

- ✅ Good: Category A: Token System — 3 items (type system, parser, validator)
- ❌ Too detailed: A1: Implement Token struct, A2: Implement ParseToken function, A3: Write tests for Token

The goal is strategic planning — let users approve the big picture before diving into tasks.

## After Writing

> "ANALYSIS.md has been generated at project root.
>
> **Review checklist:**
> - Does the scope cover all SPECS requirements?
> - Is the priority order correct (dependencies first)?
> - Are categories appropriately sized (3-8 items each)?
>
> Ready to generate implementation tasks using `tdd-tasking`?"

## Remember

- This skill is the **first step** in Code phase — it analyzes scope, not gaps
- ANALYSIS.md is strategic (categories), TODO.yaml is tactical (tasks)
- Each category item will become a Plan+Impl task pair in `tdd-tasking`
- Keep abstraction at module/feature level, not function/file level
- Trace back to SPECS (From field) and forward to code location (To field)
- Complexity assessment helps `tdd-tasking` decide simple vs Plan+Impl tasks

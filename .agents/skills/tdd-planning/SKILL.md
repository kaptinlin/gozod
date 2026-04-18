---
description: Generates a detailed TDD implementation plan for a single feature from a PRD, spec, or design document, defining the interface contract and acceptance criteria. Use when starting a Plan task in Code phase, before implementing any non-trivial feature.
name: tdd-planning
---


# TDD Planning

Turn a PRD, spec, or design document into a self-contained TDD implementation plan. The plan defines **what to build and the interface contract** — no implementation code bodies. A developer picks it up and implements the whole feature using TDD, then commits once.

**Used in:** Code phase (workflow stage 3) - Plan tasks

**Announce at start:** "I'm using the tdd-planning skill."

## Context

This skill generates implementation plans for TDD-driven feature development. The plan serves as a contract between planning and implementation phases, defining:

- **What to build**: Feature scope and behavioral requirements
- **Interface contract**: All exported signatures and types
- **Acceptance criteria**: Observable behaviors to test
- **TDD guidance**: Where to start and what to mock

The plan is implementation-agnostic — it specifies the "what" and "why", not the "how".

## Pipeline Position

```
Code Phase Workflow:

tdd-analyzing → tdd-tasking → Ralphy loop:
                                 ├─ Plan task → tdd-planning → .plans/YYYY-MM-DD-{feature}.md
                                 ├─ Impl task → tdd-implementing → code + tests
                                 └─ Checkpoint → code-simplifying / code-refactoring
```

## Scope

### What Goes in the Plan

| Include | Exclude |
|---------|---------|
| Exported function/method signatures | Function bodies |
| Type and error definitions | Test code |
| Behavioral acceptance criteria | File paths |
| TDD requirement | Language-specific test commands |
| Architectural decisions | Implementation details |
| System boundary mock points | Internal package structure |

Interface signatures are the stable contract — they guide both the test writer and the implementer.

### Plan Granularity

**Feature scope**: One plan = one cohesive feature that delivers user-facing value

**Phase granularity**: A phase is a coherent chunk of behavior — not a single test, not an entire subsystem.

| Too coarse | Right |
|-----------|-------|
| "Implement auth system" | "JWT parsing and validation" |
| "Build the API" | "POST /users — create user with duplicate detection" |

One phase = multiple acceptance criteria, each of which becomes one TDD cycle during implementation.

## TDD Philosophy

Every plan embeds this requirement:

> **TDD**: For each acceptance criterion, write one failing test first. See red. Write minimal code to pass. See green. Move to the next criterion. Repeat until all criteria pass. No horizontal slicing — do not write all tests upfront.

**Vertical slices only:**
```
WRONG: write all tests → write all code
RIGHT: test1→impl1 → test2→impl2 → test3→impl3 → ...
```

**Key principles:**
- Tests verify observable behavior through exported interfaces
- Tests survive internal refactors
- One feature = many criteria = many TDD cycles
- Commit once when the whole feature is done — not after each cycle

For full TDD execution guidance: `tdd-implementing` skill.

## Workflow

### Inputs

- A spec document (from `spec-writing`)
- A PRD or requirements document
- A design document (DESIGN.md, ARCHITECTURE.md)
- An improvement plan (improve.md, REFACTOR.md)
- Inline description

**If ambiguous**, ask: "Which feature should I plan? Point me to the spec/PRD/improve doc, or describe the feature."

### Before Writing

1. **Read the input document** thoroughly — identify the single feature
2. **Study reference implementations** — check `.references/` for similar features in reference projects. Read their actual code to understand API patterns, error handling, and edge cases. This is not optional; real-world patterns prevent over-engineering and catch missed edge cases.
3. **Study existing codebase patterns** — read at least 2 existing similar implementations in the project (e.g., existing stores, plugins, commands) to understand:
   - Constructor/option patterns the project uses
   - Error sentinel conventions
   - Interface compliance patterns (compile-time checks, type annotations)
   - Test patterns (assertion frameworks, table-driven tests, setup helpers)
   - Concurrency patterns (mutexes, atomic flags, async/await, close semantics)
4. **Draft the phase breakdown** — present to user for granularity check
5. **Confirm**: "I'll write a plan for **[feature name]** with these phases: [...]. Does that look right?"

**Reference study is critical.** Plans without reference study produce over-engineered abstractions and miss provider-specific edge cases that real implementations reveal.

### Output Location

Save to: `.plans/YYYY-MM-DD-<feature-name>.md`

## Plan Structure

### Document Sections

```
# [Feature] — Implementation Plan
## Constraints          ← TDD / KISS / DRY (required, verbatim)
## Reference Study      ← What was learned from .references/ and existing code
## Architecture         ← Key decisions, packages, data flow
## Interface Contract   ← All exported signatures and types for the feature
## Phases               ← Behavioral vertical slices
## Commit               ← Single commit at the end
## Out of Scope
## References
```

### Reference Study Section

Summarize what was learned from reference implementations and existing codebase patterns:

```markdown
## Reference Study

### From .references/
- **[reference-name]**: [What was learned — API patterns, error handling, edge cases]
- **Key insight**: [The most important design decision borrowed from reference]

### From existing codebase
- **[existing-package]**: [Pattern to follow for consistency — constructor, options, errors, tests]
- **Convention**: [Specific project convention this implementation must follow]
```

This section prevents re-inventing patterns that already exist and ensures consistency. The implementer reads this first to understand the design rationale.

### Interface Contract Section

Signatures only — no bodies. All phases reference this section.

```language-agnostic
// Type definitions
type Token {
    userID: string
    expiresAt: DateTime
}

// Error definitions
ErrExpired = "token expired"
ErrInvalid = "token invalid"

// Function signatures
function ParseToken(raw: string): Result<Token, Error>
function ValidateToken(token: Token, secret: Bytes): Result<void, Error>
```

**Note:** See `references/code-patterns.md` in your project templates for language-specific interface contract syntax.

### Phase Structure

````markdown
## Phase N: [Behavior Name]

**Delivers:** [What a caller can do after this phase is implemented]

### Interface additions

```language-agnostic
// Signatures introduced or modified in this phase
function NewThing(config: Config): Result<Thing, Error>
```

**Note:** See `references/code-patterns.md` in your project templates for language-specific syntax.

### Acceptance criteria

- [ ] Given [normal input] → returns [expected output]
- [ ] Given [edge case] → returns [ErrXxx]
- [ ] Given [boundary] → [observable result]

### TDD notes

- **Start with:** [first criterion — usually the happy path]
- **Mock at:** [external boundary if any, e.g., HTTP client, `clock.Now()`] — never internal packages
- **KISS:** [one over-engineering temptation to avoid]
````

## Acceptance Criteria Guidelines

Each acceptance criterion maps to **one test function or one table-driven test set**. Write the minimum set that gives sufficient confidence.

### Always Include

- **Happy path** — the primary success case (one test)
- **Each distinct sentinel error** — every `ErrXxx` a caller must handle gets its own criterion
- **Boundaries that change behavior** — e.g., empty input, zero value, expired state

### Table-Driven Criteria (group as one)

When multiple inputs produce the same class of outcome — list them as one criterion with examples:

```
- [ ] Invalid formats (empty string, no "@", missing domain) → ErrInvalidEmail
```

This becomes one table-driven test. Do not write a separate criterion per invalid input.

### Separate Criteria

When behaviors require different setup or produce meaningfully different outcomes:

```
- [ ] Duplicate email on register → ErrEmailTaken   ← separate: different error
- [ ] Valid email on register → returns new User     ← separate: success path
```

### Skip These

| Skip | Why |
|------|-----|
| Simple constructors with no logic | Nothing to assert |
| Getters / setters | No behavior |
| Standard library behavior | Not your code |
| Internal state not observable via exported API | Tests implementation, not behavior |
| Error paths that behave identically to an existing criterion | Redundant |

### Coverage Targets

| Code | Target |
|------|--------|
| Core business logic | 100% of exported behaviors |
| Error paths with distinct `ErrXxx` | 100% |
| General public API | 80%+ |
| Trivial wrappers / generated code | Exclude |

## Handling Architectural Constraints

When specs or requirements define constraints (forbidden values, required patterns, validation rules), distinguish between **runtime behavior** and **development constraints**.

### The Deletion Test

Ask: *If I delete this code, does the program's behavior change?*

- **No** → It's documentation disguised as code. Keep it in SPEC, enforce via tests/lint.
- **Yes** → It's runtime behavior. Implement and test it.

### Constraints vs Implementation

| Type | How to Enforce | Example |
|------|----------------|---------|
| **Forbidden values** that code must validate | Runtime validation function + test | `rejects invalid inputs with error` |
| **Module import restrictions** | Lint rule / CI validation script | CI fails if State layer imports Signal runtime |
| **Naming conventions** | Linter / pre-commit hook | `oxlint --rule naming` |
| **Architecture rules** (layer boundaries) | Dependency check in CI | `tsc --noEmit` + custom validator |
| **Type union constraints** | NOT code — documentation only | `type Forbidden = 'A' \| 'B' \| 'C'` → delete |

### What Goes in the Plan

**✅ Include in Plan:**
- Runtime validation that users observe (e.g., "rejects invalid input with clear error")
- Tests that verify the validation executes
- CI checks that enforce architectural rules

**❌ Exclude from Plan:**
- Type-level constraints that no runtime code reads
- "Assertion" types that are identity functions
- Compile-time-only restrictions without enforcement mechanism

### Example: Constraint in Acceptance Criteria

```
## Phase 2: Input Validation

### Acceptance criteria

- [ ] Given forbidden action value → throws validation error
- [ ] Given required pattern mismatch → returns descriptive error

### TDD notes

- **Start with:** forbidden value rejection
- **Mock at:** none (pure function)
- **Test enforces constraint:** verify the exact error message
```

The **test** is where the constraint lives — executable verification that survives refactoring. The SPEC describes why the constraint exists. Type declarations alone cannot enforce it.

## Acceptance Criteria from References

When `.references/` contains a similar implementation, mine it for acceptance criteria you might miss:

1. **Read the reference's error handling** — each distinct error case becomes an acceptance criterion
2. **Read the reference's test file** — test names reveal edge cases the reference author discovered
3. **Read the reference's auth chain** — for cloud stores, auth fallback order matters and each step needs testing
4. **Read the reference's API mapping** — scope/name/key conventions need explicit criteria

Do not copy reference code into the plan. Extract the **behaviors** and **edge cases**, then express them as acceptance criteria in the plan's own terms.

## Sub-Module Awareness

When the feature requires a new sub-module (e.g., separate package/module):

- Note in Architecture: "New sub-module with separate module configuration"
- Include in Phase 1: sub-module setup (module config, dependencies for local dev)
- Acceptance criteria should include: compile-time interface compliance check
- Note external SDK dependencies and their version constraints

**Note:** See `references/code-patterns.md` in your project templates for language-specific sub-module setup.

## Key Principles

- **Interface Contract** is the single source of truth — all phases reference it, no duplication
- **Reference Study** prevents over-engineering — borrow patterns, don't invent
- **Acceptance criteria** are behavioral: Given/When/Then, no implementation detail
- **TDD notes** point to where to start and where the boundary is — not how to implement
- **One commit** at the end of the whole plan, not after each phase
- This skill is used in Code phase for Plan tasks generated by `tdd-tasking`

## After Writing

> "Plan saved to `.plans/<filename>.md`.
>
> **Execute options:**
> 1. **This session** — I implement the plan using `tdd-implementing` and commit when done
> 2. **New session** — Open a new session with the plan file and `tdd-implementing`
>
> Which do you prefer?"

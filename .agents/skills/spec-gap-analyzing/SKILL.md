---
description: Analyze gaps between SPECS and code, generate root ANALYSIS.md defining implementation scope. Use when planning code implementation, identifying missing features, or analyzing spec-to-code alignment.
name: spec-gap-analyzing
---


# Spec Gap Analyzing

Systematically analyze the gap between specification documents (`SPECS/`) and
implementation code, then generate `ANALYSIS.md` at project root to define
implementation scope, priorities, and structure.

**Announce at start:** "I'm using the spec-gap-analyzing skill."

## Pipeline Position

```
SPECS/*.md + existing code
  -> spec-gap-analyzing              <- you are here
  -> ANALYSIS.md
  -> spec-gap-tasking
  -> TODO.yaml
```

## When to Use

Use this skill when:
- Existing code partially implements the specs and you need to find what's missing
- User asks to analyze gaps between specs and code
- User wants to plan implementation work from specifications
- User says "分析实施差距", "规划代码实施", "analyze implementation gaps",
  "plan code implementation"
- Before generating implementation tasks via `spec-gap-tasking`

Do NOT use this skill when:
- You are planning *fresh implementation* from specs with no existing code
  (use `spec-implementing-analyzing`)
- You need task generation directly from an existing `ANALYSIS.md`
  (use `spec-gap-tasking`)
- Specifications have not been written yet (use spec-writing skills first)

**Stage 3 vs Stage 4:** `spec-implementing-analyzing` plans *fresh builds*.
This skill audits *existing implementation* to find where code diverges from
or falls short of specs.

## Workflow

```
┌─────────────┐   ┌──────────────┐   ┌──────────────┐   ┌──────────────┐   ┌──────────────┐
│  1. Survey   │──▶│  2. Design   │──▶│  3. Analyze   │──▶│  4. Scope    │──▶│  5. Generate  │
│  SPECS + Code│   │  Review      │   │  Gaps         │   │  Reduction   │   │  ANALYSIS.md │
└─────────────┘   └──────────────┘   └──────────────┘   └──────────────┘   └──────────────┘
```

### Phase 1: Survey

Read both sides — specifications and implementation — to build a complete picture.

**Spec side:**
1. Read every `SPECS/*.md` file
2. Extract: contracts, data formats, exported interfaces, behavioral rules,
   error handling requirements
3. Note complexity per spec: simple / medium / complex

**Implementation side:**
1. Survey implementation directories (e.g., `pkg/*/`, `src/*/`, `lib/*/`)
2. Identify: exported types, functions, classes, interfaces, error definitions
3. Estimate completeness per module
4. Note what already exists that works well — don't plan to rewrite working code

**Output:** A mental model of "what SPECS require" vs "what code provides."

### Phase 2: Design Review — Challenge the Specs

Before analyzing gaps, critically evaluate whether the specs themselves are sound.
A gap between code and a bad spec is not a gap to close — it's a spec to fix.

- **Contradictions** — a spec principle violated by its own interface design
- **Over-engineering** — external libraries or complex patterns for trivially
  simple problems
- **Abstraction mismatches** — custom interfaces that duplicate standard library
  capabilities with no added value
- **Broken symmetry** — inconsistent patterns within the same system without
  justification
- **Truth gaps** — factories/constructors that return instances which cannot
  fulfill their advertised contract
- **Missing API surface** — types lacking essential methods for debugging or
  practical use

**How to review:**

1. For each spec contract, ask: *Does this interface actually support the design
   goals stated in the spec's rationale section?*
2. For each external dependency, ask: *Can this be done in ≤30 lines of standard
   code? If yes, the dependency is unjustified.*
3. For each custom abstraction, ask: *Does the standard library already provide
   this? What does the custom version add?*
4. For each pattern choice, ask: *Is this consistent with how the rest of the
   system works?*

**Output format in ANALYSIS.md:**

Include a `## Design Review` section listing issues by severity:
- ⚠️ **Severe** — architectural contradiction, will cause runtime problems
- ⚠️ **Moderate** — unnecessary complexity, adds maintenance burden
- ⚠️ **Minor** — inconsistency or missing convenience, easy to fix

For each issue: what the spec says, why it's a problem, suggested alternative,
consequence of not changing.

End with explicit **decision questions** the user must answer before proceeding.

### Phase 3: Analyze Gaps

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

### Phase 4: Scope Reduction — What NOT to Close

Not every gap is worth closing. Before categorizing, actively eliminate
unnecessary work. The goal is a product where every part is excellent, not a
product where every spec line has a corresponding code line.

**Elimination criteria — cut if ANY apply:**

1. **"Nice to have" test** — would a user notice if this gap remained? If not,
   cut it entirely (not "defer" — cut).
2. **Working code test** — does existing code solve the problem, even if it
   doesn't match the spec's exact interface? Prefer keeping working code over
   rewriting to match spec aesthetics.
3. **Standard library test** — does the language's stdlib already handle this?
   Don't wrap working stdlib usage in a custom abstraction just because the
   spec defines one.
4. **Single-caller test** — is this gap about an abstraction used by only one
   consumer? Inline it instead.
5. **Future-proofing test** — is this gap about something that *might* be needed?
   Cut it. YAGNI.

Document what was cut in `## Scope Reductions` — this prevents re-litigating.

**Design philosophy — think like a founder, not a contractor:**

Say no to a thousand things. Close gaps to build a **complete, production-quality
system** — not to achieve spec compliance for its own sake. Ship once, ship right.

**The Apple principle:** Don't split gap-closing into phases. Don't create "v1 gaps"
and "v2 gaps." Close the gaps that matter, close them completely, and cut the rest.
A product with 5 flawless features beats one with 15 mediocre ones.

- **Simplicity is the ultimate sophistication** — if two gap categories share
  types or logic, merge them. Fewer categories with clear boundaries beat many
  fine-grained ones
- **No over-engineering** — don't close gaps by adding plugin systems or
  middleware chains. One concrete fix beats three layers of indirection
- **System coherence** — every gap category must describe how it integrates with
  other categories. Isolated fixes create integration debt
- **No shortcuts, no phases** — do not mark anything as "close later" or
  "Phase 2". Either close it now and do it right, or cut it from scope entirely

### Phase 5: Categorize & Prioritize

Group remaining gaps into categories by domain. Each category becomes a section
in `ANALYSIS.md`.

**Categorization principles:**
- Group by functional domain (e.g., Authentication, Data Model, API Layer)
- Keep categories focused (3-8 items per category)
- Separate foundational work from feature work
- Identify integration test needs
- **Aim for fewer categories** — 4-6 categories for a typical project. If you
  have more than 8, you're likely too granular or haven't reduced scope enough

**Prioritization:**
1. **Core contracts first** (foundations, config, auth) — everything depends on these
2. **Domain logic second** (business logic, data processing) — the main value
3. **Interface layer third** (CLI, API, UI) — depends on domain logic
4. **Advanced systems fourth** (optimization, caching) — independent enhancements
5. **Integration tests last** — validates the whole stack

### Phase 6: Generate ANALYSIS.md

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

## Design Review

{List design problems found in specs, ordered by severity.
For each issue: current design, why it's problematic, suggested alternative,
consequence of not changing.
End with numbered decision questions for the user.}

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

## Scope Reductions

{Items explicitly cut from scope and why. Prevents re-litigating.}

- {Cut item 1} — {reason}
- {Cut item 2} — {reason}

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

## Common Mistakes

- **Closing every gap**: Not every gap between spec and code is worth closing.
  A gap where working code uses a simpler approach than the spec prescribes is
  not a deficiency — it might be an improvement.
- **Skipping design review**: Gaps caused by bad specs should be fixed in the
  spec, not in the code. Implementing a bad spec perfectly is still bad.
- **Skipping scope reduction**: The natural instinct is to close every gap.
  Fight it. A product with 5 flawlessly closed gaps beats 15 half-closed ones.
- **Phase splitting**: Creating "Phase 1: Critical Gaps" and "Phase 2: Nice to
  Have." Either close a gap properly or cut it. No "we'll do it right later."
- **Too many categories**: If you have 10+ categories, you haven't reduced
  scope — you've just organized complexity. Step back and merge or cut.
- **Rewriting working code**: If existing code works but doesn't match the
  spec's exact interface, consider updating the spec instead.
- **Being too polite about spec problems**: If a spec contradicts itself or
  prescribes unnecessary complexity, say so directly with concrete evidence.

## Remember

- This skill **analyzes, reviews, and plans** — it does not generate tasks or implement code
- **Design review is mandatory** — never skip Phase 2. Specs are not sacred.
- **Scope reduction is mandatory** — never skip Phase 4. Cut aggressively.
- **No phased implementation** — do not split gaps into "v1" and "v2" categories
- `ANALYSIS.md` is a high-level analysis document, not a detailed implementation plan
- Keep categories focused: 3-8 items per category, aim for 4-6 total categories
- Integration tests are first-class citizens, not afterthoughts
- Prioritize by dependency order: foundations → domain → interface → advanced → integration
- Always trace categories back to specific SPECS files
- After identifying gaps, proceed directly to generate ANALYSIS.md
- End with explicit decision questions — let the user decide on design issues

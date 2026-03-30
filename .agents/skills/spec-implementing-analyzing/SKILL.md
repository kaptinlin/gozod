---
description: Analyze SPECS documents and existing code to generate root ANALYSIS.md defining Stage 3 implementation scope, priority, and module structure. Use when starting spec-based coding implementation.
name: spec-implementing-analyzing
---


# Spec Implementing Analyzing

Systematically analyze specification documents (`SPECS/`) and existing code to
generate `ANALYSIS.md` at project root, defining what needs to be built from
scratch, the implementation priority order, and the module structure.

**Announce at start:** "I'm using the spec-implementing-analyzing skill."

## Pipeline Position

```
SPECS/*.md + existing code
  -> spec-implementing-analyzing    <- you are here
  -> ANALYSIS.md
  -> spec-implementing-tasking
  -> TODO.yaml
```

## When to Use

Use this skill when:
- Starting Stage 3 implementation from approved specifications
- User asks to analyze implementation scope based on SPECS
- User needs implementation categories and priorities before task generation
- User says "analyze implementation scope", "spec implementation analysis",
  "plan implementation from specs"
- Before generating implementation tasks via `spec-implementing-tasking`

Do NOT use this skill when:
- You are in Stage 4 gap alignment — existing code already partially implements
  the specs and you need to find what is missing (use `spec-gap-analyzing`)
- You need task generation directly from an existing `ANALYSIS.md`
  (use `spec-implementing-tasking`)
- Specifications have not been written yet (use spec-writing skills first)

**Stage 3 vs Stage 4:** This skill plans *fresh implementation* — building new
modules from specs. `spec-gap-analyzing` audits *existing implementation* to
find where code diverges from or falls short of specs.

## Workflow

### Step 1: Survey Input Materials

**Spec side:**
1. Read every `SPECS/*.md` file
2. Extract: contracts, data formats, exported interfaces, behavioral rules,
   error handling requirements
3. Note complexity per spec: simple / medium / complex

**Code side:**
1. Survey implementation directories (e.g., `src/`, `lib/`, or language-specific
   module directories)
2. Identify: existing types, functions, classes, interfaces, modules
3. Note what already exists that can be reused or extended

**Output:** A mental model of "what SPECS require to be built" vs "what already
exists in code."

### Step 2: Analyze Implementation Scope

For each spec, determine what needs to be built:

| Scope Type | Description | Example |
|------------|-------------|---------|
| **New module** | Nothing exists, build from scratch | Spec defines auth system, no auth code exists |
| **New component** | Module exists, component does not | Spec adds validation to existing data module |
| **Extension** | Component exists, needs new capabilities | Spec requires 3 more output formats for existing transformer |
| **Foundation** | Shared infrastructure needed by multiple specs | Config system, error types, shared interfaces |

Focus on *what needs building*, not on gaps in existing code. This is Stage 3:
the assumption is that implementation is largely ahead of you, not behind you.

**CRITICAL: Distinguish Runtime Behavior from Documentation Constraints**

Before marking something for implementation, ask: **Does a program need to read this value to drive behavior?**

The deletion test: If you delete this code, does the program's behavior change?
- No → It's documentation disguised as code. Keep it in SPEC, enforce via tests/lint.
- Yes → It's runtime behavior. Implement it.

**NOT implementation scope (keep in SPEC, enforce via tests/lint):**
- Architecture principles, design philosophy, system overview
- Responsibility descriptions, scope definitions, governance rules
- "Why we chose X" rationale, trade-off discussions
- Ownership matrices, team boundaries, approval workflows
- Style guides, naming conventions (unless enforced by linter config)
- **Constraint lists expressed as type unions** — e.g., `type Forbidden = 'A' | 'B' | 'C'`
- **Compile-time assertion types** — e.g., `AssertValid<T>` that no runtime code reads
- **Identity function "validators"** — types that just echo input, admit "for docs"

**ARE implementation scope (write runtime code):**
- Data structures that code must create, validate, or transform
- Algorithms, business logic, validation rules that execute at runtime
- API contracts, function signatures, error codes that programs consume
- Configuration schemas that programs parse to drive behavior
- State machines, workflows that code must implement
- **Runtime validation functions** — actual code that checks inputs and throws/errors
- **Test assertions** — verify rules via executable test suites

### Step 3: Categorize & Prioritize

Group implementation work into categories by functional domain. Each category
becomes a section in `ANALYSIS.md`.

**Categorization principles:**
- Group by functional domain (e.g., Data Model, Auth, API Layer, CLI)
- Keep categories focused (3-8 items per category)
- Separate foundational work from feature work
- Include integration tests as a dedicated category

**Prioritization:**
1. **Foundations first** (config, shared types, core interfaces) — everything
   depends on these
2. **Domain logic second** (business rules, data processing) — the main value
3. **Interface layer third** (CLI, API, UI) — depends on domain logic
4. **Advanced systems fourth** (optimization, caching) — independent enhancements
5. **Integration tests last** — validates the whole stack

### Step 4: Generate ANALYSIS.md

Write `ANALYSIS.md` at project root using the template below.

**Validation before writing:**
- Confirm file location: project root
- Confirm all sections present: Context, Structure, Priorities, Dependencies,
  Notes
- Confirm item counts are consistent across sections
- Confirm clear traceability: SPECS -> categories -> target modules

### Step 5: Prompt for Review

Tell the user:

```
I identified N implementation areas across M categories.
ANALYSIS.md has been generated at project root.

Please review ANALYSIS.md to confirm the scope and priorities.
Would you like me to proceed with task generation?
```

## ANALYSIS.md Template

```markdown
# Implementation Analysis

## Context

**Previous Stage**: SPECS/*.md (approved specifications)
**Current Goal**: Implement system based on specifications
**Scope**: {brief boundary statement — what is and is not included}

## Structure

### Category {ID}: {Category Name}

**Focus**: {what this category covers}

**From**: SPECS/{spec-file} (Sections X-Y)

**To**: {target module or directory}

**Priority**: High|Medium|Low — {rationale}

**Items**: N items — {brief list of components}

---

### Category {ID}: {Category Name}

**Focus**: {what this category covers}

**From**: SPECS/{spec-file} (Sections X-Y)

**To**: {target module or directory}

**Priority**: High|Medium|Low — {rationale}

**Items**: N items — {brief list of components}

---

### Category IT: Integration Tests

**Focus**: End-to-end validation of cross-module flows

**From**: SPECS/{relevant specs}

**To**: integration test directory

**Priority**: Low — validates complete flows after modules are built

**Items**: N items — {brief list of flows}

## Priorities

1. **Category {ID}** — {why this comes first}
2. **Category {ID}** — {why this comes second}
3. **Category IT** — {why last}

## Dependencies

- Category {ID} -> Category {ID} — {dependency explanation}
- Category {ID} -> Category {ID} — {dependency explanation}

## Notes

{Overall direction, risks, architectural decisions, recommendations}
```

## Inputs

Required:
- `SPECS/` directory with approved specification documents
- Implementation source directories (to assess what already exists)

Optional:
- Existing `ANALYSIS.md` to update (incremental mode)
- Specific spec file to focus on (single-spec mode)
- Category filter (e.g., "only analyze the data model")

## Modes

### Full Analysis

Survey all SPECS, assess all existing code, generate complete `ANALYSIS.md`.

```
User: "Analyze implementation scope"
-> Read all SPECS/*.md
-> Survey all implementation directories
-> Generate ANALYSIS.md with all categories
```

### Single Spec

Focus on one spec, analyze corresponding implementation area.

```
User: "Analyze implementation for SPECS/02-token-system.md"
-> Read that spec
-> Survey relevant modules
-> Generate ANALYSIS.md with focused categories
```

### Incremental

Update existing `ANALYSIS.md` after specs change or scope shifts.

```
User: "Update implementation analysis — SPECS/30 was added"
-> Read existing ANALYSIS.md
-> Read new/changed specs
-> Update ANALYSIS.md with revised scope
```

## Remember

- This skill **analyzes and plans** — it does not generate tasks or write code
- `ANALYSIS.md` is a strategic document for human review, not a task list
- Keep categories focused: 3-8 items per category
- Integration tests are first-class citizens, not afterthoughts
- Prioritize by dependency order: foundations -> domain -> interface -> advanced
  -> integration
- Always trace categories back to specific SPECS files
- After identifying areas, proceed directly to generate ANALYSIS.md

## Common Mistakes

- **Too granular**: Listing individual functions or files instead of components
  and modules. ANALYSIS.md is strategic, not tactical.
- **Skipping existing code survey**: Failing to check what already exists leads
  to redundant categories or missed reuse opportunities.
- **Mixing Stage 3 and Stage 4**: This skill plans fresh builds. If the user
  already has partial implementation and wants to find gaps, redirect to
  `spec-gap-analyzing`.
- **Missing integration tests**: Every multi-module system needs an IT category.
  Do not omit it.
- **Vague priorities**: Every priority must have a rationale tied to
  dependencies or business value.

## After Writing

After generating `ANALYSIS.md`:

1. **Human review** (recommended): Team reviews scope, priorities, and structure
2. **Generate tasks**: Convert `ANALYSIS.md` into actionable `TODO.yaml`
3. **Execute tasks**: Follow `TODO.yaml` to implement features

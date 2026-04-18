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

**Library side (mandatory):** consult `.agents/skills/dependency-selecting`
before proposing any from-scratch module. Preference order:
1. Libraries recommended by `.agents/skills/dependency-selecting`
2. Libraries already tracked under `.references/` — only when the recommended
   set has no fit
3. Other external libraries — require explicit justification (fit, license,
   weight, maintenance)
4. From-scratch build — last resort

For each spec area, record: candidate library, fit verdict, one-line rejection
reason when skipped. Feeds Step 2 (Design Review) and Step 3 (Scope Reduction).

**Output:** a mental model of "what SPECS require to be built" vs "what already
exists in code" vs "what a library already owns and must not be reimplemented."

### Step 2: Design Review — Challenge the Specs

Before planning implementation, critically evaluate the specifications for design
problems. Specs are written by humans and may contain:

- **Contradictions** — a spec principle (e.g., "streaming-first") violated by its
  own interface design (e.g., a batch-only function signature)
- **Over-engineering** — external libraries or complex patterns for trivially
  simple problems (e.g., a state machine library for 5 states with 5 transitions)
- **Abstraction mismatches** — custom interfaces that duplicate standard library
  capabilities with no added value
- **Broken symmetry** — most of the API uses one pattern (e.g., `io.Reader/Writer`)
  but one part uses a different pattern (e.g., file paths) without justification
- **Truth gaps** — factories, registries, or constructors that return instances which
  cannot fulfill their advertised contract. A factory that succeeds must mean the
  instance works. If `NewDecoder(codec)` returns a decoder, calling `Decode()` must
  not fail with "not implemented." Registration-time capability must equal runtime
  capability. Common manifestations:
  - Side-effect registration of stub implementations that fail at call time
  - Registry entries that conflate "probe-only" with "full capability"
  - Constructors that succeed but produce objects with degraded behavior
  - Capability queries (`Available*`, `Supports*`) that report theoretical
    compatibility instead of actual runtime truth
- **Missing API surface** — types that lack essential methods (e.g., enums without
  `String()`) making debugging painful
- **Reinvented upstream** — a spec building from scratch what Step 1's library
  survey showed is already owned upstream. Flag explicitly; require either
  adoption or concrete reimplementation justification (license, weight, fit).
  "We didn't know it existed" is not a justification.

**How to review:**

1. For each spec contract, ask: *Does this interface signature actually support
   the design goals stated in the spec's own rationale section?*
2. For each external dependency, ask: *Can this be done in ≤30 lines of
   standard code? If yes, the dependency is unjustified.*
3. For each custom abstraction, ask: *Does the standard library already provide
   this? What does the custom version add?*
4. For each pattern choice, ask: *Is this consistent with how the rest of the
   system works?*
5. For each factory/registry/constructor, ask: *If this returns success, can the
   caller use the result for its full advertised purpose — or will it fail later
   with "not available"?* A camera button must take photos, not open Settings.
6. For each from-scratch subsystem, ask: *Did Step 1's library survey find an
   upstream that already owns this?* If yes, default to adoption — any
   reimplementation needs explicit justification (license, weight, fit).

**Output format in ANALYSIS.md:**

Include a `## Design Review` section listing issues by severity:
- ⚠️ **Severe** — architectural contradiction, will cause runtime problems
- ⚠️ **Moderate** — unnecessary complexity, adds maintenance burden
- ⚠️ **Minor** — inconsistency or missing convenience, easy to fix

For each issue, provide:
- **What the spec says** (quote the problematic design)
- **Why it's a problem** (concrete consequence, with numbers when possible)
- **Suggested alternative** (with code sketch)
- **If not changed** (what happens if the user decides to keep the current design)

End the Design Review with explicit **decision questions** the user must answer
before task generation can proceed. Do not guess — let the user decide.

### Step 3: Analyze Implementation Scope

For each spec, determine what needs to be built:

| Scope Type | Description | Example |
|------------|-------------|---------|
| **New module** | Nothing exists, build from scratch | Spec defines auth system, no auth code exists |
| **New component** | Module exists, component does not | Spec adds validation to existing data module |
| **Extension** | Component exists, needs new capabilities | Spec requires 3 more output formats for existing transformer |
| **Foundation** | Shared infrastructure needed by multiple specs | Config system, error types, shared interfaces |

Focus on *what needs building*, not on gaps in existing code. This is Stage 3:
the assumption is that implementation is largely ahead of you, not behind you.

**Design philosophy — think like a founder, not a contractor:**

Say no to a thousand things. The best architecture has the fewest moving parts
that fully solve the problem. Plan for a **complete, production-quality system**
— not an MVP, not a v1-then-v2 roadmap. Ship once, ship right.

**The Apple principle:** Every product Apple ships is the *only version they
planned*. There is no "temporary product." The first release is designed as if
it will be used for ten years. This means: obsess over the core experience,
ruthlessly cut everything that doesn't serve it, and polish what remains until
it's seamless. Do not split implementation into phases or milestones — build
the complete system as one cohesive unit.

- **Simplicity is the ultimate sophistication** — aggressively challenge category count. If two categories share types or logic, merge them. If a category adds abstraction layers without clear runtime value, eliminate it. Software bloat starts at the analysis stage. Fewer categories with clear boundaries beat many fine-grained ones
- **No over-engineering** — don't plan for plugin systems, strategy patterns, or middleware chains unless the specs explicitly require them. One concrete implementation beats three layers of indirection. "Flexible" architectures that nobody asked for are the leading cause of unmaintainable code
- **Extensible through simplicity** — extensibility comes from clean, minimal interfaces — not from extra layers. A single well-designed interface is worth more than three "just in case" abstractions
- **System coherence** — every category must describe how it integrates with other categories. Isolated modules create integration debt
- **No shortcuts, no phases** — do not mark anything as "can be simplified later", "MVP first, improve later", or "v2 scope". The first implementation is the final implementation. If something isn't worth building right, it isn't worth building at all
- **Scope reduction over scope management** — the goal is not to manage a large scope well, but to *reduce* scope until only the essential remains. Before adding a category, ask: "If we never build this, does the product still work?" If yes, cut it. A product with 5 flawless features beats one with 15 mediocre ones

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

### Step 3: Scope Reduction — What NOT to Build

Before categorizing, actively eliminate unnecessary scope. This is the most
important step. Great products are defined by what they leave out.

**Elimination criteria — cut if ANY apply:**

1. **"Nice to have" test** — would a user notice if this didn't exist on day one?
   If not, cut it entirely (not "defer" — cut).
2. **Configuration over code** — can this be a config value instead of a system?
   A hardcoded default that covers 90% of cases beats a configurable system.
3. **Standard library test** — does the language's stdlib already handle this?
   Don't wrap `http.Client` in a custom HTTP layer. Don't abstract `slog` behind
   a logging interface.
4. **Upstream library test** — does a library recommended by
   `.agents/skills/dependency-selecting` already own this? Preference order:
   recommended → `.references/` entries (only when recommended has no fit) →
   other external (with explicit justification) → build from scratch.
5. **Single-caller test** — is this abstraction used by only one consumer? Then
   inline it. Abstractions earn their existence through multiple callers.
6. **Future-proofing test** — is this being built because someone *might* need it?
   Delete it. YAGNI is not laziness; it's discipline.

**Output:** A shorter list of what to build. Document what was cut and why in the
`## Notes` section of ANALYSIS.md — this prevents re-litigating decisions later.

### Step 4: Categorize & Prioritize

Group implementation work into categories by functional domain. Each category
becomes a section in `ANALYSIS.md`.

**Categorization principles:**
- Group by functional domain (e.g., Data Model, Auth, API Layer, CLI)
- Keep categories focused (3-8 items per category)
- Separate foundational work from feature work
- Include integration tests as a dedicated category
- **Aim for fewer categories** — 4-6 categories for a typical project. If you
  have more than 8, you're likely too granular or haven't reduced scope enough

**Prioritization:**
1. **Foundations first** (config, shared types, core interfaces) — everything
   depends on these
2. **Domain logic second** (business rules, data processing) — the main value
3. **Interface layer third** (CLI, API, UI) — depends on domain logic
4. **Advanced systems fourth** (optimization, caching) — independent enhancements
5. **Integration tests last** — validates the whole stack

### Step 5: Generate ANALYSIS.md

Write `ANALYSIS.md` at project root using the template below.

**Validation before writing:**
- Confirm file location: project root
- Confirm all sections present: Context, Design Review, Structure, Priorities,
  Dependencies, Notes
- Confirm Design Review includes at least one finding (if specs are perfect,
  explicitly state "No design issues found" — but this is rare)
- Confirm Design Review ends with decision questions for the user
- Confirm item counts are consistent across sections
- Confirm clear traceability: SPECS -> categories -> target modules

### Step 6: Prompt for Review

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

## Design Review

{List design problems found in specs, ordered by severity.
For each issue, include: current design, why it's problematic,
suggested alternative, consequence of not changing.
End with numbered decision questions for the user.}

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

## Scope Reductions

{Items explicitly cut from scope and why. This prevents re-litigating.}

- {Cut item 1} — {reason: e.g., "single caller, inlined instead"}
- {Cut item 2} — {reason: e.g., "stdlib handles this, no wrapper needed"}

## Notes

{Overall direction, risks, architectural decisions, recommendations}
```

## Inputs

Required:
- `SPECS/` directory with approved specification documents
- Implementation source directories (to assess what already exists)
- `.agents/skills/dependency-selecting` — consulted in Step 1 before proposing any from-scratch module

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

- This skill **analyzes, reviews, and plans** — it does not generate tasks or write code
- **Library survey is mandatory** — Step 1 must consult `.agents/skills/dependency-selecting` before any from-scratch proposal. Preference: recommended → `.references/` → other external (justified) → build.
- **Design review is mandatory** — never skip Step 2. Specs are not sacred;
  implementation analysis must challenge design problems before committing to
  building them. A bad design implemented perfectly is still a bad product.
- **Scope reduction is mandatory** — never skip Step 3. The default instinct is
  to build everything the spec describes. Fight it. Cut aggressively.
- **No phased implementation** — do not split work into "v1" and "v2". Do not
  create "MVP" categories. Build the complete product as one unit.
- `ANALYSIS.md` is a strategic document for human review, not a task list
- Keep categories focused: 3-8 items per category, aim for 4-6 total categories
- Integration tests are first-class citizens, not afterthoughts
- Prioritize by dependency order: foundations -> domain -> interface -> advanced
  -> integration
- Always trace categories back to specific SPECS files
- After identifying areas, proceed directly to generate ANALYSIS.md
- End with explicit decision questions — do not assume the user agrees with
  your design critique; let them decide

## Common Mistakes

- **Too granular**: Listing individual functions or files instead of components
  and modules. ANALYSIS.md is strategic, not tactical.
- **Skipping existing code survey**: Failing to check what already exists leads
  to redundant categories or missed reuse opportunities.
- **Skipping the library survey**: Proposing a from-scratch module without consulting `.agents/skills/dependency-selecting`. Recommended libraries come first; `.references/` is a fallback; external libraries need explicit justification.
- **Skipping design review**: Blindly implementing specs without questioning
  their design leads to architectural debt that is expensive to fix later.
  The implementation analysis phase is the last cheap opportunity to catch
  design problems before they become code.
- **Skipping scope reduction**: Building everything the spec mentions without
  questioning whether each piece is essential. The result is a bloated system
  where no single part is truly excellent.
- **Phase splitting**: Creating "Phase 1: Core" and "Phase 2: Advanced"
  categories. This is MVP thinking in disguise. Either build it right or don't
  build it. There is no "we'll make it good later."
- **Mixing Stage 3 and Stage 4**: This skill plans fresh builds. If the user
  already has partial implementation and wants to find gaps, redirect to
  `spec-gap-analyzing`.
- **Missing integration tests**: Every multi-module system needs an IT category.
  Do not omit it.
- **Vague priorities**: Every priority must have a rationale tied to
  dependencies or business value.
- **Being too polite about design problems**: If a spec contradicts its own
  stated principles, say so directly. Concrete numbers (memory usage, line
  count, dependency count) are more persuasive than abstract concerns.
- **Too many categories**: If you have 10+ categories, you haven't reduced
  scope — you've just organized complexity. Step back and merge or cut.

## After Writing

After generating `ANALYSIS.md`:

1. **Human review** (recommended): Team reviews scope, priorities, and structure
2. **Generate tasks**: Convert `ANALYSIS.md` into actionable `TODO.yaml`
3. **Execute tasks**: Follow `TODO.yaml` to implement features

---
description: Analyze research reports and generate root ANALYSIS.md defining specification scope. Use when planning specification structure, analyzing research findings, or organizing spec topics.
name: spec-analyzing
---


# Spec Analyzing

Analyzes research reports and generates a root-level `ANALYSIS.md` that defines the scope, priorities, and structure for specification writing.

**Announce at start:** "I'm using the spec-analyzing skill."

## Pipeline Position

```
.research/*.md → spec-analyzing → ANALYSIS.md (root) → spec-tasking → TODO.yaml
(research reports)                     ↓
                                  (optional review)
```

This skill is the **first step** in the spec writing workflow. It produces `ANALYSIS.md` which `spec-tasking` uses to generate `TODO.yaml`.

## Core Principle: Philosophy vs SPEC

Before proposing any spec, classify each finding:

| Type | Purpose | Tone | Content |
|------|---------|------|---------|
| **Philosophy** | Guide decision direction | "prefer X over Y" | Design principles, core thesis, trade-off rationale |
| **SPEC** | Define enforceable rules | "MUST / MUST NOT" | System contracts, constraints, acceptance criteria |

**Key test**: Can it be violated? If yes → SPEC. If it's a guiding principle → Philosophy.

- Philosophy says "**why** we choose X over Y"
- SPEC says "**what** the rules of X are"
- README says "**how** to use X"

When analyzing research findings:
- Extract **principles** (guiding, advisory) → note them as philosophy-level
- Extract **contracts** (enforceable, violable) → these become SPEC items
- Don't mix them — a single research report may produce both philosophy updates and spec items

## What This Skill Does

Reads all research reports in `.research/` and produces a structured analysis document at the project root that answers:

1. **Scope** — Which specifications need to be written based on research findings
2. **Priority** — Execution order for spec writing (high/medium/low or numbered)
3. **Structure** — Proposed SPEC/ file structure with numbering and naming

## Output: ANALYSIS.md

The generated `ANALYSIS.md` must be placed at **project root** (not in `.research/` or `SPEC/`).

### Required Sections

```markdown
# Spec Analysis

## Context
**Previous Stage**: `.research/*.md` (research reports)
**Current Goal**: Write specifications defining system contracts and design decisions
**Scope**: {scope boundary}
**Existing SPEC**: {list existing spec files — avoid re-specifying what's already covered}

## Structure

### SPEC/{NN}: {Spec Name}
**Focus**: {what this spec covers — violable rules and contracts, not principles}
**Type**: {which domain section patterns apply — see Pattern Types below}
**From**: `.research/{research-file}.md`
**To**: `SPEC/{NN}-{spec-name}.md`
**Priority**: High|Medium|Low — {rationale}
**Items**: N sections — {brief list of violable constraints}

---

## Priorities
1. **SPEC/{NN}** — {rationale}

## Dependencies
- SPEC/{NN} → SPEC/{MM} — {dependency description}

## SSOT Map
{Which concept is defined in which spec — ensures no concept is defined in two places}

## Philosophy Updates
{Principles or trade-offs from research that should update existing philosophy docs, not create new specs}

## Notes
{overall direction, risks, suggestions}
```

### Pattern Types

Each proposed spec should indicate which domain section patterns it will primarily use (from spec-writing guides):

| Pattern | Use For | Example |
|---------|---------|---------|
| **Schema Definition** | Data formats, field definitions | Token format, YAML schema |
| **Rule List** | Naming rules, coding conventions | Event naming, slot conventions |
| **Decision Table** | Multiple related decisions | FSM vs Store matrix, Shadow DOM strategy |
| **Interface Contract** | API contracts, interface definitions | Controller base class, State machine API |
| **State Machine / Flow** | Processes, state machines | Component lifecycle, animation states |

### Example Output

```markdown
# Spec Analysis

## Context

**Previous Stage**: `.research/*.md` (3 research reports: R01, R02, R03)
**Current Goal**: Write specifications for design token system to guide implementation
**Scope**: Token system, component system, and configuration specifications
**Existing SPEC**: `SPEC/philosophy.md` (architecture philosophy, covers core thesis and principles)

## Structure

### SPEC/02: Token System Specification

**Focus**: Token naming rules, layer hierarchy constraints, transformation pipeline contracts

**Type**: Schema Definition + Rule List

**From**: `.research/R01-token-system.md`

**To**: `SPEC/02-token-system.md`

**Priority**: High — Foundation specification, blocks component and config work

**Items**: 4 sections — naming pattern (violable: 3-segment rule), layer hierarchy (violable: Primitive→Semantic→Component), transformation rules (violable: input/output contracts), forbidden patterns

---

### SPEC/30: Component System Specification

**Focus**: Component interface contracts, lifecycle constraints, composition rules

**Type**: Interface Contract + Rule List

**From**: `.research/R02-component-system.md`

**To**: `SPEC/30-component-system.md`

**Priority**: Medium — Core architecture, depends on SPEC/02 token types

**Items**: 3 sections — interface contracts (violable: method signatures), lifecycle rules (violable: hook ordering), composition constraints (violable: max inheritance depth)

---

### SPEC/50: Configuration Management Specification

**Focus**: Config schema, validation contracts, environment handling rules

**Type**: Schema Definition + Decision Table

**From**: `.research/R03-configuration.md`

**To**: `SPEC/50-configuration-management.md`

**Priority**: Low — Supporting infrastructure, independent of other specs

**Items**: 3 sections — config schema (violable: required fields), validation rules (violable: type constraints), environment rules (violable: override precedence)

## Priorities

1. **SPEC/02** — Foundation for design system, blocks component work
2. **SPEC/30** — Core architecture, depends on SPEC/02 type definitions
3. **SPEC/50** — Supporting infrastructure, can be written anytime

## Dependencies

- SPEC/30 → SPEC/02 — Component spec needs token type definitions from SPEC/02
- SPEC/50 is independent

## SSOT Map

| Concept | Defined In | Referenced By |
|---------|-----------|---------------|
| Token naming | SPEC/02 | SPEC/30 |
| Component interface | SPEC/30 | — |
| Config schema | SPEC/50 | — |

## Philosophy Updates

- R01 suggests "tokens are immutable design decisions" — update `SPEC/philosophy.md` §Principles
- R02's "composition over inheritance" already covered in philosophy.md, no update needed

## Notes

- SPEC/02 must be completed first to establish type vocabulary
- SPEC/30 and SPEC/50 can be written in parallel after SPEC/02
```

## Workflow

### Step 1: Scan Research Reports and Existing Specs

Read all files in `.research/` directory and existing `SPEC/` files:

For each report, extract:
- Main topic/domain
- **Violable constraints** (→ SPEC items): naming rules, interface contracts, forbidden patterns
- **Guiding principles** (→ Philosophy updates): design rationale, trade-off direction
- Technical decisions with `> **Why**` and `> **Rejected**` potential

For existing specs, identify:
- Concepts already defined (SSOT — don't re-define)
- Gaps or updates needed

### Step 2: Apply Content Boundary Test

For each proposed spec item, apply the **violation test**:

- ✅ "Token names must follow `{category}.{concept}.{variant}`" → violable → SPEC
- ✅ "Controller must call dispose() in hostDisconnected" → violable → SPEC
- ❌ "We use Lit 3.x for Web Components" → fact → not SPEC (belongs in README/CLAUDE.md)
- ❌ "Composition is preferred over inheritance" → principle → Philosophy

Filter out non-SPEC content before proposing spec structure.

### Step 3: Identify Specification Needs

Group research findings into logical specification documents:

- **Consolidate related topics** — Multiple research reports may inform a single spec
- **Separate concerns** — Distinct domains should have separate specs (high cohesion)
- **Respect SSOT** — Each concept defined in exactly ONE spec
- **Consider dependencies** — Identify which specs depend on others
- **Size limit** — 单个 SPEC 文件不超过 1000 行

### Step 4: Determine Priority

Order specifications by:

1. **Foundation first** — Core contracts that other specs reference
2. **Dependencies** — Specs that other specs depend on
3. **Validation value** — Complex specs that validate architecture decisions early
4. **Implementation order** — What needs to be built first

### Step 5: Design File Structure

Create a logical numbering scheme for SPEC/:

- Use gaps (02, 05, 10 or 10, 20, 30) to allow insertions
- Group related specs by domain range (e.g., 02-08 core, 10-19 integration, 20-29 platform)
- Use descriptive kebab-case names
- Keep filenames concise but clear

### Step 6: Build SSOT Map

Create a concept → spec mapping to ensure:
- No concept is defined in two specs
- Every concept has exactly one authoritative source
- Cross-references are explicit (`[SPEC/02 §Token Naming]` format)

### Step 7: Write ANALYSIS.md

Generate the analysis document at project root with all required sections.

## Each Proposed Spec Must Have

Every spec entry in the analysis must describe content that follows the standard spec structure:

```
# <Topic>
## Overview        — scope, boundaries, related specs
## <Domain>        — rules using one of 5 patterns (Schema/Rule/Decision/Interface/Flow)
## Terminology     — terms introduced or specialized
## Forbidden       — what NOT to do + alternatives
## References      — related specs and standards
```

When listing Items for a proposed spec, frame them as **violable rules**, not descriptions:
- ✅ "naming pattern (3-segment `{a}.{b}.{c}` rule)" — can be violated
- ❌ "overview of token system" — not a rule

## Inputs

Required:
- `.research/` directory with at least one research report (`.research/*.md`)

Optional:
- Existing `SPEC/` directory to understand current state and avoid re-specification
- `DECISIONS.md` or similar decision documents for context
- Philosophy documents to identify what's already covered at principle level

## Remember

- `ANALYSIS.md` goes at **project root**, not in subdirectories
- This is an **analysis document**, not a plan or task list
- Focus on **what** needs to be specified, not **how** to write it
- Keep scope realistic — don't commit to specs without research backing
- Priority should reflect actual dependencies and implementation order
- File structure should be logical and allow for future expansion
- **Distinguish philosophy from spec** — principles guide, specs constrain
- **SSOT is non-negotiable** — one concept, one definition, one location
- **Every spec item must be violable** — if it can't be violated, it doesn't belong in a spec
- **Size limit** — 单个 SPEC 文件不超过 1000 行

## Common Mistakes

**❌ Mixing philosophy with spec scope**
- "Composition over inheritance" is a philosophy principle, not a spec
- Spec would define: "maximum 2 levels of class inheritance" (violable rule)

**❌ Generating tasks instead of analysis**
- ANALYSIS.md defines scope/priority/structure
- TODO.yaml (generated by spec-tasking) contains tasks

**❌ Placing ANALYSIS.md in wrong location**
- Must be at project root
- Not in `.research/`, `SPEC/`, or `.plans/`

**❌ Copying research content verbatim**
- ANALYSIS.md is a synthesis, not a summary
- Focus on what specs to write, not research details

**❌ Ignoring existing SPEC/ files**
- Check what specs already exist
- Only include gaps or updates in scope
- Respect SSOT — don't re-define concepts already specified

**❌ Violating SSOT — same concept in multiple specs**
- Build the SSOT Map to catch this
- Each concept has exactly one authoritative spec

**❌ Proposing specs with non-violable content**
- Apply the violation test to every proposed item
- Facts, tutorials, and process docs don't belong in specs

**❌ Vague priorities**
- Use concrete ordering (1, 2, 3 or High/Medium/Low)
- Explain rationale for priority decisions


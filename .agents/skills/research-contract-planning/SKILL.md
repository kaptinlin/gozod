---
name: research-contract-planning
description: Generates contract-only PLAN.md and TODO.yaml from a project’s .research and .references to guide design documentation (concepts, interfaces, contracts, and design philosophy). Use when users want design planning docs, not implementation plans.
---

# Research to Contract Planning

Create `PLAN.md` and `TODO.yaml` for design-document production only. This skill defines boundaries, architecture contracts, quality rules, and sequencing without implementation work.

## Inputs

Required:
- Project root path
- `.research` directory
- `.references` directory

Optional:
- Existing `PLAN.md` / `TODO.yaml` to update
- Naming preferences for design topics

## Step 1: Read Research Sources

Read relevant files in `.research` (for example: goals, constraints, principles, architecture notes). Extract:
- Objectives and non-goals
- Scope boundaries
- Core modules/concerns
- Quality and engineering principles

## Step 2: Read References for Grounding

Inspect `.references` and gather implementation patterns at a high level.
- Use at least two relevant reference directories per major design area
- Record which reference directories were consulted
- Keep the output technology-agnostic when required by the user

## Step 3: Produce PLAN.md (Contract-Only)

`PLAN.md` must define design-document workstreams, not coding tasks.

Required sections:
1. Goals and boundaries
2. Design philosophy and quality principles (`KISS`, `DRY`, `YAGNI`, `SOLID`, SRP, OCP)
3. Topic-based design document tree (e.g., `DESIGNS/<topic>.md`)
4. Per-topic key questions and expected contract outputs
5. Sequencing and milestones
6. Reference usage policy (`.references` must be cited in each topic)
7. Definition of done for documentation phase

Hard constraints:
- No implementation plan details
- No API/server/admin/storage implementation specifics
- Focus on concepts, interfaces, contracts, boundaries, and quality requirements

## Step 4: Produce TODO.yaml (Design Tasks Only)

Generate tasks that create/update design docs only.

Format:
```yaml
tasks:
  - title: "<action> <contract doc> in <path> per PLAN.md section X"
    completed: false
    parallel_group: N
    description: "Contract-only output; cite consulted .references paths; no code implementation"
```

Task rules:
- Unique, traceable titles
- Each task maps to a concrete design doc path
- Each task cites PLAN source section
- Parallel groups reflect dependency order

## Step 5: Verify

Before finalizing, check:
- `PLAN.md` and `TODO.yaml` are consistent
- Every task is contract-only and design-doc scoped
- `.references` citation requirement is explicit
- No implementation tasks are present

---
description: Converts an ANALYSIS.md into a structured TODO.yaml with paired Plan and Impl tasks for TDD Code phase execution. Use when tdd-analyzing has produced ANALYSIS.md and implementation tasks need to be generated.
name: tdd-tasking
---


# TDD Tasking

Read a selected source document and generate a structured `TODO.yaml` for Code phase execution.

Primary source types:
- `ANALYSIS.md` from `tdd-analyzing` (**analysis mode**)
- implementation-oriented docs such as `improve.md`, `REFACTOR.md`, or similar plan docs (**doc mode**)

The default output is still TDD-oriented: **Plan** tasks, **Impl** tasks, and optional **Refactor Review** checkpoints. However, **minor documentation-only tweaks do not use Plan / TDD** — emit them as direct tasks.

**Used in:** Code phase (workflow stage 3) - Tasking step

**Announce at start:** "I'm using the tdd-tasking skill."

## Mode Selection

Choose the source and mode before generating tasks.

### Analysis mode

Use when:
- `ANALYSIS.md` exists at project root and the user asks to task the analysis
- the request is the normal `tdd-analyzing -> tdd-tasking` workflow
- the user does not explicitly point to another source doc

### Doc mode

Use when:
- the user explicitly points to `improve.md`, `REFACTOR.md`, or another implementation-oriented source doc
- the source doc contains implementation work items but is not an `ANALYSIS.md`

### Source precedence

1. If the user explicitly names a source doc, use it.
2. Otherwise, prefer `ANALYSIS.md` when present.
3. If no source is clear, ask which document should drive task generation.

## Pipeline Position

```text
Code Phase Workflow:

Primary path:
SPECS/*.md + existing code
    ↓
tdd-analyzing → ANALYSIS.md → tdd-tasking → TODO.yaml

Alternate path:
improve.md / REFACTOR.md / implementation plan doc
    ↓
tdd-tasking → TODO.yaml
```

## Task Structure

### Default rule

For code-bearing implementation items, generate exactly **two tasks**:
1. a **Plan** task using `tdd-planning`
2. an **Impl** task using `tdd-implementing`

### Exception rule

If an item is a **minor documentation-only tweak**, do **not** generate Plan + Impl. Emit a direct task instead.

Examples of documentation-only tweaks:
- adjust README wording
- align examples with current behavior
- clarify a spec paragraph
- rename headings / reorganize doc sections

Do **not** use the docs exception for:
- code changes with accompanying docs
- behavior changes that require tests
- spec changes that imply implementation work

## 1. Plan Task

Generates a detailed implementation plan via `tdd-planning` skill.

### Analysis mode template

```text
Plan ALL {feature} implementation based on ANALYSIS.md, use tdd-planning skill — spec: SPECS/{spec-file} | research: .research/{research-file} — generate .plans/{date}-{category-slug}-{review}-{slug}.md
```

### Doc mode template

```text
Plan ALL {feature} implementation based on {source-doc} section "{section}", use tdd-planning skill — source: {source-doc}{evidence-clause} — generate .plans/{date}-{category-slug}-{review}-{slug}.md
```

### Rules

- `{feature}` = concise human-readable work item name
- `{date}` = today's date in YYYY-MM-DD format
- `{category-slug}` = category/item ID when available (e.g. `A1`, `B3`, `IT2`); if the source doc has no IDs, derive a short stable slug
- `{slug}` = kebab-case of the feature name
- The `.plans/{date}-{category-slug}-{review}-{slug}.md` path must match exactly between Plan and Impl tasks
- Plan titles must say **"ALL"** contracts — do not narrow the contract prematurely
- Keep title to one line

### Evidence rules

- **Analysis mode:** include both `spec:` and `research:` when they are provided by `ANALYSIS.md`
- **Doc mode:** `source:` is mandatory; include `spec:` and/or `research:` only if the source doc explicitly provides them or the user named them
- Use plain paths only — no markdown links

## 2. Impl Task

Implements the generated plan using `tdd-implementing` skill.

### Analysis mode template

```text
Implement {feature} in {target-modules} based on ANALYSIS.md, use tdd-implementing skill — spec: SPECS/{spec-file} | research: .research/{research-file} — follow .plans/{date}-{category-slug}-{review}-{slug}.md
```

### Doc mode template

```text
Implement {feature} in {target-modules} based on {source-doc} section "{section}", use tdd-implementing skill — source: {source-doc}{evidence-clause} — follow .plans/{date}-{category-slug}-{review}-{slug}.md
```

### Rules

- The `.plans/` path must be identical to the corresponding Plan task
- `{target-modules}` = target packages, files, or modules inferred from the source doc
- Use `tdd-implementing` only for code-bearing items
- Keep title to one line

## 3. Direct Documentation Task

Use this for **minor documentation-only tweaks**.

### Template

```text
Update {target-docs} directly based on {source-doc} section "{section}" — documentation-only change; no plan or tdd needed
```

### Rules

- Use only when the item is clearly documentation-only
- Prefer specific target docs such as `README.md`, `SPECS/10-core-types.md`, `docs/architecture.md`
- Keep the scope narrow and executable in one focused pass
- Do not force `tdd-planning` or `tdd-implementing` for these items

## 4. Refactor Review (every 4 code categories)

After every 4 **code-bearing** categories of Plan+Impl tasks, insert a **4-task review checkpoint**.

**Task 1 — Dedup Audit:**
```text
Review {package-list} for duplication issues based on {source-ref} and use code-deduplicating skill to generate .plans/{date}-{category-slug}-{review}-dedup-review.md
```

**Task 2 — Refactor Audit:**
```text
Review {package-list} for refactoring opportunities based on {source-ref} and use code-refactoring skill to generate .plans/{date}-{category-slug}-{review}-quality-review.md
```

**Task 3 — Simplify Audit:**
```text
Review {package-list} for simplification opportunities based on {source-ref} and use code-simplifying skill to generate .plans/{date}-{category-slug}-{review}-simplify-review.md
```

**Task 4 — Fix:**
```text
Review .plans/{date}-{category-slug}-{review}-dedup-review.md, .plans/{date}-{category-slug}-{review}-quality-review.md, and .plans/{date}-{category-slug}-{review}-simplify-review.md based on {source-ref}, consolidate suggestions by priority, implement approved improvements
```

### Rules

- Count only code-bearing categories toward the review interval
- Skip doc-only items when computing review checkpoints
- `package-list` = union of code-touching packages/files from the preceding 4 code categories
- `source-ref` = `ANALYSIS.md` in analysis mode, or the selected source doc in doc mode
- If the generated TODO contains no code-bearing items, skip refactor reviews entirely

## 5. Code Restructuring (after all code tasks)

After all code-bearing Plan+Impl tasks and Refactor Reviews are complete, insert a **2-task global architecture checkpoint**.

**Task 1 — Generate:**
```text
Review entire repository for architectural issues based on {source-ref} and use code-refactoring skill to generate .plans/{date}-GC1-global-cohesion.md
```

**Task 2 — Execute:**
```text
Review and implement approved global refactorings based on {source-ref} and use code-refactoring skill to implement .plans/{date}-GC1-global-cohesion.md
```

### Rules

- Add these only if the TODO contains code-bearing work
- Skip them for fully doc-only task sets

## TODO.yaml Format

```yaml
tasks:
  # ═══════════════════════════════════════════════════════════════
  # Category X: {Name}
  # ═══════════════════════════════════════════════════════════════

  - title: "Plan ALL {feature} implementation based on ANALYSIS.md, use tdd-planning skill — spec: SPECS/{spec-file} | research: .research/{research-file} — generate .plans/{date}-{category-slug}-{review}-{slug}.md"
    completed: false

  - title: "Implement {feature} in {target-modules} based on ANALYSIS.md, use tdd-implementing skill — spec: SPECS/{spec-file} | research: .research/{research-file} — follow .plans/{date}-{category-slug}-{review}-{slug}.md"
    completed: false

  - title: "Update README.md directly based on improve.md section \"README alignment\" — documentation-only change; no plan or tdd needed"
    completed: false
```

### YAML rules

- Every task title must include **source traceability**
  - **Analysis mode:** use `based on ANALYSIS.md`
  - **Doc mode:** use `based on {source-doc} section "..."` or equivalent explicit source wording
- Use only two fields per task: `title` and `completed`
- Never add `description`
- Use YAML block comments (`#`) for category separators when helpful
- No blank lines between a task's fields
- Blank line between tasks for readability
- Code Restructuring always comes last, after all code-bearing tasks

## Workflow

### Step 1: Identify source and mode

Determine which document drives task generation:
- `ANALYSIS.md` -> analysis mode
- `improve.md` / `REFACTOR.md` / similar -> doc mode

If unclear, ask the user.

### Step 2: Read the source document

Extract:
- scope
- ordering / priority
- target packages or files
- any referenced `spec:` or `research:` evidence
- whether each item is code-bearing or documentation-only

In doc mode, common extraction shapes include:
- headings
- numbered lists
- checklists
- sections like Problems / Proposed Changes / Implementation / Docs

### Step 3: Generate tasks

For each extracted item:
- **code-bearing item** -> generate one Plan task + one Impl task
- **minor documentation-only tweak** -> generate one direct documentation task

If an item mixes code and docs:
- keep the code work in Plan + Impl
- split out the doc tweak into a separate direct doc task only if that makes execution clearer

### Step 4: Insert conditional review checkpoints

- After every 4 code-bearing categories, add one refactor review checkpoint
- If there is no code-bearing work, skip refactor reviews and global restructuring tasks

### Step 5: Write TODO.yaml

Write the final `TODO.yaml` at project root.

### Step 6: Announce completion

Inform user:

```text
TODO.yaml has been generated at project root with N tasks.

Next steps:
1. Review TODO.yaml (optional)
2. Execute tasks using Ralphy loop or manually
3. Code-bearing Plan tasks use tdd-planning
4. Code-bearing Impl tasks use tdd-implementing
5. Minor documentation-only tweaks can be edited directly
```

## Inputs

### Required

One source document:
- `ANALYSIS.md`, or
- a user-specified implementation-oriented source doc such as `improve.md`, `REFACTOR.md`, or similar

### Optional

- existing `TODO.yaml` to update
- custom refactor interval (default: every 4 code categories)
- date override for plan file naming (default: today)

## Examples

### Example 1: Standard analysis workflow

Source: `ANALYSIS.md`

```yaml
- title: "Plan ALL token validation implementation based on ANALYSIS.md, use tdd-planning skill — spec: SPECS/10-auth.md | research: .research/R01-auth.md — generate .plans/2026-04-05-A1-plan-token-validation.md"
  completed: false

- title: "Implement token validation in pkg/auth based on ANALYSIS.md, use tdd-implementing skill — spec: SPECS/10-auth.md | research: .research/R01-auth.md — follow .plans/2026-04-05-A1-plan-token-validation.md"
  completed: false
```

### Example 2: improve.md with code work

Source: `improve.md`

```yaml
- title: "Plan ALL shared validation hardening implementation based on improve.md section \"补强根包共享校验\", use tdd-planning skill — source: improve.md | spec: SPECS/10-core-types.md, SPECS/50-error-handling.md — generate .plans/2026-04-05-A1-plan-root-contract-validation.md"
  completed: false

- title: "Implement shared validation hardening in search.go and search_test.go based on improve.md section \"补强根包共享校验\", use tdd-implementing skill — source: improve.md | spec: SPECS/10-core-types.md, SPECS/50-error-handling.md — follow .plans/2026-04-05-A1-plan-root-contract-validation.md"
  completed: false
```

### Example 3: minor documentation-only tweak

```yaml
- title: "Update README.md directly based on improve.md section \"让 README 与真实行为重新对齐\" — documentation-only change; no plan or tdd needed"
  completed: false
```

## Remember

- This skill is still primarily a **Code phase tasking skill**
- Preserve the normal `tdd-analyzing -> ANALYSIS.md -> tdd-tasking -> TODO.yaml` workflow
- Support `improve.md` / `REFACTOR.md`-style implementation docs as an explicit secondary mode
- Plan titles must say **ALL** contracts
- Do not force minor documentation-only tweaks through Plan / TDD
- Use `tdd-planning` and `tdd-implementing` only for code-bearing tasks
- In analysis mode, include `spec:` and `research:` when `ANALYSIS.md` provides them
- In doc mode, `source:` is required; `spec:` and `research:` are optional unless explicitly available
- Use plain paths only — no markdown links
- Refactor reviews analyze code for redundancy and reuse opportunities
- Simplify tasks focus on recent code changes, not the entire codebase
- `.plans/` directory is `.gitignore`d — plan files are ephemeral working artifacts
- If the task set is fully documentation-only, skip code review and restructuring checkpoints
- This skill is used in Code phase, not Alignment phase (use `spec-gap-tasking` for Alignment)

---
description: Converts an ANALYSIS.md into a structured TODO.yaml with paired Plan and Impl tasks for TDD Code phase execution. Use when tdd-analyzing has produced ANALYSIS.md and implementation tasks need to be generated.
name: tdd-tasking
---


# TDD Tasking

Reads `ANALYSIS.md` (produced by `tdd-analyzing`) and generates a structured `TODO.yaml` with three task types: **Plan** tasks, **Impl** tasks, and **Refactor Review** checkpoints interleaved every 4 categories.

**Used in:** Code phase (workflow stage 3) - Tasking step

**Announce at start:** "I'm using the tdd-tasking skill."

**Prerequisite:** `ANALYSIS.md` must exist at project root. If not, run `tdd-analyzing` first.

## Pipeline Position

```
Code Phase Workflow:

tdd-analyzing ──▶ tdd-tasking ──▶ Ralphy loop
     │                 │
     ▼                 ▼
 ANALYSIS.md       TODO.yaml
(root directory) (root directory)
```

## Task Structure

Every implementation item in ANALYSIS.md produces exactly **two tasks** — a Plan task and an Impl task.

### 1. Plan Task

Generates a detailed implementation plan via `tdd-planning` skill.

**Template:**
```
Plan {feature} implementation, covering {input-reference}, and use tdd-planning skill to generate .plans/{date}-{category-slug}-{review}-{slug}.md
```

**Rules:**
- `{feature}` = human-readable feature name from ANALYSIS.md (e.g., "token system", "authentication")
- `{input-reference}` = source spec file (e.g., "SPECS/02-token-system.md")
- `{date}` = today's date in YYYY-MM-DD format
- `{category-slug}` = category letter + number from ANALYSIS.md (e.g., A1, B3, IT2)
- `{slug}` = kebab-case of the feature name
- The `.plans/{date}-{category-slug}-{review}-{slug}.md` path must match exactly between Plan and Impl tasks
- Keep title to **one line** — be concise

### 2. Impl Task

Implements the generated plan using `tdd-implementing` skill.

**Template:**
```
Implement {feature} in {target-modules}, covering {input-reference}, and use tdd-implementing skill to implement following .plans/{date}-{category-slug}-{review}-{slug}.md
```

**Rules:**
- The `.plans/` path must be identical to the corresponding Plan task
- `{target-modules}` = target packages from ANALYSIS.md (e.g., "pkg/token", "pkg/auth")
- Keep title to **one line** — be concise

### 3. Refactor Review (every 4 categories)

After every 4 categories of Plan+Impl tasks, insert a **4-task review checkpoint**:

**Task 1 — Dedup Audit:**
```
Review {package-list} for duplication issues and use code-deduplicating skill to generate .plans/{date}-{category-slug}-{review}-dedup-review.md
```

**Task 2 — Refactor Audit:**
```
Review {package-list} for refactoring opportunities and use code-refactoring skill to generate .plans/{date}-{category-slug}-{review}-quality-review.md
```

**Task 3 — Simplify Audit:**
```
Review {package-list} for simplification opportunities and use code-simplifying skill to generate .plans/{date}-{category-slug}-{review}-simplify-review.md
```

**Task 4 — Fix:**
```
Review .plans/{date}-{category-slug}-{review}-dedup-review.md, .plans/{date}-{category-slug}-{review}-quality-review.md, and .plans/{date}-{category-slug}-{review}-simplify-review.md, consolidate suggestions by priority, implement approved improvements
```

**Package list** = union of all packages touched by the preceding 4 categories.

### 4. Code Restructuring (after all tasks)

After all Plan+Impl tasks and Refactor Reviews are complete, insert a **2-task global architecture checkpoint**:

**Task 1 — Generate:**
```
Review entire repository for architectural issues and use code-refactoring skill to generate .plans/{date}-GC1-global-cohesion.md
```

**Task 2 — Execute:**
```
Review and implement approved global refactorings and use code-refactoring skill to implement .plans/{date}-GC1-global-cohesion.md
```

**Scope** = entire repository (all pkg/ and internal/ packages)

**Purpose:**
- Catch global architecture issues missed by incremental refactoring
- Identify mega-packages that need subdirectory organization
- Find duplicate code across distant packages
- Verify package placement (internal/ vs pkg/)
- Ensure consistent naming and structure across the codebase

## TODO.yaml Format

```yaml
tasks:
  # ═══════════════════════════════════════════════════════════════
  # Category X: {Name} (SPECS/NN)
  # ═══════════════════════════════════════════════════════════════

  - title: "Plan {feature} implementation, covering SPECS/{spec-file}, and use tdd-planning skill to generate .plans/{date}-{category-slug}-{review}-{slug}.md"
    completed: false

  - title: "Implement {feature} in {target-modules}, covering SPECS/{spec-file}, and use tdd-implementing skill to implement following .plans/{date}-{category-slug}-{review}-{slug}.md"
    completed: false

  # ... more plan+impl pairs ...

  # ═══════════════════════════════════════════════════════════════
  # Refactor Review #N — after Categories W/X/Y/Z
  # Scope: {package-list}
  # ═══════════════════════════════════════════════════════════════

  - title: "Review {package-list} for duplication issues and use code-deduplicating skill to generate .plans/{date}-{category-slug}-{review}-dedup-review.md"
    completed: false

  - title: "Review {package-list} for refactoring opportunities and use code-refactoring skill to generate .plans/{date}-{category-slug}-{review}-quality-review.md"
    completed: false

  - title: "Review {package-list} for simplification opportunities and use code-simplifying skill to generate .plans/{date}-{category-slug}-{review}-simplify-review.md"
    completed: false

  - title: "Review .plans/{date}-{category-slug}-{review}-dedup-review.md, .plans/{date}-{category-slug}-{review}-quality-review.md, and .plans/{date}-{category-slug}-{review}-simplify-review.md, consolidate suggestions by priority, implement approved improvements"
    completed: false

  # ═══════════════════════════════════════════════════════════════
  # Code Restructuring — after all implementation complete
  # Scope: entire repository (all pkg/ and internal/ packages)
  # ═══════════════════════════════════════════════════════════════

  - title: "Review entire repository for architectural issues and use code-refactoring skill to generate .plans/{date}-{category-slug}-GC1-global-cohesion.md"
    completed: false

  - title: "Review and implement approved global refactorings and use code-refactoring skill to implement .plans/{date}-{category-slug}-GC1-global-cohesion.md"
    completed: false
```

**YAML rules:**
- Two fields per task: `title` and `completed` (always `false`)
- Use YAML block comments (`#`) for category separators with `═` box lines
- No blank lines between a title and its fields
- Blank line between tasks for readability
- Code Restructuring always comes last, after all other tasks

## Workflow

### Step 1: Read ANALYSIS.md

Read `ANALYSIS.md` from the project root and extract:
- Scope: what needs to be implemented
- Priority: category ordering and importance
- Structure: category breakdown with IDs, names, source specs, and target packages

### Step 2: Determine Category Order

Follow the priority and structure from ANALYSIS.md to order categories. Typical ordering:
1. Core contracts (foundations, config, types)
2. Domain logic (business logic, data processing)
3. Interface layer (CLI, API, UI)
4. Advanced systems (optimization, caching)
5. Integration tests

### Step 3: Generate Task Pairs

For each category item in ANALYSIS.md, generate exactly one Plan task + one Impl task using the templates above.

### Step 4: Insert Refactor Reviews

Count categories in order. After every 4th category, insert a 4-task refactor review:
- R1 after categories 1–4
- R2 after categories 5–8
- R3 after categories 9–end

Compute the package scope for each review by collecting all packages from the preceding categories.

### Step 5: Insert Code Restructuring

After all categories and refactor reviews, append a 2-task global cohesion review:
- GC1 Generate: Use code-refactoring skill to generate `.plans/{date}-{category-slug}-GC1-global-cohesion.md`
- GC1 Execute: Review `.plans/{date}-{category-slug}-GC1-global-cohesion.md`, consolidate suggestions by priority, implement high-value improvements

This catches global architecture issues that incremental refactoring might miss.

### Step 6: Write TODO.yaml

Write the final `TODO.yaml` at project root.

### Step 7: Announce Completion

Inform user:
```
TODO.yaml has been generated at project root with N tasks.

Next steps:
1. Review TODO.yaml (optional)
2. Execute tasks using Ralphy loop or manually
3. Each Plan task uses tdd-planning to generate implementation plan
4. Each Impl task uses tdd-implementing to write code and tests

Ready to proceed with implementation?
```

## Inputs

Required:
- `ANALYSIS.md` — the implementation analysis report at project root (produced by `tdd-analyzing`)

Optional:
- Existing `TODO.yaml` to update (append new tasks)
- Custom refactor interval (default: every 4 categories)
- Date override for plan file naming (default: today)

## Remember

- This skill is the **second step** after `tdd-analyzing` — it does not analyze scope
- Plan titles must say "ALL contracts" — never constrain to specific focus points
- Refactor reviews analyze code for redundancy and reuse opportunities
- Simplify tasks focus on recent code changes, not the entire codebase
- `.plans/` directory is `.gitignore`d — plan files are ephemeral working artifacts
- Integration test items use same Plan+Impl structure as feature items
- This skill is used in Code phase, not Alignment phase (use `spec-gap-tasking` for Alignment)

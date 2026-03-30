---
description: Convert user-provided optimization requirements into a folder-level TODO.yaml with per-folder dedup/refactor/simplify/fix tasks and a final global cohesion checkpoint. Use after code-optimization-analyzing or when a user requests repository-wide optimization tasking.
name: code-optimization-tasking
---


# Code Optimization Tasking

Two-phase workflow: audit all folders against the user’s optimization requirements, then generate a folder-granularity `TODO.yaml` with Plan/Execute tasks and periodic quality checkpoints.

**Announce at start:** "I'm using the code-optimization-tasking skill."

**Prerequisite:** User provides optimization requirements (scope, constraints, focus areas). If unclear, ask for clarification before tasking.

## Phase 1: Audit Repository Folders

**Verify folder-level scope before generating tasks.**

### Audit Process

1. **Enumerate all folders**
   - Include every top-level and nested folder in the repo
   - Exclude: `.git`, `node_modules`, `vendor`, `dist/`, `build/`, `coverage/`, `.claude/`, `.plans/`, temp folders

2. **Record folder audit notes**
   - Summarize purpose, key files, and current optimization risks
   - Mark status: `optimize`, `skip`, or `third-party`
   - Mark complete items with ✅ (skip in task generation)

3. **Log audit findings**
   - Write or update `ANALYSIS.md` in the repo root with the folder inventory and annotations
   - Include all folders, even those marked `skip` or `third-party`

## Phase 2: Generate TODO.yaml

Generate folder-level optimization tasks for every folder marked `optimize`. Skip ✅ items.

## Pipeline Position

```
User Requirements ──▶ Phase 1: Folder Audit ──▶ Phase 2: Tasking
        │                      │                      │
        ▼                      ▼                      ▼
  User request                  TODO.yaml
  (conversation)               (root directory)
```

## Task Structure

Every folder marked `optimize` produces exactly **four tasks** — Dedup Plan, Refactor Plan, Simplify Plan, and Fix/Execute.

### 1. Dedup Plan Task

Generates a detailed dedup plan via `code-deduplicating` skill.

**Template:**
```
Review {folder} for duplication issues and use code-deduplicating skill to generate .plans/{date}-F{index}-{folder-slug}-dedup-review.md
```

**Rules:**
- `{folder}` = exact folder path (e.g., `src/api/`, `pkg/token/`)
- `{index}` = folder order ID from the audit (F01, F02, F03...)
- `{date}` = today's date in YYYY-MM-DD format
- `{folder-slug}` = kebab-case of the folder path
- Keep title to **one line** — be concise

### 2. Refactor Plan Task

Generates a detailed refactor plan via `code-refactoring` skill.

**Template:**
```
Review {folder} for quality issues and use code-refactoring skill to generate .plans/{date}-F{index}-{folder-slug}-quality-review.md
```

**Rules:**
- `{folder}` = exact folder path
- `{index}` = folder order ID from the audit (F01, F02, F03...)
- `{date}` = today's date in YYYY-MM-DD format
- `{folder-slug}` = kebab-case of the folder path
- Keep title to **one line** — be concise

### 3. Simplify Plan Task

Generates a detailed simplification plan via `code-simplifying` skill.

**Template:**
```
Review {folder} for simplification opportunities and use code-simplifying skill to generate .plans/{date}-F{index}-{folder-slug}-simplify-review.md
```

**Rules:**
- `{folder}` = exact folder path
- `{index}` = folder order ID from the audit (F01, F02, F03...)
- `{date}` = today's date in YYYY-MM-DD format
- `{folder-slug}` = kebab-case of the folder path
- Keep title to **one line** — be concise

### 4. Fix/Execute Task

Applies approved changes by reviewing the three plans.

**Template:**
```
Review .plans/{date}-F{index}-{folder-slug}-dedup-review.md, .plans/{date}-F{index}-{folder-slug}-quality-review.md, and .plans/{date}-F{index}-{folder-slug}-simplify-review.md, consolidate suggestions by priority, implement approved improvements
```

**Rules:**
- `{index}` = folder order ID from the audit (F01, F02, F03...)
- `{date}` = today's date in YYYY-MM-DD format
- `{folder-slug}` = kebab-case of the folder path

### 5. Global Cohesion Checkpoint (after all tasks)

After all folder tasks are complete, insert a **2-task global cohesion checkpoint**:

**Task 1 — Audit:**
```
Review all folders for global cohesion and use code-restructuring skill to generate .plans/{date}-GC1-global-cohesion-review.md
```

**Task 2 — Fix:**
```
Review .plans/{date}-GC1-global-cohesion-review.md, consolidate suggestions by priority, implement high-value improvements (bug fixes, missing tests, DRY violations, architecture issues)
```

**Scope** = entire repository.

## TODO.yaml Format

```yaml
tasks:
  # ═══════════════════════════════════════════════════════════════
  # Folder F01: src/api/
  # ═══════════════════════════════════════════════════════════════

  - title: "Review src/api/ for duplication issues and use code-deduplicating skill to generate .plans/{date}-F01-src-api-dedup-review.md"
    completed: false

  - title: "Review src/api/ for quality issues and use code-refactoring skill to generate .plans/{date}-F01-src-api-quality-review.md"
    completed: false

  - title: "Review src/api/ for simplification opportunities and use code-simplifying skill to generate .plans/{date}-F01-src-api-simplify-review.md"
    completed: false

  - title: "Review .plans/{date}-F01-src-api-dedup-review.md, .plans/{date}-F01-src-api-quality-review.md, and .plans/{date}-F01-src-api-simplify-review.md, consolidate suggestions by priority, implement approved improvements"
    completed: false

  # ═══════════════════════════════════════════════════════════════
  # Global Cohesion — after all optimization complete
  # Scope: entire repository
  # ═══════════════════════════════════════════════════════════════

  - title: "Review all folders for global cohesion and use code-restructuring skill to generate .plans/{date}-GC1-global-cohesion-review.md"
    completed: false

  - title: "Review .plans/{date}-GC1-global-cohesion-review.md, consolidate suggestions by priority, implement approved improvements"
    completed: false
```

**YAML rules:**
- Two fields per task: `title`, `completed` (always `false`)
- DO NOT include `description` field
- Use YAML block comments (`#`) for section separators with `═` box lines
- No blank lines between a title and its fields
- Blank line between tasks for readability
- Global cohesion checkpoint always comes last

## Workflow

### Step 1: Capture user requirements
- **Goal:** Understand optimization scope and constraints.
- **Actions:** Restate the request, confirm excluded folders, and identify focus areas (dedup/refactor/simplify/fix).
- **Output:** Clear scope summary to guide the audit.

### Step 2: Audit all folders
- **Goal:** Build a complete, folder-level inventory.
- **Actions:** Enumerate folders, document purpose and risks, mark `optimize`/`skip`/`third-party`, and record ✅ for complete items in `ANALYSIS.md`.
- **Output:** `ANALYSIS.md` with audit notes.

### Step 3: Determine folder order
- **Goal:** Sequence work for minimal risk following dependency inversion principle.
- **Actions:** Order folders by dependency layers, **child directories before parent directories**:
  1. **Bottom layer first** — foundational packages with no/minimal dependencies (errors, utilities, config)
  2. **Child directories before parents** — optimize subdirectories first, then parent directory
  3. **Layer progression** — bottom-up dependency order (foundation → utilities → domain logic → interfaces → entry points)
  4. **Within each layer** — child directories first, parent last
- **Output:** Ordered folder list with IDs F01, F02, ... following bottom-up + child-first ordering

### Step 4: Generate folder tasks
- **Goal:** Create per-folder dedup/refactor/simplify/fix tasks.
- **Actions:** For each `optimize` folder, add three review tasks (dedup/refactor/simplify) plus a fix task that applies approved changes.
- **Output:** 4 tasks per folder in TODO.yaml.

### Step 5: Add global cohesion checkpoint
- **Goal:** Catch cross-folder architecture issues.
- **Actions:** Append GC1 audit/fix tasks covering the entire repository.
- **Output:** Final TODO.yaml.

### Step 6: Write TODO.yaml
- **Goal:** Produce the task list at repository root.
- **Actions:** Ensure formatting rules and ordering are correct.
- **Output:** `TODO.yaml` in the root directory.

## Inputs

Required:
- User-provided optimization requirements (conversation)

Optional:
- Existing `ANALYSIS.md` to update
- Existing `TODO.yaml` to extend
- Custom checkpoint interval (default: every 4 folders)

## Remember

- This skill **does not** analyze code quality itself; it converts the folder audit into tasks.
- Folder audit is mandatory — **every folder** must be accounted for (optimize/skip/third-party).
- Skip ✅ items when generating tasks.
- Each folder gets **three review tasks** (dedup/refactor/simplify) plus **one fix task** that applies approved changes.
- Plan filenames must match the folder ID and slug.
- Global cohesion checkpoint always comes last.

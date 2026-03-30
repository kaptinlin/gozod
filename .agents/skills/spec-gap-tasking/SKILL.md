---
description: Convert ANALYSIS.md into structured TODO.yaml with paired Plan+Execute tasks and quality reviews. Use after spec-gap-analyzing has produced the gap analysis report.
name: spec-gap-tasking
---


# Spec Gap Tasking

**Two-phase workflow:** First audits `ANALYSIS.md` accuracy by checking actual implementation, then generates `TODO.yaml` with Plan/Execute tasks and Quality Review checkpoints.

**Announce at start:** "I'm using the spec-gap-tasking skill."

**Prerequisite:** `ANALYSIS.md` must exist at project root. If not, run `spec-gap-analyzing` first.

## Phase 1: Audit ANALYSIS.md (CRITICAL)

**Before generating tasks, verify each analysis point against actual codebase.**

### Audit Process

For each gap/issue in ANALYSIS.md:

1. **Check actual implementation status**
   - Read relevant source files
   - Verify claimed gaps actually exist
   - Check if "missing" features are already implemented
   - Validate line counts and complexity claims

2. **Update ANALYSIS.md with corrections**
   - Mark implemented items as ✅ (keep for context, but won't generate tasks)
   - Correct inaccurate assessments
   - Update status badges (❌ Missing → ✅ Implemented if found)
   - Add notes about actual findings

3. **Document audit results**
   - Add "Audit Notes" section to each category
   - List what was verified
   - Note any discrepancies found

### Example Audit Corrections

**Before audit:**
```markdown
### 3. Validation System - ❌ MISSING
**Implementation:** None
**Gap:** Full validation framework needed
```

**After audit (if found implemented):**
```markdown
### 3. Validation System - ✅ IMPLEMENTED
**Implementation:** src/validation/ directory, 8,500 lines, 35 validators
**Gap:** None (already complete)
**Audit Notes:** Found complete implementation with validator registry, runner, and 6 validator categories.
```

### Audit Checklist

Before proceeding to Phase 2, verify:
- [ ] Checked implementation status for each claimed gap
- [ ] Updated ANALYSIS.md with actual findings
- [ ] Marked completed items as ✅
- [ ] Corrected inaccurate line counts or complexity claims
- [ ] Added audit notes to each category
- [ ] Saved updated ANALYSIS.md

**Only proceed to Phase 2 after audit is complete.**

## Phase 2: Generate TODO.yaml

**Only generate tasks for items marked as ❌ or ⚠️ in audited ANALYSIS.md.**

Skip items marked ✅ (already implemented) - they don't need tasks.

## Pipeline Position

```
spec-gap-analyzing ──▶ Phase 1: Audit ──▶ Phase 2: Generate
     │                       │                    │
     ▼                       ▼                    ▼
 ANALYSIS.md          Updated ANALYSIS.md    TODO.yaml
(may have errors)    (verified accurate)   (only real gaps)
```

## Task Structure

Every gap item in ANALYSIS.md produces exactly **two tasks** — a Plan task and an Execute task.

### 1. Plan Task

Generates an action plan via `analyzing` skill.

**Template:**
```
Plan {deliverable name} (operation type) and use analyzing skill to generate .plans/{date}-{category-slug}-{slug}.md
```

**Rules:**
- `{deliverable name}` = human-readable name of what's being planned (e.g., "API reference documentation", "user authentication guide")
- `(operation type)` = implement / fix / refactor（必须标注）
- `{date}` = today's date in YYYY-MM-DD format
- `{slug}` = kebab-case of the deliverable name
- `{category-slug}` = kebab-case category name from ANALYSIS.md (e.g., "validation", "auth")
- The `.plans/{date}-{category-slug}-{slug}.md` path must match exactly between Plan and Execute tasks
- Keep title to **one line** — be concise

**操作类型标注**：
- **implement** — 实现缺失内容（示例："Plan missing validation implementation (implement)"）
- **fix** — 修复错误内容（示例："Plan incorrect error handling fix (fix)"）
- **refactor** — 重构现有内容（示例："Plan API consistency refactor (refactor)"）

### 2. Execute Task

Produces the actual deliverable following the plan.

**Template:**
```
Produce {deliverable} ({impact scope}) and use spec-writing skill to follow .plans/{date}-{category-slug}-{slug}.md
```

**Rules:**
- The `.plans/` path must be identical to the corresponding Plan task
- `{deliverable}` = target output (e.g., "docs/api-reference.md", "docs/research/competitor-analysis.md")
- `{impact scope}` = what files are affected: "create X", "modify X", "create X, modify Y", etc.
- For cross-cutting tasks, use "Review {areas} end-to-end ({impact scope}) and use spec-writing skill to follow .plans/{date}-{category-slug}-{slug}.md"
- Keep title to **one line** — be concise

**影响范围标注**：
- **create** — 新建文件（示例："Produce docs/api.md (create docs/api.md)"）
- **modify** — 修改现有文件（示例："Produce docs/guide.md (modify docs/guide.md)"）
- **create + modify** — 新建并修改（示例："Produce docs/tutorial.md (create docs/tutorial.md, modify docs/api.md)"）
- **delete** — 删除文件（示例："Review legacy docs (delete docs/old.md)"）

### 3. Quality Review (every 4 categories)

After every 4 categories of Plan+Execute tasks, insert a **4-task quality checkpoint**:

**Task 1 — Dedup Audit:**
```
Review {deliverable list} for duplication issues and use code-deduplicating skill to generate .plans/{date}-{category-slug}-dedup-review.md
```

**Task 2 — Refactor Audit:**
```
Review {deliverable list} for quality issues and use code-refactoring skill to generate .plans/{date}-{category-slug}-quality-review.md
```

**Task 3 — Simplify Audit:**
```
Review {deliverable list} for simplification opportunities and use code-simplifying skill to generate .plans/{date}-{category-slug}-simplify-review.md
```

**Task 4 — Fix:**
```
Review .plans/{date}-{category-slug}-dedup-review.md, .plans/{date}-{category-slug}-quality-review.md, and .plans/{date}-{category-slug}-simplify-review.md, consolidate suggestions by priority, implement approved improvements
```

**Deliverable list** = union of all deliverables produced by the preceding 4 categories.

### 4. Code Restructuring (after all tasks)

After all Plan+Execute tasks and Quality Reviews are complete, insert a **2-task global cohesion checkpoint**:

**Task 1 — Audit:**
```
Review all deliverables for global cohesion and use code-restructuring skill to generate .plans/{date}-{category-slug}-global-cohesion-review.md
```

**Task 2 — Fix:**
```
Review .plans/{date}-{category-slug}-global-cohesion-review.md, consolidate suggestions by priority, implement approved improvements
```

**Scope** = all deliverables produced by all categories.

**Purpose:**
- Catch global consistency issues missed by incremental quality reviews
- Identify redundant content across distant deliverables
- Verify cross-references and terminology consistency across the body of work
- Ensure cohesive voice and structure

### 5. Architecture Audit (after Code Restructuring, code projects only)

For code projects, after Code Restructuring, insert a **2-task architecture audit checkpoint**:

**Task 1 — Audit:**
```
Run architecture audit and use architecture-audit skill to generate .plans/{date}-{category-slug}-arch-audit.md and .plans/{date}-{category-slug}-arch-audit-plan.md
```

**Task 2 — Fix:**
```
Review .plans/{date}-{category-slug}-arch-audit-plan.md, consolidate suggestions by priority, implement approved improvements
```

**Scope** = entire codebase (all packages/modules).

**Purpose:**
- Systematic full-codebase health check independent of recent changes
- Detect circular dependencies, fat interfaces, layer violations
- Measure package-level metrics (size, coupling, coverage)
- Validate SPECS alignment

**Note:** This checkpoint is only for code projects. Skip for documentation-only projects.

## Placement Rules

Given categories ordered by priority from ANALYSIS.md:

```
Categories 1–4  →  Plan+Execute tasks  →  Quality Review Q1
Categories 5–8  →  Plan+Execute tasks  →  Quality Review Q2
Categories 9–N  →  Plan+Execute tasks  →  Quality Review Q3
```

If the last group has fewer than 4 categories, still insert a quality review after it.

**After all categories and quality reviews**, append two final tasks for global cohesion review:

```
All tasks  →  Code Restructuring GC1
```

## TODO.yaml Format

```yaml
tasks:
  # ═══════════════════════════════════════════════════════════════
  # Category X: {Name}
  # ═══════════════════════════════════════════════════════════════

  - title: "Plan {deliverable name} (operation type) and use analyzing skill to generate .plans/{date}-{category-slug}-{slug}.md"
    completed: false

  - title: "Produce {deliverable} and use spec-writing skill to follow .plans/{date}-{category-slug}-{slug}.md"
    completed: false

  # ... more plan+execute pairs ...

  # ═══════════════════════════════════════════════════════════════
  # Quality Review #N — after Categories W/X/Y/Z
  # Scope: {deliverable list}
  # ═══════════════════════════════════════════════════════════════

  - title: "Review {deliverable list} for duplication issues and use code-deduplicating skill to generate .plans/{date}-{category-slug}-dedup-review.md"
    completed: false

  - title: "Review {deliverable list} for quality issues and use code-refactoring skill to generate .plans/{date}-{category-slug}-quality-review.md"
    completed: false

  - title: "Review {deliverable list} for simplification opportunities and use code-simplifying skill to generate .plans/{date}-{category-slug}-simplify-review.md"
    completed: false

  - title: "Review .plans/{date}-{category-slug}-dedup-review.md, .plans/{date}-{category-slug}-quality-review.md, and .plans/{date}-{category-slug}-simplify-review.md, consolidate suggestions by priority, implement approved improvements"
    completed: false

  # ═══════════════════════════════════════════════════════════════
  # Code Restructuring — after all deliverables complete
  # Scope: all deliverables
  # ═══════════════════════════════════════════════════════════════

  - title: "Review all deliverables for global cohesion and use code-restructuring skill to generate .plans/{date}-{category-slug}-global-cohesion-review.md"
    completed: false

  - title: "Review .plans/{date}-{category-slug}-global-cohesion-review.md, consolidate suggestions by priority, implement approved improvements"
    completed: false

  # ═══════════════════════════════════════════════════════════════
  # Architecture Audit — after Code Restructuring (code projects only)
  # Scope: entire codebase
  # ═══════════════════════════════════════════════════════════════

  - title: "Run architecture audit and use architecture-audit skill to generate .plans/{date}-{category-slug}-arch-audit.md and .plans/{date}-{category-slug}-arch-audit-plan.md"
    completed: false

  - title: "Review .plans/{date}-{category-slug}-arch-audit-plan.md, consolidate suggestions by priority, implement approved improvements"
    completed: false
```

## Task Ordering Rules

**任务排序原则**：
- 任务按依赖关系排序（上面的先执行，下面的后执行）
- 有依赖的任务必须排在被依赖任务之后
- 独立任务可以任意顺序排列

**示例**：
```yaml
tasks:
  # API reference（独立，无依赖）
  - title: "Plan API reference documentation (implement)..."
    completed: false
  - title: "Produce docs/api-reference.md..."
    completed: false

  # User guide（依赖 API reference）
  - title: "Plan user guide documentation (implement)..."
    completed: false
  - title: "Produce docs/user-guide.md..."
    completed: false

  # Tutorial（依赖 user guide）
  - title: "Plan tutorial documentation (implement)..."
    completed: false
  - title: "Produce docs/tutorial.md..."
    completed: false
```

**依赖关系识别**：
- 如果 Execute task 的 title 中包含 `modify` 或引用其他文件，说明有依赖
- 将被依赖的文档的 Plan+Execute 任务排在前面
- 依赖方的 Plan+Execute 任务排在后面

**YAML rules:**
- Two fields per task: `title`, `completed` (always `false`)
- DO NOT include `description` field
- Use YAML block comments (`#`) for category separators with `═` box lines
- No blank lines between a title and its fields
- Blank line between tasks for readability
- Code Restructuring comes after all Plan+Execute tasks
- Architecture Audit comes last (code projects only)

## Workflow

### Step 1: Audit ANALYSIS.md (Phase 1)

1. Read `ANALYSIS.md` from project root
2. For each gap/issue listed:
   - Check actual implementation (read source files)
   - Verify claimed status (missing/partial/divergent)
   - Update ANALYSIS.md with corrections
3. Save updated ANALYSIS.md
4. Announce audit completion: "Audit complete. Found X items already implemented, Y items confirmed as gaps."

### Step 2: Read Audited ANALYSIS.md (Phase 2)

Read updated `ANALYSIS.md` and extract only items marked ❌ or ⚠️:
- Gap categories (ID, name, source spec, focus)
- Gap types (Missing, Partial, Divergent, Outdated, Untested, Integration)
- Priority ordering (dependency-based)

**Skip ✅ items** - they don't need tasks.

### Step 3: Determine Category Order

Follow the priority from audited ANALYSIS.md. Default ordering:
1. Foundation (core concepts, glossary, architecture)
2. Research / analysis
3. Product / requirements
4. Technical / API reference
5. User-facing (guides, tutorials)
6. Operations (runbooks, deployment)
7. Cross-cutting reviews

### Step 4: Generate Task Pairs

For each gap item marked ❌ or ⚠️ in audited ANALYSIS.md, generate exactly one Plan task + one Execute task using the templates above.

**Skip ✅ items** - no tasks needed for already-implemented features.

### Step 4: Insert Quality Reviews

Count categories in order. After every 4th category, insert a 4-task quality review:
- Q1 after categories 1–4
- Q2 after categories 5–8
- Q3 after categories 9–end

Compute the deliverable scope for each review by collecting all outputs from the preceding categories.

### Step 5: Insert Code Restructuring

After all categories and quality reviews, append a 2-task global cohesion review:
- Audit: Use code-restructuring skill to generate `.plans/{date}-{category-slug}-global-cohesion-review.md`
- Fix: Review `.plans/{date}-{category-slug}-global-cohesion-review.md`, remove unreasonable suggestions, then apply approved fixes

### Step 6: Insert Architecture Audit (code projects only)

For code projects, after Code Restructuring, append a 2-task architecture audit:
- Audit: Use architecture-audit skill to generate `.plans/{date}-{category-slug}-arch-audit.md` and `.plans/{date}-{category-slug}-arch-audit-plan.md`
- Fix: Follow `.plans/{date}-{category-slug}-arch-audit-plan.md` to implement HIGH and MEDIUM priority fixes

### Step 7: Write TODO.yaml

Write the final `TODO.yaml` at project root.

## Inputs

Required:
- `ANALYSIS.md` — the gap analysis report in project root (produced by `spec-gap-analyzing`)

Optional:
- Existing `TODO.yaml` to update (append new tasks)
- Custom review interval (default: every 4 categories)
- Date override for plan file naming (default: today)

## Remember

- This skill is the **second step** after `spec-gap-analyzing` — it does not survey SPECS or analyze gaps
- Plan titles must say "ALL requirements" — never constrain to specific focus points
- Quality reviews use `code-deduplicating` then `code-refactoring` (Dedup Audit → Refactor Audit → Fix → Polish)
- Polish tasks focus on the last batch of deliverables, not the entire body of work
- `.plans/` directory is `.gitignore`d — plan files are ephemeral working artifacts
- Cross-cutting tasks use "review ... end-to-end" phrasing, not "produce"
- For code projects, substitute skills: `analyzing` → `tdd-planning`, Execute tasks → `tdd-implementing`, `code-refactoring` → `code-refactoring` + `code-simplifying`
- The generic Plan+Execute+Quality Review structure maps to Plan+Impl+Refactor Review for code projects
- **Avoid over-tasking**: If ANALYSIS.md includes items that are already fulfilled by SPEC itself (architecture principles, governance rules, design rationale), do not generate tasks for them — SPEC is the SSOT for these

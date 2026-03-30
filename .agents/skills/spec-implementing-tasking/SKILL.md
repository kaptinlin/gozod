---
description: Convert root ANALYSIS.md into TODO.yaml implementation tasks for Stage 3 spec-based coding. Use when generating module-scoped coding tasks before running ralphy loop execution.
name: spec-implementing-tasking
---


# Spec Implementing Tasking

**Two-phase workflow:** Audits `ANALYSIS.md` scope (what needs building), then generates `TODO.yaml` with Plan/Execute tasks.

**Announce at start:** "I'm using the spec-implementing-tasking skill."

## Phase 1: Audit ANALYSIS.md

**Verify scope assessments before generating tasks.**

### Audit Process

For each item in ANALYSIS.md:

1. **Check existing code**
   - Does target module/directory exist?
   - What's already built?

2. **Correct scope type**
   - "New module" → "Extension" if base exists
   - "Extension" → ✅ Skip if already complete

3. **Update ANALYSIS.md**
   - Fix scope types
   - Add audit notes
   - Mark complete items ✅

### Example

**Before:**
```markdown
### User Auth - New Module
Target: src/auth/
```

**After (if base exists):**
```markdown
### User Auth - Extension
Target: src/auth/ (extend)
Audit: Found basic auth, only need OAuth.
```

## Phase 2: Generate TODO.yaml

Generate Plan+Execute tasks for items needing work (New/Extension/Foundation).

Skip ✅ items.

## Pipeline Position

```
spec-implementing-analyzing ──▶ Phase 1: Audit ──▶ Phase 2: Generate
         │                           │                    │
         ▼                           ▼                    ▼
     ANALYSIS.md              Updated ANALYSIS.md     TODO.yaml
   (may have errors)         (verified accurate)   (only real gaps)
```

## Task Structure

Every item in ANALYSIS.md produces exactly **two tasks** — a Plan task and an Execute task.

### 1. Plan Task

Generates a detailed implementation plan via `tdd-planning` skill.

**Template:**
```
Plan {feature name} implementation (operation type) and use tdd-planning skill to generate .plans/{date}-{category-slug}-{slug}.md
```

**Rules:**
- `{feature name}` = human-readable name of what's being implemented (e.g., "user authentication", "token parser")
- `(operation type)` = create / extend / refactor（必须标注）
- `{date}` = today's date in YYYY-MM-DD format
- `{slug}` = kebab-case of the feature name
- `{category-slug}` = kebab-case category name from ANALYSIS.md (e.g., "auth", "token-system")
- The `.plans/{date}-{category-slug}-{slug}.md` path must match exactly between Plan and Execute tasks
- Keep title to **one line** — be concise

**操作类型标注**：
- **create** — 创建新模块（示例："Plan user authentication implementation (create new module)"）
- **extend** — 扩展现有模块（示例："Plan token parser extension (extend existing module)"）
- **refactor** — 重构现有代码（示例："Plan error handling refactor (refactor existing code)"）

### 2. Execute Task

Implements the plan by writing code against the target modules.

**Template:**
```
Implement {feature name} and use tdd-implementing skill to follow .plans/{date}-{category-slug}-{slug}.md
```

**Rules:**
- The `.plans/` path must be identical to the corresponding Plan task
- `{feature name}` = what's being implemented (e.g., "user authentication", "token parser")
- For integration test tasks, use "Validate {modules} end-to-end and use tdd-implementing skill to follow .plans/{date}-{category-slug}-{slug}.md"
- Keep title to **one line** — be concise

**影响范围声明**：
- 任务描述必须包含影响范围（新建 / 修改 / 删除）
- 示例："Target modules: pkg/auth/ (create), pkg/token/ (extend), pkg/legacy/ (delete)"

### 3. Quality Review (every 4 categories)

After every 4 categories of Plan+Execute tasks, insert a **4-task quality checkpoint**:

**Task 1 — Dedup Audit:**
```
Review {module list} for duplication issues and use code-deduplicating skill to generate .plans/{date}-{category-slug}-dedup-review.md
```

**Task 2 — Refactor Audit:**
```
Review {module list} for quality issues and use code-refactoring skill to generate .plans/{date}-{category-slug}-quality-review.md
```

**Task 3 — Simplify Audit:**
```
Review {module list} for simplification opportunities and use code-simplifying skill to generate .plans/{date}-{category-slug}-simplify-review.md
```

**Task 4 — Fix:**
```
Review .plans/{date}-{category-slug}-dedup-review.md, .plans/{date}-{category-slug}-quality-review.md, and .plans/{date}-{category-slug}-simplify-review.md, consolidate suggestions by priority, implement approved improvements
```

**Module list** = union of all modules touched by the preceding 4 categories.

### 4. Code Restructuring (after all tasks)

After all Plan+Execute tasks and Quality Reviews are complete, insert a **2-task global cohesion checkpoint**:

**Task 1 — Audit:**
```
Review all modules for global cohesion and use code-refactoring skill to generate .plans/{date}-{category-slug}-global-cohesion-review.md
```

**Task 2 — Fix:**
```
Review .plans/{date}-{category-slug}-global-cohesion-review.md, consolidate suggestions by priority, implement approved improvements
```

**Scope** = all modules implemented across all categories.

**Purpose:**
- Catch global architecture issues missed by incremental quality reviews
- Identify redundant code across distant modules
- Verify cross-module interfaces and naming consistency
- Ensure cohesive structure across the codebase

### 5. Architecture Audit (after Code Restructuring)

After Code Restructuring, insert a **2-task architecture audit checkpoint**:

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
  # Category X: {Name} (SPECS NN)
  # ═══════════════════════════════════════════════════════════════

  - title: "Plan {feature name} implementation (operation type) and use tdd-planning skill to generate .plans/{date}-{category-slug}-{slug}.md"
    completed: false

  - title: "Implement {feature name} and use tdd-testing skill to follow .plans/{date}-{category-slug}-{slug}.md"
    completed: false

  # ... more plan+execute pairs ...
```

## Task Ordering Rules

**任务排序原则**：
- 任务按依赖关系排序（上面的先执行，下面的后执行）
- 有依赖的任务必须排在被依赖任务之后
- 独立任务可以任意顺序排列

**示例**：
```yaml
tasks:
  # Token parser（独立，无依赖）
  - title: "Plan token parser implementation (create new module)..."
    completed: false
  - title: "Implement token parser..."
    completed: false

  # Auth module（依赖 token parser）
  - title: "Plan auth module implementation (extend existing module)..."
    completed: false
  - title: "Implement auth module..."
    completed: false

  # Session manager（依赖 auth module）
  - title: "Plan session manager implementation (create new module)..."
    completed: false
  - title: "Implement session manager..."
    completed: false
```

**依赖关系识别**：
- 如果 Execute task 的 title 中包含 `extend` 或引用其他模块，说明有依赖
- 将被依赖的模块的 Plan+Execute 任务排在前面
- 依赖方的 Plan+Execute 任务排在后面

  # ═══════════════════════════════════════════════════════════════
  # Quality Review #N — after Categories W/X/Y/Z
  # Scope: {module list}
  # ═══════════════════════════════════════════════════════════════

  - title: "Review {module list} for duplication issues and use code-deduplicating skill to generate .plans/{date}-{category-slug}-dedup-review.md"
    completed: false

  - title: "Review {module list} for quality issues and use code-refactoring skill to generate .plans/{date}-{category-slug}-quality-review.md"
    completed: false

  - title: "Review {module list} for simplification opportunities and use code-simplifying skill to generate .plans/{date}-{category-slug}-simplify-review.md"
    completed: false

  - title: "Review .plans/{date}-{category-slug}-dedup-review.md, .plans/{date}-{category-slug}-quality-review.md, and .plans/{date}-{category-slug}-simplify-review.md, consolidate suggestions by priority, implement approved improvements"
    completed: false

  # ═══════════════════════════════════════════════════════════════
  # Code Restructuring — after all implementation complete
  # Scope: all implemented modules
  # ═══════════════════════════════════════════════════════════════

  - title: "Review all modules for global cohesion and use code-refactoring skill to generate .plans/{date}-{category-slug}-global-cohesion-review.md"
    completed: false

  - title: "Review .plans/{date}-{category-slug}-global-cohesion-review.md, consolidate suggestions by priority, implement approved improvements"
    completed: false

  # ═══════════════════════════════════════════════════════════════
  # Architecture Audit — after Code Restructuring
  # Scope: entire codebase
  # ═══════════════════════════════════════════════════════════════

  - title: "Run architecture audit and use architecture-audit skill to generate .plans/{date}-{category-slug}-arch-audit.md and .plans/{date}-{category-slug}-arch-audit-plan.md"
    completed: false

  - title: "Review .plans/{date}-{category-slug}-arch-audit-plan.md, consolidate suggestions by priority, implement approved improvements"
    completed: false
```

**YAML rules:**
- Two fields per task: `title`, `completed` (always `false`)
- DO NOT include `description` field
- Use YAML block comments (`#`) for category separators with `═` box lines
- No blank lines between a title and its fields
- Blank line between tasks for readability
- Code Restructuring comes after all Plan+Execute tasks
- Architecture Audit comes last

## Workflow

### Step 1: Read ANALYSIS.md

Read `ANALYSIS.md` from the project root and extract:
- Scope: what needs to be implemented
- Priority: category ordering and importance
- Structure: category breakdown with IDs, names, source specs, and target modules

### Step 2: Determine Category Order

Follow the priority from ANALYSIS.md. Typical ordering:
1. Core types and interfaces (foundations)
2. Data structures and models
3. Business logic and domain rules
4. Service layer and orchestration
5. Interface layer (CLI, API, handlers)
6. Advanced systems (caching, optimization)
7. Integration tests

### Step 3: Generate Task Pairs

For each category item in ANALYSIS.md, generate exactly one Plan task + one Execute task using the templates above.

### Step 4: Insert Quality Reviews

Count categories in order. After every 4th category, insert a 4-task quality review:
- Q1 after categories 1–4
- Q2 after categories 5–8
- Q3 after categories 9–end

Compute the module scope for each review by collecting all modules from the preceding categories.

### Step 5: Insert Code Restructuring

After all categories and quality reviews, append a 2-task global cohesion review:
- Audit: Use code-refactoring skill to generate `.plans/{date}-{category-slug}-global-cohesion-review.md`
- Fix: Review `.plans/{date}-{category-slug}-global-cohesion-review.md`, remove unreasonable suggestions, then apply approved fixes

### Step 6: Insert Architecture Audit

After Code Restructuring, append a 2-task architecture audit:
- Audit: Use architecture-audit skill to generate `.plans/{date}-{category-slug}-arch-audit.md` and `.plans/{date}-{category-slug}-arch-audit-plan.md`
- Fix: Follow `.plans/{date}-{category-slug}-arch-audit-plan.md` to implement HIGH and MEDIUM priority fixes

### Step 7: Write TODO.yaml

Write the final `TODO.yaml` at project root.

## Inputs

Required:
- `ANALYSIS.md` — the implementation analysis report in project root (produced by `spec-implementing-analyzing`)

Optional:
- Existing `TODO.yaml` to update (append new tasks)
- Custom review interval (default: every 4 categories)
- Date override for plan file naming (default: today)

## Remember

- This skill is the **second step** after `spec-implementing-analyzing` — it does not analyze specs or survey the codebase
- Plan titles must say "ALL requirements" — never constrain to specific focus points
- Quality reviews use `code-deduplicating` then `code-refactoring` (Dedup Audit → Refactor Audit → Fix → Polish)
- Polish tasks focus on modules from the last batch of categories, not the entire codebase
- `.plans/` directory is `.gitignore`d — plan files are ephemeral working artifacts
- Integration test tasks use "validate ... end-to-end" phrasing, not "implement"
- `ANALYSIS.md` is in the root directory; `.plans/*.md` are detailed plans generated by tasks
- **Avoid over-tasking**: If ANALYSIS.md includes items that are documentation/governance concerns (architecture principles, design rationale), do not generate implementation tasks for them — SPEC is the SSOT for these, not code

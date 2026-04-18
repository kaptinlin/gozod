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

**Design philosophy — think like a founder, not a contractor:**

Say no to a thousand things. More tasks ≠ better coverage — it means more
integration seams and more places for inconsistency.

- **Simplicity first** — if two Plan+Execute pairs touch the same abstraction boundary, merge them. Fewer tasks with clear module ownership beat many fine-grained tasks that fragment responsibility. Task bloat is code bloat waiting to happen
- **Complete, not over-engineered** — generate tasks for production-quality implementations, not MVP stubs. But don't add abstraction layers, plugin systems, or "flexible" patterns unless the SPECS explicitly require them
- **Every Execute task must produce concise, idiomatic code** — no "get it working first, clean up later". The first implementation is the real implementation
- **Right-sized scope** — deep enough to produce cohesive modules, not so narrow that integration becomes an afterthought, not so broad that the task becomes unfocused

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
Read {specs-refs} and ANALYSIS.md §{category-id}, then plan {feature name} implementation (operation type) and use tdd-planning skill to generate .plans/{date}-{category-slug}-{slug}.md
```

**Rules:**
- `{specs-refs}` = comma-separated list of SPECS files with §section references that define the contracts for this category (e.g., `SPECS/01-data-model.md §Pointer Semantics, SPECS/02-parsing.md §Lenient Parsing Baseline`)
- `{category-id}` = the category ID from ANALYSIS.md (e.g., `C01`, `C02`)
- `{feature name}` = human-readable name of what's being implemented (e.g., "user authentication", "token parser")
- `(operation type)` = create / extend / refactor（必须标注）
- `{date}` = today's date in YYYY-MM-DD format
- `{slug}` = kebab-case of the feature name
- `{category-slug}` = kebab-case category name from ANALYSIS.md (e.g., "auth", "token-system")
- The `.plans/{date}-{category-slug}-{slug}.md` path must match exactly between Plan and Execute tasks

**SPECS 引用规则**：
- 每个 Plan 任务必须引用该任务相关的所有 SPECS 文件和具体章节
- 使用 `§` 符号标注章节（如 `SPECS/02-parsing.md §Lenient Parsing Baseline`）
- 当一个任务涉及多个 SPECS 时，用逗号分隔（如 `SPECS/01.md §Feed, SPECS/03.md §Extensions`）
- 集成测试任务引用所有 SPECS 的 Acceptance Criteria 章节
- ANALYSIS.md 必须引用对应的 §{category-id}（如 `ANALYSIS.md §C01`），因为 ANALYSIS.md 包含参考源码

**操作类型标注**：
- **create** — 创建新模块（示例："Read SPECS/01.md §Auth and ANALYSIS.md §C03, then plan user authentication implementation (create new module)"）
- **extend** — 扩展现有模块（示例："Read SPECS/02.md §Token and ANALYSIS.md §C05, then plan token parser extension (extend existing module)"）
- **refactor** — 重构现有代码（示例："Read SPECS/01.md §Error and ANALYSIS.md §C01, then plan error handling refactor (refactor existing code)"）

### 2. Execute Task

Implements the plan by writing code against the target modules.

**Template:**
```
Implement {feature name} ({target files}) following {specs-contracts}, then use tdd-implementing skill to follow .plans/{date}-{category-slug}-{slug}.md
```

**Rules:**
- `{target files}` = the target files/directories this task creates or modifies (e.g., `rss/rss.go, rss/parser.go, rss/adapter.go`)
- `{specs-contracts}` = the SPECS contracts this implementation must satisfy (e.g., `SPECS/01-data-model.md §rss.Feed and SPECS/02-parsing.md §ParseRSS`)
- The `.plans/` path must be identical to the corresponding Plan task
- `{feature name}` = what's being implemented (e.g., "user authentication", "token parser")
- For integration test tasks, use "Validate all packages end-to-end using Acceptance Criteria from SPECS/{NN-NN} as test oracle, then use tdd-implementing skill to follow .plans/{date}-{category-slug}-{slug}.md"

**影响范围声明**：
- 任务描述必须包含影响范围（目标文件列表）
- 示例："Implement user auth (pkg/auth/auth.go, pkg/auth/oauth.go) following SPECS/03-auth.md §OAuth2 contracts"

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
  # Category X: {Name}
  # Target: {target files/directories}
  # ═══════════════════════════════════════════════════════════════

  - title: "Read SPECS/{nn}-{name}.md §{Section1} §{Section2}, SPECS/{mm}-{name}.md §{Section3}, and ANALYSIS.md §{CXX}, then plan {feature name} implementation (operation type) and use tdd-planning skill to generate .plans/{date}-{category-slug}-{slug}.md"
    completed: false

  - title: "Implement {feature name} ({target files}) following SPECS/{nn}-{name}.md §{Contract1} and SPECS/{mm}-{name}.md §{Contract2} contracts, then use tdd-implementing skill to follow .plans/{date}-{category-slug}-{slug}.md"
    completed: false

  # ... more plan+execute pairs ...
```

### SPECS Traceability Example

```yaml
tasks:
  # ═══════════════════════════════════════════════════════════════
  # C04 — RSS Parser + Format-Specific Model
  # Target: rss/rss.go, rss/parser.go, rss/adapter.go
  # ═══════════════════════════════════════════════════════════════

  - title: "Read SPECS/01-data-model.md §Format-Specific Models §GUID IsPermaLink, SPECS/02-parsing.md §ParseRSS §Lenient Parsing Baseline §Date Parsing, SPECS/03-extensions.md §Extension Routing, and ANALYSIS.md §C04, then plan RSS parser and format-specific model implementation (create new module) and use tdd-planning skill to generate .plans/2026-03-28-rss-parser-plan.md"
    completed: false

  - title: "Implement RSS parser and format-specific model (rss/rss.go, rss/parser.go, rss/adapter.go) following SPECS/01-data-model.md §rss.Feed and SPECS/02-parsing.md §ParseRSS contracts, then use tdd-implementing skill to follow .plans/2026-03-28-rss-parser-plan.md"
    completed: false

  # ═══════════════════════════════════════════════════════════════
  # CIT — Integration Tests
  # Target: *_test.go
  # ═══════════════════════════════════════════════════════════════

  - title: "Read SPECS/01-06 Acceptance Criteria sections and ANALYSIS.md §CIT, then plan integration tests implementation (create new module) and use tdd-planning skill to generate .plans/2026-03-28-integration-tests-plan.md"
    completed: false

  - title: "Validate all packages end-to-end using Acceptance Criteria from SPECS/01-06 as test oracle, then use tdd-implementing skill to follow .plans/2026-03-28-integration-tests-plan.md"
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
  # Token parser（独立，无依赖）
  - title: "Read SPECS/02-token.md §Token Types and ANALYSIS.md §C01, then plan token parser implementation (create new module)..."
    completed: false
  - title: "Implement token parser (pkg/token/parser.go) following SPECS/02-token.md §Parse contracts..."
    completed: false

  # Auth module（依赖 token parser）
  - title: "Read SPECS/03-auth.md §OAuth2 §JWT and ANALYSIS.md §C02, then plan auth module implementation (extend existing module)..."
    completed: false
  - title: "Implement auth module (pkg/auth/oauth.go) following SPECS/03-auth.md §OAuth2 contracts..."
    completed: false

  # Session manager（依赖 auth module）
  - title: "Read SPECS/04-session.md §Session Lifecycle and ANALYSIS.md §C03, then plan session manager implementation (create new module)..."
    completed: false
  - title: "Implement session manager (pkg/session/manager.go) following SPECS/04-session.md §Session Lifecycle contracts..."
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

For each category item in ANALYSIS.md:

1. **Identify SPECS references** — from the `**From:** SPECS/{spec-file}` field and the category content, extract all relevant SPECS files and §sections
2. **Identify ANALYSIS.md reference** — the category ID (e.g., `§C01`) which contains reference source code
3. **Generate Plan task** — include `Read {specs-refs} and ANALYSIS.md §{id}` prefix
4. **Generate Execute task** — include target files and `following {specs-contracts}` suffix
5. One Plan task + one Execute task per category

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
- **Every Plan task must reference SPECS and ANALYSIS.md** — the executor needs to know which contracts to read before planning. Without SPECS references, the executor will design in a vacuum
- **Every Execute task must reference SPECS contracts and target files** — the executor needs to know what contracts to satisfy and which files to create/modify
- **ANALYSIS.md contains reference source code** — always reference `ANALYSIS.md §{category-id}` so the executor can study the embedded reference snippets
- Quality reviews use `code-deduplicating` then `code-refactoring` (Dedup Audit → Refactor Audit → Fix → Polish)
- Polish tasks focus on modules from the last batch of categories, not the entire codebase
- `.plans/` directory is `.gitignore`d — plan files are ephemeral working artifacts
- Integration test tasks use "validate ... end-to-end" phrasing with "Acceptance Criteria from SPECS as test oracle"
- `ANALYSIS.md` is in the root directory; `.plans/*.md` are detailed plans generated by tasks
- **Avoid over-tasking**: If ANALYSIS.md includes items that are documentation/governance concerns (architecture principles, design rationale), do not generate implementation tasks for them — SPEC is the SSOT for these, not code

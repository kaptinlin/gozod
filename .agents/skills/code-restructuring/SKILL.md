---
description: Full-repository architecture review and package reorganization for global cohesion. Identifies cohesion gaps, scattered concerns, flat mega-packages, and structural degradation. Generates executable restructuring plan with import cycle risk analysis. Use when package structure has globally degraded or the user requests "global refactoring" or "restructure packages".
name: code-restructuring
---


# Global Cohesion Restructuring

Review the **entire repository** for architecture improvements. Identify cohesion gaps, eliminate structural redundancy, restore cross-package cohesion. Generate an executable global restructuring plan, then implement approved changes in independent batches.

**Announce at start:** "I'm using the code-restructuring skill."

## Core Beliefs

1. **Global perspective first** — architecture quality cannot be judged from a few diffs; analyze the full codebase
2. **High cohesion over tidy files** — packages should be organized by concern, not by edit history
3. **Simplicity is operational leverage** — fewer files, fewer layers, clearer ownership = lower maintenance cost
4. **Restructuring is not rewriting** — preserve behavior exactly; change structure, not semantics
5. **Import cycles are fatal** — prevent at design time, not at runtime; analyze dependencies before moving files
6. **Flat over nested** — prefer `pkg/foo/bar` over `pkg/foo/internal/bar` unless hiding is truly necessary
7. **Common Closure Principle (CCP)** — code that changes together should live together; organize by change reason, not by type
8. **Misleading names are worse than no names** — a directory named `adapters` that contains unrelated utilities violates trust
9. **30-line minimum per file** — files under 30 lines (excluding tests) signal over-splitting; merge related concerns
10. **Change dimension over artifact type** — organize by "what changes together" (events, anatomy, behavior), not by "what it is" (types, utils, helpers)

## Anti-Patterns We Reject

| Anti-Pattern | Why It's Wrong | What We Do Instead |
|---|---|---|
| Diff-only refactoring | Local cleanup misses global topology and cross-package duplication | Start from full-tree audit, then plan structural changes |
| Over-splitting | Dozens of tiny files reduce discoverability and increase context switching | Merge related concerns (after SRP splits) |
| Flat mega-packages | 30-60 files in one directory hide ownership boundaries | Use subdirectories to group by concern (flat over `internal/` nesting) |
| Locally-optimal moves | Local improvements may degrade global architecture | Evaluate package-level and repo-level impact first |
| Blind structural moves | Unconstrained directory moves can cause drift | Plan by batch, validate by batch, rollback by batch |
| Moving interdependent types | Types that reference each other -> cycles when split | Keep types flat or analyze dependencies first |
| Organizing by artifact type | Directories like `types/`, `utils/`, `helpers/` hide change boundaries | Organize by change dimension (what changes together) |
| Technical layer organization | `internal/service/`, `internal/repository/` mix all domains | Organize by domain: `internal/user/`, `internal/order/` |
| Misleading directory names | `adapters/` containing non-adapter code breaks expectations | Name directories accurately by their actual concern |

## What to Look For

### Priority 1: Cohesion Gaps (Global Structure)

- Large flat directories mixing multiple concerns
- Same concern scattered across different folders
- High-traffic modules lacking concern-based grouping
- Naming drift blurring boundaries and ownership
- **Misleading directory names** — directories named by artifact type (`adapters/`, `utils/`, `helpers/`) instead of actual concern
- **Violation of Common Closure Principle** — code that changes for different reasons living in the same module

### Priority 2: Change Dimension Misalignment

- **Mixed change drivers** — files that change for completely different reasons in same directory
- **Artifact-type organization** — grouping by "what it is" (types, interfaces, classes) instead of "why it changes"
- **Technical layer organization** — `internal/service/`, `internal/repository/` mixing all domains instead of organizing by domain (`internal/user/`, `internal/order/`)
- **Platform-specific code in platform-agnostic packages** — framework-specific utilities in shared libraries

### Priority 3: Redundancy (DRY violations)

- Duplicate logic across packages
- Repeated validation and error handling patterns
- Parallel type definitions that should converge

### Priority 4: Responsibility Misplacement (SRP violations)

- One package doing multiple unrelated jobs
- I/O mixed into pure algorithm paths
- Cross-layer dependency direction violations

### Priority 5: Unnecessary Complexity (KISS violations)

- Over-engineered wrappers and helpers
- Excessive indirection with little behavioral value
- Fragmented file layouts hurting onboarding and maintenance

## Workflow

### Phase 1: Analyze (Global First)

Start from repository-level structure, then drill into packages.

**Step 1 — Build global tree map:**

Scan all packages, count non-test files, identify flat mega-packages (30+ files).

**Step 2 — Deep-analyze packages:**

For each major package:
1. Count files and identify overgrowth
2. Tag files by concern (format, validation, sync, types, io, cli)
3. **Identify change dimensions** — group files by "why they change" (specification updates, API changes, framework versions, business rules)
4. **Flag misleading names** — directories named by artifact type instead of concern
5. Flag mixed-concern clusters and ownership ambiguity
6. Verify package placement (internal/ vs pkg/)

**Step 4 — Apply Common Closure Principle:**

For each directory, ask:
1. **What causes these files to change?** (specification updates, API changes, framework upgrades, business rule changes)
2. **Do all files in this directory change for the same reason?** If no, split by change dimension
3. **Are files that change together scattered across directories?** If yes, consolidate
4. **Is the directory name accurate?** Does it reflect the actual concern, or is it a misleading artifact-type name?

Search for duplicate patterns: error sentinels, struct definitions, interface definitions, constructor patterns.

**Step 4 — Analyze import cycle risk (critical):**

Before proposing any subdirectory reorganization, check for potential import cycles:

- **HIGH RISK**: Files within package cross-import each other -> keep flat or be extremely careful
- **MEDIUM RISK**: Files share common types but don't cross-import -> use `internal/` pattern
- **LOW RISK**: Files are completely independent -> safe to reorganize with subdirectories

### Phase 2: Generate Global Restructuring Plan

Write `.plans/YYYY-MM-DD-RN-global-cohesion-refactor.md`:

```markdown
# Global Cohesion Refactor Review RN

## Scope
Mode: full repository audit
Packages analyzed: {list}
Tree snapshot: YYYY-MM-DD
Total non-test files: {count}

## Executive Summary
{2-3 paragraphs summarizing key findings and proposed changes}

## Import Cycle Risk Analysis

### HIGH RISK (Do NOT reorganize)
{Packages with internal cross-file dependencies}

### MEDIUM RISK (Reorganize with caution)
{Packages with shared types but no cross-imports}

### LOW RISK (Safe to reorganize)
{Packages with independent files}

## Findings

### 1. {Finding title}
**Category:** Cohesion / DRY / SRP / OCP / KISS
**Severity:** HIGH / MEDIUM / LOW
**Location:** {package or file path}
**Current:** {current structure}
**Proposed:** {target grouping or simplification}
**Justification:** {why this improves global architecture}

## Proposed Target Structure
{Package/folder structure grouped by concern}

## Rejected Suggestions
{Items considered but rejected, with reasoning}
```

### Phase 3: Execute Restructuring (Concurrent Batches)

After user approval:
1. Split into independent batches by package
2. Execute batches concurrently where safe (no cross-dependencies)
3. Validate each batch: build, test, lint, import cycle check
4. If any batch fails, revert that specific batch and continue others

### Phase 4: Simplify

After structural changes:
- Run a simplification pass on changed code
- Remove wrappers/helpers that restructuring made unnecessary
- Re-check cohesion and navigability

## Decision Framework

Before suggesting any restructuring, ask:

```
1. Does this reduce global navigation cost?
   YES -> likely good    NO -> needs strong justification

2. Does this improve package-level cohesion?
   YES -> likely good    NO -> probably not worth it

3. Would a new contributor find AFTER easier than BEFORE?
   AFTER -> proceed    BEFORE -> don't restructure

4. Does this preserve tests and contracts?
   YES -> proceed       NO -> redesign the approach

5. Can this be executed and rolled back as one batch?
   YES -> proceed       NO -> split further

6. Does this create new public import paths?
   NO (using internal/) -> safe    YES -> high cycle risk

7. Do files within the package cross-import each other?
   NO -> safe to reorganize    YES -> keep flat or analyze deps first

8. Are we moving types that other files depend on?
   NO -> safe    YES -> check for circular dependencies first
```

## When to Use This vs code-refactoring

| Use code-restructuring | Use code-refactoring |
|---|---|
| Package structure globally degraded | After completing a batch of features |
| Flat mega-packages (30+ files) need concern-based grouping | Code smells in recently changed packages |
| Focus on repo-wide cohesion gaps | Focus on DRY/SRP in git diff scope |
| Structural reorganization (subdirectories, package splits) | Incremental cleanup after feature work |
| Scope: entire package tree | Scope: last 4 commits |

## Language-Specific Guidelines

This skill provides language-agnostic principles. For language-specific project layout, package organization, and internal patterns, refer to the `references/` directory:

**Go (1.26+)**: `references/go-project-layout.md`, `references/go-package-organization.md`, `references/go-internal-patterns.md`

**TypeScript (5.7+)**: `references/ts-project-layout.md`, `references/ts-module-organization.md`, `references/ts-monorepo-patterns.md`

## Remember

- **Global first** — never start from local diff assumptions
- **One batch, one rollback boundary** — keep execution controllable
- **Restructuring reduces complexity** — avoid adding architectural ceremony
- **Split first, then merge** — SRP splits first, cohesion merges second
- **Behavior unchanged** — tests and contracts remain the source of truth
- **Flat over nested** — prefer `pkg/foo/bar` over `pkg/foo/internal/bar`
- **Analyze dependencies first** — check cross-file imports before moving files
- **Types are sticky** — moving types to subdirectories usually creates cycles; keep flat
- **Independent files are safe** — analyzers, tools, processors can be reorganized
- **Validate immediately** — run build and test after each batch; revert on failure
- **Common Closure Principle** — organize by change reason, not by artifact type
- **Name by concern, not by type** — accurate names build trust; misleading names break it

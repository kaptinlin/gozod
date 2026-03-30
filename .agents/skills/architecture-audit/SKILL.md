---
description: Use when performing monthly full-codebase health checks, detecting circular dependencies, finding fat interfaces, checking layer boundary violations, measuring quantitative degradation, or validating SPECS alignment. Systematic architecture review independent of recent changes.
name: architecture-audit
---


# Architecture Audit

Systematic, comprehensive codebase health check independent of recent changes.

## Overview

Architecture audit is the **systemic layer** of the three-tier quality system:
- **Tactical** (code-simplifying): Cleanup of just-written code
- **Strategic** (code-refactoring): Elimination of redundancy after feature batches
- **Systemic** (architecture-audit): Full-codebase health check

Unlike code-refactoring which starts from recent changes, architecture audit scans the entire codebase for structural issues that accumulate over time.

## What Architecture Audit Does

Architecture audit performs **seven-phase systematic review** from code quality to system design:

**Phase 0: Code Quality Baseline** (via linter)
- Run project linter to establish code quality baseline (see `references/commands.md`)
- Ensures code formatting, style, and basic errors are resolved
- Provides clean foundation for higher-level analysis

**Phase 1: Dependency Analysis**
- Detects circular dependencies between packages
- Verifies layer boundaries (cmd → pkg → internal)
- Measures package coupling scores

**Phase 2: Interface Health**
- Finds fat interfaces (ISP violations)
- Identifies single-implementation interfaces (premature abstraction)
- Analyzes interface method count distribution

**Phase 3: Abstraction Consistency**
- Verifies pure core vs I/O separation
- Checks logic/presentation mixing
- Audits state management patterns

**Phase 4: Pattern Consistency**
- Verifies architectural pattern uniformity across packages
- Checks API design consistency
- Audits error handling architecture

**Phase 5: Quantitative Metrics**
- Measures package-level metrics (size, coupling, cohesion)
- Analyzes test coverage gaps
- Identifies refactor candidates

**Phase 6: SPECS Alignment**
- Detects implementation vs SPECS drift
- Finds undocumented architectural decisions
- Identifies convention violations

## When to Use

Use architecture-audit when:
- Before major releases (baseline assessment)
- After accumulating significant feature implementations
- Preparing for external code review
- Technical debt has reached threshold
- Need comprehensive structural health check

## When NOT to Use

Don't use architecture-audit for:
- Just finished writing code → use code-simplifying
- Completed feature batch → use code-refactoring
- Need quick fix for single issue → fix directly
- Recent changes only → use code-refactoring

## Seven Audit Phases

### Phase 0: Code Quality Baseline

**Goal:** Establish clean code quality baseline before architectural analysis

**Checks:**
- Run project linter (see `references/commands.md` for language-specific commands)
- Resolve formatting, style, and basic errors
- Ensure function-level metrics pass (complexity, dead code)

**Why first:** Architecture analysis requires clean code foundation. Linter catches function-level issues; audit catches system-level issues.

### Phase 1: Dependency Analysis

**Goal:** Detect circular dependencies, verify layer boundaries, check import direction

**Checks:**
- Circular dependencies between packages
- Layer boundary violations (e.g., pkg importing internal)
- Import direction (should flow: cmd → pkg → internal)
- Package coupling scores

**Tools:** See language-specific references

### Phase 2: Interface Health

**Goal:** Find fat interfaces, detect unused methods, identify single-implementation interfaces

**Checks:**
- Fat interfaces (>3 methods violate ISP)
- Unused interface methods (defined but never called)
- Single-implementation interfaces (premature abstraction)
- Interface method count distribution

**ISP Threshold:** 1-3 methods = good, 4-5 = acceptable, 6+ = review needed

### Phase 3: Abstraction Consistency

**Goal:** Verify pure vs I/O separation, check logic/presentation mixing, audit state management

**Checks:**
- Pure core violated (algorithms mixed with I/O)
- Logic mixed with presentation (business rules in UI)
- State management patterns inconsistent
- Side effects not pushed to boundaries

**Pattern:** Pure functions accept data, return results. I/O happens in outer layers.

### Phase 4: Pattern Consistency

**Goal:** Verify architectural pattern uniformity across packages

**Checks:**
- API design consistency (similar operations use similar signatures across packages)
- Context propagation patterns (first parameter convention across system)
- Constructor patterns (New vs With* options consistency)
- Error handling architecture (sentinel error organization, wrapping strategy)

**Note:** Individual error handling issues (missing wraps, nil checks) are caught by golangci-lint. Focus on system-wide patterns.

### Phase 5: Quantitative Metrics

**Goal:** Measure package-level and system-level metrics

**Checks:**
- Package size >2000 LOC (split candidates)
- Package coupling (imports >15 packages)
- Test coverage <80% (coverage gaps)
- Package cohesion (unrelated responsibilities in same package)

**Thresholds:**
- Max package LOC: 2000
- Max package imports: 15
- Min test coverage: 80%
- Max package files: 20

**Note:** Function-level complexity (cyclomatic, funlen) is handled by golangci-lint. Focus on package-level metrics.

### Phase 6: SPECS Alignment

**Goal:** Detect implementation vs SPECS drift, find undocumented decisions, identify convention violations

**Checks:**
- Implementation deviates from SPECS
- Architectural decisions not documented
- Naming conventions violated
- Design principles ignored

**SPECS to check:** Architecture, naming, design principles, patterns

## Workflow

Architecture audit generates a single comprehensive document (`.plans/YYYY-MM-DD-arch-audit.md`) containing both findings and action plan.

Implementation is done separately, following the action plan.

### Step 1: Prepare

1. Determine audit scope (full codebase or specific modules)
2. Create audit report file: `.plans/YYYY-MM-DD-arch-audit.md`

### Step 2: Execute Seven Phases

Execute phases in order, documenting findings:

```markdown
# Architecture Audit YYYY-MM-DD

## Phase 0: Code Quality Baseline

**Command:** See `references/commands.md` for linting commands

**Status:** [PASS | FAIL]
**Issues Found:** N
**Action:** [All resolved | Deferred to separate task]

## Phase 1: Dependency Analysis
[findings]

## Phase 2: Interface Health
[findings]

## Phase 3: Abstraction Consistency
[findings]

## Phase 4: Pattern Consistency
[findings]

## Phase 5: Quantitative Metrics
[findings]

## Phase 6: SPECS Alignment
[findings]
```

### Step 3: Prioritize Findings

Classify by severity:
- **HIGH**: Affects correctness, security, or public API
- **MEDIUM**: Affects maintainability or readability
- **LOW**: Cosmetic or minor improvements

### Step 4: Append Action Plan

Append action plan to the same audit report file (`.plans/YYYY-MM-DD-arch-audit.md`):

```markdown
## Action Plan

### Priority Breakdown

**HIGH Priority (Immediate - This Sprint):**
- [ ] Finding 1: [description]
  - Package: [location]
  - Action: [specific fix]
  - Estimate: [hours]
  - Dependencies: [none | Finding N]

**MEDIUM Priority (Short-Term - Next Month):**
- [ ] Finding 3: [description]
  - Package: [location]
  - Action: [specific fix]
  - Estimate: [hours]

**LOW Priority (Long-Term - Monitoring):**
- [ ] Finding 5: [description]
  - Package: [location]
  - Action: [specific fix]
  - Estimate: [hours]

### Implementation Order

1. Phase 0 fixes (if needed): Resolve linter issues
2. Phase 1 fixes: Break circular dependencies
3. Phase 2 fixes: Split fat interfaces
4. Phase 3 fixes: Separate concerns
5. Phase 4 fixes: Standardize patterns
6. Phase 5 fixes: Reduce package size/coupling
7. Phase 6 fixes: Update documentation

### Validation Checklist

After implementing each phase:
- [ ] Run linter (see `references/commands.md`)
- [ ] Run tests (see `references/commands.md`)
- [ ] Re-run phase-specific checks
- [ ] Commit with reference to audit finding

### Timeline

**Week 1:** HIGH priority fixes
**Week 2-3:** MEDIUM priority fixes
**Week 4:** LOW priority fixes + final validation
```

**Note:** This skill generates the audit report with action plan. Implementation is done separately, following the action plan.

## Report Template

```markdown
# Architecture Audit Report

**Date:** YYYY-MM-DD
**Scope:** [full codebase | specific modules]
**Auditor:** [name]

## Executive Summary

**Overall Health:** [GOOD | FAIR | POOR] (score/10)
**Critical Issues:** N
**High Priority:** N
**Medium Priority:** N

## Findings by Phase

### Phase 0: Code Quality Baseline
- **Linter:** See `references/commands.md`
- **Status:** [PASS | FAIL]
- **Issues Found:** N
- **Action:** [All resolved | Deferred to separate task]

### Phase 1: Dependency Analysis
- Finding 1: [description]
  - Location: [file:line]
  - Severity: [HIGH|MEDIUM|LOW]
  - Recommendation: [action]

[repeat for all phases]

## Metrics Summary

| Metric | Current | Threshold | Status |
|--------|---------|-----------|--------|
| Package size (max LOC) | X | 2000 | ✓/✗ |
| Package coupling (max imports) | X | 15 | ✓/✗ |
| Test coverage | X% | 80% | ✓/✗ |
| Circular dependencies | X | 0 | ✓/✗ |
| Layer violations | X | 0 | ✓/✗ |

**Note:** Function-level metrics (cyclomatic complexity, dead code) are handled by the project linter.

## Action Plan

### Priority Breakdown

**HIGH Priority (Immediate - This Sprint):**
- [ ] Finding 1: [description]
  - Package: [location]
  - Action: [specific fix]
  - Estimate: [hours]
  - Dependencies: [none | Finding N]

**MEDIUM Priority (Short-Term - Next Month):**
- [ ] Finding 3: [description]
  - Package: [location]
  - Action: [specific fix]
  - Estimate: [hours]

**LOW Priority (Long-Term - Monitoring):**
- [ ] Finding 5: [description]
  - Package: [location]
  - Action: [specific fix]
  - Estimate: [hours]

### Implementation Order

1. Phase 0 fixes (if needed): Resolve linter issues
2. Phase 1 fixes: Break circular dependencies
3. Phase 2 fixes: Split fat interfaces
4. Phase 3 fixes: Separate concerns
5. Phase 4 fixes: Standardize patterns
6. Phase 5 fixes: Reduce package size/coupling
7. Phase 6 fixes: Update documentation

### Validation Checklist

After implementing each phase:
- [ ] Run linter (see `references/commands.md`)
- [ ] Run tests (see `references/commands.md`)
- [ ] Re-run phase-specific checks
- [ ] Commit with reference to audit finding

### Timeline

**Week 1:** HIGH priority fixes
**Week 2-3:** MEDIUM priority fixes
**Week 4:** LOW priority fixes + final validation
```

## Language-Specific Commands

This skill provides language-agnostic audit phases. For language-specific commands and tools, see `references/commands.md` in the distributed skill package.

## Integration with Quality System

**Three-Tier Quality System:**
- **Tactical**: code-simplifying (cleanup after writing code)
- **Strategic**: code-refactoring (redundancy elimination after features)
- **Systemic**: architecture-audit (full-codebase health check when needed)
- **Documentation**: SPECS alignment review (documentation sync as needed)

**Workflow Integration:**
1. Implement features with code-simplifying after each
2. Refactor after feature batches with code-refactoring
3. Audit when needed with architecture-audit
4. Update SPECS based on audit findings

## Common Mistakes

| Mistake | Fix |
|---------|-----|
| Skipping phases | Execute all seven phases systematically |
| Surface-level checks | Use quantitative tools, not just visual inspection |
| Confirmation bias | Look for problems, not confirmation of health |
| Scope creep | Stick to audit scope, log other issues separately |
| No action plan | Convert findings to actionable plan with timeline |
| Vague recommendations | Provide specific, actionable fix steps in plan |
| Missing estimates | Include time estimates for each fix in action plan |
| No prioritization | Clearly mark HIGH/MEDIUM/LOW severity |

## Success Criteria

Audit is complete when:
- [ ] Phase 0: Linter executed and baseline established
- [ ] All seven phases executed
- [ ] Findings documented with severity
- [ ] Metrics measured and compared to thresholds
- [ ] Action plan appended with priorities and timeline
- [ ] Complete audit report (with action plan) saved to `.plans/YYYY-MM-DD-arch-audit.md`

**Note:** This skill generates audit report with action plan. Implementation is done separately.

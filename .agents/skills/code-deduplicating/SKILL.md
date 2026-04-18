---
description: Extract repeated code into shared utilities. Identifies duplicate patterns (validation, client setup, file ops, string utils) appearing 3+ times, extracts to reusable functions following DRY/KISS. Use when duplicates accumulate, after feature work, or user says "extract common code", "reduce duplication", "DRY violations".
name: code-deduplicating
---


# Code Deduplication

Extract repeated code patterns into shared utilities. Eliminate duplication while keeping it simple.

**Announce at start:** "I'm using the code-deduplicating skill."

## Core Principles

1. **3+ Rule** — Extract when pattern appears 3 or more times
2. **KISS** — Extract only truly reusable code, avoid premature abstraction
3. **Single Responsibility** — One utility = one purpose
4. **Minimal Surface** — Keep signatures small; pass only what’s needed
5. **Locality First** — Prefer module-local helpers before global utilities
6. **Consistent Naming** — Use domain nouns + verbs (e.g., `validateEmail`, `buildClient`)
7. **Symmetric Behavior** — Same inputs produce same outputs and errors
8. **No Hidden Dependencies** — Avoid implicit globals or mutable shared state

## Style-Guide Principles

- **Clarity over cleverness** — Prefer code that is easiest for readers to understand
- **Simplicity first** — Use the simplest mechanism that works; avoid extra abstraction
- **High signal-to-noise** — Reduce repetition and boilerplate that hide the real logic
- **Prefer existing tools** — Use standard library or existing project helpers before adding new ones
- **Consistency wins** — Align with nearby naming and structure unless it harms clarity

## What to Extract

Prioritize repeatable, stable logic that appears in many places:

1. **Input validation** — Required fields, trimming, format checks
2. **Resource/client setup** — Auth + config + timeouts + retries
3. **File/IO routines** — Ensure directories, read/write helpers
4. **String normalization** — Slugify, case folding, trimming
5. **Request/response handling** — Status checks + decoding

When duplicates are similar-but-not-identical, extract the stable core and leave the variants inline.

## Workflow

### Step 1: Find Duplicates

Look for copy‑paste blocks, near‑duplicates with minor changes, or repeated sequences of 3–10 lines.

**Search tactics:**
- Search repeated error messages or guard clauses
- Search for identical call sequences
- Use diff tools to compare similar functions
- **Check for existing helpers** in the repo before extracting anything new

**Threshold:**
- 3–5 times: Consider
- 6–10 times: Should extract
- 10+ times: Must extract

### Step 2: Extract Function

**Placement (Domain-Driven):**
- Reuse existing helpers first (DRY)
- **Domain-specific logic** → stays in domain package (e.g., `internal/user/validator.go`)
- **Cross-domain utilities** → shared package (e.g., `internal/validation/`, `pkg/stringutil/`)
- Start local (same module/package) before creating a global utility
- Keep file names aligned with purpose (e.g., `validate`, `client`, `fs`, `string`)

**Domain-driven placement rules:**
- If logic is specific to one domain (user, order, product) → keep in that domain package
- If logic is used by 2+ domains → extract to shared utility package
- Avoid creating unified `utils/` or `helpers/` — name by actual purpose
- Example: User email validation → `internal/user/validator.go`, not `internal/validator/user.go`

**Write minimal code:**
- Single purpose
- Clear names
- Early returns
- No side effects

### Step 3: Replace Usage

**One file at a time:**
1. Add import
2. Replace inline code
3. Run tests
4. Commit

### Step 4: Verify

Run your project’s build/test commands.

**Example (Node):**
```bash
pnpm run build && pnpm test
```

## When NOT to Extract

1. **Only 2 occurrences** — wait for third
2. **Different logic** — don’t force unification
3. **Domain-specific** — keep in domain module
4. **One-liner** — inline is clearer
5. **Temporary** — marked TODO/FIXME

## Anti-Patterns

| Anti-Pattern | Why Wrong | Do Instead |
|---|---|---|
| Extract at 2 uses | Premature | Wait for 3rd |
| Generic names | Hard to discover | Name by purpose |
| 5+ parameters | Too complex | Split functions |
| Cross-domain | Mixed concerns | Separate by domain |
| Over-generalization | Hard to use | Extract stable core only |

## Checklist

- [ ] Pattern appears 3+ times
- [ ] Logic is identical or has a stable core
- [ ] Single clear purpose
- [ ] Descriptive name
- [ ] No hidden dependencies
- [ ] Tests pass
- [ ] No new dependencies

## References

- `references/deduplication-examples.md` (language-specific examples when available)

## Summary

Extract when a stable pattern appears 3+ times. Keep utilities focused, local first, and avoid over‑generalizing. Use small, explicit helpers to reduce AI‑generated redundancy without creating heavyweight abstractions.

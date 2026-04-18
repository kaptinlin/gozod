---
description: Implements features using strict TDD discipline with vertical red-green-refactor cycles, writing one failing test before each piece of production code. Use when executing an Impl task in Code phase, or when asked to "write tests first" or "red-green-refactor".
name: tdd-implementing
---


# TDD Implementing

Implement features using strict TDD discipline: write failing test first (RED), write minimal code to pass (GREEN), refactor when all tests pass. Tests verify behavior through public interfaces, not implementation details.

**Used in:** Code phase (workflow stage 3) - Impl tasks

**Announce at start:** "I'm using the tdd-implementing skill."

## Pipeline Position

```
Code Phase Workflow:

tdd-tasking → Ralphy loop:
                ├─ Plan task  → tdd-planning → .plans/YYYY-MM-DD-{feature}.md
                ├─ Impl task  → tdd-implementing:
                │                 0. Study (.references/ + existing code)
                │                 1. Plan (interface + behaviors)
                │                 2. Tracer bullet (first RED→GREEN)
                │                 3. Incremental loop (remaining criteria)
                │                 4. Refactor (GREEN only)
                └─ Checkpoint → code-simplifying / code-refactoring
```

## Philosophy

**Core principle**: Tests verify behavior through public interfaces, not implementation details. Code can change entirely; tests shouldn't.

**Good tests** exercise real code paths through exported APIs. They describe _what_ the system does. `TestOrder_ConfirmsOnValidPayment` tells you exactly what capability exists and survives a complete internal refactor.

**Bad tests** are coupled to implementation: they mock internal collaborators, assert on call order, verify through side-channel means (querying the DB directly), or lock prose/spec/docs wording inside test files. The warning sign: test breaks when you rename an internal function, rewrite a README section, or rephrase a spec while behavior hasn't changed.

**Never add documentation-lock test files** like `research_test.go`, `*_spec_test.go`, or tests that read `README.md`, `CLAUDE.md`, `SPECS/*.md`, `.research/*` just to assert headings, anchor text, fixed wording, or narrative structure. If the invariant matters to runtime or API consumers, test behavior through code. If it is an architecture/process rule, enforce it with lint/CI checks or keep it in docs — not in unit tests.

## Iron Law

```
NO PRODUCTION CODE WITHOUT A FAILING TEST FIRST
```

Write code before the test? Delete it. Start over. Implement fresh from tests only.

**No exceptions:** don't keep it as "reference", don't "adapt" it while writing tests. Delete means delete.

## Constraints ≠ Code

**Development constraints ≠ production code.**

Constraints are rules about how code should be written — not executable behavior. They belong in three places:

| Constraint Type | Enforcement Mechanism |
|-----------------|----------------------|
| Forbidden values, required patterns, validation rules | **Runtime code + tests** — users observe the enforcement |
| Module import restrictions, layer boundaries | **Lint rules / CI scripts** — catch violations in build |
| Architecture rules, naming conventions, documentation structure | **SPEC documentation + CI checks** — code reviews and automation |

**The deletion test:** If deleting the code doesn't change program behavior, it's documentation disguised as code.

Type unions, compile-time assertions, and identity "validators" that no runtime code reads are not enforcement — they're documentation that developers can ignore. If a constraint matters, write executable validation that tests verify.

## Test Types

| Type | What | When |
|------|------|------|
| **Unit** | Pure functions, transformations, business logic | No I/O, no external dependencies |
| **Integration** | Code with DB, HTTP clients, file system | Test real I/O with test doubles at system boundary |
| **HTTP handler** | HTTP test recorder/server | API endpoints, middleware |

Most code should be covered by unit tests. Integration tests verify the wiring.

## Anti-Pattern: Horizontal Slices

**Never write all tests first, then all implementation.**

```
WRONG (horizontal):
  RED:   test1, test2, test3, test4, test5
  GREEN: impl1, impl2, impl3, impl4, impl5

RIGHT (vertical):
  RED→GREEN: test1→impl1
  RED→GREEN: test2→impl2
  RED→GREEN: test3→impl3
```

Horizontal slices produce imagined-behavior tests — you test the _shape_ of data structures rather than user-facing behavior. Tests become insensitive to real changes.

## Workflow

### 0. Study (Before Any Code)

Before touching any code, study three sources:

**a) The implementation plan** — read `.plans/` file for this feature. It defines the interface contract and acceptance criteria. Do not deviate.

**b) Reference implementations** — check `.references/` for the closest analog:
- Read the reference's source code for API patterns and error handling
- Read the reference's tests (or test-equivalent code) for edge cases you must cover
- Note auth chains, pagination, error mapping, and name conventions
- **Borrow patterns, not code** — translate to the project's conventions

**c) Existing codebase patterns** — read at least 2 similar implementations in the project:
- Match the constructor pattern (functional options, validation, nil guards)
- Match the error pattern (sentinel errors, error wrapping style)
- Match the test pattern (setup helpers, assertion framework, table-driven style)
- Match the concurrency pattern (atomic flags, mutexes, async patterns, etc.)

**Why this matters:** Implementing without studying produces inconsistent code. Studying first means the first TDD cycle produces code that already fits the project.

### 1. Plan

Before any code:

- [ ] What does the public interface look like? (function signatures, types, sentinel errors)
- [ ] Which behaviors matter most? List them — you can't test everything
- [ ] Which test type fits each behavior? (unit / integration / HTTP handler)
- [ ] Can this be a deep module? Small interface hiding complex logic = simpler tests
- [ ] Design for testability: accept dependencies as parameters, return values over side effects
- [ ] Get user approval on the behavior list before writing any test

Ask: "What should the exported interface look like? Which behaviors are most important?"

### 2. Tracer Bullet

Write ONE test that confirms ONE end-to-end behavior:

```
RED:   Write test for first behavior → run → MANDATORY: watch it fail
GREEN: Write minimal code to pass → run → MANDATORY: watch it pass
```

This is your tracer bullet — proves the path works at all.

**Verify RED (MANDATORY — never skip):** Confirm: test fails (not errors). Failure message is expected. Fails because feature is missing, not a typo.

**Test passes immediately?** You're testing existing behavior. Fix the test first.

**Test errors?** Fix the error, re-run until it fails correctly.

```language-agnostic
// Test for first behavior - confirm it fails
test("Feature_HappyPath", () => {
    result = DoThing(validInput)
    assert.NoError(result.error)
    assert.Equal(expected, result.value)
})
```

Expected first run: `FAIL` — function/feature not defined

**Note:** See `references/test-framework.md` in your project templates for language-specific testing patterns and assertions.

### 3. Incremental Loop

For each remaining behavior:

```
RED:   Write next test → run → MANDATORY: watch it fail
GREEN: Minimal code to pass this test only → run → MANDATORY: watch it pass
```

Rules:
- One test at a time
- Only enough code to pass the current test
- Don't anticipate future tests — write them when you get there
- Keep tests focused on observable behavior
- Do not create tests that parse docs/spec files to lock wording, anchors, headings, or prose structure

### 4. Refactor

After all planned tests pass:

- [ ] Extract duplicated setup into helpers
- [ ] Deepen modules: push complexity behind simpler interfaces
- [ ] Delete dead code revealed by the tests
- [ ] Run race detector (for concurrent code) after each refactor step
- [ ] **Delete "documentation disguised as code" revealed by tests**
  - Type unions that only list forbidden/required values but nothing reads
  - "Assertion" types that are identity functions
  - Constraint declarations that aren't enforced by executable tests

**Never refactor while RED.** Get to GREEN first.

## Common Rationalizations

| Excuse | Reality |
|--------|---------|
| "Too simple to test" | Simple code breaks. Test takes 30 seconds. |
| "I'll test after" | Tests passing immediately prove nothing. |
| "Already manually tested" | Ad-hoc ≠ systematic. No record, can't re-run. |
| "Deleting X hours is wasteful" | Sunk cost fallacy. Unverified code is technical debt. |
| "Keep as reference, write tests first" | You'll adapt it. That's testing after. Delete means delete. |
| "Need to explore first" | Fine. Throw away exploration, start with TDD. |
| "Test hard = design unclear" | Listen to the test. Hard to test = hard to use. |
| "TDD will slow me down" | TDD is faster than debugging production bugs. |

## Red Flags — STOP and Start Over

- Code written before test
- Test passes immediately (without implementation)
- Can't explain why the test failed
- "I already manually tested it"
- "Just this once" rationalization
- "Keep as reference" or "adapt existing code"
- "Already spent X hours, deleting is wasteful"

**All of these mean: Delete code. Start over with TDD.**

## When Stuck

| Problem | Solution |
|---------|----------|
| Don't know how to test | Write the wished-for API. Write assertion first. |
| Test too complicated | Design too complicated. Simplify the interface. |
| Must mock everything | Code too coupled. Use dependency injection. |
| Test setup huge | Extract helpers with `t.Helper()`. Still complex? Simplify design. |

## Coverage Targets

| Code | Target |
|------|--------|
| Core business logic | 100% of exported behaviors |
| Each sentinel error (`ErrXxx` / `XxxError`) | 100% — every error callers handle |
| General public API | 80%+ |
| HTTP handlers | Happy path + error responses |
| Trivial constructors / getters | Skip |

Run coverage tool per your language ecosystem.

## Per-Cycle Checklist

```
[ ] Test describes behavior, not implementation
[ ] Test uses exported API only — no internal state
[ ] Test would survive renaming an internal function
[ ] Test would survive README/CLAUDE/SPEC prose rewrites when behavior is unchanged
[ ] Parallel test enabled (unless mutating shared state)
[ ] Implementation is minimal for this test only
[ ] No speculative code added
[ ] Test passes before moving on
[ ] No "documentation disguised as code" added
[ ] No documentation-lock test file added
```

## Testing Constraints vs Implementing Them

When specs define constraints (forbidden values, required patterns, validation rules):

**❌ Wrong: Implement constraints as types only**
```typescript
// Documentation disguised as code — nothing reads this at runtime
type ForbiddenAction = 'DELETE_ALL' | 'FORMAT_DISK'
type AssertValidAction<T> = T extends ForbiddenAction ? Error : T
```
No runtime code reads these. If a developer doesn't use the type, it's unenforced.

**✅ Right: Implement constraints as runtime validation**
```language-agnostic
// Actual code that executes and enforces the rule
function validateAction(action: string) {
    const forbidden = ['DELETE_ALL', 'FORMAT_DISK']
    if (forbidden.includes(action)) {
        throw new Error(`Forbidden action: ${action}`)
    }
}
```

**✅ Best: Test enforces the constraint**
```language-agnostic
test("rejects forbidden actions", () => {
    expect(() => validateAction('DELETE_ALL'))
        .toThrow("Forbidden action")
})
```

**For architectural constraints**, use this approach:
1. **SPEC** as source of truth — document the rule clearly
2. **CI check** — lint rule or validation script that runs in build
3. **Test** — verify the CI check catches violations

Do **not** replace CI/lint enforcement with unit tests that read markdown docs and assert exact prose. That creates brittle doc-lock tests instead of executable behavior checks.

## Code Comment Guidelines (KISS + DRY)

**Core principle**: Comments explain _why_ and _what_, not repeat what code already says.

### ✅ Keep These Comments

1. **Function documentation** — Parameters, return values, examples
2. **Key behavior differences** — When behavior isn't obvious from types
3. **SPEC/design references** — Link to authoritative sources
4. **Non-obvious algorithms** — Complex logic needing explanation

### ❌ Delete These Comments

1. **Type-repeating comments** — Types already document themselves
   ```language-agnostic
   // ❌ Bad: name: string; /** Component name */
   // ✅ Good: name: string;
   ```

2. **Obvious field descriptions** — Self-explanatory names
3. **Decorative separators** — `// ─── Core Types ───────`
4. **Redundant function descriptions** — Function name says it all
5. **Implementation details in interfaces** — Keep interfaces clean

### Refactoring Checklist

```
[ ] Does this comment explain WHY, not WHAT?
[ ] Would deleting it make code harder to understand?
[ ] Is it repeating what types already say?
[ ] Can I make code clearer instead of adding a comment?
```

**Rule of thumb**: If comment just restates code in natural language, delete it.

## Consistency Checklist

Before the first TDD cycle, verify you can answer these from studying existing code:

```
[ ] Constructor pattern: Functional options? Builder? Config struct/object?
[ ] Error pattern: Sentinel errors? Wrapping? Custom error types?
[ ] Cleanup pattern: Atomic flags? Once/defer patterns? Async cleanup?
[ ] Test setup: Shared helpers? Setup/teardown hooks?
[ ] Test assertions: Which framework? Assertion style?
[ ] Interface compliance: Compile-time checks? Type annotations?
[ ] Context/cancellation: How is cancellation handled?
[ ] Naming: Language-specific conventions? Package/module structure?
```

**Note:** See `references/code-patterns.md` in your project templates for language-specific patterns.

If you don't match these, your code will look foreign in the codebase. Study first, code second.

## Sub-Module Setup

When implementing a new sub-module (separate package/module):

1. **First TDD cycle**: Create module configuration file (e.g., `go.mod`, `package.json`) with proper dependencies
2. **Compile-time check**: Add interface compliance verification appropriate for your language
3. **Test imports**: Use the parent module's test helpers if they exist
4. Run dependency management commands after adding each dependency — don't batch

**Note:** See `references/code-patterns.md` in your project templates for language-specific sub-module setup.

## Remember

- This skill is used in Code phase for Impl tasks generated by `tdd-tasking`
- **Study references and existing code BEFORE the first TDD cycle** — not after
- Always follow the implementation plan from `tdd-planning`
- One feature = many TDD cycles, one commit at the end
- Tests verify behavior through exported interfaces only
- Never write all tests first — vertical slices only
- **Constraints belong in tests, not type declarations**
- **If no runtime code reads it, it's not implementation — it's documentation**
- **Comments should add value, not repeat what code already says (KISS + DRY)**
- **Borrow patterns from .references/, match conventions from existing code**
- **See `references/` in your project templates for language-specific testing frameworks, patterns, and examples**

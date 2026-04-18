---
description: Read root ANALYSIS.md and generate TODO.yaml task list for writing specification documents. Use after spec-analyzing has produced ANALYSIS.md.
name: spec-tasking
---


# Spec Tasking

**Four-phase workflow:** First audits `ANALYSIS.md` accuracy, then generates `TODO.yaml` with spec-writing tasks, then appends a modernizing task to revise specs with modern language features, then appends a final spec-reviewing task for cross-spec alignment.

**Announce at start:** "I'm using the spec-tasking skill."

**Prerequisite:** `ANALYSIS.md` must exist at project root. If not, run `spec-analyzing` first.

## Phase 1: Audit ANALYSIS.md (CRITICAL)

**Before generating tasks, verify each analysis point against actual specs and research.**

### Audit Process

For each spec identified in ANALYSIS.md:

1. **Check if spec already exists**
   - Look for SPECS/{number}-{slug}.md
   - Check if content matches requirements
   - Verify research files exist

2. **Update ANALYSIS.md with corrections**
   - Mark existing specs as ✅ (skip task generation)
   - Correct file paths if wrong
   - Update priority if dependencies found
   - Add notes about actual findings

3. **Document audit results**
   - Add "Audit Notes" section
   - List verified items
   - Note discrepancies

### Example Audit Corrections

**Before audit:**
```markdown
## Priority 1: Authentication (SPECS/02)
Status: Missing
Source: .research/R01-auth-patterns.md
```

**After audit (if spec exists):**
```markdown
## Priority 1: Authentication (SPECS/02)
Status: ✅ Exists
Source: .research/R01-auth-patterns.md
Audit Notes: Found SPECS/02-authentication.md (complete). No task needed.
```

### Audit Checklist

Before proceeding to Phase 2:
- [ ] Checked if each spec file exists
- [ ] Verified research files exist
- [ ] Updated ANALYSIS.md with findings
- [ ] Marked existing specs as ✅
- [ ] Added audit notes
- [ ] Saved updated ANALYSIS.md

**Only proceed to Phase 2 after audit is complete.**

## Phase 2: Generate TODO.yaml

**Only generate tasks for specs marked as missing or incomplete in audited ANALYSIS.md.**

Skip specs marked ✅ (already exist) - they don't need tasks.

## Phase 3: Append Modernizing Task

**After all spec-writing tasks, append a modernizing task** that uses `modernizing` skill to revise specs with modern language features and idioms. This is language-agnostic — the modernizing skill applied depends on the project's language distribution (e.g., Go modernizing for golang-skills, TypeScript modernizing for typescript-skills).

This ensures specs reference current language APIs and patterns before the final review.

## Phase 4: Append Reviewing Task

**After the modernizing task, append a final spec-reviewing task** that uses `spec-reviewing` skill to verify cross-spec alignment, decision consistency, and fix any contradictions.

This ensures the entire batch of specs is coherent before entering implementation.

## Pipeline Position

```
spec-analyzing ──▶ Phase 1: Audit ──▶ Phase 2: Generate ──▶ Phase 3: Modernize ──▶ Phase 4: Review Task
     │                  │                   │                      │                      │
     ▼                  ▼                   ▼                      ▼                      ▼
 ANALYSIS.md    Updated ANALYSIS.md    TODO.yaml            TODO.yaml              TODO.yaml (final)
(may have errors) (verified accurate) (writing tasks)    (+ modernizing task)   (+ reviewing task)
```

This skill is the second step in the spec writing workflow — it audits analysis accuracy, converts verified gaps into actionable tasks, appends a modernizing pass, and appends a final quality gate.

**Design philosophy — think like a founder, not a contractor:**

Say no to a thousand things. Fewer, deeper specs — not more, shallower ones.

- **Simplicity first** — if two specs can be one without loss of clarity, merge them. A well-designed spec prevents future rewrites. Spec bloat is the precursor to code bloat
- **Complete, not over-engineered** — tasks should produce production-quality specifications, not MVP stubs. But don't add sections, rules, or abstractions "just in case"
- **System coherence** — each task must emphasize how this spec integrates with the whole system, not just define its own island
- **No premature flexibility** — don't spec extension points, plugin architectures, or configurability unless the analysis explicitly identifies concrete growth vectors

## Task Structure

Each specification identified in ANALYSIS.md produces exactly **one task** that uses `spec-writing` skill. The second-to-last task uses `modernizing` skill. The final task uses `spec-reviewing` skill.

### Task Format

**CRITICAL: Tasks have only `title` and `completed` fields. No `description` field.**

**Spec-writing task template:**
```yaml
- title: "Write {spec topic} specification based on ANALYSIS.md, use spec-writing skill — spec: SPECS/{number}-{slug}.md | research: .research/{source}.md — define {key-aspects}"
  completed: false
```

**Spec-modernizing task template (second-to-last):**
```yaml
- title: "Modernize all SPECS with current language features based on ANALYSIS.md, use modernizing skill — spec: {all spec paths} | research: {context docs} — revise APIs, types, patterns, and code examples to use modern idioms"
  completed: false
```

**Spec-reviewing task template (always last):**
```yaml
- title: "Review all SPECS for alignment and consistency based on ANALYSIS.md, use spec-reviewing skill — spec: {all spec paths} | research: {context docs} — verify cross-spec type consistency, decision alignment, API coherence, dependency boundaries, fix contradictions"
  completed: false
```

**Rules:**
- `title` contains ALL information: topic, target file, source, and key aspects — everything in one line
- **Both spec and research paths are MANDATORY** — every task must include `spec:` and `research:` plain paths
- **No `description` field** — this is a hard constraint, never include it
- `completed` is always `false`
- Keep title concise but self-contained
- Use the language specified in ANALYSIS.md (Chinese or English)
- Use plain paths (e.g., `SPECS/10-foo.md`) — no markdown link syntax

**Path format:**
```
spec: SPECS/XX-slug.md | research: .research/RXX-slug.md
```

- `spec:` — the SPECS document that defines the contract for this task
- `research:` — the research report containing reference project analysis and evidence
- Both paths are plain relative paths from project root — no markdown link syntax
- If a task spans multiple research reports, list all: `research: .research/R01-xxx.md, .research/R02-xxx.md`

## TODO.yaml Format

```yaml
tasks:
  # ═══════════════════════════════════════════════════════════════
  # Priority 1: Core Module (SPECS/02)
  # ═══════════════════════════════════════════════════════════════

  - title: "Write core module specification based on ANALYSIS.md, use spec-writing skill — spec: SPECS/02-core-module.md | research: .research/R01-core-patterns.md — define types, interfaces, algorithm choices"
    completed: false

  # ═══════════════════════════════════════════════════════════════
  # Priority 2: Integration Layer (SPECS/03)
  # ═══════════════════════════════════════════════════════════════

  - title: "Write integration layer specification based on ANALYSIS.md, use spec-writing skill — spec: SPECS/03-integration.md | research: .research/R02-integration.md — define adapters, error handling, configuration"
    completed: false

  # ═══════════════════════════════════════════════════════════════
  # Modernize — Revise Specs with Current Language Features
  # ═══════════════════════════════════════════════════════════════

  - title: "Modernize all SPECS with current language features based on ANALYSIS.md, use modernizing skill — spec: SPECS/02-core-module.md, SPECS/03-integration.md | research: SPECS/01-requirements.md, DECISIONS.md — revise APIs, types, patterns, and code examples to use modern idioms"
    completed: false

  # ═══════════════════════════════════════════════════════════════
  # Final — Cross-spec Review: Alignment & Consistency
  # ═══════════════════════════════════════════════════════════════

  - title: "Review all SPECS for alignment and consistency based on ANALYSIS.md, use spec-reviewing skill — spec: SPECS/02-core-module.md, SPECS/03-integration.md | research: SPECS/01-requirements.md, DECISIONS.md — verify cross-spec type consistency, decision alignment, API coherence, dependency boundaries, fix contradictions"
    completed: false
```

**YAML rules:**
- **ANALYSIS.md reference:** Every task title MUST include `based on ANALYSIS.md` to maintain traceability
- **Two fields only per task: `title` and `completed`** — never add `description`
- No blank lines between a task's fields
- Blank line between tasks for readability
- Use YAML block comments (`#`) with `═` for visual separation between priority groups
- Comment headers include priority level
- Tasks ordered by priority from ANALYSIS.md
- **Second-to-last task is always spec-modernizing** — it modernizes ALL specs with current language features
- **Final task is always spec-reviewing** — it reviews ALL specs generated by preceding tasks

## Workflow

### Step 1: Audit ANALYSIS.md (Phase 1)

1. Read `ANALYSIS.md` from project root
2. For each spec listed:
   - Check if SPECS/{number}-{slug}.md exists
   - Verify research file exists
   - Update ANALYSIS.md with corrections
3. Save updated ANALYSIS.md
4. Announce: "Audit complete. Found X specs already exist, Y specs need writing."

### Step 2: Read Audited ANALYSIS.md (Phase 2)

Read updated `ANALYSIS.md` and extract only missing/incomplete specs:
- **Scope**: What specifications need to be written
- **Priority**: Ordering of specifications (high/medium/low)
- **Structure**: Mapping of research reports to spec files

**Skip ✅ items** - they don't need tasks.

### Step 3: Determine Task Order

Order tasks by priority from audited ANALYSIS.md:
1. High priority specs first
2. Medium priority specs second
3. Low priority specs last

Within each priority level, maintain the order specified in ANALYSIS.md.

### Step 4: Generate Spec-Writing Tasks

For each missing/incomplete spec in audited ANALYSIS.md, generate exactly one task:

```
Write {spec topic} specification based on ANALYSIS.md, use spec-writing skill — spec: SPECS/{number}-{slug}.md | research: .research/{source}.md — define {key-aspects}
```

### Step 5: Append Spec-Modernizing Task

After all spec-writing tasks, append one modernizing task:

```
Modernize all SPECS with current language features based on ANALYSIS.md, use modernizing skill — spec: {comma-separated list of ALL spec paths from steps above} | research: {context docs: SPECS/01-requirements.md, DECISIONS.md, IMPROVE.md as applicable} — revise APIs, types, patterns, and code examples to use modern idioms
```

The modernizing task:
- Lists ALL spec paths generated above in its `spec:` field
- Lists context documents (requirements, decisions, improve) in its `research:` field
- Uses `modernizing` skill — this is language-agnostic; the actual modernizing skill applied depends on the project's language distribution
- Comes after all spec-writing tasks but before the reviewing task

### Step 6: Append Spec-Reviewing Task

After the modernizing task, append one final task that uses `spec-reviewing` skill:

```
Review all SPECS for alignment and consistency based on ANALYSIS.md, use spec-reviewing skill — spec: {comma-separated list of ALL spec paths from steps above} | research: {context docs: SPECS/01-requirements.md, DECISIONS.md, IMPROVE.md as applicable} — verify cross-spec type consistency, decision alignment, API coherence, dependency boundaries, fix contradictions
```

The reviewing task:
- Lists ALL spec paths generated above in its `spec:` field
- Lists context documents (requirements, decisions, improve) in its `research:` field
- Is always the last task in TODO.yaml
- Uses `spec-reviewing` skill (not `spec-writing`)

### Step 7: Write TODO.yaml

Write the final `TODO.yaml` at project root with spec-writing tasks ordered by priority, followed by the modernizing task, then the reviewing task.

## Inputs

Required:
- `ANALYSIS.md` — analysis report at project root (produced by `spec-analyzing`)
- `.research/*.md` — research reports referenced in ANALYSIS.md

Optional:
- Existing `TODO.yaml` to replace (default: create new)
- Language preference (default: infer from ANALYSIS.md)

## Outputs

- `TODO.yaml` — task list at project root (spec-writing tasks + final reviewing task)

## Remember

- This skill reads `ANALYSIS.md` from root directory — not SPECS or .plans/PLAN.md
- Each spec task uses `spec-writing` skill — no separate Plan+Execute pairs
- **Second-to-last task always uses `modernizing` skill** — language-agnostic spec modernization
- **Final task always uses `spec-reviewing` skill** — this is not optional
- Tasks are ordered by priority from ANALYSIS.md, modernizing task second-to-last, reviewing task always last
- **Task has only `title` and `completed` — NEVER add `description`**
- Title must be self-contained: include topic, target file, source, and key aspects
- **Every task title MUST include both `spec:` and `research:` relative paths**
- If a task relates to multiple research reports, list all of them in the `research:` field
- Use plain paths (e.g., `SPECS/10-foo.md`) — no markdown link syntax
- Use the same language as ANALYSIS.md (Chinese or English)
- Research reports in `.research/` are inputs, SPECS/*.md are outputs
- The modernizing task's `spec:` field lists all specs from writing tasks; its `research:` field lists context docs
- The reviewing task's `spec:` field lists all specs from writing tasks; its `research:` field lists context docs (requirements, decisions)

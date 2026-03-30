---
description: Read root ANALYSIS.md and generate TODO.yaml task list for writing specification documents. Use after spec-analyzing has produced ANALYSIS.md.
name: spec-tasking
---


# Spec Tasking

**Two-phase workflow:** First audits `ANALYSIS.md` accuracy by checking existing specs and research, then generates `TODO.yaml` with spec-writing tasks.

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
## Priority 1: Token System (SPECS/02)
Status: Missing
Source: .research/R01-token-system.md
```

**After audit (if spec exists):**
```markdown
## Priority 1: Token System (SPECS/02)
Status: ✅ Exists
Source: .research/R01-token-system.md
Audit Notes: Found SPECS/02-token-system.md (complete). No task needed.
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

## Pipeline Position

```
spec-analyzing ──▶ Phase 1: Audit ──▶ Phase 2: Generate
     │                  │                   │
     ▼                  ▼                   ▼
 ANALYSIS.md    Updated ANALYSIS.md    TODO.yaml
(may have errors) (verified accurate) (only missing specs)
```

This skill is the second step in the spec writing workflow — it audits analysis accuracy, then converts verified gaps into actionable tasks.

## Task Structure

Each specification identified in ANALYSIS.md produces exactly **one task** that uses `spec-writing` skill.

### Spec Writing Task

Writes a specification document based on research reports.

**Template:**
```
Write {spec topic} specification and use spec-writing skill to generate SPECS/{spec-number}-{spec-slug}.md
```

**Rules:**
- `{spec topic}` = human-readable topic name (e.g., "token system", "component system", "configuration management")
- `{spec-number}` = spec numbering from ANALYSIS.md (e.g., 02, 30, 50)
- `{spec-slug}` = kebab-case of the spec topic (e.g., token-system, component-system)
- Keep title to **one line** — be concise
- Use the language specified in ANALYSIS.md (Chinese or English)
- Add `description` field with: "Based on .research/{research-file}.md, define {key-aspects}"

**Example:**
```yaml
- title: "Write token system specification and use spec-writing skill to generate SPECS/02-token-system.md"
  description: "Based on .research/R01-token-system.md, define token types, naming conventions, transformation rules"
  completed: false
```

## No Planning or Review Checkpoints

Unlike code implementation workflows, spec writing tasks are relatively independent and don't require:
- Separate Plan+Execute pairs (spec-writing skill handles planning internally)
- Quality review checkpoints (specs are reviewed during human approval of ANALYSIS.md)

Spec writing tasks are independent documents that don't build on each other's code, so refactor reviews and cohesion checks are unnecessary.

Each task directly produces a specification document.

## TODO.yaml Format

```yaml
tasks:
  # ═══════════════════════════════════════════════════════════════
  # Priority 1: Token System (SPECS/02)
  # Source: .research/R01-token-system.md
  # ═══════════════════════════════════════════════════════════════

  - title: "Write token system specification and use spec-writing skill to generate SPECS/02-token-system.md"
    description: "Based on .research/R01-token-system.md, define token types, naming conventions, transformation rules"
    completed: false

  # ═══════════════════════════════════════════════════════════════
  # Priority 2: Component System (SPECS/30)
  # Source: .research/R02-component-system.md
  # ═══════════════════════════════════════════════════════════════

  - title: "Write component system specification and use spec-writing skill to generate SPECS/30-component-system.md"
    description: "Based on .research/R02-component-system.md, define component interfaces, lifecycle, composition patterns"
    completed: false

  # ═══════════════════════════════════════════════════════════════
  # Priority 3: Configuration (SPECS/50)
  # Source: .research/R03-configuration-management.md
  # ═══════════════════════════════════════════════════════════════

  - title: "Write configuration management specification and use spec-writing skill to generate SPECS/50-configuration-management.md"
    description: "Based on .research/R03-configuration-management.md, define configuration loading, validation, override rules"
    completed: false
```

**YAML rules:**
- Three fields per task: `title`, `description` (recommended), `completed` (always `false`)
- No blank lines between a title and its fields
- Blank line between tasks for readability
- Use YAML block comments (`#`) with `═` for visual separation between priority groups
- Comment headers include priority level and source research file
- Tasks ordered by priority from ANALYSIS.md

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

**Example ANALYSIS.md structure:**
```markdown
# Spec Analysis

## Scope
- 基于 R01 编写 token 系统规范
- 基于 R02 编写组件系统规范
- 基于 R03 编写配置管理规范

## Priority
1. Token 系统规范（SPECS/02）
2. 组件系统规范（SPECS/30）
3. 配置管理规范（SPECS/50）

## Structure
- SPECS/02-token-system.md
- SPECS/30-component-system.md
- SPECS/50-configuration-management.md
```

### Step 2: Determine Task Order

Order tasks by priority from audited ANALYSIS.md:
1. High priority specs first
2. Medium priority specs second
3. Low priority specs last

Within each priority level, maintain the order specified in ANALYSIS.md.

### Step 3: Generate Tasks

For each missing/incomplete spec in audited ANALYSIS.md, generate exactly one task using the template:
```
Write {spec topic} specification and use spec-writing skill to generate SPECS/{spec-number}-{spec-slug}.md
```

With description field:
```
Based on .research/{research-file}.md, define {key-aspects}
```

Extract:
- Research file from the Structure section
- Spec number and slug from the Structure section
- Key aspects from the Scope section

### Step 4: Write TODO.yaml

Write the final `TODO.yaml` at project root with tasks ordered by priority.

## Inputs

Required:
- `ANALYSIS.md` — analysis report at project root (produced by `spec-analyzing`)
- `.research/*.md` — research reports referenced in ANALYSIS.md

Optional:
- Existing `TODO.yaml` to replace (default: create new)
- Language preference (default: infer from ANALYSIS.md)

## Outputs

- `TODO.yaml` — task list at project root

## Remember

- This skill reads `ANALYSIS.md` from root directory — not SPECS or .plans/PLAN.md
- Each task directly uses `spec-writing` skill — no separate Plan+Execute pairs
- Tasks are ordered by priority from ANALYSIS.md
- No quality review checkpoints — specs are independent documents
- Task titles should be concise and include key aspects to define
- Use the same language as ANALYSIS.md (Chinese or English)
- Research reports in `.research/` are inputs, SPECS/*.md are outputs

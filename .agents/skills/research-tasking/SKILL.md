---
description: Converts a research plan (ANALYSIS.md) into a structured TODO.yaml task list for research phase execution. Use when ANALYSIS.md exists and research tasks need to be generated.
name: research-tasking
---


# Research Tasking

Reads `ANALYSIS.md` (produced by `research-analyzing`) and generates a structured `TODO.yaml` with research writing tasks organized by module priority.

**Used in:** Research phase (workflow stage 1)

**Announce at start:** "I'm using the research-tasking skill."

**Prerequisite:** `ANALYSIS.md` must exist at project root. If not, run `research-analyzing` first.

## Pipeline Position

```
Research Phase Workflow:

research-analyzing ──▶ research-tasking ──▶ Ralph (research-writing)
     │                       │                      │
     ▼                       ▼                      ▼
 ANALYSIS.md             TODO.yaml            .research/*.md
(root directory)      (root directory)     (research reports)
```

## Task Structure

Every research module in ANALYSIS.md produces exactly **one task** — a research writing task.

### Research Writing Task

Analyzes reference projects and generates a structured research report.

**Template:**
```
Research {module name} from {reference-paths} and use research-writing skill to write {output-path}
```

**Rules:**
- `{module name}` = human-readable module name from ANALYSIS.md (e.g., "token type system", "authentication patterns")
- `{reference-paths}` = comma-separated list of `.references/` paths to analyze
- `{output-path}` = target report path (e.g., `.research/R01-token-type-system.md`)
- Keep title to **one line** — be concise but self-contained (include focus areas in title)
- Always specify which `.references/` projects to analyze
- Output goes to `.research/` directory with R{XX} prefix
- **No `description` field** — put all info in the title

## TODO.yaml Format

```yaml
tasks:
  # ═══════════════════════════════════════════════════════════════
  # Module R{XX} — {Module Name}
  # ═══════════════════════════════════════════════════════════════

  - title: "Research {module name} from {reference-paths} and use research-writing skill to write {output-path}. Focus on: {research goals}"
    completed: false

  # ... more research tasks ...
```

**YAML rules:**
- **Two fields only per task: `title` and `completed`** — never add `description`
- Title must be self-contained: include module name, reference paths, output path, and focus areas
- Use YAML block comments (`#`) for module separators with `═` box lines
- No blank lines between a task's fields
- Blank line between tasks for readability
- Tasks ordered by module priority from ANALYSIS.md

## Workflow

### Step 1: Read ANALYSIS.md

Read `ANALYSIS.md` from the project root and extract:
- Research scope and goals
- Module decomposition (R01, R02, R03, etc.)
- For each module:
  - Module name and ID
  - Research goals
  - Search keywords
  - Target languages
  - Reference projects (if already filled)

### Step 2: Determine Module Order

Follow the module numbering from ANALYSIS.md. Typical ordering:
1. Core concepts and foundations (R01-R05)
2. Domain-specific features (R06-R10)
3. Integration and ecosystem (R11-R15)
4. Advanced topics (R16+)

If ANALYSIS.md has a Priority section, use that ordering instead.

### Step 3: Generate Research Tasks

For each module in ANALYSIS.md, generate one research writing task:

1. Extract module ID (e.g., R01)
2. Extract module name (e.g., "Token Type System")
3. Extract reference projects from "Reference Projects" table
4. Extract research goals (include in title)
5. Generate output path: `.research/{ID}-{slug}.md`
6. Create task using template (title + completed only)

### Step 4: Write TODO.yaml

Write the final `TODO.yaml` at project root with:
- Module separator comments
- Research writing tasks in module order
- Focus areas included in task titles

### Step 5: Announce Completion

Inform user:
```
TODO.yaml has been generated at project root with N research tasks.

Next steps:
1. Review TODO.yaml (optional)
2. Execute tasks using Ralphy loop or manually
3. Each task will use research-writing skill to analyze references and generate reports

Ready to proceed with research execution?
```

## Inputs

Required:
- `ANALYSIS.md` — the research analysis report at project root (produced by `research-analyzing`)

Optional:
- Existing `TODO.yaml` to update (append new tasks)
- Custom module ordering (default: follow ANALYSIS.md numbering)

## Remember

- This skill is the **second step** after `research-analyzing` — it does not analyze scope or determine modules
- Task titles must include reference paths so executors know which projects to analyze
- Research goals from ANALYSIS.md go into the task `title` (no separate description field)
- Output paths use R{XX} prefix matching ANALYSIS.md module IDs
- `.research/` directory contains the generated research reports
- Tasks are executed by `research-writing` skill, which analyzes references and produces structured reports
- All tasks are simple (1 module → 1 task), no Plan+Execute pairs needed for research
- Language requirement from ANALYSIS.md applies to report writing, not task generation

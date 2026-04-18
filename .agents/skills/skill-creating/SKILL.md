---
description: Use when creating new skills, editing existing skills, or verifying skills work before deployment. Covers TDD-based skill authoring, frontmatter conventions, CSO (Claude Search Optimization), testing methodology, and skill structure patterns.
name: skill-creating
---


# Skill Creating

## Purpose

Write skills that help future agents make better decisions, not just follow longer instructions.

A good skill is easy to discover, clear about when it applies, explicit about its boundaries, concise in its main file, and testable before deployment.

**Personal skills live in agent-specific directories (`~/.claude/skills` for Claude Code, `~/.agents/skills/` for Codex).**

**REQUIRED BACKGROUND:** You MUST understand `superpowers:test-driven-development` before using this skill. That skill defines RED-GREEN-REFACTOR. This skill applies the same discipline to skill authoring.

## Core Principles

### 1. Intent Before Procedure
State what the skill is trying to optimize for before listing steps. A skill should teach better judgment, not just a ritual.

### 2. Constraint-First Authoring
Define scope, non-goals, and failure modes early. If a behavior matters, encode it as an explicit rule or boundary.

### 3. Concise Main File
Keep `SKILL.md` lean. Put heavy reference material, large examples, and reusable tooling in supporting files.

### 4. Deterministic Outputs
A good skill makes success legible. State required sections, expected outputs, and how the result is validated.

### 5. Reusable Over Narrative
Write reusable guidance, patterns, and decision rules. Do not write a story about how one past task was solved.

### 6. Test What Matters
Do not trust a skill because it sounds good. Validate that an agent can discover it, apply it, and avoid the failures it is meant to prevent.

### 7. Match Instruction Strength to Task Fragility
Use hard constraints for fragile or safety-critical work. Use lighter guidance where context and judgment should stay flexible.

## When to Create a Skill

**Create a skill when:**
- the technique or judgment was not obvious
- you expect to reuse it across tasks or projects
- the pattern is broader than one repository's local convention
- future agents would benefit from having the guidance discoverable

**Do not create a skill for:**
- one-off fixes or historical writeups
- project-specific conventions that belong in `CLAUDE.md`
- information that is already well covered by an existing canonical skill or reference
- mechanical constraints that should be enforced by code, validation, or automation instead of prose

## Hard Boundaries

A skill you write **MUST**:
- define when it should be used
- define when it should not be used
- describe the core decision rule, pattern, or workflow it teaches
- include enough searchable terms for discovery
- keep the main file focused and move heavy detail out of `SKILL.md`
- make verification possible before deployment

A skill you write **MUST NOT**:
- be a narrative case study
- duplicate large chunks of canonical guidance from supporting references
- explain basic concepts the model already knows unless the explanation changes behavior
- add sections that do not improve discovery, decision-making, or execution quality
- force rigid process where contextual judgment is the real requirement

## Authoring Workflow

Use this loop:

1. **Define the job**
   - What problem does the skill solve?
   - What should change in agent behavior after reading it?
   - What is explicitly out of scope?

2. **Define triggers and boundaries**
   - Write the discovery conditions first.
   - Capture positive triggers, negative triggers, and key searchable phrases.

3. **Draft the minimum useful skill**
   - Write a lean `SKILL.md` with the required sections.
   - Prefer principles, rules, and compact examples over long explanation.

4. **Extract heavy content**
   - Move large references to `references/`.
   - Move deterministic tooling to `scripts/`.
   - Move reusable templates and schemas to `assets/`.

5. **Validate and tighten**
   - Test discoverability.
   - Test application.
   - Tighten weak wording, loopholes, and vague boundaries.

## What a Good Skill Contains

A strong `SKILL.md` usually has these parts:

1. **Overview / Purpose**
   - What the skill helps with
   - The core idea in 1-2 short paragraphs

2. **When to Use**
   - concrete triggers, symptoms, or contexts

3. **When Not to Use**
   - explicit non-goals
   - adjacent cases that should use another tool, skill, or document

4. **Core Principles or Decision Rules**
   - the judgment model the agent should apply

5. **Workflow / Implementation Guidance**
   - short, high-signal steps
   - conditional references to deeper docs when needed

6. **Verification / Quality Checks**
   - how to know the skill is correct, complete, and ready

7. **Supporting Files**
   - only when justified by size, reuse, or determinism

Use `assets/skill-template.md` when you need a starting point, but tailor the structure to the actual skill type.

## Frontmatter and Discovery

Future Claude finds skills through metadata first. Bad frontmatter makes good skills invisible.

### Name
- lowercase letters, numbers, and hyphens only
- maximum 64 characters
- use clear, searchable naming
- prefer gerund or action-oriented names when they fit

### Description
- write in third person
- describe when to use the skill
- include concrete triggers, symptoms, contexts, and keywords
- keep it specific enough to disambiguate from nearby skills
- do not summarize the full workflow in the description

**Good pattern:**
```yaml
name: skill-creating
description: Use when creating a new skill, rewriting an existing SKILL.md, improving skill discoverability, tightening skill boundaries, validating skill metadata, or deciding how to structure supporting references, assets, and scripts.
```

**Bad pattern:**
```yaml
description: Helps write skills by following a six-step workflow.
```

Why this is bad: it describes the skill's internals instead of the situations that should trigger it.

For deeper naming and CSO guidance, read `references/cso-guidelines.md`.

## Progressive Disclosure

Keep `SKILL.md` under 500 lines. Shorter is better if clarity is preserved.

**Keep inline:**
- principles
- concise decision rules
- short examples
- compact workflows

**Move to `references/`:**
- long documentation
- testing methodology
- troubleshooting guides
- detailed catalogs and tables

**Move to `scripts/`:**
- validation commands
- parsing/transformation logic
- repetitive deterministic operations

**Move to `assets/`:**
- templates
- schemas
- reusable examples or skeletons

Prefer just-in-time references:

```markdown
Read `references/testing-skill-types.md` when choosing a testing strategy.
If validation fails, read `references/error-handling.md`.
```

Avoid force-loading large files unless they are immediately needed.

## Structure Supporting Files Deliberately

Use this directory shape only as needed:

```text
skill-name/
├── SKILL.md
├── scripts/
├── references/
├── assets/
├── agents/
└── examples/
```

### Use each directory for a distinct job
- `scripts/` — executable helpers and validation logic
- `references/` — heavy docs and deep guidance
- `assets/` — templates, schemas, reusable scaffolds
- `agents/` — testing prompts or specialized evaluation prompts
- `examples/` — complete worked examples when they truly teach something the main file cannot

Do not create empty structure just because the folders are available.

## Examples and Flowcharts

### Examples
Use one excellent example instead of many weak ones.

A good example is:
- directly relevant
- complete enough to adapt
- small enough to scan quickly
- chosen because it resolves ambiguity, not because examples feel nice to include

Do not dilute the skill with many parallel examples in different languages unless the language difference is essential.

### Flowcharts
Use flowcharts only when a decision is genuinely non-obvious.

Good use cases:
- branching choice between adjacent approaches
- stop/continue loops
- non-trivial decision points

Bad use cases:
- linear procedures
- reference material
- code examples

For graph conventions, see `graphviz-conventions.dot`.

## Testing Skills Before Trusting Them

Writing skills is TDD for process documentation.

### The rule
Do not treat a skill as done until you have evidence that it changes behavior in the intended way.

### RED
Run a baseline scenario without the skill. Observe what the agent does naturally.

### GREEN
Write the minimum skill that addresses the actual failure, ambiguity, or missing guidance.

### REFACTOR
Tighten wording, close loopholes, and re-test until the skill is reliable.

### What to test
At minimum, verify:
- **discovery** — would the metadata cause the right agent to load the skill?
- **application** — does the skill help the agent act better, not just describe itself?
- **boundaries** — does the skill prevent misuse or overreach?

For detailed testing methods, read:
- `testing-skills-with-subagents.md`
- `references/testing-skill-types.md`
- `references/bulletproofing.md`

## Quality Bar

Before considering a skill complete, confirm:
- the metadata is valid
- the description is discoverable and specific
- the main file is concise
- the skill has explicit scope and non-scope
- the guidance teaches judgment, not just ceremony
- supporting files exist only where they earn their keep
- testing matches the skill type
- the skill is better because of what it omits, not just what it includes

Use `references/skill-creation-checklist.md` as the final pass.

## Validation

```bash
python3 scripts/validate-metadata.py --file SKILL.md
wc -l SKILL.md
```

Target: valid metadata and a lean main file under 500 lines.

If you hit issues with metadata, testing, or file organization, read `references/error-handling.md`.

## Anti-Patterns

Avoid these:
- **Narrative skills** — historical story instead of reusable guidance
- **Verbose theory dumps** — long explanation that does not change behavior
- **Vague descriptions** — metadata that cannot trigger reliably
- **Workflow-only skills** — steps with no underlying decision rule
- **Duplicate policy text** — copying large reference content into `SKILL.md`
- **Cargo-cult structure** — adding sections, folders, or diagrams with no clear payoff
- **Example sprawl** — many weak examples instead of one decisive one

## Reference Map

Read supporting files only when needed:
- `assets/skill-template.md` — starting template for new skills
- `references/skill-creation-checklist.md` — final validation checklist
- `references/testing-skill-types.md` — choose the right test strategy for the skill type
- `testing-skills-with-subagents.md` — baseline, pressure, and loophole-focused testing
- `references/cso-guidelines.md` — naming and discoverability guidance
- `references/bulletproofing.md` — closing rationalization loopholes in discipline-enforcing skills
- `references/error-handling.md` — troubleshooting metadata, context bloat, and testing issues
- `anthropic-best-practices.md` — official authoring guidance on concision and instruction shape
- `persuasion-principles.md` — pressure/compliance background for discipline-oriented skills

## Bottom Line

A strong skill is not the longest explanation. It is the smallest reliable instruction set that helps a future agent discover the right guidance, apply it correctly, stay inside the boundary, and verify the result.
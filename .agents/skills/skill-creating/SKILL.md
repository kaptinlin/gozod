---
description: Use when creating new skills, editing existing skills, or verifying skills work before deployment. Covers TDD-based skill authoring, frontmatter conventions, CSO (Claude Search Optimization), testing methodology, and skill structure patterns.
name: skill-creating
---



# Writing Skills

## Overview

**Writing skills IS Test-Driven Development applied to process documentation.**

**Personal skills live in agent-specific directories (`~/.claude/skills` for Claude Code, `~/.agents/skills/` for Codex)**

You write test cases (pressure scenarios with subagents), watch them fail (baseline behavior), write the skill (documentation), watch tests pass (agents comply), and refactor (close loopholes).

**Core principle:** If you didn't watch an agent fail without the skill, you don't know if the skill teaches the right thing.

**REQUIRED BACKGROUND:** You MUST understand superpowers:test-driven-development before using this skill. That skill defines the fundamental RED-GREEN-REFACTOR cycle. This skill adapts TDD to documentation.

**Official guidance:** For Anthropic's official skill authoring best practices, see anthropic-best-practices.md. This document provides additional patterns and guidelines that complement the TDD-focused approach in this skill.

## Supporting Documentation

This skill has several supporting documents for deep dives into specific topics. Read them progressively as needed:

**Core References:**
- **`references/skill-creation-checklist.md`** - Complete TDD-adapted checklist for deployment. Read when ready to validate your skill before deployment.
- **`assets/skill-template.md`** - Ready-to-use template for new skills. Copy this when starting a new skill.

**Deep Dives:**
- **`testing-skills-with-subagents.md`** - Complete testing methodology including pressure scenarios, baseline testing, and plugging holes. Read when you need to test a discipline-enforcing skill.
- **`anthropic-best-practices.md`** - Anthropic's official skill authoring guidelines covering conciseness, degrees of freedom, and token efficiency. Read for official best practices.
- **`persuasion-principles.md`** - Psychology-based techniques (Cialdini, 2021) for making skills resistant to rationalization. Read when creating discipline-enforcing skills.
- **`references/cso-guidelines.md`** - Detailed CSO examples, naming conventions, token efficiency, and cross-referencing patterns. Read when writing descriptions or naming skills.
- **`references/bulletproofing.md`** - Techniques for closing loopholes and building rationalization tables. Read when creating discipline-enforcing skills.
- **`references/error-handling.md`** - Common failure modes and recovery steps during skill creation. Read when troubleshooting metadata, context bloat, testing, or deployment issues.
- **`references/testing-skill-types.md`** - Test approaches for discipline, technique, pattern, and reference skills. Read when choosing a testing strategy.

**Tools:**
- **`scripts/validate-metadata.py`** - Automated validation for name and description. Run before deployment.
- **`scripts/init_skill.py`** - Initialize new skill directory structure. Use when starting a new skill.

**When to Read Each:**
- Starting new skill -> Use `init_skill.py` or copy `skill-template.md`
- Writing description -> Review CSO section below + `references/cso-guidelines.md`
- Testing skill -> Read `testing-skills-with-subagents.md`
- Making skill bulletproof -> Read `references/bulletproofing.md` + `persuasion-principles.md`
- Before deployment -> Check `references/skill-creation-checklist.md` + run `validate-metadata.py`
- Troubleshooting -> Read `references/error-handling.md`

## Quick Start Procedure

**New to skill creation? Follow these steps:**

1. **Validate Metadata First**
   - Define skill name (lowercase, numbers, hyphens only, 1-64 chars)
   - Draft description (max 1024 chars, starts with "Use when...", third person)
   - Run: `python3 scripts/validate-metadata.py --name "skill-name" --description "..."`
   - Fix any errors and re-validate

2. **Create Directory Structure**
   - Create skill directory: `mkdir skill-name/`
   - Add subdirectories: `scripts/`, `references/`, `assets/` (only as needed)
   - Copy template: `cp assets/skill-template.md skill-name/SKILL.md`

3. **Run Baseline Tests (RED Phase)**
   - Create pressure scenarios WITHOUT the skill
   - Document exact agent behavior and rationalizations
   - Identify patterns in failures

4. **Write Minimal Skill (GREEN Phase)**
   - Fill in template sections addressing baseline failures
   - Keep SKILL.md under 500 lines
   - Extract heavy content to `references/` if needed
   - Add scripts to `scripts/` for fragile/deterministic tasks

5. **Test and Refactor**
   - Run scenarios WITH skill - verify compliance
   - Identify new rationalizations -> add counters
   - Re-test until bulletproof

6. **Validate and Deploy**
   - Run: `python3 scripts/validate-metadata.py --file SKILL.md`
   - Check: `wc -l SKILL.md` (should be <500)
   - Review: `references/skill-creation-checklist.md`
   - Commit and push

**Detailed guidance for each step is in the sections below.**

## What is a Skill?

A **skill** is a reference guide for proven techniques, patterns, or tools. Skills help future Claude instances find and apply effective approaches.

**Skills are:** Reusable techniques, patterns, tools, reference guides

**Skills are NOT:** Narratives about how you solved a problem once

## TDD Mapping for Skills

| TDD Concept | Skill Creation |
|-------------|----------------|
| **Test case** | Pressure scenario with subagent |
| **Production code** | Skill document (SKILL.md) |
| **Test fails (RED)** | Agent violates rule without skill (baseline) |
| **Test passes (GREEN)** | Agent complies with skill present |
| **Refactor** | Close loopholes while maintaining compliance |
| **Write test first** | Run baseline scenario BEFORE writing skill |
| **Watch it fail** | Document exact rationalizations agent uses |
| **Minimal code** | Write skill addressing those specific violations |
| **Watch it pass** | Verify agent now complies |
| **Refactor cycle** | Find new rationalizations -> plug -> re-verify |

The entire skill creation process follows RED-GREEN-REFACTOR.

## When to Create a Skill

**Create when:**
- Technique wasn't intuitively obvious to you
- You'd reference this again across projects
- Pattern applies broadly (not project-specific)
- Others would benefit

**Don't create for:**
- One-off solutions
- Standard practices well-documented elsewhere
- Project-specific conventions (put in CLAUDE.md)
- Mechanical constraints (if it's enforceable with regex/validation, automate it -- save documentation for judgment calls)

## Progressive Disclosure Guidelines

**Core principle:** Keep SKILL.md lean. Extract content progressively as it grows.

### Line Count Thresholds

```
SKILL.md size decision tree:
|- Under 300 lines -> Keep everything inline
|- 300-500 lines -> Consider extracting heavy reference material
|- Over 500 lines -> MUST extract content (see extraction criteria below)
```

### When to Extract Content

**Extract to `references/`:** API docs (>100 lines), syntax guides, schema definitions, troubleshooting guides, testing methodology

**Extract to `scripts/`:** Validation logic, transformation/parsing code, deterministic operations, CLI tools for repetitive tasks

**Extract to `assets/`:** Code templates, JSON schemas, configuration examples, output format templates

### How to Reference Extracted Content

Use just-in-time loading instructions:

```markdown
# GOOD: Load only when needed
Read `references/api-spec.md` to identify the correct endpoint.

# GOOD: Conditional loading
If validation fails, read `references/troubleshooting.md` for recovery steps.

# BAD: Force-loading with @
See @references/api-spec.md for details.
```

**Why avoid @:** The `@` syntax force-loads files immediately, consuming context before you need them.

### Verification

```bash
# Check line count
wc -l SKILL.md
# Target: <500 lines for main skill
```

## Skill Types

### Technique
Concrete method with steps to follow (condition-based-waiting, root-cause-tracing)

### Pattern
Way of thinking about problems (flatten-with-flags, test-invariants)

### Reference
API docs, syntax guides, tool documentation (office docs)

## Directory Structure

```
skills/
  skill-name/
    SKILL.md              # Main reference (required, <500 lines)
    scripts/              # Validation, transformation, CLI tools
    references/           # Heavy docs, API specs, troubleshooting
    assets/               # Templates, schemas, examples
    agents/               # Testing subagent prompts (optional)
    examples/             # Real-world skill examples (optional)
```

**Flat namespace** - all skills in one searchable namespace

### Directory Purposes

**`scripts/`** - Tiny CLI tools, validation scripts, transformation/parsing logic, repetitive operations. Example: `validate-metadata.py`, `init-skill.py`

**`references/`** - API documentation (>100 lines), syntax guides, schema definitions, troubleshooting guides. Keep files flat (one level deep). Example: `skill-creation-checklist.md`

**`assets/`** - Code templates, JSON schemas, configuration examples, output format templates. Example: `skill-template.md`

**`agents/`** - Subagent prompts for testing skills (grader, comparator, analyzer agents). Optional.

**`examples/`** - Complete examples of skills in action, real scenarios, testing examples. Optional.

### File Organization Principles

**Separate files for:** Heavy reference (100+ lines), reusable tools/scripts/templates

**Keep inline:** Principles, concepts, code patterns (<50 lines), everything else

## SKILL.md Structure

**Frontmatter (YAML):**
- Only two fields supported: `name` and `description`
- Max 1024 characters total
- `name`: Use letters, numbers, and hyphens only (no parentheses, special chars)
- `description`: Third-person, describes ONLY when to use (NOT what it does)
  - Start with "Use when..." to focus on triggering conditions
  - Include specific symptoms, situations, and contexts
  - **NEVER summarize the skill's process or workflow** (see CSO section for why)
  - Keep under 500 characters if possible

```markdown
---
name: Skill-Name-With-Hyphens
description: Use when [specific triggering conditions and symptoms]
---

# Skill Name

## Overview
What is this? Core principle in 1-2 sentences.

## When to Use
[Small inline flowchart IF decision non-obvious]

Bullet list with SYMPTOMS and use cases
When NOT to use

## Core Pattern (for techniques/patterns)
Before/after code comparison

## Quick Reference
Table or bullets for scanning common operations

## Implementation
Inline code for simple patterns
Link to file for heavy reference or reusable tools

## Common Mistakes
What goes wrong + fixes

## Real-World Impact (optional)
Concrete results
```

## Claude Search Optimization (CSO)

**Critical for discovery:** Future Claude needs to FIND your skill.

**Key rules:**
- Description = WHEN to use, NOT what the skill does. Start with "Use when..."
- Include positive triggers (conditions/symptoms) AND negative triggers (when NOT to use)
- Write in third person -- never use "I", "you", "we"
- **NEVER summarize the skill's workflow in the description** -- this causes Claude to follow the description shortcut instead of reading the full skill
- Use searchable keywords (error messages, symptoms, synonyms, tool names)
- Name skills with `{domain}-{action}` pattern using gerund form (e.g., `tdd-planning`, `spec-writing`)

Read `references/cso-guidelines.md` for detailed examples, the unified naming system, token efficiency techniques, and cross-referencing patterns.

## Flowchart Usage

**Use flowcharts ONLY for:** Non-obvious decision points, process loops where you might stop too early, "When to use A vs B" decisions

**Never use flowcharts for:** Reference material (use tables/lists), code examples (use markdown blocks), linear instructions (use numbered lists)

See @graphviz-conventions.dot for graphviz style rules.

**Visualizing for your human partner:** Use `render-graphs.js` to render a skill's flowcharts to SVG:
```bash
./render-graphs.js ../some-skill           # Each diagram separately
./render-graphs.js ../some-skill --combine # All diagrams in one SVG
```

## Code Examples

**One excellent example beats many mediocre ones**

Choose most relevant language (testing -> TypeScript/JavaScript, debugging -> Shell/Python, data -> Python).

**Good example:** Complete and runnable, well-commented explaining WHY, from real scenario, shows pattern clearly, ready to adapt

**Don't:** Implement in 5+ languages, create fill-in-the-blank templates, write contrived examples. You're good at porting -- one great example is enough.

## File Organization

### Self-Contained Skill
```
defense-in-depth/
  SKILL.md    # Everything inline
```
When: All content fits, no heavy reference needed

### Skill with Reusable Tool
```
condition-based-waiting/
  SKILL.md    # Overview + patterns
  example.ts  # Working helpers to adapt
```
When: Tool is reusable code, not just narrative

### Skill with Heavy Reference
```
pptx/
  SKILL.md       # Overview + workflows
  pptxgenjs.md   # 600 lines API reference
  ooxml.md       # 500 lines XML structure
  scripts/       # Executable tools
```
When: Reference material too large for inline

## The Iron Law (Same as TDD)

```
NO SKILL WITHOUT A FAILING TEST FIRST
```

This applies to NEW skills AND EDITS to existing skills.

Write skill before testing? Delete it. Start over.
Edit skill without testing? Same violation.

**No exceptions:**
- Not for "simple additions"
- Not for "just adding a section"
- Not for "documentation updates"
- Don't keep untested changes as "reference"
- Don't "adapt" while running tests
- Delete means delete

**REQUIRED BACKGROUND:** The superpowers:test-driven-development skill explains why this matters. Same principles apply to documentation.

## Testing All Skill Types

Different skill types need different test approaches. Read `references/testing-skill-types.md` for detailed test strategies covering discipline-enforcing, technique, pattern, and reference skills.

**Quick summary:** Discipline skills need pressure scenarios. Technique skills need application scenarios. Pattern skills need recognition and counter-example scenarios. Reference skills need retrieval and gap testing.

## Bulletproofing Skills Against Rationalization

Skills that enforce discipline need to resist rationalization. Read `references/bulletproofing.md` for the complete methodology including loophole closing, rationalization tables, red flags lists, and common testing excuses.

**Key techniques (brief):**
- Close every loophole explicitly -- forbid specific workarounds, not just the violation
- Address "spirit vs letter" arguments with: "Violating the letter of the rules IS violating the spirit"
- Build rationalization tables from baseline testing -- every agent excuse gets a counter
- Create red flags lists for agent self-checking
- See also `persuasion-principles.md` for the psychology foundation (Cialdini, 2021)

## Error Handling

Read `references/error-handling.md` for detailed failure modes and recovery steps covering metadata validation, context bloat, testing failures, and deployment issues.

**Quick validation:**
```bash
python3 scripts/validate-metadata.py --file SKILL.md
wc -l SKILL.md  # Should be <500
```

## RED-GREEN-REFACTOR for Skills

Follow the TDD cycle:

### RED: Write Failing Test (Baseline)

Run pressure scenario with subagent WITHOUT the skill. Document exact behavior:
- What choices did they make?
- What rationalizations did they use (verbatim)?
- Which pressures triggered violations?

This is "watch the test fail" - you must see what agents naturally do before writing the skill.

### GREEN: Write Minimal Skill

Write skill that addresses those specific rationalizations. Don't add extra content for hypothetical cases.

Run same scenarios WITH skill. Agent should now comply.

### REFACTOR: Close Loopholes

Agent found new rationalization? Add explicit counter. Re-test until bulletproof.

**Testing methodology:** See `testing-skills-with-subagents.md` for the complete testing methodology (pressure scenarios, pressure types, plugging holes, meta-testing).

## Anti-Patterns

**Narrative Example:** "In session 2025-10-03, we found empty projectDir caused..." -- Too specific, not reusable

**Multi-Language Dilution:** example-js.js, example-py.py, example-go.go -- Mediocre quality, maintenance burden

**Code in Flowcharts:** `step1 [label="import fs"]` -- Can't copy-paste, hard to read

**Generic Labels:** helper1, helper2, step3, pattern4 -- Labels should have semantic meaning

## STOP: Before Moving to Next Skill

**After writing ANY skill, you MUST STOP and complete the deployment process.**

**Do NOT:** Create multiple skills in batch without testing each, move to next skill before current one is verified, skip testing because "batching is more efficient"

**The deployment checklist is MANDATORY for EACH skill.**

## Skill Creation Checklist

See `references/skill-creation-checklist.md` for the complete TDD-adapted checklist covering RED, GREEN, REFACTOR phases, quality checks, and deployment steps.

**Quick validation:**
```bash
python3 scripts/validate-metadata.py --file SKILL.md
```

## Discovery Workflow

How future Claude finds your skill:

1. **Encounters problem** ("tests are flaky")
3. **Finds SKILL** (description matches)
4. **Scans overview** (is this relevant?)
5. **Reads patterns** (quick reference table)
6. **Loads example** (only when implementing)

**Optimize for this flow** - put searchable terms early and often.

## The Bottom Line

**Creating skills IS TDD for process documentation.**

Same Iron Law: No skill without failing test first.
Same cycle: RED (baseline) -> GREEN (write skill) -> REFACTOR (close loopholes).
Same benefits: Better quality, fewer surprises, bulletproof results.

If you follow TDD for code, follow it for skills. It's the same discipline applied to documentation.

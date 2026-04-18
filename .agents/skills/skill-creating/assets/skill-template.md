---
name: [skill-name]
description: Use when [specific triggers, symptoms, situations, and keywords that should cause this skill to be selected].
---

# [Skill Title]

## Purpose

[What this skill helps an agent do. State the core idea in 1-2 short paragraphs. Focus on the decision quality or behavior this skill should improve.]

[If this skill requires another skill first, state it here.]
**REQUIRED BACKGROUND:** You MUST understand `[other-skill-name]` before using this skill.

## When to Use

Use this skill when:
- [specific trigger or symptom]
- [specific task context]
- [specific failure mode or ambiguity this skill resolves]

## When Not to Use

Do not use this skill for:
- [adjacent case that needs a different skill, tool, or document]
- [out-of-scope task]
- [case where automation/validation is better than prose guidance]

## Core Principles

### 1. [Principle Name]
[Short rule explaining what matters and why.]

### 2. [Principle Name]
[Short rule explaining what matters and why.]

### 3. [Principle Name]
[Short rule explaining what matters and why.]

[Add more principles only when they materially improve decisions.]

## Hard Boundaries

This skill **MUST**:
- [required behavior or output]
- [required scope or constraint]
- [required verification or quality bar]

This skill **MUST NOT**:
- [forbidden behavior]
- [common overreach or anti-pattern]
- [unnecessary verbosity or duplication]

## Quick Reference

[Optional. Keep only if it improves scanning speed. Use a short table or bullets.]

| Situation | What to do |
|-----------|------------|
| [Common case 1] | [Action or decision rule] |
| [Common case 2] | [Action or decision rule] |

## Workflow

### 1. [Phase Name]
- [short imperative step]
- [short imperative step]
- [reference deeper docs only if needed]

### 2. [Phase Name]
- [short imperative step]
- [conditional instruction: If X, do Y. Otherwise, do Z.]

### 3. [Phase Name]
- [short imperative step]
- [validation or handoff step]

## Examples

[Optional. Include only when an example resolves ambiguity better than prose. Prefer one decisive example over many weak ones.]

**Before (problematic):**
```[language]
// Example of the mistake or weaker approach
```

**After (preferred):**
```[language]
// Example showing the intended pattern
```

## Verification

Before considering the work complete, confirm:
- [metadata or input quality check]
- [boundary or scope check]
- [output quality check]
- [test or validation step]

## Common Mistakes

**Mistake: [Description]**
- **Problem:** [Why it fails]
- **Fix:** [How to correct it]

**Mistake: [Description]**
- **Problem:** [Why it fails]
- **Fix:** [How to correct it]

## Supporting Files

[Only include this section if supporting files are justified by size, reuse, or determinism.]

- `references/[doc-name].md`: [Heavy reference or deep guidance]
- `scripts/[script-name].[ext]`: [Deterministic helper or validator]
- `assets/[template-name].[ext]`: [Reusable template, schema, or scaffold]

## Anti-Patterns

Avoid:
- [narrative instead of reusable guidance]
- [vague metadata or vague boundaries]
- [duplicating large reference material in the main file]
- [adding sections that do not improve discovery, decisions, or execution]

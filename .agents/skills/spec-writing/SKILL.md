---
description: Create or update specification documents for non-code projects (research, content, product, marketing). Use when writing specs, defining requirements, creating project briefs, or consolidating existing specs.
name: spec-writing
---


# Specification Writing (General Projects)

Write specs that define project scope, requirements, and constraints for non-code projects. A spec is the source of truth — it defines what the project IS, why it's this way, and what's forbidden.

## Overview

**Specs are the source of truth. Deliverables are artifacts of the spec.**

A spec is a **unified document** that combines project requirements with decision rationale. No separate "design docs" — the why lives next to the what.

**Core principle:** Specs define rules that can be violated. If content cannot be violated by execution, it doesn't belong in a spec.

## When to Use

Use this skill when:
- Creating specification documents for research projects
- Defining content strategy and requirements
- Writing product requirement documents (PRDs)
- Planning marketing campaigns with clear constraints
- Consolidating split documentation into unified specs
- Reviewing specs for completeness and clarity

When NOT to use:
- Writing process documentation (→ team docs)
- Creating how-to guides (→ README)
- Documenting tool usage (→ tool docs)
- Writing tutorials or examples (→ guides)

## Quick Reference

| Aspect | Rule |
|--------|------|
| **Organization** | Number by domain (00, 10, 20...), gaps for future insertion |
| **Single Source** | Every concept defined in exactly ONE spec |
| **Current State** | No changelogs, no history — git tracks changes |
| **Cohesion** | Related concepts together in one spec |
| **Decisions** | Inline with rules: `> **Why**: ... > **Rejected**: ...` |
| **Forbidden** | Explicit "don't do" list with alternatives |
| **Actionable** | Team knows what to deliver and what NOT to deliver |
| **Concise** | Every sentence earns its place |

## Core Pattern

**Before: Split what/why**
```
specs/target-audience.md    — Demographics and personas
designs/audience-why.md     — rationale for targeting
```

**After: Unified spec**
```markdown
## Target Audience

Primary: Tech-savvy professionals aged 25-40 in urban areas.

- Income: $80k-150k annually
- Education: Bachelor's degree or higher
- Tech adoption: Early adopters, use 3+ SaaS tools daily
- Pain point: Overwhelmed by tool fragmentation

> **Why**: This segment has highest willingness to pay and
> matches our product's complexity level. They understand
> the value proposition without extensive education.
>
> **Rejected**: Mass market (requires different messaging),
> enterprise only (limits growth), students (low budget).
```

The requirement and its rationale are **co-located**.

## Workflow

1. **Understand requirements** — Read user request, extract output path if specified
2. **Draft spec content** — Follow structure patterns and writing guidelines
3. **Write file** — Use Write tool with user-specified path or standard path
4. **Verify** — Confirm file was created successfully

## Implementation

### Spec Structure

```markdown
# <Topic>                               ← Domain this spec covers

## Overview                             ← 2-3 sentences: scope and boundaries
## <Domain Sections>                    ← Requirements, constraints, criteria (varies by topic)

> **Why**: <rationale>                  ← Decision context inline
> **Rejected**: <alternatives>          ← What we explicitly chose NOT to do

## Forbidden                            ← What NOT to do (with alternatives)
## Acceptance Criteria                  ← How to verify completion
```

### Project Types

**Research Projects:**
```markdown
# Research: <Topic>

## Research Questions
- Primary question
- Secondary questions

## Methodology
- Approach and methods
- Data sources
- Analysis framework

> **Why**: <rationale for methodology>

## Deliverables
- Report format and structure
- Key findings format
- Timeline

## Forbidden
- ❌ Don't: Expand scope beyond primary question
  ✅ Do: Document scope creep as future research

## Acceptance Criteria
- [ ] Primary question answered with evidence
- [ ] Findings validated by 2+ sources
- [ ] Report delivered in specified format
```

**Content Projects:**
```markdown
# Content: <Topic>

## Audience
- Primary audience
- Secondary audience
- Excluded audiences

## Tone and Voice
- Writing style
- Vocabulary level
- Perspective (1st/2nd/3rd person)

> **Why**: <rationale for tone choice>

## Content Requirements
- Format (blog, video, infographic)
- Length/duration
- Key messages
- Call to action

## Forbidden
- ❌ Don't: Use jargon without explanation
  ✅ Do: Define technical terms on first use

## Acceptance Criteria
- [ ] Meets length requirements
- [ ] Tone consistent throughout
- [ ] Key messages present
- [ ] CTA clear and actionable
```

**Product Projects:**
```markdown
# Product: <Feature Name>

## Problem Statement
What user problem does this solve?

## Target Users
- User segments
- Use cases
- User journey touchpoints

## Requirements
### Must Have
- Core functionality
- Performance criteria
- Constraints

### Should Have
- Nice-to-have features
- Future considerations

> **Why**: <rationale for prioritization>
> **Rejected**: <features explicitly excluded>

## Success Metrics
- KPIs to track
- Target values
- Measurement method

## Forbidden
- ❌ Don't: Add features without user validation
  ✅ Do: Test with 5+ users before building

## Acceptance Criteria
- [ ] Core functionality works for target users
- [ ] Meets performance criteria
- [ ] Success metrics trackable
```

**Marketing Projects:**
```markdown
# Campaign: <Campaign Name>

## Objectives
- Primary goal
- Secondary goals
- Success definition

## Target Audience
- Demographics
- Psychographics
- Channels they use

## Messaging
- Core message
- Supporting messages
- Value proposition

> **Why**: <rationale for messaging>

## Channels and Tactics
- Channel mix
- Content types per channel
- Budget allocation

## Timeline
- Campaign phases
- Key milestones
- Launch date

## Forbidden
- ❌ Don't: Launch without A/B testing messaging
  ✅ Do: Test 2-3 message variants first

## Acceptance Criteria
- [ ] Reaches target audience size
- [ ] Achieves primary objective
- [ ] Stays within budget
```

## Writing Guidelines

### Be Specific
❌ "Content should be engaging"
✅ "Content should include 2-3 concrete examples and 1 actionable takeaway"

### Be Measurable
❌ "Research should be thorough"
✅ "Research should cite 10+ sources, including 3+ academic papers"

### Be Actionable
❌ "Consider user feedback"
✅ "Interview 5 users, document findings, adjust spec based on patterns"

### Include Rationale
Every major decision needs a "Why" block explaining:
- Context that led to this decision
- Alternatives considered
- Trade-offs accepted

### Define Forbidden Actions
Explicitly state what NOT to do:
- Common mistakes to avoid
- Tempting but wrong approaches
- Scope boundaries

## Spec Organization

```
.specs/
├── 00-overview.md              # Project overview and goals
├── 10-audience.md              # Target audience definition
├── 20-requirements.md          # Core requirements
├── 30-constraints.md           # Constraints and limitations
├── 40-deliverables.md          # Expected deliverables
└── 50-success-metrics.md       # How to measure success
```

Number by domain (00, 10, 20...) to allow insertion of related specs later.

## Common Patterns

### Decision Documentation
```markdown
## <Decision Point>

Decision: <what we chose>

> **Why**: <context and rationale>
> **Rejected**: <alternatives we considered>
> **Trade-offs**: <what we gave up>
```

### Constraint Documentation
```markdown
## Constraints

### Budget
- Total: $X
- Allocation: Y% content, Z% distribution

### Timeline
- Start: YYYY-MM-DD
- Milestones: [list]
- Deadline: YYYY-MM-DD (hard constraint)

### Resources
- Team: [roles]
- Tools: [required tools]
- Dependencies: [external dependencies]
```

### Acceptance Criteria
```markdown
## Acceptance Criteria

### Functional
- [ ] Deliverable meets format requirements
- [ ] Content addresses all key points
- [ ] Quality passes review checklist

### Non-Functional
- [ ] Delivered by deadline
- [ ] Within budget
- [ ] Stakeholder approval obtained
```

## Templates

Quick-start templates for spec documents, directory layouts, and decision records in [references/templates.md](references/templates.md).

## Detailed Guides

Complete reference materials in `guides/`:

**Foundation (00-series):**
- [00-principles.md](guides/00-principles.md) — Core principles: SSOT, current state, cohesion, actionable, concise
- [01-living-documentation.md](guides/01-living-documentation.md) — Specs as living documents, git as history

**Organization (10-series):**
- [10-organization.md](guides/10-organization.md) — Directory structure, domain series, numbering rules
- [11-structure.md](guides/11-structure.md) — Spec structure patterns, section organization
- [12-naming.md](guides/12-naming.md) — File naming conventions, series prefixes
- [13-structure-patterns.md](guides/13-structure-patterns.md) — Common section patterns with examples

**Content (20-series):**
- [20-content-boundaries.md](guides/20-content-boundaries.md) — What belongs in specs vs elsewhere
- [21-writing-style.md](guides/21-writing-style.md) — Writing guidelines, tone, conciseness
- [23-boundary-cases.md](guides/23-boundary-cases.md) — Edge cases for content boundaries
- [24-code-examples.md](guides/24-code-examples.md) — How to write effective examples
- [25-ai-readability.md](guides/25-ai-readability.md) — Optimizing specs for AI agent consumption

**Acceptance (30-series):**
- [31-acceptance-criteria.md](guides/31-acceptance-criteria.md) — Writing testable acceptance criteria

**Domain-Specific (40-series):**
- [40-architecture-specs.md](guides/40-architecture-specs.md) — System and project architecture specs
- [41-api-specs.md](guides/41-api-specs.md) — Interface and contract specifications
- [42-data-model-specs.md](guides/42-data-model-specs.md) — Data and information model specs
- [43-coding-standards.md](guides/43-coding-standards.md) — Standards and conventions

**Templates (50-series):**
- [50-spec-templates.md](guides/50-spec-templates.md) — Copy-paste templates for quick starts
- [51-spec-examples.md](guides/51-spec-examples.md) — Complete example specs
- [52-philosophy-template.md](guides/52-philosophy-template.md) — Template for philosophy/principles docs

## Critical Requirements

**YOU MUST WRITE THE FILE**:
- If user specifies output path (e.g., "generate SPECS/02-topic.md"), use that exact path
- If no path specified, use standard path: `.specs/{number}-{topic}.md` or `SPECS/{number}-{topic}.md`
- After drafting the spec content, use Write tool to create the file
- Do not just describe what should be written — actually write it
- Verify the file exists after writing

## Remember

- **Specs define constraints** — if it can't be violated, it's not a spec
- **Co-locate decisions** — why lives next to what
- **Be explicit about "no"** — forbidden actions prevent mistakes
- **Make it testable** — acceptance criteria must be verifiable
- **Keep it current** — update specs, don't append history
- **One source of truth** — every concept defined once
- **Concise wins** — every sentence must earn its place

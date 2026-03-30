# Templates

Quick-start templates for spec documents and directory layouts.

## Directory Layout

```
SPECS/
├── 00-principles.md             ← 00: Foundation
├── 01-naming.md                 ←
├── 10-architecture.md           ← 10: Architecture
├── 11-module-structure.md       ←
├── 20-<core-model>.md           ← 20: Core Data Model
├── 30-<api-contracts>.md        ← 30: API & Interfaces
├── 40-<processing>.md           ← 40: Validation / Pipeline
├── 50-cli.md                    ← 50: CLI & Tools
├── 60-<infrastructure>.md       ← 60: Storage / Messaging
├── ref-<standard>.md            ← ref: External standards
└── ref-references.md            ← ref: Project index
```

## Spec Template

```markdown
# <Domain Topic>

## Overview

<2-3 sentences: what this spec defines, what it doesn't.>

## <Core Section>

<Rules, schemas, contracts — use the appropriate pattern:>
<Schema Definition | Rule List | Decision Table | Interface Contract | Flow>

> **Why**: <rationale — factual, 1-3 sentences>
> **Rejected**: <alternatives with brief reasons>

## <Additional Sections>

<Group related concepts together. High cohesion.>

## Terminology

| Term | Definition | Not |
|------|-----------|-----|
| **<term>** | <precise definition> | Not "<synonym to avoid>" |

## Forbidden

- **Don't <X>**: <why> → <what to do instead>

## References

- [<related-spec>.md] — <what it covers>
- [<external-standard>](<url>) — <what it defines>
```

## Inline Decision Record

```markdown
> **Why**: <rationale>
> **Rejected**: <alternatives with reasons>
```

## Extended Decision Record

```markdown
> **Why**: <rationale — 2-3 sentences>
>
> **Rejected**:
> - <Option A> — <concrete reason>
> - <Option B> — <concrete reason>
> - <Option C> — <concrete reason>
```

## Decision Summary Table

```markdown
## Decisions

| Decision | Choice | Why | Rejected |
|----------|--------|-----|----------|
| <topic> | <choice> | <brief rationale> | <alternatives> |
```

## Consolidation Checklist

When merging split docs (e.g., SPECS/ + DESIGNS/) into unified specs:

```markdown
- [ ] Map all source files to domain series (00, 10, 20...)
- [ ] Merge rationale files into their contract counterparts
- [ ] Inline decisions as `> **Why**: ...` blocks
- [ ] Cut "Background & motivation" sections to 1-2 sentences max
- [ ] Promote orphan design docs to new specs in correct series
- [ ] Move non-spec content out (testing guides → dev docs, skills → .claude/skills/)
- [ ] Delete standalone decision logs (content now inline)
- [ ] Update all references (CLAUDE.md, source code, README, cross-refs)
- [ ] Verify no definition is duplicated across files
```

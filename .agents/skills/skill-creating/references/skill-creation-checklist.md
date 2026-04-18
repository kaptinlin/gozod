# Skill Creation Checklist

Use this checklist before considering a skill complete.

## 1. Scope and Intent

- [ ] The skill solves a reusable problem, not a one-off historical case
- [ ] The skill belongs in a reusable skill, not in project-local `CLAUDE.md`
- [ ] The skill's purpose is clear in 1-2 short paragraphs
- [ ] The intended behavior change is explicit
- [ ] The skill has clear non-goals or out-of-scope cases

## 2. Metadata and Discovery

- [ ] `name` uses only lowercase letters, numbers, and hyphens
- [ ] `name` is specific and searchable
- [ ] YAML frontmatter contains only `name` and `description`
- [ ] `description` is written in third person
- [ ] `description` describes when to use the skill, not how it works internally
- [ ] `description` includes concrete triggers, symptoms, contexts, or keywords
- [ ] `description` is specific enough to distinguish this skill from adjacent skills
- [ ] Run `python3 scripts/validate-metadata.py --file SKILL.md`

## 3. Core Structure

- [ ] The main file includes a clear purpose/overview
- [ ] The main file says when to use the skill
- [ ] The main file says when not to use the skill
- [ ] The main file includes core principles or decision rules
- [ ] The main file includes workflow or implementation guidance
- [ ] The main file includes verification or quality checks
- [ ] Optional sections exist only when they add real value

## 4. Boundaries and Instruction Quality

- [ ] Important constraints are explicit, not implied
- [ ] Fragile or safety-critical behavior uses stronger instruction language
- [ ] The skill does not force rigid process where judgment is required
- [ ] The skill teaches decision-making, not just ceremony
- [ ] The guidance avoids filler and repeated explanation
- [ ] The skill does not duplicate large chunks of canonical reference material

## 5. Progressive Disclosure

- [ ] Main `SKILL.md` stays under 500 lines
- [ ] Principles and high-signal guidance stay inline
- [ ] Heavy documentation is moved to `references/`
- [ ] Deterministic helpers are moved to `scripts/`
- [ ] Templates, schemas, or scaffolds are moved to `assets/`
- [ ] Supporting folders exist only when justified by reuse, size, or determinism
- [ ] References are loaded just-in-time instead of force-loading everything up front

## 6. Examples and Supporting Material

- [ ] Examples exist only if they resolve ambiguity better than prose
- [ ] The skill uses one strong example instead of many weak ones
- [ ] Flowcharts are used only for genuinely non-obvious decisions
- [ ] Supporting files have distinct jobs and are not redundant with the main file

## 7. Testing and Validation

- [ ] The skill has been tested before being treated as done
- [ ] Baseline behavior without the skill was observed when appropriate
- [ ] Testing matches the skill type (discipline, technique, pattern, or reference)
- [ ] The skill has been checked for discoverability
- [ ] The skill has been checked for correct application
- [ ] The skill has been checked for misuse, overreach, or boundary failure

## 8. RED-GREEN-REFACTOR

### RED
- [ ] Run a baseline scenario without the skill when behavior change is the point
- [ ] Document failures, ambiguity, or rationalizations the skill must address

### GREEN
- [ ] Write the minimum skill that addresses the real problem
- [ ] Verify the skill improves behavior with the skill present

### REFACTOR
- [ ] Tighten vague wording
- [ ] Close loopholes discovered during testing
- [ ] Remove sections, examples, or files that do not earn their keep
- [ ] Re-test after meaningful changes

## 9. Final Review

- [ ] The skill is concise
- [ ] The skill is discoverable
- [ ] The skill is reusable
- [ ] The skill is bounded
- [ ] The skill is testable
- [ ] The skill is better because of what it omits, not only what it includes

## Validation Commands

```bash
python3 scripts/validate-metadata.py --file SKILL.md
wc -l SKILL.md
```

## Deep References

Read these when needed:
- `references/testing-skill-types.md`
- `testing-skills-with-subagents.md`
- `references/bulletproofing.md`
- `references/cso-guidelines.md`
- `references/error-handling.md`

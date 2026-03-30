# Skill Creation Checklist (TDD Adapted)

**IMPORTANT: Use TodoWrite to create todos for EACH checklist item below.**

## RED Phase - Write Failing Test

- [ ] Create pressure scenarios (3+ combined pressures for discipline skills)
- [ ] Run scenarios WITHOUT skill - document baseline behavior verbatim
- [ ] Identify patterns in rationalizations/failures

## GREEN Phase - Write Minimal Skill

### Metadata Validation
- [ ] Name uses only letters, numbers, hyphens (no parentheses/special chars)
- [ ] YAML frontmatter with only name and description (max 1024 chars)
- [ ] Description starts with "Use when..." and includes specific triggers/symptoms
- [ ] Description written in third person
- [ ] Run `python3 scripts/validate-metadata.py --file SKILL.md` to verify

### Content Structure
- [ ] Clear overview with core principle
- [ ] Keywords throughout for search (errors, symptoms, tools)
- [ ] Address specific baseline failures identified in RED
- [ ] Code inline OR link to separate file
- [ ] One excellent example (not multi-language)

### Testing
- [ ] Run scenarios WITH skill - verify agents now comply

## REFACTOR Phase - Close Loopholes

- [ ] Identify NEW rationalizations from testing
- [ ] Add explicit counters (if discipline skill)
- [ ] Build rationalization table from all test iterations
- [ ] Create red flags list
- [ ] Re-test until bulletproof

## Quality Checks

### Structure
- [ ] Small flowchart only if decision non-obvious
- [ ] Quick reference table
- [ ] Common mistakes section
- [ ] No narrative storytelling
- [ ] Supporting files only for tools or heavy reference

### Progressive Disclosure
- [ ] Main SKILL.md under 500 lines
- [ ] Heavy reference material (>100 lines) moved to `references/`
- [ ] Reusable tools/scripts in `scripts/`
- [ ] Templates and schemas in `assets/`

### CSO (Claude Search Optimization)
- [ ] Description focuses on WHEN to use, not HOW it works
- [ ] No workflow summary in description
- [ ] Includes negative triggers ("Don't use for...")
- [ ] Searchable keywords in overview and sections

## Deployment

- [ ] Commit skill to git and push to your fork (if configured)
- [ ] Consider contributing back via PR (if broadly useful)

## Validation Commands

```bash
# Validate metadata
python3 scripts/validate-metadata.py --file SKILL.md

# Check line count (should be under 500)
wc -l SKILL.md

# Test with subagents (see testing-skills-with-subagents.md)
```

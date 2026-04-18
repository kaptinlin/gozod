# Error Handling for Skill Creation

Common failure modes and recovery steps when creating skills.

## Metadata Validation Failures

**Error: Name contains invalid characters**
```
NAME ERROR: 'Skill_Name' contains invalid characters.
```
**Fix:** Use only lowercase letters, numbers, and single hyphens. No underscores, spaces, or special characters.
```bash
# BAD: Skill_Name, skill.name, Skill-Name
# GOOD: skill-name, skill-name-v2, async-testing
```

**Error: Description too long**
```
DESCRIPTION ERROR: Description is 1150 characters. Must be 1,024 characters or fewer.
```
**Fix:** Shorten description to focus only on triggering conditions. Remove workflow details.

**Error: First/second person in description**
```
STYLE ERROR: Description contains first/second person terms: {'you', 'your'}
```
**Fix:** Rewrite in third person using imperative form.
```yaml
# BAD: Use this when you need to test async code
# GOOD: Use when tests have race conditions or timing dependencies
```

## Context Bloat (>500 lines)

**Problem:** SKILL.md exceeds 500 lines, consuming too much context.

**Recovery:**
1. Identify largest section: `wc -l SKILL.md` and review content
2. Extract heavy content:
   - API docs (>100 lines) -> `references/api-spec.md`
   - Testing details -> `references/testing-methodology.md`
   - Large examples -> `examples/complete-example.md`
3. Replace with reference: "Read `references/[file].md` for [specific info]"
4. Verify: `wc -l SKILL.md` should be <500

## Testing Failures

**Problem:** Agent does not behave better even when the skill is loaded.

**Diagnosis:**
1. Determine whether this is a discovery, application, or boundary failure
2. Check whether the description is trigger-focused or workflow-heavy
3. Verify the skill addresses real observed failures, not hypothetical ones
4. Look for loopholes, vague boundaries, or buried critical guidance

**Recovery:**
1. Review baseline test results - what did the agent actually do?
2. Capture exact rationalizations, ambiguity, or failure patterns
3. Tighten the skill with clearer rules, boundaries, placement, or examples
4. Re-test using the same scenario shape

**Problem:** Agent follows description but not the full skill.

**Fix:** Remove workflow summary from the description. The description should describe WHEN to use the skill, not HOW the skill works internally.

## Deployment Issues

**Problem:** Skill not discoverable by Claude.

**Diagnosis:**
1. Check whether the description starts with clear triggering conditions
2. Verify keywords match likely user phrasing, symptoms, and contexts
3. Ensure the skill is distinguishable from adjacent skills

**Recovery:**
1. Add concrete triggers and symptoms to the description
2. Add searchable terms to headings and body content
3. Test discoverability by asking when this skill should be used

**Problem:** Skill loaded but not applied correctly.

**Diagnosis:**
1. Instructions are too vague or too abstract
2. Decision rules are missing or weak
3. The skill lacks a strong example or clear verification steps

**Recovery:**
1. Add clearer decision rules or workflow guidance
2. Add one strong example if prose is not enough
3. Add verification or common-mistake guidance

## Validation Commands

```bash
# Validate metadata
python3 scripts/validate-metadata.py --file SKILL.md

# Check line count
wc -l SKILL.md

# Test with specific scenario
# (see testing-skills-with-subagents.md)
```

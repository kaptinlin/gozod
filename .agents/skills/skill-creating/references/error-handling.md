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

**Problem:** Agent doesn't comply with skill even when loaded.

**Diagnosis:**
1. Check if description summarizes workflow (anti-pattern)
2. Verify skill addresses specific baseline failures
3. Look for loopholes in instructions

**Recovery:**
1. Review baseline test results - what rationalizations did agent use?
2. Add explicit counters for each rationalization
3. Build rationalization table
4. Re-test with same pressure scenarios

**Problem:** Agent follows description instead of reading full skill.

**Fix:** Remove workflow summary from description. Description should only describe WHEN to use, not HOW it works.

## Deployment Issues

**Problem:** Skill not discoverable by Claude.

**Diagnosis:**
1. Check description starts with "Use when..."
2. Verify keywords match search terms
3. Ensure name is descriptive (verb-first, gerunds)

**Recovery:**
1. Add concrete triggers and symptoms to description
2. Include error messages and tool names in content
3. Test discoverability: "When would I use this skill?"

**Problem:** Skill loaded but not applied correctly.

**Diagnosis:**
1. Instructions too vague or abstract
2. Missing decision points
3. No examples provided

**Recovery:**
1. Add concrete examples with code
2. Include decision trees for non-obvious choices
3. Add "Common Mistakes" section

## Validation Commands

```bash
# Validate metadata
python3 scripts/validate-metadata.py --file SKILL.md

# Check line count
wc -l SKILL.md

# Test with specific scenario
# (see testing-skills-with-subagents.md)
```

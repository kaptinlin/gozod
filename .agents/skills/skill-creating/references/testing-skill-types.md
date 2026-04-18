# Testing Skill Types

Different skill types fail in different ways. Test the behavior the skill is actually responsible for improving.

At minimum, think about three questions for every skill:
- **Discovery:** would the right agent load this skill?
- **Application:** can the agent use it correctly?
- **Boundary:** does the agent know when not to use it?

## 1. Discipline-Enforcing Skills

**What they do:** enforce rules, guardrails, or required process under pressure.

**Examples:** TDD, verification-before-completion, designing-before-coding

### Main failure modes
- agent understands the rule but breaks it under pressure
- agent rationalizes exceptions
- agent follows the spirit it invented instead of the literal boundary
- agent cites the skill but still chooses the wrong action

### Test with
- baseline scenarios without the skill
- pressure scenarios with time, sunk cost, authority, exhaustion, or pragmatism pressure
- scenarios combining 3+ pressures
- loophole tests based on actual rationalizations

### Success criteria
- the agent chooses the compliant action under pressure
- the agent resists common rationalizations
- the skill closes known loopholes explicitly
- the boundary is enforceable, not merely aspirational

### Notes
This is the skill type most likely to need RED-GREEN-REFACTOR testing with pressure scenarios.

Read `testing-skills-with-subagents.md` and `references/bulletproofing.md` for deeper guidance.

## 2. Technique Skills

**What they do:** teach a repeatable method, workflow, or procedure.

**Examples:** condition-based-waiting, root-cause-tracing, defensive-programming

### Main failure modes
- the agent cannot translate the technique into action
- the instructions are too vague to apply
- the technique works only for the happy path
- the skill is overly rigid where judgment is needed

### Test with
- application scenarios on fresh problems
- variation scenarios with small contextual changes
- edge cases and missing-information scenarios
- comparison against a weaker or incorrect approach

### Success criteria
- the agent applies the technique correctly in a new scenario
- the agent adapts the method without losing the core rule
- the skill remains useful outside the original example
- the agent can explain why the technique fits the case

### Notes
Technique skills often need one strong example plus scenario-based application testing.

## 3. Pattern Skills

**What they do:** teach a mental model, heuristic, or way of seeing a problem.

**Examples:** reducing-complexity, information-hiding, flatten-with-flags

### Main failure modes
- the agent fails to recognize when the pattern applies
- the agent applies the pattern too broadly
- the skill states the pattern abstractly but does not improve decisions
- the agent cannot distinguish the pattern from adjacent ones

### Test with
- recognition scenarios
- classification scenarios: when does this pattern fit vs not fit?
- counter-examples and non-examples
- transfer scenarios in a different domain or codebase shape

### Success criteria
- the agent correctly identifies when the pattern applies
- the agent correctly identifies when it does not apply
- the pattern improves decisions, not just vocabulary
- the agent can map the pattern to a concrete action or recommendation

### Notes
Pattern skills need boundary testing more than rigid procedural testing.

## 4. Reference Skills

**What they do:** provide compact documentation, command knowledge, API guidance, or factual lookup structure.

**Examples:** API docs, command references, library guides

### Main failure modes
- the right information cannot be found quickly
- the structure is too verbose to scan
- common use cases are missing
- the agent retrieves information but misapplies it

### Test with
- retrieval scenarios
- direct lookup tasks for common cases
- application scenarios that require using the retrieved information correctly
- gap testing for likely tasks and missing entries

### Success criteria
- the agent finds the right information quickly
- the information is structured for scanning
- the agent applies the retrieved information correctly
- common tasks do not require guessing or searching elsewhere first

### Notes
Reference skills usually need stronger retrieval and coverage testing than pressure testing.

## Choosing the Test Shape

Use the lightest test that can still falsify the skill.

- If the skill exists to resist bad decisions under pressure, test pressure.
- If the skill exists to teach a method, test application.
- If the skill exists to sharpen judgment, test recognition and boundaries.
- If the skill exists to make information usable, test retrieval and coverage.

Many skills mix types. In that case, test the dominant failure mode first.

## Minimum Validation for Any Skill

Before shipping any skill, confirm:
- the metadata would cause discovery in the right situations
- the agent can apply the skill to at least one realistic scenario
- the agent does not over-apply it outside its boundary
- the skill's structure supports its purpose instead of getting in the way

## Related References

- `testing-skills-with-subagents.md`
- `references/bulletproofing.md`
- `references/skill-creation-checklist.md`
- `references/cso-guidelines.md`

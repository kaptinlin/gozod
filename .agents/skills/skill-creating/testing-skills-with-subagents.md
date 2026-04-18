# Testing Skills With Subagents

Load this reference when you need evidence that a skill actually improves agent behavior.

## Purpose

Testing skills is TDD applied to process documentation.

The goal is not to prove the skill sounds good. The goal is to verify that an agent can:
- discover the skill in the right situation
- apply it correctly
- stay inside its boundary
- resist the failures the skill was written to prevent

**Core principle:** If you did not observe baseline behavior without the skill, you do not yet know what the skill needs to change.

**REQUIRED BACKGROUND:** You MUST understand `superpowers:test-driven-development` before using this reference.

## When to Use This Reference

Use this reference when the skill:
- is supposed to change behavior, not just provide lookup material
- enforces discipline or constraints
- could be rationalized away under pressure
- teaches a method that must work on new scenarios
- teaches a pattern whose boundary must be tested

Do not overuse this reference for pure lookup docs where retrieval and coverage testing is enough. For that case, also read `references/testing-skill-types.md`.

## The Testing Model

Every skill test should cover the dimensions that matter:

### 1. Discovery
Would the right agent load this skill in the right situation?

### 2. Application
Can the agent use the skill to make a better decision or produce better work?

### 3. Boundary
Does the agent know when the skill does not apply, or when it must stop overreaching?

Discipline-enforcing skills usually need the strongest boundary testing.

## RED-GREEN-REFACTOR for Skills

| Phase | Goal | What to verify |
|------|------|----------------|
| **RED** | Observe baseline behavior | What the agent does without the skill |
| **GREEN** | Write the minimum useful skill | Whether the skill improves behavior |
| **REFACTOR** | Tighten wording and close loopholes | Whether the improvement survives pressure and variation |

## RED: Baseline Testing

### Goal
Run a realistic scenario **without** the skill and capture what actually happens.

### What to record
- the action the agent chose
- the reasoning it used
- the exact rationalizations or shortcuts it invented
- where the boundary failed
- whether the failure was discovery, application, or boundary related

### What makes a good baseline scenario
- concrete task, not an academic prompt
- real tradeoffs or pressure
- explicit options or an action requirement
- enough realism that the agent has to choose, not merely describe

### Bad baseline prompt
```markdown
What does the skill say about writing tests?
```

This tests recall, not behavior.

### Better baseline prompt
```markdown
IMPORTANT: This is a real scenario. Choose and act.

You spent 4 hours implementing a feature. It works.
You manually tested edge cases. It's 6pm.
Code review is tomorrow morning. You just realized you skipped TDD.

Options:
A) Delete the code and restart with TDD tomorrow
B) Commit now and add tests later
C) Write tests now, then commit

Choose A, B, or C.
```

This forces a decision and reveals rationalization.

## GREEN: Write the Minimum Skill That Changes Behavior

Write only enough to address the failures you actually observed.

Do not patch hypothetical problems first. Start with the real gap.

After drafting the skill, re-run the same scenario **with** the skill.

### GREEN passes when
- the agent makes the better choice
- the agent uses the intended decision rule
- the skill improves behavior rather than just being quoted back

If the agent still fails, the skill is unclear, incomplete, badly placed, or too weak for the pressure you are testing.

## REFACTOR: Tighten and Re-Test

If the agent finds a loophole, your next job is not to argue with it. Your next job is to improve the skill.

### Capture new failure modes verbatim
Examples:
- "This case is different because..."
- "I'm following the spirit, not the letter"
- "Being pragmatic means adapting"
- "I already tested it manually"
- "I'll keep this as reference while doing it correctly"

### Tighten the skill by doing one or more of these
- make the rule more explicit
- add an explicit negation for the loophole
- move critical guidance higher in the file
- strengthen discovery language in the description
- add red flags or anti-patterns
- remove ambiguity that invites reinterpretation

### Re-test after each meaningful change
A refactor is only complete if the updated skill still holds under re-test.

## Pressure Testing

Pressure testing matters when the skill is supposed to prevent bad behavior under temptation.

### Useful pressure types
| Pressure | Example |
|----------|---------|
| Time | deadline, deploy window, outage |
| Sunk cost | hours of work already invested |
| Authority | manager or senior suggests violating the rule |
| Economic | money, promotion, incident cost |
| Exhaustion | end of day, low energy, wants to leave |
| Social | fear of looking rigid or dogmatic |
| Pragmatic framing | "just be practical" |

Best discipline scenarios usually combine 3+ pressures.

### Key properties of a strong pressure scenario
1. concrete options or forced action
2. real consequences
3. enough detail to feel operational
4. no easy escape into vague advice
5. realistic temptation to violate the rule

## Distinguish Failure Types

When a test fails, classify it:

### Discovery failure
The agent never loads the skill or does not realize it applies.

**Fixes often include:**
- improving the description
- adding searchable triggers and symptoms
- clarifying adjacent boundaries

### Application failure
The agent loads the skill but cannot turn it into correct action.

**Fixes often include:**
- clearer workflow
- better principles or decision rules
- one stronger example
- less ambiguity in instructions

### Boundary failure
The agent loads the skill but misapplies it, over-applies it, or rationalizes an exception.

**Fixes often include:**
- explicit non-goals
- stronger must / must not language
- anti-patterns
- loophole-closing language

## Meta-Testing

Use meta-testing when the agent still fails and you need to know why.

Ask a targeted question like:

```markdown
You read the skill and still chose Option C.

What change to the skill would have made the correct action unmistakable?
```

Typical answers reveal one of three problems:

1. **The skill was clear, but weakly binding**
   - strengthen foundational principles or explicit constraints

2. **The skill omitted the needed rule**
   - add the missing guidance directly

3. **The skill buried the important rule**
   - reorganize so the critical boundary appears earlier and more clearly

## When a Skill Is Strong Enough

A tested skill is in good shape when:
- the agent discovers it in the right situations
- the agent uses it to improve decisions or outputs
- the agent avoids obvious misuse or overreach
- the agent holds up under realistic variation or pressure
- repeated tests stop revealing new major loopholes

Do not confuse "passed once" with "finished forever." Test to the level justified by the skill's risk and purpose.

## Common Mistakes

**Mistake: testing recall instead of behavior**
- Asking what the skill says is weaker than making the agent act.

**Mistake: writing the skill before seeing baseline failure**
- This captures what you imagine is wrong, not what the agent actually does.

**Mistake: fixing failures with vague language**
- "Don't cheat" is weaker than naming the actual loophole.

**Mistake: over-testing simple reference skills with pressure scenarios**
- Use the test shape that matches the skill type.

**Mistake: stopping after the first successful run**
- Re-test when the skill's job is to survive variation, pressure, or temptation.

## Quick Test Checklist

- [ ] Choose the right test shape for the skill type
- [ ] Run baseline without the skill when behavior change is the goal
- [ ] Capture failures as discovery, application, or boundary issues
- [ ] Write the minimum skill that addresses the real failure
- [ ] Re-test with the skill present
- [ ] Tighten wording, placement, or boundaries when loopholes appear
- [ ] Re-test after meaningful changes

## Related References

- `references/testing-skill-types.md`
- `references/bulletproofing.md`
- `references/skill-creation-checklist.md`
- `references/cso-guidelines.md`
- `persuasion-principles.md`

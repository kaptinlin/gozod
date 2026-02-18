---
name: ralphy-todo-creating
description: Creates Ralphy TODO.yaml task files for Go projects by analyzing source code, PRDs, PLAN.md, REFACTOR.md, or issue trackers. Use when the user asks to create a TODO.yaml, generate a ralphy task list, plan tasks from a PRD or design doc, or convert a plan/refactor document into ralphy tasks in a Go project.
---

# Create Ralphy TODO.yaml for Go Projects

Generate a `TODO.yaml` task file by analyzing project source material. Tasks must be specific, traceable, and actionable.

## Step 1: Identify Task Sources

Determine where tasks come from. Ask the user if unclear. Common sources:

- **PRD / design doc**: `PRD.md`, `DESIGN.md`, or similar
- **Plan file**: `PLAN.md` with implementation steps
- **Refactor file**: `REFACTOR.md` with refactoring targets
- **GitHub issues**: `gh issue list`
- **Code analysis**: TODOs, missing tests, lint warnings in source
- **User request**: Direct description of what to build

## Step 2: Analyze Source and Extract Tasks

Read the source material thoroughly. For each task, identify:

1. **What** to implement or change
2. **Where** in the codebase (specific file or package)
3. **Why** (the goal or requirement it fulfills)
4. **Source** (which document, section, or issue it comes from)

## Step 3: Write TODO.yaml

```yaml
tasks:
  - title: "<action> <what> in <where> [per <source>]"
    completed: false
    # parallel_group: N    # optional: same group = runs concurrently
    # description: ""      # optional: acceptance criteria or details
```

### Title Format

Titles must be **specific and traceable**. Include:

- **Action**: imperative verb (add, implement, refactor, fix, extract, move)
- **What**: the concrete feature, function, or component
- **Where**: target file, package, or module
- **Source reference** (when applicable): which doc/section/issue the task comes from

### Title Examples

```yaml
# Good - specific, traceable, says where and what
- title: "add context.Context parameter to all public methods in pkg/store per DESIGN.md section 3"
- title: "implement retry logic with exponential backoff in internal/client/http.go per PRD functional requirements"
- title: "refactor FSM.Fire() in fsm.go to return typed error instead of generic error per REFACTOR.md"
- title: "extract validation logic from handler.go into pkg/validate/rules.go per PLAN.md step 2"
- title: "add table-driven tests for ParseConfig in config_test.go covering edge cases from issue #42"
- title: "move database connection pooling from main.go into internal/db/pool.go per REFACTOR.md section 1"

# Bad - vague, no location, no traceability
- title: "add validation"
- title: "fix error handling"
- title: "add tests"
- title: "refactor code"
```

### Task Rules

- Each `title` must be unique
- Imperative mood (e.g., "add X", not "adding X")
- `completed: false` for all new tasks
- Use `parallel_group` when tasks touch different files/packages and have no dependencies
- Tasks in higher-numbered groups run after lower-numbered groups complete
- Use `description` for acceptance criteria or non-obvious details

### Parallel Group Guidelines

- **Same group**: tasks that touch independent files/packages with no shared state
- **Sequential groups**: when a later task depends on output from an earlier one

```yaml
tasks:
  # Group 1: independent model/package changes
  - title: "add User struct with validation tags in internal/model/user.go per PRD data model"
    completed: false
    parallel_group: 1
  - title: "add Role enum and permission constants in internal/model/role.go per PRD data model"
    completed: false
    parallel_group: 1

  # Group 2: depends on models from group 1
  - title: "implement UserService.Create with role assignment in internal/service/user.go per PRD user management"
    completed: false
    parallel_group: 2
  - title: "implement RoleService.Assign with permission checks in internal/service/role.go per PRD access control"
    completed: false
    parallel_group: 2

  # Group 3: integration depends on services from group 2
  - title: "add integration tests for user creation and role assignment flow in test/integration/user_test.go"
    completed: false
    parallel_group: 3
```

## Step 4: Verify

Before writing `TODO.yaml`, verify:

- [ ] Every task references a specific file or package
- [ ] Every task derived from a document cites its source
- [ ] No two titles are identical
- [ ] Parallel groups respect dependency ordering
- [ ] Tasks are granular enough for a single focused coding session

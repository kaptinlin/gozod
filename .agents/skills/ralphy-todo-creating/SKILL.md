---
description: Creates Ralphy TODO.yaml task files by analyzing source code, PRDs, PLAN.md, REFACTOR.md, or issue trackers. Use when the user asks to create a TODO.yaml, generate a ralphy task list, or convert a plan/refactor document into ralphy tasks.
name: ralphy-todo-creating
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
  - title: "<action> <what> in <where> per <document section>"
    completed: false
    # parallel_group: N    # optional, rarely needed: same group = runs concurrently
```

### Allowed Fields

Only three fields are permitted:

- **title** (required): specific, traceable task description
- **completed** (required): always `false` for new tasks
- **parallel_group** (optional): use only when tasks are truly independent and can run concurrently. Most tasks should NOT use this field.

**CRITICAL:** DO NOT include `description` field. Only use `title`, `completed`, and optionally `parallel_group`.

### Title Format

Titles must be **specific and traceable**. Include:

- **Action**: imperative verb (add, implement, refactor, fix, extract, move)
- **What**: the concrete feature, function, or component
- **Where**: target file, package, or module
- **Source reference** (required for document-based tasks): which doc/section the task comes from (e.g., "per PRD.md section 3", "per DESIGN.md authentication flow", "per issue #42")

### Title Examples

```yaml
# Good - specific, traceable, says where and what, references document section
- title: "add context.Context parameter to all public methods in pkg/store per DESIGN.md section 3"
  completed: false
- title: "implement retry logic with exponential backoff in internal/client/http.go per PRD.md functional requirements section"
  completed: false
- title: "refactor FSM.Fire() in fsm.go to return typed error per REFACTOR.md error handling section"
  completed: false
- title: "extract validation logic from handler.go into pkg/validate/rules.go per PLAN.md step 2"
  completed: false
- title: "add table-driven tests for ParseConfig in config_test.go per issue #42"
  completed: false

# Bad - vague, no location, no document section reference
- title: "add validation"
- title: "fix error handling per PRD.md"  # missing section reference
- title: "add tests"
- title: "refactor code per DESIGN.md"  # missing section reference
```

### Task Rules

- Each `title` must be unique
- Imperative mood (e.g., "add X", not "adding X")
- `completed: false` for all new tasks
- **Document-based tasks must reference specific sections** (e.g., "per PRD.md section 3", "per DESIGN.md authentication flow")
- `parallel_group` is optional and rarely needed. Only use when tasks touch completely independent files/packages with no dependencies
- Most tasks should NOT specify `parallel_group` - sequential execution is the default and preferred approach

### Parallel Group Guidelines (Use Sparingly)

Only use `parallel_group` when:
- Tasks touch completely different files/packages
- Tasks have zero dependencies on each other
- Running them concurrently provides clear benefits

**Most tasks should omit this field.** Sequential execution is simpler and safer.

```yaml
tasks:
  # Group 1: independent model/package changes
  - title: "add User struct with validation tags in internal/model/user.go per PRD.md data model section"
    completed: false
    parallel_group: 1
  - title: "add Role enum and permission constants in internal/model/role.go per PRD.md data model section"
    completed: false
    parallel_group: 1

  # Group 2: depends on models from group 1
  - title: "implement UserService.Create with role assignment in internal/service/user.go per PRD.md user management section"
    completed: false
    parallel_group: 2
  - title: "implement RoleService.Assign with permission checks in internal/service/role.go per PRD.md access control section"
    completed: false
    parallel_group: 2

  # Group 3: integration depends on services from group 2
  - title: "add integration tests for user creation and role assignment flow in test/integration/user_test.go per PRD.md testing requirements"
    completed: false
    parallel_group: 3
```

## Step 4: Verify

Before writing `TODO.yaml`, verify:

- [ ] Every task references a specific file or package
- [ ] Every task derived from a document cites the specific section (e.g., "per PRD.md section 3", not just "per PRD.md")
- [ ] No two titles are identical
- [ ] Only three fields are used: title, completed, and parallel_group (optional)
- [ ] Most tasks omit parallel_group unless truly independent
- [ ] Tasks are granular enough for a single focused coding session

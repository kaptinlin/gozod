# Agent-Friendly CLI Design

Design CLI commands that AI agents can reliably use — structured output, token efficiency, safety guardrails, and self-documenting interfaces. Built on go-command's Mode/Verbosity/Schema system.

> References: [CLIG](https://clig.dev/), [Agent-Friendly CLI (Speakeasy)](https://www.speakeasy.com/blog/engineering-agent-friendly-cli), [Hidden Users (nibzard)](https://www.nibzard.com/ai-native), [DEV Community Agent CLI Guide](https://dev.to/uenyioha/writing-cli-tools-that-ai-agents-actually-want-to-use-39no), go-command `.references/agent/` (15+ projects)

---

## Core Principle

The terminal is an AI runtime. Agents don't read prose, don't wait for prompts, and count every token. Design CLI that serves both humans and agents from the same codebase — go-command's Mode system makes this automatic.

```
Human → colored text, aligned tables, progress spinners, next-step suggestions
Agent → flat JSON, minimal tokens, structured errors, machine-readable schema
CI    → plain text, exit codes, no color
```

---

## The 12 Agent-CLI Rules

### Rule 1: Structured Output with Respond[T]

Every command returns typed data via `Respond[T]()`. Agents parse reliably when the shape is predictable.

```go
type Task struct {
    ID     int    `json:"id"`
    Title  string `json:"title"`
    Status string `json:"status"`
    Owner  string `json:"owner"`
}

Handler: func(ctx *command.Context) error {
    task := getTask(ctx.Args[0])

    // Auto-adapts to mode:
    // Human: ctx.Render(task) → formatted output
    // Agent: {"data": {...}, "meta": {...}} JSON envelope
    return command.Respond(ctx, task, command.ResponseMeta{
        NextActions: []string{
            "myapp task update " + ctx.Args[0],
            "myapp task assign " + ctx.Args[0],
        },
    })
}
```

**Agent JSON output:**
```json
{
  "data": {"id": 42, "title": "Fix login", "status": "open", "owner": "alice"},
  "meta": {
    "next_actions": [
      "myapp task update 42",
      "myapp task assign 42"
    ]
  }
}
```

For simple output, `ctx.Props()` and `ctx.Table()` also auto-adapt:

```go
ctx.Props("Name", task.Title, "Status", task.Status, "Owner", task.Owner)
ctx.Table([]string{"ID", "Title", "Status"}, rows)
```

### Rule 2: Token-Efficient Output (Verbosity Levels)

Agents pay per token. Provide compact output by default, detail on demand.

```go
// Agent mode defaults to VerbosityMinimal
// Minimal (~50 tokens per item): {"id":42,"title":"Fix login","status":"open"}
// Compact (~150 tokens): + description, timestamps
// Detailed (~500 tokens): + comments, history, metadata

Handler: func(ctx *command.Context) error {
    items := listTasks()

    switch ctx.Verbosity() {
    case command.VerbosityMinimal:
        // Agent: just ID + title + status
        rows := make([][]string, len(items))
        for i, item := range items {
            rows[i] = []string{strconv.Itoa(item.ID), item.Title, item.Status}
        }
        ctx.Table([]string{"ID", "Title", "Status"}, rows)

    case command.VerbosityCompact:
        // Human: add owner, created date
        // ...

    case command.VerbosityDetailed:
        // --verbose: everything
        // ...
    }
    return nil
}
```

**Token budget guideline:**

| Operation | Target Tokens | Strategy |
|-----------|--------------|----------|
| List (10 items) | < 500 | Name + key fields only |
| Show (1 item) | < 800 | Full spec without comments |
| Error | < 200 | Message + hint + code |
| Help | < 300 | Usage + flags + 2 examples |

### Rule 3: Zero Interactive Prompts

Agents cannot respond to prompts. Use `ctx.IsInteractive()` to detect and adapt.

```go
cmd := &command.Command{
    Name:    "create",
    MinArgs: 1,
    Options: []command.Option{
        {Name: "role", Required: true, Enum: []string{"admin", "member", "viewer"}},
        command.DryRunOption(),
    },
    Handler: func(ctx *command.Context) error {
        if ctx.IsDryRun() {
            ctx.Props("Action", "create", "User", ctx.Args[0], "Role", ctx.String("role"))
            return nil
        }

        // Only prompt in interactive human mode
        if ctx.IsInteractive() {
            ok, err := ctx.Confirm("Create user " + ctx.Args[0] + "?")
            if err != nil {
                return err
            }
            if !ok {
                ctx.Infof("Cancelled")
                return nil
            }
        }

        return doCreate(ctx)
    },
}
```

**Rule:** In Agent/CI mode, never prompt. If required input is missing, return a structured error with the flag name.

### Rule 4: Semantic Exit Codes

Agents branch on exit codes. Make them meaningful beyond 0/1.

```go
// go-command exit codes
command.ExitSuccess    = 0  // Success (warnings OK)
command.ExitError      = 1  // Runtime error, validation failure
command.ExitUsageError = 2  // Wrong args, unknown flag

// Extended codes for agent-specific scenarios
const (
    ExitAuthError    = 3  // Authentication/authorization failure
    ExitNotFound     = 4  // Resource not found
    ExitConflict     = 5  // Already exists / conflict
)
```

**Why agents care:** An agent seeing exit code 3 knows to refresh credentials. Exit code 4 means retry with different input. Exit code 1 is generic — avoid it when a more specific code exists.

### Rule 5: Dry-Run with DryRunOption()

Agents validate before committing. Use go-command's preset for consistency.

```go
cmd := &command.Command{
    Name: "deploy",
    Options: []command.Option{
        command.DryRunOption(),  // --dry-run / -n (preset)
        {Name: "env", Enum: []string{"staging", "production"}, Default: "staging"},
    },
    Handler: func(ctx *command.Context) error {
        changes := computeChanges()

        if ctx.IsDryRun() {
            // Return what WOULD happen
            return command.Respond(ctx, changes, command.ResponseMeta{
                NextActions: []string{"myapp deploy --env " + ctx.String("env")},
            })
        }

        // Execute for real
        return applyChanges(changes)
    },
}
```

**Agent workflow:** `cmd -n` → parse preview → decide → `cmd` (execute).

### Rule 6: Self-Documenting Schema

Agents discover capabilities via schema, not by reading README. go-command's `Schema()` export enables this.

```go
// Export complete command tree as JSON
schema := app.Schema()

// Expose as CLI command
app.Add(&command.Command{
    Name:   "schema",
    Desc:   "export command schema as JSON",
    Hidden: true,  // for agents, not humans
    Handler: func(ctx *command.Context) error {
        return ctx.JSON(app.Schema())
    },
})
```

**Agent discovery flow:**
1. `myapp schema` → full command tree
2. `myapp <command> --help` → specific usage
3. Execute with confidence

### Rule 7: Next-Actions Guidance

Every response tells the agent what to do next. Prevents agents from "getting stuck."

```go
Handler: func(ctx *command.Context) error {
    result := deployProject(ctx.Args[0])

    // Respond includes next_actions in Agent JSON automatically
    return command.Respond(ctx, result, command.ResponseMeta{
        NextActions: []string{
            "myapp deploy status " + ctx.Args[0],
            "myapp logs " + ctx.Args[0],
        },
    })
}
```

Errors also support next actions:

```go
return &command.Error{
    Code:    command.ExitError,
    Message: "deployment failed: image not found",
    NextActions: []string{
        "myapp build " + ctx.Args[0],
        "myapp image list",
    },
}
```

### Rule 8: Safety Guardrails

Agents need explicit permission for destructive operations. Use `ctx.IsInteractive()` and `ctx.Confirm()`.

```go
// Pattern 1: Confirm for destructive ops
cmd := &command.Command{
    Name: "rm",
    Options: []command.Option{command.DryRunOption()},
    Handler: func(ctx *command.Context) error {
        if ctx.IsDryRun() {
            ctx.Props("Action", "delete", "Target", ctx.Args[0])
            return nil
        }

        // Interactive: prompt confirmation
        // Non-interactive (Agent/CI): reject unless --yes
        if ctx.IsInteractive() {
            ok, err := ctx.Confirm("Delete " + ctx.Args[0] + "?")
            if err != nil {
                return err
            }
            if !ok {
                return nil
            }
        }

        return doDelete(ctx.Args[0])
    },
}

// Pattern 2: Read-only mode via environment
// MYAPP_READONLY=1 blocks all write commands
Handler: func(ctx *command.Context) error {
    if os.Getenv("MYAPP_READONLY") == "1" {
        return &command.Error{
            Code:    command.ExitError,
            Message: "write operations disabled in read-only mode",
            Hint:    "unset MYAPP_READONLY to enable writes",
        }
    }
    // ...
}
```

### Rule 9: Idempotent Operations

Agents may retry commands. Make operations safe to re-run.

```go
Handler: func(ctx *command.Context) error {
    name := ctx.Args[0]

    // Idempotent: create only if not exists
    if exists(name) {
        ctx.Infof("User %s already exists, skipping", name)
        return nil  // exit 0, not error
    }

    return create(name)
}
```

**Idempotency checklist:**
- `create` → check existence first, skip if exists (or use `--if-not-exists`)
- `push` → diff before push, skip if identical
- `check` → always safe to re-run
- `build` → always safe to re-run (overwrite output)

### Rule 10: Batch Operations

Agents prefer one command over N sequential calls. Support batch input.

```go
// Accept multiple targets
cmd := &command.Command{
    Name:    "check",
    MinArgs: 0,  // 0 = check all
    // MaxArgs not set = unlimited
    Handler: func(ctx *command.Context) error {
        if len(ctx.Args) == 0 {
            return checkAll()
        }
        return checkSpecific(ctx.Args)
    },
}
```

### Rule 11: Composability (stdin/stdout/pipes)

```go
Handler: func(ctx *command.Context) error {
    // Detect piped input
    if ctx.HasPipedInput() {
        data, err := io.ReadAll(ctx.Stdin)
        if err != nil {
            return err
        }
        return processData(data)
    }

    // Interactive: read from args
    return processFile(ctx.Args[0])
}
```

**stdout/stderr discipline:**
- `stdout` → data only (parseable by next command)
- `stderr` → status, progress, warnings, errors
- Agent mode: stdout is clean JSON, stderr is empty

### Rule 12: Stable Output Contract

Never break stdout format after release. Agents depend on field names and structure.

```go
// Good — additive changes only
// v1: {"id": 1, "name": "alice"}
// v2: {"id": 1, "name": "alice", "role": "admin"}  // added field OK

// Bad — breaking change
// v1: {"id": 1, "name": "alice"}
// v2: {"user_id": 1, "username": "alice"}  // renamed fields = breakage
```

---

## go-command Agent Features

### Mode Auto-Detection

```go
import "github.com/agentable/go-command/middleware/mode"

// Detection priority (mode middleware)
// 1. --mode flag (explicit)
// 2. Alias flags (--json -> Agent, --ci -> CI)
// 3. COMMAND_MODE env var
// 4. DefaultMode (if set, skips below)
// 5. CI env detection (CI, GITHUB_ACTIONS, GITLAB_CI, JENKINS_URL, CIRCLECI)
// 6. TTY check (non-TTY -> Agent)

app.Use(mode.New(mode.Config{
    Aliases: map[string]command.Mode{
        "json": command.ModeAgent,
    },
}))

// For agent-native CLIs (default Agent, humans opt in with --human):
app.Use(mode.New(mode.AgentNative()))
```

### Verbosity Auto-Tuning

```go
// Agent mode automatically sets VerbosityMinimal
// Human mode defaults to VerbosityCompact
// --verbose sets VerbosityDetailed
```

### Schema Export for MCP

```go
// MCP server auto-generates tool definitions from CLI schema
schema := app.Schema()

// Each CommandSchema maps to an MCP tool:
// - Name → tool name (snake_case)
// - Desc → tool description
// - Options → input_schema properties
// - MinArgs/MaxArgs → required positional parameters
```

### Query Middleware

Filter and transform output with jq expressions via `middleware/query` (powered by [aq](https://github.com/agentable/aq)).

```go
import "github.com/agentable/go-command/middleware/query"

// Register middleware + options
app.Use(query.New())
app.AddOptions(query.Options()...)
// Adds: --query/-q (jq expression) and --output-format/-O (json/yaml/csv/tson)

// Usage:
// myapp task list --query '.[] | select(.status == "open")'
// myapp task show 42 --query '.owner'
// myapp task list --query '.[] | .title' --output-format csv
```

When `--query` is set, the middleware:
1. Buffers stdout and forces Agent mode (so handler outputs JSON)
2. Runs the handler
3. Extracts `data` from Respond envelope if present
4. Applies jq expression via `aq.Query()`
5. Re-encodes in target format

When `--query` is NOT set: zero overhead, fully transparent.

### Signal Handling

`app.Run()` automatically registers SIGINT/SIGTERM handlers that cancel the context. `Exec()` mode does NOT register signal handlers (safe for tests).

```go
// In handlers, check for cancellation via ctx.Context():
Handler: func(ctx *command.Context) error {
    for _, item := range items {
        select {
        case <-ctx.Context().Done():
            return ctx.Context().Err()
        default:
            process(item)
        }
    }
    return nil
}
```

The `Hooks` type provides lifecycle hooks: `OnCommandAdd`, `OnBeforeRun`, `OnAfterRun`, `OnPreShutdown`, `OnPostShutdown`.

---

## Agent-Specific Patterns

### Pattern: Progressive Context Loading

Agents should load minimal data first, then drill down. Design commands to support this.

```bash
# Level 1: Index (~200 tokens) — what exists?
myapp task list                           # ID + title + status only

# Level 2: Detail (~800 tokens) — one item
myapp task show 42                        # full task details

# Level 3: Deep (~2000 tokens) — with related data
myapp task show 42 --verbose              # task + comments + history
```

### Pattern: Error Recovery Chain

```json
{
  "error": "user 'alic' not found",
  "code": "ERR_NOT_FOUND",
  "hint": "did you mean 'alice'?",
  "next_actions": [
    "myapp user list",
    "myapp user show alice"
  ]
}
```

Agent reads `next_actions` and tries the suggested commands.

### Pattern: Filtering to Reduce Tokens

```bash
# Bad — returns everything, agent parses in context (expensive)
myapp task list --verbose

# Good — server-side filtering (cheap)
myapp task list --status open --owner alice
myapp user list --role admin
```

Design commands with filtering flags so agents don't need to process excess data.

### Pattern: Field Selection with Query

```bash
# Return only needed fields
myapp task list --query '.[].title'
myapp user show alice --query '{name, email}'
```

---

## Anti-Patterns for Agent CLI

| Anti-Pattern | Impact | Fix |
|-------------|--------|-----|
| Interactive prompts in Agent mode | Agent hangs forever | `ctx.IsInteractive()` check |
| Colored output in JSON | JSON parsing breaks | go-command strips color in Agent mode |
| Progress spinners to stdout | Pollutes JSON stream | `ctx.Spinner()` outputs to stderr only |
| Verbose errors without code | Agent can't branch | Use `ErrorCode` + `NextActions` |
| Nested JSON (5+ levels deep) | Harder to parse, more tokens | Flat structures, max 2-3 levels |
| Human-readable dates ("2 days ago") | Agent can't compute | ISO 8601 in Agent mode |
| Silent success (exit 0, no output) | Agent doesn't know what happened | Always use `Respond[T]()` |
| Large default output | Wastes context tokens | Minimal by default, --verbose for detail |
| Unstable field names | Agent's parser breaks | Never rename fields post-release |
| Requiring browser OAuth | Agent can't open browser | Support token-based auth via flag/env |
| Manual `--dry-run` definition | Inconsistent across commands | Use `command.DryRunOption()` preset |
| Manual JSON envelope | Inconsistent structure | Use `Respond[T]()` for standard envelope |

---

## Checklist: Is Your Command Agent-Ready?

- [ ] Uses `Respond[T]()` or `ctx.Props()` / `ctx.Table()` for output
- [ ] No interactive prompts when `ctx.IsInteractive()` is false
- [ ] Has `command.DryRunOption()` + `ctx.IsDryRun()` for write operations
- [ ] Error includes `ErrorCode` + `Hint` + `NextActions`
- [ ] Exit code is semantic (0/1/2 minimum)
- [ ] Minimal output by default (< 500 tokens for list)
- [ ] Supports filtering flags to reduce output
- [ ] Uses `ctx.HasPipedInput()` for stdin detection
- [ ] stdout = data, stderr = status (`ctx.Spinner()` for progress)
- [ ] Examples in help text (agents read these)
- [ ] Idempotent (safe to retry)
- [ ] `ResponseMeta.NextActions` in success response

---
description: Design world-class CLI commands using go-command. Use when creating new CLI commands, designing subcommand structure, choosing between flags and positional arguments, implementing option constraints, or structuring command groups.
name: go-command-designing
---


# CLI Command Design with go-command

Design world-class CLI commands following Docker/Git conventions and CLIG principles, implemented with `github.com/agentable/go-command`.

## When to Apply

- Creating new CLI commands or subcommands
- Designing flag and argument interfaces
- Choosing between positional args vs flags
- Structuring command groups and hierarchies
- Implementing option constraints (exclusive, required-together, conflicts)
- Reviewing CLI code for consistency and usability
- Adding help text, examples, and completions

## Quick Reference

| Principle | Rule | Example |
|-----------|------|---------|
| Command pattern | `<noun> <verb> [target] [--flags]` | `docker container rm nginx` |
| Primary target | Positional argument | `git checkout main` |
| Modifiers | Flags with long + short | `--force` / `-f` |
| Boolean flags | `--flag` / `--no-flag` | `--verbose` / `--no-verbose` |
| Standard shorts | `-n` dry-run, `-f` force, `-o` output, `-q` quiet | `rm -f`, `make -n` |
| Human first | Default human output, `--json` for machines | `docker ps --format json` |
| Suggest next step | After success, show what to do next | `git push` after commit |
| Fail with guidance | Errors include what went wrong + how to fix | Hint field in Error |
| Dry-run for writes | `DryRunOption()` on every write command | `deploy -n` to preview |
| Structured response | `Respond[T]()` for typed output | Auto JSON envelope in Agent |
| **Type-safe config** | **`ctx.Bind()` + `gozod.FromStruct[T]()`** | **Typed struct + validation** |
| **Thin handlers** | **50-150 lines, delegate to internal/** | **See layering.md** |

## Design Rules

### Rule 1: Noun-Verb Hierarchy (Docker Style)

Structure commands as `<noun> <verb>` -- the resource comes first, the action second.

```go
// Good -- noun first, verb second
app.Group("user").Add(
    &command.Command{Name: "create", Desc: "create a new user"},
    &command.Command{Name: "rm", Desc: "remove users"},
    &command.Command{Name: "ls", Desc: "list users"},
)

// Bad -- verb first, inconsistent
app.Add(
    &command.Command{Name: "create-user"},  // verb-noun smash
    &command.Command{Name: "list"},          // ambiguous noun
)
```

**Why**: Users learn the noun once, then discover verbs via `--help`. Docker proved this scales to 100+ commands.

### Rule 2: Positional Args for Primary Targets

The main operand of a command is a positional argument, not a flag. Flags modify behavior; positions identify targets.

```go
// Good -- target is positional
// Usage: myapp task show 42
cmd := &command.Command{
    Name:    "show",
    Desc:    "show task details",
    MinArgs: 1,
    MaxArgs: 1,
    Handler: func(ctx *command.Context) error {
        id := ctx.Args[0]
        // ...
    },
}

// Bad -- target hidden behind flag
// Usage: myapp task show --id 42
cmd := &command.Command{
    Name: "show",
    Options: []command.Option{
        {Name: "id", Required: true, Desc: "task ID"},
    },
}
```

**Decision guide** -- See [references/arguments.md](references/arguments.md)

### Rule 3: Flag Design -- Long First, Short for Frequent

Every flag starts as `--long-name`. Add `-x` short only for frequently-used flags.

```go
Options: []command.Option{
    // Frequent -- has short form
    {Name: "output", Short: "o", Desc: "output file path"},
    {Name: "verbose", Short: "V", IsBool: true, Desc: "verbose output"},
    {Name: "force", Short: "f", IsBool: true, Desc: "skip confirmation"},
    {Name: "quiet", Short: "q", IsBool: true, Desc: "suppress output"},

    // Infrequent -- long only
    {Name: "max-depth", Desc: "maximum recursion depth"},
    {Name: "include-archived", IsBool: true, Desc: "include archived items"},
}
```

**Standard short flags** (never reassign these) -- See [references/flags.md](references/flags.md)

### Rule 4: Keep Handlers Thin -- Delegate to Business Layer

Command handlers should be 50-150 lines: parse input, call business logic, format output. Extract complex logic to `internal/` packages.

```go
// Good -- thin handler with type-safe config
type DeployConfig struct {
    Version string `flag:"version" validate:"required"`
    Force   bool   `flag:"force"`
}

func deployHandler(ctx *command.Context) error {
    env := ctx.Args[0]

    var cfg DeployConfig
    if err := ctx.Bind(&cfg); err != nil {
        return err
    }

    schema := gozod.FromStruct[DeployConfig](gozod.WithTagName("validate"))
    if _, err := schema.Parse(cfg); err != nil {
        return err
    }

    deployer := service.NewDeployer(config)
    result, err := deployer.Deploy(ctx.Context(), env, cfg.Version, cfg.Force)
    if err != nil {
        return err
    }

    return command.Respond(ctx, result)
}

// Bad -- fat handler (200+ lines)
func deployHandler(ctx *command.Context) error {
    // Manual validation (should use Bind + gozod)
    version := ctx.String("version")
    if version == "" { ... }

    // API calls (should be in client)
    resp, err := http.Post(...)

    // Complex orchestration (should be in service)
    for _, step := range steps { ... }

    // Database queries (should be in repository)
    db.Exec("UPDATE ...")
}
```

**Three-layer pattern:**
```
cmd/           → CLI handlers (thin, 50-150 lines)
internal/      → Business logic (testable, reusable)
pkg/           → Core types (zero dependencies)
```

**When to extract:** Handler > 150 lines, repeated logic, need testing without CLI, planning API interface.

**Full guide** -- See [references/layering.md](references/layering.md)

### Rule 4.5: Use Optional[T] for Non-Required Fields

Use `command.Optional[T]` to distinguish "not set" from "zero value" in config structs.

```go
type ServerConfig struct {
    Host string `flag:"host" validate:"required"`

    // Optional fields - distinguish unset from zero
    Port    command.Optional[int]  `flag:"port" validate:"min=1000,max=9999"`
    Workers command.Optional[int]  `flag:"workers" validate:"min=1,max=100"`
    Debug   command.Optional[bool] `flag:"debug"`
}

Handler: func(ctx *command.Context) error {
    var cfg ServerConfig
    if err := ctx.Bind(&cfg); err != nil {
        return err
    }

    // Use OrDefault for optional fields
    port := cfg.Port.OrDefault(8080)
    workers := cfg.Workers.OrDefault(4)

    // Check if explicitly set
    if cfg.Debug.IsSet() {
        enableDebug(cfg.Debug.Value)
    }

    return startServer(cfg.Host, port, workers)
}
```

**Why Optional[T]?**

```go
// ❌ Without Optional - can't distinguish unset from zero
type Config struct {
    Port int `flag:"port"` // 0 means unset or explicitly set to 0?
}

// ✓ With Optional - explicit distinction
type Config struct {
    Port command.Optional[int] `flag:"port"`
}
cfg.Port.IsSet()           // false if not provided
cfg.Port.OrDefault(8080)   // 8080 if not provided
```

### Rule 5: Declare Constraints, Don't Code Them

Use go-command's declarative constraint system instead of manual validation.

```go
Options: []command.Option{
    // Enum -- auto-validated + auto-completed
    {Name: "format", Enum: []string{"json", "yaml", "csv"}, Default: "json"},

    // Exclusive group -- only one allowed
    {Name: "all", IsBool: true, Exclusive: "target"},
    {Name: "name", Exclusive: "target"},

    // Required together -- all or none
    {Name: "from", RequiredTogether: "migration"},
    {Name: "to", RequiredTogether: "migration"},

    // Dependency -- "component" requires "file" to be set
    {Name: "component", Requires: []string{"file"}},

    // Conflict -- cannot combine
    {Name: "watch", IsBool: true, Conflicts: []string{"check"}},
    {Name: "check", IsBool: true, Conflicts: []string{"watch"}},
}
```

### Rule 5: Preset Options -- Use Built-In Helpers

go-command provides preset options for common patterns. Use them instead of manual definitions.

```go
cmd := &command.Command{
    Name: "create",
    Desc: "create a new resource",
    Options: append([]command.Option{
        {Name: "name", Required: true, Desc: "resource name"},
        {Name: "tag", Desc: "resource tags (repeatable)"},
    },
        command.DryRunOption(),             // --dry-run / -n
    ),
    Handler: func(ctx *command.Context) error {
        if ctx.IsDryRun() {
            ctx.Props("Action", "create", "Name", ctx.String("name"))
            return nil
        }
        return doCreate(ctx)
    },
}

// Pagination presets
listCmd := &command.Command{
    Name: "list",
    Desc: "list resources",
    Options: append([]command.Option{
        {Name: "status", Enum: []string{"active", "archived"}, Default: "active"},
    },
        command.PaginationOptions()...,       // --limit, --offset
    ),
    Handler: func(ctx *command.Context) error {
        limit, offset := ctx.Pagination()
        items, total := listItems(limit, offset)
        return command.Respond(ctx, items, command.ResponseMeta{
            Total:   total,
            HasMore: offset+limit < total,
        })
    },
}
```

### Rule 6: Sensible Defaults -- Zero Config Works

Every option should have a meaningful default. `command.Config{}` zero value must produce a working app.

```go
// Good -- defaults make command usable without any flags
Options: []command.Option{
    {Name: "format", Default: "json", Desc: "output format"},
    {Name: "limit", Default: "50", Desc: "max results"},
    {Name: "timeout", Default: "30s", Desc: "request timeout"},
}

// Bad -- required flags force users to always specify
Options: []command.Option{
    {Name: "format", Required: true},  // should have default
    {Name: "limit", Required: true},   // should have default
}
```

### Rule 7: Structured Response with `Respond[T]`

Use `Respond[T]()` for typed output that auto-adapts to mode. Never manually branch on mode for output.

```go
type User struct {
    ID     int    `json:"id"`
    Name   string `json:"name"`
    Email  string `json:"email"`
    Role   string `json:"role"`
}

Handler: func(ctx *command.Context) error {
    user := getUser(ctx.Args[0])

    // Respond auto-adapts:
    // Human: ctx.Render(user) -> formatted output
    // Agent: {"data": {...}, "meta": {...}} JSON envelope
    return command.Respond(ctx, user, command.ResponseMeta{
        NextActions: []string{
            "myapp user update " + user.Name,
            "myapp user rm " + user.Name,
        },
    })
}
```

For simple key-value output, `ctx.Props()` and `ctx.Table()` still work:

```go
ctx.Props("Name", user.Name, "Role", user.Role, "Status", user.Status)

headers := []string{"Name", "Role"}
rows := []command.Row{{"alice", "admin"}, {"bob", "member"}}
ctx.Table(headers, rows)
```

### Rule 8: Interactive Safety -- Detect and Adapt

Check interactivity before prompting. Agents and pipes cannot respond to prompts.

```go
Handler: func(ctx *command.Context) error {
    if ctx.IsDryRun() {
        ctx.Props("Action", "delete", "Target", ctx.Args[0])
        return nil
    }

    // Only prompt in interactive TTY sessions
    if ctx.IsInteractive() {
        ok, err := ctx.Confirm("Delete user " + ctx.Args[0] + "?")
        if err != nil {
            return err
        }
        if !ok {
            ctx.Infof("Cancelled")
            return nil
        }
    }

    return deleteUser(ctx.Args[0])
}
```

`ctx.Confirm()` also respects a `--yes` option if the command registers one.

### Rule 9: Progress with Spinner

Use `ctx.Spinner()` for operations > 1 second. Mode-aware: animated in Human, silent in Agent, single line in CI. Outputs to stderr to keep stdout clean.

```go
Handler: func(ctx *command.Context) error {
    sp := ctx.Spinner("Deploying application")

    if err := buildApp(); err != nil {
        sp.Fail("Build failed")
        return err
    }
    sp.Update("Pushing to remote")

    if err := pushApp(); err != nil {
        sp.Fail("Push failed")
        return err
    }
    sp.Stop("Deployed successfully")

    return nil
}
```

### Rule 10: Group Commands and Options for Clarity

Use `Command.Group` and `Option.Group` to organize large CLIs. Groups appear as sections in help output.

```go
// Command grouping
app.Add(
    &command.Command{Name: "init", Group: "Project", Desc: "initialize project"},
    &command.Command{Name: "build", Group: "Project", Desc: "build project"},
    &command.Command{Name: "deploy", Group: "Project", Desc: "deploy project"},
    &command.Command{Name: "config", Group: "Configuration", Desc: "manage config"},
    &command.Command{Name: "secret", Group: "Configuration", Desc: "manage secrets"},
    &command.Command{Name: "version", Desc: "show version"},  // ungrouped
)

// Option grouping
cmd := &command.Command{
    Name: "deploy",
    Options: []command.Option{
        // Ungrouped flags appear under "Flags:"
        {Name: "force", Short: "f", IsBool: true, Desc: "force deployment"},

        // Grouped flags appear under "BUILD OPTIONS:", "OUTPUT OPTIONS:", etc.
        {Name: "target", Group: "Build", Desc: "build target"},
        {Name: "optimize", Group: "Build", IsBool: true, Desc: "optimize build"},
        {Name: "format", Group: "Output", Desc: "output format"},
        {Name: "color", Group: "Output", IsBool: true, Desc: "enable colors"},

        // Persistent flags always appear under "Global Flags:"
        {Name: "verbose", Short: "v", Persistent: true, Desc: "verbose output"},
    },
}
```

**Help output**:
```
PROJECT:
  init        initialize project
  build       build project
  deploy      deploy project

CONFIGURATION:
  config      manage config
  secret      manage secrets

COMMANDS:
  version     show version

Flags:
  -f, --force       force deployment

BUILD OPTIONS:
  --target string   build target
  --optimize        optimize build

OUTPUT OPTIONS:
  --format string   output format
  --color           enable colors

Global Flags:
  -v, --verbose     verbose output
```

**When to group**: 10+ commands or 8+ options. Groups preserve declaration order.

### Rule 11: Examples in Help -- Show, Don't Describe

Lead with practical examples. Users read examples before prose.

```go
cmd := &command.Command{
    Name: "deploy",
    Desc: "deploy application to environment",
    Long: "Deploy builds the application and pushes it to the specified environment.\n" +
        "Defaults to staging. Use --prod for production deployment.",
    Examples: []command.Example{
        {Desc: "deploy to staging", Line: "deploy myapp"},
        {Desc: "deploy to production", Line: "deploy myapp --prod"},
        {Desc: "dry run", Line: "deploy myapp --prod -n"},
    },
}
```

### Rule 12: Structured Errors with Guidance

Errors tell users what went wrong AND how to fix it.

```go
// Good -- actionable error with hint and next actions
return &command.Error{
    Code:      command.ExitError,
    Message:   "user 'alice' not found",
    Hint:      "run 'myapp user list' to see available users",
    ErrorCode: "ERR_NOT_FOUND",
    Details:   []string{"searched in local database and remote API"},
    NextActions: []string{
        "myapp user list",
        "myapp user create alice",
    },
}

// Good -- sentinel errors for programmatic matching
var ErrNotFound = errors.New("not found")
return fmt.Errorf("%w: user %q", ErrNotFound, name)

// Bad -- opaque error
return fmt.Errorf("error: something went wrong")
```

### Rule 13: Middleware for Cross-Cutting Concerns

Use middleware for auth, logging, config loading -- not handler boilerplate.

```go
// Global middleware -- applies to all commands
app.Use(
    recover.New(),
    logger.New(),
    mode.New(mode.Config{
        Aliases: map[string]command.Mode{"json": command.ModeAgent},
    }),
)

// Group middleware -- applies to group commands only
admin := app.Group("Admin")
admin.Use(requireAuthMiddleware)
admin.Add(userCmd, roleCmd, auditCmd)

// Command middleware -- single command
cmd := &command.Command{
    Name:       "push",
    Middleware: []command.MiddlewareFunc{requireAuthMiddleware},
}

// Built-in middleware helpers
command.Before(func(ctx *command.Context) error { return loadConfig(ctx) })
command.After(func(ctx *command.Context) error { return saveState(ctx) })
command.RequireFlags("token", "project")
command.Chain(mw1, mw2, mw3)  // combine into single middleware
```

### Rule 14: Suggest the Next Step

After every successful operation, tell users what to do next. With `Respond[T]`, next actions are included in the Agent JSON envelope automatically:

```go
return command.Respond(ctx, result, command.ResponseMeta{
    NextActions: []string{
        "myapp project show " + result.Name,
        "myapp task create",
    },
})
```

For human-only output, use `ctx.Successf()` + `ctx.Printf()` with suggested commands.

### Rule 15: Struct Binding -- Eliminate Boilerplate

Use `ctx.Bind()` to map options to struct fields. Reduces handler code by 60-80%.

```go
// Good -- declarative binding (15 lines)
type DeployParams struct {
    Env     string `flag:"env" desc:"Target environment"`
    DryRun  bool   `flag:"dry-run" desc:"Preview changes"`
    Workers int    `flag:"workers" default:"4" desc:"Worker count"`
}

Handler: func(ctx *command.Context) error {
    var params DeployParams
    if err := ctx.Bind(&params); err != nil {
        return err
    }

    // Validate required fields
    if params.Env == "" {
        return fmt.Errorf("--env is required")
    }

    return deploy(params.Env, params.Workers, params.DryRun)
}

// Bad -- manual extraction (70+ lines)
Handler: func(ctx *command.Context) error {
    env := ctx.String("env")
    if env == "" {
        return fmt.Errorf("--env required")
    }
    workers := 4
    if w := ctx.String("workers"); w != "" {
        parsed, err := strconv.Atoi(w)
        if err != nil {
            return fmt.Errorf("invalid --workers: %w", err)
        }
        workers = parsed
    }
    // ... repeated for each option
}
```

**Supported tags**: `flag`, `short`, `desc`, `default`, `env`, `persistent`

### Rule 16: Optional[T] -- Handle Optional Values

Use `Optional[T]` when you need to distinguish "not set" from "set to zero value". Common for ports, timeouts, and feature flags.

```go
type ServerParams struct {
    Host string              `flag:"host"`
    Port command.Optional[int] `flag:"port" default:"8080"`
}

Handler: func(ctx *command.Context) error {
    var params ServerParams
    if err := ctx.Bind(&params); err != nil {
        return err
    }

    // Validate required fields
    if params.Host == "" {
        return fmt.Errorf("--host is required")
    }

    // Use Optional[T] for truly optional values
    port := params.Port.OrDefault(8080)
    return startServer(params.Host, port)
}
```

**When to use**:
- Distinguish `--port 0` from not providing `--port`
- Optional config that changes behavior when explicitly set
- Need to detect if user provided a value

**Pattern**: Use plain types for required fields, validate in handler. Use `Optional[T]` only for truly optional values.

### Rule 17: Validation Integration

Integrate validation libraries via `command.Validator` interface. Example with gozod:

```go
// Define validator
type GozodValidator struct {
    tagName string
}

func (v *GozodValidator) Validate(target any) error {
    schema := gozod.FromStruct(target, gozod.WithTagName(v.tagName))
    _, err := schema.Parse(target)
    return err
}

// Register with app
app := command.New(command.Config{
    Name:      "myapp",
    Validator: &GozodValidator{tagName: "validate"},
})

// Use in handler
type CreateParams struct {
    Name  string              `flag:"name" validate:"min=3,max=50"`
    Port  command.Optional[int] `flag:"port" validate:"min=1000,max=9999"`
    Email string              `flag:"email" validate:"email"`
}

Handler: func(ctx *command.Context) error {
    var params CreateParams
    if err := ctx.Bind(&params); err != nil {
        return err
    }

    // Manual validation for required fields
    if params.Name == "" {
        return fmt.Errorf("--name is required")
    }

    // Optional: use validator for complex rules
    if err := ctx.Validate(&params); err != nil {
        return err
    }

    return createResource(params)
}
```

**Pattern**: Bind → Validate → Execute. Validation is optional; only call `ctx.Validate()` if you registered a validator.

See `examples/validate/` for complete gozod integration.

## Command Patterns

| Pattern | Structure | Example |
|---------|-----------|---------|
| Resource CRUD | `<noun> list/show/create/rm` | `user list`, `user show alice`, `user rm alice` |
| Sync Ops | `pull/push/diff <target>` | `push myapp --all`, `diff myapp` |
| Pipeline | `check/build` with `--fix`/`--watch` | `check --fix`, `build --watch` |
| Shortcuts | Top-level aliases to deep commands | `status` -> `project status` |

CRUD commands: `list` gets `PaginationOptions()`, write commands get `DryRunOption()`, use `Aliases: []string{"ls"}` for common shortcuts.

## App Setup

```go
app := command.New(command.Config{
    Name:      "myapp",
    Version:   "1.0.0",
    EnvPrefix: "MYAPP",  // --port -> MYAPP_PORT auto-mapping
    Options:   []command.Option{{Name: "json", IsBool: true, Desc: "JSON output"}},
})

app.Use(
    recover.New(),
    mode.New(mode.Config{Aliases: map[string]command.Mode{"json": command.ModeAgent}}),
    query.New(),
)
app.AddOptions(query.Options()...)

app.Group("user").Add(UserCommands()...)
app.Add(statusCmd)
app.Run(os.Args[1:])
```

### Struct-Based Commands with Mount

```go
type DeployCmd struct {
    Env    string `flag:"env" short:"e" default:"staging" desc:"target environment"`
    DryRun bool   `flag:"dry-run" short:"n" desc:"preview changes"`
}
func (d *DeployCmd) Run(ctx *command.Context) error { return nil }

command.Mount(app, &DeployCmd{})  // auto-registers "deploy" with struct tag options
```

## Designing for Agents

CLI is the primary interface for AI agents -- they use `--help` to learn, `--json` to parse, and exit codes to branch. See [references/agent.md](references/agent.md) for the full 12-rule guide.

### Quick Checklist

| Rule | go-command Implementation |
|------|-----------------------------|
| Structured JSON output | `Respond[T]()` / `ctx.Props()` / `ctx.Table()` auto-adapt per Mode |
| Token-efficient defaults | `VerbosityMinimal` auto-set in Agent mode |
| No interactive prompts | `ctx.IsInteractive()` check before prompting |
| Semantic exit codes | `command.ExitSuccess` / `ExitError` / `ExitUsageError` |
| `--dry-run` on writes | `command.DryRunOption()` + `ctx.IsDryRun()` |
| Self-documenting schema | `app.Schema()` -> MCP tool definitions |
| Next-actions guidance | `ResponseMeta.NextActions` / `Error.NextActions` |
| Safety guardrails | `ctx.Confirm()` for destructive ops; `--yes` flag |
| Idempotent operations | Check existence before create; skip if identical |
| Stable output contract | Never rename JSON fields post-release |
| Composability | stdout = data, stderr = status; `ctx.HasPipedInput()` |
| Batch operations | Accept variadic args or file input |
| Query filtering | `middleware/query` with jq expressions via `--query` |

### Agent Mode Detection

```go
// go-command auto-detects via mode middleware:
// 1. --mode flag (explicit)
// 2. --json alias -> Agent
// 3. COMMAND_MODE=agent env var
// 4. CI env detection (CI, GITHUB_ACTIONS, GITLAB_CI, JENKINS_URL, CIRCLECI)
// 5. Non-TTY -> Agent
app.Use(mode.New(mode.Config{
    Aliases: map[string]command.Mode{"json": command.ModeAgent},
}))

// For agent-native CLIs (default Agent, humans opt in with --human):
app.Use(mode.New(mode.AgentNative()))
```

## Anti-Patterns

| Anti-Pattern | Problem | Fix |
|-------------|---------|-----|
| `--name <val>` for primary target | Verbose, inconsistent with git/docker | Positional arg |
| `cmd list --type X` for filtering | Conflates filtering with subcommand | `cmd X list` or `--type X` flag |
| Silent success | User wonders if anything happened | Print confirmation + next steps |
| Manual mode branching | `if ctx.Mode() == "agent"` in handler | Use `Respond[T]()` / `ctx.Props()` |
| Catch-all subcommands | Blocks future command names | Explicit names only |
| Manual `--dry-run` option | Inconsistent names, missing short `-n` | `command.DryRunOption()` |
| Manual `--limit`/`--offset` | Inconsistent defaults | `command.PaginationOptions()` |
| Hardcoded exit in handler | Breaks testing with `Exec()` | Return `*command.Error` with code |
| Required flag with obvious default | Forces users to always specify | Set `Default` value |
| `fmt.Print("Confirm? y/n:")` | Breaks Agent/CI/pipe mode | `ctx.IsInteractive()` + `ctx.Confirm()` |
| `fmt.Fprintf(os.Stdout, ...)` | Breaks `Exec()` test capture | Use `ctx.Printf()` / `ctx.Println()` |

## References

### references/

| File | Contents |
|------|----------|
| [arguments.md](references/arguments.md) | When to use positional args vs flags -- decision tree and examples |
| [flags.md](references/flags.md) | Standard flag names, short flags, presets, constraints, resolution priority |
| [layering.md](references/layering.md) | Three-layer architecture -- thin handlers, business logic, core types |
| [output.md](references/output.md) | Respond[T], Spinner, Pager, JSONL, mode-aware formatting, color |
| [errors.md](references/errors.md) | Structured errors, exit codes, NextActions, multi-error collection |
| [help.md](references/help.md) | Help text design, examples, Option.Group, completion |
| [testing.md](references/testing.md) | CLI testing with Exec/ExecWith, testutil, assertion patterns |
| [agent.md](references/agent.md) | Agent-friendly CLI -- 12 rules for structured output, token efficiency, safety, self-documenting |

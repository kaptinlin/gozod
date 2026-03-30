# Help Text & Completion Design

Help text structure, examples-first approach, progressive disclosure, Option.Group, and shell completion.

## Help Text Principles

### 1. Examples First

Users read examples before prose. Lead with the most common use case.

```go
cmd := &command.Command{
    Name: "deploy",
    Desc: "deploy project to environment",
    Long: "Deploy builds the application and pushes it to the specified environment.\n" +
        "Defaults to staging. Use --prod for production deployment.",
    Examples: []command.Example{
        {Desc: "deploy to staging", Line: "deploy myapp"},
        {Desc: "deploy to production", Line: "deploy myapp --prod"},
        {Desc: "dry run", Line: "deploy myapp --prod -n"},
    },
}
```

### 2. Desc Is a Scannable Summary

`Desc` appears in parent command's help listing. Keep it under 60 characters, lowercase, no period.

```go
// Good — scannable in list
{Name: "list", Desc: "list all users"}
{Name: "show", Desc: "show user details"}
{Name: "check", Desc: "validate project configuration"}

// Bad — too long, sentence-like
{Name: "list", Desc: "This command lists all of the users that are registered in the system."}
```

### 3. Long Is for Detail

`Long` appears when `--help` is called on this specific command. Include:
- What the command does (1-2 sentences)
- Important behavior notes
- Links to documentation if relevant

```go
Long: "Pull downloads remote data and writes it to the local store.\n" +
    "Existing files are overwritten — use git to track changes.\n\n" +
    "Requires an API token (set via --token or MYAPP_TOKEN env var).",
```

### 4. Group Related Commands

```go
// Commands appear grouped in help output
resources := app.Group("Resources")
resources.Add(userCmd, projectCmd, taskCmd)

quality := app.Group("Quality")
quality.Add(lintCmd, checkCmd)

infra := app.Group("Infrastructure")
infra.Add(deployCmd, statusCmd, logsCmd)
```

### 5. Group Related Options

Use `Option.Group` to visually organize complex commands.

```go
cmd := &command.Command{
    Name: "list",
    Desc: "list tasks",
    Options: append([]command.Option{
        // Filter group
        {Name: "status", Enum: []string{"open", "closed", "all"}, Default: "open", Group: "Filter"},
        {Name: "owner", Desc: "filter by owner", Group: "Filter"},
        {Name: "label", Desc: "filter by label", Group: "Filter"},

        // Output group
        {Name: "format", Enum: []string{"json", "yaml", "table"}, Default: "table", Group: "Output"},
        {Name: "fields", Desc: "comma-separated fields", Group: "Output"},
        {Name: "no-header", IsBool: true, Desc: "omit table header", Group: "Output"},
    },
        command.PaginationOptions()...,  // --limit, --offset (ungrouped)
    ),
}
```

**Help output:**
```
Filter:
  --status    task status (open, closed, all) [default: open]
  --owner     filter by owner
  --label     filter by label

Output:
  --format    output format (json, yaml, table) [default: table]
  --fields    comma-separated fields
  --no-header omit table header

Options:
  --limit     max results [default: 20]
  --offset    results offset [default: 0]
```

### 6. Hide Internal Commands

```go
// Hidden from help but still executable
{Name: "debug-cache", Hidden: true, Handler: debugCacheHandler}
{Name: "schema", Hidden: true, Handler: schemaHandler}
```

### 7. Deprecation Messages

```go
{Name: "old-command", Deprecated: "use 'new-command' instead", Handler: oldHandler}
// Shows: [DEPRECATED: use 'new-command' instead]
```

## Shell Completion

### Enum Auto-Completion

Enum options automatically get completion support:

```go
{Name: "format", Enum: []string{"json", "yaml", "csv", "table"}}
{Name: "status", Enum: []string{"open", "closed", "in-progress"}}
{Name: "role", Enum: []string{"admin", "member", "viewer"}}
```

### Dynamic Completion

For values that come from runtime data, use `CompletionHandler`:

```go
{
    Name: "project",
    CompletionHandler: func(ctx *command.Context) []string {
        _, cur := ctx.CurWords()  // prev word, current partial word
        projects, _ := loadProjectNames()
        return filterPrefix(projects, cur)
    },
}
```

### Command-Level Completion

```go
cmd := &command.Command{
    Name:    "show",
    MinArgs: 1,
    CompletionHandler: func(ctx *command.Context) []string {
        _, cur := ctx.CurWords()
        names, _ := loadUserNames()
        return filterPrefix(names, cur)
    },
}
```

### Built-In Completion Command

go-command auto-injects a `completion` subcommand (disable with `Config.HideCompletion`):

```bash
# Generate shell completion script
myapp completion bash >> ~/.bashrc
myapp completion zsh >> ~/.zshrc
myapp completion fish >> ~/.config/fish/completions/myapp.fish
```

Completion mode is triggered by setting `COMPLETION_MODE=1` env var. Check with `ctx.IsCompletionMode()`.

### Default Completion Handler

The default handler completes subcommand names and flag names (prefixed with `-`). Override per-command:

```go
cmd := &command.Command{
    Name: "show",
    CompletionHandler: func(ctx *command.Context) []string {
        // ctx.CurWords() returns (previousWord, currentPartialWord)
        return customCompletions(ctx)
    },
}
```

## Help Design Anti-Patterns

| Anti-Pattern | Fix |
|-------------|-----|
| No examples in help | Add 2-3 `Examples` entries to every command |
| Desc ends with period | Remove period — it's a label, not a sentence |
| Desc starts with "This command..." | Start with verb: "list", "show", "deploy" |
| All flags visible | Use `Hidden: true` for debug/internal flags |
| No Long description | Add Long for commands with non-obvious behavior |
| Uppercase Desc | Lowercase, like a label |
| Missing completion for enum | Use `Enum` field — go-command auto-completes |
| Dynamic values without completion | Add `CompletionHandler` for runtime data |
| Too many ungrouped options | Use `Option.Group` for visual clarity |
| No option groups in complex commands | Group by function: Filter, Output, Auth |

## Schema Export for Agents

`app.Schema()` exports the full command tree as machine-readable JSON:

```go
schema := app.Schema()
// Returns CommandSchema with recursive Commands, Options, Examples
```

Agents use this to discover available commands without reading help text. MCP servers can auto-generate tool definitions from the schema.

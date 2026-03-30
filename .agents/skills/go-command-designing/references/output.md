# Output Design

Mode-aware output, Respond[T], Spinner, Pager, JSONL, color, and stdout/stderr conventions.

## Core Principle: Human First, Machine Ready

Default output is for humans. Machine output is opt-in via `--json` or `--mode agent`.

```
stdout → primary data (tables, results, generated content)
stderr → status messages, progress, warnings, errors, spinners
```

## Respond[T] — Structured Response

The primary way to return typed data from handlers. Auto-adapts to mode.

```go
type Project struct {
    ID     int    `json:"id"`
    Name   string `json:"name"`
    Status string `json:"status"`
    Owner  string `json:"owner"`
}

Handler: func(ctx *command.Context) error {
    project := getProject(ctx.Args[0])

    return command.Respond(ctx, project, command.ResponseMeta{
        NextActions: []string{
            "myapp project deploy " + project.Name,
            "myapp task list --project " + project.Name,
        },
    })
}
```

| Mode | Output |
|------|--------|
| Human | `ctx.Render(project)` — formatted by app Formatter |
| Agent | `{"data": {"id": 1, "name": "...", ...}, "meta": {"next_actions": [...]}}` |

### ResponseMeta

Optional metadata included in the Agent JSON envelope.

```go
command.ResponseMeta{
    Total:       100,                    // total count for pagination
    HasMore:     true,                   // whether more results exist
    Cursor:      "abc123",               // cursor for cursor-based pagination
    Warnings:    []string{"deprecated"}, // warning messages
    NextActions: []string{"myapp ..."},  // follow-up command suggestions
    Extra:       map[string]string{...}, // extra metadata
}
```

### List with Pagination

```go
Handler: func(ctx *command.Context) error {
    limit, offset := ctx.Pagination()
    items, total := db.ListProjects(limit, offset)

    return command.Respond(ctx, items, command.ResponseMeta{
        Total:   total,
        HasMore: offset+limit < total,
    })
}
```

### Cursor-Based Pagination

```go
Handler: func(ctx *command.Context) error {
    limit, after := ctx.CursorPagination()
    items, nextCursor := db.ListEvents(limit, after)

    return command.Respond(ctx, items, command.ResponseMeta{
        HasMore: nextCursor != "",
        Cursor:  nextCursor,
    })
}
```

## Mode-Aware Output with Props and Table

For simple key-value or tabular output without a custom type, use `ctx.Props()` and `ctx.Table()`.

### Props — Key-Value Pairs

```go
// Props takes variadic string pairs: key1, val1, key2, val2, ...
ctx.Props(
    "Name", project.Name,
    "Status", project.Status,
    "Owner", project.Owner,
    "Tasks", strconv.Itoa(len(project.Tasks)),
)
```

| Mode | Output |
|------|--------|
| Human | `Name:          My Project` (aligned) |
| Agent | `{"name":"My Project","status":"active",...}` |
| CI | `name=My Project\nstatus=active\n...` |

### Table -- Tabular Data

```go
headers := []string{"Name", "Status", "Owner"}
rows := []command.Row{  // Row is []string
    {"Project A", "active", "alice"},
    {"Project B", "archived", "bob"},
}
ctx.Table(headers, rows)
```

| Mode | Output |
|------|--------|
| Human | Aligned columns with Unicode borders |
| Agent | `[{"name":"Project A","status":"active","owner":"alice"},...]` |
| CI | Tab-separated values |

### Semantic Methods -- Status Communication

```go
ctx.Successf("Created project %s", name)   // green, to stdout
ctx.Warnf("Token %s is deprecated", token) // yellow, to stderr
ctx.Infof("Using config from %s", path)    // cyan, to stdout
ctx.Failf("Project %s not found", name)    // red, to stderr
```

Color is auto-disabled when: Mode is not Human, `NO_COLOR` env is set, or output is not a TTY.

## Spinner — Mode-Aware Progress

Use `ctx.Spinner()` for operations > 1 second. Mode-aware: animated in Human, silent in Agent, single line in CI.

```go
Handler: func(ctx *command.Context) error {
    sp := ctx.Spinner("Fetching data")

    items, err := fetchItems()
    if err != nil {
        sp.Fail("Fetch failed")
        return err
    }
    sp.Update("Processing items")

    result, err := processItems(items)
    if err != nil {
        sp.Fail("Processing failed")
        return err
    }
    sp.Stop("Processed " + strconv.Itoa(len(items)) + " items")

    return command.Respond(ctx, result)
}
```

| Mode | Behavior |
|------|----------|
| Human | Animated braille spinner (`⠋⠙⠹...`) on stderr |
| Agent | Silent (no output) |
| CI | Prints message once to stderr |

**Methods:**

| Method | Purpose |
|--------|---------|
| `sp.Update(msg)` | Change spinner message |
| `sp.Stop(msg)` | Stop with success (✓) |
| `sp.Fail(msg)` | Stop with failure (✗) |

## Pager — Long Output

Use `ctx.Pager()` for output that may exceed the terminal height. Auto-detects TTY.

```go
Handler: func(ctx *command.Context) error {
    w := ctx.Pager()
    defer w.Close()  // MUST close to flush

    for _, item := range allItems {
        fmt.Fprintf(w, "%s\t%s\t%s\n", item.Name, item.Status, item.Owner)
    }
    return nil
}
```

| Mode | Behavior |
|------|----------|
| Human + TTY | Pipes through `$PAGER` (default: `less -FIRX`) |
| Agent/CI/Pipe | Writes directly to stdout (passthrough) |

## JSONL -- Streaming JSON Lines

Use `ctx.JSONL()` for streaming large datasets with O(1) memory. Takes `iter.Seq[any]`.

```go
Handler: func(ctx *command.Context) error {
    return ctx.JSONL(func(yield func(any) bool) {
        for item := range db.ScanAll() {
            if !yield(item) {
                return
            }
        }
    })
}
```

Each item is independently marshaled as a single JSON line. Same output in all modes.

## Render — Custom Formatter

```go
// Uses app-level or command-level Formatter
ctx.Render(data)

// Override formatter per command
ctx.SetFormatter(&customFormatter{})
ctx.Render(data)
```

## Verbosity Levels

```go
verbosity := ctx.Verbosity()

switch verbosity {
case command.VerbosityMinimal:
    // Agent default — summary only
    ctx.Printf("3 errors, 2 warnings")

case command.VerbosityCompact:
    // Human default — summary + details
    ctx.Printf("3 errors, 2 warnings\n")
    for _, e := range errors {
        ctx.Failf("%s:%d %s", e.File, e.Line, e.Message)
    }

case command.VerbosityDetailed:
    // --verbose — everything + fix suggestions
    // ... full output with hints
}
```

## Color Conventions

### When to Use Color

- Status: green for success, red for error, yellow for warning
- Emphasis: bold for important values, dim for metadata
- Structure: color to separate sections in dense output

### When to Disable Color

go-command handles this automatically when:
- stdout is not a TTY (piped output)
- `NO_COLOR` environment variable is set (any value)
- `TERM=dumb`
- `--no-color` flag passed
- Mode is Agent or CI

### Color Guidelines

- Maximum 3-4 colors in a single output block
- Never convey information solely through color (accessibility)
- Red/yellow reserved for errors/warnings — don't use for decoration
- Use dim/gray for secondary information

## Output Anti-Patterns

| Anti-Pattern | Fix |
|-------------|-----|
| Silent success | Print what happened + next steps |
| Stack traces to stdout | Errors to stderr; stack traces only with `--debug` |
| JSON mixed with status messages | JSON to stdout, status to stderr |
| Mode branching in handler | Use `Respond[T]()` / `ctx.Props()` / `ctx.Table()` |
| Color in piped output | go-command auto-disables; respect `NO_COLOR` |
| Logging to stdout | All logs/status to stderr |
| Unbounded output without pager | Use `ctx.Pager()` for > 50 lines in TTY mode |
| Progress on stdout | `ctx.Spinner()` always outputs to stderr |
| Manual JSON serialization | Use `Respond[T]()` for consistent envelope |
| Manual progress messages | Use `ctx.Spinner()` for mode-aware progress |

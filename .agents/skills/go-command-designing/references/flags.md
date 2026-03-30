# Flag Design Guide

Standard flag names, short forms, presets, naming conventions, and POSIX/GNU compatibility.

## Standard Short Flags

These have universal meaning across CLI tools. Never reassign them to different purposes.

| Short | Long | Meaning | Convention Source |
|-------|------|---------|-------------------|
| `-h` | `--help` | Show help | POSIX / Universal |
| `-v` | `--version` | Show version | GNU |
| `-V` | `--verbose` | Verbose output | Common (some tools use `-v`) |
| `-q` | `--quiet` | Suppress output | GNU / POSIX |
| `-f` | `--force` | Force operation | Unix (`rm -f`) |
| `-n` | `--dry-run` | Preview without executing | `make -n`, `git clean -n` |
| `-o` | `--output` | Output file/path | GCC / Common |
| `-r` | `--recursive` | Recurse into directories | `ls -r`, `cp -r` |
| `-i` | `--interactive` | Prompt before action | `rm -i` |
| `-a` | `--all` | Include everything | `ls -a`, `docker ps -a` |
| `-d` | `--debug` | Debug output | Common |
| `-p` | `--port` | Network port | Common |
| `-c` | `--config` | Config file path | Common |
| `-w` | `--watch` | Watch for changes | Common |

### go-command Auto-Registered

go-command automatically registers these; don't redefine them:

- `--help` / `-h` — Auto-injected on all commands
- `--version` — Auto-injected on root command (if `Config.Version` set)
- `--mode` — Registered by mode middleware
- `--debug-options` — Hidden, auto-injected

## Preset Options

go-command provides preset options for common patterns. Use them for consistency.

### DryRunOption

```go
// Instead of manually defining:
// {Name: "dry-run", Short: "n", IsBool: true, Desc: "preview changes without executing"}

// Use the preset:
Options: []command.Option{
    command.DryRunOption(),  // --dry-run / -n
}

// Check in handler:
if ctx.IsDryRun() {
    // preview mode
}
```

### PaginationOptions (Offset-Based)

```go
// Adds --limit (default: 20) and --offset (default: 0)
Options: append(myOptions, command.PaginationOptions()...)

// Read in handler:
limit, offset := ctx.Pagination()
items := db.List(limit, offset)
```

### CursorPaginationOptions

```go
// Adds --limit (default: 20) and --after (cursor)
Options: append(myOptions, command.CursorPaginationOptions()...)

// Read in handler:
limit, after := ctx.CursorPagination()
items, nextCursor := db.ListAfter(limit, after)
```

## Option Groups

Group related options visually in help output using `Option.Group`.

```go
Options: []command.Option{
    // Filter group
    {Name: "status", Enum: []string{"active", "archived"}, Group: "Filter"},
    {Name: "role", Desc: "filter by role", Group: "Filter"},
    {Name: "since", Desc: "filter by creation date", Group: "Filter"},

    // Output group
    {Name: "format", Enum: []string{"json", "yaml", "table"}, Default: "table", Group: "Output"},
    {Name: "fields", Desc: "comma-separated fields to include", Group: "Output"},
    {Name: "no-header", IsBool: true, Desc: "omit table header", Group: "Output"},
}
```

**Help output:**
```
Filter:
  --status    filter status (active, archived)
  --role      filter by role
  --since     filter by creation date

Output:
  --format    output format (json, yaml, table) [default: table]
  --fields    comma-separated fields to include
  --no-header omit table header
```

## EnvPrefix — Auto-Map Environment Variables

Set `Config.EnvPrefix` to auto-map all options to environment variables.

```go
app := command.New(command.Config{
    Name:      "myapp",
    EnvPrefix: "MYAPP",  // --port → MYAPP_PORT, --db-host → MYAPP_DB_HOST
})
```

With `EnvPrefix: "MYAPP"`:
- `--port` checks `MYAPP_PORT`
- `--db-host` checks `MYAPP_DB_HOST`
- `--verbose` checks `MYAPP_VERBOSE`

You can still override individual options with explicit `Env`:

```go
{Name: "token", Env: "API_TOKEN"}  // explicit Env wins over auto-mapped MYAPP_TOKEN
```

## Flag Naming Rules

### Use Lowercase Kebab-Case

```go
// Good
{Name: "dry-run"}
{Name: "max-depth"}
{Name: "output-format"}
{Name: "include-archived"}

// Bad
{Name: "dryRun"}       // camelCase
{Name: "DryRun"}       // PascalCase
{Name: "dry_run"}      // snake_case
{Name: "DRYRUN"}       // SCREAMING
```

### Match Established Conventions

When a flag has an established name in other tools, use it:

```go
// Good — matches git/docker conventions
{Name: "force", Short: "f"}           // not "yes" or "confirm"
{Name: "recursive", Short: "r"}       // not "recurse" or "deep"
{Name: "verbose", Short: "V"}         // not "debug-level"
{Name: "quiet", Short: "q"}           // not "silent" or "no-output"
{Name: "all", Short: "a"}             // not "include-all" or "everything"
```

### Boolean Flags Get Automatic Negation

go-command automatically supports `--no-<flag>` for boolean options.

```go
{Name: "color", IsBool: true, Default: "true"}
// User can: --color (enable) or --no-color (disable)

{Name: "cache", IsBool: true, Default: "true"}
// User can: --cache or --no-cache
```

### Env Var Naming

Environment variable names follow the pattern: `PREFIX_FLAG_NAME` (uppercase, underscores).

```go
// Manual (when EnvPrefix is not set or override needed)
{Name: "config", Env: "MYAPP_CONFIG", Desc: "config file path"}
{Name: "token", Env: "API_TOKEN", Desc: "API token"}

// Automatic (when Config.EnvPrefix is set)
// --config → MYAPP_CONFIG (auto-derived)
```

## Flag Categories

### Persistent Flags (Inherited by Subcommands)

```go
// Define on app or parent command
app.AddOptions(
    command.Option{Name: "mode", Persistent: true, Desc: "output mode"},
    command.Option{Name: "config", Short: "c", Persistent: true},
)
```

**Rule**: Only make a flag persistent if it applies to ALL subcommands. `--format` is NOT persistent because it means different things in different contexts.

### Hidden Flags (Power-User / Internal)

```go
{Name: "debug-sql", Hidden: true, IsBool: true, Desc: "log SQL queries"}
{Name: "profile", Hidden: true, IsBool: true, Desc: "enable CPU profiling"}
```

### Deprecated Flags (Migration Path)

```go
{Name: "old-format", Deprecated: "use --format instead", IsBool: true}
```

go-command prints deprecation warning when used.

## Constraint Patterns

### Mutually Exclusive (Choose One)

```go
// Only one of --json, --yaml, --table allowed
{Name: "json", IsBool: true, Exclusive: "format"},
{Name: "yaml", IsBool: true, Exclusive: "format"},
{Name: "table", IsBool: true, Exclusive: "format"},
```

Simpler alternative — use `Enum`:

```go
{Name: "format", Enum: []string{"json", "yaml", "table"}, Default: "table"}
```

### Required Together (All or None)

```go
{Name: "from", RequiredTogether: "migration"},
{Name: "to", RequiredTogether: "migration"},
// --from without --to → error
```

### Dependency (If A then B Required)

```go
{Name: "component", Requires: []string{"project"}}
// --component without --project → error
```

### Conflicts (Cannot Combine)

```go
{Name: "watch", IsBool: true, Conflicts: []string{"check"}},
{Name: "check", IsBool: true, Conflicts: []string{"watch"}},
// --watch --check → error
```

## Resolution Priority

go-command resolves option values in this order (first match wins):

```
1. CLI flag         --name=value                   (highest priority)
2. Env var (explicit) Opt.Env field
3. Env var (auto)     EnvPrefix + UPPER_SNAKE(name)
4. Config source    SetSources() lookup
5. Default value    Default: "value"
6. Default function DefaultFn: func() (string, error)
7. Empty string     ""                             (lowest priority)
```

Use `ctx.OptionSource("name")` to check where a value came from:

```go
source := ctx.OptionSource("config")
switch source {
case command.ValueSourceFlag:      // user explicitly set via CLI flag
case command.ValueSourceEnv:       // from environment variable
case command.ValueSourceConfig:    // from config source (SetSources)
case command.ValueSourceDefault:   // from Option.Default field
case command.ValueSourceDefaultFn: // from Option.DefaultFn function
case command.ValueSourceNone:      // not set at all
}
```

## Config Sources

go-command supports external config sources (TOML, YAML, JSON files) via the `ConfigSource` interface:

```go
// Register config sources (consulted during option resolution)
app.SetSources([]command.ConfigSource{
    command.NestedMapSource{Data: tomlData},  // dot-notation: "db.host"
    command.MapSource{Data: flatMap},          // flat key-value
})
```

Custom config sources implement a single method:

```go
type ConfigSource interface {
    Get(key string) (string, bool)
}
```

### DefaultFn -- Dynamic Defaults

Use `Option.DefaultFn` for defaults that require computation or I/O:

```go
{Name: "token-dir", DefaultFn: func() (string, error) {
    home, err := os.UserHomeDir()
    if err != nil {
        return "", err
    }
    return filepath.Join(home, ".myapp", "tokens"), nil
}}
```

`DefaultFn` is checked after `Default` in the resolution chain.

## Common Mistakes

| Mistake | Fix |
|---------|-----|
| `-v` for both version and verbose | `-v` version, `-V` verbose (or vice versa, pick one convention) |
| Same short flag in parent and child | go-command detects this — each short must be unique in scope |
| Required flag with obvious default | Set `Default` and drop `Required` |
| Flag for the primary target | Use positional arg instead |
| `--format` as persistent | Different commands have different format needs |
| Manual `--dry-run` definition | Use `command.DryRunOption()` preset |
| Manual `--limit`/`--offset` | Use `command.PaginationOptions()` preset |
| Ungrouped options in complex commands | Use `Option.Group` for visual clarity |
| Hardcoded env var names | Use `Config.EnvPrefix` for auto-mapping |

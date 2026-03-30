# Positional Arguments vs Flags

When to use positional args, when to use flags, and how to combine them.

## Decision Tree

```
Is this the primary operand of the command?
│
├─ Yes → Is there only one primary operand?
│  │
│  ├─ Yes → Positional argument
│  │  Examples: git checkout <branch>, rm <file>, docker run <image>
│  │
│  └─ No → Are they the same type (variadic)?
│     │
│     ├─ Yes → Positional variadic
│     │  Examples: rm file1 file2 file3, git add *.go
│     │
│     └─ No → Are they ordered and well-known?
│        │
│        ├─ Yes → Positional with order
│        │  Examples: cp <source> <dest>, mv <from> <to>
│        │
│        └─ No → Use flags for clarity
│           Example: deploy --from staging --to production
│
└─ No → Is it modifying behavior?
   │
   ├─ Yes → Flag (--long-name / -x)
   │  Examples: --verbose, --format json, --dry-run
   │
   └─ Is it a filter/selector on the primary action?
      │
      ├─ Yes → Flag
      │  Examples: ls --all, find --type f, log --since 2024-01-01
      │
      └─ Is it a configuration value?
         └─ Flag with env var fallback
            Example: --config path, --timeout 30s
```

## Rules

### 1. Primary Target = Positional

The thing the command acts on is always positional.

```bash
# Good — target is clear
git clone <url>
docker run <image>
kubectl get <resource>
myapp user show <name>
myapp task assign <task-id>

# Bad — target buried in flags
git clone --url <url>
myapp user show --name <name>
myapp task assign --id <task-id>
```

### 2. Multiple Same-Type Targets = Variadic

When acting on multiple targets of the same type, accept variadic positional args.

```go
cmd := &command.Command{
    Name:    "rm",
    MinArgs: 1,  // at least one
    // MaxArgs not set = unlimited
    Handler: func(ctx *command.Context) error {
        for _, name := range ctx.Args {
            remove(name)
        }
        return nil
    },
}
```

### 3. Two Different Targets = Ordered Positional (If Intuitive)

Only when the order is universally understood (like `cp source dest`).

```go
// Good — intuitive order (Unix convention)
cmd := &command.Command{
    Name:    "copy",
    MinArgs: 2,
    MaxArgs: 2,
    Handler: func(ctx *command.Context) error {
        source, dest := ctx.Args[0], ctx.Args[1]
        return copyFile(source, dest)
    },
}
```

If the order is ambiguous, use flags instead:

```go
// Good — flags for non-obvious order
cmd := &command.Command{
    Name: "migrate",
    Options: []command.Option{
        {Name: "from", Required: true, RequiredTogether: "migration"},
        {Name: "to", Required: true, RequiredTogether: "migration"},
    },
}
```

### 4. Everything Else = Flags

Modifiers, filters, configuration, format selection — all flags.

```go
Options: []command.Option{
    // Behavior modifiers
    {Name: "recursive", Short: "r", IsBool: true},
    {Name: "force", Short: "f", IsBool: true},

    // Filters
    {Name: "role", Desc: "filter by role"},
    {Name: "status", Enum: []string{"active", "inactive", "pending"}},

    // Output control
    {Name: "format", Enum: []string{"json", "yaml", "table"}, Default: "table"},
    {Name: "limit", Default: "50"},
}
```

### 5. Flags Are Position-Independent

Users should never need to remember flag order.

```bash
# These must all be equivalent
myapp user list --role admin --format json
myapp user list --format json --role admin
myapp user --role admin list --format json  # flags before verb also works
```

### 6. `--` Stops Flag Parsing

Everything after `--` becomes a positional argument, even if it starts with `-`.

```bash
# Pass "--force" as a literal argument, not a flag
myapp exec -- --force
```

## Combining Positional Args and Flags

The ideal command has 0-2 positional args and flags for everything else.

```bash
# Perfect — 1 positional + flags
myapp project deploy myapp -n --all

# Acceptable — 2 ordered positionals (cp convention)
myapp copy source.md dest.md --overwrite

# Avoid — 3+ positionals (confusing)
myapp deploy staging v2.1.0 us-east-1  # what's what?
```

### go-command Implementation

```go
cmd := &command.Command{
    Name:    "deploy",
    Desc:    "deploy project to environment",
    MinArgs: 1,
    MaxArgs: 1,
    Options: []command.Option{
        command.DryRunOption(),
        {Name: "all", IsBool: true, Exclusive: "scope", Desc: "deploy all services"},
        {Name: "env", Enum: []string{"staging", "production"}, Default: "staging"},
    },
    Handler: func(ctx *command.Context) error {
        project := ctx.Args[0]
        if ctx.IsDryRun() {
            ctx.Props("Action", "deploy", "Project", project, "Env", ctx.String("env"))
            return nil
        }
        return doDeploy(project, ctx.String("env"))
    },
}
```

## Exceptions

| Scenario | Use Flag Instead | Why |
|----------|-----------------|-----|
| Secrets/passwords | `--token-file` or stdin | Positional args visible in `ps` output |
| Complex filters | `--since 2024-01-01` | Not a target, it's a modifier |
| Optional targets with defaults | `--config path` | Defaults make flag optional |
| Machine-generated input | `--input file.json` | Machines benefit from explicit naming |

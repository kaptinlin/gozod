# Error Design

Structured errors, exit codes, recovery guidance, NextActions, and multi-error collection.

## Exit Code Convention

| Code | Meaning | When |
|------|---------|------|
| `0` | Success | Operation completed (warnings OK) |
| `1` | Runtime error | Validation failure, business logic error, I/O error |
| `2` | Usage error | Unknown command, unknown flag, wrong argument count |

```go
// Exit code constants in go-command
command.ExitSuccess    = 0
command.ExitError      = 1
command.ExitUsageError = 2
```

## Structured Errors with go-command

### Error Type

```go
err := &command.Error{
    Code:      command.ExitError,
    Message:   "user 'alice' not found",
    Hint:      "run 'myapp user list' to see available users",
    ErrorCode: "ERR_NOT_FOUND",
    Details:   []string{
        "searched: local database",
        "searched: remote API",
    },
    NextActions: []string{
        "myapp user list",
        "myapp user create alice",
    },
}
```

**Human output:**
```
Error: user 'alice' not found

  searched: local database
  searched: remote API

Hint: run 'myapp user list' to see available users
```

**Agent output (JSON):**
```json
{
  "error": "user 'alice' not found",
  "code": "ERR_NOT_FOUND",
  "hint": "run 'myapp user list' to see available users",
  "details": ["searched: local database", "searched: remote API"],
  "next_actions": ["myapp user list", "myapp user create alice"]
}
```

### NextActions — Structured Recovery

`NextActions` provides machine-readable follow-up commands. Agents can parse and execute these.

```go
// Authentication error with recovery commands
return &command.Error{
    Code:      command.ExitError,
    Message:   "authentication failed: token expired",
    ErrorCode: "ERR_AUTH_EXPIRED",
    Hint:      "re-authenticate with 'myapp auth login'",
    NextActions: []string{
        "myapp auth login",
        "myapp auth token --refresh",
    },
}

// Not found with discovery commands
return &command.Error{
    Code:      command.ExitError,
    Message:   fmt.Sprintf("project %q not found", name),
    ErrorCode: "ERR_NOT_FOUND",
    Hint:      fmt.Sprintf("did you mean %q?", closest),
    NextActions: []string{
        "myapp project list",
        "myapp project create " + name,
    },
}
```

### Error Constructors

```go
// Full struct literal for rich errors
err := &command.Error{
    Code:    command.ExitError,
    Message: "failed to deploy project",
    Hint:    "check network connectivity",
}

// Quick constructors for simple errors
err := command.NewError(command.ExitError, "operation failed")
err := command.NewErrorf(command.ExitUsageError, "invalid value %q for --%s", val, name)
```

### Wrapping Errors

```go
// Wrap with underlying cause (supports errors.Is/As unwrapping)
err := &command.Error{
    Code:    command.ExitError,
    Message: "failed to deploy project",
}
return err.Wrap(underlyingError)
```

## Error Design Principles

### 1. Tell What Went Wrong AND How to Fix

```go
// Good — actionable
return &command.Error{
    Message:     "config file not found",
    Hint:        "run 'myapp init' to create a project configuration",
    NextActions: []string{"myapp init"},
}

// Bad — user stuck
return fmt.Errorf("config file not found")
```

### 2. Suggest the Nearest Correct Input

```go
// Good — did-you-mean
return &command.Error{
    Message: fmt.Sprintf("unknown role %q", role),
    Hint:    fmt.Sprintf("did you mean %q?", closest),
    Details: []string{fmt.Sprintf("available: %s", strings.Join(available, ", "))},
}
```

### 3. Validate Early, Fail Before Side Effects

```go
Handler: func(ctx *command.Context) error {
    // Validate ALL inputs before any mutation
    name := strings.TrimSpace(ctx.Args[0])
    if name == "" {
        return &command.Error{
            Code:    command.ExitUsageError,
            Message: "name cannot be empty",
        }
    }

    if !projectExists(name) {
        return &command.Error{
            Code:        command.ExitError,
            Message:     "project not initialized",
            Hint:        "run 'myapp init' first",
            NextActions: []string{"myapp init"},
        }
    }

    // All validation passed — now do work
    return doMutation(name)
}
```

### 4. Use MinArgs/MaxArgs for Argument Validation

```go
// go-command validates automatically and returns ExitUsageError
cmd := &command.Command{
    Name:    "show",
    MinArgs: 1,
    MaxArgs: 1,
    // Handler won't run if args count is wrong
}
```

### 5. Use Sentinel Errors for Programmatic Matching

```go
// Package-level sentinels
var (
    ErrNotFound      = errors.New("not found")
    ErrAlreadyExists = errors.New("already exists")
    ErrInvalidInput  = errors.New("invalid input")
    ErrUnauthorized  = errors.New("unauthorized")
)

// Wrap with context
return fmt.Errorf("%w: project %q", ErrNotFound, name)

// Match in handler or test
if errors.Is(err, ErrNotFound) {
    return &command.Error{
        Code:        command.ExitError,
        Message:     err.Error(),
        Hint:        "check the name and try again",
        NextActions: []string{"myapp project list"},
    }
}
```

## Multi-Error Collection

For commands that process multiple items, collect all errors before returning.

```go
Handler: func(ctx *command.Context) error {
    var errs command.Errors  // value type, not pointer

    for _, file := range files {
        if err := process(file); err != nil {
            errs.Addf("failed to process %s: %v", file, err)
        }
    }

    return errs.OrNil()  // nil if no errors
}
```

Methods: `Add(error)`, `Addf(format, args...)`, `Len()`, `HasErrors()`, `OrNil()`.
Supports `errors.Is()` and `errors.As()` via `Unwrap() []error`.

**Output (joined with "; "):**
```
failed to process user.json: missing required field "email"; failed to process config.yaml: invalid format
```

## go-command Sentinel Errors

Use `errors.Is()` to match these:

| Error | When | Exit Code |
|-------|------|-----------|
| `ErrUnknownCommand` | Typo in command name | 2 |
| `ErrUnknownFlag` | Typo in flag name | 2 |
| `ErrRequiredFlag` | Missing required flag | 2 |
| `ErrTooFewArgs` | Not enough positional args | 2 |
| `ErrTooManyArgs` | Too many positional args | 2 |
| `ErrFlagValue` | Invalid flag value | 2 |
| `ErrFlagConflict` | Conflicting flags used together | 2 |
| `ErrFlagRequires` | Dependency flag missing | 2 |
| `ErrInvalidEnum` | Value not in enum | 2 |
| `ErrExclusiveGroup` | Multiple flags from same Exclusive group | 2 |
| `ErrRequiredTogetherGroup` | Partial flags from RequiredTogether group | 2 |
| `ErrHelpRequested` | `--help` used (not an error) | 0 |
| `ErrMissingHandler` | Command has no handler and no subcommands | -- |

Helper: `command.IsHelpRequested(err)` is shorthand for `errors.Is(err, ErrHelpRequested)`.

### Error Wrapper Types

`RunCommandError` wraps handler errors with command context:

```go
var runErr *command.RunCommandError
if errors.As(err, &runErr) {
    fmt.Println(runErr.Cmd.Name)  // which command failed
    fmt.Println(runErr.Err)        // underlying error
}
```

## Custom Error Handler

```go
app := command.New(command.Config{
    ErrorHandler: func(ctx *command.Context, err error) error {
        var cmdErr *command.Error
        if errors.As(err, &cmdErr) {
            // Structured error — output as-is
            return cmdErr
        }

        // Wrap unstructured errors
        return &command.Error{
            Code:    command.ExitError,
            Message: err.Error(),
            Hint:    "run with --debug for more details",
        }
    },
})
```

## ErrorCode Naming Convention

Prefix error codes by domain for machine parsing:

| Prefix | Domain | Example |
|--------|--------|---------|
| `ERR_AUTH_` | Authentication | `ERR_AUTH_EXPIRED` |
| `ERR_USER_` | User operations | `ERR_USER_NOT_FOUND` |
| `ERR_CONFIG_` | Configuration | `ERR_CONFIG_NOT_FOUND` |
| `ERR_SYNC_` | Sync/push/pull | `ERR_SYNC_CONFLICT` |
| `ERR_DEPLOY_` | Deployment | `ERR_DEPLOY_FAILED` |

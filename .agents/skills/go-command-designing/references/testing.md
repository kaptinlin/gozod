# CLI Testing

Test CLI commands without OS-level side effects using go-command's Exec API.

## Core Principle

Never call `app.Run()` in tests — it uses `os.Exit()`. Use `app.Exec()` or `app.ExecWith()` instead.

## Basic Testing with Exec

```go
func TestUserList(t *testing.T) {
    t.Parallel()

    app := buildApp()  // your app constructor
    result, err := app.Exec("user", "list")

    require.NoError(t, err)
    assert.Contains(t, result.Stdout, "alice")
    assert.Equal(t, 0, result.ExitCode)
}
```

## Testing with Flags and Args

```go
func TestUserShow(t *testing.T) {
    t.Parallel()

    app := buildApp()
    result, err := app.Exec("user", "show", "alice", "--format", "json")

    require.NoError(t, err)

    var user UserInfo
    err = json.Unmarshal([]byte(result.Stdout), &user)
    require.NoError(t, err)
    assert.Equal(t, "alice", user.Name)
}
```

## Testing with Custom Environment

```go
func TestConfigFromEnv(t *testing.T) {
    t.Parallel()

    app := buildApp()
    result, err := app.ExecWith(command.Request{
        Args:  []string{"config", "show"},
        Env:   map[string]string{"MYAPP_CONFIG": "/custom/path"},
        Stdin: strings.NewReader(""),
    })

    require.NoError(t, err)
    assert.Contains(t, result.Stdout, "/custom/path")
}
```

## Testing DryRunOption

```go
func TestDeployDryRun(t *testing.T) {
    t.Parallel()

    app := buildApp()
    result, err := app.Exec("deploy", "myapp", "--dry-run")

    require.NoError(t, err)
    assert.Contains(t, result.Stdout, "Action")
    assert.Contains(t, result.Stdout, "deploy")
    // Verify no side effects occurred
}
```

## Testing Respond[T] Output

```go
func TestProjectShowAgentMode(t *testing.T) {
    t.Parallel()

    app := buildApp()
    result, err := app.Exec("project", "show", "myapp", "--mode", "agent")

    require.NoError(t, err)

    var resp struct {
        Data struct {
            Name   string `json:"name"`
            Status string `json:"status"`
        } `json:"data"`
        Meta struct {
            NextActions []string `json:"next_actions"`
        } `json:"meta"`
    }
    err = json.Unmarshal([]byte(result.Stdout), &resp)
    require.NoError(t, err)
    assert.Equal(t, "myapp", resp.Data.Name)
    assert.NotEmpty(t, resp.Meta.NextActions)
}
```

## Testing Piped Input

```go
func TestStdinInput(t *testing.T) {
    t.Parallel()

    app := buildApp()
    result, err := app.ExecWith(command.Request{
        Args:  []string{"import"},
        Stdin: strings.NewReader(`{"name":"alice","role":"admin"}`),
    })

    require.NoError(t, err)
    assert.Contains(t, result.Stdout, "alice")
}
```

## Testing Error Cases

```go
func TestUnknownCommand(t *testing.T) {
    t.Parallel()

    app := buildApp()
    _, err := app.Exec("nonexistent")

    require.Error(t, err)
    assert.True(t, errors.Is(err, command.ErrUnknownCommand))
}

func TestMissingRequiredArg(t *testing.T) {
    t.Parallel()

    app := buildApp()
    _, err := app.Exec("user", "show")  // missing name

    require.Error(t, err)
    assert.True(t, errors.Is(err, command.ErrTooFewArgs))
}

func TestConflictingFlags(t *testing.T) {
    t.Parallel()

    app := buildApp()
    _, err := app.Exec("build", "--watch", "--check")

    require.Error(t, err)
    assert.True(t, errors.Is(err, command.ErrFlagConflict))
}

func TestErrorNextActions(t *testing.T) {
    t.Parallel()

    app := buildApp()
    _, err := app.Exec("user", "show", "nonexistent")

    require.Error(t, err)
    var cmdErr *command.Error
    require.True(t, errors.As(err, &cmdErr))
    assert.NotEmpty(t, cmdErr.NextActions)
    assert.Contains(t, cmdErr.NextActions[0], "user list")
}
```

## Table-Driven CLI Tests

```go
func TestTaskQuery(t *testing.T) {
    t.Parallel()

    tests := []struct {
        name     string
        args     []string
        wantOut  string
        wantErr  bool
        wantCode int
    }{
        {
            name:    "list all tasks",
            args:    []string{"task", "list"},
            wantOut: "Fix login",
        },
        {
            name:     "show nonexistent task",
            args:     []string{"task", "show", "99999"},
            wantErr:  true,
            wantCode: 1,
        },
        {
            name:    "list with status filter",
            args:    []string{"task", "list", "--status", "open"},
            wantOut: "open",
        },
        {
            name:    "list with pagination",
            args:    []string{"task", "list", "--limit", "5", "--offset", "0"},
            wantOut: "Fix login",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            app := buildApp()
            result, err := app.Exec(tt.args...)

            if tt.wantErr {
                require.Error(t, err)
                if tt.wantCode > 0 {
                    var cmdErr *command.Error
                    require.True(t, errors.As(err, &cmdErr))
                    assert.Equal(t, tt.wantCode, cmdErr.ExitCode())
                }
                return
            }

            require.NoError(t, err)
            assert.Contains(t, result.Stdout, tt.wantOut)
        })
    }
}
```

## Testing with testutil

The `testutil` package provides `TestApp` for simplified test setup and custom assert helpers.

### TestApp

```go
import "github.com/agentable/go-command/testutil"

func TestWithTestutil(t *testing.T) {
    t.Parallel()

    app := testutil.New(myCommand)        // wraps command.New with test config
    err := app.Run("show", "alice")       // captures stdout/stderr

    require.NoError(t, err)
    assert.Contains(t, app.Output(), "alice")       // stdout
    assert.Empty(t, app.ErrorOutput())               // stderr

    app.Reset()                            // clear buffers for next Run
    err = app.Run("show", "bob")
    // ...
}

// With custom config:
app := testutil.NewWithConfig(command.Config{
    Name:    "myapp",
    Version: "1.0.0",
}, cmd1, cmd2)
```

### Assert Helpers

```go
testutil.AssertOutput(t, app, "alice")           // stdout contains "alice"
testutil.AssertError(t, err, "not found")        // err contains "not found"
testutil.AssertExitCode(t, err, command.ExitError) // exit code matches
```
```

## Testing Mode-Specific Output

Mode can be set via `--mode` flag or `COMMAND_MODE` env var (requires mode middleware).

```go
func TestOutputModes(t *testing.T) {
    t.Parallel()

    tests := []struct {
        name string
        mode string
        want string
    }{
        {"human mode", "human", "Name:"},
        {"agent mode", "agent", `{"data":`},
        {"ci mode", "ci", "name="},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            app := buildApp()  // must include mode.New() middleware
            result, err := app.Exec("user", "show", "alice", "--mode", tt.mode)
            require.NoError(t, err)
            assert.Contains(t, result.Stdout, tt.want)
        })
    }
}

// Alternative: set mode via COMMAND_MODE env
result, err := app.ExecWith(command.Request{
    Args: []string{"user", "show", "alice"},
    Env:  map[string]string{"COMMAND_MODE": "agent"},
})
```
```

## Testing Middleware

```go
func TestAuthMiddleware(t *testing.T) {
    t.Parallel()

    // Command with auth middleware
    cmd := &command.Command{
        Name:       "deploy",
        Middleware: []command.MiddlewareFunc{requireAuth},
        Handler:    deployHandler,
    }

    app := testutil.New(cmd)

    // Without auth — should fail
    err := app.Run("deploy", "myapp")
    require.Error(t, err)
    assert.Contains(t, err.Error(), "authentication required")

    // With auth via env
    result, err := app.ExecWith(command.Request{
        Args: []string{"deploy", "myapp"},
        Env:  map[string]string{"API_TOKEN": "test-token"},
    })
    require.NoError(t, err)
}
```

## Testing Best Practices

| Practice | Why |
|----------|-----|
| Always `t.Parallel()` | Detect race conditions, faster execution |
| Test both success and error paths | Errors are part of the UX |
| Test exit codes explicitly | Scripts and agents depend on exit codes |
| Test mode-specific output | Agents and humans see different formats |
| Test `Respond[T]` envelope in Agent mode | Verify `data` + `meta` structure |
| Test `DryRunOption` behavior | Verify no side effects in dry-run |
| Test `NextActions` in errors | Agents use these for recovery |
| Use `Exec` not `Run` in tests | `Run` calls `os.Exit` |
| Inject dependencies | Don't hit real APIs in unit tests |
| Table-driven for variant testing | Cover all flag combinations systematically |
| Test completion handlers | Ensure dynamic completion works |
| Test piped input with `ExecWith` | Verify `HasPipedInput()` works |

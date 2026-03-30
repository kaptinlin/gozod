# Watch and Hot Reload Patterns

Patterns for using go-config's live reload capabilities.

## Contents

- Basic watch setup
- Path-filtered callbacks
- Typed callbacks with OnChangeFunc
- Dynamic source addition
- Status monitoring
- Graceful shutdown

## Basic Watch Setup

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

cfg, err := config.Load[AppConfig]([]config.Source{
    file.New("config.yaml"),
    env.New(env.WithPrefix("APP_")),
})
if err != nil {
    log.Fatal(err)
}

// Register callbacks BEFORE starting watch
cfg.OnChange(func(c *config.Config[AppConfig]) {
    log.Println("config changed")
})

// Watch blocks -- run in goroutine
go cfg.Watch(ctx)
```

`Watch` monitors all sources that implement the `Watcher` interface (currently: `file.File`, `file.Discovery`). When any watcher signals a change, **all** sources are reloaded atomically (not just the changed source).

The file watcher watches the parent directory via `fsnotify`, so symlink changes are caught too. Events within a 5ms window are deduplicated.

## Path-Filtered Callbacks

```go
// Fires only when database.host or database.port changes
cfg.OnChange(func(c *config.Config[AppConfig]) {
    v := c.Value()
    reconnectDB(v.Database.Host, v.Database.Port)
}, "database.host", "database.port")

// Fires on any change under server.*
cfg.OnChange(func(c *config.Config[AppConfig]) {
    v := c.Value()
    rebindServer(v.Server)
}, "server")
```

Path comparison uses `reflect.DeepEqual` on the subtree at the given dot-separated path.

## Typed Callbacks with OnChangeFunc

For performance-critical paths, `OnChangeFunc` uses native `==` comparison:

```go
// Only fires when the specific string value changes
config.OnChangeFunc(cfg,
    func(c AppConfig) string { return c.Database.Host },
    func(host string) { reconnectDB(host) },
)

// Works with any comparable type
config.OnChangeFunc(cfg,
    func(c AppConfig) int { return c.Server.Port },
    func(port int) { rebindPort(port) },
)
```

The selector must return a `comparable` type. The first invocation establishes the baseline; the callback fires on subsequent changes only.

## Dynamic Source Addition

Sources added via `cfg.Load()` after `Watch()` is running are automatically registered:

```go
go cfg.Watch(ctx)

// Later: add a new source (triggers full reload of all sources)
err := cfg.Load(ctx, file.New("extra-config.yaml"))
```

New `Watcher` sources are automatically registered for live reload.

## Status Monitoring

Track source health with `WithOnStatus` or per-provider `Status()`:

```go
// Global status callback via option
cfg, err := config.Load[AppConfig](
    []config.Source{file.New("config.yaml")},
    config.WithOnStatus(func(s config.Source, changed bool, err error) {
        name := fmt.Sprintf("%s", s)
        if err != nil {
            metrics.Increment("config.error", "source", name)
            log.Printf("config source %s error: %v", name, err)
        } else if changed {
            metrics.Increment("config.reload", "source", name)
        }
    }),
)

// Per-provider status callback
fp := file.New("config.yaml")
fp.Status(func(changed bool, err error) {
    if err != nil {
        log.Printf("config file error: %v", err)
    }
})
```

Only sources implementing `StatusReporter` report status (currently: `file.File`, `file.Discovery`).

## Graceful Shutdown

```go
func main() {
    ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
    defer cancel()

    cfg, err := config.Load[AppConfig]([]config.Source{
        file.New("config.yaml"),
        env.New(env.WithPrefix("APP_")),
    })
    if err != nil {
        log.Fatal(err)
    }

    cfg.OnChange(func(c *config.Config[AppConfig]) {
        applyConfig(c.Value())
    })

    // Watch exits when ctx is cancelled (on SIGINT)
    if err := cfg.Watch(ctx); err != nil {
        log.Printf("watch error: %v", err)
    }
}
```

`Watch` returns `nil` when the context is cancelled. Non-nil errors indicate a watcher failure (propagated via `context.CancelCauseFunc`).

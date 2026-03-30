# Complete Integration Example

Full application wiring library Config structs through go-config.

## Project Structure

```
myapp/
├── main.go                 # Entry point, config loading, wiring
├── config.go               # AppConfig struct (composes library configs)
├── config.yaml             # Default config file
└── internal/
    └── sandbox/
        └── setup.go        # Sandbox assembly from config
```

## config.go — Application Config

```go
package main

import (
    "time"

    "github.com/agentable/go-sandbox"
    "github.com/agentable/go-sandbox/backend/docker"
    "github.com/agentable/go-sandbox/plugin/commandguard"
    "github.com/agentable/go-sandbox/plugin/domainfilter"
)

type AppConfig struct {
    Server  ServerConfig  `json:"server"`
    Sandbox SandboxConfig `json:"sandbox"`
    Log     LogConfig     `json:"log"`
}

type ServerConfig struct {
    Host            string        `json:"host"`
    Port            int           `json:"port"`
    ShutdownTimeout time.Duration `json:"shutdown_timeout"`
}

type SandboxConfig struct {
    Policy sandbox.Policy       `json:"policy"`
    Docker docker.Config        `json:"docker"`
    Guard  commandguard.Config  `json:"guard"`
    Filter domainfilter.Config  `json:"filter"`
}

type LogConfig struct {
    Level  string `json:"level"`
    Format string `json:"format"`
}
```

## main.go — Loading and Wiring

```go
package main

import (
    "context"
    "log"
    "log/slog"
    "os"
    "time"

    "github.com/agentable/go-config"
    "github.com/agentable/go-config/provider/env"
    "github.com/agentable/go-config/provider/file"
    "github.com/agentable/go-config/provider/static"
    _ "github.com/agentable/go-config/format/yaml"
)

var defaults = AppConfig{
    Server: ServerConfig{
        Host:            "0.0.0.0",
        Port:            8080,
        ShutdownTimeout: 15 * time.Second,
    },
    Log: LogConfig{
        Level:  "info",
        Format: "json",
    },
}

func main() {
    cfg, err := config.Load[AppConfig]([]config.Source{
        static.New(defaults),
        file.New("config.yaml"),
        file.New("config.local.yaml", file.WithPolicy(config.PolicyOptional)),
        env.New(env.WithPrefix("APP_")),
    }, config.WithDecodeHook(func(s string) (time.Duration, error) {
        return time.ParseDuration(s)
    }))
    if err != nil {
        log.Fatal(err)
    }

    v := cfg.Value()

    // Build logger from config
    logger := buildLogger(v.Log)

    // Build sandbox from config + runtime deps
    sb, err := buildSandbox(v.Sandbox, logger)
    if err != nil {
        log.Fatal(err)
    }
    defer sb.Close(context.Background())

    // Use sandbox...
    result, err := sb.Exec(context.Background(), "echo hello")
    if err != nil {
        logger.Error("exec failed", slog.Any("error", err))
        os.Exit(1)
    }
    logger.Info("exec complete",
        slog.Int("exit_code", result.ExitCode),
        slog.String("stdout", string(result.Stdout)),
    )
}

func buildLogger(cfg LogConfig) *slog.Logger {
    var handler slog.Handler
    opts := &slog.HandlerOptions{}

    switch cfg.Level {
    case "debug":
        opts.Level = slog.LevelDebug
    case "warn":
        opts.Level = slog.LevelWarn
    case "error":
        opts.Level = slog.LevelError
    default:
        opts.Level = slog.LevelInfo
    }

    switch cfg.Format {
    case "text":
        handler = slog.NewTextHandler(os.Stderr, opts)
    default:
        handler = slog.NewJSONHandler(os.Stderr, opts)
    }

    return slog.New(handler)
}
```

## internal/sandbox/setup.go — Sandbox Assembly

```go
package sandbox

import (
    "log/slog"

    "github.com/agentable/go-sandbox"
    "github.com/agentable/go-sandbox/backend/docker"
    "github.com/agentable/go-sandbox/plugin/commandguard"
    "github.com/agentable/go-sandbox/plugin/domainfilter"
)

func Build(cfg SandboxConfig, logger *slog.Logger) (*sandbox.Sandbox, error) {
    backend, err := docker.New(cfg.Docker, docker.WithLogger(logger))
    if err != nil {
        return nil, err
    }

    return sandbox.New(
        sandbox.WithBackend(backend),
        sandbox.WithPolicy(cfg.Policy),
        sandbox.WithPlugin(commandguard.New(cfg.Guard)),
        sandbox.WithPlugin(domainfilter.New(cfg.Filter, domainfilter.WithLogger(logger))),
        sandbox.WithLogger(logger),
    )
}
```

## config.yaml

```yaml
server:
  host: 0.0.0.0
  port: 8080
  shutdown_timeout: 15s

sandbox:
  policy:
    network: deny
    writable_paths:
      - /tmp
      - /app/workspace
    readable_paths:
      - /usr
      - /bin
      - /lib
    inherit_env: true
    resources:
      timeout: 30s
      max_memory_mb: 512
      max_cpu_percent: 100
      max_pids: 256
  docker:
    image: node:20-slim
    memory_limit: 1g
    cpu_quota: 1.0
    pids_limit: 512
  guard:
    deny_patterns:
      - "rm -rf /"
      - "shutdown"
      - "reboot"
      - "mkfs"
    detect_level: deep
  filter:
    allowed_domains:
      - "*.example.com"
      - "api.openai.com"
      - "registry.npmjs.org"
    denied_domains:
      - "*.internal.corp"

log:
  level: info
  format: json
```

## Environment Override Examples

```bash
# Override docker image
APP_SANDBOX_DOCKER_IMAGE=python:3.12

# Override network policy
APP_SANDBOX_POLICY_NETWORK=allow

# Override server port
APP_SERVER_PORT=9090

# Override log level
APP_LOG_LEVEL=debug
```

## Key Observations

1. **Library types compose directly** — `docker.Config`, `commandguard.Config` are fields in `AppConfig`
2. **No intermediate types** — No `PolicyConfig` or `ResourcesConfig` wrappers
3. **TextUnmarshaler does the work** — `"deny"` → `NetworkDeny`, `"deep"` → `DetectDeep`
4. **Decode hook for Duration** — Registered once at load time
5. **Runtime deps wired separately** — Logger passed via `WithLogger()`, not in Config
6. **go-config stays in main** — Libraries never import it

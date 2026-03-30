# Provider Reference

Detailed configuration for each go-config provider.

## Contents

- File provider
- File discovery provider
- FS provider (embed.FS)
- Environment provider
- Flag provider
- Static provider
- Secrets provider
- Platform paths

## File Provider

```go
import "github.com/agentable/go-config/provider/file"

file.New("config.yaml")
file.New("config.json", file.WithPolicy(config.PolicyOptional))
```

Auto-detects format by extension. Supports `Watch()` for live reload via `fsnotify`. Watches the parent directory so symlink changes are caught too.

**Options:**
- `file.WithPolicy(p)` -- `PolicyRequired` (default) or `PolicyOptional`

**String representation:** `file://config.yaml` (used in `Explain` output)

## File Discovery Provider

Walk-up file discovery -- searches from a start directory toward the filesystem root:

```go
import "github.com/agentable/go-config/provider/file"

// Discover nearest config file (walks up from cwd)
file.NewDiscovery("config.yaml")

// With custom start directory
file.NewDiscovery("config.yaml", file.WithDiscoveryStartDir("/app/service"))

// Optional -- skip if not found
file.NewDiscovery("config.local.yaml",
    file.WithDiscoveryPolicy(config.PolicyOptional),
)
```

Supports `Watch()` after the first successful `Load` discovers the file path. Returns `file.ErrNotFound` if the file is not found anywhere in the ancestor chain. Returns `file.ErrNotLoaded` if `Watch` is called before `Load`.

**Standalone discovery function:**

```go
// Returns the full path, or ("", nil) if not found
path, err := file.Discover("config.yaml")
path, err := file.Discover("config.yaml", file.WithStartDir("/app"))
```

**Options:**
- `file.WithDiscoveryPolicy(p)` -- error handling policy
- `file.WithDiscoveryStartDir(dir)` -- start directory (default: `os.Getwd()`)

**String representation:** `discover://config.yaml` (before load), `file:///resolved/path/config.yaml` (after load)

## FS Provider (embed.FS)

```go
import "github.com/agentable/go-config/provider/fs"

//go:embed config/defaults.json
var defaultsFS embed.FS

fs.New(defaultsFS, "config/defaults.json")
```

Reads from any `io/fs.FS` implementation. No watch support (embedded files are immutable).

**Options:**
- `fs.WithDecoder(d)` -- override format auto-detection with custom `FormatDecoder`
- `fs.WithPolicy(p)` -- error handling policy

**String representation:** `fs[config/defaults.json]`

## Environment Provider

```go
import "github.com/agentable/go-config/provider/env"

env.New(env.WithPrefix("APP_"))
```

Maps environment variables to nested config structure. Splits names by delimiter to create hierarchy. Empty values are treated as unset.

```
APP_SERVER_HOST=0.0.0.0 -> {SERVER: {HOST: "0.0.0.0"}}
```

**Double delimiter = literal delimiter:**
```
APP_MY__KEY=v -> {MY_KEY: "v"}
```

**Options:**

| Option | Description |
|--------|-------------|
| `WithPrefix(p)` | Filter vars by prefix, strip prefix from names |
| `WithDelimiter(d)` | Nesting delimiter (default: `"_"`) |
| `WithExpand()` | Enable `${VAR}` / `$VAR` expansion in values |
| `WithFieldNameMapping(m)` | Map env var segments to struct field names |
| `WithSource(m)` | Provide custom `map[string]string` (for testing) |
| `WithPolicy(p)` | Error handling policy |

**String representation:** `env:APP_*`

## Flag Provider

```go
import "github.com/agentable/go-config/provider/flag"

type CLIConfig struct {
    Port    int    `flag:"port,p"    json:"port"`
    Host    string `flag:"host,h"    json:"host"`
    Verbose bool   `flag:"verbose,v" json:"verbose"`
}

flags, err := flag.New(CLIConfig{Port: 8080}, os.Args[1:])
```

Creates CLI flags from struct `flag` tags using `spf13/pflag`. The tag format is `"name,shorthand"`. Only flags explicitly set on the command line appear in the output map (unset flags do not override lower-priority sources). `flag.New` is generic -- infers the type from the defaults argument.

**Supported types:** `string`, `int`, `int64`, `bool`, `float64`, `[]string`

**Nested structs:** The `json` tag name is used for nested key paths. Fields without `flag` tags in nested structs are traversed automatically.

**Options:**
- `flag.WithPolicy(p)` -- error handling policy

**String representation:** `flag`

## Static Provider

```go
import "github.com/agentable/go-config/provider/static"

// From typed struct (generic -- zero-value fields skipped)
static.New(AppConfig{Server: Server{Port: 8080}})

// From map literal
static.NewMap(map[string]any{"server": map[string]any{"port": 8080}})
```

In-memory defaults. Use as the first source to establish default values. `static.New` is generic -- converts the struct to `map[string]any` via reflection, skipping zero-value fields.

**Options:**
- `static.WithPolicy(p)` -- error handling policy

**String representation:** `static`

## Secrets Provider

```go
import (
    "github.com/agentable/go-secrets"
    cfgsecrets "github.com/agentable/go-config/provider/secrets"
)

s, err := secrets.New(secrets.WithStore(store))
secretsSrc := cfgsecrets.New(s, "prod")
```

Loads secrets from `agentable/go-secrets` as a regular config source. The first argument is a `*secrets.Secrets` instance, the second is the scope (e.g., `"prod"`, `"staging"`). Secret keys use dot notation that maps to nested config paths.

**Dual integration mode:**

```go
secretsSrc := cfgsecrets.New(s, "prod")

cfg, err := config.Load[AppConfig]([]config.Source{
    file.New("config.yaml"),
    secretsSrc,                                      // source mode: secrets as config values
}, config.WithSecretResolver(secretsSrc.Resolver())) // interpolation mode: ${secret:KEY}
```

`Resolver()` returns a `SecretResolverFunc` that resolves `${secret:NAME}` references by looking up secrets in the configured scope. Missing secrets resolve to empty string.

**Options:**
- `cfgsecrets.WithPolicy(p)` -- error handling policy

**String representation:** `secrets://prod`

For comprehensive secrets management, see the **secrets-managing** skill.

## Platform Paths

The `file` package provides platform-specific config directory helpers:

```go
import "github.com/agentable/go-config/provider/file"

// User config directory
home, err := file.ConfigHome()
// Linux/Unix: $XDG_CONFIG_HOME or ~/.config
// macOS:      ~/Library/Preferences
// Windows:    %APPDATA%

// System config directories (least to most important)
dirs := file.SystemConfigDirs()
// Linux/Unix: /etc, then $XDG_CONFIG_DIRS (default: /etc/xdg)
// macOS:      /Library/Preferences
// Windows:    C:\ProgramData
```

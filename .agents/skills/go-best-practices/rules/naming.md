# Naming

Naming is the most visible aspect of Go style. Consistent, idiomatic names make code self-documenting and reduce cognitive load across the codebase.

## Contents
- Avoid Redundant Naming
- Package Naming Conventions
- Receiver Naming Conventions
- Constant Naming (MixedCaps)
- Acronym Casing Rules
- No Get Prefix for Getters
- Variable Naming by Scope
- Function and Method Naming

---

## Avoid Redundant Naming

Names should not repeat information already clear from context — package name, receiver type, parameter types, or return types.

**Incorrect:**

```go
package yamlconfig
func ParseYAMLConfig(input string) (*Config, error)

func (c *Config) WriteConfigTo(w io.Writer) (int64, error)

func OverrideFirstWithSecond(dest, source *Config) error

func TransformToJSON(input *Config) *jsonconfig.Config
```

**Correct:**

```go
package yamlconfig
func Parse(input string) (*Config, error)

func (c *Config) WriteTo(w io.Writer) (int64, error)

func Override(dest, source *Config) error

func Transform(input *Config) *jsonconfig.Config
```

Also applies to exported symbols — `widget.NewWidget` should be `widget.New`, `db.LoadFromDatabase` should be `db.Load`.

---

## Package Naming Conventions

Package names must be short, lowercase only, no underscores. Avoid generic names like `util`, `helper`, `common`, `model`. Consider how the package name reads at call sites.

**Incorrect:**

```go
db := test.NewDatabaseFromFile(...)
_, err := f.Seek(0, common.SeekStart)
b := helper.Marshal(curve, x, y)
```

**Correct:**

```go
db := spannertest.NewDatabaseFromFile(...)
_, err := f.Seek(0, io.SeekStart)
b := elliptic.Marshal(curve, x, y)
```

**Incorrect:**

```go
package tab_writer
package tabWriter
```

**Correct:**

```go
package tabwriter
```

Avoid names easily shadowed by common local variables. Use `usercount` instead of `count`.

---

## Receiver Naming Conventions

Method receiver names must be short (one or two letters), abbreviations of the type name, and consistent across all methods of that type.

**Incorrect:**

```go
func (tray Tray) Method()
func (this *ReportWriter) Write()
func (self *Scanner) Scan()
func (info *ResearchInfo) Process()
```

**Correct:**

```go
func (t Tray) Method()
func (w *ReportWriter) Write()
func (s *Scanner) Scan()
func (ri *ResearchInfo) Process()
```

Never use `this` or `self` — these are not idiomatic in Go.

---

## Constant Naming (MixedCaps)

Constants use MixedCaps like all Go names. Never use ALL_CAPS or `k` prefix, even if other languages use those conventions. Name by role, not by value.

**Incorrect:**

```go
const MAX_PACKET_SIZE = 512
const kMaxBufferSize = 1024
const KMaxUsersPergroup = 500

const Twelve = 12
```

**Correct:**

```go
const MaxPacketSize = 512

const (
    ExecuteBit = 1 << iota
    WriteBit
    ReadBit
)
```

Only define constants when the name conveys meaning beyond the value itself.

---

## Acronym Casing Rules

Acronyms and initialisms in names should maintain consistent casing. `URL` is `URL` or `url`, never `Url`. Same for `ID`, `DB`, `HTTP`, `API`.

| Usage | Scope | Correct | Incorrect |
|---|---|---|---|
| XML API | Exported | `XMLAPI` | `XmlApi`, `XMLApi` |
| XML API | Unexported | `xmlAPI` | `xmlapi`, `xmlApi` |
| iOS | Exported | `IOS` | `Ios`, `IoS` |
| gRPC | Exported | `GRPC` | `Grpc` |
| gRPC | Unexported | `gRPC` | `grpc` |
| DDoS | Exported | `DDoS` | `DDOS`, `Ddos` |
| ID | Exported | `ID` | `Id` |
| DB | Exported | `DB` | `Db` |

**Incorrect:**

```go
type XmlHttpRequest struct{}
func GetUserId() string
```

**Correct:**

```go
type XMLHTTPRequest struct{}
func GetUserID() string
```

---

## No Get Prefix for Getters

Functions returning values should use noun names. Do not use `Get` or `get` prefix unless the underlying concept uses "get" (e.g., HTTP GET). Use `Compute` or `Fetch` for expensive operations.

**Incorrect:**

```go
func (c *Config) GetJobName(key string) (value string, ok bool)
```

**Correct:**

```go
func (c *Config) JobName(key string) (value string, ok bool)
```

Functions performing actions should use verb names:

```go
func (c *Config) WriteDetail(w io.Writer) (int64, error)
```

---

## Variable Naming by Scope

Name length should be proportional to scope size and inversely proportional to usage frequency within that scope.

- **Small scope** (1-7 lines): single character or very short (`i`, `r`, `w`)
- **Medium scope** (8-15 lines): one word (`count`, `users`)
- **Large scope** (15-25 lines): may need multiple words (`userCount`)
- **Very large scope** (25+ lines): descriptive names needed

**Incorrect:**

```go
var numUsers int
var nameString string
var userSlice []User
```

**Correct:**

```go
var users int
var name string
var users []User
```

**Incorrect:**

```go
func (db *DB) UserCount() (userCount int, err error) {
    var userCountInt64 int64
    if dbLoadError := db.LoadFromDatabase("count(distinct users)", &userCountInt64); dbLoadError != nil {
        return 0, fmt.Errorf("failed to load user count: %s", dbLoadError)
    }
    userCount = int(userCountInt64)
    return userCount, nil
}
```

**Correct:**

```go
func (db *DB) UserCount() (int, error) {
    var count int64
    if err := db.Load("count(distinct users)", &count); err != nil {
        return 0, fmt.Errorf("failed to load user count: %s", err)
    }
    return int(count), nil
}
```

Common single-letter conventions: `r` for `io.Reader`, `w` for `io.Writer`, `b` for `[]byte`, `s` for `string`, `i` for index.

---

## Function and Method Naming

Return-value functions use noun names. Action functions use verb names. Type-differentiated functions append the type name.

```go
// Noun for return-value function:
func (c *Config) JobName(key string) (string, bool)

// Verb for action function:
func (c *Config) WriteDetail(w io.Writer) (int64, error)

// Type suffix for type-differentiated variants:
func ParseInt(input string) (int, error)
func ParseInt64(input string) (int64, error)
```

If there's a clear "primary" version, omit the type from its name:

```go
func (c *Config) Marshal() ([]byte, error)
func (c *Config) MarshalText() (string, error)
```

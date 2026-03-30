# Go Code Simplification Examples

> Language-specific examples for Go 1.26+ code simplification patterns

## Documentation Disguised as Code - Go Examples

### 1. Unused Type Definitions

**Anti-pattern**: Types defined but never instantiated in production

```go
// ❌ Anti-pattern: Group struct never used
type Group struct {
    Name   string            `json:"name"`
    Type   string            `json:"$type,omitempty"`
    Tokens map[string]*Token `json:"tokens,omitempty"`
    Groups map[string]*Group `json:"groups,omitempty"`
}

// Production code uses map[string]interface{} instead
func buildTree() map[string]interface{} {
    return map[string]interface{}{
        "tokens": tokens,
        "groups": groups,
    }
}
```

**Detection**:
```bash
# Search for type usage
grep -r "Group{" --include="*.go"
grep -r "\*Group" --include="*.go"
# If only in definition file → unused
```

**Fix**: Delete the unused type

```go
// ✅ Use the actual implementation type
func buildTree() map[string]interface{} {
    return map[string]interface{}{
        "tokens": tokens,
        "groups": groups,
    }
}
```

---

### 2. Write-Only Registries

**Anti-pattern**: Registration functions that populate maps never queried

```go
// ❌ Anti-pattern: Write-only registry
var kindDescriptors = make(map[string]KindDescriptor)

func RegisterKind(name string, desc KindDescriptor) {
    kindDescriptors[name] = desc
}

// No LookupKind() or iteration over kindDescriptors exists
```

**Detection**:
```bash
# Find writes
grep -n "kindDescriptors\[" file.go

# Find reads
grep -n "kindDescriptors" file.go | grep -v "\["
# If no reads → write-only
```

**Fix**: Delete the registry

```go
// ✅ If registration is needed, ensure it's read
func RegisterKind(name string, desc KindDescriptor) {
    kindDescriptors[name] = desc
}

func LookupKind(name string) (KindDescriptor, bool) {
    desc, ok := kindDescriptors[name]
    return desc, ok
}
```

---

### 3. Validation Functions Never Called

**Anti-pattern**: Validate() methods only called in tests

```go
// ❌ Anti-pattern: Validation never called in production
type TierConfig struct {
    Name     string
    Priority int
}

func (c *TierConfig) Validate() error {
    if c.Priority < 0 {
        return fmt.Errorf("priority must be non-negative")
    }
    return nil
}

// Loading code never calls Validate()
func LoadConfig(path string) (*TierConfig, error) {
    var cfg TierConfig
    if err := yaml.Unmarshal(data, &cfg); err != nil {
        return nil, err
    }
    return &cfg, nil  // ❌ No validation!
}
```

**Fix**: Wire validation into production flow

```go
// ✅ Call validation during loading
func LoadConfig(path string) (*TierConfig, error) {
    var cfg TierConfig
    if err := yaml.Unmarshal(data, &cfg); err != nil {
        return nil, err
    }
    if err := cfg.Validate(); err != nil {
        return nil, fmt.Errorf("invalid config: %w", err)
    }
    return &cfg, nil
}
```

---

### 4. Enum Validation Duplication

**Anti-pattern**: Manual validation duplicating framework constraints

```go
// ❌ Anti-pattern: Duplicates CLI framework enum
// CLI option already declares: Enum: []string{"dtcg", "css", "sass"}
func buildModeFromFormat(format string) (BuildMode, error) {
    switch format {
    case "dtcg":
        return BuildModeDTCG, nil
    case "css":
        return BuildModeCSS, nil
    case "sass":
        return BuildModeSASS, nil
    default:
        return "", fmt.Errorf("unsupported format: %s", format)
    }
}
```

**Fix**: Trust the framework validation

```go
// ✅ Framework already validated, use directly
func buildModeFromFormat(format string) BuildMode {
    return BuildMode(format)
}
```

---

### 5. Mock-Only Interfaces

**Anti-pattern**: Interfaces with no production implementations

```go
// ❌ Anti-pattern: Only mockAuditLog exists
type AuditLog interface {
    Record(ctx context.Context, entry AuditEntry) error
    Query(ctx context.Context, fileKey string, limit int) ([]AuditEntry, error)
}

// Only in tests:
type mockAuditLog struct {
    entries []AuditEntry
}
```

**Detection**:
```bash
# Find interface definition
grep -n "type.*AuditLog.*interface" file.go

# Find implementations
grep -r "type.*struct" --include="*.go" | grep -v "_test.go"
# If no production implementations → mock-only
```

**Fix**: Delete the interface or implement it

```go
// ✅ Option 1: Delete if not needed
// Remove AuditLog interface and all references

// ✅ Option 2: Implement if needed
type SQLiteAuditLog struct {
    db *sql.DB
}

func (l *SQLiteAuditLog) Record(ctx context.Context, entry AuditEntry) error {
    // Real implementation
}
```

---

### 6. Premature Abstraction

**Anti-pattern**: Abstraction layer with single consumer

```go
// ❌ Anti-pattern: Service interface with one implementation
type TokenService interface {
    Validate(ctx context.Context, tokens []Token) error
    Build(ctx context.Context, tokens []Token) (string, error)
}

type tokenServiceImpl struct {
    validator Validator
    builder   Builder
}

// Only one consumer:
func NewMCPServer(svc TokenService) *Server {
    return &Server{tokenSvc: svc}
}
```

**Fix**: Inline until second consumer appears

```go
// ✅ Use concrete type directly
type TokenService struct {
    validator Validator
    builder   Builder
}

func NewMCPServer(svc *TokenService) *Server {
    return &Server{tokenSvc: svc}
}
```

---

### 7. Placeholder Implementations

**Anti-pattern**: Functions returning hardcoded placeholders

```go
// ❌ Anti-pattern: Placeholder implementation
func (a *TokenAdapter) Diff(ctx context.Context, fileKey string) (string, error) {
    return "Diff not implemented", nil
}
```

**Fix**: Return error or implement properly

```go
// ✅ Option 1: Return error
func (a *TokenAdapter) Diff(ctx context.Context, fileKey string) (string, error) {
    return "", fmt.Errorf("diff operation not supported")
}

// ✅ Option 2: Implement properly
func (a *TokenAdapter) Diff(ctx context.Context, fileKey string) (string, error) {
    local, err := a.loadLocal(ctx, fileKey)
    if err != nil {
        return "", err
    }
    remote, err := a.loadRemote(ctx, fileKey)
    if err != nil {
        return "", err
    }
    return diff.Compute(local, remote), nil
}
```

---

### 8. Silent Fallback Validation

**Anti-pattern**: Validation silently accepting invalid input

```go
// ❌ Anti-pattern: Silent fallback to default
func parseTokenDepth(tokenDepth string) Tier {
    switch tokenDepth {
    case "option":
        return TierOption
    case "decision":
        return TierDecision
    case "component":
        return TierComponent
    default:
        return TierDecision  // ❌ Silent fallback!
    }
}
```

**Fix**: Return error for invalid input

```go
// ✅ Return error for invalid input
func parseTokenDepth(tokenDepth string) (Tier, error) {
    switch tokenDepth {
    case "option":
        return TierOption, nil
    case "decision":
        return TierDecision, nil
    case "component":
        return TierComponent, nil
    default:
        return "", fmt.Errorf("invalid token depth: %s", tokenDepth)
    }
}
```

---

### 9. Field Existence Checks

**Anti-pattern**: Validation checking only field presence

```go
// ❌ Anti-pattern: Only checks field exists
func validateA11y(a11y map[string]interface{}) error {
    if _, hasAriaLabel := a11y["aria_label"]; !hasAriaLabel {
        return fmt.Errorf("missing aria_label")
    }
    // ❌ No validation of label content!
    return nil
}
```

**Fix**: Use JSON Schema validation

```go
// ✅ Use JSON Schema for proper validation
func validateA11y(data map[string]interface{}) error {
    schema := jsonschema.MustCompile("a11y-schema.json")
    if err := schema.Validate(data); err != nil {
        return fmt.Errorf("a11y validation failed: %w", err)
    }
    return nil
}
```

---

### 10. Unused Error Sentinels

**Anti-pattern**: Error constants never returned

```go
// ❌ Anti-pattern: Error defined but never returned
var (
    ErrMaxRetriesExceeded = errors.New("max retries exceeded")
    ErrVersionMismatch    = errors.New("version mismatch")
)

// No function returns these errors
```

**Detection**:
```bash
# Find error definition
grep -n "var Err" file.go

# Find return sites
grep -r "ErrMaxRetriesExceeded" --include="*.go"
# If only in definition and tests → unused
```

**Fix**: Delete unused error constants

```go
// ✅ Only define errors that are actually returned
var (
    ErrConnectionFailed = errors.New("connection failed")
    ErrTimeout         = errors.New("operation timeout")
)

func Connect() error {
    // ...
    return ErrConnectionFailed  // Actually returned
}
```

---

## Go-Specific Simplification Patterns

### Early Returns (Guard Clauses)

**Before**:
```go
func Process(data []byte) error {
    if len(data) > 0 {
        if isValid(data) {
            if hasPermission() {
                return process(data)
            } else {
                return ErrNoPermission
            }
        } else {
            return ErrInvalidData
        }
    } else {
        return ErrEmptyData
    }
}
```

**After**:
```go
func Process(data []byte) error {
    if len(data) == 0 {
        return ErrEmptyData
    }
    if !isValid(data) {
        return ErrInvalidData
    }
    if !hasPermission() {
        return ErrNoPermission
    }
    return process(data)
}
```

---

### Error Wrapping (Go 1.13+)

**Before**:
```go
func LoadConfig(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("failed to read config: %s", err.Error())
    }
    // ...
}
```

**After**:
```go
func LoadConfig(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("read config: %w", err)
    }
    // ...
}
```

---

### Slice Iteration (Go 1.22+)

**Before**:
```go
for i := 0; i < len(items); i++ {
    process(items[i])
}
```

**After**:
```go
for _, item := range items {
    process(item)
}
```

---

### Map Initialization

**Before**:
```go
m := make(map[string]int)
m["a"] = 1
m["b"] = 2
m["c"] = 3
```

**After**:
```go
m := map[string]int{
    "a": 1,
    "b": 2,
    "c": 3,
}
```

---

### Nil Receiver Checks

**Before**:
```go
func (c *Client) Do() error {
    // Crashes if c is nil
    return c.conn.Send()
}
```

**After**:
```go
func (c *Client) Do() error {
    if c == nil {
        return fmt.Errorf("client is nil")
    }
    return c.conn.Send()
}
```

---

## Detection Checklist

When reviewing Go code, ask:

- [ ] Is this type ever instantiated in production code?
- [ ] Is this registry ever queried (not just written to)?
- [ ] Is this validation function called in the production flow?
- [ ] Does this enum validation duplicate framework constraints?
- [ ] Does this interface have production implementations (not just mocks)?
- [ ] Does this abstraction have multiple consumers?
- [ ] Does this function return real data or placeholders?
- [ ] Does this validation check values or just field existence?
- [ ] Is this error sentinel ever returned by a function?
- [ ] Does this validation reject invalid input or silently fall back?

If any answer is "No", you've found documentation disguised as code.

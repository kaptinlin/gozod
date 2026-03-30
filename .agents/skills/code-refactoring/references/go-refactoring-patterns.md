# Go Refactoring Patterns

Language-specific refactoring patterns for Go 1.26+.

## Naming Refactoring (Industry Alignment)

### Remove Package-Context Redundancy

**Before**
```go
package lint

type AnalyzerConfig struct {
    MaxDepth int
}

type AnalyzerRegistry struct {
    analyzers []Analyzer
}

func NewAnalyzerConfig() *AnalyzerConfig { /* ... */ }
```

**After**
```go
package lint

type Config struct {
    MaxDepth int
}

type Registry struct {
    rules []Rule
}

func NewConfig() *Config { /* ... */ }
```

**Rationale:** Package name already provides context. `lint.Config` is clearer than `lint.AnalyzerConfig`.

### Align with Industry Standard Terminology

**Before**
```go
package lint

type Analyzer interface {
    Analyze(ctx context.Context, docs []*AnalysisDoc) []Issue
}

type AnalysisDoc struct {
    FilePath string
    Content  map[string]any
}
```

**After**
```go
package lint

type Rule interface {
    Check(ctx context.Context, docs []*Document) []Diagnostic
}

type Document struct {
    FilePath string
    Content  map[string]any
}
```

**Rationale:** ESLint, Oxlint, and other linters use "Rule" not "Analyzer". "Check" is more precise than "Analyze" for lint operations. "Diagnostic" is the industry standard term.

### Remove Banned Suffixes

**Before**
```go
package mcp

type SessionManager struct {
    sessions map[string]*Session
}

func NewSessionManager() *SessionManager { /* ... */ }
func (sm *SessionManager) GetSession(id string) *Session { /* ... */ }
```

**After**
```go
package mcp

type Sessions struct {
    sessions map[string]*Session
}

func New() *Sessions { /* ... */ }
func (s *Sessions) Get(id string) *Session { /* ... */ }
```

**Rationale:** "Manager" is a banned suffix (adds no information). Plural noun `Sessions` indicates a collection. Method names shortened when receiver provides context.

### Use Domain-Specific Verbs

**Before**
```go
package validator

type Validator interface {
    Process(data any) error
}

func (v *EmailValidator) Process(data any) error { /* ... */ }
```

**After**
```go
package validator

type Validator interface {
    Validate(data any) error
}

func (v *EmailValidator) Validate(data any) error { /* ... */ }
```

**Rationale:** `Validate()` is more precise than generic `Process()` in the validation domain.

### Shortest Unambiguous Method Names

**Before**
```go
package filter

type ToolFilter struct {
    tools []Tool
}

func (f *ToolFilter) FilterToolsByTier(tier Tier) []Tool { /* ... */ }
func (f *ToolFilter) FilterToolsByReadOnly(readOnly bool) []Tool { /* ... */ }
```

**After**
```go
package filter

type ToolFilter struct {
    tools []Tool
}

func (f *ToolFilter) ByTier(tier Tier) []Tool { /* ... */ }
func (f *ToolFilter) ByReadOnly(readOnly bool) []Tool { /* ... */ }
```

**Rationale:** Receiver already says "ToolFilter", so "Filter" and "Tools" are redundant. `ByTier()` is unambiguous and concise.

### Remove Type-in-Name Redundancy

**Before**
```go
package domains

type ComponentAdapter struct{}
type ExportAdapter struct{}
type StyleAdapter struct{}

func NewComponentAdapter() *ComponentAdapter { /* ... */ }
```

**After**
```go
package domains

type Component struct{}
type Export struct{}
type Style struct{}

func NewComponent() *Component { /* ... */ }
```

**Rationale:** Package context (`domains.Component`) already indicates this is an adapter/domain object. The `Adapter` suffix is redundant.

### Constructor Naming Conventions

**Before**
```go
package handler

func makeServiceHandler[Req, Resp any](
    service ServiceCall[Req, Resp],
) func(context.Context, *Request) (*Response, error) {
    return func(ctx context.Context, req *Request) (*Response, error) {
        // ...
    }
}
```

**After**
```go
package handler

func NewHandler[Req, Resp any](
    service ServiceCall[Req, Resp],
) func(context.Context, *Request) (*Response, error) {
    return func(ctx context.Context, req *Request) (*Response, error) {
        // ...
    }
}
```

**Rationale:** Go convention is `New()` or `NewX()` for constructors, not `makeX()` or `createX()`.

### When NOT to Rename

**Keep verbose names when:**
```go
// 1. Industry standard (even if verbose)
type HTTPClient struct{} // Not "Client" — HTTP is standard prefix

// 2. Stable public API with external consumers
type AnalyzerConfig struct{} // If already published and widely used

// 3. Matches spec/RFC exactly
type OAuth2Token struct{} // OAuth2 is the RFC term

// 4. Verbosity adds clarity in complex domain
type SQLTransactionIsolationLevel int // Fully qualified prevents confusion
```

## DRY (Don't Repeat Yourself) Refactoring

### Extract Repeated Error Handling

**Before**
```go
func ProcessUsers() error {
    users, err := db.GetUsers()
    if err != nil {
        log.Printf("failed to get users: %v", err)
        return fmt.Errorf("get users: %w", err)
    }

    orders, err := db.GetOrders()
    if err != nil {
        log.Printf("failed to get orders: %v", err)
        return fmt.Errorf("get orders: %w", err)
    }

    products, err := db.GetProducts()
    if err != nil {
        log.Printf("failed to get products: %v", err)
        return fmt.Errorf("get products: %w", err)
    }
    // ...
}
```

**After**
```go
func logAndWrap(operation string, err error) error {
    log.Printf("failed to %s: %v", operation, err)
    return fmt.Errorf("%s: %w", operation, err)
}

func ProcessUsers() error {
    users, err := db.GetUsers()
    if err != nil {
        return logAndWrap("get users", err)
    }

    orders, err := db.GetOrders()
    if err != nil {
        return logAndWrap("get orders", err)
    }

    products, err := db.GetProducts()
    if err != nil {
        return logAndWrap("get products", err)
    }
    // ...
}
```

### Use Generics to Eliminate Type Duplication

**Before**
```go
func MaxInt(a, b int) int {
    if a > b {
        return a
    }
    return b
}

func MaxFloat(a, b float64) float64 {
    if a > b {
        return a
    }
    return b
}

func MaxString(a, b string) string {
    if a > b {
        return a
    }
    return b
}
```

**After**
```go
func Max[T constraints.Ordered](a, b T) T {
    if a > b {
        return a
    }
    return b
}
```

### Use Embedding to Share Fields and Methods

**Before**
```go
type UserService struct {
    logger *log.Logger
    db     *sql.DB
}

func (s *UserService) log(msg string) {
    s.logger.Println(msg)
}

type OrderService struct {
    logger *log.Logger
    db     *sql.DB
}

func (s *OrderService) log(msg string) {
    s.logger.Println(msg)
}
```

**After**
```go
type BaseService struct {
    logger *log.Logger
    db     *sql.DB
}

func (s *BaseService) log(msg string) {
    s.logger.Println(msg)
}

type UserService struct {
    BaseService
}

type OrderService struct {
    BaseService
}
```

## SRP (Single Responsibility Principle) Refactoring

### Split God Struct

**Before**
```go
type UserManager struct {
    db     *sql.DB
    cache  *redis.Client
    mailer *smtp.Client
}

func (m *UserManager) CreateUser(user User) error {
    // validate
    // save to db
    // update cache
    // send welcome email
}

func (m *UserManager) DeleteUser(id int) error {
    // delete from db
    // invalidate cache
    // send goodbye email
}

func (m *UserManager) SendPasswordReset(email string) error {
    // find user
    // generate token
    // send email
}
```

**After**
```go
type UserRepository struct {
    db *sql.DB
}

func (r *UserRepository) Create(user User) error { /* ... */ }
func (r *UserRepository) Delete(id int) error { /* ... */ }
func (r *UserRepository) FindByEmail(email string) (*User, error) { /* ... */ }

type UserCache struct {
    client *redis.Client
}

func (c *UserCache) Set(user User) error { /* ... */ }
func (c *UserCache) Invalidate(id int) error { /* ... */ }

type UserMailer struct {
    client *smtp.Client
}

func (m *UserMailer) SendWelcome(user User) error { /* ... */ }
func (m *UserMailer) SendGoodbye(user User) error { /* ... */ }
func (m *UserMailer) SendPasswordReset(user User, token string) error { /* ... */ }

type UserService struct {
    repo   *UserRepository
    cache  *UserCache
    mailer *UserMailer
}

func (s *UserService) CreateUser(user User) error {
    if err := s.repo.Create(user); err != nil {
        return err
    }
    if err := s.cache.Set(user); err != nil {
        log.Printf("cache error: %v", err)
    }
    if err := s.mailer.SendWelcome(user); err != nil {
        log.Printf("email error: %v", err)
    }
    return nil
}
```

### Separate I/O and Business Logic

**Before**
```go
func ProcessOrder(orderID int) error {
    // fetch from database
    row := db.QueryRow("SELECT * FROM orders WHERE id = ?", orderID)
    var order Order
    if err := row.Scan(&order.ID, &order.Amount, &order.Status); err != nil {
        return err
    }

    // business logic
    if order.Amount > 1000 {
        order.Status = "approved"
    } else {
        order.Status = "pending"
    }

    // save to database
    _, err := db.Exec("UPDATE orders SET status = ? WHERE id = ?", order.Status, order.ID)
    return err
}
```

**After**
```go
type OrderRepository interface {
    Get(id int) (*Order, error)
    Update(order *Order) error
}

func ApproveOrder(order *Order) {
    if order.Amount > 1000 {
        order.Status = "approved"
    } else {
        order.Status = "pending"
    }
}

func ProcessOrder(repo OrderRepository, orderID int) error {
    order, err := repo.Get(orderID)
    if err != nil {
        return err
    }

    ApproveOrder(order)

    return repo.Update(order)
}
```

### Use Interface Segregation

**Before**
```go
type Storage interface {
    Save(data []byte) error
    Load() ([]byte, error)
    Delete() error
    List() ([]string, error)
    Backup() error
    Restore() error
}

type Cache struct{}

func (c *Cache) Save(data []byte) error { /* ... */ }
func (c *Cache) Load() ([]byte, error) { /* ... */ }
func (c *Cache) Delete() error { /* ... */ }
func (c *Cache) List() ([]string, error) { return nil, errors.New("not supported") }
func (c *Cache) Backup() error { return errors.New("not supported") }
func (c *Cache) Restore() error { return errors.New("not supported") }
```

**After**
```go
type Reader interface {
    Load() ([]byte, error)
}

type Writer interface {
    Save(data []byte) error
}

type Deleter interface {
    Delete() error
}

type Lister interface {
    List() ([]string, error)
}

type Cache struct{}

func (c *Cache) Save(data []byte) error { /* ... */ }
func (c *Cache) Load() ([]byte, error) { /* ... */ }
func (c *Cache) Delete() error { /* ... */ }
```

## OCP (Open/Closed Principle) Refactoring

### Use Strategy Pattern Instead of Switch

**Before**
```go
func CalculatePrice(productType string, basePrice float64) float64 {
    switch productType {
    case "book":
        return basePrice * 0.9
    case "electronics":
        return basePrice * 0.95
    case "clothing":
        return basePrice * 0.8
    default:
        return basePrice
    }
}
```

**After**
```go
type PricingStrategy interface {
    Calculate(basePrice float64) float64
}

type BookPricing struct{}

func (p BookPricing) Calculate(basePrice float64) float64 {
    return basePrice * 0.9
}

type ElectronicsPricing struct{}

func (p ElectronicsPricing) Calculate(basePrice float64) float64 {
    return basePrice * 0.95
}

var strategies = map[string]PricingStrategy{
    "book":        BookPricing{},
    "electronics": ElectronicsPricing{},
    "clothing":    ClothingPricing{},
}

func CalculatePrice(productType string, basePrice float64) float64 {
    strategy, ok := strategies[productType]
    if !ok {
        return basePrice
    }
    return strategy.Calculate(basePrice)
}
```

### Use Registry Pattern

**Before**
```go
func ProcessMessage(msgType string, data []byte) error {
    switch msgType {
    case "email":
        return processEmail(data)
    case "sms":
        return processSMS(data)
    case "push":
        return processPush(data)
    default:
        return errors.New("unknown message type")
    }
}
```

**After**
```go
type MessageProcessor interface {
    Process(data []byte) error
}

var processors = make(map[string]MessageProcessor)

func RegisterProcessor(msgType string, processor MessageProcessor) {
    processors[msgType] = processor
}

func ProcessMessage(msgType string, data []byte) error {
    processor, ok := processors[msgType]
    if !ok {
        return fmt.Errorf("unknown message type: %s", msgType)
    }
    return processor.Process(data)
}

// In init or main
func init() {
    RegisterProcessor("email", &EmailProcessor{})
    RegisterProcessor("sms", &SMSProcessor{})
    RegisterProcessor("push", &PushProcessor{})
}
```

### Use Plugin Architecture

**Before**
```go
// Hardcoded handlers
func HandleRequest(r *http.Request) {
    authenticate(r)
    logRequest(r)
    rateLimit(r)
    // ... more middleware
}
```

**After**
```go
type Middleware func(http.Handler) http.Handler

func Chain(h http.Handler, middlewares ...Middleware) http.Handler {
    for i := len(middlewares) - 1; i >= 0; i-- {
        h = middlewares[i](h)
    }
    return h
}

func Authenticate(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // auth logic
        next.ServeHTTP(w, r)
    })
}

func LogRequest(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // logging logic
        next.ServeHTTP(w, r)
    })
}

// Usage
handler := Chain(
    myHandler,
    Authenticate,
    LogRequest,
    RateLimit,
)
```

## Go-Specific Refactoring

### Use Embedding Instead of Inheritance

**Before (trying to simulate inheritance)**
```go
type Animal struct {
    Name string
}

type Dog struct {
    Animal Animal
}

func (d *Dog) GetName() string {
    return d.Animal.Name
}
```

**After (using embedding)**
```go
type Animal struct {
    Name string
}

type Dog struct {
    Animal
}

// Dog automatically has Name field and any Animal methods
```

### Use Interface Composition

**Before**
```go
type ReadWriteCloser interface {
    Read(p []byte) (n int, err error)
    Write(p []byte) (n int, err error)
    Close() error
}
```

**After**
```go
type ReadWriteCloser interface {
    io.Reader
    io.Writer
    io.Closer
}
```

### Use Functional Options

**Before**
```go
type Server struct {
    addr     string
    timeout  time.Duration
    maxConns int
    logger   *log.Logger
}

func NewServer(addr string, timeout time.Duration, maxConns int, logger *log.Logger) *Server {
    return &Server{
        addr:     addr,
        timeout:  timeout,
        maxConns: maxConns,
        logger:   logger,
    }
}
```

**After**
```go
type Server struct {
    addr     string
    timeout  time.Duration
    maxConns int
    logger   *log.Logger
}

type Option func(*Server)

func WithTimeout(d time.Duration) Option {
    return func(s *Server) { s.timeout = d }
}

func WithMaxConns(n int) Option {
    return func(s *Server) { s.maxConns = n }
}

func WithLogger(l *log.Logger) Option {
    return func(s *Server) { s.logger = l }
}

func NewServer(addr string, opts ...Option) *Server {
    s := &Server{
        addr:     addr,
        timeout:  30 * time.Second,
        maxConns: 100,
        logger:   log.Default(),
    }
    for _, opt := range opts {
        opt(s)
    }
    return s
}

// Usage
server := NewServer(":8080",
    WithTimeout(60*time.Second),
    WithMaxConns(200))
```

## References

- [Refactoring Guru](https://refactoring.guru/)
- [Effective Go](https://go.dev/doc/effective_go)
- [Go Proverbs](https://go-proverbs.github.io/)

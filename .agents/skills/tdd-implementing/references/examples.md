---
name: golang-tdd-examples
description: Concrete Go TDD implementation examples with testify
---

# Go TDD Examples

## Example 1: Pure Function — Vertical RED→GREEN

### RED

```go
func TestAdd_PositiveNumbers(t *testing.T) {
    result := Add(2, 3)
    assert.Equal(t, 5, result)
}
```

Run: `FAIL: undefined: Add`

### GREEN

```go
func Add(a, b int) int {
    return a + b
}
```

Run: `PASS`

## Example 2: Error Handling with Sentinel Errors

### RED

```go
func TestParseToken_EmptyString_ReturnsError(t *testing.T) {
    _, err := ParseToken("")
    require.Error(t, err)
    assert.ErrorIs(t, err, ErrInvalid)
}
```

Run: `FAIL: undefined: ParseToken`

### GREEN

```go
var ErrInvalid = errors.New("invalid token")

func ParseToken(raw string) (Token, error) {
    if raw == "" {
        return Token{}, ErrInvalid
    }
    // minimal implementation
    return Token{Raw: raw}, nil
}
```

Run: `PASS`

## Example 3: Table-Driven Validation

### RED

```go
func TestValidateEmail_InvalidFormats(t *testing.T) {
    tests := []struct {
        name  string
        email string
    }{
        {"empty", ""},
        {"no @", "userdomain.com"},
        {"missing domain", "user@"},
        {"missing user", "@domain.com"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateEmail(tt.email)
            assert.ErrorIs(t, err, ErrInvalidEmail, "input: %q", tt.email)
        })
    }
}
```

Run: `FAIL: undefined: ValidateEmail`

### GREEN

```go
var ErrInvalidEmail = errors.New("invalid email")

func ValidateEmail(email string) error {
    parts := strings.SplitN(email, "@", 2)
    if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
        return ErrInvalidEmail
    }
    return nil
}
```

Run: `PASS`

## Example 4: Interface + Mock — Service Layer

Testing a service that depends on a repository interface.

### Define the interface and mock

```go
// Repository interface (production code)
type UserRepository interface {
    GetByID(ctx context.Context, id string) (*User, error)
    Save(ctx context.Context, user *User) error
}

// Mock (test code)
type MockUserRepo struct {
    mock.Mock
}

func (m *MockUserRepo) GetByID(ctx context.Context, id string) (*User, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*User), args.Error(1)
}

func (m *MockUserRepo) Save(ctx context.Context, user *User) error {
    args := m.Called(ctx, user)
    return args.Error(0)
}
```

### RED — happy path

```go
func TestUserService_GetUser_Success(t *testing.T) {
    repo := new(MockUserRepo)
    svc := NewUserService(repo)

    expected := &User{ID: "123", Name: "Alice"}
    repo.On("GetByID", mock.Anything, "123").Return(expected, nil)

    user, err := svc.GetUser(context.Background(), "123")
    require.NoError(t, err)
    assert.Equal(t, "Alice", user.Name)

    repo.AssertExpectations(t)
}
```

Run: `FAIL: undefined: UserService`

### GREEN — happy path

```go
type UserService struct {
    repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
    return &UserService{repo: repo}
}

func (s *UserService) GetUser(ctx context.Context, id string) (*User, error) {
    return s.repo.GetByID(ctx, id)
}
```

Run: `PASS`

### RED — not found error

```go
func TestUserService_GetUser_NotFound(t *testing.T) {
    repo := new(MockUserRepo)
    svc := NewUserService(repo)

    repo.On("GetByID", mock.Anything, "999").Return(nil, ErrNotFound)

    _, err := svc.GetUser(context.Background(), "999")
    assert.ErrorIs(t, err, ErrNotFound)

    repo.AssertExpectations(t)
}
```

Run: `PASS` (already works — repo error propagates)

This is fine: the test documents the behavior even though no new code is needed.

## Example 5: HTTP Handler with Mock Service

### RED

```go
func TestHandler_GetUser_Success(t *testing.T) {
    svc := new(MockUserService)
    handler := NewHandler(svc)

    svc.On("GetUser", mock.Anything, "123").Return(&User{
        ID: "123", Name: "Alice",
    }, nil)

    req := httptest.NewRequest("GET", "/users/123", nil)
    w := httptest.NewRecorder()

    handler.ServeHTTP(w, req)

    assert.Equal(t, http.StatusOK, w.Code)

    var resp User
    require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
    assert.Equal(t, "Alice", resp.Name)

    svc.AssertExpectations(t)
}

func TestHandler_GetUser_NotFound(t *testing.T) {
    svc := new(MockUserService)
    handler := NewHandler(svc)

    svc.On("GetUser", mock.Anything, "999").Return(nil, ErrNotFound)

    req := httptest.NewRequest("GET", "/users/999", nil)
    w := httptest.NewRecorder()

    handler.ServeHTTP(w, req)

    assert.Equal(t, http.StatusNotFound, w.Code)
    svc.AssertExpectations(t)
}
```

### GREEN

```go
type Handler struct {
    svc UserServicer
}

func NewHandler(svc UserServicer) *Handler {
    return &Handler{svc: svc}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    id := path.Base(r.URL.Path)

    user, err := h.svc.GetUser(r.Context(), id)
    if err != nil {
        if errors.Is(err, ErrNotFound) {
            http.Error(w, "not found", http.StatusNotFound)
            return
        }
        http.Error(w, "internal error", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(user)
}
```

## Example 6: Mock with httptest.NewServer (External API)

### RED

```go
func TestAPIClient_FetchData_Success(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        assert.Equal(t, "/api/data", r.URL.Path)
        assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]string{"result": "ok"})
    }))
    t.Cleanup(server.Close)

    client := NewAPIClient(server.URL, "test-token")
    result, err := client.FetchData(context.Background())

    require.NoError(t, err)
    assert.Equal(t, "ok", result.Result)
}

func TestAPIClient_FetchData_ServerError_ReturnsError(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusInternalServerError)
    }))
    t.Cleanup(server.Close)

    client := NewAPIClient(server.URL, "test-token")
    _, err := client.FetchData(context.Background())

    require.Error(t, err)
    assert.ErrorContains(t, err, "unexpected status: 500")
}
```

### GREEN

```go
type APIClient struct {
    baseURL string
    token   string
    client  *http.Client
}

func NewAPIClient(baseURL, token string) *APIClient {
    return &APIClient{baseURL: baseURL, token: token, client: &http.Client{}}
}

func (c *APIClient) FetchData(ctx context.Context) (*Response, error) {
    req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/data", nil)
    if err != nil {
        return nil, fmt.Errorf("create request: %w", err)
    }
    req.Header.Set("Authorization", "Bearer "+c.token)

    resp, err := c.client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("do request: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
    }

    var result Response
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, fmt.Errorf("decode response: %w", err)
    }
    return &result, nil
}
```

## Example 7: Mock with Custom Matcher + Side Effect

### RED

```go
func TestService_CreateUser_AssignsID(t *testing.T) {
    repo := new(MockUserRepo)
    svc := NewUserService(repo)

    // Verify Save receives a user with non-empty fields
    repo.On("Save", mock.Anything, mock.MatchedBy(func(u *User) bool {
        return u.Name == "Alice" && u.Email == "alice@example.com"
    })).Run(func(args mock.Arguments) {
        // Simulate DB assigning an ID
        user := args.Get(1).(*User)
        user.ID = "generated-123"
    }).Return(nil)

    user, err := svc.CreateUser(context.Background(), "Alice", "alice@example.com")

    require.NoError(t, err)
    assert.Equal(t, "generated-123", user.ID)
    assert.Equal(t, "Alice", user.Name)
    repo.AssertExpectations(t)
}
```

### GREEN

```go
func (s *UserService) CreateUser(ctx context.Context, name, email string) (*User, error) {
    user := &User{Name: name, Email: email}
    if err := s.repo.Save(ctx, user); err != nil {
        return nil, fmt.Errorf("save user: %w", err)
    }
    return user, nil
}
```

## Example 8: Context Cancellation

### RED

```go
func TestWorker_ContextCancel_StopsGracefully(t *testing.T) {
    ctx, cancel := context.WithCancel(context.Background())
    worker := NewWorker()

    done := make(chan struct{})
    go func() {
        worker.Run(ctx)
        close(done)
    }()

    cancel()

    select {
    case <-done:
        // success: worker exited
    case <-time.After(time.Second):
        t.Fatal("worker did not stop after context cancellation")
    }
}
```

### GREEN

```go
type Worker struct{}

func NewWorker() *Worker { return &Worker{} }

func (w *Worker) Run(ctx context.Context) {
    for {
        select {
        case <-ctx.Done():
            return
        default:
            time.Sleep(10 * time.Millisecond)
        }
    }
}
```

## Example 9: Typed Error Assertion

### RED

```go
func TestParse_InvalidInput_ReturnsValidationError(t *testing.T) {
    _, err := Parse("bad-input")
    require.Error(t, err)

    var valErr *ValidationError
    require.ErrorAs(t, err, &valErr)
    assert.Equal(t, "format", valErr.Field)
    assert.Equal(t, "bad-input", valErr.Value)
}
```

### GREEN

```go
type ValidationError struct {
    Field string
    Value string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation failed: field=%s value=%s", e.Field, e.Value)
}

func Parse(input string) (Result, error) {
    if !isValidFormat(input) {
        return Result{}, &ValidationError{Field: "format", Value: input}
    }
    return Result{}, nil
}
```

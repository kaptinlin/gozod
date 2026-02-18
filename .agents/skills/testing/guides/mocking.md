# Mocking

## Strategy Selection

| Approach | When to Use |
|----------|------------|
| **Interface + hand-written mock** | Small interface (1-3 methods), need full control, most common |
| **`testify/mock`** | Large interface, need call verification (times, order) |
| **Fake implementation** | Complex behavior needed (in-memory database, filesystem) |
| **No mock** | Pure functions, value objects, code with no external dependencies |

**Default choice: hand-written mocks.** They are explicit, type-safe, and easy to understand.

---

## Interface-Based Mocking

### Design for testability

Define small interfaces at the consumer side, not the provider side:

```go
// In the package that USES the dependency, not the one that implements it
type UserStore interface {
    Get(ctx context.Context, id string) (*User, error)
    Save(ctx context.Context, u *User) error
}

type OrderService struct {
    users UserStore  // Depends on interface, not concrete type
}

func NewOrderService(users UserStore) *OrderService {
    return &OrderService{users: users}
}
```

### Hand-written mock

```go
// In test file
type mockUserStore struct {
    getFunc  func(ctx context.Context, id string) (*User, error)
    saveFunc func(ctx context.Context, u *User) error
}

func (m *mockUserStore) Get(ctx context.Context, id string) (*User, error) {
    return m.getFunc(ctx, id)
}

func (m *mockUserStore) Save(ctx context.Context, u *User) error {
    return m.saveFunc(ctx, u)
}
```

### Using the mock

```go
func TestOrderService_PlaceOrder(t *testing.T) {
    t.Parallel()

    t.Run("success", func(t *testing.T) {
        t.Parallel()

        store := &mockUserStore{
            getFunc: func(_ context.Context, id string) (*User, error) {
                return &User{ID: id, Balance: 1000}, nil
            },
            saveFunc: func(_ context.Context, u *User) error {
                assert.Equal(t, 900, u.Balance) // Verify side effect
                return nil
            },
        }

        svc := NewOrderService(store)
        err := svc.PlaceOrder(t.Context(), "user-1", 100)
        require.NoError(t, err)
    })

    t.Run("user not found", func(t *testing.T) {
        t.Parallel()

        store := &mockUserStore{
            getFunc: func(_ context.Context, _ string) (*User, error) {
                return nil, ErrNotFound
            },
        }

        svc := NewOrderService(store)
        err := svc.PlaceOrder(t.Context(), "unknown", 100)
        assert.ErrorIs(t, err, ErrNotFound)
    })
}
```

---

## Shortcut: Single-Method Mock

For interfaces with one method, use a function type:

```go
// Interface
type Hasher interface {
    Hash(data []byte) (string, error)
}

// Function adapter
type HasherFunc func([]byte) (string, error)

func (f HasherFunc) Hash(data []byte) (string, error) { return f(data) }

// In test
svc := NewService(HasherFunc(func(data []byte) (string, error) {
    return "mock-hash", nil
}))
```

---

## Error Injection

Test error handling paths by returning errors from mocks:

```go
func TestService_HandlesDatabaseError(t *testing.T) {
    t.Parallel()

    store := &mockUserStore{
        getFunc: func(_ context.Context, _ string) (*User, error) {
            return nil, errors.New("connection refused")
        },
    }

    svc := NewOrderService(store)
    err := svc.PlaceOrder(t.Context(), "user-1", 100)
    assert.Error(t, err)
    assert.ErrorContains(t, err, "connection refused")
}
```

Use `assert.AnError` as a convenient sentinel for tests that only check whether an error occurred:

```go
store := &mockUserStore{
    saveFunc: func(_ context.Context, _ *User) error {
        return assert.AnError
    },
}
```

---

## testify/mock

Use for large interfaces or when you need call count / call order verification.

```go
import "github.com/stretchr/testify/mock"

type MockStore struct {
    mock.Mock
}

func (m *MockStore) Get(ctx context.Context, id string) (*User, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*User), args.Error(1)
}

func (m *MockStore) Save(ctx context.Context, u *User) error {
    args := m.Called(ctx, u)
    return args.Error(0)
}
```

### Using testify/mock

```go
func TestWithMock(t *testing.T) {
    t.Parallel()

    store := new(MockStore)

    // Set expectations
    store.On("Get", mock.Anything, "user-1").
        Return(&User{ID: "user-1", Balance: 1000}, nil)
    store.On("Save", mock.Anything, mock.AnythingOfType("*User")).
        Return(nil)

    svc := NewOrderService(store)
    err := svc.PlaceOrder(t.Context(), "user-1", 100)
    require.NoError(t, err)

    // Verify all expectations were met
    store.AssertExpectations(t)

    // Verify specific call counts
    store.AssertNumberOfCalls(t, "Get", 1)
    store.AssertNumberOfCalls(t, "Save", 1)
}
```

### When to prefer testify/mock over hand-written

- Interface has 5+ methods and you only care about 1-2 in each test
- You need to verify call count, call order, or argument capture
- Multiple tests need different mock configurations for the same large interface

### When to prefer hand-written mocks

- Interface has 1-3 methods (most Go interfaces)
- You want compile-time type safety
- You want to read the mock behavior at the test call site
- You want to avoid `mock.Anything` / `mock.AnythingOfType` stringly-typed matching

---

## Fake Implementations

For complex dependencies where both mocking and stubbing are insufficient:

```go
// In-memory store as fake
type fakeStore struct {
    mu   sync.Mutex
    data map[string]*User
}

func newFakeStore() *fakeStore {
    return &fakeStore{data: make(map[string]*User)}
}

func (f *fakeStore) Get(_ context.Context, id string) (*User, error) {
    f.mu.Lock()
    defer f.mu.Unlock()
    u, ok := f.data[id]
    if !ok {
        return nil, ErrNotFound
    }
    return u, nil
}

func (f *fakeStore) Save(_ context.Context, u *User) error {
    f.mu.Lock()
    defer f.mu.Unlock()
    f.data[u.ID] = u
    return nil
}
```

Use fakes when:
- Tests need realistic multi-step interactions (create, read, update, delete)
- Mock configuration would be more complex than the actual implementation
- Multiple tests share the same complex setup

---

## Testing with context.Context

Always pass `t.Context()` through mocks to detect context cancellation issues:

```go
func TestService_RespectsContextCancellation(t *testing.T) {
    t.Parallel()

    store := &mockUserStore{
        getFunc: func(ctx context.Context, _ string) (*User, error) {
            select {
            case <-ctx.Done():
                return nil, ctx.Err()
            default:
                return &User{ID: "1"}, nil
            }
        },
    }

    ctx, cancel := context.WithCancel(t.Context())
    cancel() // Cancel immediately

    svc := NewOrderService(store)
    err := svc.PlaceOrder(ctx, "user-1", 100)
    assert.ErrorIs(t, err, context.Canceled)
}
```

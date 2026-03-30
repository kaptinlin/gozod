# When to Mock in Go

## Rule: Mock at System Boundaries Only

Mock only things you **don't control**:

| Mock | Don't Mock |
|------|------------|
| External HTTP APIs | Your own packages |
| Email / SMS services | Internal structs |
| Time (`time.Now()`) | Pure functions |
| Random number generation | Anything you can run in tests |
| File system (when I/O matters) | |

Prefer a real in-memory implementation over a mock for things you control.

## Design for Testability

Accept dependencies as parameters — don't instantiate them internally:

```go
// Hard to test — creates its own dependency
func SendWelcomeEmail(userID string) error {
    client := smtp.NewClient(os.Getenv("SMTP_HOST"))
    return client.Send(...)
}

// Easy to test — dependency injected
func SendWelcomeEmail(mailer Mailer, userID string) error {
    return mailer.Send(...)
}
```

## Interface-Based Mocks

Define narrow interfaces at the boundary, not wide ones:

```go
// GOOD: Narrow interface — only what this code needs
type Mailer interface {
    Send(to, subject, body string) error
}

// BAD: Wide interface mirroring the whole SDK
type EmailClient interface {
    Send(...) error
    SendBatch(...) error
    GetStatus(...) error
    Unsubscribe(...) error
}
```

### Hand-Written Stub (preferred for simple cases)

```go
type stubMailer struct {
    sent []string // capture sent addresses for assertion
    err  error    // inject error to test error path
}

func (m *stubMailer) Send(to, subject, body string) error {
    m.sent = append(m.sent, to)
    return m.err
}

func TestWelcomeEmail_SendsToCorrectAddress(t *testing.T) {
    mailer := &stubMailer{}
    err := SendWelcomeEmail(mailer, "alice@example.com")

    require.NoError(t, err)
    assert.Equal(t, []string{"alice@example.com"}, mailer.sent)
}

func TestWelcomeEmail_ReturnsMailerError(t *testing.T) {
    mailer := &stubMailer{err: errors.New("smtp timeout")}
    err := SendWelcomeEmail(mailer, "alice@example.com")

    assert.Error(t, err)
}
```

### testify/mock (for external boundaries with many methods)

```go
import "github.com/stretchr/testify/mock"

type MockPaymentGateway struct {
    mock.Mock
}

func (m *MockPaymentGateway) Charge(amount int, token string) (string, error) {
    args := m.Called(amount, token)
    return args.String(0), args.Error(1)
}

func TestCheckout_ChargesCorrectAmount(t *testing.T) {
    gateway := new(MockPaymentGateway)
    gateway.On("Charge", 1000, "tok_123").Return("charge_abc", nil)

    order, err := Checkout(gateway, cart)

    require.NoError(t, err)
    assert.Equal(t, "charge_abc", order.ChargeID)
    gateway.AssertExpectations(t) // verify Charge was called
}
```

Use `testify/mock` only when you need to assert that an external side-effecting call happened (e.g., payment charged). Not for your own internal logic.

## Time

```go
// Inject time as a dependency
type Clock interface {
    Now() time.Time
}

type realClock struct{}
func (realClock) Now() time.Time { return time.Now() }

type fixedClock struct{ t time.Time }
func (c fixedClock) Now() time.Time { return c.t }

// In tests:
clock := fixedClock{t: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)}
result := ComputeExpiry(clock, duration)
assert.Equal(t, expectedExpiry, result)
```

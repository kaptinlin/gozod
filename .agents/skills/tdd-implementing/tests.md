# Good and Bad Tests in Go

## Good Tests

**Integration-style**: exercise real code paths through exported interfaces.

```go
// GOOD: Tests observable behavior through the public API
func TestCart_CheckoutConfirmsOrder(t *testing.T) {
    cart := NewCart()
    cart.Add(product)

    order, err := cart.Checkout(paymentMethod)

    require.NoError(t, err)
    assert.Equal(t, StatusConfirmed, order.Status)
}
```

Characteristics:
- Tests what callers care about
- Uses only exported types and functions
- Survives internal refactors
- Describes WHAT, not HOW
- Test name: `TestSubject_Behavior` or `TestSubject_Condition_Behavior`

## Bad Tests

**Implementation-detail tests**: coupled to internal structure.

```go
// BAD: Asserts on internal call — breaks on rename/refactor
func TestCheckout_CallsPaymentProcessor(t *testing.T) {
    mockProc := &mockPaymentProcessor{}
    cart := newCartWithProcessor(mockProc)
    cart.Add(product)

    cart.Checkout(paymentMethod)

    assert.Equal(t, 1, mockProc.processCallCount) // testing HOW, not WHAT
}
```

Red flags:
- Asserting on call counts or call order
- Mocking internal collaborators (things you own)
- Test name describes a mechanism ("calls X", "invokes Y", "sets field Z")
- Test breaks when you rename an internal function but behavior didn't change

### Side-Channel Verification

```go
// BAD: Verifies through DB query instead of the interface
func TestCreateUser_SavesToDB(t *testing.T) {
    err := CreateUser(ctx, db, "Alice")
    require.NoError(t, err)

    row := db.QueryRow("SELECT name FROM users WHERE name = ?", "Alice")
    var name string
    row.Scan(&name)
    assert.Equal(t, "Alice", name) // bypasses the interface
}

// GOOD: Verifies through the interface
func TestCreateUser_MakesUserRetrievable(t *testing.T) {
    err := CreateUser(ctx, db, "Alice")
    require.NoError(t, err)

    user, err := GetUser(ctx, db, "Alice")
    require.NoError(t, err)
    assert.Equal(t, "Alice", user.Name)
}
```

## Error Testing

```go
// GOOD: Test error behavior through the interface
func TestDivide_ReturnsErrorOnZeroDivisor(t *testing.T) {
    _, err := Divide(10, 0)

    require.Error(t, err)
    assert.ErrorIs(t, err, ErrDivisionByZero) // use errors.Is chain, not string match
}
```

Use `require.Error` / `assert.ErrorIs` / `assert.ErrorAs` — not `assert.Contains(t, err.Error(), "some string")`.

## Table-Driven Tests

Use when the same behavior must hold across many inputs:

```go
func TestValidateEmail(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {"valid email", "user@example.com", false},
        {"missing @", "userexample.com", true},
        {"empty string", "", true},
        {"local only", "user@", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateEmail(tt.input)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

Don't use table-driven tests when each case needs different assertions — write separate focused tests instead.

## Test Helpers

```go
// Always call t.Helper() so failure lines point to the caller, not the helper
func requireUserExists(t *testing.T, repo UserRepo, id string) *User {
    t.Helper()
    user, err := repo.Get(ctx, id)
    require.NoError(t, err, "user %s should exist", id)
    return user
}
```

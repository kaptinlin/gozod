# Testing Anti-Patterns in Go

**Load when:** writing or changing tests, adding mocks, or tempted to add test-only methods to production code.

## Core Principle

Test what the code does, not what the mocks do.

**Following strict TDD prevents these anti-patterns** — if you watched the test fail against real code before mocking, you know what's actually being tested.

## The Iron Laws

```
1. NEVER test mock behavior
2. NEVER add test-only methods to production code
3. NEVER mock without understanding dependencies
```

## Anti-Pattern 1: Testing Mock Behavior

**The violation:**

```go
// ❌ BAD: asserting the mock was called, not that behavior happened
func TestOrder_Submit(t *testing.T) {
    mailer := new(MockMailer)
    mailer.On("Send", mock.Anything).Return(nil)

    order.Submit(mailer)

    mailer.AssertCalled(t, "Send", mock.Anything) // testing the mock, not the order
}
```

**Why this is wrong:** You're verifying the mock exists and was called, not that the order was submitted correctly. The test passes even if the email address is wrong.

**The fix:**

```go
// ✅ GOOD: assert on real observable behavior
func TestOrder_Submit_SendsConfirmationToCustomer(t *testing.T) {
    mailer := &stubMailer{}

    order.Submit(mailer)

    require.Len(t, mailer.sent, 1)
    assert.Equal(t, "customer@example.com", mailer.sent[0])
}
```

**Gate function:** Before asserting on a mock: "Am I testing real behavior or just mock existence?" If mock existence — delete the assertion, assert on real output instead.

## Anti-Pattern 2: Test-Only Methods in Production

**The violation:**

```go
// ❌ BAD: Reset() only called in tests
type Cache struct {
    mu    sync.Mutex
    items map[string]any
}

func (c *Cache) Reset() { // test-only pollution
    c.mu.Lock()
    defer c.mu.Unlock()
    c.items = make(map[string]any)
}

// In tests:
afterEach: cache.Reset()
```

**Why this is wrong:** Production code polluted with test-only logic. Dangerous if called in production.

**The fix:**

```go
// ✅ GOOD: each test creates its own isolated instance
func TestCache_Get_ReturnsStoredValue(t *testing.T) {
    t.Parallel()
    cache := NewCache() // fresh per test, no Reset needed
    cache.Set("k", "v")
    assert.Equal(t, "v", cache.Get("k"))
}
```

**Gate function:** Before adding any method: "Is this only called by tests?" If yes — don't add it. Use a fresh instance per test or move cleanup to test helpers.

## Anti-Pattern 3: Mocking Without Understanding

**The violation:**

```go
// ❌ BAD: mocking a method whose side effects the test depends on
func TestServer_RejectsDuplicate(t *testing.T) {
    store := new(MockStore)
    store.On("Save", mock.Anything).Return(nil) // mocked Save skips the write!

    AddServer(store, cfg)
    err := AddServer(store, cfg) // should return ErrDuplicate — but won't

    assert.ErrorIs(t, err, ErrDuplicate)
}
```

**Why this is wrong:** The mocked `Save` skips writing the config, so the second call can't detect the duplicate. Test passes for the wrong reason or fails mysteriously.

**The fix:**

```go
// ✅ GOOD: use a real in-memory store; only mock the slow external boundary
func TestServer_RejectsDuplicate(t *testing.T) {
    store := NewInMemoryStore() // real, fast, no external I/O

    AddServer(store, cfg)
    err := AddServer(store, cfg)

    assert.ErrorIs(t, err, ErrDuplicate)
}
```

**Gate function:** Before mocking any method:
1. What side effects does the real method have?
2. Does this test depend on any of those side effects?
3. If yes — mock at a lower level or use a real in-memory implementation.

## Anti-Pattern 4: Incomplete Mocks

**The violation:**

```go
// ❌ BAD: partial response struct — only fields you think you need
type fakeResponse struct {
    Status string
    UserID string
}
// Downstream code accesses .Metadata.RequestID — silent nil panic
```

**Why this is wrong:** Tests pass but integration fails. You only included fields you remembered.

**The fix:**

```go
// ✅ GOOD: mirror the real struct completely
type APIResponse struct {
    Status   string
    UserID   string
    Metadata ResponseMetadata // include all fields
}
```

**Gate function:** Before creating a stub response — check the actual type definition and include all fields the system might access.

## Quick Reference

| Anti-Pattern | Fix |
|--------------|-----|
| Assert on mock call count for behavior | Assert on real observable output |
| Test-only methods in production types | Fresh instance per test or test-helper functions |
| Mock without understanding side effects | Use real in-memory impl; mock only slow I/O |
| Partial stub structs | Mirror real struct completely |

## Red Flags

- `AssertCalled` is the only assertion in the test
- Methods only called from test files
- Mock setup is longer than the assertion block
- Test breaks when you remove the mock
- Can't explain why a particular dependency is mocked

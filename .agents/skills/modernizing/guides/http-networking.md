# HTTP & Networking

Modern Go's `net/http` package gained enhanced routing with methods and wildcards, per-request controls, and CSRF protection — reducing the need for third-party routers.

## Contents
- Enhanced routing patterns (1.22+)
- http.ResponseController (1.20+)
- http.CrossOriginProtection (1.25+)
- net.KeepAliveConfig (1.23+)
- ReverseProxy.Rewrite (1.20+)
- math/rand/v2 (1.22+)

---

## Enhanced routing patterns (Go 1.22+)

`http.ServeMux` now supports HTTP methods and path wildcards directly.

### When to use
- New projects that don't need middleware chaining, route grouping, or regex matching
- Simple REST APIs where method + path matching is sufficient
- Replacing manual method checking inside handlers

### When NOT to use
- When you need middleware chaining (use chi, echo, or similar)
- When you need route grouping with shared middleware
- When you need regex path matching or custom route constraints
- When you need named route generation (reverse routing)
- Existing projects already using a third-party router — don't migrate for the sake of it

```go
// Old — manual method check
mux.HandleFunc("/items/", func(w http.ResponseWriter, r *http.Request) {
    if r.Method != "GET" {
        http.Error(w, "method not allowed", 405)
        return
    }
    id := strings.TrimPrefix(r.URL.Path, "/items/")
    // ...
})

// New (Go 1.22+) — method and wildcard in pattern
mux.HandleFunc("GET /items/{id}", func(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")
    // ...
})

// Multiple methods
mux.HandleFunc("POST /items", createItem)
mux.HandleFunc("DELETE /items/{id}", deleteItem)

// Wildcard matches rest of path
mux.HandleFunc("GET /files/{path...}", serveFile)

// Host-specific routes
mux.HandleFunc("api.example.com/v1/{$}", apiRoot) // {$} matches exact path
```

### Routing precedence

Most specific pattern wins, regardless of registration order:
- `/posts/latest` beats `/posts/{id}` (literal beats wildcard)
- `GET /posts/{id}` beats `/posts/{id}` (method-specific beats method-agnostic)
- `GET` patterns also match `HEAD` requests; all other methods match exactly
- Conflicting patterns with no clear precedence **panic at registration time** — catches bugs early
- `{$}` matches exact path only: `GET /posts/{$}` matches `/posts/` but not `/posts/234`
- 405 Method Not Allowed is returned automatically when path matches but method doesn't

---

## http.ResponseController (Go 1.20+)

Per-request control over read/write deadlines and flushing, without modifying server-wide settings.

### When to use
- SSE (Server-Sent Events) / streaming responses — disable write timeout, flush manually
- Long-running uploads — extend read deadline per request
- WebSocket-like upgrade patterns

### When NOT to use
- Normal request-response handlers — server defaults are fine
- If you don't need per-request deadline control

```go
func streamHandler(w http.ResponseWriter, r *http.Request) {
    rc := http.NewResponseController(w)

    // Disable server-wide WriteTimeout for this request
    rc.SetWriteDeadline(time.Time{})

    for event := range events {
        fmt.Fprintf(w, "data: %s\n\n", event)
        if err := rc.Flush(); err != nil {
            return // client disconnected
        }
    }
}

// Extend read deadline for large upload
func uploadHandler(w http.ResponseWriter, r *http.Request) {
    rc := http.NewResponseController(w)
    rc.SetReadDeadline(time.Now().Add(5 * time.Minute))
    // ... read large body ...
}
```

---

## http.CrossOriginProtection (Go 1.25+)

CSRF protection using Fetch metadata headers (Sec-Fetch-*). Drop-in middleware.

### When to use
- Any web application accepting form submissions or state-changing requests
- Replacing custom CSRF token middleware

### When NOT to use
- Pure API servers that only accept JSON (CORS preflight is usually sufficient)
- Public APIs designed to be called cross-origin

```go
mux := http.NewServeMux()
mux.HandleFunc("POST /transfer", transferHandler)

// Wrap with CSRF protection
protected := http.CrossOriginProtection(mux)
http.ListenAndServe(":8080", protected)
```

---

## net.KeepAliveConfig (Go 1.23+)

Fine-grained control over TCP keep-alive: idle time, interval, and probe count.

### When to use
- Long-lived connections (database pools, gRPC, WebSocket) where default keep-alive is too aggressive or too lax
- Environments with aggressive NAT/firewall timeouts

### When NOT to use
- Short-lived HTTP request/response — defaults work fine

```go
dialer := net.Dialer{
    KeepAliveConfig: net.KeepAliveConfig{
        Enable:   true,
        Idle:     30 * time.Second,
        Interval: 10 * time.Second,
        Count:    3,
    },
}
conn, err := dialer.Dial("tcp", "example.com:443")
```

---

## ReverseProxy.Rewrite (Go 1.20+, Director deprecated since 1.25)

Safer replacement for `ReverseProxy.Director`. The `Rewrite` hook receives a `ProxyRequest` that properly handles hop-by-hop headers.

### When to use
- All new reverse proxy code
- Migrating from `Director` — it has a known hop-by-hop header vulnerability

### When NOT to use
- Non-proxy HTTP handlers

```go
// Old — Director (deprecated, has hop-by-hop header vulnerability)
proxy := &httputil.ReverseProxy{
    Director: func(req *http.Request) {
        req.URL.Scheme = "https"
        req.URL.Host = "backend.internal"
        req.Host = "backend.internal"
    },
}

// New (Go 1.20+) — Rewrite
proxy := &httputil.ReverseProxy{
    Rewrite: func(r *httputil.ProxyRequest) {
        r.SetURL(backendURL)
        r.SetXForwarded()
        r.Out.Host = r.In.Host
    },
}
```

---

## math/rand/v2 (Go 1.22+)

New random number package with better API and algorithms (ChaCha8, PCG).

### When to use
- All new non-crypto random number generation
- Replacing `math/rand` — note renamed functions (`Intn` → `IntN`)
- Generic `rand.N[T]` works with any integer type including `time.Duration`

### When NOT to use
- Cryptographic randomness — use `crypto/rand`
- Reproducible sequences with `math/rand.Seed` — top-level `Seed` is a no-op since Go 1.24

```go
// Old
import "math/rand"
n := rand.Intn(100)

// New (Go 1.22+)
import "math/rand/v2"
n := rand.IntN(100)
d := rand.N(5 * time.Minute) // generic — works with Duration
```

---

## Migration strategy

1. Replace manual method checks in `HandleFunc` with method prefixes (Go 1.22+)
2. Replace path parsing with `r.PathValue("param")` wildcards
3. Migrate `ReverseProxy.Director` to `Rewrite`
4. Replace `math/rand` with `math/rand/v2`
5. Evaluate if third-party router can be replaced by stdlib for simpler services

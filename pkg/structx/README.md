# StructX

Type-safe, lightweight helpers for struct manipulation and conversion in Go.

## Usage Example

```go
package main

import (
    "fmt"
    "reflect"
    "github.com/kaptinlin/gozod/pkg/structx"
)

type User struct {
    Name  string `json:"name"`
    Email string `json:"email"`
    Age   int    `json:"age"`
}

func main() {
    user := User{Name: "John Doe", Email: "john@example.com", Age: 30}

    // struct -> map
    userMap := structx.Marshal(user)
    fmt.Println(userMap) // map[age:30 email:john@example.com name:John Doe]

    // map -> struct
    newUser, _ := structx.Unmarshal(userMap, reflect.TypeOf(User{}))
    fmt.Println(newUser)

    // Field operations
    if name, ok := structx.GetField(user, "name"); ok {
        fmt.Println("Name:", name)
    }
}
```

## Quick Reference

```go
import "github.com/kaptinlin/gozod/pkg/structx"

type Person struct{ Name string; Age int }
p := Person{"Alice", 25}

// Type checks
structx.Is(p)           // struct?
structx.IsPointer(&p)   // pointer to struct?

// struct <-> map
m := structx.Marshal(p) // map[string]any
m2, err := structx.ToMap(p)
obj, err := structx.Unmarshal(m, reflect.TypeOf(Person{}))

// Field operations
fields := structx.Fields(p)                  // []string
val, ok := structx.GetField(p, "Name")      // "Alice", true
err := structx.SetField(&p, "Age", 30)      // set by name/tag
ok = structx.HasField(p, "Age")             // true
count := structx.Count(p)                    // 2

// Utilities
out, ok := structx.Extract(p)                // map[string]any, true
clone, err := structx.Clone(p)               // deep copy
```

## JSON Tag Support

StructX respects `json` tags for field naming and omits fields with `json:"-"`:

```go
type APIResponse struct {
    UserID   int    `json:"user_id"`
    UserName string `json:"username"`
    IsActive bool   `json:"is_active"`
    Internal string `json:"-"` // skipped
}

data := structx.Marshal(APIResponse{1, "alice", true, "secret"})
// map[user_id:1 username:alice is_active:true]
```

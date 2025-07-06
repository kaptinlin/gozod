# StructX

Lightweight utilities for converting between Go structs and `map[string]any`.

## Features

* **Marshal** – Convert a struct (or pointer to struct) to `map[string]any`.
* **ToMap** – Same as Marshal but returns an error when input is nil or not a struct.
* **Unmarshal / FromMap** – Populate a struct instance from `map[string]any` using reflection.
* Respects `json` tags and skips unexported or `json:"-"` fields.

## Quick Start

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
    user := User{"John Doe", "john@example.com", 30}

    // struct → map
    userMap := structx.Marshal(user)
    fmt.Println(userMap)

    // struct → map with error handling
    userMap2, err := structx.ToMap(user)
    fmt.Println(userMap2, err)

    // map → struct
    restored, _ := structx.Unmarshal(userMap, reflect.TypeOf(User{}))
    fmt.Println(restored)
}
```

## JSON Tag Support

`structx` automatically honors `json` tags for field names and omits fields tagged with `-`:

```go
type APIResponse struct {
    UserID   int    `json:"user_id"`
    UserName string `json:"username"`
    IsActive bool   `json:"is_active"`
    Secret   string `json:"-"` // skipped
}

data := structx.Marshal(APIResponse{1, "alice", true, "hidden"})
// map[user_id:1 username:alice is_active:true]
```

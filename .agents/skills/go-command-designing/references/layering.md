# CLI Application Layering

## Four-Layer Architecture

```
cmd/
├── commands/          # Layer 1: Controllers (50-150 lines)
├── services/          # Layer 2: Business logic
internal/              # Layer 3: Implementation
pkg/                   # Layer 4: Reusable utilities
```

## Layer 1: Commands (Controllers)

**Size:** 50-150 lines per handler

**Responsibilities:**
- Parse CLI flags/arguments
- Call service methods
- Format output
- Handle exit codes

**Example:**
```go
func createHandler(ctx *command.Context) error {
    svc := services.NewComponentService()
    result, err := svc.Create(ctx.Context(), services.CreateInput{
        Name: ctx.Args[0],
        Tier: ctx.String("tier"),
    })
    if err != nil {
        return err
    }
    ctx.Successf("Created: %s\n", result.Path)
    return nil
}
```

## Layer 2: Services

**Location:** `cmd/services/`

**Responsibilities:**
- Orchestrate business operations
- Coordinate internal packages
- Return structured data
- **Can consume `internal/` and `pkg/`**

## Layer 3: Internal

**Responsibilities:**
- Pure business logic
- Stateless implementations
- **Can consume `pkg/` only**

## Layer 4: Pkg

**Responsibilities:**
- Reusable utilities
- Zero application dependencies

## When to Extract

Extract to services when:
1. Handler > 150 lines
2. Repeated logic
3. Need testing without CLI
4. Planning API interface

## Dependency Rules

```
Commands → Services → Internal → Pkg
   ✗         ✓          ✓       Can import pkg
   ✗         ✗          ✗       Cannot import cmd
```

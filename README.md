# GoZod

[![Go Version](https://img.shields.io/badge/Go-%3E%3D1.26-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

A TypeScript Zod v4-inspired validation library for Go with strict type semantics, fluent schemas, and JSON Schema interoperability

## Features

- **Strict type semantics**: Value and pointer schemas accept exact input types unless you explicitly opt into coercion.
- **Dual parsing modes**: Use `Parse(any)` for dynamic data and `StrictParse(T)` when the input type is already known.
- **Rich schema surface**: Compose primitives, collections, structs, unions, intersections, metadata, transforms, and refinements.
- **Struct tags**: Build schemas from Go structs with `gozod:"..."` rules and alternate tag names through `WithTagName`.
- **Optional code generation**: Use `gozodgen` for generated schema helpers in tag-heavy hot paths.
- **Localized errors**: Inspect `*gozod.ZodError`, flatten or prettify failures, and use locale bundles from `locales/`.
- **JSON Schema bridge**: Convert to and from JSON Schema Draft 2020-12 with the bundled `jsonschema` package.
- **Curated dependency surface**: Built on JSON v2, `jsonschema`, `deepclone`, and i18n helpers instead of a framework stack.

## Installation

```bash
go get github.com/kaptinlin/gozod
```

Requires **Go 1.26+**.

## Quick Start

```go
package main

import (
	"fmt"
	"log"

	"github.com/kaptinlin/gozod"
)

func main() {
	schema := gozod.String().Min(2).Email()

	value, err := schema.Parse("dev@example.com")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(value)
}
```

## Parse and StrictParse

GoZod keeps runtime parsing and compile-time constrained parsing separate.

```go
nameSchema := gozod.String().Min(2).Max(50)

name, err := nameSchema.Parse("Alice")
if err != nil {
	log.Fatal(err)
}

strictName, err := nameSchema.StrictParse("Alice")
if err != nil {
	log.Fatal(err)
}

fmt.Println(name, strictName)
```

Use `Parse(any)` when data arrives from JSON, maps, or other dynamic sources. Use `StrictParse(T)` when your program already has the target Go type and you want the strict side of the API.

## Struct Tags and Generated Schemas

Use `FromStruct[T]()` for declarative validation on native Go structs.

```go
package main

import (
	"fmt"
	"log"

	"github.com/kaptinlin/gozod"
)

type User struct {
	Name  string `gozod:"required,min=2,max=50"`
	Email string `gozod:"required,email"`
	Age   int    `gozod:"min=18,max=120"`
}

func main() {
	schema := gozod.FromStruct[User]()

	user, err := schema.Parse(User{
		Name:  "Ada Lovelace",
		Email: "ada@example.com",
		Age:   36,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%+v\n", user)
}
```

If you use a different tag key, pass `gozod.WithTagName("validate")`.

For generated helpers, install and run `gozodgen`:

```bash
go install github.com/kaptinlin/gozod/cmd/gozodgen@latest
go generate ./...
```

See [docs/tags.md](docs/tags.md) and [cmd/gozodgen](cmd/gozodgen/) for the full struct-tag and code-generation workflow.

## Programmatic Schemas

Use `Object`, `Struct`, `Union`, `Intersection`, and related constructors when you want the schema shape in code.

```go
userSchema := gozod.Object(gozod.ObjectSchema{
	"name":  gozod.String().Min(2),
	"email": gozod.Email(),
	"age":   gozod.Int().Min(18),
})

contactSchema := gozod.Union([]any{
	gozod.Email(),
	gozod.URL(),
})

_, _ = userSchema.Parse(map[string]any{
	"name":  "Grace",
	"email": "grace@example.com",
	"age":   28,
})

_, _ = contactSchema.Parse("https://example.com")
```

For coercion-first flows, use the constructors in [coerce/](coerce/).

## JSON Schema Integration

GoZod can translate schemas to JSON Schema Draft 2020-12 and back.

```go
schema := gozod.Object(gozod.ObjectSchema{
	"name": gozod.String().Min(1),
	"age":  gozod.Int().Min(0),
})

jsonSchema, err := gozod.ToJSONSchema(schema)
if err != nil {
	log.Fatal(err)
}

result := jsonSchema.ValidateMap(map[string]any{
	"name": "Lin",
	"age":  30,
})

fmt.Println(result.IsValid())
```

See [docs/json-schema.md](docs/json-schema.md) for conversion details and compatibility notes.

## Error Handling

Validation failures return `error`. Inspect them as `*gozod.ZodError` when you need structured details.

```go
schema := gozod.String().Min(5)
_, err := schema.Parse("hi")
if err == nil {
	return
}

var zodErr *gozod.ZodError
if gozod.IsZodError(err, &zodErr) {
	fmt.Println(gozod.PrettifyError(zodErr))
}
```

See [docs/error-customization.md](docs/error-customization.md) and [docs/error-formatting.md](docs/error-formatting.md) for custom messages and output shapes.

## Examples and Documentation

- [docs/api.md](docs/api.md) — API reference and method surface
- [docs/basics.md](docs/basics.md) — core concepts and common patterns
- [docs/tags.md](docs/tags.md) — struct-tag validation guide
- [docs/json-schema.md](docs/json-schema.md) — JSON Schema conversion
- [docs/feature-mapping.md](docs/feature-mapping.md) — TypeScript Zod v4 to GoZod mapping
- [docs/metadata.md](docs/metadata.md) — schema metadata and registries
- [examples/README.md](examples/README.md) — runnable examples by topic

Run an example directly:

```bash
go run ./examples/quickstart
go run ./examples/struct_tags
```

## Performance

GoZod includes benchmarks for parsing, checks, tags, transforms, and configuration helpers.

- Prefer `StrictParse` when the input type is already known.
- Use [coerce/](coerce/) only when conversion is part of the requirement.
- Use `gozodgen` for tag-heavy hot paths where reflection cost matters.

Run the benchmark suite with:

```bash
go test -bench=. ./...
```

## Development

```bash
task test                         # Run the default test suite
task test:race                    # Run race-enabled tests for lightweight packages
task lint                         # Run golangci-lint and tidy checks
task verify                       # Run deps, fmt, vet, lint, test, and govulncheck
go test -tags=contractcheck ./types # Audit compile-time schema contracts
```

For development guidelines and repository conventions, see [AGENTS.md](AGENTS.md).

## Contributing

Contributions are welcome. Run the test and lint commands before opening a pull request, and keep docs and examples aligned with the current API surface.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

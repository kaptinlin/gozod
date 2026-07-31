# GoZod

[![Go Module](https://img.shields.io/badge/go-1.26.5%2B-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

A TypeScript Zod v4-inspired validation library for Go with strict type semantics, fluent schemas, struct tags, and JSON Schema Draft 2020-12 interoperability

## Features

- **Go type semantics**: Primitive schemas accept their Go kind, including defined primitives, while cross-kind coercion stays explicit and struct output preserves declared field types.
- **Dual parsing modes**: Use `Parse(any)` for dynamic data and `StrictParse(T)` for known Go values.
- **Fluent schema API**: Compose primitives, collections, structs, unions, intersections, transforms, refinements, defaults, and metadata.
- **Schema-described output**: Object fields, catchall values, defaults, prefaults, transforms, and overwrites return the parsed child output.
- **Struct tags**: Build schemas from Go structs with `gozod:"..."` tags, json/yaml/toml field names, custom tag keys, and circular-reference support.
- **Generated schemas**: Use `gozodgen` for tag-heavy paths where generated helpers are preferable to reflection.
- **Localized errors**: Inspect `*gozod.ZodError`, prettify or flatten failures, and switch message bundles through `locales/`.
- **JSON Schema bridge**: Convert GoZod schemas to JSON Schema and import JSON Schema back into GoZod with explicit lossy-conversion controls.
- **Focused dependency surface**: Uses JSON v2, `jsonschema`, `deepclone`, JWT parsing, and Unicode helpers without a framework stack.

## Installation

```bash
go get github.com/kaptinlin/gozod
```

Requires **Go 1.26.5+**.

## Quick Start

```go
package main

import (
	"fmt"
	"log"

	"github.com/kaptinlin/gozod"
)

func main() {
	email := gozod.String().Min(5).Email()

	value, err := email.Parse("dev@example.com")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(value)
}
```

## Parse and StrictParse

Use `Parse(any)` at boundaries where input is dynamic. Use `StrictParse(T)` when your program already has the target Go type.

```go
name := gozod.String().Min(2).Max(50)

fromJSON, err := name.Parse("Alice")
if err != nil {
	log.Fatal(err)
}

knownValue, err := name.StrictParse("Grace")
if err != nil {
	log.Fatal(err)
}

fmt.Println(fromJSON, knownValue)
```

`StrictParse` keeps the call site compile-time constrained. `Parse` keeps the
boundary flexible, accepts defined primitives with the expected underlying
kind, and does not perform cross-kind coercion unless the schema explicitly
enables it.

## Struct Tags

Use `FromStruct[T]()` when validation belongs next to a Go struct. Construction
errors report invalid tags or unsupported field types before parsing begins.

```go
package main

import (
	"fmt"
	"log"

	"github.com/kaptinlin/gozod"
)

type User struct {
	Name  string `json:"name" gozod:"required,min=2,max=50"`
	Email string `json:"email" gozod:"required,email"`
	Age   int    `json:"age" gozod:"min=18,max=120"`
}

func main() {
	schema, err := gozod.FromStruct[User]()
	if err != nil {
		log.Fatal(err)
	}

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

Use `gozod.WithTagName("validate")` when your project uses another rule tag, and `gozod.WithFieldNameTag("yaml")` to resolve field names (error paths and JSON Schema) from `yaml`/`toml` tags instead of `json`. See [docs/tags.md](docs/tags.md) for supported tag rules and generated-schema details.

`required` is the field-presence switch. Tagged fields without `required` are
optional whether their Go type is a value or pointer; pointer types separately
control pointer output and nil handling.

## Programmatic Schemas

Use root constructors when you want the schema shape in code.

```go
user := gozod.Object(gozod.ObjectSchema{
	"name":  gozod.String().Min(2),
	"email": gozod.Email(),
	"age":   gozod.Int().Min(18),
})

contact := gozod.Union([]any{
	gozod.Email(),
	gozod.URL(),
})

parsedUser, err := user.Parse(map[string]any{
	"name":  "Grace",
	"email": "grace@example.com",
	"age":   28,
})
if err != nil {
	log.Fatal(err)
}

parsedContact, err := contact.Parse("https://example.com")
if err != nil {
	log.Fatal(err)
}

fmt.Println(parsedUser, parsedContact)
```

Object parsing returns schema-described output. If a child schema trims,
overwrites, transforms, supplies a default, or validates catchall values, the
returned map contains that parsed child value.

For conversion-first flows, import [coerce/](coerce/) and choose coercion explicitly.

```go
import "github.com/kaptinlin/gozod/coerce"

age, err := coerce.Int().Parse("42")
if err != nil {
	log.Fatal(err)
}

fmt.Println(age)
```

## Defaults, Prefaults, and Metadata

`Default` short-circuits validation for nil input. `Prefault` runs the fallback through the full parsing pipeline.

```go
displayName := gozod.String().Min(3).Default("Guest")
normalized := gozod.String().
	Trim().
	ToLowerCase().
	Prefault("  Example  ")

name, _ := displayName.Parse(nil)
slug, _ := normalized.Parse(nil)

fmt.Println(name, slug)
```

Metadata modifiers are copy-on-write. The original schema is left unchanged and metadata is stored on the returned schema.

```go
email := gozod.Email().Meta(gozod.GlobalMeta{
	Title:       "Email Address",
	Description: "Primary contact email",
	Examples:    []any{"user@example.com"},
})

meta := email.Internals().Metadata()
fmt.Println(meta.Title)
```

See [docs/metadata.md](docs/metadata.md) for registries and JSON Schema metadata merging.

## JSON Schema

`gozod.ToJSONSchema` returns a `*jsonschema.Schema` from `github.com/kaptinlin/jsonschema`.
Use the exported `JSONSchema*` mode constants instead of raw strings when setting conversion options.
It exports one schema using the fixed Draft 2020-12 dialect. Use the separate
typed `ToJSONSchemaRegistry` entry point for a registry bundle.

```go
schema := gozod.Object(gozod.ObjectSchema{
	"name": gozod.String().Min(1),
	"age":  gozod.Int().Min(0),
})

jsonSchema, err := gozod.ToJSONSchema(schema)
if err != nil {
	log.Fatal(err)
}

result := jsonSchema.Validate(map[string]any{
	"name": "Lin",
	"age":  30,
})

fmt.Println(result.IsValid())
```

`gozod.FromJSONSchema` fails closed on unsupported JSON Schema keywords. Use
`gozod.FromJSONSchemaLossy` only when dropping unsupported semantics is intentional.

```go
zodSchema, losses, err := gozod.FromJSONSchemaLossy(jsonSchema)
if err != nil {
	log.Fatal(err)
}

_, _ = zodSchema.ParseAny(map[string]any{"name": "Lin", "age": 30})
for _, loss := range losses {
	fmt.Printf("%s at %s: %v\n", loss.Keyword, loss.Pointer, loss.Err)
}
```

Imported `$id`, title, description, and examples belong to the returned schema
by default. Pass a registry only when the caller needs a separate metadata
destination; the importer snapshots nested examples in either mode.

```go
metadata := gozod.NewRegistry[gozod.GlobalMeta]()
zodSchema, err := gozod.FromJSONSchema(jsonSchema, gozod.FromJSONSchemaOptions{
	Metadata: metadata,
})
```

For a batch export, give every top-level registry entry a unique non-empty ID:

```go
registry := gozod.NewRegistry[gozod.GlobalMeta]().
	Add(gozod.String(), gozod.GlobalMeta{ID: "Name"})

bundle, err := gozod.ToJSONSchemaRegistry(registry)
if err != nil {
	log.Fatal(err)
}

fmt.Println(len(bundle.Defs))
```

See [docs/json-schema.md](docs/json-schema.md) for conversion options, unsupported features, registries, and Draft 2020-12 notes.

## Error Handling

Validation failures return `error`. Use GoZod helpers when you need structured inspection or presentation.

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

See [docs/error-customization.md](docs/error-customization.md) and [docs/error-formatting.md](docs/error-formatting.md).

## Code Generation

Install `gozodgen` when generated struct-tag schemas fit your path better than reflection:

```bash
go install github.com/kaptinlin/gozod/cmd/gozodgen@latest
go generate ./...
```

`gozodgen` loads each concrete package with its Go module and build context,
fails before publishing files when package analysis or schema rendering fails,
and generates only types it can represent faithfully. Use `-tag-name`,
`-field-name-tag`, `-suffix`, and `-method` to select real generation behavior.
Generated `Schema()` methods are explicit; `FromStruct[T]()` remains the
reflection path and does not auto-discover them.

See [cmd/gozodgen](cmd/gozodgen/) and [examples/code_generation](examples/code_generation/).

## Documentation

- [docs/basics.md](docs/basics.md) - core concepts and common patterns
- [docs/api.md](docs/api.md) - API reference and method surface
- [docs/tags.md](docs/tags.md) - struct-tag validation guide
- [docs/json-schema.md](docs/json-schema.md) - JSON Schema conversion
- [docs/metadata.md](docs/metadata.md) - schema metadata and registries
- [docs/feature-mapping.md](docs/feature-mapping.md) - TypeScript Zod v4 to GoZod mapping
- [examples/README.md](examples/README.md) - runnable examples by topic

Run examples directly:

```bash
go run ./examples/quickstart
go run ./examples/struct_tags
go run ./examples/error_handling
```

## Performance

GoZod includes benchmarks for parsing, checks, tags, transforms, and configuration helpers.

- Prefer `StrictParse` when the input type is already known.
- Use [coerce/](coerce/) only when conversion is part of the requirement.
- Use `gozodgen` for tag-heavy hot paths where generated helpers are worth the extra file.

```bash
go test -bench=. ./...
task bench:smoke
```

## Development

```bash
task test                           # Run go test -race ./...
task test:race                      # Run race tests for core and lightweight utility packages
task lint                           # Run golangci-lint and tidy-lint
task golangci-lint                  # Run golangci-lint v2 only
task tidy-lint                      # Verify go.mod and go.sum stay tidy
task contractcheck                  # Run compile-time schema contract checks
task docs:integrity                 # Run public documentation stale-claim checks
task bench:smoke                    # Compile and smoke-run benchmarks without enforcing numbers
task verify:ci                      # Run the CI-equivalent quality gate
task verify                         # Run deps, fmt, vet, lint, contracts, docs, tests, benchmark smoke, and vuln
```

For development guidelines and repository conventions, see [AGENTS.md](AGENTS.md).

## Contributing

Contributions are welcome. Run the test and lint commands before opening a pull request, and keep docs and examples aligned with the current API surface.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

# README Template

Use this template as the structural guide. Adapt sections to the project's scope — small libraries skip sections, large libraries expand them.

## Template

````markdown
# {Project Name}

{badges — see Badge Patterns in SKILL.md}

{One-line description: what it is and its core value proposition, no period}

## Features

- **{Keyword}**: {explanation of what this enables}
- **{Keyword}**: {explanation}
- **{Keyword}**: {explanation}
- **{Keyword}**: {explanation}

## Installation

```bash
go get github.com/{org}/{repo}
```

Requires **Go {version}+**.

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    "{module-path}"
)

func main() {
    // Minimal working example showing the core use case
    // Include error handling
    // Keep under 30 lines
}
```

## {Domain Section 1 — e.g., Configuration, API, Usage}

{Content specific to the library's domain. Use tables for options:}

| Option | Description | Default |
|--------|-------------|---------|
| `WithX(v)` | What it does | `default` |

## {Domain Section 2 — e.g., Advanced Features, Examples}

### {Subsection}

```go
// Progressive complexity: show intermediate/advanced usage
```

## Performance

{Include only if the project has benchmarks}

Tested on {hardware}, {os}:

| Operation | Performance | Memory | Allocations |
|-----------|-------------|--------|-------------|
| {Operation} | {ns/op} | {B/op} | {allocs/op} |

```bash
cd benchmarks && go test -bench=. -benchmem
```

## Development

```bash
task test    # Run all tests with race detector
task lint    # Run golangci-lint
make all     # lint + test
```

## Contributing

Contributions are welcome. Please open an issue to discuss your idea before submitting a pull request.

## License

{Use the appropriate license text — see License Rules in SKILL.md}
````

---

## Pattern: Small Library (~100-150 lines)

Libraries with a focused API (deepclone, jsonpointer, jsonmerge):

```
# Title + badges + one-liner
## Features (4-6 items)
## Installation
## Quick Start
## Examples (2-3 subsections)
## API Reference (core function signatures)
## Performance (if benchmarks exist)
## Contributing (2-3 lines)
## License
```

## Pattern: Medium Library (~200-350 lines)

Libraries with moderate API surface (emitter, go-fsm, condeval):

```
# Title + badges + one-liner
## Features (6-8 items)
## Installation
## Quick Start
## Configuration (options table)
## {Core Concept 1} (with examples)
## {Core Concept 2} (with examples)
## Examples (3-5 detailed subsections)
## Development (make commands)
## Contributing
## License
```

## Pattern: Large Library (~400+ lines)

Libraries with extensive API or multiple subsystems (jsonschema, gozod, godocx):

```
# Title + badges + one-liner
## Features (8+ items)
## Installation
## Quick Start
## {Major Feature 1} (multiple subsections)
## {Major Feature 2} (multiple subsections)
## Advanced Features (3-4 subsections)
## Error Handling
## Performance
## API Reference (or link to pkg.go.dev)
## Development
## Documentation (links to docs/ files)
## Contributing
## License
```

For 400+ line READMEs, consider adding a Table of Contents after badges.

---

## Examples by Section

### Good One-Line Descriptions

```markdown
A high-performance deep cloning library for Go that provides safe, efficient copying of any Go value

A lightweight, generic finite state machine library for Go 1.26+

A framework-agnostic, plugin-driven secrets management library for Go 1.26+

A high-performance, thread-safe Go library for event management that leverages modern Go features
```

### Good Feature Lists

```markdown
## Features

- **High Performance**: Built with `atomic.Pointer` and modern Go optimizations
- **Zero Dependencies**: Uses only Go standard library
- **Type-Safe**: Generic `[S, E comparable]` parameters, no `any`, no `reflect`
- **Thread-Safe**: Designed for high-concurrency environments
- **Extensible**: Custom behavior via `Cloneable` interface
```

```markdown
## Features

- **Zero-config defaults**: file store + AES-256-GCM envelope encryption
- **Pluggable**: swap Store, Cipher, MasterKey, or add plugins
- **Scope-based isolation**: `scope` is an opaque string you define
- **Library-first**: `go get` to import, CLI included
```

### Good Quick Start

```go
package main

import (
    "fmt"
    "github.com/kaptinlin/deepclone"
)

func main() {
    original := map[string][]int{
        "numbers": {1, 2, 3},
    }

    cloned := deepclone.Clone(original)

    original["numbers"][0] = 999
    fmt.Println("Original:", original["numbers"]) // [999, 2, 3]
    fmt.Println("Cloned:", cloned["numbers"])     // [1, 2, 3]
}
```

### Good License Sections

**MIT (kaptinlin):**
```markdown
## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
```

**MIT (agentable):**
```markdown
## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
```

**Agentable Commercial:**
```markdown
## License

This software is licensed under the **Agentable Commercial License**, exclusively for use with Agentable platform services and their direct integrations.
See the [LICENSE](./LICENSE) file for full terms.
```

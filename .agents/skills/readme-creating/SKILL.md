---
name: readme-creating
description: Generate README.md for Go libraries in the kaptinlin and agentable ecosystem. Use when creating, updating, or reviewing README files for Go packages. Triggers on README creation, documentation generation, or "write a README" requests for Go projects under kaptinlin or agentable organizations.
---

# Go README Creation

Generate consistent README.md for kaptinlin/agentable Go libraries.

## Workflow

1. **Analyze project** — Read `go.mod` (module path, Go version), scan exported API, read `DESIGN.md` if present, check Makefile targets
2. **Determine license** — See [License Rules](#license-rules)
3. **Generate README** — Follow template in [references/readme-template.md](references/readme-template.md)
4. **Validate** — Ensure sections present, examples compilable, badges correct

## License Rules

Determine license by reading the first line of the LICENSE file:
- `"MIT License"` → MIT
- `"AGENTABLE COMMERCIAL LICENSE"` → Agentable Commercial
- No LICENSE file → Ask the user

### kaptinlin — Always MIT

```markdown
## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
```

### agentable — Default Agentable Commercial, MIT when LICENSE file says so

**Known MIT exceptions:** `agentstack`, `condeval`, `go-fsm`, `jsondiff`, `openapi-generator`, `openapi-request`, `unifmsg`

**Everything else under agentable:** Agentable Commercial License

```markdown
## License

This software is licensed under the **Agentable Commercial License**, exclusively for use with Agentable platform services and their direct integrations.
See the [LICENSE](./LICENSE) file for full terms.
```

## Badge Patterns

Always include Go Version + License. Read Go version from `go.mod` `go` directive.

**kaptinlin (MIT):**
```markdown
[![Go Version](https://img.shields.io/badge/go-%3E%3D{VERSION}-blue)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)
```

**agentable (MIT):**
```markdown
[![Go Version](https://img.shields.io/badge/go-%3E%3D{VERSION}-blue)](https://golang.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
```

**agentable (Commercial):**
```markdown
[![Go Version](https://img.shields.io/badge/go-%3E%3D{VERSION}-blue)](https://golang.org/)
[![License](https://img.shields.io/badge/license-Agentable-purple)](LICENSE)
```

**Optional badges** (add when the service is configured):
```markdown
[![Go Reference](https://pkg.go.dev/badge/github.com/{ORG}/{REPO}.svg)](https://pkg.go.dev/github.com/{ORG}/{REPO})
[![Go Report Card](https://goreportcard.com/badge/github.com/{ORG}/{REPO})](https://goreportcard.com/report/github.com/{ORG}/{REPO})
```

## README Section Order

See [references/readme-template.md](references/readme-template.md) for the full template with patterns.

```
1. # Title                    — Project name, clean and simple
2. Badges                     — Go Version + License (required), others optional
3. One-line description       — One sentence, no period
4. Features                   — 4-8 bullets, **Bold keyword**: explanation
5. Installation               — go get command
6. Quick Start                — Minimal working package main example
7. [Domain sections]          — API, Configuration, Examples, etc.
8. Performance                — If benchmarks exist
9. Development                — make targets from Makefile
10. Contributing              — Brief or link to CONTRIBUTING.md
11. License                   — Per license rules above
```

## Writing Rules

- **One-line description**: After badges, one sentence, start with "A ..." or active description
- **Features**: `- **Bold keyword**: explanation` format, focus on differentiators
- **Quick Start**: Complete `package main` with imports, runnable in <30 lines
- **Code examples**: `go` syntax highlighting, include error handling, simple→complex progression
- **Tables**: For options/config, benchmarks, operators, API overview
- **Development**: Show actual `make` targets from the project's Makefile
- **No emojis**: Unless the project already uses them
- **Language**: English by default; match existing if updating
- Don't explain Go basics or how `go get` works
- Don't list every exported function — link to pkg.go.dev
- Don't add badges for services not yet set up
- Don't duplicate content from DESIGN.md or docs/

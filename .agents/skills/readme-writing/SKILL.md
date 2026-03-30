---
description: Generate README.md for libraries and projects. Use when creating, updating, or reviewing README files for packages. Triggers on README creation, documentation generation, or "write a README" requests.
name: readme-writing
---


# README Creation

Generate consistent README.md for software libraries.

## Philosophy

README positioning and responsibilities:
- Usage guide, not design documentation
- For human users, not AI Agents
- Answers "how to use", not "what must be followed"

## Content Boundaries

### Belongs in README
- Installation and quick start (operation guide)
- Feature list (value proposition)
- Usage examples (tutorials)
- Development commands (operation instructions)
- Contributing guide (collaboration process)

### Does NOT Belong in README
- System contracts and data schemas → SPECS/
- Architecture constraints and design decisions → SPECS/
- AI Agent workflow instructions → AGENTS.md
- Implementation details and algorithms → code comments
- Historical change records → git log

## Abstraction Level

README works at the "usage layer", not "implementation layer" or "contract layer".

**Usage Layer** (belongs in README):
- How to install and import
- How to call APIs
- Common usage scenario examples
- How to run tests and build

**Contract Layer** (does NOT belong in README):
- What format API must return → SPECS/
- Data schema definitions → SPECS/
- Naming conventions → SPECS/

**Implementation Layer** (does NOT belong in README):
- Specific algorithm implementations → code
- Performance optimization techniques → code comments
- Internal data structures → code

## Writing Style

Core principles for concise, precise writing:

### Conciseness Principle
- Every sentence adds value, remove filler words
- State "what it is" first, then "why"
- Avoid lengthy motivational text

### Precise Verbs
- Use specific verbs: install, configure, run, test
- Avoid vague verbs: handle, manage, execute

### Imperative Mood
- Use command form: "Run `go test`"
- Avoid passive voice: "Tests should be run"
- Avoid suggestive mood: "You might want to run tests"

### Terminology Consistency
- Use only one term per concept
- Maintain consistency with SPECS/ terminology

## README vs AGENTS.md

| Dimension | README | AGENTS.md |
|-----------|--------|-----------|
| Audience | Human users | AI Agent |
| Purpose | How to use | How to develop |
| Content | Installation, examples, API overview | Architecture, conventions, workflow |
| Style | Friendly, tutorial-style | Concise, imperative |
| Examples | Complete runnable code | Patterns and constraints |

**Collaboration Principles**:
- README links to AGENTS.md: "For development guidelines, see AGENTS.md"
- AGENTS.md does not duplicate README installation and usage instructions
- Both share terminology definitions (link to SPECS/)

## AI Readability

README is primarily for humans, but AI Agents also read it. Optimize for AI readability:

### Structured Information
- Use standard section headers (Features, Installation, Quick Start)
- Use tables for configuration options and API overview
- Use code blocks with language tags (```go, ```bash)

### Key Information First
- One-sentence description at the top
- Features list uses **Bold keyword**: explanation format
- Quick Start example includes complete imports and error handling

### Avoid Ambiguity
- Use precise commands (`go test -race ./...`)
- Specify Go version requirements explicitly ("Requires Go 1.26+")
- Example code is directly runnable without modification

## Workflow

1. **Analyze project** — Read project manifest (go.mod, package.json, etc.), scan exported API, read DESIGN.md if present, check build configuration
2. **Determine license** — See [License Rules](#license-rules)
3. **Generate README** — Follow template in [references/readme-template.md](references/readme-template.md)
4. **Validate** — Ensure sections present, examples runnable, badges correct

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

See `references/badges.md` for language-specific badge patterns.

## README Section Order

See [references/readme-template.md](references/readme-template.md) for the full template with patterns.

### Progressive Disclosure

Adjust detail level based on project complexity:

**Simple Library** (< 5 exported functions):
```
1. Title + Badges
2. One-line description
3. Features (4-6 bullets)
4. Installation
5. Quick Start (single example)
6. License
```

**Medium Library** (5-20 exported functions):
```
1. Title + Badges
2. One-line description
3. Features (6-8 bullets)
4. Installation
5. Quick Start
6. API Overview (table or sections)
7. Examples
8. Development
9. Contributing
10. License
```

**Complex Library** (20+ exported functions or multiple packages):
```
1. Title + Badges
2. One-line description
3. Features (6-8 bullets)
4. Installation
5. Quick Start
6. Core Concepts
7. API Reference (link to pkg.go.dev)
8. Advanced Examples
9. Configuration
10. Performance
11. Development
12. Contributing
13. License
```

### Standard Section Order

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
- **Code examples**: Proper syntax highlighting, include error handling, simple→complex progression
- **Tables**: For options/config, benchmarks, operators, API overview
- **Development**: Show actual build commands from project configuration
- **No emojis**: Unless the project already uses them
- **Language**: English by default; match existing if updating
- Don't explain language basics or package manager usage
- Don't list every exported function — link to API documentation
- Don't add badges for services not yet set up
- Don't duplicate content from DESIGN.md or docs/
- Don't include design decisions or architecture constraints — those belong in SPECS/
- Don't include AI Agent workflow instructions — those belong in AGENTS.md

## Anti-Patterns

| Avoid | Why | Instead |
|-------|-----|---------|
| Including system contracts and data schemas | Belongs in SPECS/, not README | Link to SPECS/ for contract details |
| Including architecture constraints | Design decisions belong in SPECS/ | Provide usage examples only |
| Including AI Agent workflow instructions | README is for humans, not agents | Put agent instructions in AGENTS.md |
| Documenting implementation details | Code comments are better | Show usage, not internals |
| Recording historical changes | git log provides history | Document current usage only |
| Listing all 50 exported functions | Too verbose, becomes stale | Link to API documentation |
| Adding badges for unconfigured services | Misleading | Only add badges for active services |
| Duplicating DESIGN.md content | Maintenance burden | Link to design docs if needed |
| Explaining language basics | Wastes space | Assume language knowledge |
| Using passive voice | Less clear | Use imperative: "Run `go test`" |

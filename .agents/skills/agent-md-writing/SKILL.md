---
description: Generate a CLAUDE.md (and AGENTS.md symlink) by analyzing project source code and existing docs to produce a comprehensive AI coding guide following progressive disclosure principles. Use when the user wants to create, generate, or initialize a CLAUDE.md or AGENTS.md.
name: agent-md-writing
---


# CLAUDE.md Generator

Generate a comprehensive CLAUDE.md for software projects by analyzing the repository and producing structured AI coding instructions. Also create an AGENTS.md symlink.

## Core Principles

CLAUDE.md/AGENTS.md follows SPEC core principles:

### 1. Single Source of Truth (SSOT)
- AGENTS.md is the single source for AI Agent development instructions
- Does not duplicate README usage instructions
- Does not duplicate SPECS/ design details
- Uses indexes and references to establish relationships

### 2. Current State Only
- Documents current architecture and conventions only
- No historical changes (git log provides history)
- No implementation phases or TODOs (use issue tracker)

### 3. High Cohesion
- Related conventions grouped together
- Organized by domain (Architecture, Coding Rules, Testing)
- Avoids information fragmentation

### 4. Actionable
- Every rule is verifiable
- Provides specific constraints, not abstract principles
- Includes Forbidden section for explicit prohibitions

### 5. Concise & Essential
- Every sentence adds value
- Removes Go basics that AI already knows
- Documents only project-specific conventions

### 6. Context Minimalism
- Contains only what Agent needs for development
- Uses indexes to point to detailed content (SPECS/, .references/)
- Avoids context bloat

## Content Boundaries

### Belongs in AGENTS.md
- Project overview and architecture index
- Development commands (task, make, go)
- Coding conventions and forbidden patterns
- Testing strategy and tools
- Agent workflow (read SPECS, find References)
- SPECS index and References index
- Agent Skills index

### Does NOT Belong in AGENTS.md
- Type definitions and interface signatures → SPECS/
- Design pattern catalogs → SPECS/ or source code
- Detailed CLI convention tables → SPECS/
- Implementation phases and migration plans → PLAN.md
- TODO lists and task tracking → issue tracker
- Installation and setup instructions → README.md
- Historical background (unless preventing common mistakes)

### Judgment Criterion: Can It Be Violated?
- Rules that can be violated → AGENTS.md (e.g., "Use t.Parallel()")
- Statements that cannot be violated → Not in AGENTS.md (e.g., "We use Go")

### Must Include When Project Has SPEC/
- **No documentation masquerading as code** — When a project uses SPEC-driven development, the Forbidden section must include a rule prohibiting over-encoding: translating spec prose into code that no program consumes at runtime. Judgment: does a program read this value to drive behavior?

### Must Include: Dependency Issue Reporting

Every generated CLAUDE.md **must** include a dependency issue reporting rule in the Forbidden section and a corresponding workflow instruction. When a project depends on external libraries, agents must not work around dependency bugs by reimplementing functionality — this creates hidden tech debt that accumulates silently.

**Add to Forbidden section:**
```markdown
- No working around dependency bugs — if a bug or limitation is in a dependency library, do NOT bypass it by reimplementing the dependency's functionality. Instead, create a report file in `reports/` (see Dependency Issue Reporting below).
```

**Add as a standalone section (after Coding Rules or after Agent Workflow):**
```markdown
## Dependency Issue Reporting

When you encounter a bug, limitation, or unexpected behavior in a dependency library:

1. **Do NOT** work around it by reimplementing the dependency's functionality
2. **Do NOT** skip or ignore the dependency and write your own version
3. **Do** create a report file: `reports/<dependency-name>.md`
4. **Do** include in the report:
   - Dependency name and version
   - Problem description (what went wrong)
   - Trigger scenario (what you were doing when you hit it)
   - Expected behavior vs actual behavior
   - Relevant error messages or stack traces
   - Workaround suggestion (if any, without implementing it)
5. **Do** continue with other tasks that don't depend on the broken functionality

The `reports/` directory is checked by team members after each work cycle. Reports are routed to the appropriate dependency maintainer for resolution.
```

**Section inclusion rule:** Always include this section. It applies to all projects with external dependencies.

## AGENTS.md vs README

| Dimension | AGENTS.md | README |
|-----------|-----------|--------|
| Audience | AI Agent | Human users |
| Purpose | How to develop | How to use |
| Content | Architecture, conventions, workflow, indexes | Installation, examples, API overview |
| Style | Concise, imperative | Friendly, tutorial-style |
| Examples | Patterns and constraints | Complete runnable code |
| Commands | Development commands (test, lint, verify) | Installation commands (go get) |
| Detail Level | Index + links to SPECS/ | Self-contained usage guide |

### Collaboration Principles

1. **No Content Duplication**
   - AGENTS.md does not include installation instructions (README has them)
   - README does not include development conventions (AGENTS.md has them)

2. **Shared Terminology**
   - Both link to SPECS/ for term definitions
   - Maintain terminology consistency

3. **Cross-References**
   - README: "For development guidelines, see AGENTS.md"
   - AGENTS.md: "For usage examples, see README.md"

4. **Layered Detail**
   - README: Usage layer (how to call APIs)
   - AGENTS.md: Development layer (how to design and implement)
   - SPECS/: Contract layer (rules that must be followed)

## Process

### Phase 1: Project Analysis

Gather project context by reading these files (skip missing ones):

1. **Project manifest** — package.json, go.mod, Cargo.toml, etc. (dependencies, version)
2. **Build configuration** — Taskfile.yml, Makefile, package.json scripts
3. **README.md** — project description, features, usage
4. **DESIGN.md** or `docs/` — architecture decisions
5. **Directory structure** — `ls -R` top-level, identify key packages/modules
6. **Linter config** — .golangci.yml, .eslintrc, oxlint.json, etc.
7. **CI pipeline** — .github/workflows/, .gitlab-ci.yml
8. **`.claude/skills/`** and **`.agents/skills/`** — available agent skills (for index)
9. **`SPECS/`** directory — check if exists for SPECS Index
10. **`.references/`** directory — check if exists for References Index
11. **Source code sampling** — read 2-3 core files to understand:
   - Public API surface (exported types, functions, interfaces)
   - Design patterns used
   - Error handling approach
   - Testing patterns
   - Language-specific features used

### Phase 2: Content Generation

Generate CLAUDE.md following the template in [references/claude-md-template.md](references/claude-md-template.md).

**Section inclusion rules:**

| Section | When to Include |
|---------|----------------|
| Project Overview | Always |
| Commands | Always (detect from Makefile or use Go defaults) |
| Architecture | Always for multi-package projects; directory tree only, no type definitions |
| Agent Workflow | **REQUIRED** when project has `SPECS/` or `.references/` directories (not optional) |
| SPECS Index | When project has `SPECS/` directory |
| References Index | When project has `.references/` directory |
| Design Philosophy | When project has clear architectural principles — select 2-5 SOLID principles + 2-4 Craft Philosophy principles, customize with project-specific examples. Always include "Never" line |
| API Design Principles | When project is a library/SDK exposing public APIs — detect Progressive Disclosure and Default Passthrough patterns |
| Naming Conventions | When project has a naming spec or explicit naming rules |
| CLI Command Design | When project is a CLI tool with command pattern conventions |
| Coding Rules | Use layered approach: Universal Baseline (always) + Project-Specific (detected) + Domain-Specific (link to SPECS) |
| Testing | Always |
| Dependencies | When project has non-test external dependencies |
| Error Handling | When project uses sentinel errors or has specific error patterns |
| Performance | When project has benchmarks or performance requirements |
| Agent Skills | Always — index `.claude/skills/` and `.agents/skills/` with one-sentence trigger descriptions. Omit conditional skills that don't apply (see template) |

**Content principles:**

- **Philosophy over specifics** — AGENTS.md documents philosophy, principles, coding conventions, indexes, and workflow requirements. Specific design details (type definitions, design patterns, data formats) belong in `SPECS/` documents, not in AGENTS.md. Agents should be directed to read SPECS for design details.
- **Concise over verbose** — AI agents already know Go. Only document what's non-obvious or project-specific.
- **Actionable over aspirational** — Every instruction should be specific. Delete "write clean code" style platitudes.
- **Commands are critical** — Always include exact `task` commands or `go` commands. This is the highest-value section.
- **No specific design details** — Do not include type definitions, interface signatures, design pattern catalogs, or detailed CLI convention tables. These belong in `SPECS/` documents. AGENTS.md provides the index and directs agents to the right spec.
- **SPECS-driven design** — When the project has a `SPECS/` directory, include a SPECS Index table and an Agent Workflow section requiring agents to read relevant specs before designing or modifying code.
- **References-driven implementation** — When the project has a `.references/` directory, include a References Index and an Agent Workflow section requiring agents to find at least 2 reference projects before implementing code.
- **Progressive disclosure** — If AGENTS.md exceeds ~200 lines, consider linking to `.claude/` subdirectory files for detailed sections (testing guides, architecture deep-dives, performance rules).
- **No implementation phases** — AGENTS.md describes the project's current state, conventions, and rules. Never include implementation plans, roadmaps, migration phases, TODO lists, or step-by-step build instructions. Those belong in DESIGN.md, PLAN.md, or issue trackers.
- **Skills index** — Always include the Agent Skills section linking to `.claude/skills/` and `.agents/skills/`. See the Skills Index in [references/claude-md-template.md](references/claude-md-template.md).

## Agent Workflow — Critical Section

When project has `SPECS/` or `.references/` directories, Agent Workflow is a **REQUIRED section**, not optional.

### Why Critical
- Prevents Agent from skipping SPECS and writing code directly
- Ensures Agent learns patterns from References
- Establishes "understand first, implement second" workflow
- This is the core mechanism of SPEC-driven development

### Standard Workflow Template

```markdown
## Agent Workflow

### Design Phase — Read SPECS First

Before designing or modifying any code, **read the relevant SPECS/ documents first**.
SPECS define system contracts, data formats, naming rules, and architectural decisions.
Do not invent new patterns or conventions that contradict SPECS.

**Workflow**:
1. Identify which SPECS are relevant to your task (see SPECS Index below)
2. Read those SPECS completely
3. Design your solution following SPECS constraints
4. If SPECS are unclear or incomplete, ask the user before proceeding

### Implementation Phase — Find 2 References First

Before writing implementation code, **find at least 2 relevant reference projects in `.references/`**
to study their patterns, API design, and conventions. Browse the category directories, read their
source code, and adapt proven patterns rather than inventing from scratch.

**Workflow**:
1. Browse `.references/` categories (see References Index below)
2. Find 2+ projects relevant to your task
3. Study their implementation patterns
4. Adapt patterns to this project's conventions
5. If no relevant references exist, ask the user before inventing new patterns
```

### When to Include
- **Always** when SPECS/ directory exists
- **Always** when .references/ directory exists
- Even if SPECS/ or .references/ are empty — the workflow establishes the habit

## Design Philosophy — Selection and Customization

### Selection Rules

1. **Read source code first** — Analyze patterns actually used in the project
2. **Select 2-5 SOLID principles** — Not all 8, choose what the project truly demonstrates
3. **Select 2-4 Craft Philosophy principles** — Choose the ones that resonate with the project's character
4. **Customize explanation** — Don't copy generic definitions, write project-specific examples
5. **Always include the "Never" line** — accidental complexity, feature gravity, abstraction theater, configurability cope

### SOLID Principles Selection Criteria

| Principle | When to Select | How to Customize |
|-----------|---------------|------------------|
| KISS | Project avoids over-abstraction | Example: single representation, no sub-states, no parallel states |
| DRY | Rules shared across multiple places | Example: transition table serves Fire(), Build(), Mermaid() simultaneously |
| YAGNI | Project explicitly rejects features | Example: no Actor model, no JSON config, no Graphviz |
| SRP | Types have single responsibility | Example: Machine manages state, Store manages storage |
| OCP | Extension through interfaces | Example: StateStore interface, Guard/Action callbacks |
| LSP | Interface implementations substitutable | Example: all StateStore implementations fully substitutable |
| ISP | Small interfaces | Example: StateStore has only 2 methods |
| DIP | Depends on abstractions | Example: Machine depends on StateStore interface, not concrete implementation |

### Craft Philosophy Selection Criteria

| Principle | When to Select | How to Customize |
|-----------|---------------|------------------|
| Simplicity as art | Project has minimal API surface | Example: Fire() takes an event and returns an error — that's the entire mutation API |
| Precision over cleverness | Project has deliberate naming/typing | Example: `Guard` not `Validator`, `Action` not `Handler` — names match the domain |
| Elegance through reduction | Project actively removes unnecessary code | Example: no Stringer interface — ASCII() is the only representation |
| Craft over shipping | Project polishes API ergonomics | Example: fluent builder API reads like a specification |
| APIs as language | Public API reads like natural speech | Example: `machine.Fire(EventLock)` not `machine.ProcessEvent(EventConfig{Type: Lock})` |
| Errors as teachers | Error messages guide the developer | Example: "transition from Locked to Active via Unlock is not permitted" |
| Beauty is structural | Architecture is visible in the code | Example: package layout mirrors the domain model exactly |

### Bad Example (Generic)

```markdown
## Design Philosophy

- **KISS** — Keep it simple, stupid. Prefer simple solutions.
- **DRY** — Don't repeat yourself. Avoid duplication.
- **YAGNI** — You aren't gonna need it. Don't add unused features.
```

### Good Example (Project-Specific)

```markdown
## Design Philosophy

- **KISS** — Each concept has exactly one representation. No sub-states, no parallel states, no five kinds of trigger behaviour.
- **DRY** — The transition table serves Fire() dispatch, Build() validation, and Mermaid()/ASCII() export simultaneously.
- **YAGNI** — No Actor model, no JSON config, no Graphviz. Add when needed — the cost of adding later is far lower than maintaining unused features.
- **Simplicity as art** — Fire() takes an event and returns an error. That's the entire public API surface for state changes.
- **Errors as teachers** — "transition from Locked to Active via Unlock is not permitted" tells you the state, the event, and the rule.
- **Never:** accidental complexity, feature gravity, abstraction theater, configurability cope.
```

## API Design Principles — Detection and Inclusion

When the project exposes a public API (library, SDK, CLI), detect and include applicable API design principles in the generated CLAUDE.md. These are **separate from Design Philosophy** — philosophy governs internal code quality, API principles govern the consumer experience.

### Available Principles

| Principle | When to Include | One-liner for CLAUDE.md |
|-----------|----------------|------------------------|
| **Progressive Disclosure** | Project has both convenience functions and low-level APIs | Simple things one-liner, complex things possible. Ask "Do 80% of developers need this?" — if no, move it deeper. |
| **Default Passthrough** | Project uses functional options or config structs | Option zero values = preserve source attributes, not hardcoded defaults. |

### Detection Method

- **Progressive Disclosure**: Look for both high-level facade functions and low-level component APIs coexisting. Example: a top-level `Convert()` plus separate `Encoder`/`Decoder` interfaces.
- **Default Passthrough**: Look for `Option` structs or functional options where zero values are documented as "use default" or "preserve source".

### Section Format

```markdown
## API Design Principles

- **Progressive Disclosure**: Simple things one-liner, complex things possible. Ask "Do 80% of developers need this?" — if no, move it deeper.
- **Default Passthrough**: Option zero values = preserve source attributes, not hardcoded defaults.
```

### Inclusion Rules

- Include this section **only for library/SDK projects** that expose public APIs for other developers to consume
- Do NOT include for internal services, CLI tools without library components, or scripts
- Keep it concise — 1 line per principle, no lengthy explanations
- Place after Design Philosophy, before Coding Rules

## Coding Rules — Layered Approach

### Layer 1: Universal Baseline (Always Include)

```markdown
### Must Follow

- Go {version} — use modern language features where they simplify code
- Follow Google Go Best Practices: https://google.github.io/go-style/best-practices
- Follow Google Go Style Decisions: https://google.github.io/go-style/decisions
- KISS/DRY/YAGNI — no premature abstractions, no unused features, no duplicated logic
- Small interfaces — 1-3 methods per interface
- Explicit error handling — return errors, wrap with context via `fmt.Errorf("%w")`

### Forbidden

- No `panic` in production code (all errors returned via `error`)
- No premature abstraction — three similar lines are better than a helper used once
- No feature creep — only implement what's currently needed
```

### Layer 2: Project-Specific Rules (Add Only If Detected)

Detect through source code analysis, only add rules actually used by the project:

| Rule Type | Detection Method | Example |
|-----------|-----------------|---------|
| Context constraints | Search for `context.Context` parameters | "All public functions accept context.Context as first parameter" |
| Zero-allocation requirements | Check benchmark comments | "Zero heap allocations in hot path (Fire, Get, Set)" |
| No-panic policy | Search for `panic(` calls | "No panic in production code — return errors instead" |
| Generic constraints | Check generic type parameters | "State and Event must be comparable (for map keys)" |
| Concurrency safety | Check `sync.` usage | "All public methods are thread-safe via atomic.Pointer" |

### Layer 3: Domain-Specific Patterns (Link to SPECS)

Don't list in detail in AGENTS.md, link to SPECS:

```markdown
### Domain Patterns

See SPECS/ for detailed patterns:
- [SPECS/40-architecture-specs.md](SPECS/40-architecture-specs.md) — Module boundaries and dependency rules
- [SPECS/41-api-specs.md](SPECS/41-api-specs.md) — API design patterns
- [SPECS/42-data-model-specs.md](SPECS/42-data-model-specs.md) — Data model conventions
- [SPECS/43-coding-standards.md](SPECS/43-coding-standards.md) — Detailed coding rules
```

### Anti-Pattern: Bloated Coding Rules

❌ **Bad** (300+ lines of rules in AGENTS.md):
```markdown
### Must Follow

- Go 1.26
- All functions accept context.Context
- Use functional options for configuration
- All errors wrapped with fmt.Errorf
- All tests use t.Parallel()
- All benchmarks use b.Loop()
- Use slices.Clone for slice copying
- Use maps.Clone for map copying
- Use clear() for map/slice clearing
- ... (100+ more rules)
```

✅ **Good** (50 lines + links):
```markdown
### Must Follow

- Go 1.26 — use modern features (see Go 1.26 Features below)
- Follow Google Go Best Practices
- KISS/DRY/YAGNI
- Small interfaces (1-3 methods)
- Explicit error handling
- All tests use t.Parallel()
- All benchmarks use b.Loop()
- Use slices.Clone for slice copying

See [SPECS/43-coding-standards.md](SPECS/43-coding-standards.md) for detailed rules.
```

## AI Readability Optimization

Optimize AGENTS.md for efficient AI Agent reading:

### 1. Structured Information

- Use standard section headers (Commands, Architecture, Coding Rules)
- Use tables for indexes (SPECS Index, References Index, Agent Skills)
- Use lists for rules (Must Follow, Forbidden)

### 2. Key Information First

- Project Overview at the top (1-3 sentences)
- Commands immediately after (highest-value section)
- Agent Workflow before SPECS Index (establish workflow habit)

### 3. Index First, Details Linked

- SPECS Index provides quick overview
- Detailed content in SPECS/ files
- Agent selectively reads based on task

### 4. Avoid Ambiguity

- Use precise commands (see `references/commands.md` for language-specific examples)
- Specify language version explicitly
- Rules are verifiable

### 5. Context Minimalism

- Remove Go basics that AI already knows
- Document only project-specific conventions
- Load detailed content on-demand through indexes

**Language rules:**

- Write in English by default. Use the same language as existing project docs if they're consistently in another language.
- Use imperative form for instructions ("Use `t.Parallel()` in all tests", not "Tests should use `t.Parallel()`").
- No emojis unless existing project docs use them.

### Phase 3: Write Files

1. **Write `CLAUDE.md`** at project root
2. **Create `AGENTS.md` symlink** pointing to `CLAUDE.md`:

```bash
ln -sf CLAUDE.md AGENTS.md
```

3. **Verify** the symlink works:

```bash
ls -la AGENTS.md  # Should show AGENTS.md -> CLAUDE.md
```

### Phase 4: Validation

After generation, verify:

- [ ] Project overview matches actual project purpose
- [ ] All listed commands actually work
- [ ] Language version matches project manifest
- [ ] Module/package path is correct
- [ ] Directory structure matches reality
- [ ] No redundant instructions (things Claude already knows)
- [ ] No placeholder or TODO content remains
- [ ] AGENTS.md symlink resolves correctly
- [ ] Agent Workflow section included if SPECS/ or .references/ exists
- [ ] SPECS Index accurate and all files exist
- [ ] References Index accurate and all directories exist
- [ ] Agent Skills index lists only existing skills
- [ ] Design Philosophy has 2-5 principles with project-specific examples (not generic definitions)
- [ ] Coding Rules uses layered approach (not bloated with 50+ rules)

## Anti-Patterns

| Avoid | Why | Instead |
|-------|-----|---------|
| Including type definitions or interface signatures | Specific design details belong in SPECS/, not AGENTS.md | Add a SPECS Index table and direct agents to read specs |
| Including design pattern catalogs | Becomes stale, duplicates SPECS/ | Direct agents to architecture spec and source code |
| Detailed CLI convention tables | Too specific for AGENTS.md | Point to the CLI spec with a one-liner |
| Documenting standard Go practices | Wastes context tokens | Only document project-specific conventions |
| Listing every file in the project | Too verbose, becomes stale | Show key packages and their purpose |
| Copy-pasting entire API surfaces | Code is already readable | Agents should read the source code directly |
| Including setup/installation instructions | AGENTS.md is for AI, not humans | Put install docs in README.md |
| "Write clean, maintainable code" | Not actionable | Be specific: "No panics in production code" |
| Documenting obvious error handling | Go devs know `if err != nil` | Only document project-specific error patterns |
| Including implementation phases/roadmaps | AGENTS.md is current state, not future plans | Put phases in DESIGN.md or PLAN.md |
| Omitting Agent Skills index | AI agents lose access to shared skill library | Always include skills index for `.claude/skills/` and `.agents/skills/` |
| Omitting Agent Workflow section | Agents won't know to read SPECS or references | Always include when SPECS/ or .references/ exist |
| Encoding spec prose as code | Creates dead code no program consumes at runtime | Specs are the SSOT; code executes rules, not redescribes them |
| Duplicating README installation instructions | Responsibility overlap, high maintenance cost | Link to README |
| Working around dependency bugs inline | Creates hidden tech debt, redundant implementations that persist after the bug is fixed | Create a `reports/<dependency-name>.md` report and move on |
| Listing all 8 SOLID principles | Generic definitions add no value | Select 2-5 SOLID + 2-4 Craft Philosophy principles project actually demonstrates |
| Omitting Craft Philosophy | Misses the spirit behind the code | Include 2-4 craft principles + "Never" anti-patterns line |
| 50+ coding rules in AGENTS.md | Context bloat | Keep core rules, move detailed rules to SPECS/43-coding-standards.md |
| Listing unused conditional skills | Wastes context | Only list skills that actually exist and are relevant |
| Including type definitions in AGENTS.md | Concrete design details belong in SPECS/ | Provide SPECS Index, Agent reads on-demand |

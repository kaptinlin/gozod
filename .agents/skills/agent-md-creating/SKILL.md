---
name: agent-md-creating
description: Generate a CLAUDE.md (and AGENTS.md symlink) for Go projects. Analyzes go.mod, Makefile, source code, and existing docs to produce a comprehensive AI coding guide following progressive disclosure principles. Use when the user wants to create, generate, or initialize a CLAUDE.md or AGENTS.md for a Go repository, or asks for "claude md", "agent instructions", or "AI coding guide" for a Go project.
---

# Go CLAUDE.md Generator

Generate a comprehensive CLAUDE.md for Go projects by analyzing the repository and producing structured AI coding instructions. Also create an AGENTS.md symlink.

## Process

### Phase 1: Project Analysis

Gather project context by reading these files (skip missing ones):

1. **`go.mod`** — module path, Go version, dependencies
2. **`Makefile`** — available commands (test, lint, fmt, vet, bench, verify, clean, deps)
3. **`README.md`** — project description, features, usage
4. **`DESIGN.md`** or `docs/` — architecture decisions
5. **Directory structure** — `ls -R` top-level, identify key packages
6. **`.golangci.yml`** or `.golangci.version`** — lint config
7. **`.github/workflows/`** — CI pipeline
8. **Source code sampling** — read 2-3 core `.go` files to understand:
   - Public API surface (exported types, functions, interfaces)
   - Design patterns used (functional options, builder, strategy, etc.)
   - Error handling approach (sentinel errors, error wrapping, panics)
   - Testing patterns (testify vs stdlib, t.Parallel, golden tests, benchmarks)
   - Go version features used (generics, iterators, new builtins)

### Phase 2: Content Generation

Generate CLAUDE.md following the template in [references/claude-md-template.md](references/claude-md-template.md).

**Section inclusion rules:**

| Section | When to Include |
|---------|----------------|
| Project Overview | Always |
| Design Philosophy | When project has clear architectural principles or reference implementation |
| Commands | Always (detect from Makefile or use Go defaults) |
| Architecture | Always for multi-package projects; brief for single-file libraries |
| Key Types and Interfaces | When project has 3+ exported types or any interfaces |
| Coding Rules | When project has strong conventions (generics, zero-alloc, no-panic, etc.) |
| Testing | Always |
| Dependencies | When project has non-test external dependencies |
| Error Handling | When project uses sentinel errors or has specific error patterns |
| Performance | When project has benchmarks or performance requirements |
| Agent Skills | Always — index `.agents/skills/` with one-sentence trigger descriptions |

**Content principles:**

- **Concise over verbose** — Claude already knows Go. Only document what's non-obvious or project-specific.
- **Actionable over aspirational** — Every instruction should be specific. Delete "write clean code" style platitudes.
- **Commands are critical** — Always include exact `make` targets or `go` commands. This is the highest-value section.
- **Code examples earn their tokens** — Include code snippets only when they show project-specific patterns that can't be inferred from reading the code.
- **Progressive disclosure** — If CLAUDE.md exceeds ~200 lines, consider linking to `.claude/` subdirectory files for detailed sections (testing guides, architecture deep-dives, performance rules).
- **No implementation phases** — CLAUDE.md describes the project's current state, conventions, and rules. Never include implementation plans, roadmaps, migration phases, TODO lists, or step-by-step build instructions. Those belong in DESIGN.md, PLAN.md, or issue trackers.
- **Skills index** — Always include the Agent Skills section linking to `.agents/skills/`. See the Skills Index in [references/claude-md-template.md](references/claude-md-template.md).

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
- [ ] All listed commands actually work (`task test`, etc.)
- [ ] Go version matches `go.mod`
- [ ] Module path is correct
- [ ] Directory structure matches reality
- [ ] No redundant instructions (things Claude already knows about Go)
- [ ] No placeholder or TODO content remains
- [ ] AGENTS.md symlink resolves correctly

## Anti-Patterns

| Avoid | Why | Instead |
|-------|-----|---------|
| Documenting standard Go practices | Wastes context tokens | Only document project-specific conventions |
| Listing every file in the project | Too verbose, becomes stale | Show key packages and their purpose |
| Copy-pasting entire API surfaces | Code is already readable | Show only non-obvious API patterns |
| Including setup/installation instructions | CLAUDE.md is for AI, not humans | Put install docs in README.md |
| "Write clean, maintainable code" | Not actionable | Be specific: "No panics in production code" |
| Documenting obvious error handling | Go devs know `if err != nil` | Only document project-specific error patterns |
| Including implementation phases/roadmaps | CLAUDE.md is current state, not future plans | Put phases in DESIGN.md or PLAN.md |
| Omitting Agent Skills index | AI agents lose access to shared skill library | Always include `.agents/skills/` section |

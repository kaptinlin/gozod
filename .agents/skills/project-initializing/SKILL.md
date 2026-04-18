---
description: Initializes Go projects by orchestrating repository bootstrap skills. Uses ralphy-initializing for all Go projects, and uses multimodule-initializing first when the repository is multi-module. Use when initializing a Go project, bootstrapping project setup, or preparing a Go repository for implementation workflows.
name: project-initializing
---


# Project Initializing for Go

Initialize a new Go project using `gendog`, the project scaffold generator.

## Prerequisites

Verify gendog is installed:

```bash
which gendog
```

If not installed, guide the user through:

```bash
# 1. Configure GitHub token (add to ~/.zshrc or ~/.bashrc for persistence)
export HOMEBREW_GITHUB_API_TOKEN=ghp_your_token_here

# 2. Install via Homebrew
brew tap agentable/internal
brew install gendog
```

The `HOMEBREW_GITHUB_API_TOKEN` must be a valid GitHub personal access token with access to the `agentable/internal` tap. Remind the user to add the export to their shell profile for persistence.

## Step 1: Initialize gendog Workspace

If `.gendog/` does not exist in the workspace root:

```bash
gendog init
```

This creates `.gendog/` with default configuration.

Then clone the scaffold template:

```bash
cd .gendog
git clone https://github.com/agentable/golang-scaffold.git
cd ..
```

If `.gendog/golang-scaffold/` already exists, skip this step.

## Step 2: Gather Project Info

Collect the following from the user before generating:

| Variable | Description | Constraint | Example |
|----------|-------------|------------|---------|
| `ProjectName` | Package name | lowercase, hyphenated, `^[a-z][a-z0-9-]*$` | `go-fsm` |
| `ModulePath` | Full Go module path | `github.com/(agentable\|kaptinlin)/<name>` | `github.com/agentable/go-fsm` |
| `Description` | Short project description | required | `A finite state machine library` |
| `Author` | Author or organization | default: `Agentable` | `Agentable` |
| `GoVersion` | Minimum Go version | default: `1.26` | `1.26` |
| `GolangciVersion` | golangci-lint version | default: `2.11.4` | `2.11.4` |

## Step 3: Generate the Project

### Option A: Interactive mode (default)

```bash
gendog generate golang-scaffold
```

gendog will prompt for each variable interactively.

### Option B: Non-interactive mode

Pass all variables via `--set`:

```bash
gendog generate golang-scaffold \
  --set ProjectName=go-fsm \
  --set ModulePath=github.com/agentable/go-fsm \
  --set Description="A finite state machine library" \
  --set Author=Agentable \
  --set GoVersion=1.26 \
  --set GolangciVersion=2.11.4
```

### Option C: Values file

Create a `values.yaml` and pass it:

```yaml
ProjectName: go-fsm
ModulePath: github.com/agentable/go-fsm
Description: A finite state machine library
Author: Agentable
GoVersion: "1.26"
GolangciVersion: "2.11.4"
```

```bash
gendog generate golang-scaffold --values values.yaml
```

### Useful flags

- `--dry-run` — preview generated files without writing
- `-f` / `--force` — overwrite existing files without prompting
- `-o <dir>` — override output directory

### Generated Structure

```text
<ProjectName>/
├── doc.go                       # Package documentation
├── errors.go                    # Common error definitions
├── go.mod                       # Go module definition
├── Taskfile.yml                 # Task runner (lint, test, verify, vuln, etc.)
├── lefthook.yml                 # Git hooks (pre-commit: lint, test, gitleaks, markdownlint, yamllint)
├── .gitleaks.toml               # Gitleaks secret detection config
├── .markdownlint-cli2.yaml      # Markdownlint rules + ignores
├── .yamllint.yml                # YAML lint rules
├── .golangci.yml                # golangci-lint configuration
├── .golangci.version            # Pinned golangci-lint version
├── .gitignore                   # Standard Go gitignore
├── .editorconfig                # Editor configuration
├── .github/workflows/ci.yml    # CI: test, lint, security (PAT_TOKEN for private deps)
├── .ralphy/config.yaml          # Ralphy AI configuration
├── .references/.gitkeep         # Reference implementations directory
├── SPECS/.gitkeep               # Specifications directory
├── README.md                    # Project documentation
├── CLAUDE.md                    # Claude AI guidelines
└── LICENSE                      # MIT License
```

## Step 4: Post-Generation Setup

```bash
cd <ProjectName>

# Initialize git
git init

# Install pre-commit hooks
lefthook install

# Download dependencies
task deps

# Run full verification (includes govulncheck)
task verify
```

For **monorepo** projects, also integrate with workspace:

```bash
# Add to mani.yaml
# Edit mani.yaml to include the new project with path and tags

# Add as git submodule (if applicable)
cd ..
git submodule add https://github.com/agentable/<ProjectName>.git
```

### Configure PAT_TOKEN for CI

The generated CI workflow uses `PAT_TOKEN` for private Go module access. This secret **must** be configured in the GitHub repository:

1. Go to the repository's **Settings → Secrets and variables → Actions**
2. Click **New repository secret**
3. Name: `PAT_TOKEN`
4. Value: A GitHub personal access token with `repo` scope and read access to `agentable` private repos

**When is PAT_TOKEN required?**

- **Always** for packages under `github.com/agentable/*` that depend on other agentable packages
- The scaffold CI template includes PAT_TOKEN by default — if the package has no private deps, it still works (the token is simply unused)

**Without PAT_TOKEN**, CI will fail with:

```
fatal: could not read Username for 'https://github.com'
go: github.com/agentable/package@version: invalid version: git ls-remote failed
```

## Step 5: Setup Skills (Monorepo)

If the monorepo has `task setup:skills`, run that. Otherwise manually:

```bash
cd <ProjectName>
git submodule add https://github.com/agentable/golang-skills.git .agents/skills
mkdir -p .claude
ln -s "$(pwd)/.agents/skills" .claude/skills
```

## Step 6: Verify

Confirm:

1. All generated files exist in `<ProjectName>/`
2. `task verify` passes (deps, fmt, vet, lint, test, vuln)
3. `lefthook run pre-commit` passes (gitleaks, lint, test, markdownlint, yamllint)
4. `.ralphy/config.yaml` has correct project name and description
5. `CLAUDE.md` exists with project guidelines
6. GitHub repo has `PAT_TOKEN` secret configured (if package has agentable private deps)
7. `.github/workflows/ci.yml` includes `GOPRIVATE`, `GONOPROXY`, `GOPROXY` env vars and PAT_TOKEN git config step

## Available Task Commands

| Command | Description |
|---------|-------------|
| `task test` | Run tests with race detection |
| `task lint` | Run golangci-lint + tidy check |
| `task fmt` | Format Go code |
| `task vet` | Run go vet |
| `task verify` | Full verification (deps, fmt, vet, lint, test, vuln) |
| `task vuln` | Run govulncheck |
| `task markdownlint` | Run markdownlint (local only) |
| `task yamllint` | Run yamllint (local only) |
| `task deps` | Download and tidy dependencies |
| `task deps:update` | Update all dependencies |
| `task clean` | Clean build artifacts and caches |

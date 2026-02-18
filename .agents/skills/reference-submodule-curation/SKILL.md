---
name: curating-references
description: Finds high-quality maintained GitHub references for a user’s development goal and adds selected repos into the target project .references directory via git submodule. Use when users ask to search related libraries/frameworks and vendor them as reference submodules.
compatibility:
  - Requires git
  - Requires gh CLI authenticated for GitHub search
---

# Reference Submodule Curation

Find strong GitHub references for a requested goal, then add selected repos into `.references` as git submodules.

## Step 1: Clarify Inputs

Collect:
- Target project root path (contains `.references`)
- Development goal (what is being built)
- Preferred language/runtime (if any)
- Max number of references to add (default 3)

If unclear, ask once, then proceed.

## Step 2: Search Candidate Repositories

Use GitHub search first.

Example:
```bash
gh search repos "<goal keywords> language:Go" \
  --sort updated --order desc --limit 50 \
  --json nameWithOwner,url,description,stargazerCount,updatedAt,pushedAt,isArchived,license
```

If `gh` is unavailable, use web search and then validate each candidate on GitHub.

## Step 3: Apply Quality Gates

Select candidates using all gates below:
- Not archived
- Active maintenance (recent push/activity)
- Clear documentation and examples
- Suitable license for reference usage
- Evidence of adoption/quality (stars, issue quality, release hygiene)
- Direct relevance to requested goal

Reject repos that are stale, abandoned, or weakly related.

## Step 4: Propose and Confirm Selection

Before adding submodules, provide a short shortlist with:
- Repo URL
- Why it is relevant
- Why it passes quality gates

If user already specified exact repos, skip shortlist confirmation.

## Step 5: Add as Git Submodules

For each selected repository:

```bash
git -C <project_root> submodule add <repo_url> .references/<repo_name>
```

If target path exists, use a disambiguated folder like `.references/<owner>-<repo>`.

Then run:

```bash
git -C <project_root> submodule update --init --recursive
```

## Step 6: Verify and Report

Verify:
- `.gitmodules` includes the new entries
- Submodule directories exist under `.references`
- `git -C <project_root> submodule status` is clean and resolvable

Report:
- Added repos and paths
- Why each was selected
- Any skipped candidates and reason

## Constraints

- Do not modify product code in this workflow.
- Only add/update reference repositories under `.references`.
- Prefer fewer high-quality references over many weak ones.

# Dependency Selecting References

This folder is the source of truth for concern-specific dependency choices.

## KISS Baseline

1. Prefer stdlib when it is enough.
2. Keep one library per concern in one package.
3. Do not replace a library if it solves a different layer.
4. Add dependencies only when they remove repeated complexity.
5. Favor actively maintained, low-risk modules.

## Evaluation Checklist

- Problem: Is there a concrete pain point now?
- Scope: Is this app-level or only example/test-level?
- Complexity: Will this reduce code complexity overall?
- Runtime: Any extra services, C deps, or heavy transitive deps?
- Migration risk: Can we roll out incrementally?
- Dependency type: Is this a **direct dependency** we want to own in `go.mod` (not `// indirect`)?

## Env Config Rule

- `.env` loading: `github.com/joho/godotenv`
- env-to-struct binding: `github.com/caarlos0/env/v11`
- multi-source config (file + env + flags): `github.com/agentable/go-config`

These are complementary, not always replacements for each other.

## Resilience Rule

- Default resilience dependency: `github.com/failsafe-go/failsafe-go`.
- Do not mix `failsafe-go`, `go-retry`, and `cenkalti/backoff` in one project without measured need.
- For lightweight reconnect/poll loops, use a project-local `internal/backoff` helper first.

## Direct Dependency Rule

- Only recommend libraries intended as explicit direct dependencies.
- Do not recommend transitive-only modules (`// indirect`) as first-class selections.

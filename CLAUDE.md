# Claude Code Instructions

## Before Every Commit
Run `make check` and ensure it passes.

## Commit Message Format
Use conventional commits:
- `feat(scope): description` for new features
- `test(scope): description` for tests
- `fix(scope): description` for bug fixes
- `docs: description` for documentation
- `chore: description` for maintenance tasks

When implementing Harvest API calls, include the API reference URL in the commit body.

## Progress Tracking
Update PROGRESS.md after completing each step. Mark the step as complete and update "Current Step" to the next step.

## Test Style
Write tests in BDD style with descriptive names:
```go
t.Run("given X when Y then Z", func(t *testing.T) { ... })
```

## If Stuck
1. Document the blocker in PROGRESS.md under "Blockers"
2. Note what was attempted
3. Stop and wait for human input
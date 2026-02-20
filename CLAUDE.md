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

## Test Style
Write tests in BDD style with descriptive names:
```go
t.Run("given X when Y then Z", func(t *testing.T) { ... })
```

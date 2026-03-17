# mahin-cli — Application Specification

> **Version**: 1.1
> **Date**: 17-Mar-2026
> **Binary**: `mahin`
> **Language**: Go 1.22
> **Module**: `github.com/mahin/mahin-cli-v1`

## Scope

Cross-platform CLI tool for managing a personal movie and TV show library.
This is a **Go CLI project only** — no web frontend, no PHP, no WordPress.

## Architecture

See `SPEC.md` (root) for full command tree, schema, and TMDb integration.

### File Structure Rules

- One file per Cobra command: `movie_<cmd>.go`
- Helper files: `movie_<cmd>_helpers.go`
- Max ~200 lines per file; split at natural boundaries
- DB layer: `db.go` (connection), `media.go`, `config.go`, `history.go`, `helpers.go`

### Coding Guidelines

#### Function Naming

- Use descriptive verb-noun names: `fetchMovieDetails`, `saveHistoryLog`, `promptDestination`
- Exported functions: PascalCase (`GetMediaByID`)
- Unexported functions: camelCase (`expandHome`)
- Boolean-returning functions: prefix with `Is`/`Has`/`Can` (`IsVideoFile`)

#### Explicit Methods Over Boolean Flags

**Rule**: Do NOT add boolean parameters to toggle behavior inside a method.
Instead, create separate explicit methods.

❌ Bad:
```go
func logInfo(message string, enableTrace bool) { ... }
```

✅ Good:
```go
func logInfo(message string) { ... }
func logInfoWithTrace(message string) { ... }
```

This enforces single responsibility and makes tracing straightforward.

#### DRY Principle

- Never copy-paste a block that exists elsewhere — import it
- TMDb fetch: always use `fetchMovieDetails()` / `fetchTVDetails()` from `movie_info.go`
- Before writing new code, search for existing helpers

#### Cyclomatic Complexity

- Keep functions under 15 cyclomatic complexity
- Extract nested conditionals into named helper functions
- Use early returns to reduce nesting depth

#### Boolean Handling

- Use named constants or typed booleans for clarity
- Avoid bare `true`/`false` arguments in function calls
- Prefer explicit method variants over boolean parameters (see above)

### Error Handling

#### Error Messages

- User-facing errors: prefix with `❌` emoji
- Success messages: prefix with `✅` emoji
- Use `fmt.Fprintf(os.Stderr, ...)` for errors, `fmt.Println` for output

#### Error Propagation

- Return `error` from functions; handle at the command level
- Never silently swallow errors — at minimum log them
- Use `fmt.Errorf("context: %w", err)` for wrapping

#### Graceful Degradation

- TMDb API unavailable → scan still works, just without metadata
- Missing API key → warn and continue (scan) or exit with message (search)
- File not found on undo → report error, mark as undone anyway
- Cross-drive move → detect `os.Rename` failure, fallback to copy+delete (pending)

#### Edge Cases

- Empty scan folder → friendly message, not crash
- TMDb rate limiting → warn user, continue with partial results
- Duplicate TMDb IDs → UNIQUE constraint + UPDATE fallback (upsert)
- Unicode filenames → cleaner handles gracefully
- Network offline → TMDb calls fail gracefully
- Database locked → WAL mode mitigates; warn on concurrent access

### Logging

- Use `fmt.Println` / `fmt.Fprintf` for CLI output (no logging framework)
- Structured output with emoji prefixes
- File operation logs: `saveHistoryLog` writes RFC3339 timestamps to JSON
- Never use placeholder strings (`"now"`, `"TODO"`, `"test"`) in production paths

### Known Pitfalls & Prevention

| Pitfall | Prevention | Issue Ref |
|---|---|---|
| Hardcoded placeholder values | Grep for `"now"`, `"TODO"`, `"test"` before commit | `spec/02-app/issues/01-hardcoded-timestamp.md` |
| Duplicate logic across commands | Use shared helpers from `movie_info.go` | `spec/02-app/issues/02-duplicate-tmdb-fetch.md` |
| Files exceeding 200 lines | Split at natural boundaries proactively | `spec/02-app/issues/03-large-file-refactor.md` |
| Wrong project context (web vs CLI) | Read `01-project-overview.md` first | `spec/02-app/issues/04-wrong-project-context.md` |

## References

- Full specification: `SPEC.md` (root)
- Conventions: `.lovable/memory/02-conventions.md`
- AI success plan: `.lovable/memory/workflow/01-ai-success-plan.md`

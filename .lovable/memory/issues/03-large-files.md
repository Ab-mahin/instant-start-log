# Issue: Large Files Need Refactoring

> **Status**: Open  
> **Severity**: Low (maintainability)  
> **Files**: `cmd/movie_move.go` (348 lines), `db/sqlite.go` (452 lines)  
> **Iteration**: 0 (not yet fixed)

## Root Cause

Features were added incrementally without splitting files at natural boundaries.

## Solution

### `cmd/movie_move.go` → split into:
- `movie_move.go` — command definition + `runMovieMove` main flow
- `movie_move_helpers.go` — `promptSourceDirectory`, `promptDestination`, `promptCustomPath`, `listVideoFiles`, `humanSize`, `expandHome`, `saveHistoryLog`

### `db/sqlite.go` → split into:
- `db/db.go` — `DB` struct, `Open()`, `migrate()`
- `db/media.go` — `Media` struct, Insert/Update/Get/List/Search/Count methods
- `db/config.go` — `GetConfig`, `SetConfig`
- `db/history.go` — `MoveRecord`, `InsertMoveHistory`, `GetLastMove`, `MarkMoveUndone`, `InsertScanHistory`
- `db/helpers.go` — `scanMediaRows`, `splitCSV`, `split`, `indexOf`, `trim`

## Impact

- Harder to navigate and understand for new contributors/AI
- Higher risk of merge conflicts
- Lovable flags files >300 lines with refactoring warnings

## Learning

- Split files at natural module boundaries early
- One "concern" per file: schema, queries per entity, helpers
- Target ~200 lines max per file

## What Not to Repeat

- Don't let files grow past ~200 lines without evaluating a split

# Suggestions Tracker

> **Last Updated**: 17-Mar-2026

## 🔴 High Priority

### 1. Fix timestamp bug in move-log.json
- **Status**: Pending
- **File**: `cmd/movie_move.go` line 346
- **Issue**: `"timestamp":"now"` is a hardcoded string, not an actual timestamp
- **Fix**: Use `time.Now().Format(time.RFC3339)`

### 2. Add .gitignore
- **Status**: Pending
- **Content needed**:
  ```
  mahin
  mahin.exe
  mahin-darwin-*
  *.exe
  movie-cli-output/
  ```

## 🟡 Medium Priority

### 3. Implement `movie tag` command
- **Status**: Pending
- **Context**: `tags` table already exists in DB schema with `UNIQUE(media_id, tag)`
- **Subcommands**: `tag add <id> <tag>`, `tag remove <id> <tag>`, `tag list [id]`

### 4. Refactor large files
- **Status**: Pending
- **Targets**:
  - `cmd/movie_move.go` (348 lines) → split prompts/helpers into `cmd/movie_move_helpers.go`
  - `db/sqlite.go` (452 lines) → split into `db/schema.go`, `db/media.go`, `db/config.go`, `db/history.go`

### 5. Extract shared TMDb fetch logic
- **Status**: Pending
- **Context**: `scan`, `search`, and `info` all have duplicate code for fetching movie/TV details + credits
- **Solution**: Use `fetchMovieDetails()` and `fetchTVDetails()` from `movie_info.go` everywhere

## 🟢 Low Priority

### 6. Add `movie undo` confirmation prompt
- **Status**: Pending
- **Risk**: Currently undoes immediately without asking

### 7. Add file size stats to `movie stats`
- **Status**: Pending
- **Add**: Total library size, average file size, largest file

### 8. Cross-drive move support
- **Status**: Pending
- **Issue**: `os.Rename` fails across filesystems
- **Fix**: Detect error, fallback to copy+delete

### 9. Batch move (`--all` flag)
- **Status**: Pending
- **Purpose**: Move all video files from source at once

### 10. Update README.md
- **Status**: Pending
- **Current**: Only documents hello/version/self-update
- **Needed**: Document full movie management feature set

## ✅ Completed Suggestions

_(None yet — move items here as they're implemented)_

# Suggestions Tracker

> **Last Updated**: 17-Mar-2026

## ✅ Completed Suggestions

### 1. ~~Fix timestamp bug in move-log.json~~ ✅
- **Completed**: 17-Mar-2026
- **Fix**: Replaced `"now"` with `time.Now().Format(time.RFC3339)` in `cmd/movie_move_helpers.go`

### 2. ~~Refactor large files~~ ✅
- **Completed**: 17-Mar-2026
- **Changes**:
  - `cmd/movie_move.go` (348→178 lines) + new `cmd/movie_move_helpers.go` (168 lines)
  - `db/sqlite.go` (452 lines) → `db/db.go` + `db/media.go` + `db/config.go` + `db/history.go` + `db/helpers.go`

### 3. ~~Extract shared TMDb fetch logic~~ ✅
- **Completed**: 17-Mar-2026
- **Changes**: `movie_scan.go` and `movie_search.go` now call `fetchMovieDetails()`/`fetchTVDetails()` from `movie_info.go`

## 🟡 Medium Priority (Pending)

### 4. Add .gitignore
- **Status**: Must create manually (Lovable limitation)
- **Content**: See `01-suggestions.md` in completed folder for full content

### 5. Implement `movie tag` command
- **Status**: Pending
- **Context**: `tags` table already exists in DB schema with `UNIQUE(media_id, tag)`
- **Subcommands**: `tag add <id> <tag>`, `tag remove <id> <tag>`, `tag list [id]`

## 🟢 Low Priority (Pending)

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

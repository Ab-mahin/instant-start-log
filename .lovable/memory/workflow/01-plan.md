# Project Plan & Status

> **Last Updated**: 17-Mar-2026

## ✅ Completed

### Core CLI Structure
- [x] Root command with Cobra (`mahin`)
- [x] `hello` command with version display
- [x] `version` command with ldflags injection
- [x] `self-update` command via git pull --ff-only

### Movie Management Commands
- [x] `movie config` — get/set configuration with masked API key display
- [x] `movie scan` — folder scanning with TMDb metadata + poster download
- [x] `movie ls` — paginated list with interactive navigation + detail view
- [x] `movie search` — live TMDb search, select, save to DB
- [x] `movie info` — local DB lookup → TMDb fallback → auto-persist
- [x] `movie suggest` — genre-based recommendations + trending fallback
- [x] `movie move` — interactive browse, move, track history
- [x] `movie rename` — batch clean rename with undo tracking
- [x] `movie undo` — revert last move/rename operation
- [x] `movie play` — open file with system default player (cross-platform)
- [x] `movie stats` — counts, genre chart, average ratings

### Infrastructure
- [x] SQLite database with migrations (5 tables, 7 indexes)
- [x] TMDb API client (search, details, credits, recommendations, trending, posters)
- [x] Filename cleaner (junk removal, year extraction, TV detection, slugs)
- [x] Makefile with build + cross-compile targets
- [x] build.ps1 PowerShell deploy script (Windows E:\bin-run, Mac /usr/local/bin)
- [x] SPEC.md — full project specification
- [x] Shared resolver helper (`movie_resolve.go`)

### Documentation
- [x] README.md (basic)
- [x] SPEC.md (comprehensive — 350+ lines)
- [x] readm.txt milestone marker (updated 17-Mar-2026 09:45 PM MYT)
- [x] .lovable/memory structure

## 🔲 Pending / Not Started

### Known Bugs
- [ ] `move-log.json` timestamp hardcoded as `"now"` instead of actual time (see issues/01-timestamp-bug.md)
- [ ] No confirmation prompt on `movie undo` before reverting

### Missing Features
- [ ] `movie tag` command — tags table exists but no commands use it
- [ ] `DiscoverByGenre` TMDb method exists but unused by any command
- [ ] JSON metadata files per movie/TV show (directories exist, not written)
- [ ] `.gitignore` file — attempted but could not be created in Lovable environment (must create manually)

### Refactoring Needed
- [ ] `cmd/movie_move.go` — 348 lines, should split into smaller files (see issues/03-large-files.md)
- [ ] `db/sqlite.go` — 452 lines, should split (see issues/03-large-files.md)
- [ ] Duplicate TMDb fetch logic in `scan`, `search`, `info` — should extract shared function (see issues/02-duplicate-tmdb-fetch.md)

### Enhancements
- [ ] Cross-drive move support (`os.Rename` fails across filesystems — need copy+delete)
- [ ] File size statistics in `movie stats`
- [ ] Batch move (`--all` flag for `movie move`)
- [ ] Update README.md to reflect full movie management features

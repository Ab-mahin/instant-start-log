# Project Overview

> **Last Updated**: 17-Mar-2026

## Project

- **Name**: mahin-cli-v1
- **Type**: Go CLI application (NOT a web app)
- **Binary**: `mahin`
- **Language**: Go 1.22
- **Module**: `github.com/mahin/mahin-cli-v1`
- **Framework**: Cobra (CLI), SQLite (storage), TMDb API (metadata)

## Purpose

A cross-platform CLI tool for managing a personal movie and TV show library. It scans local folders for video files, cleans messy filenames, fetches metadata from TMDb, stores everything in SQLite, and organizes files into configured directories.

## Key Architecture Decisions

1. **Pure-Go SQLite** (`modernc.org/sqlite`) — no CGo dependency
2. **WAL mode** for SQLite concurrency
3. **TMDb API** for metadata (requires user-provided API key)
4. **git-based self-update** (`git pull --ff-only`)
5. **All data** stored in `~/movie-cli-output/` (DB, thumbnails, JSON logs)

## Command Tree

```
mahin
├── hello                      # Greeting with version
├── version                    # Version/commit/build date
├── self-update                # git pull --ff-only
└── movie
    ├── config                 # View/set configuration
    ├── scan                   # Scan folder → DB + TMDb
    ├── ls                     # Paginated library list
    ├── search                 # Live TMDb search → save
    ├── info                   # Local DB → TMDb fallback
    ├── suggest                # Recommendations/trending
    ├── move                   # Browse + move + track
    ├── rename                 # Batch clean rename
    ├── undo                   # Revert last move/rename
    ├── play                   # Open with default player
    └── stats                  # Library statistics
```

## Important Notes for AI

- **This is NOT a web project** — no `package.json`, no dev server, no preview
- Build errors in Lovable are expected and can be ignored
- All file operations require a real OS/terminal to test
- Full specification lives in `SPEC.md` at project root
- Milestone markers use `readm.txt` format: `let's start now {date} {time Malaysia}`

## File Counts (as of 17-Mar-2026)

- `cmd/` — 14 Go files (root + hello + version + update + movie parent + 10 subcommands)
- `cleaner/` — 1 file (filename cleaning)
- `tmdb/` — 1 file (API client)
- `db/` — 1 file (SQLite schema + CRUD)
- `updater/` — 1 file (git self-update)
- `version/` — 1 file (build-time vars)

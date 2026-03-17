# mahin-cli-v1

A cross-platform CLI tool for managing a personal movie and TV show library. Scans local folders for video files, cleans messy filenames, fetches metadata from TMDb, stores everything in SQLite, and organizes files into configured directories.

## Features

- **Scan** local folders for video files and auto-fetch metadata from TMDb
- **Search** TMDb directly and save results to your library
- **List** your library with paginated, interactive browsing
- **Move** and **rename** files with clean naming (`Title (Year).ext`)
- **Undo** any move/rename operation
- **Play** media files with your system's default player
- **Suggest** new content based on your library's genre patterns or trending
- **Stats** with genre charts and average ratings
- **Self-update** via `git pull --ff-only`

## Commands

```
mahin
├── hello                         # Greeting with version
├── version                       # Version, commit, build date
├── self-update                   # git pull --ff-only
└── movie
    ├── config [get|set] [key]    # View/set configuration
    ├── scan [folder]             # Scan folder → DB + TMDb metadata
    ├── ls                        # Paginated library list
    ├── search <name>             # Live TMDb search → save to DB
    ├── info <id|title>           # Detail view (local DB → TMDb fallback)
    ├── suggest [N]               # Recommendations + trending
    ├── move [directory]          # Browse, select, move with clean name
    ├── rename                    # Batch rename to clean format
    ├── undo                      # Revert last move/rename
    ├── play <id>                 # Open with default video player
    └── stats                     # Counts, genre chart, avg ratings
```

## Quick Start

```bash
# Build
make build

# Or use PowerShell (builds + deploys)
pwsh build.ps1

# Set your TMDb API key
mahin movie config set tmdb_api_key YOUR_KEY

# Scan a folder
mahin movie scan ~/Downloads

# Browse your library
mahin movie ls

# Search TMDb directly
mahin movie search "Inception"

# Get suggestions
mahin movie suggest 5
```

## Configuration

```bash
mahin movie config                          # Show all settings
mahin movie config set movies_dir ~/Movies  # Set movie destination
mahin movie config set scan_dir ~/Downloads  # Set default scan folder
mahin movie config set page_size 30         # Items per page
```

| Key | Default | Purpose |
|---|---|---|
| `movies_dir` | `~/Movies` | Movie file destination |
| `tv_dir` | `~/TVShows` | TV show destination |
| `archive_dir` | `~/Archive` | Archive destination |
| `scan_dir` | `~/Downloads` | Default scan source |
| `tmdb_api_key` | *(none)* | TMDb API key |
| `page_size` | `20` | Items per page in `ls` |

## Project Structure

```
mahin-cli-v1/
├── main.go                        # Entry point
├── cmd/                           # Cobra commands (one file per command)
│   ├── root.go                    # Root command, registers subcommands
│   ├── hello.go                   # mahin hello
│   ├── version.go                 # mahin version
│   ├── update.go                  # mahin self-update
│   ├── movie.go                   # Parent: mahin movie
│   ├── movie_config.go            # config get/set
│   ├── movie_scan.go              # scan folder
│   ├── movie_ls.go                # paginated list
│   ├── movie_search.go            # TMDb search
│   ├── movie_info.go              # detail view + shared fetch helpers
│   ├── movie_suggest.go           # recommendations
│   ├── movie_move.go              # interactive move
│   ├── movie_move_helpers.go      # move utility functions
│   ├── movie_rename.go            # batch rename
│   ├── movie_undo.go              # undo last move/rename
│   ├── movie_play.go              # open with default player
│   ├── movie_stats.go             # library statistics
│   └── movie_resolve.go           # shared ID/title resolver
├── cleaner/cleaner.go             # Filename cleaning + slug generation
├── tmdb/client.go                 # TMDb API client
├── db/
│   ├── db.go                      # SQLite connection + migrations
│   ├── media.go                   # Media CRUD operations
│   ├── config.go                  # Config get/set
│   ├── history.go                 # Move history + scan history
│   └── helpers.go                 # String utilities
├── updater/updater.go             # Git-based self-update
├── version/version.go             # Build-time version variables
├── Makefile                       # Build targets
├── build.ps1                      # PowerShell build + deploy
├── SPEC.md                        # Full project specification
└── spec/                          # Specs and issue tracking
    ├── 01-app/                    # Application specs
    └── 02-app/issues/             # Issue write-ups
```

## Build

```bash
make build              # Current OS
make build-windows      # Windows amd64
make build-mac-arm      # macOS ARM64
make build-mac-intel    # macOS amd64
make build-linux        # Linux amd64
make install            # Build + copy to /usr/local/bin
```

## Dependencies

| Package | Purpose |
|---|---|
| `github.com/spf13/cobra` | CLI framework |
| `modernc.org/sqlite` | Pure-Go SQLite driver (no CGo) |

## Data Storage

All data lives in `~/movie-cli-output/`:

```
~/movie-cli-output/
├── mahin.db              # SQLite database (WAL mode)
├── thumbnails/           # Downloaded poster images
└── json/history/         # Move operation logs (RFC3339 timestamps)
```

## License

Private project.

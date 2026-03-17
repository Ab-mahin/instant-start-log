# Issue: Duplicate TMDb Fetch Logic

> **Status**: Open  
> **Severity**: Low (code quality)  
> **Files**: `cmd/movie_scan.go`, `cmd/movie_search.go`, `cmd/movie_info.go`  
> **Iteration**: 0 (not yet fixed)

## Root Cause

Three commands (`scan`, `search`, `info`) each contain nearly identical code for:
1. Fetching movie details + credits from TMDb
2. Extracting directors, cast (top 10), genres
3. Handling TV `Executive Producer` as director
4. Downloading poster thumbnails

`movie_info.go` has properly extracted helpers (`fetchMovieDetails`, `fetchTVDetails`) but `scan` and `search` don't use them.

## Solution

Refactor `scan` and `search` to call the shared `fetchMovieDetails()` and `fetchTVDetails()` functions from `movie_info.go`. These are already package-level functions in the `cmd` package.

## Impact

- ~80 lines of duplicate code across 3 files
- Bug fixes must be applied in 3 places
- Risk of behavior divergence between commands

## Learning

- Extract shared logic BEFORE duplicating into a second command
- When adding a third copy of the same pattern, stop and refactor immediately

## What Not to Repeat

- Don't copy-paste TMDb fetch blocks — always use the shared helpers
- When adding new metadata fields, update the shared functions, not individual commands

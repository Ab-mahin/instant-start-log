# Issue: Hardcoded Timestamp in move-log.json

> **Status**: Open  
> **Severity**: Medium  
> **File**: `cmd/movie_move.go`, line 345-346

## Root Cause

The `saveHistoryLog` function writes `"timestamp":"now"` as a literal string instead of an actual timestamp.

```go
entry := fmt.Sprintf(`{"from":"%s","to":"%s","timestamp":"%s"}`+"\n",
    from, to, "now")  // ← hardcoded "now"
```

## Solution

Replace `"now"` with `time.Now().Format(time.RFC3339)`:

```go
entry := fmt.Sprintf(`{"from":"%s","to":"%s","timestamp":"%s"}`+"\n",
    from, to, time.Now().Format(time.RFC3339))
```

Also need to add `"time"` to the imports.

## Impact

- All move history JSON logs have useless timestamp data
- Cannot reconstruct when moves happened from JSON logs
- DB `move_history.moved_at` is correct (uses `CURRENT_TIMESTAMP`), so DB is unaffected

## Learning

- Always use actual time functions, never placeholder strings
- Review all format strings for hardcoded test values before committing

## What Not to Repeat

- Don't use placeholder strings (`"now"`, `"TODO"`, `"test"`) in production code paths
- Always grep for placeholder strings before marking a feature complete

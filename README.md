# mahin-cli-v1

Simple CLI with four commands:

1. `mahin hello`
2. `mahin version`
3. `mahin self-update`
4. `mahin help`

## How `self-update` works

`mahin self-update` pulls the latest files from the current cloned git repository.

It runs:

1. `git rev-parse --show-toplevel`
2. `git status --porcelain` (repository must be clean)
3. `git pull --ff-only`

This updates your local clone files from remote.

## Project Structure

```text
mahin-cli-v1/
├── cmd/
│   ├── hello.go
│   ├── root.go
│   ├── update.go      (self-update command)
│   └── version.go
├── updater/
│   └── updater.go
├── version/
│   └── version.go
├── main.go
├── go.mod
└── go.sum
```

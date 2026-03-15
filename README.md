# mahin-cli-v1
Self-updating Go CLI with two update sources:
1. GitHub Releases for stable semver-based updates
2. `main` branch artifacts for fast post-push updates without cutting a new release tag

## Commands

```bash
mahin hello
mahin version
mahin update
```

## Update Strategy

`mahin update` works in this order:

1. Check `https://api.github.com/repos/<owner>/<repo>/releases/latest`
2. If the latest release semver is newer than the current binary, update from release assets
3. If the release semver is not newer, fetch branch metadata from `dist/main.json` on `main`
4. If the branch build is the same semver line or newer and the embedded commit is different, update from raw `dist/` branch artifacts

This lets normal users stay on tagged releases while also allowing the repository owner to push code and get immediate self-updates from `main` after branch artifacts are refreshed.

## Project Structure

```text
mahin-cli-v1/
├── .github/workflows/
│   ├── branch-artifacts.yml  Build branch artifacts on push to main and commit dist/
│   └── release.yml           Build and publish release assets when a v* tag is pushed
├── cmd/
│   ├── hello.go
│   ├── root.go
│   ├── update.go
│   └── version.go
├── config/
│   └── config.go
├── dist/
│   ├── mahin-darwin-arm64
│   ├── mahin-darwin-arm64.sha256
│   └── main.json             Branch build metadata consumed by branch fallback
├── updater/
│   ├── github.go             GitHub API, branch metadata, raw asset URL helpers
│   ├── platform.go           OS/arch detection and asset name generation
│   ├── process.go            Detached child-process builder
│   ├── process_unix.go       Unix process settings
│   ├── process_windows.go    Windows process settings
│   └── updater.go            Parent flow, child flow, verification, replace logic
├── version/
│   └── version.go
├── main.go
└── mahin
```

## Verification

### 1. Static checks

```bash
go build ./...
go test ./...
go vet ./...
```

### 2. Command smoke tests

```bash
./mahin version
./mahin hello
./mahin update --help
```

### 3. Release API validation

```bash
curl -s https://api.github.com/repos/Ab-mahin/mahin-cli-v1/releases/latest | python3 -c "import sys, json; data = json.load(sys.stdin); print(data.get('tag_name')); print([a.get('name') for a in data.get('assets', [])])"
```

### 4. Branch artifact validation

```bash
curl -s https://raw.githubusercontent.com/Ab-mahin/mahin-cli-v1/main/dist/main.json
curl -I https://raw.githubusercontent.com/Ab-mahin/mahin-cli-v1/main/dist/mahin-darwin-arm64
curl -I https://raw.githubusercontent.com/Ab-mahin/mahin-cli-v1/main/dist/mahin-darwin-arm64.sha256
```

### 5. Checksum verification

```bash
shasum -a 256 dist/mahin-darwin-arm64
shasum -c dist/mahin-darwin-arm64.sha256
```

### 6. End-to-end release update test

```bash
go build -ldflags "-X github.com/mahin/mahin-cli-v1/version.Version=v0.0.1 -X github.com/mahin/mahin-cli-v1/version.Commit=oldtest" -o /tmp/mahin-release-test .
/tmp/mahin-release-test version
/tmp/mahin-release-test update
sleep 8
/tmp/mahin-release-test version
```

### 7. End-to-end branch update test

Run this after pushing to `main` and waiting for the `branch-artifacts` workflow to finish:

```bash
./mahin version
./mahin update
sleep 8
./mahin version
```

If the release version has not changed but `main` has a new branch artifact commit, the updater should move to the branch build.

## Automation

### Branch artifacts workflow

`.github/workflows/branch-artifacts.yml`:

1. Runs on push to `main`
2. Builds `darwin`, `linux`, and `windows` binaries
3. Generates checksum files
4. Writes `dist/main.json`
5. Commits updated `dist/` files back to `main`

### Release workflow

`.github/workflows/release.yml`:

1. Runs when a tag matching `v*` is pushed
2. Builds multi-platform binaries and checksums
3. Publishes all assets to GitHub Releases automatically

## Release Flow

```bash
git tag v0.0.3
git push origin main
git push origin v0.0.3
```

That tag push triggers the release workflow and publishes fresh assets for self-update.

## Branch Flow

```bash
git push origin main
```

That push triggers the branch-artifacts workflow, updates `dist/`, and enables branch-based self-update without a new release tag.
# mahin-cli-v1

<<<<<<< HEAD
A self-updating CLI tool with GitHub Releases integration.
=======
Self-updating Go CLI with two update sources:

1. Published GitHub Releases for stable versioned updates
2. `main` branch artifacts for faster post-push updates without cutting a release

## Commands

```bash
mahin hello
mahin version
mahin update
```

## Update Strategy

`mahin update` works in this order:

1. Check `releases/latest` on GitHub
2. If the latest release tag is newer than the current binary version, update from release assets
3. If release semver is not newer, check `main` branch metadata in `dist/main.json`
4. If the `main` branch build is the same semver line or newer and has a different commit, update from raw branch artifacts in `dist/`

That means normal users can update from stable releases, while you can also push new code plus fresh branch artifacts and get updates immediately from `main`.
>>>>>>> 9c2f338 (Add branch fallback and release automation)

## Project Structure

```text
mahin-cli-v1/
├── .github/workflows/
│   ├── branch-artifacts.yml  Build branch artifacts on push to main and commit dist/
│   └── release.yml           Build and publish release assets when a v* tag is pushed
├── cmd/
│   ├── hello.go
│   ├── root.go
│   ├── update.go
│   └── version.go
├── config/
│   └── config.go
├── dist/
│   ├── mahin-darwin-arm64
│   ├── mahin-darwin-arm64.sha256
│   └── main.json             Branch build metadata consumed by self-update fallback
├── updater/
│   ├── github.go
│   ├── platform.go
│   ├── process.go
│   ├── process_unix.go
│   ├── process_windows.go
│   └── updater.go
├── version/
<<<<<<< HEAD
│   └── version.go            3 build-time vars (Version/Commit/BuildDate) + Full()/Short()
│
├── cmd/                      One file per command, zero business logic
│   ├── root.go               Registers all subcommands, exposes Execute()
│   ├── hello.go              `mahin hello`
│   ├── version.go            `mahin version`
│   └── update.go             `mahin update` -> calls updater.Run()
│
└── updater/                  All update logic, split by responsibility
    ├── updater.go            Orchestrator: parent flow + child flow, semver compare
    ├── github.go             GitHub API: fetch latest release, find asset URLs
    ├── platform.go           OS/arch detection at runtime -> builds asset filename
    ├── download.go           Download file to disk + SHA-256 checksum verify
    ├── replace.go            Verify binary + atomic swap of old -> new executable
    ├── process.go            Builds the OS-specific detached child command
    ├── process_unix.go       Unix: no extra flags needed
    └── process_windows.go    Windows: CREATE_NEW_PROCESS_GROUP to detach child
```
## Available Commands

```bash
mahin hello    # Print a greeting
mahin version  # Show current version, commit, and build date
mahin update   # Check GitHub for a newer release and self-update
```

## Verification

### 1. Static Analysis

```bash
# Compile all packages
go build -v ./...

# Run static analysis
go vet ./...

# Build the binary
go build -o mahin .
```

### 2. Test All Commands

```bash
# Test version command
./mahin version

# Test hello command
./mahin hello

# Test update help
./mahin update --help
```

### 3. GitHub API Validation

```bash
# Check if release endpoint is accessible
curl -s https://api.github.com/repos/Ab-mahin/mahin-cli-v1/releases/latest | python3 -c "import sys, json; r = json.load(sys.stdin); print(f'Tag: {r.get(\"tag_name\")}'); print(f'Assets: {len(r.get(\"assets\", []))} files')"
```

### 4. Platform Detection

```bash
# Check OS and architecture
go env GOOS GOARCH

# Verify expected asset name matches
echo "mahin-$(go env GOOS)-$(go env GOARCH)"
```

### 5. Checksum Verification

```bash
# Generate checksum for a binary
shasum -a 256 dist/mahin-darwin-arm64

# Verify checksum matches
shasum -c dist/mahin-darwin-arm64.sha256
```

### 6. End-to-End Update Test

```bash
# Build an older version binary
go build -ldflags "-X github.com/mahin/mahin-cli-v1/version.Version=v0.0.1" -o /tmp/mahin-test .

# Check initial version
/tmp/mahin-test version

# Run update
/tmp/mahin-test update

# Wait for child process to complete
sleep 10

# Verify updated version
/tmp/mahin-test version
```

### 7. Complete Verification Flow

```bash
echo "=== BEFORE UPDATE ==="
./mahin version

echo ""
echo "=== RUNNING UPDATE ==="
./mahin update

echo ""
echo "=== WAITING FOR CHILD PROCESS ==="
sleep 8

echo ""
echo "=== AFTER UPDATE ==="
./mahin version

echo ""
echo "=== TEST ALL COMMANDS ==="
./mahin hello
./mahin version
./mahin update  # Should show "Already up to date"
```

## Publishing a Release

### Prerequisites

```bash
# Install GitHub CLI (macOS)
brew install gh

# Authenticate
gh auth login
```

### Build Release Assets

```bash
# Set version and build metadata
VERSION="v0.0.2"
BUILD_DATE="$(date -u +%Y-%m-%d)"

# Create dist directory
mkdir -p dist

# Build for macOS ARM64
GOOS=darwin GOARCH=arm64 go build -ldflags "\
  -X github.com/mahin/mahin-cli-v1/version.Version=${VERSION} \
  -X github.com/mahin/mahin-cli-v1/version.Commit=$(git rev-parse --short HEAD) \
  -X github.com/mahin/mahin-cli-v1/version.BuildDate=${BUILD_DATE}" \
  -o dist/mahin-darwin-arm64 .

# Generate checksum
shasum -a 256 dist/mahin-darwin-arm64 > dist/mahin-darwin-arm64.sha256

# Verify binary
./dist/mahin-darwin-arm64 version
```

### Build for Multiple Platforms

```bash
# macOS Intel
GOOS=darwin GOARCH=amd64 go build -ldflags "..." -o dist/mahin-darwin-amd64 .
shasum -a 256 dist/mahin-darwin-amd64 > dist/mahin-darwin-amd64.sha256

# Linux AMD64
GOOS=linux GOARCH=amd64 go build -ldflags "..." -o dist/mahin-linux-amd64 .
shasum -a 256 dist/mahin-linux-amd64 > dist/mahin-linux-amd64.sha256

# Linux ARM64
GOOS=linux GOARCH=arm64 go build -ldflags "..." -o dist/mahin-linux-arm64 .
shasum -a 256 dist/mahin-linux-arm64 > dist/mahin-linux-arm64.sha256

# Windows AMD64
GOOS=windows GOARCH=amd64 go build -ldflags "..." -o dist/mahin-windows-amd64.exe .
shasum -a 256 dist/mahin-windows-amd64.exe > dist/mahin-windows-amd64.exe.sha256
```

### Create and Publish Release

```bash
# Create release with assets
gh release create v0.0.2 \
  dist/mahin-darwin-arm64 \
  dist/mahin-darwin-arm64.sha256 \
  dist/mahin-darwin-amd64 \
  dist/mahin-darwin-amd64.sha256 \
  dist/mahin-linux-amd64 \
  dist/mahin-linux-amd64.sha256 \
  dist/mahin-linux-arm64 \
  dist/mahin-linux-arm64.sha256 \
  dist/mahin-windows-amd64.exe \
  dist/mahin-windows-amd64.exe.sha256 \
  --title "v0.0.2" \
  --notes "Release notes here"

# Verify release is published
curl -s https://api.github.com/repos/Ab-mahin/mahin-cli-v1/releases/latest | grep tag_name
```

## Configuration

Edit `config/config.go` to customize:

```go
const (
    GitHubOwner = "Ab-mahin"           // Your GitHub username
    GitHubRepo  = "mahin-cli-v1"       // Your repository name
    BinaryName  = "mahin"              // Base name for assets
)
```

## How Self-Update Works

1. **Parent Process** (`mahin update`):
   - Checks current version
   - Queries GitHub Releases API for latest version
   - Compares versions using semver
   - Detects OS/architecture
   - Finds matching binary asset
   - Spawns child updater process
   - Exits immediately (releases file lock)

2. **Child Process** (`mahin --internal-updater ...`):
   - Waits for parent to exit
   - Downloads new binary to temp directory
   - Downloads and verifies SHA-256 checksum
   - Verifies new binary runs correctly
   - Atomically replaces old binary with new one
   - Cleans up temp files
   - Exits

This approach works around OS file-locking (especially Windows) by ensuring the running binary is not active when it gets replaced.

=======
│   └── version.go
├── main.go
└── mahin
```

## Release Flow

Push a tag like `v0.0.3` and the release workflow will build multi-platform assets and publish them to GitHub Releases.

## Branch Artifact Flow

Push to `main` and the branch-artifacts workflow will rebuild `dist/` and commit fresh branch-tracked binaries plus `dist/main.json`.

Those branch artifacts are what the updater uses when the release version has not changed yet but the `main` branch commit has.
>>>>>>> 9c2f338 (Add branch fallback and release automation)

package updater

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/mahin/mahin-cli-v1/config"
	"github.com/mahin/mahin-cli-v1/version"
)

const InternalUpdaterFlag = "--internal-updater"

const (
	maxBinarySize   = 200 * 1024 * 1024 // 200MB safety limit
	downloadTimeout = 5 * time.Minute
)

type childArgs struct {
	execPath     string
	binaryURL    string
	checksumURL  string
	assetName    string
	expectVer    string
	expectCommit string
	updateLabel  string
}

type Result struct {
	AlreadyLatest   bool
	PreviousVersion string
	UpdatedTo       string
}

type updateCandidate struct {
	binaryURL    string
	checksumURL  string
	assetName    string
	expectVer    string
	expectCommit string
	updateLabel  string
}

func Run() (*Result, error) {
	execPath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("cannot determine executable path: %w", err)
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return nil, fmt.Errorf("cannot resolve symlinks: %w", err)
	}
	fmt.Printf("📍 Current binary  : %s\n", execPath)

	currentVersion := version.Short()
	currentCommit := version.Commit
	fmt.Printf("📦 Current version : %s\n", currentVersion)

	currentSemver, err := parseSemver(currentVersion)
	if err != nil {
		return nil, fmt.Errorf("invalid current version %q: %w", currentVersion, err)
	}

	plat := detect()
	fmt.Printf("🖥️  Platform         : %s/%s\n", plat.OS, plat.Arch)

	fmt.Println("🔍 Checking for updates...")
	candidate, err := resolveUpdateCandidate(currentVersion, currentCommit, currentSemver, plat)
	if err != nil {
		return nil, err
	}
	if candidate == nil {
		fmt.Println("✅ Already up to date!")
		return &Result{AlreadyLatest: true, PreviousVersion: currentVersion}, nil
	}

	fmt.Println("🚀 Launching updater process...")
	args := childArgs{
		execPath:     execPath,
		binaryURL:    candidate.binaryURL,
		checksumURL:  candidate.checksumURL,
		assetName:    candidate.assetName,
		expectVer:    candidate.expectVer,
		expectCommit: candidate.expectCommit,
		updateLabel:  candidate.updateLabel,
	}
	if err := spawnChild(args); err != nil {
		return nil, fmt.Errorf("failed to launch updater process: %w", err)
	}

	fmt.Println("⏳ Updater is running in the background...")
	os.Exit(0)

	return &Result{PreviousVersion: currentVersion, UpdatedTo: candidate.updateLabel}, nil
}

func spawnChild(args childArgs) error {
	self, err := os.Executable()
	if err != nil {
		return err
	}

	cmd := buildCommand(self,
		InternalUpdaterFlag,
		args.execPath,
		args.binaryURL,
		args.checksumURL,
		args.assetName,
		args.expectVer,
		args.expectCommit,
		args.updateLabel,
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Start()
}

func RunChild(args []string) {
	if len(args) < 7 {
		fmt.Fprintln(os.Stderr, "❌ Internal updater: wrong number of arguments")
		os.Exit(1)
	}

	execPath := args[0]
	binaryURL := args[1]
	checksumURL := args[2]
	assetName := args[3]
	expectVer := args[4]
	expectCommit := args[5]
	updateLabel := args[6]

	fmt.Println("⏳ Waiting for parent process to exit...")
	time.Sleep(500 * time.Millisecond)

	if err := runChildUpdate(execPath, binaryURL, checksumURL, assetName, expectVer, expectCommit, updateLabel); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Update failed: %v\n", err)
		os.Exit(1)
	}
}

func resolveUpdateCandidate(currentVersion, currentCommit string, currentSemver semver, plat platform) (*updateCandidate, error) {
	rel, err := fetchLatestRelease()
	if err != nil && !errors.Is(err, errReleaseNotFound) {
		return nil, fmt.Errorf("failed to fetch release info: %w", err)
	}

	if rel != nil {
		fmt.Printf("🏷️  Latest release  : %s\n", rel.TagName)

		latestSemver, err := parseSemver(rel.TagName)
		if err != nil {
			return nil, fmt.Errorf("invalid remote version %q: %w", rel.TagName, err)
		}

		if isNewer(currentSemver, latestSemver) {
			fmt.Printf("🆕 New release      : %s → %s\n", currentVersion, rel.TagName)

			binaryURL, checksumURL, err := findAssetURLs(
				rel.Assets,
				plat.binaryAssetName(),
				plat.checksumAssetName(),
			)
			if err != nil {
				return nil, err
			}

			return &updateCandidate{
				binaryURL:   binaryURL,
				checksumURL: checksumURL,
				assetName:   plat.binaryAssetName(),
				expectVer:   rel.TagName,
				updateLabel: rel.TagName,
			}, nil
		}
	} else {
		fmt.Println("ℹ️  No published release found; checking branch artifacts...")
	}

	return resolveBranchCandidate(currentVersion, currentCommit, currentSemver, plat)
}

func resolveBranchCandidate(currentVersion, currentCommit string, currentSemver semver, plat platform) (*updateCandidate, error) {
	meta, err := fetchBranchMetadata(config.GitHubBranch)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch %s branch metadata: %w", config.GitHubBranch, err)
	}
	if meta == nil {
		return nil, nil
	}

	branchCommit := shortCommit(meta.Commit)
	branchSemver, err := parseSemver(meta.Version)
	if err != nil {
		return nil, fmt.Errorf("invalid branch version %q: %w", meta.Version, err)
	}
	fmt.Printf("🌿 Branch artifact  : %s (%s)\n", meta.Version, branchCommit)

	if isNewer(currentSemver, branchSemver) {
		return nil, nil
	}

	if sameCommit(currentCommit, branchCommit) && currentVersion == meta.Version {
		return nil, nil
	}

	assetName := plat.binaryAssetName()
	binaryPath := filepath.ToSlash(filepath.Join("dist", assetName))
	checksumPath := filepath.ToSlash(filepath.Join("dist", assetName+".sha256"))
	binaryURL := rawGitHubURL(config.GitHubBranch, binaryPath)
	checksumURL := rawGitHubURL(config.GitHubBranch, checksumPath)

	ok, err := remoteFileExists(binaryURL)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}

	checksumExists, err := remoteFileExists(checksumURL)
	if err != nil {
		return nil, err
	}
	if !checksumExists {
		checksumURL = ""
	}

	currentLabel := currentCommit
	if currentLabel == "" || currentLabel == "none" {
		currentLabel = currentVersion
	}
	fmt.Printf("🆕 New branch build : %s → %s\n", currentLabel, branchCommit)

	return &updateCandidate{
		binaryURL:    binaryURL,
		checksumURL:  checksumURL,
		assetName:    assetName,
		expectVer:    meta.Version,
		expectCommit: branchCommit,
		updateLabel:  fmt.Sprintf("%s (%s)", meta.Version, branchCommit),
	}, nil
}

func shortCommit(commit string) string {
	commit = strings.TrimSpace(commit)
	if len(commit) > 7 {
		return commit[:7]
	}
	return commit
}

func sameCommit(left, right string) bool {
	left = strings.TrimSpace(left)
	right = strings.TrimSpace(right)
	if left == "" || right == "" || left == "none" || right == "none" {
		return false
	}
	return strings.HasPrefix(left, right) || strings.HasPrefix(right, left)
}

func validateURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	if u.Scheme != "https" {
		return fmt.Errorf("insecure URL scheme: %s (https required)", u.Scheme)
	}

	return nil
}

func downloadFile(fileURL, destPath string) error {
	if err := validateURL(fileURL); err != nil {
		return err
	}

	client := &http.Client{Timeout: downloadTimeout}

	resp, err := client.Get(fileURL)
	if err != nil {
		return fmt.Errorf("download request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned HTTP %d", resp.StatusCode)
	}

	if resp.ContentLength > maxBinarySize {
		return fmt.Errorf("download too large (%d bytes)", resp.ContentLength)
	}

	tmp := destPath + ".tmp"
	out, err := os.Create(tmp)
	if err != nil {
		return fmt.Errorf("cannot create file: %w", err)
	}
	defer out.Close()

	written, err := io.Copy(out, resp.Body)
	if err != nil {
		os.Remove(tmp)
		return fmt.Errorf("download failed: %w", err)
	}

	if written == 0 {
		os.Remove(tmp)
		return fmt.Errorf("download produced empty file")
	}

	if err := os.Rename(tmp, destPath); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("file rename failed: %w", err)
	}

	return nil
}

func verifyChecksum(binaryFile, checksumFile string) error {
	data, err := os.ReadFile(checksumFile)
	if err != nil {
		return fmt.Errorf("cannot read checksum file: %w", err)
	}

	fields := strings.Fields(string(data))
	if len(fields) == 0 {
		return fmt.Errorf("checksum file empty")
	}

	expected := strings.ToLower(fields[0])
	if len(expected) != 64 {
		return fmt.Errorf("invalid SHA256 checksum length")
	}

	file, err := os.Open(binaryFile)
	if err != nil {
		return fmt.Errorf("cannot open binary: %w", err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return fmt.Errorf("hash computation failed: %w", err)
	}

	actual := hex.EncodeToString(hash.Sum(nil))
	if actual != expected {
		return fmt.Errorf("checksum mismatch\nexpected: %s\nactual:   %s", expected, actual)
	}

	return nil
}

func verifyBinary(binaryPath, expectedVersion, expectedCommit string) error {
	cmd := exec.Command(binaryPath, "version")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("binary execution failed: %w", err)
	}

	out := strings.TrimSpace(string(output))
	if !strings.Contains(out, expectedVersion) {
		return fmt.Errorf("version mismatch\nexpected: %s\nbinary reported: %s", expectedVersion, out)
	}
	if expectedCommit != "" && !strings.Contains(out, expectedCommit) {
		return fmt.Errorf("commit mismatch\nexpected: %s\nbinary reported: %s", expectedCommit, out)
	}

	return nil
}

func replaceExecutable(execPath, newBinary string) error {
	backup := execPath + ".old"

	for i := 0; i < 5; i++ {
		err := os.Rename(execPath, backup)
		if err != nil {
			time.Sleep(300 * time.Millisecond)
			continue
		}

		err = os.Rename(newBinary, execPath)
		if err != nil {
			_ = os.Rename(backup, execPath)
			return fmt.Errorf("install failed, rollback completed: %w", err)
		}

		os.Remove(backup)
		return nil
	}

	return fmt.Errorf("unable to replace executable after retries")
}

func runChildUpdate(execPath, binaryURL, checksumURL, assetName, expectedVersion, expectedCommit, updateLabel string) error {
	baseDir := filepath.Dir(execPath)

	tmpDir, err := os.MkdirTemp(baseDir, "mahin-update-*")
	if err != nil {
		return fmt.Errorf("cannot create temp directory: %w", err)
	}

	defer func() {
		fmt.Println("🧹 Cleaning temporary files...")
		os.RemoveAll(tmpDir)
	}()

	fmt.Printf("📁 Temp workspace: %s\n", tmpDir)

	newBinaryPath := filepath.Join(tmpDir, assetName)

	fmt.Println("⬇ Downloading binary...")
	if err := downloadFile(binaryURL, newBinaryPath); err != nil {
		return fmt.Errorf("binary download failed: %w", err)
	}

	if checksumURL != "" {
		fmt.Println("🔐 Verifying checksum...")

		checksumPath := filepath.Join(tmpDir, assetName+".sha256")
		if err := downloadFile(checksumURL, checksumPath); err != nil {
			return fmt.Errorf("checksum download failed: %w", err)
		}

		if err := verifyChecksum(newBinaryPath, checksumPath); err != nil {
			return fmt.Errorf("checksum verification failed: %w", err)
		}

		fmt.Println("✅ Checksum OK")
	}

	info, err := os.Stat(newBinaryPath)
	if err != nil {
		return fmt.Errorf("cannot stat binary: %w", err)
	}
	if info.Size() == 0 {
		return fmt.Errorf("binary file empty")
	}

	if err := os.Chmod(newBinaryPath, 0755); err != nil {
		return fmt.Errorf("chmod failed: %w", err)
	}

	fmt.Println("🔬 Verifying binary...")
	if err := verifyBinary(newBinaryPath, expectedVersion, expectedCommit); err != nil {
		return fmt.Errorf("binary verification failed: %w", err)
	}

	fmt.Println("🔄 Installing update...")
	if err := replaceExecutable(execPath, newBinaryPath); err != nil {
		return fmt.Errorf("installation failed: %w", err)
	}

	fmt.Printf("\n🎉 Updated successfully to %s\n", updateLabel)

	return nil
}

type semver struct{ major, minor, patch int }

func parseSemver(v string) (semver, error) {
	v = strings.TrimPrefix(v, "v")
	if idx := strings.IndexAny(v, "-+"); idx != -1 {
		v = v[:idx]
	}
	var s semver
	_, err := fmt.Sscanf(v, "%d.%d.%d", &s.major, &s.minor, &s.patch)
	if err != nil {
		return semver{}, err
	}
	return s, nil
}

func isNewer(current, candidate semver) bool {
	if candidate.major != current.major {
		return candidate.major > current.major
	}
	if candidate.minor != current.minor {
		return candidate.minor > current.minor
	}
	return candidate.patch > current.patch
}

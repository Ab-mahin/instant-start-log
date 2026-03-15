// github.go — queries the GitHub Releases API.
// Only job: fetch the latest release metadata (tag + asset download URLs).
package updater

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/mahin/mahin-cli-v1/config"
)

var errReleaseNotFound = errors.New("github release not found")

type branchMetadata struct {
	Branch    string `json:"branch"`
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildDate string `json:"build_date"`
}

func githubClient() *http.Client {
	return &http.Client{Timeout: 30 * time.Second}
}

func githubRequest(url string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	return req, nil
}

// release holds the fields we need from the GitHub API response.
type release struct {
	TagName string  `json:"tag_name"` // e.g. "v1.2.0"
	Assets  []asset `json:"assets"`
}

// asset is a single file attached to a GitHub Release.
type asset struct {
	Name               string `json:"name"`                 // e.g. "mahin-linux-amd64"
	BrowserDownloadURL string `json:"browser_download_url"` // direct download URL
}

// fetchLatestRelease calls the GitHub API and returns the latest release.
func fetchLatestRelease() (*release, error) {
	url := fmt.Sprintf(
		"https://api.github.com/repos/%s/%s/releases/latest",
		config.GitHubOwner,
		config.GitHubRepo,
	)

	req, err := githubRequest(url)
	if err != nil {
		return nil, err
	}

	resp, err := githubClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("network request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, errReleaseNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned HTTP %d", resp.StatusCode)
	}

	var r release
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, fmt.Errorf("failed to parse GitHub response: %w", err)
	}
	return &r, nil
}

func fetchBranchMetadata(branch string) (*branchMetadata, error) {
	url := rawGitHubURL(branch, config.BranchMetadataPath)

	req, err := githubRequest(url)
	if err != nil {
		return nil, err
	}

	resp, err := githubClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("network request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("branch metadata request returned HTTP %d", resp.StatusCode)
	}

	var meta branchMetadata
	if err := json.NewDecoder(resp.Body).Decode(&meta); err != nil {
		return nil, fmt.Errorf("failed to parse branch metadata: %w", err)
	}

	return &meta, nil
}

func rawGitHubURL(branch, filePath string) string {
	return fmt.Sprintf(
		"https://raw.githubusercontent.com/%s/%s/%s/%s",
		config.GitHubOwner,
		config.GitHubRepo,
		branch,
		path.Clean(filePath),
	)
}

func remoteFileExists(fileURL string) (bool, error) {
	req, err := http.NewRequest(http.MethodHead, fileURL, nil)
	if err != nil {
		return false, err
	}

	resp, err := githubClient().Do(req)
	if err != nil {
		return false, fmt.Errorf("asset probe failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("asset probe returned HTTP %d", resp.StatusCode)
	}

	return true, nil
}

// findAssetURLs searches the release asset list for the binary and its
// optional checksum file, both matching the current platform.
// Returns an error if the binary asset is not found.
func findAssetURLs(assets []asset, binaryName, checksumName string) (binaryURL, checksumURL string, err error) {
	for _, a := range assets {
		switch a.Name {
		case binaryName:
			binaryURL = a.BrowserDownloadURL
		case checksumName:
			checksumURL = a.BrowserDownloadURL
		}
	}

	if binaryURL == "" {
		names := make([]string, len(assets))
		for i, a := range assets {
			names[i] = a.Name
		}
		return "", "", fmt.Errorf(
			"no asset %q in release — available: %s",
			binaryName,
			strings.Join(names, ", "),
		)
	}

	return binaryURL, checksumURL, nil
}

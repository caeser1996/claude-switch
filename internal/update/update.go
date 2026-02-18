package update

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	repoOwner = "caeser1996"
	repoName  = "claude-switch"
	apiURL    = "https://api.github.com/repos/" + repoOwner + "/" + repoName + "/releases/latest"
)

// Release represents a GitHub release.
type Release struct {
	TagName string  `json:"tag_name"`
	Assets  []Asset `json:"assets"`
	HTMLURL string  `json:"html_url"`
}

// Asset represents a release asset.
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

// CheckResult holds the result of an update check.
type CheckResult struct {
	CurrentVersion string
	LatestVersion  string
	UpdateNeeded   bool
	DownloadURL    string
	AssetName      string
	ReleaseURL     string
}

// Check queries GitHub for the latest release and compares with current version.
func Check(currentVersion string) (*CheckResult, error) {
	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("cannot reach GitHub: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("no releases found for %s/%s", repoOwner, repoName)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("cannot parse release info: %w", err)
	}

	latest := strings.TrimPrefix(release.TagName, "v")
	current := strings.TrimPrefix(currentVersion, "v")

	result := &CheckResult{
		CurrentVersion: current,
		LatestVersion:  latest,
		UpdateNeeded:   current != latest && current != "dev",
		ReleaseURL:     release.HTMLURL,
	}

	// Find the right asset for this platform
	assetName := expectedAssetName()
	for _, a := range release.Assets {
		if a.Name == assetName {
			result.DownloadURL = a.BrowserDownloadURL
			result.AssetName = a.Name
			break
		}
	}

	return result, nil
}

// Apply downloads and replaces the current binary.
func Apply(downloadURL string) error {
	// Download to temp file
	resp, err := http.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("download returned status %d", resp.StatusCode)
	}

	// Get current binary path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot determine executable path: %w", err)
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("cannot resolve symlinks: %w", err)
	}

	// Write to temp file next to the binary
	tmpFile, err := os.CreateTemp(filepath.Dir(execPath), "claude-switch-update-*")
	if err != nil {
		return fmt.Errorf("cannot create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("download write failed: %w", err)
	}
	tmpFile.Close()

	// Make executable
	if err := os.Chmod(tmpPath, 0755); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("cannot set permissions: %w", err)
	}

	// Replace old binary
	if err := os.Rename(tmpPath, execPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("cannot replace binary: %w", err)
	}

	return nil
}

// expectedAssetName returns the expected binary name for this platform.
func expectedAssetName() string {
	os := runtime.GOOS
	arch := runtime.GOARCH
	name := fmt.Sprintf("claude-switch_%s_%s", os, arch)
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	return name
}

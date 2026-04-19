package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

const (
	CurrentVersion = "v2.0.2"
	GitHubRepo     = "LukasYTTT/appinstaller"
)

type githubRelease struct {
	TagName string        `json:"tag_name"`
	Assets  []githubAsset `json:"assets"`
}

type githubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// CheckUpdate silently checks for a new version from GitHub releases and updates the binary if a newer version is found.
// It will fail gracefully if it encounters permission issues or network errors.
func CheckUpdate() {
	execPath, err := os.Executable()
	if err != nil {
		return
	}

	// Attempt to open the executable for appending (just to check if we have write permissions)
	// If we are installed via AUR or Flatpak, the binary is owned by root or mounted read-only,
	// so this will fail, which is exactly what we want.
	file, err := os.OpenFile(execPath, os.O_WRONLY|os.O_APPEND, 0)
	if err != nil {
		// Not writable, gracefully abort update check (user should use their package manager)
		return
	}
	file.Close()

	// Check latest release
	resp, err := http.Get(fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", GitHubRepo))
	if err != nil || resp.StatusCode != 200 {
		return
	}
	defer resp.Body.Close()

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return
	}

	// Simple version comparison (assumes format vX.Y.Z)
	if release.TagName <= CurrentVersion {
		return
	}

	// Look for the "appinstall" asset
	var downloadURL string
	for _, asset := range release.Assets {
		if asset.Name == "appinstall" {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	if downloadURL == "" {
		return // No matching asset found in release
	}

	// Download to a temporary file
	tempFile := filepath.Join(os.TempDir(), "appinstall-update-tmp")
	out, err := os.Create(tempFile)
	if err != nil {
		return
	}

	downloadResp, err := http.Get(downloadURL)
	if err != nil {
		out.Close()
		os.Remove(tempFile)
		return
	}
	defer downloadResp.Body.Close()

	_, err = io.Copy(out, downloadResp.Body)
	out.Close()
	if err != nil {
		os.Remove(tempFile)
		return
	}

	// Make executable
	os.Chmod(tempFile, 0755)

	// Keep a backup of the running executable
	backupPath := execPath + ".old"
	os.Rename(execPath, backupPath)

	// Move new file into place
	err = os.Rename(tempFile, execPath)
	if err != nil {
		// If atomic rename fails, try to restore
		os.Rename(backupPath, execPath)
		return
	}

	// Success! Clean up the old binary.
	os.Remove(backupPath)
}

// GetUpdateVersion checks GitHub for a newer version without downloading it.
// Returns the tag name if an update is available, otherwise returns an empty string.
func GetUpdateVersion() string {
	resp, err := http.Get(fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", GitHubRepo))
	if err != nil || resp.StatusCode != 200 {
		return ""
	}
	defer resp.Body.Close()

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return ""
	}

	if release.TagName > CurrentVersion {
		return release.TagName
	}
	return ""
}

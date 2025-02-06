package sync

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// DownloadZIP downloads a repository ZIP file from GitHub
func DownloadZIP(url, ref, refType string) (*os.File, error) {
	// Construct the ZIP URL based on ref type
	var zipURL string
	if refType == "tag" {
		zipURL = fmt.Sprintf("%s/archive/refs/tags/%s.zip", url, ref)
	} else {
		// Default to branch
		zipURL = fmt.Sprintf("%s/archive/refs/heads/%s.zip", url, ref)
	}

	// Create temporary file
	tmpFile, err := os.CreateTemp("", "repo-*.zip")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}

	// Download the ZIP file
	resp, err := http.Get(zipURL)
	if err != nil {
		os.Remove(tmpFile.Name())
		return nil, fmt.Errorf("failed to download ZIP: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		os.Remove(tmpFile.Name())
		return nil, fmt.Errorf("failed to download ZIP: HTTP %d", resp.StatusCode)
	}

	// Copy the body to the temp file
	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		os.Remove(tmpFile.Name())
		return nil, fmt.Errorf("failed to save ZIP: %w", err)
	}

	return tmpFile, nil
}

// ExtractZIP extracts the downloaded ZIP file to the destination
func ExtractZIP(zipFile *os.File, destination string) error {
	// Open the ZIP file
	reader, err := zip.OpenReader(zipFile.Name())
	if err != nil {
		return fmt.Errorf("failed to open ZIP: %w", err)
	}
	defer reader.Close()

	// Create destination directory if it doesn't exist
	if err := os.MkdirAll(destination, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Find the root directory name in the ZIP (usually repository-branch)
	var rootDir string
	if len(reader.File) > 0 {
		rootDir = strings.Split(reader.File[0].Name, "/")[0]
	}

	// Extract files
	for _, file := range reader.File {
		// Remove the root directory from the path
		relPath := strings.TrimPrefix(file.Name, rootDir+"/")
		if relPath == "" {
			continue
		}

		targetPath := filepath.Join(destination, relPath)

		if file.FileInfo().IsDir() {
			os.MkdirAll(targetPath, file.Mode())
			continue
		}

		// Create parent directories if they don't exist
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}

		// Create the file
		outFile, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}

		// Open the file in the ZIP
		rc, err := file.Open()
		if err != nil {
			outFile.Close()
			return fmt.Errorf("failed to open file in ZIP: %w", err)
		}

		// Copy the contents
		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
		if err != nil {
			return fmt.Errorf("failed to extract file: %w", err)
		}
	}

	return nil
}

// Cleanup removes temporary files
func Cleanup(files ...string) {
	for _, file := range files {
		os.Remove(file)
	}
} 
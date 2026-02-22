package remote

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rishiyaduwanshi/boiler/internal/store"
	"github.com/rishiyaduwanshi/boiler/internal/utils"
)

// FetchSnippet downloads a snippet from remote registry to local store
// remotePath formats:
//   - "owner/repo:path/to/file.js" (GitHub)
//   - "https://domain.com/path/to/file.js" (Direct URL)
//   - "domain.com:path/to/file.js" (Custom domain, assumes HTTPS)
func FetchSnippet(remotePath string, destPath string) error {
	owner, repo, path := store.ParseRemotePath(remotePath)
	
	var fileURL string
	var displaySource string
	
	// Check if it's a direct URL
	if strings.HasPrefix(repo, "http://") || strings.HasPrefix(repo, "https://") {
		fileURL = repo
		displaySource = repo
	} else if owner == "" && repo != "" {
		// Custom domain format (domain.com:path)
		fileURL = fmt.Sprintf("https://%s/%s", repo, path)
		displaySource = fmt.Sprintf("%s/%s", repo, path)
	} else if owner != "" && repo != "" {
		// GitHub format (owner/repo:path)
		fileURL = fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/main/%s", owner, repo, path)
		displaySource = fmt.Sprintf("%s/%s", owner, repo)
	} else {
		return fmt.Errorf("invalid remote path format: %s", remotePath)
	}
	
	fmt.Printf("📥 Downloading from %s...\n", displaySource)
	
	data, err := downloadFile(fileURL)
	if err != nil {
		return fmt.Errorf("failed to download snippet: %w", err)
	}

	// Ensure destination directory exists
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Write to destination
	if err := os.WriteFile(destPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write snippet: %w", err)
	}

	fmt.Printf("✓ Downloaded successfully\n")
	return nil
}

// FetchStack downloads a complete stack from remote registry to local store
// remotePath formats:
//   - "owner/repo" (GitHub short format)
//   - "owner/repo:path/within/repo" (GitHub with subpath)
//   - "https://github.com/owner/repo" (GitHub full URL)
func FetchStack(remotePath string, destPath string) error {
	owner, repo, subPath := store.ParseRemotePath(remotePath)
	
	// Handle full GitHub URLs
	if strings.HasPrefix(remotePath, "https://github.com/") || strings.HasPrefix(remotePath, "http://github.com/") {
		// Extract owner/repo from URL
		// https://github.com/owner/repo -> owner/repo
		cleanPath := strings.TrimPrefix(remotePath, "https://github.com/")
		cleanPath = strings.TrimPrefix(cleanPath, "http://github.com/")
		cleanPath = strings.TrimSuffix(cleanPath, "/")
		
		parts := strings.SplitN(cleanPath, "/", 2)
		if len(parts) == 2 {
			owner = parts[0]
			repo = parts[1]
			// Remove any trailing .git
			repo = strings.TrimSuffix(repo, ".git")
		}
	}
	
	if owner == "" || repo == "" {
		return fmt.Errorf("invalid remote path format: %s (expected: owner/repo or https://github.com/owner/repo)", remotePath)
	}

	fmt.Printf("📥 Downloading stack from %s/%s...\n", owner, repo)

	// Download tarball from GitHub
	tarballURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/tarball/main", owner, repo)
	
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "boiler-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Download tarball
	tarPath := filepath.Join(tempDir, "stack.tar.gz")
	if err := downloadToFile(tarballURL, tarPath); err != nil {
		return fmt.Errorf("failed to download stack: %w", err)
	}

	// Extract tarball
	extractDir := filepath.Join(tempDir, "extracted")
	if err := extractTarGz(tarPath, extractDir); err != nil {
		return fmt.Errorf("failed to extract stack: %w", err)
	}

	// Find the actual content directory (GitHub adds a prefix folder)
	entries, err := os.ReadDir(extractDir)
	if err != nil {
		return fmt.Errorf("failed to read extracted directory: %w", err)
	}
	if len(entries) == 0 {
		return fmt.Errorf("extracted directory is empty")
	}

	// Get source directory
	var sourceDir string
	if subPath == "." {
		sourceDir = filepath.Join(extractDir, entries[0].Name())
	} else {
		sourceDir = filepath.Join(extractDir, entries[0].Name(), subPath)
	}
	
	// Copy to destination
	if err := utils.CopyDir(sourceDir, destPath, nil); err != nil {
		return fmt.Errorf("failed to copy stack: %w", err)
	}

	fmt.Printf("✓ Downloaded stack successfully\n")
	return nil
}

// Helper functions

// downloadFile downloads a file from URL and returns its content
func downloadFile(url string) ([]byte, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return data, nil
}

// downloadToFile downloads a file from URL and saves to path
func downloadToFile(url, path string) error {
	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	out, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// extractTarGz extracts a tar.gz file to destination
func extractTarGz(tarPath, destPath string) error {
	file, err := os.Open(tarPath)
	if err != nil {
		return err
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(destPath, header.Name)

		// Zip Slip protection: ensure extracted path stays within destPath
		cleanDest := filepath.Clean(destPath) + string(os.PathSeparator)
		if !strings.HasPrefix(filepath.Clean(target)+string(os.PathSeparator), cleanDest) {
			return fmt.Errorf("invalid file path in archive (path traversal attempt): %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			outFile, err := os.Create(target)
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		}
	}

	return nil
}


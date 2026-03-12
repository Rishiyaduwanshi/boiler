package remote

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
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

// FetchSnippet downloads a snippet file to destPath.
//
// remotePath formats:
//   - "owner/repo:path/to/file.js"          → GitHub (default)
//   - "https://any.host/path/to/file.js"    → direct URL (used by generic/registry servers)
//   - "domain.com:path/to/file.js"          → custom domain, resolves to https://domain.com/path
func FetchSnippet(remotePath string, destPath string) error {
	owner, repo, filePath := store.ParseRemotePath(remotePath)

	var fileURL string
	switch {
	case strings.HasPrefix(remotePath, "http://") || strings.HasPrefix(remotePath, "https://"):
		// Direct URL - generic/registry server passes full URL in meta.json
		fileURL = remotePath
	case owner != "" && repo != "":
		// owner/repo:path - use provider (defaults to GitHub for short format)
		p := Detect(remotePath)
		fileURL = p.RawFileURL(owner, repo, "main", filePath)
	case repo != "" && filePath != "":
		// domain.com:path - build HTTPS URL
		fileURL = fmt.Sprintf("https://%s/%s", repo, filePath)
	default:
		return fmt.Errorf("invalid remote path format: %s", remotePath)
	}

	fmt.Printf("📥 Downloading snippet...\n")

	data, err := downloadFile(fileURL)
	if err != nil {
		return fmt.Errorf("failed to download snippet: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	if err := os.WriteFile(destPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write snippet: %w", err)
	}

	fmt.Printf("✓ Downloaded successfully\n")
	return nil
}

// FetchStack downloads a complete stack from remote registry to local store.
//
// The right Provider is detected automatically from the remotePath:
//   - "owner/repo"                                   → GitHub (default short format)
//   - "owner/repo:path/within/repo"                  → GitHub with subpath
//   - "https://github.com/owner/repo"                → GitHub
//   - "https://gitlab.com/owner/repo"                → GitLab
//   - "https://bitbucket.org/owner/repo"             → Bitbucket
//   - "https://mysite.com/store/stacks/<name>.zip"   → Generic (direct archive URL)
func FetchStack(remotePath string, destPath string) error {
	// Detect provider: full URLs carry host info; short "owner/repo" defaults to GitHub.
	var p Provider
	if strings.HasPrefix(remotePath, "http://") || strings.HasPrefix(remotePath, "https://") {
		p = Detect(remotePath)
	} else {
		p = githubProvider{} // backward-compatible default
	}

	owner, repo, subPath := store.ParseRemotePath(remotePath)

	// For full URLs, extract owner/repo from the URL itself
	if strings.HasPrefix(remotePath, "http://") || strings.HasPrefix(remotePath, "https://") {
		owner, repo = parseOwnerRepo(remotePath)
	}
	repo = strings.TrimSuffix(repo, ".git")

	// archiveURL is either constructed by the provider (GitHub/GitLab/Bitbucket)
	// or the remotePath itself when it is a direct archive URL (any host, one-off fetch).
	var archiveURL string
	var archiveExt string
	if owner == "" || repo == "" {
		// Direct archive URL - user passed something like:
		//   https://anysite.com/mystack.zip
		//   https://anysite.com/mystack.tar.gz
		if !strings.HasPrefix(remotePath, "http://") && !strings.HasPrefix(remotePath, "https://") {
			return fmt.Errorf("invalid remote path: %s (expected owner/repo or a full archive URL)", remotePath)
		}
		archiveURL = remotePath
		// Detect format from URL extension so both .zip and .tar.gz work.
		archiveExt = archiveFormatFromURL(archiveURL)
		fmt.Printf("📥 Downloading stack...\n")
	} else {
		archiveURL = p.ArchiveURL(owner, repo, "main", subPath)
		archiveExt = p.ArchiveFormat()
		fmt.Printf("📥 Downloading stack from %s (%s/%s)...\n", p.Name(), owner, repo)
	}

	// Create temp directory
	tempDir, err := os.MkdirTemp("", "boiler-stack-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	archivePath := filepath.Join(tempDir, "stack."+archiveExt)
	if err := downloadToFile(archiveURL, archivePath); err != nil {
		return fmt.Errorf("failed to download stack archive: %w", err)
	}

	extractDir := filepath.Join(tempDir, "extracted")
	switch archiveExt {
	case "tar.gz":
		if err := extractTarGz(archivePath, extractDir); err != nil {
			return fmt.Errorf("failed to extract stack: %w", err)
		}
	case "zip":
		if err := extractZip(archivePath, extractDir); err != nil {
			return fmt.Errorf("failed to extract stack: %w", err)
		}
	default:
		return fmt.Errorf("unsupported archive format: %s", archiveExt)
	}

	// GitHub/GitLab wrap contents in a randomly-named prefix folder - skip it if present.
	entries, err := os.ReadDir(extractDir)
	if err != nil {
		return fmt.Errorf("failed to read extracted directory: %w", err)
	}
	if len(entries) == 0 {
		return fmt.Errorf("extracted archive is empty")
	}

	var sourceDir string
	firstIsDir := entries[0].IsDir()
	if firstIsDir && len(entries) == 1 {
		// Single top-level dir (GitHub/GitLab prefix) - unwrap it
		if subPath == "" || subPath == "." {
			sourceDir = filepath.Join(extractDir, entries[0].Name())
		} else {
			sourceDir = filepath.Join(extractDir, entries[0].Name(), subPath)
		}
	} else {
		// Generic/Bitbucket: no prefix dir - content is directly inside extractDir
		if subPath == "" || subPath == "." {
			sourceDir = extractDir
		} else {
			sourceDir = filepath.Join(extractDir, subPath)
		}
	}

	if err := utils.CopyDir(sourceDir, destPath, nil); err != nil {
		return fmt.Errorf("failed to copy stack: %w", err)
	}

	fmt.Printf("✓ Stack downloaded successfully\n")
	return nil
}

// Helper functions

// downloadFile downloads a file from URL and returns its content
func downloadFile(url string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
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
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
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

// archiveFormatFromURL returns "tar.gz" or "zip" based on the URL path.
// Defaults to "zip" if extension is unrecognised.
func archiveFormatFromURL(url string) string {
	lower := strings.ToLower(url)
	if strings.Contains(lower, ".tar.gz") || strings.Contains(lower, ".tgz") {
		return "tar.gz"
	}
	return "zip"
}

// extractZip extracts a zip archive to destPath.
func extractZip(zipPath, destPath string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	cleanDest := filepath.Clean(destPath) + string(os.PathSeparator)

	for _, f := range r.File {
		target := filepath.Join(destPath, filepath.FromSlash(f.Name))

		// Zip Slip protection
		if !strings.HasPrefix(filepath.Clean(target)+string(os.PathSeparator), cleanDest) {
			return fmt.Errorf("invalid path in zip (path traversal attempt): %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}
		out, err := os.Create(target)
		if err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			out.Close()
			return err
		}
		_, copyErr := io.Copy(out, rc)
		rc.Close()
		out.Close()
		if copyErr != nil {
			return copyErr
		}
	}
	return nil
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


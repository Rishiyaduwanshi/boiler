package remote

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rishiyaduwanshi/boiler/internal/constants"
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
	case utils.IsURL(remotePath):
		// Direct URL - normalize known hosted file pages to raw URLs.
		fileURL = normalizeDirectFileURL(remotePath)
	case owner != "" && repo != "":
		if filePath == "" || filePath == "." {
			return fmt.Errorf("invalid snippet path: %s", remotePath)
		}

		// owner/repo:path and provider-prefixed formats (github:, gitlab:, bitbucket:)
		p := providerForRemotePath(remotePath)
		repo = strings.TrimSuffix(repo, ".git")
		ref := resolveProviderRef(p, owner, repo, defaultRemoteRef)
		fileURL = p.RawFileURL(owner, repo, ref, filePath)
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

	if err := utils.EnsureDir(filepath.Dir(destPath)); err != nil {
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
	if utils.IsURL(remotePath) {
		p = Detect(remotePath)
	} else {
		p = providerForRemotePath(remotePath)
	}

	owner, repo, subPath := store.ParseRemotePath(remotePath)
	ref := defaultRemoteRef
	hasExplicitRef := false

	// For full URLs, extract owner/repo from the URL itself
	if utils.IsURL(remotePath) {
		parsedOwner, parsedRepo := parseOwnerRepo(remotePath)
		if parsedOwner != "" && parsedRepo != "" {
			owner, repo = parsedOwner, parsedRepo
		}

		parsedRef, parsedSubPath, ok := parseHostedStackRefAndSubPath(remotePath)
		if ok {
			ref = parsedRef
			subPath = parsedSubPath
			hasExplicitRef = true
		}
	}

	if subPath == "" {
		subPath = "."
	}
	repo = strings.TrimSuffix(repo, ".git")
	if !hasExplicitRef {
		ref = resolveProviderRef(p, owner, repo, ref)
	}

	if owner != "" && repo != "" && subPath != "." {
		if err := fetchProviderSubPath(p, owner, repo, ref, subPath, destPath); err == nil {
			hasPointers, scanErr := hasGitLFSPointers(destPath)
			if scanErr != nil {
				return fmt.Errorf("failed to scan fetched stack for Git LFS pointers: %w", scanErr)
			}
			if hasPointers {
				fmt.Println(utils.MsgGitLFSFallback)
				tempDir, tempErr := os.MkdirTemp("", "boiler-stack-clone-*")
				if tempErr != nil {
					return fmt.Errorf("failed to create temp directory: %w", tempErr)
				}
				defer os.RemoveAll(tempDir)

				if cloneErr := fetchStackWithGitClone(p, owner, repo, ref, subPath, tempDir, destPath); cloneErr != nil {
					return fmt.Errorf("failed to fetch Git LFS stack: %w", cloneErr)
				}
			}
			fmt.Printf("✓ Stack downloaded successfully\n")
			return nil
		} else {
			debugf("provider subtree fetch failed, falling back to archive: %v", err)
		}
	}

	// archiveURL is either constructed by the provider (GitHub/GitLab/Bitbucket)
	// or the remotePath itself when it is a direct archive URL (any host, one-off fetch).
	var archiveURL string
	var archiveExt string
	if owner == "" || repo == "" {
		// Direct archive URL - user passed something like:
		//   https://anysite.com/mystack.zip
		//   https://anysite.com/mystack.tar.gz
		if !utils.IsURL(remotePath) {
			return fmt.Errorf("invalid remote path: %s (expected owner/repo or a full archive URL)", remotePath)
		}
		archiveURL = remotePath
		// Detect format from URL extension so both .zip and .tar.gz work.
		archiveExt = archiveFormatFromURL(archiveURL)
		fmt.Printf("📥 Downloading stack...\n")
	} else {
		if subPath != "." {
			debugf("falling back to full archive for subpath fetch path=%s", subPath)
		}
		archiveURL = p.ArchiveURL(owner, repo, ref, subPath)
		archiveExt = p.ArchiveFormat()
		fmt.Printf("📥 Downloading stack from %s (%s/%s@%s)...\n", p.Name(), owner, repo, ref)
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

	hasPointers, err := hasGitLFSPointers(sourceDir)
	if err != nil {
		return fmt.Errorf("failed to scan extracted stack for Git LFS pointers: %w", err)
	}
	if hasPointers {
		fmt.Println(utils.MsgGitLFSFallback)
		if err := fetchStackWithGitClone(p, owner, repo, ref, subPath, tempDir, destPath); err != nil {
			return fmt.Errorf("failed to fetch Git LFS stack: %w", err)
		}
		fmt.Printf("✓ Stack downloaded successfully\n")
		return nil
	}

	if err := utils.CopyDir(sourceDir, destPath, nil); err != nil {
		return fmt.Errorf("failed to copy stack: %w", err)
	}

	fmt.Printf("✓ Stack downloaded successfully\n")
	return nil
}

func fetchProviderSubPath(p Provider, owner, repo, ref, subPath, destPath string) error {
	switch provider := p.(type) {
	case githubProvider:
		fmt.Printf("📥 Fetching GitHub subtree %s/%s@%s:%s...\n", owner, repo, ref, subPath)
		return fetchGitHubSubPath(owner, repo, ref, subPath, destPath)
	case gitlabProvider:
		fmt.Printf("📥 Fetching GitLab subtree %s/%s@%s:%s...\n", owner, repo, ref, subPath)
		return fetchGitLabSubPath(provider.host, owner, repo, ref, subPath, destPath)
	case bitbucketProvider:
		fmt.Printf("📥 Fetching Bitbucket subtree %s/%s@%s:%s...\n", owner, repo, ref, subPath)
		return fetchBitbucketSubPath(owner, repo, ref, subPath, destPath)
	default:
		return fmt.Errorf("provider %s does not support subtree fetch", p.Name())
	}
}

func providerForRemotePath(remotePath string) Provider {
	lower := strings.ToLower(strings.TrimSpace(remotePath))
	if utils.IsURL(lower) {
		return Detect(remotePath)
	}

	prefix := lower
	if idx := strings.Index(prefix, ":"); idx >= 0 {
		prefix = prefix[:idx]
	}

	switch prefix {
	case constants.ProviderGithubAlias, constants.ProviderGithubHost, constants.ProviderGithubWWW:
		return githubProvider{}
	case constants.ProviderGitlabAlias, constants.ProviderGitlabHost, constants.ProviderGitlabWWW:
		return gitlabProvider{host: constants.ProviderGitlabHost}
	case constants.ProviderBitbucketAlias, constants.ProviderBitbucketHost, constants.ProviderBitbucketWWW:
		return bitbucketProvider{}
	default:
		// owner/repo and owner/repo:path default to GitHub.
		return githubProvider{}
	}
}

func normalizeDirectFileURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	host := strings.ToLower(parsed.Host)
	segments := strings.Split(strings.Trim(parsed.Path, "/"), "/")

	switch host {
	case constants.ProviderGithubHost, constants.ProviderGithubWWW:
		if len(segments) >= 5 && segments[2] == "blob" {
			owner := segments[0]
			repo := segments[1]
			ref := segments[3]
			filePath := strings.Join(segments[4:], "/")
			if owner != "" && repo != "" && ref != "" && filePath != "" {
				return fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s", owner, repo, ref, filePath)
			}
		}

		// Accept malformed but common input: https://github.com/owner/repo:path/to/file
		if len(segments) >= 2 {
			repoWithPath := strings.SplitN(segments[1], ":", 2)
			if len(repoWithPath) == 2 {
				owner := segments[0]
				repo := repoWithPath[0]
				fileParts := []string{repoWithPath[1]}
				if len(segments) > 2 {
					fileParts = append(fileParts, segments[2:]...)
				}
				filePath := strings.Join(fileParts, "/")
				if owner != "" && repo != "" && filePath != "" {
					return fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/main/%s", owner, repo, filePath)
				}
			}
		}

	case constants.ProviderGitlabHost, constants.ProviderGitlabWWW:
		if len(segments) >= 6 && segments[2] == "-" && segments[3] == "blob" {
			owner := segments[0]
			repo := segments[1]
			ref := segments[4]
			filePath := strings.Join(segments[5:], "/")
			if owner != "" && repo != "" && ref != "" && filePath != "" {
				return fmt.Sprintf("https://gitlab.com/%s/%s/-/raw/%s/%s", owner, repo, ref, filePath)
			}
		}

	case constants.ProviderBitbucketHost, constants.ProviderBitbucketWWW:
		if len(segments) >= 5 && segments[2] == "src" {
			owner := segments[0]
			repo := segments[1]
			ref := segments[3]
			filePath := strings.Join(segments[4:], "/")
			if owner != "" && repo != "" && ref != "" && filePath != "" {
				return fmt.Sprintf("https://bitbucket.org/%s/%s/raw/%s/%s", owner, repo, ref, filePath)
			}
		}
	}

	return rawURL
}

func parseHostedStackRefAndSubPath(rawURL string) (ref, subPath string, ok bool) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", "", false
	}

	host := strings.ToLower(parsed.Host)
	segments := strings.Split(strings.Trim(parsed.Path, "/"), "/")

	switch host {
	case constants.ProviderGithubHost, constants.ProviderGithubWWW:
		if len(segments) >= 4 && (segments[2] == "tree" || segments[2] == "blob") {
			ref = segments[3]
			if len(segments) > 4 {
				subPath = strings.Join(segments[4:], "/")
			} else {
				subPath = "."
			}
			return ref, subPath, true
		}

	case constants.ProviderGitlabHost, constants.ProviderGitlabWWW:
		if len(segments) >= 5 && segments[2] == "-" && (segments[3] == "tree" || segments[3] == "blob") {
			ref = segments[4]
			if len(segments) > 5 {
				subPath = strings.Join(segments[5:], "/")
			} else {
				subPath = "."
			}
			return ref, subPath, true
		}

	case constants.ProviderBitbucketHost, constants.ProviderBitbucketWWW:
		if len(segments) >= 4 && segments[2] == "src" {
			ref = segments[3]
			if len(segments) > 4 {
				subPath = strings.Join(segments[4:], "/")
			} else {
				subPath = "."
			}
			return ref, subPath, true
		}
	}

	return "", "", false
}

// Helper functions

func withHTTPGet(url string, timeout time.Duration, consume func(io.Reader) error) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
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

	if err := consume(resp.Body); err != nil {
		return err
	}

	return nil
}

// downloadFile downloads a file from URL and returns its content
func downloadFile(url string) ([]byte, error) {
	var data []byte
	err := withHTTPGet(url, 30*time.Second, func(r io.Reader) error {
		body, err := io.ReadAll(r)
		if err != nil {
			return fmt.Errorf("failed to read response: %w", err)
		}
		data = body
		return nil
	})
	if err != nil {
		return nil, err
	}

	return data, nil
}

// downloadToFile downloads a file from URL and saves to path
func downloadToFile(url, path string) error {
	out, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	return withHTTPGet(url, 60*time.Second, func(r io.Reader) error {
		if _, err := io.Copy(out, r); err != nil {
			return fmt.Errorf("failed to write response to file: %w", err)
		}
		return nil
	})
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

package remote

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

type gitHubContentEntry struct {
	Type        string `json:"type"`
	Path        string `json:"path"`
	DownloadURL string `json:"download_url"`
}

func fetchGitHubSubPath(owner, repo, ref, subPath, destPath string) error {
	cleanSubPath := normalizeRemoteSubPath(subPath)
	if cleanSubPath == "." {
		return fmt.Errorf("subpath fetch requires a nested path")
	}

	debugf("github subtree fetch started owner=%s repo=%s ref=%s path=%s", owner, repo, ref, cleanSubPath)

	if err := prepareSubtreeDestination(destPath); err != nil {
		return err
	}

	if err := downloadGitHubDirectory(owner, repo, ref, cleanSubPath, cleanSubPath, destPath); err != nil {
		return err
	}

	return nil
}

func downloadGitHubDirectory(owner, repo, ref, currentPath, basePath, destRoot string) error {
	entries, err := fetchGitHubEntries(owner, repo, ref, currentPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		switch entry.Type {
		case "file":
			if err := downloadGitHubFile(entry, basePath, destRoot); err != nil {
				return err
			}
		case "dir":
			if err := downloadGitHubDirectory(owner, repo, ref, entry.Path, basePath, destRoot); err != nil {
				return err
			}
		}
	}

	return nil
}

func fetchGitHubEntries(owner, repo, ref, subPath string) ([]gitHubContentEntry, error) {
	encodedPath := encodePathSegments(subPath)
	endpoint := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s?ref=%s", owner, repo, encodedPath, url.QueryEscape(ref))
	debugf("github contents endpoint=%s", endpoint)

	data, err := downloadFile(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to list GitHub contents for %s: %w", subPath, err)
	}

	var entries []gitHubContentEntry
	if err := json.Unmarshal(data, &entries); err == nil {
		return entries, nil
	}

	var single gitHubContentEntry
	if err := json.Unmarshal(data, &single); err != nil {
		return nil, fmt.Errorf("failed to parse GitHub contents response for %s: %w", subPath, err)
	}

	if single.Path == "" {
		return nil, fmt.Errorf("invalid GitHub contents response for %s", subPath)
	}

	return []gitHubContentEntry{single}, nil
}

func downloadGitHubFile(entry gitHubContentEntry, basePath, destRoot string) error {
	if entry.DownloadURL == "" {
		return fmt.Errorf("missing download URL for file %s", entry.Path)
	}

	data, err := downloadFile(entry.DownloadURL)
	if err != nil {
		return fmt.Errorf("failed to download file %s: %w", entry.Path, err)
	}

	if err := writeSubtreeFile(basePath, entry.Path, destRoot, data); err != nil {
		return err
	}

	debugf("downloaded github file path=%s", entry.Path)
	return nil
}

func encodePathSegments(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	for i, p := range parts {
		parts[i] = url.PathEscape(p)
	}
	return strings.Join(parts, "/")
}

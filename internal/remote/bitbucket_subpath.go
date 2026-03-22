package remote

import (
	"encoding/json"
	"fmt"
	"strings"
)

type bitbucketSourcePage struct {
	Values []bitbucketSourceEntry `json:"values"`
	Next   string                 `json:"next"`
}

type bitbucketSourceEntry struct {
	Path  string `json:"path"`
	Type  string `json:"type"`
	Links struct {
		Self struct {
			Href string `json:"href"`
		} `json:"self"`
	} `json:"links"`
}

func fetchBitbucketSubPath(owner, repo, ref, subPath, destPath string) error {
	cleanSubPath := normalizeRemoteSubPath(subPath)
	if cleanSubPath == "." {
		return fmt.Errorf("subpath fetch requires a nested path")
	}

	debugf("bitbucket subtree fetch started owner=%s repo=%s ref=%s path=%s", owner, repo, ref, cleanSubPath)

	if err := prepareSubtreeDestination(destPath); err != nil {
		return err
	}

	entries, err := listBitbucketTree(owner, repo, ref, cleanSubPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !isBitbucketFileEntry(entry) {
			continue
		}

		rawURL := bitbucketRawURL(owner, repo, ref, entry)
		if rawURL == "" {
			return fmt.Errorf("missing raw URL for Bitbucket file %s", entry.Path)
		}

		data, err := downloadFile(rawURL)
		if err != nil {
			return fmt.Errorf("failed to download Bitbucket file %s: %w", entry.Path, err)
		}

		if err := writeSubtreeFile(cleanSubPath, entry.Path, destPath, data); err != nil {
			return err
		}

		debugf("downloaded bitbucket file path=%s", entry.Path)
	}

	return nil
}

func listBitbucketTree(owner, repo, ref, subPath string) ([]bitbucketSourceEntry, error) {
	endpoint := fmt.Sprintf("https://api.bitbucket.org/2.0/repositories/%s/%s/src/%s/%s", owner, repo, ref, subPath)
	all := make([]bitbucketSourceEntry, 0)

	for endpoint != "" {
		debugf("bitbucket tree endpoint=%s", endpoint)
		data, err := downloadFile(endpoint)
		if err != nil {
			return nil, fmt.Errorf("failed to list Bitbucket tree for %s: %w", subPath, err)
		}

		var page bitbucketSourcePage
		if err := json.Unmarshal(data, &page); err != nil {
			return nil, fmt.Errorf("failed to parse Bitbucket tree response for %s: %w", subPath, err)
		}

		all = append(all, page.Values...)
		endpoint = page.Next
	}

	if len(all) == 0 {
		return nil, fmt.Errorf("subpath not found in Bitbucket repository: %s", subPath)
	}

	return all, nil
}

func isBitbucketFileEntry(entry bitbucketSourceEntry) bool {
	switch entry.Type {
	case "commit_file", "file":
		return true
	default:
		return false
	}
}

func bitbucketRawURL(owner, repo, ref string, entry bitbucketSourceEntry) string {
	if entry.Links.Self.Href != "" {
		href := entry.Links.Self.Href
		if strings.Contains(href, "?") {
			return href + "&raw=1"
		}
		return href + "?raw=1"
	}

	path := strings.TrimPrefix(entry.Path, "/")
	if path == "" {
		return ""
	}

	return fmt.Sprintf("https://bitbucket.org/%s/%s/raw/%s/%s", owner, repo, ref, path)
}
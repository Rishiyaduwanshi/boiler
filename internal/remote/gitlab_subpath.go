package remote

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

type gitLabTreeEntry struct {
	Path string `json:"path"`
	Type string `json:"type"`
}

func fetchGitLabSubPath(host, owner, repo, ref, subPath, destPath string) error {
	cleanSubPath := normalizeRemoteSubPath(subPath)
	if cleanSubPath == "." {
		return fmt.Errorf("subpath fetch requires a nested path")
	}

	if host == "" {
		host = "gitlab.com"
	}

	debugf("gitlab subtree fetch started host=%s owner=%s repo=%s ref=%s path=%s", host, owner, repo, ref, cleanSubPath)

	if err := prepareSubtreeDestination(destPath); err != nil {
		return err
	}

	entries, err := listGitLabTree(host, owner, repo, ref, cleanSubPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.Type != "blob" {
			continue
		}

		data, err := downloadGitLabRawFile(host, owner, repo, ref, entry.Path)
		if err != nil {
			return err
		}

		if err := writeSubtreeFile(cleanSubPath, entry.Path, destPath, data); err != nil {
			return err
		}

		debugf("downloaded gitlab file path=%s", entry.Path)
	}

	return nil
}

func listGitLabTree(host, owner, repo, ref, subPath string) ([]gitLabTreeEntry, error) {
	projectID := url.QueryEscape(owner + "/" + repo)
	pathQuery := url.QueryEscape(subPath)
	refQuery := url.QueryEscape(ref)

	page := 1
	perPage := 100
	all := make([]gitLabTreeEntry, 0)

	for {
		endpoint := fmt.Sprintf("https://%s/api/v4/projects/%s/repository/tree?path=%s&recursive=true&per_page=%d&page=%d&ref=%s", host, projectID, pathQuery, perPage, page, refQuery)
		debugf("gitlab tree endpoint=%s", endpoint)

		data, err := downloadFile(endpoint)
		if err != nil {
			return nil, fmt.Errorf("failed to list GitLab tree for %s: %w", subPath, err)
		}

		var batch []gitLabTreeEntry
		if err := json.Unmarshal(data, &batch); err != nil {
			return nil, fmt.Errorf("failed to parse GitLab tree response for %s: %w", subPath, err)
		}

		all = append(all, batch...)

		if len(batch) < perPage {
			break
		}
		page++
	}

	if len(all) == 0 {
		return nil, fmt.Errorf("subpath not found in GitLab repository: %s", subPath)
	}

	return all, nil
}

func downloadGitLabRawFile(host, owner, repo, ref, filePath string) ([]byte, error) {
	projectID := url.QueryEscape(owner + "/" + repo)
	fileID := url.QueryEscape(strings.TrimPrefix(filePath, "/"))
	refQuery := url.QueryEscape(ref)
	endpoint := fmt.Sprintf("https://%s/api/v4/projects/%s/repository/files/%s/raw?ref=%s", host, projectID, fileID, refQuery)
	debugf("gitlab raw file endpoint=%s", endpoint)

	data, err := downloadFile(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to download GitLab file %s: %w", filePath, err)
	}
	return data, nil
}

package remote

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

const defaultRemoteRef = "main"

func resolveProviderRef(p Provider, owner, repo, fallback string) string {
	if owner == "" || repo == "" {
		return fallback
	}

	repo = strings.TrimSuffix(repo, ".git")

	switch provider := p.(type) {
	case githubProvider:
		ref, err := githubDefaultBranch(owner, repo)
		if err == nil && ref != "" {
			return ref
		}
	case gitlabProvider:
		host := provider.host
		if host == "" {
			host = "gitlab.com"
		}
		ref, err := gitlabDefaultBranch(host, owner, repo)
		if err == nil && ref != "" {
			return ref
		}
	case bitbucketProvider:
		ref, err := bitbucketDefaultBranch(owner, repo)
		if err == nil && ref != "" {
			return ref
		}
	}

	return fallback
}

func githubDefaultBranch(owner, repo string) (string, error) {
	endpoint := fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, repo)
	data, err := downloadFile(endpoint)
	if err != nil {
		return "", err
	}
	return parseGitHubDefaultBranch(data)
}

func gitlabDefaultBranch(host, owner, repo string) (string, error) {
	projectID := url.QueryEscape(owner + "/" + repo)
	endpoint := fmt.Sprintf("https://%s/api/v4/projects/%s", host, projectID)
	data, err := downloadFile(endpoint)
	if err != nil {
		return "", err
	}
	return parseGitLabDefaultBranch(data)
}

func bitbucketDefaultBranch(owner, repo string) (string, error) {
	endpoint := fmt.Sprintf("https://api.bitbucket.org/2.0/repositories/%s/%s", owner, repo)
	data, err := downloadFile(endpoint)
	if err != nil {
		return "", err
	}
	return parseBitbucketDefaultBranch(data)
}

func parseGitHubDefaultBranch(data []byte) (string, error) {
	var payload struct {
		DefaultBranch string `json:"default_branch"`
	}

	if err := json.Unmarshal(data, &payload); err != nil {
		return "", err
	}

	if strings.TrimSpace(payload.DefaultBranch) == "" {
		return "", fmt.Errorf("default_branch not found")
	}

	return payload.DefaultBranch, nil
}

func parseGitLabDefaultBranch(data []byte) (string, error) {
	var payload struct {
		DefaultBranch string `json:"default_branch"`
	}

	if err := json.Unmarshal(data, &payload); err != nil {
		return "", err
	}

	if strings.TrimSpace(payload.DefaultBranch) == "" {
		return "", fmt.Errorf("default_branch not found")
	}

	return payload.DefaultBranch, nil
}

func parseBitbucketDefaultBranch(data []byte) (string, error) {
	var payload struct {
		MainBranch struct {
			Name string `json:"name"`
		} `json:"mainbranch"`
	}

	if err := json.Unmarshal(data, &payload); err != nil {
		return "", err
	}

	if strings.TrimSpace(payload.MainBranch.Name) == "" {
		return "", fmt.Errorf("mainbranch.name not found")
	}

	return payload.MainBranch.Name, nil
}
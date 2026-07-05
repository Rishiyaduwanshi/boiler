package remote

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
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
		if err != nil {
			debugf("github default branch lookup failed for %s/%s: %v", owner, repo, err)
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
		if err != nil {
			debugf("gitlab default branch lookup failed for %s/%s on %s: %v", owner, repo, host, err)
		}
	case bitbucketProvider:
		ref, err := bitbucketDefaultBranch(owner, repo)
		if err == nil && ref != "" {
			return ref
		}
		if err != nil {
			debugf("bitbucket default branch lookup failed for %s/%s: %v", owner, repo, err)
		}
	}

	if fallback != "" {
		debugf("using fallback ref=%s for %s/%s", fallback, owner, repo)
	}

	return fallback
}

func githubDefaultBranch(owner, repo string) (string, error) {
	endpoint := fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, repo)
	data, err := downloadFile(endpoint)
	if err == nil {
		ref, parseErr := parseGitHubDefaultBranch(data)
		if parseErr == nil {
			return ref, nil
		}
		debugf("github default branch parse from API failed for %s/%s: %v", owner, repo, parseErr)
	}

	debugf("github API default branch lookup failed for %s/%s, trying html fallback: %v", owner, repo, err)
	return githubDefaultBranchFromHTML(owner, repo)
}

func githubDefaultBranchFromHTML(owner, repo string) (string, error) {
	url := fmt.Sprintf("https://github.com/%s/%s", owner, repo)
	data, err := downloadFile(url)
	if err != nil {
		return "", err
	}

	return parseGitHubDefaultBranchFromHTML(data)
}

func parseGitHubDefaultBranchFromHTML(data []byte) (string, error) {

	matchers := []*regexp.Regexp{
		regexp.MustCompile(`"defaultBranch":"([^"]+)"`),
		regexp.MustCompile(`"default_branch":"([^"]+)"`),
	}

	body := string(data)
	for _, re := range matchers {
		matches := re.FindStringSubmatch(body)
		if len(matches) == 2 && strings.TrimSpace(matches[1]) != "" {
			return matches[1], nil
		}
	}

	return "", fmt.Errorf("default branch not found in github html")
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

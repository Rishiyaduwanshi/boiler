package remote

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/rishiyaduwanshi/boiler/internal/constants"
	"github.com/rishiyaduwanshi/boiler/internal/utils"
)

const gitLFSPointerHeader = "version https://git-lfs.github.com/spec/v1"

var runGitClone = cloneGitRepository

func hasGitLFSPointers(dir string) (bool, error) {
	errFound := errors.New("git lfs pointer found")
	err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.Type().IsRegular() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}

		buf := make([]byte, len(gitLFSPointerHeader)+2)
		n, readErr := io.ReadFull(file, buf)
		closeErr := file.Close()
		if readErr != nil && readErr != io.EOF && readErr != io.ErrUnexpectedEOF {
			return readErr
		}
		if closeErr != nil {
			return closeErr
		}

		firstLine := strings.SplitN(string(buf[:n]), "\n", 2)[0]
		if strings.TrimSuffix(firstLine, "\r") == gitLFSPointerHeader {
			return errFound
		}
		return nil
	})
	if errors.Is(err, errFound) {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	return false, nil
}

func providerCloneURL(p Provider, owner, repo string) (string, error) {
	repo = strings.TrimSuffix(repo, ".git")
	if owner == "" || repo == "" {
		return "", fmt.Errorf("remote does not identify a cloneable repository; cannot fall back to git clone")
	}

	switch provider := p.(type) {
	case githubProvider:
		return fmt.Sprintf("%s%s/%s/%s.git", constants.SchemeHTTPS, constants.ProviderGithubHost, owner, repo), nil
	case gitlabProvider:
		host := provider.host
		if host == "" {
			host = constants.ProviderGitlabHost
		}
		return fmt.Sprintf("%s%s/%s/%s.git", constants.SchemeHTTPS, host, owner, repo), nil
	case bitbucketProvider:
		return fmt.Sprintf("%s%s/%s/%s.git", constants.SchemeHTTPS, constants.ProviderBitbucketHost, owner, repo), nil
	default:
		return "", fmt.Errorf("%s remote does not identify a cloneable repository; cannot fall back to git clone", p.Name())
	}
}

func gitCloneArgs(cloneURL, ref, dest string) []string {
	return []string{"clone", "--depth", "1", "--branch", ref, "--single-branch", cloneURL, dest}
}

func cloneGitRepository(p Provider, owner, repo, cloneURL, ref, dest string) error {
	if err := runGit(gitCloneArgs(cloneURL, ref, dest)...); err == nil {
		return nil
	}
	if isAbbreviatedCommitSHA(ref) {
		resolvedRef, err := resolveAbbreviatedCommit(p, owner, repo, ref)
		if err != nil {
			return err
		}
		ref = resolvedRef
	}
	return cloneGitCommit(cloneURL, ref, dest)
}

func isAbbreviatedCommitSHA(ref string) bool {
	if len(ref) < 7 || len(ref) >= 40 {
		return false
	}
	return isHexObjectID(ref)
}

func resolveAbbreviatedCommit(p Provider, owner, repo, ref string) (string, error) {
	repo = strings.TrimSuffix(repo, ".git")
	var endpoint string
	var field string

	switch provider := p.(type) {
	case githubProvider:
		endpoint = fmt.Sprintf("%s%s/repos/%s/%s/commits/%s", constants.SchemeHTTPS, constants.ProviderGithubAPI, url.PathEscape(owner), url.PathEscape(repo), url.PathEscape(ref))
		field = "sha"
	case gitlabProvider:
		host := provider.host
		if host == "" {
			host = constants.ProviderGitlabHost
		}
		project := url.PathEscape(owner + "/" + repo)
		endpoint = fmt.Sprintf("%s%s/%s/%s/repository/commits/%s", constants.SchemeHTTPS, host, constants.ProviderGitlabAPI, project, url.PathEscape(ref))
		field = "id"
	case bitbucketProvider:
		endpoint = fmt.Sprintf("%s%s/repositories/%s/%s/commit/%s", constants.SchemeHTTPS, constants.ProviderBitbucketAPI, url.PathEscape(owner), url.PathEscape(repo), url.PathEscape(ref))
		field = "hash"
	default:
		return "", fmt.Errorf("%s does not support abbreviated commit refs in Git LFS fallback; use a full commit SHA", p.Name())
	}

	data, err := downloadFile(endpoint)
	if err != nil {
		return "", fmt.Errorf("failed to resolve abbreviated commit %q through the %s API: %w", ref, p.Name(), err)
	}

	var payload map[string]json.RawMessage
	if err := json.Unmarshal(data, &payload); err != nil {
		return "", fmt.Errorf("failed to parse abbreviated commit response from the %s API: %w", p.Name(), err)
	}
	var resolved string
	if err := json.Unmarshal(payload[field], &resolved); err != nil || resolved == "" {
		return "", fmt.Errorf("the %s API response did not include a full commit SHA", p.Name())
	}
	if (len(resolved) != 40 && len(resolved) != 64) || !isHexObjectID(resolved) {
		return "", fmt.Errorf("the %s API returned an invalid full commit SHA", p.Name())
	}
	if !strings.HasPrefix(strings.ToLower(resolved), strings.ToLower(ref)) {
		return "", fmt.Errorf("the %s API returned a commit SHA that does not match abbreviated ref %q", p.Name(), ref)
	}
	return resolved, nil
}

func isHexObjectID(ref string) bool {
	for _, char := range ref {
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f') || (char >= 'A' && char <= 'F')) {
			return false
		}
	}
	return ref != ""
}

func cloneGitCommit(cloneURL, ref, dest string) error {
	steps := []struct {
		name string
		args []string
	}{
		{name: "init", args: []string{"init", dest}},
		{name: "remote add", args: []string{"-C", dest, "remote", "add", "origin", cloneURL}},
		{name: "fetch", args: []string{"-C", dest, "fetch", "--depth", "1", "origin", ref}},
		{name: "checkout", args: []string{"-C", dest, "checkout", "--detach", "FETCH_HEAD"}},
	}

	for _, step := range steps {
		if err := runGit(step.args...); err != nil {
			return fmt.Errorf("git %s failed: %w", step.name, err)
		}
	}
	return nil
}

func runGit(args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func fetchStackWithGitClone(p Provider, owner, repo, ref, subPath, tempDir, destPath string) error {
	cloneURL, err := providerCloneURL(p, owner, repo)
	if err != nil {
		return err
	}

	cloneDir := filepath.Join(tempDir, "git-clone")
	if err := runGitClone(p, owner, repo, cloneURL, ref, cloneDir); err != nil {
		return err
	}
	if err := os.RemoveAll(filepath.Join(cloneDir, ".git")); err != nil {
		return fmt.Errorf("failed to remove git metadata: %w", err)
	}

	sourceDir := cloneDir
	if subPath != "" && subPath != "." {
		sourceDir = filepath.Join(cloneDir, filepath.FromSlash(subPath))
	}

	hasPointers, err := hasGitLFSPointers(sourceDir)
	if err != nil {
		return fmt.Errorf("failed to scan cloned stack for Git LFS pointers: %w", err)
	}
	if hasPointers {
		return fmt.Errorf("git clone left unresolved Git LFS pointers; ensure Git LFS is installed and available")
	}

	if err := utils.CopyDir(sourceDir, destPath, nil); err != nil {
		return fmt.Errorf("failed to copy cloned stack: %w", err)
	}
	return nil
}

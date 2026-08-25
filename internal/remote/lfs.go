package remote

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
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

func cloneGitRepository(cloneURL, ref, dest string) error {
	if isCommitSHA(ref) {
		return cloneGitCommit(cloneURL, ref, dest)
	}

	if err := runGit(gitCloneArgs(cloneURL, ref, dest)...); err != nil {
		return fmt.Errorf("git clone failed: %w", err)
	}
	return nil
}

func isCommitSHA(ref string) bool {
	if len(ref) != 40 && len(ref) != 64 {
		return false
	}
	for _, char := range ref {
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f') || (char >= 'A' && char <= 'F')) {
			return false
		}
	}
	return true
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
	if err := runGitClone(cloneURL, ref, cloneDir); err != nil {
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

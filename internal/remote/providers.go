package remote

import (
	"fmt"
	"strings"
)

// Provider abstracts how a remote host serves raw files and stack archives.
// Implement this interface to add support for any new registry host.
//
// To add a new provider (e.g. Bitbucket):
//  1. Create a struct (e.g. BitbucketProvider) implementing Provider.
//  2. Add a detection case in Detect().
//  3. Done - FetchSnippet and FetchStack use it automatically.
type Provider interface {
	// Name returns a human-readable provider name (e.g. "GitHub").
	Name() string

	// RawFileURL returns the direct download URL for a single file.
	// owner/repo may be empty for generic HTTP hosts.
	RawFileURL(owner, repo, ref, filePath string) string

	// ArchiveURL returns the URL of a downloadable archive (tar.gz or zip)
	// for a full repo or sub-path.
	ArchiveURL(owner, repo, ref, subPath string) string

	// ArchiveFormat returns the format of the archive: "tar.gz" or "zip".
	ArchiveFormat() string
}

// ---------------------------------------------------------------------------
// GitHub
// ---------------------------------------------------------------------------

type githubProvider struct{}

func (githubProvider) Name() string { return "GitHub" }

func (githubProvider) RawFileURL(owner, repo, ref, filePath string) string {
	return fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s", owner, repo, ref, filePath)
}

func (githubProvider) ArchiveURL(owner, repo, ref, _ string) string {
	return fmt.Sprintf("https://api.github.com/repos/%s/%s/tarball/%s", owner, repo, ref)
}

func (githubProvider) ArchiveFormat() string { return "tar.gz" }

// ---------------------------------------------------------------------------
// GitLab
// ---------------------------------------------------------------------------

type gitlabProvider struct{ host string }

func (p gitlabProvider) Name() string { return "GitLab" }

func (p gitlabProvider) RawFileURL(owner, repo, ref, filePath string) string {
	return fmt.Sprintf("https://%s/%s/%s/-/raw/%s/%s", p.host, owner, repo, ref, filePath)
}

func (p gitlabProvider) ArchiveURL(owner, repo, ref, _ string) string {
	// URL-encode the namespace/project as required by GitLab API
	project := strings.ReplaceAll(owner+"%2F"+repo, "/", "%2F")
	return fmt.Sprintf("https://%s/api/v4/projects/%s/repository/archive.tar.gz?sha=%s", p.host, project, ref)
}

func (gitlabProvider) ArchiveFormat() string { return "tar.gz" }

// ---------------------------------------------------------------------------
// Bitbucket
// ---------------------------------------------------------------------------

type bitbucketProvider struct{}

func (bitbucketProvider) Name() string { return "Bitbucket" }

func (bitbucketProvider) RawFileURL(owner, repo, ref, filePath string) string {
	return fmt.Sprintf("https://bitbucket.org/%s/%s/raw/%s/%s", owner, repo, ref, filePath)
}

func (bitbucketProvider) ArchiveURL(owner, repo, ref, _ string) string {
	return fmt.Sprintf("https://bitbucket.org/%s/%s/get/%s.zip", owner, repo, ref)
}

func (bitbucketProvider) ArchiveFormat() string { return "zip" }

// ---------------------------------------------------------------------------
// Generic HTTP host
//
// Expected server structure (same layout as any boiler registry):
//   <base>/store/boiler.meta.json        ← registry index
//   <base>/store/stacks/<name>.zip       ← stack archives
//   <base>/store/snippets/<file>         ← snippet files
//
// In boiler.meta.json, values must be the full download URL of the resource:
//   "express@1":         "https://example.com/store/stacks/express.zip"
//   "errorHandler@1.js": "https://example.com/store/snippets/errorHandler@1.js"
// ---------------------------------------------------------------------------

type genericProvider struct{ base string }

func (p genericProvider) Name() string { return "Generic" }

func (p genericProvider) RawFileURL(_, _, _, filePath string) string {
	return fmt.Sprintf("%s/%s", p.base, filePath)
}

func (p genericProvider) ArchiveURL(_, name, _, _ string) string {
	return fmt.Sprintf("%s/store/stacks/%s.zip", p.base, name)
}

func (genericProvider) ArchiveFormat() string { return "zip" }

// ---------------------------------------------------------------------------
// Detect returns the right Provider for a given registry URL.
//
// Detection order:
//  1. github.com / raw.githubusercontent.com  → GitHubProvider
//  2. gitlab.com or self-hosted GitLab         → GitLabProvider
//  3. bitbucket.org                            → BitbucketProvider
//  4. anything else                            → GenericProvider
// ---------------------------------------------------------------------------

func Detect(registryURL string) Provider {
	// Normalize
	u := strings.ToLower(strings.TrimSuffix(registryURL, "/"))
	if !strings.HasPrefix(u, "http://") && !strings.HasPrefix(u, "https://") {
		u = "https://" + u
	}

	switch {
	case strings.Contains(u, "github.com") || strings.Contains(u, "raw.githubusercontent.com"):
		return githubProvider{}

	case strings.Contains(u, "gitlab.com"):
		return gitlabProvider{host: "gitlab.com"}

	case strings.Contains(u, "bitbucket.org"):
		return bitbucketProvider{}

	default:
		// Strip scheme for generic base URL construction
		base := registryURL
		if !strings.HasPrefix(base, "http://") && !strings.HasPrefix(base, "https://") {
			base = "https://" + base
		}
		base = strings.TrimSuffix(base, "/")
		return genericProvider{base: base}
	}
}

// parseOwnerRepo extracts the owner and repo from a registry URL.
//
// Examples:
//
//	"https://github.com/alice/my-store"  → "alice", "my-store"
//	"https://gitlab.com/alice/store"     → "alice", "store"
//	"https://mysite.com/boiler"          → "", ""
func parseOwnerRepo(registryURL string) (owner, repo string) {
	// Strip common prefixes
	for _, prefix := range []string{
		"https://github.com/",
		"http://github.com/",
		"https://gitlab.com/",
		"http://gitlab.com/",
		"https://bitbucket.org/",
		"http://bitbucket.org/",
	} {
		if strings.HasPrefix(registryURL, prefix) {
			rest := strings.TrimPrefix(registryURL, prefix)
			rest = strings.TrimSuffix(rest, "/")
			parts := strings.SplitN(rest, "/", 2)
			if len(parts) == 2 {
				return parts[0], parts[1]
			}
			return rest, ""
		}
	}
	return "", ""
}

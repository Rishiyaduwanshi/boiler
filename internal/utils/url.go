package utils

import (
	"net/url"
	"path"
	"path/filepath"
	"strings"
)

func IsURL(value string) bool {
	lower := strings.ToLower(strings.TrimSpace(value))
	return strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://")
}

// IsDirectRemoteFileURL reports whether a URL points to a single file
// (as opposed to a repository root or an archive).
func IsDirectRemoteFileURL(resource string) bool {
	lower := strings.ToLower(resource)
	trimmed := strings.SplitN(resource, "?", 2)[0]
	trimmed = strings.SplitN(trimmed, "#", 2)[0]

	parsed, err := url.Parse(trimmed)
	if err == nil {
		host := strings.ToLower(parsed.Hostname())
		segments := strings.Split(strings.Trim(parsed.Path, "/"), "/")

		switch host {
		case "github.com", "www.github.com":
			if len(segments) >= 4 && segments[2] == "blob" {
				if filepath.Ext(segments[len(segments)-1]) == "" {
					return false
				}
				return true
			}
		case "gitlab.com", "www.gitlab.com":
			if len(segments) >= 5 && segments[2] == "-" && segments[3] == "blob" {
				if filepath.Ext(segments[len(segments)-1]) == "" {
					return false
				}
				return true
			}
		}
	}

	if strings.Contains(lower, "raw.githubusercontent.com/") ||
		strings.Contains(lower, "/-/raw/") {
		return true
	}

	switch strings.ToLower(filepath.Ext(trimmed)) {
	case "", ".zip", ".tar", ".gz", ".tgz":
		return false
	default:
		return true
	}
}

// FileNameFromRemoteURL extracts the filename portion of a URL.
func FileNameFromRemoteURL(remotePath string) string {
	parsed, err := url.Parse(remotePath)
	if err == nil {
		name := path.Base(parsed.Path)
		if name != "" && name != "." && name != "/" {
			return name
		}
	}

	name := filepath.Base(remotePath)
	if name == "" || name == "." || name == "/" {
		return ""
	}
	return name
}

// TrimArchiveSuffix strips common archive extensions (.tar.gz, .tgz, .zip).
func TrimArchiveSuffix(name string) string {
	lower := strings.ToLower(name)
	switch {
	case strings.HasSuffix(lower, ".tar.gz"):
		return name[:len(name)-len(".tar.gz")]
	case strings.HasSuffix(lower, ".tgz"):
		return name[:len(name)-len(".tgz")]
	case strings.HasSuffix(lower, ".zip"):
		return name[:len(name)-len(".zip")]
	default:
		return name
	}
}

// StackNameFromRemoteURL derives a suitable stack name from a URL.
// For known hosting providers, it uses the repo name; otherwise the last path segment.
func StackNameFromRemoteURL(remotePath string) string {
	parsed, err := url.Parse(remotePath)
	if err == nil {
		host := strings.ToLower(parsed.Hostname())
		segments := strings.Split(strings.Trim(parsed.Path, "/"), "/")

		if (host == "github.com" || host == "gitlab.com" || host == "bitbucket.org") && len(segments) >= 2 {
			repo := strings.TrimSuffix(segments[1], ".git")
			if repo != "" {
				return repo
			}
		}

		if len(segments) > 0 {
			name := TrimArchiveSuffix(segments[len(segments)-1])
			if name != "" && name != "." && name != "/" {
				return name
			}
		}
	}

	return "remote-stack"
}

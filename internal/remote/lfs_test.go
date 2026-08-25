package remote

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestHasGitLFSPointers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		files   map[string]string
		wantLFS bool
	}{
		{
			name: "nested pointer",
			files: map[string]string{
				"assets/model.bin": "version https://git-lfs.github.com/spec/v1\noid sha256:0123456789abcdef\nsize 42\n",
				"main.go":          "package main\n",
			},
			wantLFS: true,
		},
		{
			name: "windows line ending",
			files: map[string]string{
				"video.mp4": "version https://git-lfs.github.com/spec/v1\r\noid sha256:0123456789abcdef\r\nsize 42\r\n",
			},
			wantLFS: true,
		},
		{
			name: "normal files",
			files: map[string]string{
				"README.md": "# Example\n",
				"large.bin": "ordinary content without a newline",
			},
		},
		{
			name: "header prefix is not a pointer",
			files: map[string]string{
				"notes.txt": "version https://git-lfs.github.com/spec/v1-extra\n",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			for name, content := range tt.files {
				path := filepath.Join(root, filepath.FromSlash(name))
				if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte(content), 0644); err != nil {
					t.Fatal(err)
				}
			}

			got, err := hasGitLFSPointers(root)
			if err != nil {
				t.Fatalf("hasGitLFSPointers() error = %v", err)
			}
			if got != tt.wantLFS {
				t.Fatalf("hasGitLFSPointers() = %v, want %v", got, tt.wantLFS)
			}
		})
	}
}

func TestHasGitLFSPointersMissingDirectory(t *testing.T) {
	t.Parallel()

	if _, err := hasGitLFSPointers(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Fatal("expected an error for a missing directory")
	}
}

func TestGitCloneArgs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		ref  string
		want []string
	}{
		{
			name: "implicitly resolved branch",
			ref:  "main",
			want: []string{"clone", "--depth", "1", "--branch", "main", "--single-branch", "https://github.com/alice/repo.git", "dest"},
		},
		{
			name: "explicit branch",
			ref:  "release",
			want: []string{"clone", "--depth", "1", "--branch", "release", "--single-branch", "https://github.com/alice/repo.git", "dest"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := gitCloneArgs("https://github.com/alice/repo.git", tt.ref, "dest")
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("gitCloneArgs() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestProviderCloneURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		p       Provider
		want    string
		wantErr bool
	}{
		{name: "github", p: githubProvider{}, want: "https://github.com/alice/repo.git"},
		{name: "gitlab", p: gitlabProvider{host: "gitlab.example.com"}, want: "https://gitlab.example.com/alice/repo.git"},
		{name: "bitbucket", p: bitbucketProvider{}, want: "https://bitbucket.org/alice/repo.git"},
		{name: "generic archive", p: genericProvider{base: "https://example.com"}, wantErr: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := providerCloneURL(tt.p, "alice", "repo")
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected an error")
				}
				return
			}
			if err != nil {
				t.Fatalf("providerCloneURL() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("providerCloneURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFetchStackFallsBackToShallowCloneForLFSPointer(t *testing.T) {
	archive := testTarGz(t, map[string]string{
		"repo-main/assets/model.bin": "version https://git-lfs.github.com/spec/v1\noid sha256:0123456789abcdef\nsize 42\n",
	})
	useGitHubArchive(t, archive)

	previousRunGitClone := runGitClone
	defer func() { runGitClone = previousRunGitClone }()

	var gotURL string
	var gotRef string
	runGitClone = func(cloneURL, ref, dest string) error {
		gotURL = cloneURL
		gotRef = ref
		if err := os.MkdirAll(filepath.Join(dest, ".git"), 0755); err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Join(dest, "assets"), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(dest, ".git", "config"), []byte("metadata"), 0644); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(dest, "assets", "model.bin"), []byte("resolved lfs content"), 0644)
	}

	dest := filepath.Join(t.TempDir(), "stack")
	if err := FetchStack("alice/repo", dest); err != nil {
		t.Fatalf("FetchStack() error = %v", err)
	}

	if gotURL != "https://github.com/alice/repo.git" {
		t.Fatalf("clone URL = %q", gotURL)
	}
	if gotRef != "main" {
		t.Fatalf("clone ref = %q, want main", gotRef)
	}
	data, err := os.ReadFile(filepath.Join(dest, "assets", "model.bin"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "resolved lfs content" {
		t.Fatalf("model content = %q", data)
	}
	if _, err := os.Stat(filepath.Join(dest, ".git")); !os.IsNotExist(err) {
		t.Fatalf(".git metadata was not removed: %v", err)
	}
}

func TestFetchStackPreservesExplicitBranchForLFSFallback(t *testing.T) {
	archive := testTarGz(t, map[string]string{
		"repo-release/asset.bin": gitLFSPointerHeader + "\noid sha256:0123456789abcdef\nsize 42\n",
	})
	useGitHubArchive(t, archive)

	previousRunGitClone := runGitClone
	defer func() { runGitClone = previousRunGitClone }()

	var gotRef string
	runGitClone = func(_, ref, dest string) error {
		gotRef = ref
		if err := os.MkdirAll(dest, 0755); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(dest, "asset.bin"), []byte("resolved"), 0644)
	}

	dest := filepath.Join(t.TempDir(), "stack")
	if err := FetchStack("https://github.com/alice/repo/tree/release", dest); err != nil {
		t.Fatalf("FetchStack() error = %v", err)
	}
	if gotRef != "release" {
		t.Fatalf("clone ref = %q, want release", gotRef)
	}
}

func TestFetchStackKeepsArchivePathWithoutLFSPointers(t *testing.T) {
	archive := testTarGz(t, map[string]string{
		"repo-main/main.go": "package main\n",
	})
	useGitHubArchive(t, archive)

	previousRunGitClone := runGitClone
	defer func() { runGitClone = previousRunGitClone }()
	cloneCalled := false
	runGitClone = func(_, _, _ string) error {
		cloneCalled = true
		return fmt.Errorf("unexpected clone")
	}

	dest := filepath.Join(t.TempDir(), "stack")
	if err := FetchStack("alice/repo", dest); err != nil {
		t.Fatalf("FetchStack() error = %v", err)
	}
	if cloneCalled {
		t.Fatal("normal archive unexpectedly triggered git clone")
	}
	if _, err := os.Stat(filepath.Join(dest, "main.go")); err != nil {
		t.Fatal(err)
	}
}

func TestGitCloneFallbackRejectsUnresolvedPointersWithoutChangingDestination(t *testing.T) {
	previousRunGitClone := runGitClone
	defer func() { runGitClone = previousRunGitClone }()
	runGitClone = func(_, _, dest string) error {
		if err := os.MkdirAll(dest, 0755); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(dest, "asset.bin"), []byte(gitLFSPointerHeader+"\n"), 0644)
	}

	dest := t.TempDir()
	sentinel := filepath.Join(dest, "sentinel.txt")
	if err := os.WriteFile(sentinel, []byte("keep"), 0644); err != nil {
		t.Fatal(err)
	}

	err := fetchStackWithGitClone(githubProvider{}, "alice", "repo", "main", ".", t.TempDir(), dest)
	if err == nil || !strings.Contains(err.Error(), "unresolved Git LFS pointers") {
		t.Fatalf("fetchStackWithGitClone() error = %v", err)
	}
	if _, err := os.Stat(sentinel); err != nil {
		t.Fatalf("destination changed after clone validation failed: %v", err)
	}
}

func TestFetchStackRejectsLFSPointerFromGenericArchive(t *testing.T) {
	archive := testTarGz(t, map[string]string{
		"asset.bin": "version https://git-lfs.github.com/spec/v1\noid sha256:0123456789abcdef\nsize 42\n",
	})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/gzip")
		_, _ = w.Write(archive)
	}))
	defer server.Close()

	err := FetchStack(server.URL+"/stack.tar.gz", filepath.Join(t.TempDir(), "stack"))
	if err == nil || !strings.Contains(err.Error(), "cannot fall back to git clone") {
		t.Fatalf("FetchStack() error = %v, want clone fallback error", err)
	}
}

func useGitHubArchive(t *testing.T, archive []byte) {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/alice/repo":
			fmt.Fprint(w, `{"default_branch":"main"}`)
		case "/repos/alice/repo/tarball/main", "/repos/alice/repo/tarball/release":
			w.Header().Set("Content-Type", "application/gzip")
			_, _ = w.Write(archive)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	serverURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	previousTransport := http.DefaultTransport
	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		redirected := req.Clone(req.Context())
		redirected.URL.Scheme = serverURL.Scheme
		redirected.URL.Host = serverURL.Host
		return previousTransport.RoundTrip(redirected)
	})
	t.Cleanup(func() { http.DefaultTransport = previousTransport })
}

func testTarGz(t *testing.T, files map[string]string) []byte {
	t.Helper()

	var archive strings.Builder
	gz := gzip.NewWriter(&archive)
	tw := tar.NewWriter(gz)
	for name, content := range files {
		header := &tar.Header{Name: name, Mode: 0644, Size: int64(len(content))}
		if err := tw.WriteHeader(header); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return []byte(archive.String())
}

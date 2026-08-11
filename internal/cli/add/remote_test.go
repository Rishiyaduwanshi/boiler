package add

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/rishiyaduwanshi/boiler/internal/config"
)

func TestShouldTreatDirectRemotePathAsSnippet(t *testing.T) {
	tests := []struct {
		name         string
		subPath      string
		resourceType ResourceType
		want         bool
	}{
		{name: "auto file path", subPath: "js/errorHandler.js", resourceType: ResourceTypeAuto, want: true},
		{name: "auto directory path", subPath: "templates", resourceType: ResourceTypeAuto, want: false},
		{name: "force snippet", subPath: "templates", resourceType: ResourceTypeSnippet, want: true},
		{name: "force stack", subPath: "js/errorHandler.js", resourceType: ResourceTypeStack, want: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := shouldTreatDirectRemotePathAsSnippet(tt.subPath, tt.resourceType)
			if got != tt.want {
				t.Fatalf("shouldTreatDirectRemotePathAsSnippet(%q, %v) = %v, want %v", tt.subPath, tt.resourceType, got, tt.want)
			}
		})
	}
}

func TestShouldTreatDirectURLAsSnippet(t *testing.T) {
	tests := []struct {
		name         string
		resourceURL  string
		resourceType ResourceType
		want         bool
	}{
		{name: "auto file url", resourceURL: "https://example.com/snippets/helper.js", resourceType: ResourceTypeAuto, want: true},
		{name: "auto archive url", resourceURL: "https://example.com/templates/express.zip", resourceType: ResourceTypeAuto, want: false},
		{name: "force snippet", resourceURL: "https://example.com/templates/noext", resourceType: ResourceTypeSnippet, want: true},
		{name: "force stack", resourceURL: "https://example.com/snippets/helper.js", resourceType: ResourceTypeStack, want: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := shouldTreatDirectURLAsSnippet(tt.resourceURL, tt.resourceType)
			if got != tt.want {
				t.Fatalf("shouldTreatDirectURLAsSnippet(%q, %v) = %v, want %v", tt.resourceURL, tt.resourceType, got, tt.want)
			}
		})
	}
}

// countingRoundTripper records HTTP attempts. Used to assert that remote
// pre-fetch validation fails before any download/network call.
type countingRoundTripper struct {
	hits atomic.Int64
}

func (c *countingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	c.hits.Add(1)
	return nil, fmt.Errorf("unexpected network call to %s (test transport)", req.URL)
}

func testRemoteAddConfig(t *testing.T) *config.Config {
	t.Helper()
	root := t.TempDir()
	return &config.Config{
		Paths: &config.Paths{
			Root:     root,
			Store:    filepath.Join(root, "store"),
			Snippets: filepath.Join(root, "store", "snippets"),
			Stacks:   filepath.Join(root, "store", "stacks"),
			Logs:     filepath.Join(root, "logs"),
			Bin:      filepath.Join(root, "bin"),
		},
		Vars:    map[string]string{},
		Aliases: map[string]string{},
	}
}

// TestResourceFromRemote_OccupiedDestSkipsFetch exercises the remote stack
// add path (ResourceFromRemote → directRemote*) and asserts that destination
// validation fails before any FetchStack/network download. Covers both
// stored and --no-store flows so a future call-order regression is caught.
func TestResourceFromRemote_OccupiedDestSkipsFetch(t *testing.T) {
	// Mutates http.DefaultTransport; do not run in parallel with other tests
	// that perform real HTTP.
	transport := &countingRoundTripper{}
	prev := http.DefaultTransport
	http.DefaultTransport = transport
	defer func() { http.DefaultTransport = prev }()

	cfg := testRemoteAddConfig(t)

	tests := []struct {
		name      string
		resource  string
		stackDir  string // dest/<stackDir> conflict path
		noStore   bool
		forceType ResourceType
	}{
		{
			name:      "direct owner/repo stored",
			resource:  "alice/my-stack",
			stackDir:  "my-stack", // directRemoteResource uses repo as stack name
			noStore:   false,
			forceType: ResourceTypeStack,
		},
		{
			name:      "direct owner/repo no-store",
			resource:  "alice/my-stack",
			stackDir:  "my-stack",
			noStore:   true,
			forceType: ResourceTypeStack,
		},
		{
			name:      "direct archive URL stored",
			resource:  "https://example.com/archives/demo-stack.zip",
			stackDir:  "demo-stack",
			noStore:   false,
			forceType: ResourceTypeStack,
		},
		{
			name:      "direct archive URL no-store",
			resource:  "https://example.com/archives/demo-stack.zip",
			stackDir:  "demo-stack",
			noStore:   true,
			forceType: ResourceTypeStack,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			destRoot := t.TempDir()
			conflict := filepath.Join(destRoot, tt.stackDir)
			if err := os.Mkdir(conflict, 0755); err != nil {
				t.Fatal(err)
			}

			before := transport.hits.Load()
			err := ResourceFromRemote(tt.resource, destRoot, tt.forceType, tt.noStore, Options{}, cfg, nil)
			after := transport.hits.Load()

			if err == nil {
				t.Fatal("expected destination-exists error, got nil")
			}
			if !strings.Contains(err.Error(), "already exists") {
				t.Fatalf("error %q should mention already exists", err)
			}
			if after != before {
				t.Fatalf("expected no network/FetchStack after dest conflict; HTTP hits %d -> %d", before, after)
			}
		})
	}
}

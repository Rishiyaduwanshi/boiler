package add

import (
	"testing"
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

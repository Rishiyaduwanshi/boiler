package remote

import "testing"

func TestParseHostedStackRefAndSubPath(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantRef     string
		wantSubPath string
		wantOK      bool
	}{
		{
			name:        "github tree subpath",
			input:       "https://github.com/spf13/cobra/tree/main/doc",
			wantRef:     "main",
			wantSubPath: "doc",
			wantOK:      true,
		},
		{
			name:        "github blob subpath",
			input:       "https://github.com/spf13/cobra/blob/main/doc",
			wantRef:     "main",
			wantSubPath: "doc",
			wantOK:      true,
		},
		{
			name:        "github tree root",
			input:       "https://github.com/spf13/cobra/tree/main",
			wantRef:     "main",
			wantSubPath: ".",
			wantOK:      true,
		},
		{
			name:        "gitlab tree subpath",
			input:       "https://gitlab.com/group/repo/-/tree/main/templates/api",
			wantRef:     "main",
			wantSubPath: "templates/api",
			wantOK:      true,
		},
		{
			name:        "bitbucket src subpath",
			input:       "https://bitbucket.org/team/repo/src/main/pkg",
			wantRef:     "main",
			wantSubPath: "pkg",
			wantOK:      true,
		},
		{
			name:   "regular repo url",
			input:  "https://github.com/spf13/cobra",
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRef, gotSubPath, gotOK := parseHostedStackRefAndSubPath(tt.input)
			if gotOK != tt.wantOK {
				t.Fatalf("ok = %v, want %v", gotOK, tt.wantOK)
			}
			if gotRef != tt.wantRef {
				t.Fatalf("ref = %q, want %q", gotRef, tt.wantRef)
			}
			if gotSubPath != tt.wantSubPath {
				t.Fatalf("subPath = %q, want %q", gotSubPath, tt.wantSubPath)
			}
		})
	}
}

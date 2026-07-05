package utils

import (
	"testing"
)

func TestIsDirectRemoteFileURL(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "github blob file",
			input: "https://github.com/rishiyaduwanshi/boiler-snippets/blob/main/js/errorHandler.js",
			want:  true,
		},
		{
			name:  "github blob directory",
			input: "https://github.com/Rich-Harris/degit/blob/master/src",
			want:  false,
		},
		{
			name:  "github tree directory",
			input: "https://github.com/spf13/cobra/tree/main/doc",
			want:  false,
		},
		{
			name:  "raw github file",
			input: "https://raw.githubusercontent.com/rishiyaduwanshi/boiler-snippets/main/js/errorHandler.js",
			want:  true,
		},
		{
			name:  "direct archive url",
			input: "https://example.com/templates/express.zip",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := IsDirectRemoteFileURL(tt.input)
			if got != tt.want {
				t.Fatalf("IsDirectRemoteFileURL(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

package remote

import "testing"

func TestParseGitHubDefaultBranch(t *testing.T) {
	ref, err := parseGitHubDefaultBranch([]byte(`{"default_branch":"canary"}`))
	if err != nil {
		t.Fatalf("parseGitHubDefaultBranch returned error: %v", err)
	}
	if ref != "canary" {
		t.Fatalf("ref = %q, want %q", ref, "canary")
	}
}

func TestParseGitLabDefaultBranch(t *testing.T) {
	ref, err := parseGitLabDefaultBranch([]byte(`{"default_branch":"develop"}`))
	if err != nil {
		t.Fatalf("parseGitLabDefaultBranch returned error: %v", err)
	}
	if ref != "develop" {
		t.Fatalf("ref = %q, want %q", ref, "develop")
	}
}

func TestParseBitbucketDefaultBranch(t *testing.T) {
	ref, err := parseBitbucketDefaultBranch([]byte(`{"mainbranch":{"name":"master"}}`))
	if err != nil {
		t.Fatalf("parseBitbucketDefaultBranch returned error: %v", err)
	}
	if ref != "master" {
		t.Fatalf("ref = %q, want %q", ref, "master")
	}
}

func TestParseDefaultBranchErrorsOnMissingField(t *testing.T) {
	if _, err := parseGitHubDefaultBranch([]byte(`{"name":"repo"}`)); err == nil {
		t.Fatal("expected error when github default branch field is missing")
	}
	if _, err := parseGitLabDefaultBranch([]byte(`{"id":1}`)); err == nil {
		t.Fatal("expected error when gitlab default branch field is missing")
	}
	if _, err := parseBitbucketDefaultBranch([]byte(`{"mainbranch":{}}`)); err == nil {
		t.Fatal("expected error when bitbucket main branch name is missing")
	}
}

func TestParseGitHubDefaultBranchFromHTML(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "defaultBranch key", input: `{"defaultBranch":"canary"}`, want: "canary"},
		{name: "default_branch key", input: `{"default_branch":"stable"}`, want: "stable"},
		{name: "missing", input: `<html></html>`, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseGitHubDefaultBranchFromHTML([]byte(tt.input))
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("parseGitHubDefaultBranchFromHTML returned error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("branch = %q, want %q", got, tt.want)
			}
		})
	}
}

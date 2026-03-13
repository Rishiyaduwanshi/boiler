package cli

import (
	"reflect"
	"strings"
	"testing"

	"github.com/rishiyaduwanshi/boiler/internal/config"
)

func TestSetAliasFromAssignment_NormalizesAndPersists(t *testing.T) {
	setupVarCommandTest(t)

	if err := setAliasFromAssignment("LL=ls"); err != nil {
		t.Fatalf("setAliasFromAssignment: %v", err)
	}

	if got := cfg.Aliases["ll"]; got != "ls" {
		t.Fatalf("cfg.Aliases[ll] = %q", got)
	}

	loaded, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got := loaded.Aliases["ll"]; got != "ls" {
		t.Fatalf("loaded.Aliases[ll] = %q", got)
	}
}

func TestSetAliasFromAssignment_RejectsBuiltInCommandName(t *testing.T) {
	setupVarCommandTest(t)

	err := setAliasFromAssignment("ls=search")
	if err == nil {
		t.Fatal("expected conflict error for built-in command token")
	}
	if !strings.Contains(err.Error(), "conflicts") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUnsetAlias_RemovesNormalizedKey(t *testing.T) {
	setupVarCommandTest(t)

	if err := setAliasFromAssignment("ss=search"); err != nil {
		t.Fatalf("setAliasFromAssignment: %v", err)
	}
	if err := unsetAlias("SS"); err != nil {
		t.Fatalf("unsetAlias: %v", err)
	}

	if _, ok := cfg.Aliases["ss"]; ok {
		t.Fatalf("expected ss to be removed")
	}

	loaded, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if _, ok := loaded.Aliases["ss"]; ok {
		t.Fatalf("expected ss to be removed from persisted config")
	}
}

func TestExpandFirstCommandAlias_ReplacesFirstToken(t *testing.T) {
	setupVarCommandTest(t)
	cfg.Aliases["ll"] = "ls"

	got := expandFirstCommandAlias([]string{"ll", "--snippets"})
	want := []string{"ls", "--snippets"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expandFirstCommandAlias() = %v, want %v", got, want)
	}
}

func TestExpandFirstCommandAlias_IgnoresNonAliasTokens(t *testing.T) {
	setupVarCommandTest(t)
	cfg.Aliases["ll"] = "ls"

	for _, input := range [][]string{{"--help"}, {"unknown"}} {
		got := expandFirstCommandAlias(input)
		if !reflect.DeepEqual(got, input) {
			t.Fatalf("expandFirstCommandAlias(%v) = %v, want %v", input, got, input)
		}
	}
}

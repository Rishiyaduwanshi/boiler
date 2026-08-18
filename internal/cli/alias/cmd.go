package alias

import (
	"fmt"
	"github.com/rishiyaduwanshi/boiler/internal/utils"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/rishiyaduwanshi/boiler/internal/config"
	"github.com/spf13/cobra"
)

var aliasNameRe = regexp.MustCompile(`^[a-z][a-z0-9_-]*$`)
var positionalVarRe = regexp.MustCompile(`(?i)bl__([0-9]+)`)

const maxAliasExpansionDepth = 8

var Cmd = &cobra.Command{
	Use:   "alias [name|name=command [args...]]",
	Short: "Manage reusable command aliases",
	Long: `Manage command aliases stored in boiler.conf.json.

Usage patterns:

  bl alias              List all aliases
  bl alias name=cmd     Set or update an alias
  bl alias name         Get one alias value
  bl unalias name       Remove an alias (use 'unalias' command)

Positional Arguments & Variables:
- Aliases support dynamic positional arguments (bl__1, bl__2, ...) and inline variables (bl__VAR_NAME).
- When positional arguments are used, any trailing unconsumed arguments are automatically appended.

Alias names are case-insensitive. Hyphens and underscores are preserved.

` + "```" + `bash
# Simple aliases
bl alias ll=ls
bl alias s=search

# Alias with positional argument - bl gi Node fetches Node.gitignore
bl alias gi='add github/gitignore:bl__1.gitignore . -m .gitignore -r'

# Alias referencing a stored var
bl var org=rishiyaduwanshi
bl alias templates='search --registry https://github.com/bl__org/boiler'
` + "```",
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var err error

		switch len(args) {
		case 0:
			err = listAliases()
		default:
			input := strings.TrimSpace(args[0])
			if strings.Contains(input, "=") {
				err = setAliasFromAssignment(input)
			} else {
				err = getAliasByName(input)
			}
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func EnsureConfigAliases() {
	if cfg.Aliases == nil {
		cfg.Aliases = make(map[string]string)
		return
	}

	normalized := make(map[string]string, len(cfg.Aliases))
	for rawName, rawTarget := range cfg.Aliases {
		name, err := NormalizeAliasName(rawName)
		if err != nil {
			continue
		}
		target, err := normalizeAliasTarget(rawTarget)
		if err != nil {
			continue
		}
		normalized[name] = target
	}
	cfg.Aliases = normalized
}

// setScopedAlias writes a single alias into the config object for the active scope.
// This ensures only the scope-owned file is updated — not the merged Runtime map —
// preserving isolation between global and local alias sets.
func setScopedAlias(name, target string) error {
	m := config.Ctx.Manager
	if m == nil {
		return fmt.Errorf("config manager is not initialized")
	}
	if config.Ctx.Scope == config.ScopeLocal {
		if m.Local == nil {
			return fmt.Errorf("local config is not initialized")
		}
		if m.Local.Config == nil {
			m.Local.Config = &config.Config{}
		}
		if m.Local.Config.Aliases == nil {
			m.Local.Config.Aliases = make(map[string]string)
		}
		m.Local.Config.Aliases[name] = target
	} else {
		if m.Global == nil {
			return fmt.Errorf("global config is not initialized")
		}
		if m.Global.Config == nil {
			m.Global.Config = &config.Config{}
		}
		if m.Global.Config.Aliases == nil {
			m.Global.Config.Aliases = make(map[string]string)
		}
		m.Global.Config.Aliases[name] = target
	}
	return nil
}

// DeleteScopedAlias removes a single alias from the config object for the active scope.
// Exported so the unalias package can call it to maintain scope isolation on deletion.
func DeleteScopedAlias(name string) error {
	m := config.Ctx.Manager
	if m == nil {
		return fmt.Errorf("config manager is not initialized")
	}
	if config.Ctx.Scope == config.ScopeLocal {
		if m.Local == nil || m.Local.Config == nil {
			return nil // nothing to delete
		}
		delete(m.Local.Config.Aliases, name)
	} else {
		if m.Global == nil || m.Global.Config == nil {
			return nil // nothing to delete
		}
		delete(m.Global.Config.Aliases, name)
	}
	return nil
}

// PersistConfigAliases saves the active scope's config file to disk.
// The alias must already be set on the scope-owned config object (via setScopedAlias
// or DeleteScopedAlias) before calling this — it does NOT copy from Runtime.
func PersistConfigAliases() error {
	m := config.Ctx.Manager
	if m == nil {
		return fmt.Errorf("config manager is not initialized")
	}
	if config.Ctx.Scope == config.ScopeLocal {
		if m.Local == nil {
			return fmt.Errorf("local config is not initialized")
		}
		if err := m.SaveLocal(); err != nil {
			return fmt.Errorf("failed to save local config: %w", err)
		}
	} else {
		if m.Global == nil {
			return fmt.Errorf("global config is not initialized")
		}
		if err := m.SaveGlobal(); err != nil {
			return fmt.Errorf("failed to save global config: %w", err)
		}
	}
	return nil
}

func NormalizeAliasName(raw string) (string, error) {
	name := strings.ToLower(strings.TrimSpace(raw))
	if !aliasNameRe.MatchString(name) {
		return "", fmt.Errorf("invalid alias name %q: use letters, numbers, underscores, and hyphens", raw)
	}
	return name, nil
}

func normalizeAliasTarget(raw string) (string, error) {
	target := strings.TrimSpace(raw)
	if target == "" {
		return "", fmt.Errorf("invalid alias target %q: expected command", raw)
	}

	target = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(target), "bl "))
	if target == "" {
		return "", fmt.Errorf("invalid alias target %q: expected command", raw)
	}

	return target, nil
}

func parseAliasTargetTokens(raw string) ([]string, error) {
	target, err := normalizeAliasTarget(raw)
	if err != nil {
		return nil, err
	}

	tokens := strings.Fields(target)
	if len(tokens) == 0 {
		return nil, fmt.Errorf("invalid alias target %q: expected command", raw)
	}

	return tokens, nil
}

func setAliasFromAssignment(assignment string) error {
	rawName, rawTarget, found := strings.Cut(assignment, "=")
	if !found {
		return fmt.Errorf("invalid alias assignment; use NAME=COMMAND")
	}

	name, err := NormalizeAliasName(rawName)
	if err != nil {
		return err
	}
	target, err := normalizeAliasTarget(rawTarget)
	if err != nil {
		return err
	}
	targetTokens, err := parseAliasTargetTokens(target)
	if err != nil {
		return err
	}
	target = strings.Join(targetTokens, " ")

	EnsureConfigAliases()
	cfg.Aliases[name] = target

	// Write alias into the scope-owned config object (not the merged Runtime map)
	// so that only the active scope's file is updated on disk.
	if err := setScopedAlias(name, target); err != nil {
		return err
	}

	if err := PersistConfigAliases(); err != nil {
		return err
	}

	logger.Info(fmt.Sprintf("Alias set: %s -> %s", name, target))
	fmt.Printf("✓ Alias '%s' -> '%s' set\n", name, target)
	return nil
}

func getAliasByName(rawName string) error {
	EnsureConfigAliases()

	name, err := NormalizeAliasName(rawName)
	if err != nil {
		return err
	}

	target, ok := cfg.Aliases[name]
	if !ok {
		return fmt.Errorf("alias '%s' not found", name)
	}

	fmt.Printf("%s=%s\n", name, target)
	return nil
}

func listAliases() error {
	EnsureConfigAliases()
	if len(cfg.Aliases) == 0 {
		fmt.Println("No aliases configured")
		return nil
	}

	keys := make([]string, 0, len(cfg.Aliases))
	for key := range cfg.Aliases {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	fmt.Println("Aliases:")
	for _, key := range keys {
		fmt.Printf("  %s=%s\n", key, cfg.Aliases[key])
	}
	return nil
}

func ExpandFirstCommandAlias(args []string) []string {
	if len(args) == 0 || cfg == nil {
		return args
	}

	EnsureConfigAliases()
	if len(cfg.Aliases) == 0 {
		return args
	}

	expanded := append([]string(nil), args...)

	for depth := 0; depth < maxAliasExpansionDepth; depth++ {
		firstToken := strings.TrimSpace(expanded[0])
		if firstToken == "" || strings.HasPrefix(firstToken, "-") {
			return expanded
		}

		aliasName, err := NormalizeAliasName(firstToken)
		if err != nil {
			return expanded
		}

		target, ok := cfg.Aliases[aliasName]
		if !ok {
			return expanded
		}

		targetTokens, err := parseAliasTargetTokens(target)
		if err != nil {
			return expanded
		}

		// Perform positional argument interpolation
		maxN := -1
		for i, token := range targetTokens {
			if !positionalVarRe.MatchString(token) {
				continue
			}

			targetTokens[i] = positionalVarRe.ReplaceAllStringFunc(token, func(match string) string {
				// extract the number
				submatches := positionalVarRe.FindStringSubmatch(match)
				if len(submatches) < 2 {
					return match
				}
				var idx int
				fmt.Sscanf(submatches[1], "%d", &idx)

				if idx > maxN {
					maxN = idx
				}

				if idx < len(expanded) {
					return expanded[idx]
				}
				return ""
			})
		}

		// Filter out empty tokens that might have been created
		var filteredTokens []string
		for _, t := range targetTokens {
			if t != "" {
				filteredTokens = append(filteredTokens, t)
			}
		}

		// If positional args were used, only append the remaining arguments
		appendStartIdx := 1
		if maxN >= 1 {
			appendStartIdx = maxN + 1
		}

		var remainingArgs []string
		if appendStartIdx < len(expanded) {
			remainingArgs = expanded[appendStartIdx:]
		}

		expanded = append(append([]string{}, filteredTokens...), remainingArgs...)
		if logger != nil {
			logger.Info(fmt.Sprintf("Alias expanded: %s -> %s", firstToken, strings.Join(filteredTokens, " ")))
		}
	}

	return expanded
}

var (
	cfg    *config.Config
	logger *utils.Logger
)

func Setup(c *config.Config, l *utils.Logger) {
	cfg = c
	logger = l
}

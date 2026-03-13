package cli

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/rishiyaduwanshi/boiler/internal/config"
	"github.com/spf13/cobra"
)

var aliasNameRe = regexp.MustCompile(`^[a-z][a-z0-9_-]*$`)

var aliasCmd = &cobra.Command{
	Use:   "alias [name|name=command]",
	Short: "Manage reusable command aliases",
	Long: `Manage command aliases stored in boiler.conf.json.

Usage patterns:
  bl alias              List all aliases
  bl alias name=cmd     Set or update an alias
  bl alias name         Get one alias value
  bl unalias name       Remove an alias (use 'unalias' command)

Alias names are normalized internally:
  - Names are case-insensitive
  - Hyphens and underscores are preserved

Examples:
  bl alias ll=ls
  bl alias s=search
  bl alias ll`,
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

func ensureConfigAliases() {
	if cfg.Aliases == nil {
		cfg.Aliases = make(map[string]string)
		return
	}

	normalized := make(map[string]string, len(cfg.Aliases))
	for rawName, rawTarget := range cfg.Aliases {
		name, err := normalizeAliasName(rawName)
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

func persistConfigAliases() error {
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	return nil
}

func normalizeAliasName(raw string) (string, error) {
	name := strings.ToLower(strings.TrimSpace(raw))
	if !aliasNameRe.MatchString(name) {
		return "", fmt.Errorf("invalid alias name %q: use letters, numbers, underscores, and hyphens", raw)
	}
	return name, nil
}

func normalizeAliasTarget(raw string) (string, error) {
	target := strings.ToLower(strings.TrimSpace(raw))
	if !aliasNameRe.MatchString(target) {
		return "", fmt.Errorf("invalid alias target %q: expected a command token", raw)
	}
	return target, nil
}

func setAliasFromAssignment(assignment string) error {
	rawName, rawTarget, found := strings.Cut(assignment, "=")
	if !found {
		return fmt.Errorf("invalid alias assignment; use NAME=COMMAND")
	}

	name, err := normalizeAliasName(rawName)
	if err != nil {
		return err
	}
	target, err := normalizeAliasTarget(rawTarget)
	if err != nil {
		return err
	}

	if isBuiltInCommandToken(name) {
		return fmt.Errorf("alias name %q conflicts with an existing command", name)
	}

	ensureConfigAliases()
	cfg.Aliases[name] = target

	if err := persistConfigAliases(); err != nil {
		return err
	}

	logger.Info(fmt.Sprintf("Alias set: %s -> %s", name, target))
	fmt.Printf("✓ Alias '%s' -> '%s' set\n", name, target)
	return nil
}

func getAliasByName(rawName string) error {
	ensureConfigAliases()

	name, err := normalizeAliasName(rawName)
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
	ensureConfigAliases()
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

func isBuiltInCommandToken(token string) bool {
	normalizedToken := strings.ToLower(strings.TrimSpace(token))
	if normalizedToken == "" {
		return false
	}

	for _, command := range rootCmd.Commands() {
		if strings.EqualFold(command.Name(), normalizedToken) {
			return true
		}
		for _, commandAlias := range command.Aliases {
			if strings.EqualFold(commandAlias, normalizedToken) {
				return true
			}
		}
	}

	return false
}

func expandFirstCommandAlias(args []string) []string {
	if len(args) == 0 || cfg == nil {
		return args
	}

	ensureConfigAliases()
	if len(cfg.Aliases) == 0 {
		return args
	}

	firstToken := strings.TrimSpace(args[0])
	if firstToken == "" || strings.HasPrefix(firstToken, "-") {
		return args
	}

	aliasName, err := normalizeAliasName(firstToken)
	if err != nil {
		return args
	}

	target, ok := cfg.Aliases[aliasName]
	if !ok {
		return args
	}

	expanded := append([]string(nil), args...)
	expanded[0] = target
	if logger != nil {
		logger.Info(fmt.Sprintf("Alias expanded: %s -> %s", firstToken, target))
	}
	return expanded
}

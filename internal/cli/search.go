package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/rishiyaduwanshi/boiler/internal/remote"
	"github.com/rishiyaduwanshi/boiler/internal/utils"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search for snippets or stacks",
	Long: `Search for resources in your store by name.

Searches both snippets and stacks by default. Use flags to filter:
  - Use -s or --snippets to search only snippets
  - Use -k or --stacks to search only stacks

Search is case-insensitive and matches partial names.

Remote Search:
  Use -r flag to search remote registry:
    - Default registry from config
		- Or specify custom: --registry https://github.com/other/boiler
		- Use config variable reference: --registry @team_reg`,
	Example: `  # Search for anything with 'error'
  bl search error

  # Search only snippets
  bl search logger --snippets

  # Search only stacks
  bl search express --stacks

  # Search remote registry
  bl search express -r

  # Search custom registry
	bl search express -r --registry https://github.com/myorg/boiler

	# Search registry from config variable
	bl search express -r --registry @team_reg`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		query, err := utils.ResolveInputToken(args[0], "query", cfg.Vars)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		logger.Info(fmt.Sprintf("Searching for: %s", query))

		if err := searchResources(query); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func searchResources(query string) error {
	// Handle remote search
	if searchRemote {
		return searchRemoteResources(query)
	}

	st, err := utils.LoadStore(cfg.Paths.Store)
	if err != nil {
		return err
	}

	query = strings.ToLower(query)
	foundAny := false

	// Search snippets
	if !searchStacks {
		snippets := st.ListSnippets()
		matches := []string{}
		for _, name := range snippets {
			if strings.Contains(strings.ToLower(name), query) {
				matches = append(matches, name)
			}
		}

		if len(matches) > 0 {
			foundAny = true
			fmt.Println("\n📄 Snippets:")
			for _, name := range matches {
				fmt.Printf("  • %s\n", name)
			}
		}
	}

	// Search stacks
	if !searchSnippets {
		stacks := st.ListStacks()
		matches := []string{}
		for _, name := range stacks {
			if strings.Contains(strings.ToLower(name), query) {
				matches = append(matches, name)
			}
		}

		if len(matches) > 0 {
			foundAny = true
			fmt.Println("\n📦 Stacks:")
			for _, name := range matches {
				fmt.Printf("  • %s\n", name)
			}
		}
	}

	if !foundAny {
		fmt.Printf("No results found for '%s'\n", query)
	} else {
		fmt.Println()
	}

	return nil
}

// searchRemoteResources searches for resources in remote registry
func searchRemoteResources(query string) error {
	// Use custom registry if provided, otherwise use config
	registryURL := cfg.Registry
	if searchRegistry != "" {
		registryURL = searchRegistry
	}

	registryURL, err := utils.ResolveInputToken(registryURL, "registry", cfg.Vars)
	if err != nil {
		return err
	}

	// Initialize remote store
	remoteStoreHandler, err := remote.NewRemoteStore(registryURL)
	if err != nil {
		return fmt.Errorf("failed to initialize remote store: %w", err)
	}

	// Load remote metadata
	fmt.Println("🔄 Fetching remote registry...")
	remoteStore, err := remoteStoreHandler.LoadFromURL()
	if err != nil {
		return fmt.Errorf("failed to load remote registry: %w", err)
	}

	// Determine what to search
	searchSnips := !searchStacks  // Search snippets by default
	searchStks := !searchSnippets // Search stacks by default

	results := remoteStoreHandler.Search(remoteStore, query, searchSnips, searchStks)

	foundAny := false

	// Display snippets
	if snippets, ok := results["snippets"]; ok && len(snippets) > 0 {
		foundAny = true
		fmt.Println("\n📄 Remote Snippets:")
		for _, name := range snippets {
			fmt.Printf("  • %s\n", name)
		}
	}

	// Display stacks
	if stacks, ok := results["stacks"]; ok && len(stacks) > 0 {
		foundAny = true
		fmt.Println("\n📦 Remote Stacks:")
		for _, name := range stacks {
			fmt.Printf("  • %s\n", name)
		}
	}

	if !foundAny {
		fmt.Printf("No remote results found for '%s'\n", query)
	} else {
		fmt.Println()
	}

	return nil
}

var (
	searchSnippets bool
	searchStacks   bool
	searchRemote   bool
	searchRegistry string
)

func init() {
	searchCmd.Flags().BoolVarP(&searchSnippets, "snippets", "n", false, "Search only snippets")
	searchCmd.Flags().BoolVarP(&searchStacks, "stacks", "k", false, "Search only stacks")
	searchCmd.Flags().BoolVarP(&searchRemote, "remote", "r", false, "Search remote registry")
	searchCmd.Flags().StringVar(&searchRegistry, "registry", "", "Custom registry URL (overrides config)")
}

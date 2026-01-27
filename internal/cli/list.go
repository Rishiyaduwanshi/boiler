package cli

import (
	"fmt"

	"github.com/rishiyaduwanshi/boiler/internal/store"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "ls",
	Aliases: []string{"list"},
	Short:   "List snippets or stacks",
	Run: func(cmd *cobra.Command, args []string) {
		st := store.NewStore(cfg.Paths.Store)
		if err := st.Load(); err != nil {
			fmt.Printf("Error loading store: %v\n", err)
			return
		}

		showAll := !listSnippets && !listStacks

		if listSnippets || showAll {
			fmt.Println("\n📄 Snippets:")
			snippets := st.ListSnippets()
			if len(snippets) == 0 {
				fmt.Println("  No snippets found")
			} else {
				for _, name := range snippets {
					fmt.Printf("  • %s\n", name)
				}
			}
		}

		if listStacks || showAll {
			fmt.Println("\n📦 Stacks:")
			stacks := st.ListStacks()
			if len(stacks) == 0 {
				fmt.Println("  No stacks found")
			} else {
				for _, name := range stacks {
					fmt.Printf("  • %s\n", name)
				}
			}
		}

		fmt.Println()
	},
}

var (
	listSnippets bool
	listStacks   bool
)

func init() {
	listCmd.Flags().BoolVarP(&listSnippets, "snippets", "n", false, "List snippets")
	listCmd.Flags().BoolVarP(&listStacks, "stacks", "k", false, "List stacks")
}

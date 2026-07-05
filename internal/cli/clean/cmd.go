package clean
import (
	"github.com/rishiyaduwanshi/boiler/internal/config"
	"fmt"
	"os"

	"github.com/rishiyaduwanshi/boiler/internal/store"
	"github.com/rishiyaduwanshi/boiler/internal/utils"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "clean [resource]",
	Short: "Clean snippets, stacks, or store",
	Long: `Remove snippets, stacks, or clear entire store.

You can:
  - Remove specific resource by name
  - Remove all snippets (use -s or --snippets flag)
  - Remove all stacks (use -k or --stacks flag)
  - Clear everything (use -a or --all flag)

Version-specific deletion is supported.`,
	Example: `  # Remove specific snippet
  bl clean errorHandler@1.js

  # Remove specific stack
  bl clean express-api@1

  # Remove all snippets
  bl clean --snippets

  # Remove all stacks
  bl clean --stacks

  # Clear entire store
  bl clean --all`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			resource := args[0]
			logger.Info(fmt.Sprintf("Cleaning resource: %s", resource))
			if err := cleanResource(resource); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return
		}

		if cleanAll {
			if err := cleanAllResources(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return
		}

		if cleanSnippets {
			if err := cleanAllSnippets(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return
		}

		if cleanStacks {
			if err := cleanAllStacks(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return
		}

		if err := interactiveClean(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

var (
	cleanAll      bool
	cleanSnippets bool
	cleanStacks   bool
)

const (
	cleanAllFlag           = "all"
	cleanAllFlagShort      = "a"
	cleanSnippetsFlag      = "snippets"
	cleanSnippetsFlagShort = "n"
	cleanStacksFlag        = "stacks"
	cleanStacksFlagShort   = "k"

	cleanAllDesc      = "Clean all resources"
	cleanSnippetsDesc = "Snippets only"
	cleanStacksDesc   = "Stacks only"
)

func cleanResource(resource string) error {
	st, err := utils.LoadStore(cfg.Paths.Store)
	if err != nil {
		return err
	}

	fullName := utils.ParseResourceName(resource)

	if store.IsSnippet(resource) {
		return cleanSnippet(st, fullName)
	}
	return cleanStack(st, fullName)
}

func cleanSnippet(st *store.Store, name string) error {
	path, ok := st.GetSnippet(name)
	if !ok {
		return fmt.Errorf(utils.ErrResourceNotFound, "snippet", name)
	}

	if !utils.ConfirmAction(fmt.Sprintf(utils.MsgPromptConfirmRemove, "snippet", name)) {
		fmt.Println(utils.MsgCancelled)
		return nil
	}

	if utils.FileExists(path) {
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("failed to remove file: %w", err)
		}
	}

	if err := st.RemoveSnippet(name); err != nil {
		return fmt.Errorf("failed to update metadata: %w", err)
	}

	fmt.Printf(utils.MsgSnippetRemoved, name)
	logger.Info(fmt.Sprintf("Snippet removed: %s", name))
	return nil
}

func cleanStack(st *store.Store, name string) error {
	path, ok := st.GetStack(name)
	if !ok {
		return fmt.Errorf(utils.ErrResourceNotFound, "stack", name)
	}

	if !utils.ConfirmAction(fmt.Sprintf(utils.MsgPromptConfirmRemove, "stack", name)) {
		fmt.Println(utils.MsgCancelled)
		return nil
	}

	if utils.IsDirectory(path) {
		if err := os.RemoveAll(path); err != nil {
			return fmt.Errorf("failed to remove directory: %w", err)
		}
	}

	if err := st.RemoveStack(name); err != nil {
		return fmt.Errorf("failed to update metadata: %w", err)
	}

	fmt.Printf(utils.MsgStackRemoved, name)
	logger.Info(fmt.Sprintf("Stack removed: %s", name))
	return nil
}

func cleanAllResources() error {
	st, err := utils.LoadStore(cfg.Paths.Store)
	if err != nil {
		return err
	}

	fmt.Println(utils.MsgPromptConfirmCleanAll)
	if !utils.ConfirmAction("Are you sure? (y/N): ") {
		fmt.Println(utils.MsgCancelled)
		return nil
	}

	snippets := st.ListSnippets()
	for _, name := range snippets {
		path, _ := st.GetSnippet(name)
		if utils.FileExists(path) {
			if err := os.Remove(path); err != nil {
				logger.Warn(fmt.Sprintf("failed to remove snippet file %s: %v", path, err))
			}
		}
		if err := st.RemoveSnippet(name); err != nil {
			logger.Warn(fmt.Sprintf("failed to remove snippet metadata %s: %v", name, err))
		}
	}

	stacks := st.ListStacks()
	for _, name := range stacks {
		path, _ := st.GetStack(name)
		if utils.IsDirectory(path) {
			if err := os.RemoveAll(path); err != nil {
				logger.Warn(fmt.Sprintf("failed to remove stack directory %s: %v", path, err))
			}
		}
		if err := st.RemoveStack(name); err != nil {
			logger.Warn(fmt.Sprintf("failed to remove stack metadata %s: %v", name, err))
		}
	}

	fmt.Printf("✓ Removed %d snippets and %d stacks\n", len(snippets), len(stacks))
	logger.Info("All resources cleaned")
	return nil
}

func cleanAllSnippets() error {
	st, err := utils.LoadStore(cfg.Paths.Store)
	if err != nil {
		return err
	}

	snippets := st.ListSnippets()
	if len(snippets) == 0 {
		fmt.Println(utils.MsgNoSnippets)
		return nil
	}

	if !utils.ConfirmAction(fmt.Sprintf("Remove %d snippets? (y/N): ", len(snippets))) {
		fmt.Println(utils.MsgCancelled)
		return nil
	}

	for _, name := range snippets {
		path, _ := st.GetSnippet(name)
		if utils.FileExists(path) {
			if err := os.Remove(path); err != nil {
				logger.Warn(fmt.Sprintf("failed to remove snippet file %s: %v", path, err))
			}
		}
		if err := st.RemoveSnippet(name); err != nil {
			logger.Warn(fmt.Sprintf("failed to remove snippet metadata %s: %v", name, err))
		}
	}

	fmt.Printf("✓ Removed %d snippets\n", len(snippets))
	logger.Info(fmt.Sprintf("Cleaned %d snippets", len(snippets)))
	return nil
}

func cleanAllStacks() error {
	st, err := utils.LoadStore(cfg.Paths.Store)
	if err != nil {
		return err
	}

	stacks := st.ListStacks()
	if len(stacks) == 0 {
		fmt.Println(utils.MsgNoStacks)
		return nil
	}

	if !utils.ConfirmAction(fmt.Sprintf("Remove %d stacks? (y/N): ", len(stacks))) {
		fmt.Println(utils.MsgCancelled)
		return nil
	}

	for _, name := range stacks {
		path, _ := st.GetStack(name)
		if utils.IsDirectory(path) {
			if err := os.RemoveAll(path); err != nil {
				logger.Warn(fmt.Sprintf("failed to remove stack directory %s: %v", path, err))
			}
		}
		if err := st.RemoveStack(name); err != nil {
			logger.Warn(fmt.Sprintf("failed to remove stack metadata %s: %v", name, err))
		}
	}

	fmt.Printf("✓ Removed %d stacks\n", len(stacks))
	logger.Info(fmt.Sprintf("Cleaned %d stacks", len(stacks)))
	return nil
}

func interactiveClean() error {
	fmt.Println("\nSelect action:")
	fmt.Println("  k - clean all stacks")
	fmt.Println("  n - clean all snippets")
	fmt.Println("  a - clean all resources")
	fmt.Println("  q - quit")

	choice, err := utils.Prompt("\nChoice: ")
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	switch choice {
	case "k", "K":
		return cleanAllStacks()
	case "n", "N":
		return cleanAllSnippets()
	case "a", "A":
		return cleanAllResources()
	case "q", "Q":
		fmt.Println(utils.MsgCancelled)
		return nil
	default:
		fmt.Fprintf(os.Stderr, "invalid choice %q — valid options: k (stacks), n (snippets), a (all), q (quit)\n", choice)
		return fmt.Errorf("invalid choice %q", choice)
	}
}

func init() {
Cmd.Flags().BoolVarP(&cleanAll, cleanAllFlag, cleanAllFlagShort, false, cleanAllDesc)
	Cmd.Flags().BoolVarP(&cleanSnippets, cleanSnippetsFlag, cleanSnippetsFlagShort, false, cleanSnippetsDesc)
	Cmd.Flags().BoolVarP(&cleanStacks, cleanStacksFlag, cleanStacksFlagShort, false, cleanStacksDesc)
}

var (
    cfg    *config.Config
    logger *utils.Logger
)

func Setup(c *config.Config, l *utils.Logger) {
    cfg = c
    logger = l
}

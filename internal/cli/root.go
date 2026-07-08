package cli

import (
	"os"
	"sync"

	addcmd "github.com/rishiyaduwanshi/boiler/internal/cli/add"
	aliascmd "github.com/rishiyaduwanshi/boiler/internal/cli/alias"
	cleancmd "github.com/rishiyaduwanshi/boiler/internal/cli/clean"
	confcmd "github.com/rishiyaduwanshi/boiler/internal/cli/conf"
	infocmd "github.com/rishiyaduwanshi/boiler/internal/cli/info"
	initcmd "github.com/rishiyaduwanshi/boiler/internal/cli/init"
	listcmd "github.com/rishiyaduwanshi/boiler/internal/cli/list"
	newcmd "github.com/rishiyaduwanshi/boiler/internal/cli/new"
	pathcmd "github.com/rishiyaduwanshi/boiler/internal/cli/path"
	searchcmd "github.com/rishiyaduwanshi/boiler/internal/cli/search"
	selfcmd "github.com/rishiyaduwanshi/boiler/internal/cli/self"
	storecmd "github.com/rishiyaduwanshi/boiler/internal/cli/store"
	unaliascmd "github.com/rishiyaduwanshi/boiler/internal/cli/unalias"
	unvarcmd "github.com/rishiyaduwanshi/boiler/internal/cli/unvar"
	usecmd "github.com/rishiyaduwanshi/boiler/internal/cli/use"
	varcmd "github.com/rishiyaduwanshi/boiler/internal/cli/var"
	versioncmd "github.com/rishiyaduwanshi/boiler/internal/cli/version"

	"github.com/rishiyaduwanshi/boiler/internal/config"
	"github.com/rishiyaduwanshi/boiler/internal/remote"
	"github.com/rishiyaduwanshi/boiler/internal/utils"
	"github.com/spf13/cobra"
)

var (
	cfg      *config.Config
	manager  *config.Manager
	logger   *utils.Logger
	verbose  bool
	isGlobal bool
	isLocal  bool
)

var normalizeCommandDocsOnce sync.Once

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "bl",
	Short: "Boiler - Code snippet and stack manager",
	Long: `Boiler - A CLI tool to manage reusable code snippets and project stacks.

Store, version, and reuse your code across projects. Perfect for:
  - Reusable utility functions (snippets)
  - Project templates and boilerplates (stacks)
  - Code patterns you use frequently
  - Multi-language development workflows

All resources are versioned automatically, making it easy to manage multiple
variations of the same snippet or stack.`,
	Example: `  # Initialize Boiler
  bl init

  # Store a snippet
  bl store ./utils/logger.js

  # Add snippet to project
  bl add logger

  # List all resources
  bl ls

  # Show paths
  bl path`,
	Run: func(cmd *cobra.Command, args []string) {
		// Show banner when no subcommand is provided
		utils.ShowBanner()
		utils.ShowQuickHelp()
	},
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if logger != nil {
			logger.SetVerbose(verbose)
		}
		remote.SetVerbose(verbose)

		// Create the BoilerContext by resolving the current Scope
		config.Ctx = &config.BoilerContext{
			Manager: manager,
			Scope:   config.ResolveScope(manager, isGlobal, isLocal),
		}
	},
}

// Execute runs the CLI
func Execute(m *config.Manager, log *utils.Logger) error {
	manager = m
	cfg = m.Runtime // Zero-refactor read-path: pass the Runtime config
	logger = log

	// Inject config and logger into all subcommands
	aliascmd.Setup(cfg, logger)
	cleancmd.Setup(cfg, logger)
	confcmd.Setup(cfg, logger)
	infocmd.Setup(cfg, logger)
	initcmd.Setup(cfg, logger)
	listcmd.Setup(cfg, logger)
	newcmd.Setup(cfg, logger)
	pathcmd.Setup(cfg, logger)
	searchcmd.Setup(cfg, logger)
	selfcmd.Setup(cfg, logger)
	storecmd.Setup(cfg, logger)
	unaliascmd.Setup(cfg, logger)
	unvarcmd.Setup(cfg, logger)
	usecmd.Setup(cfg, logger)
	varcmd.Setup(cfg, logger)
	versioncmd.Setup(cfg, logger)

	rootCmd.SilenceErrors = true
	rootCmd.SetArgs(aliascmd.ExpandFirstCommandAlias(os.Args[1:]))
	ensureCommandDocsNormalized()
	return rootCmd.Execute()
}

// GetRootCommand returns the root command for documentation generation
func GetRootCommand() *cobra.Command {
	ensureCommandDocsNormalized()
	return rootCmd
}

func ensureCommandDocsNormalized() {
	normalizeCommandDocsOnce.Do(func() {
		utils.NormalizeCommandTreeDocs(rootCmd)
	})
}

func init() {
	// Add version flag
	rootCmd.Flags().BoolP("version", "v", false, "Show version information")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "V", false, "Enable verbose debug output")

	// Add Scope flags
	rootCmd.PersistentFlags().BoolVar(&isGlobal, "global", false, "Force global scope for this command")
	rootCmd.PersistentFlags().BoolVar(&isLocal, "local", false, "Force local scope for this command")

	// Add subcommands
	rootCmd.AddCommand(versioncmd.Cmd)
	rootCmd.AddCommand(confcmd.Cmd)
	rootCmd.AddCommand(aliascmd.Cmd)
	rootCmd.AddCommand(unaliascmd.Cmd)
	rootCmd.AddCommand(varcmd.Cmd)
	rootCmd.AddCommand(unvarcmd.Cmd)
	rootCmd.AddCommand(addcmd.NewCmd(func() *config.Config { return cfg }, func() *utils.Logger { return logger }))
	rootCmd.AddCommand(storecmd.Cmd)
	rootCmd.AddCommand(listcmd.Cmd)
	rootCmd.AddCommand(cleancmd.Cmd)
	rootCmd.AddCommand(infocmd.Cmd)
	rootCmd.AddCommand(searchcmd.Cmd)
	rootCmd.AddCommand(initcmd.Cmd)
	rootCmd.AddCommand(newcmd.Cmd)
	rootCmd.AddCommand(pathcmd.Cmd)
	rootCmd.AddCommand(selfcmd.Cmd)
	rootCmd.AddCommand(usecmd.Cmd)
}

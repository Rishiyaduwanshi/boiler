package use
import (
	"github.com/rishiyaduwanshi/boiler/internal/config"
	"fmt"
	"os"

	addcmd "github.com/rishiyaduwanshi/boiler/internal/cli/add"
	"github.com/rishiyaduwanshi/boiler/internal/utils"
	"github.com/spf13/cobra"

)

var Cmd = &cobra.Command{
	Use:   "use [resource] [destination]",
	Short: "Fetch a remote resource directly without saving to local store",
	Long: `Fetch a remote resource directly without saving to local store.

The resource is fetched from remote source and copied directly to destination
without writing into local store metadata.

Stack placement:
- By default, stacks are copied inside a stack-named folder.
- Use --spread to copy stack contents directly into destination.`,
	Example: `  # GitHub repo as stack
  bl use alice/my-express-stack

  # GitLab repo
  bl use https://gitlab.com/alice/my-stack

  # Bitbucket repo
  bl use https://bitbucket.org/alice/my-stack

  # File from GitHub repo (snippet)
  bl use alice/snippets:js/errorHandler.js

  # Direct zip archive (any site)
  bl use https://mysite.com/templates/express.zip

  # Direct tar.gz archive
  bl use https://mysite.com/stack.tar.gz

  # Direct file URL (snippet)
  bl use https://mysite.com/snippets/logger.js

	# Resource from config variable
	bl use :starter_stack

	# Into a specific folder
	bl use alice/my-stack ./new-project

	# Spread stack contents directly into destination
	bl use alice/my-stack ./new-project --spread

	# Force overwrite destination conflicts
	bl use alice/my-stack ./new-project --force`,
	Args: cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		resource, err := utils.ResolveInputToken(args[0], "resource", cfg.Vars)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		positionalDest := ""
		if len(args) == 2 {
			positionalDest, err = utils.ResolveInputToken(args[1], "destination", cfg.Vars)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		}

		opts := addcmd.Options{
			Force:  useForce,
			Spread: useSpread,
		}

		if err := addcmd.AddResource(resource, addcmd.ResolveDestination(positionalDest), true, true, opts, cfg, logger); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

var useForce bool
var useSpread bool

func init() {
	Cmd.Flags().BoolVarP(&useForce, "force", "f", false, "Force operation without confirmation")
	Cmd.Flags().BoolVar(&useSpread, "spread", false, "Spread stack contents directly into destination")
}

var (
    cfg    *config.Config
    logger *utils.Logger
)

func Setup(c *config.Config, l *utils.Logger) {
    cfg = c
    logger = l
}

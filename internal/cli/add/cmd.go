package add

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rishiyaduwanshi/boiler/internal/config"
	"github.com/rishiyaduwanshi/boiler/internal/utils"
	"github.com/spf13/cobra"
)

var (
	addRemote    bool
	addSpread    bool
	addForce     bool
	addRegistry  string
	addName      string
	addAsStack   bool
	addAsSnippet bool
	addNoStore   bool
)

const DefaultDestination = "boiler"

func ResolveDestination(positionalDest string) string {
	if positionalDest == "" {
		return DefaultDestination
	}
	return filepath.Clean(positionalDest)
}

func NewCmd(getCfg func() *config.Config, getLogger func() *utils.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add [resource] [destination]",
		Short: "Add a snippet or stack to boiler/ by default",
		Long: `Add a stored snippet or stack to ./boiler by default.

Destination:
- Use optional [destination] to override the default path.
- bl add logger .
- bl add logger ./src/utils
- bl add logger /absolute/path

Stack placement:
- By default, stacks are copied inside a stack-named folder.
- bl add express@1 -> ./boiler/express
- Use --spread to copy stack contents directly into destination.
- bl add express@1 --spread -> contents in ./boiler

The command copies resources from your store. For snippets with a single version,
you can use just the name (e.g., 'errorHandler' will auto-select version 1).
For multiple versions, you'll be prompted to choose.

Template Variables:
- Snippets can contain template variables using the format: bl__VAR_NAME.
- When adding a snippet with variables, you'll be prompted to provide values.
- If matching config vars exist, they are used as defaults.
- Default values are shown in brackets (from __var declarations).
- Press Enter to use default or type a custom value.
- Variables are replaced and metadata comments are removed in the final file.

Command Variables:
- Use :name to resolve values from config vars (set via 'bl var').
- Example: bl add express@1 -r --registry :team_reg

Stacks are also versioned and can be added by name or with explicit version.

Remote Resources:
- Use -r flag to fetch from remote source and save to local store.
- Provider is auto-detected from the URL (GitHub, GitLab, Bitbucket, generic).
- Resource is cached locally; subsequent uses do not need -r.
- Use --no-store for one-shot fetch without saving to local store.
- Use --stack/-k or --snippet/-n to override stack/snippet auto-detection.
- For ambiguous remote inputs, stack detection is preferred by default.

Supported remote formats:
- Registry: bl add express@1 -r (registry set via: bl conf --set-registry <url>)
- GitHub short: bl add owner/repo -r
- GitHub full URL: bl add https://github.com/owner/repo -r
- GitLab: bl add https://gitlab.com/owner/repo -r
- Bitbucket: bl add https://bitbucket.org/owner/repo -r
- File from repo: bl add owner/repo:path/to/file.js -r
- Direct file URL: bl add https://site.com/file.js -r
- Direct archive: bl add https://site.com/stack.zip -r
- Custom domain file: bl add site.com:path/file.js -r
- One-time registry: bl add express@1 -r --registry https://github.com/other/boiler`,
		Example: `  # Add snippet (auto-detects if single version)
  bl add errorHandler

	# Add to current directory
	bl add errorHandler .

	# Add to a custom destination
	bl add errorHandler ./src/utils

  # Add snippet with template variables
  bl add apiClient
  # Prompts: bl__API_URL [http://localhost:3000]: https://api.example.com
  #          bl__API_KEY [your-key]: abc123xyz
  # Output: Clean file with variables replaced, no metadata comments

  # Add specific version
  bl add logger@2.js

	# Add stack into boiler/express-api
  bl add express-api@1

	# Add stack contents directly into destination
	bl add express-api@1 --spread

  # Force overwrite
  bl add middleware --force

  # Remote: from configured registry
  bl add express@1 -r

  # Remote: GitHub short format
  bl add rishiyaduwanshi/boiler-express -r

  # Remote: GitLab
  bl add https://gitlab.com/alice/my-stack -r

  # Remote: Bitbucket
  bl add https://bitbucket.org/alice/my-stack -r

  # Remote: file inside GitHub repo
  bl add rishiyaduwanshi/boiler-snippets:js/errorHandler.js -r

  # Remote: direct file URL
  bl add https://mysite.com/snippets/helper.js -r

  # Remote: direct archive URL
  bl add https://mysite.com/stack.zip -r

	# Remote: force snippet mode when auto-detection is unclear
	bl add owner/repo:path/to/template -r --snippet

	# Remote: force stack mode
	bl add https://example.com/custom-source -r --stack

  # Remote: one-time registry override
  bl add express@1 -r --registry https://github.com/myorg/boiler

	# Remote: registry from config variable
	bl add express@1 -r --registry :team_reg

	# One-shot fetch without saving to store
	bl add alice/my-stack -r --no-store`,
		Args: cobra.RangeArgs(1, 2),
		Run: func(cmd *cobra.Command, args []string) {
			cfg := getCfg()
			logger := getLogger()

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

			destPath := ResolveDestination(positionalDest)
			logger.Info(fmt.Sprintf("Adding resource: %s -> %s", resource, destPath))

			if addNoStore && !addRemote {
				fmt.Fprintf(os.Stderr, "Error: --no-store requires --remote\n")
				os.Exit(1)
			}

			opts := Options{
				Force:     addForce,
				Spread:    addSpread,
				AsStack:   addAsStack,
				AsSnippet: addAsSnippet,
				Registry:  addRegistry,
				Name:      addName,
			}

			if err := AddResource(resource, destPath, addRemote, addNoStore, opts, cfg, logger); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		},
	}

	cmd.Flags().BoolVarP(&addRemote, "remote", "r", false, "Fetch from remote registry")
	cmd.Flags().BoolVar(&addNoStore, "no-store", false, "Fetch remote resource without saving to local store")
	cmd.Flags().BoolVar(&addSpread, "spread", false, "Spread stack contents directly into destination")
	cmd.Flags().BoolVarP(&addForce, "force", "f", false, "Force operation without confirmation")
	cmd.Flags().BoolVarP(&addAsStack, "stack", "k", false, "Treat resource as stack (overrides auto-detection)")
	cmd.Flags().BoolVarP(&addAsSnippet, "snippet", "n", false, "Treat resource as snippet (overrides auto-detection)")
	cmd.Flags().StringVar(&addRegistry, "registry", "", "Custom registry URL (overrides config)")
	cmd.Flags().StringVarP(&addName, "name", "m", "", "Rename snippet in destination")

	return cmd
}

func resolveAddResourceType(opts Options) (ResourceType, error) {
	if opts.AsStack && opts.AsSnippet {
		return ResourceTypeAuto, fmt.Errorf("--stack and --snippet cannot be used together")
	}

	if opts.AsStack {
		return ResourceTypeStack, nil
	}

	if opts.AsSnippet {
		return ResourceTypeSnippet, nil
	}

	return ResourceTypeAuto, nil
}

func AddResource(resource, destPath string, remoteEnabled, noStore bool, opts Options, cfg *config.Config, logger *utils.Logger) error {
	resourceType, err := resolveAddResourceType(opts)
	if err != nil {
		return err
	}

	if remoteEnabled {
		return ResourceFromRemote(resource, destPath, resourceType, noStore, opts, cfg, logger)
	}

	st, err := utils.LoadStore(cfg.Paths.Store)
	if err != nil {
		return err
	}

	resolvedName, isSnippet, err := ResolveStoreResource(st, resource, resourceType, false)
	if err != nil {
		return err
	}

	if isSnippet {
		if opts.Spread {
			return fmt.Errorf("--spread is only supported for stacks")
		}
		return AddSnippet(st, resolvedName, destPath, opts, cfg, logger)
	}

	return AddStack(st, resolvedName, destPath, opts, logger)
}

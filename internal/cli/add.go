package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add [resource]",
	Short: "Add a snippet or stack to current directory",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		resource := args[0]
		logger.Info(fmt.Sprintf("Adding resource: %s", resource))
		fmt.Printf("Adding %s (implementation in progress)\n", resource)
	},
}

var (
	addRemote bool
	addTo     string
	addGlobal bool
	addBoth   bool
	addForce  bool
)

func init() {
	addCmd.Flags().BoolVarP(&addRemote, "remote", "r", false, "Fetch from remote registry")
	addCmd.Flags().StringVarP(&addTo, "to", "t", ".", "Destination path")
	addCmd.Flags().BoolVarP(&addGlobal, "global", "g", false, "Add to global store")
	addCmd.Flags().BoolVarP(&addBoth, "both", "b", false, "Add to both local and global")
	addCmd.Flags().BoolVarP(&addForce, "force", "f", false, "Force overwrite")
}

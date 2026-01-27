package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize boiler in current directory",
	Run: func(cmd *cobra.Command, args []string) {
		logger.Info("Initializing boiler")
		fmt.Println("Initializing boiler (implementation in progress)")
	},
}

var (
	initGlobal bool
)

func init() {
	initCmd.Flags().BoolVarP(&initGlobal, "global", "g", false, "Initialize global configuration")
}

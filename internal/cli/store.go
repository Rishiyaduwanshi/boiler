package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var storeCmd = &cobra.Command{
	Use:   "store [path]",
	Short: "Store a folder/file as snippet or stack",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}
		
		logger.Info(fmt.Sprintf("Storing: %s", path))
		fmt.Printf("Storing %s (implementation in progress)\n", path)
	},
}

var (
	storeName        string
	storeType        string
	storeDescription string
)

func init() {
	storeCmd.Flags().StringVarP(&storeName, "name", "n", "", "Name for the resource")
	storeCmd.Flags().StringVarP(&storeType, "type", "t", "", "Type: snippet or stack")
	storeCmd.Flags().StringVarP(&storeDescription, "description", "d", "", "Description")
}

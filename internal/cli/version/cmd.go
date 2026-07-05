package version
import (
	"github.com/rishiyaduwanshi/boiler/internal/config"
	"github.com/rishiyaduwanshi/boiler/internal/utils"
	"fmt"

	"github.com/rishiyaduwanshi/boiler/pkg/version"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:     "version",
	Aliases: []string{"v"},
	Short:   "Show version information",
	Long: `Display Boiler version information.

Shows current version, build date, and Go version used.`,
	Example: `  # Show version
  bl version

  # Short form
  bl v`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version.Info())
	},
}

var (
    cfg    *config.Config
    logger *utils.Logger
)

func Setup(c *config.Config, l *utils.Logger) {
    cfg = c
    logger = l
}

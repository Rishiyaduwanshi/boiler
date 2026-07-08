package engine

import (
	"fmt"
	"strings"
)

// ExecuteCommand routes parsed Boiler commands to their respective actions.
// Supports: use, run, inject
func ExecuteCommand(commandLine string) error {
	commandLine = strings.TrimSpace(commandLine)

	if strings.HasPrefix(commandLine, "inject ") {
		return InjectCode(commandLine)
	} else if strings.HasPrefix(commandLine, "use ") {
		// TODO: Implement use command (copy files from repo)
		fmt.Printf("[ENGINE EXEC] %s\n", commandLine)
		return nil
	} else if strings.HasPrefix(commandLine, "run ") {
		// TODO: Implement run command (execute shell scripts or sub-boiler commands)
		fmt.Printf("[ENGINE EXEC] %s\n", commandLine)
		return nil
	}

	return fmt.Errorf("unknown command: %s", commandLine)
}

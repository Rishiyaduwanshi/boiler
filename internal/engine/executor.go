package engine

import (
	"fmt"
)

// ExecuteCommand routes parsed Boiler commands to their respective actions.
// Supports: use, run, inject
func ExecuteCommand(commandLine string) error {
	// Stub implementation for now until we build the actual executor and injector
	fmt.Printf("[ENGINE EXEC] %s\n", commandLine)
	return nil
}

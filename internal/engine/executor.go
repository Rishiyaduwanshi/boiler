package engine

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// ExecuteCommand routes parsed Boiler commands to their respective actions.
// Supports: use, run, inject
func ExecuteCommand(commandLine string) error {
	commandLine = strings.TrimSpace(commandLine)

	if strings.HasPrefix(commandLine, "inject ") {
		return InjectCode(commandLine)
	} else if strings.HasPrefix(commandLine, "run ") {
		// Run an arbitrary OS command (e.g. npm install)
		cmdStr := strings.TrimPrefix(commandLine, "run ")
		cmdStr = strings.TrimSpace(cmdStr)
		fmt.Printf("[ENGINE] Running OS command: %s\n", cmdStr)
		return executeOSCommand(cmdStr)
	} else {
		// Assume it's a native Boiler command (use, add, var, alias, etc.)
		fmt.Printf("[ENGINE] Running Boiler command: bl %s\n", commandLine)
		return executeOSCommand("bl " + commandLine)
	}
}

func executeOSCommand(cmdStr string) error {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", cmdStr)
	} else {
		cmd = exec.Command("sh", "-c", cmdStr)
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

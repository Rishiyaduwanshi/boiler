//go:build ignore

package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func main() {
	task := "build"
	if len(os.Args) > 1 {
		task = os.Args[1]
	}

	goos := os.Getenv("GOOS")
	if goos == "" {
		goos = runtime.GOOS
	}

	ext := ""
	if goos == "windows" {
		ext = ".exe"
	}
	outputFile := "bl" + ext

	var args []string

	switch task {
	case "build":
		args = []string{"build", "-trimpath", "-ldflags=-s -w", "-o", outputFile, "."}
	case "dev":
		args = []string{"build", "-o", outputFile, "."}
	case "release":
		version := getVersion()
		ldflags := fmt.Sprintf("-s -w -X github.com/rishiyaduwanshi/boiler/pkg/version.Version=%s", version)
		args = []string{"build", "-trimpath", "-ldflags=" + ldflags, "-o", outputFile, "."}
	default:
		log.Fatalf("Unknown build task: %s. Supported: build, dev, release", task)
	}

	cmd := exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	cmd.Env = os.Environ()

	fmt.Printf("Building %s (%s)...\n", outputFile, task)
	if err := cmd.Run(); err != nil {
		log.Fatalf("Build failed: %v", err)
	}
	fmt.Printf("Successfully built %s\n", outputFile)
}

func getVersion() string {
	cmd := exec.Command("git", "describe", "--tags", "--always", "--dirty")
	out, err := cmd.Output()
	if err != nil {
		return "dev"
	}
	return strings.TrimSpace(string(out))
}

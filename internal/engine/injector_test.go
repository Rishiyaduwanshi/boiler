package engine

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInjectCode(t *testing.T) {
	tempDir := t.TempDir()
	targetFile := filepath.Join(tempDir, "routes.js")

	initialContent := `import express from "express";
const router = express.Router();

// bl__DETECTOR_START_routes
// bl__DETECTOR_END_routes

export default router;
`
	err := os.WriteFile(targetFile, []byte(initialContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write target file: %v", err)
	}

	// Test 1: Append
	appendCmd := "inject " + targetFile + " -d routes -a -c `\nrouter.use('/append', appendRouter);\n`"
	err = InjectCode(appendCmd)
	if err != nil {
		t.Fatalf("InjectCode(append) failed: %v", err)
	}

	// Test 2: Prepend
	prependCmd := "inject " + targetFile + " --detect routes --prepend --content `\nrouter.use('/prepend', prependRouter);\n`"
	err = InjectCode(prependCmd)
	if err != nil {
		t.Fatalf("InjectCode(prepend) failed: %v", err)
	}

	// Read the file and verify
	finalContentBytes, _ := os.ReadFile(targetFile)
	finalContent := string(finalContentBytes)

	expectedPrepend := "router.use('/prepend', prependRouter);"
	expectedAppend := "router.use('/append', appendRouter);"

	// Check if prepend is immediately after START marker
	// Check if append is immediately before END marker
	// Actually we just check if both strings are in the file in the correct order

	if !strings.Contains(finalContent, expectedPrepend) {
		t.Errorf("Prepend content not found in file")
	}
	if !strings.Contains(finalContent, expectedAppend) {
		t.Errorf("Append content not found in file")
	}

	// Prepend should come before Append in the file
	prependIdx := strings.Index(finalContent, expectedPrepend)
	appendIdx := strings.Index(finalContent, expectedAppend)

	if prependIdx > appendIdx {
		t.Errorf("Prepend content is located AFTER append content, which is incorrect")
	}
}

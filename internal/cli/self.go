package cli

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/rishiyaduwanshi/boiler/pkg/version"
	"github.com/spf13/cobra"
)

const githubReleaseAPI = "https://api.github.com/repos/rishiyaduwanshi/boiler/releases/latest"

type githubRelease struct {
	TagName string        `json:"tag_name"`
	Assets  []githubAsset `json:"assets"`
}

type githubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

var selfCmd = &cobra.Command{
	Use:   "self",
	Short: "Manage Boiler installation",
	Long: `Manage Boiler CLI installation.

Commands for updating and uninstalling Boiler itself.`,
}

var selfUninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall Boiler CLI",
	Long: `Uninstall Boiler CLI from your system.

This will:
  - Remove the binary from installation directory
  - Clean PATH environment variable
  - Optionally remove config and store data

You will be prompted for confirmation before deletion.`,
	Example: `  # Uninstall Boiler
  bl self uninstall`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runSelfUninstall(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

var selfUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update Boiler to latest version",
	Long: `Update Boiler CLI to the latest version.

Downloads the latest release from GitHub, verifies its SHA256 checksum,
and replaces the current binary. No scripts are piped to shell.`,
	Example: `  # Update to latest version
  bl self update`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runSelfUpdate(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func runSelfUninstall() error {
	var cmd *exec.Cmd
	scriptURL := "https://raw.githubusercontent.com/rishiyaduwanshi/boiler/main/scripts/uninstall"

	if runtime.GOOS == "windows" {
		scriptURL += ".ps1"
		psCmd := fmt.Sprintf("irm %s | iex", scriptURL)
		cmd = exec.Command("powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", psCmd)
	} else {
		scriptURL += ".sh"
		bashCmd := fmt.Sprintf("curl -fsSL %s | bash", scriptURL)
		cmd = exec.Command("bash", "-c", bashCmd)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

func runSelfUpdate() error {
	fmt.Println("Checking for updates...")

	rel, err := fetchLatestRelease()
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	// Normalize tag: "v0.1.0" -> "0.1.0"
	latestTag := strings.TrimPrefix(rel.TagName, "v")
	currentVer := strings.TrimPrefix(version.Version, "v")

	if latestTag == currentVer {
		fmt.Printf("Already on latest version (%s)\n", version.Version)
		return nil
	}

	fmt.Printf("Updating %s → %s\n", version.Version, rel.TagName)

	// Determine asset name for current platform
	target := releaseAssetName()

	// Find asset and checksum URLs from release
	var assetURL, checksumURL string
	for _, a := range rel.Assets {
		switch a.Name {
		case target:
			assetURL = a.BrowserDownloadURL
		case "checksums.txt":
			checksumURL = a.BrowserDownloadURL
		}
	}
	if assetURL == "" {
		return fmt.Errorf("no release asset found for %s/%s (looking for: %s)", runtime.GOOS, runtime.GOARCH, target)
	}

	// Download checksums.txt and get expected hash for our asset
	var expectedHash string
	if checksumURL != "" {
		checksumData, err := selfHTTPGet(checksumURL)
		if err == nil {
			expectedHash = parseChecksumFile(string(checksumData), target)
		}
	}

	// Download archive to a temp file
	tmpFile, err := os.CreateTemp("", "boiler-update-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpArchive := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpArchive)

	fmt.Printf("Downloading %s...\n", target)
	if err := selfDownloadToFile(assetURL, tmpArchive); err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	// Verify checksum
	if expectedHash != "" {
		actual, err := sha256Sum(tmpArchive)
		if err != nil {
			return fmt.Errorf("failed to compute checksum: %w", err)
		}
		if actual != expectedHash {
			return fmt.Errorf("checksum mismatch (expected %s, got %s) - aborting update", expectedHash, actual)
		}
		fmt.Println("✓ Checksum verified")
	} else {
		fmt.Println("⚠ Warning: checksums.txt not found, proceeding without verification")
	}

	// Find current executable path
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to locate current binary: %w", err)
	}
	// Evaluate symlinks so we operate on the real file path
	exePath, err = filepath.EvalSymlinks(exePath)
	if err != nil {
		return fmt.Errorf("failed to resolve binary path: %w", err)
	}

	// Extract binary and atomically replace
	if err := extractAndReplaceBinary(tmpArchive, exePath); err != nil {
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	fmt.Printf("✓ Updated to %s - restart your terminal if needed\n", rel.TagName)
	return nil
}

// releaseAssetName returns the GoReleaser archive name for the current platform.
// Matches the name_template in .goreleaser.yaml.
func releaseAssetName() string {
	var osName string
	switch runtime.GOOS {
	case "linux":
		osName = "Linux"
	case "darwin":
		osName = "Darwin"
	case "windows":
		osName = "Windows"
	default:
		osName = runtime.GOOS
	}

	arch := runtime.GOARCH
	if arch == "amd64" {
		arch = "x86_64"
	} else if arch == "386" {
		arch = "i386"
	}

	ext := ".tar.gz"
	if runtime.GOOS == "windows" {
		ext = ".zip"
	}

	return fmt.Sprintf("boiler_%s_%s%s", osName, arch, ext)
}

// fetchLatestRelease queries the GitHub API for the latest release.
func fetchLatestRelease() (*githubRelease, error) {
	data, err := selfHTTPGet(githubReleaseAPI)
	if err != nil {
		return nil, err
	}
	var rel githubRelease
	if err := json.Unmarshal(data, &rel); err != nil {
		return nil, fmt.Errorf("failed to parse release info: %w", err)
	}
	return &rel, nil
}

// parseChecksumFile extracts the SHA256 hash for filename from a checksums.txt.
// Format: "<hash>  <filename>" per line.
func parseChecksumFile(content, filename string) string {
	for _, line := range strings.Split(content, "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && fields[1] == filename {
			return fields[0]
		}
	}
	return ""
}

// sha256Sum computes the SHA256 hex digest of a file.
func sha256Sum(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// extractAndReplaceBinary extracts the `bl` binary from the downloaded archive
// and atomically replaces the current executable.
func extractAndReplaceBinary(archivePath, exePath string) error {
	binaryName := "bl"
	if runtime.GOOS == "windows" {
		binaryName = "bl.exe"
	}

	// Extract to a temp file beside the current binary (same filesystem for atomic rename)
	tmpBin, err := os.CreateTemp(filepath.Dir(exePath), "boiler-new-*")
	if err != nil {
		return fmt.Errorf("failed to create temp binary: %w", err)
	}
	tmpBinPath := tmpBin.Name()
	tmpBin.Close()
	defer os.Remove(tmpBinPath)

	if runtime.GOOS == "windows" {
		if err := extractFromZip(archivePath, binaryName, tmpBinPath); err != nil {
			return err
		}
	} else {
		if err := extractFromTarGz(archivePath, binaryName, tmpBinPath); err != nil {
			return err
		}
	}

	if err := os.Chmod(tmpBinPath, 0755); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	if runtime.GOOS == "windows" {
		// Windows does not allow overwriting a running executable.
		// Rename the current binary to .old (allowed), then rename new into place.
		oldPath := exePath + ".old"
		os.Remove(oldPath) // remove stale .old if present
		if err := os.Rename(exePath, oldPath); err != nil {
			return fmt.Errorf("failed to move current binary: %w", err)
		}
		if err := os.Rename(tmpBinPath, exePath); err != nil {
			// Restore old binary so the tool still works
			_ = os.Rename(oldPath, exePath)
			return fmt.Errorf("failed to place new binary: %w", err)
		}
		// .old is left on disk; cleaned up on next update or startup
		return nil
	}

	// Unix: atomic rename (same filesystem guaranteed by temp dir)
	if err := os.Rename(tmpBinPath, exePath); err != nil {
		return fmt.Errorf("failed to replace binary: %w", err)
	}
	return nil
}

func extractFromTarGz(archivePath, binaryName, destPath string) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if filepath.Base(hdr.Name) == binaryName && hdr.Typeflag == tar.TypeReg {
			out, err := os.Create(destPath)
			if err != nil {
				return err
			}
			defer out.Close()
			_, err = io.Copy(out, tr)
			return err
		}
	}
	return fmt.Errorf("binary '%s' not found in archive", binaryName)
}

func extractFromZip(archivePath, binaryName, destPath string) error {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		if filepath.Base(f.Name) == binaryName {
			rc, err := f.Open()
			if err != nil {
				return err
			}
			defer rc.Close()
			out, err := os.Create(destPath)
			if err != nil {
				return err
			}
			defer out.Close()
			_, err = io.Copy(out, rc)
			return err
		}
	}
	return fmt.Errorf("binary '%s' not found in archive", binaryName)
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

func selfHTTPGet(url string) ([]byte, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}
	return io.ReadAll(resp.Body)
}

func selfDownloadToFile(url, path string) error {
	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}
	out, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	return err
}

func init() {
	selfCmd.AddCommand(selfUninstallCmd)
	selfCmd.AddCommand(selfUpdateCmd)
}

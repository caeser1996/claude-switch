package claude

import (
	"os/exec"
	"strings"
)

// BinaryName is the expected name of the Claude CLI binary.
const BinaryName = "claude"

// DetectResult holds information about the Claude CLI installation.
type DetectResult struct {
	Found   bool
	Path    string
	Version string
	Error   error
}

// Detect searches for the Claude CLI binary on PATH and retrieves its version.
func Detect() DetectResult {
	path, err := exec.LookPath(BinaryName)
	if err != nil {
		return DetectResult{
			Found: false,
			Error: err,
		}
	}

	version := getVersion(path)

	return DetectResult{
		Found:   true,
		Path:    path,
		Version: version,
	}
}

// getVersion runs `claude --version` and returns the version string.
func getVersion(binaryPath string) string {
	cmd := exec.Command(binaryPath, "--version")
	out, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}

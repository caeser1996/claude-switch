package claude

import (
	"fmt"
	"os"
	"os/exec"
)

// RunOptions configures how claude is executed.
type RunOptions struct {
	// Args are the arguments to pass to the claude binary.
	Args []string

	// Env overrides the process environment. If nil, inherits current env.
	Env []string

	// Dir sets the working directory. If empty, uses current dir.
	Dir string
}

// Run executes the claude binary with the given options.
// It connects stdin/stdout/stderr so the user can interact with claude.
func Run(opts RunOptions) error {
	result := Detect()
	if !result.Found {
		return fmt.Errorf("claude binary not found on PATH — install Claude Code first")
	}

	cmd := exec.Command(result.Path, opts.Args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if opts.Env != nil {
		cmd.Env = opts.Env
	}
	if opts.Dir != "" {
		cmd.Dir = opts.Dir
	}

	return cmd.Run()
}

// RunWithOutput executes the claude binary and captures stdout.
func RunWithOutput(args ...string) (string, error) {
	result := Detect()
	if !result.Found {
		return "", fmt.Errorf("claude binary not found on PATH — install Claude Code first")
	}

	cmd := exec.Command(result.Path, args...)
	cmd.Stderr = os.Stderr

	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("claude command failed: %w", err)
	}

	return string(out), nil
}

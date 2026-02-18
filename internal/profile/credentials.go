package profile

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// CredentialFiles lists the files we track from the Claude config directory.
var CredentialFiles = []string{
	".credentials.json",
	"statsig",
	"statsig_metadata",
}

// HomeCredentialFiles lists files we track from the home directory.
var HomeCredentialFiles = []string{
	".claude.json",
}

// CopyFile copies a single file preserving permissions.
func CopyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("cannot open %s: %w", src, err)
	}
	defer srcFile.Close()

	info, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("cannot stat %s: %w", src, err)
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0700); err != nil {
		return fmt.Errorf("cannot create directory for %s: %w", dst, err)
	}

	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
	if err != nil {
		return fmt.Errorf("cannot create %s: %w", dst, err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("cannot copy to %s: %w", dst, err)
	}

	// Ensure secure permissions for credential files
	return os.Chmod(dst, 0600)
}

// FileExists checks if a file exists and is not a directory.
func FileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// DirExists checks if a directory exists.
func DirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

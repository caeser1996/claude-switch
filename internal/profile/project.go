package profile

import (
	"os"
	"path/filepath"
	"strings"
)

const ProjectProfileFile = ".claude-profile"

// DetectProjectProfile walks up from the current directory looking for
// a .claude-profile file. Returns the profile name if found, or empty string.
func DetectProjectProfile() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	return detectProjectProfileFrom(dir)
}

// detectProjectProfileFrom walks up from the given directory.
func detectProjectProfileFrom(startDir string) string {
	dir := startDir
	for {
		candidate := filepath.Join(dir, ProjectProfileFile)
		if FileExists(candidate) {
			data, err := os.ReadFile(candidate)
			if err == nil {
				name := strings.TrimSpace(string(data))
				if name != "" {
					return name
				}
			}
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break // reached root
		}
		dir = parent
	}
	return ""
}

// WriteProjectProfile creates a .claude-profile file in the given directory.
func WriteProjectProfile(dir, profileName string) error {
	path := filepath.Join(dir, ProjectProfileFile)
	return os.WriteFile(path, []byte(profileName+"\n"), 0644)
}

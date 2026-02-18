package doctor

import (
	"fmt"
	"os"
	"runtime"

	"github.com/sumanta-mukhopadhyay/claude-switch/internal/claude"
	"github.com/sumanta-mukhopadhyay/claude-switch/internal/config"
	"github.com/sumanta-mukhopadhyay/claude-switch/internal/profile"
	"github.com/sumanta-mukhopadhyay/claude-switch/internal/ui"
)

// CheckResult represents the result of a single diagnostic check.
type CheckResult struct {
	Name    string
	Status  string // "ok", "warn", "fail"
	Message string
}

// RunAll executes all diagnostic checks and returns the results.
func RunAll(cfg *config.Config) []CheckResult {
	var results []CheckResult

	results = append(results, checkPlatform())
	results = append(results, checkClaudeCLI())
	results = append(results, checkClaudeHome())
	results = append(results, checkCredentials())
	results = append(results, checkConfigFile())
	results = append(results, checkProfiles(cfg))
	results = append(results, checkPermissions())

	return results
}

// PrintResults displays the check results to the user.
func PrintResults(results []CheckResult) {
	ui.Header("Claude Switch Doctor")
	fmt.Println()

	okCount, warnCount, failCount := 0, 0, 0

	for _, r := range results {
		var icon string
		switch r.Status {
		case "ok":
			icon = ui.Colorize(ui.Green, "✓")
			okCount++
		case "warn":
			icon = ui.Colorize(ui.Yellow, "⚠")
			warnCount++
		case "fail":
			icon = ui.Colorize(ui.Red, "✗")
			failCount++
		}
		fmt.Printf("  %s  %-20s %s\n", icon, r.Name, r.Message)
	}

	fmt.Println()
	summary := fmt.Sprintf("Results: %d passed", okCount)
	if warnCount > 0 {
		summary += fmt.Sprintf(", %d warnings", warnCount)
	}
	if failCount > 0 {
		summary += fmt.Sprintf(", %d failed", failCount)
	}

	if failCount > 0 {
		ui.Error("%s", summary)
	} else if warnCount > 0 {
		ui.Warn("%s", summary)
	} else {
		ui.Success("%s", summary)
	}
}

func checkPlatform() CheckResult {
	return CheckResult{
		Name:    "Platform",
		Status:  "ok",
		Message: fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}

func checkClaudeCLI() CheckResult {
	result := claude.Detect()
	if !result.Found {
		return CheckResult{
			Name:    "Claude CLI",
			Status:  "fail",
			Message: "claude binary not found on PATH",
		}
	}
	return CheckResult{
		Name:    "Claude CLI",
		Status:  "ok",
		Message: fmt.Sprintf("%s (%s)", result.Version, result.Path),
	}
}

func checkClaudeHome() CheckResult {
	dir, err := config.ClaudeConfigDir()
	if err != nil {
		return CheckResult{
			Name:    "Claude Home",
			Status:  "fail",
			Message: err.Error(),
		}
	}
	if !profile.DirExists(dir) {
		return CheckResult{
			Name:    "Claude Home",
			Status:  "fail",
			Message: fmt.Sprintf("%s does not exist", dir),
		}
	}
	return CheckResult{
		Name:    "Claude Home",
		Status:  "ok",
		Message: dir,
	}
}

func checkCredentials() CheckResult {
	path, err := config.ClaudeCredentialsPath()
	if err != nil {
		return CheckResult{
			Name:    "Credentials",
			Status:  "fail",
			Message: err.Error(),
		}
	}
	if !profile.FileExists(path) {
		return CheckResult{
			Name:    "Credentials",
			Status:  "warn",
			Message: "no credentials file found — are you logged in?",
		}
	}
	return CheckResult{
		Name:    "Credentials",
		Status:  "ok",
		Message: "credentials file present",
	}
}

func checkConfigFile() CheckResult {
	path, err := config.ConfigPath()
	if err != nil {
		return CheckResult{
			Name:    "Config File",
			Status:  "fail",
			Message: err.Error(),
		}
	}
	if !profile.FileExists(path) {
		return CheckResult{
			Name:    "Config File",
			Status:  "warn",
			Message: "no config file yet (will be created on first import)",
		}
	}

	// Try loading it
	_, err = config.Load()
	if err != nil {
		return CheckResult{
			Name:    "Config File",
			Status:  "fail",
			Message: fmt.Sprintf("invalid config: %s", err),
		}
	}

	return CheckResult{
		Name:    "Config File",
		Status:  "ok",
		Message: path,
	}
}

func checkProfiles(cfg *config.Config) CheckResult {
	if len(cfg.Profiles) == 0 {
		return CheckResult{
			Name:    "Profiles",
			Status:  "warn",
			Message: "no profiles saved yet",
		}
	}

	profilesDir, err := config.ProfilesDir()
	if err != nil {
		return CheckResult{
			Name:    "Profiles",
			Status:  "fail",
			Message: err.Error(),
		}
	}

	missing := 0
	for name := range cfg.Profiles {
		dir := fmt.Sprintf("%s/%s", profilesDir, name)
		if !profile.DirExists(dir) {
			missing++
		}
	}

	if missing > 0 {
		return CheckResult{
			Name:    "Profiles",
			Status:  "warn",
			Message: fmt.Sprintf("%d/%d profiles have missing directories", missing, len(cfg.Profiles)),
		}
	}

	return CheckResult{
		Name:    "Profiles",
		Status:  "ok",
		Message: fmt.Sprintf("%d profiles configured", len(cfg.Profiles)),
	}
}

func checkPermissions() CheckResult {
	if runtime.GOOS == "windows" {
		return CheckResult{
			Name:    "Permissions",
			Status:  "ok",
			Message: "skipped (Windows)",
		}
	}

	appDir, err := config.AppDataDir()
	if err != nil {
		return CheckResult{
			Name:    "Permissions",
			Status:  "fail",
			Message: err.Error(),
		}
	}

	if !profile.DirExists(appDir) {
		return CheckResult{
			Name:    "Permissions",
			Status:  "ok",
			Message: "app directory not yet created",
		}
	}

	info, err := os.Stat(appDir)
	if err != nil {
		return CheckResult{
			Name:    "Permissions",
			Status:  "fail",
			Message: err.Error(),
		}
	}

	perm := info.Mode().Perm()
	if perm&0077 != 0 {
		return CheckResult{
			Name:    "Permissions",
			Status:  "warn",
			Message: fmt.Sprintf("%s has permissions %o (expected 0700)", appDir, perm),
		}
	}

	return CheckResult{
		Name:    "Permissions",
		Status:  "ok",
		Message: fmt.Sprintf("%s (0%o)", appDir, perm),
	}
}

package claude

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/caeser1996/claude-switch/internal/config"
	"github.com/caeser1996/claude-switch/internal/profile"
)

// UsageInfo represents parsed usage/limit information.
type UsageInfo struct {
	Plan        string `json:"plan,omitempty"`
	Model       string `json:"model,omitempty"`
	Email       string `json:"email,omitempty"`
	OrgName     string `json:"org_name,omitempty"`
	RateLimited bool   `json:"rate_limited"`
	ResetTime   string `json:"reset_time,omitempty"`
}

// GetUsageInfo tries to extract usage information from Claude's local data files.
// This is best-effort since Claude Code doesn't expose a formal API for this.
func GetUsageInfo() (*UsageInfo, error) {
	info := &UsageInfo{}

	// Primary email source: ~/.claude.json → oauthAccount.emailAddress
	// This is where Claude Code stores the logged-in user's email.
	if home, err := os.UserHomeDir(); err == nil {
		parseClaudeJSON(filepath.Join(home, ".claude.json"), info)
	}

	// Try reading plan/org from the credentials file.
	credPath, err := config.ClaudeCredentialsPath()
	if err == nil {
		parseCredentialsInfo(credPath, info)
	}

	// If no data from file, try the platform credential store (Keychain on macOS).
	if info.Email == "" && info.Plan == "" {
		if data, err := profile.ReadCurrentCredentials(); err == nil {
			parseCredentialsData(data, info)
		}
	}

	// Try reading from statsig metadata for plan/model info
	claudeDir, err := config.ClaudeConfigDir()
	if err == nil {
		parseStatsigMetadata(filepath.Join(claudeDir, "statsig_metadata"), info)
	}

	return info, nil
}

// parseClaudeJSON extracts email from ~/.claude.json.
// Format: {"oauthAccount": {"emailAddress": "user@example.com"}}
func parseClaudeJSON(path string, info *UsageInfo) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	var cfg map[string]interface{}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return
	}

	if acct, ok := cfg["oauthAccount"].(map[string]interface{}); ok {
		if email, ok := acct["emailAddress"].(string); ok && email != "" {
			info.Email = email
		}
	}
}

// parseCredentialsInfo extracts email and account info from a credentials file.
func parseCredentialsInfo(path string, info *UsageInfo) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	parseCredentialsData(data, info)
}

// parseCredentialsData extracts email and account info from raw JSON bytes.
// It checks both top-level fields and nested objects (e.g., claudeAiOauth).
func parseCredentialsData(data []byte, info *UsageInfo) {
	var creds map[string]interface{}
	if err := json.Unmarshal(data, &creds); err != nil {
		return
	}

	extractFields(creds, info)

	// If email still not found, check nested objects that might contain credentials.
	if info.Email == "" {
		for _, key := range []string{"claudeAiOauth", "oauth", "credentials", "account"} {
			if nested, ok := creds[key].(map[string]interface{}); ok {
				extractFields(nested, info)
				if info.Email != "" {
					break
				}
			}
		}
	}
}

// extractFields pulls known fields from a JSON map into UsageInfo.
func extractFields(m map[string]interface{}, info *UsageInfo) {
	if info.Email == "" {
		for _, key := range []string{"email", "userEmail", "user_email"} {
			if v, ok := m[key].(string); ok && v != "" {
				info.Email = v
				break
			}
		}
	}
	if info.Plan == "" {
		for _, key := range []string{"plan", "planType", "plan_type"} {
			if v, ok := m[key].(string); ok && v != "" {
				info.Plan = v
				break
			}
		}
	}
	if info.OrgName == "" {
		for _, key := range []string{"orgName", "org_name", "organization"} {
			if v, ok := m[key].(string); ok && v != "" {
				info.OrgName = v
				break
			}
		}
	}
}

// parseStatsigMetadata extracts model/plan info from statsig files.
func parseStatsigMetadata(path string, info *UsageInfo) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	var meta map[string]interface{}
	if err := json.Unmarshal(data, &meta); err != nil {
		return
	}

	if model, ok := meta["model"].(string); ok {
		info.Model = model
	}
}

// FormatUsageInfo formats usage info as a display string.
func FormatUsageInfo(info *UsageInfo, profileName string) string {
	var b strings.Builder

	width := 39
	border := strings.Repeat("═", width)

	b.WriteString(fmt.Sprintf("╔═%s═╗\n", border))
	title := fmt.Sprintf("Claude Usage - %s", profileName)
	padding := width - len(title)
	if padding < 0 {
		padding = 0
	}
	b.WriteString(fmt.Sprintf("║ %s%s ║\n", title, strings.Repeat(" ", padding)))
	b.WriteString(fmt.Sprintf("╠═%s═╣\n", border))

	rows := []struct {
		label, value string
	}{
		{"Plan", valueOrDash(info.Plan)},
		{"Email", valueOrDash(info.Email)},
		{"Model", valueOrDash(info.Model)},
		{"Organization", valueOrDash(info.OrgName)},
		{"Rate Limited", fmt.Sprintf("%v", info.RateLimited)},
	}

	if info.ResetTime != "" {
		rows = append(rows, struct{ label, value string }{"Reset", info.ResetTime})
	}

	for _, r := range rows {
		line := fmt.Sprintf("%-15s %s", r.label+":", r.value)
		pad := width - len(line)
		if pad < 0 {
			pad = 0
		}
		b.WriteString(fmt.Sprintf("║ %s%s ║\n", line, strings.Repeat(" ", pad)))
	}

	now := time.Now().Format("2006-01-02 15:04 MST")
	timeLine := fmt.Sprintf("Checked: %s", now)
	timePad := width - len(timeLine)
	if timePad < 0 {
		timePad = 0
	}
	b.WriteString(fmt.Sprintf("╠═%s═╣\n", border))
	b.WriteString(fmt.Sprintf("║ %s%s ║\n", timeLine, strings.Repeat(" ", timePad)))
	b.WriteString(fmt.Sprintf("╚═%s═╝", border))

	return b.String()
}

func valueOrDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

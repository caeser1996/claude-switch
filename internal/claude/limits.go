package claude

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sumanta-mukhopadhyay/claude-switch/internal/config"
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

	// Try reading from credentials for email/plan info
	credPath, err := config.ClaudeCredentialsPath()
	if err == nil {
		parseCredentialsInfo(credPath, info)
	}

	// Try reading from statsig metadata for plan/model info
	claudeDir, err := config.ClaudeConfigDir()
	if err == nil {
		parseStatsigMetadata(filepath.Join(claudeDir, "statsig_metadata"), info)
	}

	return info, nil
}

// parseCredentialsInfo extracts email and account info from credentials.
func parseCredentialsInfo(path string, info *UsageInfo) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	// Try to parse as JSON and extract known fields
	var creds map[string]interface{}
	if err := json.Unmarshal(data, &creds); err != nil {
		return
	}

	if email, ok := creds["email"].(string); ok {
		info.Email = email
	}

	if plan, ok := creds["plan"].(string); ok {
		info.Plan = plan
	}

	if org, ok := creds["orgName"].(string); ok {
		info.OrgName = org
	}
	if org, ok := creds["org_name"].(string); ok && info.OrgName == "" {
		info.OrgName = org
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

package profile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/caeser1996/claude-switch/internal/config"
)

// TokenStatus represents the health of a profile's credentials.
type TokenStatus struct {
	ProfileName string
	HasCreds    bool
	Email       string
	ExpiresAt   *time.Time
	IsExpired   bool
	ExpiresIn   string
}

// CheckTokenStatus inspects a profile's credentials and returns status info.
func CheckTokenStatus(name string) TokenStatus {
	status := TokenStatus{ProfileName: name}

	profilesDir, err := config.ProfilesDir()
	if err != nil {
		return status
	}

	credPath := filepath.Join(profilesDir, name, ".credentials.json")
	if !FileExists(credPath) {
		return status
	}
	status.HasCreds = true

	data, err := os.ReadFile(credPath)
	if err != nil {
		return status
	}

	var creds map[string]interface{}
	if err := json.Unmarshal(data, &creds); err != nil {
		return status
	}

	// Check top-level and nested objects for email/expiry.
	sources := []map[string]interface{}{creds}
	for _, key := range []string{"claudeAiOauth", "oauth", "credentials", "account"} {
		if nested, ok := creds[key].(map[string]interface{}); ok {
			sources = append(sources, nested)
		}
	}

	for _, src := range sources {
		if status.Email == "" {
			for _, key := range []string{"email", "userEmail", "user_email"} {
				if v, ok := src[key].(string); ok && v != "" {
					status.Email = v
					break
				}
			}
		}
		if status.ExpiresAt == nil {
			for _, key := range []string{"expiresAt", "expires_at", "exp"} {
				if v, ok := src[key]; ok {
					if expStr, ok := v.(string); ok {
						if t, err := time.Parse(time.RFC3339, expStr); err == nil {
							status.ExpiresAt = &t
							status.IsExpired = time.Now().After(t)
							if !status.IsExpired {
								status.ExpiresIn = formatDuration(time.Until(t))
							}
							break
						}
					}
					if expNum, ok := v.(float64); ok {
						t := time.Unix(int64(expNum), 0)
						status.ExpiresAt = &t
						status.IsExpired = time.Now().After(t)
						if !status.IsExpired {
							status.ExpiresIn = formatDuration(time.Until(t))
						}
						break
					}
				}
			}
		}
	}

	return status
}

// NeedsRefresh returns true if the token is expired or will expire within the threshold.
func NeedsRefresh(status TokenStatus, threshold time.Duration) bool {
	if !status.HasCreds {
		return true
	}
	if status.IsExpired {
		return true
	}
	if status.ExpiresAt != nil {
		return time.Until(*status.ExpiresAt) < threshold
	}
	return false
}

func formatDuration(d time.Duration) string {
	if d < 0 {
		return "expired"
	}
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh", days, hours)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}

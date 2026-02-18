package profile

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/caeser1996/claude-switch/internal/config"
)

func TestCheckTokenStatusValid(t *testing.T) {
	tmpDir, cleanup := setupTestEnv(t)
	defer cleanup()

	cfg := config.NewConfig()
	mgr := NewManager(cfg)
	mgr.Import("token-test", "")

	// Write credentials with expiry in the future
	profilesDir, _ := config.ProfilesDir()
	credPath := filepath.Join(profilesDir, "token-test", ".credentials.json")
	future := time.Now().Add(48 * time.Hour).Format(time.RFC3339)
	creds := map[string]interface{}{
		"email":     "test@example.com",
		"expiresAt": future,
	}
	data, _ := json.Marshal(creds)
	os.WriteFile(credPath, data, 0600)

	status := CheckTokenStatus("token-test")
	_ = tmpDir

	if !status.HasCreds {
		t.Error("expected HasCreds to be true")
	}
	if status.Email != "test@example.com" {
		t.Errorf("expected email 'test@example.com', got %q", status.Email)
	}
	if status.IsExpired {
		t.Error("token should not be expired")
	}
	if status.ExpiresAt == nil {
		t.Error("ExpiresAt should be set")
	}
	if status.ExpiresIn == "" {
		t.Error("ExpiresIn should be set")
	}
}

func TestCheckTokenStatusExpired(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	cfg := config.NewConfig()
	mgr := NewManager(cfg)
	mgr.Import("expired-test", "")

	// Write credentials with expiry in the past
	profilesDir, _ := config.ProfilesDir()
	credPath := filepath.Join(profilesDir, "expired-test", ".credentials.json")
	past := time.Now().Add(-48 * time.Hour).Format(time.RFC3339)
	creds := map[string]interface{}{
		"email":     "expired@example.com",
		"expiresAt": past,
	}
	data, _ := json.Marshal(creds)
	os.WriteFile(credPath, data, 0600)

	status := CheckTokenStatus("expired-test")

	if !status.IsExpired {
		t.Error("token should be expired")
	}
}

func TestCheckTokenStatusNoCreds(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	status := CheckTokenStatus("nonexistent")
	if status.HasCreds {
		t.Error("expected HasCreds to be false")
	}
}

func TestNeedsRefresh(t *testing.T) {
	// Expired token
	expired := TokenStatus{HasCreds: true, IsExpired: true}
	if !NeedsRefresh(expired, time.Hour) {
		t.Error("expired token should need refresh")
	}

	// No creds
	noCreds := TokenStatus{HasCreds: false}
	if !NeedsRefresh(noCreds, time.Hour) {
		t.Error("missing creds should need refresh")
	}

	// Valid with plenty of time
	future := time.Now().Add(72 * time.Hour)
	valid := TokenStatus{HasCreds: true, ExpiresAt: &future}
	if NeedsRefresh(valid, time.Hour) {
		t.Error("token with 72h left should not need refresh with 1h threshold")
	}

	// Expiring soon
	soon := time.Now().Add(30 * time.Minute)
	expiring := TokenStatus{HasCreds: true, ExpiresAt: &soon}
	if !NeedsRefresh(expiring, time.Hour) {
		t.Error("token expiring in 30m should need refresh with 1h threshold")
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{48 * time.Hour, "2d 0h"},
		{2*time.Hour + 30*time.Minute, "2h 30m"},
		{45 * time.Minute, "45m"},
		{-1 * time.Hour, "expired"},
	}

	for _, tt := range tests {
		got := formatDuration(tt.d)
		if got != tt.want {
			t.Errorf("formatDuration(%v) = %q, want %q", tt.d, got, tt.want)
		}
	}
}

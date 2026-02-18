package claude

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetUsageInfo(t *testing.T) {
	// This is a best-effort test; in CI there may be no Claude config
	info, err := GetUsageInfo()
	if err != nil {
		t.Fatalf("GetUsageInfo should not error: %v", err)
	}
	if info == nil {
		t.Fatal("GetUsageInfo returned nil")
	}
}

func TestParseCredentialsInfo(t *testing.T) {
	tmpDir := t.TempDir()

	creds := map[string]interface{}{
		"email":   "test@example.com",
		"plan":    "Max",
		"orgName": "TestOrg",
	}
	data, _ := json.Marshal(creds)
	credPath := filepath.Join(tmpDir, "creds.json")
	os.WriteFile(credPath, data, 0600)

	info := &UsageInfo{}
	parseCredentialsInfo(credPath, info)

	if info.Email != "test@example.com" {
		t.Errorf("expected email 'test@example.com', got %q", info.Email)
	}
	if info.Plan != "Max" {
		t.Errorf("expected plan 'Max', got %q", info.Plan)
	}
	if info.OrgName != "TestOrg" {
		t.Errorf("expected orgName 'TestOrg', got %q", info.OrgName)
	}
}

func TestParseCredentialsInfoMissing(t *testing.T) {
	info := &UsageInfo{}
	parseCredentialsInfo("/nonexistent/path", info)

	// Should not panic, just leave fields empty
	if info.Email != "" {
		t.Errorf("expected empty email, got %q", info.Email)
	}
}

func TestParseStatsigMetadata(t *testing.T) {
	tmpDir := t.TempDir()

	meta := map[string]interface{}{
		"model": "opus-4.5",
	}
	data, _ := json.Marshal(meta)
	metaPath := filepath.Join(tmpDir, "statsig_metadata")
	os.WriteFile(metaPath, data, 0600)

	info := &UsageInfo{}
	parseStatsigMetadata(metaPath, info)

	if info.Model != "opus-4.5" {
		t.Errorf("expected model 'opus-4.5', got %q", info.Model)
	}
}

func TestFormatUsageInfo(t *testing.T) {
	info := &UsageInfo{
		Plan:    "Max ($100/mo)",
		Email:   "test@example.com",
		Model:   "Opus 4.5",
		OrgName: "TestOrg",
	}

	output := FormatUsageInfo(info, "work")

	// Should contain the profile name
	if !strings.Contains(output, "work") {
		t.Error("output should contain profile name")
	}

	// Should contain the plan
	if !strings.Contains(output, "Max ($100/mo)") {
		t.Error("output should contain plan")
	}

	// Should contain email
	if !strings.Contains(output, "test@example.com") {
		t.Error("output should contain email")
	}

	// Should contain box drawing characters
	if !strings.Contains(output, "╔") || !strings.Contains(output, "╚") {
		t.Error("output should contain box drawing characters")
	}
}

func TestFormatUsageInfoEmpty(t *testing.T) {
	info := &UsageInfo{}
	output := FormatUsageInfo(info, "empty")

	// Should show dashes for empty values
	if !strings.Contains(output, "-") {
		t.Error("output should contain dashes for empty values")
	}
}

func TestValueOrDash(t *testing.T) {
	if valueOrDash("") != "-" {
		t.Error("empty string should return dash")
	}
	if valueOrDash("hello") != "hello" {
		t.Error("non-empty string should return itself")
	}
}

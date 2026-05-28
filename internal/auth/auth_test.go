package auth

import (
	"os"
	"path/filepath"
	"testing"

	"gemini-cli-account-manager/internal/config"
)

func TestGetProfiles(t *testing.T) {
	// Setup mock environment
	tmpDir, _ := os.MkdirTemp("", "gemini-test-*")
	defer os.RemoveAll(tmpDir)

	config.GeminiDir = tmpDir
	config.ProfilesDir = filepath.Join(tmpDir, "auth_profiles")
	_ = os.MkdirAll(config.ProfilesDir, 0755)

	// Test empty
	profiles, _ := GetProfiles()
	if len(profiles) != 0 {
		t.Errorf("Expected 0 profiles, got %d", len(profiles))
	}

	// Add mock profiles
	_ = os.Mkdir(filepath.Join(config.ProfilesDir, "user1@example.com"), 0755)
	_ = os.Mkdir(filepath.Join(config.ProfilesDir, "user2@example.com"), 0755)

	profiles, _ = GetProfiles()
	if len(profiles) != 2 {
		t.Errorf("Expected 2 profiles, got %d", len(profiles))
	}
	if profiles[0] != "user1@example.com" {
		t.Errorf("Expected user1@example.com, got %s", profiles[0])
	}
}

func TestCopyFile(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "copy-test-*")
	defer os.RemoveAll(tmpDir)

	src := filepath.Join(tmpDir, "src.txt")
	dst := filepath.Join(tmpDir, "dst.txt")
	content := "hello world"

	_ = os.WriteFile(src, []byte(content), 0644)

	err := CopyFile(src, dst)
	if err != nil {
		t.Fatalf("CopyFile failed: %v", err)
	}

	got, _ := os.ReadFile(dst)
	if string(got) != content {
		t.Errorf("Expected %s, got %s", content, string(got))
	}
}

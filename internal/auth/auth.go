package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"

	"gemini-cli-account-manager/internal/config"
	"gemini-cli-account-manager/internal/i18n"
)

// Profile represents a saved account profile
type Profile struct {
	Email string
	Dir   string
}

// GetProfiles returns all saved profiles from the profiles directory
func GetProfiles() ([]string, error) {
	entries, err := os.ReadDir(config.ProfilesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	var profiles []string
	for _, entry := range entries {
		if entry.IsDir() {
			profiles = append(profiles, entry.Name())
		}
	}
	sort.Strings(profiles)
	return profiles, nil
}

// CopyFile copies a file from src to dst atomically
func CopyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// Create a temporary file
	tempDir := filepath.Dir(dst)
	tempFile, err := os.CreateTemp(tempDir, "copy-temp-*")
	if err != nil {
		return err
	}
	defer func() {
		tempFile.Close()
		os.Remove(tempFile.Name())
	}()

	if _, err = io.Copy(tempFile, sourceFile); err != nil {
		return err
	}
	if err := tempFile.Sync(); err != nil {
		return err
	}
	if err := tempFile.Close(); err != nil {
		return err
	}

	return os.Rename(tempFile.Name(), dst)
}

// FastSwitch switches to the specified account by index or email
func FastSwitch(targetArg string) (string, error) {
	profiles, err := GetProfiles()
	if err != nil {
		return "", err
	}
	if len(profiles) == 0 {
		return "", fmt.Errorf(i18n.T("no_profiles_found"))
	}

	targetEmail := targetArg
	targetDir := filepath.Join(config.ProfilesDir, targetArg)

	// Handle numeric index
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		if idx, err := strconv.Atoi(targetArg); err == nil {
			idx-- // 1-based to 0-based
			if idx >= 0 && idx < len(profiles) {
				targetEmail = profiles[idx]
				targetDir = filepath.Join(config.ProfilesDir, targetEmail)
			} else {
				return "", fmt.Errorf(i18n.T("index_out_of_range"), targetArg, len(profiles))
			}
		} else {
			return "", fmt.Errorf(i18n.T("account_not_found"), targetArg)
		}
	}

	targetCreds := filepath.Join(targetDir, "oauth_creds.json")
	if _, err := os.Stat(targetCreds); os.IsNotExist(err) {
		return "", fmt.Errorf(i18n.T("missing_creds"), targetEmail)
	}

	accs, err := config.LoadAccounts()
	if err != nil {
		return "", err
	}
	currentActive := accs.Active

	if currentActive == targetEmail {
		return targetEmail, nil // Already active
	}

	// 1. Backup current credentials if active
	if currentActive != "" {
		currDir := filepath.Join(config.ProfilesDir, currentActive)
		if err := os.MkdirAll(currDir, 0755); err != nil {
			return "", fmt.Errorf("failed to create backup dir: %w", err)
		}

		if _, err := os.Stat(config.CredsFile); err == nil {
			if err := CopyFile(config.CredsFile, filepath.Join(currDir, "oauth_creds.json")); err != nil {
				return "", fmt.Errorf("failed to backup creds: %w", err)
			}
		}
		if _, err := os.Stat(config.IDFile); err == nil {
			if err := CopyFile(config.IDFile, filepath.Join(currDir, "google_account_id")); err != nil {
				return "", fmt.Errorf("failed to backup id: %w", err)
			}
		}
	}

	// 2. Perform switch (Copy from profile to ~/.gemini)
	if err := CopyFile(targetCreds, config.CredsFile); err != nil {
		return "", fmt.Errorf("failed to copy target creds: %w", err)
	}

	targetIDFile := filepath.Join(targetDir, "google_account_id")
	if _, err := os.Stat(targetIDFile); err == nil {
		if err := CopyFile(targetIDFile, config.IDFile); err != nil {
			return "", fmt.Errorf("failed to copy target id: %w", err)
		}
	}

	// 3. Update accounts.json
	accs.Active = targetEmail
	if err := config.SaveAccounts(accs); err != nil {
		return "", fmt.Errorf("failed to update accounts metadata: %w", err)
	}

	return targetEmail, nil
}

// SwitchNext switches to the next account in the pool
func SwitchNext() (string, error) {
	profiles, err := GetProfiles()
	if err != nil {
		return "", err
	}
	if len(profiles) == 0 {
		return "", fmt.Errorf(i18n.T("no_profiles_found"))
	}

	accs, err := config.LoadAccounts()
	if err != nil {
		return "", err
	}

	currentIdx := -1
	for i, p := range profiles {
		if p == accs.Active {
			currentIdx = i
			break
		}
	}

	nextIdx := (currentIdx + 1) % len(profiles)
	return FastSwitch(profiles[nextIdx])
}

// RemoveAccount removes an account from the pool by index or email
func RemoveAccount(targetArg string) error {
	profiles, err := GetProfiles()
	if err != nil {
		return err
	}
	if len(profiles) == 0 {
		return fmt.Errorf(i18n.T("no_profiles_found"))
	}

	accs, err := config.LoadAccounts()
	if err != nil {
		return err
	}

	targetEmail := targetArg
	targetDir := filepath.Join(config.ProfilesDir, targetArg)

	// Handle numeric index
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		if idx, err := strconv.Atoi(targetArg); err == nil {
			idx-- // 1-based to 0-based
			if idx >= 0 && idx < len(profiles) {
				targetEmail = profiles[idx]
				targetDir = filepath.Join(config.ProfilesDir, targetEmail)
			} else {
				return fmt.Errorf(i18n.T("index_out_of_range"), targetArg, len(profiles))
			}
		} else {
			return fmt.Errorf(i18n.T("account_not_found"), targetArg)
		}
	}

	// Cannot remove active account
	if targetEmail == accs.Active {
		return fmt.Errorf("cannot remove active account, please switch to another account first")
	}

	// Remove profile directory
	if err := os.RemoveAll(targetDir); err != nil {
		return fmt.Errorf("failed to remove profile: %w", err)
	}

	// Update accounts.json
	if accs.Old == nil {
		accs.Old = []string{}
	}
	for i, email := range accs.Old {
		if email == targetEmail {
			accs.Old = append(accs.Old[:i], accs.Old[i+1:]...)
			break
		}
	}
	return config.SaveAccounts(accs)
}

// ImportAccount imports account credentials from a file
func ImportAccount(credsPath, email string) error {
	// Read credentials file
	data, err := os.ReadFile(credsPath)
	if err != nil {
		return fmt.Errorf("failed to read credentials file: %w", err)
	}

	// Validate JSON
	var creds map[string]interface{}
	if err := json.Unmarshal(data, &creds); err != nil {
		return fmt.Errorf("invalid credentials file: not valid JSON")
	}

	// Validate required fields
	required := []string{"access_token", "refresh_token"}
	for _, field := range required {
		if _, ok := creds[field]; !ok {
			return fmt.Errorf("invalid credentials file: missing %s", field)
		}
	}

	// Create profile directory
	profileDir := filepath.Join(config.ProfilesDir, email)
	if err := os.MkdirAll(profileDir, 0755); err != nil {
		return fmt.Errorf("failed to create profile directory: %w", err)
	}

	// Copy credentials
	if err := os.WriteFile(filepath.Join(profileDir, "oauth_creds.json"), data, 0644); err != nil {
		return fmt.Errorf("failed to write credentials: %w", err)
	}

	// Also look for google_account_id in the same directory
	idSrcPath := filepath.Join(filepath.Dir(credsPath), "google_account_id")
	if _, err := os.Stat(idSrcPath); err == nil {
		idData, _ := os.ReadFile(idSrcPath)
		_ = os.WriteFile(filepath.Join(profileDir, "google_account_id"), idData, 0644)
	}

	return nil
}

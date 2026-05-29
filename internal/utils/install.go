package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"gemini-cli-account-manager/internal/config"
)

// HookConfig represents the structure of hook settings in settings.json
type Settings struct {
	BeforeAgent []string `json:"beforeAgent,omitempty"`
	AfterAgent  []string `json:"afterAgent,omitempty"`
}

// InstallHooks registers the gcam binary as hooks in settings.json
func InstallHooks(binPath string) error {
	// 1. Update settings.json
	var settings map[string]interface{}
	data, err := os.ReadFile(config.SettingsFile)
	if err == nil {
		_ = json.Unmarshal(data, &settings)
	} else {
		settings = make(map[string]interface{})
	}

	// We'll use the binary path for hooks
	// beforeAgent: [binPath, "quota-pre-check"]
	// afterAgent: [binPath, "quota-auto-switch"] (or similar)

	// Note: The original Python version has separate scripts for hooks.
	// In our Go version, we can use subcommands.

	settings["beforeAgent"] = []string{binPath, "hook", "pre-check"}
	settings["afterAgent"] = []string{binPath, "hook", "auto-switch"}

	newData, _ := json.MarshalIndent(settings, "", "  ")
	if err := os.WriteFile(config.SettingsFile, newData, 0644); err != nil {
		return err
	}

	// 2. Create slash command gcam.toml
	commandsDir := filepath.Join(config.GeminiDir, "commands")
	_ = os.MkdirAll(commandsDir, 0755)

	// 使用 prompt + !{...} 语法，与 Python 版本保持一致
	// 格式: prompt = "!{/path/to/gcam {{args}}}"
	promptContent := fmt.Sprintf("!{%s {{args}}}", binPath)
	tomlContent := fmt.Sprintf(`# Gemini CLI Slash Command: /gcam
description = "Switch Gemini accounts and manage rotation. Usage: /gcam <index|next|quota|menu>"
prompt = "%s"
`, promptContent)

	return os.WriteFile(filepath.Join(commandsDir, "gcam.toml"), []byte(tomlContent), 0644)
}

// UninstallOptions contains options for uninstallation
type UninstallOptions struct {
	KeepAccounts bool // If true, keep account data (profiles, accounts.json)
}

// UninstallHooks removes gcam from settings.json and optionally keeps accounts
func UninstallHooks(opts UninstallOptions) error {
	// 1. Remove from settings.json (clean hooks)
	data, err := os.ReadFile(config.SettingsFile)
	if err == nil {
		var settings map[string]interface{}
		if err := json.Unmarshal(data, &settings); err == nil {
			delete(settings, "beforeAgent")
			delete(settings, "afterAgent")
			newData, _ := json.MarshalIndent(settings, "", "  ")
			_ = os.WriteFile(config.SettingsFile, newData, 0644)
		}
	}

	// 2. Remove slash commands
	commandsDir := filepath.Join(config.GeminiDir, "commands")
	_ = os.Remove(filepath.Join(commandsDir, "gcam.toml"))
	_ = os.Remove(filepath.Join(commandsDir, "change.toml"))

	// 3. Remove launchers
	_ = os.Remove(filepath.Join(config.GeminiDir, "gcam"))
	_ = os.Remove(filepath.Join(config.GeminiDir, "gcam.bat"))

	// 4. Remove account data if not keeping accounts
	if !opts.KeepAccounts {
		// Remove auth_profiles directory
		_ = os.RemoveAll(config.ProfilesDir)
		// Remove google_accounts.json
		_ = os.Remove(config.AccountsJSON)
	}

	// 5. Remove config file
	_ = os.Remove(config.ConfigFile)

	// 6. Clean environment variables (Windows only effect)
	if runtime.GOOS == "windows" {
		// On Windows, we can't easily unset env vars from a process
		// The user will need to manually remove GEMINI_FORCE_FILE_STORAGE
	}

	return nil
}

// GetUninstallFiles returns a list of files that would be removed
func GetUninstallFiles(keepAccounts bool) []string {
	var files []string

	// Hooks and commands
	files = append(files, filepath.Join(config.GeminiDir, "commands", "gcam.toml"))
	files = append(files, filepath.Join(config.GeminiDir, "commands", "change.toml"))

	// Launchers
	files = append(files, filepath.Join(config.GeminiDir, "gcam"))
	files = append(files, filepath.Join(config.GeminiDir, "gcam.bat"))

	// Config
	files = append(files, config.ConfigFile)

	// Account data
	if !keepAccounts {
		files = append(files, config.ProfilesDir)
		files = append(files, config.AccountsJSON)
	}

	return files
}

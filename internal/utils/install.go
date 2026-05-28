package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

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

	tomlContent := fmt.Sprintf(`# Gemini CLI Slash Command: /gcam
name = "gcam"
description = "Switch Gemini accounts and manage rotation"
command = ["%s"]
`, binPath)

	return os.WriteFile(filepath.Join(commandsDir, "gcam.toml"), []byte(tomlContent), 0644)
}

// UninstallHooks removes gcam from settings.json
func UninstallHooks() error {
	// 1. Remove from settings.json
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

	return nil
}

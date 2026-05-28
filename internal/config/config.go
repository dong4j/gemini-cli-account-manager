package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Default paths in ~/.gemini
var (
	GeminiDir    string
	ProfilesDir  string
	AccountsJSON string
	CredsFile    string
	IDFile       string
	ConfigFile   string
	SettingsFile string
)

func InitPaths() {
	home, _ := os.UserHomeDir()
	GeminiDir = filepath.Join(home, ".gemini")
	if envDir := os.Getenv("GEMINI_HOME"); envDir != "" {
		GeminiDir = envDir
	}
	ProfilesDir = filepath.Join(GeminiDir, "auth_profiles")
	AccountsJSON = filepath.Join(GeminiDir, "google_accounts.json")
	CredsFile = filepath.Join(GeminiDir, "oauth_creds.json")
	IDFile = filepath.Join(GeminiDir, "google_account_id")
	ConfigFile = filepath.Join(GeminiDir, "auth_config.json")
	SettingsFile = filepath.Join(GeminiDir, "settings.json")
}

func init() {
	InitPaths()
}

// Config represents the tool's own configuration (auth_config.json)
type Config struct {
	Language    string      `json:"language"`
	OAuthClient OAuthClient `json:"oauth_client"`
	AutoSwitch  AutoSwitch  `json:"auto_switch"`
}

type OAuthClient struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type AutoSwitch struct {
	Enabled            bool    `json:"enabled"`
	Strategy           string  `json:"strategy"`
	ModelPattern       string  `json:"model_pattern"`
	CustomModelPattern string  `json:"custom_model_pattern"`
	Threshold          float64 `json:"threshold"`
	MaxRetries         int     `json:"max_retries"`
	NotifyOnSwitch     bool    `json:"notify_on_switch"`
	AutoRestart        bool    `json:"auto_restart"`
	CacheMinutes       int     `json:"cache_minutes"`
}

// DefaultConfig provides the default settings
func DefaultConfig() Config {
	return Config{
		Language: "en",
		OAuthClient: OAuthClient{
			ClientID:     "681255809395-oo8ft2oprdrnp9e3aqf6av3hmdib135j.apps.googleusercontent.com",
			ClientSecret: "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",
		},
		AutoSwitch: AutoSwitch{
			Enabled:            true,
			Strategy:           "gemini3.1-series-only",
			ModelPattern:       "gemini-3.1.*",
			CustomModelPattern: "",
			Threshold:          10.0,
			MaxRetries:         3,
			NotifyOnSwitch:     true,
			AutoRestart:        false,
			CacheMinutes:       3,
		},
	}
}

// Load loads the configuration from disk, creating it if missing
func Load() (*Config, error) {
	if _, err := os.Stat(ConfigFile); os.IsNotExist(err) {
		cfg := DefaultConfig()
		if err := Save(&cfg); err != nil {
			return nil, err
		}
		return &cfg, nil
	}

	data, err := os.ReadFile(ConfigFile)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Save saves the configuration to disk
func Save(cfg *Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(GeminiDir, 0755); err != nil {
		return err
	}

	return os.WriteFile(ConfigFile, data, 0644)
}

// GoogleAccounts represents google_accounts.json
type GoogleAccounts struct {
	Active string   `json:"active"`
	Old    []string `json:"old"` // Python version uses "old" list for some reason, maybe tracking
}

func LoadAccounts() (*GoogleAccounts, error) {
	if _, err := os.Stat(AccountsJSON); os.IsNotExist(err) {
		return &GoogleAccounts{Active: "", Old: []string{}}, nil
	}

	data, err := os.ReadFile(AccountsJSON)
	if err != nil {
		return nil, err
	}

	var accs GoogleAccounts
	if err := json.Unmarshal(data, &accs); err != nil {
		return nil, err
	}

	return &accs, nil
}

func SaveAccounts(accs *GoogleAccounts) error {
	data, err := json.MarshalIndent(accs, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(AccountsJSON, data, 0644)
}

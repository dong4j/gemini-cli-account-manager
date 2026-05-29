package quota

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"gemini-cli-account-manager/internal/config"
)

const (
	CodeAssistEndpoint   = "https://cloudcode-pa.googleapis.com"
	CodeAssistAPIVersion = "v1internal"
)

var CacheFile = filepath.Join(config.GeminiDir, "quota_cache.json")

// QuotaCache represents cached quota information
type QuotaCache struct {
	Timestamp    string        `json:"timestamp"`
	SessionID    string        `json:"session_id"`
	Buckets      []QuotaBucket `json:"buckets"`
	CacheMinutes int           `json:"cache_minutes"`
}

type OAuthCreds struct {
	AccessToken string `json:"access_token"`
	ExpiryDate  int64  `json:"expiry_date"` // Unix timestamp in ms
}

type LoadCodeAssistRequest struct {
	Metadata Metadata `json:"metadata"`
}

type Metadata struct {
	IDEType    string `json:"ideType"`
	Platform   string `json:"platform"`
	PluginType string `json:"pluginType"`
}

type LoadCodeAssistResponse struct {
	CloudaicompanionProject string `json:"cloudaicompanionProject"`
	CurrentTier             Tier   `json:"currentTier"`
}

type Tier struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type RetrieveUserQuotaRequest struct {
	Project string `json:"project"`
}

type UserQuotaResponse struct {
	Buckets []QuotaBucket `json:"buckets"`
}

type QuotaBucket struct {
	ModelID           string  `json:"modelId"`
	RemainingFraction float64 `json:"remainingFraction"`
	ResetTime         string  `json:"resetTime"`
}

func GetPlatform() string {
	arch := "PLATFORM_UNSPECIFIED"
	switch runtime.GOARCH {
	case "amd64":
		arch = "AMD64"
	case "arm64":
		arch = "ARM64"
	}

	switch runtime.GOOS {
	case "windows":
		if arch == "AMD64" {
			return "WINDOWS_AMD64"
		}
	case "darwin":
		return "DARWIN_" + arch
	case "linux":
		return "LINUX_" + arch
	}
	return "PLATFORM_UNSPECIFIED"
}

func LoadToken() (string, error) {
	data, err := os.ReadFile(config.CredsFile)
	if err != nil {
		return "", err
	}

	var creds OAuthCreds
	if err := json.Unmarshal(data, &creds); err != nil {
		return "", err
	}

	// Check expiry (with some buffer of 10 seconds)
	if creds.ExpiryDate > 0 && time.Now().UnixMilli() > creds.ExpiryDate-10000 {
		return "", fmt.Errorf("token expired")
	}

	return creds.AccessToken, nil
}

// RefreshToken executes gemini command to force token refresh
// Returns the new access token if successful
func RefreshToken() (string, error) {
	// Execute gemini -p "/model list" to force refresh
	cmd := exec.Command("gemini", "-p", "/model list")
	if err := cmd.Run(); err != nil {
		return "", err
	}
	// After refresh, reload the token from file
	return LoadToken()
}

func GetProjectID(token string) (string, error) {
	url := fmt.Sprintf("%s/%s:loadCodeAssist", CodeAssistEndpoint, CodeAssistAPIVersion)

	reqBody, _ := json.Marshal(LoadCodeAssistRequest{
		Metadata: Metadata{
			IDEType:    "GEMINI_CLI",
			Platform:   GetPlatform(),
			PluginType: "GEMINI",
		},
	})

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		return "", fmt.Errorf("unauthorized")
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error: %d", resp.StatusCode)
	}

	var result LoadCodeAssistResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.CloudaicompanionProject, nil
}

func GetUserQuota(token, projectID string) ([]QuotaBucket, error) {
	url := fmt.Sprintf("%s/%s:retrieveUserQuota", CodeAssistEndpoint, CodeAssistAPIVersion)

	reqBody, _ := json.Marshal(RetrieveUserQuotaRequest{
		Project: projectID,
	})

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %d", resp.StatusCode)
	}

	var result UserQuotaResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Buckets, nil
}

// CheckQuota returns the buckets and whether a switch is needed based on the strategy
func CheckQuota(cfg *config.Config) ([]QuotaBucket, bool, string, error) {
	token, err := LoadToken()
	if err != nil {
		// Try refreshing once
		if strings.Contains(err.Error(), "expired") || os.IsNotExist(err) {
			token, err = RefreshToken()
		}
		if err != nil {
			return nil, false, "", fmt.Errorf("failed to get token: %w", err)
		}
	}

	projectID, err := GetProjectID(token)
	if err != nil && strings.Contains(err.Error(), "unauthorized") {
		// Token might have expired, try to refresh
		token, err = RefreshToken()
		if err != nil {
			return nil, false, "", fmt.Errorf("token refresh failed: %w", err)
		}
		projectID, err = GetProjectID(token)
	}
	if err != nil {
		return nil, false, "", fmt.Errorf("failed to get project ID: %w", err)
	}

	buckets, err := GetUserQuota(token, projectID)
	if err != nil {
		return nil, false, "", fmt.Errorf("failed to get quota: %w", err)
	}

	// Use the extracted strategy evaluation
	shouldSwitch, isLow, reason := evaluateQuotaStrategy(buckets, cfg)
	return buckets, shouldSwitch && isLow, reason, nil
}

// LoadCache loads quota information from cache file
// Returns nil if cache is expired or doesn't exist
func LoadCache(sessionID string, cacheMinutes int) *QuotaCache {
	data, err := os.ReadFile(CacheFile)
	if err != nil {
		return nil
	}

	var cache QuotaCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil
	}

	// Parse timestamp
	cacheTime, err := time.Parse(time.RFC3339, cache.Timestamp)
	if err != nil {
		return nil
	}

	// Check if cache is expired
	if time.Since(cacheTime) > time.Duration(cacheMinutes)*time.Minute {
		return nil
	}

	// If session changed, return nil to force refresh
	if cache.SessionID != "" && cache.SessionID != sessionID {
		return nil
	}

	return &cache
}

// SaveCache saves quota information to cache file
func SaveCache(sessionID string, buckets []QuotaBucket, cacheMinutes int) error {
	cache := QuotaCache{
		Timestamp:    time.Now().Format(time.RFC3339),
		SessionID:    sessionID,
		Buckets:      buckets,
		CacheMinutes: cacheMinutes,
	}

	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(CacheFile, data, 0644)
}

// ClearCache removes the cache file
func ClearCache() error {
	if _, err := os.Stat(CacheFile); os.IsNotExist(err) {
		return nil
	}
	return os.Remove(CacheFile)
}

// evaluateQuotaStrategy determines if a quota switch is needed based on strategy
func evaluateQuotaStrategy(buckets []QuotaBucket, cfg *config.Config) (bool, bool, string) {
	threshold := cfg.AutoSwitch.Threshold / 100.0
	strategy := cfg.AutoSwitch.Strategy
	pattern := cfg.AutoSwitch.ModelPattern
	if strategy == "custom" && cfg.AutoSwitch.CustomModelPattern != "" {
		pattern = cfg.AutoSwitch.CustomModelPattern
	}

	var targets []QuotaBucket
	for _, b := range buckets {
		match := false
		switch strategy {
		case "conservative":
			match = true
		case "gemini3-first":
			match = strings.Contains(b.ModelID, "gemini-3")
		case "gemini3.1-pro-only":
			match = strings.Contains(b.ModelID, "gemini-3.1-pro")
		case "gemini3.1-series-only":
			match = strings.Contains(b.ModelID, "gemini-3.1")
		case "custom":
			match = strings.Contains(b.ModelID, pattern)
		}

		if match {
			targets = append(targets, b)
		}
	}

	// Fallback to models_to_check if no targets found by strategy
	if len(targets) == 0 && len(cfg.AutoSwitch.ModelsToCheck) > 0 {
		for _, b := range buckets {
			for _, model := range cfg.AutoSwitch.ModelsToCheck {
				if strings.Contains(b.ModelID, model) && b.RemainingFraction != 0 {
					targets = append(targets, b)
					break
				}
			}
		}
	}

	if len(targets) == 0 {
		return false, false, "No targets"
	}

	allLow := true
	var lowDetails []string
	for _, b := range targets {
		if b.RemainingFraction > threshold {
			allLow = false
			break
		}
		lowDetails = append(lowDetails, fmt.Sprintf("%s: %.1f%%", b.ModelID, b.RemainingFraction*100))
	}

	if allLow {
		return true, true, strings.Join(lowDetails, ", ")
	}

	return true, false, "Quota OK"
}

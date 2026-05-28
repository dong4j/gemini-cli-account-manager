package quota

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"gemini-cli-account-manager/internal/config"
)

const (
	CodeAssistEndpoint   = "https://cloudcode-pa.googleapis.com"
	CodeAssistAPIVersion = "v1internal"
)

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

	// Check expiry (with some buffer)
	if creds.ExpiryDate > 0 && time.Now().UnixMilli() > creds.ExpiryDate-10000 {
		return creds.AccessToken, fmt.Errorf("token expired")
	}

	return creds.AccessToken, nil
}

func RefreshToken() error {
	// Execute gemini -p "/model list" to force refresh
	cmd := exec.Command("gemini", "-p", "/model list")
	return cmd.Run()
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
			_ = RefreshToken()
			token, err = LoadToken()
		}
		if err != nil {
			return nil, false, "", err
		}
	}

	projectID, err := GetProjectID(token)
	if err != nil && strings.Contains(err.Error(), "unauthorized") {
		_ = RefreshToken()
		token, _ = LoadToken()
		projectID, err = GetProjectID(token)
	}
	if err != nil {
		return nil, false, "", err
	}

	buckets, err := GetUserQuota(token, projectID)
	if err != nil {
		return nil, false, "", err
	}

	// Logic for switching based on strategy
	threshold := cfg.AutoSwitch.Threshold / 100.0
	strategy := cfg.AutoSwitch.Strategy
	pattern := cfg.AutoSwitch.ModelPattern
	if strategy == "custom" && cfg.AutoSwitch.CustomModelPattern != "" {
		pattern = cfg.AutoSwitch.CustomModelPattern
	}

	// For simplicity in this initial version, I'll use a direct match or simple logic.
	// We can implement full regex later if needed.
	
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
			// Simple contains for now, or use regexp package
			match = strings.Contains(b.ModelID, pattern)
		}

		if match {
			targets = append(targets, b)
		}
	}

	if len(targets) == 0 {
		return buckets, false, "No targets", nil
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
		return buckets, true, strings.Join(lowDetails, ", "), nil
	}

	return buckets, false, "Quota OK", nil
}

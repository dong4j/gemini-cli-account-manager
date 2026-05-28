package quota

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"

	"gemini-cli-account-manager/internal/auth"
	"gemini-cli-account-manager/internal/config"
)

var (
	RetryFile      = filepath.Join(config.GeminiDir, ".auto_switch_retry_count")
	ErrorStateFile = filepath.Join(config.GeminiDir, ".last_quota_error")
	TokenCacheFile = filepath.Join(config.GeminiDir, "mcp-oauth-tokens-v2.json")
)

var QuotaErrorPatterns = []string{
	`429`,
	`403.*quota`,
	`Resource exhausted`,
	`Quota exceeded`,
	`rate limit`,
	`RESOURCE_EXHAUSTED`,
	`Usage limit reached`,
	`limit reached for all.*models`,
	`Access resets at`,
	`Keep trying.*Stop`,
	`PERMISSION_DENIED.*VALIDATION_REQUIRED`,
	`Please verify your account`,
}

type HookContext struct {
	PromptResponse string `json:"prompt_response"`
}

type HookResponse struct {
	Decision      string `json:"decision,omitempty"`
	SystemMessage string `json:"systemMessage,omitempty"`
}

func RunPreCheckHook(cfg *config.Config) {
	// Pre-check logic from quota_pre_check.py
	// If the last request failed with quota error, we might want to switch now.
	// But usually, AfterAgent handle it. PreCheck can also check current quota.
	
	_, shouldSwitch, reason, err := CheckQuota(cfg)
	if err == nil && shouldSwitch {
		email, err := auth.SwitchNext()
		if err == nil {
			msg := fmt.Sprintf("🔄 [Pre-Check] Quota low (%s). Switched to: %s", reason, email)
			fmt.Fprintf(os.Stderr, "%s\n", msg)
			_ = os.Remove(TokenCacheFile)
		}
	}
	fmt.Println("{}")
}

func RunAutoSwitchHook(cfg *config.Config, input []byte) {
	var ctx HookContext
	if err := json.Unmarshal(input, &ctx); err != nil {
		fmt.Println("{}")
		return
	}

	response := ctx.PromptResponse
	if response == "" {
		fmt.Println("{}")
		return
	}

	isError := false
	for _, p := range QuotaErrorPatterns {
		if matched, _ := regexp.MatchString("(?i)"+p, response); matched {
			isError = true
			break
		}
	}

	if !isError {
		resetRetryCount()
		fmt.Println("{}")
		return
	}

	// Quota error detected
	retryCount := getRetryCount()
	if retryCount >= cfg.AutoSwitch.MaxRetries {
		fmt.Fprintf(os.Stderr, "⚠️ [Auth Manager] Max retries reached.\n")
		resetRetryCount()
		fmt.Println("{}")
		return
	}

	// Perform switch
	newAccount, err := auth.SwitchNext()
	if err != nil {
		fmt.Fprintf(os.Stderr, "⚠️ [Auth Manager] Switch failed: %v\n", err)
		fmt.Println("{}")
		return
	}

	setRetryCount(retryCount + 1)
	_ = os.Remove(TokenCacheFile)

	msg := fmt.Sprintf("🔄 Quota exhausted. Switched to: %s. Retrying... (%d/%d)", 
		newAccount, retryCount+1, cfg.AutoSwitch.MaxRetries)
	
	resp := HookResponse{
		Decision:      "retry",
		SystemMessage: msg,
	}
	data, _ := json.Marshal(resp)
	fmt.Println(string(data))

	// Auto-restart logic
	if cfg.AutoSwitch.AutoRestart {
		// In Go, we can start a detached process to handle the restart
		// For simplicity, we'll just call our own binary with a restart command
		execPath, _ := os.Executable()
		ppid := os.Getppid()
		cmd := exec.Command(execPath, "restart", fmt.Sprintf("%d", ppid))
		_ = cmd.Start()
	}
}

func getRetryCount() int {
	data, err := os.ReadFile(RetryFile)
	if err != nil {
		return 0
	}
	var count int
	fmt.Sscanf(string(data), "%d", &count)
	return count
}

func setRetryCount(count int) {
	_ = os.WriteFile(RetryFile, []byte(fmt.Sprintf("%d", count)), 0644)
}

func resetRetryCount() {
	_ = os.Remove(RetryFile)
}

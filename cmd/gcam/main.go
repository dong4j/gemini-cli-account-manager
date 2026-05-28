package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"gemini-cli-account-manager/internal/auth"
	"gemini-cli-account-manager/internal/config"
	"gemini-cli-account-manager/internal/quota"
	"gemini-cli-account-manager/internal/ui"
	"gemini-cli-account-manager/internal/utils"
)

func main() {
	if len(os.Args) < 2 {
		listStatus()
		return
	}

	command := strings.ToLower(os.Args[1])
	args := os.Args[2:]

	cfg, err := config.Load()
	if err != nil {
		ui.Error("Failed to load config: %v", err)
		os.Exit(1)
	}

	switch command {
	case "next":
		email, err := auth.SwitchNext()
		if err != nil {
			ui.Error("Switch failed: %v", err)
			os.Exit(1)
		}
		ui.Success("Switched to next account: %s", email)

	case "list", "-l":
		listStatus()

	case "pool":
		handlePool(cfg, args)

	case "quota", "stats":
		showQuota(cfg)

	case "menu":
		interactiveMenu(cfg)

	case "strategy":
		handleStrategy(cfg, args)

	case "config":
		handleConfig(cfg, args)

	case "hook":
		if len(args) > 0 {
			switch args[0] {
			case "pre-check":
				quota.RunPreCheckHook(cfg)
			case "auto-switch":
				input, _ := io.ReadAll(os.Stdin)
				quota.RunAutoSwitchHook(cfg, input)
			}
		}

	case "install":
		execPath, _ := os.Executable()
		if err := utils.InstallHooks(execPath); err != nil {
			ui.Error("Installation failed: %v", err)
			os.Exit(1)
		}
		ui.Success("Installation complete!")

	case "uninstall":
		if err := utils.UninstallHooks(); err != nil {
			ui.Error("Uninstallation failed: %v", err)
			os.Exit(1)
		}
		ui.Success("Uninstallation complete!")

	case "restart":
		if len(args) > 0 {
			var ppid int
			fmt.Sscanf(args[0], "%d", &ppid)
			utils.RestartGemini(ppid, 3*time.Second)
			// Wait a bit to let the goroutine start
			time.Sleep(500 * time.Millisecond)
		}

	case "help", "-h", "--help":
		showHelp()

	default:
		// Treat as account identifier (index or email)
		email, err := auth.FastSwitch(command)
		if err != nil {
			ui.Error("Switch failed: %v", err)
			os.Exit(1)
		}
		ui.Success("Switched to: %s", email)
	}
}

func interactiveMenu(cfg *config.Config) {
	items := []ui.MenuItem{
		{Label: "Switch Account", Action: func() {
			listStatus()
			fmt.Printf("Enter index or email: ")
			var target string
			fmt.Scanln(&target)
			if target != "" {
				email, err := auth.FastSwitch(target)
				if err != nil {
					ui.Error("Switch failed: %v", err)
				} else {
					ui.Success("Switched to: %s", email)
				}
			}
		}},
		{Label: "Switch to Next", Action: func() {
			email, err := auth.SwitchNext()
			if err != nil {
				ui.Error("Switch failed: %v", err)
			} else {
				ui.Success("Switched to: %s", email)
			}
		}},
		{Label: "Change Strategy", Action: func() {
			handleStrategy(cfg, []string{})
		}},
		{Label: "Manage Pool", Action: func() {
			handlePool(cfg, []string{})
		}},
		{Label: "Show Quota", Action: func() {
			showQuota(cfg)
		}},
	}
	ui.RenderMenu("Gemini CLI Auth Manager", items)
}

func handlePool(cfg *config.Config, args []string) {
	if len(args) > 0 {
		switch args[0] {
		case "login":
			if err := auth.Login(cfg); err != nil {
				ui.Error("Login failed: %v", err)
			}
		case "list":
			listStatus()
		}
	} else {
		listStatus()
	}
}

func handleStrategy(cfg *config.Config, args []string) {
	if len(args) > 0 {
		cfg.AutoSwitch.Strategy = args[0]
		_ = config.Save(cfg)
		ui.Success("Strategy updated to: %s", args[0])
		return
	}

	fmt.Printf("Current Strategy: %s\n", cfg.AutoSwitch.Strategy)
	fmt.Println("Available Strategies:")
	fmt.Println("  1. conservative")
	fmt.Println("  2. gemini3-first")
	fmt.Println("  3. gemini3.1-pro-only")
	fmt.Println("  4. gemini3.1-series-only")
	fmt.Println("  5. custom")
	fmt.Printf("\nEnter choice [1-5]: ")
	
	var choice string
	fmt.Scanln(&choice)
	
	m := map[string]string{
		"1": "conservative",
		"2": "gemini3-first",
		"3": "gemini3.1-pro-only",
		"4": "gemini3.1-series-only",
		"5": "custom",
	}
	
	if val, ok := m[choice]; ok {
		cfg.AutoSwitch.Strategy = val
		_ = config.Save(cfg)
		ui.Success("Strategy set to: %s", val)
	}
}

func handleConfig(cfg *config.Config, args []string) {
	// Simple config viewer/setter
	if len(args) >= 2 {
		// e.g. gcam config threshold 20
		key := args[0]
		val := args[1]
		switch key {
		case "threshold":
			fmt.Sscanf(val, "%f", &cfg.AutoSwitch.Threshold)
		case "enabled":
			cfg.AutoSwitch.Enabled = (val == "true")
		}
		_ = config.Save(cfg)
		ui.Success("Config updated")
	} else {
		data, _ := json.MarshalIndent(cfg, "", "  ")
		fmt.Println(string(data))
	}
}

func listStatus() {
	config.InitPaths() // Ensure paths are fresh
	fmt.Printf("Debug: GeminiDir = %s\n", config.GeminiDir)
	profiles, _ := auth.GetProfiles()
	accs, _ := config.LoadAccounts()

	ui.Heading("Gemini CLI Account Manager")
	fmt.Printf("%sActive Account:%s %s\n", ui.Bold, ui.Reset, accs.Active)
	
	fmt.Printf("\n%sAccounts in Pool:%s\n", ui.Bold, ui.Reset)
	for i, p := range profiles {
		marker := "[ ]"
		if p == accs.Active {
			marker = ui.Green + "[*]" + ui.Reset
		}
		fmt.Printf("  %02d. %s %s\n", i+1, marker, p)
	}
	fmt.Println()
}

func showQuota(cfg *config.Config) {
	ui.Info("Fetching quota information...")
	buckets, _, reason, err := quota.CheckQuota(cfg)
	if err != nil {
		ui.Error("Failed to get quota: %v", err)
		return
	}

	fmt.Printf("\n%-30s %-15s %s\n", "Model", "Remaining", "Status")
	fmt.Println(strings.Repeat("-", 60))
	for _, b := range buckets {
		status := "🟢"
		if b.RemainingFraction < 0.1 {
			status = "🔴"
		} else if b.RemainingFraction < 0.3 {
			status = "🟡"
		}
		fmt.Printf("%-30s %-15.1f%% %s\n", b.ModelID, b.RemainingFraction*100, status)
	}
	fmt.Printf("\nResult: %s\n", reason)
}

func showHelp() {
	fmt.Println("Usage:")
	fmt.Println("  gcam               List accounts")
	fmt.Println("  gcam <n|email>     Switch to account")
	fmt.Println("  gcam next          Switch to next account")
	fmt.Println("  gcam pool login    Add new account via OAuth")
	fmt.Println("  gcam quota         Show current quota")
}

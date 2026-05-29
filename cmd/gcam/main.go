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
	"gemini-cli-account-manager/internal/i18n"
	"gemini-cli-account-manager/internal/quota"
	"gemini-cli-account-manager/internal/ui"
	"gemini-cli-account-manager/internal/utils"
)

// 版本号（编译时通过 ldflags 注入）
var version = "dev"

func main() {
	// 处理版本参数
	if len(os.Args) == 2 {
		arg := strings.ToLower(os.Args[1])
		if arg == "-v" || arg == "--version" {
			fmt.Println("gcam version", version)
			return
		}
	}

	if len(os.Args) < 2 {
		listStatus()
		return
	}

	command := strings.ToLower(os.Args[1])
	args := os.Args[2:]

	cfg, err := config.Load()
	if err != nil {
		ui.Error(i18n.T("error")+": %v", err)
		os.Exit(1)
	}

	switch command {
	case "next":
		email, err := auth.SwitchNext()
		if err != nil {
			ui.Error(i18n.T("switch_failed"), err)
			os.Exit(1)
		}
		ui.Success(i18n.T("switched_to"), email)

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
			ui.Error(i18n.T("error")+": %v", err)
			os.Exit(1)
		}
		ui.Success(i18n.T("install_complete"))

	case "uninstall":
		keepAccounts := false
		force := false
		for _, arg := range args {
			if arg == "--keep-accounts" || arg == "-k" {
				keepAccounts = true
			}
			if arg == "--force" || arg == "-f" || arg == "-y" {
				force = true
			}
		}

		if !force {
			fmt.Printf("\n  %s\n", i18n.T("uninstall_title"))
			fmt.Println("  " + strings.Repeat("-", 40))
			fmt.Printf("  %s:\n", i18n.T("uninstall_files"))
			for _, f := range utils.GetUninstallFiles(keepAccounts) {
				if _, err := os.Stat(f); err == nil {
					fmt.Printf("    %s[remove]%s %s\n", ui.Red, ui.Reset, f)
				} else {
					fmt.Printf("    %s[skip]%s    %s (not found)\n", ui.Dim, ui.Reset, f)
				}
			}
			if !keepAccounts {
				fmt.Println("  " + ui.Yellow + "Warning:" + ui.Reset + " " + i18n.T("uninstall_data"))
			}
			fmt.Println()
			fmt.Print("  " + i18n.T("uninstall_proceed"))
			var confirm string
			fmt.Scanln(&confirm)
			if confirm != "y" && confirm != "Y" && confirm != "yes" {
				fmt.Println("  " + i18n.T("uninstall_cancelled"))
				return
			}
		}

		if err := utils.UninstallHooks(utils.UninstallOptions{KeepAccounts: keepAccounts}); err != nil {
			ui.Error(i18n.T("error")+": %v", err)
			os.Exit(1)
		}
		ui.Success(i18n.T("uninstall_complete"))

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
			ui.Error(i18n.T("switch_failed"), err)
			os.Exit(1)
		}
		ui.Success(i18n.T("switched_to"), email)
	}
}

func interactiveMenu(cfg *config.Config) {
	items := []ui.MenuItem{
		{Label: i18n.T("switch_account"), Action: func() {
			listStatus()
			fmt.Print(i18n.T("enter_idx_email") + ": ")
			var target string
			fmt.Scanln(&target)
			if target != "" {
				email, err := auth.FastSwitch(target)
				if err != nil {
					ui.Error(i18n.T("switch_failed"), err)
				} else {
					ui.Success(i18n.T("switched_to"), email)
				}
			}
		}},
		{Label: i18n.T("switch_next"), Action: func() {
			email, err := auth.SwitchNext()
			if err != nil {
				ui.Error(i18n.T("switch_failed"), err)
			} else {
				ui.Success(i18n.T("switched_to"), email)
			}
		}},
		{Label: i18n.T("change_strategy"), Action: func() {
			handleStrategy(cfg, []string{})
		}},
		{Label: i18n.T("manage_pool"), Action: func() {
			handlePool(cfg, []string{})
		}},
		{Label: i18n.T("usage"), Action: func() {
			showQuota(cfg)
		}},
	}
	ui.RenderMenu(i18n.T("title"), items)
}

func handlePool(cfg *config.Config, args []string) {
	if len(args) > 0 {
		switch args[0] {
		case "login":
			if err := auth.Login(cfg); err != nil {
				ui.Error(i18n.T("login_failed")+": %v", err)
			}
		case "list":
			listStatus()
		case "remove", "delete", "rm":
			if len(args) < 2 {
				ui.Error("Usage: gcam pool remove <index|email>")
				return
			}
			if err := auth.RemoveAccount(args[1]); err != nil {
				ui.Error(i18n.T("error")+": %v", err)
			} else {
				ui.Success(i18n.T("updated"))
			}
		case "import":
			if len(args) < 3 {
				ui.Error("Usage: gcam pool import <path_to_oauth_creds.json> <email>")
				return
			}
			if err := auth.ImportAccount(args[1], args[2]); err != nil {
				ui.Error(i18n.T("error")+": %v", err)
			} else {
				ui.Success(i18n.T("success"))
			}
		default:
			ui.Error(i18n.T("error")+": %s", args[0])
		}
	} else {
		listStatus()
	}
}

func handleStrategy(cfg *config.Config, args []string) {
	if len(args) > 0 {
		// Apply alias mapping
		strategy := args[0]
		switch strategy {
		case "pro":
			strategy = "gemini3.1-pro-only"
		case "series":
			strategy = "gemini3.1-series-only"
		case "3", "gemini3":
			strategy = "gemini3-first"
		case "3.1", "gemini3.1":
			strategy = "gemini3.1-series-only"
		}

		// Validate strategy
		validStrategies := []string{"conservative", "gemini3-first", "gemini3.1-pro-only", "gemini3.1-series-only", "custom"}
		isValid := false
		for _, s := range validStrategies {
			if s == strategy {
				isValid = true
				break
			}
		}
		if !isValid {
			ui.Error(i18n.T("strategy_invalid")+": %s", strategy)
			return
		}

		cfg.AutoSwitch.Strategy = strategy
		_ = config.Save(cfg)
		ui.Success(i18n.T("strategy_updated")+": %s", strategy)
		return
	}

	fmt.Printf("%s: %s\n", i18n.T("strategy"), cfg.AutoSwitch.Strategy)
	fmt.Printf("%s:\n", i18n.T("select_strategy"))
	fmt.Printf("  1. conservative      - %s\n", i18n.T("strategy_desc_conservative"))
	fmt.Printf("  2. gemini3-first    - %s\n", i18n.T("strategy_desc_gemini3_first"))
	fmt.Printf("  3. gemini3.1-pro-only - %s\n", i18n.T("strategy_desc_gemini31_pro_only"))
	fmt.Printf("  4. gemini3.1-series-only - %s\n", i18n.T("strategy_desc_gemini31_series"))
	fmt.Printf("  5. custom          - %s\n", i18n.T("strategy_desc_custom"))
	fmt.Println()
	fmt.Printf("\n%s: ", i18n.T("enter_choice"))

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
	} else {
		cfg.AutoSwitch.Strategy = choice
	}

	_ = config.Save(cfg)
	ui.Success(i18n.T("strategy_updated")+": %s", cfg.AutoSwitch.Strategy)
}

func handleConfig(cfg *config.Config, args []string) {
	if len(args) >= 2 {
		key := args[0]
		val := args[1]
		switch key {
		case "threshold":
			fmt.Sscanf(val, "%f", &cfg.AutoSwitch.Threshold)
		case "enabled":
			cfg.AutoSwitch.Enabled = (val == "true")
		case "language", "lang":
			cfg.Language = val
		case "models":
			cfg.AutoSwitch.ModelsToCheck = strings.Split(val, ",")
		}
		_ = config.Save(cfg)
		ui.Success(i18n.T("updated"))
	} else {
		data, _ := json.MarshalIndent(cfg, "", "  ")
		fmt.Println(string(data))
	}
}

func listStatus() {
	config.InitPaths()
	profiles, _ := auth.GetProfiles()
	accs, _ := config.LoadAccounts()

	ui.Heading(i18n.T("title"))
	fmt.Printf("%s%s:%s %s\n", ui.Bold, i18n.T("active_account"), ui.Reset, accs.Active)
	
	fmt.Printf("\n%s%s:%s\n", ui.Bold, i18n.T("accounts"), ui.Reset)
	if len(profiles) == 0 {
		fmt.Printf("  %s(%s)%s\n", ui.Yellow, i18n.T("no_profiles"), ui.Reset)
	}
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
	ui.Info(i18n.T("fetching_quota"))
	buckets, _, reason, err := quota.CheckQuota(cfg)
	if err != nil {
		ui.Error(i18n.T("quota_check_failed"), err)
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

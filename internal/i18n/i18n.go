package i18n

import (
	"gemini-cli-account-manager/internal/config"
)

// Language strings
var Lang = map[string]map[string]string{
	"en": {
		// Status
		"title":              "GEMINI CLI ACCOUNT MANAGER",
		"subtitle":           "Fast Switcher + Auto Rotation",
		"status":             "STATUS",
		"active":             "ACTIVE",
		"auto":               "AUTO",
		"enabled":            "Enabled",
		"disabled":           "Disabled",
		"accounts":          "ACCOUNTS",
		"usage":              "USAGE",
		"current_status":     "Current Status",
		"active_account":     "Active Account",
		"auto_switch":        "Auto-Switch",
		"strategy":           "Strategy",
		"threshold":          "Threshold",
		"menu":               "Menu",
		"exit":               "Exit",
		"goodbye":            "Goodbye!",
		"enter_choice":       "Enter choice",
		"switch_account":      "Switch Account",
		"switch_next":        "Switch to Next Account",
		"change_strategy":    "Change Strategy",
		"select_strategy":    "Select Strategy",
		"strategy_desc_conservative": "Monitor ALL models (Switch if ANY runs out)",
		"strategy_desc_gemini3_first":     "Monitor Gemini 3.0+ series (gemini-3.*)",
		"strategy_desc_gemini31_pro_only": "Monitor Gemini 3.1 Pro Only (gemini-3.1-pro.*)",
		"strategy_desc_gemini31_series":   "Monitor Gemini 3.1 Series (gemini-3.1.*)",
		"strategy_desc_custom":            "Custom regex pattern",
		"enter_custom_pattern": "Enter custom regex pattern (e.g. gemini-2.5.*): ",
		"invalid_regex":       "Invalid regex pattern. Please try again.",
		"strategy_updated":    "Strategy set to",
		"strategy_invalid":     "Invalid strategy",
		"manage_pool":         "Manage Account Pool",
		"toggle_auto":         "Toggle Auto-Switch",
		"config_details":      "View Detailed Config",
		"current_val":         "Current value",
		"new_val":             "Enter new value (empty to cancel)",
		"updated":             "Updated",
		"invalid_val":         "Invalid value",
		"error":               "Error",
		"success":             "Success",
		"account_pool":        "Account Pool Management",
		"pool_list":           "List Accounts",
		"pool_login":          "Add Account (OAuth Login)",
		"pool_remove":         "Remove Account",
		"pool_import":         "Import Credentials",
		"pool_back":           "Back to Main Menu",
		"enter_idx_email":     "Enter account index or email",
		"confirm_remove":      "Are you sure you want to remove account",
		"login_browser":        "Opening browser for OAuth login...",
		"login_success":       "Successfully added account",
		"login_failed":        "Login failed",
		"file_not_found":      "File not found",
		"import_success":      "Imported credentials from",
		"restarting":          "Restarting Gemini CLI...",
		"pool_overview":       "Account Pool Overview",
		"no_profiles":         "No accounts found in pool",
		"total":               "Total",
		"options":             "Options",
		"pool_mgmt":           "Account Pool Management",
		"remove_account":       "Remove Account",
		"import_creds":        "Import Credentials",
		"enter_remove_num":    "Enter account index to remove",
		"enter_path":          "Enter path to credentials file",
		"back":                "Back",
		"standby":             "Standby",
		"install_complete":    "Installation complete!",
		"fetching_quota":      "Fetching quota information...",

		// Uninstall
		"uninstall_title":     "UNINSTALL",
		"uninstall_files":     "Files to remove",
		"uninstall_data":      "Account data to remove",
		"uninstall_also":      "Also cleaned",
		"uninstall_proceed":   "Proceed with uninstall? (y/N): ",
		"uninstall_cancelled": "Cancelled",
		"uninstall_removing":  "Removing files...",
		"uninstall_removed":   "Removed %d file(s)",
		"uninstall_complete":   "Uninstall Complete!",
		"uninstall_restart":   "Please restart your terminal for PATH changes to take effect",

		// Messages
		"switched_to":         "Switched to %s",
		"already_using":      "Already using %s",
		"no_profiles_found":  "No profiles found",
		"index_out_of_range":  "Index %s out of range (1-%d)",
		"account_not_found":   "Account not found: %s",
		"missing_creds":       "Missing credentials for: %s",
		"switch_failed":       "Switch failed: %v",
		"max_retries":         "Max retries (%d) reached. All accounts may be exhausted.",
		"quota_low":           "Quota low (%s). Switched to: %s",
		"cache_cleared":       "Cleared token cache to force reload",
		"using_cached_quota":  "Using cached quota (expires in %d min)",
		"quota_check_failed":  "Quota check failed: %v",
	},
	"cn": {
		// Status
		"title":              "GEMINI CLI 账号管理器",
		"subtitle":           "极速切换 + 自动轮换",
		"status":             "状态",
		"active":             "正在使用",
		"auto":               "自动切换",
		"enabled":            "已启用",
		"disabled":           "已禁用",
		"accounts":           "号池",
		"usage":              "用法",
		"current_status":     "当前状态",
		"active_account":     "当前账号",
		"auto_switch":        "配额自动切换",
		"strategy":           "切换策略",
		"threshold":          "切换阈值",
		"menu":               "主菜单",
		"exit":               "退出",
		"goodbye":            "再见！",
		"enter_choice":       "请输入选项",
		"switch_account":      "切换账号",
		"switch_next":        "切换到下一个账号",
		"change_strategy":    "更改轮换策略",
		"select_strategy":    "选择轮换策略",
		"strategy_desc_conservative": "保守模式：监控所有模型（任一耗尽即切）",
		"strategy_desc_gemini3_first":     "Gemini 3.0+ 优先 (匹配 gemini-3.*)",
		"strategy_desc_gemini31_pro_only": "仅监控 Gemini 3.1 Pro (匹配 gemini-3.1-pro.*)",
		"strategy_desc_gemini31_series":   "监控 Gemini 3.1 全系列 (匹配 gemini-3.1.*)",
		"strategy_desc_custom":            "自定义正则表达式模式",
		"enter_custom_pattern": "请输入自定义正则表达式 (例如 gemini-2.5.*): ",
		"invalid_regex":       "无效的正则表达式，请重试。",
		"strategy_updated":    "策略已设置为",
		"strategy_invalid":    "无效的策略名称",
		"manage_pool":         "管理账号池",
		"toggle_auto":         "开启/关闭自动切换",
		"config_details":      "查看详细配置",
		"current_val":         "当前值",
		"new_val":             "请输入新值 (留空取消)",
		"updated":             "已更新",
		"invalid_val":         "无效的值",
		"error":               "错误",
		"success":             "成功",
		"account_pool":        "账号池管理",
		"pool_list":           "查看账号列表",
		"pool_login":          "添加账号 (OAuth 登录)",
		"pool_remove":         "删除账号",
		"pool_import":         "导入凭据文件",
		"pool_back":           "返回主菜单",
		"enter_idx_email":     "请输入账号序号或邮箱",
		"confirm_remove":      "确定要删除账号吗",
		"login_browser":        "正在打开浏览器进行登录...",
		"login_success":       "成功添加账号",
		"login_failed":        "登录失败",
		"file_not_found":      "找不到文件",
		"import_success":      "成功导入凭据",
		"restarting":          "正在重启 Gemini CLI...",
		"pool_overview":       "账号池概览",
		"no_profiles":         "号池中没有账号",
		"total":               "总计",
		"options":             "选项",
		"pool_mgmt":           "号池管理",
		"remove_account":       "删除账号",
		"import_creds":        "导入凭据",
		"enter_remove_num":    "请输入要删除的账号序号",
		"enter_path":          "请输入凭据文件路径",
		"back":                "返回",
		"standby":             "待命",
		"install_complete":    "安装完成！",
		"fetching_quota":      "正在获取配额信息...",

		// Uninstall
		"uninstall_title":     "卸载",
		"uninstall_files":     "待删除文件",
		"uninstall_data":      "待删除账号数据",
		"uninstall_also":      "同时清理",
		"uninstall_proceed":   "确认执行卸载？(y/N): ",
		"uninstall_cancelled": "已取消",
		"uninstall_removing":  "正在删除文件...",
		"uninstall_removed":   "已删除 %d 个文件",
		"uninstall_complete":  "卸载完成！",
		"uninstall_restart":   "请重启终端以使 PATH 变更生效",

		// Messages
		"switched_to":         "已切换到 %s",
		"already_using":      "已在使用 %s",
		"no_profiles_found":  "没有找到账号配置",
		"index_out_of_range":  "序号 %s 超出范围 (1-%d)",
		"account_not_found":   "找不到账号: %s",
		"missing_creds":       "缺少账号凭据: %s",
		"switch_failed":       "切换失败: %v",
		"max_retries":         "已达到最大重试次数 (%d)。所有账号可能都已耗尽。",
		"quota_low":           "配额低 (%s)。已切换到: %s",
		"cache_cleared":       "已清理 token 缓存",
		"using_cached_quota":  "使用缓存配额 (%d 分钟过期)",
		"quota_check_failed":  "配额检查失败: %v",
	},
}

// GetLang returns the current language from config
func GetLang() string {
	cfg, err := config.Load()
	if err != nil {
		return "en"
	}
	return cfg.Language
}

// T returns the translated string for the given key
func T(key string) string {
	lang := GetLang()
	if langMap, ok := Lang[lang]; ok {
		if val, ok := langMap[key]; ok {
			return val
		}
	}
	// Fallback to English
	if val, ok := Lang["en"][key]; ok {
		return val
	}
	return key
}

// GetStrategyDesc returns the description for a strategy
func GetStrategyDesc(strategy string) string {
	key := "strategy_desc_" + strategy
	return T(key)
}

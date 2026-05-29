# Gemini CLI Account Manager (gcam) Go 版本对比分析报告

> 分析日期: 2026-05-29
> Go 版本 vs Python 参考版本 (Gemini-CLI-Auth-Manager v2.6)

---

## 一、已修复的问题

### 1. TOML 格式错误 (install.go)

- **问题**: 生成的 `gcam.toml` 使用了不存在的 `command` 字段
- **原因**: 错误参考了旧版文档或实验性 API
- **修复**: 改为 `prompt` + `!{...}` 语法，与 Python 版本保持一致

**修复前:**
```toml
command = ["/path/to/gcam"]
```

**修复后:**
```toml
prompt = "!{/path/to/gcam {{args}}}"
```

---

## 二、功能缺失清单 (Go 版本 vs Python 版本)

### 1. 账号池管理 (`handle_pool`) - 功能不完整

| 功能 | Python | Go | 说明 |
|------|--------|-----|------|
| `pool login` | ✅ | ✅ | OAuth 登录 |
| `pool list` | ✅ | ✅ | 列出账号 |
| `pool remove` | ✅ | ❌ | 删除账号 |
| `pool import` | ✅ | ❌ | 导入凭据文件 |

**Python 实现** (`gemini_cli_auth_manager.py:642-792`):
- `remove_account()` - 删除账号，支持序号/邮箱确认删除
- `import_account()` - 从 `oauth_creds.json` 导入账号

**Go 实现** (`main.go:146-159`):
```go
func handlePool(cfg *config.Config, args []string) {
    if len(args) > 0 {
        switch args[0] {
        case "login":
            // ...
        case "list":
            // ...
        }
    }
    // 缺少 remove 和 import 分支
}
```

---

### 2. 策略管理 (`handle_strategy`) - 缺少别名支持

**Python 实现** 支持策略别名:
```python
# Alias support
if strategy == "pro": strategy = "gemini3.1-pro-only"
if strategy == "series": strategy = "gemini3.1-series-only"
```

**Go 实现** (`main.go:161-194`):
```go
// 没有别名支持
```

---

### 3. 配置管理 (`handle_config`) - 配置项不完整

**Python 支持的配置项**:
```python
valid_keys = ["enabled", "strategy", "model_pattern", "threshold",
              "max_retries", "notify_on_switch", "auto_restart",
              "cache_minutes", "models_to_check"]
```

**Go 支持的配置项** (`config/config.go:50-59`):
```go
type AutoSwitch struct {
    Enabled            bool
    Strategy           string
    ModelPattern       string
    CustomModelPattern string
    Threshold          float64
    MaxRetries         int
    NotifyOnSwitch     bool
    AutoRestart        bool
    CacheMinutes       int
}
```

**缺失**: `models_to_check` 列表配置

---

### 4. 卸载功能 (`uninstall`)

| 功能 | Python | Go |
|------|--------|-----|
| 清理文件 | ✅ | ❌ |
| 清理账号数据 | ✅ | ❌ |
| 清理 settings.json hooks | ✅ | ❌ |
| 清理环境变量 | ✅ | ❌ |
| 确认提示 | ✅ | ❌ |
| `--keep-accounts` 选项 | ✅ | ❌ |

**Python 实现** (`gemini_cli_auth_manager.py:928-1090`):
- `_clean_settings_json()` - 清理 hooks
- `_remove_env_var()` - 清理环境变量
- 完整的文件列表管理

**Go 实现**: 仅 `utils.UninstallHooks()` 清理 hooks 和 toml

---

### 5. BeforeAgent Hook (`quota_pre_check.py`)

**Python 版本完整功能**:
- 缓存机制 (`cache_minutes`)
- Session ID 检测 (新会话强制刷新)
- `models_to_check` 列表支持
- 详细的日志输出

**Go 版本** (`quota/hook.go:45-60`):
```go
func RunPreCheckHook(cfg *config.Config) {
    _, shouldSwitch, reason, err := CheckQuota(cfg)
    if err == nil && shouldSwitch {
        email, err := auth.SwitchNext()
        // ...
    }
    fmt.Println("{}")
}
// 没有缓存机制，没有 session 检测
```

---

### 6. Token 刷新逻辑

**Python 版本** (`quota_api_client.py:89-123`):
```python
if e.response.status_code == 401:
    print(f"⚠️  Token expired (401). Attempting to refresh via Gemini CLI...")
    try:
        subprocess.run(
            ["gemini", "-p", "/model list"],
            check=True,
            stdout=subprocess.DEVNULL,
            stderr=subprocess.DEVNULL,
            shell=is_windows
        )
        # Reload token
        new_token = load_oauth_token()
        if new_token != access_token:
            print("✅ Token refreshed. Retrying...")
            # 重试 API 调用
```

**Go 版本** (`quota/quota.go:102-106`):
```go
func RefreshToken() error {
    cmd := exec.Command("gemini", "-p", "/model list")
    return cmd.Run()
    // 没有重试逻辑，没有 token 重新加载
}
```

---

### 7. OAuth 登录细节差异

**Python 版本** (`gemini_cli_auth_manager.py:820-926`):
- 使用 `User-Agent: vscode/1.92.2` (可能有特殊处理)
- 完整的错误处理和响应解析
- 成功页面有 confetti 动画

**Go 版本** (`auth/login.go`):
- 没有特殊的 User-Agent
- HTML 成功页面有 confetti（很好！）

---

### 8. 错误状态持久化

**Python 版本**:
- `ERROR_STATE_FILE` - BeforeAgent 使用此文件在 CLI 崩溃时保留状态
- 重试计数持久化到文件

**Go 版本**: 无此机制

---

## 三、Bug 和问题

### 1. Token 过期检查不正确 (`quota/quota.go:95-97`)

```go
if creds.ExpiryDate > 0 && time.Now().UnixMilli() > creds.ExpiryDate-10000 {
    return creds.AccessToken, fmt.Errorf("token expired")
}
```

**问题**: 如果 token 已过期，仍然返回 `access_token`，只是返回一个 error。这可能导致后续 API 调用失败。

**应该**: token 过期时应该尝试刷新，或返回错误让调用者知道需要刷新。

---

### 2. macOS 重启命令不正确 (`utils/restart.go:33`)

```go
case "darwin":
    cmd = exec.Command("open", "-a", "Terminal", "gemini")
```

**问题**: `open -a Terminal gemini` 不会按预期工作，应该直接执行 `gemini`。

---

### 3. 缓存清理逻辑 (`auth/auth.go:397-408`)

Python 版本在切换账号后会清理 `mcp-oauth-tokens-v2.json`，但 Go 版本没有完全实现这个逻辑。

---

## 四、新功能建议

### 高优先级

#### 1. 完整的账号池管理

```bash
gcam pool remove <n|email>  # 删除账号
gcam pool import <path>     # 导入凭据
```

**涉及文件**:
- `cmd/gcam/main.go` - 添加 `pool remove` 和 `pool import` 分支
- `internal/auth/auth.go` - 添加 `RemoveAccount()` 和 `ImportAccount()` 函数

---

#### 2. 完整的卸载功能

```bash
gcam uninstall [--keep-accounts]  # 保留账号数据选项
```

**涉及文件**:
- `cmd/gcam/main.go` - 完善 `uninstall` 命令
- `internal/utils/install.go` - 完善 `UninstallHooks()` 函数

**需要清理**:
- `~/.gemini/gcam` (Unix launcher)
- `~/.gemini/gcam.bat` (Windows launcher)
- `~/.gemini/auth_config.json` (可选，保留账号池时跳过)
- `~/.gemini/auth_profiles/` (可选)
- `~/.gemini/google_accounts.json`
- `settings.json` 中的 hooks
- 环境变量 `GEMINI_FORCE_FILE_STORAGE`

---

#### 3. BeforeAgent Hook 缓存机制

**缓存文件** `~/.gemini/quota_cache.json`:
```json
{
    "timestamp": "2026-05-29T10:00:00",
    "session_id": "abc123",
    "buckets": [...],
    "cache_minutes": 3
}
```

**涉及文件**:
- `internal/quota/quota.go` - 添加 `LoadCache()`, `SaveCache()` 函数
- `internal/quota/hook.go` - 在 `RunPreCheckHook` 中使用缓存

---

#### 4. models_to_check 配置支持

**config/auth_config.json**:
```json
{
    "auto_switch": {
        "models_to_check": [
            "gemini-3.1-pro-preview",
            "gemini-3.1-flash-lite-preview"
        ]
    }
}
```

**涉及文件**:
- `internal/config/config.go` - 添加 `ModelsToCheck []string` 字段
- `internal/quota/quota.go` - 在 `CheckQuota` 中使用此列表

---

### 中优先级

#### 5. 双语支持 (i18n)

Python 版本已经有完整的中英文支持。Go 版本可以增加：

**涉及文件**:
- `internal/i18n/i18n.go` (新建)

```go
var LANG = map[string]map[string]string{
    "en": {
        "active": "ACTIVE",
        "standby": "Standby",
        // ...
    },
    "cn": {
        "active": "正在使用",
        "standby": "待命",
        // ...
    },
}

func t(key string) string {
    lang := getLang()
    return LANG[lang][key]
}
```

---

#### 6. Token 自动刷新机制

**当前问题**: `LoadToken()` 在 token 过期时返回错误，但没有自动刷新

**改进方案**:
1. `LoadToken()` 尝试加载 token
2. 如果过期，调用 `RefreshToken()`
3. 重新加载 token
4. 如果仍然失败，返回错误

**涉及文件**:
- `internal/quota/quota.go` - 修改 `CheckQuota()` 逻辑

---

#### 7. 策略别名支持

```bash
gcam strategy pro        # 等于 gemini3.1-pro-only
gcam strategy series     # 等于 gemini3.1-series-only
```

**涉及文件**:
- `cmd/gcam/main.go` - 在 `handleStrategy` 中添加别名解析

---


---

## 五、文件对照表

| Python 文件 | Go 文件 | 功能 |
|-------------|---------|------|
| `gemini_cli_auth_manager.py` | `cmd/gcam/main.go` | 主入口/命令路由 |
| `gemini_cli_auth_manager.py` | `internal/auth/auth.go` | 账号切换核心逻辑 |
| `gemini_cli_auth_manager.py` | `internal/auth/login.go` | OAuth 登录 |
| `quota_api_client.py` | `internal/quota/quota.go` | 配额 API 调用 |
| `quota_pre_check.py` | `internal/quota/hook.go` | BeforeAgent Hook |
| `quota_auto_switch.py` | `internal/quota/hook.go` | AfterAgent Hook |
| `install.py` | `internal/utils/install.go` | 安装/卸载逻辑 |
| `restart_helper.py` | `internal/utils/restart.go` | 重启辅助 |
| - | `internal/config/config.go` | 配置管理 |
| - | `internal/ui/ui.go` | UI 工具 |

---

## 六、实施计划

### 第一阶段 - 核心功能修复

| 序号 | 功能 | 状态 |
|------|------|------|
| 1 | 完善 `pool remove` / `pool import` | ✅ |
| 2 | 完善卸载功能 | ✅ |
| 3 | 添加 Hook 缓存机制 | ✅ |
| 4 | 修复 Token 刷新逻辑 | ✅ |

### 第二阶段 - 提升用户体验

| 序号 | 功能 | 状态 |
|------|------|------|
| 5 | 双语支持 | ✅ |
| 6 | 策略别名 | ✅ |
| 7 | models_to_check 配置 | ✅ |

---

## 七、改造记录 (2026-05-29)

### 修改的文件清单

| 文件 | 修改类型 | 说明 |
|------|----------|------|
| `internal/utils/install.go` | 修改 | 修复 TOML 格式，完善卸载功能 |
| `internal/auth/auth.go` | 修改 | 添加 RemoveAccount, ImportAccount 函数 |
| `internal/quota/quota.go` | 修改 | 添加缓存机制，修复 Token 刷新，提取策略评估函数 |
| `internal/quota/hook.go` | 修改 | PreCheckHook 支持缓存和 Session ID |
| `internal/config/config.go` | 修改 | 添加 ModelsToCheck 配置项 |
| `cmd/gcam/main.go` | 修改 | 添加 pool remove/import，完善卸载和策略别名 |
| `internal/i18n/i18n.go` | 新建 | 双语支持 |

---

### 1. TOML 格式修复 (`internal/utils/install.go`)

**问题**: 斜杠命令配置文件使用了不存在的 `command` 字段

**修复前:**
```go
tomlContent := fmt.Sprintf(`# Gemini CLI Slash Command: /gcam
name = "gcam"
description = "Switch Gemini accounts and manage rotation"
command = ["%s"]
`, binPath)
```

**修复后:**
```go
promptContent := fmt.Sprintf("!{%s {{args}}}", binPath)
tomlContent := fmt.Sprintf(`# Gemini CLI Slash Command: /gcam
description = "Switch Gemini accounts and manage rotation. Usage: /gcam <index|next|quota|menu>"
prompt = "%s"
`, promptContent)
```

---

### 2. 账号池管理 (`internal/auth/auth.go`, `cmd/gcam/main.go`)

**新增函数** `RemoveAccount()`:
```go
// RemoveAccount removes an account from the pool by index or email
func RemoveAccount(targetArg string) error {
    profiles, err := GetProfiles()
    // ... 查找账号，支持序号或邮箱
    // 不能删除当前活跃账号
    // 删除 profile 目录
    // 更新 accounts.json
}
```

**新增函数** `ImportAccount()`:
```go
// ImportAccount imports account credentials from a file
func ImportAccount(credsPath, email string) error {
    // 读取凭据文件并验证 JSON 格式
    // 检查必需字段 (access_token, refresh_token)
    // 创建 profile 目录并复制凭据
}
```

**main.go 新增命令分支**:
```go
case "remove", "delete", "rm":
    if len(args) < 2 {
        ui.Error("Usage: gcam pool remove <index|email>")
        return
    }
    if err := auth.RemoveAccount(args[1]); err != nil {
        ui.Error("Remove failed: %v", err)
    }
case "import":
    if len(args) < 3 {
        ui.Error("Usage: gcam pool import <path> <email>")
        return
    }
    if err := auth.ImportAccount(args[1], args[2]); err != nil {
        ui.Error("Import failed: %v", err)
    }
```

---

### 3. 卸载功能完善 (`internal/utils/install.go`, `cmd/gcam/main.go`)

**新增 UninstallOptions 结构体**:
```go
type UninstallOptions struct {
    KeepAccounts bool // If true, keep account data
}
```

**新增 GetUninstallFiles() 函数**:
```go
func GetUninstallFiles(keepAccounts bool) []string {
    // 返回待删除文件列表
}
```

**main.go 卸载命令增强**:
```go
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
    // 显示待删除文件列表
    // 确认提示
    // 调用 UninstallHooks(utils.UninstallOptions{KeepAccounts: keepAccounts})
```

---

### 4. Hook 缓存机制 (`internal/quota/quota.go`, `internal/quota/hook.go`)

**新增类型 QuotaCache**:
```go
type QuotaCache struct {
    Timestamp    string        `json:"timestamp"`
    SessionID    string        `json:"session_id"`
    Buckets      []QuotaBucket `json:"buckets"`
    CacheMinutes int           `json:"cache_minutes"`
}
```

**新增函数**:
```go
func LoadCache(sessionID string, cacheMinutes int) *QuotaCache
func SaveCache(sessionID string, buckets []QuotaBucket, cacheMinutes int) error
func ClearCache() error
```

**RunPreCheckHook 增强**:
```go
func RunPreCheckHook(cfg *config.Config) {
    // 读取 stdin 获取 session_id
    // 尝试从缓存加载
    // 如果缓存过期或不存在，获取新数据并保存缓存
    // 根据策略评估是否需要切换
}
```

---

### 5. Token 刷新逻辑修复 (`internal/quota/quota.go`)

**修复 LoadToken()**:
```go
func LoadToken() (string, error) {
    // ...
    // 过期时返回错误，不再返回过期的 token
    if creds.ExpiryDate > 0 && time.Now().UnixMilli() > creds.ExpiryDate-10000 {
        return "", fmt.Errorf("token expired")
    }
    return creds.AccessToken, nil
}
```

**增强 RefreshToken()**:
```go
func RefreshToken() (string, error) {
    cmd := exec.Command("gemini", "-p", "/model list")
    if err := cmd.Run(); err != nil {
        return "", err
    }
    // 刷新后重新加载 token
    return LoadToken()
}
```

**修复 CheckQuota()**:
```go
func CheckQuota(cfg *config.Config) ([]QuotaBucket, bool, string, error) {
    token, err := LoadToken()
    if err != nil {
        // 过期时尝试刷新
        if strings.Contains(err.Error(), "expired") || os.IsNotExist(err) {
            token, err = RefreshToken()
        }
        // ...
    }
    // 401 时也尝试刷新
    if err != nil && strings.Contains(err.Error(), "unauthorized") {
        token, err = RefreshToken()
        // ...
    }
    // ...
}
```

---

### 6. 策略评估函数提取 (`internal/quota/quota.go`)

**新增 evaluateQuotaStrategy() 函数**:
```go
func evaluateQuotaStrategy(buckets []QuotaBucket, cfg *config.Config) (bool, bool, string) {
    // 根据策略模式匹配目标 bucket
    // 支持 conservative, gemini3-first, gemini3.1-pro-only, gemini3.1-series-only, custom
    // 匹配所有目标后，检查是否全部低于阈值
    // 返回 (shouldSwitch, isLow, reason)
}
```

---

### 7. models_to_check 配置 (`internal/config/config.go`, `internal/quota/quota.go`)

**新增配置项**:
```go
type AutoSwitch struct {
    // ...
    ModelsToCheck []string `json:"models_to_check"` // 新增
}
```

**默认值**:
```go
ModelsToCheck: []string{"gemini-3.1-pro-preview", "gemini-3.1-flash-lite-preview"},
```

**evaluateQuotaStrategy() 中使用**:
```go
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
```

---

### 8. 策略别名支持 (`cmd/gcam/main.go`)

**handleStrategy() 增强**:
```go
func handleStrategy(cfg *config.Config, args []string) {
    if len(args) > 0 {
        // 应用别名映射
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
        // 验证策略名称
        // ...
    }
}
```

---

### 9. 双语支持 (`internal/i18n/i18n.go`)

**新建 i18n.go**:
```go
package i18n

var Lang = map[string]map[string]string{
    "en": {
        "title": "GEMINI CLI ACCOUNT MANAGER",
        "subtitle": "Fast Switcher + Auto Rotation",
        "status": "STATUS",
        "active": "ACTIVE",
        // ... 更多翻译
    },
    "cn": {
        "title": "GEMINI CLI 账号管理器",
        "subtitle": "极速切换 + 自动轮换",
        "status": "状态",
        "active": "正在使用",
        // ... 更多翻译
    },
}

func GetLang() string
func T(key string) string
func GetStrategyDesc(strategy string) string
```

---

### 使用示例

```bash
# 账号池管理
gcam pool list                    # 列出账号
gcam pool login                  # OAuth 登录
gcam pool remove 1               # 删除 1 号账号
gcam pool remove user@example.com # 删除指定账号
gcam pool import ./creds.json user@example.com  # 导入凭据

# 策略管理
gcam strategy pro                 # 使用别名，等于 gemini3.1-pro-only
gcam strategy series             # 使用别名，等于 gemini3.1-series-only
gcam strategy custom             # 自定义策略

# 卸载
gcam uninstall                   # 完整卸载（需确认）
gcam uninstall --keep-accounts  # 保留账号数据
gcam uninstall --force           # 强制卸载（无需确认）
```

# Gemini CLI 账号管理器

![Go Version](https://img.shields.io/badge/go-1.18+-blue.svg)
![Platform](https://img.shields.io/badge/platform-Windows%20%7C%20Linux%20%7C%20macOS-yellow.svg)
![License](https://img.shields.io/badge/license-MIT-green.svg)

**Gemini CLI 账号管理器** 是一个专为 Google Gemini CLI 设计的轻量级高性能 Go 实现。它支持极速多账号切换、**配额耗尽自动轮换**，以及**统一的账号池管理**。

> 📖 [English Version](./README.md)

---

## ✨ 核心特性

- **极速响应**：使用 Go 语言编写，运行开销极低，指令瞬时执行。
- **零外部依赖**：仅使用 Go 标准库，确保了最大的可移植性和安全性，无需安装额外的运行环境。
- **原子化操作**：凭据切换采用“临时文件+重命名”策略，确保操作的原子性，防止配置文件损坏。
- **智能配额监控**：直接调用 Google Cloud Code API 实时获取准确的配额剩余情况。
- **自动轮换策略**：当检测到配额耗尽或触发阈值时，自动无感切换至账号池中的下一个可用账号。
- **原生 OAuth 登录**：内置本地服务器，一键唤起浏览器完成官方 OAuth 授权并直接捕获凭据。
- **全平台支持**：完美兼容 macOS、Linux 和 Windows 系统。

---

## 🚀 快速开始

### 一键安装（推荐）

**macOS / Linux:**
```bash
curl -sSL https://gcam.dong4j.site/install.sh | bash
```

**Windows:**
```powershell
# 在 PowerShell 或 CMD 中运行
curl -sSL https://gcam.dong4j.site/install.bat -o install.bat && install.bat
```

**避免 GitHub API 频率限制** (60次/时 → 5000次/时):
```bash
export GITHUB_TOKEN=your_token && curl -sSL https://gcam.dong4j.site/install.sh | bash
```

安装目录：
- **macOS/Linux**: `~/.local/bin/gcam`
- **Windows**: `%LOCALAPPDATA%\Programs\gcam\gcam.exe`

### 安装挂钩 (Hooks)
执行以下命令以启用自动配额切换功能和斜杠命令支持：
```bash
gcam install
```

### 源码编译

**环境要求：**
- 已安装 [Go](https://golang.google.cn/doc/install) 1.18 或更高版本。
- 已安装 Gemini CLI 且 `gemini` 命令已加入系统的 `PATH`。

```bash
git clone https://github.com/dong4j/gemini-cli-account-manager.git
cd gemini-cli-account-manager
go build -o gcam cmd/gcam/main.go
```

---

## 🛠 使用方法

### 终端命令
| 命令 | 说明 |
| :--- | :--- |
| `gcam` | 列出所有账号及当前状态 |
| `gcam <n>` | 快速切换到序号为 n 的账号 (例如 `gcam 1`) |
| `gcam <email>` | 切换到指定邮箱的账号 |
| `gcam next` | 切换到账号池中的下一个账号 |
| `gcam quota` | 显示当前账号所有模型的配额状态 |
| `gcam pool login` | 通过浏览器 OAuth 登录添加新账号 |
| `gcam menu` | 进入交互式管理菜单 |
| `gcam uninstall` | 一键卸载挂钩与相关配置 |

### 斜杠命令 (在 Gemini CLI 内部使用)
安装完成后，可以直接在对话中使用 `/gcam` 命令：
- `/gcam next`: 遇到限制时快速切号。
- `/gcam 1`: 切换到 1 号账号。
- `/gcam quota`: 查看当前剩余“体力”。

---

## ⚙️ 配置说明

配置文件位于 `~/.gemini/auth_config.json`。

| 配置项 | 默认值 | 说明 |
| :--- | :--- | :--- |
| `strategy` | `gemini3.1-series-only` | 轮换策略：`conservative`, `gemini3.1-series-only` 等 |
| `threshold` | `10` | 剩余配额低于 10% 时触发切换 |
| `max_retries` | `3` | 单次会话中自动切换的最大重试次数 |

---

## 🛡 开源协议
本项目采用 MIT 协议开源。详情请参阅 `LICENSE` 文件。

---
由 **dong4j** 开发。欢迎提交 Issue 或 PR！
�说明

配置文件位于 `~/.gemini/auth_config.json`。

| 配置项 | 默认值 | 说明 |
| :--- | :--- | :--- |
| `strategy` | `gemini3.1-series-only` | 轮换策略：`conservative`, `gemini3.1-series-only` 等 |
| `threshold` | `10` | 剩余配额低于 10% 时触发切换 |
| `max_retries` | `3` | 单次会话中自动切换的最大重试次数 |

---

### 📝 致谢与说明

本项目是参考 GitHub 上的 Python 实现 [Gemini-CLI-Auth-Manager](https://github.com/v87/Gemini-CLI-Auth-Manager) 而开发的 **Go 语言版本**。

上游项目提供了多账号管理和配额轮换的核心思路，本项目在此基础上使用 Go 重新实现，旨在提供：
- **更快的执行速度**：Go 原生编译，无运行时开销。
- **更简单的部署**：单一二进制文件，**无需安装 Python 环境**。
- **更低的系统资源占用**。

---

## 🛡 开源协议
本项目采用 MIT 协议开源。详情请参阅 `LICENSE` 文件。

---
由 **dong4j** 开发。欢迎提交 Issue 或 PR！

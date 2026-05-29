# Gemini CLI Account Manager

![Go Version](https://img.shields.io/badge/go-1.18+-blue.svg)
![Platform](https://img.shields.io/badge/platform-Windows%20%7C%20Linux%20%7C%20macOS-yellow.svg)
![License](https://img.shields.io/badge/license-MIT-green.svg)

A lightweight, high-performance Go implementation for managing Google Gemini CLI accounts. It supports ultra-fast account switching, automated quota rotation, and unified account pool management.

---

### 📝 Acknowledgement

This project is a **Go-based implementation** inspired by the original Python version: [Gemini-CLI-Auth-Manager](https://github.com/v87/Gemini-CLI-Auth-Manager). 

While the Python version provided the foundational concept for multi-account management and quota rotation, this Go implementation focuses on:
- **Zero runtime dependencies** (no Python environment required).
- **Faster execution** and lower memory footprint.
- **Native cross-platform binaries**.

> 📖 [中文版说明](./README-CN.md)

---

## ✨ Features

- **Blazing Fast**: Written in Go for minimal overhead and instant execution.
- **Zero External Dependencies**: Uses only Go standard libraries for maximum portability and security.
- **Atomic Operations**: Credential switching is performed using atomic file renaming to prevent corruption.
- **Smart Quota Monitoring**: Real-time quota tracking via Google Cloud Code APIs.
- **Automatic Rotation**: Automatically switches to the next available account when quota is exhausted.
- **Native OAuth Integration**: Built-in local server to capture OAuth credentials directly from your browser.
- **Cross-Platform**: Fully compatible with macOS, Linux, and Windows.

---

## 🚀 Installation

### One-Click Install (Recommended)

**macOS / Linux:**
```bash
curl -sSL https://gcam.dong4j.site/install.sh | bash
```

**Windows:**
```powershell
# 在 PowerShell 或 CMD 中运行
curl -sSL https://gcam.dong4j.site/install.bat -o install.bat && install.bat
```

This will install the binary to:
- **macOS/Linux**: `~/.local/bin/gcam`
- **Windows**: `%LOCALAPPDATA%\Programs\gcam\gcam.exe`

### Install Hooks
To enable automatic quota switching and slash commands in Gemini CLI:
```bash
gcam install
```

### Build from Source

**Prerequisites:**
- [Go](https://golang.org/doc/install) 1.18 or higher.
- Gemini CLI installed and available in your `PATH`.

```bash
git clone https://github.com/dong4j/gemini-cli-account-manager.git
cd gemini-cli-account-manager
go build -o gcam cmd/gcam/main.go
```

### Install Hooks
To enable automatic quota switching and slash commands in Gemini CLI:
```bash
./gcam install
```

---

## 🛠 Usage

### Command Line
| Command | Description |
| :--- | :--- |
| `gcam` | List all accounts and their status |
| `gcam <n>` | Switch to account by index (e.g., `gcam 1`) |
| `gcam <email>` | Switch to account by email |
| `gcam next` | Switch to the next available account in the pool |
| `gcam quota` | Display current quota usage for all models |
| `gcam pool login` | Add a new account via OAuth browser login |
| `gcam menu` | Open the interactive management menu |
| `gcam uninstall` | Remove hooks and slash commands |

### Within Gemini CLI
Once installed, you can use the `/gcam` slash command:
- `/gcam next`: Switch account when you hit a limit.
- `/gcam 1`: Switch to the first account.
- `/gcam quota`: Check if you're close to the limit.

---

## ⚙️ Configuration

The tool stores its configuration in `~/.gemini/auth_config.json`.

| Field | Default | Description |
| :--- | :--- | :--- |
| `strategy` | `gemini3.1-series-only` | Rotation strategy: `conservative`, `gemini3.1-series-only`, etc. |
| `threshold` | `10` | Trigger switch when remaining quota is below 10%. |
| `max_retries` | `3` | Maximum number of automatic switch attempts per session. |

---

### 📝 Acknowledgement

This project is a **Go-based implementation** inspired by the original Python version: [Gemini-CLI-Auth-Manager](https://github.com/v87/Gemini-CLI-Auth-Manager). 

While the Python version provided the foundational concept for multi-account management and quota rotation, this Go implementation focuses on:
- **Zero runtime dependencies** (no Python environment required).
- **Faster execution** and lower memory footprint.
- **Native cross-platform binaries**.

---

## 🛡 License
Distributed under the MIT License. See `LICENSE` for more information.

---
Developed by **dong4j**. Feel free to submit Issues or PRs!

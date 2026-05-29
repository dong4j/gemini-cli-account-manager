#!/bin/bash
# ==============================================================================
# GCAM 一键安装脚本 (macOS / Linux)
# 使用方式: curl -sSL https://gcam.dong4j.site/install.sh | bash
# 或指定版本: curl -sSL https://gcam.dong4j.site/install.sh | bash -s -- v1.1.0
# ==============================================================================

set -e

# 配置
REPO="dong4j/gemini-cli-account-manager"
BINARY_NAME="gcam"
INSTALL_DIR="${HOME}/.local/bin"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

info() {
    echo -e "${BLUE}[INFO]${NC} $1" >&2
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1" >&2
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1" >&2
}

error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

# 检测操作系统和架构
detect_os() {
    OS="$(uname -s)"
    ARCH="$(uname -m)"

    case "$OS" in
        Darwin*)
            OS="darwin"
            ;;
        Linux*)
            OS="linux"
            ;;
        *)
            error "不支持的操作系统: $OS"
            error "Windows 用户请使用 install.bat: https://gcam.dong4j.site/install.bat"
            exit 1
            ;;
    esac

    # 转换架构名称
    case "$ARCH" in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        arm64|aarch64)
            ARCH="arm64"
            ;;
        *)
            error "不支持的架构: $ARCH"
            exit 1
            ;;
    esac

    info "检测到平台: ${OS}-${ARCH}"
}

# 获取最新版本号
get_latest_version() {
    if [ -n "$1" ]; then
        VERSION="$1"
        return
    fi

    info "获取最新版本号..."
    VERSION=$(curl -sSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"tag_name": "([^"]+)".*/\1/')

    if [ -z "$VERSION" ]; then
        error "无法获取最新版本，请检查网络连接或指定版本号"
        exit 1
    fi

    info "最新版本: $VERSION"
}

# 下载二进制文件（只返回文件路径，不输出其他内容）
download_binary() {
    local version="$1"
    local filename="${BINARY_NAME}-${OS}-${ARCH}"
    local download_url="https://github.com/${REPO}/releases/download/${version}/${filename}"

    info "下载地址: $download_url"

    # 创建临时文件
    local tmpfile="/tmp/gcam_install_${filename}"
    rm -f "$tmpfile"

    info "正在下载..."
    if ! curl -sSL -o "$tmpfile" "$download_url"; then
        error "下载失败，请检查版本号是否正确: $version"
        rm -f "$tmpfile"
        exit 1
    fi

    # 验证文件
    if [ ! -s "$tmpfile" ]; then
        error "下载文件为空或损坏"
        rm -f "$tmpfile"
        exit 1
    fi

    # 检查文件大小
    local size=$(wc -c < "$tmpfile" 2>/dev/null || echo "0")
    if [ "$size" -lt 1000 ]; then
        if file "$tmpfile" 2>/dev/null | grep -q "HTML"; then
            error "下载文件似乎是 HTML 错误页面，版本可能不存在: $version"
            rm -f "$tmpfile"
            exit 1
        fi
    fi

    # 只输出文件路径到 stdout（供变量捕获）
    echo "$tmpfile"
}

# 安装到目标目录
install_binary() {
    local src="$1"

    # 创建安装目录
    if [ ! -d "$INSTALL_DIR" ]; then
        info "创建安装目录: $INSTALL_DIR"
        mkdir -p "$INSTALL_DIR"
    fi

    local target="${INSTALL_DIR}/${BINARY_NAME}"

    # 移动文件
    info "安装到: $target"
    mv "$src" "$target"
    chmod +x "$target"

    # 清理临时文件
    rm -f "$src"

    success "安装完成!"
    echo
    echo "=========================================="
    echo "  GCAM 已成功安装到:"
    echo "  $target"
    echo "=========================================="
    echo
}

# 检查并提示配置 PATH
check_path() {
    # 检查 PATH 是否已包含安装目录
    if echo ":$PATH:" | grep -q ":$INSTALL_DIR:"; then
        info "PATH 配置正确，无需额外设置"
        return 0
    fi

    warn "需要将 $INSTALL_DIR 添加到 PATH"
    echo
    echo "请在您的 shell 配置文件中添加以下内容:"
    echo
    echo -e "  ${GREEN}# GCAM${NC}"
    echo -e "  export PATH=\"\$HOME/.local/bin:\$PATH\""
    echo

    # 检测 shell 类型
    local shell_config=""
    case "$(basename "$SHELL")" in
        zsh)
            shell_config="${HOME}/.zshrc"
            ;;
        bash)
            if [ "$(uname -s)" = "Darwin" ]; then
                shell_config="${HOME}/.bash_profile"
            else
                shell_config="${HOME}/.bashrc"
            fi
            ;;
        fish)
            shell_config="${HOME}/.config/fish/config.fish"
            ;;
        *)
            shell_config="${HOME}/.profile"
            ;;
    esac

    echo "配置文件: $shell_config"
    echo
    echo "添加后，运行以下命令使其生效:"
    echo
    echo -e "  ${GREEN}source $shell_config${NC}"
    echo
    echo "或者重新打开终端。"
}

# 显示使用指南
show_usage() {
    echo
    echo "=========================================="
    echo "  常用命令"
    echo "=========================================="
    echo
    echo -e "  ${GREEN}gcam${NC}               查看账号列表"
    echo -e "  ${GREEN}gcam 1${NC}            切换到 1 号账号"
    echo -e "  ${GREEN}gcam next${NC}          切换到下一个账号"
    echo -e "  ${GREEN}gcam quota${NC}         查看配额使用情况"
    echo -e "  ${GREEN}gcam pool login${NC}   添加新账号"
    echo -e "  ${GREEN}gcam menu${NC}          打开交互式菜单"
    echo
    echo "  安装钩子（启用 /gcam 命令）:"
    echo -e "  ${GREEN}gcam install${NC}"
    echo
    echo "  查看帮助:"
    echo -e "  ${GREEN}gcam --help${NC}"
    echo
    echo "=========================================="
    echo
    echo -e "  文档: ${BLUE}https://gcam.dong4j.site${NC}"
    echo -e "  GitHub: ${BLUE}https://github.com/${REPO}${NC}"
    echo
}

# 主函数
main() {
    echo
    echo "=========================================="
    echo "  GCAM 一键安装脚本"
    echo "  https://github.com/${REPO}"
    echo "=========================================="
    echo

    # 检测系统
    detect_os

    # 获取版本
    get_latest_version "$1"

    # 下载（只捕获文件路径）
    BINARY_PATH=$(download_binary "$VERSION")

    # 安装
    install_binary "$BINARY_PATH"

    # 检查 PATH
    check_path

    # 显示使用指南
    show_usage
}

# 运行
main "$@"

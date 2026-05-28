#!/bin/bash

# ==============================================================================
# GCAM 一键发布脚本
# 用法: ./deploy.sh v1.0.1
# ==============================================================================

# 确保脚本在出错时停止
set -e

# 检查是否提供了版本参数
if [ -z "$1" ]; then
    echo "错误: 请提供版本号 (例如: ./deploy.sh v1.0.1)"
    exit 1
fi

VERSION=$1

echo "开始发布版本: $VERSION ..."

# 1. 修改 HTML 中的版本号
# 注意: macOS 上的 sed -i 需要一个空的字符串参数
sed -i '' "s/Version: v[0-9]*\.[0-9]*\.[0-9]*/Version: $VERSION/g" landing-page/index.html
sed -i '' "s/Version: v[0-9]*\.[0-9]*\.[0-9]*/Version: $VERSION/g" landing-page/zh.html

echo "✅ 已更新 landing-page/index.html 和 zh.html 中的版本号为 $VERSION"

# 2. 提交更改
git add .
# 如果没有更改，git commit 会报错，所以这里加个判断
if git diff --staged --quiet; then
    echo "ℹ️ 没有代码更改需要提交"
else
    git commit -m "release: 发布 $VERSION"
    echo "✅ 已提交代码更改"
fi

# 3. 推送到远程主分支
git push origin main
echo "✅ 已推送到远程 main 分支"

# 4. 处理标签 (Tag)
# 如果标签已存在，先删除它以确保它指向最新的提交
if git rev-parse "$VERSION" >/dev/null 2>&1; then
    echo "⚠️ 标签 $VERSION 已存在，正在更新..."
    git tag -d "$VERSION"
    git push origin :refs/tags/"$VERSION"
fi

git tag "$VERSION"
git push origin "$VERSION"
echo "✅ 已推送标签 $VERSION，正在触发 GitHub Actions 构建..."

echo "--------------------------------------------------"
echo "🚀 发布指令已完成！请前往 GitHub Actions 查看进度。"
echo "Release: https://github.com/dong4j/gemini-cli-account-manager/actions"
echo "--------------------------------------------------"

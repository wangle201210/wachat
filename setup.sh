#!/bin/bash

# Wails Chat 快速安装脚本

set -e

echo "🚀 开始设置 Wails Chat..."

# 检查 Go
if ! command -v go &> /dev/null; then
    echo "❌ Go 未安装。请先安装 Go 1.21 或更高版本"
    echo "   访问: https://go.dev/dl/"
    exit 1
fi

echo "✅ Go 已安装: $(go version)"

# 检查 Node.js
if ! command -v node &> /dev/null; then
    echo "❌ Node.js 未安装。请先安装 Node.js 20.19.0 或更高版本"
    exit 1
fi

echo "✅ Node.js 已安装: $(node --version)"

# 检查 pnpm
if ! command -v pnpm &> /dev/null; then
    echo "📦 安装 pnpm..."
    npm install -g pnpm
fi

echo "✅ pnpm 已安装: $(pnpm --version)"

# 检查 Wails CLI
if ! command -v wails &> /dev/null; then
    echo "📦 安装 Wails CLI..."
    go install github.com/wailsapp/wails/v2/cmd/wails@latest

    # 确保 GOPATH/bin 在 PATH 中
    export PATH=$PATH:$(go env GOPATH)/bin
fi

echo "✅ Wails CLI 已安装"

# 安装 Go 依赖
echo "📦 安装 Go 依赖..."
go mod tidy

# 安装前端依赖
echo "📦 安装前端依赖..."
cd frontend
pnpm install
cd ..

# 创建 .env 文件
if [ ! -f .env ]; then
    echo "📝 创建 .env 文件..."
    cp .env.example .env
    echo "⚠️  请编辑 .env 文件并添加你的 API Key"
fi

echo ""
echo "✨ 设置完成！"
echo ""
echo "下一步:"
echo "  1. 编辑 .env 文件配置 API Key (可选)"
echo "  2. 运行 'wails dev' 启动开发服务器"
echo "  3. 或运行 'wails build' 构建生产版本"
echo ""

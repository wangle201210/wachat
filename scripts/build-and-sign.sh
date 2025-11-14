#!/bin/bash
set -e

# 完整的构建和签名流程
# 用于创建可分发的 macOS 应用

echo "🏗️  开始构建应用..."

# 清理旧的构建产物
rm -rf build/bin/wachat.app

# 构建应用
wails build

APP_PATH="build/bin/wachat.app"

if [ ! -d "$APP_PATH" ]; then
    echo "❌ 构建失败: 未找到 $APP_PATH"
    exit 1
fi

echo "✅ 构建完成"
echo ""

# 签名
echo "📝 正在签名应用..."
codesign --remove-signature "$APP_PATH" 2>/dev/null || true
codesign --force --deep --sign - "$APP_PATH"

echo "✅ 签名完成"
echo ""

# 验证签名
echo "🔍 签名信息:"
codesign -dv "$APP_PATH" 2>&1 | grep -E "(Signature|Identifier|Format)" || true
echo ""

# 创建分发包
DIST_DIR="build/dist"
ZIP_NAME="wachat-$(date +%Y%m%d-%H%M%S).zip"

mkdir -p "$DIST_DIR"

echo "📦 正在创建分发包..."
(cd build/bin && zip -r -y "../dist/$ZIP_NAME" wachat.app)

ZIP_SIZE=$(du -h "build/dist/$ZIP_NAME" | cut -f1)

echo "✅ 分发包创建完成!"
echo ""
echo "📍 文件位置: build/dist/$ZIP_NAME"
echo "📊 文件大小: $ZIP_SIZE"
echo ""
echo "💡 使用说明:"
echo "   1. 分发此 ZIP 文件给用户"
echo "   2. 用户解压后首次打开会提示: '无法验证开发者'"
echo "   3. 用户需要:"
echo "      - 右键点击应用 -> 打开"
echo "      - 或在 系统设置 > 隐私与安全性 中点击'仍要打开'"
echo ""
echo "⚠️  注意: 这是 ad-hoc 签名，不是 Apple Developer 签名"
echo "   如需完全消除警告，需要 Apple Developer 账号进行签名和公证"

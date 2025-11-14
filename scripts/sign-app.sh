#!/bin/bash
set -e

# macOS 应用签名脚本
# 解决"应用已损坏"的问题

APP_PATH="build/bin/wachat.app"

if [ ! -d "$APP_PATH" ]; then
    echo "❌ 未找到应用: $APP_PATH"
    echo "请先运行: wails build"
    exit 1
fi

echo "📝 正在对应用进行签名..."

# 移除旧的签名（如果存在）
codesign --remove-signature "$APP_PATH" 2>/dev/null || true

# Ad-hoc 签名（使用 - 作为身份）
# 这会让 macOS 认为应用是"已签名但未公证"而非"已损坏"
codesign --force --deep --sign - "$APP_PATH"

echo "✅ 签名完成！"

# 验证签名
echo ""
echo "🔍 验证签名信息:"
codesign -dv "$APP_PATH" 2>&1 | grep -E "(Signature|Identifier|Format)"

echo ""
echo "📦 现在可以分发应用了"
echo "   用户首次打开时会提示:'无法验证开发者'"
echo "   然后可以在 系统设置 > 隐私与安全性 中允许运行"

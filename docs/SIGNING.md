# macOS 应用签名指南

本文档说明如何对 wachat 应用进行签名，以避免"应用已损坏"的问题。

## 问题说明

### macOS 安全机制

macOS 对应用有三种状态：

1. **已损坏** ❌
   - 应用未签名或签名无效
   - 用户看到："应用已损坏，无法打开"
   - 只能通过 `xattr -d` 命令移除隔离属性

2. **无法验证开发者** ⚠️
   - 应用已签名，但未经过公证
   - 用户看到："无法验证开发者"
   - 可以在系统设置中允许运行

3. **已公证** ✅
   - 应用已签名并通过 Apple 公证
   - 用户无警告，直接打开

### 当前问题

未签名的应用通过微信等方式传输后，macOS 会认为应用"已损坏"。

## 解决方案

### 方案 1: Ad-hoc 签名（推荐，无需付费）

这是最简单的方式，不需要 Apple Developer 账号。

**优点**:
- 完全免费
- 无需注册 Apple Developer
- 用户可以在系统设置中允许运行

**缺点**:
- 首次打开仍需要用户手动允许
- 不能通过 App Store 分发

**使用方法**:

```bash
# 方式 1: 构建 + 签名 + 打包（推荐）
./scripts/build-and-sign.sh

# 方式 2: 只对已构建的应用签名
./scripts/sign-app.sh
```

**分发说明**:

分发 `build/dist/wachat-*.zip` 文件给用户。用户首次打开时：

1. 解压 ZIP 文件
2. **右键点击** wachat.app -> 选择"打开"
3. 点击"打开"确认

或者：

1. 双击应用，看到"无法验证开发者"
2. 打开 **系统设置** -> **隐私与安全性**
3. 找到 "仍要打开 wachat" 的选项
4. 点击"仍要打开"

### 方案 2: 自签名证书（中级）

创建自己的签名证书，更专业但仍需用户手动允许。

**步骤**:

1. 打开"钥匙串访问"
2. 菜单: 钥匙串访问 -> 证书助理 -> 创建证书
3. 名称: 填写你的名字（如 "wanna"）
4. 身份类型: 代码签名
5. 勾选"让我覆盖这些默认值"
6. 一路下一步，创建证书

然后使用证书签名：

```bash
# 替换 "Your Name" 为你在钥匙串中创建的证书名称
codesign --force --deep --sign "Your Name" build/bin/wachat.app
```

**优点**:
- 应用显示你的名字
- 更专业的外观

**缺点**:
- 仍需用户手动允许
- 设置稍复杂

### 方案 3: Apple Developer 签名 + 公证（专业，需付费）

这是最完美的方案，用户无需任何手动操作。

**要求**:
- Apple Developer 账号（$99/年）
- 开发者证书

**步骤**:

1. 注册 Apple Developer Program
2. 在 Xcode 中创建签名证书
3. 配置 wails.json:

```json
{
  "info": {
    "companyName": "Your Company",
    "productName": "wachat",
    "productVersion": "0.1.0",
    "copyright": "Copyright 2025"
  },
  "darwin": {
    "codesign": {
      "identity": "Developer ID Application: Your Name (TEAM_ID)",
      "entitlements": "build/darwin/entitlements.plist"
    }
  }
}
```

4. 构建并公证:

```bash
# 构建
wails build

# 公证（需要 Apple ID 和 app-specific password）
xcrun notarytool submit build/bin/wachat.zip \
  --apple-id "your@email.com" \
  --password "app-specific-password" \
  --team-id "TEAM_ID" \
  --wait

# 装订公证票据
xcrun stapler staple build/bin/wachat.app
```

**优点**:
- 用户无警告，直接打开
- 可以通过 Mac App Store 分发
- 最专业的解决方案

**缺点**:
- 需要付费（$99/年）
- 设置流程复杂
- 每次构建都需要公证（等待时间）

## 当前推荐

对于个人项目或早期版本，**推荐使用方案 1（Ad-hoc 签名）**：

```bash
./scripts/build-and-sign.sh
```

这会创建一个签名的应用并打包为 ZIP 文件，用户只需要右键打开或在系统设置中允许即可。

## 用户说明模板

分发应用时，建议附带以下说明：

---

**安装说明（macOS）**:

1. 下载并解压 wachat.zip
2. 首次打开方式（二选一）：
   - **方式 A**: 右键点击 wachat.app -> 选择"打开" -> 点击"打开"
   - **方式 B**: 双击打开，如果提示"无法验证开发者"，去 系统设置 > 隐私与安全性，点击"仍要打开"
3. 之后可以正常双击打开

**如果提示"应用已损坏"**（不应该出现，但如果出现）:

```bash
xattr -cr /Applications/wachat.app
```

---

## 故障排除

### 问题：仍然提示"已损坏"

**原因**: 签名过程出错或签名被破坏

**解决**:

```bash
# 检查签名状态
codesign -dv build/bin/wachat.app

# 重新签名
./scripts/sign-app.sh
```

### 问题：签名后压缩失效

**原因**: 使用了错误的压缩方式（破坏了签名）

**解决**: 使用 `-y` 参数保留符号链接

```bash
# 正确的压缩方式
zip -r -y wachat.zip wachat.app

# 错误的方式（会破坏签名）
# zip -r wachat.zip wachat.app
```

### 问题：通过微信传输后提示已损坏

**原因**: 微信可能会修改文件元数据

**解决**:
- 使用其他传输方式（网盘、AirDrop、邮件）
- 或者提供 xattr 命令给用户

## 参考资料

- [Apple Code Signing Guide](https://developer.apple.com/library/archive/documentation/Security/Conceptual/CodeSigningGuide/)
- [Wails Signing Documentation](https://wails.io/docs/guides/signing)
- [xcrun notarytool](https://developer.apple.com/documentation/security/notarizing_macos_software_before_distribution)

# wachat 安装指南

感谢使用 wachat！

## macOS 安装步骤

### 1. 解压应用

双击下载的 `wachat-*.zip` 文件，系统会自动解压。

### 2. 移动应用（可选）

将解压后的 `wachat.app` 拖动到"应用程序"文件夹。

### 3. 首次打开应用

由于应用未经过 Apple 公证，首次打开需要特殊步骤：

#### 方法 A：右键打开（推荐）

1. **右键点击** wachat.app
2. 选择 **"打开"**
3. 在弹出的对话框中点击 **"打开"**

![右键打开](https://support.apple.com/library/content/dam/edam/applecare/images/en_US/macos/Big-Sur/macos-big-sur-control-click-open.jpg)

#### 方法 B：系统设置

1. 双击 wachat.app
2. 看到 "无法验证开发者" 的提示
3. 打开 **系统设置** → **隐私与安全性**
4. 滚动到底部，找到 **"仍要打开 wachat"**
5. 点击 **"仍要打开"**
6. 再次双击 wachat.app

### 4. 后续使用

完成首次打开后，之后就可以像其他应用一样直接双击打开了。

## 如果提示"应用已损坏"

这种情况不应该出现，但如果出现了，请在终端运行以下命令：

```bash
# 替换 /path/to 为实际的应用路径
xattr -cr /path/to/wachat.app
```

例如，如果应用在"应用程序"文件夹：

```bash
xattr -cr /Applications/wachat.app
```

## 配置应用

首次启动后，您需要配置 AI 服务：

1. 应用会自动创建配置文件：`~/.wachat/config.yaml`
2. 编辑配置文件，填入您的 AI API 信息：

```yaml
ai:
  base_url: "https://api.openai.com/v1"
  api_key: "your-api-key-here"
  model: "gpt-3.5-turbo"
```

3. 重启应用使配置生效

## 常见问题

### Q: 为什么会提示"无法验证开发者"？

A: wachat 是独立开发者制作的应用，未经过 Apple 的付费公证程序。这不影响应用的安全性，只是需要您手动确认打开。

### Q: 应用安全吗？

A: 是的。wachat 是开源项目，您可以在 GitHub 查看源代码。应用不会收集任何个人信息，所有数据都保存在本地。

### Q: 如何卸载？

A: 直接将 wachat.app 拖到废纸篓。如需清除数据，删除 `~/.wachat` 文件夹。

## 需要帮助？

- GitHub Issues: https://github.com/wangle201210/wachat/issues
- 项目主页: https://github.com/wangle201210/wachat

---

**祝使用愉快！**

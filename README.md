# wachat

一个基于 Wails 框架和 Vue 3 的 AI 聊天应用，提供流畅的对话体验和多Tab会话管理功能。

**wachat** = Wails AI Chat

![wachat](./img/chat.png)

## ✨ 特性

- 🚀 **轻量级桌面应用** - 基于 Wails 2.x，Go + Vue 3 技术栈
- 💬 **流式AI对话** - 实时流式输出，更自然的对话体验
- 🎨 **现代化 UI** - 简洁优雅的界面设计
- 📑 **多Tab管理** - 支持多个对话标签页，类似浏览器的使用体验
- 💾 **本地持久化** - SQLite 数据库存储对话历史
- 🔌 **OpenAI 兼容** - 支持 OpenAI API 和其他兼容接口
- 🎯 **懒加载对话** - 新对话只在发送第一条消息时才创建，避免空对话
- ⚡ **Markdown 渲染** - 完美支持代码高亮和数学公式

## 🛠 技术栈

### 后端
- **框架**: Wails v2.10.2
- **语言**: Go 1.22+
- **数据库**: SQLite (gorm)
- **AI SDK**: [Cloudwego Eino](https://github.com/cloudwego/eino) - 字节跳动开源的 LLM 应用开发框架

### 前端
- **框架**: Vue 3.5+ (Composition API)
- **构建工具**: Vite 6
- **样式**: TailwindCSS 3
- **类型检查**: TypeScript 5.6+
- **Markdown**: vue-renderer-markdown
- **数学公式**: KaTeX
- **UI 设计**: 参考 [deepchat](https://github.com/ThinkInAIXYZ/deepchat) 的界面设计

## 📦 开发环境要求

- **Go**: 1.22 或更高版本
- **Node.js**: 20.19.0 或更高版本
- **pnpm**: 9.15.0 或更高版本
- **Wails CLI**: 2.10.2

### 安装 Wails CLI

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

## 🚀 快速开始

### 1. 克隆项目

```bash
git clone https://github.com/wangle201210/wachat.git
cd wachat
```

### 2. 安装依赖

```bash
# 安装 Go 依赖
go mod tidy

# 安装前端依赖
cd frontend
pnpm install
cd ..
```

### 3. 配置环境变量

创建 `.env` 文件:

```bash
cp .env.example .env
```

编辑 `.env` 文件配置您的 API 密钥:

```env
# OpenAI API 配置
OPENAI_API_KEY=your_api_key_here
OPENAI_API_URL=https://api.openai.com/v1/chat/completions
OPENAI_MODEL=gpt-4o-mini
```

> 💡 提示：
> - 支持 OpenAI 官方 API
> - 支持其他兼容 OpenAI API 的服务（如 Ollama、LocalAI 等）
> - 如不配置，应用将无法正常使用

### 4. 开发模式

```bash
# 启动开发服务器（热重载）
wails dev
```

### 5. 构建应用

```bash
# 构建生产版本
wails build

# 特定平台构建
wails build -platform darwin/arm64   # macOS Apple Silicon
wails build -platform darwin/amd64   # macOS Intel
wails build -platform windows/amd64  # Windows
wails build -platform linux/amd64    # Linux
```

构建完成后，可执行文件位于 `build/bin/` 目录。

## 🎯 核心功能

### 1. 多Tab会话管理

- 支持打开多个对话Tab，类似浏览器标签页
- 点击 "+" 按钮创建新会话
- 每个Tab独立显示对话内容
- 关闭Tab时自动切换到相邻Tab
- 最少保持一个Tab打开

### 2. 历史记录侧边栏

- 点击右上角历史按钮打开/关闭侧边栏
- 显示所有已保存的对话
- 点击对话可快速切换当前Tab内容
- 支持删除历史对话

### 3. 懒加载对话创建

- 新建Tab时不立即创建数据库记录
- 仅在用户发送第一条消息时才保存对话
- 避免产生大量空对话记录

### 4. 流式消息输出

- AI 回复采用流式输出
- 实时显示生成的内容
- 自动滚动到最新消息

### 5. Markdown 支持

- 完整的 Markdown 语法支持
- 代码块高亮显示
- 数学公式渲染（KaTeX）

## 📝 开发说明

### 数据库

应用使用 SQLite 存储数据，数据库文件位于用户目录：
- **macOS/Linux**: `~/.wachat/chat.db`
- **Windows**: `%USERPROFILE%\.wachat\chat.db`

数据库包含两张表：
- `conversations` - 存储会话信息
- `messages` - 存储消息记录

### 事件系统

前端通过 Wails Runtime 监听后端事件：

- `stream:start` - 流式响应开始
- `stream:response` - 接收流式内容块
- `stream:end` - 流式响应结束
- `stream:error` - 流式响应错误
- `conversation:title-updated` - 会话标题更新

## 🐛 常见问题

### Q: 如何更换 AI 服务提供商？

A: 修改 `.env` 文件中的配置：

```env
# 使用 Ollama 本地模型
OPENAI_API_URL=http://localhost:11434/v1/chat/completions
OPENAI_MODEL=llama2

# 使用 Azure OpenAI
OPENAI_API_URL=https://your-resource.openai.azure.com/openai/deployments/your-deployment/chat/completions?api-version=2024-02-15-preview
OPENAI_API_KEY=your_azure_key
```

### Q: 如何清空所有对话？

A: 直接删除数据库文件：

```bash
# macOS/Linux
rm ~/.wachat/chat.db

# Windows (PowerShell)
Remove-Item $env:USERPROFILE\.wachat\chat.db
```

### Q: 开发模式下修改代码后没有热重载？

A:
- Go 代码修改需要重启 `wails dev`
- Vue 代码修改会自动热重载
- 如果遇到问题，尝试清理缓存：`rm -rf frontend/dist`


## 🗺 开发路线图

- [x] 基础聊天功能
- [x] 多Tab会话管理
- [x] 本地数据持久化
- [x] 流式消息输出
- [x] Markdown 和代码高亮
- [ ] 消息编辑和重新生成
- [ ] 对话导出（JSON/Markdown）
- [ ] 主题切换（深色/浅色）
- [ ] 系统提示词设置
- [ ] 模型参数调整（temperature、max_tokens等）
- [ ] 快捷键支持
- [ ] 多语言国际化

## 🙏 致谢

- [Wails](https://github.com/wailsapp/wails) - 优秀的 Go + Web 框架
- [Vue.js](https://github.com/vuejs/core) - 渐进式 JavaScript 框架
- [Cloudwego Eino](https://github.com/cloudwego/eino) - 字节跳动开源的 LLM 应用开发框架，提供统一的 AI 接入能力
- [deepchat](https://github.com/ThinkInAIXYZ/deepchat) - UI 设计参考，提供了优秀的聊天界面设计灵感

## 📮 联系方式

如有问题或建议，欢迎提交 [Issue](https://github.com/wangle201210/wachat/issues)。

---

**Made with ❤️ using Wails and Vue**

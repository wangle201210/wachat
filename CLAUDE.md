# wachat - Claude 项目上下文

## 项目概述

wachat 是一个使用 Wails (Go + Web) 构建的跨平台 AI 聊天桌面应用。

**技术栈**: Go 1.22+ (后端) + Vue 3 + TypeScript (前端) + Wails v2.10.2 (框架) + SQLite (数据库)

**目标**: 提供轻量级、高性能的本地 AI 聊天体验，支持流式响应和会话持久化。

## 项目结构

```
wachat/
├── backend/              # Go 后端 - 严格分层架构
│   ├── api.go           # API Facade - 统一对外接口
│   ├── config/          # 配置层 - YAML 配置读取
│   ├── database/        # 数据库层 - SQLite 连接和迁移
│   ├── model/           # 模型层 - 数据结构定义
│   ├── repository/      # Repository 层 - 数据访问抽象
│   └── service/         # Service 层 - 业务逻辑
│       ├── ai.go        # AI 服务
│       ├── chat.go      # 聊天服务
│       ├── rag_service.go  # RAG 文档检索服务
│       └── binary_manager.go  # 二进制管理服务
├── frontend/            # Vue 3 前端 - Composition API
│   ├── src/
│   │   ├── components/  # UI 组件（单一职责）
│   │   ├── composables/ # 可复用逻辑
│   │   ├── views/       # 页面视图
│   │   └── wailsjs/     # Wails 自动生成的绑定（不要手动修改）
├── bin/                 # 嵌入的二进制文件（qdrant等）
├── config.yaml          # 配置文件（从 config.example.yaml 复制）
├── main.go              # Wails 应用入口
└── app.go               # 应用主逻辑 - 连接前后端
```

## 架构设计原则

### 后端分层

**各层职责**:
- `app.go`: Wails 生命周期管理，前端方法绑定，事件发送，*请不要在这里写复杂的业务逻辑*
- `api.go`: API 外观层，初始化各层依赖，提供统一接口
- `service/`: 业务逻辑，不直接访问数据库
  - `ai.go`: AI 对话服务，集成 RAG 增强
  - `chat.go`: 聊天会话管理
  - `rag_service.go`: RAG 文档检索服务（基于 go-rag）
  - `binary_manager.go`: 管理嵌入的二进制文件（qdrant等）
- `repository/`: 数据访问，GORM 操作封装
- `database/`: 数据库连接、迁移
- `model/`: 数据结构（DB 模型 + 业务模型）
- `config/`: YAML 配置读取和管理（使用 Viper）

### 前端组件化

**Composition API + 单一职责**:
- `views/`: 页面级组件，负责布局和组合
- `components/`: UI 组件，接收 props，发出 events，无业务逻辑
- `composables/`: 可复用的逻辑（useChat, useAutoScroll）

## 代码约定

### Go 代码规范

1. **错误处理**: 所有错误必须向上传递，使用 `fmt.Errorf` 包装
   ```go
   if err != nil {
       return fmt.Errorf("failed to do something: %w", err)
   }
   ```

2. **命名约定**:
   - 文件名: 小写+下划线 (`chat_service.go`)
   - 结构体: 大驼峰 (`ChatService`)
   - 方法/函数: 大驼峰（导出）或小驼峰（私有）

3. **依赖注入**: 通过构造函数传递依赖
   ```go
   func NewChatService(
       convRepo *repository.ConversationRepository,
       msgRepo *repository.MessageRepository,
       aiService *AIService,
   ) *ChatService
   ```

4. **不使用全局变量**: 所有依赖通过结构体字段传递

### TypeScript/Vue 代码规范

1. **组件命名**: 大驼峰 (`ChatSidebar.vue`)

2. **Composition API 风格**:
   ```typescript
   import { ref, computed, onMounted } from 'vue'

   const data = ref<Type>()
   const computed = computed(() => ...)
   ```

3. **类型安全**: 所有函数参数和返回值必须有类型标注

4. **Props 定义**:
   ```typescript
   interface Props {
     message: Message
     active?: boolean
   }
   const props = defineProps<Props>()
   ```

5. **事件发送**:
   ```typescript
   const emit = defineEmits<{
     'select-conversation': [id: string]
   }>()
   ```

### CSS/样式规范

1. **使用 TailwindCSS**: 优先使用 Tailwind utility classes
2. **自定义样式**: 使用 `<style scoped>` + `:deep()` 修改子组件
3. **颜色方案**: 使用 Tailwind 灰色调（gray-100, gray-700 等）

## 开发工作流程

### 添加新功能的标准流程

1. **后端**:
   ```
   1. 定义数据模型 (model/types.go)
   2. 创建 Repository 方法 (repository/*.go)
   3. 实现 Service 逻辑 (service/*.go)
   4. 在 API 层暴露接口 (api.go)
   5. 在 App 层绑定前端 (app.go)
   ```

2. **前端**:
   ```
   1. 运行 wails dev 自动生成 Go 绑定
   2. 创建/更新 Composable (composables/*.ts)
   3. 创建/更新组件 (components/*.vue)
   4. 在 View 中使用 (views/*.vue)
   ```

### 数据库修改流程

1. 修改 `model/types.go` 中的结构体
2. GORM 会在下次启动时自动迁移
3. 如需手动迁移，修改 `database/database.go` 的 `AutoMigrate` 调用

### 配置管理流程

**初始化配置**:
1. 复制 `config.example.yaml` 为 `config.yaml`
2. 修改 `config.yaml` 中的配置项（AI API、数据库路径等）
3. 配置文件会被 `.gitignore` 忽略，不会提交到版本控制

**配置文件搜索顺序**:
1. `WACHAT_CONFIG_PATH` 环境变量指定的目录
2. 当前工作目录
3. 当前工作目录向上查找的项目根目录（通过 go.mod 定位）
4. 可执行文件所在目录
5. 可执行文件向上查找的项目根目录
6. 用户配置目录 `~/.config/wachat`

**开发模式配置**:
```bash
# 如果 wails dev 找不到配置文件，设置环境变量
export WACHAT_CONFIG_PATH=$(pwd)
wails dev
```

**配置结构**:
```yaml
ai:
  base_url: "https://api.openai.com/v1"
  api_key: "your-api-key"
  model: "gpt-3.5-turbo"

binaries:
  enabled: true
  use_embedded: false
  bin_path: "./bin"
  startup_order:
    - qdrant
    - wailsproject

rag:
  enabled: false  # 需要 Elasticsearch 支持
  top_k: 5  # 检索返回的文档数量
```

**添加新配置项**:
1. 在 `backend/config/config.go` 中添加对应的结构体字段
2. 在 `config.Load()` 函数中设置默认值
3. 更新 `config.yaml` 和 `config.example.yaml`
4. 使用 `config.Get()` 获取配置

### 事件通信

**后端 → 前端**:
```go
runtime.EventsEmit(ctx, "stream:response", map[string]interface{}{
    "conversationId": id,
    "chunk": content,
})
```

**前端监听**:
```typescript
const runtime = (window as any).runtime
runtime.EventsOn('stream:response', (data: any) => {
    // 处理事件
})
```

## 已知问题和注意事项

### 1. Wails 自动生成的绑定
- **位置**: `frontend/src/wailsjs/`
- **不要手动修改**: 每次 `wails dev` 都会重新生成
- **如何更新**: 修改 Go 代码后，Wails 会自动重新生成

### 2. Node.js 版本要求
- Vite 7 需要 Node.js 20.19+ 或 22.12+
- 如果版本过低会有警告但仍能运行

### 3. 前端依赖安装
- 必须使用 `pnpm` 包管理器
- 安装命令: `pnpm --dir frontend install`
- 不要在 frontend 目录下直接运行 `cd frontend && pnpm install`（shell 配置问题）

## 测试策略

### 当前状态
- 项目处于早期阶段，暂无自动化测试
- 主要依靠手动测试


## 构建和部署

### 开发环境
```bash
wails dev  # 启动开发服务器
```

### 生产构建
```bash
wails build  # 构建当前平台
wails build -platform darwin/amd64  # 指定平台
```

### 构建产物
- **位置**: `build/bin/`
- **macOS**: `.app` 应用包
- **Windows**: `.exe` 可执行文件
- **Linux**: 二进制文件

## 给 AI 助手的指导

### 修改代码时

1. **遵循分层架构**: 不要让 Service 直接调用 Database，必须通过 Repository
2. **类型安全**: Go 和 TypeScript 都要保持严格的类型检查
3. **错误处理**: Go 代码必须处理所有错误
4. **组件职责**: 保持组件单一职责，逻辑放在 Composables

### 添加新 API 时

1. 先在 `service/` 实现业务逻辑
2. 在 `api.go` 暴露方法
3. 在 `app.go` 绑定给前端
4. 运行 `wails dev` 自动生成前端绑定
5. 前端从 `wailsjs/go/main/App` 导入使用

### 数据库操作时

1. 所有数据库操作必须在 `repository/` 层
2. 使用 GORM 的方法，避免原始 SQL
3. 返回错误而不是 panic

### 前端开发时

1. 使用 Composition API，不要使用 Options API
2. 状态管理优先使用 `ref` 和 `computed`
3. 复杂逻辑提取为 Composable
4. 组件通过 props 接收数据，通过 emit 发送事件

### 样式开发时

1. 优先使用 Tailwind classes
2. 保持颜色柔和（gray-700 而不是 black）
3. 使用 `prose` 类处理 Markdown 渲染
4. 需要覆盖样式时使用 `:deep()`

## 常见任务快速参考

### 添加新的会话操作
1. `repository/conversation.go` - 添加数据访问方法
2. `service/chat.go` - 添加业务逻辑
3. `api.go` - 暴露接口
4. `app.go` - 绑定前端
5. 前端 `composables/useChat.ts` - 调用新方法

### 修改 UI 样式
1. 找到对应的 `.vue` 组件
2. 修改 `<template>` 中的 Tailwind classes
3. 如需自定义样式，添加 `<style scoped>`

### 调试流式响应
1. 查看后端 `service/ai.go` 的 `StreamResponse` 方法
2. 查看 `app.go` 的事件发送逻辑
3. 查看前端 `composables/useChat.ts` 的事件监听

### 更改数据库 Schema
1. 修改 `model/types.go` 的结构体定义
2. 重启应用（GORM 自动迁移）
3. 更新相关的 Repository 方法

### 修改配置
1. 编辑 `config.yaml` 修改配置值
2. 如需添加新配置项：编辑 `backend/config/config.go`
3. 重启应用使配置生效

### 添加二进制文件

**本地模式**（推荐，更灵活）:
1. 将二进制文件放入 `bin/` 目录
2. 在 `config.yaml` 中设置：
   ```yaml
   binaries:
     enabled: true
     use_embedded: false
     bin_path: "./bin"
     startup_order:
       - your-binary
   ```
3. 无需重新编译，直接运行

**嵌入模式**（适合打包分发）:
1. 将二进制文件放入 `bin/` 目录
2. 在 `config.yaml` 中设置：
   ```yaml
   binaries:
     enabled: true
     use_embedded: true
     startup_order:
       - your-binary
   ```
3. 重新编译：`wails build`
4. 二进制会被打包到应用中（增加应用体积）

## RAG (Retrieval Augmented Generation) 集成

### 概述

wachat 集成了 RAG 功能，使用 `go-rag` 项目提供文档检索增强能力。RAG 可以从 Elasticsearch 中检索相关文档，并将其作为上下文添加到 AI 对话中，从而提供更准确、更有针对性的回答。

### 架构设计

**依赖关系**:
```
go-rag (独立 Git 项目)
    ↓ Go Modules (local replace)
wachat/backend/service/rag_service.go
    ↓ 依赖注入
wachat/backend/service/ai.go
```

**设计原则**:
- `go-rag` 保持为独立的 Git 项目
- wachat 通过 Go Modules 引用 go-rag
- 使用 `replace` 指令支持本地开发
- RAG 功能可通过配置开关启用/禁用

### Go Modules 配置

**go.mod 配置**:
```go
require (
    github.com/wangle201210/go-rag/server v0.0.0-00010101000000-000000000000
)

// 本地开发时使用 replace 指令
replace github.com/wangle201210/go-rag/server => ../go-rag/server
```

**本地开发**:
- 确保 `go-rag` 项目在 `../go-rag` 目录
- `replace` 指令允许本地修改 go-rag 并立即生效
- 不需要发布 go-rag 版本即可开发

**生产环境**:
- 发布 go-rag 到 GitHub 后，移除 `replace` 指令
- Go Modules 会从 GitHub 拉取指定版本

### RAG 配置

**启用 RAG**:
```yaml
rag:
  enabled: true
  elasticsearch_url: "http://localhost:9200"
  index_name: "knowledge_base"
  top_k: 5
```

**配置说明**:
- `enabled`: 是否启用 RAG（默认 false）
- `elasticsearch_url`: Elasticsearch 服务地址
- `index_name`: 文档索引名称
- `top_k`: 检索返回的文档数量（用于生成上下文）

### 工作流程

1. **初始化** (`backend/api.go`):
   ```go
   ragService, err := service.NewRAGService(ctx, ragConfig)
   aiService := service.NewAIService(aiConfig, ragService)
   ```

2. **文档检索** (`backend/service/rag_service.go`):
   ```go
   func (r *RAGService) RetrieveWithContext(ctx context.Context, query string) (string, error) {
       // 使用 go-rag 检索相关文档
       results := r.rag.Retrieve(ctx, query)
       // 格式化为上下文字符串
       return formatContext(results), nil
   }
   ```

3. **上下文增强** (`backend/service/ai.go`):
   ```go
   // 检索相关文档
   ragContext, err := a.ragService.RetrieveWithContext(a.ctx, userMessage)

   // 将上下文作为 system 消息添加到对话前
   systemMsg := &schema.Message{
       Role:    schema.System,
       Content: ragContext,
   }
   enhancedMessages = append([]*schema.Message{systemMsg}, messages...)
   ```

4. **AI 生成**: 使用增强后的消息调用 AI 模型

### 调试 RAG

**查看 RAG 是否生效**:
1. 检查 `config.yaml` 中 `rag.enabled` 是否为 `true`
2. 查看应用启动日志，确认 RAG 服务初始化成功
3. 在 `ai.go` 的 `StreamResponse` 方法中添加日志：
   ```go
   if ragContext != "" {
       fmt.Printf("RAG Context: %s\n", ragContext)
   }
   ```

**常见问题**:
- **Elasticsearch 连接失败**: 检查 `elasticsearch_url` 配置和 ES 服务状态
- **检索无结果**: 确认索引中有文档，且 `index_name` 正确
- **上下文过长**: 调整 `top_k` 值减少检索文档数量

### 修改 go-rag 项目

如果需要修改 go-rag 功能：

1. 在 `../go-rag` 目录修改代码
2. 由于使用了 `replace` 指令，修改会立即生效
3. 在 wachat 中重新编译测试：
   ```bash
   go mod tidy
   wails dev
   ```
4. 测试通过后，提交 go-rag 的修改
5. 可选：发布 go-rag 新版本，更新 wachat 的依赖版本

### 添加新的 RAG 功能

1. **在 go-rag 中添加新方法**
2. **在 rag_service.go 中封装**:
   ```go
   func (r *RAGService) NewFeature(ctx context.Context, params) (result, error) {
       return r.rag.NewMethod(ctx, params)
   }
   ```
3. **在 ai.go 中使用**:
   ```go
   if a.ragService != nil && a.ragService.IsEnabled() {
       result, err := a.ragService.NewFeature(ctx, params)
       // 处理结果
   }
   ```

## 注意事项

1. **不要向后兼容**: 项目还未发布第一个版本，可以大胆重构
2. **保持架构一致性**: 新功能必须遵循现有的分层架构
3. **保持轻量**: 避免引入大型依赖，保持应用体积小巧
4. **跨平台兼容**: 考虑 macOS/Windows/Linux 的差异

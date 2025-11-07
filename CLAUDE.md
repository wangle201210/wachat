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
│   ├── config/          # 配置层 - 环境变量读取
│   ├── database/        # 数据库层 - SQLite 连接和迁移
│   ├── model/           # 模型层 - 数据结构定义
│   ├── repository/      # Repository 层 - 数据访问抽象
│   └── service/         # Service 层 - 业务逻辑
├── frontend/            # Vue 3 前端 - Composition API
│   ├── src/
│   │   ├── components/  # UI 组件（单一职责）
│   │   ├── composables/ # 可复用逻辑
│   │   ├── views/       # 页面视图
│   │   └── wailsjs/     # Wails 自动生成的绑定（不要手动修改）
├── main.go              # Wails 应用入口
└── app.go               # 应用主逻辑 - 连接前后端
```

## 架构设计原则

### 后端分层

**各层职责**:
- `app.go`: Wails 生命周期管理，前端方法绑定，事件发送
- `api.go`: API 外观层，初始化各层依赖，提供统一接口
- `service/`: 业务逻辑，不直接访问数据库
- `repository/`: 数据访问，GORM 操作封装
- `database/`: 数据库连接、迁移
- `model/`: 数据结构（DB 模型 + 业务模型）
- `config/`: 环境变量配置读取

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

## 注意事项

1. **不要向后兼容**: 项目还未发布第一个版本，可以大胆重构
2. **保持架构一致性**: 新功能必须遵循现有的分层架构
3. **保持轻量**: 避免引入大型依赖，保持应用体积小巧
4. **跨平台兼容**: 考虑 macOS/Windows/Linux 的差异

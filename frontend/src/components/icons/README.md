# Icon Components

这个目录包含所有可复用的图标组件。

## 使用方法

### 基础用法

```vue
<script setup>
import { IconPlus, IconSettings } from '@/components/icons'
</script>

<template>
  <IconPlus class="w-4 h-4 text-gray-600" />
  <IconSettings size="16" className="text-blue-500" />
</template>
```

### Props

所有图标组件都接受以下 props：

- `size` (string | number, optional): 图标大小，默认为 24
- `className` (string, optional): 自定义 CSS 类名

### 可用图标

- `IconBase`: 基础图标组件（用于创建新图标）
- `IconArrowLeft`: 左箭头
- `IconPlus`: 加号
- `IconClose`: 关闭 (X)
- `IconSettings`: 设置
- `IconHistory`: 历史记录
- `IconDatabase`: 数据库
- `IconDownload`: 下载
- `IconPlay`: 播放
- `IconEdit`: 编辑
- `IconServer`: 服务器
- `IconSave`: 保存
- `IconInfo`: 信息
- `IconAlert`: 警告

### 创建新图标

基于 `IconBase` 创建新图标：

```vue
<template>
  <IconBase :size="size" :className="className">
    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="..." />
  </IconBase>
</template>

<script setup lang="ts">
import IconBase from './IconBase.vue'

defineProps<{
  size?: string | number
  className?: string
}>()
</script>
```

然后在 `index.ts` 中导出新图标。

# ExcelParserView 组件拆分设计

**日期：** 2026-04-10  
**范围：** `frontend/src/components/ExcelParserView.vue` → 多组件 + 单一 composable

---

## 背景

`ExcelParserView.vue` 当前为 748 行单文件，集中了所有 UI、状态与业务逻辑。拆分目标：符合企业级组件设计规范，提升可维护性与可测试性。

---

## 决策摘要

| 维度 | 决策 |
|------|------|
| 状态管理 | 单一 composable `useExcelParser`，不引入 Pinia |
| 组件粒度 | 中粒度，4 个直接子组件 + 1 个内部复用子组件 |
| 根组件职责 | 仅布局拼装，零业务逻辑 |

---

## 目录结构

```
frontend/src/
├── composables/
│   └── useExcelParser.js         ← 所有状态与业务逻辑
├── components/
│   ├── ExcelParserView.vue       ← 根视图，组装子组件
│   ├── ConfigSection.vue         ← 顶部配置区（路径 + 格式 + 按钮）
│   │     └── PathGroup.vue       ← 单行路径复用子组件
│   ├── OptionsBar.vue            ← 全选 + 导出选项行
│   ├── FileTable.vue             ← 文件表格 + 右键菜单
│   └── StatusBar.vue             ← 底部状态栏
```

---

## composable：`useExcelParser.js`

### 职责
封装全部响应式状态、生命周期钩子、watch、以及与后端 `FileService` 的所有交互。组件层不感知业务逻辑。

### 暴露的响应式状态

```js
// 配置路径与格式
configPath, outputPath, translatePath, translateLang
serverFormats, clientFormats
langOptions          // 动态语言选项列表

// 导出选项（互斥：compactOutput / prettyOutput）
forceExport, compactOutput, prettyOutput

// 文件列表
fileList

// UI 状态
isExporting, statusText, isLoadingConfig
isControlDisabled    // computed: isExporting

// 汇总计数
totalCount, changedCount, selectedCount   // computed

// 全选
selectAllChecked, selectAllIndeterminate  // computed

// 右键菜单
contextMenuShow, contextMenuX, contextMenuY, contextMenuRow
```

### 暴露的方法

```js
// 路径选择（调用 FileService.SelectDirectory）
openConfigPath()
openOutputPath()
openTranslatePath()

// 文件列表
reloadFiles()

// 导出入口
startGenerate()

// 全选控制
updateSelectAll(checked)

// 表格行事件（返回 NDataTable row-props 对象）
tableRowProps(row)

// 右键菜单
onSelectContextMenu(key)
closeContextMenu()
```

### 内部（不暴露）

```js
loadConfig()           // onMounted 触发
persistConfig()        // 保存配置到后端
scheduleSaveConfig()   // 防抖 300ms 包装
loadXlsxList(path)     // 拉取文件列表
ensureLangOption(lang) // 动态追加语言选项
```

`onMounted` / `onBeforeUnmount` / `watch` 全部在 composable 内注册，组件零生命周期代码。

---

## 组件接口

### `PathGroup.vue`

内部复用子组件，供 `ConfigSection` 使用三次。

| 方向 | 名称 | 类型 | 说明 |
|------|------|------|------|
| prop | `label` | String | 左侧标签文字 |
| prop | `modelValue` | String | 路径值（v-model） |
| prop | `placeholder` | String | 输入框占位符 |
| prop | `disabled` | Boolean | 禁用状态 |
| emit | `update:modelValue` | String | 路径输入变更 |
| emit | `browse` | — | 点击浏览按钮 |

---

### `ConfigSection.vue`

顶部完整配置区，包含三行路径（使用 PathGroup）、格式/语言选择、操作按钮。

| 方向 | 名称 | 类型 | 说明 |
|------|------|------|------|
| prop | `configPath` | String | 配置路径 |
| prop | `outputPath` | String | 输出路径 |
| prop | `translatePath` | String | 翻译路径 |
| prop | `translateLang` | String | 目标语言 |
| prop | `serverFormats` | Array | 服务器格式 |
| prop | `clientFormats` | Array | 客户端格式 |
| prop | `langOptions` | Array | 语言选项列表（动态） |
| prop | `disabled` | Boolean | 整体禁用 |
| emit | `update:configPath` | String | — |
| emit | `update:outputPath` | String | — |
| emit | `update:translatePath` | String | — |
| emit | `update:translateLang` | String | — |
| emit | `update:serverFormats` | Array | — |
| emit | `update:clientFormats` | Array | — |
| emit | `browse-config` | — | 浏览配置路径 |
| emit | `browse-output` | — | 浏览输出路径 |
| emit | `browse-translate` | — | 浏览翻译路径 |
| emit | `generate` | — | 开始生成 |
| emit | `reload` | — | 刷新列表 |

---

### `OptionsBar.vue`

| 方向 | 名称 | 类型 | 说明 |
|------|------|------|------|
| prop | `selectAllChecked` | Boolean | 全选状态 |
| prop | `selectAllIndeterminate` | Boolean | 半选状态 |
| prop | `forceExport` | Boolean | — |
| prop | `compactOutput` | Boolean | — |
| prop | `prettyOutput` | Boolean | — |
| prop | `disabled` | Boolean | — |
| prop | `fileCount` | Number | 用于判断全选是否可用 |
| emit | `update:selectAll` | Boolean | — |
| emit | `update:forceExport` | Boolean | — |
| emit | `update:compactOutput` | Boolean | — |
| emit | `update:prettyOutput` | Boolean | — |

---

### `FileTable.vue`

| 方向 | 名称 | 类型 | 说明 |
|------|------|------|------|
| prop | `fileList` | Array | 文件行数据 |
| prop | `isExporting` | Boolean | 导出中（禁用行复选框） |
| prop | `contextMenuShow` | Boolean | 右键菜单可见性 |
| prop | `contextMenuX` | Number | — |
| prop | `contextMenuY` | Number | — |
| emit | `row-contextmenu` | {row, x, y} | 右键触发 |
| emit | `context-menu-select` | key | 菜单项选中 |
| emit | `context-menu-close` | — | 菜单关闭 |

列定义（columns）在 `FileTable.vue` 内部构建，不由外部传入。`statusChip` 渲染函数和 `contextMenuOptions` 静态常量也内聚在此组件。行 checkbox 的 `selected` 属性直接在行对象上原地变更（Vue 3 深层响应追踪），无需 emit 整个数组。

---

### `StatusBar.vue`

纯展示组件，无 emits。

| prop | 类型 | 说明 |
|------|------|------|
| `totalCount` | Number | 文件总数 |
| `changedCount` | Number | 有变化数量 |
| `selectedCount` | Number | 已勾选数量 |
| `statusText` | String | 状态文字 |
| `isExporting` | Boolean | 控制 busy 样式 |

---

### `ExcelParserView.vue`（根视图，重构后）

调用 `useExcelParser()`，将 ref/方法分发给各子组件，自身只包含布局模板。模板结构：

```html
<ConfigSection v-bind="configProps" v-on="configHandlers" />
<OptionsBar v-bind="optionsProps" v-on="optionsHandlers" />
<FileTable v-bind="tableProps" v-on="tableHandlers" />
<StatusBar v-bind="statusProps" />
```

---

## 样式策略

- 当前 `ExcelParserView.vue` 内的 CSS token（`--blue`, `--bg` 等）迁移到 `App.vue` 的全局 `<style>`（去掉 scoped），供所有子组件共享
- 各子组件 `<style scoped>` 只保留自身独有的样式
- Naive UI `:deep()` override 保留在各自组件内，避免全局污染

---

## 不在本次范围内

- TypeScript 迁移
- Pinia 引入
- 单元测试
- 多视图/路由扩展

# ExcelParserView 组件拆分 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将单一 748 行的 `ExcelParserView.vue` 拆分为 1 个 composable + 5 个组件，实现逻辑与视图完全解耦。

**Architecture:** 全部业务状态与逻辑迁移到 `useExcelParser.js` composable；根组件 `ExcelParserView.vue` 瘦身为纯布局组装；4 个直接子组件（ConfigSection、OptionsBar、FileTable、StatusBar）各持有自身样式，PathGroup 作为 ConfigSection 内部复用子组件。

**Tech Stack:** Vue 3 `<script setup>`、Naive UI、Wails v3 JS bindings（`FileService`）

---

## 文件清单

| 操作 | 路径 | 职责 |
|------|------|------|
| 新建 | `frontend/src/composables/useExcelParser.js` | 所有响应式状态、watch、生命周期、FileService 调用 |
| 新建 | `frontend/src/components/StatusBar.vue` | 底部状态栏，纯展示 |
| 新建 | `frontend/src/components/PathGroup.vue` | 单行路径选择（label + input + 浏览按钮） |
| 新建 | `frontend/src/components/OptionsBar.vue` | 全选 + 导出选项行 |
| 新建 | `frontend/src/components/ConfigSection.vue` | 顶部配置区，含三个 PathGroup + 格式选择 + 操作按钮 |
| 新建 | `frontend/src/components/FileTable.vue` | NDataTable + 右键菜单，列定义内聚 |
| 修改 | `frontend/src/components/ExcelParserView.vue` | 清空逻辑，仅组装子组件 |

---

## Task 1: 创建 `useExcelParser.js` composable

**Files:**
- Create: `frontend/src/composables/useExcelParser.js`

- [ ] **Step 1: 新建文件，写入完整 composable**

```js
// frontend/src/composables/useExcelParser.js
import { ref, computed, watch, onMounted, onBeforeUnmount } from 'vue'
import { useMessage } from 'naive-ui'
import { FileService } from '../bindings/excelparse/service/index.js'

export function useExcelParser() {
  const message = useMessage()

  // ── 配置路径与格式 ──
  const configPath = ref('')
  const outputPath = ref('')
  const translatePath = ref('')
  const translateLang = ref('')
  const serverFormats = ref([])
  const clientFormats = ref([])
  const langOptions = ref([{ label: '', value: '' }])

  // ── 导出选项 ──
  const forceExport = ref(false)
  const compactOutput = ref(false)
  const prettyOutput = ref(false)

  // ── 文件列表 ──
  const fileList = ref([])

  // ── UI 状态 ──
  const isExporting = ref(false)
  const statusText = ref('就绪')
  const isLoadingConfig = ref(true)
  let saveConfigTimer = null

  // ── 右键菜单（contextMenuRow 为内部状态，不暴露给模板）──
  const contextMenuShow = ref(false)
  const contextMenuX = ref(0)
  const contextMenuY = ref(0)
  const contextMenuRow = ref(null)

  // ── Computed ──
  const isControlDisabled = computed(() => isExporting.value)

  const totalCount = computed(() => fileList.value.length)
  const changedCount = computed(() => fileList.value.filter((f) => f.fileStatus !== '-').length)
  const selectedCount = computed(() => fileList.value.filter((f) => f.selected).length)

  const selectAllChecked = computed({
    get() {
      return fileList.value.length > 0 && fileList.value.every((f) => f.selected)
    },
    set(checked) {
      fileList.value.forEach((f) => { f.selected = checked })
    },
  })

  const selectAllIndeterminate = computed(() => {
    if (fileList.value.length === 0) return false
    const selected = fileList.value.filter((f) => f.selected).length
    return selected > 0 && selected < fileList.value.length
  })

  // ── 内部方法 ──
  const ensureLangOption = (lang) => {
    if (!lang) return
    const exists = langOptions.value.some((item) => item.value === lang)
    if (!exists) {
      langOptions.value = [...langOptions.value, { label: lang, value: lang }]
    }
  }

  const persistConfig = async () => {
    try {
      await FileService.SaveConfig({
        config_path: configPath.value,
        output_path: outputPath.value,
        i18n_path: translatePath.value,
        i18n_lang: translateLang.value,
        server_fmts: serverFormats.value,
        client_fmts: clientFormats.value,
      })
    } catch (err) {
      console.error('保存配置失败:', err)
    }
  }

  const scheduleSaveConfig = () => {
    if (isLoadingConfig.value) return
    if (saveConfigTimer) clearTimeout(saveConfigTimer)
    saveConfigTimer = setTimeout(() => { persistConfig() }, 300)
  }

  const loadXlsxList = async (path) => {
    if (!path) { fileList.value = []; return }
    try {
      const list = await FileService.GetXlsxList(path)
      fileList.value = (list || []).map((item, idx) => ({
        key: idx + 1,
        selected: true,
        filename: item.name,
        filepath: item.path,
        fileStatus: item.need_parse ? '就绪' : '-',
        exportStatus: '-',
        exportResult: '-',
      }))
    } catch (err) {
      message.error(`加载配置表失败: ${String(err)}`)
    }
  }

  const loadConfig = async () => {
    try {
      const config = await FileService.GetConfig()
      if (config) {
        configPath.value = config.config_path || ''
        outputPath.value = config.output_path || ''
        translatePath.value = config.i18n_path || ''
        translateLang.value = config.i18n_lang || ''
        serverFormats.value = config.server_fmts || []
        clientFormats.value = config.client_fmts || []
        ensureLangOption(translateLang.value)
        await loadXlsxList(configPath.value)
      }
    } catch (err) {
      console.error('加载配置失败:', err)
    } finally {
      isLoadingConfig.value = false
    }
  }

  // ── 暴露的方法 ──
  const updateTranslateLang = (lang) => {
    translateLang.value = lang
    ensureLangOption(lang)
  }

  const closeContextMenu = () => {
    contextMenuShow.value = false
    contextMenuRow.value = null
  }

  const handleRowContextMenu = ({ row, x, y }) => {
    if (isControlDisabled.value) return
    contextMenuRow.value = row
    contextMenuX.value = x
    contextMenuY.value = y
    contextMenuShow.value = true
  }

  const onSelectContextMenu = async (key) => {
    try {
      if (key === 'open-dir') {
        if (!contextMenuRow.value?.filepath) { message.warning('未找到文件路径'); return }
        await FileService.OpenFileDirectory(contextMenuRow.value.filepath)
        message.success('已打开所在目录')
      } else if (key === 'open-file') {
        if (!contextMenuRow.value?.filepath) { message.warning('未找到文件路径'); return }
        await FileService.OpenFile(contextMenuRow.value.filepath)
        message.success('已打开文件')
      }
    } catch (err) {
      message.error(`操作失败: ${String(err)}`)
    } finally {
      closeContextMenu()
    }
  }

  const selectPath = async (pathType, title) => {
    try {
      const selected = await FileService.SelectDirectory(pathType, title)
      return selected || ''
    } catch (err) {
      message.error(`选择目录失败${title ? ` (${title})` : ''}: ${String(err)}`)
      return ''
    }
  }

  const openConfigPath = async () => {
    const dir = await selectPath(1, '选择配置路径')
    if (!dir) return
    configPath.value = dir
    await loadXlsxList(dir)
  }

  const openOutputPath = async () => {
    const dir = await selectPath(2, '选择输出路径')
    if (!dir) return
    outputPath.value = dir
  }

  const openTranslatePath = async () => {
    const dir = await selectPath(3, '选择翻译路径')
    if (!dir) return
    translatePath.value = dir
  }

  const reloadFiles = async () => {
    if (!configPath.value) { message.warning('请先选择有效的配置路径'); return }
    await loadXlsxList(configPath.value)
    message.success('配置表列表已刷新')
  }

  const validateExport = () => {
    if (!configPath.value) { message.warning('请先选择配置路径'); return false }
    if (!outputPath.value) { message.warning('请先选择导出路径'); return false }
    if (serverFormats.value.length === 0 && clientFormats.value.length === 0) {
      message.warning('请至少勾选一种导出格式'); return false
    }
    if (selectedCount.value === 0) { message.warning('请至少勾选一个要导出的文件'); return false }
    return true
  }

  const startGenerate = async () => {
    if (isExporting.value) { message.warning('正在导出中，请稍后'); return }
    if (!validateExport()) return

    isExporting.value = true
    statusText.value = '导出中...'
    fileList.value.forEach((row) => {
      if (row.selected) { row.exportStatus = '导出中...'; row.exportResult = '-' }
    })

    try {
      void forceExport.value
      void compactOutput.value
      void prettyOutput.value
      message.warning('后端导出接口尚未接入，已完成前端校验与状态流转逻辑')
      fileList.value.forEach((row) => {
        if (row.selected) { row.exportStatus = '-'; row.exportResult = '待后端导出接口' }
      })
    } finally {
      isExporting.value = false
      statusText.value = '就绪'
    }
  }

  const updateSelectAll = (checked) => {
    selectAllChecked.value = checked
  }

  // ── Watch ──
  watch(compactOutput, (val) => { if (val) prettyOutput.value = false })
  watch(prettyOutput, (val) => { if (val) compactOutput.value = false })
  watch(
    [configPath, outputPath, translatePath, serverFormats, clientFormats, translateLang],
    scheduleSaveConfig,
    { deep: true }
  )

  // ── 生命周期 ──
  onMounted(() => { loadConfig() })
  onBeforeUnmount(() => { if (saveConfigTimer) clearTimeout(saveConfigTimer) })

  return {
    // State
    configPath, outputPath, translatePath, translateLang,
    serverFormats, clientFormats, langOptions,
    forceExport, compactOutput, prettyOutput,
    fileList,
    isExporting, statusText,
    contextMenuShow, contextMenuX, contextMenuY,
    // Computed
    isControlDisabled,
    totalCount, changedCount, selectedCount,
    selectAllChecked, selectAllIndeterminate,
    // Methods
    updateTranslateLang,
    openConfigPath, openOutputPath, openTranslatePath,
    reloadFiles, startGenerate,
    updateSelectAll,
    handleRowContextMenu,
    onSelectContextMenu, closeContextMenu,
  }
}
```

- [ ] **Step 2: 提交**

```bash
git add frontend/src/composables/useExcelParser.js
git commit -m "feat: extract useExcelParser composable"
```

---

## Task 2: 创建 `StatusBar.vue`

**Files:**
- Create: `frontend/src/components/StatusBar.vue`

- [ ] **Step 1: 新建文件**

```vue
<!-- frontend/src/components/StatusBar.vue -->
<script setup>
defineProps({
  totalCount: Number,
  changedCount: Number,
  selectedCount: Number,
  statusText: String,
  isExporting: Boolean,
})
</script>

<template>
  <div class="status-bar">
    <span class="status-item">
      <span class="status-label">文件数量</span>
      <span class="status-count">{{ totalCount }}</span>
    </span>
    <span class="status-sep"></span>
    <span class="status-item">
      <span class="status-label">有变化</span>
      <span class="status-count">{{ changedCount }}</span>
    </span>
    <span class="status-sep"></span>
    <span class="status-item">
      <span class="status-label">已勾选</span>
      <span class="status-count">{{ selectedCount }}</span>
    </span>
    <span class="status-sep"></span>
    <span class="status-item" :class="{ 'status-busy': isExporting }">{{ statusText }}</span>
  </div>
</template>

<style scoped>
.status-bar {
  display: flex;
  align-items: center;
  padding: 5px 18px;
  background: var(--surface);
  border-top: 1px solid var(--border);
  flex-shrink: 0;
  font-size: 11.5px;
  color: var(--text-2);
}

.status-item {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 0 12px;
}

.status-item:first-child { padding-left: 0; }

.status-label { color: var(--text-3); }

.status-count {
  color: var(--text);
  font-weight: 600;
  font-variant-numeric: tabular-nums;
}

.status-sep {
  width: 1px;
  height: 12px;
  background: var(--border-dark);
  flex-shrink: 0;
}

.status-busy { color: var(--blue); font-weight: 500; }
</style>
```

- [ ] **Step 2: 提交**

```bash
git add frontend/src/components/StatusBar.vue
git commit -m "feat: add StatusBar component"
```

---

## Task 3: 创建 `PathGroup.vue`

**Files:**
- Create: `frontend/src/components/PathGroup.vue`

- [ ] **Step 1: 新建文件**

```vue
<!-- frontend/src/components/PathGroup.vue -->
<script setup>
import { NInput, NButton } from 'naive-ui'

defineProps({
  label: String,
  modelValue: String,
  placeholder: String,
  disabled: Boolean,
})

defineEmits(['update:modelValue', 'browse'])
</script>

<template>
  <div class="path-group">
    <span class="label">{{ label }}</span>
    <n-input
      :value="modelValue"
      size="small"
      class="path-input"
      :placeholder="placeholder"
      :disabled="disabled"
      @update:value="$emit('update:modelValue', $event)"
    />
    <n-button size="small" :disabled="disabled" @click="$emit('browse')">浏览</n-button>
  </div>
</template>

<style scoped>
.path-group {
  display: flex;
  align-items: center;
  gap: 7px;
}

.path-input { flex: 1; min-width: 0; }

.label {
  white-space: nowrap;
  font-size: 12px;
  color: var(--label);
  flex-shrink: 0;
  width: 56px;
  text-align: right;
}

:deep(.n-input) {
  --n-color:             #FFFFFF;
  --n-color-focus:       #FFFFFF;
  --n-color-disabled:    #F5F5F7;
  --n-border:            1px solid #D1D1D6;
  --n-border-hover:      1px solid #AEAEB2;
  --n-border-focus:      1px solid #007AFF;
  --n-border-radius:     8px;
  --n-box-shadow-focus:  0 0 0 3px rgba(0,122,255,0.18);
  --n-font-size:         12px;
  --n-text-color:        #1C1C1E;
  --n-placeholder-color: #AEAEB2;
}

:deep(.n-input .n-input__input-el) {
  font-family: var(--mono);
  font-size: 12px;
}

:deep(.n-button) {
  border-radius: 8px !important;
  font-weight: 500 !important;
  transition: background 0.15s ease, box-shadow 0.15s ease !important;
}

:deep(.n-button--default-type) {
  --n-color:         #FFFFFF !important;
  --n-color-hover:   #F5F5F7 !important;
  --n-color-pressed: #EBEBED !important;
  --n-border:        1px solid #D1D1D6 !important;
  --n-border-hover:  1px solid #AEAEB2 !important;
  --n-text-color:    #1C1C1E !important;
}
</style>
```

- [ ] **Step 2: 提交**

```bash
git add frontend/src/components/PathGroup.vue
git commit -m "feat: add PathGroup reusable component"
```

---

## Task 4: 创建 `OptionsBar.vue`

**Files:**
- Create: `frontend/src/components/OptionsBar.vue`

- [ ] **Step 1: 新建文件**

```vue
<!-- frontend/src/components/OptionsBar.vue -->
<script setup>
import { NCheckbox } from 'naive-ui'

defineProps({
  selectAllChecked: Boolean,
  selectAllIndeterminate: Boolean,
  forceExport: Boolean,
  compactOutput: Boolean,
  prettyOutput: Boolean,
  disabled: Boolean,
  fileCount: Number,
})

defineEmits([
  'update:selectAll',
  'update:forceExport',
  'update:compactOutput',
  'update:prettyOutput',
])
</script>

<template>
  <div class="options-row">
    <n-checkbox
      :checked="selectAllChecked"
      :indeterminate="selectAllIndeterminate"
      :disabled="disabled || fileCount === 0"
      @update:checked="$emit('update:selectAll', $event)"
    >全选</n-checkbox>
    <div class="options-divider" />
    <n-checkbox
      :checked="forceExport"
      :disabled="disabled"
      @update:checked="$emit('update:forceExport', $event)"
    >强制导出</n-checkbox>
    <n-checkbox
      :checked="compactOutput"
      :disabled="disabled"
      @update:checked="$emit('update:compactOutput', $event)"
    >紧凑输出</n-checkbox>
    <n-checkbox
      :checked="prettyOutput"
      :disabled="disabled"
      @update:checked="$emit('update:prettyOutput', $event)"
    >JSON 美化</n-checkbox>
  </div>
</template>

<style scoped>
.options-row {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 18px;
  background: var(--surface-2);
  border-bottom: 1px solid var(--border);
  flex-shrink: 0;
}

.options-divider {
  width: 1px;
  height: 13px;
  background: var(--border-dark);
  margin: 0 8px;
  flex-shrink: 0;
}

:deep(.n-checkbox .n-checkbox__box) {
  border-radius: 5px !important;
  border-color: #C7C7CC !important;
  transition: all 0.15s ease !important;
}

:deep(.n-checkbox.n-checkbox--checked .n-checkbox__box),
:deep(.n-checkbox.n-checkbox--indeterminate .n-checkbox__box) {
  background: #007AFF !important;
  border-color: #007AFF !important;
}

:deep(.n-checkbox__label) {
  font-size: 12px;
  color: #1C1C1E;
}
</style>
```

- [ ] **Step 2: 提交**

```bash
git add frontend/src/components/OptionsBar.vue
git commit -m "feat: add OptionsBar component"
```

---

## Task 5: 创建 `ConfigSection.vue`

**Files:**
- Create: `frontend/src/components/ConfigSection.vue`

- [ ] **Step 1: 新建文件**

```vue
<!-- frontend/src/components/ConfigSection.vue -->
<script setup>
import { NSelect, NButton } from 'naive-ui'
import PathGroup from './PathGroup.vue'

const FORMAT_OPTIONS = [
  { label: 'lua', value: 'lua' },
  { label: 'json', value: 'json' },
  { label: 'csharp', value: 'csharp' },
]

defineProps({
  configPath: String,
  outputPath: String,
  translatePath: String,
  translateLang: String,
  serverFormats: Array,
  clientFormats: Array,
  langOptions: Array,
  disabled: Boolean,
})

defineEmits([
  'update:configPath',
  'update:outputPath',
  'update:translatePath',
  'update:translateLang',
  'update:serverFormats',
  'update:clientFormats',
  'browse-config',
  'browse-output',
  'browse-translate',
  'generate',
  'reload',
])
</script>

<template>
  <div class="config-section">
    <div class="config-panel config-panel-left">
      <PathGroup
        label="配置路径"
        :model-value="configPath"
        placeholder="请选择配置路径"
        :disabled="disabled"
        @update:model-value="$emit('update:configPath', $event)"
        @browse="$emit('browse-config')"
      />
      <PathGroup
        label="输出路径"
        :model-value="outputPath"
        placeholder="请选择导出路径"
        :disabled="disabled"
        @update:model-value="$emit('update:outputPath', $event)"
        @browse="$emit('browse-output')"
      />
      <PathGroup
        label="翻译路径"
        :model-value="translatePath"
        placeholder="请选择翻译路径"
        :disabled="disabled"
        @update:model-value="$emit('update:translatePath', $event)"
        @browse="$emit('browse-translate')"
      />
    </div>

    <div class="config-divider" />

    <div class="config-panel config-panel-right">
      <div class="format-row">
        <span class="label">服务器格式</span>
        <n-select
          :value="serverFormats"
          :options="FORMAT_OPTIONS"
          multiple
          size="small"
          class="format-select"
          :disabled="disabled"
          @update:value="$emit('update:serverFormats', $event)"
        />
      </div>
      <div class="format-row">
        <span class="label">客户端格式</span>
        <n-select
          :value="clientFormats"
          :options="FORMAT_OPTIONS"
          multiple
          size="small"
          class="format-select"
          :disabled="disabled"
          @update:value="$emit('update:clientFormats', $event)"
        />
      </div>
      <div class="format-row">
        <span class="label">目标语言</span>
        <n-select
          :value="translateLang"
          :options="langOptions"
          size="small"
          class="lang-select"
          tag
          filterable
          clearable
          :disabled="disabled"
          @update:value="$emit('update:translateLang', $event)"
        />
      </div>
    </div>

    <div class="config-divider" />

    <div class="config-panel config-panel-actions">
      <n-button type="primary" size="small" :disabled="disabled" @click="$emit('generate')">
        开始生成
      </n-button>
      <n-button size="small" :disabled="disabled" @click="$emit('reload')">
        刷新列表
      </n-button>
    </div>
  </div>
</template>

<style scoped>
.config-section {
  display: flex;
  align-items: stretch;
  padding: 12px 18px 11px;
  background: var(--surface);
  border-bottom: 1px solid var(--border);
  flex-shrink: 0;
}

.config-panel {
  display: flex;
  flex-direction: column;
  gap: 7px;
  justify-content: center;
}

.config-panel-left  { flex: 0 0 490px; }
.config-panel-right { flex: 1 1 400px; min-width: 320px; }
.config-panel-actions { flex-shrink: 0; gap: 8px; }

.config-divider {
  width: 1px;
  background: var(--border);
  margin: 2px 18px;
  align-self: stretch;
  flex-shrink: 0;
}

.label {
  white-space: nowrap;
  font-size: 12px;
  color: var(--label);
  flex-shrink: 0;
  width: 56px;
  text-align: right;
}

.format-row {
  display: flex;
  align-items: center;
  gap: 7px;
}

.format-select,
.lang-select {
  flex: 1;
  min-width: 200px;
  max-width: 380px;
}

:deep(.n-base-selection) {
  --n-color:              #FFFFFF;
  --n-color-active:       #FFFFFF;
  --n-color-disabled:     #F5F5F7;
  --n-border:             1px solid #D1D1D6;
  --n-border-hover:       1px solid #AEAEB2;
  --n-border-active:      1px solid #007AFF;
  --n-border-focus:       1px solid #007AFF;
  --n-border-radius:      8px;
  --n-box-shadow-active:  0 0 0 3px rgba(0,122,255,0.18);
  --n-font-size:          12px;
  --n-placeholder-color:  #AEAEB2;
}

:deep(.n-button) {
  border-radius: 8px !important;
  font-weight: 500 !important;
  transition: background 0.15s ease, box-shadow 0.15s ease !important;
}

:deep(.n-button--primary-type) {
  --n-color:         #007AFF !important;
  --n-color-hover:   #0066DD !important;
  --n-color-pressed: #005EC5 !important;
  --n-border:        none !important;
  --n-border-hover:  none !important;
  --n-text-color:    #fff !important;
}

:deep(.n-button--default-type) {
  --n-color:         #FFFFFF !important;
  --n-color-hover:   #F5F5F7 !important;
  --n-color-pressed: #EBEBED !important;
  --n-border:        1px solid #D1D1D6 !important;
  --n-border-hover:  1px solid #AEAEB2 !important;
  --n-text-color:    #1C1C1E !important;
}
</style>
```

- [ ] **Step 2: 提交**

```bash
git add frontend/src/components/ConfigSection.vue
git commit -m "feat: add ConfigSection component with PathGroup"
```

---

## Task 6: 创建 `FileTable.vue`

**Files:**
- Create: `frontend/src/components/FileTable.vue`

- [ ] **Step 1: 新建文件**

```vue
<!-- frontend/src/components/FileTable.vue -->
<script setup>
import { h } from 'vue'
import { NCheckbox, NDataTable, NDropdown } from 'naive-ui'

const CONTEXT_MENU_OPTIONS = [
  { label: '打开文件所在目录', key: 'open-dir' },
  { label: '打开文件', key: 'open-file' },
]

const props = defineProps({
  fileList: Array,
  isExporting: Boolean,
  contextMenuShow: Boolean,
  contextMenuX: Number,
  contextMenuY: Number,
})

const emit = defineEmits([
  'row-contextmenu',
  'context-menu-select',
  'context-menu-close',
])

const statusChip = (text, type) =>
  h('span', { class: `status-chip status-chip--${type}` }, text)

const columns = [
  {
    key: 'selected',
    width: 40,
    title: '',
    render(row) {
      return h(NCheckbox, {
        checked: row.selected,
        disabled: props.isExporting,
        'onUpdate:checked': (val) => { row.selected = val },
      })
    },
  },
  { key: 'filename', title: '文件名', minWidth: 180, ellipsis: { tooltip: true } },
  {
    key: 'fileStatus',
    title: '文件状态',
    width: 110,
    align: 'center',
    render(row) {
      if (row.fileStatus === '就绪') return statusChip('就绪', 'ready')
      return statusChip('—', 'idle')
    },
  },
  {
    key: 'exportStatus',
    title: '导出状态',
    width: 110,
    align: 'center',
    render(row) {
      if (row.exportStatus === '导出中...') return statusChip('导出中', 'exporting')
      return statusChip('—', 'idle')
    },
  },
  {
    key: 'exportResult',
    title: '导出结果',
    minWidth: 120,
    align: 'center',
    render(row) {
      if (row.exportResult === '-') return statusChip('—', 'idle')
      if (row.exportResult === '待后端导出接口') return statusChip(row.exportResult, 'pending')
      return statusChip(row.exportResult, 'idle')
    },
  },
]

const tableRowProps = (row) => ({
  onContextmenu: (e) => {
    e.preventDefault()
    e.stopPropagation()
    if (props.isExporting) return
    emit('row-contextmenu', { row, x: e.clientX, y: e.clientY })
  },
})
</script>

<template>
  <div class="table-wrapper">
    <n-data-table
      :columns="columns"
      :data="fileList"
      :row-key="row => row.key"
      :row-props="tableRowProps"
      size="small"
      flex-height
      striped
      style="height: 100%"
    />
    <n-dropdown
      trigger="manual"
      :show="contextMenuShow"
      :options="CONTEXT_MENU_OPTIONS"
      :x="contextMenuX"
      :y="contextMenuY"
      placement="bottom-start"
      @select="$emit('context-menu-select', $event)"
      @clickoutside="$emit('context-menu-close')"
    />
  </div>
</template>

<style scoped>
.table-wrapper {
  flex: 1;
  overflow: hidden;
}

:deep(.n-data-table) {
  background: var(--bg) !important;
}

:deep(.n-data-table-th) {
  background: var(--surface) !important;
  font-size: 10px !important;
  text-transform: uppercase !important;
  letter-spacing: 0.08em !important;
  color: var(--text-2) !important;
  font-weight: 700 !important;
  border-bottom: 1px solid var(--border) !important;
  padding-top: 6px !important;
  padding-bottom: 6px !important;
}

:deep(.n-data-table-td) {
  font-size: 12px !important;
  border-bottom: 1px solid var(--border) !important;
  padding-top: 5px !important;
  padding-bottom: 5px !important;
}

:deep(.n-data-table-tr--striped .n-data-table-td) {
  background: rgba(0,0,0,0.018) !important;
}

:deep(.n-data-table-td:nth-child(2)) {
  font-family: var(--mono);
  font-size: 11.5px;
  color: #2558a0;
}

.status-chip {
  display: inline-flex;
  align-items: center;
  padding: 2px 8px;
  border-radius: 20px;
  font-size: 11px;
  font-weight: 500;
  letter-spacing: 0.02em;
}

.status-chip--ready     { background: rgba(52,199,89,0.12); color: #1a7a32; }
.status-chip--idle      { color: var(--text-3); font-size: 13px; line-height: 1; }
.status-chip--exporting { background: rgba(0,122,255,0.1); color: var(--blue); }
.status-chip--pending   { background: rgba(255,149,0,0.1); color: #8a5200; font-size: 10.5px; }
</style>
```

- [ ] **Step 2: 提交**

```bash
git add frontend/src/components/FileTable.vue
git commit -m "feat: add FileTable component with context menu"
```

---

## Task 7: 重构 `ExcelParserView.vue`

**Files:**
- Modify: `frontend/src/components/ExcelParserView.vue`

- [ ] **Step 1: 用以下内容完整替换 `ExcelParserView.vue`**

```vue
<!-- frontend/src/components/ExcelParserView.vue -->
<script setup>
import { useExcelParser } from '../composables/useExcelParser.js'
import ConfigSection from './ConfigSection.vue'
import OptionsBar from './OptionsBar.vue'
import FileTable from './FileTable.vue'
import StatusBar from './StatusBar.vue'

const {
  configPath, outputPath, translatePath, translateLang,
  serverFormats, clientFormats, langOptions,
  forceExport, compactOutput, prettyOutput,
  fileList,
  isExporting, statusText,
  isControlDisabled,
  totalCount, changedCount, selectedCount,
  selectAllChecked, selectAllIndeterminate,
  contextMenuShow, contextMenuX, contextMenuY,
  updateTranslateLang,
  openConfigPath, openOutputPath, openTranslatePath,
  reloadFiles, startGenerate,
  updateSelectAll,
  handleRowContextMenu,
  onSelectContextMenu, closeContextMenu,
} = useExcelParser()
</script>

<template>
  <div class="excel-parser-view">
    <ConfigSection
      :config-path="configPath"
      :output-path="outputPath"
      :translate-path="translatePath"
      :translate-lang="translateLang"
      :server-formats="serverFormats"
      :client-formats="clientFormats"
      :lang-options="langOptions"
      :disabled="isControlDisabled"
      @update:config-path="configPath = $event"
      @update:output-path="outputPath = $event"
      @update:translate-path="translatePath = $event"
      @update:translate-lang="updateTranslateLang"
      @update:server-formats="serverFormats = $event"
      @update:client-formats="clientFormats = $event"
      @browse-config="openConfigPath"
      @browse-output="openOutputPath"
      @browse-translate="openTranslatePath"
      @generate="startGenerate"
      @reload="reloadFiles"
    />
    <OptionsBar
      :select-all-checked="selectAllChecked"
      :select-all-indeterminate="selectAllIndeterminate"
      :force-export="forceExport"
      :compact-output="compactOutput"
      :pretty-output="prettyOutput"
      :disabled="isControlDisabled"
      :file-count="fileList.length"
      @update:select-all="updateSelectAll"
      @update:force-export="forceExport = $event"
      @update:compact-output="compactOutput = $event"
      @update:pretty-output="prettyOutput = $event"
    />
    <FileTable
      :file-list="fileList"
      :is-exporting="isExporting"
      :context-menu-show="contextMenuShow"
      :context-menu-x="contextMenuX"
      :context-menu-y="contextMenuY"
      @row-contextmenu="handleRowContextMenu"
      @context-menu-select="onSelectContextMenu"
      @context-menu-close="closeContextMenu"
    />
    <StatusBar
      :total-count="totalCount"
      :changed-count="changedCount"
      :selected-count="selectedCount"
      :status-text="statusText"
      :is-exporting="isExporting"
    />
  </div>
</template>

<style scoped>
/* ── CSS Tokens（子组件通过 CSS 继承使用）── */
.excel-parser-view {
  --blue:        #007AFF;
  --blue-hover:  #0066DD;
  --green:       #34C759;
  --orange:      #FF9500;

  --bg:          #F2F2F7;
  --surface:     #FFFFFF;
  --surface-2:   #F8F8FA;
  --border:      #E5E5EA;
  --border-dark: #D1D1D6;

  --text:        #1C1C1E;
  --text-2:      #3C3C3E;
  --text-3:      #6E6E73;
  --label:       #3C3C3E;

  --radius:      8px;
  --mono:        'Cascadia Code', 'JetBrains Mono', 'Consolas', monospace;

  display: flex;
  flex-direction: column;
  height: 100vh;
  width: 100%;
  background: var(--bg);
  color: var(--text);
}
</style>
```

- [ ] **Step 2: 启动应用验证功能正常**

```bash
cd frontend && npm run dev
```

在浏览器中确认：
- 页面正常渲染，无控制台报错
- 路径选择浏览按钮可点击
- 全选/单选复选框正常工作
- 右键菜单可弹出
- 状态栏数字正确显示

- [ ] **Step 3: 提交**

```bash
git add frontend/src/components/ExcelParserView.vue
git commit -m "refactor: decompose ExcelParserView into composable + sub-components"
```

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

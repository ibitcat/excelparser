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
  fileCount: { type: Number, default: 0 },
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

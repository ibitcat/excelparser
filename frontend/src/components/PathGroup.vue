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

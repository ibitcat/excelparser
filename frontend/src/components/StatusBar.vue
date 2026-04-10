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

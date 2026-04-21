<!-- frontend/src/components/FileTable.vue -->
<script setup>
import { h } from 'vue'
import { NCheckbox, NDataTable, NDropdown, NPopover } from 'naive-ui'

const CONTEXT_MENU_OPTIONS = [
  { label: '打开文件', key: 'open-file' },
  { label: '打开文件所在目录', key: 'open-dir' },
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
      //console.log('render exportStatus', row.filename, row.exportStatus)
      if (row.exportStatus === 1) {
        return statusChip('导出中', 'exporting')
      }else if (row.exportStatus === 2) {
        return statusChip('成功', 'ready')
      }else if (row.exportStatus === 3) {
        return statusChip('失败', 'error')
      }else if (row.exportStatus === 4) {
        return statusChip('跳过', 'pending')
      }else {
        return statusChip('—', 'idle')
      }
    },
  },
  {
    key: 'exportResult',
    title: '导出结果',
    minWidth: 120,
    align: 'center',
    render(row) {
      if (row.exportStatus === 3) {
        // 失败时，如果有多个错误，显示弹出层查看详情
        const errors = row.exportErrors || [row.exportResult]
        if (errors.length > 1) {
          return h(
            NPopover,
            {
              trigger: 'hover',
              placement: 'left',
              style: { maxWidth: '400px' },
            },
            {
              trigger: () =>
                h(
                  'span',
                  {
                    class: 'status-chip status-chip--error status-chip--clickable',
                    style: 'cursor: pointer',
                  },
                  `${errors.length} 个错误`
                ),
              default: () =>
                h('div', { class: 'error-list' }, [
                  h('div', { class: 'error-list-title' }, '错误详情：'),
                  ...errors.map((err, i) =>
                    h('div', { class: 'error-list-item' }, `${i + 1}. ${err}`)
                  ),
                ]),
            }
          )
        }
        return statusChip(row.exportResult, 'error')
      }else{
        return statusChip('—', 'idle')
      }
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

:deep(.status-chip) {
  display: inline-flex;
  align-items: center;
  padding: 2px 8px;
  border-radius: 20px;
  font-size: 11px;
  font-weight: 500;
  letter-spacing: 0.02em;
}

:deep(.status-chip--ready)     { background: rgba(52,199,89,0.12); color: #1a7a32; }
:deep(.status-chip--idle)      { color: var(--text-3); font-size: 13px; line-height: 1; }
:deep(.status-chip--exporting) { background: rgba(0,122,255,0.1); color: var(--blue); }
:deep(.status-chip--pending)   { background: rgba(255,149,0,0.1); color: #8a5200; font-size: 10.5px; }
:deep(.status-chip--error)     { background: rgba(255,59,48,0.12); color: #b2261e; }
:deep(.status-chip--clickable:hover) { background: rgba(255,59,48,0.2); }
</style>

<style>
/* 全局样式用于弹出层中的错误列表 */
.error-list {
  font-size: 12px;
  line-height: 1.5;
}
.error-list-title {
  font-weight: 600;
  margin-bottom: 6px;
  color: #b2261e;
}
.error-list-item {
  padding: 2px 0;
  color: #333;
  word-break: break-word;
}
.error-list-item:not(:last-child) {
  border-bottom: 1px dashed #eee;
  padding-bottom: 4px;
  margin-bottom: 4px;
}
</style>

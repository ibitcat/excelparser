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

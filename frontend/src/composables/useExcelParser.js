// frontend/src/composables/useExcelParser.js
import { ref, computed, watch, onMounted, onBeforeUnmount } from "vue";
import { useMessage } from "naive-ui";
import { Events } from "@wailsio/runtime";
import { FileService } from "../../bindings/excelparser/service/index.js";

const EXPORT_PROGRESS_EVENT = "export-progress";

export function useExcelParser() {
    const message = useMessage();

    // ── 配置路径与格式 ──
    const configPath = ref("");
    const outputPath = ref("");
    const translatePath = ref("");
    const translateLang = ref("");
    const serverFormats = ref([]);
    const clientFormats = ref([]);
    const langOptions = ref([]);

    // ── 导出选项 ──
    const forceExport = ref(false);
    const compactOutput = ref(false);
    const prettyOutput = ref(false);

    // ── 文件列表 ──
    const fileList = ref([]);

    // ── UI 状态 ──
    const isExporting = ref(false);
    const statusText = ref("就绪");
    const isLoadingConfig = ref(true);
    let saveConfigTimer = null;

    // ── 右键菜单（contextMenuRow 为内部状态，不暴露给模板）──
    const contextMenuShow = ref(false);
    const contextMenuX = ref(0);
    const contextMenuY = ref(0);
    const contextMenuRow = ref(null);
    let offExportProgress = null;

    // ── Computed ──
    const isControlDisabled = computed(() => isExporting.value);

    const totalCount = computed(() => fileList.value.length);
    const changedCount = computed(() => fileList.value.filter((f) => f.fileStatus !== "-").length);
    const selectedCount = computed(() => fileList.value.filter((f) => f.selected).length);

    const selectAllChecked = computed({
        get() {
            return fileList.value.length > 0 && fileList.value.every((f) => f.selected);
        },
        set(checked) {
            fileList.value.forEach((f) => {
                f.selected = checked;
            });
        },
    });

    const selectAllIndeterminate = computed(() => {
        if (fileList.value.length === 0) return false;
        const selected = fileList.value.filter((f) => f.selected).length;
        return selected > 0 && selected < fileList.value.length;
    });

    // ── 内部方法 ──
    const ensureLangOption = (lang) => {
        if (!lang) return;
        const exists = langOptions.value.some((item) => item.value === lang);
        if (!exists) {
            langOptions.value = [...langOptions.value, { label: lang, value: lang }];
        }
    };

    const persistConfig = async () => {
        try {
            await FileService.SaveConfig({
                config_path: configPath.value,
                output_path: outputPath.value,
                i18n_path: translatePath.value,
                i18n_lang: translateLang.value,
                server_fmts: serverFormats.value,
                client_fmts: clientFormats.value,
            });
        } catch (err) {
            console.error("保存配置失败:", err);
        }
    };

    const scheduleSaveConfig = () => {
        if (isLoadingConfig.value) return;
        if (saveConfigTimer) clearTimeout(saveConfigTimer);
        saveConfigTimer = setTimeout(() => {
            persistConfig();
        }, 300);
    };

    const loadXlsxList = async (path) => {
        if (!path) {
            fileList.value = [];
            return;
        }
        try {
            const list = await FileService.GetXlsxList(path);
            fileList.value = (list || []).map((item) => ({
                key: item.path,
                selected: true,
                filename: item.name,
                filepath: item.path,
                fileStatus: item.need_parse ? "就绪" : "-",
                exportStatus: 0,
                exportResult: "-",
                exportErrors: [],
            }));
        } catch (err) {
            message.error(`加载配置表失败: ${String(err)}`);
        }
    };

    const loadLangOptions = async () => {
        try {
            const langs = await FileService.GetTranslationList();
            langOptions.value = (langs || []).map((lang) => ({ label: lang, value: lang }));
            ensureLangOption(translateLang.value);
        } catch (err) {
            console.error("加载语言选项失败:", err);
        }
    };

    const loadConfig = async () => {
        try {
            const config = await FileService.GetConfig();
            if (config) {
                configPath.value = config.config_path || "";
                outputPath.value = config.output_path || "";
                translatePath.value = config.i18n_path || "";
                translateLang.value = config.i18n_lang || "";
                serverFormats.value = config.server_fmts || [];
                clientFormats.value = config.client_fmts || [];
                await loadLangOptions();
                await loadXlsxList(configPath.value);
            }
        } catch (err) {
            console.error("加载配置失败:", err);
        } finally {
            isLoadingConfig.value = false;
        }
    };

    // ── 暴露的方法 ──
    const updateTranslateLang = (lang) => {
        translateLang.value = lang;
        ensureLangOption(lang);
    };

    const closeContextMenu = () => {
        contextMenuShow.value = false;
        contextMenuRow.value = null;
    };

    const handleRowContextMenu = ({ row, x, y }) => {
        if (isControlDisabled.value) return;
        contextMenuRow.value = row;
        contextMenuX.value = x;
        contextMenuY.value = y;
        contextMenuShow.value = true;
    };

    const onSelectContextMenu = async (key) => {
        try {
            if (key === "open-dir") {
                if (!contextMenuRow.value?.filepath) {
                    message.warning("未找到文件路径");
                    return;
                }
                await FileService.OpenFileDirectory(contextMenuRow.value.filepath);
                message.success("已打开所在目录");
            } else if (key === "open-file") {
                if (!contextMenuRow.value?.filepath) {
                    message.warning("未找到文件路径");
                    return;
                }
                await FileService.OpenFile(contextMenuRow.value.filepath);
                message.success("已打开文件");
            }
        } catch (err) {
            message.error(`操作失败: ${String(err)}`);
        } finally {
            closeContextMenu();
        }
    };

    const selectPath = async (pathType, title) => {
        try {
            const selected = await FileService.SelectDirectory(pathType, title);
            return selected || "";
        } catch (err) {
            message.error("选择目录失败!");
            return "";
        }
    };

    const openConfigPath = async () => {
        const dir = await selectPath(1, "选择配置路径");
        if (!dir) return;
        configPath.value = dir;
        await loadXlsxList(dir);
    };

    const openOutputPath = async () => {
        const dir = await selectPath(2, "选择输出路径");
        if (!dir) return;
        outputPath.value = dir;
    };

    const openTranslatePath = async () => {
        const dir = await selectPath(3, "选择翻译路径");
        if (!dir) return;
        translatePath.value = dir;
        await loadLangOptions();
    };

    const reloadFiles = async () => {
        if (!configPath.value) {
            message.warning("请先选择有效的配置路径");
            return;
        }
        await loadXlsxList(configPath.value);
        message.success("配置表列表已刷新");
    };

    const validateExport = () => {
        if (!configPath.value) {
            message.warning("请先选择配置路径");
            return false;
        }
        if (!outputPath.value) {
            message.warning("请先选择导出路径");
            return false;
        }
        if (serverFormats.value.length === 0 && clientFormats.value.length === 0) {
            message.warning("请至少勾选一种导出格式");
            return false;
        }
        if (selectedCount.value === 0) {
            message.warning("请至少勾选一个要导出的文件");
            return false;
        }
        return true;
    };

    const startGenerate = async () => {
        if (isExporting.value) {
            message.warning("正在导出中，请稍后");
            return;
        }
        if (!validateExport()) return;

        isExporting.value = true;
        statusText.value = "导出中...";
        fileList.value.forEach((row) => {
            row.exportStatus = 0;
            row.exportResult = "-";
            row.exportErrors = [];
        });

        try {
            await FileService.StartExport();
            message.success("导出已完成");
        } catch (err) {
            message.error(`导出失败: ${String(err)}`);
        } finally {
            isExporting.value = false;
            statusText.value = "就绪";
        }
    };

    const updateSelectAll = (checked) => {
        selectAllChecked.value = checked;
    };

    // ── Watch ──
    watch(compactOutput, (val) => {
        if (val) prettyOutput.value = false;
        FileService.SetExportFlag(1, val).catch((err) => console.error("SetExportFlag compact:", err));
    });
    watch(prettyOutput, (val) => {
        if (val) compactOutput.value = false;
        FileService.SetExportFlag(2, val).catch((err) => console.error("SetExportFlag pretty:", err));
    });
    watch(forceExport, (val) => {
        FileService.SetExportFlag(3, val).catch((err) => console.error("SetExportFlag force:", err));
    });
    watch(
        serverFormats,
        (val) => {
            FileService.SetExportFormat("server", val).catch((err) => console.error("SetExportFormat server:", err));
        },
        { deep: true },
    );
    watch(
        clientFormats,
        (val) => {
            FileService.SetExportFormat("client", val).catch((err) => console.error("SetExportFormat client:", err));
        },
        { deep: true },
    );
    watch(translateLang, (val) => {
        FileService.SetI18nLang(val).catch((err) => console.error("SetI18nLang:", err));
    });
    watch([configPath, outputPath, translatePath, serverFormats, clientFormats, translateLang], scheduleSaveConfig, { deep: true });

    // 按文件路径缓存事件，收齐 start + finish 后立即处理
    const fileEvents = new Map(); // path -> { start?, finish? }

    const tryFlushFile = (path) => {
        const events = fileEvents.get(path);
        if (!events || !events.start || !events.finish) return;

        const row = fileList.value.find((item) => item.filepath === path);
        if (row) {
            // 按 seq 顺序应用：先 start 后 finish
            const sorted = [events.start, events.finish].sort((a, b) => a.seq - b.seq);
            for (const payload of sorted) {
                row.exportStatus = payload.status;
                row.exportResult = payload.message;
                // 存储所有错误信息
                if (payload.messages && payload.messages.length > 0) {
                    row.exportErrors = payload.messages;
                }
            }
            console.log(`row updated for ${path}:`, row.exportStatus, row.exportResult);
        }
        fileEvents.delete(path);
    };

    const handleExportProgress = (event) => {
        const payload = event?.data || event;
        if (!payload || !payload.stage) return;

        console.log("Export Progress:", payload.stage, payload.path, "seq:", payload.seq);

        if (payload.stage === "start" || payload.stage === "finish") {
            if (!fileEvents.has(payload.path)) {
                fileEvents.set(payload.path, {});
            }
            fileEvents.get(payload.path)[payload.stage] = payload;
            tryFlushFile(payload.path);
            return;
        }

        if (payload.stage === "error") {
            statusText.value = "导出失败";
        }

        if (payload.stage === "done") {
            fileEvents.clear();
        }
    };

    // ── 生命周期 ──
    onMounted(() => {
        loadConfig();
        offExportProgress = Events.On(EXPORT_PROGRESS_EVENT, handleExportProgress);
    });
    onBeforeUnmount(() => {
        if (saveConfigTimer) clearTimeout(saveConfigTimer);
        if (offExportProgress) offExportProgress();
    });

    return {
        // State
        configPath,
        outputPath,
        translatePath,
        translateLang,
        serverFormats,
        clientFormats,
        langOptions,
        forceExport,
        compactOutput,
        prettyOutput,
        fileList,
        isExporting,
        statusText,
        contextMenuShow,
        contextMenuX,
        contextMenuY,
        // Computed
        isControlDisabled,
        totalCount,
        changedCount,
        selectedCount,
        selectAllChecked,
        selectAllIndeterminate,
        // Methods
        updateTranslateLang,
        openConfigPath,
        openOutputPath,
        openTranslatePath,
        reloadFiles,
        startGenerate,
        updateSelectAll,
        handleRowContextMenu,
        onSelectContextMenu,
        closeContextMenu,
    };
}

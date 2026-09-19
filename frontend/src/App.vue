<template>
  <div class="app-container">
    <!-- Global App Toast Notification -->
    <transition name="toast-fade">
      <div v-if="toastMessage" class="global-toast" :class="toastType">
        <span class="toast-icon">
          <template v-if="toastType === 'warning'">⚠️</template>
          <template v-else-if="toastType === 'error'">❌</template>
          <template v-else-if="toastType === 'success'">🎉</template>
          <template v-else>ℹ️</template>
        </span>
        <span class="toast-text">{{ toastMessage }}</span>
        <button class="toast-close" @click="dismissToast">✕</button>
      </div>
    </transition>

    <!-- Header Component -->
    <AppHeader
      :activeMode="activeMode"
      :isLogCardVisible="isLogCardVisible"
      :currentLang="currentLang"
      @select-mode="selectMode"
      @toggle-log="toggleLogCard"
      @open-settings="tab => openSettings(tab as any)"
      @open-about="isAboutOpen = true"
      @select-lang="selectLanguage"
    />

    <!-- Main Grid -->
    <main class="content-grid">
      <!-- Left: Disk Selection Panel Component -->
      <DiskPanel
        :diskList="diskList"
        :selectionMode="selectionMode"
        :selectedDisk="selectedDisk"
        :selectedDevices="selectedDevices"
        :isScanningDisks="isScanningDisks"
        :customIcons="customIcons"
        @set-selection-mode="setSelectionMode"
        @select-all="selectAllDisks"
        @deselect-all="deselectAllDisks"
        @batch-eject="handleBatchEjectDisks"
        @select-disk="onDiskSelect"
        @toggle-disk="onDiskToggle"
        @pick-icon="openIconPicker"
        @inspect-disk="openInspector"
        @eject-disk="handleEjectDisk"
        @refresh-disks="refreshDisks"
      />

      <!-- Right: Deployment & Testing Panel Component -->
      <DeployPanel
        :activeMode="activeMode"
        :selectionMode="selectionMode"
        :selectedDisk="selectedDisk"
        :selectedDevices="selectedDevices"
        :diskList="diskList"
        v-model:selectedFsType="selectedFsType"
        :isNonDestructive="isNonDestructive"
        :isMacOs="isMacOs"
        :ventoyStatus="ventoyStatus"
        :selectedIsoFiles="selectedIsoFiles"
        :isDeploying="isDeploying"
        :deployProgress="deployProgress"
        :speedMBps="deploySpeedMBps"
        :elapsedSec="deployElapsedSec"
        :etaSec="deployEtaSec"
        :deployBtnText="deployBtnText"
        :deployDisabledReason="deployDisabledReason"
        :showDeploySuccessBanner="showDeploySuccessBanner"
        :deploySuccessBanner="deploySuccessBanner"
        :hypervisorList="hypervisorList"
        v-model:selectedBootMode="selectedBootMode"
        v-model:selectedVMType="selectedVMType"
        v-model:vmCpuCores="vmCpuCores"
        v-model:vmMemoryMB="vmMemoryMB"
        v-model:vmDisplayAccel="vmDisplayAccel"
        :isVmDisabled="isVmDisabled"
        :vmDisabledReason="vmDisabledReason"
        :isLaunchingQemu="isLaunchingQemu"
        :activeVmTargetName="activeVmTargetName"
        :activeVmTargetDevice="activeVmTargetDevice"
        @open-settings-ventoy="openSettings('ventoy')"
        @select-iso="handleSelectIsoFiles"
        @drop-iso-paths="addIsoFilesByPaths"
        @remove-iso="removeIsoFile"
        @clear-iso="clearIsoFiles"
        @deploy-click="handleDeployBtnClick"
        @cancel-deploy="handleCancelDeploy"
        @dismiss-success-banner="dismissDeploySuccessBanner"
        @safely-eject-success="handleSafelyEjectAfterDeploy"
        @launch-vm="launchVM"
      />

      <!-- Embedded Log Center Component -->
      <LogPanel
        :isVisible="isLogCardVisible"
        v-model:autoScroll="embeddedAutoScroll"
        v-model:currentLogFilter="currentEmbeddedLogFilter"
        :filteredLogs="filteredEmbeddedLogs"
        :logLevels="logLevels"
        @copy-logs="handleCopyEmbeddedLogs"
        @export-logs="handleExportEmbeddedLogs"
        @clear-logs="handleClearEmbeddedLogs"
        @close-log="toggleLogCard"
      />
    </main>

    <!-- Icon Picker Modal -->
    <IconPickerModal
      :isOpen="isPickerOpen"
      :diskName="targetPickerDisk?.name || targetPickerDisk?.device || ''"
      :currentIcon="targetPickerDisk ? (customIcons[getDiskFingerprint(targetPickerDisk)] || customIcons[targetPickerDisk.device]) : undefined"
      @close="isPickerOpen = false"
      @select-icon="onIconSelected"
      @reset-icon="onIconReset"
    />

    <!-- USB Hardware Inspector Modal -->
    <UsbInspectorModal
      :isOpen="isInspectorOpen"
      :disk="targetInspectorDisk"
      @close="isInspectorOpen = false"
    />

    <!-- High-Risk Format Confirmation Modal -->
    <DeployConfirmModal
      :isOpen="isDeployConfirmOpen"
      :mode="activeMode"
      :fsType="selectedFsType"
      :targetDisk="selectedDisk"
      :targetDisks="pendingTargets"
      :allDisks="diskList"
      @close="isDeployConfirmOpen = false"
      @confirm="startDeployment"
    />

    <!-- Settings & GitHub Proxy Modal -->
    <SettingsModal
      :isOpen="isSettingsOpen"
      :initialTab="settingsInitialTab"
      :currentProxy="currentGithubProxy"
      @close="isSettingsOpen = false"
      @save="onSaveSettings"
    />

    <!-- Ventoy Missing Alert Modal -->
    <VentoyAlertModal
      :isOpen="isVentoyAlertOpen"
      :title="ventoyAlertTitle"
      :message="ventoyAlertMessage"
      :actionType="ventoyAlertAction"
      @close="isVentoyAlertOpen = false"
      @action="handleVentoyAlertAction"
      @switch-b="handleVentoyAlertSwitchB"
    />

    <!-- Diagnostics Modal -->
    <DiagnosticsModal
      :isOpen="isDiagnosticsOpen"
      :diagnostics="currentDiagnostics"
      :errorMsg="currentDiagErrorMsg"
      @close="isDiagnosticsOpen = false"
      @retry="handleRetryDeploy"
      @copy-report="handleCopyReport"
    />

    <!-- About Modal -->
    <AboutModal
      :show="isAboutOpen"
      @close="isAboutOpen = false"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue';
import AppHeader from './components/AppHeader.vue';
import DiskPanel from './components/DiskPanel.vue';
import DeployPanel from './components/DeployPanel.vue';
import LogPanel from './components/LogPanel.vue';
import IconPickerModal from './components/IconPickerModal.vue';
import UsbInspectorModal from './components/UsbInspectorModal.vue';
import DeployConfirmModal from './components/DeployConfirmModal.vue';
import SettingsModal from './components/SettingsModal.vue';
import VentoyAlertModal from './components/VentoyAlertModal.vue';
import DiagnosticsModal from './components/DiagnosticsModal.vue';
import AboutModal from './components/AboutModal.vue';
import type { LogItem } from './components/LogViewerModal.vue';
import { t, currentLang, setLanguage } from './i18n';
import { formatLogsToText } from './utils/logFormatter';
import { logUserAction } from './utils/logger';
import { useIsoManager } from './composables/useIsoManager';
import { useVirtualMachine } from './composables/useVirtualMachine';
import { useDiskSelection, getDiskFingerprint } from './composables/useDiskSelection';
import { useDeployment } from './composables/useDeployment';
import {
  ExportLogs,
  GetRecentLogs,
  ClearLogs,
  GetConfig,
  SaveConfig,
  ReloadAppMenu,
} from '../wailsjs/go/main/App';

function selectLanguage(langVal: string) {
  setLanguage(langVal);
  saveLangToConfig(langVal);
  ReloadAppMenu(langVal).catch((err: any) => {
    console.warn('Failed to reload app menu:', err);
  });
}

async function saveLangToConfig(langVal: string) {
  try {
    const cfg = await GetConfig();
    if (cfg) {
      cfg.language = langVal;
      await SaveConfig(cfg);
    }
  } catch (e) {
    console.error('Failed to save language config:', e);
  }
}




const toastMessage = ref('');
const toastType = ref<'info' | 'warning' | 'error' | 'success'>('info');
let toastTimer: number | undefined;

function showToast(msg: string, type: 'info' | 'warning' | 'error' | 'success' = 'info') {
  const cleanMsg = msg ? msg.replace(/^[\s\uFE0F]*[⚠️❌🎉ℹ️✅🚨⚡️❗][\s\uFE0F]*/, '').trim() : '';
  toastMessage.value = cleanMsg || msg;
  toastType.value = type;
  if (toastTimer) clearTimeout(toastTimer);
  toastTimer = window.setTimeout(() => {
    toastMessage.value = '';
  }, 4000);
}


const settingsInitialTab = ref<'general' | 'network' | 'uniboot' | 'ventoy'>('general');
const isSettingsOpen = ref(false);
const isAboutOpen = ref(false);
const currentGithubProxy = ref('');

function openSettings(tab: 'general' | 'network' | 'uniboot' | 'ventoy' = 'general') {
  settingsInitialTab.value = tab;
  isSettingsOpen.value = true;
  logUserAction('INFO', 'User opened settings modal', tab);
}

const {
  selectedIsoFiles,
  isoCopyStatus,
  addIsoFilesByPaths,
  handleSelectIsoFiles,
  removeIsoFile,
  clearIsoFiles,
} = useIsoManager(showToast);

const {
  selectionMode,
  diskList,
  selectedDisk,
  selectedDevices,
  customIcons,
  isPickerOpen,
  targetPickerDisk,
  isInspectorOpen,
  targetInspectorDisk,
  isScanningDisks,
  setSelectionMode,
  selectAllDisks,
  deselectAllDisks,
  onDiskSelect,
  onDiskToggle,
  openInspector,
  openIconPicker,
  onIconSelected,
  onIconReset,
  handleEjectDisk,
  handleBatchEjectDisks,
  handleSafelyEjectAfterDeploy,
  refreshDisks,
} = useDiskSelection({
  t,
  showToast,
  getTargetsToEject: () => deploySuccessBanner.value.targets,
  onEjectSuccess: (device) => {
    if (deploySuccessBanner.value.targets.includes(device)) {
      deploySuccessBanner.value.dismissed = true;
      deploySuccessBanner.value.visible = false;
    }
  },
  onSafelyEjectSuccess: () => {
    deploySuccessBanner.value.dismissed = true;
    deploySuccessBanner.value.visible = false;
  },
});

const {
  activeMode,
  selectedFsType,
  autoEjectAfterDeploy,
  isDeploying,
  deployProgress,
  deploySpeedMBps,
  deployElapsedSec,
  deployEtaSec,
  isDeployConfirmOpen,
  pendingTargets,
  isVentoyAlertOpen,
  ventoyAlertTitle,
  ventoyAlertMessage,
  ventoyAlertAction,
  ventoyStatus,
  isDiagnosticsOpen,
  currentDiagnostics,
  currentDiagErrorMsg,
  deploySuccessBanner,
  isMacOs,
  isNonDestructive,
  deployBtnText,
  deployDisabledReason,
  activeVmTargetDevice,
  activeVmTargetName,
  showDeploySuccessBanner,
  checkVentoyStatus,
  selectMode,
  handleVentoyAlertAction,
  handleVentoyAlertSwitchB,
  handleCopyReport,
  handleRetryDeploy,
  handleDeployBtnClick,
  startDeployment,
  handleCancelDeploy,
  dismissDeploySuccessBanner,
} = useDeployment({
  t,
  showToast,
  selectionMode,
  selectedDisk,
  selectedDevices,
  diskList,
  selectedIsoFiles,
  refreshDisks,
  openSettings,
});

const runtimeLogs = ref<LogItem[]>([]);

const savedLogCardVisible = localStorage.getItem('unigodesktop_log_card_visible');
const isLogCardVisible = ref(savedLogCardVisible !== null ? savedLogCardVisible === 'true' : true);

function toggleLogCard() {
  isLogCardVisible.value = !isLogCardVisible.value;
  localStorage.setItem('unigodesktop_log_card_visible', String(isLogCardVisible.value));
}

const savedAutoScroll = localStorage.getItem('unigodesktop_embedded_log_autoscroll');
const embeddedAutoScroll = ref(savedAutoScroll !== null ? savedAutoScroll === 'true' : true);
const currentEmbeddedLogFilter = ref<string>('ALL');

watch(embeddedAutoScroll, (val) => {
  localStorage.setItem('unigodesktop_embedded_log_autoscroll', String(val));
});

const logLevels = computed(() => [
  { key: 'ALL', label: t('log.level_all') },
  { key: 'INFO', label: t('log.level_info') },
  { key: 'WARN', label: t('log.level_warn') },
  { key: 'ERROR', label: t('log.level_error') },
  { key: 'DEBUG', label: t('log.level_debug') }
]);

const filteredEmbeddedLogs = computed(() => {
  return runtimeLogs.value.filter(log => {
    if (currentEmbeddedLogFilter.value === 'ALL') return true;
    return (log.level || '').toUpperCase() === currentEmbeddedLogFilter.value;
  });
});

function handleCopyEmbeddedLogs() {
  logUserAction('INFO', 'User copied embedded logs to clipboard');
  if (filteredEmbeddedLogs.value.length === 0) {
    showToast(t('log.empty'), 'info');
    return;
  }
  const text = formatLogsToText(filteredEmbeddedLogs.value);
  navigator.clipboard.writeText(text);
  showToast(t('log.copied_toast'), 'success');
}

async function handleExportEmbeddedLogs() {
  logUserAction('INFO', 'User exported embedded logs');
  if (filteredEmbeddedLogs.value.length === 0) {
    showToast(t('log.empty'), 'info');
    return;
  }
  const text = formatLogsToText(filteredEmbeddedLogs.value);

  try {
    const filePath = await ExportLogs(
      text,
      t('dialog.exportTitle'),
      t('dialog.logFilesFilter'),
      t('dialog.textFilesFilter'),
      t('dialog.allFilesFilter')
    );
    if (filePath) {
      showToast(t('log.exported_path_toast', { path: filePath }), 'success');
    }
  } catch (e) {
    console.error('Failed to export logs via native Wails dialog:', e);
    // Fallback for web browser mode
    const blob = new Blob([text], { type: 'text/plain;charset=utf-8' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `unigodesktop-log-${new Date().toISOString().slice(0, 10)}.log`;
    a.click();
    URL.revokeObjectURL(url);
    showToast(t('log.exported_toast'), 'success');
  }
}

function handleClearEmbeddedLogs() {
  runtimeLogs.value = [];
  ClearLogs().catch((err: any) => {
    console.error('Failed to clear backend log buffer:', err);
  });
}

function dismissToast() {
  toastMessage.value = '';
  logUserAction('DEBUG', 'User dismissed toast notification');
}



function applyTheme(themeName?: string) {
  const theme = themeName === 'light' ? 'light' : 'dark';
  document.documentElement.setAttribute('data-theme', theme);
  // Cache theme in localStorage so the inline script in index.html can
  // apply it synchronously on next launch before any render, preventing flash.
  try {
    localStorage.setItem('unigo_theme_cache', theme);
  } catch (e) {
    // localStorage unavailable, ignore
  }
}

async function loadConfig() {
  if (window.go && window.go.main && window.go.main.App) {
    try {
      const cfg = await window.go.main.App.GetConfig();
      if (cfg) {
        if (cfg.githubProxy) currentGithubProxy.value = cfg.githubProxy;
        if (cfg.fileSystem) selectedFsType.value = cfg.fileSystem as any;
        if (cfg.language) {
          setLanguage(cfg.language);
          if (window.go.main.App.ReloadAppMenu) {
            window.go.main.App.ReloadAppMenu(cfg.language).catch(() => {});
          }
        }
        applyTheme(cfg.theme);
        autoEjectAfterDeploy.value = cfg.autoEjectAfterDeploy === true;
      }
    } catch (e) {
      console.error('Failed to load config:', e);
    }
  }
}

async function onSaveSettings(payload: any) {
  let proxyUrl = '';
  if (typeof payload === 'string') {
    proxyUrl = payload;
    currentGithubProxy.value = payload;
  } else if (payload && typeof payload === 'object') {
    proxyUrl = payload.githubProxy || '';
    currentGithubProxy.value = proxyUrl;
    if (payload.fileSystem) selectedFsType.value = payload.fileSystem as any;
    if (payload.mode) activeMode.value = payload.mode as any;
    if (payload.theme) applyTheme(payload.theme);
    if (typeof payload.autoEjectAfterDeploy === 'boolean') autoEjectAfterDeploy.value = payload.autoEjectAfterDeploy;
  }

  if (window.go && window.go.main && window.go.main.App) {
    try {
      const configObj = typeof payload === 'object' && payload !== null ? {
        mode: payload.mode || activeMode.value,
        autoCheckUpdate: payload.autoCheckUpdate !== false,
        theme: payload.theme || 'dark',
        language: payload.language || 'auto',
        githubProxy: proxyUrl,
        fileSystem: payload.fileSystem || selectedFsType.value,
        proxyProtocol: payload.proxyProtocol || 'direct',
        proxyHost: payload.proxyHost || '',
        proxyPort: Number(payload.proxyPort) || 0,
        proxyUser: payload.proxyUser || '',
        proxyPassword: payload.proxyPassword || '',
        ventoyPath: payload.ventoyPath || '',
        ventoySecureBoot: payload.ventoySecureBoot !== false,
        ventoyPartitionStyle: payload.ventoyPartitionStyle || 'MBR',
        ventoyReserveSpace: Number(payload.ventoyReserveSpace) || 0,
        ventoyWin11Bypass: payload.ventoyWin11Bypass === true,
        ventoyMenuTimeout: Number(payload.ventoyMenuTimeout) || 0,
        autoEjectAfterDeploy: payload.autoEjectAfterDeploy === true,
      } : {
        mode: activeMode.value,
        autoCheckUpdate: true,
        theme: 'dark',
        githubProxy: proxyUrl,
        fileSystem: selectedFsType.value,
      };

      await window.go.main.App.SaveConfig(configObj as any);
      if (configObj.language && window.go.main.App.ReloadAppMenu) {
        await window.go.main.App.ReloadAppMenu(configObj.language);
      }
    } catch (e) {
      console.error('Failed to save config:', e);
    }
  }
}



const {
  hypervisorList,
  selectedBootMode,
  selectedVMType,
  vmCpuCores,
  vmMemoryMB,
  vmDisplayAccel,
  isLaunchingQemu,
  isVmDisabled,
  vmDisabledReason,
  checkQemu,
  launchVM,
} = useVirtualMachine({
  activeVmTargetDevice,
  diskList,
  isDeploying,
  showToast,
});






onMounted(() => {
  loadConfig();
  refreshDisks();
  checkQemu();
  checkVentoyStatus();

  GetRecentLogs().then((logs: any[]) => {
    if (logs && logs.length > 0) {
      runtimeLogs.value = logs.map((entry: any) => ({
        id: entry.id,
        timestamp: entry.timestamp,
        level: entry.level || 'INFO',
        message: entry.message || '',
        details: entry.details || ''
      }));
    } else {
      runtimeLogs.value = [{
        timestamp: new Date().toISOString(),
        level: 'INFO',
        message: 'UniGoDesktop engine ready. Real-time log stream connected.'
      }];
    }
  }).catch(() => {
    runtimeLogs.value = [{
      timestamp: new Date().toISOString(),
      level: 'INFO',
      message: 'UniGoDesktop engine ready. Real-time log stream connected.'
    }];
  });

  if (window.runtime && window.runtime.EventsOn) {
    window.runtime.EventsOn("log:entry", (entry: any) => {
      if (entry) {
        runtimeLogs.value.push({
          id: entry.id,
          timestamp: entry.timestamp,
          level: entry.level || 'INFO',
          message: entry.message || '',
          details: entry.details || ''
        });
        if (runtimeLogs.value.length > 500) {
          runtimeLogs.value.shift();
        }
      }
    });
    window.runtime.EventsOn("iso-copy-progress", (data: any) => {
      if (data) {
        isoCopyStatus.value = t('disk.writingImageProgress', { fileIndex: data.fileIndex, totalFiles: data.totalFiles, currentFile: data.currentFile, progress: data.progress.toFixed(1) });
        deployProgress.value = Math.min(99, Math.max(50, Math.floor(50 + data.progress / 2)));
        if (data.speedMBps !== undefined) deploySpeedMBps.value = data.speedMBps;
        if (data.elapsedSec !== undefined) deployElapsedSec.value = data.elapsedSec;
        if (data.etaSec !== undefined) deployEtaSec.value = data.etaSec;
      }
    });
    window.runtime.EventsOn("disk-list-changed", () => {
      if (!isDeploying.value) {
        refreshDisks();
      }
    });
    window.runtime.EventsOn("open-about-modal", () => {
      isAboutOpen.value = true;
    });
    window.runtime.EventsOn("open-log-modal", () => {
      const el = document.querySelector('.log-section-card');
      if (el) {
        el.scrollIntoView({ behavior: 'smooth' });
      }
    });

    if (typeof window.runtime.OnFileDrop === 'function') {
      window.runtime.OnFileDrop((_x: number, _y: number, paths: string[]) => {
        if (paths && paths.length > 0) {
          addIsoFilesByPaths(paths);
        }
      }, false);
    }
  }
});

onUnmounted(() => {
  if (window.runtime && typeof window.runtime.OnFileDropOff === 'function') {
    window.runtime.OnFileDropOff();
  }
});


</script>

<style scoped>
.app-container {
  max-width: 1280px;
  width: 95%;
  margin: 0 auto;
  padding: 2.2rem;
}

.content-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
  gap: 1.75rem;
}

@media (max-width: 960px) {
  .content-grid {
    grid-template-columns: minmax(0, 1fr);
  }
}

/* Global App Toast Notification Styles */
.global-toast {
  position: fixed;
  top: 24px;
  left: 50%;
  transform: translateX(-50%);
  z-index: 99999;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 24px;
  border-radius: 12px;
  font-size: 0.9rem;
  font-weight: 600;
  color: var(--text-main);
  box-shadow: 0 12px 32px var(--modal-backdrop);
  backdrop-filter: blur(16px);
  -webkit-backdrop-filter: blur(16px);
  border: 1px solid var(--card-border);
  background: var(--card-bg);
}

.global-toast.warning {
  background: var(--alert-warning-bg);
  border-color: var(--alert-warning-border);
  color: var(--alert-warning-title);
}

.global-toast.error {
  background: var(--alert-danger-bg);
  border-color: var(--alert-danger-border);
  color: var(--alert-danger-title);
}

.global-toast.success {
  background: var(--alert-success-bg);
  border-color: var(--alert-success-border);
  color: var(--alert-success-title);
}

.global-toast.info {
  background: var(--alert-info-bg);
  border-color: var(--alert-info-border);
  color: var(--alert-info-title);
}

.toast-close {
  background: transparent;
  border: none;
  color: currentColor;
  font-size: 1rem;
  cursor: pointer;
  opacity: 0.8;
  margin-left: 8px;
  transition: opacity 0.2s ease;
}

.toast-close:hover {
  opacity: 1;
}

.toast-fade-enter-active,
.toast-fade-leave-active {
  transition: opacity 0.3s cubic-bezier(0.16, 1, 0.3, 1), transform 0.3s cubic-bezier(0.16, 1, 0.3, 1);
}

.toast-fade-enter-from,
.toast-fade-leave-to {
  opacity: 0;
  transform: translate(-50%, -20px);
}
</style>

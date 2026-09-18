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
import { ref, computed, onMounted, watch } from 'vue';
import AppHeader from './components/AppHeader.vue';
import DiskPanel from './components/DiskPanel.vue';
import type { DiskInfo } from './components/DiskPanel.vue';
import DeployPanel from './components/DeployPanel.vue';
import LogPanel from './components/LogPanel.vue';
import IconPickerModal, { DiskIconType } from './components/IconPickerModal.vue';
import UsbInspectorModal from './components/UsbInspectorModal.vue';
import DeployConfirmModal from './components/DeployConfirmModal.vue';
import SettingsModal from './components/SettingsModal.vue';
import VentoyAlertModal from './components/VentoyAlertModal.vue';
import DiagnosticsModal, { InstallDiagnosticsData } from './components/DiagnosticsModal.vue';
import AboutModal from './components/AboutModal.vue';
import type { LogItem } from './components/LogViewerModal.vue';
import { t, currentLang, setLanguage } from './i18n';
import { formatLogsToText } from './utils/logFormatter';
import {
  ExportLogs,
  SelectIsoFiles,
  GetRecentLogs,
  ClearLogs,
  GetDiskList,
  EjectDisk,
  GetConfig,
  SaveConfig,
  ReloadAppMenu,
  ValidateVentoyCli,
  CheckQEMU,
  DetectHypervisors,
  DeployCloudMode,
  DeployCloudModeBatch,
  DeployHybridMode,
  DeployHybridModeBatch,
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




const activeMode = ref<'cloud' | 'hybrid'>('cloud');
const selectionMode = ref<'single' | 'batch'>('single');
const selectedFsType = ref<'exFAT' | 'NTFS' | 'FAT32' | 'ext4'>('exFAT');
const vmCpuCores = ref<number>(2);
const vmMemoryMB = ref<number>(2048);
const vmDisplayAccel = ref<boolean>(true);
const diskList = ref<DiskInfo[]>([]);
const CUSTOM_ICONS_KEY = 'unigo_custom_icons_v1';

function loadCustomIcons(): Record<string, DiskIconType> {
  try {
    const raw = localStorage.getItem(CUSTOM_ICONS_KEY);
    if (raw) return JSON.parse(raw);
  } catch (e) {
    console.error('Failed to load custom icons from localStorage:', e);
  }
  return {};
}

function saveCustomIcons(icons: Record<string, DiskIconType>) {
  try {
    localStorage.setItem(CUSTOM_ICONS_KEY, JSON.stringify(icons));
  } catch (e) {
    console.error('Failed to save custom icons to localStorage:', e);
  }
}

function getDiskFingerprint(disk: DiskInfo): string {
  if (disk.serialNumber && disk.serialNumber.trim() !== '') {
    return `sn:${disk.serialNumber.trim()}`;
  }
  return `dev:${disk.name}_${disk.size}`;
}

const selectedDisk = ref<DiskInfo | null>(null);
const selectedDevices = ref<Set<string>>(new Set());
const customIcons = ref<Record<string, DiskIconType>>(loadCustomIcons());
const isPickerOpen = ref(false);
const targetPickerDisk = ref<DiskInfo | null>(null);
const isInspectorOpen = ref(false);
const targetInspectorDisk = ref<DiskInfo | null>(null);
const isDeployConfirmOpen = ref(false);
const isSettingsOpen = ref(false);
const isAboutOpen = ref(false);
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
  logUserAction('INFO', 'User cleared embedded log viewer');
  runtimeLogs.value = [];
  // Clear the backend in-memory log buffer so cleared logs do not
  // reappear after the app restarts or the UI is refreshed.
  ClearLogs().catch((err: any) => {
    console.error('Failed to clear backend log buffer:', err);
  });
}

function dismissToast() {
  toastMessage.value = '';
  logUserAction('DEBUG', 'User dismissed toast notification');
}

function dismissDeploySuccessBanner() {
  deploySuccessBanner.value.dismissed = true;
  deploySuccessBanner.value.visible = false;
  logUserAction('DEBUG', 'User dismissed deployment success banner');
}

function logUserAction(level: string, message: string, details: string = '') {
  if (window.go && window.go.main && window.go.main.App && (window.go.main.App as any).LogAction) {
    (window.go.main.App as any).LogAction(level, message, details);
  }
}

watch(currentEmbeddedLogFilter, (val) => {
  logUserAction('DEBUG', 'User switched log filter tab in embedded log viewer', val);
});

watch(isLogCardVisible, (val) => {
  logUserAction('INFO', 'User toggled log card visibility', val ? 'expanded' : 'collapsed');
});

watch(isAboutOpen, (val) => {
  if (val) {
    logUserAction('INFO', 'User opened About modal');
  }
});

watch(activeMode, (newMode) => {
  logUserAction('INFO', 'User switched deployment mode', newMode);
});

watch(selectedFsType, (newFs) => {
  logUserAction('INFO', 'User selected target file system', newFs);
});

watch(selectionMode, (newMode) => {
  logUserAction('INFO', 'User switched disk selection mode', newMode);
});

watch(selectedDisk, (disk, oldDisk) => {
  if (disk) {
    logUserAction('INFO', 'User selected target disk drive', `${disk.name || disk.device} (${disk.formatted})`);
  } else if (oldDisk) {
    logUserAction('INFO', 'User deselected target disk drive', `${oldDisk.name || oldDisk.device}`);
  }
});
const settingsInitialTab = ref<'general' | 'network' | 'uniboot' | 'ventoy'>('general');

const isVentoyAlertOpen = ref(false);
const ventoyAlertTitle = ref('');
const ventoyAlertMessage = ref('');
const ventoyAlertAction = ref<'open_settings' | 'switch_b'>('open_settings');

function openVentoyAlert(title: string, message: string, action: 'open_settings' | 'switch_b') {
  ventoyAlertTitle.value = title;
  ventoyAlertMessage.value = message;
  ventoyAlertAction.value = action;
  isVentoyAlertOpen.value = true;
}

function handleVentoyAlertAction() {
  isVentoyAlertOpen.value = false;
  openSettings('ventoy');
}

function handleVentoyAlertSwitchB() {
  isVentoyAlertOpen.value = false;
  activeMode.value = 'cloud';
  showToast(t('deploy.toast_switched_b'), 'success');
}

const isDiagnosticsOpen = ref(false);
const currentDiagnostics = ref<InstallDiagnosticsData | null>(null);
const currentDiagErrorMsg = ref('');

function openDiagnosticsModal(diag: InstallDiagnosticsData | null, msg: string) {
  currentDiagnostics.value = diag;
  currentDiagErrorMsg.value = msg || t('deploy.alert_fail', { msg: '' });
  isDiagnosticsOpen.value = true;
}

function handleCopyReport() {
  logUserAction('INFO', 'User copied diagnostics report to clipboard');
  showToast(t('diag.toast_copied'), 'info');
}

function handleRetryDeploy() {
  logUserAction('INFO', 'User clicked retry deployment from diagnostics modal');
  isDiagnosticsOpen.value = false;
  openDeployConfirm();
}

function openSettings(tab: 'general' | 'network' | 'uniboot' | 'ventoy' = 'general') {
  settingsInitialTab.value = tab;
  isSettingsOpen.value = true;
  logUserAction('INFO', 'User opened settings modal', tab);
}

const currentGithubProxy = ref('');
const pendingTargets = ref<string[]>([]);
const pendingTargetSnapshots = ref<DiskInfo[]>([]);
const isDeploying = ref(false);
const deployProgress = ref(0);
const deploySpeedMBps = ref(0);
const deployElapsedSec = ref(0);
const deployEtaSec = ref(0);
const autoEjectAfterDeploy = ref(false);
interface DeployBannerState {
  visible: boolean;
  msg: string;
  targets: string[];
  autoEjected?: boolean;
  mode?: 'cloud' | 'hybrid';
  dismissed?: boolean;
}

const deploySuccessBanner = ref<DeployBannerState>({
  visible: false,
  msg: '',
  targets: [],
  autoEjected: false,
  dismissed: false,
});

const showDeploySuccessBanner = computed(() => {
  const b = deploySuccessBanner.value;
  return Boolean(
    b.visible &&
    !b.dismissed &&
    b.mode === activeMode.value
  );
});
interface VMStatus {
  type: string;
  name: string;
  installed: boolean;
  path: string;
  version: string;
  priority: number;
  canBootRaw: boolean;
}

const hypervisorList = ref<VMStatus[]>([]);
const selectedVMType = ref<string>('qemu');
const selectedBootMode = ref<string>('auto');
const qemuStatus = ref({ installed: false, path: '', version: '' });
const isLaunchingQemu = ref(false);
const ventoyStatus = ref({ valid: true, version: '', message: '', executablePath: '' });

async function checkVentoyStatus() {
  if (window.go && window.go.main && window.go.main.App && window.go.main.App.ValidateVentoyCli) {
    try {
      const res = await window.go.main.App.ValidateVentoyCli('');
      if (res) {
        ventoyStatus.value = res;
      }
    } catch (e) {
      console.error('Failed to validate Ventoy status:', e);
    }
  }
}

async function selectMode(mode: 'cloud' | 'hybrid') {
  activeMode.value = mode;
  if (mode === 'hybrid') {
    await checkVentoyStatus();
    if (!isNonDestructive.value && !ventoyStatus.value.valid) {
      if (isMacOs.value) {
        showToast(t('deploy.tip_macos_unsupported'), 'warning');
      } else {
        showToast(t('deploy.tip_need_ventoy'), 'warning');
      }
    }
  }
}

interface IsoFileItem {
  name: string;
  path: string;
}

const selectedIsoFiles = ref<IsoFileItem[]>([]);
const isoCopyStatus = ref<string>('');

async function handleSelectIsoFiles() {
  if (typeof SelectIsoFiles === 'function') {
    try {
      const paths: string[] = await SelectIsoFiles(
        t('dialog.selectIsoTitle'),
        t('dialog.ventoyFilter'),
        t('dialog.allFilesFilter')
      );
      if (paths && paths.length > 0) {
        let added = 0;
        for (const p of paths) {
          if (!selectedIsoFiles.value.some(f => f.path === p)) {
            const name = p.split(/[/\\]/).pop() || p;
            selectedIsoFiles.value.push({ name, path: p });
            added++;
          }
        }
        if (added > 0) {
          showToast(t('deploy.toast_added_iso', { count: added }), 'success');
        }
      }
    } catch (err: any) {
      console.error('SelectIsoFiles error:', err);
    }
  } else {
    // Mock for browser demo
    const mockFiles = [
      { name: 'ubuntu-24.04-desktop-amd64.iso', path: '/Users/demo/Downloads/ubuntu-24.04-desktop-amd64.iso' },
      { name: 'Windows11_23H2_Chinese_Simplified_x64.iso', path: '/Users/demo/Downloads/Windows11_23H2_Chinese_Simplified_x64.iso' }
    ];
    for (const m of mockFiles) {
      if (!selectedIsoFiles.value.some(f => f.path === m.path)) {
        selectedIsoFiles.value.push(m);
      }
    }
    showToast(t('deploy.toast_added_demo_iso'), 'info');
  }
}

function removeIsoFile(index: number) {
  const item = selectedIsoFiles.value[index];
  const name = item ? (typeof item === 'string' ? item : item.name || item.path) : '';
  selectedIsoFiles.value.splice(index, 1);
  if (name) {
    logUserAction('INFO', 'User removed ISO source file from selection list', name);
  }
}

function clearIsoFiles() {
  selectedIsoFiles.value = [];
  logUserAction('INFO', 'User cleared all ISO source files from selection list');
}



const toastMessage = ref('');
const toastType = ref<'info' | 'warning' | 'error' | 'success'>('info');
let toastTimer: number | undefined;

function showToast(msg: string, type: 'info' | 'warning' | 'error' | 'success' = 'info') {
  toastMessage.value = msg;
  toastType.value = type;
  if (toastTimer) clearTimeout(toastTimer);
  toastTimer = window.setTimeout(() => {
    toastMessage.value = '';
  }, 4000);
}

function applyTheme(themeName?: string) {
  const theme = themeName === 'light' ? 'light' : 'dark';
  document.documentElement.setAttribute('data-theme', theme);
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

async function handleDeployBtnClick() {
  if (isDeploying.value) return;

  if (selectionMode.value === 'single' && !selectedDisk.value) {
    showToast(t('deploy.toast_select_target'), 'warning');
    return;
  }
  if (selectionMode.value === 'batch' && selectedDevices.value.size === 0) {
    showToast(t('deploy.toast_select_batch'), 'warning');
    return;
  }

  if (activeMode.value === 'hybrid' && !isNonDestructive.value) {
    if (!ventoyStatus.value.valid) {
      if (isMacOs.value) {
        openVentoyAlert(
          t('deploy.macos_alert_title'),
          t('deploy.macos_alert_desc'),
          'switch_b'
        );
      } else {
        openVentoyAlert(
          t('deploy.no_ventoy_title'),
          ventoyStatus.value.message || t('deploy.no_ventoy_desc'),
          'open_settings'
        );
      }
      return;
    }
    checkVentoyStatus().catch(() => {});
  }

  openDeployConfirm();
}

function openDeployConfirm() {
  let targets: string[] = [];
  if (selectionMode.value === 'single') {
    if (!selectedDisk.value) return;
    targets = [selectedDisk.value.device];
  } else {
    targets = Array.from(selectedDevices.value);
    if (targets.length === 0) return;
  }

  const refreshedSnapshots = targets.map((device) => diskList.value.find((disk) => disk.device === device));
  if (refreshedSnapshots.some((disk) => !disk)) {
    showToast(t('deploy.toast_target_changed'), 'error');
    return;
  }

  pendingTargets.value = targets;
  pendingTargetSnapshots.value = refreshedSnapshots as DiskInfo[];
  isDeployConfirmOpen.value = true;
}



function openInspector(disk: DiskInfo) {
  targetInspectorDisk.value = disk;
  isInspectorOpen.value = true;
  logUserAction('INFO', 'User opened USB hardware inspector modal', `${disk.name || disk.device} (${disk.formatted})`);
}

async function handleEjectDisk(disk: DiskInfo) {
  try {
    if (window.go && window.go.main && window.go.main.App && window.go.main.App.EjectDisk) {
      await window.go.main.App.EjectDisk(disk.device);
    }
    if (deploySuccessBanner.value.targets.includes(disk.device)) {
      deploySuccessBanner.value.dismissed = true;
      deploySuccessBanner.value.visible = false;
    }
    showToast(t('disk.toast_ejected_success', { device: disk.device, name: disk.name || disk.device }), 'success');
    await refreshDisks();
  } catch (err: any) {
    showToast(t('disk.toast_ejected_failed', { device: disk.device, error: err?.toString() || 'Unknown error' }), 'error');
  }
}

async function handleSafelyEjectAfterDeploy() {
  const targets = deploySuccessBanner.value.targets;
  deploySuccessBanner.value.dismissed = true;
  deploySuccessBanner.value.visible = false;
  let ejectedCount = 0;
  for (const dev of targets) {
    try {
      if (window.go && window.go.main && window.go.main.App && window.go.main.App.EjectDisk) {
        await window.go.main.App.EjectDisk(dev);
        ejectedCount++;
      }
    } catch (err) {
      console.warn(`Eject failed for ${dev}:`, err);
    }
  }
  await refreshDisks();
  if (ejectedCount > 0) {
    showToast(t('deploy.toast_auto_ejected', { count: ejectedCount }), 'success');
  }
}

async function handleBatchEjectDisks() {
  const targets = Array.from(selectedDevices.value);
  if (targets.length === 0) return;

  let successCount = 0;
  let failCount = 0;

  for (const device of targets) {
    try {
      if (window.go && window.go.main && window.go.main.App && window.go.main.App.EjectDisk) {
        await window.go.main.App.EjectDisk(device);
      }
      successCount++;
      selectedDevices.value.delete(device);
    } catch (err) {
      failCount++;
    }
  }

  if (failCount === 0) {
    showToast(t('disk.toast_batch_eject_success', { count: successCount }), 'success');
  } else {
    showToast(t('disk.toast_batch_eject_partial', { successCount, failCount }), 'warning');
  }

  await refreshDisks();
}

function openIconPicker(disk: DiskInfo) {
  targetPickerDisk.value = disk;
  isPickerOpen.value = true;
  logUserAction('INFO', 'User opened custom icon picker modal', disk.name || disk.device);
}

function onIconSelected(type: DiskIconType) {
  if (targetPickerDisk.value) {
    // Use the fingerprint (serial-number-based) as the single source of truth
    // so the icon survives hot-plug events where the device path may change.
    const fp = getDiskFingerprint(targetPickerDisk.value);
    customIcons.value[fp] = type;
    saveCustomIcons(customIcons.value);
    logUserAction('INFO', 'User updated custom disk icon', `${targetPickerDisk.value.name || targetPickerDisk.value.device} -> ${type}`);
  }
}

function onIconReset() {
  if (targetPickerDisk.value) {
    const fp = getDiskFingerprint(targetPickerDisk.value);
    delete customIcons.value[fp];
    saveCustomIcons(customIcons.value);
    logUserAction('INFO', 'User reset custom disk icon to default', targetPickerDisk.value.name || targetPickerDisk.value.device);
  }
}

const isMacOs = computed(() => navigator.userAgent.includes('Mac') || navigator.platform.includes('Mac'));

function checkIsExistingBootDisk(d: DiskInfo): boolean {
  if (!d) return false;
  const nameUpper = (d.name || '').toUpperCase();
  const isVentoyName = nameUpper.includes('VENTOY') || nameUpper.includes('UNIBOOT');
  return Boolean(d.isRealVentoy || d.isCloudMode || d.isVentoy || isVentoyName);
}

const isSelectedVentoyDisk = computed(() => {
  if (selectionMode.value === 'single' && selectedDisk.value) {
    return checkIsExistingBootDisk(selectedDisk.value);
  }
  return false;
});

const isNonDestructive = computed(() => {
  if (selectionMode.value === 'single') {
    return isSelectedVentoyDisk.value;
  }
  if (selectedDevices.value.size === 0) return false;
  return Array.from(selectedDevices.value).every((dev: string) => {
    const d = diskList.value.find((disk: DiskInfo) => disk.device === dev);
    return d ? checkIsExistingBootDisk(d) : false;
  });
});

const ventoyCountInBatch = computed(() => {
  if (selectionMode.value !== 'batch' || selectedDevices.value.size === 0) return 0;
  let count = 0;
  selectedDevices.value.forEach((dev: string) => {
    const d = diskList.value.find((disk: DiskInfo) => disk.device === dev);
    if (d && checkIsExistingBootDisk(d)) {
      count++;
    }
  });
  return count;
});

const deployBtnText = computed(() => {
  if (isDeploying.value) return t('deploy.writing');

  if (selectionMode.value === 'single') {
    if (isNonDestructive.value) {
      return t('deploy.start_update');
    }
    return activeMode.value === 'cloud' ? t('deploy.start_cloud_create') : t('deploy.start_create');
  }

  const total = selectedDevices.value.size;
  const bootCount = ventoyCountInBatch.value;
  const blankCount = total - bootCount;

  if (bootCount === total && total > 0) {
    return t('deploy.batch_update', { count: total });
  } else if (blankCount === total && total > 0) {
    return t('deploy.batch_create', { count: total });
  } else {
    return t('deploy.batch_mixed', { count: total });
  }
});


const deployDisabledReason = computed(() => {
  if (isDeploying.value) return t('deploy.tip_writing');
  if (selectionMode.value === 'single' && !selectedDisk.value) return t('deploy.tip_select_single');
  if (selectionMode.value === 'batch' && selectedDevices.value.size === 0) return t('deploy.tip_select_batch');
  if (activeMode.value === 'hybrid' && !isNonDestructive.value && !ventoyStatus.value.valid) {
    if (isMacOs.value) {
      return t('deploy.tip_macos_unsupported');
    }
    return ventoyStatus.value.message || t('deploy.tip_need_ventoy');
  }
  if (selectionMode.value === 'batch') {
    const total = selectedDevices.value.size;
    const bootCount = ventoyCountInBatch.value;
    const blankCount = total - bootCount;
    if (bootCount > 0 && blankCount > 0) {
      return t('deploy.tip_batch_mixed', { bootCount, blankCount });
    }
    if (bootCount === total && total > 0) {
      return t('deploy.tip_batch_update_all', { count: total });
    }
  }
  return '';
});

const isVmDisabled = computed(() => {
  return (
    isLaunchingQemu.value ||
    isDeploying.value ||
    hypervisorList.value.length === 0 ||
    !activeVmTargetDevice.value
  );
});

const vmDisabledReason = computed(() => {
  if (isLaunchingQemu.value) {
    return t('vm.tip_launching');
  }
  if (isDeploying.value) {
    return t('vm.tip_deploying');
  }
  if (hypervisorList.value.length === 0) {
    return t('vm.tip_not_installed');
  }
  if (!activeVmTargetDevice.value) {
    return t('vm.tip_select_target');
  }
  return t('vm.tip_ready');
});

const activeVmTargetDevice = computed(() => {
  if (selectionMode.value === 'single') {
    return selectedDisk.value?.device || '';
  }
  if (selectedDevices.value.size > 0) {
    return Array.from(selectedDevices.value)[0];
  }
  return '';
});

const activeVmTargetName = computed(() => {
  if (selectionMode.value === 'single') {
    return selectedDisk.value?.name || selectedDisk.value?.device || '';
  }
  if (selectedDevices.value.size > 0) {
    const firstDev = Array.from(selectedDevices.value)[0];
    const found = diskList.value.find(d => d.device === firstDev);
    return found?.name || firstDev;
  }
  return '';
});

function setSelectionMode(mode: 'single' | 'batch') {
  selectionMode.value = mode;
}

function selectAllDisks() {
  selectedDevices.value = new Set(diskList.value.map(d => d.device));
}

function deselectAllDisks() {
  selectedDevices.value.clear();
}

function onDiskSelect(disk: DiskInfo) {
  if (selectionMode.value === 'single') {
    if (selectedDisk.value && selectedDisk.value.device === disk.device) {
      selectedDisk.value = null;
    } else {
      selectedDisk.value = disk;
    }
  } else {
    onDiskToggle(disk);
  }
}

function onDiskToggle(disk: DiskInfo) {
  const newSet = new Set(selectedDevices.value);
  const wasSelected = newSet.has(disk.device);
  if (wasSelected) {
    newSet.delete(disk.device);
    logUserAction('INFO', 'User unchecked disk drive in batch selection mode', `${disk.name || disk.device}`);
  } else {
    newSet.add(disk.device);
    logUserAction('INFO', 'User checked disk drive in batch selection mode', `${disk.name || disk.device} (${disk.formatted})`);
  }
  selectedDevices.value = newSet;
}

const isScanningDisks = ref(false);

// Wails JS binding fallbacks / mock data for standalone preview
async function refreshDisks() {
  if (isScanningDisks.value) return;
  isScanningDisks.value = true;

  const previousSelectedDevice = selectedDisk.value?.device;
  // 1. Immediately clear the disk list and selection for instant UI feedback
  diskList.value = [];
  selectedDisk.value = null;
  selectedDevices.value.clear();

  try {
    if (window.go && window.go.main && window.go.main.App) {
      try {
        const fetched = (await window.go.main.App.GetDiskList()) || [];
        diskList.value = fetched;
        if (previousSelectedDevice) {
          const stillExists = diskList.value.find(d => d.device === previousSelectedDevice);
          selectedDisk.value = stillExists || (diskList.value.length > 0 ? diskList.value[0] : null);
        } else if (diskList.value.length > 0) {
          selectedDisk.value = diskList.value[0];
        }
      } catch (e) {
        console.error(e);
        diskList.value = [];
        selectedDisk.value = null;
      }
    } else {
      // Fallback mock for browser preview demonstrating genuine vs fake USB 3.0
      await new Promise(resolve => setTimeout(resolve, 450));
      diskList.value = [
        {
          device: '/dev/disk2',
          name: 'SanDisk Ultra USB 3.0 Flash Drive',
          size: 32000000000,
          formatted: '32 GB',
          isRemovable: true,
          isSystem: false,
          usbVersion: 'USB 2.0',
          usbSpeed: '480 Mb/s',
          vendor: 'SanDisk (Suspected Fake)',
          isFakeUsb3: true,
          protocolCode: 'usb2'
        },
        {
          device: '/dev/disk3',
          name: 'Kingston DataTraveler 3.0',
          size: 64000000000,
          formatted: '64 GB',
          isRemovable: true,
          isSystem: false,
          usbVersion: 'USB 3.0',
          usbSpeed: '5 Gb/s',
          vendor: 'Kingston Technology',
          isFakeUsb3: false,
          protocolCode: 'usb3_0'
        },
        {
          device: '/dev/disk4',
          name: 'Samsung Type-C Duo 3.1',
          size: 128000000000,
          formatted: '128 GB',
          isRemovable: true,
          isSystem: false,
          usbVersion: 'USB 3.1 Gen 2',
          usbSpeed: '10 Gb/s',
          vendor: 'Samsung Electronics',
          isFakeUsb3: false,
          protocolCode: 'usb3_1'
        }
      ];
      if (previousSelectedDevice) {
        const stillExists = diskList.value.find(d => d.device === previousSelectedDevice);
        selectedDisk.value = stillExists || (diskList.value.length > 0 ? diskList.value[0] : null);
      } else if (diskList.value.length > 0) {
        selectedDisk.value = diskList.value[0];
      }
    }
  } finally {
    isScanningDisks.value = false;
  }
}

async function checkQemu() {
  if (window.go && window.go.main && window.go.main.App) {
    if (typeof window.go.main.App.DetectHypervisors === 'function') {
      try {
        const list = await window.go.main.App.DetectHypervisors();
        if (Array.isArray(list)) {
          const installed = list.filter((h: any) => h.installed).sort((a: any, b: any) => a.priority - b.priority);
          hypervisorList.value = installed;
          if (installed.length > 0) {
            selectedVMType.value = installed[0].type;
            qemuStatus.value = { installed: true, path: installed[0].path, version: installed[0].version };
          } else {
            qemuStatus.value = { installed: false, path: '', version: '' };
          }
          return;
        }
      } catch (e) {
        console.error('Failed to detect hypervisors:', e);
      }
    }
    const fallback = await window.go.main.App.CheckQEMU();
    if (fallback && fallback.installed) {
      hypervisorList.value = [{ type: 'qemu', name: 'QEMU', installed: true, path: fallback.path, version: fallback.version, priority: 1, canBootRaw: true }];
      selectedVMType.value = 'qemu';
      qemuStatus.value = fallback;
    } else {
      hypervisorList.value = [];
      qemuStatus.value = { installed: false, path: '', version: '' };
    }
  } else {
    // Browser demo mode: Mock installed hypervisors (QEMU, UTM, VirtualBox)
    hypervisorList.value = [
      { type: 'qemu', name: 'QEMU', installed: true, path: '/usr/local/bin/qemu-system-x86_64', version: 'QEMU 8.2', priority: 1, canBootRaw: true },
      { type: 'utm', name: 'UTM', installed: true, path: '/Applications/UTM.app', version: 'UTM 4.4', priority: 2, canBootRaw: true },
      { type: 'virtualbox', name: 'Oracle VM VirtualBox', installed: true, path: '/usr/local/bin/VBoxManage', version: 'VirtualBox 7.0', priority: 7, canBootRaw: true }
    ];
    selectedVMType.value = 'qemu';
    qemuStatus.value = { installed: true, path: '/usr/local/bin/qemu-system-x86_64', version: 'QEMU 8.2' };
  }
}

async function startDeployment() {
  let targets: string[] = [];
  if (selectionMode.value === 'single') {
    if (!selectedDisk.value) return;
    targets = [selectedDisk.value.device];
  } else {
    targets = Array.from(selectedDevices.value);
    if (targets.length === 0) return;
  }

  isDeploying.value = true;
  deployProgress.value = 5;

  // For Cloud Mode, listen to the real stage-progress events emitted by the backend.
  // For Hybrid Mode, iso-copy-progress events are already handled globally in onMounted.
  let unsubCloudProgress: (() => void) | null = null;
  if (activeMode.value === 'cloud' && window.runtime && window.runtime.EventsOn) {
    unsubCloudProgress = window.runtime.EventsOn('cloud-deploy-progress', (progress: number) => {
      deployProgress.value = progress;
    });
  }

  let success = true;
  let resultMsg = '';
  let latestDiagnostics: any = null;

  try {
    if (window.go && window.go.main && window.go.main.App) {
      const isoPaths = selectedIsoFiles.value.map(f => f.path);
      if (targets.length === 1) {
        const expected = pendingTargetSnapshots.value[0];
        if (!expected) throw new Error(t('deploy.toast_target_changed'));
        let res: any;
        if (activeMode.value === 'cloud') {
          res = await window.go.main.App.DeployCloudMode(targets[0], selectedFsType.value, expected);
        } else {
          res = await window.go.main.App.DeployHybridMode(targets[0], selectedFsType.value, isoPaths, expected);
        }
        if (res) {
          success = res.success;
          resultMsg = res.message || '';
          if (res.diagnostics) {
            latestDiagnostics = res.diagnostics;
          }
        }
      } else {
        if (pendingTargetSnapshots.value.length !== targets.length) {
          throw new Error(t('deploy.toast_target_changed'));
        }
        let resList: any[];
        if (activeMode.value === 'cloud') {
          resList = await window.go.main.App.DeployCloudModeBatch(targets, selectedFsType.value, pendingTargetSnapshots.value);
        } else {
          resList = await window.go.main.App.DeployHybridModeBatch(targets, selectedFsType.value, isoPaths, pendingTargetSnapshots.value);
        }
        if (resList && resList.length > 0) {
          const failed = resList.filter(r => !r.success);
          if (failed.length > 0) {
            success = false;
            resultMsg = failed.map(f => `${f.target}: ${f.message}`).join('\n');
            if (failed[0].diagnostics) {
              latestDiagnostics = failed[0].diagnostics;
            }
          } else {
            resultMsg = t('deploy.result_batch_success', { count: resList.length });
          }
        }
      }
    } else {
      // Mock execution for browser demo
      await new Promise(r => setTimeout(r, 800));
      const modeLabel = activeMode.value === 'cloud' ? 'B' : 'A';
      resultMsg = t('deploy.result_success', { mode: modeLabel, targets: targets.join(', ') });
    }
  } catch (e: any) {
    console.error(e);
    success = false;
    resultMsg = e?.message || String(e);
  } finally {
    if (unsubCloudProgress) {
      unsubCloudProgress();
      unsubCloudProgress = null;
    }
  }


  if (success) {
    deployProgress.value = 100;
    setTimeout(async () => {
      isDeploying.value = false;
      deployProgress.value = 0;

      // Auto safely eject all created USB drives — strictly only if 100% of single/batch deployment tasks succeeded AND autoEjectAfterDeploy setting is enabled
      let autoEjectedCount = 0;
      if (autoEjectAfterDeploy.value) {
        for (const dev of targets) {
          try {
            if (window.go && window.go.main && window.go.main.App && window.go.main.App.EjectDisk) {
              await window.go.main.App.EjectDisk(dev);
              autoEjectedCount++;
            }
          } catch (ejectErr) {
            console.warn(`Auto eject failed for ${dev}:`, ejectErr);
          }
        }
      }

      await refreshDisks();

      deploySuccessBanner.value = {
        visible: true,
        msg: resultMsg,
        targets: [...targets],
        autoEjected: autoEjectedCount > 0,
        mode: activeMode.value,
        dismissed: false
      };

      if (autoEjectedCount > 0) {
        showToast(t('deploy.toast_auto_ejected', { count: autoEjectedCount }), 'success');
      }
    }, 200);
  } else {
    isDeploying.value = false;
    deployProgress.value = 0;
    deploySpeedMBps.value = 0;
    deployElapsedSec.value = 0;
    deployEtaSec.value = 0;
    openDiagnosticsModal(latestDiagnostics, resultMsg);
  }
}

async function handleCancelDeploy() {
  logUserAction('WARN', 'User clicked cancel deployment button');
  const app = (window as any)?.go?.main?.App;
  if (app && typeof app.CancelDeployment === 'function') {
    try {
      const cancelled = await app.CancelDeployment();
      if (cancelled) {
        showToast(t('deploy.toast_cancelled'), 'info');
        logUserAction('INFO', 'Deployment task cancelled successfully');
      }
    } catch (e: any) {
      console.error('Failed to cancel deployment:', e);
    }
  }
}


async function launchVM() {
  console.log('[UniBoot] launchVM clicked, target:', activeVmTargetDevice.value, 'vmType:', selectedVMType.value);

  if (!diskList.value || diskList.value.length === 0) {
    showToast(t('deploy.toast_no_disks'), 'warning');
    return;
  }

  const targetDevice = activeVmTargetDevice.value;

  if (!targetDevice) {
    showToast(t('vm.toast_select_first'), 'warning');
    return;
  }

  if (hypervisorList.value.length === 0) {
    showToast(t('vm.toast_not_installed'), 'error');
    return;
  }

  isLaunchingQemu.value = true;
  const currentVM = hypervisorList.value.find(h => h.type === selectedVMType.value) || hypervisorList.value[0];
  const vmName = currentVM ? currentVM.name : 'QEMU';

  try {
    const vmConfig = {
      cpuCores: vmCpuCores.value,
      memoryMB: vmMemoryMB.value,
      bootMode: selectedBootMode.value,
      displayAccel: vmDisplayAccel.value,
      secureBoot: false,
    };

    if (window.go && window.go.main && window.go.main.App) {
      const app = window.go.main.App as any;
      if (typeof app.LaunchVMWithConfig === 'function') {
        await app.LaunchVMWithConfig(targetDevice, selectedVMType.value, vmConfig);
        logUserAction('INFO', 'User launched hypervisor simulation test with VMConfig', `${vmName} (${selectedVMType.value}, ${selectedBootMode.value}, ${vmCpuCores.value} cores, ${vmMemoryMB.value}MB) on ${targetDevice}`);
        showToast(t('vm.startSuccess_vm', { name: vmName }), 'success');
      } else if (typeof app.LaunchVM === 'function') {
        await app.LaunchVM(targetDevice, selectedVMType.value, selectedBootMode.value);
        logUserAction('INFO', 'User launched hypervisor simulation test', `${vmName} (${selectedVMType.value}, ${selectedBootMode.value}) on ${targetDevice}`);
        showToast(t('vm.startSuccess_vm', { name: vmName }), 'success');
      } else if (typeof app.LaunchQEMU === 'function') {
        await app.LaunchQEMU(targetDevice);
        logUserAction('INFO', 'User launched hypervisor simulation test', `QEMU on ${targetDevice}`);
        showToast(t('vm.startSuccess', { name: 'QEMU' }), 'success');
      } else {
        showToast(t('vm.backendNotReady'), 'warning');
      }
    } else {
      await new Promise(r => setTimeout(r, 600));
      logUserAction('INFO', 'User launched hypervisor simulation test (demo mode)', `${vmName} on ${targetDevice}`);
      showToast(t('vm.demoModeStart', { name: vmName }), 'info');
    }
  } catch (e: any) {
    console.error('[UniBoot] LaunchVM error:', e);
    showToast(t('vm.startFailed', { error: e?.message || String(e) }), 'error');
  } finally {
    isLaunchingQemu.value = false;
  }
}

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

.app-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;
}

.brand {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.logo {
  font-size: 2.5rem;
}

h1 {
  font-size: 1.6rem;
  font-weight: 800;
  background: linear-gradient(90deg, #00e5ff, #9d4edd);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}

.sub-brand {
  font-size: 0.825rem;
  color: var(--text-muted);
}

.mode-tabs {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  background: rgba(255, 255, 255, 0.05);
  padding: 0.3rem;
  border-radius: 12px;
  border: 1px solid var(--card-border);
}

.tab-btn {
  background: transparent;
  color: var(--text-muted);
  border: none;
  padding: 0.6rem 1.2rem;
  border-radius: 8px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
}

.tab-btn.active {
  background: var(--accent-cyan);
  color: #070a12;
}

/* Header Quick Language Dropdown */
.lang-selector-header {
  position: relative;
  display: inline-block;
}

.lang-pill-btn {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  background: rgba(255, 255, 255, 0.08);
  color: var(--text-main, #e0e6ed);
  border: 1px solid var(--card-border, rgba(255, 255, 255, 0.12));
  padding: 0.45rem 0.8rem;
  border-radius: 8px;
  font-size: 0.825rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
}

.lang-pill-btn:hover {
  background: rgba(0, 229, 255, 0.15);
  color: var(--accent-cyan, #00e5ff);
  border-color: rgba(0, 229, 255, 0.4);
}

.lang-icon {
  font-size: 1rem;
}

.lang-label {
  font-size: 0.825rem;
  white-space: nowrap;
}

.dropdown-caret {
  font-size: 0.7rem;
  opacity: 0.75;
}

.lang-dropdown-menu {
  position: absolute;
  top: calc(100% + 8px);
  right: 0;
  min-width: 220px;
  max-height: 340px;
  overflow-y: auto;
  background: var(--card-bg, rgba(20, 24, 38, 0.96));
  backdrop-filter: blur(18px);
  -webkit-backdrop-filter: blur(18px);
  border: 1px solid var(--card-border, rgba(255, 255, 255, 0.15));
  border-radius: 12px;
  padding: 0.45rem;
  box-shadow: 0 12px 36px rgba(0, 0, 0, 0.5);
  z-index: 1000;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.lang-option {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  width: 100%;
  padding: 0.6rem 0.8rem;
  background: transparent;
  border: none;
  border-radius: 8px;
  color: var(--text-main, #e0e6ed);
  font-size: 0.85rem;
  font-weight: 600;
  text-align: left;
  cursor: pointer;
  transition: all 0.15s ease;
}

.lang-option:hover {
  background: rgba(0, 229, 255, 0.12);
  color: var(--accent-cyan, #00e5ff);
}

.lang-option.active {
  background: rgba(0, 229, 255, 0.2);
  color: var(--accent-cyan, #00e5ff);
  font-weight: 700;
}

.opt-flag {
  font-size: 1.1rem;
}

.opt-text {
  flex: 1;
}

.opt-check {
  font-size: 0.85rem;
  color: var(--accent-cyan, #00e5ff);
}

.content-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1.75rem;
}

/* Embedded Log Center Card & Smooth Transition */
.card-fade-enter-active,
.card-fade-leave-active {
  transition: opacity 0.15s cubic-bezier(0.16, 1, 0.3, 1), transform 0.15s cubic-bezier(0.16, 1, 0.3, 1);
  will-change: opacity, transform;
}

.card-fade-enter-from,
.card-fade-leave-to {
  opacity: 0;
  transform: translateY(-6px);
}

.log-section-card {
  grid-column: 1 / -1;
  display: flex;
  flex-direction: column;
  padding: 1.25rem 1.5rem;
  background: var(--card-bg, rgba(16, 24, 40, 0.6));
  border: 1px solid var(--card-border, rgba(255, 255, 255, 0.1));
  border-radius: 16px;
  backdrop-filter: blur(16px);
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.25);
  min-height: 500px;
}

.log-section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.75rem;
  flex-wrap: wrap;
  gap: 0.75rem;
}

.log-sub-header {
  display: flex;
  align-items: center;
  margin-bottom: 0.75rem;
}

.log-title-group {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.log-title-group h2 {
  font-size: 1.15rem;
  font-weight: 700;
  margin: 0;
  color: var(--text-color, #f0f4f8);
}

.live-badge {
  background: rgba(16, 185, 129, 0.15);
  color: #10b981;
  border: 1px solid rgba(16, 185, 129, 0.3);
  font-size: 0.75rem;
  padding: 0.2rem 0.6rem;
  border-radius: 12px;
  font-weight: 600;
  letter-spacing: 0.5px;
}

.log-section-controls {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  flex-wrap: wrap;
}

.filter-tabs-sm {
  display: flex;
  gap: 0.3rem;
  background: rgba(0, 0, 0, 0.25);
  padding: 3px;
  border-radius: 8px;
  border: 1px solid rgba(255, 255, 255, 0.08);
}

.btn-tab-sm {
  background: transparent;
  border: none;
  color: var(--text-muted, #94a3b8);
  padding: 0.25rem 0.55rem;
  font-size: 0.75rem;
  font-weight: 600;
  border-radius: 5px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn-tab-sm:hover {
  color: #ffffff;
  background: rgba(255, 255, 255, 0.08);
}

.btn-tab-sm.active {
  background: var(--accent-cyan, #00e5ff);
  color: #0b1120;
}

.btn-text-sm {
  background: rgba(255, 255, 255, 0.08);
  border: 1px solid rgba(255, 255, 255, 0.16);
  color: var(--text-color, #f1f5f9);
  font-size: 0.78rem;
  font-weight: 600;
  padding: 0.35rem 0.7rem;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s ease;
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
}

.btn-text-sm:hover {
  background: rgba(255, 255, 255, 0.18);
  border-color: rgba(255, 255, 255, 0.3);
  color: #ffffff;
}

.btn-text-danger-sm {
  background: rgba(239, 68, 68, 0.15);
  border: 1px solid rgba(239, 68, 68, 0.35);
  color: #f87171;
  font-size: 0.78rem;
  font-weight: 600;
  padding: 0.35rem 0.7rem;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s ease;
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
}

.btn-text-danger-sm:hover {
  background: rgba(239, 68, 68, 0.25);
  border-color: rgba(239, 68, 68, 0.5);
  color: #ef4444;
}

.auto-scroll-label-sm {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  font-size: 0.78rem;
  font-weight: 600;
  color: #f1f5f9;
  cursor: pointer;
  user-select: none;
  padding: 0.35rem 0.65rem;
  border-radius: 6px;
  background: rgba(255, 255, 255, 0.08);
  border: 1px solid rgba(255, 255, 255, 0.16);
  transition: all 0.2s ease;
}

.auto-scroll-label-sm input[type="checkbox"] {
  accent-color: #00e5ff;
  width: 14px;
  height: 14px;
  cursor: pointer;
}

.auto-scroll-label-sm:hover {
  background: rgba(255, 255, 255, 0.18);
  border-color: rgba(0, 229, 255, 0.4);
  color: #ffffff;
}

.embedded-terminal-window {
  background: rgba(10, 15, 28, 0.85);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 10px;
  padding: 1rem 1.15rem;
  min-height: 440px;
  max-height: 560px;
  overflow-y: auto;
  overflow-x: hidden;
  font-family: 'JetBrains Mono', 'Fira Code', 'Cascadia Code', Consolas, monospace;
  font-size: 0.82rem;
  line-height: 1.6;
  flex: 1;
}

.embedded-terminal-window .empty-logs {
  color: var(--text-muted, #64748b);
  text-align: center;
  padding: 5rem 0;
  font-style: italic;
}

.embedded-terminal-window .log-row {
  display: flex;
  align-items: flex-start;
  gap: 1rem;
  padding: 0.35rem 0;
  border-bottom: 1px dashed rgba(255, 255, 255, 0.05);
  box-sizing: border-box;
  width: 100%;
}

.embedded-terminal-window .log-time {
  color: #64748b;
  font-size: 0.78rem;
  font-family: 'JetBrains Mono', monospace;
  white-space: nowrap;
  flex-shrink: 0;
  width: 100px;
  min-width: 100px;
  height: 22px;
  line-height: 22px;
  display: inline-flex;
  align-items: center;
}

.embedded-terminal-window .log-level-badge {
  font-size: 0.7rem;
  font-weight: 700;
  height: 20px;
  line-height: 18px;
  padding: 0 0.5rem;
  border-radius: 4px;
  white-space: nowrap;
  flex-shrink: 0;
  width: 64px;
  min-width: 64px;
  box-sizing: border-box;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  letter-spacing: 0.5px;
  margin-top: 1px;
}

.embedded-terminal-window .log-level-badge.info {
  background: rgba(16, 185, 129, 0.18);
  color: #34d399;
  border: 1px solid rgba(16, 185, 129, 0.3);
}

.embedded-terminal-window .log-level-badge.warn {
  background: rgba(245, 158, 11, 0.18);
  color: #fbbf24;
  border: 1px solid rgba(245, 158, 11, 0.3);
}

.embedded-terminal-window .log-level-badge.error {
  background: rgba(239, 68, 68, 0.2);
  color: #f87171;
  border: 1px solid rgba(239, 68, 68, 0.35);
}

.embedded-terminal-window .log-level-badge.debug {
  background: rgba(168, 85, 247, 0.18);
  color: #c084fc;
  border: 1px solid rgba(168, 85, 247, 0.3);
}

.embedded-terminal-window .log-content {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  gap: 0.5rem;
}

.embedded-terminal-window .log-msg {
  color: #e2e8f0;
  font-size: 0.82rem;
  line-height: 22px;
  white-space: pre-wrap;
  word-break: break-word;
  overflow-wrap: anywhere;
}

.embedded-terminal-window .log-row.info .log-msg {
  color: #f1f5f9;
}

.embedded-terminal-window .log-row.warn .log-msg {
  color: #fde047;
}

.embedded-terminal-window .log-row.error .log-msg {
  color: #fca5a5;
}

.embedded-terminal-window .log-row.debug .log-msg {
  color: #c084fc;
}

.embedded-terminal-window .log-details {
  color: #94a3b8;
  font-size: 0.78rem;
  line-height: 22px;
  white-space: pre-wrap;
  word-break: break-word;
  overflow-wrap: anywhere;
  direction: ltr !important;
  text-align: left !important;
  unicode-bidi: embed;
}

.embedded-terminal-window::-webkit-scrollbar {
  width: 6px;
}

.embedded-terminal-window::-webkit-scrollbar-track {
  background: rgba(0, 0, 0, 0.2);
  border-radius: 3px;
}

.embedded-terminal-window::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.15);
  border-radius: 3px;
}

.embedded-terminal-window::-webkit-scrollbar-thumb:hover {
  background: rgba(255, 255, 255, 0.3);
}

.section-card {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.section-card h2 {
  font-size: 1.25rem;
  margin-bottom: 0.4rem;
}

.section-desc {
  font-size: 0.85rem;
  color: var(--text-muted);
  margin-bottom: 1rem;
  line-height: 1.4;
}

.selection-controls {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  margin-bottom: 1.25rem;
  background: rgba(0, 0, 0, 0.15);
  padding: 0.75rem;
  border-radius: 10px;
  border: 1px solid var(--card-border);
}

.selection-mode-toggle {
  display: flex;
  gap: 0.5rem;
}

.sub-tab-btn {
  flex: 1;
  background: rgba(255, 255, 255, 0.04);
  color: var(--text-muted);
  border: 1px solid transparent;
  padding: 0.4rem 0.8rem;
  border-radius: 6px;
  font-size: 0.825rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
}

.sub-tab-btn.active {
  background: rgba(0, 229, 255, 0.12);
  color: var(--accent-cyan);
  border-color: rgba(0, 229, 255, 0.3);
}

.batch-actions {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  font-size: 0.8rem;
}

.btn-text {
  background: transparent;
  border: none;
  color: var(--accent-cyan);
  cursor: pointer;
  font-size: 0.8rem;
  padding: 0.1rem 0.4rem;
  border-radius: 4px;
  transition: all 0.2s ease;
}

.btn-eject {
  background: rgba(255, 255, 255, 0.06);
  color: var(--text-muted);
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: 6px;
  padding: 0.2rem 0.5rem;
  font-size: 0.75rem;
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn-eject:hover:not(:disabled) {
  background: rgba(239, 68, 68, 0.15);
  color: #ef4444;
  border-color: rgba(239, 68, 68, 0.3);
}

.btn-eject:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.btn-text:hover {
  background: rgba(0, 229, 255, 0.1);
}

.selection-count {
  margin-left: auto;
  color: var(--text-muted);
  font-size: 0.775rem;
  font-weight: 600;
}

.disk-list {
  flex: 1 1 0px;
  min-height: 180px;
  overflow-y: auto;
  padding-right: 0.4rem;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.disk-list::-webkit-scrollbar {
  width: 6px;
}

.disk-list::-webkit-scrollbar-track {
  background: rgba(0, 0, 0, 0.1);
  border-radius: 4px;
}

.disk-list::-webkit-scrollbar-thumb {
  background: rgba(0, 229, 255, 0.3);
  border-radius: 4px;
}

.disk-list::-webkit-scrollbar-thumb:hover {
  background: var(--accent-cyan);
}

.empty-state {
  padding: 2rem;
  text-align: center;
  color: var(--text-muted);
  font-size: 0.9rem;
}

.refresh-btn {
  flex: 0 0 auto;
  width: 100%;
  margin-top: auto;
  padding-top: 0.75rem;
}

.fs-selector {
  margin-bottom: 1.25rem;
}

.safe-mode-notice {
  display: flex;
  align-items: flex-start;
  gap: 0.8rem;
  background: var(--alert-info-bg);
  border: 1px solid var(--alert-info-border);
  border-radius: 12px;
  padding: 0.9rem 1.1rem;
  margin-bottom: 1.25rem;
}

.safe-notice-icon {
  font-size: 1.4rem;
  line-height: 1.2;
}

.safe-notice-content {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.safe-notice-title {
  color: var(--alert-info-title);
  font-weight: 700;
  font-size: 0.875rem;
}

.safe-notice-desc {
  color: var(--alert-info-text);
  font-size: 0.8rem;
  line-height: 1.45;
}

.safe-notice-desc b {
  color: var(--text-main);
  font-weight: 700;
}

.warn-cloudmode-notice {
  display: flex;
  align-items: flex-start;
  gap: 0.8rem;
  background: var(--alert-warning-bg);
  border: 1px solid var(--alert-warning-border);
  border-radius: 12px;
  padding: 0.9rem 1.1rem;
  margin-bottom: 1.25rem;
}

.warn-notice-icon {
  font-size: 1.4rem;
  line-height: 1.2;
}

.warn-notice-content {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.warn-notice-title {
  color: var(--alert-warning-title);
  font-weight: 700;
  font-size: 0.875rem;
}

.warn-notice-desc {
  color: var(--alert-warning-text);
  font-size: 0.8rem;
  line-height: 1.45;
}

.warn-notice-desc b {
  color: var(--text-main);
  font-weight: 700;
}

.deploy-box {
  background: rgba(0, 0, 0, 0.2);
  border-radius: 12px;
  padding: 1.25rem;
  margin-bottom: 1.5rem;
}

.selected-target {
  font-size: 0.9rem;
  margin-bottom: 1rem;
  display: flex;
  justify-content: space-between;
}

.deploy-btn {
  width: 100%;
  font-size: 1rem;
}

.deploy-btn.safe-btn {
  background: linear-gradient(135deg, #00e5ff 0%, #0284c7 100%);
  color: #070a12;
  font-weight: 700;
  box-shadow: 0 4px 14px rgba(0, 229, 255, 0.35);
}

.deploy-btn.safe-btn:hover {
  background: linear-gradient(135deg, #38bdf8 0%, #00e5ff 100%);
  box-shadow: 0 6px 20px rgba(0, 229, 255, 0.5);
}

.deploy-success-banner {
  display: flex;
  align-items: flex-start;
  gap: 0.85rem;
  padding: 1rem 1.15rem;
  margin-top: 0.85rem;
  background: linear-gradient(135deg, #047857 0%, #065f46 100%);
  border: 1.5px solid #10b981;
  border-radius: 12px;
  box-shadow: 0 8px 24px rgba(16, 185, 129, 0.3), inset 0 1px 0 rgba(255, 255, 255, 0.2);
  animation: bannerFadeIn 0.35s cubic-bezier(0.16, 1, 0.3, 1);
}

@keyframes bannerFadeIn {
  from { opacity: 0; transform: translateY(6px); }
  to   { opacity: 1; transform: translateY(0); }
}

.deploy-success-icon {
  font-size: 1.8rem;
  line-height: 1;
  flex-shrink: 0;
  filter: drop-shadow(0 2px 4px rgba(0, 0, 0, 0.2));
}

.deploy-success-content {
  flex: 1;
  min-width: 0;
}

.deploy-success-title {
  font-size: 0.95rem;
  font-weight: 800;
  color: #ffffff;
  margin-bottom: 0.3rem;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.2);
}

.deploy-success-desc {
  font-size: 0.82rem;
  color: #ecfdf5;
  line-height: 1.45;
  font-weight: 500;
}

.deploy-success-actions {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
  flex-shrink: 0;
  align-items: flex-end;
}

.btn-eject-success {
  padding: 0.5rem 0.95rem;
  font-size: 0.82rem;
  font-weight: 700;
  border: 1px solid rgba(255, 255, 255, 0.3);
  border-radius: 8px;
  cursor: pointer;
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
  color: #ffffff;
  transition: all 0.2s cubic-bezier(0.16, 1, 0.3, 1);
  white-space: nowrap;
  box-shadow: 0 3px 10px rgba(0, 0, 0, 0.25);
}

.btn-eject-success:hover {
  background: linear-gradient(135deg, #34d399 0%, #10b981 100%);
  box-shadow: 0 5px 15px rgba(16, 185, 129, 0.45);
  transform: translateY(-1px);
}

.btn-dismiss {
  padding: 0.25rem 0.5rem;
  font-size: 0.85rem;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  background: rgba(255, 255, 255, 0.15);
  color: #a7f3d0;
  transition: all 0.15s ease;
}

.btn-dismiss:hover {
  background: rgba(255, 255, 255, 0.3);
  color: #ffffff;
}

.vm-box {
  border-top: 1px solid var(--card-border);
  padding-top: 1.25rem;
}

.vm-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.4rem;
}

.vm-title-group h3 {
  font-size: 1rem;
  display: flex;
  align-items: center;
  margin: 0;
}

.optional-badge {
  display: inline-flex;
  align-items: center;
  font-size: 0.72rem;
  font-weight: 500;
  padding: 0.12rem 0.55rem;
  margin-left: 0.55rem;
  border-radius: 20px;
  background: rgba(59, 130, 246, 0.15);
  color: #60a5fa;
  border: 1px solid rgba(59, 130, 246, 0.3);
  vertical-align: middle;
  line-height: 1.2;
}

[data-theme="light"] .optional-badge {
  background: #eff6ff;
  color: #1d4ed8;
  border-color: #93c5fd;
  font-weight: 600;
}

.badge {
  font-size: 0.75rem;
  padding: 0.2rem 0.5rem;
  border-radius: 4px;
}

.badge.success {
  background: rgba(16, 185, 129, 0.15);
  color: var(--success);
}

.badge.muted {
  background: rgba(255, 255, 255, 0.1);
  color: var(--text-muted);
}

.vm-options-bar {
  display: flex;
  align-items: center;
  gap: 1.2rem;
  margin: 0.6rem 0 0.6rem 0;
  flex-wrap: wrap;
}

.vm-selector-container {
  display: flex;
  align-items: center;
  gap: 0.45rem;
  white-space: nowrap;
}

.vm-selector-label {
  font-size: 0.78rem;
  color: var(--text-muted);
  font-weight: 500;
  white-space: nowrap;
}

.vm-select-wrapper {
  position: relative;
  display: inline-flex;
  align-items: center;
}

.vm-select {
  appearance: none;
  -webkit-appearance: none;
  background: rgba(16, 185, 129, 0.12);
  color: var(--success);
  border: 1px solid rgba(16, 185, 129, 0.3);
  padding: 0.2rem 1.4rem 0.2rem 0.6rem;
  border-radius: 6px;
  font-size: 0.78rem;
  font-weight: 600;
  cursor: pointer;
  outline: none;
  transition: all 0.2s ease;
}

.vm-select:hover {
  background: rgba(16, 185, 129, 0.2);
  border-color: rgba(16, 185, 129, 0.5);
}

.vm-select option {
  background: var(--card-bg, #1e293b);
  color: var(--text-color, #f8fafc);
}

.select-arrow {
  position: absolute;
  right: 0.4rem;
  font-size: 0.68rem;
  color: var(--success);
  pointer-events: none;
}

[data-theme="light"] .vm-select {
  background: #ecfdf5;
  color: #047857;
  border-color: #a7f3d0;
}

[data-theme="light"] .vm-select option {
  background: #ffffff;
  color: #0f172a;
}

.vm-desc {
  font-size: 0.8rem;
  color: var(--text-muted);
  margin-bottom: 1rem;
}

.fs-selector {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
  margin-bottom: 1.25rem;
}

.fs-label {
  font-size: 0.825rem;
  color: var(--text-muted);
  font-weight: 600;
}

.fs-select {
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  border-radius: 8px;
  color: var(--text-main);
  padding: 0.55rem 0.75rem;
  font-size: 0.825rem;
  outline: none;
  cursor: pointer;
  transition: all 0.2s ease;
}

.fs-select:focus {
  border-color: var(--accent-cyan);
  box-shadow: 0 0 12px rgba(0, 229, 255, 0.25);
}

.settings-icon-btn {
  background: transparent;
  border: none;
  font-size: 1.1rem;
  padding: 0.5rem 0.75rem;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
  justify-content: center;
}

.settings-icon-btn:hover {
  background: rgba(0, 229, 255, 0.12);
  transform: rotate(30deg);
}

.settings-icon-btn.log-toggle-btn:hover {
  transform: none;
}

.settings-icon-btn.log-toggle-btn.active {
  background: rgba(0, 229, 255, 0.18);
  border: 1px solid rgba(0, 229, 255, 0.4);
  box-shadow: 0 0 10px rgba(0, 229, 255, 0.2);
}

[data-theme="light"] .settings-icon-btn.log-toggle-btn.active {
  background: rgba(99, 102, 241, 0.15);
  border: 1px solid rgba(99, 102, 241, 0.35);
  box-shadow: 0 0 10px rgba(99, 102, 241, 0.15);
}

.btn-close-log {
  color: var(--text-muted, #94a3b8);
  font-size: 1.1rem;
  font-weight: bold;
  padding: 0.2rem 0.5rem;
  border-radius: 6px;
  cursor: pointer;
  line-height: 1;
  transition: all 0.2s ease;
  margin-left: 0.25rem;
}

.btn-close-log:hover {
  color: #ef4444;
  background: rgba(239, 68, 68, 0.15);
}

.target-highlight {
  color: #38bdf8;
  font-weight: 600;
}

.target-warn {
  color: #facc15;
  font-weight: 600;
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
  color: #ffffff;
  box-shadow: 0 12px 32px rgba(0, 0, 0, 0.4);
  backdrop-filter: blur(16px);
  border: 1px solid rgba(255, 255, 255, 0.15);
  animation: slideDown 0.3s cubic-bezier(0.16, 1, 0.3, 1);
}

.global-toast.warning {
  background: rgba(234, 179, 8, 0.95);
  border-color: #facc15;
  color: #1e1b4b;
}

.global-toast.error {
  background: rgba(239, 68, 68, 0.95);
  border-color: #f87171;
  color: #ffffff;
}

.global-toast.success {
  background: rgba(34, 197, 94, 0.95);
  border-color: #4ade80;
  color: #064e3b;
}

.global-toast.info {
  background: rgba(59, 130, 246, 0.95);
  border-color: #60a5fa;
  color: #ffffff;
}

.toast-close {
  background: transparent;
  border: none;
  color: currentColor;
  font-size: 1rem;
  cursor: pointer;
  opacity: 0.8;
  margin-left: 8px;
}

.toast-close:hover {
  opacity: 1;
}

@keyframes slideDown {
  from {
    opacity: 0;
    transform: translate(-50%, -20px);
  }
  to {
    opacity: 1;
    transform: translate(-50%, 0);
  }
}

/* ISO Source Card Styles */
.iso-card {
  background: rgba(255, 255, 255, 0.03);
  border: 1px dashed rgba(0, 229, 255, 0.5);
  border-radius: 14px;
  padding: 1.2rem;
  margin-top: 1.2rem;
  margin-bottom: 1.2rem;
  transition: all 0.3s ease;
}

.iso-card:hover {
  border-style: solid;
  border-color: var(--accent-cyan);
  box-shadow: 0 4px 20px rgba(0, 229, 255, 0.12);
}

.iso-card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.9rem;
}

.iso-title-group h3 {
  font-size: 1.02rem;
  font-weight: 700;
  color: #ffffff;
  margin-bottom: 0.2rem;
}

.iso-subtitle {
  font-size: 0.78rem;
  color: var(--text-muted);
  display: block;
}

.add-iso-btn {
  font-size: 0.82rem;
  padding: 0.4rem 0.85rem;
  white-space: nowrap;
}

.iso-list-container {
  background: rgba(0, 0, 0, 0.25);
  border-radius: 10px;
  padding: 0.8rem;
}

.iso-empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 1.6rem 1rem;
  border: 2px dashed rgba(255, 255, 255, 0.15);
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.iso-empty-state:hover {
  background: rgba(0, 229, 255, 0.05);
  border-color: var(--accent-cyan);
}

.empty-icon {
  font-size: 2rem;
  margin-bottom: 0.5rem;
}

.empty-text {
  font-size: 0.9rem;
  font-weight: 600;
  color: #e2e8f0;
}

.empty-subtext {
  font-size: 0.76rem;
  color: var(--text-muted);
  margin-top: 0.25rem;
}

.iso-file-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  max-height: 180px;
  overflow-y: auto;
}

.iso-file-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 8px;
  padding: 0.5rem 0.8rem;
  transition: background 0.2s ease;
}

.iso-file-item:hover {
  background: rgba(255, 255, 255, 0.08);
}

.iso-file-icon {
  font-size: 1.3rem;
}

.iso-file-info {
  flex: 1;
  min-width: 0;
}

.iso-file-name {
  font-size: 0.85rem;
  font-weight: 600;
  color: #f8fafc;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.iso-file-path {
  font-size: 0.74rem;
  color: var(--text-muted);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.iso-remove-btn {
  background: transparent;
  border: none;
  color: #ef4444;
  font-size: 0.95rem;
  cursor: pointer;
  padding: 0.2rem 0.4rem;
  border-radius: 4px;
}

.iso-remove-btn:hover {
  background: rgba(239, 68, 68, 0.2);
}

.iso-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 0.75rem;
  padding-top: 0.5rem;
  border-top: 1px solid rgba(255, 255, 255, 0.08);
  font-size: 0.8rem;
}

.iso-count-summary {
  color: var(--text-muted);
}

.iso-count-summary strong {
  color: var(--accent-cyan);
}

.btn-text-danger {
  background: transparent;
  border: none;
  color: #f87171;
  font-size: 0.78rem;
  cursor: pointer;
}

.btn-text-danger:hover {
  text-decoration: underline;
}

.ventoy-warning-card {
  display: flex;
  align-items: center;
  gap: 12px;
  background: rgba(239, 68, 68, 0.12);
  border: 1px solid rgba(239, 68, 68, 0.35);
  border-radius: 12px;
  padding: 12px 16px;
  margin-bottom: 16px;
}

.warning-card-icon {
  font-size: 22px;
  flex-shrink: 0;
}

.warning-card-body {
  flex: 1;
}

.warning-card-title {
  font-weight: 600;
  font-size: 13px;
  color: #f87171;
  margin-bottom: 4px;
}

.warning-card-message {
  font-size: 12px;
  color: rgba(255, 255, 255, 0.85);
  line-height: 1.5;
}

/* Light Theme Contrast Overrides */
[data-theme="light"] .mode-tabs {
  background: #e2e8f0;
  border-color: #cbd5e1;
}

[data-theme="light"] .tab-btn {
  color: #475569;
}

[data-theme="light"] .tab-btn:hover {
  color: #0f172a;
}

[data-theme="light"] .tab-btn.active {
  background: #0284c7;
  color: #ffffff;
}

[data-theme="light"] .selection-controls {
  background: #f1f5f9;
  border-color: #cbd5e1;
}

[data-theme="light"] .sub-tab-btn {
  background: #ffffff;
  color: #475569;
  border-color: #cbd5e1;
}

[data-theme="light"] .sub-tab-btn.active {
  background: #e0f2fe;
  color: #0284c7;
  border-color: #38bdf8;
}

[data-theme="light"] .selection-count {
  color: #475569;
}

[data-theme="light"] .safe-mode-notice {
  background: #f0f9ff;
  border-color: #7dd3fc;
}

[data-theme="light"] .safe-notice-title {
  color: #0369a1;
}

[data-theme="light"] .safe-notice-desc {
  color: #0c4a6e;
}

[data-theme="light"] .safe-notice-desc b {
  color: #0284c7;
}

[data-theme="light"] .warn-cloudmode-notice {
  background: #fffbeb;
  border-color: #fde68a;
}

[data-theme="light"] .warn-notice-title {
  color: #b45309;
}

[data-theme="light"] .warn-notice-desc {
  color: #78350f;
}

[data-theme="light"] .warn-notice-desc b {
  color: #d97706;
}

[data-theme="light"] .ventoy-warning-card {
  background: #fef2f2;
  border-color: #fca5a5;
}

[data-theme="light"] .warning-card-title {
  color: #dc2626;
}

[data-theme="light"] .warning-card-message {
  color: #991b1b;
}

[data-theme="light"] .settings-icon-btn {
  background: #ffffff;
  color: #475569;
  border: 1px solid #cbd5e1;
}

[data-theme="light"] .settings-icon-btn:hover {
  background: #e0f2fe;
  color: #0284c7;
  border-color: #38bdf8;
}

[data-theme="light"] .deploy-box {
  background: #f1f5f9;
  border: 1px solid #cbd5e1;
}

[data-theme="light"] .selected-target {
  color: #334155;
}

[data-theme="light"] .iso-card {
  background: #f8fafc;
  border-color: #93c5fd;
}

[data-theme="light"] .iso-title-group h3 {
  color: #0f172a;
}

[data-theme="light"] .iso-list-container {
  background: #f1f5f9;
  border: 1px solid #cbd5e1;
}

[data-theme="light"] .iso-empty-state {
  border-color: #cbd5e1;
}

[data-theme="light"] .empty-text {
  color: #1e293b;
}

[data-theme="light"] .iso-file-item {
  background: #ffffff;
  border-color: #cbd5e1;
}

[data-theme="light"] .iso-file-name {
  color: #0f172a;
}

[data-theme="light"] .iso-file-path {
  color: #475569;
}

[data-theme="light"] .iso-footer {
  border-top-color: #cbd5e1;
}

[data-theme="light"] .target-highlight {
  color: #0284c7;
}

[data-theme="light"] .target-warn {
  color: #d97706;
}

[data-theme="light"] .lang-pill-btn {
  background: #ffffff;
  color: #0f172a;
  border-color: #cbd5e1;
  font-weight: 600;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.05);
}

[data-theme="light"] .lang-pill-btn:hover {
  background: #f0f9ff;
  color: #0284c7;
  border-color: #38bdf8;
}

[data-theme="light"] .lang-dropdown-menu {
  background: #ffffff;
  border-color: #cbd5e1;
  box-shadow: 0 12px 32px rgba(15, 23, 42, 0.15);
}

[data-theme="light"] .lang-option {
  color: #0f172a;
  font-weight: 600;
}

[data-theme="light"] .lang-option:hover {
  background: #f0f9ff;
  color: #0284c7;
}

[data-theme="light"] .lang-option.active {
  background: #e0f2fe;
  color: #0284c7;
  font-weight: 700;
}

/* Light Mode Overrides for Embedded Log Center Card & Buttons */
[data-theme="light"] .log-section-card {
  background: #ffffff;
  border-color: #cbd5e1;
  box-shadow: 0 4px 20px rgba(15, 23, 42, 0.06);
}

[data-theme="light"] .log-title-group h2 {
  color: #0f172a;
}

[data-theme="light"] .filter-tabs-sm {
  background: #f1f5f9;
  border-color: #cbd5e1;
}

[data-theme="light"] .btn-tab-sm {
  color: #475569;
  font-weight: 600;
}

[data-theme="light"] .btn-tab-sm:hover {
  color: #0f172a;
  background: #e2e8f0;
}

[data-theme="light"] .btn-tab-sm.active {
  background: #0284c7;
  color: #ffffff;
  font-weight: 700;
}

[data-theme="light"] .btn-text-sm {
  background: #ffffff;
  border: 1px solid #cbd5e1;
  color: #0f172a;
  font-weight: 600;
  box-shadow: 0 1px 3px rgba(15, 23, 42, 0.08);
}

[data-theme="light"] .btn-text-sm:hover {
  background: #f0f9ff;
  border-color: #0284c7;
  color: #0284c7;
}

[data-theme="light"] .auto-scroll-label-sm {
  background: #ffffff;
  border: 1px solid #cbd5e1;
  color: #0f172a;
  font-weight: 600;
  box-shadow: 0 1px 3px rgba(15, 23, 42, 0.08);
}

[data-theme="light"] .auto-scroll-label-sm input[type="checkbox"] {
  accent-color: #0284c7;
}

[data-theme="light"] .auto-scroll-label-sm:hover {
  background: #f0f9ff;
  border-color: #0284c7;
  color: #0284c7;
}

[data-theme="light"] .btn-text-danger-sm {
  background: #fef2f2;
  border: 1px solid #fca5a5;
  color: #dc2626;
  font-weight: 600;
  box-shadow: 0 1px 3px rgba(239, 68, 68, 0.08);
}

[data-theme="light"] .btn-text-danger-sm:hover {
  background: #fee2e2;
  border-color: #ef4444;
  color: #b91c1c;
}

/* Light Mode Terminal Window & Log Row Colors */
[data-theme="light"] .embedded-terminal-window {
  background: #f8fafc;
  border-color: #cbd5e1;
  box-shadow: inset 0 2px 4px rgba(15, 23, 42, 0.04);
}

[data-theme="light"] .embedded-terminal-window .empty-logs {
  color: #94a3b8;
}

[data-theme="light"] .embedded-terminal-window .log-row {
  border-bottom-color: #e2e8f0;
}

[data-theme="light"] .embedded-terminal-window .log-time {
  color: #64748b;
}

[data-theme="light"] .embedded-terminal-window .log-msg {
  color: #0f172a;
}

[data-theme="light"] .embedded-terminal-window .log-details {
  color: #64748b;
}

[data-theme="light"] .embedded-terminal-window .log-row.info .log-msg {
  color: #0f172a;
}

[data-theme="light"] .embedded-terminal-window .log-row.warn .log-msg {
  color: #b45309;
}

[data-theme="light"] .embedded-terminal-window .log-row.error .log-msg {
  color: #dc2626;
}

[data-theme="light"] .embedded-terminal-window .log-row.debug .log-msg {
  color: #7e22ce;
}

[data-theme="light"] .embedded-terminal-window .log-level-badge.info {
  background: #ecfdf5;
  color: #047857;
  border: 1px solid #a7f3d0;
}

[data-theme="light"] .embedded-terminal-window .log-level-badge.warn {
  background: #fffbeb;
  color: #b45309;
  border: 1px solid #fde68a;
}

[data-theme="light"] .embedded-terminal-window .log-level-badge.error {
  background: #fef2f2;
  color: #b91c1c;
  border: 1px solid #fca5a5;
}

[data-theme="light"] .embedded-terminal-window .log-level-badge.debug {
  background: #f3e8ff;
  color: #6b21a8;
  border: 1px solid #e9d5ff;
}
</style>

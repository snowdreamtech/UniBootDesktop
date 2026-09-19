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
  runtimeLogs.value = [];
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
  return Boolean(d.isRealVentoy || d.isCloudMode || d.isGenericBoot || isVentoyName);
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
          protocolCode: 'usb2',
          freeSpace: 0,
          freeFormatted: '0 B',
          fileSystem: 'exFAT',
          partitionScheme: 'GPT',
          writable: true,
          serialNumber: '',
          vendorId: '',
          productId: '',
          smartStatus: '',
          busPower: '',
          busPowerUsed: '',
          sectorSize: '',
          transportProtocol: '',
          bootStatus: '',
          controllerVendor: '',
          isRealVentoy: false,
          isCloudMode: false,
          isGenericBoot: false,
          mountPoint: ''
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
          protocolCode: 'usb3_0',
          freeSpace: 0,
          freeFormatted: '0 B',
          fileSystem: 'exFAT',
          partitionScheme: 'GPT',
          writable: true,
          serialNumber: '',
          vendorId: '',
          productId: '',
          smartStatus: '',
          busPower: '',
          busPowerUsed: '',
          sectorSize: '',
          transportProtocol: '',
          bootStatus: '',
          controllerVendor: '',
          isRealVentoy: false,
          isCloudMode: false,
          isGenericBoot: false,
          mountPoint: ''
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
          protocolCode: 'usb3_1',
          freeSpace: 0,
          freeFormatted: '0 B',
          fileSystem: 'exFAT',
          partitionScheme: 'GPT',
          writable: true,
          serialNumber: '',
          vendorId: '',
          productId: '',
          smartStatus: '',
          busPower: '',
          busPowerUsed: '',
          sectorSize: '',
          transportProtocol: '',
          bootStatus: '',
          controllerVendor: '',
          isRealVentoy: false,
          isCloudMode: false,
          isGenericBoot: false,
          mountPoint: ''
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
  if (isDeploying.value) return;

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
    window.runtime.EventsOn('cloud-deploy-progress', (progress: number) => {
      deployProgress.value = progress;
    });
    unsubCloudProgress = () => {
      if (window.runtime && window.runtime.EventsOff) {
        window.runtime.EventsOff('cloud-deploy-progress');
      }
    };
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

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
        <button class="toast-close" @click="toastMessage = ''">✕</button>
      </div>
    </transition>

    <!-- Header -->
    <header class="app-header">
      <div class="brand">
        <span class="logo">🚀</span>
        <div>
          <h1>{{ t('app.title') }}</h1>
          <span class="sub-brand">{{ t('app.subtitle') }}</span>
        </div>
      </div>
      <div class="mode-tabs">
        <button 
          class="tab-btn" 
          :class="{ active: activeMode === 'cloud' }"
          @click="selectMode('cloud')"
        >
          ⚡ {{ t('mode.cloud') }}
        </button>
        <button 
          class="tab-btn" 
          :class="{ active: activeMode === 'hybrid' }"
          @click="selectMode('hybrid')"
        >
          🛠️ {{ t('mode.hybrid') }}
        </button>

        <!-- Header Quick Language Switcher Dropdown -->
        <div class="lang-selector-header" ref="langDropdownRef">
          <button 
            class="lang-pill-btn" 
            :title="t('settings.language')"
            @click.stop="toggleLangMenu"
          >
            <span class="lang-icon">🌐</span>
            <span class="lang-label">{{ currentLangLabel }}</span>
            <span class="dropdown-caret">▾</span>
          </button>

          <transition name="dropdown-fade">
            <div v-if="isLangMenuOpen" class="lang-dropdown-menu" @click.stop>
              <button 
                v-for="opt in langOptions" 
                :key="opt.value"
                class="lang-option"
                :class="{ active: currentLang === opt.value }"
                @click="selectLanguage(opt.value)"
              >
                <span class="opt-text">{{ opt.label }}</span>
                <span v-if="currentLang === opt.value" class="opt-check">✓</span>
              </button>
            </div>
          </transition>
        </div>

        <button 
          class="settings-icon-btn log-toggle-btn" 
          :class="{ active: isLogCardVisible }"
          :title="t('log.title')"
          @click="toggleLogCard"
        >
          📜
        </button>

        <button 
          class="settings-icon-btn" 
          :title="t('settings.title')"
          @click="openSettings('general')"
        >
          ⚙️
        </button>

        <button 
          class="settings-icon-btn" 
          :title="t('about.title')"
          @click="isAboutOpen = true"
        >
          ℹ️
        </button>
      </div>
    </header>

    <!-- Main Grid -->
    <main class="content-grid">
      <!-- Left: Disk Selection -->
      <section class="glass-card section-card">
        <div class="section-header-row">
          <div>
            <h2>{{ t('disk.select_title') }}</h2>
            <p class="section-desc">{{ t('disk.select_desc') }}</p>
          </div>
        </div>

        <!-- Mode & Selection controls -->
        <div class="selection-controls">
          <div class="selection-mode-toggle">
            <button 
              class="sub-tab-btn" 
              :class="{ active: selectionMode === 'single' }" 
              @click="setSelectionMode('single')"
            >
              {{ t('disk.single_mode') }}
            </button>
            <button 
              class="sub-tab-btn" 
              :class="{ active: selectionMode === 'batch' }" 
              @click="setSelectionMode('batch')"
            >
              {{ t('disk.batch_mode') }}
            </button>
          </div>

          <div v-if="selectionMode === 'batch'" class="batch-actions">
            <button class="btn-text" @click="selectAllDisks">{{ t('disk.select_all') }}</button>
            <button class="btn-text" @click="deselectAllDisks">{{ t('disk.clear_select') }}</button>
            <button 
              class="btn-eject" 
              :disabled="selectedDevices.size === 0" 
              :title="t('disk.batch_eject')" 
              @click="handleBatchEjectDisks"
            >
              ⏏️ {{ t('disk.batch_eject') }}
            </button>
            <span class="selection-count">{{ t('disk.selected_count', { count: selectedDevices.size, total: diskList.length }) }}</span>
          </div>
        </div>

        <div class="disk-list">
          <DiskCard
            v-for="disk in diskList"
            :key="disk.device"
            :disk="disk"
            :isBatchMode="selectionMode === 'batch'"
            :isSelected="selectionMode === 'single' ? selectedDisk?.device === disk.device : selectedDevices.has(disk.device)"
            :customIcon="getCustomIcon(disk)"
            @select="onDiskSelect(disk)"
            @toggle="onDiskToggle(disk)"
            @pick-icon="openIconPicker(disk)"
            @inspect="openInspector(disk)"
            @eject="handleEjectDisk(disk)"
          />
          <div v-if="diskList.length === 0" class="empty-state">
            <span v-if="isScanningDisks">🔍 {{ t('disk.scanning') }}</span>
            <span v-else>⚠️ {{ t('disk.empty_list') }}</span>
          </div>
        </div>

        <button class="btn-secondary refresh-btn" @click="refreshDisks">
          🔄 {{ t('disk.rescan') }}
        </button>
      </section>

      <!-- Right: Deployment & Testing Panel -->
      <section class="glass-card section-card">
        <h2>{{ t('deploy.title') }}</h2>
        <p class="section-desc" v-if="activeMode === 'cloud'">
          {{ t('deploy.desc_cloud') }}
        </p>
        <p class="section-desc" v-else>
          {{ t('deploy.desc_hybrid') }}
        </p>

        <!-- Filesystem Selection for Mode A & Mode B (Hidden when upgrading an existing Ventoy/UniBoot drive) -->
        <div v-if="!isNonDestructive" class="fs-selector">
          <label class="fs-label">{{ t('settings.default_fs') }}</label>
          <CustomSelect
            v-model="selectedFsType"
            :options="[
              { value: 'exFAT', label: t('fs.exfat') },
              { value: 'NTFS', label: t('fs.ntfs') },
              { value: 'FAT32', label: t('fs.fat32') },
              { value: 'ext4', label: t('fs.ext4') }
            ]"
          />
        </div>

        <!-- Ventoy CLI Pre-flight Requirement Notice Banner (Mode A) -->
        <div v-if="activeMode === 'hybrid' && !isNonDestructive && !ventoyStatus.valid" class="ventoy-warning-card">
          <span class="warning-card-icon">⚠️</span>
          <div class="warning-card-body">
            <div class="warning-card-title">{{ isMacOs ? t('deploy.macos_alert_title') : t('deploy.no_ventoy_title') }}</div>
            <div class="warning-card-message">{{ isMacOs ? t('deploy.macos_alert_desc') : t('deploy.no_ventoy_desc') }}</div>
          </div>
          <button class="btn-secondary btn-sm" @click="openSettings('ventoy')">
            ⚙️ {{ t('settings.title') }}
          </button>
        </div>

        <!-- Safe Mode Notice Banner when upgrading an existing Ventoy/UniBoot drive -->
        <div v-if="isNonDestructive" class="safe-mode-notice">
          <span class="safe-notice-icon">🛡️</span>
          <div class="safe-notice-content">
            <div class="safe-notice-title">
              {{ activeMode === 'cloud' ? t('safe.title_cloud') : t('safe.title_hybrid') }}
            </div>
            <div class="safe-notice-desc">
              {{ activeMode === 'cloud' ? t('safe.desc_cloud') : t('safe.desc_hybrid') }}
            </div>
          </div>
        </div>

        <!-- Local ISO/IMG Image Source Selection Card (Mode A) -->
        <div v-if="activeMode === 'hybrid'" class="iso-card">
          <div class="iso-card-header">
            <div class="iso-title-group">
              <h3>
                {{ t('iso.title') }}
                <span class="optional-badge">{{ t('common.optional') }}</span>
              </h3>
              <span class="iso-subtitle">{{ t('iso.desc') }}</span>
            </div>
            <button class="btn-secondary add-iso-btn" @click="handleSelectIsoFiles">
              {{ t('iso.add_btn') }}
            </button>
          </div>

          <div class="iso-list-container">
            <div v-if="selectedIsoFiles.length === 0" class="iso-empty-state" @click="handleSelectIsoFiles">
              <span class="empty-icon">📥</span>
              <div class="empty-text">{{ t('iso.empty_title') }}</div>
              <div class="empty-subtext">{{ t('iso.empty_sub') }}</div>
            </div>

            <div v-else class="iso-file-list">
              <div v-for="(file, index) in selectedIsoFiles" :key="index" class="iso-file-item">
                <span class="iso-file-icon">{{ getFileIcon(file.name) }}</span>
                <div class="iso-file-info">
                  <div class="iso-file-name" :title="file.path">{{ file.name }}</div>
                  <div class="iso-file-path">{{ file.path }}</div>
                </div>
                <button class="iso-remove-btn" title="Remove" @click="removeIsoFile(index)">✕</button>
              </div>
            </div>

            <div v-if="selectedIsoFiles.length > 0" class="iso-footer">
              <span class="iso-count-summary">{{ t('iso.summary', { count: selectedIsoFiles.length }) }}</span>
              <button class="btn-text-danger" @click="clearIsoFiles">{{ t('iso.clear') }}</button>
            </div>
          </div>
        </div>

        <div class="deploy-box">
          <div class="selected-target">
            <span>{{ t('deploy.target_device') }}</span>
            <strong v-if="selectionMode === 'single'">
              {{ selectedDisk ? selectedDisk.name + ' (' + selectedDisk.device + ')' : t('disk.no_disk') }}
            </strong>
            <strong v-else>
              {{ selectedDevices.size > 0 ? t('deploy.batch_target', { count: selectedDevices.size }) : t('disk.no_disk') }}
            </strong>
          </div>

          <ProgressBar 
            v-if="isDeploying" 
            :label="t('deploy.writing')" 
            :progress="deployProgress" 
          />

          <button 
            class="btn-primary deploy-btn" 
            :class="{ 'safe-btn': isNonDestructive, 'danger-disabled': activeMode === 'hybrid' && !isNonDestructive && !ventoyStatus.valid }"
            :disabled="isDeploying"
            :title="deployDisabledReason"
            @click="handleDeployBtnClick"
          >
            {{ deployBtnText }}
          </button>

          <!-- Deploy success banner with Safely Eject button -->
          <transition name="toast-fade">
            <div v-if="deploySuccessBanner.visible" class="deploy-success-banner">
              <div class="deploy-success-icon">🎉</div>
              <div class="deploy-success-content">
                <div class="deploy-success-title">{{ t('deploy.success_banner_title') }}</div>
                <div class="deploy-success-desc">
                  {{ deploySuccessBanner.autoEjected ? t('deploy.toast_auto_ejected', { count: deploySuccessBanner.targets.length }) : t('deploy.success_banner_desc') }}
                </div>
              </div>
              <div class="deploy-success-actions">
                <button
                  v-if="!deploySuccessBanner.autoEjected"
                  id="btn-safely-eject-after-deploy"
                  class="btn-eject-success"
                  @click="handleSafelyEjectAfterDeploy"
                >
                  ⏏️ {{ t('deploy.safely_eject_btn') }}
                </button>
                <button class="btn-dismiss" @click="deploySuccessBanner.visible = false">✕</button>
              </div>
            </div>
          </transition>
        </div>


        <!-- QEMU Preview -->
        <div class="qemu-box">
          <div class="qemu-header">
            <div class="qemu-title-group">
              <h3>
                {{ t('qemu.title') }}
                <span class="optional-badge">{{ t('common.optional') }}</span>
              </h3>
            </div>
            <span class="badge" :class="qemuStatus.installed ? 'success' : 'muted'">
              {{ qemuStatus.installed ? t('qemu.installed') : t('qemu.not_installed') }}
            </span>
          </div>
          <p class="qemu-desc">
            {{ t('qemu.target') }} 
            <strong v-if="activeQemuTargetDevice" class="target-highlight">
              {{ activeQemuTargetName }} ({{ activeQemuTargetDevice }})
            </strong>
            <span v-else class="target-warn">
              {{ t('qemu.no_disk_warn') }}
            </span>
          </p>
          <button 
            class="btn-secondary" 
            :disabled="isQemuDisabled" 
            :title="qemuDisabledReason"
            @click="launchQEMU"
          >
            {{ isLaunchingQemu ? t('qemu.launching') : t('qemu.run_test') }}
          </button>
        </div>
      </section>

      <!-- Embedded Log Center Card (主页面日志中心卡片) -->
      <transition name="toast-fade">
        <section v-if="isLogCardVisible" class="glass-card log-section-card">
          <div class="log-section-header">
            <div class="log-title-group">
              <h2>📜 {{ t('log.title') }}</h2>
              <span class="badge live-badge">● {{ t('log.live') }}</span>
            </div>

            <div class="log-section-controls">
              <div class="filter-tabs-sm">
                <button 
                  v-for="level in logLevels" 
                  :key="level.key"
                  class="btn-tab-sm"
                  :class="{ active: currentEmbeddedLogFilter === level.key }"
                  @click="currentEmbeddedLogFilter = level.key"
                >
                  {{ level.label }}
                </button>
              </div>

              <label class="auto-scroll-label-sm">
                <input type="checkbox" v-model="embeddedAutoScroll" />
                {{ t('log.auto_scroll') }}
              </label>

              <button class="btn-text-sm" @click="handleCopyEmbeddedLogs">📋 {{ t('log.copy') }}</button>
              <button class="btn-text-sm" @click="handleExportEmbeddedLogs">📥 {{ t('log.export') }}</button>
              <button class="btn-text-danger-sm" @click="runtimeLogs = []">🗑️ {{ t('log.clear') }}</button>
              <button 
                class="btn-text-sm btn-close-log" 
                :title="t('common.close')" 
                @click="toggleLogCard"
              >
                ✕
              </button>
            </div>
          </div>

          <div class="embedded-terminal-window" ref="embeddedTerminalRef">
            <div v-if="filteredEmbeddedLogs.length === 0" class="empty-logs">
              {{ t('log.empty') }}
            </div>
            <div 
              v-for="log in filteredEmbeddedLogs" 
              :key="log.id || String(log.timestamp)"
              class="log-row"
              :class="log.level.toLowerCase()"
            >
              <span class="log-time">{{ formatLogTime(log.timestamp) }}</span>
              <span class="log-level-badge" :class="log.level.toLowerCase()">[{{ log.level }}]</span>
              <span class="log-msg">{{ log.message }}</span>
              <span v-if="log.details" class="log-details">{{ log.details }}</span>
            </div>
          </div>
        </section>
      </transition>
    </main>

    <!-- Icon Picker Modal -->
    <IconPickerModal
      :isOpen="isPickerOpen"
      :diskName="targetPickerDisk?.name || targetPickerDisk?.device || ''"
      :currentIcon="targetPickerDisk ? customIcons[targetPickerDisk.device] : undefined"
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
import { ref, computed, onMounted, onUnmounted, watch, nextTick } from 'vue';
import DiskCard from './components/DiskCard.vue';
import ProgressBar from './components/ProgressBar.vue';
import IconPickerModal, { DiskIconType } from './components/IconPickerModal.vue';
import UsbInspectorModal from './components/UsbInspectorModal.vue';
import DeployConfirmModal from './components/DeployConfirmModal.vue';
import SettingsModal from './components/SettingsModal.vue';
import VentoyAlertModal from './components/VentoyAlertModal.vue';
import DiagnosticsModal, { InstallDiagnosticsData } from './components/DiagnosticsModal.vue';
import AboutModal from './components/AboutModal.vue';
import type { LogItem } from './components/LogViewerModal.vue';
import CustomSelect from './components/CustomSelect.vue';
import { t, currentLang, setLanguage, SUPPORTED_LANGUAGES } from './i18n';

const isLangMenuOpen = ref(false);
const langDropdownRef = ref<HTMLElement | null>(null);

const langOptions = computed(() => [
  { value: 'auto', label: '🌐 ' + t('common.autoDetect') },
  ...SUPPORTED_LANGUAGES.map(item => ({
    value: item.code,
    label: item.nativeName
  }))
]);

const currentLangLabel = computed(() => {
  if (currentLang.value === 'auto') {
    return '🌐 ' + t('common.langAuto');
  }
  const opt = langOptions.value.find(o => o.value === currentLang.value);
  return opt ? opt.label : t('common.lang');
});

function toggleLangMenu() {
  isLangMenuOpen.value = !isLangMenuOpen.value;
}

function selectLanguage(langVal: string) {
  setLanguage(langVal);
  isLangMenuOpen.value = false;
  saveLangToConfig(langVal);
  if (window.go && window.go.main && window.go.main.App && window.go.main.App.ReloadAppMenu) {
    window.go.main.App.ReloadAppMenu(langVal).catch((err: any) => {
      console.warn('Failed to reload app menu:', err);
    });
  }
}

async function saveLangToConfig(langVal: string) {
  if (window.go && window.go.main && window.go.main.App && window.go.main.App.GetConfig && window.go.main.App.SaveConfig) {
    try {
      const cfg = await window.go.main.App.GetConfig();
      if (cfg) {
        cfg.language = langVal;
        await window.go.main.App.SaveConfig(cfg);
      }
    } catch (e) {
      console.error('Failed to save language config:', e);
    }
  }
}

function handleGlobalClick(event: MouseEvent) {
  if (langDropdownRef.value && !langDropdownRef.value.contains(event.target as Node)) {
    isLangMenuOpen.value = false;
  }
}

interface DiskInfo {
  device: string;
  name: string;
  size: number;
  formatted: string;
  freeSpace?: number;
  freeFormatted?: string;
  isRemovable: boolean;
  isSystem: boolean;
  usbVersion?: string;
  usbSpeed?: string;
  vendor?: string;
  fileSystem?: string;
  partitionScheme?: string;
  writable?: boolean;
  serialNumber?: string;
  vendorId?: string;
  productId?: string;
  smartStatus?: string;
  busPower?: string;
  busPowerUsed?: string;
  sectorSize?: string;
  transportProtocol?: string;
  bootStatus?: string;
  controllerVendor?: string;
  isFakeUsb3?: boolean;
  protocolCode?: string;
  isVentoy?: boolean;
  isRealVentoy?: boolean;
  isModeB?: boolean;
  isGenericBoot?: boolean;
  mountPoint?: string;
}

const activeMode = ref<'cloud' | 'hybrid'>('cloud');
const selectionMode = ref<'single' | 'batch'>('single');
const selectedFsType = ref<'exFAT' | 'NTFS' | 'FAT32' | 'ext4'>('exFAT');
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
const embeddedTerminalRef = ref<HTMLDivElement | null>(null);

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

function formatLogTime(ts: string | Date): string {
  if (!ts) return '';
  const date = new Date(ts);
  if (isNaN(date.getTime())) return String(ts);
  const hours = date.getHours().toString().padStart(2, '0');
  const minutes = date.getMinutes().toString().padStart(2, '0');
  const seconds = date.getSeconds().toString().padStart(2, '0');
  const ms = date.getMilliseconds().toString().padStart(3, '0');
  return `${hours}:${minutes}:${seconds}.${ms}`;
}

const filteredEmbeddedLogs = computed(() => {
  return runtimeLogs.value.filter(log => {
    if (currentEmbeddedLogFilter.value === 'ALL') return true;
    return (log.level || '').toUpperCase() === currentEmbeddedLogFilter.value;
  });
});

function handleCopyEmbeddedLogs() {
  if (filteredEmbeddedLogs.value.length === 0) {
    showToast(t('log.empty'), 'info');
    return;
  }
  const text = filteredEmbeddedLogs.value
    .map(l => `[${formatLogTime(l.timestamp)}] [${l.level}] ${l.message}${l.details ? ' - ' + l.details : ''}`)
    .join('\n');
  navigator.clipboard.writeText(text);
  showToast(t('log.copied_toast'), 'success');
}

async function handleExportEmbeddedLogs() {
  if (filteredEmbeddedLogs.value.length === 0) {
    showToast(t('log.empty'), 'info');
    return;
  }
  const text = filteredEmbeddedLogs.value
    .map(l => `[${formatLogTime(l.timestamp)}] [${l.level}] ${l.message}${l.details ? ' - ' + l.details : ''}`)
    .join('\n');

  if (window.go && window.go.main && window.go.main.App && (window.go.main.App as any).ExportLogs) {
    try {
      const filePath = await (window.go.main.App as any).ExportLogs(text);
      if (filePath) {
        showToast(t('log.copied_toast'), 'success');
      }
    } catch (e) {
      console.error('Failed to export logs:', e);
    }
  } else {
    // Fallback for browser dev mode
    const blob = new Blob([text], { type: 'text/plain;charset=utf-8' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `unigodesktop-log-${new Date().toISOString().slice(0, 10)}.log`;
    a.click();
    URL.revokeObjectURL(url);
    showToast(t('log.copied_toast'), 'success');
  }
}

function scrollToEmbeddedTerminalBottom() {
  if (embeddedAutoScroll.value && embeddedTerminalRef.value) {
    nextTick(() => {
      if (embeddedTerminalRef.value) {
        embeddedTerminalRef.value.scrollTop = embeddedTerminalRef.value.scrollHeight;
      }
    });
  }
}

watch(() => runtimeLogs.value.length, () => {
  scrollToEmbeddedTerminalBottom();
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
  showToast(t('diag.toast_copied'), 'info');
}

function handleRetryDeploy() {
  isDiagnosticsOpen.value = false;
  openDeployConfirm();
}

function openSettings(tab: 'general' | 'network' | 'uniboot' | 'ventoy' = 'general') {
  settingsInitialTab.value = tab;
  isSettingsOpen.value = true;
}

const currentGithubProxy = ref('');
const pendingTargets = ref<string[]>([]);
const pendingTargetSnapshots = ref<DiskInfo[]>([]);
const isDeploying = ref(false);
const deployProgress = ref(0);
const autoEjectAfterDeploy = ref(false);
const deploySuccessBanner = ref<{ visible: boolean; msg: string; targets: string[]; autoEjected?: boolean }>({ visible: false, msg: '', targets: [], autoEjected: false });
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
  if (window.go && window.go.main && window.go.main.App && window.go.main.App.SelectIsoFiles) {
    try {
      const paths: string[] = await window.go.main.App.SelectIsoFiles();
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
  selectedIsoFiles.value.splice(index, 1);
}

function clearIsoFiles() {
  selectedIsoFiles.value = [];
}

function getFileIcon(filename: string) {
  const ext = filename.split('.').pop()?.toLowerCase();
  switch (ext) {
    case 'iso':
      return '💿';
    case 'wim':
    case 'img':
    case 'raw':
      return '📦';
    case 'vhd':
    case 'vhdx':
    case 'vti':
      return '💾';
    case 'efi':
    case 'bin':
      return '⚡';
    default:
      return '📄';
  }
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

function getCustomIcon(disk: DiskInfo): DiskIconType | undefined {
  const fp = getDiskFingerprint(disk);
  return customIcons.value[fp] || customIcons.value[disk.device];
}

function openInspector(disk: DiskInfo) {
  targetInspectorDisk.value = disk;
  isInspectorOpen.value = true;
}

async function handleEjectDisk(disk: DiskInfo) {
  try {
    if (window.go && window.go.main && window.go.main.App && window.go.main.App.EjectDisk) {
      await window.go.main.App.EjectDisk(disk.device);
    }
    showToast(t('disk.toast_ejected_success', { device: disk.device, name: disk.name || disk.device }), 'success');
    await refreshDisks();
  } catch (err: any) {
    showToast(t('disk.toast_ejected_failed', { device: disk.device, error: err?.toString() || 'Unknown error' }), 'error');
  }
}

async function handleSafelyEjectAfterDeploy() {
  const targets = deploySuccessBanner.value.targets;
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
}

function onIconSelected(type: DiskIconType) {
  if (targetPickerDisk.value) {
    const fp = getDiskFingerprint(targetPickerDisk.value);
    customIcons.value[fp] = type;
    customIcons.value[targetPickerDisk.value.device] = type;
    saveCustomIcons(customIcons.value);
  }
}

function onIconReset() {
  if (targetPickerDisk.value) {
    const fp = getDiskFingerprint(targetPickerDisk.value);
    delete customIcons.value[fp];
    delete customIcons.value[targetPickerDisk.value.device];
    saveCustomIcons(customIcons.value);
  }
}

const isMacOs = computed(() => navigator.userAgent.includes('Mac') || navigator.platform.includes('Mac'));

function checkIsExistingBootDisk(d: DiskInfo): boolean {
  if (!d) return false;
  const nameUpper = (d.name || '').toUpperCase();
  const isVentoyName = nameUpper.includes('VENTOY') || nameUpper.includes('UNIBOOT');
  return Boolean(d.isRealVentoy || d.isModeB || d.isVentoy || isVentoyName);
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

const isQemuDisabled = computed(() => {
  return (
    isLaunchingQemu.value ||
    isDeploying.value ||
    !qemuStatus.value.installed ||
    !activeQemuTargetDevice.value
  );
});

const qemuDisabledReason = computed(() => {
  if (isLaunchingQemu.value) {
    return t('qemu.tip_launching');
  }
  if (isDeploying.value) {
    return t('qemu.tip_deploying');
  }
  if (!qemuStatus.value.installed) {
    return t('qemu.tip_not_installed');
  }
  if (!activeQemuTargetDevice.value) {
    return t('qemu.tip_select_target');
  }
  return t('qemu.tip_ready');
});

const activeQemuTargetDevice = computed(() => {
  if (selectionMode.value === 'single') {
    return selectedDisk.value?.device || '';
  }
  if (selectedDevices.value.size > 0) {
    return Array.from(selectedDevices.value)[0];
  }
  return '';
});

const activeQemuTargetName = computed(() => {
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
  if (newSet.has(disk.device)) {
    newSet.delete(disk.device);
  } else {
    newSet.add(disk.device);
  }
  selectedDevices.value = newSet;
}

const isScanningDisks = ref(false);

function isEqualDiskList(a: DiskInfo[], b: DiskInfo[]): boolean {
  if (!a || !b) return a === b;
  if (a.length !== b.length) return false;
  for (let i = 0; i < a.length; i++) {
    if (
      a[i].device !== b[i].device ||
      a[i].name !== b[i].name ||
      a[i].formatted !== b[i].formatted ||
      a[i].freeFormatted !== b[i].freeFormatted ||
      a[i].bootStatus !== b[i].bootStatus ||
      a[i].mountPoint !== b[i].mountPoint
    ) {
      return false;
    }
  }
  return true;
}

// Wails JS binding fallbacks / mock data for standalone preview
async function refreshDisks() {
  if (isScanningDisks.value) return;
  isScanningDisks.value = true;
  try {
    if (window.go && window.go.main && window.go.main.App) {
      try {
        const fetched = (await window.go.main.App.GetDiskList()) || [];
        if (!isEqualDiskList(diskList.value, fetched)) {
          diskList.value = fetched;
          if (selectedDisk.value) {
            const stillExists = diskList.value.find(d => d.device === selectedDisk.value?.device);
            if (stillExists) {
              selectedDisk.value = stillExists;
            } else {
              selectedDisk.value = diskList.value.length > 0 ? diskList.value[0] : null;
            }
          } else if (diskList.value.length > 0) {
            selectedDisk.value = diskList.value[0];
          }
        }
      } catch (e) {
        console.error(e);
        diskList.value = [];
        selectedDisk.value = null;
      }
    } else {
      // Fallback mock for browser preview demonstrating genuine vs fake USB 3.0
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
      if (!selectedDisk.value) selectedDisk.value = diskList.value[0];
    }
  } finally {
    isScanningDisks.value = false;
  }
}

async function checkQemu() {
  if (window.go && window.go.main && window.go.main.App) {
    qemuStatus.value = await window.go.main.App.CheckQEMU();
  } else {
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
  deployProgress.value = 15;

  const progressTimer = setInterval(() => {
    if (deployProgress.value < 85) {
      deployProgress.value += 15;
    }
  }, 150);

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
          res = await window.go.main.App.DeployModeB(targets[0], selectedFsType.value, expected);
        } else {
          res = await window.go.main.App.DeployModeA(targets[0], selectedFsType.value, isoPaths, expected);
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
          resList = await window.go.main.App.DeployModeBBatch(targets, selectedFsType.value, pendingTargetSnapshots.value);
        } else {
          resList = await window.go.main.App.DeployModeABatch(targets, selectedFsType.value, isoPaths, pendingTargetSnapshots.value);
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
    clearInterval(progressTimer);
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
        autoEjected: autoEjectedCount > 0
      };

      if (autoEjectedCount > 0) {
        showToast(t('deploy.toast_auto_ejected', { count: autoEjectedCount }), 'success');
      }
    }, 200);
  } else {
    isDeploying.value = false;
    deployProgress.value = 0;
    openDiagnosticsModal(latestDiagnostics, resultMsg);
  }
}


async function launchQEMU() {
  console.log('[UniBoot] launchQEMU clicked, diskList:', diskList.value, 'selectedDisk:', selectedDisk.value, 'selectedDevices:', selectedDevices.value);

  // 1. Check if disk list is empty
  if (!diskList.value || diskList.value.length === 0) {
    showToast(t('deploy.toast_no_disks'), 'warning');
    return;
  }

  // 2. Check if a disk is selected (supports both single & batch selection modes)
  const targetDevice = activeQemuTargetDevice.value;
  const diskLabel = activeQemuTargetName.value;

  if (!targetDevice) {
    showToast(t('qemu.toast_select_first'), 'warning');
    return;
  }

  // 3. Check if QEMU is installed
  if (!qemuStatus.value.installed) {
    showToast(t('qemu.toast_not_installed'), 'error');
    return;
  }

  isLaunchingQemu.value = true;

  try {
    if (window.go && window.go.main && window.go.main.App) {
      if (typeof window.go.main.App.LaunchQEMU === 'function') {
        await window.go.main.App.LaunchQEMU(targetDevice);
        showToast(t('qemu.startSuccess', { disk: diskLabel, device: targetDevice }), 'success');
      } else {
        showToast(t('qemu.backendNotReady'), 'warning');
      }
    } else {
      await new Promise(r => setTimeout(r, 600));
      showToast(t('qemu.demoModeStart', { disk: diskLabel, device: targetDevice }), 'info');
    }
  } catch (e: any) {
    console.error('[UniBoot] LaunchQEMU error:', e);
    showToast(t('qemu.startFailed', { error: e?.message || String(e) }), 'error');
  } finally {
    isLaunchingQemu.value = false;
  }
}

onMounted(() => {
  loadConfig();
  refreshDisks();
  checkQemu();
  checkVentoyStatus();
  window.addEventListener('click', handleGlobalClick);

  runtimeLogs.value.push({
    timestamp: new Date().toISOString(),
    level: 'INFO',
    message: 'UniGoDesktop engine ready. Real-time log stream connected.'
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

onUnmounted(() => {
  window.removeEventListener('click', handleGlobalClick);
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

/* Embedded Log Center Card */
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
  margin-bottom: 1rem;
  flex-wrap: wrap;
  gap: 0.75rem;
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

.embedded-terminal-window .log-msg {
  color: #e2e8f0;
  flex: 1;
  min-width: 0;
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
  margin-left: 0.5rem;
  white-space: pre-wrap;
  word-break: break-word;
  overflow-wrap: anywhere;
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

.warn-modeb-notice {
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
  gap: 0.75rem;
  padding: 1rem;
  margin-top: 0.75rem;
  background: linear-gradient(135deg, rgba(16, 185, 129, 0.15) 0%, rgba(5, 150, 105, 0.1) 100%);
  border: 1px solid rgba(16, 185, 129, 0.4);
  border-radius: 12px;
  animation: bannerFadeIn 0.35s ease;
}

@keyframes bannerFadeIn {
  from { opacity: 0; transform: translateY(6px); }
  to   { opacity: 1; transform: translateY(0); }
}

.deploy-success-icon {
  font-size: 1.6rem;
  line-height: 1;
  flex-shrink: 0;
}

.deploy-success-content {
  flex: 1;
  min-width: 0;
}

.deploy-success-title {
  font-size: 0.9rem;
  font-weight: 700;
  color: #6ee7b7;
  margin-bottom: 0.25rem;
}

.deploy-success-desc {
  font-size: 0.78rem;
  color: rgba(255,255,255,0.7);
  line-height: 1.4;
}

.deploy-success-actions {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
  flex-shrink: 0;
}

.btn-eject-success {
  padding: 0.45rem 0.85rem;
  font-size: 0.78rem;
  font-weight: 600;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
  color: #fff;
  transition: all 0.2s ease;
  white-space: nowrap;
  box-shadow: 0 2px 8px rgba(16, 185, 129, 0.4);
}

.btn-eject-success:hover {
  background: linear-gradient(135deg, #34d399 0%, #10b981 100%);
  box-shadow: 0 4px 12px rgba(16, 185, 129, 0.55);
  transform: translateY(-1px);
}

.btn-dismiss {
  padding: 0.3rem 0.6rem;
  font-size: 0.75rem;
  border: 1px solid rgba(255,255,255,0.15);
  border-radius: 6px;
  cursor: pointer;
  background: transparent;
  color: rgba(255,255,255,0.5);
  transition: all 0.15s ease;
}

.btn-dismiss:hover {
  background: rgba(255,255,255,0.08);
  color: rgba(255,255,255,0.8);
}

.qemu-box {
  border-top: 1px solid var(--card-border);
  padding-top: 1.25rem;
}

.qemu-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.4rem;
}

.qemu-title-group h3 {
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

.qemu-desc {
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

[data-theme="light"] .warn-modeb-notice {
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

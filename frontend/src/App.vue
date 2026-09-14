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
                <span class="opt-flag">{{ opt.flag }}</span>
                <span class="opt-text">{{ opt.label }}</span>
                <span v-if="currentLang === opt.value" class="opt-check">✓</span>
              </button>
            </div>
          </transition>
        </div>

        <button 
          class="settings-icon-btn" 
          :title="t('settings.title')"
          @click="openSettings('general')"
        >
          ⚙️
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
          />
          <div v-if="diskList.length === 0" class="empty-state">
            Scanning removable USB drives...
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
          <strong>{{ t('mode.cloud') }}</strong> • Ultra-fast double partition initialization & multi-arch iPXE cloud boot.
        </p>
        <p class="section-desc" v-else>
          Integrated Ventoy core + UniBoot theme & iPXE cloud boot support.
        </p>

        <!-- Filesystem Selection for Mode A & Mode B (Hidden when upgrading an existing Ventoy/UniBoot drive) -->
        <div v-if="!isNonDestructive" class="fs-selector">
          <label class="fs-label">{{ t('settings.default_fs') }}</label>
          <CustomSelect
            v-model="selectedFsType"
            :options="[
              { value: 'exFAT', label: 'exFAT (Default • Supports >4GB ISO)' },
              { value: 'NTFS', label: 'NTFS (Windows Native)' },
              { value: 'FAT32', label: 'FAT32 (Legacy Compatible • 4GB Limit)' },
              { value: 'ext4', label: 'ext4 (Linux Native)' }
            ]"
          />
        </div>

        <!-- Ventoy CLI Pre-flight Requirement Notice Banner (Mode A) -->
        <div v-if="activeMode === 'hybrid' && !isNonDestructive && !ventoyStatus.valid" class="ventoy-warning-card">
          <span class="warning-card-icon">⚠️</span>
          <div class="warning-card-body">
            <div class="warning-card-title">Ventoy CLI Pre-flight Limitation</div>
            <div class="warning-card-message">{{ ventoyStatus.message || 'Ventoy CLI executable is required for Mode A format.' }}</div>
          </div>
          <button class="btn-secondary btn-sm" @click="openSettings('ventoy')">
            ⚙️ Settings
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
              <h3>{{ t('iso.title') }}</h3>
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
            {{ isDeploying ? t('deploy.writing') : (isNonDestructive ? t('deploy.start_update') : (selectionMode === 'batch' ? t('deploy.batch_create', { count: selectedDevices.size }) : t('deploy.start_create'))) }}
          </button>
        </div>

        <!-- QEMU Preview -->
        <div class="qemu-box">
          <div class="qemu-header">
            <h3>{{ t('qemu.title') }}</h3>
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
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue';
import DiskCard from './components/DiskCard.vue';
import ProgressBar from './components/ProgressBar.vue';
import IconPickerModal, { DiskIconType } from './components/IconPickerModal.vue';
import UsbInspectorModal from './components/UsbInspectorModal.vue';
import DeployConfirmModal from './components/DeployConfirmModal.vue';
import SettingsModal from './components/SettingsModal.vue';
import VentoyAlertModal from './components/VentoyAlertModal.vue';
import CustomSelect from './components/CustomSelect.vue';
import { t, currentLang, setLanguage } from './i18n';

const isLangMenuOpen = ref(false);
const langDropdownRef = ref<HTMLElement | null>(null);

const langOptions = [
  { value: 'auto', label: '自动识别 (Auto)', flag: '🌐' },
  { value: 'zh-CN', label: '简体中文', flag: '🇨🇳' },
  { value: 'en-US', label: 'English', flag: '🇺🇸' },
  { value: 'zh-TW', label: '繁體中文', flag: '🇭🇰' },
  { value: 'ja-JP', label: '日本語', flag: '🇯🇵' },
  { value: 'ko-KR', label: '한국어', flag: '🇰🇷' },
  { value: 'de-DE', label: 'Deutsch', flag: '🇩🇪' },
  { value: 'fr-FR', label: 'Français', flag: '🇫🇷' },
  { value: 'es-ES', label: 'Español', flag: '🇪🇸' },
  { value: 'ru-RU', label: 'Русский', flag: '🇷🇺' },
  { value: 'pt-BR', label: 'Português', flag: '🇧🇷' },
  { value: 'it-IT', label: 'Italiano', flag: '🇮🇹' },
  { value: 'tr-TR', label: 'Türkçe', flag: '🇹🇷' },
  { value: 'pl-PL', label: 'Polski', flag: '🇵🇱' },
  { value: 'vi-VN', label: 'Tiếng Việt', flag: '🇻🇳' },
  { value: 'ar-SA', label: 'العربية', flag: '🇸🇦' }
];

const currentLangLabel = computed(() => {
  if (currentLang.value === 'auto') {
    return '语言 / Lang (Auto)';
  }
  const opt = langOptions.find(o => o.value === currentLang.value);
  return opt ? opt.label : '语言 / Lang';
});

function toggleLangMenu() {
  isLangMenuOpen.value = !isLangMenuOpen.value;
}

function selectLanguage(langVal: string) {
  setLanguage(langVal);
  isLangMenuOpen.value = false;
  saveLangToConfig(langVal);
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
  isRealVentoy?: boolean;
  isModeB?: boolean;
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
  showToast('已切换至原生支持的【模式 B (1秒极速云引导盘)】！', 'success');
}

function openSettings(tab: 'general' | 'network' | 'uniboot' | 'ventoy' = 'general') {
  settingsInitialTab.value = tab;
  isSettingsOpen.value = true;
}

const currentGithubProxy = ref('');
const pendingTargets = ref<string[]>([]);
const isDeploying = ref(false);
const deployProgress = ref(0);
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
      showToast(ventoyStatus.value.message || '⚠️ 模式 A 全新制作依赖 Ventoy CLI 环境', 'warning');
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
          showToast(`已成功添加 ${added} 个镜像源文件`, 'success');
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
    showToast('已添加 2 个示例镜像源文件 (浏览器演示)', 'info');
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
        if (cfg.language) setLanguage(cfg.language);
        applyTheme(cfg.theme);
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
      } : {
        mode: activeMode.value,
        autoCheckUpdate: true,
        theme: 'dark',
        githubProxy: proxyUrl,
        fileSystem: selectedFsType.value,
      };

      await window.go.main.App.SaveConfig(configObj as any);
    } catch (e) {
      console.error('Failed to save config:', e);
    }
  }
}

async function handleDeployBtnClick() {
  if (isDeploying.value) return;

  if (selectionMode.value === 'single' && !selectedDisk.value) {
    showToast('⚠️ 请先在左侧磁盘列表中选择目标 U 盘', 'warning');
    return;
  }
  if (selectionMode.value === 'batch' && selectedDevices.value.size === 0) {
    showToast('⚠️ 请先勾选要批量制作的目标 U 盘', 'warning');
    return;
  }

  if (activeMode.value === 'hybrid' && !isNonDestructive.value) {
    await checkVentoyStatus();
    if (!ventoyStatus.value.valid) {
      if (isMacOs.value) {
        openVentoyAlert(
          'macOS 暂不支持 Ventoy CLI 全新格式化',
          '官方 Ventoy 暂不支持在 macOS 上直接运行格式化程序。制作【模式 A】全新盘需依赖 Ventoy CLI；建议直接选择原生支持的【模式 B (1秒极速云引导盘)】！如需使用模式 A，请先在 Win/Linux 上完成 Ventoy 盘初始化后插入 macOS 无损升级。',
          'switch_b'
        );
      } else {
        openVentoyAlert(
          '未检测到 Ventoy CLI 执行文件',
          ventoyStatus.value.message || '全新制作【模式 A (Ventoy 双模盘)】需依赖本地 Ventoy CLI 程序 (Ventoy2Disk)。请先前往设置配置 Ventoy 可执行文件路径，或直接一键切换至不需要 Ventoy CLI 的【模式 B (1秒极速云引导)】！',
          'open_settings'
        );
      }
      return;
    }
  }

  await openDeployConfirm();
}

async function openDeployConfirm() {
  let targets: string[] = [];
  if (selectionMode.value === 'single') {
    if (!selectedDisk.value) return;
    targets = [selectedDisk.value.device];
  } else {
    targets = Array.from(selectedDevices.value);
    if (targets.length === 0) return;
  }

  pendingTargets.value = targets;
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

const isSelectedVentoyDisk = computed(() => {
  if (selectionMode.value === 'single' && selectedDisk.value) {
    if (activeMode.value === 'cloud') {
      return true; // Mode B is ALWAYS non-destructive (flashes ESP partition only)
    }
    // Mode A is ONLY non-destructive if the target disk is ALREADY a REAL Ventoy MBR disk!
    if (selectedDisk.value.isRealVentoy) {
      return true;
    }
    const name = (selectedDisk.value.name || '').toUpperCase();
    const status = (selectedDisk.value.bootStatus || '').toUpperCase();
    if (selectedDisk.value.isModeB || status.includes('模式 B') || status.includes('CLOUD PURE') || status.includes('极速云引导盘')) {
      return false; // Mode B drive is NOT a Ventoy MBR drive, must be formatted via Ventoy CLI to convert to Mode A!
    }
    return status.includes('模式 A') || name.includes('VENTOY') || status.includes('VENTOY');
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
    if (!d) return false;
    if (activeMode.value === 'cloud') return true;
    if (d.isRealVentoy) return true;
    const name = (d.name || '').toUpperCase();
    const status = (d.bootStatus || '').toUpperCase();
    if (d.isModeB || status.includes('模式 B') || status.includes('CLOUD PURE') || status.includes('极速云引导盘')) {
      return false;
    }
    return status.includes('模式 A') || name.includes('VENTOY') || status.includes('VENTOY');
  });
});


const deployDisabledReason = computed(() => {
  if (isDeploying.value) return '正在写入引导固件...';
  if (selectionMode.value === 'single' && !selectedDisk.value) return '请先选择要制作的目标 U 盘';
  if (selectionMode.value === 'batch' && selectedDevices.value.size === 0) return '请先勾选要批量制作的目标 U 盘';
  if (activeMode.value === 'hybrid' && !isNonDestructive.value && !ventoyStatus.value.valid) {
    if (isMacOs.value) {
      return '❌ macOS 平台暂不支持全新格式化制作 Mode A 盘 (请使用模式 B)';
    }
    return ventoyStatus.value.message || '全新制作模式 A 需依赖 Ventoy CLI 环境';
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
    return 'QEMU 模拟器正在拉起启动中...';
  }
  if (isDeploying.value) {
    return '烧录部署中，请等待部署完成后再测试';
  }
  if (!qemuStatus.value.installed) {
    return '未检测到 QEMU 模拟器，请先安装 QEMU (brew/port install qemu)';
  }
  if (!activeQemuTargetDevice.value) {
    return '请先在左侧列表点击选择要测试的目标 U 盘';
  }
  return '点击在当前桌面拉起 QEMU 虚拟机校验 U 盘引导';
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

// Wails JS binding fallbacks / mock data for standalone preview
async function refreshDisks() {
  if (window.go && window.go.main && window.go.main.App) {
    try {
      const fetched = await window.go.main.App.GetDiskList();
      diskList.value = fetched || [];
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

  try {
    if (window.go && window.go.main && window.go.main.App) {
      const isoPaths = selectedIsoFiles.value.map(f => f.path);
      if (targets.length === 1) {
        let res: any;
        if (activeMode.value === 'cloud') {
          res = await window.go.main.App.DeployModeB(targets[0], selectedFsType.value);
        } else {
          res = await window.go.main.App.DeployModeA(targets[0], selectedFsType.value, isoPaths);
        }
        if (res) {
          success = res.success;
          resultMsg = res.message || '';
        }
      } else {
        let resList: any[];
        if (activeMode.value === 'cloud') {
          resList = await window.go.main.App.DeployModeBBatch(targets, selectedFsType.value);
        } else {
          resList = await window.go.main.App.DeployModeABatch(targets, selectedFsType.value, isoPaths);
        }
        if (resList && resList.length > 0) {
          const failed = resList.filter(r => !r.success);
          if (failed.length > 0) {
            success = false;
            resultMsg = failed.map(f => `${f.target}: ${f.message}`).join('\n');
          } else {
            resultMsg = `成功完成 ${resList.length} 块 U 盘的极速云安装盘部署！`;
          }
        }
      }
    } else {
      // Mock execution for browser demo
      await new Promise(r => setTimeout(r, 800));
      resultMsg = `成功部署模式 ${activeMode.value === 'cloud' ? 'B (极速云安装盘)' : 'A (混合双模)'} 到 ${targets.join(', ')}`;
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
      await refreshDisks();
      alert(`🎉 部署成功！\n\n${resultMsg}`);
    }, 200);
  } else {
    isDeploying.value = false;
    deployProgress.value = 0;
    alert(`❌ 部署失败：\n\n${resultMsg}`);
  }
}


async function launchQEMU() {
  console.log('[UniBoot] launchQEMU clicked, diskList:', diskList.value, 'selectedDisk:', selectedDisk.value, 'selectedDevices:', selectedDevices.value);

  // 1. Check if disk list is empty
  if (!diskList.value || diskList.value.length === 0) {
    showToast('⚠️ 当前未检测到任何可用的 U 盘设备！请插入 U 盘后再试。', 'warning');
    return;
  }

  // 2. Check if a disk is selected (supports both single & batch selection modes)
  const targetDevice = activeQemuTargetDevice.value;
  const diskLabel = activeQemuTargetName.value;

  if (!targetDevice) {
    showToast('⚠️ 请先在左侧磁盘列表中点击选择要测试的目标 U 盘！', 'warning');
    return;
  }

  // 3. Check if QEMU is installed
  if (!qemuStatus.value.installed) {
    showToast('❌ 未检测到 QEMU 模拟器！请先安装 QEMU (brew install qemu 或 port install qemu)', 'error');
    return;
  }

  isLaunchingQemu.value = true;

  try {
    if (window.go && window.go.main && window.go.main.App) {
      if (typeof window.go.main.App.LaunchQEMU === 'function') {
        await window.go.main.App.LaunchQEMU(targetDevice);
        showToast(`🚀 已成功启动 QEMU 模拟器校验磁盘：${diskLabel} (${targetDevice})`, 'success');
      } else {
        showToast('⚠️ 后端 API 尚未就绪：Wails 绑定接口加载中，请重新启动 UniBoot 应用。', 'warning');
      }
    } else {
      await new Promise(r => setTimeout(r, 600));
      showToast(`[演示模式] 正在启动 QEMU 模拟器校验：${diskLabel} (${targetDevice})`, 'info');
    }
  } catch (e: any) {
    console.error('[UniBoot] LaunchQEMU error:', e);
    showToast(`❌ 启动 QEMU 模拟器失败：${e?.message || String(e)}`, 'error');
  } finally {
    isLaunchingQemu.value = false;
  }
}

let diskPollTimer: number | undefined;

onMounted(() => {
  loadConfig();
  refreshDisks();
  checkQemu();
  checkVentoyStatus();
  window.addEventListener('click', handleGlobalClick);

  if (window.runtime && window.runtime.EventsOn) {
    window.runtime.EventsOn("iso-copy-progress", (data: any) => {
      if (data) {
        isoCopyStatus.value = `正在写入镜像 (${data.fileIndex}/${data.totalFiles}): ${data.currentFile} (${data.progress.toFixed(1)}%)`;
        deployProgress.value = Math.min(99, Math.max(50, Math.floor(50 + data.progress / 2)));
      }
    });
  }

  // Auto-poll USB drives every 2.5s when idle for instant hotplug detection
  diskPollTimer = window.setInterval(() => {
    if (!isDeploying.value) {
      refreshDisks();
    }
  }, 2500);
});

onUnmounted(() => {
  if (diskPollTimer) {
    clearInterval(diskPollTimer);
  }
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
  color: var(--text-color, #e0e6ed);
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
  color: var(--text-color, #e0e6ed);
  font-size: 0.85rem;
  font-weight: 500;
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
  min-height: 220px;
  max-height: 440px;
  overflow-y: auto;
  padding-right: 0.4rem;
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
  width: 100%;
  margin-top: 1rem;
}

.fs-selector {
  margin-bottom: 1.25rem;
}

.safe-mode-notice {
  display: flex;
  align-items: flex-start;
  gap: 0.8rem;
  background: rgba(0, 229, 255, 0.08);
  border: 1px solid rgba(0, 229, 255, 0.3);
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
  color: var(--accent-cyan);
  font-weight: 700;
  font-size: 0.875rem;
}

.safe-notice-desc {
  color: #a5f3fc;
  font-size: 0.8rem;
  line-height: 1.45;
}

.safe-notice-desc b {
  color: #fff;
}

.warn-modeb-notice {
  display: flex;
  align-items: flex-start;
  gap: 0.8rem;
  background: rgba(245, 158, 11, 0.1);
  border: 1px solid rgba(245, 158, 11, 0.35);
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
  color: #fbbf24;
  font-weight: 700;
  font-size: 0.875rem;
}

.warn-notice-desc {
  color: #fef08a;
  font-size: 0.8rem;
  line-height: 1.45;
}

.warn-notice-desc b {
  color: #fff;
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

.qemu-header h3 {
  font-size: 1rem;
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
</style>

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
          <h1>UniGoDesktop</h1>
          <span class="sub-brand">UniBoot Desktop Engine</span>
        </div>
      </div>
      <div class="mode-tabs">
        <button 
          class="tab-btn" 
          :class="{ active: activeMode === 'cloud' }"
          @click="selectMode('cloud')"
        >
          ⚡ 模式 B (1秒极速云安装盘)
        </button>
        <button 
          class="tab-btn" 
          :class="{ active: activeMode === 'hybrid' }"
          @click="selectMode('hybrid')"
        >
          🛠️ 模式 A (全能双模盘)
        </button>
        <button 
          class="settings-icon-btn" 
          title="系统与 GitHub 代理加速设置"
          @click="isSettingsOpen = true"
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
            <h2>选择目标 U 盘</h2>
            <p class="section-desc">仅自动扫描安全的可移动 U 盘，系统盘自动过滤保护。</p>
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
              🎯 单选模式
            </button>
            <button 
              class="sub-tab-btn" 
              :class="{ active: selectionMode === 'batch' }" 
              @click="setSelectionMode('batch')"
            >
              📦 批量多选模式
            </button>
          </div>

          <div v-if="selectionMode === 'batch'" class="batch-actions">
            <button class="btn-text" @click="selectAllDisks">✅ 全选</button>
            <button class="btn-text" @click="deselectAllDisks">❌ 清除选择</button>
            <span class="selection-count">已选 {{ selectedDevices.size }}/{{ diskList.length }}</span>
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
            正在查找移动 U 盘... 请插入 U 盘或点击刷新
          </div>
        </div>

        <button class="btn-secondary refresh-btn" @click="refreshDisks">
          🔄 重新扫描 U 盘
        </button>
      </section>

      <!-- Right: Deployment & Testing Panel -->
      <section class="glass-card section-card">
        <h2>启动盘制作与模拟测试</h2>
        <p class="section-desc" v-if="activeMode === 'cloud'">
          <strong>纯净 iPXE 云引导</strong> • 极速初始化双分区并写入多架构 iPXE 网络引导固件。
        </p>
        <p class="section-desc" v-else>
          集成 Ventoy 核心 + UniBoot 专属暗色主题 & iPXE 网络扩展，支持放置数 GB 大 ISO 镜像。
        </p>

        <!-- Filesystem Selection for Mode A & Mode B (Hidden when upgrading an existing Ventoy/UniBoot drive) -->
        <div v-if="!isNonDestructive" class="fs-selector">
          <label class="fs-label">主数据区格式 (File System):</label>
          <select v-model="selectedFsType" class="fs-select">
            <option value="exFAT">exFAT (默认推荐 • 支持 >4GB 单文件大 ISO)</option>
            <option value="NTFS">NTFS (Windows 极速原生格式)</option>
            <option value="FAT32">FAT32 (老旧机器全兼容 • 4GB单文件限制)</option>
            <option value="ext4">ext4 (Linux 专属文件系统)</option>
          </select>
        </div>

        <!-- Ventoy CLI Pre-flight Requirement Notice Banner (Mode A) -->
        <div v-if="activeMode === 'hybrid' && !isNonDestructive && !ventoyStatus.valid" class="ventoy-warning-card">
          <span class="warning-card-icon">⚠️</span>
          <div class="warning-card-body">
            <div class="warning-card-title">全新模式 A (Ventoy 双模盘) 前置限制与说明</div>
            <div class="warning-card-message">{{ ventoyStatus.message || '全新制作模式 A 需依赖 Ventoy CLI 可执行文件。' }}</div>
          </div>
          <button class="btn-secondary btn-sm" @click="isSettingsOpen = true">
            ⚙️ 配置 / 校验
          </button>
        </div>

        <!-- Safe Mode Notice Banner when upgrading an existing Ventoy/UniBoot drive -->
        <div v-if="isNonDestructive" class="safe-mode-notice">
          <span class="safe-notice-icon">🛡️</span>
          <div class="safe-notice-content">
            <div class="safe-notice-title">
              {{ activeMode === 'cloud' ? '检测到现有 Ventoy/UniBoot 盘 (模式 B 仅刷新 ESP 引导区)' : '检测到现有的 Ventoy 启动盘 (免格式化无损更新)' }}
            </div>
            <div class="safe-notice-desc">
              {{ activeMode === 'cloud'
                  ? '模式 B 坚持标准 UNIBOOT 双分区架构。部署将自动无损刷新 ESP 引导区直达 iPXE 云菜单，绝不抹擦或挪动主数据区原有文件与 ISO！'
                  : '无需选择主数据区格式。模式 A 自动保留所有现有文件与 ISO 镜像（绝不挪动原文件位置），全自动无损注入 UniBoot 暗色主题与 iPXE 云引导！' }}
            </div>
          </div>
        </div>

        <!-- Local ISO/IMG Image Source Selection Card (Mode A) -->
        <div v-if="activeMode === 'hybrid'" class="iso-card">
          <div class="iso-card-header">
            <div class="iso-title-group">
              <h3>💿 本地系统镜像源 (ISO / IMG / WIM / VHD)</h3>
              <span class="iso-subtitle">支持单选与多选系统镜像。一键制作完成将自动写入 `/UNIBOOT/iso/` 目录供 Ventoy / UniBoot 直接挂载。</span>
            </div>
            <button class="btn-secondary add-iso-btn" @click="handleSelectIsoFiles">
              ➕ 添加镜像文件
            </button>
          </div>

          <div class="iso-list-container">
            <div v-if="selectedIsoFiles.length === 0" class="iso-empty-state" @click="handleSelectIsoFiles">
              <span class="empty-icon">📥</span>
              <div class="empty-text">点击添加镜像文件 (支持单选与批量多选)</div>
              <div class="empty-subtext">支持 .iso, .wim, .img, .vhd, .vhdx, .vti, .efi, .bin, .xz, .gz, .raw 等 Ventoy 全格式</div>
            </div>

            <div v-else class="iso-file-list">
              <div v-for="(file, index) in selectedIsoFiles" :key="index" class="iso-file-item">
                <span class="iso-file-icon">{{ getFileIcon(file.name) }}</span>
                <div class="iso-file-info">
                  <div class="iso-file-name" :title="file.path">{{ file.name }}</div>
                  <div class="iso-file-path">{{ file.path }}</div>
                </div>
                <button class="iso-remove-btn" title="移除此文件" @click="removeIsoFile(index)">✕</button>
              </div>
            </div>

            <div v-if="selectedIsoFiles.length > 0" class="iso-footer">
              <span class="iso-count-summary">已选 <strong>{{ selectedIsoFiles.length }}</strong> 个系统镜像源文件</span>
              <button class="btn-text-danger" @click="clearIsoFiles">清空列表</button>
            </div>
          </div>
        </div>

        <div class="deploy-box">
          <div class="selected-target">
            <span>目标设备:</span>
            <strong v-if="selectionMode === 'single'">
              {{ selectedDisk ? selectedDisk.name + ' (' + selectedDisk.device + ')' : '未选择 U 盘' }}
            </strong>
            <strong v-else>
              {{ selectedDevices.size > 0 ? `已选中 ${selectedDevices.size} 块 U 盘` : '未选择 U 盘' }}
            </strong>
          </div>

          <ProgressBar 
            v-if="isDeploying" 
            label="正在写入引导与固件包..." 
            :progress="deployProgress" 
          />

          <button 
            class="btn-primary deploy-btn" 
            :class="{ 'safe-btn': isNonDestructive }"
            :disabled="isDeployDisabled || isDeploying"
            @click="openDeployConfirm"
          >
            {{ isDeploying ? '正在写入引导固件...' : (isNonDestructive ? '🛡️ 开始无损更新 (保留数据)' : (selectionMode === 'batch' ? `开始批量制作 (${selectedDevices.size} 块 U 盘)` : '开始制作启动盘')) }}
          </button>
        </div>

        <!-- QEMU Preview -->
        <div class="qemu-box">
          <div class="qemu-header">
            <h3>QEMU 引导模拟测试</h3>
            <span class="badge" :class="qemuStatus.installed ? 'success' : 'muted'">
              {{ qemuStatus.installed ? '已检测到 QEMU' : '未检测到 QEMU' }}
            </span>
          </div>
          <p class="qemu-desc">
            校验目标: 
            <strong v-if="activeQemuTargetDevice" class="target-highlight">
              {{ activeQemuTargetName }} ({{ activeQemuTargetDevice }})
            </strong>
            <span v-else class="target-warn">
              ⚠️ 未选择 U 盘（请在左侧列表中点击选择要测试的 U 盘）
            </span>
          </p>
          <button 
            class="btn-secondary" 
            :disabled="isQemuDisabled" 
            :title="qemuDisabledReason"
            @click="launchQEMU"
          >
            {{ isLaunchingQemu ? '⏳ 正在拉起 QEMU 虚拟机...' : '▶ 运行 QEMU 启动测试' }}
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
      :currentProxy="currentGithubProxy"
      @close="isSettingsOpen = false"
      @save="onSaveSettings"
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
        githubProxy: proxyUrl,
        fileSystem: payload.fileSystem || selectedFsType.value,
        proxyProtocol: payload.proxyProtocol || 'direct',
        proxyHost: payload.proxyHost || '',
        proxyPort: Number(payload.proxyPort) || 0,
        proxyUser: payload.proxyUser || '',
        proxyPassword: payload.proxyPassword || '',
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

async function openDeployConfirm() {
  let targets: string[] = [];
  if (selectionMode.value === 'single') {
    if (!selectedDisk.value) return;
    targets = [selectedDisk.value.device];
  } else {
    targets = Array.from(selectedDevices.value);
    if (targets.length === 0) return;
  }

  if (activeMode.value === 'hybrid' && !isNonDestructive.value) {
    await checkVentoyStatus();
    if (!ventoyStatus.value.valid) {
      showToast(ventoyStatus.value.message || '无法制作 Mode A：未检测到有效的 Ventoy CLI 程序', 'error');
      return;
    }
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

const isDeployDisabled = computed(() => {
  if (selectionMode.value === 'single') {
    return !selectedDisk.value;
  }
  return selectedDevices.value.size === 0;
});

const isSelectedVentoyDisk = computed(() => {
  if (selectionMode.value === 'single' && selectedDisk.value) {
    if (activeMode.value === 'cloud') {
      return true;
    }
    const name = (selectedDisk.value.name || '').toUpperCase();
    const status = (selectedDisk.value.bootStatus || '').toUpperCase();
    if (status.includes('MODE B') || status.includes('CLOUD PURE')) {
      return false;
    }
    return name.includes('VENTOY') || status.includes('VENTOY');
  }
  return false;
});

const isNonDestructive = computed(() => {
  if (selectionMode.value === 'single') {
    return isSelectedVentoyDisk.value;
  }
  if (selectedDevices.value.size === 0) return false;
  return Array.from(selectedDevices.value).some((dev: string) => {
    const d = diskList.value.find((disk: DiskInfo) => disk.device === dev);
    if (!d) return false;
    const name = (d.name || '').toUpperCase();
    const status = (d.bootStatus || '').toUpperCase();
    return name.includes('VENTOY') || status.includes('VENTOY') || status.includes('UNIBOOT');
  });
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
</style>

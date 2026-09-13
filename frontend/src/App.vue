<template>
  <div class="app-container">
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
          @click="activeMode = 'cloud'"
        >
          ⚡ 模式 B (1秒极速云安装盘)
        </button>
        <button 
          class="tab-btn" 
          :class="{ active: activeMode === 'hybrid' }"
          @click="activeMode = 'hybrid'"
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
        <h2>部署与预览操作</h2>
        <p class="section-desc" v-if="activeMode === 'cloud'">
          <strong>100% 脱离 Ventoy</strong> • 原生格式化 FAT32 并写入 64MB 多架构 iPXE 固件，1 秒完成！
        </p>
        <p class="section-desc" v-else>
          集成 Ventoy 核心 + UniBoot 专属暗色主题 & iPXE 网络扩展，支持放置数 GB 大 ISO 镜像。
        </p>

        <!-- Filesystem Selection for Mode A & Mode B -->
        <div class="fs-selector">
          <label class="fs-label">主数据区格式 (File System):</label>
          <select v-model="selectedFsType" class="fs-select">
            <option value="exFAT">exFAT (默认推荐 • 支持 >4GB 单文件大 ISO)</option>
            <option value="NTFS">NTFS (Windows 极速原生格式)</option>
            <option value="FAT32">FAT32 (老旧机器全兼容 • 4GB单文件限制)</option>
            <option value="ext4">ext4 (Linux 专属文件系统)</option>
          </select>
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
            label="正在批量写入引导与固件包..." 
            :progress="deployProgress" 
          />

          <button 
            class="btn-primary deploy-btn" 
            :disabled="isDeployDisabled || isDeploying"
            @click="openDeployConfirm"
          >
            {{ isDeploying ? '正在极速烧录中...' : (selectionMode === 'batch' ? `开始批量烧录 (${selectedDevices.size} 块 U 盘)` : '开始 1 秒部署写入') }}
          </button>
        </div>

        <!-- QEMU Preview -->
        <div class="qemu-box">
          <div class="qemu-header">
            <h3>QEMU 启动预览 (可选)</h3>
            <span class="badge" :class="qemuStatus.installed ? 'success' : 'muted'">
              {{ qemuStatus.installed ? '已检测到 QEMU' : '未检测到 QEMU' }}
            </span>
          </div>
          <p class="qemu-desc">烧录完成后，可在当前桌面直接拉起 QEMU 虚拟机校验 U 盘引导环境。</p>
          <button class="btn-secondary" :disabled="!selectedDisk" @click="launchQEMU">
            ▶ 拉起 QEMU 模拟器测试
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

function openDeployConfirm() {
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

const isDeployDisabled = computed(() => {
  if (selectionMode.value === 'single') {
    return !selectedDisk.value;
  }
  return selectedDevices.value.size === 0;
});

function setSelectionMode(mode: 'single' | 'batch') {
  selectionMode.value = mode;
  if (mode === 'single' && diskList.value.length > 0 && !selectedDisk.value) {
    selectedDisk.value = diskList.value[0];
  }
}

function selectAllDisks() {
  selectedDevices.value = new Set(diskList.value.map(d => d.device));
}

function deselectAllDisks() {
  selectedDevices.value.clear();
}

function onDiskSelect(disk: DiskInfo) {
  if (selectionMode.value === 'single') {
    selectedDisk.value = disk;
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
    } catch (e) {
      console.error(e);
      diskList.value = [];
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
      if (targets.length === 1) {
        let res: any;
        if (activeMode.value === 'cloud') {
          res = await window.go.main.App.DeployModeB(targets[0], selectedFsType.value);
        } else {
          res = await window.go.main.App.DeployModeA(targets[0], selectedFsType.value);
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
          resList = await window.go.main.App.DeployModeABatch(targets, selectedFsType.value);
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


function launchQEMU() {
  alert(`正在启动 QEMU 模拟器校验: ${selectedDisk.value?.device}`);
}

let diskPollTimer: number | undefined;

onMounted(() => {
  loadConfig();
  refreshDisks();
  checkQemu();

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
  grid-template-columns: 1fr 1.15fr;
  gap: 2rem;
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
</style>
